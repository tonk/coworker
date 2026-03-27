package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/config"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/models"
)

const (
	settingRegistrationEnabled    = "registration_enabled"
	settingDefaultDateTimeFormat  = "default_date_time_format"
	settingDefaultTimezone        = "default_timezone"
	settingDefaultTheme           = "default_theme"
	settingDefaultFont            = "default_font"
	settingDefaultFontSize        = "default_font_size"
	settingDefaultLocale          = "default_locale"
)

var systemSettingDefaults = map[string]string{
	settingRegistrationEnabled:   "true",
	settingDefaultDateTimeFormat: "YYYY-MM-DD HH:mm",
	settingDefaultTimezone:       "UTC",
	settingDefaultTheme:          "system",
	settingDefaultFont:           "system",
	settingDefaultFontSize:       "14",
	settingDefaultLocale:         "en",
}

// InitSystemDefaults seeds the in-memory defaults from the config file so that
// settings not yet stored in the database reflect the operator's preferences.
func InitSystemDefaults(cfg *config.Config) {
	if cfg.DefaultLocale != "" {
		systemSettingDefaults[settingDefaultLocale] = cfg.DefaultLocale
	}
}

// GetSystemSettings returns public system settings (registration + global UI defaults).
func GetSystemSettings(c *gin.Context) {
	all := loadAllSettings()
	c.JSON(http.StatusOK, gin.H{
		"registration_enabled":        all[settingRegistrationEnabled] != "false",
		"default_date_time_format":    all[settingDefaultDateTimeFormat],
		"default_timezone":            all[settingDefaultTimezone],
		"default_theme":               all[settingDefaultTheme],
		"default_font":                all[settingDefaultFont],
		"default_font_size":           all[settingDefaultFontSize],
		"default_locale":              all[settingDefaultLocale],
	})
}

// AdminGetSystemSettings returns all system settings for admins.
func AdminGetSystemSettings(c *gin.Context) {
	all := loadAllSettings()
	c.JSON(http.StatusOK, all)
}

// AdminUpdateSystemSettings updates system settings.
func AdminUpdateSystemSettings(c *gin.Context) {
	var req struct {
		RegistrationEnabled   *bool  `json:"registration_enabled"`
		DefaultDateTimeFormat string `json:"default_date_time_format"`
		DefaultTimezone       string `json:"default_timezone"`
		DefaultTheme          string `json:"default_theme"`
		DefaultFont           string `json:"default_font"`
		DefaultFontSize       string `json:"default_font_size"`
		DefaultLocale         string `json:"default_locale"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RegistrationEnabled != nil {
		val := "true"
		if !*req.RegistrationEnabled {
			val = "false"
		}
		database.DB.Save(&models.SystemSetting{Key: settingRegistrationEnabled, Value: val})
	}
	if req.DefaultDateTimeFormat != "" {
		database.DB.Save(&models.SystemSetting{Key: settingDefaultDateTimeFormat, Value: req.DefaultDateTimeFormat})
	}
	if req.DefaultTimezone != "" {
		database.DB.Save(&models.SystemSetting{Key: settingDefaultTimezone, Value: req.DefaultTimezone})
	}
	if req.DefaultTheme == "light" || req.DefaultTheme == "dark" || req.DefaultTheme == "system" {
		database.DB.Save(&models.SystemSetting{Key: settingDefaultTheme, Value: req.DefaultTheme})
	}
	if req.DefaultFont != "" {
		database.DB.Save(&models.SystemSetting{Key: settingDefaultFont, Value: req.DefaultFont})
	}
	if req.DefaultFontSize != "" {
		database.DB.Save(&models.SystemSetting{Key: settingDefaultFontSize, Value: req.DefaultFontSize})
	}
	validLocales := map[string]bool{"en": true, "nl": true, "de": true, "fr": true, "es": true}
	if validLocales[req.DefaultLocale] {
		database.DB.Save(&models.SystemSetting{Key: settingDefaultLocale, Value: req.DefaultLocale})
	}

	AdminGetSystemSettings(c)
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
