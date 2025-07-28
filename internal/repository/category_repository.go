package repository

import (
	"github.com/novel/internal/model"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	FindOrCreate(name string) (*model.Category, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) FindOrCreate(name string) (*model.Category, error) {
	var category model.Category
	// FirstOrCreate 会查找，如果找不到，就根据给定的条件创建
	err := r.db.Where("name = ?", name).FirstOrCreate(&category).Error
	return &category, err
}
