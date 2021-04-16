package public

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"o-ten/global"
	"o-ten/model"
	"time"
)

type tvStruct struct {
	EpisodesInfo string `json:"episodes_info"`
	Rate string `json:"rate"`
	Title string `json:"title"`
	Url string `json:"url"`
	Playable bool `json:"palyable"`
	Cover string `json:"cover"`
	ID string `json:"id"`
	ISNew bool `json:"is_new"`
}

func fetchDoubanTvHot() []tvStruct {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", "https://movie.douban.com/j/search_subjects?type=tv&tag=%E7%83%AD%E9%97%A8&sort=recommend&page_limit=20&page_start=0", nil)
	req.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36`)
	req.Header.Add("Host", "movie.douban.com")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("fetchHot", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	m := make(map[string][]tvStruct)
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil
	}

	return m["subjects"]
}

func IndexController(c *gin.Context) {
	// 获取 DB
	db := global.DB
	var result interface{}
	result = fetchDoubanTvHot()

	type tvDbStruct struct {
		Name string `json:"title"`
		Url string `json:"url"`
		Playable bool `json:"palyable"`
		Cover string `json:"cover"`
		ID string `json:"id"`
		ISNew bool `json:"is_new"`
	}

	var televisions []tvDbStruct

	if result == nil {
		db.Model(&model.Television{}).Order("updated_at desc, name").Limit(20).Scan(&televisions)
		result = televisions
	}

	c.JSON(
		200,
		gin.H{
			"success": true,
			"data": gin.H{
				"hots": result,
			},
		},
	)
}
