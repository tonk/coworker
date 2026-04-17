package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
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
)

var backupCfg *config.Config

const backupsDir = "./backups"

// InitBackup stores the config reference for the backup handler.
func InitBackup(cfg *config.Config) {
	backupCfg = cfg
}

// AdminBackupDatabase creates a database backup and returns the filename and path.
// For SQLite: uses VACUUM INTO for an atomic online backup.
// For PostgreSQL: runs pg_dump.
// For MySQL: runs mysqldump.
func AdminBackupDatabase(c *gin.Context) {
	ts := time.Now().Format("20060102_1504")
	filename := "warmdesk_db_" + ts

	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create backups directory: " + err.Error()})
		return
	}

	switch backupCfg.DBDriver {
	case "sqlite", "sqlite3", "":
		adminBackupSQLite(c, filename)
	case "postgres", "postgresql":
		adminBackupPostgres(c, filename)
	case "mysql":
		adminBackupMySQL(c, filename)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported database driver: " + backupCfg.DBDriver})
	}
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
		if e.IsDir() || !strings.HasPrefix(e.Name(), "warmdesk_db_") {
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
		return result[i].Filename > result[j].Filename
	})

	if result == nil {
		result = []BackupInfo{}
	}
	c.JSON(http.StatusOK, result)
}

