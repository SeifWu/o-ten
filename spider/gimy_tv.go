package spider

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/gocolly/redisstorage"
	"log"
	"o-ten/global"
	"o-ten/model"
	"strconv"
	"strings"
)

type MediaData struct {
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	PlayPath []PlayPath
}

type PlayPathDetail struct {
	Name string
	Url  string `json:"url"`
}

type PlayPath struct {
	Key  string
	Name string
	List []PlayPathDetail
}

// 剧迷 Tv 爬虫
func GimyTvSpider() {
	// 获取 DB
	DB := global.DB
	// 创建主采集器
	c := colly.NewCollector(
		colly.AllowedDomains("gimytv.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	// 创建 Redis Storage
	storage := &redisstorage.Storage{
		Address:  "127.0.0.1:6379",
		Password: "",
		DB:       0,
		Prefix:   "gimyTvSpider",
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
	// 播放页面收集器
	playerCollector := detailCollector.Clone()
	// 查看请求的地址
	c.OnRequest(func(request *colly.Request) {
		log.Println("开始请求 =>", request.URL)
	})

	// 视频列表选择器
	videoListDomSelector := `
		.myui-panel.active.myui-panel-bg.clearfix > .myui-panel-box.clearfix > .myui-panel_bd > .myui-vodlist.clearfix
	`

	c.OnHTML(videoListDomSelector, func(element *colly.HTMLElement) {
		// 视频列表选择器
		videoLinkSelector := "ul > li > div.myui-vodlist__box > a.myui-vodlist__thumb.lazyload"
		element.ForEach(videoLinkSelector, func(i int, videoCardElement *colly.HTMLElement) {
			// 获取视频名称
			title := videoCardElement.Attr("title")

			// 获取视频封面
			cover := videoCardElement.Attr("data-original")

			// 视频详情页网站
			requestDetailURL := videoCardElement.Request.AbsoluteURL(videoCardElement.Attr("href"))
			// TODO 查询数据库 requestDetailURL存在且已完成

			categoryName := ""
			currentUrl := element.Request.URL.String()
			if strings.HasPrefix(currentUrl, "https://gimytv.com/genre/2") {
				categoryName = "电视剧"
			}
			if strings.HasPrefix(currentUrl, "https://gimytv.com/genre/1") {
				categoryName = "电影"
			}
			if strings.HasPrefix(currentUrl, "https://gimytv.com/genre/4") {
				categoryName = "动漫"
			}

			// 通过Context上下文对象将【c 采集器】采集到的数据传递到【detailCollector 采集器】
			ctx := colly.NewContext()
			ctx.Put("title", title)
			ctx.Put("cover", cover)
			ctx.Put("categoryName", categoryName)

			detailCollector.Request("GET", requestDetailURL, nil, ctx, nil)
		})
	})

	c.OnHTML("ul.myui-page.text-center.clearfix", func(element *colly.HTMLElement) {
		nextPage := element.ChildAttr("li:nth-last-of-type(2) > a", "href")
		nextPageUrl := element.Request.AbsoluteURL(nextPage)
		currentUrl := element.Request.URL.String()

		if currentUrl != nextPageUrl {
			_ = c.Visit(nextPageUrl)
		}
	})

	// 解析视频详情页面
	detailCollector.OnHTML("body", func(detailElement *colly.HTMLElement) {
		title := detailElement.Request.Ctx.Get("title")
		cover := detailElement.Request.Ctx.Get("cover")
		categoryName := detailElement.Request.Ctx.Get("categoryName")

		// 其他信息 css 选择器
		otherInfoSelector := `
			.container > div.row:nth-of-type(1) > div.myui-panel.myui-panel-bg.clearfix > div.myui-panel-box.clearfix >
			div.col-xs-1:nth-of-type(2) > .myui-content__detail > p:nth-of-type(1)
		`
		// 视频类型
		mediaType := detailElement.DOM.Find(otherInfoSelector + " > a:nth-of-type(1)").Text()
		// 地区
		region := detailElement.DOM.Find(otherInfoSelector + " > a:nth-of-type(2)").Text()
		// 年份
		year := detailElement.DOM.Find(otherInfoSelector + " > a:nth-of-type(3)").Text()

		descSelector := `
			.container > div.row:nth-of-type(2) > .col-md-wide-7.col-xs-1.padding-0 >
			div#desc > .myui-panel-box.clearfix >
			.myui-panel_bd > .col-pd.text-collapse.content > span.sketch.content
		`
		// 简介
		introduction := detailElement.DOM.Find(descSelector).Text()
		// 播放地址 - 视频来源选择器
		videoSourceSelector := `
			.container > div.row:nth-of-type(2) > .col-md-wide-7.col-xs-1.padding-0 >
			div.myui-panel.myui-panel-bg.clearfix:nth-of-type(2) > .myui-panel-box.clearfix >
			.myui-panel_hd .myui-panel__head.active.bottom-line.clearfix > .nav.nav-tabs.active > li
		`

		var playerList []PlayPath
		detailElement.ForEach(videoSourceSelector, func(i int, childElement *colly.HTMLElement) {
			key := childElement.ChildAttr("a", "href")
			name := childElement.ChildText("a")
			tempPlayerPath := PlayPath{
				Key:  key,
				Name: name,
			}
			playerList = append(playerList, tempPlayerPath)
		})

		for _, v := range playerList {
			querySelector := fmt.Sprintf(".tab-content.myui-panel_bd %s ul.myui-content__list.sort-list.scrollbar.clearfix li", v.Key)
			detailElement.ForEach(querySelector, func(i int, liElement *colly.HTMLElement) {
				url := liElement.ChildAttr("a", "href")
				name := liElement.ChildText("a")
				requestDetailURL := liElement.Request.AbsoluteURL(url)

				ctx := colly.NewContext()
				ctx.Put("title", title)
				ctx.Put("cover", cover)
				ctx.Put("mediaType", mediaType)
				ctx.Put("region", region)
				ctx.Put("year", year)
				ctx.Put("introduction", introduction)
				ctx.Put("sourceUrl", detailElement.Request.URL.String())
				ctx.Put("sourceName", v.Name)
				ctx.Put("videoName", name)
				ctx.Put("categoryName", categoryName)

				playerCollector.Request("GET", requestDetailURL, nil, ctx, nil)

				//_ = playerCollector.Visit(requestDetailURL)
			})
		}
	})

	playerCollector.OnHTML(
		".myui-player.clearfix .container .row .myui-player__item.clearfix .col-md-wide-75.padding-0.relative.clearfix .myui-player__box .myui-player__video.embed-responsive.clearfix script:first-of-type",
		func(playerElement *colly.HTMLElement) {
			title := playerElement.Request.Ctx.Get("title")
			cover := playerElement.Request.Ctx.Get("cover")
			mediaType := playerElement.Request.Ctx.Get("mediaType")
			region := playerElement.Request.Ctx.Get("region")
			year := playerElement.Request.Ctx.Get("year")
			introduction := playerElement.Request.Ctx.Get("introduction")
			sourceUrl := playerElement.Request.Ctx.Get("sourceUrl")
			sourceName := playerElement.Request.Ctx.Get("sourceName")
			videoName := playerElement.Request.Ctx.Get("videoName")
			categoryName := playerElement.Request.Ctx.Get("categoryName")
			dat := playerElement.Text
			jsonData := dat[strings.Index(dat, "{"):]
			data := &PlayPathDetail{}
			err := json.Unmarshal([]byte(jsonData), data)
			if err != nil {
				log.Fatal(err)
			}

			log.Println("------------------------ ", title, " ---------------------------------")
			log.Println("封面地址: ", cover)
			log.Println("播放页地址: ", sourceUrl)
			log.Println("类别：", categoryName)
			log.Println("分类: ", mediaType)
			log.Println("地区: ", region)
			log.Println("年份: ", year)
			log.Println("简介: ", introduction)
			log.Println("视频来源地址: ", sourceName)
			log.Println("视频标题: ", videoName)
			log.Println("解析后的 url: ", data.Url)
			log.Println()

			// 1. 查询是否存在该影视
			var media model.Media
			DB.Debug().Model(&model.Media{}).Where("title = ?", title).First(&media)

			// 不存在时新建
			if media.ID == 0 {
				mediaYear, err := strconv.Atoi(year)
				if err != nil {
					mediaYear = 0
				}
				media = model.Media{
					Title:        title,
					Cover:        cover,
					Type:         mediaType,
					Region:       region,
					Year:         mediaYear,
					Introduction: introduction,
					CategoryName: categoryName,
				}
				_ = DB.Debug().Create(&media)
			}

			// 查询来源是否存在
			var source model.Source
			DB.Debug().Where(&model.Source{
				Title:      "剧迷TV",
				SourceName: sourceName,
				SourceUrl:  sourceUrl,
				MediaId:    media.ID,
			}).First(&source)
			// 来源不存在时新建
			if source.ID == 0 {
				source = model.Source{
					Title:      "剧迷TV",
					SourceName: sourceName,
					SourceUrl:  sourceUrl,
					MediaId:    media.ID,
				}
				_ = DB.Debug().Create(&source)
			}

			// 查询视频数量用于排序
			var count int64
			DB.Model(&model.Video{}).Where(&model.Video{SourceId: source.ID}).Count(&count)

			var video model.Video
			DB.Debug().Where(&model.Video{Src: data.Url, SourceId: source.ID}).First(&video)
			if video.ID == 0 {
				video = model.Video{
					Title:    videoName,
					Src:      data.Url,
					SourceId: source.ID,
					OrderSeq: float64(count),
				}
				_ = DB.Debug().Create(&video)
			}
		},
	)

	urls := []string{
		"https://gimytv.com/genre/1--------1---.html",
		"https://gimytv.com/genre/2--------1---.html",
		"https://gimytv.com/genre/4--------1---.html",
	}

	q, _ := queue.New(
		2,
		storage,
	)

	// add URLs to the queue
	for _, u := range urls {
		_ = q.AddURL(u)
	}
	// consume requests
	_ = q.Run(c)
}
