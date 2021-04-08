package model

import "time"

type Source struct {
	Base
	Name          string    `json:"name" gorm:"comment:来源名称"`
	OwnerID       int64     `json:"ownerId" gorm:"comment:拥有者ID"`
	OwnerType     string    `json:"ownerType" gorm:"comment:拥有者表名"`
	Domain        string    `json:"domain" gorm:"comment:采集网站"`
	SourceUrl     string    `json:"sourceUrl" gorm:"comment:来源链接"`
	UpdatedFlag   string    `json:"updatedFlag" gorm:"comment:来源是否更新标识"`
	IsCompleted   bool      `json:"isCompleted" gorm:"comment:'是否完成'"`
	LastCollectAt time.Time `json:"lastCollectAt" gorm:"comment:'最后采集时间'"`
}
