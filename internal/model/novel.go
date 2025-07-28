package model

import "gorm.io/gorm"

type Novel struct {
	gorm.Model
	Title         string   `json:"title"`
	Author        string   `json:"author"`
	Description   string   `json:"description"`
	CoverImageURL string   `json:"cover_image_url"`
	Ratings       []Rating `json:"ratings"`                     // 一本小说可以有多个评分
	WeightedScore float64  `json:"weighted_score" gorm:"index"` // 加权平均分，并添加索引以备排序
	RatingsCount  int      `json:"ratings_count"`               // 总评分数
}
