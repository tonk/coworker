package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// nextInvoiceNumber generates the next invoice number using the configured prefix
// and the count of existing invoices (e.g. "INV-0042").
func nextInvoiceNumber(prefix string) string {
	if prefix == "" {
		prefix = "INV"
	}
	var count int64
	database.DB.Model(&models.Invoice{}).Count(&count)
	return fmt.Sprintf("%s-%04d", prefix, count+1)
}

// ListInvoices returns all invoices for a customer, newest first.
func ListInvoices(c *gin.Context) {
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
	var invoices []models.Invoice
	database.DB.Where("customer_id = ?", custID).Order("created_at desc").Find(&invoices)
	c.JSON(http.StatusOK, invoices)
}

// GetInvoice returns a single invoice.
func GetInvoice(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	invoiceID, err := strconv.ParseUint(c.Param("invoiceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice id"})
		return
	}
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)
	if err := requireCustomerAccess(uint(custID), userID, globalRole); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var invoice models.Invoice
	if err := database.DB.Preload("Customer").First(&invoice, invoiceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if invoice.CustomerID != uint(custID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, invoice)
}

// CreateInvoiceRequest is the body for POST /customers/:id/invoices.
type CreateInvoiceRequest struct {
	PeriodStart string                   `json:"period_start" binding:"required"`
	PeriodEnd   string                   `json:"period_end" binding:"required"`
	LineItems   []models.InvoiceLineItem `json:"line_items" binding:"required"`
	Currency    string                   `json:"currency"`
	VATRate     float64                  `json:"vat_rate"`
	DueDate     string                   `json:"due_date"`
	Notes       string                   `json:"notes"`
}

// CreateInvoice creates an invoice for a customer from a supplied set of line items.
func CreateInvoice(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)
	if !canManageCustomer(c, uint(custID)) && globalRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_start"})
		return
	}
	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_end"})
		return
	}

	// Compute totals from line items.
	subtotal := 0.0
	currency := req.Currency
	for _, li := range req.LineItems {
		subtotal += li.Amount
		if currency == "" && li.Currency != "" {
			currency = li.Currency
		}
	}
	vatAmount := subtotal * req.VATRate / 100.0
	total := subtotal + vatAmount

	lineItemsJSON, err := json.Marshal(req.LineItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode line items"})
		return
	}

	settings := loadAllSettings()
	prefix := settings[settingInvoiceNumberPrefix]

	invoice := models.Invoice{
		CustomerID:    uint(custID),
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		Status:        models.InvoiceStatusDraft,
		Currency:      currency,
		LineItems:     string(lineItemsJSON),
		Subtotal:      subtotal,
		VATRate:       req.VATRate,
		VATAmount:     vatAmount,
		Total:         total,
		Notes:         req.Notes,
		CreatedByID:   &userID,
		InvoiceNumber: nextInvoiceNumber(prefix),
	}

	if req.DueDate != "" {
		if dd, err := time.Parse("2006-01-02", req.DueDate); err == nil {
			invoice.DueDate = &dd
		}
	}

	if err := database.DB.Create(&invoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invoice"})
		return
	}
	database.DB.Preload("Customer").First(&invoice, invoice.ID)
	c.JSON(http.StatusCreated, invoice)
}

// UpdateInvoiceRequest is the body for PUT /customers/:id/invoices/:invoiceId.
type UpdateInvoiceRequest struct {
	Status  string   `json:"status"`
	Notes   string   `json:"notes"`
	DueDate string   `json:"due_date"`
	VATRate *float64 `json:"vat_rate"`
}

// UpdateInvoice updates an invoice's status, notes, due date, or VAT rate.
func UpdateInvoice(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	invoiceID, err := strconv.ParseUint(c.Param("invoiceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice id"})
		return
	}
	globalRole := middleware.GetGlobalRole(c)
	if !canManageCustomer(c, uint(custID)) && globalRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var invoice models.Invoice
	if err := database.DB.First(&invoice, invoiceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if invoice.CustomerID != uint(custID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var req UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{"notes": req.Notes}
	if req.Status != "" {
		valid := map[string]bool{
			models.InvoiceStatusDraft: true,
			models.InvoiceStatusSent:  true,
			models.InvoiceStatusPaid:  true,
		}
		if !valid[req.Status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		updates["status"] = req.Status
	}
	if req.DueDate != "" {
		if dd, err := time.Parse("2006-01-02", req.DueDate); err == nil {
			updates["due_date"] = dd
		}
	} else {
		updates["due_date"] = nil
	}
	if req.VATRate != nil {
		vatAmount := invoice.Subtotal * *req.VATRate / 100.0
		updates["vat_rate"] = *req.VATRate
		updates["vat_amount"] = vatAmount
		updates["total"] = invoice.Subtotal + vatAmount
	}

	database.DB.Model(&invoice).Updates(updates)
	database.DB.Preload("Customer").First(&invoice, invoice.ID)
	c.JSON(http.StatusOK, invoice)
}

// DeleteInvoice deletes an invoice (only allowed in draft status).
func DeleteInvoice(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	invoiceID, err := strconv.ParseUint(c.Param("invoiceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice id"})
		return
	}
	globalRole := middleware.GetGlobalRole(c)
	if !canManageCustomer(c, uint(custID)) && globalRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var invoice models.Invoice
	if err := database.DB.First(&invoice, invoiceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if invoice.CustomerID != uint(custID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if invoice.Status != models.InvoiceStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only draft invoices can be deleted"})
		return
	}
	database.DB.Delete(&invoice)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
