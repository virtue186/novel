package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/novel/internal/dto"
	"github.com/novel/internal/middleware"
	"github.com/novel/internal/model"
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

// GetNovels 获取小说列表 (已为排序和分页做好准备)
func (h *NovelHandler) GetNovels(c *gin.Context) {
	var query dto.ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数错误")
		return
	}
	paginatedResult, err := h.svc.GetRankedNovels(&query)
	if err != nil {
		response.ServerError(c)
		return
	}

	response.Ok(c, paginatedResult)
}

// GetNovelByID 获取单本小说的详情 (已为高性能预计算做好准备)
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
	response.Ok(c, details)
}

// CreateRating 为小说创建新评分 (已完善)
func (h *NovelHandler) CreateRating(c *gin.Context) {
	// 1. 解析小说ID，确保类型正确
	novelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的小说ID")
		return
	}

	// 2. 从 Context 中获取当前登录用户的ID
	userIDVal, exists := c.Get(middleware.CtxUserIDKey)
	if !exists {
		response.Fail(c, "无法获取用户信息，请重新登录")
		return
	}
	// 进行类型断言，确保 userID 是 uint 类型
	userID, ok := userIDVal.(uint)
	if !ok {
		response.ServerError(c) // 如果类型不匹配，说明中间件或代码逻辑有问题
		return
	}

	// 3. 绑定请求体到专用的DTO
	var req dto.CreateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不合法")
		return
	}

	// 4. 调用 Service 处理业务，现在参数完全匹配
	rating, err := h.svc.CreateRatingForNovel(userID, uint(novelID), req.Score, req.Comment)
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

// VoteForRating 为评分投票 (已完善)
func (h *NovelHandler) VoteForRating(c *gin.Context) {
	// 1. 解析评分ID
	ratingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的评分ID")
		return
	}

	// 2. 从 Context 中获取当前登录用户的ID
	userID, exists := c.Get(middleware.CtxUserIDKey)
	if !exists {
		response.Fail(c, "无法获取用户信息，请重新登录")
		return
	}

	// 3. 绑定请求体到专用的DTO
	var req dto.VoteForRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不合法")
		return
	}

	// 4. 调用 Service 处理业务
	err = h.svc.VoteForRating(userID.(uint), uint(ratingID), model.VoteType(req.Vote))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c)
		} else {
			response.ServerError(c)
		}
		return
	}

	response.OkWithMessage(c, "投票成功", nil)
}

func (h *NovelHandler) CreateNovel(c *gin.Context) {
	// 1. 绑定并校验请求体到我们专用的 DTO
	var req dto.CreateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 返回一个对用户友好的、通用的参数错误信息
		response.BadRequest(c, "请求参数不合法或缺少必要字段")
		return
	}

	// 2. 调用 Service 层，将创建的复杂业务逻辑委托出去
	novel, err := h.svc.CreateNovel(&req)
	if err != nil {
		// 在这里，我们可以根据 Service 返回的不同错误类型，给出更具体的提示
		// 但为了保持简洁，我们暂时统一返回服务器内部错误
		response.ServerError(c)
		return
	}

	// 3. 成功创建后，返回一个 201 Created 状态码和新创建的小说数据
	c.JSON(201, response.Response{
		Code: 0,
		Msg:  "小说创建成功",
		Data: novel,
	})
}
