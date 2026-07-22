// gc-uploads finds (and optionally deletes) files under upload_dir that are no
// longer referenced by any database row — orphans left behind by avatar/logo
// replacements from before that path started cleaning up after itself.
//
// Usage (run from the backend/ directory):
//
//	go run ./cmd/gc-uploads               # dry run: list orphaned files and total size
//	go run ./cmd/gc-uploads --delete       # actually remove them
//	go run ./cmd/gc-uploads --config /path/to/warmdesk.yaml
//
// A file is considered referenced if its name appears in any of: users.avatar_url,
// customers.logo_url, projects.avatar, user_groups.avatar, conversations.avatar,
// the company_logo/company_logo_dark system settings, or attachments.stored_name
// (chat/card/ticket attachments — a separate upload path that lives in the same
// directory).
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

func must(err error) {
	if err != nil {
		log.Fatalf("gc-uploads: %v", err)
	}
}

// localName returns the bare filename for a local "/uploads/xxx" URL, or ""
// for an empty/external (http(s)) value that doesn't correspond to a local file.
func localName(url string) string {
	if url == "" || !strings.HasPrefix(url, "/uploads/") {
		return ""
	}
	return filepath.Base(url)
}

func main() {
	configPath := flag.String("config", "", "path to warmdesk.yaml (optional)")
	del := flag.Bool("delete", false, "actually delete orphaned files (default: dry run, list only)")
	flag.Parse()

	cfg := config.Load(*configPath)
	cfg.DBLog = "silent"
	must(database.Init(cfg))
	db := database.DB

	uploadDir := cfg.UploadDir
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	entries, err := os.ReadDir(uploadDir)
	must(err)

	referenced := map[string]bool{}

	var avatarURLs []string
	db.Model(&models.User{}).Pluck("avatar_url", &avatarURLs)
	var logoURLs []string
	db.Model(&models.Customer{}).Pluck("logo_url", &logoURLs)
	var projectAvatars []string
	db.Model(&models.Project{}).Pluck("avatar", &projectAvatars)
	var groupAvatars []string
	db.Model(&models.UserGroup{}).Pluck("avatar", &groupAvatars)
	var convAvatars []string
	db.Model(&models.Conversation{}).Pluck("avatar", &convAvatars)
	var settingValues []string
	db.Model(&models.SystemSetting{}).Where("key IN ?", []string{"company_logo", "company_logo_dark"}).Pluck("value", &settingValues)

	for _, list := range [][]string{avatarURLs, logoURLs, projectAvatars, groupAvatars, convAvatars, settingValues} {
		for _, v := range list {
			if name := localName(v); name != "" {
				referenced[name] = true
			}
		}
	}

	var storedNames []string
	db.Model(&models.Attachment{}).Pluck("stored_name", &storedNames)
	for _, name := range storedNames {
		if name != "" {
			referenced[name] = true
		}
	}

	var orphans []string
	var totalSize int64
	for _, e := range entries {
		if e.IsDir() || referenced[e.Name()] {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		orphans = append(orphans, e.Name())
		totalSize += info.Size()
	}
	sort.Strings(orphans)

	if len(orphans) == 0 {
		fmt.Println("✓ No orphaned files found.")
		return
	}

	fmt.Printf("Found %d orphaned file(s) in %s, totalling %.2f MB:\n\n", len(orphans), uploadDir, float64(totalSize)/(1024*1024))
	for _, name := range orphans {
		fmt.Println("  " + name)
	}

	if !*del {
		fmt.Println("\nDry run — nothing deleted. Re-run with --delete to remove these files.")
		return
	}

	fmt.Println()
	removed := 0
	for _, name := range orphans {
		if err := os.Remove(filepath.Join(uploadDir, name)); err != nil {
			fmt.Printf("  ✗ failed to remove %s: %v\n", name, err)
			continue
		}
		removed++
	}
	fmt.Printf("Removed %d/%d orphaned file(s).\n", removed, len(orphans))
}
