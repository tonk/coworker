package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	"github.com/tonk/coworker/services"
)

func AddWatcher(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}
	watcherID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
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

	// Ensure the watcher is a member of the project
	if err := services.RequireProjectRole(project.ID, uint(watcherID), "", "viewer"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user is not a project member"})
		return
	}

	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	var watcher models.User
	if err := database.DB.First(&watcher, watcherID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	database.DB.Model(&card).Association("Watchers").Append(&watcher)
	c.JSON(http.StatusOK, watcher)
}

func RemoveWatcher(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}
	watcherID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
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

	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	var watcher models.User
	if err := database.DB.First(&watcher, watcherID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	database.DB.Model(&card).Association("Watchers").Delete(&watcher)
	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}
