package models

import "time"

// TimeEntryRowOrder persists the user's custom row ordering for the time-tracking grid.
// One row per user; OrderedKeys is a JSON array of row-key strings.
type TimeEntryRowOrder struct {
	UserID      uint   `gorm:"primaryKey" json:"user_id"`
	OrderedKeys string `gorm:"type:text" json:"ordered_keys"`
}

// TimeEntryWeekRowOrder persists row-key order for a specific ISO week, including empty rows.
// Comments is a JSON object mapping rowKey → comment text.
type TimeEntryWeekRowOrder struct {
	UserID      uint   `gorm:"primaryKey" json:"user_id"`
	Year        int    `gorm:"primaryKey" json:"year"`
	Week        int    `gorm:"primaryKey" json:"week"`
	OrderedKeys string `gorm:"type:text" json:"ordered_keys"`
	Comments    string `gorm:"type:text" json:"comments"`
}

type TimeEntry struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CustomerID  *uint     `gorm:"index" json:"customer_id"`
	Customer    *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	ProjectID   *uint     `gorm:"index" json:"project_id"`
	Project     *Project  `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	ContractID  *uint     `gorm:"index" json:"contract_id"`
	Contract    *Contract `gorm:"foreignKey:ContractID" json:"contract,omitempty"`
	TicketID    *uint     `gorm:"index" json:"ticket_id,omitempty"`
	Ticket      *Ticket   `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	Date        time.Time `gorm:"not null;index" json:"date"`
	Minutes     int     `gorm:"not null" json:"minutes"`
	Description string  `json:"description"`
	IsHoliday   bool     `gorm:"default:false" json:"is_holiday"`
	StartTime   *string  `gorm:"size:5" json:"start_time,omitempty"`
	EndTime     *string  `gorm:"size:5" json:"end_time,omitempty"`
	Distance    *float64 `json:"distance,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
