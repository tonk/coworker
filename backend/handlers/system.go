package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

const (
	settingRegistrationEnabled    = "registration_enabled"
	settingDefaultDateTimeFormat  = "default_date_time_format"
	settingDefaultTimezone        = "default_timezone"
	settingDefaultTheme           = "default_theme"
	settingDefaultFont            = "default_font"
	settingDefaultFontSize        = "default_font_size"
	settingDefaultLocale          = "default_locale"
	settingSMTPHost               = "smtp_host"
	settingSMTPPort               = "smtp_port"
	settingSMTPFrom               = "smtp_from"
	settingSMTPUsername           = "smtp_username"
	settingSMTPPassword           = "smtp_password"
	settingSessionTimeoutMinutes  = "session_timeout_minutes"
	settingCompanyName            = "company_name"
	settingCompanyLogo            = "company_logo"
	settingDefaultColumns         = "default_columns"
	settingDefaultLabels          = "default_labels"
	settingMFARequired            = "mfa_required"
	settingPasswordMinLength      = "password_min_length"
	settingPasswordRequireUpper   = "password_require_upper"
	settingPasswordRequireLower   = "password_require_lower"
	settingPasswordRequireDigit   = "password_require_digit"
	settingPasswordRequireSpecial = "password_require_special"
	settingBackupSchedule         = "backup_schedule"
	settingBackupStartTime        = "backup_start_time"
	settingBackupLastRun          = "backup_last_run"
	settingBackupKeep             = "backup_keep"
)

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
	settingRegistrationEnabled:    "true",
	settingDefaultDateTimeFormat:  "YYYY-MM-DD HH:mm",
	settingDefaultTimezone:        "UTC",
	settingDefaultTheme:           "system",
	settingDefaultFont:            "system",
	settingDefaultFontSize:        "14",
	settingDefaultLocale:          "en",
	settingSMTPHost:               "",
	settingSMTPPort:               "587",
	settingSMTPFrom:               "",
	settingSMTPUsername:           "",
	settingSMTPPassword:           "",
	settingSessionTimeoutMinutes:  "60",
	settingCompanyName:            "",
	settingCompanyLogo:            "",
	settingDefaultColumns:         "Backlog",
	settingDefaultLabels:          "Bug\nFeature\nDesign\nContent",
	settingMFARequired:            "false",
	settingPasswordMinLength:      "12",
	settingPasswordRequireUpper:   "false",
	settingPasswordRequireLower:   "false",
	settingPasswordRequireDigit:   "false",
	settingPasswordRequireSpecial: "false",
	settingBackupSchedule:         "disabled",
	settingBackupStartTime:        "",
	settingBackupLastRun:          "",
	settingBackupKeep:             "10",
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
		"registration_enabled":        all[settingRegistrationEnabled] != "false",
		"default_date_time_format":    all[settingDefaultDateTimeFormat],
		"default_timezone":            all[settingDefaultTimezone],
		"default_theme":               all[settingDefaultTheme],
		"default_font":                all[settingDefaultFont],
		"default_font_size":           all[settingDefaultFontSize],
		"default_locale":              all[settingDefaultLocale],
		"session_timeout_minutes":     timeoutMinutes,
		"company_name":                all[settingCompanyName],
		"company_logo":                all[settingCompanyLogo],
		"mfa_required":                all[settingMFARequired] == "true",
		"password_policy":             GetPasswordPolicy(),
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
	c.JSON(http.StatusOK, all)
}

// AdminUpdateSystemSettings updates system settings.
func AdminUpdateSystemSettings(c *gin.Context) {
	var req struct {
		MFARequired            *bool   `json:"mfa_required"`
		RegistrationEnabled    *bool   `json:"registration_enabled"`
		DefaultDateTimeFormat  string  `json:"default_date_time_format"`
		DefaultTimezone        string  `json:"default_timezone"`
		DefaultTheme           string  `json:"default_theme"`
		DefaultFont            string  `json:"default_font"`
		DefaultFontSize        string  `json:"default_font_size"`
		DefaultLocale          string  `json:"default_locale"`
		PasswordMinLength      *int    `json:"password_min_length"`
		PasswordRequireUpper   *bool   `json:"password_require_upper"`
		PasswordRequireLower   *bool   `json:"password_require_lower"`
		PasswordRequireDigit   *bool   `json:"password_require_digit"`
		PasswordRequireSpecial *bool   `json:"password_require_special"`
		SMTPHost               *string `json:"smtp_host"`
		SMTPPort               json.Number `json:"smtp_port"` // accepts "587" or 587
		SMTPFrom               *string `json:"smtp_from"`
		SMTPUsername           *string `json:"smtp_username"` // pointer so empty string clears it
		SMTPPassword           *string `json:"smtp_password"` // pointer so empty string clears it
		SessionTimeoutMinutes  *int    `json:"session_timeout_minutes"`
		CompanyName            *string `json:"company_name"`
		CompanyLogo            *string `json:"company_logo"`
		DefaultColumns         *string `json:"default_columns"`
		DefaultLabels          *string `json:"default_labels"`
		BackupSchedule         *string `json:"backup_schedule"`
		BackupStartTime        *string `json:"backup_start_time"`
		BackupKeep             *int    `json:"backup_keep"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	validLocales := map[string]bool{"en": true, "nl": true, "de": true, "fr": true, "es": true}
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

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	from := cfg.From
	if from == "" {
		from = "warmdesk@localhost"
	}
	body := "This is a test email from WarmDesk. Your SMTP configuration is working correctly."
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: WarmDesk SMTP test\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, req.To, body)

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}
	if err := smtp.SendMail(addr, auth, from, []string{req.To}, []byte(msg)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
