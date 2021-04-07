package model

type Category struct {
	Base
	Name     string `json:"name" gorm:"index; comment:名称"`
	ParentID int64  `json:"parentId" gorm:"comment:父ID"`
}
