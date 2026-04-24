package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

type SearchResult struct {
	Type string      `json:"type"`
	Item interface{} `json:"item"`
}

// GlobalSearch godoc
// @Summary      Search cards, topics, and projects
// @Tags         search
// @Produce      json
// @Security     BearerAuth
// @Param        q query string true "Search query (min 2 characters)"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Router       /search [get]
func GlobalSearch(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)
	q := c.Query("q")
	if len(q) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too short (minimum 3 characters)"})
		return
	}
	// Strip SQL LIKE wildcards to prevent deliberate full-table scans.
	q = strings.NewReplacer("%", "", "_", "").Replace(q)
	if len(q) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too short after sanitization"})
		return
	}
	pattern := "%" + q + "%"

	// Build project scope:
	// - global admins can search all projects
	// - regular users are restricted to projects they are a member of
	var memberProjectIDs []uint
	if globalRole == "admin" {
		database.DB.Model(&models.Project{}).Pluck("id", &memberProjectIDs)
	} else {
		database.DB.Model(&models.ProjectMember{}).
			Where("user_id = ?", userID).
			Pluck("project_id", &memberProjectIDs)
	}

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
