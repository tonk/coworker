package models

import "time"

type TicketTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	TicketID  uint      `gorm:"not null;uniqueIndex:idx_ticket_tag" json:"ticket_id"`
	Name      string    `gorm:"not null;size:100;uniqueIndex:idx_ticket_tag" json:"name"`
}

type TicketLink struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	SourceTicketID  uint      `gorm:"not null;index;uniqueIndex:idx_ticket_link_pair" json:"source_ticket_id"`
	TargetTicketID  uint      `gorm:"not null;index;uniqueIndex:idx_ticket_link_pair" json:"target_ticket_id"`
	CreatedAt       time.Time `json:"created_at"`
}
