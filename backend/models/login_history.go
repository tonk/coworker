package models

import "time"

type LoginHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"` // 0 = unknown/unauthenticated user
	Username  string    `gorm:"size:100" json:"username"`
	Event     string    `gorm:"size:64;not null" json:"event"`
	Detail    string    `gorm:"size:128" json:"detail"`
	IP        string    `gorm:"size:64" json:"ip"`
	Client    string    `gorm:"size:128" json:"client"`
	CreatedAt time.Time `json:"created_at"`
}
