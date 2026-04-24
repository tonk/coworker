package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

type AuthHandler struct {
	authSvc *services.AuthService
}

func NewAuthHandler(authSvc *services.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type registerRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"` // email or username
	Password string `json:"password" binding:"required"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body registerRequest true "Registration details"
// @Success      201 {object} tokenResponse
// @Failure      400 {object} map[string]string
// @Failure      403 {object} map[string]string "Registration disabled"
// @Failure      409 {object} map[string]string "Email or username already exists"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	if !IsRegistrationEnabled() {
		authLog(c, "register_failed", 0, "", "reason=registration_disabled host="+c.Request.Host)
		c.JSON(http.StatusForbidden, gin.H{"error": "registration is disabled"})
		return
	}

	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		authLog(c, "register_failed", 0, "", "reason=bad_payload host="+c.Request.Host)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authLog(c, "register_attempt", 0, "", "email="+req.Email+" username="+req.Username+" host="+c.Request.Host)

	if msg := ValidatePasswordPolicy(req.Password); msg != "" {
		authLog(c, "register_failed", 0, "", "username="+req.Username+" reason=password_policy")
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	hash, err := h.authSvc.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}

	defs := GetGlobalDefaults()
	user := models.User{
		Email:          strings.ToLower(req.Email),
		Username:       req.Username,
		PasswordHash:   hash,
		DisplayName:    displayName,
		GlobalRole:     "user",
		Locale:         defs["locale"],
		IsActive:       true,
		DateTimeFormat: defs["date_time_format"],
		Timezone:       defs["timezone"],
		Theme:          defs["theme"],
		Font:           defs["font"],
		FontSize:       defs["font_size"],
	}

	// First user becomes admin
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		user.GlobalRole = "admin"
	}

	if err := database.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "Duplicate") {
			authLog(c, "register_failed", 0, "", "username="+req.Username+" reason=duplicate")
			c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
			return
		}
		authLog(c, "register_failed", 0, "", "username="+req.Username+" reason=internal_error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	tokens, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	setAuthCookies(c, tokens)
	authLog(c, "register_ok", user.ID, user.Username, "")
	c.JSON(http.StatusCreated, tokens)
}

// Login godoc
// @Summary      Login with email/username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body loginRequest true "Login credentials"
// @Success      200 {object} tokenResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		authLog(c, "login_failed", 0, "", "reason=bad_payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authLog(c, "login_attempt", 0, "", "login="+req.Login+" host="+c.Request.Host)

	var user models.User
	login := strings.ToLower(req.Login)
	if err := database.DB.Where("email = ? OR username = ?", login, req.Login).First(&user).Error; err != nil {
		authLog(c, "login_failed", 0, "", "login="+req.Login+" reason=unknown_user")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !user.IsActive {
		authLog(c, "login_failed", user.ID, user.Username, "reason=account_deactivated")
		c.JSON(http.StatusForbidden, gin.H{"error": "account deactivated"})
		return
	}

	if !h.authSvc.CheckPassword(user.PasswordHash, req.Password) {
		authLog(c, "login_failed", user.ID, user.Username, "reason=wrong_password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	now := time.Now()
	database.DB.Model(&user).Update("last_login_at", now)

	// If the user has TOTP enabled, issue an MFA challenge token instead of full tokens.
	if user.TOTPEnabled {
		mfaToken, err := h.authSvc.IssueMFAToken(user.ID, user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		authLog(c, "login_mfa_challenge", user.ID, user.Username, "")
		c.JSON(http.StatusOK, gin.H{"mfa_required": true, "mfa_token": mfaToken})
		return
	}

	tokens, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	setAuthCookies(c, tokens)
	authLog(c, "login_ok", user.ID, user.Username, "")
	resp := gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	}
	if IsMFARequired() {
		resp["mfa_setup_required"] = true
	}
	c.JSON(http.StatusOK, resp)
}

// Refresh godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body map[string]string true "Refresh token"
// @Success      200 {object} tokenResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	// Body is optional — browser clients send no body and rely on the httpOnly cookie.
	_ = c.ShouldBindJSON(&req)
	if req.RefreshToken == "" {
		req.RefreshToken, _ = c.Cookie("refresh_token")
	}
	if req.RefreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	claims, err := h.authSvc.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	tokens, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	setAuthCookies(c, tokens)
	c.JSON(http.StatusOK, tokens)
}

// Me godoc
// @Summary      Get current user profile
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.User
// @Failure      404 {object} map[string]string
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.CanViewReports = userCanViewReports(userID, user.GlobalRole)
	c.JSON(http.StatusOK, user)
}

// userCanViewReports returns true if the user is a global admin or is an
// admin/owner of at least one project.
func userCanViewReports(userID uint, globalRole string) bool {
	if globalRole == "admin" {
		return true
	}
	var count int64
	database.DB.Model(&models.ProjectMember{}).
		Where("user_id = ? AND role IN ?", userID, []string{"admin", "owner"}).
		Count(&count)
	return count > 0
}

// UpdateMe godoc
// @Summary      Update current user profile and preferences
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body map[string]interface{} true "Profile fields to update"
// @Success      200 {object} models.User
// @Failure      400 {object} map[string]string
// @Router       /auth/me [put]
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		FirstName          string  `json:"first_name"`
		LastName           string  `json:"last_name"`
		DisplayName        string  `json:"display_name"`
		Email              string  `json:"email"`
		AvatarURL          *string `json:"avatar_url"`
		Locale             string  `json:"locale"`
		Theme              string  `json:"theme"`
		DateTimeFormat     string  `json:"date_time_format"`
		Timezone           string  `json:"timezone"`
		Font               string  `json:"font"`
		FontSize           string  `json:"font_size"`
		SidebarPosition    string  `json:"sidebar_position"`
		ShowBreadcrumbs    *bool   `json:"show_breadcrumbs"`
		AccentColor        string  `json:"accent_color"`
		EmailNotifications  *bool   `json:"email_notifications"`
		TimeTrackingEnabled *bool   `json:"time_tracking_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Email != "" {
		updates["email"] = strings.ToLower(req.Email)
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL
	}
	validLocales := map[string]bool{"en": true, "nl": true, "de": true, "fr": true, "es": true}
	if validLocales[req.Locale] {
		updates["locale"] = req.Locale
	}
	if req.Theme == "light" || req.Theme == "dark" || req.Theme == "system" {
		updates["theme"] = req.Theme
	}
	if req.DateTimeFormat != "" {
		updates["date_time_format"] = req.DateTimeFormat
	}
	if req.Timezone != "" {
		updates["timezone"] = req.Timezone
	}
	if req.Font != "" {
		updates["font"] = req.Font
	}
	if req.FontSize != "" {
		updates["font_size"] = req.FontSize
	}
	if req.SidebarPosition == "left" || req.SidebarPosition == "right" {
		updates["sidebar_position"] = req.SidebarPosition
	}
	if req.ShowBreadcrumbs != nil {
		updates["show_breadcrumbs"] = *req.ShowBreadcrumbs
	}
	validAccents := map[string]bool{"blue": true, "red": true, "green": true, "orange": true}
	if validAccents[req.AccentColor] {
		updates["accent_color"] = req.AccentColor
	}
	if req.EmailNotifications != nil {
		updates["email_notifications"] = *req.EmailNotifications
	}
	if req.TimeTrackingEnabled != nil {
		updates["time_tracking_enabled"] = *req.TimeTrackingEnabled
	}

	now := time.Now()
	updates["settings_updated_at"] = now

	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var user models.User
	database.DB.First(&user, userID)
	c.JSON(http.StatusOK, user)
}

