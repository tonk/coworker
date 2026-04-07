package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
	"gorm.io/gorm"
)

// ListSubCards returns all sub-cards for a given parent card.
// Sub-cards are ordered by creation time (oldest first).
func ListSubCards(c *gin.Context) {
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

	// Verify parent card exists and belongs to the project
	var parent models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&parent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	var subCards []models.Card
	database.DB.Preload("Labels").Preload("Assignee").Preload("Assignees").
		Where("parent_card_id = ?", cardID).
		Order("created_at asc, id asc").Find(&subCards)

	c.JSON(http.StatusOK, subCards)
}

// CreateSubCard creates a sub-card under a parent card.
// The sub-card inherits the parent's project and column.
func CreateSubCard(c *gin.Context) {
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

	var parent models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&parent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "parent card not found"})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required,min=1"`
		Description string `json:"description"`
		AssigneeID  *uint  `json:"assignee_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Increment project card counter for a unique card number
	database.DB.Model(&models.Project{}).Where("id = ?", project.ID).
		UpdateColumn("card_counter", gorm.Expr("card_counter + 1"))
	var updatedProject models.Project
	database.DB.Select("card_counter").First(&updatedProject, project.ID)

	parentID := uint(cardID)
	subCard := models.Card{
		ColumnID:     parent.ColumnID,
		ProjectID:    project.ID,
		ParentCardID: &parentID,
		Title:        req.Title,
		Description:  req.Description,
		Priority:     "none",
		AssigneeID:   req.AssigneeID,
		CreatedByID:  userID,
		Position:     0, // sub-cards ordered by created_at, position unused
		CardNumber:   updatedProject.CardCounter,
	}
	if err := database.DB.Create(&subCard).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	database.DB.Preload("Labels").Preload("Assignee").First(&subCard, subCard.ID)

	c.JSON(http.StatusCreated, subCard)
}
