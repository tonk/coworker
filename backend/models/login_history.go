package models

import "time"

type LoginHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	IP        string    `gorm:"size:64" json:"ip"`
	Client    string    `gorm:"size:128" json:"client"`
	CreatedAt time.Time `json:"created_at"`
}
