package model

import (
	"gorm.io/gorm"
	"time"
)

type Video struct {
	ID        int64          `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Title     string
	Src       string
	OrderSeq  float64
	SourceId  int64
	Source    Source
}
