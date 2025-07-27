package middleware

import (
	"github.com/google/uuid"
	"github.com/novel/internal/pkg/logger"
	"go.uber.org/zap"
	"net/http"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 为请求生成一个唯一的ID
		requestID := uuid.NewString()

		// 2. 将 ID 设置到响应头中，方便客户端追踪
		w.Header().Set("X-Request-ID", requestID)

		// 3. 使用 logger.InjectLoggerIntoContext 创建一个携带 request_id 的新 context
		ctx := logger.InjectLoggerIntoContext(r.Context(), zap.String("request_id", requestID))

		// 4. 创建一个带有新 context 的新请求对象
		reqWithCtx := r.WithContext(ctx)

		// 5. 调用后续的处理器
		next.ServeHTTP(w, reqWithCtx)
	})
}
