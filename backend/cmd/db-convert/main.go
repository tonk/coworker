// db-convert copies every WarmDesk table from one database to another,
// across any combination of the three supported drivers (sqlite, postgres,
// mysql). It is meant for one-off moves — e.g. restoring a production SQLite
// backup onto a MySQL-driven test server — since the built-in backup/restore
// system (handlers/backup.go) always restores using the destination server's
// own configured driver and has no cross-driver conversion path.
//
// It does NOT copy upload_dir (attachments, avatars, logos) — copy that
// directory separately (e.g. the "uploads/" folder inside a WarmDesk backup
// archive) to the destination server's own upload_dir.
//
// Usage (run from the backend/ directory):
//
//	# extract the backup archive first, e.g.:
//	#   tar xzf warmdesk_backup_20260726_0201_97704e2a.tar.gz
//	#   -> db/warmdesk.db (or db/dump.sql for postgres/mysql sources)
//
//	go run ./cmd/db-convert \
//	  --src-driver sqlite --src-dsn ./db/warmdesk.db \
//	  --dst-driver mysql  --dst-dsn "user:pass@tcp(host:3306)/warmdesk?charset=utf8mb4&parseTime=True&loc=Local"
//
// The destination is schema-migrated automatically (the same GORM
// AutoMigrate the server runs on every startup) before any data is copied,
// so it can point at a brand-new empty database. Pass --truncate to delete
// existing rows in each destination table first, making the run repeatable.
package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func must(err error) {
	if err != nil {
		log.Fatalf("db-convert: %v", err)
	}
}

// tableOps is the pair of generic operations db-convert needs for one model
// type T, captured as closures so they can be stored in a plain slice
// (Go generics can't be used directly as slice element types across
// different T without this kind of erasure).
type tableOps struct {
	truncate func(dst *gorm.DB) (table string, err error)
	copy     func(src, dst *gorm.DB, batchSize int) (table string, rows int, err error)
}

func opsFor[T any]() tableOps {
	return tableOps{
		truncate: truncateTable[T],
		copy: func(src, dst *gorm.DB, batchSize int) (string, int, error) {
			return copyTable[T](src, dst, batchSize)
		},
	}
}

// jobs mirrors database.AutoMigrate's model list and order — parent tables
// before the tables that reference them via foreign key. The copy pass below
// runs this list forward; the truncate pass runs it in reverse (children
// before parents), so both are FK-safe even on a destination where FK checks
// couldn't be disabled (see disableFKChecks — guaranteed on MySQL/SQLite,
// best-effort on Postgres which requires superuser to fully suppress them).
var jobs = []tableOps{
	opsFor[models.User](),
	opsFor[models.Project](),
	opsFor[models.ProjectMember](),
	opsFor[models.Column](),
	opsFor[models.Card](),
	opsFor[models.CardLabel](),
	opsFor[models.CardComment](),
	opsFor[models.Label](),
	opsFor[models.CardHistory](),
	opsFor[models.ChatMessage](),
	opsFor[models.DirectMessage](),
	opsFor[models.Conversation](),
	opsFor[models.ConversationMember](),
	opsFor[models.ConversationMessage](),
	opsFor[models.SystemSetting](),
	opsFor[models.StarredProject](),
	opsFor[models.APIKey](),
	opsFor[models.Attachment](),
	opsFor[models.MessageReaction](),
	opsFor[models.ProjectWebhook](),
	opsFor[models.CardTag](),
	opsFor[models.CardAssignee](),
	opsFor[models.CardChecklistItem](),
	opsFor[models.Topic](),
	opsFor[models.TopicReply](),
	opsFor[models.FavoriteUser](),
	opsFor[models.CardLink](),
	opsFor[models.Customer](),
	opsFor[models.Contract](),
	opsFor[models.ContractTimeSlot](),
	opsFor[models.CustomerFavorite](),
	opsFor[models.CustomerAccess](),
	opsFor[models.CardReference](),
	opsFor[models.Epic](),
	opsFor[models.Sprint](),
	opsFor[models.SprintCard](),
	opsFor[models.Release](),
	opsFor[models.ReleaseSprint](),
	opsFor[models.UserGroup](),
	opsFor[models.GroupMember](),
	opsFor[models.GroupProjectAccess](),
	opsFor[models.GroupCustomerAccess](),
	opsFor[models.TimeEntry](),
	opsFor[models.TimeEntryRowOrder](),
	opsFor[models.TimeEntryWeekRowOrder](),
	opsFor[models.TimeMacroLibrary](),
	opsFor[models.NewsItem](),
	opsFor[models.PasskeyCredential](),
	opsFor[models.MFATrustedDevice](),
	opsFor[models.Ticket](),
	opsFor[models.TicketTag](),
	opsFor[models.TicketLink](),
	opsFor[models.TicketCardLink](),
	opsFor[models.TicketMessage](),
	opsFor[models.TicketHistory](),
	opsFor[models.TicketView](),
	opsFor[models.SlaPolicy](),
	opsFor[models.Macro](),
	opsFor[models.TicketChecklistTemplate](),
	opsFor[models.TicketChecklistItem](),
	opsFor[models.LoginHistory](),
	opsFor[models.Invoice](),
	opsFor[models.InvoiceTemplate](),
	opsFor[models.CustomerContact](),
}

// truncateTable deletes every row of T from dst. Unscoped is required: models
// with a gorm.DeletedAt field would otherwise turn this into a soft-delete
// (UPDATE ... SET deleted_at) that leaves the rows — and their primary
// keys — in place, which then collides with the copy pass's inserts.
func truncateTable[T any](dst *gorm.DB) (string, error) {
	var zero T
	name, _, _, err := tableMeta(dst, &zero)
	if err != nil {
		return "", fmt.Errorf("inspect schema: %w", err)
	}
	if err := dst.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&zero).Error; err != nil {
		return name, fmt.Errorf("truncate %s: %w", name, err)
	}
	return name, nil
}

