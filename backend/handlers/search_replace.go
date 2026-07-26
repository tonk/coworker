package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
	"github.com/tonk/warmdesk/ws"
)

// replaceableTypes whitelists which search result types search-replace may
// target. Chat messages are excluded: WarmDesk has no message-edit capability
// for project chat at all (only sender-or-project-owner delete), so there is
// no existing single-record rule to reuse for it.
var replaceableTypes = map[string]bool{
	"card":         true,
	"card_comment": true,
	"dm_message":   true,
	"ticket":       true,
	"time_entry":   true,
}

// replaceAllCI replaces all case-insensitive occurrences of old in s with
// replacement, preserving the original casing of everything outside the
// match, and returns the replacement count. SQLite's LIKE (the project's
// default DB) is already case-insensitive for ASCII, so a case-sensitive
// replace could match a row via the DB query yet change nothing at all -
// case-insensitive replacement guarantees anything the DB surfaced is
// actually found and replaced.
func replaceAllCI(s, old, replacement string) (string, int) {
	if old == "" {
		return s, 0
	}
	lowerS := strings.ToLower(s)
	lowerOld := strings.ToLower(old)
	var b strings.Builder
	count := 0
	i := 0
	for {
		idx := strings.Index(lowerS[i:], lowerOld)
		if idx < 0 {
			b.WriteString(s[i:])
			break
		}
		idx += i
		b.WriteString(s[i:idx])
		b.WriteString(replacement)
		i = idx + len(old)
		count++
	}
	return b.String(), count
}

// snippet returns a short excerpt of s centered on the first case-insensitive
// occurrence of match, for compact before/after previews.
func snippet(s, match string) string {
	const radius = 40
	if match == "" || len(s) <= 2*radius {
		return s
	}
	idx := strings.Index(strings.ToLower(s), strings.ToLower(match))
	if idx < 0 {
		return s[:2*radius] + "…"
	}
	start := idx - radius
	prefix := ""
	if start < 0 {
		start = 0
	} else {
		prefix = "…"
	}
	end := idx + len(match) + radius
	suffix := ""
	if end >= len(s) {
		end = len(s)
	} else {
		suffix = "…"
	}
	return prefix + s[start:end] + suffix
}

