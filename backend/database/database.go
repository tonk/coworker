package database

import (
	"fmt"
	"log"

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

	return autoMigrate(db)
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
	)
}
