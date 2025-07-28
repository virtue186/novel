package model

import "gorm.io/gorm"

// PublicationType 定义了出版类型的枚举
type PublicationType int

const (
	TypeWebNovel  PublicationType = 1 // 网络小说
	TypePublished PublicationType = 2 // 实体书籍
)

// SerializationStatus 定义了连载状态的枚举
type SerializationStatus int

const (
	StatusSerializing SerializationStatus = 1 // 连载中
	StatusCompleted   SerializationStatus = 2 // 已完结
)

type Novel struct {
	gorm.Model
	Title         string `json:"title"`
	Author        string `json:"author"`
	Description   string `json:"description"`
	CoverImageURL string `json:"cover_image_url"`

	Ratings       []Rating `json:"ratings"`                     // 一本小说可以有多个评分
	WeightedScore float64  `json:"weighted_score" gorm:"index"` // 加权平均分，并添加索引以备排序
	RatingsCount  int      `json:"ratings_count"`               // 总评分数

	// --- 新增的核心区分字段 ---
	PublicationType PublicationType `gorm:"not null;index"`

	// --- 实体书特有字段 ---
	Publisher *string `gorm:"size:100"` // 出版社 (指针，可为空)
	Isbn      *string `gorm:"size:20"`  // ISBN (指针，可为空)

	// --- 网络小说特有字段 ---
	WordCount           int                 `json:"word_count"`
	PublicationSite     string              `json:"publication_site"`
	SerializationStatus SerializationStatus `json:"serialization_status"`

	// --- 统一的分类与标签 ---
	CategoryID uint     `json:"category_id"`
	Category   Category `json:"category"`
	Tags       []*Tag   `gorm:"many2many:novel_tags;" json:"tags"`
}
