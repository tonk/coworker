package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/testutil"
)

// A regression guard for a bug where UpdateInvoice unconditionally
// overwrote notes/due_date/payment_reference/payment_date from the request
// body, so the web UI's status-only update (used to mark an invoice as
// "sent") silently wiped notes and due_date on every call.
func TestUpdateInvoicePreservesOmittedFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	prevDB := database.DB
	database.DB = db
	defer func() { database.DB = prevDB }()

	cust := &models.Customer{Name: "Acme"}
	require.NoError(t, db.Create(cust).Error)

	due := time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC)
	inv := &models.Invoice{
		InvoiceNumber:    "INV-0001",
		CustomerID:       cust.ID,
		Status:           "draft",
		Notes:            "Q1 support services",
		DueDate:          &due,
		PaymentReference: "REF-1",
	}
	require.NoError(t, db.Create(inv).Error)

	admin := &models.User{Username: "admin", Email: "admin@example.com", PasswordHash: "x", GlobalRole: "admin"}
	require.NoError(t, db.Create(admin).Error)

	// Mirrors the web UI's changeInvoiceStatus() call: only {status} is sent.
	c, w := ginTestContext(t, http.MethodPut, "/invoices/1", `{"status":"sent"}`)
	c.Set(middleware.ContextUserID, admin.ID)
	c.Set(middleware.ContextGlobalRole, "admin")
	c.Params = gin.Params{{Key: "customerId", Value: "1"}, {Key: "invoiceId", Value: "1"}}

	UpdateInvoice(c)
	require.Equal(t, http.StatusOK, w.Code)

	var reloaded models.Invoice
	require.NoError(t, db.First(&reloaded, inv.ID).Error)
	require.Equal(t, "sent", reloaded.Status)
	require.Equal(t, "Q1 support services", reloaded.Notes, "notes must survive a status-only update")
	require.NotNil(t, reloaded.DueDate, "due_date must survive a status-only update")
	require.Equal(t, "REF-1", reloaded.PaymentReference)
}
