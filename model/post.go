package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	Id           uint           `gorm:"primaryKey" json:"id"`
	Title        string         `gorm:"size:100" json:"title"`
	Content      string         `gorm:"type:text" json:"content"`
	AuthorId     uint           `json:"author_id"`
	Author       User           `gorm:"foreignKey:AuthorId" json:"author"`
	Comments     []Comment      `gorm:"foreignKey:PostId" json:"comments,omitempty"`
	CommentCount uint           `gorm:"default:0" json:"comment_count"`
	LikeCount    uint           `gorm:"default:0" json:"like_count"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	CreatedAt    time.Time      `json:"created_at"`
}
