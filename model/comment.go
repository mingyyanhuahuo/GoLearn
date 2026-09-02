package model

import (
	"time"
)

type Comment struct {
	Id        uint      `gorm:"primaryKey" json:"id"`
	PostId    uint      `json:"post_id"`
	Content   string    `gorm:"type:text" json:"content"`
	AuthorId  uint      `json:"-"`
	Author    User      `gorm:"foreignKey:AuthorId" json:"author"`
	CreatedAt time.Time `json:"created_at"`
}
