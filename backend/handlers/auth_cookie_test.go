package handlers

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newAuthTestContext(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

func accessTokenCookieSecure(t *testing.T, rec *httptest.ResponseRecorder) bool {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == "access_token" {
			return c.Secure
		}
	}
	t.Fatal("access_token cookie not found")
	return false
}

func TestCookieSecure(t *testing.T) {
	t.Run("plain HTTP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		c, _ := newAuthTestContext(req)
		assert.False(t, cookieSecure(c))
	})

	t.Run("TLS connection", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.TLS = &tls.ConnectionState{}
		c, _ := newAuthTestContext(req)
		assert.True(t, cookieSecure(c))
	})

	t.Run("X-Forwarded-Proto https", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		c, _ := newAuthTestContext(req)
		assert.True(t, cookieSecure(c))
	})

	t.Run("X-Forwarded-Proto http", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.Header.Set("X-Forwarded-Proto", "http")
		c, _ := newAuthTestContext(req)
		assert.False(t, cookieSecure(c))
	})
}

func TestSetAuthCookiesSecureFlag(t *testing.T) {
	tokens := &tokenResponse{AccessToken: "access", RefreshToken: "refresh"}

	t.Run("plain HTTP omits Secure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		c, rec := newAuthTestContext(req)
		setAuthCookies(c, tokens)
		assert.False(t, accessTokenCookieSecure(t, rec))
	})

	t.Run("reverse proxy HTTPS sets Secure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		c, rec := newAuthTestContext(req)
		setAuthCookies(c, tokens)
		assert.True(t, accessTokenCookieSecure(t, rec))
	})

	t.Run("MFA trust cookie follows same Secure rule", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/mfa/verify", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		c, rec := newAuthTestContext(req)
		setMFATrustCookie(c, "trust-token", 3600)
		var trust *http.Cookie
		for _, c := range rec.Result().Cookies() {
			if c.Name == "mfa_trust" {
				trust = c
				break
			}
		}
		require.NotNil(t, trust)
		assert.True(t, trust.Secure)
		assert.True(t, trust.HttpOnly)
	})
}
