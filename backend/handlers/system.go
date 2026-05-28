package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

const (
	settingRegistrationEnabled       = "registration_enabled"
	settingDefaultDateTimeFormat     = "default_date_time_format"
	settingDefaultTimezone           = "default_timezone"
	settingDefaultTheme              = "default_theme"
	settingDefaultFont               = "default_font"
	settingDefaultFontSize           = "default_font_size"
	settingDefaultLocale             = "default_locale"
	settingSMTPHost                  = "smtp_host"
	settingSMTPPort                  = "smtp_port"
	settingSMTPFrom                  = "smtp_from"
	settingSMTPUsername              = "smtp_username"
	settingSMTPPassword              = "smtp_password"
	settingSessionTimeoutMinutes     = "session_timeout_minutes"
	settingCompanyName               = "company_name"
	settingCompanyLogo               = "company_logo"
	settingDefaultColumns            = "default_columns"
	settingDefaultLabels             = "default_labels"
	settingMFARequired               = "mfa_required"
	settingPasswordMinLength         = "password_min_length"
	settingPasswordRequireUpper      = "password_require_upper"
	settingPasswordRequireLower      = "password_require_lower"
	settingPasswordRequireDigit      = "password_require_digit"
	settingPasswordRequireSpecial    = "password_require_special"
	settingBackupSchedule            = "backup_schedule"
	settingBackupStartTime           = "backup_start_time"
	settingBackupLastRun             = "backup_last_run"
	settingBackupKeep                = "backup_keep"
	settingBackupEmailEnabled        = "backup_email_enabled"
	settingBackupEmailAddress        = "backup_email_address"
	settingBackupLastSuccess         = "backup_last_success"
	settingScrumStorypointsEnabled   = "scrum_storypoints_enabled"
	settingGravatarEnabled           = "gravatar_enabled"
	settingExternalImageProxyEnabled = "external_image_proxy_enabled"
	settingLoginBrandingEnabled      = "login_branding_enabled"
	settingCompanyLogoDark           = "company_logo_dark"
	settingAllowedIPs                = "allowed_ips"
	settingPasswordChangePeriodDays  = "password_change_period_days"
	settingIMAPEnabled               = "imap_enabled"
	settingIMAPHost                  = "imap_host"
	settingIMAPPort                  = "imap_port"
	settingIMAPUsername              = "imap_username"
	settingIMAPPassword              = "imap_password"
	settingIMAPUseTLS                = "imap_use_tls"
	settingIMAPMailbox               = "imap_mailbox"
	settingIMAPProcessedMailbox       = "imap_processed_mailbox"
	settingIMAPPollInterval          = "imap_poll_interval"
	settingIMAPAuthMechanism         = "imap_auth_mechanism"
	settingIMAPOAuth2Provider        = "imap_oauth2_provider"
	settingIMAPAccessToken           = "imap_access_token"
	settingIMAPRefreshToken          = "imap_refresh_token"
	settingIMAPTokenExpiry           = "imap_token_expiry"
)

func init() {
	models.GravatarEnabledFn = IsGravatarEnabled
}

// IsGravatarEnabled returns true when the admin has enabled Gravatar avatars.
func IsGravatarEnabled() bool {
	return loadAllSettings()[settingGravatarEnabled] != "false"
}

// IsExternalImageProxyEnabled returns true when the media proxy should be used
// for external avatar/image URLs.
func IsExternalImageProxyEnabled() bool {
	return loadAllSettings()[settingExternalImageProxyEnabled] != "false"
}

// GetPasswordChangePeriodDays returns the configured password change period in
// days. 0 means the policy is disabled.
func GetPasswordChangePeriodDays() int {
	v := loadAllSettings()[settingPasswordChangePeriodDays]
	n, _ := strconv.Atoi(v)
	if n < 0 {
		return 0
	}
	return n
}

// configuredBaseURL stores the value of base_url from the config file so
// auth handlers can build absolute URLs (e.g. password-reset links) without
// requiring a DB setting.
var configuredBaseURL string

// SetBaseURL is called from main.go after loading the config.
func SetBaseURL(u string) { configuredBaseURL = u }

