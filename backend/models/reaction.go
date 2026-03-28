package models

import "time"

// MessageReaction is a single emoji reaction by a user on a chat or conversation message.
type MessageReaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	OwnerType string    `gorm:"not null;size:50;uniqueIndex:idx_react_unique" json:"owner_type"` // "chat_message" | "conv_message"
	OwnerID   uint      `gorm:"not null;uniqueIndex:idx_react_unique" json:"owner_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_react_unique" json:"user_id"`
	Emoji     string    `gorm:"not null;size:10;uniqueIndex:idx_react_unique" json:"emoji"`
}

// ReactionSummary groups reaction counts and user IDs for a single emoji.
type ReactionSummary struct {
	Emoji   string `json:"emoji"`
	Count   int    `json:"count"`
	UserIDs []uint `json:"user_ids"`
}
