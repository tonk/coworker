package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// AdminListSlaPolicies GET /api/v1/admin/sla-policies
func AdminListSlaPolicies(c *gin.Context) {
	var policies []models.SlaPolicy
	database.DB.Order("name asc").Find(&policies)
	if policies == nil {
		policies = []models.SlaPolicy{}
	}
	c.JSON(http.StatusOK, policies)
}

// AdminCreateSlaPolicy POST /api/v1/admin/sla-policies
func AdminCreateSlaPolicy(c *gin.Context) {
	var req struct {
		Name                  string `json:"name" binding:"required"`
		ResponseTimeMinutes   int    `json:"response_time_minutes"`
		ResolutionTimeMinutes int    `json:"resolution_time_minutes"`
		PriorityFilter        string `json:"priority_filter"`
		IsActive              *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	policy := models.SlaPolicy{
		Name:                  req.Name,
		ResponseTimeMinutes:   req.ResponseTimeMinutes,
		ResolutionTimeMinutes: req.ResolutionTimeMinutes,
		PriorityFilter:        req.PriorityFilter,
		IsActive:              isActive,
	}
	if err := database.DB.Create(&policy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create sla policy"})
		return
	}
	c.JSON(http.StatusCreated, policy)
}

// AdminUpdateSlaPolicy PUT /api/v1/admin/sla-policies/:id
func AdminUpdateSlaPolicy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var policy models.SlaPolicy
	if err := database.DB.First(&policy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sla policy not found"})
		return
	}

	var req struct {
		Name                  *string `json:"name"`
		ResponseTimeMinutes   *int    `json:"response_time_minutes"`
		ResolutionTimeMinutes *int    `json:"resolution_time_minutes"`
		PriorityFilter        *string `json:"priority_filter"`
		IsActive              *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.ResponseTimeMinutes != nil {
		updates["response_time_minutes"] = *req.ResponseTimeMinutes
	}
	if req.ResolutionTimeMinutes != nil {
		updates["resolution_time_minutes"] = *req.ResolutionTimeMinutes
	}
	if req.PriorityFilter != nil {
		updates["priority_filter"] = *req.PriorityFilter
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		database.DB.Model(&policy).Updates(updates)
	}
	database.DB.First(&policy, id)
	c.JSON(http.StatusOK, policy)
}

// AdminDeleteSlaPolicy DELETE /api/v1/admin/sla-policies/:id
func AdminDeleteSlaPolicy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var policy models.SlaPolicy
	if err := database.DB.First(&policy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sla policy not found"})
		return
	}

	database.DB.Delete(&policy)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// MatchSlaPolicy finds an active SLA policy matching the given priority.
func MatchSlaPolicy(priority string) *models.SlaPolicy {
	var policies []models.SlaPolicy
	database.DB.Where("is_active = ?", true).Find(&policies)
	for _, p := range policies {
		if p.PriorityFilter == "" {
			return &p
		}
		for _, pri := range strings.Split(p.PriorityFilter, ",") {
			if strings.TrimSpace(pri) == priority {
				return &p
			}
		}
	}
	return nil
}

// ComputeSlaDeadlines computes SLA deadlines for a ticket given a policy.
func ComputeSlaDeadlines(policy *models.SlaPolicy, now time.Time) (responseDeadline, resolutionDeadline *time.Time) {
	if policy.ResponseTimeMinutes > 0 {
				d := now.Add(time.Duration(policy.ResponseTimeMinutes) * time.Minute)
				responseDeadline = &d
	}
	if policy.ResolutionTimeMinutes > 0 {
		d := now.Add(time.Duration(policy.ResolutionTimeMinutes) * time.Minute)
		resolutionDeadline = &d
	}
	return
}
