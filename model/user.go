package model

import (
	"time"
)

type User struct {
	Id          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"size:32" json:"username"`
	Name        string    `gorm:"size:32" json:"name"`
	PhoneNumber string    `gorm:"size:11 unique" json:"-"`
	PassHash    string    `json:"-"`
	Role        string    `gorm:"default:student" json:"role"`
	CreatedAt   time.Time `json:"-"`
}
