package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

type SearchResult struct {
	Type string      `json:"type"`
	Item interface{} `json:"item"`
}

// searchScope carries the per-request access boundaries shared by GlobalSearch
// and the search-replace preview/apply endpoints, so both compute "what can
// this user see/touch" exactly the same way.
type searchScope struct {
	userID     uint
	globalRole string

	// memberProjectIDs: projects the user may search cards/chat/comments in
	// (all projects for admins, ProjectMember rows otherwise).
	memberProjectIDs []uint
	// convIDs: conversations the user is a member of (DM search scope).
	convIDs []uint

	// helpdeskAccess mirrors middleware.RequireFeature("helpdesk_enabled"):
	// true for admins, the "customer" role, or users with the flag enabled.
	helpdeskAccess bool
	// timeTrackingAccess mirrors middleware.RequireFeature("time_tracking_enabled"),
	// but is always false for the "customer" role (BlockCustomerRole).
	timeTrackingAccess bool
	// timeTrackingAllUsers is true for admins and time_tracking_viewer users,
	// who may *see* every user's time entries in search (though editing one
	// still requires being its owner - see UpdateTimeEntry).
	timeTrackingAllUsers bool
}

func buildSearchScope(userID uint, globalRole string) searchScope {
	scope := searchScope{userID: userID, globalRole: globalRole}

	if globalRole == "admin" {
		database.DB.Model(&models.Project{}).Pluck("id", &scope.memberProjectIDs)
	} else {
		database.DB.Model(&models.ProjectMember{}).
			Where("user_id = ?", userID).
			Pluck("project_id", &scope.memberProjectIDs)
	}

	database.DB.Model(&models.ConversationMember{}).
		Where("user_id = ?", userID).
		Pluck("conversation_id", &scope.convIDs)

	var featureUser models.User
	if globalRole != "admin" {
		database.DB.Select("helpdesk_enabled", "time_tracking_enabled", "time_tracking_viewer").First(&featureUser, userID)
	}
	scope.helpdeskAccess = globalRole == "admin" || globalRole == "customer" || featureUser.HelpdeskEnabled
	scope.timeTrackingAllUsers = globalRole == "admin" || featureUser.TimeTrackingViewer
	scope.timeTrackingAccess = globalRole != "customer" && (scope.timeTrackingAllUsers || featureUser.TimeTrackingEnabled)

	return scope
}

// sanitizeSearchQuery validates and cleans a raw search query. It returns the
// cleaned query (no SQL LIKE wildcards), a ready-to-use "%q%" LIKE pattern,
// and false if the query is too short to search with (before or after
// stripping wildcard characters).
func sanitizeSearchQuery(q string) (raw string, pattern string, ok bool) {
	if len(q) < 3 {
		return "", "", false
	}
	q = strings.NewReplacer("%", "", "_", "").Replace(q)
	if len(q) < 3 {
		return "", "", false
	}
	return q, "%" + q + "%", true
}

// defaultSearchLimit is the per-type result cap used by the quick-search
// dropdown (GlobalSearch). Search-replace uses the same default but lets the
// user raise it - see maxSearchLimit in search_replace.go.
const defaultSearchLimit = 20

func searchCards(scope searchScope, pattern string, limit int) []SearchResult {
	if len(scope.memberProjectIDs) == 0 {
		return nil
	}
	var cards []models.Card
	database.DB.Preload("Assignee").
		Where("project_id IN ? AND (title LIKE ? OR description LIKE ?)", scope.memberProjectIDs, pattern, pattern).
		Limit(limit).Find(&cards)
	results := make([]SearchResult, 0, len(cards))
	for _, c := range cards {
		results = append(results, SearchResult{Type: "card", Item: c})
	}
	return results
}

func searchChatMessages(scope searchScope, pattern string, limit int) []SearchResult {
	if len(scope.memberProjectIDs) == 0 {
		return nil
	}
	var msgs []models.ChatMessage
	database.DB.Preload("User").
		Where("project_id IN ? AND body LIKE ? AND is_deleted = false AND is_bot = false", scope.memberProjectIDs, pattern).
		Limit(limit).Find(&msgs)
	results := make([]SearchResult, 0, len(msgs))
	for _, m := range msgs {
		results = append(results, SearchResult{Type: "chat_message", Item: m})
	}
	return results
}

