package model

import (
	"gorm.io/gorm"
	"time"
)

type Media struct {
	ID           int64          `json:"id" gorm:"primaryKey"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Title        string
	Introduction string
	Region       string
	year         int64
	Cover        string
}
