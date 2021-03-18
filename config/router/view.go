package router

import (
	"github.com/gin-gonic/gin"
	"o-ten/controller/view"
)

// View 视图模版路由
func View(r *gin.Engine) {
	r.GET("/", view.HomePage)
}
