package handlers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	wa "github.com/go-webauthn/webauthn/webauthn"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

type PasskeyHandler struct {
	authSvc *services.AuthService
}

func NewPasskeyHandler(authSvc *services.AuthService) *PasskeyHandler {
	return &PasskeyHandler{authSvc: authSvc}
}

// passkeyUser implements wa.User for the go-webauthn library.
type passkeyUser struct {
	user        models.User
	credentials []wa.Credential
}

func (u *passkeyUser) WebAuthnID() []byte {
	id := make([]byte, 8)
	binary.BigEndian.PutUint64(id, uint64(u.user.ID))
	return id
}

func (u *passkeyUser) WebAuthnName() string { return u.user.Email }

func (u *passkeyUser) WebAuthnDisplayName() string {
	if u.user.DisplayName != "" {
		return u.user.DisplayName
	}
	return u.user.Username
}

func (u *passkeyUser) WebAuthnCredentials() []wa.Credential { return u.credentials }

// webAuthnFromRequest creates a WebAuthn instance configured for the request's origin.
// RPID is derived from the Origin header (scheme+host stripped to hostname).
func webAuthnFromRequest(c *gin.Context) (*wa.WebAuthn, error) {
	origin := c.GetHeader("Origin")
	if origin == "" {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.Request.Host
		if host == "" {
			host = "localhost"
		}
		origin = scheme + "://" + host
	}
	u, err := url.Parse(origin)
	if err != nil {
		return nil, fmt.Errorf("invalid origin: %w", err)
	}
	return wa.New(&wa.Config{
		RPDisplayName: "WarmDesk",
		RPID:          u.Hostname(),
		RPOrigins:     []string{origin},
	})
}

// dbCredsToWebAuthn converts stored PasskeyCredentials to wa.Credential slice.
func dbCredsToWebAuthn(creds []models.PasskeyCredential) []wa.Credential {
	result := make([]wa.Credential, len(creds))
	for i, c := range creds {
		var transports []protocol.AuthenticatorTransport
		for _, t := range strings.Split(c.Transports, ",") {
			if t != "" {
				transports = append(transports, protocol.AuthenticatorTransport(t))
			}
		}
		result[i] = wa.Credential{
			ID:        c.CredentialID,
			PublicKey: c.PublicKey,
			Authenticator: wa.Authenticator{
				AAGUID:    c.AAGUID,
				SignCount: c.SignCount,
			},
			Transport: transports,
		}
	}
	return result
}

