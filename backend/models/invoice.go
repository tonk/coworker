package models

import "time"

// InvoiceStatus values.
const (
	InvoiceStatusDraft      = "draft"
	InvoiceStatusSent       = "sent"
	InvoiceStatusPaid       = "paid"
	InvoiceStatusCreditNote = "credit_note"
)

// Invoice is a billable document generated from time entries for a customer.
type Invoice struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	InvoiceNumber     string     `gorm:"size:50;uniqueIndex" json:"invoice_number"`
	CustomerID        uint       `gorm:"not null;index" json:"customer_id"`
	Customer          *Customer  `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	PeriodStart       time.Time  `json:"period_start"`
	PeriodEnd         time.Time  `json:"period_end"`
	Status            string     `gorm:"size:20;default:draft" json:"status"`
	Currency          string     `gorm:"size:10;default:€" json:"currency"`
	LineItems         string     `gorm:"type:text" json:"line_items"` // JSON-encoded []InvoiceLineItem
	Subtotal          float64    `json:"subtotal"`
	VATRate           float64    `gorm:"default:0" json:"vat_rate"`
	VATAmount         float64    `json:"vat_amount"`
	Total             float64    `json:"total"`
	DueDate           *time.Time `json:"due_date,omitempty"`
	Notes             string     `gorm:"type:text" json:"notes"`
	PaymentDate       *time.Time `json:"payment_date,omitempty"`
	PaymentAmount     *float64   `json:"payment_amount,omitempty"`
	PaymentReference  string     `gorm:"size:200" json:"payment_reference"`
	CreditedInvoiceID *uint      `gorm:"index" json:"credited_invoice_id,omitempty"`
	CreatedByID       *uint      `gorm:"index" json:"created_by_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// InvoiceLineItem is one billable row on an invoice, serialised into Invoice.LineItems.
type InvoiceLineItem struct {
	Date        string  `json:"date"`
	ProjectName string  `json:"project_name"`
	Description string  `json:"description"`
	Minutes     int     `json:"minutes"`
	HourlyRate  float64 `json:"hourly_rate,omitempty"`
	Distance    float64 `json:"distance,omitempty"`
	PricePerKm  float64 `json:"price_per_km,omitempty"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Quantity    float64 `json:"quantity,omitempty"`
	UnitPrice   float64 `json:"unit_price,omitempty"`
	IsManual    bool    `json:"is_manual,omitempty"`
}
