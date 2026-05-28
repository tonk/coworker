package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type MacroAction struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type MacroActions []MacroAction

func (a MacroActions) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	return string(b), err
}

func (a *MacroActions) Scan(src any) error {
	var s string
	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("cannot scan type %T into MacroActions", src)
	}
	if s == "" || s == "null" {
		*a = MacroActions{}
		return nil
	}
	return json.Unmarshal([]byte(s), a)
}

type Macro struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"not null;size:200" json:"name"`
	Description string       `gorm:"size:500" json:"description"`
	Actions     MacroActions `gorm:"type:text" json:"actions"`
	IsActive    bool         `gorm:"not null;default:true" json:"is_active"`
	SortOrder   int          `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
