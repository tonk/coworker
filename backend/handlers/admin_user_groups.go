package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// AdminGetUserGroups returns the IDs of all groups the user is a member of.
func AdminGetUserGroups(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var members []models.GroupMember
	database.DB.Where("user_id = ?", userID).Find(&members)

	ids := make([]uint, len(members))
	for i, m := range members {
		ids[i] = m.GroupID
	}
	c.JSON(http.StatusOK, gin.H{"group_ids": ids})
}

// AdminSetUserGroups syncs a user's group memberships to exactly the given list.
// Existing memberships not in the list are removed; missing ones are added.
func AdminSetUserGroups(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		GroupIDs []uint `json:"group_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	desired := make(map[uint]bool, len(req.GroupIDs))
	for _, id := range req.GroupIDs {
		desired[id] = true
	}

	var current []models.GroupMember
	database.DB.Where("user_id = ?", userID).Find(&current)

	currentSet := make(map[uint]bool, len(current))
	for _, m := range current {
		currentSet[m.GroupID] = true
	}

	for _, m := range current {
		if !desired[m.GroupID] {
			database.DB.Where("group_id = ? AND user_id = ?", m.GroupID, userID).
				Delete(&models.GroupMember{})
			var g models.UserGroup
			if database.DB.First(&g, m.GroupID).Error == nil && g.ConversationID != nil {
				database.DB.Where("conversation_id = ? AND user_id = ?", *g.ConversationID, userID).
					Delete(&models.ConversationMember{})
			}
		}
	}

	for _, gid := range req.GroupIDs {
		if !currentSet[gid] {
			database.DB.Create(&models.GroupMember{GroupID: gid, UserID: uint(userID)})
			var g models.UserGroup
			if database.DB.First(&g, gid).Error == nil && g.ConversationID != nil {
				database.DB.FirstOrCreate(&models.ConversationMember{}, models.ConversationMember{
					ConversationID: *g.ConversationID,
					UserID:         uint(userID),
					JoinedAt:       time.Now(),
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"group_ids": req.GroupIDs})
}
