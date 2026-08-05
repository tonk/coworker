package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
	"github.com/tonk/warmdesk/ws"
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

// CreateChatMessage POST /projects/:projectSlug/chat/messages
//
// A REST alternative to the WebSocket TypeChatSend flow, for bulk/backfill
// tools (e.g. the Ryver migration importer) that need to create historical
// messages without a live connection. created_at is optional — when given,
// it overrides the default "now" timestamp so imported history keeps its
// original time, and the message is treated as a backfill: no websocket
// broadcast or @mention notification is fired, since flooding anyone
// currently online with years-old messages would be worse than silence.
func CreateChatMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Body      string     `json:"body" binding:"required"`
		CreatedAt *time.Time `json:"created_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	chatMsg := models.ChatMessage{
		ProjectID: project.ID,
		UserID:    userID,
		Body:      req.Body,
	}
	isBackfill := req.CreatedAt != nil
	if isBackfill {
		chatMsg.CreatedAt = *req.CreatedAt
	}
	database.DB.Create(&chatMsg)
	database.DB.Preload("User").First(&chatMsg, chatMsg.ID)

	if !isBackfill {
		ws.BroadcastToProject(project.ID, ws.Message{
			Type:    ws.TypeChatMessageCreated,
			Payload: chatMsg,
		})
		if notifSvc != nil {
			go notifSvc.NotifyMentions(req.Body, userID, "project chat")
		}
	}

	c.JSON(http.StatusCreated, chatMsg)
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