func containsCI(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

// --- per-type edit-permission rules, mirroring each entity's existing
// single-record edit handler exactly (see handlers/card.go UpdateCard,
// handlers/card_comment.go UpdateComment, handlers/conversation.go
// EditConversationMessage, handlers/ticket.go UpdateTicket, and
// handlers/time_entry.go UpdateTimeEntry). ---

func cardEditable(card models.Card, userID uint, role string) (bool, string) {
	if err := services.RequireProjectRole(card.ProjectID, userID, role, "member"); err != nil {
		return false, "you need member access to this project to edit cards"
	}
	return true, ""
}

func cardCommentEditable(comment models.CardComment, projectID uint, userID uint, role string) (bool, string) {
	if err := services.RequireProjectRole(projectID, userID, role, "viewer"); err != nil {
		return false, "you don't have access to this project"
	}
	if comment.UserID != userID {
		return false, "you can only edit your own comments"
	}
	return true, ""
}

func dmMessageEditable(msg models.ConversationMessage, userID uint) (bool, string) {
	if msg.SenderID != userID {
		return false, "you can only edit your own messages"
	}
	return true, ""
}

func ticketEditable(customerID uint, userID uint, role string) (bool, string) {
	if err := requireCustomerAccess(customerID, userID, role); err != nil {
		return false, "you don't have access to this customer"
	}
	if err := requireNotCustomerRole(role); err != nil {
		return false, "customer-role users can't edit tickets"
	}
	return true, ""
}

func timeEntryEditable(entry models.TimeEntry, userID uint) (bool, string) {
	if entry.UserID != userID {
		return false, "you can only edit your own time entries"
	}
	return true, ""
}

type replaceCandidate struct {
	Type     string `json:"type"`
	ID       uint   `json:"id"`
	Field    string `json:"field"`
	Before   string `json:"before"`
	After    string `json:"after"`
	Editable bool   `json:"editable"`
	Reason   string `json:"reason,omitempty"`
}

func cardCandidates(scope searchScope, pattern, raw, replacement string) []replaceCandidate {
	var out []replaceCandidate
	for _, r := range searchCards(scope, pattern) {
		card := r.Item.(models.Card)
		editable, reason := cardEditable(card, scope.userID, scope.globalRole)
		if containsCI(card.Title, raw) {
			after, _ := replaceAllCI(card.Title, raw, replacement)
			out = append(out, replaceCandidate{Type: "card", ID: card.ID, Field: "title",
				Before: snippet(card.Title, raw), After: snippet(after, replacement), Editable: editable, Reason: reason})
		}
		if containsCI(card.Description, raw) {
			after, _ := replaceAllCI(card.Description, raw, replacement)
			out = append(out, replaceCandidate{Type: "card", ID: card.ID, Field: "description",
				Before: snippet(card.Description, raw), After: snippet(after, replacement), Editable: editable, Reason: reason})
		}
	}
	return out
}

func cardCommentCandidates(scope searchScope, pattern, raw, replacement string) []replaceCandidate {
	var out []replaceCandidate
	for _, r := range searchCardComments(scope, pattern) {
		comment := r.Item.(models.CardComment)
		var card models.Card
		database.DB.Select("project_id").First(&card, comment.CardID)
		editable, reason := cardCommentEditable(comment, card.ProjectID, scope.userID, scope.globalRole)
		after, _ := replaceAllCI(comment.Body, raw, replacement)
		out = append(out, replaceCandidate{Type: "card_comment", ID: comment.ID, Field: "body",
			Before: snippet(comment.Body, raw), After: snippet(after, replacement), Editable: editable, Reason: reason})
	}
	return out
}

func dmMessageCandidates(scope searchScope, pattern, raw, replacement string) []replaceCandidate {
	var out []replaceCandidate
	for _, r := range searchDMMessages(scope, pattern) {
		msg := r.Item.(models.ConversationMessage)
		editable, reason := dmMessageEditable(msg, scope.userID)
		after, _ := replaceAllCI(msg.Body, raw, replacement)
		out = append(out, replaceCandidate{Type: "dm_message", ID: msg.ID, Field: "body",
			Before: snippet(msg.Body, raw), After: snippet(after, replacement), Editable: editable, Reason: reason})
	}
	return out
}

func ticketCandidates(scope searchScope, pattern, raw, replacement string) []replaceCandidate {
	var out []replaceCandidate
	for _, r := range searchTickets(scope, pattern) {
		ticket := r.Item.(models.Ticket)
		var editable bool
		var reason string
		if ticket.CustomerID != nil {
			editable, reason = ticketEditable(*ticket.CustomerID, scope.userID, scope.globalRole)
		} else {
			editable, reason = false, "ticket has no customer"
		}
		after, _ := replaceAllCI(ticket.Title, raw, replacement)
		out = append(out, replaceCandidate{Type: "ticket", ID: ticket.ID, Field: "title",
			Before: snippet(ticket.Title, raw), After: snippet(after, replacement), Editable: editable, Reason: reason})
	}
	return out
}

func timeEntryCandidates(scope searchScope, pattern, raw, replacement string) []replaceCandidate {
	var out []replaceCandidate
	for _, r := range searchTimeEntries(scope, pattern) {
		entry := r.Item.(models.TimeEntry)
		editable, reason := timeEntryEditable(entry, scope.userID)
		after, _ := replaceAllCI(entry.Description, raw, replacement)
		out = append(out, replaceCandidate{Type: "time_entry", ID: entry.ID, Field: "description",
			Before: snippet(entry.Description, raw), After: snippet(after, replacement), Editable: editable, Reason: reason})
	}
	return out
}

func collectCandidates(t string, scope searchScope, pattern, raw, replacement string) []replaceCandidate {
	switch t {
	case "card":
		return cardCandidates(scope, pattern, raw, replacement)
	case "card_comment":
		return cardCommentCandidates(scope, pattern, raw, replacement)
	case "dm_message":
		return dmMessageCandidates(scope, pattern, raw, replacement)
	case "ticket":
		return ticketCandidates(scope, pattern, raw, replacement)
	case "time_entry":
		return timeEntryCandidates(scope, pattern, raw, replacement)
	}
	return nil
}

type replaceRequest struct {
	Q       string   `json:"q"`
	Replace string   `json:"replace"`
	Types   []string `json:"types"`
}

// SearchReplacePreview godoc
// @Summary      Preview a search-and-replace across cards, card comments, DMs, tickets, and time entries
// @Tags         search
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Router       /search/replace/preview [post]
func SearchReplacePreview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var req replaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	raw, pattern, ok := sanitizeSearchQuery(req.Q)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too short (minimum 3 characters)"})
		return
	}
	if len(req.Types) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "select at least one type"})
		return
	}
	for _, t := range req.Types {
		if !replaceableTypes[t] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported type: " + t})
			return
		}
	}

	scope := buildSearchScope(userID, globalRole)

	var candidates []replaceCandidate
	counts := map[string]int{}
	for _, t := range req.Types {
		items := collectCandidates(t, scope, pattern, raw, req.Replace)
		candidates = append(candidates, items...)
		counts[t] = len(items)
	}
	if candidates == nil {
		candidates = []replaceCandidate{}
	}

	c.JSON(http.StatusOK, gin.H{"results": candidates, "counts": counts})
}

