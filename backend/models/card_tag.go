package models

import "time"

type CardTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	CardID    uint      `gorm:"not null;uniqueIndex:idx_card_tag" json:"card_id"`
	Name      string    `gorm:"not null;size:100;uniqueIndex:idx_card_tag" json:"name"`
}
