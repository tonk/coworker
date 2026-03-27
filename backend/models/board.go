package models

import (
	"time"

	"gorm.io/gorm"
)

type Column struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	ProjectID uint           `gorm:"not null;index" json:"project_id"`
	Project   Project        `json:"-"`
	Name      string         `gorm:"not null;size:200" json:"name"`
	Position  float64        `gorm:"not null;default:0" json:"position"`
	Color     string         `gorm:"size:7" json:"color"`
	WIPLimit  *int           `json:"wip_limit"`
	Cards     []Card         `json:"cards,omitempty"`
}

type Card struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	ColumnID    uint           `gorm:"not null;index" json:"column_id"`
	Column      Column         `json:"-"`
	ProjectID   uint           `gorm:"not null;index" json:"project_id"`
	Title       string         `gorm:"not null;size:500" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Position    float64        `gorm:"not null;default:0" json:"position"`
	DueDate     *time.Time     `json:"due_date"`
	Priority    string         `gorm:"size:20;default:'none'" json:"priority"`
	AssigneeID  *uint          `json:"assignee_id"`
	Assignee    *User          `json:"assignee,omitempty"`
	CreatedByID uint           `gorm:"not null" json:"created_by_id"`
	CreatedBy   User           `json:"created_by"`
	Labels      []Label        `gorm:"many2many:card_labels" json:"labels,omitempty"`
	Watchers    []User         `gorm:"many2many:card_watchers" json:"watchers,omitempty"`
	Comments    []CardComment  `json:"comments,omitempty"`
}

type CardComment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CardID    uint           `gorm:"not null;index" json:"card_id"`
	Card      Card           `json:"-"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `json:"user"`
	Body      string         `gorm:"type:text;not null" json:"body"`
	IsEdited  bool           `gorm:"default:false" json:"is_edited"`
}

type Label struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	ProjectID uint           `gorm:"not null;index" json:"project_id"`
	Project   Project        `json:"-"`
	Name      string         `gorm:"not null;size:100" json:"name"`
	Color     string         `gorm:"not null;size:7" json:"color"`
	Cards     []Card         `gorm:"many2many:card_labels" json:"-"`
}

type CardLabel struct {
	CardID    uint      `gorm:"primaryKey" json:"card_id"`
	LabelID   uint      `gorm:"primaryKey" json:"label_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CardHistory records every column change for a card.
type CardHistory struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	CardID       uint      `gorm:"not null;index" json:"card_id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	User         User      `json:"user"`
	FromColumnID uint      `json:"from_column_id"`
	FromColumn   Column    `gorm:"foreignKey:FromColumnID" json:"from_column"`
	ToColumnID   uint      `gorm:"not null" json:"to_column_id"`
	ToColumn     Column    `gorm:"foreignKey:ToColumnID" json:"to_column"`
}
