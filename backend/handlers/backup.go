package handlers

import (
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
	filename := "warmdesk_db_" + ts + "_" + randomHex(4)
	now := time.Now().UTC()

	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		log.Printf("backup: cannot create backups directory (user=%d, ip=%s): %v", middleware.GetUserID(c), c.ClientIP(), err)
		recordBackupResult(false, now)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create backups directory"})
		return
	}

	var finalFilename string
	var backupErr error
	switch backupCfg.DBDriver {
	case "sqlite", "sqlite3", "":
		finalFilename, backupErr = doBackupSQLite(filename)
	case "postgres", "postgresql":
		finalFilename, backupErr = doBackupPostgres(filename)
	case "mysql":
		finalFilename, backupErr = doBackupMySQL(filename)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported database driver: " + backupCfg.DBDriver})
		return
	}

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

// AdminDownloadBackup streams a backup file as an attachment.
func AdminDownloadBackup(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	if !strings.HasPrefix(filename, "warmdesk_db_") {
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
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pg_dump failed: %s", out)
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
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("mysqldump failed: %s", out)
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
			if e.IsDir() || !strings.HasPrefix(e.Name(), "warmdesk_db_") {
				continue
			}
			info, _ := e.Info()
			bi := BackupInfo{Filename: e.Name()}
			if info != nil {
				bi.Size = info.Size()
				bi.ModifiedAt = info.ModTime().UTC().Format(time.RFC3339)
			}
			backupFiles = append(backupFiles, bi)
		}
		sort.Slice(backupFiles, func(i, j int) bool {
			return backupFiles[i].Filename > backupFiles[j].Filename
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
				`<td style="padding:6px 0;font-size:14px;font-family:monospace">%s</td></tr>`,
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
				`<tr style="background:%s">`+
					`<td style="padding:6px 8px;font-family:monospace;font-size:12px;color:#333">%s</td>`+
					`<td style="padding:6px 8px;font-size:12px;color:#888;text-align:right">%s</td>`+
					`<td style="padding:6px 8px;font-size:12px;color:#888">%s</td></tr>`,
				rowBg, f.Filename, formatEmailBytes(f.Size), f.ModifiedAt,
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
			`<thead><tr style="background:#f5f5f5">`+
			`<th style="padding:7px 8px;text-align:left;font-size:12px;font-weight:600;color:#555;border-bottom:1px solid #e0e0e0">Filename</th>`+
			`<th style="padding:7px 8px;text-align:right;font-size:12px;font-weight:600;color:#555;border-bottom:1px solid #e0e0e0">Size</th>`+
			`<th style="padding:7px 8px;text-align:left;font-size:12px;font-weight:600;color:#555;border-bottom:1px solid #e0e0e0">Date</th>`+
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
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
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

	safeDSN, pgpw := pgCredentials(backupCfg.DBDSN)
	cmd := exec.Command("psql", safeDSN, "-f", backupPath) //nolint:gosec
	if pgpw != "" {
		cmd.Env = append(os.Environ(), "PGPASSWORD="+pgpw)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("backup: psql restore failed (user=%d, ip=%s): %s", middleware.GetUserID(c), c.ClientIP(), out)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database restore failed"})
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

	args, mysqlpw := mysqlSafeArgsAndPw(backupCfg.DBDSN)
	cmd := exec.Command("mysql", args...) //nolint:gosec
	if mysqlpw != "" {
		cmd.Env = append(os.Environ(), "MYSQL_PWD="+mysqlpw)
	}
	f, err := os.Open(backupPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open backup"})
		return
	}
	defer f.Close()
	cmd.Stdin = f

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("backup: mysql restore failed (user=%d, ip=%s): %s", middleware.GetUserID(c), c.ClientIP(), out)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database restore failed"})
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
	ts := time.Now().Format("20060102_1504")
	filename := "warmdesk_db_" + ts + "_" + randomHex(4)
	now := time.Now().UTC()

	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		log.Printf("backup scheduler: cannot create backups dir: %v", err)
		recordBackupResult(false, now)
		sendBackupEmail(false, "", err.Error(), now)
		return
	}

	var finalFilename string
	var backupErr error
	switch backupCfg.DBDriver {
	case "sqlite", "sqlite3", "":
		finalFilename, backupErr = doBackupSQLite(filename)
	case "postgres", "postgresql":
		finalFilename, backupErr = doBackupPostgres(filename)
	case "mysql":
		finalFilename, backupErr = doBackupMySQL(filename)
	default:
		log.Printf("backup scheduler: unsupported driver %s", backupCfg.DBDriver)
		return
	}

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

