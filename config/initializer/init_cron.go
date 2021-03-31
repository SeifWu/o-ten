package initializer

import (
	"github.com/robfig/cron/v3"
	"log"
)

func InitCron()  {
	c := cron.New()
	c.AddFunc("@daily", func() {
		log.Println("test cron")
	})

	c.Start()
}
