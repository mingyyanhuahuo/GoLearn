package model

import (
	"time"
)

type Like struct {
	UserId    uint      `gorm:"primaryKey" json:"user_id"`
	PostId    uint      `gorm:"primaryKey" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}
