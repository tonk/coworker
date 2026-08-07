package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/testutil"
)

func TestParseNullableTimestamp(t *testing.T) {
	t.Run("absent key is not provided", func(t *testing.T) {
		tm, ok := parseNullableTimestamp(nil)
		assert.False(t, ok)
		assert.Nil(t, tm)
	})

	t.Run("explicit null clears", func(t *testing.T) {
		tm, ok := parseNullableTimestamp([]byte("null"))
		assert.True(t, ok)
		assert.Nil(t, tm)
	})

	t.Run("a real timestamp parses, with or without fractional seconds", func(t *testing.T) {
		tm, ok := parseNullableTimestamp([]byte(`"2026-06-15T12:00:00Z"`))
		require.True(t, ok)
		require.NotNil(t, tm)
		assert.Equal(t, 2026, tm.Year())

		tm2, ok := parseNullableTimestamp([]byte(`"2026-06-15T12:00:00.000Z"`))
		require.True(t, ok)
		require.NotNil(t, tm2)
		assert.True(t, tm.Equal(*tm2))
	})
}

// A regression guard for a bug where clicking "clear" on a ticket's
// reminder/close date in the web UI (which sends {"reminder_at": null})
// silently did nothing: *time.Time can't distinguish an explicit JSON null
// from an absent key, so the update guard never fired.
func TestUpdateTicketCanClearReminderAndCloseDates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	prevDB := database.DB
	database.DB = db
	defer func() { database.DB = prevDB }()

	cust := &models.Customer{Name: "Acme"}
	require.NoError(t, db.Create(cust).Error)

	reminder := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	ticket := &models.Ticket{
		CustomerID: &cust.ID,
		Title:      "Waiting for vendor",
		Status:     "pending",
		ReminderAt: &reminder,
		CloseAt:    &reminder,
	}
	require.NoError(t, db.Create(ticket).Error)

	admin := &models.User{Username: "admin", Email: "admin@example.com", PasswordHash: "x", GlobalRole: "admin"}
	require.NoError(t, db.Create(admin).Error)

	c, w := ginTestContext(t, http.MethodPut, "/tickets/1", `{"reminder_at":null,"close_at":null}`)
	c.Set(middleware.ContextUserID, admin.ID)
	c.Set(middleware.ContextGlobalRole, "admin")
	c.Params = gin.Params{{Key: "customerId", Value: "1"}, {Key: "ticketId", Value: "1"}}

	UpdateTicket(c)
	require.Equal(t, http.StatusOK, w.Code)

	var reloaded models.Ticket
	require.NoError(t, db.First(&reloaded, ticket.ID).Error)
	assert.Nil(t, reloaded.ReminderAt, "reminder_at must be cleared by an explicit null")
	assert.Nil(t, reloaded.CloseAt, "close_at must be cleared by an explicit null")
}
