package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// ListContacts returns all contacts for a customer.
func ListContacts(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)
	if err := requireCustomerAccess(uint(custID), userID, globalRole); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var contacts []models.CustomerContact
	database.DB.Where("customer_id = ?", custID).Order("is_primary desc, id asc").Find(&contacts)
	c.JSON(http.StatusOK, contacts)
}

type contactRequest struct {
	Name       string `json:"name" binding:"required"`
	Department string `json:"department"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	IsPrimary  bool   `json:"is_primary"`
}

// CreateContact adds a contact person to a customer.
func CreateContact(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	if !canManageCustomer(c, uint(custID)) && middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req contactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.IsPrimary {
		database.DB.Model(&models.CustomerContact{}).
			Where("customer_id = ?", custID).
			Update("is_primary", false)
	}
	contact := models.CustomerContact{
		CustomerID: uint(custID),
		Name:       req.Name,
		Department: req.Department,
		Phone:      req.Phone,
		Email:      req.Email,
		IsPrimary:  req.IsPrimary,
	}
	if err := database.DB.Create(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create contact"})
		return
	}
	c.JSON(http.StatusCreated, contact)
}

// UpdateContact updates a contact person.
func UpdateContact(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	contactID, err := strconv.ParseUint(c.Param("contactId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact id"})
		return
	}
	if !canManageCustomer(c, uint(custID)) && middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var contact models.CustomerContact
	if err := database.DB.First(&contact, contactID).Error; err != nil || contact.CustomerID != uint(custID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req contactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.IsPrimary {
		database.DB.Model(&models.CustomerContact{}).
			Where("customer_id = ? AND id != ?", custID, contactID).
			Update("is_primary", false)
	}
	database.DB.Model(&contact).Updates(map[string]interface{}{
		"name":       req.Name,
		"department": req.Department,
		"phone":      req.Phone,
		"email":      req.Email,
		"is_primary": req.IsPrimary,
	})
	c.JSON(http.StatusOK, contact)
}

// DeleteContact removes a contact person.
func DeleteContact(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	contactID, err := strconv.ParseUint(c.Param("contactId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact id"})
		return
	}
	if !canManageCustomer(c, uint(custID)) && middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var contact models.CustomerContact
	if err := database.DB.First(&contact, contactID).Error; err != nil || contact.CustomerID != uint(custID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	database.DB.Delete(&contact)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
