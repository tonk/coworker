package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

// ListTimeTrackingProjects returns time-tracking-only projects visible to the
// current user: their own plus any created by an admin.
func ListTimeTrackingProjects(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var projects []models.Project
	if globalRole == "admin" {
		database.DB.Where("time_tracking_only = true AND deleted_at IS NULL").
			Order("name asc").Find(&projects)
	} else {
		database.DB.Raw(`
			SELECT p.* FROM projects p
			JOIN users u ON u.id = p.created_by_id
			WHERE p.time_tracking_only = true
			  AND p.deleted_at IS NULL
			  AND (p.created_by_id = ? OR u.global_role = 'admin')
			ORDER BY p.name ASC`, userID).Scan(&projects)
	}
	c.JSON(http.StatusOK, projects)
}

// CreateTimeTrackingProject creates a new time-tracking-only project owned by
// the current user. No board columns or members are created.
func CreateTimeTrackingProject(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Name                string `json:"name"`
		Color               string `json:"color"`
		UndeclarableMinutes int    `json:"undeclarable_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if len(name) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name too long"})
		return
	}
	if req.UndeclarableMinutes < 0 {
		req.UndeclarableMinutes = 0
	}

	project := models.Project{
		Name:                name,
		Color:               req.Color,
		Slug:                services.GenerateSlug(name),
		KeyPrefix:           services.GenerateKeyPrefix(name),
		BoardType:           "kanban",
		TimeTrackingOnly:    true,
		UndeclarableMinutes: req.UndeclarableMinutes,
		CreatedByID:         userID,
	}
	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, project)
}

// UpdateTimeTrackingProject updates the name/color of a time-tracking-only
// project. Users can only update their own; admins can update any.
func UpdateTimeTrackingProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var project models.Project
	if err := database.DB.Where("id = ? AND time_tracking_only = true AND deleted_at IS NULL", id).
		First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if globalRole != "admin" && project.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Name                string `json:"name"`
		Color               string `json:"color"`
		UndeclarableMinutes int    `json:"undeclarable_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if len(name) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name too long"})
		return
	}
	if req.UndeclarableMinutes < 0 {
		req.UndeclarableMinutes = 0
	}

	project.Name = name
	project.Color = req.Color
	project.UndeclarableMinutes = req.UndeclarableMinutes
	if err := database.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, project)
}

// DeleteTimeTrackingProject soft-deletes a time-tracking-only project.
// Users can only delete their own; admins can delete any.
func DeleteTimeTrackingProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var project models.Project
	if err := database.DB.Where("id = ? AND time_tracking_only = true AND deleted_at IS NULL", id).
		First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if globalRole != "admin" && project.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	database.DB.Delete(&project)
	c.JSON(http.StatusNoContent, nil)
}
