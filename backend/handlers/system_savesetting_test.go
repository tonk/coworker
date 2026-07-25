package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/testutil"
)

// A regression guard for two compounding bugs in the old "try an Update,
// then Create if nothing happened" upsert pattern, both specific to MySQL/
// MariaDB and invisible against SQLite (used by the rest of this suite):
//
//  1. A raw string Where("key = ?", key) condition bypasses GORM's per-
//     dialect identifier quoting. "key" is a reserved word in MySQL/
//     MariaDB, so the generated UPDATE failed there with a syntax error on
//     every call — silently, since the error was never checked.
//  2. Even after quoting the column correctly, MySQL's UPDATE reports
//     RowsAffected as rows *changed*, not rows *matched* — re-saving a
//     setting with the value it already has reports 0 rows affected
//     despite the row existing. "RowsAffected == 0" was wrongly read as
//     "no such row", and the fallback Create then failed on a duplicate
//     primary key. SQLite and PostgreSQL both report rows *matched*.
//
// saveSetting now uses a single atomic INSERT ... ON CONFLICT DO UPDATE,
// which has neither problem by construction.
func TestSaveSettingUpdatesExistingKeyOnSubsequentCalls(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	saveSetting("login_branding_enabled", "false")
	saveSetting("login_branding_enabled", "true")

	var brandingRow models.SystemSetting
	require.NoError(t, db.First(&brandingRow, "key = ?", "login_branding_enabled").Error)
	assert.Equal(t, "true", brandingRow.Value)

	// A key saved for the first time should also work (Create path).
	saveSetting("company_logo", "/uploads/abc123.png")
	var logoRow models.SystemSetting
	require.NoError(t, db.First(&logoRow, "key = ?", "company_logo").Error)
	assert.Equal(t, "/uploads/abc123.png", logoRow.Value)
}

// Re-saving a setting with the exact value it already has (e.g. re-running
// the seed script, or an admin clicking Save without changing anything)
// must not error — this is precisely the scenario the RowsAffected-as-
// "rows changed" MySQL behavior broke under the old update-or-create pattern.
func TestSaveSettingResavingIdenticalValueDoesNotError(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	saveSetting("default_columns", "Backlog\nIn Progress\nDone")
	saveSetting("default_columns", "Backlog\nIn Progress\nDone")
	saveSetting("default_columns", "Backlog\nIn Progress\nDone")

	var row models.SystemSetting
	require.NoError(t, db.First(&row, "key = ?", "default_columns").Error)
	assert.Equal(t, "Backlog\nIn Progress\nDone", row.Value)
}
