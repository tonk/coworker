package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/database"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestWriteAndExtractBackupArchiveRoundTrip(t *testing.T) {
	srcDir := t.TempDir()

	dbPath := filepath.Join(srcDir, "warmdesk_backup_20260101_0000_ab12.db")
	require.NoError(t, os.WriteFile(dbPath, []byte("fake-sqlite-content"), 0644))

	uploadDir := filepath.Join(srcDir, "uploads")
	require.NoError(t, os.MkdirAll(filepath.Join(uploadDir, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(uploadDir, "avatar1.png"), []byte("avatar-bytes"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(uploadDir, "sub", "logo.png"), []byte("logo-bytes"), 0644))

	archivePath := filepath.Join(srcDir, "warmdesk_backup_20260101_0000_ab12.tar.gz")
	require.NoError(t, writeBackupArchive(archivePath, dbPath, filepath.Base(dbPath), uploadDir))

	destDir := t.TempDir()
	extractedDBPath, extractedUploadsDir, err := extractBackupArchive(archivePath, destDir)
	require.NoError(t, err)
	require.NotEmpty(t, extractedDBPath)
	require.NotEmpty(t, extractedUploadsDir)

	dbContent, err := os.ReadFile(extractedDBPath)
	require.NoError(t, err)
	assert.Equal(t, "fake-sqlite-content", string(dbContent))

	avatarContent, err := os.ReadFile(filepath.Join(extractedUploadsDir, "avatar1.png"))
	require.NoError(t, err)
	assert.Equal(t, "avatar-bytes", string(avatarContent))

	logoContent, err := os.ReadFile(filepath.Join(extractedUploadsDir, "sub", "logo.png"))
	require.NoError(t, err)
	assert.Equal(t, "logo-bytes", string(logoContent))
}

func TestWriteBackupArchiveWithoutUploads(t *testing.T) {
	srcDir := t.TempDir()
	dbPath := filepath.Join(srcDir, "dump.db")
	require.NoError(t, os.WriteFile(dbPath, []byte("dump"), 0644))

	archivePath := filepath.Join(srcDir, "archive.tar.gz")
	require.NoError(t, writeBackupArchive(archivePath, dbPath, "dump.db", ""))

	destDir := t.TempDir()
	extractedDBPath, extractedUploadsDir, err := extractBackupArchive(archivePath, destDir)
	require.NoError(t, err)
	assert.NotEmpty(t, extractedDBPath)
	assert.Empty(t, extractedUploadsDir)
}

func TestReplaceDir(t *testing.T) {
	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "nested"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "nested", "b.txt"), []byte("b"), 0644))

	dst := t.TempDir()
	// Pre-existing file in dst that should be wiped by the restore.
	require.NoError(t, os.WriteFile(filepath.Join(dst, "stale.txt"), []byte("stale"), 0644))

	require.NoError(t, replaceDir(src, dst))

	_, err := os.Stat(filepath.Join(dst, "stale.txt"))
	assert.True(t, os.IsNotExist(err))

	content, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	require.NoError(t, err)
	assert.Equal(t, "a", string(content))

	content, err = os.ReadFile(filepath.Join(dst, "nested", "b.txt"))
	require.NoError(t, err)
	assert.Equal(t, "b", string(content))
}

func TestIsBackupFile(t *testing.T) {
	assert.True(t, isBackupFile("warmdesk_backup_20260101_0000_ab12.tar.gz"))
	assert.True(t, isBackupFile("warmdesk_db_20260101_0000_ab12.db"))
	assert.True(t, isBackupFile("warmdesk_db_20260101_0000_ab12.sql"))
	assert.False(t, isBackupFile("something_else.tar.gz"))
	assert.False(t, isBackupFile("../../etc/passwd"))
}

func TestBackupSortKeyOrdersMixedPrefixesChronologically(t *testing.T) {
	files := []string{
		"warmdesk_backup_20260301_0000_ffff.tar.gz", // newest
		"warmdesk_db_20260101_0000_aaaa.db",         // oldest (legacy format)
		"warmdesk_backup_20260201_0000_bbbb.tar.gz", // middle
	}

	// Sort oldest first, same as pruneOldBackups does.
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if backupSortKey(files[j]) < backupSortKey(files[i]) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	assert.Equal(t, []string{
		"warmdesk_db_20260101_0000_aaaa.db",
		"warmdesk_backup_20260201_0000_bbbb.tar.gz",
		"warmdesk_backup_20260301_0000_ffff.tar.gz",
	}, files)
}

