package database

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) error {
	var dialector gorm.Dialector

	switch cfg.DBDriver {
	case "mysql":
		dsn, err := applyMySQLTLS(cfg)
		if err != nil {
			return err
		}
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn, err := applyPostgresTLS(cfg)
		if err != nil {
			return err
		}
		dialector = postgres.Open(dsn)
	default:
		dialector = sqlite.Open(cfg.DBDSN)
	}

	logLevel := logger.Info
	switch cfg.DBLog {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	log.Printf("Connected to %s database", cfg.DBDriver)

	if err := deduplicateKeyPrefixes(db); err != nil {
		return err
	}
	if err := autoMigrate(db); err != nil {
		return err
	}
	return backfillCardNumbers(db)
}

// buildTLSConfig builds a *tls.Config from the db_tls_* config fields.
// Returns nil when TLS is disabled or not configured.
func buildTLSConfig(cfg *config.Config) (*tls.Config, error) {
	mode := strings.ToLower(cfg.DBTLSMode)
	if mode == "" || mode == "disable" {
		return nil, nil
	}

	tlsCfg := &tls.Config{
		InsecureSkipVerify: mode == "require",
	}

	if cfg.DBTLSCACert != "" {
		pem, err := os.ReadFile(cfg.DBTLSCACert)
		if err != nil {
			return nil, fmt.Errorf("db tls: read CA cert %q: %w", cfg.DBTLSCACert, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("db tls: no valid certificates found in CA cert %q", cfg.DBTLSCACert)
		}
		tlsCfg.RootCAs = pool
		tlsCfg.InsecureSkipVerify = false
	}

	if cfg.DBTLSCert != "" && cfg.DBTLSKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.DBTLSCert, cfg.DBTLSKey)
		if err != nil {
			return nil, fmt.Errorf("db tls: load client cert/key: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return tlsCfg, nil
}

// applyPostgresTLS appends sslmode and cert parameters to the PostgreSQL DSN.
func applyPostgresTLS(cfg *config.Config) (string, error) {
	mode := strings.ToLower(cfg.DBTLSMode)
	if mode == "" || mode == "disable" {
		return cfg.DBDSN, nil
	}

	// pgx accepts both key=value and URL-style DSNs; use simple appending for
	// key=value style and query-param appending for URL style.
	dsn := cfg.DBDSN
	sep := " "
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		if strings.Contains(dsn, "?") {
			sep = "&"
		} else {
			sep = "?"
		}
	}

	param := func(k, v string) {
		dsn += sep + k + "=" + v
		sep = "&"
		if !strings.HasPrefix(cfg.DBDSN, "postgres://") && !strings.HasPrefix(cfg.DBDSN, "postgresql://") {
			sep = " "
		}
	}

	param("sslmode", mode)
	if cfg.DBTLSCACert != "" {
		param("sslrootcert", cfg.DBTLSCACert)
	}
	if cfg.DBTLSCert != "" {
		param("sslcert", cfg.DBTLSCert)
	}
	if cfg.DBTLSKey != "" {
		param("sslkey", cfg.DBTLSKey)
	}

	return dsn, nil
}

// applyMySQLTLS registers a named TLS config with the MySQL driver and appends
// the tls= parameter to the DSN.
func applyMySQLTLS(cfg *config.Config) (string, error) {
	tlsCfg, err := buildTLSConfig(cfg)
	if err != nil {
		return "", err
	}
	if tlsCfg == nil {
		return cfg.DBDSN, nil
	}

	const tlsName = "warmdesk"
	if err := mysqldriver.RegisterTLSConfig(tlsName, tlsCfg); err != nil {
		return "", fmt.Errorf("db tls: register MySQL TLS config: %w", err)
	}

	dsn := cfg.DBDSN
	if strings.Contains(dsn, "?") {
		dsn += "&tls=" + tlsName
	} else {
		dsn += "?tls=" + tlsName
	}
	return dsn, nil
}

// deduplicateKeyPrefixes ensures every project has a unique, non-empty
// key_prefix before AutoMigrate adds the unique index.
//
// The tricky case is a database that predates the key_prefix column entirely:
// AutoMigrate would add the column (DEFAULT '') and the unique index in one
// step, which fails immediately because every existing row gets ''.
//
// To handle that we add the column ourselves — without the unique index — if
// it is not already present, then fill and deduplicate, and finally let
// AutoMigrate add the unique index on clean data.
func deduplicateKeyPrefixes(db *gorm.DB) error {
	// Fresh database — projects table doesn't exist yet; AutoMigrate will create
	// it with key_prefix correctly, so nothing to do here.
	if !db.Migrator().HasTable(&models.Project{}) {
		return nil
	}

	// If the column doesn't exist yet, add it without the unique index so we
	// can populate it before AutoMigrate tries to create the constraint.
	if !db.Migrator().HasColumn(&models.Project{}, "key_prefix") {
		log.Println("key_prefix column missing — adding without unique index for pre-migration backfill")
		if err := db.Exec("ALTER TABLE projects ADD COLUMN key_prefix VARCHAR(10) NOT NULL DEFAULT ''").Error; err != nil {
			return fmt.Errorf("add key_prefix column: %w", err)
		}
	}

	type row struct {
		ID        uint
		Name      string
		KeyPrefix string
	}

	// Pass 1: deduplicate existing non-empty prefixes (oldest project keeps its prefix).
	// Includes soft-deleted projects — the unique index covers all rows in the table.
	var withPrefix []row
	db.Unscoped().Model(&models.Project{}).
		Where("key_prefix != '' AND key_prefix IS NOT NULL").
		Order("id asc").
		Select("id, name, key_prefix").
		Scan(&withPrefix)

	seen := make(map[string]bool, len(withPrefix))
	for _, p := range withPrefix {
		base := p.KeyPrefix
		candidate := base
		n := 2
		for seen[candidate] {
			candidate = fmt.Sprintf("%s%d", base, n)
			n++
		}
		seen[candidate] = true
		if candidate != p.KeyPrefix {
			log.Printf("key_prefix dedup: project %d %q → %q", p.ID, p.KeyPrefix, candidate)
			db.Unscoped().Model(&models.Project{}).Where("id = ?", p.ID).UpdateColumn("key_prefix", candidate)
		}
	}

	// Pass 2: assign prefixes to projects that have none (empty or NULL).
	// Includes soft-deleted projects — all rows must be unique before AutoMigrate
	// creates the index.
	var withoutPrefix []row
	db.Unscoped().Model(&models.Project{}).
		Where("key_prefix = '' OR key_prefix IS NULL").
		Order("id asc").
		Select("id, name, key_prefix").
		Scan(&withoutPrefix)

	for _, p := range withoutPrefix {
		base := generateKeyPrefix(p.Name)
		candidate := base
		n := 2
		for seen[candidate] {
			candidate = fmt.Sprintf("%s%d", base, n)
			n++
		}
		seen[candidate] = true
		log.Printf("key_prefix backfill: project %d %q → %q", p.ID, p.Name, candidate)
		db.Unscoped().Model(&models.Project{}).Where("id = ?", p.ID).UpdateColumn("key_prefix", candidate)
	}

	return nil
}

// backfillCardNumbers assigns key_prefix to projects and card_number to existing cards
// that were created before this feature was added (card_number == 0).
func backfillCardNumbers(db *gorm.DB) error {
	// Backfill key_prefix for projects that don't have one, ensuring uniqueness.
	var projects []models.Project
	db.Where("key_prefix = '' OR key_prefix IS NULL").Find(&projects)
	for _, p := range projects {
		prefix := uniqueKeyPrefix(db, p.Name, p.ID)
		db.Model(&p).UpdateColumn("key_prefix", prefix)
	}

	// Find projects that have unnumbered cards
	var projectIDs []uint
	db.Model(&models.Card{}).Where("card_number = 0").Distinct("project_id").Pluck("project_id", &projectIDs)

	for _, pid := range projectIDs {
		var cards []models.Card
		db.Where("project_id = ? AND card_number = 0", pid).Order("created_at asc, id asc").Find(&cards)
		if len(cards) == 0 {
			continue
		}

		// Get the current max card_number for this project (from already-numbered cards)
		var maxNum int
		db.Model(&models.Card{}).Where("project_id = ? AND card_number > 0", pid).
			Select("COALESCE(MAX(card_number), 0)").Scan(&maxNum)

		counter := maxNum
		for _, card := range cards {
			counter++
			db.Model(&card).UpdateColumn("card_number", counter)
		}
		// Sync the project counter
		db.Model(&models.Project{}).Where("id = ?", pid).UpdateColumn("card_counter", counter)
	}
	return nil
}

// uniqueKeyPrefix generates a key_prefix for projectID that is not already
// used by another project. excludeID is the project being assigned (so we
// don't collide with ourselves).
func uniqueKeyPrefix(db *gorm.DB, name string, excludeID uint) string {
	base := generateKeyPrefix(name)
	candidate := base
	n := 2
	for {
		var count int64
		db.Model(&models.Project{}).
			Where("key_prefix = ? AND id != ?", candidate, excludeID).
			Count(&count)
		if count == 0 {
			return candidate
		}
		candidate = fmt.Sprintf("%s%d", base, n)
		n++
	}
}

func generateKeyPrefix(name string) string {
	var words [][]rune
	var current []rune
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current = append(current, unicode.ToUpper(r))
		} else if len(current) > 0 {
			words = append(words, current)
			current = nil
		}
	}
	if len(current) > 0 {
		words = append(words, current)
	}

	var result []rune
	for _, w := range words {
		if len(result) >= 3 {
			break
		}
		result = append(result, w[0])
	}
	if len(result) < 3 && len(words) > 0 {
		for i := 1; i < len(words[0]) && len(result) < 3; i++ {
			result = append(result, words[0][i])
		}
	}
	for len(result) < 3 {
		result = append(result, 'X')
	}
	return string(result[:3])
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.ProjectMember{},
		&models.Column{},
		&models.Card{},
		&models.CardLabel{},
		&models.CardComment{},
		&models.Label{},
		&models.CardHistory{},
		&models.ChatMessage{},
		&models.DirectMessage{},
		&models.Conversation{},
		&models.ConversationMember{},
		&models.ConversationMessage{},
		&models.SystemSetting{},
		&models.StarredProject{},
		&models.APIKey{},
		&models.Attachment{},
		&models.MessageReaction{},
		&models.ProjectWebhook{},
		&models.CardTag{},
		&models.CardAssignee{},
		&models.CardChecklistItem{},
		&models.Topic{},
		&models.TopicReply{},
		&models.FavoriteUser{},
		&models.CardLink{},
		&models.Customer{},
		&models.Contract{},
		&models.CustomerFavorite{},
		&models.CustomerAccess{},
		&models.CardReference{},
		&models.Sprint{},
		&models.SprintCard{},
		&models.UserGroup{},
		&models.GroupMember{},
		&models.GroupProjectAccess{},
		&models.GroupCustomerAccess{},
		&models.TimeEntry{},
	)
}
