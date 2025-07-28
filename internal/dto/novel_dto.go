package dto

// CreateRatingRequest 定义了创建评分API的请求体
type CreateRatingRequest struct {
	Score   int    `json:"score" binding:"required,min=1,max=10"`
	Comment string `json:"comment"` // 允许用户在评分时直接带上评论
}

// VoteForRatingRequest 定义了为评分投票API的请求体
type VoteForRatingRequest struct {
	// 校验确保了投票值只能是 1 (赞同) 或 -1 (反对)
	Vote int `json:"vote" binding:"required,oneof=-1 1"`
}
