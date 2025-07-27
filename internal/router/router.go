package router

import (
	"github.com/gin-gonic/gin"
	"github.com/novel/internal/handler"
	"github.com/novel/internal/middleware"
	"github.com/novel/internal/pkg/config"
	"github.com/novel/internal/repository"
	"github.com/novel/internal/service"
	"gorm.io/gorm"
)

// SetupRouter 设置并返回一个配置好的 Gin 引擎
func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	router := gin.New() // 使用 gin.New() 而不是 gin.Default()，以便完全自定义中间件

	// 使用我们自定义的 zap 日志中间件
	router.Use(middleware.RequestIDMiddleware())
	// 使用 Gin 自带的 Recovery 中间件，防止 panic 导致整个程序崩溃
	router.Use(gin.Recovery())

	novelRepo := repository.NewNovelRepository(db)
	novelSvc := service.NewNovelService(novelRepo, &cfg.Algorithm)
	novelHandler := handler.NewNovelHandler(novelSvc)

	// 设置 API v1 路由组
	apiV1 := router.Group("/api/v1")
	{
		novels := apiV1.Group("/novels")
		{
			novels.GET("", novelHandler.GetNovels)
			novels.POST("/:id/rate", novelHandler.CreateRating)
		}
	}

	return router
}
