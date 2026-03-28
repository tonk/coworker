package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	"github.com/tonk/coworker/services"
	"github.com/tonk/coworker/ws"
	"gorm.io/gorm"
)

func ListCards(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	colID, err := strconv.ParseUint(c.Param("columnId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid column id"})
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

	var cards []models.Card
	database.DB.Preload("Labels").Preload("Assignee").Preload("Tags").Where("column_id = ? AND project_id = ?", colID, project.ID).Order("position asc").Find(&cards)
	c.JSON(http.StatusOK, cards)
}

func CreateCard(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	colID, err := strconv.ParseUint(c.Param("columnId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid column id"})
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
		Title       string          `json:"title" binding:"required"`
		Description string          `json:"description"`
		Priority    string          `json:"priority"`
		DueDate     json.RawMessage `json:"due_date"` // "YYYY-MM-DD" string or null
		AssigneeID  *uint           `json:"assignee_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var maxPos float64
	database.DB.Model(&models.Card{}).Where("column_id = ?", colID).Select("COALESCE(MAX(position), 0)").Scan(&maxPos)

	priority := req.Priority
	if priority == "" {
		priority = "none"
	}

	var dueDate *time.Time
	if len(req.DueDate) > 0 && string(req.DueDate) != "null" {
		var dateStr string
		if err := json.Unmarshal(req.DueDate, &dateStr); err == nil && dateStr != "" {
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				dueDate = &t
			}
		}
	}

	// Atomically increment the project's card counter
	database.DB.Model(&models.Project{}).Where("id = ?", project.ID).
		UpdateColumn("card_counter", gorm.Expr("card_counter + 1"))
	var updatedProject models.Project
	database.DB.Select("card_counter").First(&updatedProject, project.ID)

	card := models.Card{
		ColumnID:    uint(colID),
		ProjectID:   project.ID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    priority,
		DueDate:     dueDate,
		AssigneeID:  req.AssigneeID,
		CreatedByID: userID,
		Position:    maxPos + 1000,
		CardNumber:  updatedProject.CardCounter,
	}
	database.DB.Create(&card)
	database.DB.Preload("Labels").Preload("Assignee").Preload("Tags").First(&card, card.ID)

	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardCardCreated, Payload: card})
	c.JSON(http.StatusCreated, card)
}

func GetCard(c *gin.Context) {
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

	var card models.Card
	if err := database.DB.Preload("Labels").Preload("Assignee").Preload("Assignees").Preload("Watchers").Preload("Comments.User").Preload("Tags").Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}
	if am := LoadAttachments("card", []uint{card.ID}); len(am[card.ID]) > 0 {
		card.Attachments = am[card.ID]
	}
	c.JSON(http.StatusOK, card)
}

func UpdateCard(c *gin.Context) {
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

	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	var req struct {
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Priority    string          `json:"priority"`
		DueDate     json.RawMessage `json:"due_date"` // "YYYY-MM-DD" string or null
		AssigneeID  *uint           `json:"assignee_id"`
	}
	c.ShouldBindJSON(&req)

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Priority != "" {
		updates["priority"] = req.Priority
	}
	if len(req.DueDate) > 0 {
		if string(req.DueDate) == "null" {
			updates["due_date"] = nil
		} else {
			var dateStr string
			if err := json.Unmarshal(req.DueDate, &dateStr); err == nil && dateStr != "" {
				if t, err := time.Parse("2006-01-02", dateStr); err == nil {
					updates["due_date"] = t
				}
			}
		}
	}
	if req.AssigneeID != nil {
		updates["assignee_id"] = req.AssigneeID
	}

	database.DB.Model(&card).Updates(updates)
	database.DB.Preload("Labels").Preload("Assignee").Preload("Tags").First(&card, card.ID)

	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardCardUpdated, Payload: card})
	c.JSON(http.StatusOK, card)
}

func DeleteCard(c *gin.Context) {
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

	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	database.DB.Delete(&card)
	ws.BroadcastToProject(project.ID, ws.Message{
		Type:    ws.TypeBoardCardDeleted,
		Payload: map[string]uint{"card_id": uint(cardID), "column_id": card.ColumnID},
	})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func MoveCard(c *gin.Context) {
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
		ColumnID uint    `json:"column_id" binding:"required"`
		Position float64 `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	oldColumnID := card.ColumnID
	database.DB.Model(&card).Updates(map[string]interface{}{
		"column_id": req.ColumnID,
		"position":  req.Position,
	})

	if oldColumnID != req.ColumnID {
		database.DB.Create(&models.CardHistory{
			CardID:       card.ID,
			UserID:       userID,
			FromColumnID: oldColumnID,
			ToColumnID:   req.ColumnID,
		})
	}

	ws.BroadcastToProject(project.ID, ws.Message{
		Type: ws.TypeBoardCardMoved,
		Payload: map[string]interface{}{
			"card_id":        card.ID,
			"from_column_id": oldColumnID,
			"to_column_id":   req.ColumnID,
			"position":       req.Position,
		},
	})
	c.JSON(http.StatusOK, card)
}

func GetCardHistory(c *gin.Context) {
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

	var history []models.CardHistory
	database.DB.Preload("User").Preload("FromColumn").Preload("ToColumn").
		Where("card_id = ?", cardID).
		Order("created_at desc").
		Find(&history)
	c.JSON(http.StatusOK, history)
}

func ReorderCards(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	colID, err := strconv.ParseUint(c.Param("columnId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid column id"})
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

	var req []struct {
		ID       uint    `json:"id"`
		Position float64 `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, item := range req {
		database.DB.Model(&models.Card{}).Where("id = ? AND column_id = ?", item.ID, colID).Update("position", item.Position)
	}

	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

func AssignLabel(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, _ := strconv.ParseUint(c.Param("cardId"), 10, 64)
	labelID, _ := strconv.ParseUint(c.Param("labelId"), 10, 64)

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	cl := models.CardLabel{CardID: uint(cardID), LabelID: uint(labelID)}
	database.DB.FirstOrCreate(&cl, cl)
	c.JSON(http.StatusOK, gin.H{"message": "assigned"})
}

func RemoveLabel(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, _ := strconv.ParseUint(c.Param("cardId"), 10, 64)
	labelID, _ := strconv.ParseUint(c.Param("labelId"), 10, 64)

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	database.DB.Where("card_id = ? AND label_id = ?", cardID, labelID).Delete(&models.CardLabel{})
	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}

func UpdateAssignee(c *gin.Context) {
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
		UserID *uint `json:"user_id"`
	}
	c.ShouldBindJSON(&req)

	database.DB.Model(&models.Card{}).Where("id = ? AND project_id = ?", cardID, project.ID).Update("assignee_id", req.UserID)

	if notifSvc != nil && req.UserID != nil {
		var card models.Card
		var assignee, assigner models.User
		database.DB.First(&card, cardID)
		database.DB.First(&assignee, *req.UserID)
		database.DB.First(&assigner, userID)
		go notifSvc.NotifyCardAssignment(card, assignee, assigner)
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
