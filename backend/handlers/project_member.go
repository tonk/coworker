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

func ListMembers(c *gin.Context) {
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

	var members []models.ProjectMember
	database.DB.Preload("User").Where("project_id = ?", project.ID).Find(&members)
	c.JSON(http.StatusOK, members)
}

func InviteMember(c *gin.Context) {
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
		Login string `json:"login" binding:"required"` // email or username
		Role  string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	role := req.Role
	if role == "" {
		role = "member"
	}
	if role != "owner" && role != "member" && role != "viewer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	login := strings.ToLower(req.Login)
	var invitee models.User
	if err := database.DB.Where("email = ? OR username = ?", login, req.Login).First(&invitee).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var existing models.ProjectMember
	if err := database.DB.Where("project_id = ? AND user_id = ?", project.ID, invitee.ID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already a member"})
		return
	}

	member := models.ProjectMember{
		ProjectID: project.ID,
		UserID:    invitee.ID,
		Role:      role,
		InvitedBy: userID,
	}
	database.DB.Create(&member)
	database.DB.Preload("User").First(&member, member.ID)
	c.JSON(http.StatusCreated, member)
}

func UpdateMemberRole(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

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
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var member models.ProjectMember
	if err := database.DB.Where("project_id = ? AND user_id = ?", project.ID, targetID).First(&member).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}

	database.DB.Model(&member).Update("role", req.Role)
	c.JSON(http.StatusOK, member)
}

func RemoveMember(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	// Allow self-removal or owner
	globalRole := middleware.GetGlobalRole(c)
	if uint(targetID) != userID {
		if err := services.RequireProjectRole(project.ID, userID, globalRole, "owner"); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	database.DB.Where("project_id = ? AND user_id = ?", project.ID, targetID).Delete(&models.ProjectMember{})
	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}
