package model

// Television 影视、电视表
type Television struct {
	Base
	Name         string `json:"name" gorm:"index; comment:名称; not null;"`
	Cover        string `json:"cover" gorm:"comment:封面"`
	Introduction string `json:"introduction" gorm:"type:text;comment:简介; "`
	TagId        int64  `json:"tagId" gorm:"comment:标签Id;"`
	CategoryId   int64  `json:"categoryId" gorm:"comment:分类Id;"`
	CategoryName string `json:"categoryName" gorm:"comment:分类名称;"`
	Region       string `json:"region" gorm:"comment:地区;"`
	Year         int    `json:"year" gorm:"comment:年份;"`
	//Toy   Toy `gorm:"polymorphic:Owner;"`
}