// PasskeyRegisterBegin handles GET /auth/passkey/register/begin (authenticated).
// Returns WebAuthn registration options and a short-lived challenge token.
func (h *PasskeyHandler) PasskeyRegisterBegin(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	var dbCreds []models.PasskeyCredential
	database.DB.Where("user_id = ?", userID).Find(&dbCreds)

	wAuth, err := webAuthnFromRequest(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webauthn setup failed"})
		return
	}

	pUser := &passkeyUser{user: user, credentials: dbCredsToWebAuthn(dbCreds)}
	options, session, err := wAuth.BeginRegistration(pUser,
		wa.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		wa.WithConveyancePreference(protocol.PreferNoAttestation),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin registration"})
		return
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	challengeToken, err := h.authSvc.IssuePasskeyChallenge(userID, "passkey_reg", string(sessionJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"options":         options,
		"challenge_token": challengeToken,
	})
}

// PasskeyRegisterFinish handles POST /auth/passkey/register/finish (authenticated).
// Verifies the attestation and stores the new credential.
func (h *PasskeyHandler) PasskeyRegisterFinish(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Name           string          `json:"name" binding:"required"`
		ChallengeToken string          `json:"challenge_token" binding:"required"`
		Credential     json.RawMessage `json:"credential" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	claims, err := h.authSvc.ValidatePasskeyChallenge(req.ChallengeToken, "passkey_reg")
	if err != nil || claims.UserID != userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired challenge"})
		return
	}

	var session wa.SessionData
	if err := json.Unmarshal([]byte(claims.PasskeySession), &session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session"})
		return
	}

	wAuth, err := webAuthnFromRequest(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webauthn setup failed"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}
	var dbCreds []models.PasskeyCredential
	database.DB.Where("user_id = ?", userID).Find(&dbCreds)
	pUser := &passkeyUser{user: user, credentials: dbCredsToWebAuthn(dbCreds)}

	parsedResp, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(req.Credential))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential response"})
		return
	}
	cred, err := wAuth.CreateCredential(pUser, session, parsedResp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "credential verification failed"})
		return
	}

	var transports []string
	for _, t := range cred.Transport {
		transports = append(transports, string(t))
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "Passkey"
	}
	if len(name) > 100 {
		name = name[:100]
	}

	now := time.Now()
	dbCred := models.PasskeyCredential{
		UserID:       userID,
		Name:         name,
		CredentialID: cred.ID,
		PublicKey:    cred.PublicKey,
		AAGUID:       cred.Authenticator.AAGUID,
		SignCount:     cred.Authenticator.SignCount,
		Transports:   strings.Join(transports, ","),
		LastUsedAt:   &now,
	}
	if err := database.DB.Create(&dbCred).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "Duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "this passkey is already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save passkey"})
		return
	}

	authLog(c, "passkey_registered", userID, user.Username, "")
	c.JSON(http.StatusCreated, dbCred)
}

// PasskeyLoginBegin handles POST /auth/passkey/login/begin (public).
// Returns discoverable-login options and a challenge token; no username required.
func (h *PasskeyHandler) PasskeyLoginBegin(c *gin.Context) {
	wAuth, err := webAuthnFromRequest(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webauthn setup failed"})
		return
	}

	options, session, err := wAuth.BeginDiscoverableLogin(
		wa.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin login"})
		return
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	// UserID 0: user identity is not known until the assertion is verified.
	challengeToken, err := h.authSvc.IssuePasskeyChallenge(0, "passkey_auth", string(sessionJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"options":         options,
		"challenge_token": challengeToken,
	})
}

// PasskeyLoginFinish handles POST /auth/passkey/login/finish (public).
// Verifies the assertion, issues full auth tokens on success.
func (h *PasskeyHandler) PasskeyLoginFinish(c *gin.Context) {
	var req struct {
		ChallengeToken string          `json:"challenge_token" binding:"required"`
		Credential     json.RawMessage `json:"credential" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	claims, err := h.authSvc.ValidatePasskeyChallenge(req.ChallengeToken, "passkey_auth")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired challenge"})
		return
	}

	var session wa.SessionData
	if err := json.Unmarshal([]byte(claims.PasskeySession), &session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session"})
		return
	}

	wAuth, err := webAuthnFromRequest(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webauthn setup failed"})
		return
	}

	parsedResp, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(req.Credential))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential response"})
		return
	}

	var authenticatedUser models.User
	var matchedCred *models.PasskeyCredential

	handler := func(rawID, userHandle []byte) (wa.User, error) {
		if len(userHandle) < 8 {
			return nil, fmt.Errorf("invalid user handle")
		}
		uid := binary.BigEndian.Uint64(userHandle)
		var u models.User
		if err := database.DB.First(&u, uid).Error; err != nil {
			return nil, fmt.Errorf("user not found")
		}
		if !u.IsActive {
			return nil, fmt.Errorf("account deactivated")
		}
		var dbCreds []models.PasskeyCredential
		database.DB.Where("user_id = ?", u.ID).Find(&dbCreds)
		for i := range dbCreds {
			if bytes.Equal(dbCreds[i].CredentialID, rawID) {
				matchedCred = &dbCreds[i]
				break
			}
		}
		authenticatedUser = u
		return &passkeyUser{user: u, credentials: dbCredsToWebAuthn(dbCreds)}, nil
	}

	cred, err := wAuth.ValidateDiscoverableLogin(handler, session, parsedResp)
	if err != nil {
		authLog(c, "passkey_login_failed", 0, "", "reason=verification_failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "passkey authentication failed"})
		return
	}

	now := time.Now()
	if matchedCred != nil {
		database.DB.Model(matchedCred).Updates(map[string]interface{}{
			"sign_count":   cred.Authenticator.SignCount,
			"last_used_at": now,
		})
	}
	database.DB.Model(&authenticatedUser).Update("last_login_at", now)

	ah := &AuthHandler{authSvc: h.authSvc}
	tokens, err := ah.issueTokens(authenticatedUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	setAuthCookies(c, tokens)
	authLog(c, "passkey_login_ok", authenticatedUser.ID, authenticatedUser.Username, "")
	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

// PasskeyList handles GET /auth/passkeys (authenticated).
func (h *PasskeyHandler) PasskeyList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var creds []models.PasskeyCredential
	database.DB.Where("user_id = ?", userID).Order("created_at asc").Find(&creds)
	c.JSON(http.StatusOK, creds)
}

// PasskeyDelete handles DELETE /auth/passkeys/:id (authenticated).
func (h *PasskeyHandler) PasskeyDelete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var cred models.PasskeyCredential
	if err := database.DB.Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&cred).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "passkey not found"})
		return
	}
	database.DB.Delete(&cred)
	authLog(c, "passkey_deleted", userID, "", "")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
