package services

import (
	"fmt"
	"net/smtp"
	"regexp"
	"strings"

	"github.com/tonk/coworker/config"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/models"
	appws "github.com/tonk/coworker/ws"
)

// smtpConfigReader is set by main to avoid an import cycle
// (services → handlers is not allowed; handlers → services is fine).
var smtpConfigReader func() config.SMTPConfig

// SetSMTPConfigReader registers the function used to read live SMTP settings.
func SetSMTPConfigReader(fn func() config.SMTPConfig) {
	smtpConfigReader = fn
}

var mentionRe = regexp.MustCompile(`@([a-zA-Z0-9_]+)`)

// EmailService sends SMTP emails, reading configuration dynamically so admin
// changes take effect without a server restart.
type EmailService struct {
	fallback config.SMTPConfig // used when smtpConfigReader is not set
}

// NewEmailService creates an EmailService with a fallback config (from the YAML file).
func NewEmailService(cfg config.SMTPConfig) *EmailService {
	return &EmailService{fallback: cfg}
}

func (s *EmailService) cfg() config.SMTPConfig {
	if smtpConfigReader != nil {
		return smtpConfigReader()
	}
	return s.fallback
}

func (s *EmailService) enabled() bool {
	return s.cfg().Host != ""
}

func (s *EmailService) Send(to, subject, body string) error {
	cfg := s.cfg()
	if cfg.Host == "" {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	from := cfg.From
	if from == "" {
		from = "coworker@localhost"
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

// NotificationService sends in-app and email notifications.
type NotificationService struct {
	email *EmailService
}

func NewNotificationService(email *EmailService) *NotificationService {
	return &NotificationService{email: email}
}

// ExtractMentions returns unique usernames found in @username patterns.
func ExtractMentions(body string) []string {
	matches := mentionRe.FindAllStringSubmatch(body, -1)
	seen := map[string]bool{}
	var names []string
	for _, m := range matches {
		if !seen[m[1]] {
			seen[m[1]] = true
			names = append(names, m[1])
		}
	}
	return names
}

// NotifyMentions sends email notifications to users mentioned in a message body.
func (ns *NotificationService) NotifyMentions(body string, senderID uint, context string) {
	usernames := ExtractMentions(body)
	if len(usernames) == 0 {
		return
	}

	var users []models.User
	database.DB.Where("username IN ?", usernames).Find(&users)

	for _, u := range users {
		if u.ID == senderID {
			continue
		}
		if !u.EmailNotifications {
			continue
		}
		if appws.IsUserOnline(u.ID) {
			continue
		}
		go ns.email.Send(u.Email, "You were mentioned in "+context,
			fmt.Sprintf("You were mentioned:\n\n%s", body))
	}
}

// NotifyCardAssignment sends an email when a card is assigned to a user.
func (ns *NotificationService) NotifyCardAssignment(card models.Card, assignee models.User, assigner models.User) {
	if assignee.ID == assigner.ID {
		return
	}
	if !assignee.EmailNotifications {
		return
	}
	if appws.IsUserOnline(assignee.ID) {
		return
	}
	subject := fmt.Sprintf("Card assigned: %s", card.Title)
	body := fmt.Sprintf("%s assigned you to the card \"%s\".", assigner.DisplayName, card.Title)
	go ns.email.Send(assignee.Email, subject, body)
}

// NotifyNewDM sends email notifications to DM conversation members who are offline.
func (ns *NotificationService) NotifyNewDM(msg models.ConversationMessage, sender models.User) {
	var memberIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", msg.ConversationID).
		Pluck("user_id", &memberIDs)

	for _, uid := range memberIDs {
		if uid == sender.ID {
			continue
		}
		var u models.User
		if err := database.DB.First(&u, uid).Error; err != nil {
			continue
		}
		if !u.EmailNotifications {
			continue
		}
		if appws.IsUserOnline(uid) {
			continue
		}
		preview := msg.Body
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		preview = strings.TrimSpace(preview)
		go ns.email.Send(u.Email,
			fmt.Sprintf("New message from %s", sender.DisplayName),
			fmt.Sprintf("%s sent you a message:\n\n%s", sender.DisplayName, preview))
	}
}
