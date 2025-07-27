package main

import (
	"context"
	"github.com/novel/internal/pkg/config"
	"github.com/novel/internal/pkg/logger"
	"log"
)

func main() {
	// 初始化项目配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}
	// 初始化日志系统
	err = logger.InitLogger(cfg)
	if err != nil {
		log.Fatalf("无法初始化日志记录器: %v", err)
	}
	defer logger.Sync()

	logger.Info(context.Background(), "日志系统初始化成功")
}
