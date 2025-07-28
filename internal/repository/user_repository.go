package repository

import (
	"github.com/novel/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByID(userID uint) (*model.User, error)
	Update(user *model.User) error
	FindByIDWithRatings(userID uint) (*model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByID(userID uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, userID).Error
	return &user, err
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) FindByIDWithRatings(userID uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Ratings").First(&user, userID).Error
	return &user, err
}
