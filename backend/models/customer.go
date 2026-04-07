package models

import "time"

// Customer represents a client organisation that can own projects and contracts.
type Customer struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;size:200" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	LogoURL     string    `gorm:"size:500" json:"logo_url"`
	Position    int       `gorm:"default:0" json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Contract is an agreement between a Customer and the organisation, under which
// one or more Projects are delivered.
type Contract struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	CustomerID  uint       `gorm:"not null;index" json:"customer_id"`
	Name        string     `gorm:"not null;size:200" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CustomerFavorite records that a user has starred a customer for quick sidebar access.
type CustomerFavorite struct {
	UserID     uint `gorm:"primaryKey" json:"user_id"`
	CustomerID uint `gorm:"primaryKey" json:"customer_id"`
}
