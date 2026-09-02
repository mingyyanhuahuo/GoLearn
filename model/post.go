package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	Id           uint           `gorm:"primaryKey" json:"id"`
	Title        string         `gorm:"size:150" json:"title"`
	Content      string         `gorm:"type:text" json:"content"`
	Author       User           `gorm:"foreignKey:AuthorId" json:"author"`
	AuthorId     uint           `json:"author_id"`
	Comments     []Comment      `gorm:"foreignKey:PostId" json:"comments,omitempty"`
	CommentCount uint           `gorm:"default:0" json:"comment_count"`
	LikeCount    uint           `gorm:"default:0" json:"like_count"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
}
