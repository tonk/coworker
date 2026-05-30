package models

import "time"

type Ticket struct {
	ID                    uint            `gorm:"primaryKey" json:"id"`
	CustomerID            *uint           `gorm:"index" json:"customer_id"`
	Customer              *Customer       `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Title                 string          `gorm:"not null;size:500" json:"title"`
	Description           string          `gorm:"type:text" json:"description"`
	Type                  string          `gorm:"not null;size:30;default:'incident'" json:"type"`
	Status                string          `gorm:"not null;size:20;default:'new'" json:"status"`
	Priority              string          `gorm:"not null;size:20;default:'medium'" json:"priority"`
	CreatedByID           uint            `gorm:"not null" json:"created_by_id"`
	AssignedToID          *uint           `gorm:"index" json:"assigned_to_id,omitempty"`
	OwnerID               *uint           `gorm:"index" json:"owner_id,omitempty"`
	GroupID               *uint           `gorm:"index" json:"group_id,omitempty"`
	SlaPolicyID           *uint           `gorm:"index" json:"sla_policy_id,omitempty"`
	FirstResponseAt       *time.Time      `json:"first_response_at,omitempty"`
	SlaResponseDeadline   *time.Time      `json:"sla_response_deadline,omitempty"`
	SlaResolutionDeadline *time.Time      `json:"sla_resolution_deadline,omitempty"`
	SlaResponseBreached   bool            `json:"sla_response_breached"`
	SlaResolutionBreached bool            `json:"sla_resolution_breached"`
	ReminderAt            *time.Time      `json:"reminder_at,omitempty"`
	CloseAt               *time.Time      `json:"close_at,omitempty"`
	IsSpam                bool            `gorm:"default:false" json:"is_spam"`
	ChecklistTemplateID   *uint           `gorm:"index" json:"checklist_template_id,omitempty"`
	EmailMessageID        *string         `gorm:"uniqueIndex;size:998" json:"email_message_id,omitempty"`
	FromEmail             *string         `gorm:"size:254" json:"from_email,omitempty"`
	FromName              *string         `gorm:"size:150" json:"from_name,omitempty"`
	RawEmail              *string         `gorm:"type:text" json:"raw_email,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	Messages              []TicketMessage `gorm:"foreignKey:TicketID" json:"messages,omitempty"`
	Tags                  []TicketTag    `gorm:"foreignKey:TicketID" json:"tags,omitempty"`
	CreatedBy             User            `json:"created_by,omitempty"`
	AssignedTo            *User           `json:"assigned_to,omitempty"`
	Owner                 *User           `json:"owner,omitempty"`
	Group                 *UserGroup      `json:"group,omitempty"`
	SlaPolicy             *SlaPolicy      `json:"sla_policy,omitempty"`
	Attachments           []Attachment           `gorm:"-" json:"attachments,omitempty"`
	ChecklistItems        []TicketChecklistItem  `gorm:"foreignKey:TicketID" json:"checklist_items,omitempty"`
}

type TicketMessage struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	TicketID  uint         `gorm:"not null;index" json:"ticket_id"`
	UserID    uint         `gorm:"not null" json:"user_id"`
	Body      string       `gorm:"type:text;not null" json:"body"`
	FromName  string       `gorm:"size:150" json:"from_name"`
	EmailSent bool         `gorm:"default:false" json:"email_sent"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	User      User         `json:"user,omitempty"`
	Attachments []Attachment `gorm:"-" json:"attachments,omitempty"`
}

// TicketView records the last time each user viewed a ticket.
type TicketView struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	TicketID uint      `gorm:"not null;uniqueIndex:idx_ticket_view" json:"ticket_id"`
	UserID   uint      `gorm:"not null;uniqueIndex:idx_ticket_view" json:"user_id"`
	ViewedAt time.Time `json:"viewed_at"`
	User     User      `json:"user"`
}

// TicketHistory records activity events on a ticket.
type TicketHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	TicketID  uint      `gorm:"not null;index" json:"ticket_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `json:"user"`
	EventType string    `gorm:"size:50;not null" json:"event_type"`
	Detail    string    `gorm:"size:500" json:"detail"`
}
