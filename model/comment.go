package model

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	Id        uint           `gorm:"primaryKey" json:"id"`
	Content   string         `gorm:"type:text" json:"content"`
	AuthorId  uint           `json:"author_id"`
	PostId    uint           `json:"post_id"`
	Author    User           `gorm:"foreignKey:AuthorId" json:"author"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	CreatedAt time.Time      `json:"created_at"`
}
