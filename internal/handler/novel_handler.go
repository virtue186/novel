package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/novel/internal/dto"
	"github.com/novel/internal/pkg/response"
	"github.com/novel/internal/service"
	"gorm.io/gorm"
	"strconv"
)

// NovelHandler 结构体，持有 Service 接口
type NovelHandler struct {
	svc service.NovelService
}

// NewNovelHandler 的构造函数，接收 NovelService 接口作为参数
func NewNovelHandler(svc service.NovelService) *NovelHandler {
	return &NovelHandler{svc: svc}
}

// GetNovels 获取小说列表 (我们将在下一步增强此方法)
func (h *NovelHandler) GetNovels(c *gin.Context) {
	// 1. 将 URL query 参数绑定到我们的 DTO 结构体上
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 2. 调用 Service 获取分页结果
	paginatedResult, err := h.svc.GetRankedNovels(&query)
	if err != nil {
		response.ServerError(c)
		return
	}

	// 3. 返回标准化的分页响应
	response.Ok(c, paginatedResult)
}

// GetNovelByID 获取单本小说的详情
func (h *NovelHandler) GetNovelByID(c *gin.Context) {
	novelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的小说ID")
		return
	}
	details, err := h.svc.GetNovelWithCalculatedScores(uint(novelID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c)
		} else {
			response.ServerError(c)
		}
		return
	}

	// 2. 直接将 Service 返回的 DTO 作为响应数据
	//    Handler 不再需要关心数据是如何组装的
	response.Ok(c, details)
}

// CreateRating 为小说创建新评分 (保持不变，已经很完美)
func (h *NovelHandler) CreateRating(c *gin.Context) {
	novelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的小说ID")
		return
	}

	var input struct {
		Score int `json:"score" binding:"required,min=1,max=10"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	rating, err := h.svc.CreateRatingForNovel(uint(novelID), input.Score)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c)
		} else {
			response.ServerError(c)
		}
		return
	}

	response.OkWithMessage(c, "评分成功", rating)
}
