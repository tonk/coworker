package models

import "time"

type PasskeyCredential struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"not null;index" json:"-"`
	Name         string     `gorm:"size:100" json:"name"`
	CredentialID []byte     `gorm:"not null;uniqueIndex" json:"-"`
	PublicKey    []byte     `gorm:"not null;type:blob" json:"-"`
	AAGUID       []byte     `gorm:"type:blob" json:"-"`
	SignCount     uint32     `json:"-"`
	Transports   string     `gorm:"size:200" json:"-"`
	// Nil means "never recorded" — true for every passkey registered before this
	// field existed. Login self-heals these on their next successful attempt
	// instead of enforcing go-webauthn's flag-consistency check against a value
	// that was never actually captured at registration time.
	BackupEligible *bool      `json:"-"`
	BackupState    *bool      `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at"`
}
