package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/testutil"
)

// A regression guard for a bug where every card creation violated the
// card_histories from/to-column foreign key constraints on any database
// that actually enforces them (MySQL, PostgreSQL). The "created" event never
// sets FromColumnID/ToColumnID, and those were plain uint (not nullable), so
// 0 — not a valid column id — was inserted, which those two databases reject
// outright. SQLite doesn't enforce foreign keys by default, which is exactly
// how this went unnoticed through the whole test suite and local dev: this
// test explicitly turns that enforcement on to catch it without needing a
// live MySQL/Postgres server.
func TestCardHistoryRespectsForeignKeysWhenEnforced(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	require.NoError(t, db.Exec("PRAGMA foreign_keys = ON").Error)

	// Sanity check: enforcement is actually active in this test, so a
	// passing test below isn't just SQLite being lenient again.
	badRefErr := db.Exec("INSERT INTO card_histories (card_id, user_id, event_type, from_column_id, to_column_id) VALUES (1, 1, 'column_move', 99999, 99999)").Error
	require.Error(t, badRefErr, "foreign_keys pragma should be enforced in this test")

	user := &models.User{Username: "u", Email: "u@example.com", PasswordHash: "x"}
	require.NoError(t, db.Create(user).Error)
	project := &models.Project{Name: "P", Slug: "p", KeyPrefix: "P", CreatedByID: user.ID}
	require.NoError(t, db.Create(project).Error)
	col1 := &models.Column{ProjectID: project.ID, Name: "To Do", Position: 1000}
	col2 := &models.Column{ProjectID: project.ID, Name: "Done", Position: 2000}
	require.NoError(t, db.Create(col1).Error)
	require.NoError(t, db.Create(col2).Error)
	card := &models.Card{ProjectID: project.ID, ColumnID: col1.ID, Title: "T", CardNumber: 1, CreatedByID: user.ID}
	require.NoError(t, db.Create(card).Error)

	// The "created" event: no from/to column transition applies.
	created := &models.CardHistory{CardID: card.ID, UserID: user.ID, EventType: "created"}
	assert.NoError(t, db.Create(created).Error)
	assert.Nil(t, created.FromColumnID)
	assert.Nil(t, created.ToColumnID)

	// A real column move: both fields are set and must reference real columns.
	moved := &models.CardHistory{
		CardID:       card.ID,
		UserID:       user.ID,
		EventType:    "column_move",
		FromColumnID: &col1.ID,
		ToColumnID:   &col2.ID,
	}
	assert.NoError(t, db.Create(moved).Error)
}
