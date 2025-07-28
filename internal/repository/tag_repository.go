package repository

import (
	"github.com/novel/internal/model"
	"gorm.io/gorm"
)

type TagRepository interface {
	FindOrCreateByNames(names []string) ([]*model.Tag, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) FindOrCreateByNames(names []string) ([]*model.Tag, error) {
	var tags []*model.Tag
	for _, name := range names {
		var tag model.Tag
		err := r.db.Where("name = ?", name).FirstOrCreate(&tag).Error
		if err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}
