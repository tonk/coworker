package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	"github.com/tonk/coworker/services"
)

// ChatMessageResponse wraps ChatMessage with extra computed fields.
type ChatMessageResponse struct {
	models.ChatMessage
	Attachments []models.Attachment    `json:"attachments"`
	Reactions   []models.ReactionSummary `json:"reactions"`
}

func ListChatMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "viewer"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	query := database.DB.Preload("User").Where("project_id = ? AND is_deleted = false", project.ID).Order("created_at desc").Limit(limit)
	if before := c.Query("before"); before != "" {
		if id, err := strconv.ParseUint(before, 10, 64); err == nil {
			query = query.Where("id < ?", id)
		}
	}

	var messages []models.ChatMessage
	query.Find(&messages)

	// Reverse to chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// Build response with attachments and reactions
	ids := make([]uint, len(messages))
	for i, m := range messages {
		ids[i] = m.ID
	}
	attachMap := LoadAttachments("chat_message", ids)
	reactMap := LoadReactionSummaries("chat_message", ids)

	out := make([]ChatMessageResponse, len(messages))
	for i, m := range messages {
		out[i] = ChatMessageResponse{
			ChatMessage: m,
			Attachments: attachMap[m.ID],
			Reactions:   reactMap[m.ID],
		}
		if out[i].Attachments == nil {
			out[i].Attachments = []models.Attachment{}
		}
		if out[i].Reactions == nil {
			out[i].Reactions = []models.ReactionSummary{}
		}
	}

	c.JSON(http.StatusOK, out)
}

func DeleteChatMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	msgID, err := strconv.ParseUint(c.Param("msgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var msg models.ChatMessage
	if err := database.DB.Where("id = ? AND project_id = ?", msgID, project.ID).First(&msg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	if msg.UserID != userID {
		if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	database.DB.Model(&msg).Update("is_deleted", true)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
