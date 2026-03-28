package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	"github.com/tonk/coworker/services"
	"github.com/tonk/coworker/ws"
)

// AddCardTag POST /projects/:projectSlug/cards/:cardId/tags
func AddCardTag(c *gin.Context) {
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
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize: lowercase, trim spaces, strip leading #
	name := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(req.Name, "#")))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag name is empty"})
		return
	}

	// Verify card belongs to project
	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	tag := models.CardTag{CardID: card.ID, Name: name}
	if err := database.DB.FirstOrCreate(&tag, tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Reload all tags for the card and broadcast
	var tags []models.CardTag
	database.DB.Where("card_id = ?", card.ID).Find(&tags)
	ws.BroadcastToProject(project.ID, ws.Message{
		Type:    ws.TypeBoardCardUpdated,
		Payload: map[string]interface{}{"id": card.ID, "tags": tags},
	})

	c.JSON(http.StatusCreated, tag)
}

// RemoveCardTag DELETE /projects/:projectSlug/cards/:cardId/tags/:tagId
func RemoveCardTag(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}
	tagID, err := strconv.ParseUint(c.Param("tagId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
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

	database.DB.Where("id = ? AND card_id = ?", tagID, cardID).Delete(&models.CardTag{})

	var tags []models.CardTag
	database.DB.Where("card_id = ?", cardID).Find(&tags)
	ws.BroadcastToProject(project.ID, ws.Message{
		Type:    ws.TypeBoardCardUpdated,
		Payload: map[string]interface{}{"id": uint(cardID), "tags": tags},
	})

	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}
