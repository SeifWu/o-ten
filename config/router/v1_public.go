package router

import (
	"github.com/gin-gonic/gin"
)

// V1Public 管理端接口 版本: v1
func V1Public(r *gin.RouterGroup) {
	r.GET("/index", func(context *gin.Context) {
		context.JSON(
			200,
			gin.H{
				"success": true,
				"message": "hello world",
			},
		)
	})
}
