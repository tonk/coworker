package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// ListActiveNews returns news items visible right now (active + within date window).
// Available to all authenticated users.
func ListActiveNews(c *gin.Context) {
	now := time.Now()
	var items []models.NewsItem
	database.DB.
		Where("active = ?", true).
		Where("start_date IS NULL OR start_date <= ?", now).
		Where("end_date IS NULL OR end_date >= ?", now).
		Order("created_at desc").
		Find(&items)
	c.JSON(http.StatusOK, items)
}

// AdminListNews returns all news items regardless of active/date state.
func AdminListNews(c *gin.Context) {
	var items []models.NewsItem
	database.DB.Unscoped().Order("created_at desc").Find(&items)
	c.JSON(http.StatusOK, items)
}

type newsInput struct {
	Title     string     `json:"title" binding:"required"`
	Text      string     `json:"text" binding:"required"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Active    *bool      `json:"active"`
}

// AdminCreateNews creates a news item.
func AdminCreateNews(c *gin.Context) {
	var in newsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	item := models.NewsItem{
		Title:     in.Title,
		Text:      in.Text,
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
		Active:    active,
	}
	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create news item"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// AdminUpdateNews updates a news item.
func AdminUpdateNews(c *gin.Context) {
	id := c.Param("id")
	var item models.NewsItem
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news item not found"})
		return
	}
	var in newsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item.Title = in.Title
	item.Text = in.Text
	item.StartDate = in.StartDate
	item.EndDate = in.EndDate
	if in.Active != nil {
		item.Active = *in.Active
	}
	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update news item"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// AdminDeleteNews permanently deletes a news item.
func AdminDeleteNews(c *gin.Context) {
	id := c.Param("id")
	var item models.NewsItem
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news item not found"})
		return
	}
	database.DB.Unscoped().Delete(&item)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
