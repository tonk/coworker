package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
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

func setupMFAUser(t *testing.T, authSvc *services.AuthService) (models.User, string) {
	t.Helper()
	secret, _, err := authSvc.GenerateTOTPSecret("mfatest", "WarmDesk")
	require.NoError(t, err)
	hash, err := authSvc.HashPassword("pass")
	require.NoError(t, err)
	user := models.User{
		Username:     "mfatest",
		Email:        "mfatest@example.com",
		PasswordHash: hash,
		TOTPEnabled:  true,
		TOTPSecret:   secret,
		IsActive:     true,
	}
	require.NoError(t, database.DB.Create(&user).Error)
	return user, secret
}

func createTrustedDevice(t *testing.T, userID uint, expires time.Duration) string {
	t.Helper()
	plaintext, hash := generateMFATrustToken()
	device := models.MFATrustedDevice{
		UserID:     userID,
		TokenHash:  hash,
		DeviceName: "Test device",
		LastUsedAt: time.Now(),
		ExpiresAt:  time.Now().Add(expires),
	}
	require.NoError(t, database.DB.Create(&device).Error)
	return plaintext
}

func loginContextWithTrust(token string) (*gin.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "mfa_trust", Value: token})
	}
	return newAuthTestContext(req)
}

func TestIssueMFAChallengeOrSkip_trustedDeviceSkipsChallenge(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db
	saveSetting(settingMFARememberDevices, "week_month")

	authSvc := services.NewAuthService("test-secret-key-that-is-long-enough-32")
	h := NewAuthHandler(authSvc)
	user, _ := setupMFAUser(t, authSvc)
	trust := createTrustedDevice(t, user.ID, 7*24*time.Hour)

	c, w := loginContextWithTrust(trust)
	assert.True(t, h.issueMFAChallengeOrSkip(c, user))
	assert.Empty(t, w.Body.Bytes())
}

func TestIssueMFAChallengeOrSkip_policyDisabledIgnoresTrust(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db
	saveSetting(settingMFARememberDevices, "disabled")

	authSvc := services.NewAuthService("test-secret-key-that-is-long-enough-32")
	h := NewAuthHandler(authSvc)
	user, _ := setupMFAUser(t, authSvc)
	trust := createTrustedDevice(t, user.ID, 7*24*time.Hour)

	c, w := loginContextWithTrust(trust)
	assert.False(t, h.issueMFAChallengeOrSkip(c, user))

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, true, body["mfa_required"])
}

func TestIssueMFAChallengeOrSkip_expiredTrustRequiresChallenge(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db
	saveSetting(settingMFARememberDevices, "week_month")

	authSvc := services.NewAuthService("test-secret-key-that-is-long-enough-32")
	h := NewAuthHandler(authSvc)
	user, _ := setupMFAUser(t, authSvc)
	plaintext, hash := generateMFATrustToken()
	device := models.MFATrustedDevice{
		UserID:     user.ID,
		TokenHash:  hash,
		DeviceName: "Expired device",
		LastUsedAt: time.Now().Add(-48 * time.Hour),
		ExpiresAt:  time.Now().Add(-time.Hour),
	}
	require.NoError(t, database.DB.Create(&device).Error)

	c, w := loginContextWithTrust(plaintext)
	assert.False(t, h.issueMFAChallengeOrSkip(c, user))

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, true, body["mfa_required"])
}

func TestMFAVerifyRememberDaysRespectsPolicy(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db
	saveSetting(settingMFARememberDevices, "week")

	authSvc := services.NewAuthService("test-secret-key-that-is-long-enough-32")
	h := NewAuthHandler(authSvc)
	user, secret := setupMFAUser(t, authSvc)
	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)

	mfaToken, err := authSvc.IssueMFAToken(user.ID, user.Username)
	require.NoError(t, err)

	verify := func(rememberDays int) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]any{
			"mfa_token":     mfaToken,
			"code":          code,
			"remember_days": rememberDays,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/mfa/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c, w := newAuthTestContext(req)
		h.MFAVerify(c)
		return w
	}

	w30 := verify(30)
	assert.Equal(t, http.StatusOK, w30.Code)
	var count int64
	database.DB.Model(&models.MFATrustedDevice{}).Where("user_id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(0), count)

	mfaToken, err = authSvc.IssueMFAToken(user.ID, user.Username)
	require.NoError(t, err)
	w7 := verify(7)
	assert.Equal(t, http.StatusOK, w7.Code)
	database.DB.Model(&models.MFATrustedDevice{}).Where("user_id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(1), count)

	var device models.MFATrustedDevice
	require.NoError(t, database.DB.Where("user_id = ?", user.ID).First(&device).Error)
	assert.WithinDuration(t, time.Now().Add(7*24*time.Hour), device.ExpiresAt, 2*time.Minute)
}
