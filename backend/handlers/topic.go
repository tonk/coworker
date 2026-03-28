package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/middleware"
	"github.com/tonk/coworker/models"
	"github.com/tonk/coworker/services"
	appws "github.com/tonk/coworker/ws"
)

// ListTopics GET /projects/:projectSlug/topics
func ListTopics(c *gin.Context) {
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

	var topics []models.Topic
	database.DB.Preload("User").Where("project_id = ?", project.ID).Order("is_pinned desc, created_at desc").Find(&topics)

	if len(topics) > 0 {
		topicIDs := make([]uint, len(topics))
		for i, t := range topics {
			topicIDs[i] = t.ID
		}
		type countRow struct {
			TopicID uint
			Count   int
		}
		var rows []countRow
		database.DB.Model(&models.TopicReply{}).
			Select("topic_id, count(*) as count").
			Where("topic_id IN ? AND deleted_at IS NULL", topicIDs).
			Group("topic_id").Scan(&rows)
		counts := make(map[uint]int, len(rows))
		for _, r := range rows {
			counts[r.TopicID] = r.Count
		}
		for i := range topics {
			topics[i].ReplyCount = counts[topics[i].ID]
		}
	}

	c.JSON(http.StatusOK, topics)
}

// CreateTopic POST /projects/:projectSlug/topics
func CreateTopic(c *gin.Context) {
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
		Title string `json:"title" binding:"required"`
		Body  string `json:"body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic := models.Topic{
		ProjectID: project.ID,
		UserID:    userID,
		Title:     req.Title,
		Body:      req.Body,
	}
	database.DB.Create(&topic)
	database.DB.Preload("User").First(&topic, topic.ID)

	appws.BroadcastToProject(project.ID, appws.Message{Type: appws.TypeTopicCreated, Payload: topic})
	c.JSON(http.StatusCreated, topic)
}

// GetTopic GET /projects/:projectSlug/topics/:topicId
func GetTopic(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	topicID, err := strconv.ParseUint(c.Param("topicId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
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

	var topic models.Topic
	if err := database.DB.Preload("User").Where("id = ? AND project_id = ?", topicID, project.ID).First(&topic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "topic not found"})
		return
	}

	var replies []models.TopicReply
	database.DB.Preload("User").Where("topic_id = ?", topic.ID).Order("created_at asc").Find(&replies)

	c.JSON(http.StatusOK, gin.H{"topic": topic, "replies": replies})
}

// UpdateTopic PUT /projects/:projectSlug/topics/:topicId
func UpdateTopic(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	topicID, err := strconv.ParseUint(c.Param("topicId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var topic models.Topic
	if err := database.DB.Where("id = ? AND project_id = ?", topicID, project.ID).First(&topic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "topic not found"})
		return
	}

	if topic.UserID != userID {
		if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	var req struct {
		Title    string `json:"title"`
		Body     string `json:"body"`
		IsPinned *bool  `json:"is_pinned"`
	}
	c.ShouldBindJSON(&req)

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
		updates["is_edited"] = true
	}
	if req.Body != "" {
		updates["body"] = req.Body
		updates["is_edited"] = true
	}
	if req.IsPinned != nil {
		updates["is_pinned"] = *req.IsPinned
	}
	if len(updates) > 0 {
		database.DB.Model(&topic).Updates(updates)
	}
	database.DB.Preload("User").First(&topic, topic.ID)

	appws.BroadcastToProject(project.ID, appws.Message{Type: appws.TypeTopicUpdated, Payload: topic})
	c.JSON(http.StatusOK, topic)
}

// DeleteTopic DELETE /projects/:projectSlug/topics/:topicId
func DeleteTopic(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	topicID, err := strconv.ParseUint(c.Param("topicId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var topic models.Topic
	if err := database.DB.Where("id = ? AND project_id = ?", topicID, project.ID).First(&topic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "topic not found"})
		return
	}

	if topic.UserID != userID {
		if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	database.DB.Delete(&topic)
	appws.BroadcastToProject(project.ID, appws.Message{
		Type:    appws.TypeTopicDeleted,
		Payload: gin.H{"id": topicID},
	})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// CreateTopicReply POST /projects/:projectSlug/topics/:topicId/replies
func CreateTopicReply(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	topicID, err := strconv.ParseUint(c.Param("topicId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
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

	var topic models.Topic
	if err := database.DB.Where("id = ? AND project_id = ?", topicID, project.ID).First(&topic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "topic not found"})
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reply := models.TopicReply{
		TopicID: uint(topicID),
		UserID:  userID,
		Body:    req.Body,
	}
	database.DB.Create(&reply)
	database.DB.Preload("User").First(&reply, reply.ID)

	appws.BroadcastToProject(project.ID, appws.Message{Type: appws.TypeTopicReplyCreated, Payload: reply})
	c.JSON(http.StatusCreated, reply)
}

// UpdateTopicReply PUT /projects/:projectSlug/topics/:topicId/replies/:replyId
func UpdateTopicReply(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	topicID, err := strconv.ParseUint(c.Param("topicId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	replyID, err := strconv.ParseUint(c.Param("replyId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var reply models.TopicReply
	if err := database.DB.Where("id = ? AND topic_id = ?", replyID, topicID).First(&reply).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reply not found"})
		return
	}

	if reply.UserID != userID {
		if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Model(&reply).Updates(map[string]interface{}{"body": req.Body, "is_edited": true})
	database.DB.Preload("User").First(&reply, reply.ID)

	appws.BroadcastToProject(project.ID, appws.Message{Type: appws.TypeTopicReplyUpdated, Payload: reply})
	c.JSON(http.StatusOK, reply)
}

// DeleteTopicReply DELETE /projects/:projectSlug/topics/:topicId/replies/:replyId
func DeleteTopicReply(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	topicID, err := strconv.ParseUint(c.Param("topicId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	replyID, err := strconv.ParseUint(c.Param("replyId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var reply models.TopicReply
	if err := database.DB.Where("id = ? AND topic_id = ?", replyID, topicID).First(&reply).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reply not found"})
		return
	}

	if reply.UserID != userID {
		if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	database.DB.Delete(&reply)
	appws.BroadcastToProject(project.ID, appws.Message{
		Type:    appws.TypeTopicReplyDeleted,
		Payload: gin.H{"id": replyID, "topic_id": topicID},
	})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
