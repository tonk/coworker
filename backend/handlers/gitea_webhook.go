package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/coworker/database"
	"github.com/tonk/coworker/models"
	appws "github.com/tonk/coworker/ws"
)

// giteaUser is the common user object in Gitea payloads.
type giteaUser struct {
	Login     string `json:"login"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
}

func (u giteaUser) display() string {
	if u.FullName != "" {
		return u.FullName
	}
	return u.Login
}

// Minimal payload structs — only the fields we actually use.

type giteaRepository struct {
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

type giteaCommit struct {
	ID      string    `json:"id"`
	Message string    `json:"message"`
	URL     string    `json:"url"`
	Author  giteaUser `json:"author"`
}

func (c giteaCommit) shortID() string {
	if len(c.ID) > 7 {
		return c.ID[:7]
	}
	return c.ID
}

func (c giteaCommit) firstLine() string {
	msg := strings.TrimSpace(c.Message)
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		return msg[:i]
	}
	return msg
}

type giteaIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	HTMLURL string `json:"html_url"`
	User   giteaUser `json:"user"`
}

type giteaPullRequest struct {
	Number  int       `json:"number"`
	Title   string    `json:"title"`
	HTMLURL string    `json:"html_url"`
	User    giteaUser `json:"user"`
	Head    struct {
		Label string `json:"label"`
	} `json:"head"`
	Base struct {
		Label string `json:"label"`
	} `json:"base"`
}

type giteaRelease struct {
	TagName string    `json:"tag_name"`
	Name    string    `json:"name"`
	HTMLURL string    `json:"html_url"`
	Author  giteaUser `json:"author"`
}

// giteaPayload is a union of all relevant Gitea event fields.
type giteaPayload struct {
	Ref        string          `json:"ref"`
	Before     string          `json:"before"`
	After      string          `json:"after"`
	Commits    []giteaCommit   `json:"commits"`
	HeadCommit *giteaCommit    `json:"head_commit"`
	Pusher     giteaUser       `json:"pusher"`
	Sender     giteaUser       `json:"sender"`
	Repository giteaRepository `json:"repository"`
	Action     string          `json:"action"`
	Issue      *giteaIssue     `json:"issue"`
	PullRequest *giteaPullRequest `json:"pull_request"`
	Comment    *struct {
		Body    string    `json:"body"`
		HTMLURL string    `json:"html_url"`
		User    giteaUser `json:"user"`
	} `json:"comment"`
	Release *giteaRelease `json:"release"`
	// create/delete events
	RefType     string `json:"ref_type"`
}

func formatGiteaEvent(event string, p giteaPayload) string {
	repo := p.Repository.FullName
	repoLink := p.Repository.HTMLURL

	switch event {
	case "push":
		branch := p.Ref
		if strings.HasPrefix(branch, "refs/heads/") {
			branch = branch[len("refs/heads/"):]
		}
		pusher := p.Pusher.display()
		if pusher == "" {
			pusher = p.Sender.display()
		}
		n := len(p.Commits)
		commitWord := "commit"
		if n != 1 {
			commitWord = "commits"
		}
		lines := []string{
			fmt.Sprintf("**%s** pushed %d %s to **%s** ([%s](%s))",
				pusher, n, commitWord, branch, repo, repoLink),
		}
		// List up to 5 commits
		limit := n
		if limit > 5 {
			limit = 5
		}
		for i := 0; i < limit; i++ {
			c := p.Commits[i]
			lines = append(lines, fmt.Sprintf("• [`%s`](%s) %s", c.shortID(), c.URL, c.firstLine()))
		}
		if n > 5 {
			lines = append(lines, fmt.Sprintf("• … and %d more", n-5))
		}
		return strings.Join(lines, "\n")

	case "issues":
		if p.Issue == nil {
			return ""
		}
		action := p.Action
		return fmt.Sprintf("**%s** %s issue [#%d %s](%s) in [%s](%s)",
			p.Sender.display(), action, p.Issue.Number, p.Issue.Title,
			p.Issue.HTMLURL, repo, repoLink)

	case "issue_comment":
		if p.Issue == nil || p.Comment == nil {
			return ""
		}
		body := p.Comment.Body
		if len(body) > 200 {
			body = body[:200] + "…"
		}
		return fmt.Sprintf("**%s** commented on issue [#%d %s](%s) in [%s](%s)\n> %s",
			p.Comment.User.display(), p.Issue.Number, p.Issue.Title,
			p.Comment.HTMLURL, repo, repoLink, body)

	case "pull_request":
		if p.PullRequest == nil {
			return ""
		}
		action := p.Action
		pr := p.PullRequest
		return fmt.Sprintf("**%s** %s pull request [#%d %s](%s) in [%s](%s)",
			p.Sender.display(), action, pr.Number, pr.Title, pr.HTMLURL, repo, repoLink)

	case "pull_request_review_comment":
		if p.PullRequest == nil || p.Comment == nil {
			return ""
		}
		body := p.Comment.Body
		if len(body) > 200 {
			body = body[:200] + "…"
		}
		pr := p.PullRequest
		return fmt.Sprintf("**%s** reviewed [PR #%d %s](%s) in [%s](%s)\n> %s",
			p.Comment.User.display(), pr.Number, pr.Title, p.Comment.HTMLURL, repo, repoLink, body)

	case "create":
		return fmt.Sprintf("**%s** created %s **%s** in [%s](%s)",
			p.Sender.display(), p.RefType, p.Ref, repo, repoLink)

	case "delete":
		return fmt.Sprintf("**%s** deleted %s **%s** in [%s](%s)",
			p.Sender.display(), p.RefType, p.Ref, repo, repoLink)

	case "release":
		if p.Release == nil {
			return ""
		}
		r := p.Release
		name := r.Name
		if name == "" {
			name = r.TagName
		}
		return fmt.Sprintf("**%s** released [%s](%s) in [%s](%s)",
			r.Author.display(), name, r.HTMLURL, repo, repoLink)

	case "fork":
		return fmt.Sprintf("**%s** forked [%s](%s)", p.Sender.display(), repo, repoLink)

	default:
		return fmt.Sprintf("**%s** event from [%s](%s)", event, repo, repoLink)
	}
}

// verifyGiteaSignature checks the HMAC-SHA256 signature from Gitea/Forgejo.
// Returns true if no secret is stored (token itself acts as the secret),
// or if the signature matches.
func verifyGiteaSignature(secret string, body []byte, sig string) bool {
	if sig == "" {
		// Gitea sends signature only when a secret is configured on the webhook.
		// If no sig header is present we allow it through (the token in the URL
		// already authenticates the request).
		return true
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

// IncomingGiteaWebhook POST /api/v1/gitea-webhook/:token (public)
// Handles Gitea and Forgejo webhook events.
func IncomingGiteaWebhook(c *gin.Context) {
	token := c.Param("token")

	var hook models.ProjectWebhook
	if err := database.DB.Where("token = ? AND type = ?", token, "gitea").First(&hook).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Verify HMAC signature if Gitea/Forgejo sent one.
	sig := c.GetHeader("X-Gitea-Signature")
	if sig == "" {
		sig = c.GetHeader("X-Forgejo-Signature")
	}
	if !verifyGiteaSignature(token, body, sig) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Determine event type.
	event := c.GetHeader("X-Gitea-Event")
	if event == "" {
		event = c.GetHeader("X-Forgejo-Event")
	}
	if event == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing event header"})
		return
	}

	var payload giteaPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	text := formatGiteaEvent(event, payload)
	if text == "" {
		// Unknown or empty event — acknowledge without posting.
		c.JSON(http.StatusOK, gin.H{"ok": true, "skipped": true})
		return
	}

	msg := models.ChatMessage{
		ProjectID: hook.ProjectID,
		UserID:    0,
		Body:      text,
		IsBot:     true,
		BotName:   hook.Name,
	}
	if err := database.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to post message"})
		return
	}

	appws.BroadcastToProject(hook.ProjectID, appws.Message{
		Type:    appws.TypeChatMessageCreated,
		Payload: msg,
	})

	c.JSON(http.StatusCreated, gin.H{"ok": true})
}
