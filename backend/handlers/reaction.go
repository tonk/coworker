package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	"github.com/tonk/coworker/services"
	appws "github.com/tonk/coworker/ws"
)

var validEmojis = map[string]bool{
	"👍": true, "❤️": true, "😂": true, "😮": true, "😢": true, "👎": true,
}

// LoadReactionSummaries fetches reactions for a set of owner IDs and returns grouped summaries.
func LoadReactionSummaries(ownerType string, ownerIDs []uint) map[uint][]models.ReactionSummary {
	result := make(map[uint][]models.ReactionSummary)
	if len(ownerIDs) == 0 {
		return result
	}
	var reactions []models.MessageReaction
	database.DB.Where("owner_type = ? AND owner_id IN ?", ownerType, ownerIDs).Find(&reactions)

	// Group by owner_id then by emoji
	type key struct{ ownerID uint; emoji string }
	counts := make(map[key][]uint)
	for _, r := range reactions {
		k := key{r.OwnerID, r.Emoji}
		counts[k] = append(counts[k], r.UserID)
	}
	for k, userIDs := range counts {
		result[k.ownerID] = append(result[k.ownerID], models.ReactionSummary{
			Emoji:   k.emoji,
			Count:   len(userIDs),
			UserIDs: userIDs,
		})
	}
	return result
}

func buildReactionSummary(ownerType string, ownerID uint) []models.ReactionSummary {
	m := LoadReactionSummaries(ownerType, []uint{ownerID})
	if s, ok := m[ownerID]; ok {
		return s
	}
	return []models.ReactionSummary{}
}

// ToggleChatReaction POST /projects/:projectSlug/chat/messages/:msgId/reactions
func ToggleChatReaction(c *gin.Context) {
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
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "viewer"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validEmojis[req.Emoji] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid emoji"})
		return
	}

	var existing models.MessageReaction
	err2 := database.DB.Where("owner_type = ? AND owner_id = ? AND user_id = ? AND emoji = ?",
		"chat_message", msgID, userID, req.Emoji).First(&existing).Error
	if err2 == nil {
		database.DB.Delete(&existing)
	} else {
		database.DB.Create(&models.MessageReaction{
			OwnerType: "chat_message",
			OwnerID:   uint(msgID),
			UserID:    userID,
			Emoji:     req.Emoji,
		})
	}

	summaries := buildReactionSummary("chat_message", uint(msgID))
	appws.BroadcastToProject(project.ID, appws.Message{
		Type: appws.TypeChatReactionUpdated,
		Payload: map[string]interface{}{
			"message_id": msgID,
			"reactions":  summaries,
		},
	})
	c.JSON(http.StatusOK, gin.H{"message_id": msgID, "reactions": summaries})
}

// ToggleConvReaction POST /conversations/:id/messages/:msgId/reactions
func ToggleConvReaction(c *gin.Context) {
	userID := middleware.GetUserID(c)
	convID, _ := strconv.Atoi(c.Param("id"))
	msgID, err := strconv.ParseUint(c.Param("msgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if !isMember(uint(convID), userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var req struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validEmojis[req.Emoji] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid emoji"})
		return
	}

	var existing models.MessageReaction
	err2 := database.DB.Where("owner_type = ? AND owner_id = ? AND user_id = ? AND emoji = ?",
		"conv_message", msgID, userID, req.Emoji).First(&existing).Error
	if err2 == nil {
		database.DB.Delete(&existing)
	} else {
		database.DB.Create(&models.MessageReaction{
			OwnerType: "conv_message",
			OwnerID:   uint(msgID),
			UserID:    userID,
			Emoji:     req.Emoji,
		})
	}

	summaries := buildReactionSummary("conv_message", uint(msgID))

	// Broadcast to all conversation members
	var memberIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", convID).
		Pluck("user_id", &memberIDs)
	for _, uid := range memberIDs {
		appws.BroadcastToUser(uid, appws.Message{
			Type: appws.TypeDMReactionUpdated,
			Payload: map[string]interface{}{
				"conversation_id": convID,
				"message_id":      msgID,
				"reactions":       summaries,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"message_id": msgID, "reactions": summaries})
}
