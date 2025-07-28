package model

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name   string  `gorm:"size:50;unique;not null"`
	Novels []Novel `json:"-"` // 一个分类下有多本小说
}
