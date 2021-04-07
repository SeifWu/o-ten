package model

type Video struct {
	Base
	Title     string `json:"name" gorm:"comment:视频名称"`
	Url       string `json:"url" gorm:"comment:视频地址"`
	OrderSeq  string `json:"orderSeq" gorm:"comment:排序"`
	SourceID  string `json:"sourceId" gorm:"comment:来源ID"`
	SourceUrl string `json:"sourceUrl" gorm:"comment:来源链接"`
	Source    Source
}
