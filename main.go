package main

import (
	"github.com/spf13/viper"
	"log"
	"o-ten/config"
	"o-ten/config/initializer"
	"o-ten/spider"
)

func main() {
	initializer.Init()
	port := viper.GetString("SERVER_PORT")
	log.Println("port: ", port)

	r := config.Router()
	spider.GimyTvSpider()
	panic(r.Run(":" + port))
}