// ChangePassword godoc
// @Summary      Change current user password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body map[string]string true "Current and new password"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Router       /auth/me/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if !h.authSvc.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		authLog(c, "password_change_failed", user.ID, user.Username, "reason=wrong_current_password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect current password"})
		return
	}

	if msg := ValidatePasswordPolicy(req.NewPassword); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	hash, err := h.authSvc.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	database.DB.Model(&user).Update("password_hash", hash)
	authLog(c, "password_changed", user.ID, user.Username, "")
	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

// MFAVerify handles POST /auth/mfa/verify.
// Accepts the short-lived mfa_token from the login challenge and a TOTP code,
// and — if valid — returns a full access+refresh token pair.
func (h *AuthHandler) MFAVerify(c *gin.Context) {
	var req struct {
		MFAToken string `json:"mfa_token" binding:"required"`
		Code     string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.authSvc.ValidateToken(req.MFAToken)
	if err != nil || !claims.MFAPending {
		// Use 400 so the axios 401 interceptor does not fire and redirect the user.
		// The frontend shows an "expired session" message and resets to step 1.
		authLog(c, "mfa_verify_failed", 0, "", "reason=session_expired")
		c.JSON(http.StatusBadRequest, gin.H{"error": "mfa_session_expired"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		authLog(c, "mfa_verify_failed", claims.UserID, claims.Username, "reason=user_not_found")
		c.JSON(http.StatusBadRequest, gin.H{"error": "mfa_session_expired"})
		return
	}

	if !user.TOTPEnabled || !h.authSvc.VerifyTOTP(user.TOTPSecret, req.Code) {
		authLog(c, "mfa_verify_failed", user.ID, user.Username, "reason=invalid_code")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_code"})
		return
	}

	tokens, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	setAuthCookies(c, tokens)
	authLog(c, "mfa_verify_ok", user.ID, user.Username, "")
	c.JSON(http.StatusOK, tokens)
}

// MFASetup handles GET /auth/mfa/setup.
// Generates a fresh TOTP secret for the current user and stores it (not yet enabled).
// Returns the base32 secret and the otpauth:// URI for QR code generation.
func (h *AuthHandler) MFASetup(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	issuer := GetCompanyName()
	if issuer == "" {
		issuer = "WarmDesk"
	}

	secret, uri, err := h.authSvc.GenerateTOTPSecret(user.Username, issuer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Store the secret but keep TOTP disabled until the user verifies with MFAEnable.
	database.DB.Model(&user).Update("totp_secret", secret)

	c.JSON(http.StatusOK, gin.H{"secret": secret, "uri": uri})
}

// MFAEnable handles POST /auth/mfa/enable.
// Verifies the provided TOTP code against the stored secret and enables TOTP.
func (h *AuthHandler) MFAEnable(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.TOTPSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no mfa setup in progress"})
		return
	}

	if !h.authSvc.VerifyTOTP(user.TOTPSecret, req.Code) {
		authLog(c, "mfa_enable_failed", user.ID, user.Username, "reason=invalid_code")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid code"})
		return
	}

	database.DB.Model(&user).Update("totp_enabled", true)
	authLog(c, "mfa_enabled", user.ID, user.Username, "")
	c.JSON(http.StatusOK, gin.H{"message": "mfa enabled"})
}

// MFADisable handles POST /auth/mfa/disable.
// Requires the user's current password; clears the TOTP secret and disables TOTP.
func (h *AuthHandler) MFADisable(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if !h.authSvc.CheckPassword(user.PasswordHash, req.Password) {
		authLog(c, "mfa_disable_failed", user.ID, user.Username, "reason=wrong_password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect password"})
		return
	}

	database.DB.Model(&user).Updates(map[string]interface{}{
		"totp_enabled": false,
		"totp_secret":  "",
	})
	authLog(c, "mfa_disabled", user.ID, user.Username, "")
	c.JSON(http.StatusOK, gin.H{"message": "mfa disabled"})
}

// ForgotPassword handles POST /auth/forgot-password.
// Looks up the user by email; if found, active, and with a non-empty email
// address it sends a one-time reset link. Always responds 200 so callers
// cannot enumerate accounts.
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// Still 200 — do not reveal validation details.
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	ip := c.ClientIP()
	client := clientStr(c)
	authLogRaw(ip, client, "password_reset_requested", 0, "", "email="+req.Email)
	go func() {
		var user models.User
		if err := database.DB.Where("email = ?", strings.ToLower(req.Email)).First(&user).Error; err != nil {
			return // unknown email — silent
		}
		if !user.IsActive || user.Email == "" {
			return // inactive or no email — silent
		}

		// Generate a 32-byte random token, hex-encoded (64 chars).
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return
		}
		token := hex.EncodeToString(raw)
		expiry := time.Now().Add(time.Hour)

		database.DB.Model(&user).Updates(map[string]any{
			"password_reset_token":  token,
			"password_reset_expiry": expiry,
		})

		resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL(c), token)
		subject := "Reset your WarmDesk password"
		plainBody := fmt.Sprintf(
			"Hi %s,\n\nClick the link below to reset your WarmDesk password.\n"+
				"The link is valid for one hour.\n\n%s\n\n"+
				"If you did not request a password reset, ignore this email.",
			user.DisplayName, resetURL,
		)
		htmlContent := fmt.Sprintf(
			`<tr><td style="padding:28px 32px;font-size:15px;color:#333;line-height:1.6">`+
				`<p style="margin:0 0 8px">Hi <strong>%s</strong>,</p>`+
				`<p style="margin:0 0 20px;color:#555">Click the button below to reset your WarmDesk password. The link is valid for one hour.</p>`+
				`<p style="margin:0 0 20px;text-align:center">`+
				`<a href="%s" class="wd-btn" style="display:inline-block;padding:12px 28px;background:#1a5fb4;color:#ffffff;text-decoration:none;border-radius:6px;font-size:15px;font-weight:bold;border:2px solid #1a5fb4">Reset password</a>`+
				`</p>`+
				`<p style="margin:0;font-size:13px;color:#999">If you did not request a password reset, you can safely ignore this email.</p>`+
				`</td></tr>`,
			user.DisplayName, resetURL,
		)

		emailSvc := services.GetEmailService()
		if emailSvc != nil {
			_ = emailSvc.SendHTML(user.Email, subject,
				services.WrapHTML("Password Reset", htmlContent),
				services.WrapText("Password Reset", plainBody))
			authLogRaw(ip, client, "password_reset_email_sent", user.ID, user.Username, "")
		}
	}()

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ResetPassword handles POST /auth/reset-password.
// Validates the token and sets the new password. Consumes the token on success.
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("password_reset_token = ?", req.Token).First(&user).Error; err != nil {
		authLog(c, "password_reset_failed", 0, "", "reason=invalid_token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired reset link"})
		return
	}

	if user.PasswordResetExpiry == nil || time.Now().After(*user.PasswordResetExpiry) {
		authLog(c, "password_reset_failed", user.ID, user.Username, "reason=expired_token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired reset link"})
		return
	}

	if msg := ValidatePasswordPolicy(req.Password); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	hash, err := h.authSvc.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	database.DB.Model(&user).Updates(map[string]any{
		"password_hash":         hash,
		"password_reset_token":  "",
		"password_reset_expiry": nil,
	})

	authLog(c, "password_reset_ok", user.ID, user.Username, "")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// authLog writes a structured audit log line for authentication events.
