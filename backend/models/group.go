package models

import "time"

// UserGroup is a named set of users that can be granted uniform access to
// projects and customers. Created and managed by admins; project/customer
// owners can add users to groups that are already linked to their resource.
type UserGroup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;uniqueIndex;size:200" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GroupMember links a user to a UserGroup.
type GroupMember struct {
	GroupID uint `gorm:"primaryKey" json:"group_id"`
	UserID  uint `gorm:"primaryKey" json:"user_id"`
}

// GroupProjectAccess grants a group a role on a specific project.
// Role mirrors ProjectMember.Role: "viewer" | "member" | "owner"
type GroupProjectAccess struct {
	GroupID   uint   `gorm:"primaryKey" json:"group_id"`
	ProjectID uint   `gorm:"primaryKey" json:"project_id"`
	Role      string `gorm:"not null;default:'member'" json:"role"`
}

// GroupCustomerAccess grants a group a role on a specific customer.
// Role: "viewer" | "member" | "owner"
type GroupCustomerAccess struct {
	GroupID    uint   `gorm:"primaryKey" json:"group_id"`
	CustomerID uint   `gorm:"primaryKey" json:"customer_id"`
	Role       string `gorm:"not null;default:'member'" json:"role"`
}
