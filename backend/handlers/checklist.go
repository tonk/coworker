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

// ListChecklistItems GET /projects/:projectSlug/cards/:cardId/checklist
func ListChecklistItems(c *gin.Context) {
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

	var items []models.CardChecklistItem
	database.DB.Where("card_id = ?", cardID).Order("position asc, id asc").Find(&items)
	c.JSON(http.StatusOK, items)
}

// CreateChecklistItem POST /projects/:projectSlug/cards/:cardId/checklist
func CreateChecklistItem(c *gin.Context) {
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
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var maxPos float64
	database.DB.Model(&models.CardChecklistItem{}).
		Where("card_id = ?", cardID).
		Select("COALESCE(MAX(position), 0)").Scan(&maxPos)

	item := models.CardChecklistItem{
		CardID:   uint(cardID),
		Body:     req.Body,
		Position: maxPos + 1000,
	}
	database.DB.Create(&item)
	c.JSON(http.StatusCreated, item)
}

// UpdateChecklistItem PUT /projects/:projectSlug/cards/:cardId/checklist/:itemId
func UpdateChecklistItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
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

	var item models.CardChecklistItem
	if err := database.DB.Where("id = ? AND card_id = ?", itemID, cardID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	var req struct {
		Body        *string `json:"body"`
		IsCompleted *bool   `json:"is_completed"`
	}
	c.ShouldBindJSON(&req)

	updates := map[string]interface{}{}
	if req.Body != nil {
		updates["body"] = *req.Body
	}
	if req.IsCompleted != nil {
		updates["is_completed"] = *req.IsCompleted
	}
	if len(updates) > 0 {
		database.DB.Model(&item).Updates(updates)
	}

	database.DB.First(&item, item.ID)
	c.JSON(http.StatusOK, item)
}

// DeleteChecklistItem DELETE /projects/:projectSlug/cards/:cardId/checklist/:itemId
func DeleteChecklistItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
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

	var item models.CardChecklistItem
	if err := database.DB.Where("id = ? AND card_id = ?", itemID, cardID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	database.DB.Delete(&item)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
