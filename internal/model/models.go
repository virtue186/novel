package model

import "gorm.io/gorm"

type Novel struct {
	gorm.Model
	Title         string   `json:"title"`
	Author        string   `json:"author"`
	Description   string   `json:"description"`
	CoverImageURL string   `json:"cover_image_url"`
	Ratings       []Rating `json:"ratings"` // 一本小说可以有多个评分
}

type Rating struct {
	gorm.Model
	NovelID uint `json:"novel_id"`                             // 外键，关联到 Novel
	Score   int  `json:"score" binding:"required,min=1,max=5"` // 评分，1-5分，binding 用于参数校验
}
