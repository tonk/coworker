package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/testutil"
)

func TestNormalizeMFARememberDevicesPolicy(t *testing.T) {
	assert.Equal(t, "disabled", normalizeMFARememberDevicesPolicy("disabled"))
	assert.Equal(t, "week", normalizeMFARememberDevicesPolicy("week"))
	assert.Equal(t, "week_month", normalizeMFARememberDevicesPolicy("week_month"))
	assert.Equal(t, "week_month", normalizeMFARememberDevicesPolicy("invalid"))
	assert.Equal(t, "week_month", normalizeMFARememberDevicesPolicy(""))
}

func TestNormalizeMFARememberDays(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	setPolicy := func(v string) {
		saveSetting(settingMFARememberDevices, v)
	}

	setPolicy("disabled")
	assert.Equal(t, 0, NormalizeMFARememberDays(0))
	assert.Equal(t, 0, NormalizeMFARememberDays(7))
	assert.Equal(t, 0, NormalizeMFARememberDays(30))

	setPolicy("week")
	assert.Equal(t, 0, NormalizeMFARememberDays(0))
	assert.Equal(t, 7, NormalizeMFARememberDays(7))
	assert.Equal(t, 0, NormalizeMFARememberDays(30))

	setPolicy("week_month")
	assert.Equal(t, 0, NormalizeMFARememberDays(0))
	assert.Equal(t, 7, NormalizeMFARememberDays(7))
	assert.Equal(t, 30, NormalizeMFARememberDays(30))
	assert.Equal(t, 0, NormalizeMFARememberDays(14))
}

func TestApplyMFARememberDevicesPolicyChange(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	now := time.Now()
	weekDevice := models.MFATrustedDevice{
		UserID:     1,
		TokenHash:  "weekhash",
		DeviceName: "Week device",
		LastUsedAt: now,
		ExpiresAt:  now.Add(7 * 24 * time.Hour),
		CreatedAt:  now,
	}
	monthDevice := models.MFATrustedDevice{
		UserID:     1,
		TokenHash:  "monthhash",
		DeviceName: "Month device",
		LastUsedAt: now,
		ExpiresAt:  now.Add(30 * 24 * time.Hour),
		CreatedAt:  now,
	}
	db.Create(&weekDevice)
	db.Create(&monthDevice)

	applyMFARememberDevicesPolicyChange("week_month", "disabled")

	var count int64
	db.Model(&models.MFATrustedDevice{}).Count(&count)
	assert.Equal(t, int64(0), count)

	db.Create(&weekDevice)
	db.Create(&monthDevice)
	applyMFARememberDevicesPolicyChange("week_month", "week")

	db.Model(&models.MFATrustedDevice{}).Count(&count)
	assert.Equal(t, int64(1), count)

	var remaining models.MFATrustedDevice
	db.First(&remaining)
	assert.Equal(t, "weekhash", remaining.TokenHash)
}
