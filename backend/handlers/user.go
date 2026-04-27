package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"golang.org/x/crypto/bcrypt"
)

func AdminListUsers(c *gin.Context) {
	var users []models.User
	database.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

func AdminGetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func AdminUpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		GlobalRole          string `json:"global_role"`
		IsActive            *bool  `json:"is_active"`
		TimeTrackingViewer  *bool  `json:"time_tracking_viewer"`
		FirstName           string `json:"first_name"`
		LastName            string `json:"last_name"`
		DisplayName         string `json:"display_name"`
		AvatarURL           string `json:"avatar_url"`
		Email               string `json:"email"`
		Password            string `json:"password"`
		Locale              string `json:"locale"`
		DateTimeFormat      string `json:"date_time_format"`
		Timezone            string `json:"timezone"`
		Font                string `json:"font"`
		FontSize            string `json:"font_size"`
		SidebarPosition     string `json:"sidebar_position"`
		AccentColor         string `json:"accent_color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updates := map[string]interface{}{}
	validRoles := map[string]bool{"admin": true, "user": true, "viewer": true, "metrics": true, "backup": true}
	if validRoles[req.GlobalRole] {
		updates["global_role"] = req.GlobalRole
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.TimeTrackingViewer != nil {
		updates["time_tracking_viewer"] = *req.TimeTrackingViewer
	}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}
	if req.Email != "" {
		updates["email"] = strings.ToLower(req.Email)
	}
	if req.Locale == "en" || req.Locale == "nl" {
		updates["locale"] = req.Locale
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
	validAccents := map[string]bool{"blue": true, "red": true, "green": true, "orange": true}
	if validAccents[req.AccentColor] {
		updates["accent_color"] = req.AccentColor
	}
	if req.Password != "" {
		if len(req.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}
		updates["password_hash"] = string(hash)
	}

	if len(updates) > 0 {
		updates["settings_updated_at"] = time.Now()
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var user models.User
	database.DB.First(&user, id)
	c.JSON(http.StatusOK, user)
}

// AdminDisableUserMFA clears the TOTP secret and disables MFA for a user.
func AdminDisableUserMFA(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	database.DB.Model(&user).Updates(map[string]interface{}{
		"totp_enabled": false,
		"totp_secret":  "",
	})
	database.DB.First(&user, id)
	c.JSON(http.StatusOK, user)
}

func AdminDeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if uint(id) == middleware.GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete your own account"})
		return
	}
	database.DB.Where("user_id = ?", id).Delete(&models.CustomerAccess{})
	database.DB.Delete(&models.User{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func AdminCreateUser(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Username    string `json:"username" binding:"required,min=3,max=50"`
		Password    string `json:"password" binding:"required,min=8"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		DisplayName string `json:"display_name"`
		GlobalRole  string `json:"global_role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	hash := string(hashBytes)

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}
	role := req.GlobalRole
	if role != "admin" && role != "user" {
		role = "user"
	}

	defs := GetGlobalDefaults()
	user := models.User{
		Email:          strings.ToLower(req.Email),
		Username:       req.Username,
		PasswordHash:   hash,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DisplayName:    displayName,
		GlobalRole:     role,
		Locale:         "en",
		IsActive:       true,
		DateTimeFormat: defs["date_time_format"],
		Timezone:       defs["timezone"],
		Theme:          defs["theme"],
		Font:           defs["font"],
		FontSize:       defs["font_size"],
	}

	if err := database.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "Duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, user)
}
