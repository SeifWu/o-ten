package migrate

import "gorm.io/gorm"

// AutoMigrate 自动迁移
func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		//&model.User{},
	)
}
