package spider

import (
	"encoding/json"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/gocolly/redisstorage"
	"log"
	"net"
	"net/http"
	"o-ten/global"
	"o-ten/model"
	"strconv"
	"strings"
	"time"
)

type playerDetail struct {
	Url string `json:"url"`
}

// fiitvr.com 爬虫
func FiitvrComSpider() {
	// 获取 DB
	DB := global.DB

	// 创建主采集器
	c := colly.NewCollector(
		colly.AllowedDomains("fiitvr.com", "www.fiitvr.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
		colly.AllowURLRevisit(),
	)
	c.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,          // 超时时间
			KeepAlive: 30 * time.Second,          // keepAlive 超时时间
		}).DialContext,
		MaxIdleConns:          100,               // 最大空闲连接数
		IdleConnTimeout:       90 * time.Second,  // 空闲连接超时
		TLSHandshakeTimeout:   10 * time.Second,  // TLS 握手超时
		ExpectContinueTimeout: 1 * time.Second,
	})

	// 创建 Redis Storage
	storage := &redisstorage.Storage{
		Address:  "127.0.0.1:6379",
		Password: "",
		DB:       0,
		Prefix:   "fiitvrComSpider",
	}

	// 向主采集器添加 Redis storage
	err := c.SetStorage(storage)
	if err != nil {
		panic(err)
	}

	// 删除以前的数据存储
	if err := storage.Clear(); err != nil {
		log.Fatal("删除以前的数据存储出现错误：", err)
	}
	defer storage.Client.Close()

	// 新建视频详情收集器
	detailCollector := c.Clone()
	// 新建视频播放页收集器
	playerCollector := c.Clone()

	// 视频列表选择器
	videoListDomSelector := `
		div#show_page > .container > .row > .pannel.clearfix > ul.vodlist.vodlist_wi
	`

	c.OnError(func(response *colly.Response, err error) {
		log.Println("访问出现错误：", err)
		// TODO 日志报错
	})

	c.OnHTML(videoListDomSelector, func(listElement *colly.HTMLElement) {
		liDonSelector := "li.vodlist_item > a"
		listElement.ForEach(liDonSelector, func(i int, videoCardElement *colly.HTMLElement) {
			// 详情链接
			videoDetailUrl := videoCardElement.Request.AbsoluteURL(videoCardElement.Attr("href"))
			// 通过Context上下文对象将【c 采集器】采集到的数据传递到【detailCollector 采集器】
			ctx := colly.NewContext()
			ctx.Put("sourceLink", videoDetailUrl)
			detailCollector.Request("GET", videoDetailUrl, nil, ctx, nil)
		})
	})

	c.OnHTML("#show_page > .container > .row > .pannel.clearfix > ul", func(element *colly.HTMLElement) {
		nextPage := element.ChildAttr("li:nth-last-of-type(2) > a", "href")
		nextPageUrl := element.Request.AbsoluteURL(nextPage)
		currentUrl := element.Request.URL.String()

		if currentUrl != nextPageUrl && strings.Contains(nextPageUrl, "/by/time.html") {
			log.Println("当前页", currentUrl, "下一页：", nextPageUrl)
			log.Println()
			_ = c.Visit(nextPageUrl)
		}
	})

	detailCollector.OnHTML(
		`body`,
		func(bodyElement *colly.HTMLElement) {
			infoTitleSelector := `
				.hot_banner > .detail_list_box > .detail_list > .content_box.clearfix > .content_thumb.fl > a.vodlist_thumb.lazyload
			`
			otherInfoSelector := `
				.hot_banner > .detail_list_box > .detail_list > .content_box.clearfix > .content_detail.content_min.fl > ul
			`
			infoDom := bodyElement.DOM.Find(infoTitleSelector)
			// 请求地址
			sourceUrl := bodyElement.Request.URL.String()
			// 视频名称
			televisionName, _ := infoDom.Attr("title")
			// 封面
			videoCover, _ := infoDom.Attr("data-original")
			// 年份
			year, _ := strconv.Atoi(bodyElement.DOM.Find(otherInfoSelector + "li.data:nth-of-type(1) > a:nth-of-type(1)").Text())
			// 地区
			region := bodyElement.DOM.Find(otherInfoSelector + "li.data:nth-of-type(1) > a:nth-of-type(2)").Text()
			// 分类
			categoryName := bodyElement.DOM.Find(otherInfoSelector + "li.data:nth-of-type(1) > a:nth-last-of-type(1)").Text()
			// 更新 flag
			updatedFlag := strings.Trim(bodyElement.DOM.Find(otherInfoSelector+"li.data:nth-of-type(2)").Text(), " ")

			var television model.Television
			DB.Model(&model.Television{}).Where("name = ?", televisionName).First(&television)
			if television.ID == 0 {
				television = model.Television{
					Name:         televisionName,
					Cover:        videoCover,
					Year:         year,
					Region:       region,
					CategoryName: categoryName,
				}
				_ = DB.Create(&television)
			}

			playerListSelector := `
				.container > div.left_row.fl > div.pannel.clearfix#bofy > div.play_source > div.play_list_box.hide > 
				div.playlist_full > ul.content_playlist.clearfix > li > a
			`

			type sourceStruct struct {
				VideoName string
				PlayerDetailUrl string
			}
			var sourceNames []string
			sources := make(map[string][]sourceStruct)

			bodyElement.ForEach(playerListSelector, func(i int, playerListElement *colly.HTMLElement) {
				link := playerListElement.Attr("href")
				hrefArray := strings.Split(link, "-")
				// 资源名
				sourceName := strings.Join(hrefArray[:2], "-")
				// 视频名称
				videoName := playerListElement.Text
				// 播放页地址
				playerDetailUrl := playerListElement.Request.AbsoluteURL(link)

				if sources[sourceName] == nil {
					sourceNames = append(sourceNames, sourceName)
					sources[sourceName] = []sourceStruct {{PlayerDetailUrl: playerDetailUrl, VideoName: videoName}}
				} else {
					sources[sourceName] = append(sources[sourceName], sourceStruct {
						VideoName: videoName,
						PlayerDetailUrl: playerDetailUrl,
					})
				}
			})

			for _, sourceName := range sourceNames {
				var source model.Source
				// 当前时间
				now := time.Now()
				// 查询 source 是否存在
				DB.Where(&model.Source{Name: sourceName, Domain: "fiitvr.com", SourceUrl: sourceUrl}).First(&source)
				// source 不存在时
				if source.ID == 0 {
					source = model.Source{
						Name:          sourceName,
						OwnerID:       television.ID,
						OwnerType:     "television",
						Domain:        "fiitvr.com",
						SourceUrl:     sourceUrl,
						UpdatedFlag:   updatedFlag,
						LastCollectAt: &now,
					}
					DB.Model(&model.Source{}).Create(&source)
				}

				// 查询 video 数量是否与 sources map 下数量相等
				// 不相等时需要收集
				// TODO source model 添加是否收集标识
				var count int64
				DB.Model(&model.Video{}).Where(&model.Video{SourceID: source.ID}).Count(&count)
				if count != int64(len(sources[sourceName])) {
					sourceIDString := strconv.Itoa(int(source.ID))
					for _, sourceDetail := range sources[sourceName] {
						ctx := colly.NewContext()
						ctx.Put("sourceID", sourceIDString)
						ctx.Put("videoName", sourceDetail.VideoName)
						playerCollector.Request("GET", sourceDetail.PlayerDetailUrl, nil, ctx, nil)
					}
				} else {
					log.Println("跳过了：", televisionName, sourceName)
				}

			}
		},
	)

	playerCollector.OnHTML(
		"body > div#play_page > .play_boxbg > .container > .play_box.play_video > .left_row.fl > div.player_video > script:nth-of-type(1)",
		func(element *colly.HTMLElement) {
			sourceID := element.Request.Ctx.Get("sourceID")
			videoName := element.Request.Ctx.Get("videoName")
			scriptString := element.Text
			jsonData := scriptString[strings.Index(scriptString, "{"):]
			data := &playerDetail{}
			err := json.Unmarshal([]byte(jsonData), data)
			if err != nil {
				// TODO 发送错误日志
				log.Fatal(err)
			}

			// 查询视频数量用于排序
			sourceIDInt64, _ := strconv.ParseInt(sourceID, 10, 64)
			var count int64
			DB.Model(&model.Video{}).Where(&model.Video{SourceID: sourceIDInt64}).Count(&count)

			var video model.Video
			DB.Where(&model.Video{
				Title:    videoName,
				SourceID: sourceIDInt64,
				Url:      data.Url,
			}).First(&video)

			if video.ID == 0 {
				video = model.Video{
					Title:     videoName,
					SourceID:  sourceIDInt64,
					Url:       data.Url,
					OrderSeq:  float64(count),
					SourceUrl: element.Request.URL.String(),
				}
				_ = DB.Create(&video)
			}
		},
	)

	urls := []string{
		"https://www.fiitvr.com/gov1---88/by/time.html",
		"https://www.fiitvr.com/gov2---6/by/time.html",
		"https://www.fiitvr.com/gov4---1/by/time.html",
	}
	q, _ := queue.New(2, storage)
	// 添加 URL 到队列
	for _, u := range urls {
		_ = q.AddURL(u)
	}
	// 使用请求
	_ = q.Run(c)
}