func TestMergeDirDoesNotDeleteExisting(t *testing.T) {
	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "new.png"), []byte("new-file"), 0644))

	dst := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dst, "existing.png"), []byte("existing-file"), 0644))

	require.NoError(t, mergeDir(src, dst))

	existing, err := os.ReadFile(filepath.Join(dst, "existing.png"))
	require.NoError(t, err)
	assert.Equal(t, "existing-file", string(existing))

	added, err := os.ReadFile(filepath.Join(dst, "new.png"))
	require.NoError(t, err)
	assert.Equal(t, "new-file", string(added))
}

func TestQuoteSQLIdent(t *testing.T) {
	assert.Equal(t, `"simple"`, quoteSQLIdent("simple"))
	assert.Equal(t, `"has""quote"`, quoteSQLIdent(`has"quote`))
}

func TestIntersectColumnNames(t *testing.T) {
	got := intersectColumnNames([]string{"id", "name", "email"}, []string{"id", "email", "extra"})
	assert.Equal(t, []string{"id", "email"}, got)
}

func TestValidateBackupArchive(t *testing.T) {
	dir := t.TempDir()

	dbPath := filepath.Join(dir, "dump.db")
	require.NoError(t, os.WriteFile(dbPath, []byte("dump"), 0644))
	goodArchive := filepath.Join(dir, "good.tar.gz")
	require.NoError(t, writeBackupArchive(goodArchive, dbPath, "dump.db", ""))
	assert.NoError(t, validateBackupArchive(goodArchive))

	notGzip := filepath.Join(dir, "notgzip.tar.gz")
	require.NoError(t, os.WriteFile(notGzip, []byte("not a gzip file"), 0644))
	assert.Error(t, validateBackupArchive(notGzip))
}

func TestValidateSQLiteFile(t *testing.T) {
	dir := t.TempDir()

	valid := filepath.Join(dir, "valid.db")
	header := append([]byte("SQLite format 3\x00"), []byte("rest-of-file")...)
	require.NoError(t, os.WriteFile(valid, header, 0644))
	assert.NoError(t, validateSQLiteFile(valid))

	invalid := filepath.Join(dir, "invalid.db")
	require.NoError(t, os.WriteFile(invalid, []byte("not a sqlite file at all"), 0644))
	assert.Error(t, validateSQLiteFile(invalid))
}

// withTestDB temporarily points the package-level database.DB at a file-based
// SQLite connection (ATTACH DATABASE needs a real file, unlike :memory:) and
// restores the previous connection afterwards.
func withTestDB(t *testing.T, dbPath string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	previous := database.DB
	database.DB = db
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		database.DB = previous
	})
}

func TestMergeSQLiteDatabaseInsertsNewRowsAndSkipsCollidingIDs(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.db")
	backupPath := filepath.Join(dir, "backup.db")

	withTestDB(t, mainPath)
	require.NoError(t, database.DB.Exec("CREATE TABLE widgets (id INTEGER PRIMARY KEY, name TEXT)").Error)
	require.NoError(t, database.DB.Exec("INSERT INTO widgets (id, name) VALUES (1, 'existing-1'), (2, 'existing-2')").Error)

	backupDB, err := gorm.Open(sqlite.Open(backupPath), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, backupDB.Exec("CREATE TABLE widgets (id INTEGER PRIMARY KEY, name TEXT, extra TEXT)").Error)
	require.NoError(t, backupDB.Exec("INSERT INTO widgets (id, name, extra) VALUES (2, 'colliding-2', 'x'), (3, 'new-3', 'y')").Error)
	backupSQLDB, _ := backupDB.DB()
	require.NoError(t, backupSQLDB.Close())

	tables, rows, err := mergeSQLiteDatabase(backupPath)
	require.NoError(t, err)
	assert.Equal(t, 1, tables)
	assert.Equal(t, int64(1), rows) // only id=3 is new; id=2 collides and is skipped

	type widget struct {
		ID   int
		Name string
	}
	var widgets []widget
	require.NoError(t, database.DB.Raw("SELECT id, name FROM widgets ORDER BY id").Scan(&widgets).Error)
	require.Len(t, widgets, 3)
	assert.Equal(t, "existing-1", widgets[0].Name)
	assert.Equal(t, "existing-2", widgets[1].Name) // untouched, not overwritten by the colliding backup row
	assert.Equal(t, "new-3", widgets[2].Name)
}
