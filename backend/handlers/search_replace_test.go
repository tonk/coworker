package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/testutil"
)

func TestReplaceAllCICaseInsensitiveAndPreservesSurroundingCasing(t *testing.T) {
	out, n := replaceAllCI("The Quick fox jumps over the QUICK dog", "quick", "slow")
	assert.Equal(t, "The slow fox jumps over the slow dog", out)
	assert.Equal(t, 2, n)
}

func TestReplaceAllCINoMatch(t *testing.T) {
	out, n := replaceAllCI("nothing here", "zzz", "yyy")
	assert.Equal(t, "nothing here", out)
	assert.Equal(t, 0, n)
}

func TestReplaceAllCIEmptyOld(t *testing.T) {
	out, n := replaceAllCI("unchanged", "", "x")
	assert.Equal(t, "unchanged", out)
	assert.Equal(t, 0, n)
}

// Cards require project role >= member to edit, mirroring UpdateCard
// (handlers/card.go:245) - a viewer can see the card in search but must not
// be able to replace its title/description.
func TestCardEditableRequiresMemberRole(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	owner := &models.User{Username: "owner", Email: "owner@example.com", PasswordHash: "x"}
	viewer := &models.User{Username: "viewer", Email: "viewer@example.com", PasswordHash: "x"}
	require.NoError(t, db.Create(owner).Error)
	require.NoError(t, db.Create(viewer).Error)
	project := &models.Project{Name: "P", Slug: "p", KeyPrefix: "P", CreatedByID: owner.ID}
	require.NoError(t, db.Create(project).Error)
	require.NoError(t, db.Create(&models.ProjectMember{ProjectID: project.ID, UserID: viewer.ID, Role: "viewer"}).Error)
	card := models.Card{ProjectID: project.ID, Title: "hello world"}

	editable, reason := cardEditable(card, viewer.ID, "user")
	assert.False(t, editable)
	assert.NotEmpty(t, reason)

	editable, _ = cardEditable(card, owner.ID, "admin")
	assert.True(t, editable, "global admins bypass project role checks")
}

// Card comments are author-only, mirroring UpdateComment (handlers/card_comment.go:181-184).
func TestCardCommentEditableRequiresAuthor(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	author := &models.User{Username: "author", Email: "author@example.com", PasswordHash: "x"}
	other := &models.User{Username: "other", Email: "other@example.com", PasswordHash: "x"}
	require.NoError(t, db.Create(author).Error)
	require.NoError(t, db.Create(other).Error)
	project := &models.Project{Name: "P", Slug: "p", KeyPrefix: "P", CreatedByID: author.ID}
	require.NoError(t, db.Create(project).Error)
	require.NoError(t, db.Create(&models.ProjectMember{ProjectID: project.ID, UserID: other.ID, Role: "viewer"}).Error)
	comment := models.CardComment{UserID: author.ID, Body: "hi"}

	editable, reason := cardCommentEditable(comment, project.ID, other.ID, "user")
	assert.False(t, editable)
	assert.Equal(t, "you can only edit your own comments", reason)

	editable, _ = cardCommentEditable(comment, project.ID, author.ID, "user")
	assert.False(t, editable, "author has no ProjectMember row here, so project access itself is missing")

	require.NoError(t, db.Create(&models.ProjectMember{ProjectID: project.ID, UserID: author.ID, Role: "viewer"}).Error)
	editable, _ = cardCommentEditable(comment, project.ID, author.ID, "user")
	assert.True(t, editable)
}

// DM messages are sender-only, mirroring EditConversationMessage (handlers/conversation.go:367).
func TestDMMessageEditableRequiresSender(t *testing.T) {
	msg := models.ConversationMessage{SenderID: 1, Body: "hi"}
	editable, reason := dmMessageEditable(msg, 2)
	assert.False(t, editable)
	assert.Equal(t, "you can only edit your own messages", reason)

	editable, _ = dmMessageEditable(msg, 1)
	assert.True(t, editable)
}

