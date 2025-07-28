package model

import "gorm.io/gorm"

// VoteType 定义了投票类型的枚举
type VoteType int

const (
	VoteTypeUp   VoteType = 1  // 赞同
	VoteTypeDown VoteType = -1 // 反对
)

// RatingVote 记录了用户对评分的投票
type RatingVote struct {
	gorm.Model
	UserID   uint     `gorm:"uniqueIndex:idx_user_rating"` // 用户ID
	RatingID uint     `gorm:"uniqueIndex:idx_user_rating"` // 评分ID
	Vote     VoteType `gorm:"type:smallint"`               // 投票类型 (1 或 -1)
}

type Rating struct {
	gorm.Model
	NovelID        uint   `json:"novel_id"`
	UserID         uint   `json:"user_id"`
	Score          int    `json:"score"`
	Comment        string `json:"comment"`
	UpvotesCount   int    `json:"upvotes_count"`
	DownvotesCount int    `json:"downvotes_count"`

	// 内部计算字段
	Weight         float64 `json:"-"`
	UserTrustScore float64 `json:"-"`

	// 关联字段：一条评分属于一个用户
	User User `json:"-"`
}
