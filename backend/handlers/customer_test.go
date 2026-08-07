package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/testutil"
)

// A regression guard for a bug where UpdateCustomer unconditionally
// overwrote color/billing/VAT/PO fields from the request body, so any
// partial update (e.g. the web UI's inline rename, which only sends
// name/description/logo_url) silently wiped them back to empty.
func TestUpdateCustomerPreservesOmittedFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	prevDB := database.DB
	database.DB = db
	defer func() { database.DB = prevDB }()

	cust := &models.Customer{
		Name:          "Acme",
		Color:         "#3b82f6",
		BillingStreet: "1 Main St",
		VATNumber:     "NL123456789B01",
		POReference:   "PO-42",
	}
	require.NoError(t, db.Create(cust).Error)

	admin := &models.User{Username: "admin", Email: "admin@example.com", PasswordHash: "x", GlobalRole: "admin"}
	require.NoError(t, db.Create(admin).Error)

	c, w := ginTestContext(t, http.MethodPut, "/customers/1", `{"name":"Acme Renamed"}`)
	c.Set(middleware.ContextUserID, admin.ID)
	c.Set(middleware.ContextGlobalRole, "admin")
	c.Params = gin.Params{{Key: "customerId", Value: "1"}}

	UpdateCustomer(c)
	require.Equal(t, http.StatusOK, w.Code)

	var reloaded models.Customer
	require.NoError(t, db.First(&reloaded, cust.ID).Error)
	require.Equal(t, "Acme Renamed", reloaded.Name)
	require.Equal(t, "#3b82f6", reloaded.Color, "color must survive a partial update that doesn't mention it")
	require.Equal(t, "1 Main St", reloaded.BillingStreet)
	require.Equal(t, "NL123456789B01", reloaded.VATNumber)
	require.Equal(t, "PO-42", reloaded.POReference)
}

func ginTestContext(t *testing.T, method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}
