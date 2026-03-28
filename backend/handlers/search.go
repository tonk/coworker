package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
)

type SearchResult struct {
	Type string      `json:"type"`
	Item interface{} `json:"item"`
}

// GlobalSearch GET /api/v1/search?q=<query>
func GlobalSearch(c *gin.Context) {
	userID := middleware.GetUserID(c)
	q := c.Query("q")
	if len(q) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too short"})
		return
	}
	pattern := "%" + q + "%"

	// Get project IDs the user is a member of
	var memberProjectIDs []uint
	database.DB.Model(&models.ProjectMember{}).
		Where("user_id = ?", userID).
		Pluck("project_id", &memberProjectIDs)

	var results []SearchResult

	// Cards
	if len(memberProjectIDs) > 0 {
		var cards []models.Card
		database.DB.Preload("Assignee").
			Where("project_id IN ? AND (title LIKE ? OR description LIKE ?)", memberProjectIDs, pattern, pattern).
			Limit(20).Find(&cards)
		for _, c := range cards {
			results = append(results, SearchResult{Type: "card", Item: c})
		}
	}

	// Chat messages
	if len(memberProjectIDs) > 0 {
		var msgs []models.ChatMessage
		database.DB.Preload("User").
			Where("project_id IN ? AND body LIKE ? AND is_deleted = false AND is_bot = false", memberProjectIDs, pattern).
			Limit(20).Find(&msgs)
		for _, m := range msgs {
			results = append(results, SearchResult{Type: "chat_message", Item: m})
		}
	}

	// Conversation messages
	var convIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("user_id = ?", userID).
		Pluck("conversation_id", &convIDs)
	if len(convIDs) > 0 {
		var dms []models.ConversationMessage
		database.DB.Preload("Sender").
			Where("conversation_id IN ? AND body LIKE ? AND is_deleted = false", convIDs, pattern).
			Limit(20).Find(&dms)
		for _, m := range dms {
			results = append(results, SearchResult{Type: "dm_message", Item: m})
		}
	}

	// Card comments
	if len(memberProjectIDs) > 0 {
		var cardIDs []uint
		database.DB.Model(&models.Card{}).
			Where("project_id IN ?", memberProjectIDs).
			Pluck("id", &cardIDs)
		if len(cardIDs) > 0 {
			var comments []models.CardComment
			database.DB.Preload("User").
				Where("card_id IN ? AND body LIKE ?", cardIDs, pattern).
				Limit(20).Find(&comments)
			for _, cm := range comments {
				results = append(results, SearchResult{Type: "card_comment", Item: cm})
			}
		}
	}

	if results == nil {
		results = []SearchResult{}
	}
	c.JSON(http.StatusOK, results)
}