// baseURL returns an absolute base URL suitable for building links in emails.
// It prefers the configured base_url; if absent it derives one from the request.
func baseURL(c *gin.Context) string {
	if configuredBaseURL != "" {
		return strings.TrimRight(configuredBaseURL, "/")
	}
	scheme := "http"
	if c.Request.Header.Get("X-Forwarded-Proto") == "https" || c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

var systemSettingDefaults = map[string]string{
	settingRegistrationEnabled:       "true",
	settingDefaultDateTimeFormat:     "YYYY-MM-DD HH:mm",
	settingDefaultTimezone:           "UTC",
	settingDefaultTheme:              "system",
	settingDefaultFont:               "system",
	settingDefaultFontSize:           "14",
	settingDefaultLocale:             "en",
	settingSMTPHost:                  "",
	settingSMTPPort:                  "587",
	settingSMTPFrom:                  "",
	settingSMTPUsername:              "",
	settingSMTPPassword:              "",
	settingSessionTimeoutMinutes:     "60",
	settingCompanyName:               "",
	settingCompanyLogo:               "",
	settingCompanyLogoDark:           "",
	settingDefaultColumns:            "Backlog",
	settingDefaultLabels:             "Bug\nFeature\nDesign\nContent",
	settingMFARequired:               "false",
	settingPasswordMinLength:         "12",
	settingPasswordRequireUpper:      "false",
	settingPasswordRequireLower:      "true",
	settingPasswordRequireDigit:      "true",
	settingPasswordRequireSpecial:    "false",
	settingBackupSchedule:            "disabled",
	settingBackupStartTime:           "",
	settingBackupLastRun:             "",
	settingBackupKeep:                "10",
	settingBackupEmailEnabled:        "false",
	settingBackupEmailAddress:        "",
	settingBackupLastSuccess:         "",
	settingScrumStorypointsEnabled:   "false",
	settingGravatarEnabled:           "true",
	settingExternalImageProxyEnabled: "true",
	settingPasswordChangePeriodDays:  "0",
	settingIMAPEnabled:               "false",
	settingIMAPHost:                  "",
	settingIMAPPort:                  "993",
	settingIMAPUsername:              "",
	settingIMAPPassword:              "",
	settingIMAPUseTLS:                "true",
	settingIMAPMailbox:               "INBOX",
	settingIMAPProcessedMailbox:       "Processed",
	settingIMAPPollInterval:          "60",
	settingIMAPAuthMechanism:         "plain",
	settingIMAPOAuth2Provider:        "",
	settingIMAPAccessToken:           "",
	settingIMAPRefreshToken:          "",
	settingIMAPTokenExpiry:           "",
}

// InitSystemDefaults seeds the in-memory defaults from the config file so that
// settings not yet stored in the database reflect the operator's preferences.
func InitSystemDefaults(cfg *config.Config) {
	if cfg.DefaultLocale != "" {
		systemSettingDefaults[settingDefaultLocale] = cfg.DefaultLocale
	}
	if cfg.SMTP.Host != "" {
		systemSettingDefaults[settingSMTPHost] = cfg.SMTP.Host
	}
	if cfg.SMTP.Port != 0 {
		systemSettingDefaults[settingSMTPPort] = fmt.Sprintf("%d", cfg.SMTP.Port)
	}
	if cfg.SMTP.From != "" {
		systemSettingDefaults[settingSMTPFrom] = cfg.SMTP.From
	}
	if cfg.SMTP.Username != "" {
		systemSettingDefaults[settingSMTPUsername] = cfg.SMTP.Username
	}
	if cfg.SMTP.Password != "" {
		systemSettingDefaults[settingSMTPPassword] = cfg.SMTP.Password
	}
}

// GetEmailBranding returns version, company name, company logo URL, and instance URL
// for use in outbound emails. Registered as the appInfoReader in services at startup.
func GetEmailBranding() (version, companyName, logoURL, instanceURL string) {
	all := loadAllSettings()
	name := all[settingCompanyName]
	if name == "" {
		name = "WarmDesk"
	}
	logo := logoAsDataURI(all[settingCompanyLogo])
	return serverVersion, name, logo, configuredBaseURL
}

// logoAsDataURI fetches the company logo and returns it as a base64 data URI so
// it is embedded directly in outbound emails (no external URL required).
// Uploaded files are read from disk; external URLs are fetched via HTTP.
func logoAsDataURI(raw string) string {
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "data:") {
		return raw
	}
	var data []byte
	var mime string
	if strings.HasPrefix(raw, "/uploads/") {
		uploadDir := "./uploads"
		if attachmentCfg != nil && attachmentCfg.UploadDir != "" {
			uploadDir = attachmentCfg.UploadDir
		}
		storedName := strings.TrimPrefix(raw, "/uploads/")
		var err error
		data, err = os.ReadFile(filepath.Join(uploadDir, storedName))
		if err != nil {
			return ""
		}
		mime = mimeFromExt(filepath.Ext(storedName))
	} else if strings.HasPrefix(raw, "https://") {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
		if err != nil {
			return raw
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			return raw // fall back to URL so the img src still has something
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return raw
		}
		mime = resp.Header.Get("Content-Type")
		if i := strings.Index(mime, ";"); i != -1 {
			mime = strings.TrimSpace(mime[:i])
		}
		if mime == "" {
			mime = "image/png"
		}
	} else {
		return ""
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}

func mimeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

// GetSMTPSettings returns the current SMTP configuration from the database.
// Used by the email service so changes take effect without a restart.
func GetSMTPSettings() config.SMTPConfig {
	all := loadAllSettings()
	port, _ := strconv.Atoi(all[settingSMTPPort])
	if port == 0 {
		port = 587
	}
	return config.SMTPConfig{
		Host:     all[settingSMTPHost],
		Port:     port,
		From:     all[settingSMTPFrom],
		Username: all[settingSMTPUsername],
		Password: all[settingSMTPPassword],
	}
}

// GetIMAPSettings returns the current IMAP configuration from the database.
func GetIMAPSettings() config.IMAPConfig {
	all := loadAllSettings()
	port, _ := strconv.Atoi(all[settingIMAPPort])
	if port == 0 {
		port = 993
	}
	pollInterval, _ := strconv.Atoi(all[settingIMAPPollInterval])
	if pollInterval == 0 {
		pollInterval = 60
	}
	return config.IMAPConfig{
		Enabled:          all[settingIMAPEnabled] == "true",
		Host:             all[settingIMAPHost],
		Port:             port,
		Username:         all[settingIMAPUsername],
		Password:         all[settingIMAPPassword],
		UseTLS:           all[settingIMAPUseTLS] != "false",
		Mailbox:          all[settingIMAPMailbox],
		ProcessedMailbox: all[settingIMAPProcessedMailbox],
		PollInterval:     pollInterval,
		AuthMechanism:    all[settingIMAPAuthMechanism],
		OAuth2Provider:   all[settingIMAPOAuth2Provider],
		AccessToken:      all[settingIMAPAccessToken],
		RefreshToken:     all[settingIMAPRefreshToken],
		TokenExpiry:      all[settingIMAPTokenExpiry],
	}
}

// GetSystemSettings godoc
// @Summary      Get public system settings (registration enabled, locale defaults)
// @Tags         system
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /system/settings [get]
func GetSystemSettings(c *gin.Context) {
	all := loadAllSettings()
	timeoutMinutes, _ := strconv.Atoi(all[settingSessionTimeoutMinutes])
	c.JSON(http.StatusOK, gin.H{
		"registration_enabled":         all[settingRegistrationEnabled] != "false",
		"default_date_time_format":     all[settingDefaultDateTimeFormat],
		"default_timezone":             all[settingDefaultTimezone],
		"default_theme":                all[settingDefaultTheme],
		"default_font":                 all[settingDefaultFont],
		"default_font_size":            all[settingDefaultFontSize],
		"default_locale":               all[settingDefaultLocale],
		"session_timeout_minutes":      timeoutMinutes,
		"company_name":                 all[settingCompanyName],
		"company_logo":                 all[settingCompanyLogo],
		"company_logo_dark":            all[settingCompanyLogoDark],
		"login_branding_enabled":       all[settingLoginBrandingEnabled] == "true",
		"mfa_required":                 all[settingMFARequired] == "true",
		"password_policy":              GetPasswordPolicy(),
		"scrum_storypoints_enabled":    all[settingScrumStorypointsEnabled] == "true",
		"gravatar_enabled":             all[settingGravatarEnabled] != "false",
		"external_image_proxy_enabled": all[settingExternalImageProxyEnabled] != "false",
	})
}

// AdminGetSystemSettings returns all system settings for admins.
// The SMTP password is never sent back — only smtp_password_set (bool) is included.
func AdminGetSystemSettings(c *gin.Context) {
	all := loadAllSettings()
	// Mask the password: send only whether one is configured
	passwordSet := all[settingSMTPPassword] != ""
	delete(all, settingSMTPPassword)
	all["smtp_password_set"] = fmt.Sprintf("%t", passwordSet)
	imapPasswordSet := all[settingIMAPPassword] != ""
	delete(all, settingIMAPPassword)
	all["imap_password_set"] = fmt.Sprintf("%t", imapPasswordSet)
	imapAccessTokenSet := all[settingIMAPAccessToken] != ""
	delete(all, settingIMAPAccessToken)
	all["imap_access_token_set"] = fmt.Sprintf("%t", imapAccessTokenSet)
	imapRefreshTokenSet := all[settingIMAPRefreshToken] != ""
	delete(all, settingIMAPRefreshToken)
	all["imap_refresh_token_set"] = fmt.Sprintf("%t", imapRefreshTokenSet)
	c.JSON(http.StatusOK, all)
}

// AdminUpdateSystemSettings updates system settings.
func AdminUpdateSystemSettings(c *gin.Context) {
	var req struct {
		MFARequired               *bool       `json:"mfa_required"`
		RegistrationEnabled       *bool       `json:"registration_enabled"`
		DefaultDateTimeFormat     string      `json:"default_date_time_format"`
		DefaultTimezone           string      `json:"default_timezone"`
		DefaultTheme              string      `json:"default_theme"`
		DefaultFont               string      `json:"default_font"`
		DefaultFontSize           string      `json:"default_font_size"`
		DefaultLocale             string      `json:"default_locale"`
		PasswordMinLength         *int        `json:"password_min_length"`
		PasswordRequireUpper      *bool       `json:"password_require_upper"`
		PasswordRequireLower      *bool       `json:"password_require_lower"`
		PasswordRequireDigit      *bool       `json:"password_require_digit"`
		PasswordRequireSpecial    *bool       `json:"password_require_special"`
		SMTPHost                  *string     `json:"smtp_host"`
		SMTPPort                  json.Number `json:"smtp_port"` // accepts "587" or 587
		SMTPFrom                  *string     `json:"smtp_from"`
		SMTPUsername              *string     `json:"smtp_username"` // pointer so empty string clears it
		SMTPPassword              *string     `json:"smtp_password"` // pointer so empty string clears it
		IMAPEnabled               *bool       `json:"imap_enabled"`
		IMAPHost                  *string     `json:"imap_host"`
		IMAPPort                  json.Number `json:"imap_port"`
		IMAPUsername              *string     `json:"imap_username"`
		IMAPPassword              *string     `json:"imap_password"`
		IMAPUseTLS                *bool       `json:"imap_use_tls"`
		IMAPMailbox               *string     `json:"imap_mailbox"`
		IMAPProcessedMailbox      *string     `json:"imap_processed_mailbox"`
		IMAPPollInterval          json.Number `json:"imap_poll_interval"`
		IMAPAuthMechanism         *string     `json:"imap_auth_mechanism"`
		IMAPOAuth2Provider        *string     `json:"imap_oauth2_provider"`
		IMAPAccessToken           *string     `json:"imap_access_token"`
		IMAPRefreshToken          *string     `json:"imap_refresh_token"`
		SessionTimeoutMinutes     *int        `json:"session_timeout_minutes"`
		CompanyName               *string     `json:"company_name"`
		CompanyLogo               *string     `json:"company_logo"`
		CompanyLogoDark           *string     `json:"company_logo_dark"`
		DefaultColumns            *string     `json:"default_columns"`
		DefaultLabels             *string     `json:"default_labels"`
		BackupSchedule            *string     `json:"backup_schedule"`
		BackupStartTime           *string     `json:"backup_start_time"`
		BackupKeep                *int        `json:"backup_keep"`
		BackupEmailEnabled        *bool       `json:"backup_email_enabled"`
		BackupEmailAddress        *string     `json:"backup_email_address"`
		ScrumStorypointsEnabled   *bool       `json:"scrum_storypoints_enabled"`
		GravatarEnabled           *bool       `json:"gravatar_enabled"`
		ExternalImageProxyEnabled *bool       `json:"external_image_proxy_enabled"`
		LoginBrandingEnabled      *bool       `json:"login_branding_enabled"`
		AllowedIPs                *string     `json:"allowed_ips"`
		PasswordChangePeriodDays  *int        `json:"password_change_period_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.PasswordMinLength != nil {
		minLen := *req.PasswordMinLength
		if minLen < 8 {
			minLen = 8
		}
		saveSetting(settingPasswordMinLength, fmt.Sprintf("%d", minLen))
	}
	boolSetting := func(ptr *bool, key string) {
		if ptr == nil {
			return
		}
		if *ptr {
			saveSetting(key, "true")
		} else {
			saveSetting(key, "false")
		}
	}
	boolSetting(req.PasswordRequireUpper, settingPasswordRequireUpper)
	boolSetting(req.PasswordRequireLower, settingPasswordRequireLower)
	boolSetting(req.PasswordRequireDigit, settingPasswordRequireDigit)
	boolSetting(req.PasswordRequireSpecial, settingPasswordRequireSpecial)
	if req.MFARequired != nil {
		val := "false"
		if *req.MFARequired {
			val = "true"
		}
		saveSetting(settingMFARequired, val)
	}
	if req.RegistrationEnabled != nil {
		val := "true"
		if !*req.RegistrationEnabled {
			val = "false"
		}
		saveSetting(settingRegistrationEnabled, val)
	}
	if req.DefaultDateTimeFormat != "" {
		saveSetting(settingDefaultDateTimeFormat, req.DefaultDateTimeFormat)
	}
	if req.DefaultTimezone != "" {
		saveSetting(settingDefaultTimezone, req.DefaultTimezone)
	}
	if req.DefaultTheme == "light" || req.DefaultTheme == "dark" || req.DefaultTheme == "system" {
		saveSetting(settingDefaultTheme, req.DefaultTheme)
	}
	if req.DefaultFont != "" {
		saveSetting(settingDefaultFont, req.DefaultFont)
	}
	if req.DefaultFontSize != "" {
		saveSetting(settingDefaultFontSize, req.DefaultFontSize)
	}
	validLocales := map[string]bool{"en": true, "nl": true, "de": true, "fr": true, "es": true, "da": true, "sv": true, "nb": true, "fi": true, "is": true, "pt": true, "it": true}
	if validLocales[req.DefaultLocale] {
		saveSetting(settingDefaultLocale, req.DefaultLocale)
	}
	// SMTP — only save fields that were explicitly included in the request
	// (pointer fields: nil means "not sent", so don't overwrite; empty string clears)
	if req.SMTPHost != nil {
		saveSetting(settingSMTPHost, *req.SMTPHost)
	}
	if req.SMTPPort != "" {
		saveSetting(settingSMTPPort, req.SMTPPort.String())
	}
	if req.SMTPFrom != nil {
		saveSetting(settingSMTPFrom, *req.SMTPFrom)
	}
	if req.SMTPUsername != nil {
		saveSetting(settingSMTPUsername, *req.SMTPUsername)
	}
	if req.SMTPPassword != nil {
		saveSetting(settingSMTPPassword, *req.SMTPPassword)
	}
	// IMAP
	boolSetting(req.IMAPEnabled, settingIMAPEnabled)
	if req.IMAPHost != nil {
		saveSetting(settingIMAPHost, *req.IMAPHost)
	}
	if req.IMAPPort != "" {
		saveSetting(settingIMAPPort, req.IMAPPort.String())
	}
	if req.IMAPUsername != nil {
		saveSetting(settingIMAPUsername, *req.IMAPUsername)
	}
	if req.IMAPPassword != nil {
		saveSetting(settingIMAPPassword, *req.IMAPPassword)
	}
	boolSetting(req.IMAPUseTLS, settingIMAPUseTLS)
	if req.IMAPMailbox != nil {
		saveSetting(settingIMAPMailbox, strings.TrimSpace(*req.IMAPMailbox))
	}
	if req.IMAPProcessedMailbox != nil {
		saveSetting(settingIMAPProcessedMailbox, strings.TrimSpace(*req.IMAPProcessedMailbox))
	}
	if req.IMAPPollInterval != "" {
		saveSetting(settingIMAPPollInterval, req.IMAPPollInterval.String())
	}
	if req.IMAPAuthMechanism != nil {
		saveSetting(settingIMAPAuthMechanism, *req.IMAPAuthMechanism)
	}
	if req.IMAPOAuth2Provider != nil {
		saveSetting(settingIMAPOAuth2Provider, *req.IMAPOAuth2Provider)
	}
	if req.IMAPAccessToken != nil {
		saveSetting(settingIMAPAccessToken, *req.IMAPAccessToken)
	}
	if req.IMAPRefreshToken != nil {
		saveSetting(settingIMAPRefreshToken, *req.IMAPRefreshToken)
	}
	if req.SessionTimeoutMinutes != nil {
		timeout := *req.SessionTimeoutMinutes
		if timeout < 0 {
			timeout = 0
		}
		saveSetting(settingSessionTimeoutMinutes, fmt.Sprintf("%d", timeout))
	}
	if req.CompanyName != nil {
		saveSetting(settingCompanyName, *req.CompanyName)
	}
	if req.CompanyLogo != nil {
		saveSetting(settingCompanyLogo, *req.CompanyLogo)
	}
	if req.CompanyLogoDark != nil {
		saveSetting(settingCompanyLogoDark, *req.CompanyLogoDark)
	}
	if req.DefaultColumns != nil {
		saveSetting(settingDefaultColumns, *req.DefaultColumns)
	}
	if req.DefaultLabels != nil {
		saveSetting(settingDefaultLabels, *req.DefaultLabels)
	}
	validSchedules := map[string]bool{"disabled": true, "6h": true, "8h": true, "12h": true, "24h": true}
	if req.BackupSchedule != nil && validSchedules[*req.BackupSchedule] {
		saveSetting(settingBackupSchedule, *req.BackupSchedule)
	}
	if req.BackupStartTime != nil {
		// Accept "" (clear) or HH:MM format
		v := *req.BackupStartTime
		if v == "" || isValidHHMM(v) {
			saveSetting(settingBackupStartTime, v)
		}
	}
	if req.BackupKeep != nil {
		keep := *req.BackupKeep
		if keep < 1 {
			keep = 1
		}
		saveSetting(settingBackupKeep, fmt.Sprintf("%d", keep))
	}
	boolSetting(req.BackupEmailEnabled, settingBackupEmailEnabled)
	if req.BackupEmailAddress != nil {
		saveSetting(settingBackupEmailAddress, *req.BackupEmailAddress)
	}
	boolSetting(req.ScrumStorypointsEnabled, settingScrumStorypointsEnabled)
	boolSetting(req.GravatarEnabled, settingGravatarEnabled)
	boolSetting(req.ExternalImageProxyEnabled, settingExternalImageProxyEnabled)
	boolSetting(req.LoginBrandingEnabled, settingLoginBrandingEnabled)
	if req.AllowedIPs != nil {
		saveSetting(settingAllowedIPs, *req.AllowedIPs)
	}
	if req.PasswordChangePeriodDays != nil {
		days := *req.PasswordChangePeriodDays
		if days < 0 {
			days = 0
		}
		saveSetting(settingPasswordChangePeriodDays, fmt.Sprintf("%d", days))
	}

	AdminGetSystemSettings(c)
}

// AdminSendTestEmail sends a test email to verify the current SMTP configuration.
func AdminSendTestEmail(c *gin.Context) {
	var req struct {
		To string `json:"to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email address is required"})
		return
	}

	cfg := GetSMTPSettings()
	if cfg.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SMTP host is not configured"})
		return
	}

	emailSvc := services.GetEmailService()
	if emailSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "email service not initialised"})
		return
	}

	plainBody := "This is a test email from WarmDesk. Your SMTP configuration is working correctly."
	htmlContent := `<tr><td style="padding:28px 32px;font-size:15px;color:#333;line-height:1.6;text-align:center">` +
		`<p style="margin:0 0 16px;font-size:18px">&#10003; &nbsp;SMTP is configured correctly</p>` +
		`<p style="margin:0;color:#666;font-size:14px">This is a test email from WarmDesk.<br>If you received this, your mail settings are working.</p>` +
		`</td></tr>`

	if err := emailSvc.SendHTML(req.To, "WarmDesk SMTP test",
		services.WrapHTML("SMTP Test", htmlContent),
		services.WrapText("SMTP Test", plainBody),
	); err != nil {
		log.Printf("test email to %s failed: %v", req.To, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test email"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Test email sent to " + req.To})
}

