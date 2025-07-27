package main

import (
	"context"
	"fmt"
	"github.com/novel/internal/pkg/config"
	"github.com/novel/internal/pkg/db"
	"github.com/novel/internal/pkg/logger"
	"github.com/novel/internal/router"
	"log"
)

func main() {
	// 初始化项目配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}
	// 初始化日志系统
	err = logger.InitLogger(&cfg.Logger)
	if err != nil {
		log.Fatalf("无法初始化日志记录器: %v", err)
	}
	defer logger.Sync()
	logger.Info(context.Background(), "日志系统初始化成功")

	// 初始化数据库
	database, err := db.InitDB(&cfg.Database)
	if err != nil {
		panic(fmt.Errorf("fatal error database connection: %w", err))
	}

	r := router.SetupRouter(database, cfg)
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Infof("Server is running on port %s", addr)

	if err := r.Run(addr); err != nil {
		logger.Fatalf("Server run failed: %v", err)
	}

}
