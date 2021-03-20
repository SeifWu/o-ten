package model

import (
	"gorm.io/gorm"
	"time"
)

type Source struct {
	ID         int64          `json:"id" gorm:"primaryKey"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Title      string
	SourceName string
	SourceUrl  float64
	MediaId    int64
	Media      Media
}
