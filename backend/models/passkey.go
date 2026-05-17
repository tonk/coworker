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
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at"`
}
