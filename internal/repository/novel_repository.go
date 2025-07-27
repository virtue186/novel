package repository

import (
	"fmt"
	"github.com/novel/internal/dto"
	"github.com/novel/internal/model"
	"gorm.io/gorm"
)

// NovelRepository 定义了与小说相关的数据库操作接口
type NovelRepository interface {
	FindAll(query *dto.PaginationQuery) ([]model.Novel, int64, error)
	FindByID(id uint) (*model.Novel, error)
	CreateRating(rating *model.Rating) error
	FindByIDWithRatings(id uint) (*model.Novel, error)
	Update(novel *model.Novel) error
}

// novelRepository 结构体实现了 NovelRepository 接口
type novelRepository struct {
	db *gorm.DB
}

// NewNovelRepository 是 novelRepository 的构造函数
func NewNovelRepository(db *gorm.DB) NovelRepository {
	return &novelRepository{db: db}
}

// FindAll 实现获取所有小说的方法
func (r *novelRepository) FindAll(query *dto.PaginationQuery) ([]model.Novel, int64, error) {
	var novels []model.Novel
	var total int64

	// 1. 构建基础查询
	db := r.db.Model(&model.Novel{})

	// 2. 首先，在不加分页条件的情况下，计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 3. 应用排序和分页
	// 构造 ORDER BY 语句，防止 SQL 注入
	orderClause := fmt.Sprintf("%s %s", query.SortBy, query.Order)
	offset := (query.Page - 1) * query.PageSize

	err := db.Order(orderClause).Limit(query.PageSize).Offset(offset).Find(&novels).Error
	if err != nil {
		return nil, 0, err
	}

	return novels, total, nil
}

// FindByID 实现根据ID查找小说的方法
func (r *novelRepository) FindByID(id uint) (*model.Novel, error) {
	var novel model.Novel
	err := r.db.First(&novel, id).Error
	// 注意：这里我们直接返回 gorm 的错误，让上层去判断是不是 ErrRecordNotFound
	return &novel, err
}

// CreateRating 实现创建评分的方法
func (r *novelRepository) CreateRating(rating *model.Rating) error {
	return r.db.Create(rating).Error
}

// FindByIDWithRatings 实现根据ID查找小说，并预加载其所有评分
func (r *novelRepository) FindByIDWithRatings(id uint) (*model.Novel, error) {
	var novel model.Novel
	// 使用 GORM 的 Preload 功能来避免 N+1 查询问题
	err := r.db.Preload("Ratings").First(&novel, id).Error
	return &novel, err
}

// Update 实现更新小说记录的方法
func (r *novelRepository) Update(novel *model.Novel) error {
	// GORM 的 Save 会更新所有字段，即使是零值
	return r.db.Save(novel).Error
}
