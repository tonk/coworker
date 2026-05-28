package services

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"strings"
	"time"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	sasl "github.com/emersion/go-sasl"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
)

var imapConfigReader func() config.IMAPConfig

// imapOAuth2TokenRefresher is called before each poll to refresh the OAuth2
// access token if it has expired. Returns true if the token was refreshed.
var imapOAuth2TokenRefresher func() bool

// SetIMAPOAuth2TokenRefresher registers a callback that refreshes the IMAP
// OAuth2 access token when it expires.
func SetIMAPOAuth2TokenRefresher(fn func() bool) {
	imapOAuth2TokenRefresher = fn
}

// SetIMAPConfigReader registers a callback that returns live IMAP settings from
// the database. Call this from main.go after the database is initialised.
func SetIMAPConfigReader(fn func() config.IMAPConfig) {
	imapConfigReader = fn
}

// IMAPService polls an IMAP mailbox and creates helpdesk tickets from new mail.
type IMAPService struct{}

// NewIMAPService creates an IMAPService. Call StartPolling in a goroutine.
func NewIMAPService() *IMAPService {
	return &IMAPService{}
}

func (s *IMAPService) cfg() config.IMAPConfig {
	if imapConfigReader != nil {
		return imapConfigReader()
	}
	return config.IMAPConfig{}
}

// StartPolling loops forever, polling the configured mailbox at the configured
// interval. It exits when stop is closed.
func (s *IMAPService) StartPolling(stop <-chan struct{}) {
	log.Println("imap: polling service started")
	for {
		cfg := s.cfg()
		interval := time.Duration(cfg.PollInterval) * time.Second
		if interval < 30*time.Second {
			interval = 30 * time.Second
		}

		if cfg.Enabled && cfg.Host != "" {
			if err := s.poll(cfg); err != nil {
				log.Printf("imap: poll error: %v", err)
			}
		}

		select {
		case <-stop:
			log.Println("imap: polling service stopped")
			return
		case <-time.After(interval):
		}
	}
}

// connectAndLogin dials the IMAP server and authenticates. It supports two
// authentication mechanisms:
//   - "oauth2": uses SASL XOAUTH2 or OAUTHBEARER with the configured access token
//   - "plain" (default): password-based LOGIN, with STARTTLS fallback when the
//     server advertises LOGINDISABLED
func (s *IMAPService) connectAndLogin(cfg config.IMAPConfig) (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	var c *client.Client
	var err error
	if cfg.UseTLS {
		c, err = client.DialTLS(addr, &tls.Config{ServerName: cfg.Host})
	} else {
		c, err = client.Dial(addr)
	}
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", addr, err)
	}

	if cfg.AuthMechanism == "oauth2" {
		if cfg.AccessToken == "" {
			c.Logout() //nolint:errcheck
			return nil, fmt.Errorf("login: OAuth2 access token is empty")
		}
		// Try XOAUTH2 first, fall back to OAUTHBEARER
		xo2 := &xoauth2Client{username: cfg.Username, token: cfg.AccessToken}
		if err := c.Authenticate(xo2); err != nil {
			// Try OAUTHBEARER as fallback
			ob := sasl.NewOAuthBearerClient(&sasl.OAuthBearerOptions{
				Username: cfg.Username,
				Token:    cfg.AccessToken,
			})
			if authErr := c.Authenticate(ob); authErr != nil {
				c.Logout() //nolint:errcheck
				return nil, fmt.Errorf("login (oauth2): %w (xoauth2: %v)", authErr, err)
			}
		}
	} else {
		if err := c.Login(cfg.Username, cfg.Password); err != nil {
			if !cfg.UseTLS && errors.Is(err, client.ErrLoginDisabled) {
				if ok, _ := c.SupportStartTLS(); ok {
					if tlsErr := c.StartTLS(&tls.Config{ServerName: cfg.Host}); tlsErr == nil {
						if loginErr := c.Login(cfg.Username, cfg.Password); loginErr == nil {
							return c, nil
						}
					}
				}
			}
			c.Logout() //nolint:errcheck
			return nil, fmt.Errorf("login: %w", err)
		}
	}

	return c, nil
}

