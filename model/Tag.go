package model

type Tag struct {
	Base
	Name string `json:"name" gorm:"index; comment:名称"`
	Heat float64 `json:"heat" gorm:"comment:热度"`
}
