package spider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/gocolly/redisstorage"
	"log"
	"o-ten/global"
	"o-ten/model"
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

	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains("gimytv.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	// create the redis storage
	storage := &redisstorage.Storage{
		Address:  "127.0.0.1:6379",
		Password: "",
		DB:       0,
		Prefix:   "gimyTvSpider",
	}

	// add storage to the collector
	err := c.SetStorage(storage)
	if err != nil {
		panic(err)
	}

	// delete previous data from storage
	if err := storage.Clear(); err != nil {
		log.Fatal(err)
	}

	// close redis client
	defer storage.Client.Close()
	// 新建视频详情收集器
	detailCollector := c.Clone()
	// 播放页面收集器
	playerCollector := detailCollector.Clone()
	// 查看请求的地址
	c.OnRequest(func(request *colly.Request) {
		fmt.Println("OnRequest: Visiting =>", request.URL)
	})

	// 拼接视频列表选择器，这样写法原因是，直接写字符串太长了
	var videoListDomSelectorBuffer bytes.Buffer
	videoListDomSelectorBuffer.WriteString(".myui-panel.active.myui-panel-bg.clearfix > ")
	videoListDomSelectorBuffer.WriteString(".myui-panel-box.clearfix > ")
	videoListDomSelectorBuffer.WriteString(".myui-panel_bd > ")
	videoListDomSelectorBuffer.WriteString(".myui-vodlist.clearfix ")
	videoListDomSelector := videoListDomSelectorBuffer.String()

	c.OnHTML(videoListDomSelector, func(element *colly.HTMLElement) {
		// 视频列表选择器
		videoLinkSelector := "ul > li > div.myui-vodlist__box > a.myui-vodlist__thumb.lazyload"
		element.ForEach(videoLinkSelector, func(i int, videoCardElement *colly.HTMLElement) {
			title := videoCardElement.Attr("title")
			cover := videoCardElement.Attr("data-original")

			// 视频详情页网站
			requestDetailURL := videoCardElement.Request.AbsoluteURL(videoCardElement.Attr("href"))
			// 通过Context上下文对象将【c 采集器】采集到的数据传递到【detailCollector 采集器】
			detailCollector.OnRequest(func(r *colly.Request) {
				r.Ctx.Put("title", title)
				r.Ctx.Put("cover", cover)
			})

			// 查询数据库给结果
			var media model.Media
			DB.Debug().Model(&model.Media{}).Where("title = ?", title).First(&media)

			// 只要是未完成就请求
			if !media.IsCompleted {
				_ = detailCollector.Visit(requestDetailURL)
				detailCollector.Wait()
			}

			_ = detailCollector.Visit(requestDetailURL)
			detailCollector.Wait()
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

	// 视频详情页选择器
	var videoDetailDomSelectorBuffer bytes.Buffer
	videoDetailDomSelectorBuffer.WriteString(".container > div.row:nth-of-type(2) > ")
	videoDetailDomSelectorBuffer.WriteString(".col-md-wide-7.col-xs-1.padding-0 > ")
	videoDetailDomSelectorBuffer.WriteString("div.myui-panel.myui-panel-bg.clearfix:nth-of-type(2) > ")
	videoDetailDomSelectorBuffer.WriteString(".myui-panel-box.clearfix")
	// 解析视频详情页面
	detailCollector.OnHTML(videoDetailDomSelectorBuffer.String(), func(detailElement *colly.HTMLElement) {
		title := detailElement.Request.Ctx.Get("title")
		cover := detailElement.Request.Ctx.Get("cover")

		fmt.Println(title, cover)
		var playerList []PlayPath
		// 视频来源选择器
		videoSourceSelector := ".myui-panel_hd .myui-panel__head.active.bottom-line.clearfix .nav.nav-tabs.active li"
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

				playerCollector.OnRequest(func(r *colly.Request) {
					r.Ctx.Put("title", title)
					r.Ctx.Put("cover", cover)
					r.Ctx.Put("sourceUrl", detailElement.Request.URL.String())
					r.Ctx.Put("sourceName", v.Name)
					r.Ctx.Put("videoName", name)
				})

				_ = playerCollector.Visit(requestDetailURL)
				playerCollector.Wait()
			})
		}
	})

	playerCollector.OnHTML(
		".myui-player.clearfix .container .row .myui-player__item.clearfix .col-md-wide-75.padding-0.relative.clearfix .myui-player__box .myui-player__video.embed-responsive.clearfix script:first-of-type",
		func(playerElement *colly.HTMLElement) {
			title := playerElement.Request.Ctx.Get("title")
			cover := playerElement.Request.Ctx.Get("cover")
			sourceUrl := playerElement.Request.Ctx.Get("sourceUrl")
			sourceName := playerElement.Request.Ctx.Get("sourceName")
			videoName := playerElement.Request.Ctx.Get("videoName")

			dat := playerElement.Text
			jsonData := dat[strings.Index(dat, "{"):]
			data := &PlayPathDetail{}
			err := json.Unmarshal([]byte(jsonData), data)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("------------------------ ", title, " ---------------------------------")
			fmt.Println("封面地址: ", cover)
			fmt.Println("播放页地址: ", sourceUrl)
			fmt.Println("视频来源地址: ", sourceName)
			fmt.Println("视频标题: ", videoName)
			fmt.Println("解析后的 url: ", data.Url)
			fmt.Println()

			// 1. 查询是否存在该影视
			var media model.Media
			DB.Debug().Model(&model.Media{}).Where("title = ?", title).First(&media)

			newMedia := model.Media{}
			// 不存在时新建
			if media.ID == 0 {
				newMedia.Title = title
				newMedia.Cover = cover
				_ = DB.Create(&newMedia)
			}

			// 查询来源是否存在
			var source model.Source
			DB.Debug().Where(&model.Source{
				Title:      "剧迷TV",
				SourceName: sourceName,
				SourceUrl:  sourceUrl,
				MediaId:    newMedia.ID,
			}).First(&source)
			// 来源不存在时新建
			if source.ID == 0 {
				source = model.Source{
					Title:      "剧迷TV",
					SourceName: sourceName,
					SourceUrl:  sourceUrl,
					MediaId:    newMedia.ID,
				}
				_ = DB.Create(&source)
			}

			// 查询视频数量用于排序
			var count int64
			DB.Model(&model.Video{}).Where(&model.Video{SourceId: source.ID}).Count(&count)

			var video model.Video
			DB.Debug().Where(&model.Video{Src: data.Url, SourceId: source.ID}).First(&video)
			if video.ID == 0 {
				video = model.Video{
					Title: videoName,
					Src: data.Url,
					SourceId: source.ID,
					OrderSeq: float64(count),
				}
				_ = DB.Create(&video)
			}
		},
	)

	urls := []string{
		//"https://gimytv.com/genre/1--------738---.html",
		"https://gimytv.com/genre/2--------327---.html",
		//"https://gimytv.com/genre/4--------110---.html",
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
	c.Wait()
}
