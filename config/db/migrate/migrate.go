package migrate

import (
	"gorm.io/gorm"
	"o-ten/model"
)

// AutoMigrate 自动迁移
func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.Media{},
		&model.Source{},
		&model.Video{},
	)

	if err != nil {
		panic("Database auto migrate fails: " + err.Error())
	}

}
