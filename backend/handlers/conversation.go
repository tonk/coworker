package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	appws "github.com/tonk/coworker/ws"
)

// ConvMessageResponse wraps ConversationMessage with attachments and reactions.
type ConvMessageResponse struct {
	models.ConversationMessage
	Attachments []models.Attachment      `json:"attachments"`
	Reactions   []models.ReactionSummary `json:"reactions"`
}

// CreateConversation POST /conversations
// Body: {"user_ids": [2, 3], "name": "optional group name"}
// If only one other user is given and a 1-on-1 conversation already exists,
// that existing conversation is returned instead of creating a duplicate.
func CreateConversation(c *gin.Context) {
	me := middleware.GetUserID(c)

	var req struct {
		UserIDs []uint `json:"user_ids" binding:"required,min=1"`
		Name    string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Deduplicate and strip current user from the list
	seen := map[uint]bool{me: true}
	var others []uint
	for _, id := range req.UserIDs {
		if !seen[id] {
			seen[id] = true
			others = append(others, id)
		}
	}
	if len(others) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid recipients"})
		return
	}

	isGroup := len(others) > 1

	// For 1-on-1: return existing conversation if one exists
	if !isGroup {
		otherID := others[0]
		var convIDs []uint
		database.DB.Model(&models.ConversationMember{}).
			Where("user_id = ?", me).
			Pluck("conversation_id", &convIDs)
		if len(convIDs) > 0 {
			var existingID uint
			database.DB.Model(&models.ConversationMember{}).
				Where("user_id = ? AND conversation_id IN ?", otherID, convIDs).
				Pluck("conversation_id", &existingID)
			if existingID != 0 {
				var existing models.Conversation
				database.DB.Preload("Members.User").
					Where("id = ? AND is_group = false", existingID).
					First(&existing)
				if existing.ID != 0 {
					c.JSON(http.StatusOK, existing)
					return
				}
			}
		}
	}

	conv := models.Conversation{
		Name:        req.Name,
		IsGroup:     isGroup,
		CreatedByID: me,
	}
	if err := database.DB.Create(&conv).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create conversation"})
		return
	}

	now := time.Now()
	allMembers := append([]uint{me}, others...)
	for _, uid := range allMembers {
		database.DB.Create(&models.ConversationMember{
			ConversationID: conv.ID,
			UserID:         uid,
			JoinedAt:       now,
		})
	}

	database.DB.Preload("Members.User").First(&conv, conv.ID)
	c.JSON(http.StatusCreated, conv)
}

// GetConversations GET /conversations
// Returns all conversations the current user is a member of.
func GetConversations(c *gin.Context) {
	me := middleware.GetUserID(c)

	var convIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("user_id = ?", me).
		Pluck("conversation_id", &convIDs)

	var convs []models.Conversation
	if len(convIDs) > 0 {
		database.DB.Preload("Members.User").
			Where("id IN ?", convIDs).
			Order("updated_at DESC").
			Find(&convs)
	}

	c.JSON(http.StatusOK, convs)
}

// GetConversationMessages GET /conversations/:id/messages
func GetConversationMessages(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var msgs []models.ConversationMessage
	database.DB.Preload("Sender").
		Where("conversation_id = ?", convID).
		Order("created_at ASC").
		Limit(200).
		Find(&msgs)

	ids := make([]uint, len(msgs))
	for i, m := range msgs {
		ids[i] = m.ID
	}
	attachMap := LoadAttachments("conv_message", ids)
	reactMap := LoadReactionSummaries("conv_message", ids)

	out := make([]ConvMessageResponse, len(msgs))
	for i, m := range msgs {
		out[i] = ConvMessageResponse{
			ConversationMessage: m,
			Attachments:         attachMap[m.ID],
			Reactions:           reactMap[m.ID],
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

// SendConversationMessage POST /conversations/:id/messages
func SendConversationMessage(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := models.ConversationMessage{
		ConversationID: uint(convID),
		SenderID:       me,
		Body:           req.Body,
	}
	if err := database.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send"})
		return
	}

	// Bump conversation updated_at so it floats to the top of the list
	database.DB.Model(&models.Conversation{}).Where("id = ?", convID).
		Update("updated_at", time.Now())

	database.DB.Preload("Sender").First(&msg, msg.ID)

	if notifSvc != nil {
		go notifSvc.NotifyNewDM(msg, msg.Sender)
		go notifSvc.NotifyMentions(msg.Body, me, "direct message")
	}

	c.JSON(http.StatusCreated, msg)
}

// DeleteConversationMessage DELETE /conversations/:id/messages/:msgId
func DeleteConversationMessage(c *gin.Context) {
	me := middleware.GetUserID(c)
	msgID, _ := strconv.Atoi(c.Param("msgId"))

	var msg models.ConversationMessage
	if err := database.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if msg.SenderID != me {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your message"})
		return
	}

	database.DB.Model(&msg).Update("is_deleted", true)

	// Broadcast to all conversation members
	var memberIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", c.Param("id")).
		Pluck("user_id", &memberIDs)
	for _, uid := range memberIDs {
		appws.BroadcastToUser(uid, appws.Message{
			Type: appws.TypeDMMessageDeleted,
			Payload: map[string]interface{}{
				"conversation_id": msg.ConversationID,
				"id":              msg.ID,
			},
		})
	}

	c.JSON(http.StatusOK, msg)
}

// AddConversationMember POST /conversations/:id/members
func AddConversationMember(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if isMember(uint(convID), req.UserID) {
		c.JSON(http.StatusOK, gin.H{"message": "already a member"})
		return
	}

	database.DB.Create(&models.ConversationMember{
		ConversationID: uint(convID),
		UserID:         req.UserID,
		JoinedAt:       time.Now(),
	})

	// Promote to group if now more than 2 members
	var count int64
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", convID).Count(&count)
	if count > 2 {
		database.DB.Model(&models.Conversation{}).
			Where("id = ?", convID).Update("is_group", true)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "added"})
}

// EditConversationMessage PATCH /conversations/:id/messages/:msgId
func EditConversationMessage(c *gin.Context) {
	me := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))
	msgID, _ := strconv.Atoi(c.Param("msgId"))

	if !isMember(uint(convID), me) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var msg models.ConversationMessage
	if err := database.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if msg.SenderID != me {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your message"})
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Model(&msg).Updates(map[string]interface{}{"body": req.Body, "is_edited": true})

	// Broadcast to all conversation members
	var memberIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", convID).
		Pluck("user_id", &memberIDs)
	for _, uid := range memberIDs {
		appws.BroadcastToUser(uid, appws.Message{
			Type: appws.TypeDMMessageUpdated,
			Payload: map[string]interface{}{
				"conversation_id": convID,
				"id":              msg.ID,
				"body":            req.Body,
				"is_edited":       true,
			},
		})
	}

	database.DB.Preload("Sender").First(&msg, msg.ID)
	c.JSON(http.StatusOK, msg)
}

func isMember(convID, userID uint) bool {
	var m models.ConversationMember
	return database.DB.Where("conversation_id = ? AND user_id = ?", convID, userID).
		First(&m).Error == nil
}
