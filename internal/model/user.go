package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string   `gorm:"size:32;unique;not null"`
	PasswordHash string   `gorm:"size:255;not null"`
	TrustScore   float64  `gorm:"default:1.0"`
	Ratings      []Rating `json:"-"`
}
