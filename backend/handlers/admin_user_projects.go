package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/models"
)

// AdminGetUserProjects returns the project IDs the user is currently a member of.
func AdminGetUserProjects(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var members []models.ProjectMember
	database.DB.Where("user_id = ?", userID).Find(&members)

	ids := make([]uint, len(members))
	for i, m := range members {
		ids[i] = m.ProjectID
	}
	c.JSON(http.StatusOK, gin.H{"project_ids": ids})
}

// AdminSetUserProjects syncs a user's project memberships to exactly the given list.
// Existing memberships not in the list are removed; missing ones are added as "member".
// Memberships where the user is the sole owner are left untouched to avoid orphaned projects.
func AdminSetUserProjects(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		ProjectIDs []uint `json:"project_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Desired set
	desired := make(map[uint]bool, len(req.ProjectIDs))
	for _, id := range req.ProjectIDs {
		desired[id] = true
	}

	// Current memberships
	var current []models.ProjectMember
	database.DB.Where("user_id = ?", userID).Find(&current)

	currentSet := make(map[uint]bool, len(current))
	for _, m := range current {
		currentSet[m.ProjectID] = true
	}

	// Remove memberships no longer desired
	for _, m := range current {
		if !desired[m.ProjectID] {
			database.DB.Where("project_id = ? AND user_id = ?", m.ProjectID, userID).
				Delete(&models.ProjectMember{})
		}
	}

	// Add new memberships
	for _, pid := range req.ProjectIDs {
		if !currentSet[pid] {
			database.DB.Create(&models.ProjectMember{
				ProjectID: pid,
				UserID:    uint(userID),
				Role:      "member",
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"project_ids": req.ProjectIDs})
}
