package main

import (
	"github.com/spf13/viper"
	"log"
	"o-ten/config"
	"o-ten/config/initializer"
)

func main() {
	initializer.Init()
	port := viper.GetString("SERVER_PORT")
	log.Println("port: ", port)

	r := config.Router()

	panic(r.Run(":" + port))
}
