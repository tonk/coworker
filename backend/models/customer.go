package models

import "time"

// Customer represents a client organisation that can own projects and contracts.
type Customer struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"not null;size:200" json:"name"`
	Description      string    `gorm:"type:text" json:"description"`
	LogoURL          string    `gorm:"size:500" json:"logo_url"`
	Position         int       `gorm:"default:0" json:"position"`
	TimeTrackingOnly bool      `gorm:"default:false" json:"time_tracking_only"`
	CreatedByID      *uint     `gorm:"index" json:"created_by_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Contract is an agreement between a Customer and the organisation, under which
// one or more Projects are delivered.
type Contract struct {
	ID            uint                 `gorm:"primaryKey" json:"id"`
	CustomerID    uint                 `gorm:"not null;index" json:"customer_id"`
	Name          string               `gorm:"not null;size:200" json:"name"`
	Description   string               `gorm:"type:text" json:"description"`
	StartDate     *time.Time           `json:"start_date,omitempty"`
	EndDate       *time.Time           `json:"end_date,omitempty"`
	PricePerHour  *float64             `json:"price_per_hour,omitempty"`
	Currency      string               `gorm:"size:3;default:€" json:"currency"`
	TimeSlots     []ContractTimeSlot   `gorm:"foreignKey:ContractID" json:"time_slots,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// ContractTimeSlot defines an alternative rate for work performed outside standard office hours.
// DayType is one of "all", "weekdays", or "weekends".
type ContractTimeSlot struct {
	ID                   uint     `gorm:"primaryKey" json:"id"`
	ContractID           uint     `gorm:"not null;index" json:"contract_id"`
	Label                string   `gorm:"size:100" json:"label"`
	StartTime            string   `gorm:"size:5;not null" json:"start_time"`
	EndTime              string   `gorm:"size:5;not null" json:"end_time"`
	DayType              string   `gorm:"size:10;default:all" json:"day_type"`
	MultiplicationFactor *float64 `json:"multiplication_factor,omitempty"`
	HourlyRate           *float64 `json:"hourly_rate,omitempty"`
}

// CustomerFavorite records that a user has starred a customer for quick sidebar access.
type CustomerFavorite struct {
	UserID     uint `gorm:"primaryKey" json:"user_id"`
	CustomerID uint `gorm:"primaryKey" json:"customer_id"`
}

// CustomerAccess grants a non-admin user explicit visibility of a customer.
// When a user has at least one row here they can only see those customers.
// Users with no rows see all customers. Admins always see all customers.
// Role "member" = read access only; "admin" = can manage contracts and access.
type CustomerAccess struct {
	UserID     uint   `gorm:"primaryKey" json:"user_id"`
	CustomerID uint   `gorm:"primaryKey" json:"customer_id"`
	Role       string `gorm:"default:member" json:"role"`
}
