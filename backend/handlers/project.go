package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	"github.com/tonk/coworker/services"
)

func ListProjects(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var members []models.ProjectMember
	database.DB.Preload("Project").Where("user_id = ?", userID).Find(&members)

	projects := make([]models.Project, 0, len(members))
	for _, m := range members {
		if m.Project.DeletedAt.Valid {
			continue
		}
		projects = append(projects, m.Project)
	}
	c.JSON(http.StatusOK, projects)
}

func CreateProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=200"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := models.Project{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Slug:        services.GenerateSlug(req.Name),
		CreatedByID: userID,
	}

	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Creator becomes owner
	member := models.ProjectMember{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "owner",
		InvitedBy: userID,
	}
	database.DB.Create(&member)

	// Default column
	database.DB.Create(&models.Column{ProjectID: project.ID, Name: "Inbox", Position: 0})

	c.JSON(http.StatusCreated, project)
}

func GetProject(c *gin.Context) {
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

	database.DB.Preload("Columns").Preload("Columns.Cards").Preload("Columns.Cards.Assignee").Preload("Columns.Cards.Labels").Preload("Labels").Preload("Members.User").First(project, project.ID)

	c.JSON(http.StatusOK, project)
}

func UpdateProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
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

	database.DB.Model(project).Updates(updates)
	database.DB.First(project, project.ID)
	c.JSON(http.StatusOK, project)
}

func DeleteProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	globalRole := middleware.GetGlobalRole(c)
	if globalRole != "admin" {
		if err := services.RequireProjectRole(project.ID, userID, globalRole, "owner"); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	database.DB.Delete(project)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