// AdminRestoreBackup replaces the active database with a named backup.
func AdminRestoreBackup(c *gin.Context) {
	var body struct {
		Filename string `json:"filename"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
		return
	}

	// Sanitise: strip any path components to prevent traversal
	filename := filepath.Base(body.Filename)
	if !strings.HasPrefix(filename, "warmdesk_db_") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup filename"})
		return
	}

	backupPath := filepath.Join(backupsDir, filename)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup file not found"})
		return
	}

	switch backupCfg.DBDriver {
	case "sqlite", "sqlite3", "":
		adminRestoreSQLite(c, backupPath)
	case "postgres", "postgresql":
		adminRestorePostgres(c, backupPath)
	case "mysql":
		adminRestoreMySQL(c, backupPath)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported database driver: " + backupCfg.DBDriver})
	}
}

// AdminDeleteBackup removes a backup file.
func AdminDeleteBackup(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	if !strings.HasPrefix(filename, "warmdesk_db_") {
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

// ---- backup helpers ----

func adminBackupSQLite(c *gin.Context, filename string) {
	backupPath := filepath.Join(backupsDir, filename+".db")

	// VACUUM INTO creates a clean, compacted copy atomically (SQLite ≥ 3.27).
	sql := fmt.Sprintf("VACUUM INTO '%s'", strings.ReplaceAll(backupPath, "'", "''"))
	if err := database.DB.Exec(sql).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "backup failed: " + err.Error()})
		return
	}

	log.Printf("backup: created %s (sqlite, user=%d, ip=%s)", filename+".db", middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"message":  "backup created",
		"filename": filename + ".db",
		"path":     backupPath,
	})
}

func adminBackupPostgres(c *gin.Context, filename string) {
	if _, err := exec.LookPath("pg_dump"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pg_dump not found in PATH"})
		return
	}

	backupPath := filepath.Join(backupsDir, filename+".sql")
	// --clean --if-exists adds DROP statements so the dump is self-contained for restore.
	cmd := exec.Command("pg_dump", "--clean", "--if-exists", "-f", backupPath, backupCfg.DBDSN) //nolint:gosec
	if out, err := cmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pg_dump failed: " + string(out)})
		return
	}

	log.Printf("backup: created %s (postgres, user=%d, ip=%s)", filename+".sql", middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"message":  "backup created",
		"filename": filename + ".sql",
		"path":     backupPath,
	})
}

func adminBackupMySQL(c *gin.Context, filename string) {
	if _, err := exec.LookPath("mysqldump"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mysqldump not found in PATH"})
		return
	}

	args := mysqlDumpArgs(backupCfg.DBDSN)
	backupPath := filepath.Join(backupsDir, filename+".sql")
	args = append([]string{"--result-file=" + backupPath}, args...)

	cmd := exec.Command("mysqldump", args...) //nolint:gosec
	if out, err := cmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mysqldump failed: " + string(out)})
		return
	}

	log.Printf("backup: created %s (mysql, user=%d, ip=%s)", filename+".sql", middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"message":  "backup created",
		"filename": filename + ".sql",
		"path":     backupPath,
	})
}

// ---- restore helpers ----

func adminRestoreSQLite(c *gin.Context, backupPath string) {
	dsn := backupCfg.DBDSN
	if idx := strings.Index(dsn, "?"); idx >= 0 {
		dsn = dsn[:idx]
	}

	// Close all connections before overwriting the file.
	sqlDB, err := database.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get DB handle: " + err.Error()})
		return
	}
	if err := sqlDB.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot close DB: " + err.Error()})
		return
	}

	if err := copyFile(backupPath, dsn); err != nil {
		// Attempt to reconnect even on failure so the server stays up.
		_ = database.Init(backupCfg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "copy failed: " + err.Error()})
		return
	}

	if err := database.Init(backupCfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB reinit failed after restore: " + err.Error()})
		return
	}

	log.Printf("backup: restored %s (sqlite, user=%d, ip=%s)", filepath.Base(backupPath), middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "database restored from " + filepath.Base(backupPath)})
}

func adminRestorePostgres(c *gin.Context, backupPath string) {
	if _, err := exec.LookPath("psql"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "psql not found in PATH"})
		return
	}

	cmd := exec.Command("psql", backupCfg.DBDSN, "-f", backupPath) //nolint:gosec
	if out, err := cmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "psql restore failed: " + string(out)})
		return
	}

	log.Printf("backup: restored %s (postgres, user=%d, ip=%s)", filepath.Base(backupPath), middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "database restored from " + filepath.Base(backupPath)})
}

func adminRestoreMySQL(c *gin.Context, backupPath string) {
	if _, err := exec.LookPath("mysql"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mysql not found in PATH"})
		return
	}

	args := mysqlClientArgs(backupCfg.DBDSN)
	cmd := exec.Command("mysql", args...) //nolint:gosec
	f, err := os.Open(backupPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open backup: " + err.Error()})
		return
	}
	defer f.Close()
	cmd.Stdin = f

	if out, err := cmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mysql restore failed: " + string(out)})
		return
	}

	log.Printf("backup: restored %s (mysql, user=%d, ip=%s)", filepath.Base(backupPath), middleware.GetUserID(c), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "database restored from " + filepath.Base(backupPath)})
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

// mysqlDumpArgs parses a Go MySQL DSN (user:pass@tcp(host:port)/dbname) into
// mysqldump flags.
func mysqlDumpArgs(dsn string) []string {
	return mysqlArgs(dsn)
}

// mysqlClientArgs parses a Go MySQL DSN into mysql client flags.
func mysqlClientArgs(dsn string) []string {
	return mysqlArgs(dsn)
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
	if lastRunStr != "" {
		lastRun, err := time.Parse(time.RFC3339, lastRunStr)
		if err == nil && time.Since(lastRun) < interval {
			return
		}
	}
	performScheduledBackup()
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
	ts := time.Now().Format("20060102_1504")
	filename := "warmdesk_db_" + ts

	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		log.Printf("backup scheduler: cannot create backups dir: %v", err)
		return
	}

	var finalFilename string

	switch backupCfg.DBDriver {
	case "sqlite", "sqlite3", "":
		backupPath := filepath.Join(backupsDir, filename+".db")
		sql := fmt.Sprintf("VACUUM INTO '%s'", strings.ReplaceAll(backupPath, "'", "''"))
		if err := database.DB.Exec(sql).Error; err != nil {
			log.Printf("backup scheduler: sqlite backup failed: %v", err)
			return
		}
		finalFilename = filename + ".db"
	case "postgres", "postgresql":
		if _, err := exec.LookPath("pg_dump"); err != nil {
			log.Printf("backup scheduler: pg_dump not found in PATH")
			return
		}
		backupPath := filepath.Join(backupsDir, filename+".sql")
		cmd := exec.Command("pg_dump", "--clean", "--if-exists", "-f", backupPath, backupCfg.DBDSN) //nolint:gosec
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("backup scheduler: pg_dump failed: %s", out)
			return
		}
		finalFilename = filename + ".sql"
	case "mysql":
		if _, err := exec.LookPath("mysqldump"); err != nil {
			log.Printf("backup scheduler: mysqldump not found in PATH")
			return
		}
		backupPath := filepath.Join(backupsDir, filename+".sql")
		args := append([]string{"--result-file=" + backupPath}, mysqlDumpArgs(backupCfg.DBDSN)...)
		cmd := exec.Command("mysqldump", args...) //nolint:gosec
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("backup scheduler: mysqldump failed: %s", out)
			return
		}
		finalFilename = filename + ".sql"
	default:
		log.Printf("backup scheduler: unsupported driver %s", backupCfg.DBDriver)
		return
	}

	saveSetting(settingBackupLastRun, time.Now().UTC().Format(time.RFC3339))
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
		if !e.IsDir() && strings.HasPrefix(e.Name(), "warmdesk_db_") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files) // oldest first (timestamp in filename)
	for len(files) > keep {
		if err := os.Remove(filepath.Join(backupsDir, files[0])); err == nil {
			log.Printf("backup scheduler: pruned %s", files[0])
		}
		files = files[1:]
	}
}

func mysqlArgs(dsn string) []string {
	var args []string
	atIdx := strings.LastIndex(dsn, "@")
	if atIdx < 0 {
		return args
	}
	userpass := dsn[:atIdx]
	rest := dsn[atIdx+1:]

	if colonIdx := strings.Index(userpass, ":"); colonIdx >= 0 {
		args = append(args, "-u", userpass[:colonIdx], "-p"+userpass[colonIdx+1:])
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
	return args
}
