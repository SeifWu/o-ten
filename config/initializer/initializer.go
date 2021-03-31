package initializer

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Init 初始化器
func Init() {
	InitEnv()
	InitDB()
	InitCron()
}

// InitEnv 初始化环境变量
func InitEnv() {
	root, _ := os.Getwd()
	viper.SetConfigFile(root + "/.env")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}
