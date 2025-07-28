package db

import (
	"fmt"
	"github.com/novel/internal/model"
	"github.com/novel/internal/pkg/config"
	"github.com/novel/internal/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
		cfg.SSLMode,
	)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.InfoRaw("Database connection initialized")

	err = db.AutoMigrate(
		&model.Novel{},
		&model.Rating{},
		&model.RatingVote{},
		&model.User{},
		&model.Category{},
		&model.Tag{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate projects: %w", err)
	}
	logger.InfoRaw("Database migration complete")
	return db, nil
}

func GetDB() *gorm.DB {
	if db == nil {
		panic("Database is not initialized")
	}
	return db
}