// PasswordPolicy holds the active password requirements.
type PasswordPolicy struct {
	MinLength      int  `json:"min_length"`
	RequireUpper   bool `json:"require_upper"`
	RequireLower   bool `json:"require_lower"`
	RequireDigit   bool `json:"require_digit"`
	RequireSpecial bool `json:"require_special"`
}

// GetPasswordPolicy returns the current password policy from system settings.
func GetPasswordPolicy() PasswordPolicy {
	all := loadAllSettings()
	minLen, _ := strconv.Atoi(all[settingPasswordMinLength])
	if minLen < 8 {
		minLen = 8
	}
	return PasswordPolicy{
		MinLength:      minLen,
		RequireUpper:   all[settingPasswordRequireUpper] == "true",
		RequireLower:   all[settingPasswordRequireLower] == "true",
		RequireDigit:   all[settingPasswordRequireDigit] == "true",
		RequireSpecial: all[settingPasswordRequireSpecial] == "true",
	}
}

// ValidatePasswordPolicy checks password against the current policy.
// Returns a human-readable error message, or "" when the password is valid.
func ValidatePasswordPolicy(password string) string {
	p := GetPasswordPolicy()
	if len(password) < p.MinLength {
		return fmt.Sprintf("password must be at least %d characters", p.MinLength)
	}
	if p.RequireUpper && !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return "password must contain at least one uppercase letter"
	}
	if p.RequireLower && !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return "password must contain at least one lowercase letter"
	}
	if p.RequireDigit && !strings.ContainsAny(password, "0123456789") {
		return "password must contain at least one digit"
	}
	if p.RequireSpecial && !strings.ContainsAny(password, `!@#$%^&*()_+-=[]{}|;':",./<>?`) {
		return "password must contain at least one special character"
	}
	return ""
}

