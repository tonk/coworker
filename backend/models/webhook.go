package models

import "time"

// ProjectWebhook is an incoming webhook URL for posting messages to a project's chat.
// Type is either "generic" (plain JSON body) or "gitea" (Gitea/Forgejo event payloads).
type ProjectWebhook struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	ProjectID   uint      `gorm:"not null;index" json:"project_id"`
	Name        string    `gorm:"size:100" json:"name"`
	Token       string    `gorm:"uniqueIndex;not null;size:64" json:"-"`
	TokenHint   string    `gorm:"size:8" json:"token_hint"`
	Type        string    `gorm:"size:20;not null;default:'generic'" json:"type"`
	CreatedByID uint      `gorm:"not null" json:"created_by_id"`
}
