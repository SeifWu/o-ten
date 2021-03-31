package model

import (
	"gorm.io/gorm"
	"time"
)

type Source struct {
	ID          int64          `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Title       string         `json:"title" gorm:"comment:'标题'"`
	SourceName  string         `json:"sourceName" gorm:"comment:'来源名称'"`
	SourceUrl   string         `json:"sourceUrl" gorm:"comment:'来源网址'"`
	IsCompleted bool           `json:"is_completed" gorm:"comment:'是否完成'"`
	MediaId     int64
	Media       Media
	Video       []Video
}
