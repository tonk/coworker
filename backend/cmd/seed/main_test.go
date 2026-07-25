package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/testutil"
)

// A regression guard for the exact failure hit by `warmdesk-seed --reset`:
// re-running the seed writes the same static setting values again, and on
// MySQL/MariaDB the old update-or-create pattern mistook "0 rows changed"
// (the value was already correct) for "no such row", then failed on a
// duplicate primary key trying to Create one that already existed.
// seedSetting now uses a single atomic INSERT ... ON CONFLICT DO UPDATE.
func TestSeedSettingResavingIdenticalValueDoesNotError(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()

	seedSetting(db, "default_columns", "Backlog\nIn Progress\nDone")
	seedSetting(db, "default_columns", "Backlog\nIn Progress\nDone")

	var row models.SystemSetting
	require.NoError(t, db.First(&row, "key = ?", "default_columns").Error)
	assert.Equal(t, "Backlog\nIn Progress\nDone", row.Value)
}

func TestSeedSettingCreatesThenUpdates(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()

	seedSetting(db, "company_name", "First Co")
	seedSetting(db, "company_name", "Second Co")

	var row models.SystemSetting
	require.NoError(t, db.First(&row, "key = ?", "company_name").Error)
	assert.Equal(t, "Second Co", row.Value)
}
