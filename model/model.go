package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	Id          uint      `gorm:"primaryKey"json:"id"`
	Name        string    `gorm:"size:15"json:"name"`
	PhoneNumber string    `gorm:"size:11 unique"json:"phone_number"`
	PassHash    string    `json:"-"`
	Role        string    `gorm:"default:user"json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}
type Post struct {
	Id           uint           `gorm:"primaryKey"json:"id"`
	Title        string         `gorm:"size:100"json:"title"`
	Content      string         `gorm:"type:text"json:"content"`
	AuthorId     uint           `json:"author_id"`
	Author       User           `gorm:"foreignKey:AuthorId"json:"author"`
	CommentCount uint           `gorm:"default:0"json:"comment_count"`
	LikeCount    uint           `gorm:"default:0"json:"like_count"`
	DeletedAt    gorm.DeletedAt `gorm:"index"json:"deleted_at"`
	CreatedAt    time.Time      `json:"created_at"`
}
type Comment struct {
	Id        uint           `gorm:"primaryKey"json:"id"`
	Content   string         `gorm:"type:text"json:"content"`
	AuthorId  uint           `json:"author_id"`
	PostId    uint           `json:"post_id"`
	Author    User           `gorm:"foreignKey:AuthorId"json:"author"`
	DeletedAt gorm.DeletedAt `gorm:"index"json:"deleted_at"`
	CreatedAt time.Time      `json:"created_at"`
}
type Like struct {
	UserId    uint      `gorm:"primaryKey"json:"user_id"`
	PostId    uint      `gorm:"primaryKey"json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}
