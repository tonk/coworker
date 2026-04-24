package models

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// GravatarEnabledFn is wired up by the handlers package so the model layer
// can check the admin setting without creating a circular import.
var GravatarEnabledFn func() bool

type User struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	Email               string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Username            string         `gorm:"uniqueIndex;not null;size:100" json:"username"`
	PasswordHash        string         `gorm:"not null" json:"-"`
	FirstName           string         `gorm:"size:100" json:"first_name"`
	LastName            string         `gorm:"size:100" json:"last_name"`
	DisplayName         string         `gorm:"size:150" json:"display_name"`
	AvatarURL           string         `gorm:"size:500" json:"avatar_url"`
	GlobalRole          string         `gorm:"not null;default:'user'" json:"global_role"` // "admin" | "user" | "viewer" | "metrics"
	Locale              string         `gorm:"size:10;default:'en'" json:"locale"`
	Theme               string         `gorm:"size:20;default:'system'" json:"theme"` // "light" | "dark" | "system"
	DateTimeFormat      string         `gorm:"size:50;default:'YYYY-MM-DD HH:mm'" json:"date_time_format"`
	Timezone            string         `gorm:"size:100;default:'UTC'" json:"timezone"`
	Font                string         `gorm:"size:100;default:'system'" json:"font"`
	FontSize            string         `gorm:"size:10;default:'14'" json:"font_size"`
	SidebarPosition     string         `gorm:"size:10;default:'left'" json:"sidebar_position"`
	ShowBreadcrumbs     bool           `gorm:"default:true" json:"show_breadcrumbs"`
	AccentColor         string         `gorm:"size:20;default:'blue'" json:"accent_color"` // "blue" | "red" | "green" | "orange"
	LastLoginAt         *time.Time     `json:"last_login_at"`
	SettingsUpdatedAt   *time.Time     `json:"settings_updated_at"`
	IsActive            bool           `gorm:"default:true" json:"is_active"`
	EmailNotifications  bool           `gorm:"default:true" json:"email_notifications"`
	TOTPSecret          string         `gorm:"size:64" json:"-"`
	TOTPEnabled         bool           `gorm:"default:false" json:"totp_enabled"`
	PasswordResetToken  string         `gorm:"size:64;index" json:"-"`
	PasswordResetExpiry *time.Time     `json:"-"`

	// Computed — not stored in DB
	GravatarURL    string `gorm:"-" json:"gravatar_url"`
	CanViewReports bool   `gorm:"-" json:"can_view_reports"`
}

// AfterFind populates the computed GravatarURL field after every DB read.
func (u *User) AfterFind(tx *gorm.DB) error {
	if u.Username == "" {
		return nil
	}
	if GravatarEnabledFn != nil && GravatarEnabledFn() && u.Email != "" {
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(u.Email))))
		u.GravatarURL = fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=mp&s=80", hash)
	} else {
		u.GravatarURL = fmt.Sprintf("https://api.dicebear.com/7.x/initials/svg?seed=%s", u.Username)
	}
	return nil
}
