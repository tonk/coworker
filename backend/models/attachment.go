package models

import "time"

// Attachment represents an uploaded file linked to a chat message, conversation message, or card comment.
type Attachment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	OwnerType  string    `gorm:"not null;size:50;index:idx_attach_owner" json:"owner_type"` // "chat_message" | "conv_message" | "card_comment"
	OwnerID    uint      `gorm:"not null;index:idx_attach_owner" json:"owner_id"`
	UploaderID uint      `gorm:"not null" json:"uploader_id"`
	Filename   string    `gorm:"not null;size:255" json:"filename"`
	StoredName string    `gorm:"not null;size:255" json:"-"`
	MimeType   string    `gorm:"size:100" json:"mime_type"`
	SizeBytes  int64     `json:"size_bytes"`
}
