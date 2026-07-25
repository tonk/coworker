package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/services"
	"github.com/tonk/warmdesk/testutil"
	"gorm.io/gorm"
)

// A regression guard for the class of bug where a genuine database error
// (e.g. a MySQL DSN missing parseTime=true, which fails to scan DATETIME
// columns on every read while writes keep working) gets collapsed into the
// same response as "no such user" — masking a server-side malfunction as an
// ordinary auth failure. Login/Me/Refresh must return 500 for a real DB
// error, and only their normal not-found response when the row truly
// doesn't exist.

const testJWTSecret = "test-secret-at-least-32-characters-long"

// closeDB closes the underlying *sql.DB so subsequent queries fail with a
// generic driver error instead of gorm.ErrRecordNotFound.
func closeDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func decodeErrorBody(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body["error"]
}

func TestMe_DBErrorReturns500NotUserNotFound(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db
	closeDB(t, db)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	c, rec := newAuthTestContext(req)
	c.Set(middleware.ContextUserID, uint(1))

	h := NewAuthHandler(services.NewAuthService(testJWTSecret))
	h.Me(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "internal error", decodeErrorBody(t, rec))
}

func TestMe_GenuineNotFoundStillReturns404(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	c, rec := newAuthTestContext(req)
	c.Set(middleware.ContextUserID, uint(999))

	h := NewAuthHandler(services.NewAuthService(testJWTSecret))
	h.Me(c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "user not found", decodeErrorBody(t, rec))
}

func TestRefresh_DBErrorReturns500NotUnauthorized(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	authSvc := services.NewAuthService(testJWTSecret)
	refreshToken, err := authSvc.IssueRefreshToken(1, "tonk", "user")
	require.NoError(t, err)
	closeDB(t, db)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	c, rec := newAuthTestContext(req)

	h := NewAuthHandler(authSvc)
	h.Refresh(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "internal error", decodeErrorBody(t, rec))
}

func TestLogin_DBErrorReturns500NotInvalidCredentials(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db
	closeDB(t, db)

	body, err := json.Marshal(map[string]string{"login": "tonk", "password": "whatever123"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c, rec := newAuthTestContext(req)

	h := NewAuthHandler(services.NewAuthService(testJWTSecret))
	h.Login(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "internal error", decodeErrorBody(t, rec))
}
