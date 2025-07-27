package main

import (
	"context"
	"fmt"
	"github.com/novel/internal/model"
	"github.com/novel/internal/pkg/config"
	"github.com/novel/internal/pkg/db"
	"github.com/novel/internal/pkg/logger"
	"gorm.io/gorm"
	"log"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}
	err = logger.InitLogger(&cfg.Logger)
	if err != nil {
		log.Fatalf("无法初始化日志记录器: %v", err)
	}
	defer logger.Sync()
	logger.Info(context.Background(), "日志系统初始化成功")

	database, err := db.InitDB(&cfg.Database)
	if err != nil {
		panic(fmt.Errorf("fatal error database connection: %w", err))
	}

	if err := AddNovels(database); err != nil {
		log.Fatalf("failed to seed novels: %v", err)
	}

	fmt.Println("Database create completed successfully!")

}

func AddNovels(db *gorm.DB) error {
	// 定义我们要填充的小说数据
	novels := []model.Novel{
		{
			Title:         "三体",
			Author:        "刘慈欣",
			Description:   "文化大革命如火如荼进行的同时，军方绝密的“红岸工程”取得了突破性进展...",
			CoverImageURL: "https://example.com/covers/three_body.jpg",
		},
		{
			Title:         "流浪地球",
			Author:        "刘慈欣",
			Description:   "科学家们发现太阳将急速老化、膨胀，吞没整个太阳系。为求生存，人类打造出巨大的“行星发动机”...",
			CoverImageURL: "https://example.com/covers/wandering_earth.jpg",
		},
		{
			Title:         "黑暗森林",
			Author:        "刘慈欣",
			Description:   "在三体世界利用科技锁死了地球的基础科学之后，人类开始了“面壁计划”...",
			CoverImageURL: "https://example.com/covers/dark_forest.jpg",
		},
	}

	// 遍历并创建小说
	for _, novel := range novels {
		// 使用 FirstOrCreate 来防止重复插入
		// 它会根据给定的条件（这里是Title和Author）查找，如果找不到，就创建一条新记录
		result := db.Where("title = ? AND author = ?", novel.Title, novel.Author).FirstOrCreate(&novel)
		if result.Error != nil {
			return fmt.Errorf("could not create novel %s: %v", novel.Title, result.Error)
		}

		// 如果是新创建的小说 (RowsAffected > 0)，我们就为它添加一些评分
		if result.RowsAffected > 0 {
			fmt.Printf("Novel '%s' created. Seeding ratings...\n", novel.Title)
			ratings := []model.Rating{
				{NovelID: novel.ID, Score: 8},
				{NovelID: novel.ID, Score: 9},
				{NovelID: novel.ID, Score: 10},
			}
			if err := db.Create(&ratings).Error; err != nil {
				return fmt.Errorf("could not create ratings for novel %s: %v", novel.Title, err)
			}
		} else {
			fmt.Printf("Novel '%s' already exists. Skipping.\n", novel.Title)
		}
	}

	return nil
}
