package migrate

import (
	"gorm.io/gorm"
	"o-ten/model"
)

// AutoMigrate 自动迁移
func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		&model.Media{},
		&model.Source{},
		&model.Video{},
	)
}
