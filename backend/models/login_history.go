package models

import "time"

type LoginHistory struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index" json:"user_id"`        // subject (0 = unknown)
	Username      string    `gorm:"size:100" json:"username"`    // subject username
	ActorID       uint      `gorm:"index" json:"actor_id"`       // who performed the action (= UserID for self-actions)
	ActorUsername string    `gorm:"size:100" json:"actor_username"` // actor display name
	Event         string    `gorm:"size:64;not null" json:"event"`
	Detail        string    `gorm:"size:128" json:"detail"`
	IP            string    `gorm:"size:64" json:"ip"`
	Client        string    `gorm:"size:128" json:"client"`
	CreatedAt     time.Time `json:"created_at"`
}
