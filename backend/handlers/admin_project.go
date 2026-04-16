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

func AdminCreateProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Color == "" {
		req.Color = "#6366f1"
	}
	project := models.Project{
		Name:        req.Name,
		Slug:        services.GenerateSlug(req.Name),
		Description: req.Description,
		Color:       req.Color,
		CreatedByID: userID,
	}
	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "project name or slug already exists"})
		return
	}
	// Add creator as owner
	database.DB.Create(&models.ProjectMember{ProjectID: project.ID, UserID: userID, Role: "owner"})

	// Default columns from system settings
	for i, name := range getDefaultColumnNames() {
		database.DB.Create(&models.Column{ProjectID: project.ID, Name: name, Position: float64((i + 1) * 1000)})
	}

	// Default labels from system settings
	for _, def := range getDefaultLabelDefs() {
		database.DB.Create(&models.Label{ProjectID: project.ID, Name: def.Name, Color: def.Color})
	}

	database.DB.Preload("CreatedBy").First(&project, project.ID)
	c.JSON(http.StatusCreated, project)
}

type AdminProjectListItem struct {
	models.Project
	OpenCardCount int64 `json:"open_card_count"`
}

func AdminListProjects(c *gin.Context) {
	var projects []models.Project
	q := database.DB.Preload("CreatedBy")
	if c.Query("deleted") == "true" {
		q = q.Unscoped().Where("deleted_at IS NOT NULL")
	} else {
		q = q.Where("deleted_at IS NULL")
	}
	q.Find(&projects)

	result := make([]AdminProjectListItem, len(projects))
	for i, p := range projects {
		var count int64
		database.DB.Model(&models.Card{}).Where("project_id = ? AND closed = false", p.ID).Count(&count)
		result[i] = AdminProjectListItem{Project: p, OpenCardCount: count}
	}
	c.JSON(http.StatusOK, result)
}

func AdminRestoreProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result := database.DB.Unscoped().Model(&models.Project{}).Where("id = ?", id).Update("deleted_at", nil)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "restored"})
}

func AdminUpdateProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		IsArchived  *bool  `json:"is_archived"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Color != "" {
		updates["color"] = req.Color
	}
	if req.IsArchived != nil {
		updates["is_archived"] = *req.IsArchived
	}

	if err := database.DB.Model(&models.Project{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var project models.Project
	database.DB.Preload("CreatedBy").First(&project, id)
	c.JSON(http.StatusOK, project)
}

func AdminDeleteProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	database.DB.Delete(&models.Project{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
