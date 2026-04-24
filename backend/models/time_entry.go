package models

import "time"

type TimeEntry struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	CustomerID  *uint     `gorm:"index" json:"customer_id"`
	Customer    *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	ProjectID   *uint     `gorm:"index" json:"project_id"`
	Project     *Project  `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Date        time.Time `gorm:"not null;index" json:"date"`
	Minutes     int       `gorm:"not null" json:"minutes"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
