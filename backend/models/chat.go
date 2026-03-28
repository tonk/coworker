package models

import "time"

type ChatMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ProjectID uint      `gorm:"not null;index" json:"project_id"`
	Project   Project   `json:"-"`
	UserID    uint      `gorm:"default:0" json:"user_id"`
	User      User      `json:"user"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	IsEdited  bool      `gorm:"default:false" json:"is_edited"`
	IsDeleted bool      `gorm:"default:false" json:"is_deleted"`
	IsBot     bool      `gorm:"default:false" json:"is_bot"`
	BotName   string    `gorm:"size:100" json:"bot_name,omitempty"`
}
