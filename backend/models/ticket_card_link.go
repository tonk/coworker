package models

import "time"

type TicketCardLink struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TicketID    uint      `gorm:"not null;index;uniqueIndex:idx_ticket_card_pair" json:"ticket_id"`
	CardID      uint      `gorm:"not null;index;uniqueIndex:idx_ticket_card_pair" json:"card_id"`
	CreatedByID uint      `json:"created_by_id"`
	CreatedAt   time.Time `json:"created_at"`
}
