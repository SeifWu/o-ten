package initializer

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"o-ten/config/db/migrate"
	"o-ten/global"
)

// InitDB 初始化数据库
func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true",
		viper.GetString("DB_USERNAME"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_HOST"),
		viper.GetString("DB_PORT"),
		viper.GetString("DB_DATABASE"),
		viper.GetString("DB_CHARSET"),
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Database connection fails: " + err.Error())
	}

	global.DB = db
	migrate.AutoMigrate(db)
}
