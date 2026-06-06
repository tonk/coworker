package models

import "time"

// MFATrustedDevice records a device that has successfully completed MFA and
// elected to be remembered. The plaintext token lives only in the browser's
// httpOnly cookie; only the SHA-256 hash is stored here.
type MFATrustedDevice struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"-"`
	TokenHash  string    `gorm:"size:64;not null;uniqueIndex" json:"-"`
	DeviceName string    `gorm:"size:200" json:"device_name"`
	LastUsedAt time.Time `json:"last_used_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}
