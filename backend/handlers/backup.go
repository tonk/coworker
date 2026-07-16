package handlers

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/services"
)

var backupCfg *config.Config

const backupsDir = "./backups"

// backupArchivePrefix is the filename prefix for the current backup format: a
// tar.gz bundling the database dump with the upload directory (attachments,
// avatars, logos). legacyBackupPrefix identifies pre-bundling backups (a bare
// .db/.sql file, database only) which are still listable/downloadable/restorable.
const (
	backupArchivePrefix = "warmdesk_backup_"
	legacyBackupPrefix  = "warmdesk_db_"
)

// InitBackup stores the config reference for the backup handler.
func InitBackup(cfg *config.Config) {
	backupCfg = cfg
}

// isBackupFile reports whether name looks like a backup file in either the
// current (tar.gz) or legacy (bare DB dump) format.
func isBackupFile(name string) bool {
	return strings.HasPrefix(name, backupArchivePrefix) || strings.HasPrefix(name, legacyBackupPrefix)
}

// backupSortKey extracts the "<timestamp>_<hex>[.ext]" portion of a backup
// filename so current and legacy formats sort chronologically together
// regardless of their differing prefixes.
func backupSortKey(name string) string {
	switch {
	case strings.HasPrefix(name, backupArchivePrefix):
		return strings.TrimPrefix(name, backupArchivePrefix)
	case strings.HasPrefix(name, legacyBackupPrefix):
		return strings.TrimPrefix(name, legacyBackupPrefix)
	default:
		return name
	}
}

// AdminBackupDatabase creates a backup and returns the filename and path. The
// backup is a tar.gz containing the database dump (SQLite VACUUM INTO,
// pg_dump, or mysqldump depending on driver) and, unless disabled via the
// backup_include_uploads setting, a copy of the upload directory (attachments,
// avatars, logos, company branding).
func AdminBackupDatabase(c *gin.Context) {
	now := time.Now().UTC()

	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		log.Printf("backup: cannot create backups directory (user=%d, ip=%s): %v", middleware.GetUserID(c), c.ClientIP(), err)
		recordBackupResult(false, now)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create backups directory"})
		return
	}

	switch backupCfg.DBDriver {
	case "sqlite", "sqlite3", "", "postgres", "postgresql", "mysql":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported database driver: " + backupCfg.DBDriver})
		return
	}

	finalFilename, backupErr := performBackup()
	if backupErr != nil {
		log.Printf("backup: failed (user=%d, ip=%s): %v", middleware.GetUserID(c), c.ClientIP(), backupErr)
		recordBackupResult(false, now)
		sendBackupEmail(false, "", backupErr.Error(), now)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "backup failed"})
		return
	}

	recordBackupResult(true, now)
	sendBackupEmail(true, finalFilename, "", now)
	log.Printf("backup: created %s (user=%d, ip=%s)", finalFilename, middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"message":  "backup created",
		"filename": finalFilename,
		"path":     filepath.Join(backupsDir, finalFilename),
	})
}

// performBackup creates the database dump, bundles it (and the upload
// directory, unless disabled) into a tar.gz under backupsDir, and returns the
// archive's filename.
func performBackup() (string, error) {
	ts := time.Now().Format("20060102_1504")
	base := backupArchivePrefix + ts + "_" + randomHex(4)

	var dbFilename string
	var err error
	switch backupCfg.DBDriver {
	case "sqlite", "sqlite3", "":
		dbFilename, err = doBackupSQLite(base)
	case "postgres", "postgresql":
		dbFilename, err = doBackupPostgres(base)
	case "mysql":
		dbFilename, err = doBackupMySQL(base)
	default:
		return "", fmt.Errorf("unsupported database driver: %s", backupCfg.DBDriver)
	}
	if err != nil {
		return "", err
	}
	dbPath := filepath.Join(backupsDir, dbFilename)
	defer os.Remove(dbPath)

	includeUploads := loadAllSettings()[settingBackupIncludeUploads] != "false"
	var uploadDir string
	if includeUploads && backupCfg.UploadDir != "" {
		uploadDir = backupCfg.UploadDir
	}

	archiveName := base + ".tar.gz"
	if err := writeBackupArchive(filepath.Join(backupsDir, archiveName), dbPath, dbFilename, uploadDir); err != nil {
		return "", err
	}
	return archiveName, nil
}

// writeBackupArchive writes a tar.gz to archivePath containing the database
// dump at dbPath (stored under "db/<dbNameInArchive>") and, if uploadDir is
// non-empty, every regular file under uploadDir (stored under "uploads/...").
func writeBackupArchive(archivePath, dbPath, dbNameInArchive, uploadDir string) error {
	out, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err := addFileToTar(tw, dbPath, "db/"+dbNameInArchive); err != nil {
		return err
	}

	if uploadDir != "" {
		if err := addDirToTar(tw, uploadDir, "uploads"); err != nil {
			return err
		}
	}

	return nil
}

