package global

import (
	"gorm.io/gorm"
)

var (
	// DB 数据库方便全局调用
	DB *gorm.DB
)
