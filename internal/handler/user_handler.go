package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/novel/internal/dto"
	"github.com/novel/internal/pkg/response"
	"github.com/novel/internal/service"
)

// UserHandler 结构体
type UserHandler struct {
	svc service.UserService
}

// NewUserHandler 构造函数
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register 处理用户注册请求
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// [TODO] 在这里添加更复杂的密码强度校验逻辑

	user, err := h.svc.Register(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			response.Fail(c, "用户名已存在")
		} else {
			response.ServerError(c)
		}
		return
	}
	response.Ok(c, user)
}

// Login 处理用户登录请求
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	token, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Fail(c, "用户名或密码错误")
		} else {
			response.ServerError(c)
		}
		return
	}

	response.Ok(c, gin.H{"token": token})
}
