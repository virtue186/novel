package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/novel/internal/pkg/response"
	"github.com/novel/internal/service"
	"strings"
)

const CtxUserIDKey = "userID"

// AuthMiddleware 创建一个认证中间件
func AuthMiddleware(userSvc service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.FailWithCode(c, 401, "请求未携带token，无权限访问")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.FailWithCode(c, 401, "请求头中auth格式有误")
			c.Abort()
			return
		}

		userID, err := userSvc.ParseToken(parts[1])
		if err != nil {
			if errors.Is(err, service.ErrInvalidToken) {
				response.FailWithCode(c, 401, "无效的token")
			} else {
				response.ServerError(c)
			}
			c.Abort()
			return
		}

		// 将解析出的 userID 存入 gin.Context，供后续的 Handler 使用
		c.Set(CtxUserIDKey, userID)
		c.Next()
	}
}
