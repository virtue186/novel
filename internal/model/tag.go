package model

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	Name   string   `gorm:"size:50;unique;not null"`
	Novels []*Novel `gorm:"many2many:novel_tags;" json:"-"` // GORM 会自动处理中间表
}
