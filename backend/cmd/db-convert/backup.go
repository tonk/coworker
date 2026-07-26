// This file handles --backup: taking a downloaded WarmDesk backup file
// directly (as produced by handlers/backup.go) instead of requiring the user
// to extract it and figure out --src-driver/--src-dsn themselves.
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// resolveBackupSource makes path's database dump available on local disk and
// reports where, along with an uploads/ directory if the backup bundled one.
// path may be a current-format archive (warmdesk_backup_*.tar.gz, bundling
// db/ + uploads/) or a legacy bare dump (warmdesk_db_*.db/.sql, database
// only) — see the "Backup / restore" section of CLAUDE.md for the format.
// The returned cleanup func removes any temp directory created for
// extraction; call it (via defer) once the dump is no longer needed.
func resolveBackupSource(path string) (dbPath, uploadsDir string, cleanup func(), err error) {
	noop := func() {}
	lower := strings.ToLower(path)
	if !strings.HasSuffix(lower, ".tar.gz") && !strings.HasSuffix(lower, ".tgz") {
		// Legacy bare dump file — nothing to extract, use it directly.
		return path, "", noop, nil
	}

	tmpDir, err := os.MkdirTemp("", "db-convert-backup-*")
	if err != nil {
		return "", "", noop, fmt.Errorf("create temp dir: %w", err)
	}
	cleanup = func() { os.RemoveAll(tmpDir) }

	dbPath, uploadsDir, err = extractArchive(path, tmpDir)
	if err != nil {
		cleanup()
		return "", "", noop, err
	}
	if dbPath == "" {
		cleanup()
		return "", "", noop, fmt.Errorf("archive has no db/ dump entry — is %q a WarmDesk backup?", path)
	}
	return dbPath, uploadsDir, cleanup, nil
}

// extractArchive mirrors handlers/backup.go's extractBackupArchive: db/* goes
// to dbPath, uploads/* (if present) to destDir/uploads.
func extractArchive(archivePath, destDir string) (dbPath, uploadsDir string, err error) {
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

// detectDumpDriver identifies which engine produced a database dump file, so
// --backup doesn't need an explicit --src-driver hint. A SQLite dump (from
// doBackupSQLite's VACUUM INTO) is a real, directly-openable database file;
// pg_dump/mysqldump dumps are plain SQL scripts identified by their standard
// header comment, falling back to file extension if that's not present.
func detectDumpDriver(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	head := buf[:n]

	if bytes.HasPrefix(head, []byte("SQLite format 3\x00")) {
		return "sqlite", nil
	}
	switch {
	case bytes.Contains(head, []byte("PostgreSQL database dump")):
		return "postgres", nil
	case bytes.Contains(head, []byte("MySQL dump")), bytes.Contains(head, []byte("MariaDB dump")):
		// mysqldump's own header says "MySQL dump"; on many self-hosted Linux
		// boxes the mysqldump binary is actually MariaDB's build, which says
		// "MariaDB dump" instead — both target the same "mysql" driver here.
		return "mysql", nil
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".db", ".sqlite", ".sqlite3":
		return "sqlite", nil
	}
	return "", fmt.Errorf("cannot determine database driver for dump %q — expected a SQLite file, or a pg_dump/mysqldump .sql script with its usual header comment", path)
}

// importSQLDump loads a pg_dump/mysqldump .sql script into dsn, an
// already-running scratch instance of the matching engine — there's no way
// to read a flat SQL script as a live database, so a real instance of that
// engine has to exist somewhere reachable first. Mirrors
// handlers/backup.go's restorePostgres/restoreMySQL (same CLI tools, same
// DSN-parsing approach), just importing a foreign dump instead of the app's
// own.
func importSQLDump(driver, dsn, dumpPath string) error {
	switch driver {
	case "postgres":
		if _, err := exec.LookPath("psql"); err != nil {
			return fmt.Errorf("psql not found in PATH (required to import a postgres dump)")
		}
		safeDSN, pw := pgCredentials(dsn)
		cmd := exec.Command("psql", safeDSN, "-f", dumpPath) //nolint:gosec
		if pw != "" {
			cmd.Env = append(os.Environ(), "PGPASSWORD="+pw)
		}
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("psql import failed: %s", out)
		}
		return nil
	case "mysql":
		if _, err := exec.LookPath("mysql"); err != nil {
			return fmt.Errorf("mysql not found in PATH (required to import a mysql dump)")
		}
		args, pw := mysqlArgsAndPw(dsn)
		f, err := os.Open(dumpPath)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd := exec.Command("mysql", args...) //nolint:gosec
		cmd.Stdin = f
		if pw != "" {
			cmd.Env = append(os.Environ(), "MYSQL_PWD="+pw)
		}
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("mysql import failed: %s", out)
		}
		return nil
	default:
		return fmt.Errorf("unsupported dump driver: %s", driver)
	}
}

// pgCredentials splits a postgres DSN into a password-free form (safe to put
// in a process argument list, visible via ps) plus the password itself,
// passed separately via PGPASSWORD.
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

// mysqlArgsAndPw parses a Go MySQL DSN (user:pass@tcp(host:port)/dbname) into
// mysql CLI flags without -p, returning the password separately for MYSQL_PWD.
func mysqlArgsAndPw(dsn string) (args []string, password string) {
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

// copyTree copies every regular file under src into dst, preserving relative
// paths. Used for the backup archive's bundled uploads/ folder — filenames
// there are random hex (see CLAUDE.md's File uploads section), so unlike a
// database there's no merge logic needed, just a collision-free file copy.
func copyTree(src, dst string) (int, error) {
	n := 0
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		n++
		return nil
	})
	return n, err
}
