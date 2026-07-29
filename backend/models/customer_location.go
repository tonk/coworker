package models

import "time"

// CustomerLocation is a physical location (site) belonging to a customer.
type CustomerLocation struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CustomerID     uint      `gorm:"not null;index" json:"customer_id"`
	AddressLine1   string    `gorm:"size:300" json:"address_line1"`
	AddressLine2   string    `gorm:"size:300" json:"address_line2"`
	City           string    `gorm:"size:200" json:"city"`
	PostalCode     string    `gorm:"size:20" json:"postal_code"`
	Region         string    `gorm:"size:200" json:"region"`
	Country        string    `gorm:"size:100" json:"country"`
	Phone          string    `gorm:"size:100" json:"phone"`
	ContactName    string    `gorm:"size:200" json:"contact_name"`
	ContactEmail   string    `gorm:"size:200" json:"contact_email"`
	ContactPhone   string    `gorm:"size:100" json:"contact_phone"`
	TravelDistance *float64  `json:"travel_distance,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
