package spider

import (
	"encoding/json"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/gocolly/redisstorage"
	"log"
	"strings"
)

type playerDetail struct {
	Url  string `json:"url"`
}

// fiitvr.com 爬虫
func FiitvrComSpider() {
	// 创建主采集器
	c := colly.NewCollector(
		colly.AllowedDomains("fiitvr.com", "www.fiitvr.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

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

	// 查看请求的地址
	c.OnRequest(func(request *colly.Request) {
		log.Println("开始请求 =>", request.URL)
	})

	// 视频列表选择器
	videoListDomSelector := `
		div#show_page > .container > .row > .pannel.clearfix > ul.vodlist.vodlist_wi
	`

	c.OnHTML(videoListDomSelector, func(listElement *colly.HTMLElement) {
		liDonSelector := "li.vodlist_item > a"
		listElement.ForEach(liDonSelector, func(i int, videoCardElement *colly.HTMLElement) {
			// 详情链接
			videoDetailUrl := videoCardElement.Request.AbsoluteURL(videoCardElement.Attr("href"))
			// 通过Context上下文对象将【c 采集器】采集到的数据传递到【detailCollector 采集器】
			ctx := colly.NewContext()
			ctx.Put("sourceLink", videoDetailUrl)
			log.Println("详情页: ", videoDetailUrl)
			detailCollector.Request("GET", videoDetailUrl, nil, ctx, nil)

		})
	})

	// TODO 请求下一页

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
			// 视频名称
			videoName, _ := infoDom.Attr("title")
			// 封面
			videoCover, _ := infoDom.Attr("data-original")
			// 年份
			year := bodyElement.DOM.Find(otherInfoSelector + "li.data:nth-of-type(1) > a:nth-of-type(1)").Text()
			// 地区
			region := bodyElement.DOM.Find(otherInfoSelector + "li.data:nth-of-type(1) > a:nth-of-type(2)").Text()
			// 更新 flag
			updatedFlag := bodyElement.DOM.Find(otherInfoSelector + "li.data:nth-of-type(2)").Text()

			// TODO 创建 Television => 传递 ID

			playerListSelector := `
				.container > div.left_row.fl > div.pannel.clearfix#bofy > div.play_source > div.play_list_box.hide > 
				div.playlist_full > ul.content_playlist.clearfix > li > a
			`
			bodyElement.ForEach(playerListSelector, func(i int, playerListElement *colly.HTMLElement) {
				link := playerListElement.Attr("href")
				hrefArray := strings.Split(link, "-")
				// 资源名
				sourceName := strings.Join(hrefArray[:2], "-")
				// 视频名称
				videoName := playerListElement.Text

				// 播放页地址
				playerDetailUrl := playerListElement.Request.AbsoluteURL(link)

				ctx := colly.NewContext()
				ctx.Put("sourceName", sourceName)
				log.Println("详情页: ", playerDetailUrl)
				playerCollector.Request("GET", playerDetailUrl, nil, ctx, nil)

				log.Println("视频名称：", videoName, sourceName)
			})

			log.Println(videoName, videoCover, year, region, updatedFlag)
		},
	)

	playerCollector.OnHTML(
		"body > div#play_page > .play_boxbg > .container > .play_box.play_video > .left_row.fl > div.player_video > script:nth-of-type(1)",
		func(element *colly.HTMLElement) {
			scriptString := element.Text
			jsonData := scriptString[strings.Index(scriptString, "{"):]
			data := &playerDetail{}
			err := json.Unmarshal([]byte(jsonData), data)
			if err !=nil {
				// TODO 发送错误日志
				log.Fatal(err)
			}

			log.Println("地址:", data.Url)
		},
	)

	//https://www.fiitvr.com/gov1---2/by/time.html
	urls := []string{
		"https://www.fiitvr.com/gov1---2/by/time.html",
	}
	q, _ := queue.New(2, storage)
	// 添加 URL 到队列
	for _, u := range urls {
		_ = q.AddURL(u)
	}
	// 使用请求
	_ = q.Run(c)
}
