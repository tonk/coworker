package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type TicketChecklistTemplateItems []string

func (a TicketChecklistTemplateItems) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	return string(b), err
}

func (a *TicketChecklistTemplateItems) Scan(src any) error {
	var s string
	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("cannot scan type %T into TicketChecklistTemplateItems", src)
	}
	if s == "" || s == "null" {
		*a = TicketChecklistTemplateItems{}
		return nil
	}
	return json.Unmarshal([]byte(s), a)
}

type TicketChecklistTemplate struct {
	ID          uint                         `gorm:"primaryKey" json:"id"`
	Name        string                       `gorm:"not null;size:200" json:"name"`
	Description string                       `gorm:"size:500" json:"description"`
	Items       TicketChecklistTemplateItems `gorm:"type:text" json:"items"`
	IsActive    bool                         `gorm:"not null;default:true" json:"is_active"`
	SortOrder   int                          `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
}

type TicketChecklistItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TicketID    uint      `gorm:"not null;index" json:"ticket_id"`
	Body        string    `gorm:"type:text;not null" json:"body"`
	IsCompleted bool      `gorm:"default:false" json:"is_completed"`
	Position    float64   `gorm:"default:0" json:"position"`
}