func addFileToTar(tw *tar.Writer, srcPath, nameInArchive string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	hdr := &tar.Header{
		Name:    filepath.ToSlash(nameInArchive),
		Mode:    0644,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

func addDirToTar(tw *tar.Writer, srcDir, nameInArchive string) error {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil
	}
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		return addFileToTar(tw, path, filepath.Join(nameInArchive, rel))
	})
}

type BackupInfo struct {
	Filename   string `json:"filename"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
}

// AdminListBackups returns all backup files in the backups directory.
func AdminListBackups(c *gin.Context) {
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []BackupInfo
	for _, e := range entries {
		if e.IsDir() || !isBackupFile(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, BackupInfo{
			Filename:   e.Name(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
		})
	}

	// Newest first
	sort.Slice(result, func(i, j int) bool {
		return backupSortKey(result[i].Filename) > backupSortKey(result[j].Filename)
	})

	if result == nil {
		result = []BackupInfo{}
	}
	c.JSON(http.StatusOK, result)
}

// AdminRestoreBackup replaces the active database (and, for archives that
// bundle one, the upload directory) with a named backup.
func AdminRestoreBackup(c *gin.Context) {
	var body struct {
		Filename string `json:"filename"`
		Mode     string `json:"mode"` // "replace" (default, full wipe) or "merge" (add to current data)
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
		return
	}
	mode := "replace"
	if body.Mode == "merge" {
		mode = "merge"
	}

	// Sanitise: strip any path components to prevent traversal
	filename := filepath.Base(body.Filename)
	if !isBackupFile(filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup filename"})
		return
	}

	backupPath := filepath.Join(backupsDir, filename)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup file not found"})
		return
	}

	if strings.HasPrefix(filename, backupArchivePrefix) {
		restoreFromArchive(c, backupPath, filename, mode)
		return
	}

	// Legacy pre-bundling format: a bare database dump, no uploads to restore.
	if mode == "merge" {
		mergeLegacyBackup(c, backupPath, filename)
		return
	}
	if err := restoreDatabase(backupPath); err != nil {
		log.Printf("backup: restore failed (user=%d, ip=%s): %v", middleware.GetUserID(c), c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("backup: restored %s (user=%d, ip=%s)", filename, middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "database restored from " + filename, "mode": "replace"})
}

// restoreFromArchive extracts a bundled backup (database dump plus, if
// present, the upload directory) and restores or merges both depending on mode.
func restoreFromArchive(c *gin.Context, archivePath, filename, mode string) {
	tmpDir, err := os.MkdirTemp("", "warmdesk_restore_*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create temp dir: " + err.Error()})
		return
	}
	defer os.RemoveAll(tmpDir)

	dbPath, uploadsPath, err := extractBackupArchive(archivePath, tmpDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "extract failed: " + err.Error()})
		return
	}
	if dbPath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "backup archive has no database dump"})
		return
	}

	if mode == "merge" {
		mergeArchiveBackup(c, filename, dbPath, uploadsPath)
		return
	}

	if err := restoreDatabase(dbPath); err != nil {
		log.Printf("backup: restore failed (user=%d, ip=%s): %v", middleware.GetUserID(c), c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	restoredUploads := false
	if uploadsPath != "" && backupCfg != nil && backupCfg.UploadDir != "" {
		if err := replaceDir(uploadsPath, backupCfg.UploadDir); err != nil {
			log.Printf("backup: uploads restore failed (user=%d, ip=%s): %v", middleware.GetUserID(c), c.ClientIP(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database restored, but uploads restore failed: " + err.Error()})
			return
		}
		restoredUploads = true
	}

	log.Printf("backup: restored %s (uploads=%t, user=%d, ip=%s)", filename, restoredUploads, middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"message":          "database restored from " + filename,
		"mode":             "replace",
		"uploads_restored": restoredUploads,
	})
}

// mergeArchiveBackup adds the contents of an extracted archive backup (which
// may have come from a different WarmDesk server) to the current data instead
// of replacing it: uploaded files are copied in alongside existing ones, and
// (SQLite only) database rows whose primary key doesn't already exist locally
// are inserted. Rows whose ID already exists are left untouched — a genuine
// cross-server merge would need ID remapping, which this does not attempt.
func mergeArchiveBackup(c *gin.Context, filename, dbPath, uploadsPath string) {
	result := gin.H{"message": "merged data from " + filename, "mode": "merge"}

	uploadsMerged := false
	if uploadsPath != "" && backupCfg != nil && backupCfg.UploadDir != "" {
		if err := mergeDir(uploadsPath, backupCfg.UploadDir); err != nil {
			log.Printf("backup: uploads merge failed (user=%d, ip=%s): %v", middleware.GetUserID(c), c.ClientIP(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "uploads merge failed: " + err.Error()})
			return
		}
		uploadsMerged = true
	}
	result["uploads_merged"] = uploadsMerged

	switch backupCfg.DBDriver {
	case "sqlite", "sqlite3", "":
		tables, rows, err := mergeSQLiteDatabase(dbPath)
		if err != nil {
			log.Printf("backup: db merge failed (user=%d, ip=%s): %v", middleware.GetUserID(c), c.ClientIP(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database merge failed: " + err.Error()})
			return
		}
		result["db_merged"] = true
		result["tables_merged"] = tables
		result["rows_merged"] = rows
	default:
		result["db_merged"] = false
		result["db_merge_unsupported"] = true
	}

	log.Printf("backup: merged %s (uploads=%t, user=%d, ip=%s)", filename, uploadsMerged, middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, result)
}

// mergeLegacyBackup merges a legacy (pre-bundling) bare database dump — no
// uploads are involved since that format never bundled them.
func mergeLegacyBackup(c *gin.Context, backupPath, filename string) {
	switch backupCfg.DBDriver {
	case "sqlite", "sqlite3", "":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "merge mode is only supported for SQLite databases"})
		return
	}

	tables, rows, err := mergeSQLiteDatabase(backupPath)
	if err != nil {
		log.Printf("backup: db merge failed (user=%d, ip=%s): %v", middleware.GetUserID(c), c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database merge failed: " + err.Error()})
		return
	}

	log.Printf("backup: merged %s (legacy, user=%d, ip=%s)", filename, middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"message":       "merged data from " + filename,
		"mode":          "merge",
		"db_merged":     true,
		"tables_merged": tables,
		"rows_merged":   rows,
	})
}

// extractBackupArchive extracts a backup tar.gz into destDir and returns the
// path to the extracted database dump and, if the archive bundled one, the
// path to the extracted uploads directory.
func extractBackupArchive(archivePath, destDir string) (dbPath, uploadsDir string, err error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return "", "", err
	}
	defer gr.Close()

	cleanDest := filepath.Clean(destDir)
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		target := filepath.Join(destDir, filepath.FromSlash(hdr.Name))
		if !strings.HasPrefix(target, cleanDest+string(os.PathSeparator)) {
			return "", "", fmt.Errorf("invalid entry path in archive: %s", hdr.Name)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", "", err
		}
		out, err := os.Create(target)
		if err != nil {
			return "", "", err
		}
		_, copyErr := io.Copy(out, tr)
		out.Close()
		if copyErr != nil {
			return "", "", copyErr
		}

		switch {
		case strings.HasPrefix(hdr.Name, "db/"):
			dbPath = target
		case strings.HasPrefix(hdr.Name, "uploads/"):
			uploadsDir = filepath.Join(destDir, "uploads")
		}
	}
	return dbPath, uploadsDir, nil
}

// replaceDir removes dst's contents and repopulates it from src. Used to
// restore the upload directory from an extracted backup archive.
func replaceDir(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

// mergeDir copies every file under src into dst without removing anything
// already there. Used to merge an archive's uploads into the live upload
// directory — safe because upload filenames are random hex, so collisions
// with existing files are effectively impossible.
func mergeDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

// mergeSQLiteDatabase attaches otherDBPath (a SQLite file — either extracted
// from an archive or a legacy raw dump) to the live database connection and,
// for every table that exists in both schemas, inserts rows from otherDBPath
// whose primary key doesn't already exist locally. It never updates or
// deletes existing rows, so current data can't be overwritten — but rows
// whose ID collides with an unrelated existing row are skipped, not merged,
// since two independently-run servers have no shared ID space to reconcile.
func mergeSQLiteDatabase(otherDBPath string) (tablesMerged int, rowsMerged int64, err error) {
	absPath, err := filepath.Abs(otherDBPath)
	if err != nil {
		return 0, 0, err
	}
	// ATTACH DATABASE has no parameterized form for the path literal.
	escaped := strings.ReplaceAll(absPath, "'", "''")
	attachSQL := fmt.Sprintf("ATTACH DATABASE '%s' AS backup_src", escaped) // nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query -- server-generated temp path, SQLite ATTACH has no parameter binding for the path
	if err := database.DB.Exec(attachSQL).Error; err != nil {
		return 0, 0, fmt.Errorf("attach failed: %w", err)
	}
	defer database.DB.Exec("DETACH DATABASE backup_src")

	var tableNames []string
	if err := database.DB.Raw("SELECT name FROM backup_src.sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'").Scan(&tableNames).Error; err != nil {
		return 0, 0, fmt.Errorf("listing backup tables failed: %w", err)
	}

	for _, table := range tableNames {
		var mainCols, backupCols []string
		if err := database.DB.Raw("SELECT name FROM pragma_table_info(?, 'main')", table).Scan(&mainCols).Error; err != nil || len(mainCols) == 0 {
			continue // table doesn't exist in the current schema — nothing to merge into
		}
		if err := database.DB.Raw("SELECT name FROM pragma_table_info(?, 'backup_src')", table).Scan(&backupCols).Error; err != nil {
			continue
		}

		common := intersectColumnNames(mainCols, backupCols)
		if len(common) == 0 {
			continue
		}

		quotedTable := quoteSQLIdent(table)
		quotedCols := make([]string, len(common))
		for i, col := range common {
			quotedCols[i] = quoteSQLIdent(col)
		}
		colList := strings.Join(quotedCols, ", ")
		// Table/column names come from sqlite_master/pragma_table_info of the
		// attached schema, not directly from user input, and are quoted as
		// SQLite identifiers below; they cannot carry SQL beyond their own name.
		mergeSQL := fmt.Sprintf("INSERT OR IGNORE INTO main.%s (%s) SELECT %s FROM backup_src.%s", quotedTable, colList, colList, quotedTable) // nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query -- identifiers from schema introspection, quoted via quoteSQLIdent
		res := database.DB.Exec(mergeSQL)
		if res.Error != nil {
			log.Printf("backup merge: table %s skipped: %v", table, res.Error)
			continue
		}
		tablesMerged++
		rowsMerged += res.RowsAffected
	}
	return tablesMerged, rowsMerged, nil
}

// quoteSQLIdent quotes name as a SQLite identifier, doubling any embedded
// double quotes so it can't break out of the quoted identifier.
func quoteSQLIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// intersectColumnNames returns the column names present in both a and b,
// preserving a's order.
func intersectColumnNames(a, b []string) []string {
	bSet := make(map[string]bool, len(b))
	for _, c := range b {
		bSet[c] = true
	}
	var out []string
	for _, c := range a {
		if bSet[c] {
			out = append(out, c)
		}
	}
	return out
}

// AdminDownloadBackup streams a backup file as an attachment.
func AdminDownloadBackup(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	if !isBackupFile(filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup filename"})
		return
	}
	backupPath := filepath.Join(backupsDir, filename)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup file not found"})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.File(backupPath)
}

// AdminDeleteBackup removes a backup file.
func AdminDeleteBackup(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	if !isBackupFile(filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup filename"})
		return
	}

	backupPath := filepath.Join(backupsDir, filename)
	if err := os.Remove(backupPath); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// maxBackupUploadBytes caps how large an uploaded backup file may be. Backups
// can be much larger than the images handled by the generic upload endpoint,
// so this is a separate, more generous limit.
const maxBackupUploadBytes = 2 << 30 // 2 GiB

// AdminUploadBackup accepts a backup file — either one downloaded from
// another WarmDesk server, or a local copy of a previous backup — and stores
// it in the backups directory under a normalised name so it can be listed,
// downloaded, and restored/merged like any locally created backup.
func AdminUploadBackup(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	if fh.Size > maxBackupUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "backup file too large (max 2 GB)"})
		return
	}

	lowerName := strings.ToLower(fh.Filename)
	ts := time.Now().Format("20060102_1504")
	var destName string
	switch {
	case strings.HasSuffix(lowerName, ".tar.gz"):
		destName = backupArchivePrefix + ts + "_" + randomHex(4) + ".tar.gz"
	case strings.HasSuffix(lowerName, ".db"):
		destName = legacyBackupPrefix + ts + "_" + randomHex(4) + ".db"
	case strings.HasSuffix(lowerName, ".sql"):
		destName = legacyBackupPrefix + ts + "_" + randomHex(4) + ".sql"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported backup file type (expected .tar.gz, .db, or .sql)"})
		return
	}

	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create backups directory"})
		return
	}
	dest := filepath.Join(backupsDir, destName)
	if err := c.SaveUploadedFile(fh, dest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	if err := validateBackupFile(dest); err != nil {
		os.Remove(dest)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup file: " + err.Error()})
		return
	}

	log.Printf("backup: uploaded %s as %s (user=%d, ip=%s)", fh.Filename, destName, middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "backup uploaded", "filename": destName})
}

// validateBackupFile does a light sanity check on an uploaded backup so
// unrelated files don't sit undetected in the backups directory until
// restore time.
func validateBackupFile(path string) error {
	switch {
	case strings.HasSuffix(path, ".tar.gz"):
		return validateBackupArchive(path)
	case strings.HasSuffix(path, ".db"):
		return validateSQLiteFile(path)
	default:
		return nil // .sql: a plain-text dump, no reliable content check beyond the extension
	}
}

// validateBackupArchive confirms path is a readable gzip+tar archive that
// contains a database dump under "db/".
func validateBackupArchive(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("not a valid gzip file")
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("archive does not contain a database dump (no db/ entry)")
		}
		if err != nil {
			return fmt.Errorf("not a valid tar archive")
		}
		if hdr.Typeflag == tar.TypeReg && strings.HasPrefix(hdr.Name, "db/") {
			return nil
		}
	}
}

// validateSQLiteFile confirms path starts with the SQLite file magic header.
func validateSQLiteFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	header := make([]byte, 16)
	if _, err := io.ReadFull(f, header); err != nil {
		return fmt.Errorf("file too small to be a SQLite database")
	}
	if string(header) != "SQLite format 3\x00" {
		return fmt.Errorf("not a valid SQLite database file")
	}
	return nil
}

// ---- backup helpers ----

func doBackupSQLite(filename string) (string, error) {
	backupPath := filepath.Join(backupsDir, filename+".db")
	// VACUUM INTO creates a clean, compacted copy atomically (SQLite ≥ 3.27).
	// backupPath is server-generated (filepath.Join of a fixed dir + server-chosen filename).
	// SQLite does not support parameterized VACUUM INTO, so manual escaping is required.
	sql := fmt.Sprintf("VACUUM INTO '%s'", strings.ReplaceAll(backupPath, "'", "''")) // nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query -- server-generated path, SQLite has no parameterized VACUUM INTO
	if err := database.DB.Exec(sql).Error; err != nil {
		return "", fmt.Errorf("backup failed: %w", err)
	}
	return filename + ".db", nil
}

func doBackupPostgres(filename string) (string, error) {
	if _, err := exec.LookPath("pg_dump"); err != nil {
		return "", fmt.Errorf("pg_dump not found in PATH")
	}
	backupPath := filepath.Join(backupsDir, filename+".sql")
	safeDSN, pgpw := pgCredentials(backupCfg.DBDSN)
	// --clean --if-exists adds DROP statements so the dump is self-contained for restore.
	cmd := exec.Command("pg_dump", "--clean", "--if-exists", "-f", backupPath, safeDSN) //nolint:gosec
	if pgpw != "" {
		cmd.Env = append(os.Environ(), "PGPASSWORD="+pgpw)
	}
	if _, err := cmd.CombinedOutput(); err != nil {
		// Do not include command output in the error — it may contain connection details.
		return "", fmt.Errorf("pg_dump failed (check server logs for details)")
	}
	return filename + ".sql", nil
}

func doBackupMySQL(filename string) (string, error) {
	if _, err := exec.LookPath("mysqldump"); err != nil {
		return "", fmt.Errorf("mysqldump not found in PATH")
	}
	args, mysqlpw := mysqlSafeArgsAndPw(backupCfg.DBDSN)
	backupPath := filepath.Join(backupsDir, filename+".sql")
	args = append([]string{"--result-file=" + backupPath}, args...)
	cmd := exec.Command("mysqldump", args...) //nolint:gosec
	if mysqlpw != "" {
		cmd.Env = append(os.Environ(), "MYSQL_PWD="+mysqlpw)
	}
	if _, err := cmd.CombinedOutput(); err != nil {
		// Do not include command output in the error — it may contain connection details.
		return "", fmt.Errorf("mysqldump failed (check server logs for details)")
	}
	return filename + ".sql", nil
}

// recordBackupResult saves last-run timestamp and success flag to system settings.
func recordBackupResult(success bool, t time.Time) {
	saveSetting(settingBackupLastRun, t.Format(time.RFC3339))
	if success {
		saveSetting(settingBackupLastSuccess, "true")
	} else {
		saveSetting(settingBackupLastSuccess, "false")
	}
}

// sendBackupEmail sends a backup notification email if the feature is enabled.
// On success pass the created filename; on failure pass the error message via errMsg.
func sendBackupEmail(success bool, filename, errMsg string, t time.Time) {
	all := loadAllSettings()
	if all[settingBackupEmailEnabled] != "true" {
		return
	}
	to := all[settingBackupEmailAddress]
	if to == "" {
		return
	}
	emailSvc := services.GetEmailService()
	if emailSvc == nil {
		return
	}

	// Collect available backup files (newest first).
	var backupFiles []BackupInfo
	if entries, err := os.ReadDir(backupsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || !isBackupFile(e.Name()) {
				continue
			}
			info, _ := e.Info()
			bi := BackupInfo{Filename: e.Name()}
			if info != nil {
				bi.Size = info.Size()
				bi.ModifiedAt = info.ModTime().UTC().Format("2006-01-02 15:04 UTC")
			}
			backupFiles = append(backupFiles, bi)
		}
		sort.Slice(backupFiles, func(i, j int) bool {
			return backupSortKey(backupFiles[i].Filename) > backupSortKey(backupFiles[j].Filename)
		})
	}

	_, companyName, _, _ := services.GetAppInfo()
	if companyName == "" {
		companyName = "WarmDesk"
	}

	subject := companyName + " — WarmDesk backup succeeded"
	if !success {
		subject = companyName + " — WarmDesk backup failed"
	}

	htmlBody := buildBackupEmailHTML(success, filename, errMsg, t, backupFiles)
	textBody := buildBackupEmailText(success, filename, errMsg, t, companyName, backupFiles)

	go emailSvc.SendHTML(to, subject, htmlBody, textBody)
}

func buildBackupEmailText(success bool, filename, errMsg string, t time.Time, companyName string, files []BackupInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s — WarmDesk Backup Notification\n", companyName)
	b.WriteString("========================================\n\n")
	if success {
		b.WriteString("Status:    SUCCESS\n")
		fmt.Fprintf(&b, "Date/time: %s\n", t.Format("2006-01-02 15:04:05 UTC"))
		fmt.Fprintf(&b, "File:      %s\n", filename)
	} else {
		b.WriteString("Status:    FAILED\n")
		fmt.Fprintf(&b, "Date/time: %s\n", t.Format("2006-01-02 15:04:05 UTC"))
		fmt.Fprintf(&b, "Error:     %s\n", errMsg)
	}
	b.WriteString("\nAvailable backups:\n")
	if len(files) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, f := range files {
			fmt.Fprintf(&b, "  %s  (%s)\n", f.Filename, formatEmailBytes(f.Size))
		}
	}
	b.WriteString("\n-- Sent by WarmDesk\n")
	return b.String()
}

func buildBackupEmailHTML(success bool, filename, errMsg string, t time.Time, files []BackupInfo) string {
	statusColor := "#22863a"
	statusBg := "#dcffe4"
	statusText := "Backup Succeeded"
	statusIcon := "&#10003;"
	if !success {
		statusColor = "#cb2431"
		statusBg = "#ffdce0"
		statusText = "Backup Failed"
		statusIcon = "&#10007;"
	}

	var detailRows string
	if success {
		detailRows = fmt.Sprintf(
			`<tr><td style="padding:6px 12px 6px 0;color:#666;white-space:nowrap;font-size:14px">Date / time</td>`+
				`<td style="padding:6px 0;font-size:14px">%s</td></tr>`+
				`<tr><td style="padding:6px 12px 6px 0;color:#666;white-space:nowrap;font-size:14px">File</td>`+
				`<td style="padding:6px 0;font-size:14px;font-family:monospace;white-space:nowrap"><nobr>%s</nobr></td></tr>`,
			t.Format("2006-01-02 15:04:05 UTC"), filename)
	} else {
		detailRows = fmt.Sprintf(
			`<tr><td style="padding:6px 12px 6px 0;color:#666;white-space:nowrap;font-size:14px">Date / time</td>`+
				`<td style="padding:6px 0;font-size:14px">%s</td></tr>`+
				`<tr><td style="padding:6px 12px 6px 0;color:#666;white-space:nowrap;font-size:14px">Error</td>`+
				`<td style="padding:6px 0;font-size:14px;color:#cb2431">%s</td></tr>`,
			t.Format("2006-01-02 15:04:05 UTC"), errMsg)
	}

	var backupRows string
	if len(files) == 0 {
		backupRows = `<tr><td colspan="3" style="padding:8px;color:#888;font-size:13px">No backups on disk.</td></tr>`
	} else {
		for i, f := range files {
			rowBg := "#ffffff"
			if i%2 == 1 {
				rowBg = "#f8f8f8"
			}
			backupRows += fmt.Sprintf(
				`<tr>`+
					`<td style="padding:6px 8px;font-family:monospace;font-size:12px;color:#333;white-space:nowrap;background-color:%s"><nobr>%s</nobr></td>`+
					`<td style="padding:6px 8px;font-family:monospace;font-size:12px;color:#888;text-align:right;white-space:nowrap;background-color:%s">%s</td>`+
					`<td style="padding:6px 8px;font-family:monospace;font-size:12px;color:#888;white-space:nowrap;background-color:%s">%s</td></tr>`,
				rowBg, f.Filename, rowBg, formatEmailBytes(f.Size), rowBg, f.ModifiedAt,
			)
		}
	}

	bodyHTML := fmt.Sprintf(
		// Status banner
		`<tr><td style="background:%s;padding:14px 32px;text-align:center;border-bottom:1px solid rgba(0,0,0,.06)">`+
			`<span style="color:%s;font-size:17px;font-weight:bold">%s &nbsp;%s</span>`+
			`</td></tr>`+
			// Detail table
			`<tr><td style="padding:24px 32px 16px">`+
			`<table cellpadding="0" cellspacing="0" width="100%%">%s</table>`+
			`</td></tr>`+
			// Backup list
			`<tr><td style="padding:8px 32px 28px">`+
			`<div style="font-size:14px;font-weight:bold;color:#333;margin-bottom:10px">Available backups</div>`+
			`<table width="100%%" cellpadding="0" cellspacing="0" style="border:1px solid #e0e0e0;border-radius:6px;overflow:hidden">`+
			`<thead><tr>`+
			`<th style="padding:7px 8px;text-align:left;font-family:monospace;font-size:12px;font-weight:600;color:#555;border-bottom:1px solid #e0e0e0;background-color:#f5f5f5">Filename</th>`+
			`<th style="padding:7px 8px;text-align:right;font-family:monospace;font-size:12px;font-weight:600;color:#555;border-bottom:1px solid #e0e0e0;background-color:#f5f5f5">Size</th>`+
			`<th style="padding:7px 8px;text-align:left;font-family:monospace;font-size:12px;font-weight:600;color:#555;border-bottom:1px solid #e0e0e0;background-color:#f5f5f5">Date</th>`+
			`</tr></thead><tbody>%s</tbody></table>`+
			`</td></tr>`,
		statusBg, statusColor, statusIcon, statusText,
		detailRows,
		backupRows,
	)

	return services.WrapHTML("Backup Notification", bodyHTML)
}

func formatEmailBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f kB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// ---- restore helpers ----

// restoreDatabase restores the database from backupPath (a raw dump file: a
// SQLite file, or a .sql script for Postgres/MySQL) using the driver
// configured in backupCfg.
func restoreDatabase(backupPath string) error {
	switch backupCfg.DBDriver {
	case "sqlite", "sqlite3", "":
		return restoreSQLite(backupPath)
	case "postgres", "postgresql":
		return restorePostgres(backupPath)
	case "mysql":
		return restoreMySQL(backupPath)
	default:
		return fmt.Errorf("unsupported database driver: %s", backupCfg.DBDriver)
	}
}

func restoreSQLite(backupPath string) error {
	dsn := backupCfg.DBDSN
	if idx := strings.Index(dsn, "?"); idx >= 0 {
		dsn = dsn[:idx]
	}

	// Close all connections before overwriting the file.
	sqlDB, err := database.DB.DB()
	if err != nil {
		return fmt.Errorf("cannot get DB handle: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("cannot close DB: %w", err)
	}

	if err := copyFile(backupPath, dsn); err != nil {
		// Attempt to reconnect even on failure so the server stays up.
		_ = database.Init(backupCfg)
		return fmt.Errorf("copy failed: %w", err)
	}

	if err := database.Init(backupCfg); err != nil {
		return fmt.Errorf("DB reinit failed after restore: %w", err)
	}
	return nil
}

func restorePostgres(backupPath string) error {
	if _, err := exec.LookPath("psql"); err != nil {
		return fmt.Errorf("psql not found in PATH")
	}

	safeDSN, pgpw := pgCredentials(backupCfg.DBDSN)
	cmd := exec.Command("psql", safeDSN, "-f", backupPath) //nolint:gosec
	if pgpw != "" {
		cmd.Env = append(os.Environ(), "PGPASSWORD="+pgpw)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("backup: psql restore failed: %s", out)
		return fmt.Errorf("database restore failed")
	}
	return nil
}

func restoreMySQL(backupPath string) error {
	if _, err := exec.LookPath("mysql"); err != nil {
		return fmt.Errorf("mysql not found in PATH")
	}

	args, mysqlpw := mysqlSafeArgsAndPw(backupCfg.DBDSN)
	cmd := exec.Command("mysql", args...) //nolint:gosec
	if mysqlpw != "" {
		cmd.Env = append(os.Environ(), "MYSQL_PWD="+mysqlpw)
	}
	f, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("cannot open backup")
	}
	defer f.Close()
	cmd.Stdin = f

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("backup: mysql restore failed: %s", out)
		return fmt.Errorf("database restore failed")
	}
	return nil
}

// ---- utilities ----

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// pgCredentials splits a PostgreSQL DSN into a password-free DSN and the
// password itself, so pg_dump / psql can receive it via PGPASSWORD env var
// rather than a command-line argument visible in ps(1).
// Handles both URL format (postgres://user:pw@host/db) and key=value format.
func pgCredentials(dsn string) (safeDSN, password string) {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err == nil && u.User != nil {
			if pw, ok := u.User.Password(); ok {
				password = pw
				u.User = url.User(u.User.Username())
				return u.String(), password
			}
		}
		return dsn, ""
	}
	// key=value format — extract password= token
	parts := strings.Fields(dsn)
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.HasPrefix(p, "password=") {
			password = strings.Trim(strings.TrimPrefix(p, "password="), "'")
		} else {
			filtered = append(filtered, p)
		}
	}
	return strings.Join(filtered, " "), password
}

// mysqlSafeArgsAndPw parses a Go MySQL DSN (user:pass@tcp(host:port)/dbname)
// into mysql / mysqldump flags without the -p argument, returning the password
// separately so callers can pass it via MYSQL_PWD env var.
func mysqlSafeArgsAndPw(dsn string) (args []string, password string) {
	atIdx := strings.LastIndex(dsn, "@")
	if atIdx < 0 {
		return
	}
	userpass := dsn[:atIdx]
	rest := dsn[atIdx+1:]

	if colonIdx := strings.Index(userpass, ":"); colonIdx >= 0 {
		password = userpass[colonIdx+1:]
		args = append(args, "-u", userpass[:colonIdx])
	} else {
		args = append(args, "-u", userpass)
	}

	if inner, ok := strings.CutPrefix(rest, "tcp("); ok {
		if closeIdx := strings.Index(inner, ")"); closeIdx >= 0 {
			hostport := inner[:closeIdx]
			dbpart := strings.TrimPrefix(inner[closeIdx+1:], "/")
			if qIdx := strings.Index(dbpart, "?"); qIdx >= 0 {
				dbpart = dbpart[:qIdx]
			}
			if colonIdx := strings.LastIndex(hostport, ":"); colonIdx >= 0 {
				args = append(args, "-h", hostport[:colonIdx], "-P", hostport[colonIdx+1:])
			} else {
				args = append(args, "-h", hostport)
			}
			args = append(args, dbpart)
		}
	}
	return
}

// StartBackupScheduler launches a background goroutine that creates automatic
// backups based on the backup_schedule system setting. It checks every 5 minutes
// whether a backup is due and prunes old files to stay within backup_keep.
func StartBackupScheduler() {
	go func() {
		runScheduledBackupIfDue()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			runScheduledBackupIfDue()
		}
	}()
}

func runScheduledBackupIfDue() {
	all := loadAllSettings()
	interval := scheduleInterval(all[settingBackupSchedule])
	if interval == 0 {
		return
	}
	lastRunStr := all[settingBackupLastRun]
	startTime := all[settingBackupStartTime]

	if startTime != "" && isValidHHMM(startTime) {
		// Slot-based scheduling: compute the most recent past slot derived from
		// the start time and interval, then run only if last_run predates it.
		slot := mostRecentSlot(startTime, interval)
		if !slot.IsZero() {
			if lastRunStr != "" {
				lastRun, err := time.Parse(time.RFC3339, lastRunStr)
				if err == nil && !lastRun.Before(slot) {
					return
				}
			}
			performScheduledBackup()
			return
		}
	}

	// Fallback: interval-from-last-run behaviour (no start time configured).
	if lastRunStr != "" {
		lastRun, err := time.Parse(time.RFC3339, lastRunStr)
		if err == nil && time.Since(lastRun) < interval {
			return
		}
	}
	performScheduledBackup()
}

// mostRecentSlot returns the most recent past backup slot for the given
// HH:MM start time and interval. Slots repeat throughout the day starting
// from the anchor (today at HH:MM) and extending backwards as needed.
func mostRecentSlot(startTime string, interval time.Duration) time.Time {
	h, _ := strconv.Atoi(startTime[0:2])
	m, _ := strconv.Atoi(startTime[3:5])
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	anchor := midnight.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute)
	if anchor.After(now) {
		anchor = anchor.Add(-24 * time.Hour)
	}
	n := int(now.Sub(anchor) / interval)
	return anchor.Add(time.Duration(n) * interval)
}

func scheduleInterval(s string) time.Duration {
	switch s {
	case "6h":
		return 6 * time.Hour
	case "8h":
		return 8 * time.Hour
	case "12h":
		return 12 * time.Hour
	case "24h":
		return 24 * time.Hour
	}
	return 0
}

func performScheduledBackup() {
	if backupCfg == nil {
		return
	}
	now := time.Now().UTC()

	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		log.Printf("backup scheduler: cannot create backups dir: %v", err)
		recordBackupResult(false, now)
		sendBackupEmail(false, "", err.Error(), now)
		return
	}

	finalFilename, backupErr := performBackup()
	if backupErr != nil {
		log.Printf("backup scheduler: backup failed: %v", backupErr)
		recordBackupResult(false, now)
		sendBackupEmail(false, "", backupErr.Error(), now)
		return
	}

	recordBackupResult(true, now)
	sendBackupEmail(true, finalFilename, "", now)
	log.Printf("backup scheduler: created %s", finalFilename)
	pruneOldBackups()
}

func pruneOldBackups() {
	keep, err := strconv.Atoi(loadAllSettings()[settingBackupKeep])
	if err != nil || keep <= 0 {
		return
	}
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && isBackupFile(e.Name()) {
			files = append(files, e.Name())
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return backupSortKey(files[i]) < backupSortKey(files[j])
	}) // oldest first
	for len(files) > keep {
		if err := os.Remove(filepath.Join(backupsDir, files[0])); err == nil {
			log.Printf("backup scheduler: pruned %s", files[0])
		}
		files = files[1:]
	}
}

