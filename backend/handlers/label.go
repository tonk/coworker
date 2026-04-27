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

func ListLabels(c *gin.Context) {
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

	var labels []models.Label
	database.DB.Where("project_id = ?", project.ID).Find(&labels)
	c.JSON(http.StatusOK, labels)
}

func CreateLabel(c *gin.Context) {
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
		Name  string `json:"name" binding:"required"`
		Color string `json:"color" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	label := models.Label{
		ProjectID: project.ID,
		Name:      req.Name,
		Color:     req.Color,
	}
	database.DB.Create(&label)
	c.JSON(http.StatusCreated, label)
}

func UpdateLabel(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	labelID, err := strconv.ParseUint(c.Param("labelId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
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

	var label models.Label
	if err := database.DB.Where("id = ? AND project_id = ?", labelID, project.ID).First(&label).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "label not found"})
		return
	}

	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	c.ShouldBindJSON(&req)

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Color != "" {
		updates["color"] = req.Color
	}

	database.DB.Model(&label).Updates(updates)
	c.JSON(http.StatusOK, label)
}

func DeleteLabel(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	labelID, err := strconv.ParseUint(c.Param("labelId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
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

	database.DB.Where("id = ? AND project_id = ?", labelID, project.ID).Delete(&models.Label{})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
