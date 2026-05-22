package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
	appws "github.com/tonk/warmdesk/ws"
)

func generateWebhookToken() (token, hint string) {
	b := make([]byte, 32)
	rand.Read(b)
	token = hex.EncodeToString(b)
	hint = token[len(token)-8:]
	return
}

// hashWebhookToken returns the SHA-256 hex digest of a plaintext webhook token.
// Tokens are stored as hashes in the DB so a DB dump does not expose usable tokens.
func hashWebhookToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// ListWebhooks GET /projects/:projectSlug/webhooks
func ListWebhooks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var hooks []models.ProjectWebhook
	database.DB.Where("project_id = ?", project.ID).Find(&hooks)
	c.JSON(http.StatusOK, hooks)
}

// CreateWebhook POST /projects/:projectSlug/webhooks
// Body: {"name": "CI Bot"}
func CreateWebhook(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
		Type string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	hookType := req.Type
	switch hookType {
	case "gitea", "github", "gitlab":
		// valid git-platform types
	default:
		hookType = "generic"
	}

	token, hint := generateWebhookToken()
	hook := models.ProjectWebhook{
		ProjectID:   project.ID,
		Name:        req.Name,
		TokenHash:   hashWebhookToken(token),
		TokenHint:   hint,
		Type:        hookType,
		CreatedByID: userID,
	}
	if err := database.DB.Create(&hook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create webhook"})
		return
	}

	// Return the plaintext token once on creation — it is not stored server-side.
	c.JSON(http.StatusCreated, gin.H{
		"id":         hook.ID,
		"name":       hook.Name,
		"type":       hook.Type,
		"token_hint": hook.TokenHint,
		"token":      token,
		"created_at": hook.CreatedAt,
	})
}

// DeleteWebhook DELETE /projects/:projectSlug/webhooks/:webhookId
func DeleteWebhook(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	hookID, err := strconv.ParseUint(c.Param("webhookId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var hook models.ProjectWebhook
	if err := database.DB.Where("id = ? AND project_id = ?", hookID, project.ID).First(&hook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	database.DB.Delete(&hook)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// RegenerateWebhookToken POST /projects/:projectSlug/webhooks/:webhookId/regenerate
func RegenerateWebhookToken(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	hookID, err := strconv.ParseUint(c.Param("webhookId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var hook models.ProjectWebhook
	if err := database.DB.Where("id = ? AND project_id = ?", hookID, project.ID).First(&hook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	token, hint := generateWebhookToken()
	database.DB.Model(&hook).Updates(map[string]interface{}{
		"token_hash": hashWebhookToken(token),
		"token_hint": hint,
		"token":      "", // clear legacy plaintext column
	})
	c.JSON(http.StatusOK, gin.H{"token": token, "token_hint": hint})
}

// MigrateWebhookTokenHashes is a one-time startup migration that hashes any
// plaintext webhook tokens left over from before hashing was introduced.
func MigrateWebhookTokenHashes() {
	var hooks []models.ProjectWebhook
	database.DB.Where("token != '' AND (token_hash = '' OR token_hash IS NULL)").Find(&hooks)
	for _, hook := range hooks {
		database.DB.Model(&hook).Updates(map[string]interface{}{
			"token_hash": hashWebhookToken(hook.Token),
			"token":      "",
		})
	}
	if len(hooks) > 0 {
		log.Printf("migrated %d webhook token(s) to hashed storage", len(hooks))
	}
}

// IncomingWebhook POST /api/v1/webhooks/:token (public)
// Body: {"text": "message", "username": "Bot Name"}
func IncomingWebhook(c *gin.Context) {
	token := c.Param("token")

	var hook models.ProjectWebhook
	// Compare by SHA-256 hash — plaintext token is never stored after hashing.
	if err := database.DB.Where("token_hash = ?", hashWebhookToken(token)).First(&hook).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	var req struct {
		Text     string `json:"text" binding:"required"`
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	botName := req.Username
	if botName == "" {
		botName = hook.Name
	}

	msg := models.ChatMessage{
		ProjectID: hook.ProjectID,
		UserID:    0,
		Body:      req.Text,
		IsBot:     true,
		BotName:   botName,
	}
	if err := database.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to post message"})
		return
	}

	appws.BroadcastToProject(hook.ProjectID, appws.Message{
		Type:    appws.TypeChatMessageCreated,
		Payload: msg,
	})

	c.JSON(http.StatusCreated, gin.H{"ok": true})
}
