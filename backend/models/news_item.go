package models

import (
	"time"

	"gorm.io/gorm"
)

type NewsItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Title     string         `gorm:"not null;size:200" json:"title"`
	Text      string         `gorm:"type:text;not null" json:"text"`
	StartDate *time.Time     `json:"start_date"`
	EndDate   *time.Time     `json:"end_date"`
	Active    bool           `gorm:"default:true;not null" json:"active"`
}
