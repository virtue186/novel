package dto

// PaginationQuery 定义了分页和排序的查询参数
type PaginationQuery struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
	SortBy   string `form:"sort_by,default=created_at"` // 默认按创建时间排序
	Order    string `form:"order,default=desc"`         // 默认降序
}

// PaginatedResponse 定义了标准的分页响应格式
type PaginatedResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}