// IsMFARequired returns true when the admin has enabled system-wide MFA enforcement.
func IsMFARequired() bool {
	return loadAllSettings()[settingMFARequired] == "true"
}

// GetCompanyName returns the configured company name (used as the TOTP issuer).
func GetCompanyName() string {
	return loadAllSettings()[settingCompanyName]
}

// IsRegistrationEnabled is a helper used by the Register handler.
func IsRegistrationEnabled() bool {
	var setting models.SystemSetting
	result := database.DB.First(&setting, "key = ?", settingRegistrationEnabled)
	if result.Error != nil {
		return true // default: enabled
	}
	return setting.Value != "false"
}

// GetGlobalDefaults returns the global default settings for new users.
func GetGlobalDefaults() map[string]string {
	all := loadAllSettings()
	return map[string]string{
		"date_time_format": all[settingDefaultDateTimeFormat],
		"timezone":         all[settingDefaultTimezone],
		"theme":            all[settingDefaultTheme],
		"font":             all[settingDefaultFont],
		"font_size":        all[settingDefaultFontSize],
		"locale":           all[settingDefaultLocale],
	}
}

// saveSetting upserts a system setting by key.
// Uses update-or-create to work reliably with all DB drivers and SQLite versions.
func saveSetting(key, value string) {
	result := database.DB.Model(&models.SystemSetting{}).Where("key = ?", key).Update("value", value)
	if result.RowsAffected == 0 {
		database.DB.Create(&models.SystemSetting{Key: key, Value: value})
	}
}

// loadAllSettings reads all system settings from DB and fills in defaults for missing keys.
func loadAllSettings() map[string]string {
	result := map[string]string{}
	for k, v := range systemSettingDefaults {
		result[k] = v
	}
	var settings []models.SystemSetting
	database.DB.Find(&settings)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result
}

// isValidHHMM returns true if s is a valid 24-hour time string in HH:MM format.
func isValidHHMM(s string) bool {
	if len(s) != 5 || s[2] != ':' {
		return false
	}
	h, err1 := strconv.Atoi(s[0:2])
	m, err2 := strconv.Atoi(s[3:5])
	return err1 == nil && err2 == nil && h >= 0 && h <= 23 && m >= 0 && m <= 59
}
