package models

import "time"

type TimeEntry struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CustomerID  *uint     `gorm:"index" json:"customer_id"`
	Customer    *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	ProjectID   *uint     `gorm:"index" json:"project_id"`
	Project     *Project  `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	TicketID    *uint     `gorm:"index" json:"ticket_id,omitempty"`
	Ticket      *Ticket   `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	Date        time.Time `gorm:"not null;index" json:"date"`
	Minutes     int     `gorm:"not null" json:"minutes"`
	Description string  `json:"description"`
	IsHoliday   bool    `gorm:"default:false" json:"is_holiday"`
	StartTime   *string `gorm:"size:5" json:"start_time,omitempty"`
	EndTime     *string `gorm:"size:5" json:"end_time,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
