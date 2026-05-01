package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
	"github.com/tonk/warmdesk/ws"
)

func ListColumns(c *gin.Context) {
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

	var columns []models.Column
	database.DB.Where("project_id = ? AND deleted_at IS NULL", project.ID).Order("position asc").Find(&columns)
	c.JSON(http.StatusOK, columns)
}

func CreateColumn(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "admin"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required"`
		Color    string `json:"color"`
		WIPLimit *int   `json:"wip_limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.WIPLimit != nil && *req.WIPLimit < 1 {
		req.WIPLimit = nil
	}

	// Find max position
	var maxPos float64
	database.DB.Model(&models.Column{}).Where("project_id = ?", project.ID).Select("COALESCE(MAX(position), 0)").Scan(&maxPos)

	col := models.Column{
		ProjectID: project.ID,
		Name:      req.Name,
		Color:     req.Color,
		WIPLimit:  req.WIPLimit,
		Position:  maxPos + 1000,
	}
	database.DB.Create(&col)

	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardColumnCreated, Payload: col})
	c.JSON(http.StatusCreated, col)
}

func UpdateColumn(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	colID, err := strconv.ParseUint(c.Param("columnId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "admin"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var col models.Column
	if err := database.DB.Where("id = ? AND project_id = ?", colID, project.ID).First(&col).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "column not found"})
		return
	}

	var req struct {
		Name          string `json:"name"`
		Color         string `json:"color"`
		WIPLimit      *int   `json:"wip_limit"`
		ClearWIPLimit bool   `json:"wip_limit_clear"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Color != "" {
		updates["color"] = req.Color
	}
	if req.ClearWIPLimit {
		updates["wip_limit"] = nil
	} else if req.WIPLimit != nil {
		if *req.WIPLimit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "wip_limit must be at least 1"})
			return
		}
		updates["wip_limit"] = *req.WIPLimit
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, col)
		return
	}

	database.DB.Model(&col).Updates(updates)
	if err := database.DB.Where("id = ? AND project_id = ?", col.ID, project.ID).First(&col).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardColumnUpdated, Payload: col})
	c.JSON(http.StatusOK, col)
}

func DeleteColumn(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	colID, err := strconv.ParseUint(c.Param("columnId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "admin"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var col models.Column
	if err := database.DB.Where("id = ? AND project_id = ?", colID, project.ID).First(&col).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "column not found"})
		return
	}

	var cardCount int64
	database.DB.Model(&models.Card{}).Where("column_id = ? AND deleted_at IS NULL", colID).Count(&cardCount)
	if cardCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "column has cards, move them first"})
		return
	}

	database.DB.Delete(&col)
	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardColumnDeleted, Payload: map[string]uint{"column_id": uint(colID)}})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func ReorderColumns(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "admin"); err != nil {
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
		database.DB.Model(&models.Column{}).Where("id = ? AND project_id = ?", item.ID, project.ID).Update("position", item.Position)
	}

	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardColumnsReordered, Payload: req})
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
