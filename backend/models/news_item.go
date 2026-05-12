package models

import (
	"time"

	"gorm.io/gorm"
)

type NewsItem struct {
	gorm.Model
	Title     string     `gorm:"not null;size:200" json:"title"`
	Text      string     `gorm:"type:text;not null" json:"text"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Active    bool       `gorm:"default:true;not null" json:"active"`
}