func searchDMMessages(scope searchScope, pattern string, limit int) []SearchResult {
	if len(scope.convIDs) == 0 {
		return nil
	}
	var dms []models.ConversationMessage
	database.DB.Preload("Sender").
		Where("conversation_id IN ? AND body LIKE ? AND is_deleted = false", scope.convIDs, pattern).
		Limit(limit).Find(&dms)
	results := make([]SearchResult, 0, len(dms))
	for _, m := range dms {
		results = append(results, SearchResult{Type: "dm_message", Item: m})
	}
	return results
}

func searchCardComments(scope searchScope, pattern string, limit int) []SearchResult {
	if len(scope.memberProjectIDs) == 0 {
		return nil
	}
	var cardIDs []uint
	database.DB.Model(&models.Card{}).
		Where("project_id IN ?", scope.memberProjectIDs).
		Pluck("id", &cardIDs)
	if len(cardIDs) == 0 {
		return nil
	}
	var comments []models.CardComment
	database.DB.Preload("User").
		Where("card_id IN ? AND body LIKE ?", cardIDs, pattern).
		Limit(limit).Find(&comments)
	results := make([]SearchResult, 0, len(comments))
	for _, cm := range comments {
		results = append(results, SearchResult{Type: "card_comment", Item: cm})
	}
	return results
}

func searchTickets(scope searchScope, pattern string, limit int) []SearchResult {
	if !scope.helpdeskAccess {
		return nil
	}
	var customerIDs []uint
	if scope.globalRole == "admin" {
		database.DB.Model(&models.Customer{}).Pluck("id", &customerIDs)
	} else {
		for cid := range getAccessibleCustomerRoles(scope.userID) {
			customerIDs = append(customerIDs, cid)
		}
	}
	if len(customerIDs) == 0 {
		return nil
	}
	var tickets []models.Ticket
	database.DB.Preload("Customer").
		Where("customer_id IN ? AND title LIKE ? AND is_spam = false", customerIDs, pattern).
		Limit(limit).Find(&tickets)
	results := make([]SearchResult, 0, len(tickets))
	for _, t := range tickets {
		results = append(results, SearchResult{Type: "ticket", Item: t})
	}
	return results
}

func searchTimeEntries(scope searchScope, pattern string, limit int) []SearchResult {
	if !scope.timeTrackingAccess {
		return nil
	}
	query := database.DB.Preload("Customer").Preload("Project").
		Where("description LIKE ?", pattern)
	if !scope.timeTrackingAllUsers {
		query = query.Where("user_id = ?", scope.userID)
	} else {
		query = query.Preload("User")
	}
	var entries []models.TimeEntry
	query.Order("date desc").Limit(limit).Find(&entries)
	results := make([]SearchResult, 0, len(entries))
	for _, e := range entries {
		results = append(results, SearchResult{Type: "time_entry", Item: e})
	}
	return results
}

// GlobalSearch godoc
// @Summary      Search cards, messages, tickets, and time entries
// @Tags         search
// @Produce      json
// @Security     BearerAuth
// @Param        q query string true "Search query (min 2 characters)"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Router       /search [get]
func GlobalSearch(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	_, pattern, ok := sanitizeSearchQuery(c.Query("q"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too short (minimum 3 characters)"})
		return
	}

	scope := buildSearchScope(userID, globalRole)

	var results []SearchResult
	results = append(results, searchCards(scope, pattern, defaultSearchLimit)...)
	results = append(results, searchChatMessages(scope, pattern, defaultSearchLimit)...)
	results = append(results, searchDMMessages(scope, pattern, defaultSearchLimit)...)
	results = append(results, searchCardComments(scope, pattern, defaultSearchLimit)...)
	results = append(results, searchTickets(scope, pattern, defaultSearchLimit)...)
	results = append(results, searchTimeEntries(scope, pattern, defaultSearchLimit)...)

	if results == nil {
		results = []SearchResult{}
	}
	c.JSON(http.StatusOK, results)
}
