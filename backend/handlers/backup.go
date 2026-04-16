package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
)

var backupCfg *config.Config

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

func adminBackupSQLite(c *gin.Context, filename string) {
	dsn := backupCfg.DBDSN
	// Strip query params (e.g. ?cache=shared&mode=rwc)
	if idx := strings.Index(dsn, "?"); idx >= 0 {
		dsn = dsn[:idx]
	}

	dbDir := filepath.Dir(dsn)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot access database directory: " + err.Error()})
		return
	}

	backupPath := filepath.Join(dbDir, filename+".db")

	// VACUUM INTO creates a clean, compacted copy atomically (SQLite ≥ 3.27).
	sql := fmt.Sprintf("VACUUM INTO '%s'", strings.ReplaceAll(backupPath, "'", "''"))
	if err := database.DB.Exec(sql).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "backup failed: " + err.Error()})
		return
	}

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

	backupDir := "./backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create backups directory: " + err.Error()})
		return
	}

	backupPath := filepath.Join(backupDir, filename+".sql")
	cmd := exec.Command("pg_dump", backupCfg.DBDSN, "-f", backupPath) //nolint:gosec
	if out, err := cmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pg_dump failed: " + string(out)})
		return
	}

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

	backupDir := "./backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create backups directory: " + err.Error()})
		return
	}

	// Parse Go MySQL DSN: user:pass@tcp(host:port)/dbname
	dsn := backupCfg.DBDSN
	args := []string{}

	// Extract credentials and database from the DSN
	if atIdx := strings.LastIndex(dsn, "@"); atIdx >= 0 {
		userpass := dsn[:atIdx]
		rest := dsn[atIdx+1:]

		if colonIdx := strings.Index(userpass, ":"); colonIdx >= 0 {
			args = append(args, "-u", userpass[:colonIdx], "-p"+userpass[colonIdx+1:])
		} else {
			args = append(args, "-u", userpass)
		}

		// rest: tcp(host:port)/dbname or unix(/path)/dbname
		if strings.HasPrefix(rest, "tcp(") {
			inner := strings.TrimPrefix(rest, "tcp(")
			if closeIdx := strings.Index(inner, ")"); closeIdx >= 0 {
				hostport := inner[:closeIdx]
				dbpart := strings.TrimPrefix(inner[closeIdx+1:], "/")
				if colonIdx := strings.LastIndex(hostport, ":"); colonIdx >= 0 {
					args = append(args, "-h", hostport[:colonIdx], "-P", hostport[colonIdx+1:])
				} else {
					args = append(args, "-h", hostport)
				}
				// Strip query params
				if qIdx := strings.Index(dbpart, "?"); qIdx >= 0 {
					dbpart = dbpart[:qIdx]
				}
				args = append(args, dbpart)
			}
		}
	}

	backupPath := filepath.Join(backupDir, filename+".sql")
	args = append([]string{"--result-file=" + backupPath}, args...)

	cmd := exec.Command("mysqldump", args...) //nolint:gosec
	if out, err := cmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mysqldump failed: " + string(out)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "backup created",
		"filename": filename + ".sql",
		"path":     backupPath,
	})
}
