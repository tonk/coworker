package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "sqlite", cfg.DBDriver)
	assert.Equal(t, "./warmdesk.db", cfg.DBDSN)
	assert.Equal(t, "change-me-in-production", cfg.JWTSecret)
	assert.Equal(t, "http://localhost:5173", cfg.AllowedOrigins)
	assert.Equal(t, "", cfg.WebDir)
	assert.Equal(t, "release", cfg.GinMode)
	assert.Equal(t, "en", cfg.DefaultLocale)
	assert.Equal(t, true, cfg.APILog)
	assert.Equal(t, "./uploads", cfg.UploadDir)
	assert.Equal(t, int64(25), cfg.MaxUploadMB)
	assert.Equal(t, "", cfg.RedisURL)
	assert.Equal(t, "", cfg.DBLog)
	assert.Equal(t, "", cfg.BaseURL)
	assert.Equal(t, "", cfg.LiveKitURL)
	assert.Equal(t, "", cfg.LiveKitAPIKey)
	assert.Equal(t, "", cfg.LiveKitAPISecret)
	assert.Equal(t, "", cfg.LiveKitRoomPrefix)
	assert.Equal(t, "", cfg.TrustedProxies)

	assert.Equal(t, 587, cfg.SMTP.Port)
	assert.Equal(t, "", cfg.SMTP.Host)
	assert.Equal(t, "", cfg.SMTP.From)
	assert.Equal(t, "", cfg.SMTP.Username)
	assert.Equal(t, "", cfg.SMTP.Password)
	assert.Equal(t, false, cfg.SMTP.UseTLS)

	assert.Equal(t, "", cfg.DBTLSMode)
	assert.Equal(t, "", cfg.DBTLSCACert)
	assert.Equal(t, "", cfg.DBTLSCert)
	assert.Equal(t, "", cfg.DBTLSKey)
	assert.Equal(t, "", cfg.TLSCert)
	assert.Equal(t, "", cfg.TLSKey)
}