// Tickets require requireCustomerAccess + requireNotCustomerRole, mirroring
// UpdateTicket (handlers/ticket.go:269-276) - the "customer" global role must
// be excluded entirely, even with a direct CustomerAccess row.
func TestTicketEditableExcludesCustomerRole(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	agent := &models.User{Username: "agent", Email: "agent@example.com", PasswordHash: "x"}
	cust := &models.User{Username: "cust", Email: "cust@example.com", PasswordHash: "x"}
	require.NoError(t, db.Create(agent).Error)
	require.NoError(t, db.Create(cust).Error)
	customer := &models.Customer{Name: "Acme"}
	require.NoError(t, db.Create(customer).Error)
	require.NoError(t, db.Create(&models.CustomerAccess{UserID: agent.ID, CustomerID: customer.ID, Role: "member"}).Error)
	require.NoError(t, db.Create(&models.CustomerAccess{UserID: cust.ID, CustomerID: customer.ID, Role: "member"}).Error)

	editable, _ := ticketEditable(customer.ID, agent.ID, "user")
	assert.True(t, editable)

	editable, reason := ticketEditable(customer.ID, cust.ID, "customer")
	assert.False(t, editable)
	assert.Equal(t, "customer-role users can't edit tickets", reason)

	// An unrelated user with no CustomerAccess row at all.
	editable, reason = ticketEditable(customer.ID, 9999, "user")
	assert.False(t, editable)
	assert.Equal(t, "you don't have access to this customer", reason)
}

// Time entries are owner-only with no admin/time_tracking_viewer bypass,
// mirroring UpdateTimeEntry's Where("id = ? AND user_id = ?") (handlers/time_entry.go:189).
func TestTimeEntryEditableRequiresOwner(t *testing.T) {
	entry := models.TimeEntry{UserID: 5, Description: "worked"}
	editable, reason := timeEntryEditable(entry, 6)
	assert.False(t, editable)
	assert.Equal(t, "you can only edit your own time entries", reason)

	editable, _ = timeEntryEditable(entry, 5)
	assert.True(t, editable)
}

// applyTimeEntryReplace must never touch another user's entry, even though
// GlobalSearch can surface it to admins/time_tracking_viewer users.
func TestApplyTimeEntryReplaceRejectsNonOwner(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	owner := &models.User{Username: "owner", Email: "owner@example.com", PasswordHash: "x"}
	admin := &models.User{Username: "admin", Email: "admin@example.com", PasswordHash: "x"}
	require.NoError(t, db.Create(owner).Error)
	require.NoError(t, db.Create(admin).Error)
	entry := &models.TimeEntry{UserID: owner.ID, Date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Minutes: 60, Description: "fixed the bug"}
	require.NoError(t, db.Create(entry).Error)

	reason, ok := applyTimeEntryReplace(entry.ID, "bug", "issue", admin.ID)
	assert.False(t, ok)
	assert.Equal(t, "you can only edit your own time entries", reason)

	_, ok = applyTimeEntryReplace(entry.ID, "bug", "issue", owner.ID)
	assert.True(t, ok)

	var reloaded models.TimeEntry
	require.NoError(t, db.First(&reloaded, entry.ID).Error)
	assert.Equal(t, "fixed the issue", reloaded.Description)
}

// applyCardReplace must only touch the single field it's asked to (a user
// who only checked the "title" preview row must not also have their
// description silently rewritten just because it happened to match too),
// and must record CardHistory the same way UpdateCard does.
func TestApplyCardReplaceOnlyTouchesRequestedField(t *testing.T) {
	db, cleanup := testutil.SetupTestDB()
	defer cleanup()
	database.DB = db

	owner := &models.User{Username: "owner", Email: "owner@example.com", PasswordHash: "x"}
	require.NoError(t, db.Create(owner).Error)
	project := &models.Project{Name: "P", Slug: "p", KeyPrefix: "P", CreatedByID: owner.ID}
	require.NoError(t, db.Create(project).Error)
	require.NoError(t, db.Create(&models.ProjectMember{ProjectID: project.ID, UserID: owner.ID, Role: "member"}).Error)
	card := &models.Card{ProjectID: project.ID, Title: "Fix Login bug", Description: "the login bug is annoying"}
	require.NoError(t, db.Create(card).Error)

	reason, ok := applyCardReplace(card.ID, "title", "login", "signin", owner.ID, "user")
	assert.True(t, ok, reason)

	var reloaded models.Card
	require.NoError(t, db.First(&reloaded, card.ID).Error)
	assert.Equal(t, "Fix signin bug", reloaded.Title)
	assert.Equal(t, "the login bug is annoying", reloaded.Description, "description must be untouched when only the title field was confirmed")

	var history []models.CardHistory
	require.NoError(t, db.Where("card_id = ?", card.ID).Find(&history).Error)
	require.Len(t, history, 1)
	assert.Equal(t, "title_changed", history[0].EventType)

	reason, ok = applyCardReplace(card.ID, "description", "login", "signin", owner.ID, "user")
	assert.True(t, ok, reason)
	require.NoError(t, db.First(&reloaded, card.ID).Error)
	assert.Equal(t, "the signin bug is annoying", reloaded.Description)
}
