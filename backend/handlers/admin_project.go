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
		Avatar      string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
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
		Avatar:      req.Avatar,
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
	if closed := c.Query("closed"); closed == "true" {
		q = q.Where("is_closed = true")
	} else if closed == "hide" {
		q = q.Where("is_closed = false")
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
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Color       string  `json:"color"`
		Avatar      *string `json:"avatar"`
		IsArchived  *bool   `json:"is_archived"`
		IsClosed    *bool   `json:"is_closed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
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
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.IsArchived != nil {
		updates["is_archived"] = *req.IsArchived
	}
	if req.IsClosed != nil {
		updates["is_closed"] = *req.IsClosed
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

// AdminPurgeProject permanently removes a soft-deleted project and all of its data.
// Only projects already soft-deleted via AdminDeleteProject can be purged.
func AdminPurgeProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var project models.Project
	if err := database.DB.Unscoped().Where("id = ? AND deleted_at IS NOT NULL", id).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found or not deleted"})
		return
	}

	db := database.DB

	// Card-level associations
	var cardIDs []uint
	db.Unscoped().Model(&models.Card{}).Where("project_id = ?", id).Pluck("id", &cardIDs)
	if len(cardIDs) > 0 {
		var commentIDs []uint
		db.Unscoped().Model(&models.CardComment{}).Where("card_id IN ?", cardIDs).Pluck("id", &commentIDs)
		if len(commentIDs) > 0 {
			db.Where("owner_type = 'card_comment' AND owner_id IN ?", commentIDs).Delete(&models.Attachment{})
		}
		db.Where("card_id IN ?", cardIDs).Delete(&models.CardChecklistItem{})
		db.Unscoped().Where("card_id IN ?", cardIDs).Delete(&models.CardComment{})
		db.Where("source_card_id IN ? OR target_card_id IN ?", cardIDs, cardIDs).Delete(&models.CardReference{})
		db.Where("card_id IN ?", cardIDs).Delete(&models.CardHistory{})
		db.Where("card_id IN ?", cardIDs).Delete(&models.CardAssignee{})
		db.Where("card_id IN ?", cardIDs).Delete(&models.CardLabel{})
		db.Exec("DELETE FROM card_watchers WHERE card_id IN ?", cardIDs)
		db.Where("card_id IN ?", cardIDs).Delete(&models.CardTag{})
		db.Where("card_id IN ?", cardIDs).Delete(&models.CardLink{})
	}
	db.Unscoped().Where("project_id = ?", id).Delete(&models.Card{})
	db.Unscoped().Where("project_id = ?", id).Delete(&models.Column{})
	db.Unscoped().Where("project_id = ?", id).Delete(&models.Label{})

	// Topics and replies
	var topicIDs []uint
	db.Unscoped().Model(&models.Topic{}).Where("project_id = ?", id).Pluck("id", &topicIDs)
	if len(topicIDs) > 0 {
		db.Unscoped().Where("topic_id IN ?", topicIDs).Delete(&models.TopicReply{})
	}
	db.Unscoped().Where("project_id = ?", id).Delete(&models.Topic{})

	// Chat messages, their reactions and attachments
	var chatMsgIDs []uint
	db.Model(&models.ChatMessage{}).Where("project_id = ?", id).Pluck("id", &chatMsgIDs)
	if len(chatMsgIDs) > 0 {
		db.Where("owner_type = 'chat_message' AND owner_id IN ?", chatMsgIDs).Delete(&models.MessageReaction{})
		db.Where("owner_type = 'chat_message' AND owner_id IN ?", chatMsgIDs).Delete(&models.Attachment{})
	}
	db.Where("project_id = ?", id).Delete(&models.ChatMessage{})

	db.Where("project_id = ?", id).Delete(&models.ProjectWebhook{})

	// Sprints and their card assignments
	var sprintIDs []uint
	db.Unscoped().Model(&models.Sprint{}).Where("project_id = ?", id).Pluck("id", &sprintIDs)
	if len(sprintIDs) > 0 {
		db.Where("sprint_id IN ?", sprintIDs).Delete(&models.SprintCard{})
	}
	db.Unscoped().Where("project_id = ?", id).Delete(&models.Sprint{})

	db.Where("project_id = ?", id).Delete(&models.APIKey{})
	db.Where("project_id = ?", id).Delete(&models.StarredProject{})
	db.Where("project_id = ?", id).Delete(&models.ProjectMember{})

	db.Unscoped().Delete(&project)
	c.JSON(http.StatusOK, gin.H{"message": "project permanently deleted"})
}
