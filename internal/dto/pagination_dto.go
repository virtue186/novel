package dto

// PaginatedResponse 定义了标准的分页响应格式
type PaginatedResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}

type ListQuery struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
	SortBy   string `form:"sort_by,default=created_at"`
	Order    string `form:"order,default=desc"`

	// --- 新增的筛选字段 ---
	PublicationType *int   `form:"publication_type"` // 使用指针以允许不传此参数
	CategoryID      *uint  `form:"category_id"`
	TagIDs          []uint `form:"tag_ids"` // 允许按多个标签筛选
}
