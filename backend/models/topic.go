package models

import (
	"time"

	"gorm.io/gorm"
)

type Topic struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	ProjectID  uint           `gorm:"not null;index" json:"project_id"`
	Project    Project        `json:"-"`
	UserID     uint           `gorm:"not null" json:"user_id"`
	User       User           `json:"user"`
	Title      string         `gorm:"not null;size:500" json:"title"`
	Body       string         `gorm:"type:text" json:"body"`
	IsPinned   bool           `gorm:"default:false" json:"is_pinned"`
	IsEdited   bool           `gorm:"default:false" json:"is_edited"`
	ReplyCount int            `gorm:"-" json:"reply_count"`
}

type TopicReply struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	TopicID   uint           `gorm:"not null;index" json:"topic_id"`
	Topic     Topic          `json:"-"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `json:"user"`
	Body      string         `gorm:"type:text;not null" json:"body"`
	IsEdited  bool           `gorm:"default:false" json:"is_edited"`
}
