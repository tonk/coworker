package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// ListLocations returns all locations for a customer.
func ListLocations(c *gin.Context) {
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
	var locations []models.CustomerLocation
	database.DB.Where("customer_id = ?", custID).Order("id asc").Find(&locations)
	c.JSON(http.StatusOK, locations)
}

type locationRequest struct {
	AddressLine1   string   `json:"address_line1"`
	AddressLine2   string   `json:"address_line2"`
	City           string   `json:"city"`
	PostalCode     string   `json:"postal_code"`
	Region         string   `json:"region"`
	Country        string   `json:"country"`
	Phone          string   `json:"phone"`
	ContactName    string   `json:"contact_name"`
	ContactEmail   string   `json:"contact_email"`
	ContactPhone   string   `json:"contact_phone"`
	TravelDistance *float64 `json:"travel_distance"`
}

// CreateLocation adds a location to a customer.
func CreateLocation(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	if !canManageCustomer(c, uint(custID)) && middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req locationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	location := models.CustomerLocation{
		CustomerID:     uint(custID),
		AddressLine1:   req.AddressLine1,
		AddressLine2:   req.AddressLine2,
		City:           req.City,
		PostalCode:     req.PostalCode,
		Region:         req.Region,
		Country:        req.Country,
		Phone:          req.Phone,
		ContactName:    req.ContactName,
		ContactEmail:   req.ContactEmail,
		ContactPhone:   req.ContactPhone,
		TravelDistance: req.TravelDistance,
	}
	if err := database.DB.Create(&location).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create location"})
		return
	}
	c.JSON(http.StatusCreated, location)
}

// UpdateLocation updates a location.
func UpdateLocation(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	locationID, err := strconv.ParseUint(c.Param("locationId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}
	if !canManageCustomer(c, uint(custID)) && middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var location models.CustomerLocation
	if err := database.DB.First(&location, locationID).Error; err != nil || location.CustomerID != uint(custID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req locationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&location).Updates(map[string]interface{}{
		"address_line1":   req.AddressLine1,
		"address_line2":   req.AddressLine2,
		"city":            req.City,
		"postal_code":     req.PostalCode,
		"region":          req.Region,
		"country":         req.Country,
		"phone":           req.Phone,
		"contact_name":    req.ContactName,
		"contact_email":   req.ContactEmail,
		"contact_phone":   req.ContactPhone,
		"travel_distance": req.TravelDistance,
	})
	c.JSON(http.StatusOK, location)
}

// DeleteLocation removes a location.
func DeleteLocation(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	locationID, err := strconv.ParseUint(c.Param("locationId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}
	if !canManageCustomer(c, uint(custID)) && middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var location models.CustomerLocation
	if err := database.DB.First(&location, locationID).Error; err != nil || location.CustomerID != uint(custID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	database.DB.Delete(&location)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
