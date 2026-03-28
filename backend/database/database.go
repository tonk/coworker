package database

import (
	"fmt"
	"log"
	"unicode"

	"github.com/tonk/coworker/config"
	"github.com/tonk/coworker/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) error {
	var dialector gorm.Dialector

	switch cfg.DBDriver {
	case "mysql":
		dialector = mysql.Open(cfg.DBDSN)
	case "postgres":
		dialector = postgres.Open(cfg.DBDSN)
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

	if err := autoMigrate(db); err != nil {
		return err
	}
	return backfillCardNumbers(db)
}

// backfillCardNumbers assigns key_prefix to projects and card_number to existing cards
// that were created before this feature was added (card_number == 0).
func backfillCardNumbers(db *gorm.DB) error {
	// Backfill key_prefix for projects that don't have one
	var projects []models.Project
	db.Where("key_prefix = '' OR key_prefix IS NULL").Find(&projects)
	for _, p := range projects {
		prefix := generateKeyPrefix(p.Name)
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
	)
}
