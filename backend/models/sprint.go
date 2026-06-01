package models

import (
	"time"

	"gorm.io/gorm"
)

type Sprint struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	ProjectID uint           `gorm:"not null;index" json:"project_id"`
	Name      string         `gorm:"not null;size:200" json:"name"`
	Goal      string         `gorm:"type:text" json:"goal"`
	Position  float64    `gorm:"default:0" json:"position"`
	// "planning" | "active" | "completed"
	Status    string     `gorm:"size:20;default:'planning'" json:"status"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	// Computed at query time, not stored
	TotalPoints     int    `gorm:"-" json:"total_points"`
	CompletedPoints int    `gorm:"-" json:"completed_points"`
	CardCount       int    `gorm:"-" json:"card_count"`
	CardIDs         []uint `gorm:"-" json:"card_ids"`
}

// SprintCard is the join table linking cards to sprints.
// Uses a composite primary key; hard deletes only.
type SprintCard struct {
	SprintID  uint      `gorm:"primaryKey;autoIncrement:false" json:"sprint_id"`
	CardID    uint      `gorm:"primaryKey;autoIncrement:false" json:"card_id"`
	Position  float64   `gorm:"default:0" json:"position"`
	CreatedAt time.Time `json:"created_at"`
}