type replaceItem struct {
	Type  string `json:"type"`
	ID    uint   `json:"id"`
	Field string `json:"field"`
}

type replaceApplyRequest struct {
	Q       string        `json:"q"`
	Replace string        `json:"replace"`
	Items   []replaceItem `json:"items"`
}

type replaceOutcome struct {
	Type   string `json:"type"`
	ID     uint   `json:"id"`
	Reason string `json:"reason,omitempty"`
}

// SearchReplaceApply godoc
// @Summary      Apply a previously previewed search-and-replace to an explicit set of rows
// @Tags         search
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Router       /search/replace/apply [post]
func SearchReplaceApply(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var req replaceApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	raw, _, ok := sanitizeSearchQuery(req.Q)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too short (minimum 3 characters)"})
		return
	}

	var updated []replaceOutcome
	var skipped []replaceOutcome

	for _, item := range req.Items {
		if !replaceableTypes[item.Type] {
			skipped = append(skipped, replaceOutcome{Type: item.Type, ID: item.ID, Reason: "unsupported type"})
			continue
		}
		reason, ok := applyReplace(item.Type, item.ID, item.Field, raw, req.Replace, userID, globalRole)
		if ok {
			updated = append(updated, replaceOutcome{Type: item.Type, ID: item.ID})
		} else {
			skipped = append(skipped, replaceOutcome{Type: item.Type, ID: item.ID, Reason: reason})
		}
	}

	if updated == nil {
		updated = []replaceOutcome{}
	}
	if skipped == nil {
		skipped = []replaceOutcome{}
	}
	c.JSON(http.StatusOK, gin.H{"updated": updated, "skipped": skipped})
}

// applyReplace dispatches a single confirmed row to its type-specific apply
// function. field disambiguates which field to touch on multi-field types
// (currently only "card", which has both title and description) so that
// checking just the "title" preview row - and leaving "description"
// unchecked - can never also silently rewrite the description.
func applyReplace(t string, id uint, field, raw, replacement string, userID uint, role string) (reason string, ok bool) {
	switch t {
	case "card":
		return applyCardReplace(id, field, raw, replacement, userID, role)
	case "card_comment":
		return applyCardCommentReplace(id, raw, replacement, userID, role)
	case "dm_message":
		return applyDMMessageReplace(id, raw, replacement, userID)
	case "ticket":
		return applyTicketReplace(id, raw, replacement, userID, role)
	case "time_entry":
		return applyTimeEntryReplace(id, raw, replacement, userID)
	}
	return "unsupported type", false
}

