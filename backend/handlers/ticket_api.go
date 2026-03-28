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
	"gorm.io/gorm"
)

// TicketAdd creates a card via API key authentication.
// POST /api/v1/ticket/:projectSlug/cards
// Body: { "title": "...", "description": "...", "column_id": 1 }
func TicketAdd(c *gin.Context) {
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
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		ColumnID    uint   `json:"column_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var col models.Column
	if err := database.DB.Where("id = ? AND project_id = ?", req.ColumnID, project.ID).First(&col).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "column not found in project"})
		return
	}

	var maxPos struct{ Pos float64 }
	database.DB.Model(&models.Card{}).Select("COALESCE(MAX(position), 0) as pos").Where("column_id = ?", col.ID).Scan(&maxPos)

	// Atomically increment the project's card counter
	database.DB.Model(&models.Project{}).Where("id = ?", project.ID).
		UpdateColumn("card_counter", gorm.Expr("card_counter + 1"))
	var updatedProject models.Project
	database.DB.Select("card_counter").First(&updatedProject, project.ID)

	card := models.Card{
		ColumnID:    col.ID,
		ProjectID:   project.ID,
		Title:       req.Title,
		Description: req.Description,
		Position:    maxPos.Pos + 1000,
		CreatedByID: userID,
		CardNumber:  updatedProject.CardCounter,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	database.DB.Preload("CreatedBy").Preload("Assignee").Preload("Labels").Preload("Tags").First(&card, card.ID)

	appws.BroadcastToProject(project.ID, appws.Message{Type: appws.TypeBoardCardCreated, Payload: card})
	c.JSON(http.StatusCreated, card)
}

// TicketComment adds a comment to a card via API key authentication.
// POST /api/v1/ticket/:projectSlug/cards/:cardId/comments
// Body: { "body": "..." }
func TicketComment(c *gin.Context) {
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

	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment := models.CardComment{CardID: card.ID, UserID: userID, Body: req.Body}
	if err := database.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	database.DB.Preload("User").First(&comment, comment.ID)

	appws.BroadcastToProject(project.ID, appws.Message{Type: appws.TypeBoardCommentCreated, Payload: comment})
	c.JSON(http.StatusCreated, comment)
}

// TicketMove moves a card to another column via API key authentication.
// PATCH /api/v1/ticket/:projectSlug/cards/:cardId/move
// Body: { "column_id": 2, "position": 1000 }
func TicketMove(c *gin.Context) {
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

	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	var req struct {
		ColumnID uint    `json:"column_id" binding:"required"`
		Position float64 `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var col models.Column
	if err := database.DB.Where("id = ? AND project_id = ?", req.ColumnID, project.ID).First(&col).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "column not found in project"})
		return
	}

	pos := req.Position
	if pos == 0 {
		var maxPos struct{ Pos float64 }
		database.DB.Model(&models.Card{}).Select("COALESCE(MAX(position), 0) as pos").Where("column_id = ?", col.ID).Scan(&maxPos)
		pos = maxPos.Pos + 1000
	}

	oldColumnID := card.ColumnID
	database.DB.Model(&card).Updates(map[string]interface{}{"column_id": col.ID, "position": pos})

	if oldColumnID != col.ID {
		database.DB.Create(&models.CardHistory{
			CardID:       card.ID,
			UserID:       userID,
			FromColumnID: oldColumnID,
			ToColumnID:   col.ID,
		})
	}

	appws.BroadcastToProject(project.ID, appws.Message{
		Type: appws.TypeBoardCardMoved,
		Payload: map[string]interface{}{
			"card_id":       card.ID,
			"from_column_id": oldColumnID,
			"to_column_id":  col.ID,
			"position":      pos,
		},
	})

	database.DB.Preload("CreatedBy").Preload("Assignee").Preload("Labels").First(&card, card.ID)
	c.JSON(http.StatusOK, card)
}
