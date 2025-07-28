package repository

import (
	"fmt"
	"github.com/novel/internal/dto"
	"github.com/novel/internal/model"
	"gorm.io/gorm"
)

// NovelRepository 定义了与小说相关的数据库操作接口
type NovelRepository interface {
	FindAll(query *dto.ListQuery) ([]model.Novel, int64, error)
	FindByID(id uint) (*model.Novel, error)
	CreateRating(rating *model.Rating) error
	FindByIDWithRatings(id uint) (*model.Novel, error)
	Update(novel *model.Novel) error
	FindRatingByID(id uint) (*model.Rating, error)
	FindUserVote(userID, ratingID uint) (*model.RatingVote, error)
	UpdateRatingVote(rating *model.Rating, oldVote, newVote *model.RatingVote) error
	UpdateRating(rating *model.Rating) error
	CreateInTx(novel *model.Novel) error
}

// novelRepository 结构体实现了 NovelRepository 接口
type novelRepository struct {
	db *gorm.DB
}

// NewNovelRepository 是 novelRepository 的构造函数
func NewNovelRepository(db *gorm.DB) NovelRepository {
	return &novelRepository{db: db}
}

func (r *novelRepository) FindAll(query *dto.ListQuery) ([]model.Novel, int64, error) {
	var novels []model.Novel
	var total int64

	// 1. 构建基础查询，并预加载关联数据以备前端展示
	db := r.db.Model(&model.Novel{}).Preload("Category").Preload("Tags")

	// 2. 动态构建 WHERE 子句
	if query.PublicationType != nil {
		db = db.Where("publication_type = ?", *query.PublicationType)
	}
	if query.CategoryID != nil {
		db = db.Where("category_id = ?", *query.CategoryID)
	}
	if len(query.TagIDs) > 0 {
		// GORM 的多对多查询，需要使用 JOIN
		// 这会查找那些 novel_tags 中 novel_id 匹配，且 tag_id 在我们指定列表中的小说
		// HAVING COUNT(*) = ? 确保小说拥有所有指定的标签
		db = db.Joins("JOIN novel_tags ON novels.id = novel_tags.novel_id").
			Where("novel_tags.tag_id IN ?", query.TagIDs).
			Group("novels.id").
			Having("COUNT(novels.id) = ?", len(query.TagIDs))
	}

	// 3. 计算总数 (在应用分页之前)
	if err := db.Count(&total).Error; err != nil {
		// 注意：对于复杂的JOIN查询，Count可能会变慢或不准，需要特别注意
		// 暂时我们先这样实现，后续可优化
		return nil, 0, err
	}

	// 4. 应用排序和分页
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

// FindRatingByID 根据ID查找评分
func (r *novelRepository) FindRatingByID(id uint) (*model.Rating, error) {
	var rating model.Rating
	err := r.db.First(&rating, id).Error
	return &rating, err
}

// FindUserVote 查找特定用户对特定评分的投票记录
func (r *novelRepository) FindUserVote(userID, ratingID uint) (*model.RatingVote, error) {
	var vote model.RatingVote
	err := r.db.Where("user_id = ? AND rating_id = ?", userID, ratingID).First(&vote).Error
	return &vote, err
}

// UpdateRatingVote 使用数据库事务来更新投票
func (r *novelRepository) UpdateRatingVote(rating *model.Rating, oldVote, newVote *model.RatingVote) error {
	// 使用事务来保证数据一致性：对 vote 表的修改和对 rating 表计数的修改，必须同时成功或失败
	return r.db.Transaction(func(tx *gorm.DB) error {
		// --- 处理 vote 表 ---
		if oldVote != nil { // 如果存在旧投票，先删除
			if err := tx.Delete(oldVote).Error; err != nil {
				return err
			}
		}
		if newVote != nil { // 如果存在新投票，创建它
			if err := tx.Create(newVote).Error; err != nil {
				return err
			}
		}

		// --- 更新 rating 表的计数 ---
		if err := tx.Save(rating).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *novelRepository) UpdateRating(rating *model.Rating) error {
	// GORM 的 Save 会根据主键(ID)来执行更新，更新所有字段
	return r.db.Save(rating).Error
}

func (r *novelRepository) CreateInTx(novel *model.Novel) error {
	// GORM 的 Transaction 方法会自动处理开始事务、提交或回滚
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 在这个事务中，所有的数据库操作都使用 tx 对象

		// 1. 创建 Novel 主记录
		if err := tx.Create(novel).Error; err != nil {
			return err
		}

		// 2. GORM 会自动处理 `novel.Tags` 的多对多关联关系
		//    因为它在模型中已经通过 `gorm:"many2many:novel_tags;"` 定义好了
		//    在创建 novel 时，关联的 tags 会被自动插入到 novel_tags 中间表

		return nil
	})
}
