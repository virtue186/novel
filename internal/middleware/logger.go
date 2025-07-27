package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/novel/internal/pkg/logger"
	"go.uber.org/zap"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()
		c.Writer.Header().Set("X-Request-ID", requestID)

		// 注入到 context.Context 中，供 logger 使用
		ctx := logger.InjectLoggerIntoContext(c.Request.Context(), zap.String("request_id", requestID))
		c.Request = c.Request.WithContext(ctx)

		// 继续处理请求
		c.Next()
	}
}
