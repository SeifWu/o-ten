package config

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"o-ten/config/router"
)

// Router 路由配置
func Router() *gin.Engine {
	r := gin.Default()

	r.Use(static.Serve("/assets", static.LocalFile("./assets", true)))

	v1 := r.Group("/api/v1")
	{
		//v1manager := v1.Group("/manager")
		//v1manager.Use(middleware.JWTAuthMiddleware())
		//router.V1Manager(v1manager)

		v1public := v1.Group("/public")
		router.V1Public(v1public)
	}

	return r
}
