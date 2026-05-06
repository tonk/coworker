package models

import (
	"time"

	"gorm.io/gorm"
)

type Release struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	ProjectID  uint           `gorm:"not null;index" json:"project_id"`
	Name       string         `gorm:"not null;size:200" json:"name"`
	Goal       string         `gorm:"type:text" json:"goal"`
	TargetDate *time.Time     `json:"target_date"`
	Sprints    []Sprint       `gorm:"-" json:"sprints,omitempty"`
}

type ReleaseSprint struct {
	ReleaseID uint `gorm:"primaryKey;autoIncrement:false" json:"release_id"`
	SprintID  uint `gorm:"primaryKey;autoIncrement:false" json:"sprint_id"`
}
