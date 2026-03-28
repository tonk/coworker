package models

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Username     string         `gorm:"uniqueIndex;not null;size:100" json:"username"`
	PasswordHash string         `gorm:"not null" json:"-"`
	FirstName    string         `gorm:"size:100" json:"first_name"`
	LastName     string         `gorm:"size:100" json:"last_name"`
	DisplayName  string         `gorm:"size:150" json:"display_name"`
	AvatarURL    string         `gorm:"size:500" json:"avatar_url"`
	GlobalRole   string         `gorm:"not null;default:'user'" json:"global_role"` // "admin" | "user" | "viewer"
	Locale         string         `gorm:"size:10;default:'en'" json:"locale"`
	Theme          string         `gorm:"size:20;default:'system'" json:"theme"` // "light" | "dark" | "system"
	DateTimeFormat  string `gorm:"size:50;default:'YYYY-MM-DD HH:mm'" json:"date_time_format"`
	Timezone        string `gorm:"size:100;default:'UTC'" json:"timezone"`
	Font            string `gorm:"size:100;default:'system'" json:"font"`
	FontSize        string `gorm:"size:10;default:'14'" json:"font_size"`
	SidebarPosition string `gorm:"size:10;default:'left'" json:"sidebar_position"`
	LastLoginAt        *time.Time `json:"last_login_at"`
	SettingsUpdatedAt  *time.Time `json:"settings_updated_at"`
	IsActive           bool       `gorm:"default:true" json:"is_active"`
	EmailNotifications bool       `gorm:"default:true" json:"email_notifications"`

	// Computed — not stored in DB
	GravatarURL string `gorm:"-" json:"gravatar_url"`
}

// AfterFind populates the computed GravatarURL field after every DB read.
func (u *User) AfterFind(tx *gorm.DB) error {
	if u.Email != "" {
		h := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(u.Email))))
		u.GravatarURL = fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=identicon&s=80", h)
	}
	return nil
}