func (s *IMAPService) poll(cfg config.IMAPConfig) error {
	// Refresh OAuth2 token if needed before connecting
	if imapOAuth2TokenRefresher != nil && cfg.AuthMechanism == "oauth2" {
		imapOAuth2TokenRefresher()
	}
	c, err := s.connectAndLogin(cfg)
	if err != nil {
		return err
	}
	defer c.Logout() //nolint:errcheck

	mailbox := cfg.Mailbox
	if mailbox == "" {
		mailbox = "INBOX"
	}
	if _, err := c.Select(mailbox, false); err != nil {
		return fmt.Errorf("select %q: %w", mailbox, err)
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	seqNums, err := c.Search(criteria)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}
	if len(seqNums) == 0 {
		return nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}
	messages := make(chan *imap.Message, 10)
	fetchErr := make(chan error, 1)
	go func() {
		fetchErr <- c.Fetch(seqSet, items, messages)
	}()

	var processed []uint32
	for msg := range messages {
		if err := s.processMessage(msg, section); err != nil {
			log.Printf("imap: process msg %d: %v", msg.SeqNum, err)
		} else {
			processed = append(processed, msg.SeqNum)
		}
	}
	if err := <-fetchErr; err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	if len(processed) > 0 {
		markSet := new(imap.SeqSet)
		markSet.AddNum(processed...)
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.SeenFlag}
		if err := c.Store(markSet, item, flags, nil); err != nil {
			log.Printf("imap: mark seen: %v", err)
		}
	}
	return nil
}

func (s *IMAPService) processMessage(msg *imap.Message, section *imap.BodySectionName) error {
	r := msg.GetBody(section)
	if r == nil {
		return fmt.Errorf("empty body")
	}

	m, err := mail.ReadMessage(r)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	// Deduplication by Message-ID
	msgID := strings.TrimSpace(m.Header.Get("Message-Id"))
	if msgID != "" {
		var count int64
		database.DB.Model(&models.Ticket{}).Where("email_message_id = ?", msgID).Count(&count)
		if count > 0 {
			return nil // already imported
		}
	}

	fromEmail := ""
	if fromAddr, err := mail.ParseAddress(m.Header.Get("From")); err == nil && fromAddr != nil {
		fromEmail = fromAddr.Address
	}

	subject, _ := decodeRFC2047(m.Header.Get("Subject"))
	subject = strings.TrimSpace(subject)
	if subject == "" {
		subject = "(no subject)"
	}

	inReplyTo := strings.TrimSpace(m.Header.Get("In-Reply-To"))

	body, err := extractPlainText(m)
	if err != nil || body == "" {
		body = subject
	}
	body = stripQuotedReplies(body)

	// customer= directive: customer=Acme or customer=42
	customerID := parseCustomerDirective(body)

	systemUserID := systemUserID()

	// Parse References header for fallback threading
	references := strings.Fields(m.Header.Get("References"))

	// findParentTicket tries to locate an existing ticket by a message ID value.
	// If found and the ticket is in a final state (resolved/closed), it reopens it.
	findParentTicket := func(msgIDVal string) *models.Ticket {
		var parent models.Ticket
		if err := database.DB.Where("email_message_id = ?", msgIDVal).First(&parent).Error; err != nil {
			return nil
		}
		closedStatuses := map[string]bool{"resolved": true, "closed": true, "done": true, "cancelled": true}
		if closedStatuses[parent.Status] {
			log.Printf("imap: reopening ticket #%d (was %s) on customer reply", parent.ID, parent.Status)
			database.DB.Model(&parent).Update("status", "open")
			parent.Status = "open"
		}
		return &parent
	}

	// Reply threading: check In-Reply-To first, then fall back to References.
	var parent *models.Ticket
	if inReplyTo != "" {
		parent = findParentTicket(inReplyTo)
	}
	if parent == nil {
		for _, ref := range references {
			ref = strings.Trim(ref, "<>")
			if ref == "" {
				continue
			}
			parent = findParentTicket(ref)
			if parent != nil {
				break
			}
		}
	}
	if parent != nil {
		database.DB.Create(&models.TicketMessage{
			TicketID: parent.ID,
			UserID:   systemUserID,
			Body:     body,
		})
		log.Printf("imap: reply added to ticket #%d", parent.ID)
		return nil
	}

	// New ticket
	var msgIDPtr *string
	if msgID != "" {
		msgIDPtr = &msgID
	}
	var fromEmailPtr *string
	if fromEmail != "" {
		fromEmailPtr = &fromEmail
	}
	ticket := models.Ticket{
		CustomerID:     customerID,
		Title:          subject,
		Description:    body,
		Type:           "service_request",
		Status:         "new",
		Priority:       "medium",
		CreatedByID:    systemUserID,
		EmailMessageID: msgIDPtr,
		FromEmail:      fromEmailPtr,
	}
	if err := database.DB.Create(&ticket).Error; err != nil {
		return fmt.Errorf("create ticket: %w", err)
	}
	log.Printf("imap: created ticket #%d %q", ticket.ID, subject)
	return nil
}

