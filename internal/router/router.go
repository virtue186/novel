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

// SetupRouter 设置并返回一个配置好的 Gin 引擎 (最终版)
func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(gin.Recovery())

	// --- 依赖注入的“总装配流水线” ---
	userRepo := repository.NewUserRepository(db)
	novelRepo := repository.NewNovelRepository(db)
	categoryRepo := repository.NewCategoryRepository(db) // <-- 确保已创建
	tagRepo := repository.NewTagRepository(db)           // <-- 确保已创建

	trustSvc := service.NewTrustService(userRepo, novelRepo)
	// 确保 NewNovelService 的签名和调用是最新版本，能接收所有需要的 repo
	novelSvc := service.NewNovelService(novelRepo, trustSvc, categoryRepo, tagRepo, &cfg.Algorithm)
	userSvc := service.NewUserService(userRepo, &cfg.JWT)

	novelHandler := handler.NewNovelHandler(novelSvc)
	userHandler := handler.NewUserHandler(userSvc)

	// --- 路由设置 ---
	apiV1 := router.Group("/api/v1")
	{
		// 开放路由
		apiV1.POST("/register", userHandler.Register)
		apiV1.POST("/login", userHandler.Login)

		novelsPublic := apiV1.Group("/novels")
		{
			novelsPublic.GET("", novelHandler.GetNovels)
			novelsPublic.GET("/:id", novelHandler.GetNovelByID)
		}

		authRequired := apiV1.Group("")                      // 首先，创建路由组，authRequired 的类型是 *gin.RouterGroup
		authRequired.Use(middleware.AuthMiddleware(userSvc)) // 然后，对这个路由组应用中间件
		{
			// 小说相关的写操作，通常需要认证，甚至需要管理员权限
			novelsProtected := authRequired.Group("/novels")
			{
				novelsProtected.POST("", novelHandler.CreateNovel) // <-- 注册新路由
				novelsProtected.POST("/:id/rate", novelHandler.CreateRating)
			}

			ratingsProtected := authRequired.Group("/ratings")
			{
				ratingsProtected.POST("/:id/vote", novelHandler.VoteForRating)
			}
		}
	}
	return router
}