// userID/username are omitted when they cannot be determined (e.g. unknown user).
func authLog(c *gin.Context, event string, userID uint, username, detail string) {
	authLogRaw(c.ClientIP(), clientStr(c), event, userID, username, detail)
}

// authLogRaw is used when the gin.Context is no longer available (e.g. inside a goroutine).
func authLogRaw(ip, client, event string, userID uint, username, detail string) {
	extra := ""
	if detail != "" {
		extra = " " + detail
	}
	if userID != 0 {
		log.Printf("auth: %s user=%d(%s)%s ip=%s via=%s", event, userID, username, extra, ip, client)
	} else {
		log.Printf("auth: %s%s ip=%s via=%s", event, extra, ip, client)
	}
}

// clientStr returns the value of X-WarmDesk-Client if present, otherwise "web".
func clientStr(c *gin.Context) string {
	if v := c.GetHeader("X-WarmDesk-Client"); v != "" {
		return v
	}
	return "web"
}

func (h *AuthHandler) issueTokens(user models.User) (*tokenResponse, error) {
	access, err := h.authSvc.IssueAccessToken(user.ID, user.Username, user.GlobalRole)
	if err != nil {
		return nil, err
	}
	refresh, err := h.authSvc.IssueRefreshToken(user.ID, user.Username, user.GlobalRole)
	if err != nil {
		return nil, err
	}
	return &tokenResponse{AccessToken: access, RefreshToken: refresh}, nil
}

// setAuthCookies writes access and refresh tokens as httpOnly, SameSite=Strict cookies.
// Browser clients use these automatically; API/Tauri clients continue to use the JSON body.
func setAuthCookies(c *gin.Context, tokens *tokenResponse) {
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("access_token", tokens.AccessToken, 15*60, "/", "", secure, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, 7*24*60*60, "/", "", secure, true)
}

// clearAuthCookies expires the auth cookies immediately.
func clearAuthCookies(c *gin.Context) {
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("access_token", "", -1, "/", "", secure, true)
	c.SetCookie("refresh_token", "", -1, "/", "", secure, true)
}

// Logout handles POST /auth/logout.
// Clears the httpOnly auth cookies so browser sessions are terminated cleanly.
func (h *AuthHandler) Logout(c *gin.Context) {
	clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
