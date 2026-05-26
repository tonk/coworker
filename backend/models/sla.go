package models

import "time"

type SlaPolicy struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	Name                  string    `gorm:"not null;size:200" json:"name"`
	ResponseTimeMinutes   int       `gorm:"not null;default:0" json:"response_time_minutes"`
	ResolutionTimeMinutes int       `gorm:"not null;default:0" json:"resolution_time_minutes"`
	PriorityFilter        string    `gorm:"size:200" json:"priority_filter"`
	IsActive              bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
