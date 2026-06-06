package testutil

import (
	"github.com/glebarez/sqlite"
	"github.com/tonk/warmdesk/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB opens an in-memory SQLite database and runs AutoMigrate
// for all models. Returns the *gorm.DB and a cleanup function.
func SetupTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}
	if err := db.AutoMigrate(allModels()...); err != nil {
		panic("failed to migrate test database: " + err.Error())
	}
	return db, func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}
}

// allModels returns the same model slice the real AutoMigrate uses.
func allModels() []interface{} {
	return []interface{}{
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
		&models.ContractTimeSlot{},
		&models.CustomerFavorite{},
		&models.CustomerAccess{},
		&models.CardReference{},
		&models.Epic{},
		&models.Sprint{},
		&models.SprintCard{},
		&models.Release{},
		&models.ReleaseSprint{},
		&models.UserGroup{},
		&models.GroupMember{},
		&models.GroupProjectAccess{},
		&models.GroupCustomerAccess{},
		&models.TimeEntry{},
		&models.NewsItem{},
		&models.PasskeyCredential{},
		&models.MFATrustedDevice{},
	}
}
