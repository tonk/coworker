package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

func populateEpic(e *models.Epic) {
	var total, done int64
	database.DB.Model(&models.Card{}).
		Where("epic_id = ? AND parent_card_id IS NULL AND deleted_at IS NULL", e.ID).Count(&total)
	if total > 0 {
		database.DB.Model(&models.Card{}).
			Where("epic_id = ? AND parent_card_id IS NULL AND closed = true AND deleted_at IS NULL", e.ID).Count(&done)
	}
	e.CardCount = int(total)
	e.DoneCount = int(done)
}

// ListEpics GET /projects/:projectSlug/epics
func ListEpics(c *gin.Context) {
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
	var epics []models.Epic
	database.DB.Where("project_id = ?", project.ID).Order("position asc, created_at asc").Find(&epics)
	for i := range epics {
		populateEpic(&epics[i])
	}
	c.JSON(http.StatusOK, epics)
}

// CreateEpic POST /projects/:projectSlug/epics
func CreateEpic(c *gin.Context) {
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
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Color       string  `json:"color"`
		Position    float64 `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	color := req.Color
	if color == "" {
		color = "#6366f1"
	}
	epic := models.Epic{
		ProjectID:   project.ID,
		Name:        req.Name,
		Description: req.Description,
		Color:       color,
		Position:    req.Position,
		Status:      "open",
	}
	if err := database.DB.Create(&epic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create epic"})
		return
	}
	populateEpic(&epic)
	c.JSON(http.StatusCreated, epic)
}

// UpdateEpic PUT /projects/:projectSlug/epics/:epicId
func UpdateEpic(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	epicID, err := strconv.ParseUint(c.Param("epicId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid epic id"})
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
	var epic models.Epic
	if err := database.DB.Where("id = ? AND project_id = ?", epicID, project.ID).First(&epic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "epic not found"})
		return
	}
	var req struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Color       *string  `json:"color"`
		Status      *string  `json:"status"`
		Position    *float64 `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Position != nil {
		updates["position"] = *req.Position
	}
	if len(updates) > 0 {
		database.DB.Model(&epic).Updates(updates)
	}
	database.DB.First(&epic, epic.ID)
	populateEpic(&epic)
	c.JSON(http.StatusOK, epic)
}

// DeleteEpic DELETE /projects/:projectSlug/epics/:epicId
func DeleteEpic(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	epicID, err := strconv.ParseUint(c.Param("epicId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid epic id"})
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
	var epic models.Epic
	if err := database.DB.Where("id = ? AND project_id = ?", epicID, project.ID).First(&epic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "epic not found"})
		return
	}
	// Unlink cards before deleting so they become epic-less, not orphaned
	database.DB.Model(&models.Card{}).Where("epic_id = ?", epicID).Update("epic_id", nil)
	database.DB.Delete(&epic)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ReorderEpics PATCH /projects/:projectSlug/epics/reorder
func ReorderEpics(c *gin.Context) {
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
	var req []struct {
		ID       uint    `json:"id"`
		Position float64 `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	for _, item := range req {
		database.DB.Model(&models.Epic{}).
			Where("id = ? AND project_id = ?", item.ID, project.ID).
			Update("position", item.Position)
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
