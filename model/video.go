package model

type Video struct {
	Base
	Title     string `json:"name" gorm:"comment:视频名称"`
	Url       string `json:"url" gorm:"comment:视频地址"`
	OrderSeq  float64 `json:"orderSeq" gorm:"comment:排序"`
	SourceID  int64 `json:"sourceId" gorm:"comment:来源ID"`
	SourceUrl string `json:"sourceUrl" gorm:"comment:来源链接,播放页地址"`
	Source    Source
}
