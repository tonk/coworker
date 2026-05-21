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
	"github.com/tonk/warmdesk/services"
	"github.com/tonk/warmdesk/ws"
)

// firstSentence returns the leading sentence of s (split on . ! ? or newline).
func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	for i, r := range s {
		if i > 0 && (r == '.' || r == '!' || r == '?' || r == '\n') {
			return strings.TrimSpace(s[:i+1])
		}
	}
	return s
}

// recalcCardTimeSpent recomputes card.time_spent_minutes as the sum of its comments.
func recalcCardTimeSpent(cardID uint) {
	var total int
	database.DB.Model(&models.CardComment{}).
		Where("card_id = ? AND deleted_at IS NULL", cardID).
		Select("COALESCE(SUM(time_spent_minutes), 0)").
		Scan(&total)
	database.DB.Model(&models.Card{}).Where("id = ?", cardID).
		UpdateColumn("time_spent_minutes", total)
}

// autoCreateTimeEntry creates a TimeEntry when the user has time-tracking enabled.
// Returns the new entry ID, or 0 if skipped.
func autoCreateTimeEntry(userID, cardID uint, minutes int, description string) uint {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil || !user.TimeTrackingEnabled {
		return 0
	}

	var card models.Card
	if err := database.DB.First(&card, cardID).Error; err != nil {
		return 0
	}

	var project models.Project
	if err := database.DB.First(&project, card.ProjectID).Error; err != nil {
		return 0
	}

	desc := description
	if len(desc) > 255 {
		desc = desc[:255]
	}

	today := time.Now().Truncate(24 * time.Hour)
	entry := models.TimeEntry{
		UserID:      userID,
		ProjectID:   &card.ProjectID,
		CustomerID:  project.CustomerID,
		Date:        today,
		Minutes:     minutes,
		Description: desc,
	}
	database.DB.Create(&entry)
	return entry.ID
}

func ListComments(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
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

	var comments []models.CardComment
	database.DB.Preload("User").Where("card_id = ?", cardID).Order("created_at asc").Find(&comments)
	c.JSON(http.StatusOK, comments)
}

func CreateComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

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
		Body             string `json:"body" binding:"required"`
		TimeSpentMinutes int    `json:"time_spent_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	comment := models.CardComment{
		CardID:           uint(cardID),
		UserID:           userID,
		Body:             req.Body,
		TimeSpentMinutes: req.TimeSpentMinutes,
	}

	if req.TimeSpentMinutes > 0 {
		desc := firstSentence(req.Body)
		if entryID := autoCreateTimeEntry(userID, uint(cardID), req.TimeSpentMinutes, desc); entryID > 0 {
			comment.TimeEntryID = &entryID
		}
	}

	database.DB.Create(&comment)
	database.DB.Create(&models.CardHistory{CardID: uint(cardID), UserID: userID, EventType: "comment_added"})
	if req.TimeSpentMinutes > 0 {
		recalcCardTimeSpent(uint(cardID))
	}
	database.DB.Preload("User").First(&comment, comment.ID)

	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardCommentCreated, Payload: comment})

	if notifSvc != nil {
		go notifSvc.NotifyMentions(req.Body, userID, "card comment")
	}

	c.JSON(http.StatusCreated, comment)
}

func UpdateComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 64)
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

	var comment models.CardComment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	if comment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "can only edit your own comments"})
		return
	}

	var req struct {
		Body             string `json:"body" binding:"required"`
		TimeSpentMinutes int    `json:"time_spent_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	oldMinutes := comment.TimeSpentMinutes
	desc := firstSentence(req.Body)
	newEntryID := comment.TimeEntryID

	switch {
	case req.TimeSpentMinutes > 0 && comment.TimeEntryID != nil:
		// Update the existing time entry
		database.DB.Model(&models.TimeEntry{}).Where("id = ?", *comment.TimeEntryID).
			Updates(map[string]interface{}{"minutes": req.TimeSpentMinutes, "description": desc})
	case req.TimeSpentMinutes > 0 && comment.TimeEntryID == nil:
		// Create a new time entry
		if entryID := autoCreateTimeEntry(userID, comment.CardID, req.TimeSpentMinutes, desc); entryID > 0 {
			newEntryID = &entryID
		}
	case req.TimeSpentMinutes == 0 && comment.TimeEntryID != nil:
		// Remove the time entry
		database.DB.Delete(&models.TimeEntry{}, *comment.TimeEntryID)
		newEntryID = nil
	}

	database.DB.Model(&comment).Updates(map[string]interface{}{
		"body":               req.Body,
		"is_edited":          true,
		"time_spent_minutes": req.TimeSpentMinutes,
		"time_entry_id":      newEntryID,
	})

	if req.TimeSpentMinutes != oldMinutes {
		recalcCardTimeSpent(comment.CardID)
	}
	database.DB.Preload("User").First(&comment, comment.ID)

	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardCommentUpdated, Payload: comment})
	c.JSON(http.StatusOK, comment)
}

func DeleteComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 64)
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

	var comment models.CardComment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	if comment.UserID != userID {
		role := services.GetMemberRole(project.ID, userID)
		if role != "owner" && middleware.GetGlobalRole(c) != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	if comment.TimeEntryID != nil {
		database.DB.Delete(&models.TimeEntry{}, *comment.TimeEntryID)
	}
	hadTime := comment.TimeSpentMinutes > 0

	database.DB.Delete(&comment)
	if hadTime {
		recalcCardTimeSpent(comment.CardID)
	}

	ws.BroadcastToProject(project.ID, ws.Message{
		Type:    ws.TypeBoardCommentDeleted,
		Payload: map[string]uint{"comment_id": uint(commentID), "card_id": comment.CardID},
	})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
