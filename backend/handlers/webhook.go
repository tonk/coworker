package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	"github.com/tonk/coworker/services"
	appws "github.com/tonk/coworker/ws"
)

func generateWebhookToken() (token, hint string) {
	b := make([]byte, 32)
	rand.Read(b)
	token = hex.EncodeToString(b)
	hint = token[len(token)-8:]
	return
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hookType := req.Type
	if hookType != "gitea" {
		hookType = "generic"
	}

	token, hint := generateWebhookToken()
	hook := models.ProjectWebhook{
		ProjectID:   project.ID,
		Name:        req.Name,
		Token:       token,
		TokenHint:   hint,
		Type:        hookType,
		CreatedByID: userID,
	}
	if err := database.DB.Create(&hook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create webhook"})
		return
	}

	// Return the token once on creation
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
	database.DB.Model(&hook).Updates(map[string]interface{}{"token": token, "token_hint": hint})
	c.JSON(http.StatusOK, gin.H{"token": token, "token_hint": hint})
}

// IncomingWebhook POST /api/v1/webhooks/:token (public)
// Body: {"text": "message", "username": "Bot Name"}
func IncomingWebhook(c *gin.Context) {
	token := c.Param("token")

	var hook models.ProjectWebhook
	if err := database.DB.Where("token = ?", token).First(&hook).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	var req struct {
		Text     string `json:"text" binding:"required"`
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
