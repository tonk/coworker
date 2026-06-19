package models

import "time"

// CustomerContact is a contact person at a customer.
type CustomerContact struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CustomerID uint      `gorm:"not null;index" json:"customer_id"`
	Name       string    `gorm:"size:200" json:"name"`
	Department string    `gorm:"size:200" json:"department"`
	Phone      string    `gorm:"size:100" json:"phone"`
	Email      string    `gorm:"size:200" json:"email"`
	IsPrimary  bool      `gorm:"default:false" json:"is_primary"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
