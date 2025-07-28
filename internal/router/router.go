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

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(gin.Recovery())

	// --- 实例化所有依赖 ---
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, &cfg.JWT)
	userHandler := handler.NewUserHandler(userSvc)

	novelRepo := repository.NewNovelRepository(db)
	trustSvc := service.NewTrustService(userRepo, novelRepo)
	novelSvc := service.NewNovelService(novelRepo, trustSvc, &cfg.Algorithm)
	novelHandler := handler.NewNovelHandler(novelSvc)

	// --- 路由设置 ---
	apiV1 := router.Group("/api/v1")
	{
		// 开放路由：注册和登录
		apiV1.POST("/register", userHandler.Register)
		apiV1.POST("/login", userHandler.Login)

		// 小说相关公共路由
		novels := apiV1.Group("/novels")
		{
			novels.GET("", novelHandler.GetNovels)
			novels.GET("/:id", novelHandler.GetNovelByID)
		}

		// 需要认证的路由组
		authRequired := apiV1.Group("").Use(middleware.AuthMiddleware(userSvc))
		{
			authRequired.POST("/novels/:id/rate", novelHandler.CreateRating)
			authRequired.POST("/ratings/:id/vote", novelHandler.VoteForRating)
		}
	}
	return router
}
