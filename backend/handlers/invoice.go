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
	"github.com/tonk/warmdesk/services"
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

// ListAllInvoices returns all invoices accessible to the requesting user,
// optionally filtered by ?customer_id= and/or ?status=.
// Admins see all invoices; regular users only see invoices for customers
// they have access to.
func ListAllInvoices(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	q := database.DB.Preload("Customer").Order("created_at desc")

	if globalRole != "admin" {
		// Restrict to customers this user can access.
		accessible := getAccessibleCustomerRoles(userID)
		if len(accessible) == 0 {
			c.JSON(http.StatusOK, []models.Invoice{})
			return
		}
		ids := make([]uint, 0, len(accessible))
		for id := range accessible {
			ids = append(ids, id)
		}
		q = q.Where("customer_id IN ?", ids)
	}

	if custStr := c.Query("customer_id"); custStr != "" {
		if custID, err := strconv.ParseUint(custStr, 10, 64); err == nil && custID > 0 {
			q = q.Where("customer_id = ?", custID)
		}
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}

	var invoices []models.Invoice
	q.Find(&invoices)
	c.JSON(http.StatusOK, invoices)
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
	Status           string                   `json:"status"`
	Notes            string                   `json:"notes"`
	DueDate          string                   `json:"due_date"`
	VATRate          *float64                 `json:"vat_rate"`
	LineItems        []models.InvoiceLineItem `json:"line_items"`
	PaymentDate      string                   `json:"payment_date"`
	PaymentAmount    *float64                 `json:"payment_amount"`
	PaymentReference string                   `json:"payment_reference"`
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
	// Update line items if provided.
	if len(req.LineItems) > 0 {
		newSubtotal := 0.0
		for _, li := range req.LineItems {
			newSubtotal += li.Amount
		}
		liJSON, _ := json.Marshal(req.LineItems)
		updates["line_items"] = string(liJSON)
		updates["subtotal"] = newSubtotal
		vatRate := invoice.VATRate
		if req.VATRate != nil {
			vatRate = *req.VATRate
		}
		vatAmt := newSubtotal * vatRate / 100.0
		updates["vat_amount"] = vatAmt
		updates["total"] = newSubtotal + vatAmt
	}
	// Payment details.
	updates["payment_reference"] = req.PaymentReference
	if req.PaymentDate != "" {
		if pd, err := time.Parse("2006-01-02", req.PaymentDate); err == nil {
			updates["payment_date"] = pd
		}
	} else {
		updates["payment_date"] = nil
	}
	if req.PaymentAmount != nil {
		updates["payment_amount"] = req.PaymentAmount
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

// SendInvoiceRequest is the body for POST /customers/:id/invoices/:id/send.
type SendInvoiceRequest struct {
	To      string `json:"to" binding:"required"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// SendInvoice emails the invoice PDF to the given address and marks it sent.
func SendInvoice(c *gin.Context) {
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
	if !canManageCustomer(c, uint(custID)) && middleware.GetGlobalRole(c) != "admin" {
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

	var req SendInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := services.GetEmailService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "email not configured"})
		return
	}

	// Build PDF bytes in memory.
	lang := c.DefaultQuery("lang", "en")
	pdfBytes, err := renderInvoicePDF(invoice, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}

	subject := req.Subject
	if subject == "" {
		subject = "Invoice " + invoice.InvoiceNumber
	}
	body := req.Body
	if body == "" {
		body = "Please find your invoice attached."
	}

	if err := svc.SendWithAttachment(req.To, subject, body, invoice.InvoiceNumber+".pdf", "application/pdf", pdfBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email: " + err.Error()})
		return
	}

	// Mark as sent if still draft.
	if invoice.Status == models.InvoiceStatusDraft {
		database.DB.Model(&invoice).Update("status", models.InvoiceStatusSent)
	}
	database.DB.Preload("Customer").First(&invoice, invoice.ID)
	c.JSON(http.StatusOK, invoice)
}

// CreateCreditNote creates a credit note that negates the specified invoice.
func CreateCreditNote(c *gin.Context) {
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
	if !canManageCustomer(c, uint(custID)) && middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var orig models.Invoice
	if err := database.DB.Preload("Customer").First(&orig, invoiceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if orig.CustomerID != uint(custID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if orig.Status == models.InvoiceStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot credit a draft invoice"})
		return
	}

	// Negate all line item amounts.
	var lineItems []models.InvoiceLineItem
	_ = json.Unmarshal([]byte(orig.LineItems), &lineItems)
	for i := range lineItems {
		lineItems[i].Amount = -lineItems[i].Amount
	}
	liJSON, _ := json.Marshal(lineItems)

	settings := loadAllSettings()
	prefix := settings[settingInvoiceNumberPrefix]
	origIDRef := orig.ID
	userID := middleware.GetUserID(c)

	cn := models.Invoice{
		InvoiceNumber:     nextInvoiceNumber(prefix),
		CustomerID:        orig.CustomerID,
		PeriodStart:       orig.PeriodStart,
		PeriodEnd:         orig.PeriodEnd,
		Status:            models.InvoiceStatusCreditNote,
		Currency:          orig.Currency,
		LineItems:         string(liJSON),
		Subtotal:          -orig.Subtotal,
		VATRate:           orig.VATRate,
		VATAmount:         -orig.VATAmount,
		Total:             -orig.Total,
		Notes:             "Credit note for " + orig.InvoiceNumber,
		CreatedByID:       &userID,
		CreditedInvoiceID: &origIDRef,
	}
	if err := database.DB.Create(&cn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create credit note"})
		return
	}
	database.DB.Preload("Customer").First(&cn, cn.ID)
	c.JSON(http.StatusCreated, cn)
}