func applyCardReplace(id uint, field, raw, replacement string, userID uint, role string) (string, bool) {
	var card models.Card
	if err := database.DB.First(&card, id).Error; err != nil {
		return "not found", false
	}
	if editable, reason := cardEditable(card, userID, role); !editable {
		return reason, false
	}

	updates := map[string]interface{}{}
	var historyEvents []models.CardHistory
	switch field {
	case "title":
		if containsCI(card.Title, raw) {
			if newTitle, n := replaceAllCI(card.Title, raw, replacement); n > 0 {
				updates["title"] = newTitle
				historyEvents = append(historyEvents, models.CardHistory{CardID: card.ID, UserID: userID, EventType: "title_changed", Detail: newTitle})
			}
		}
	case "description":
		if containsCI(card.Description, raw) {
			if newDesc, n := replaceAllCI(card.Description, raw, replacement); n > 0 {
				updates["description"] = newDesc
				historyEvents = append(historyEvents, models.CardHistory{CardID: card.ID, UserID: userID, EventType: "description_changed"})
			}
		}
	default:
		return "invalid field", false
	}
	if len(updates) == 0 {
		return "no longer matches", false
	}

	database.DB.Model(&card).Updates(updates)
	for _, h := range historyEvents {
		database.DB.Create(&h)
	}
	database.DB.Preload("Labels").Preload("Assignee").Preload("Assignees").Preload("Tags").Preload("Epic").First(&card, card.ID)
	ws.BroadcastToProject(card.ProjectID, ws.Message{Type: ws.TypeBoardCardUpdated, Payload: card})
	return "", true
}

func applyCardCommentReplace(id uint, raw, replacement string, userID uint, role string) (string, bool) {
	var comment models.CardComment
	if err := database.DB.Preload("Card").First(&comment, id).Error; err != nil {
		return "not found", false
	}
	if editable, reason := cardCommentEditable(comment, comment.Card.ProjectID, userID, role); !editable {
		return reason, false
	}
	newBody, n := replaceAllCI(comment.Body, raw, replacement)
	if n == 0 {
		return "no longer matches", false
	}

	database.DB.Model(&comment).Updates(map[string]interface{}{"body": newBody, "is_edited": true})
	database.DB.Preload("User").First(&comment, comment.ID)
	ws.BroadcastToProject(comment.Card.ProjectID, ws.Message{Type: ws.TypeBoardCommentUpdated, Payload: comment})
	return "", true
}

func applyDMMessageReplace(id uint, raw, replacement string, userID uint) (string, bool) {
	var msg models.ConversationMessage
	if err := database.DB.First(&msg, id).Error; err != nil {
		return "not found", false
	}
	if editable, reason := dmMessageEditable(msg, userID); !editable {
		return reason, false
	}
	newBody, n := replaceAllCI(msg.Body, raw, replacement)
	if n == 0 {
		return "no longer matches", false
	}

	database.DB.Model(&msg).Updates(map[string]interface{}{"body": newBody, "is_edited": true})

	var memberIDs []uint
	database.DB.Model(&models.ConversationMember{}).
		Where("conversation_id = ?", msg.ConversationID).
		Pluck("user_id", &memberIDs)
	for _, uid := range memberIDs {
		ws.BroadcastToUser(uid, ws.Message{
			Type: ws.TypeDMMessageUpdated,
			Payload: map[string]interface{}{
				"conversation_id": msg.ConversationID,
				"id":              msg.ID,
				"body":            newBody,
				"is_edited":       true,
			},
		})
	}
	return "", true
}

func applyTicketReplace(id uint, raw, replacement string, userID uint, role string) (string, bool) {
	var ticket models.Ticket
	if err := database.DB.First(&ticket, id).Error; err != nil {
		return "not found", false
	}
	if ticket.CustomerID == nil {
		return "ticket has no customer", false
	}
	if editable, reason := ticketEditable(*ticket.CustomerID, userID, role); !editable {
		return reason, false
	}
	newTitle, n := replaceAllCI(ticket.Title, raw, replacement)
	if n == 0 {
		return "no longer matches", false
	}

	database.DB.Model(&ticket).Update("title", newTitle)
	database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "title_changed", Detail: newTitle})
	return "", true
}

func applyTimeEntryReplace(id uint, raw, replacement string, userID uint) (string, bool) {
	var entry models.TimeEntry
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&entry).Error; err != nil {
		return "you can only edit your own time entries", false
	}
	newDesc, n := replaceAllCI(entry.Description, raw, replacement)
	if n == 0 {
		return "no longer matches", false
	}

	database.DB.Model(&entry).Update("description", newDesc)
	return "", true
}
