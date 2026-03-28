package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         string         `gorm:"not null;size:200" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	Slug         string         `gorm:"uniqueIndex;not null;size:100" json:"slug"`
	Color        string         `gorm:"size:7" json:"color"`
	IsArchived   bool           `gorm:"default:false" json:"is_archived"`
	KeyPrefix    string         `gorm:"size:3;not null;default:''" json:"key_prefix"`
	CardCounter  int            `gorm:"default:0" json:"-"`
	CreatedByID  uint           `gorm:"not null" json:"created_by_id"`
	CreatedBy    User           `json:"created_by"`
	Members      []ProjectMember `json:"members,omitempty"`
	Columns      []Column        `json:"columns,omitempty"`
	Labels       []Label         `json:"labels,omitempty"`
	ChatMessages []ChatMessage   `json:"chat_messages,omitempty"`
}

type ProjectMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ProjectID uint      `gorm:"not null;uniqueIndex:idx_proj_user" json:"project_id"`
	Project   Project   `json:"-"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_proj_user" json:"user_id"`
	User      User      `json:"user"`
	Role      string    `gorm:"not null;default:'member'" json:"role"` // "owner" | "member" | "viewer"
	InvitedBy uint      `json:"invited_by"`
}
