package config

import (
	"github.com/gin-gonic/gin"
	"o-ten/config/router"
)

// Router 路由配置
func Router() *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("frontend/www/**/*")
	r.Static("/static", "../static/")
	r.StaticFile("/favicon.ico", "../static/favicon.ico")
	router.View(r)

	return r
}
