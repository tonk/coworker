package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/mail"
	"net/smtp"
	"regexp"
	"strings"

	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
	appws "github.com/tonk/warmdesk/ws"
)

// envelopeAddress extracts the bare email address from a From value that may
// be either a plain address ("a@b.com") or RFC 5322 display-name form
// ("Display Name <a@b.com>").  smtp.SendMail requires the bare address for
// the SMTP envelope; the full string is used unchanged in the From: header.
func envelopeAddress(from string) string {
	if addr, err := mail.ParseAddress(from); err == nil {
		return addr.Address
	}
	return from
}

// smtpConfigReader is set by main to avoid an import cycle
// (services → handlers is not allowed; handlers → services is fine).
var smtpConfigReader func() config.SMTPConfig

// SetSMTPConfigReader registers the function used to read live SMTP settings.
func SetSMTPConfigReader(fn func() config.SMTPConfig) {
	smtpConfigReader = fn
}

// globalEmailService is the EmailService instance created in main.go, stored
// here so handlers can send emails without holding a direct reference.
var globalEmailService *EmailService

// SetEmailService stores the global EmailService created at startup.
func SetEmailService(s *EmailService) { globalEmailService = s }

// GetEmailService returns the global EmailService (nil if not yet initialised).
func GetEmailService() *EmailService { return globalEmailService }

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
		from = "warmdesk@localhost"
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return smtp.SendMail(addr, auth, envelopeAddress(from), []string{to}, []byte(msg))
}

// SendReply sends an email reply to a ticket's original sender with threading
// headers (Message-ID, In-Reply-To, References) so the customer's follow-up
// reply is threaded back to the same ticket by the IMAP polling service.
func (s *EmailService) SendReply(to, subject, body string, ticketID uint, inReplyTo, references string) error {
	cfg := s.cfg()
	if cfg.Host == "" {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	from := cfg.From
	if from == "" {
		from = "warmdesk@localhost"
	}

	// Generate a unique Message-ID with the ticket ID embedded for lookup
	randBytes := make([]byte, 8)
	rand.Read(randBytes) //nolint:errcheck
	domain := "warmdesk.local"
	if at := strings.LastIndex(from, "@"); at != -1 {
		if d := from[at+1:]; d != "" {
			domain = d
		}
	}
	msgID := fmt.Sprintf("<warmdesk-%d-%s@%s>", ticketID, hex.EncodeToString(randBytes), domain)

	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	fmt.Fprintf(&b, "Message-ID: %s\r\n", msgID)
	if inReplyTo != "" {
		fmt.Fprintf(&b, "In-Reply-To: %s\r\n", inReplyTo)
		fmt.Fprintf(&b, "References: %s\r\n", inReplyTo)
		if references != "" && references != inReplyTo {
			fmt.Fprintf(&b, "References: %s %s\r\n", inReplyTo, strings.Trim(references, "<>"))
		}
	}
	fmt.Fprintf(&b, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&b, "\r\n%s", body)

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return smtp.SendMail(addr, auth, envelopeAddress(from), []string{to}, []byte(b.String()))
}

// SendHTML sends a multipart/alternative email with both a plain-text fallback
// and an HTML body. Clients that support HTML will render the HTML version.
func (s *EmailService) SendHTML(to, subject, htmlBody, textBody string) error {
	cfg := s.cfg()
	if cfg.Host == "" {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	from := cfg.From
	if from == "" {
		from = "warmdesk@localhost"
	}

	boundary := "==WarmDesk_boundary_42=="
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=%q\r\n", boundary)
	fmt.Fprintf(&b, "\r\n")
	fmt.Fprintf(&b, "--%s\r\n", boundary)
	fmt.Fprintf(&b, "Content-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", textBody)
	fmt.Fprintf(&b, "--%s\r\n", boundary)
	fmt.Fprintf(&b, "Content-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n", htmlBody)
	fmt.Fprintf(&b, "--%s--\r\n", boundary)

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return smtp.SendMail(addr, auth, envelopeAddress(from), []string{to}, []byte(b.String()))
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

// NotifyMentions sends real-time WS notifications to online users and emails to offline users.
func (ns *NotificationService) NotifyMentions(body string, senderID uint, context string) {
	usernames := ExtractMentions(body)
	if len(usernames) == 0 {
		return
	}

	var sender models.User
	database.DB.First(&sender, senderID)
	senderName := sender.DisplayName
	if senderName == "" {
		senderName = sender.Username
	}

	preview := body
	if len(preview) > 120 {
		preview = preview[:120] + "..."
	}

	var users []models.User
	database.DB.Where("username IN ?", usernames).Find(&users)

	for _, u := range users {
		if u.ID == senderID {
			continue
		}
		if appws.IsUserOnline(u.ID) {
			appws.BroadcastToUser(u.ID, appws.Message{
				Type: appws.TypeMentionNotification,
				Payload: map[string]interface{}{
					"sender_name": senderName,
					"body":        preview,
					"context":     context,
				},
			})
			continue
		}
		if !u.EmailNotifications {
			continue
		}
		subject := "You were mentioned in " + context
		plainBody := fmt.Sprintf("You were mentioned by %s:\n\n%s", senderName, body)
		htmlContent := fmt.Sprintf(
			`<tr><td style="padding:28px 32px;font-size:15px;color:#333;line-height:1.6">`+
				`<p style="margin:0 0 12px"><strong>%s</strong> mentioned you in <em>%s</em>:</p>`+
				`<blockquote style="margin:0;padding:12px 16px;background:#f8f8f8;border-left:4px solid #1a5fb4;border-radius:4px;font-size:14px;color:#555">%s</blockquote>`+
				`</td></tr>`,
			senderName, context, plainBody,
		)
		go ns.email.SendHTML(u.Email, subject,
			WrapHTML("Mention Notification", htmlContent),
			WrapText("Mention Notification", plainBody))
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
	plainBody := fmt.Sprintf("%s assigned you to the card \"%s\".", assigner.DisplayName, card.Title)
	htmlContent := fmt.Sprintf(
		`<tr><td style="padding:28px 32px;font-size:15px;color:#333;line-height:1.6">`+
			`<p style="margin:0 0 8px"><strong>%s</strong> assigned you to a card:</p>`+
			`<div style="padding:12px 16px;background:#f8f8f8;border-left:4px solid #1a5fb4;border-radius:4px;font-size:14px;font-weight:bold;color:#333">%s</div>`+
			`</td></tr>`,
		assigner.DisplayName, card.Title,
	)
	go ns.email.SendHTML(assignee.Email, subject,
		WrapHTML("Card Assignment", htmlContent),
		WrapText("Card Assignment", plainBody))
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
		subject := fmt.Sprintf("New message from %s", sender.DisplayName)
		plainBody := fmt.Sprintf("%s sent you a message:\n\n%s", sender.DisplayName, preview)
		htmlContent := fmt.Sprintf(
			`<tr><td style="padding:28px 32px;font-size:15px;color:#333;line-height:1.6">`+
				`<p style="margin:0 0 12px"><strong>%s</strong> sent you a message:</p>`+
				`<blockquote style="margin:0;padding:12px 16px;background:#f8f8f8;border-left:4px solid #1a5fb4;border-radius:4px;font-size:14px;color:#555">%s</blockquote>`+
				`</td></tr>`,
			sender.DisplayName, preview,
		)
		go ns.email.SendHTML(u.Email, subject,
			WrapHTML("New Message", htmlContent),
			WrapText("New Message", plainBody))
	}
}