// defaultService holds the running IMAPService so handlers can trigger polls.
var defaultService *IMAPService

func SetDefaultService(s *IMAPService) { defaultService = s }
func GetDefaultService() *IMAPService  { return defaultService }

// TestIMAPConnection tests a connection without requiring a running IMAPService.
func TestIMAPConnection(cfg config.IMAPConfig) error {
	return (&IMAPService{}).TestConnection(cfg)
}

// PollOnce runs a single poll cycle using the current live config.
func (s *IMAPService) PollOnce() error {
	cfg := s.cfg()
	if !cfg.Enabled || cfg.Host == "" {
		return fmt.Errorf("IMAP not configured or disabled")
	}
	return s.poll(cfg)
}

// TestConnection connects and logs in without fetching any messages.
func (s *IMAPService) TestConnection(cfg config.IMAPConfig) error {
	if cfg.Host == "" {
		return fmt.Errorf("host is required")
	}
	c, err := s.connectAndLogin(cfg)
	if err != nil {
		return err
	}
	c.Logout() //nolint:errcheck
	return nil
}

// systemUserID returns the ID of the first admin user, or 1 as fallback.
func systemUserID() uint {
	var u models.User
	if err := database.DB.Where("global_role = ?", "admin").Order("id asc").First(&u).Error; err == nil {
		return u.ID
	}
	return 1
}

// decodeRFC2047 decodes an RFC 2047 encoded header value (e.g. =?UTF-8?B?...?=).
func decodeRFC2047(s string) (string, error) {
	dec := new(mime.WordDecoder)
	return dec.DecodeHeader(s)
}

// extractPlainText walks a MIME message and returns the first text/plain part.
func extractPlainText(m *mail.Message) (string, error) {
	ct := m.Header.Get("Content-Type")
	if ct == "" {
		ct = "text/plain"
	}
	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		// Not valid MIME — read body directly
		b, err2 := io.ReadAll(m.Body)
		return strings.TrimSpace(string(b)), err2
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(m.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			partCT := part.Header.Get("Content-Type")
			partMedia, _, _ := mime.ParseMediaType(partCT)
			if strings.EqualFold(partMedia, "text/plain") {
				return readPartBody(part, part.Header.Get("Content-Transfer-Encoding"))
			}
		}
		return "", fmt.Errorf("no text/plain part found")
	}

	return readPartBody(m.Body, m.Header.Get("Content-Transfer-Encoding"))
}

func readPartBody(r io.Reader, encoding string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "quoted-printable":
		b, err := io.ReadAll(quotedprintable.NewReader(r))
		return strings.TrimSpace(string(b)), err
	default:
		b, err := io.ReadAll(r)
		return strings.TrimSpace(string(b)), err
	}
}

var (
	reQuotedLine = regexp.MustCompile(`(?m)^>.*$`)
	reOnWrote    = regexp.MustCompile(`(?ms)^On .{0,200}wrote:.*`)
	reTrailingNL = regexp.MustCompile(`\n{3,}`)
	reCustomer   = regexp.MustCompile(`(?i)customer\s*=\s*(\S+)`)
)

// stripQuotedReplies removes quoted reply blocks from an email body.
func stripQuotedReplies(s string) string {
	s = reOnWrote.ReplaceAllString(s, "")
	s = reQuotedLine.ReplaceAllString(s, "")
	s = reTrailingNL.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

// parseCustomerDirective looks for "customer=<name or id>" in the body.
// Returns the customer ID if found, or nil.
func parseCustomerDirective(body string) *uint {
	m := reCustomer.FindStringSubmatch(body)
	if m == nil {
		return nil
	}
	value := strings.TrimSpace(m[1])
	if value == "" {
		return nil
	}
	// Try numeric ID first
	var id uint
	if _, err := fmt.Sscanf(value, "%d", &id); err == nil && id > 0 {
		var c models.Customer
		if database.DB.First(&c, id).Error == nil {
			return &c.ID
		}
	}
	// Try by name (case-insensitive)
	var c models.Customer
	if database.DB.Where("lower(name) = lower(?)", value).First(&c).Error == nil {
		return &c.ID
	}
	return nil
}
