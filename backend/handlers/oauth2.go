package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/config"
	"golang.org/x/oauth2"
)

var oauth2Cfg *config.OAuth2Config

// SetOAuth2Config stores the app-level OAuth2 client credentials.
func SetOAuth2Config(cfg *config.OAuth2Config) {
	oauth2Cfg = cfg
}

// oauthStateStore holds temporary CSRF states for the OAuth2 authorization flow.
// Keys are random state strings, values are the provider name.
var oauthStateStore = struct {
	mu     sync.Mutex
	states map[string]string
}{states: make(map[string]string)}

func generateOAuthState() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}

func oauthProviderConfig(provider string) (oauth2.Config, string, error) {
	if oauth2Cfg == nil {
		return oauth2.Config{}, "", fmt.Errorf("OAuth2 not configured (no client credentials)")
	}
	switch provider {
	case "google":
		if oauth2Cfg.GoogleClientID == "" || oauth2Cfg.GoogleClientSecret == "" {
			return oauth2.Config{}, "", fmt.Errorf("Google OAuth2 client credentials not configured")
		}
		return oauth2.Config{
			ClientID:     oauth2Cfg.GoogleClientID,
			ClientSecret: oauth2Cfg.GoogleClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
			Scopes: []string{"https://mail.google.com/"},
		}, "google", nil
	case "office365":
		if oauth2Cfg.OfficeClientID == "" || oauth2Cfg.OfficeClientSecret == "" {
			return oauth2.Config{}, "", fmt.Errorf("Office 365 OAuth2 client credentials not configured")
		}
		return oauth2.Config{
			ClientID:     oauth2Cfg.OfficeClientID,
			ClientSecret: oauth2Cfg.OfficeClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
				TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			},
			Scopes: []string{"https://outlook.office.com/IMAP.AccessAsUser.All", "offline_access"},
		}, "office365", nil
	default:
		return oauth2.Config{}, "", fmt.Errorf("unsupported OAuth2 provider: %s", provider)
	}
}

// AdminIMAPOAuth2AuthURL returns the OAuth2 authorization URL for the given provider.
// GET /api/v1/admin/imap/oauth2/auth-url?provider=google|office365
func AdminIMAPOAuth2AuthURL(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}
	oauthConf, _, err := oauthProviderConfig(provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	state := generateOAuthState()
	oauthStateStore.mu.Lock()
	oauthStateStore.states[state] = provider
	oauthStateStore.mu.Unlock()

	// Clean up old states after 10 minutes
	time.AfterFunc(10*time.Minute, func() {
		oauthStateStore.mu.Lock()
		delete(oauthStateStore.states, state)
		oauthStateStore.mu.Unlock()
	})

	redirectURL := configuredBaseURL
	if redirectURL == "" {
		redirectURL = baseURL(c)
	}
	oauthConf.RedirectURL = redirectURL + "/api/v1/admin/imap/oauth2/callback"

	url := oauthConf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// AdminIMAPOAuth2Callback handles the OAuth2 callback from Google / Office 365.
// GET /api/v1/admin/imap/oauth2/callback?code=...&state=...
func AdminIMAPOAuth2Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code and state are required"})
		return
	}

	// Verify state
	oauthStateStore.mu.Lock()
	provider, ok := oauthStateStore.states[state]
	delete(oauthStateStore.states, state)
	oauthStateStore.mu.Unlock()
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	oauthConf, _, err := oauthProviderConfig(provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	redirectURL := configuredBaseURL
	if redirectURL == "" {
		redirectURL = baseURL(c)
	}
	oauthConf.RedirectURL = redirectURL + "/api/v1/admin/imap/oauth2/callback"

	tok, err := oauthConf.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("oauth2: token exchange failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed"})
		return
	}

	expiry := ""
	if !tok.Expiry.IsZero() {
		expiry = tok.Expiry.Format(time.RFC3339)
	}

	saveSetting(settingIMAPOAuth2Provider, provider)
	saveSetting(settingIMAPAuthMechanism, "oauth2")
	saveSetting(settingIMAPAccessToken, tok.AccessToken)
	if tok.RefreshToken != "" {
		saveSetting(settingIMAPRefreshToken, tok.RefreshToken)
	}
	if expiry != "" {
		saveSetting(settingIMAPTokenExpiry, expiry)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `<html><body><script>window.close()</script><p>Authorization successful. You may close this window.</p></body></html>`)
}

// AdminIMAPOAuth2Status returns whether OAuth2 tokens are configured for IMAP.
// GET /api/v1/admin/imap/oauth2/status
func AdminIMAPOAuth2Status(c *gin.Context) {
	all := loadAllSettings()
	connected := all[settingIMAPAuthMechanism] == "oauth2" && all[settingIMAPAccessToken] != ""
	c.JSON(http.StatusOK, gin.H{
		"connected":       connected,
		"provider":        all[settingIMAPOAuth2Provider],
		"auth_mechanism":  all[settingIMAPAuthMechanism],
	})
}

// RefreshIMAPOAuth2Token checks the stored IMAP OAuth2 token expiry and
// refreshes it if needed by exchanging the refresh token for a new access token.
// Returns true if the token was refreshed.
func RefreshIMAPOAuth2Token() bool {
	all := loadAllSettings()
	if all[settingIMAPAuthMechanism] != "oauth2" {
		return false
	}
	provider := all[settingIMAPOAuth2Provider]
	if provider == "" {
		return false
	}
	expiryStr := all[settingIMAPTokenExpiry]
	if expiryStr == "" {
		return false
	}
	expiry, err := time.Parse(time.RFC3339, expiryStr)
	if err != nil {
		return false
	}
	if time.Now().Before(expiry.Add(-5 * time.Minute)) {
		return false // still valid
	}

	refreshToken := all[settingIMAPRefreshToken]
	if refreshToken == "" {
		log.Println("oauth2: cannot refresh token — no refresh token stored")
		return false
	}

	oauthConf, _, err := oauthProviderConfig(provider)
	if err != nil {
		log.Printf("oauth2: cannot refresh token: %v", err)
		return false
	}
	tok, err := oauthConf.TokenSource(context.Background(), &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		log.Printf("oauth2: token refresh failed: %v", err)
		return false
	}

	newExpiry := ""
	if !tok.Expiry.IsZero() {
		newExpiry = tok.Expiry.Format(time.RFC3339)
	}
	saveSetting(settingIMAPAccessToken, tok.AccessToken)
	if tok.RefreshToken != "" {
		saveSetting(settingIMAPRefreshToken, tok.RefreshToken)
	}
	if newExpiry != "" {
		saveSetting(settingIMAPTokenExpiry, newExpiry)
	}
	log.Println("oauth2: IMAP access token refreshed")
	return true
}

// AdminIMAPOAuth2Disconnect clears the OAuth2 tokens.
// POST /api/v1/admin/imap/oauth2/disconnect
func AdminIMAPOAuth2Disconnect(c *gin.Context) {
	saveSetting(settingIMAPAuthMechanism, "plain")
	saveSetting(settingIMAPOAuth2Provider, "")
	saveSetting(settingIMAPAccessToken, "")
	saveSetting(settingIMAPRefreshToken, "")
	saveSetting(settingIMAPTokenExpiry, "")
	c.JSON(http.StatusOK, gin.H{"message": "Disconnected"})
}
