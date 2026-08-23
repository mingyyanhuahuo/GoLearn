package model

import (
	"time"
)

type User struct {
	Id          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"size:15" json:"username"`
	PhoneNumber string    `gorm:"size:11 unique" json:"phone_number"`
	PassHash    string    `json:"-"`
	Role        string    `gorm:"default:user" json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}
