package router

import (
	"github.com/gin-gonic/gin"
	"o-ten/controller/api/v1/public"
)

// V1Public 管理端接口 版本: v1
func V1Public(r *gin.RouterGroup) {
	r.GET("/index", public.IndexController)
}
