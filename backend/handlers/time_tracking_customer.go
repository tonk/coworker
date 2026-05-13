package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// ListTimeTrackingCustomers returns time-tracking-only customers visible to the
// current user: their own plus any created by an admin.
func ListTimeTrackingCustomers(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var customers []models.Customer
	if globalRole == "admin" {
		database.DB.Where("time_tracking_only = true").
			Order("name asc").Find(&customers)
	} else {
		database.DB.Raw(`
			SELECT cu.* FROM customers cu
			JOIN users u ON u.id = cu.created_by_id
			WHERE cu.time_tracking_only = true
			  AND (cu.created_by_id = ? OR u.global_role = 'admin')
			ORDER BY cu.name ASC`, userID).Scan(&customers)
	}
	c.JSON(http.StatusOK, customers)
}

// CreateTimeTrackingCustomer creates a new time-tracking-only customer owned by
// the current user.
func CreateTimeTrackingCustomer(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if len(name) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name too long"})
		return
	}

	customer := models.Customer{
		Name:             name,
		TimeTrackingOnly: true,
		CreatedByID:      &userID,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, customer)
}

// UpdateTimeTrackingCustomer updates the name of a time-tracking-only customer.
// Users can only update their own; admins can update any.
func UpdateTimeTrackingCustomer(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var customer models.Customer
	if err := database.DB.Where("id = ? AND time_tracking_only = true", id).
		First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if globalRole != "admin" && (customer.CreatedByID == nil || *customer.CreatedByID != userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if len(name) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name too long"})
		return
	}

	customer.Name = name
	if err := database.DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, customer)
}

// DeleteTimeTrackingCustomer deletes a time-tracking-only customer.
// Users can only delete their own; admins can delete any.
func DeleteTimeTrackingCustomer(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var customer models.Customer
	if err := database.DB.Where("id = ? AND time_tracking_only = true", id).
		First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if globalRole != "admin" && (customer.CreatedByID == nil || *customer.CreatedByID != userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	database.DB.Delete(&customer)
	c.JSON(http.StatusNoContent, nil)
}
