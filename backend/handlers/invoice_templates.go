package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

// ListInvoiceTemplates returns all invoice templates.
func ListInvoiceTemplates(c *gin.Context) {
	var templates []models.InvoiceTemplate
	database.DB.Order("name asc").Find(&templates)
	c.JSON(http.StatusOK, templates)
}

type invoiceTemplateRequest struct {
	Name            string  `json:"name" binding:"required"`
	LineItems       string  `json:"line_items"`
	DefaultVATRate  float64 `json:"default_vat_rate"`
	DefaultCurrency string  `json:"default_currency"`
	Notes           string  `json:"notes"`
}

// AdminCreateInvoiceTemplate creates a new invoice template.
func AdminCreateInvoiceTemplate(c *gin.Context) {
	var req invoiceTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t := models.InvoiceTemplate{
		Name:            req.Name,
		LineItems:       req.LineItems,
		DefaultVATRate:  req.DefaultVATRate,
		DefaultCurrency: req.DefaultCurrency,
		Notes:           req.Notes,
	}
	if err := database.DB.Create(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template"})
		return
	}
	c.JSON(http.StatusCreated, t)
}

// AdminUpdateInvoiceTemplate updates an existing invoice template.
func AdminUpdateInvoiceTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var t models.InvoiceTemplate
	if err := database.DB.First(&t, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req invoiceTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t.Name = req.Name
	t.LineItems = req.LineItems
	t.DefaultVATRate = req.DefaultVATRate
	t.DefaultCurrency = req.DefaultCurrency
	t.Notes = req.Notes
	database.DB.Save(&t)
	c.JSON(http.StatusOK, t)
}

// AdminDeleteInvoiceTemplate deletes an invoice template.
func AdminDeleteInvoiceTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var t models.InvoiceTemplate
	if err := database.DB.First(&t, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	database.DB.Delete(&t)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