// copyTable reads every row of T from src and inserts it into dst, preserving
// primary keys.
func copyTable[T any](src, dst *gorm.DB, batchSize int) (string, int, error) {
	var zero T
	name, pkCol, autoIncrement, err := tableMeta(dst, &zero)
	if err != nil {
		return "", 0, fmt.Errorf("inspect schema: %w", err)
	}

	// Unscoped so soft-deleted rows (models with a gorm.DeletedAt field) are
	// copied too — a conversion should faithfully mirror the source, not
	// silently drop its soft-deleted history.
	var rows []T
	if err := src.Unscoped().Find(&rows).Error; err != nil {
		return name, 0, fmt.Errorf("read %s: %w", name, err)
	}
	if len(rows) == 0 {
		return name, 0, nil
	}
	if err := dst.CreateInBatches(&rows, batchSize).Error; err != nil {
		return name, 0, fmt.Errorf("write %s: %w", name, err)
	}

	// Postgres sequences don't advance for explicit-PK inserts (unlike
	// MySQL's AUTO_INCREMENT and SQLite's rowid, which both adapt
	// automatically) — fix up so the app's next nextval()-based insert
	// doesn't collide with an imported ID.
	if autoIncrement && dst.Dialector.Name() == "postgres" {
		q := fmt.Sprintf(
			`SELECT setval(pg_get_serial_sequence('%s', '%s'), COALESCE((SELECT MAX(%s) FROM %s), 1), true)`,
			name, pkCol, pkCol, name,
		)
		if err := dst.Exec(q).Error; err != nil {
			return name, len(rows), fmt.Errorf("reset sequence for %s: %w", name, err)
		}
	}

	return name, len(rows), nil
}

// tableMeta resolves a model's table name and, if it has a single
// auto-incrementing primary key (composite-key join tables like
// StarredProject don't), that column's name.
func tableMeta(db *gorm.DB, model any) (table, pkCol string, autoIncrement bool, err error) {
	sch, err := schema.Parse(model, &sync.Map{}, db.NamingStrategy)
	if err != nil {
		return "", "", false, err
	}
	table = sch.Table
	if f := sch.PrioritizedPrimaryField; f != nil {
		pkCol = f.DBName
		autoIncrement = f.AutoIncrement
	}
	return table, pkCol, autoIncrement, nil
}

// disableFKChecks suppresses foreign-key enforcement on dst for the duration
// of the import. Reliable on MySQL and SQLite (plain session settings); on
// Postgres it requires superuser — if that fails we log a warning and fall
// back to jobs' parent-before-child / child-before-parent ordering, which is
// FK-safe on its own.
func disableFKChecks(db *gorm.DB) {
	switch db.Dialector.Name() {
	case "mysql":
		db.Exec("SET FOREIGN_KEY_CHECKS=0")
	case "sqlite":
		db.Exec("PRAGMA foreign_keys = OFF")
	case "postgres":
		if err := db.Exec("SET session_replication_role = 'replica'").Error; err != nil {
			log.Printf("warning: could not disable FK checks on postgres (needs superuser): %v — relying on insert/truncate ordering instead", err)
		}
	}
}

func reenableFKChecks(db *gorm.DB) {
	switch db.Dialector.Name() {
	case "mysql":
		db.Exec("SET FOREIGN_KEY_CHECKS=1")
	case "sqlite":
		db.Exec("PRAGMA foreign_keys = ON")
	case "postgres":
		db.Exec("SET session_replication_role = 'origin'")
	}
}

func main() {
	srcDriver := flag.String("src-driver", "", "source database driver: sqlite | postgres | mysql")
	srcDSN := flag.String("src-dsn", "", "source database DSN")
	dstDriver := flag.String("dst-driver", "", "destination database driver: sqlite | postgres | mysql")
	dstDSN := flag.String("dst-dsn", "", "destination database DSN")
	truncate := flag.Bool("truncate", false, "delete existing rows in each destination table before importing (makes the run repeatable)")
	batchSize := flag.Int("batch-size", 200, "rows per insert batch")
	flag.Parse()

	if *srcDriver == "" || *srcDSN == "" || *dstDriver == "" || *dstDSN == "" {
		log.Fatal("db-convert: --src-driver, --src-dsn, --dst-driver, and --dst-dsn are all required (see the file header comment for usage)")
	}

	srcDB, err := database.Open(&config.Config{DBDriver: *srcDriver, DBDSN: *srcDSN, DBLog: "silent"})
	must(err)

	dstDB, err := database.Open(&config.Config{DBDriver: *dstDriver, DBDSN: *dstDSN, DBLog: "silent"})
	must(err)

	log.Printf("Migrating destination schema (%s)...", *dstDriver)
	must(database.AutoMigrate(dstDB))

	disableFKChecks(dstDB)
	defer reenableFKChecks(dstDB)

	if *truncate {
		log.Print("Truncating destination tables (children before parents)...")
		for i := len(jobs) - 1; i >= 0; i-- {
			name, err := jobs[i].truncate(dstDB)
			must(err)
			log.Printf("  cleared %s", name)
		}
	}

	log.Print("Copying tables (parents before children)...")
	total := 0
	for _, j := range jobs {
		name, n, err := j.copy(srcDB, dstDB, *batchSize)
		must(err)
		if n > 0 {
			log.Printf("  %-32s %6d rows", name, n)
			total += n
		}
	}
	log.Printf("Done — copied %d rows from %s to %s.", total, *srcDriver, *dstDriver)
}
