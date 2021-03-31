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
	Title        string         `json:"title" gorm:"comment:'标题'"`
	Introduction string         `json:"introduction" gorm:"comment:'简介'"`
	Region       string         `json:"region" gorm:"comment:'地区'"`
	Year         int            `json:"year" gorm:"comment:'年份'"`
	Cover        string         `json:"cover" gorm:"comment:'封面'"`
	Type         string         `json:"type" gorm:"comment:'类型'"`
	CategoryName string			`json:"categoryName"`
	Source       []Source
}
