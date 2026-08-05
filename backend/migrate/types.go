// Package migrate provides canonical in-memory types used as a pivot between
// WarmDesk and external project management platforms (Jira, Trello,
// OpenProject, Ryver).
package migrate

import "time"

// Project is the top-level canonical representation of a WarmDesk project.
type Project struct {
	Name        string
	Description string
	Columns     []Column
	Topics      []Topic
	Messages    []ChatMessage
}

// Column represents a Kanban column (list) inside a project.
type Column struct {
	Name  string
	Cards []Card
}

// Card represents a single work item / ticket / task.
type Card struct {
	Ref         string // e.g. "PRJ-1"
	Title       string
	Description string
	Priority    string // none | low | medium | high | critical
	StartDate   string // YYYY-MM-DD or ""
	DueDate     string // YYYY-MM-DD or ""
	Closed      bool
	Creator     string   // display name or username of whoever created the source item; used as the primary assignee if resolvable
	Assignees   []string // display names or emails
	Labels      []Label
	Tags        []string
	Checklist   []CheckItem
	Comments    []Comment
	TimeMinutes int
	Attachments []Attachment
}

// Label is a coloured tag attached to a card.
type Label struct {
	Name  string
	Color string
}

// CheckItem is a single entry in a card's checklist.
type CheckItem struct {
	Text string
	Done bool
}

// Comment is a single comment on a card.
type Comment struct {
	Author      string
	Body        string
	CreatedAt   time.Time
	Attachments []Attachment
}

// Attachment is a file attached to a card.
type Attachment struct {
	Filename string
	URL      string
	MimeType string
}

// Topic is a threaded discussion thread inside a project (forum post).
type Topic struct {
	Title   string
	Body    string
	Author  string
	Replies []TopicReply
}

// TopicReply is a single reply inside a topic thread.
type TopicReply struct {
	Author string
	Body   string
}

// ChatMessage is a single message in a project's team/group chat, as opposed
// to a Card comment or Topic reply.
type ChatMessage struct {
	Author     string
	Body       string
	CreatedAt  time.Time
	Attachment *Attachment // at most one, per Ryver's chat message model
}
