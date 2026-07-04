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
	if c.Query("deleted") == "true" {
		database.DB.Unscoped().Where("deleted_at IS NOT NULL").Find(&users)
	} else {
		database.DB.Find(&users)
	}
	c.JSON(http.StatusOK, users)
}

// AdminRestoreUser un-deletes a soft-deleted user.
func AdminRestoreUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result := database.DB.Unscoped().Model(&models.User{}).Where("id = ? AND deleted_at IS NOT NULL", id).Update("deleted_at", nil)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "deleted user not found"})
		return
	}
	var user models.User
	database.DB.First(&user, id)
	recordAdminEvent(c, user.ID, user.Username, "admin_user_restored", "")
	c.JSON(http.StatusOK, user)
}

// AdminPurgeUser permanently removes a soft-deleted user and cleans up FK references.
// Records that have value beyond the user (tickets, cards, messages) are preserved with
// their user FK nullified. Membership and personal records are deleted outright.
func AdminPurgeUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	uid := uint(id)

	// Nullify FKs on records that should survive (content / audit trail)
	database.DB.Unscoped().Model(&models.Ticket{}).Where("assigned_to_id = ?", uid).Update("assigned_to_id", nil)
	database.DB.Unscoped().Model(&models.Ticket{}).Where("owner_id = ?", uid).Update("owner_id", nil)
	database.DB.Unscoped().Model(&models.Ticket{}).Where("created_by_id = ?", uid).Update("created_by_id", nil)
	database.DB.Unscoped().Model(&models.TicketMessage{}).Where("user_id = ?", uid).Update("user_id", nil)
	database.DB.Unscoped().Model(&models.TicketHistory{}).Where("user_id = ?", uid).Update("user_id", nil)
	database.DB.Unscoped().Model(&models.CardComment{}).Where("user_id = ?", uid).Update("user_id", nil)
	database.DB.Unscoped().Model(&models.CardHistory{}).Where("user_id = ?", uid).Update("user_id", nil)
	database.DB.Unscoped().Model(&models.Card{}).Where("assignee_id = ?", uid).Update("assignee_id", nil)
	database.DB.Unscoped().Model(&models.Card{}).Where("created_by_id = ?", uid).Update("created_by_id", nil)
	database.DB.Unscoped().Model(&models.TimeEntry{}).Where("user_id = ?", uid).Update("user_id", nil)
	database.DB.Unscoped().Model(&models.Topic{}).Where("user_id = ?", uid).Update("user_id", nil)
	database.DB.Unscoped().Model(&models.TopicReply{}).Where("user_id = ?", uid).Update("user_id", nil)
	database.DB.Unscoped().Model(&models.ChatMessage{}).Where("user_id = ?", uid).Update("user_id", nil)
	database.DB.Unscoped().Model(&models.ConversationMessage{}).Where("sender_id = ?", uid).Update("sender_id", nil)
	database.DB.Unscoped().Model(&models.DirectMessage{}).Where("sender_id = ?", uid).Update("sender_id", nil)
	database.DB.Unscoped().Model(&models.DirectMessage{}).Where("receiver_id = ?", uid).Update("receiver_id", nil)
	database.DB.Unscoped().Model(&models.Attachment{}).Where("uploader_id = ?", uid).Update("uploader_id", nil)
	database.DB.Unscoped().Model(&models.TicketCardLink{}).Where("created_by_id = ?", uid).Update("created_by_id", nil)
	database.DB.Unscoped().Model(&models.ProjectWebhook{}).Where("created_by_id = ?", uid).Update("created_by_id", nil)
	database.DB.Unscoped().Model(&models.Customer{}).Where("created_by_id = ?", uid).Update("created_by_id", nil)

	// Delete personal / membership records
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.APIKey{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.StarredProject{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.CustomerFavorite{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.CustomerAccess{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.GroupMember{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.ProjectMember{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.PasskeyCredential{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.TimeEntryRowOrder{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.TimeMacroLibrary{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.FavoriteUser{})
	database.DB.Unscoped().Where("favorite_user_id = ?", uid).Delete(&models.FavoriteUser{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.CardAssignee{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.ConversationMember{})
	database.DB.Unscoped().Where("user_id = ?", uid).Delete(&models.MessageReaction{})

	// Hard-delete the user row itself (load username before it's gone)
	var purgedUser models.User
	database.DB.Unscoped().Select("id, username").First(&purgedUser, uid)
	database.DB.Unscoped().Delete(&models.User{}, uid)
	recordAdminEvent(c, uid, purgedUser.Username, "admin_user_purged", "")
	c.JSON(http.StatusOK, gin.H{"message": "purged"})
}

func AdminGetUserLoginHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var history []models.LoginHistory
	database.DB.Where("user_id = ?", id).Order("created_at DESC").Limit(500).Find(&history)
	c.JSON(http.StatusOK, history)
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
		TimeTrackingEnabled *bool  `json:"time_tracking_enabled"`
		BoardEnabled        *bool  `json:"board_enabled"`
		ChatEnabled         *bool  `json:"chat_enabled"`
		HelpdeskEnabled     *bool  `json:"helpdesk_enabled"`
		Theme               string `json:"theme"`
		ShowBreadcrumbs     *bool  `json:"show_breadcrumbs"`
		EmailNotifications  *bool  `json:"email_notifications"`
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
		TimeNotation        string `json:"time_notation"`
		WeekStart           string `json:"week_start"`
		MonWorkStart        string `json:"mon_work_start"`
		MonWorkEnd          string `json:"mon_work_end"`
		TueWorkStart        string `json:"tue_work_start"`
		TueWorkEnd          string `json:"tue_work_end"`
		WedWorkStart        string `json:"wed_work_start"`
		WedWorkEnd          string `json:"wed_work_end"`
		ThuWorkStart        string `json:"thu_work_start"`
		ThuWorkEnd          string `json:"thu_work_end"`
		FriWorkStart        string `json:"fri_work_start"`
		FriWorkEnd          string `json:"fri_work_end"`
		SatWorkStart        string `json:"sat_work_start"`
		SatWorkEnd          string `json:"sat_work_end"`
		SunWorkStart        string `json:"sun_work_start"`
		SunWorkEnd          string `json:"sun_work_end"`
		LunchBreakMinutes   int    `json:"lunch_break_minutes"`
		MustChangePassword  *bool  `json:"must_change_password"`
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
	if req.TimeTrackingEnabled != nil {
		updates["time_tracking_enabled"] = *req.TimeTrackingEnabled
	}
	if req.BoardEnabled != nil {
		updates["board_enabled"] = *req.BoardEnabled
	}
	if req.ChatEnabled != nil {
		updates["chat_enabled"] = *req.ChatEnabled
	}
	if req.HelpdeskEnabled != nil {
		updates["helpdesk_enabled"] = *req.HelpdeskEnabled
	}
	validThemes := map[string]bool{"light": true, "dark": true, "system": true}
	if validThemes[req.Theme] {
		updates["theme"] = req.Theme
	}
	if req.ShowBreadcrumbs != nil {
		updates["show_breadcrumbs"] = *req.ShowBreadcrumbs
	}
	if req.EmailNotifications != nil {
		updates["email_notifications"] = *req.EmailNotifications
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
	if req.TimeNotation == "decimal" || req.TimeNotation == "hhmm" {
		updates["time_notation"] = req.TimeNotation
	}
	if req.WeekStart == "monday" || req.WeekStart == "sunday" {
		updates["week_start"] = req.WeekStart
	}
	updates["mon_work_start"] = req.MonWorkStart
	updates["mon_work_end"] = req.MonWorkEnd
	updates["tue_work_start"] = req.TueWorkStart
	updates["tue_work_end"] = req.TueWorkEnd
	updates["wed_work_start"] = req.WedWorkStart
	updates["wed_work_end"] = req.WedWorkEnd
	updates["thu_work_start"] = req.ThuWorkStart
	updates["thu_work_end"] = req.ThuWorkEnd
	updates["fri_work_start"] = req.FriWorkStart
	updates["fri_work_end"] = req.FriWorkEnd
	updates["sat_work_start"] = req.SatWorkStart
	updates["sat_work_end"] = req.SatWorkEnd
	updates["sun_work_start"] = req.SunWorkStart
	updates["sun_work_end"] = req.SunWorkEnd
	updates["lunch_break_minutes"] = req.LunchBreakMinutes
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
		updates["password_changed_at"] = time.Now()
	}
	if req.MustChangePassword != nil {
		updates["must_change_password"] = *req.MustChangePassword
	}

	// Service-account roles must not have feature access — enforce regardless of
	// what the request sent for the feature flags.
	finalRole, _ := updates["global_role"].(string)
	if finalRole == "" {
		var existing models.User
		database.DB.Select("global_role").First(&existing, id)
		finalRole = existing.GlobalRole
	}
	if finalRole == "metrics" || finalRole == "backup" {
		updates["time_tracking_enabled"] = false
		updates["time_tracking_viewer"] = false
		updates["board_enabled"] = false
		updates["chat_enabled"] = false
		updates["helpdesk_enabled"] = false
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

	// Build a detail string for security-sensitive fields only.
	var details []string
	if _, ok := updates["global_role"]; ok {
		details = append(details, "role="+user.GlobalRole)
	}
	if _, ok := updates["password_hash"]; ok {
		details = append(details, "password_reset")
	}
	if v, ok := updates["is_active"]; ok {
		if active, _ := v.(bool); active {
			details = append(details, "activated")
		} else {
			details = append(details, "deactivated")
		}
	}
	detail := strings.Join(details, " ")
	recordAdminEvent(c, user.ID, user.Username, "admin_user_updated", detail)
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
	recordAdminEvent(c, user.ID, user.Username, "admin_mfa_disabled", "")
	c.JSON(http.StatusOK, user)
}

// AdminListUserPasskeys returns all registered passkeys for a user (admin only).
func AdminListUserPasskeys(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var creds []models.PasskeyCredential
	database.DB.Where("user_id = ?", id).Order("created_at asc").Find(&creds)
	c.JSON(http.StatusOK, creds)
}

// AdminRevokeUserPasskeys deletes all registered passkeys for a user (admin only).
func AdminRevokeUserPasskeys(c *gin.Context) {
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
	database.DB.Where("user_id = ?", id).Delete(&models.PasskeyCredential{})
	recordAdminEvent(c, user.ID, user.Username, "admin_passkeys_revoked", "")
	c.JSON(http.StatusOK, gin.H{"message": "passkeys revoked"})
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
	var targetUser models.User
	if err := database.DB.Select("id, username").First(&targetUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	database.DB.Where("user_id = ?", id).Delete(&models.CustomerAccess{})
	database.DB.Delete(&models.User{}, id)
	recordAdminEvent(c, targetUser.ID, targetUser.Username, "admin_user_deleted", "")
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func AdminCreateUser(c *gin.Context) {
	var req struct {
		Email              string `json:"email" binding:"required,email"`
		Username           string `json:"username" binding:"required,min=3,max=50"`
		Password           string `json:"password" binding:"required,min=8"`
		FirstName          string `json:"first_name"`
		LastName           string `json:"last_name"`
		DisplayName        string `json:"display_name"`
		GlobalRole         string `json:"global_role"`
		MustChangePassword bool   `json:"must_change_password"`
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
	switch role {
	case "admin", "user", "metrics", "backup":
	default:
		role = "user"
	}

	defs := GetGlobalDefaults()
	user := models.User{
		Email:              strings.ToLower(req.Email),
		Username:           req.Username,
		PasswordHash:       hash,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		DisplayName:        displayName,
		GlobalRole:         role,
		Locale:             "en",
		IsActive:           true,
		DateTimeFormat:     defs["date_time_format"],
		Timezone:           defs["timezone"],
		Theme:              defs["theme"],
		Font:               defs["font"],
		FontSize:           defs["font_size"],
		MustChangePassword: req.MustChangePassword,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "Duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	recordAdminEvent(c, user.ID, user.Username, "admin_user_created", "role="+user.GlobalRole)
	c.JSON(http.StatusCreated, user)
}
