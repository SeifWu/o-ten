package view

import "github.com/gin-gonic/gin"

type navStrut struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// HomePage 首页
func HomePage(c *gin.Context) {
	nav := make([]navStrut, 0)

	nav = append(nav, navStrut{"首页", "/"})

	c.HTML(200, "home/index.tmpl", gin.H{
		"title": "O-Ten",
		"nav":   nav,
	})
}
