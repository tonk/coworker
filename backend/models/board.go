package models

import (
	"time"

	"gorm.io/gorm"
)

type Epic struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	ProjectID   uint           `gorm:"not null;index" json:"project_id"`
	Name        string         `gorm:"not null;size:200" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Color       string         `gorm:"size:7;default:'#6366f1'" json:"color"`
	Status      string         `gorm:"size:20;default:'open'" json:"status"`
	Position    float64        `gorm:"default:0" json:"position"`
	// Computed at query time
	CardCount int `gorm:"-" json:"card_count"`
	DoneCount int `gorm:"-" json:"done_count"`
}

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
	WIPLimit  *int           `gorm:"column:wip_limit" json:"wip_limit"`
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
	StartDate   *time.Time     `json:"start_date"`
	DueDate     *time.Time     `json:"due_date"`
	Priority    string         `gorm:"size:20;default:'none'" json:"priority"`
	AssigneeID  *uint          `json:"assignee_id"`
	Assignee    *User          `json:"assignee,omitempty"`
	CreatedByID uint           `gorm:"not null" json:"created_by_id"`
	CreatedBy   User           `json:"created_by"`
	CardNumber        int            `gorm:"default:0" json:"card_number"`
	TimeSpentMinutes  int            `gorm:"default:0" json:"time_spent_minutes"`
	StoryPoints       *int           `json:"story_points"`
	Closed            bool           `gorm:"default:false" json:"closed"`
	ClosedAt          *time.Time     `json:"closed_at"`
	ExternalIssueURL  string         `gorm:"size:2000" json:"external_issue_url"`
	ExternalIssueRef  string         `gorm:"size:200" json:"external_issue_ref"`
	EpicID            *uint          `gorm:"index" json:"epic_id"`
	Epic              *Epic          `gorm:"foreignKey:EpicID" json:"epic,omitempty"`
	ParentCardID      *uint          `gorm:"index" json:"parent_card_id,omitempty"`
	SubCardCount      int            `gorm:"-" json:"sub_card_count"`
	SubCardsDone      int            `gorm:"-" json:"sub_cards_done"`
	Labels      []Label        `gorm:"many2many:card_labels" json:"labels,omitempty"`
	Tags        []CardTag      `json:"tags,omitempty"`
	Assignees   []User         `gorm:"many2many:card_assignees" json:"assignees,omitempty"`
	Watchers    []User         `gorm:"many2many:card_watchers" json:"watchers,omitempty"`
	Comments    []CardComment  `json:"comments,omitempty"`
	Attachments []Attachment   `gorm:"-" json:"attachments,omitempty"`
}

// CardAssignee is the join table for multiple card assignees.
type CardAssignee struct {
	CardID uint `gorm:"primaryKey" json:"card_id"`
	UserID uint `gorm:"primaryKey" json:"user_id"`
}

// CardChecklistItem is a single checklist item on a card.
type CardChecklistItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CardID      uint      `gorm:"not null;index" json:"card_id"`
	Body        string    `gorm:"type:text;not null" json:"body"`
	IsCompleted bool      `gorm:"default:false" json:"is_completed"`
	Position    float64   `gorm:"default:0" json:"position"`
}

type CardComment struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	CardID           uint           `gorm:"not null;index" json:"card_id"`
	Card             Card           `json:"-"`
	UserID           uint           `gorm:"not null" json:"user_id"`
	User             User           `json:"user"`
	Body             string         `gorm:"type:text;not null" json:"body"`
	IsEdited         bool           `gorm:"default:false" json:"is_edited"`
	TimeSpentMinutes int            `gorm:"default:0" json:"time_spent_minutes"`
	TimeEntryID      *uint          `gorm:"index" json:"time_entry_id,omitempty"`
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

// CardReference is a bidirectional "relates to" link between two cards.
// The link is stored once (source → target); both sides are shown when listing.
type CardReference struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SourceCardID uint      `gorm:"not null;index;uniqueIndex:idx_card_ref_pair" json:"source_card_id"`
	TargetCardID uint      `gorm:"not null;index;uniqueIndex:idx_card_ref_pair" json:"target_card_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// CardHistory records activity events on a card (creation, column moves, status changes, etc.).
type CardHistory struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	CardID       uint      `gorm:"not null;index" json:"card_id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	User         User      `json:"user"`
	EventType    string    `gorm:"size:50;default:'column_move'" json:"event_type"`
	Detail       string    `gorm:"size:500" json:"detail"`
	// Nullable: a "created" event has no from/to column transition, and 0 is
	// not a valid column id — inserting it violates the from/to foreign key
	// constraints on any database that actually enforces them (MySQL,
	// PostgreSQL; SQLite does not by default, which let this go unnoticed).
	FromColumnID *uint  `json:"from_column_id"`
	FromColumn   Column `gorm:"foreignKey:FromColumnID" json:"from_column"`
	ToColumnID   *uint  `json:"to_column_id"`
	ToColumn     Column `gorm:"foreignKey:ToColumnID" json:"to_column"`
}
