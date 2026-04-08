package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

// cardRefEntry is the enriched payload returned for each linked card.
type cardRefEntry struct {
	RefID       uint   `json:"ref_id"`      // CardReference.ID — used for deletion
	ID          uint   `json:"id"`          // linked card's ID
	CardNumber  int    `json:"card_number"`
	KeyPrefix   string `json:"key_prefix"`
	Title       string `json:"title"`
	Priority    string `json:"priority"`
	Closed      bool   `json:"closed"`
	ColumnName  string `json:"column_name"`
	ProjectSlug string `json:"project_slug"`
	ProjectName string `json:"project_name"`
}

// ListCardRefs returns all cards linked to the given card (both directions).
func ListCardRefs(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "viewer"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Verify the card exists in this project
	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	// Load all references for this card (either direction)
	var refs []models.CardReference
	database.DB.Where("source_card_id = ? OR target_card_id = ?", cardID, cardID).Find(&refs)

	// Collect IDs of the "other" card in each reference
	otherIDs := make([]uint, 0, len(refs))
	for _, r := range refs {
		if r.SourceCardID == uint(cardID) {
			otherIDs = append(otherIDs, r.TargetCardID)
		} else {
			otherIDs = append(otherIDs, r.SourceCardID)
		}
	}
	if len(otherIDs) == 0 {
		c.JSON(http.StatusOK, []cardRefEntry{})
		return
	}

	// Load the linked cards with their column and project in one query
	type row struct {
		ID          uint
		CardNumber  int
		KeyPrefix   string
		Title       string
		Priority    string
		Closed      bool
		ColumnName  string
		ProjectSlug string
		ProjectName string
	}
	var rows []row
	database.DB.Table("cards").
		Select("cards.id, cards.card_number, projects.key_prefix, cards.title, cards.priority, cards.closed, columns.name as column_name, projects.slug as project_slug, projects.name as project_name").
		Joins("JOIN columns ON columns.id = cards.column_id").
		Joins("JOIN projects ON projects.id = cards.project_id").
		Where("cards.id IN ? AND cards.deleted_at IS NULL", otherIDs).
		Scan(&rows)

	// Map card ID → row for quick lookup
	rowMap := make(map[uint]row, len(rows))
	for _, r := range rows {
		rowMap[r.ID] = r
	}

	// Build response preserving ref IDs
	result := make([]cardRefEntry, 0, len(refs))
	for _, ref := range refs {
		otherID := ref.TargetCardID
		if ref.SourceCardID == uint(cardID) {
			otherID = ref.TargetCardID
		} else {
			otherID = ref.SourceCardID
		}
		r, ok := rowMap[otherID]
		if !ok {
			continue // linked card deleted
		}
		result = append(result, cardRefEntry{
			RefID:       ref.ID,
			ID:          r.ID,
			CardNumber:  r.CardNumber,
			KeyPrefix:   r.KeyPrefix,
			Title:       r.Title,
			Priority:    r.Priority,
			Closed:      r.Closed,
			ColumnName:  r.ColumnName,
			ProjectSlug: r.ProjectSlug,
			ProjectName: r.ProjectName,
		})
	}

	c.JSON(http.StatusOK, result)
}

// CreateCardRef links another card to this card by card reference string (e.g. "WEB-4").
// Defaults to the current project when no prefix is given (e.g. just "4").
func CreateCardRef(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	var req struct {
		Ref string `json:"ref" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ref is required"})
		return
	}

	// Parse ref: "WEB-4" → prefix=WEB, number=4; "4" → prefix=current project, number=4
	ref := strings.TrimSpace(strings.ToUpper(req.Ref))
	var targetPrefix string
	var targetNumber int
	if idx := strings.LastIndex(ref, "-"); idx != -1 {
		targetPrefix = ref[:idx]
		targetNumber, err = strconv.Atoi(ref[idx+1:])
	} else {
		targetPrefix = strings.ToUpper(project.KeyPrefix)
		targetNumber, err = strconv.Atoi(ref)
	}
	if err != nil || targetNumber <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("cannot parse card ref %q", req.Ref)})
		return
	}

	// Find the target card
	var targetCard models.Card
	if err := database.DB.
		Joins("JOIN projects ON projects.id = cards.project_id").
		Where("projects.key_prefix = ? AND cards.card_number = ? AND cards.deleted_at IS NULL", targetPrefix, targetNumber).
		First(&targetCard).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("card %s not found", req.Ref)})
		return
	}

	if targetCard.ID == uint(cardID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot link a card to itself"})
		return
	}

	// Prevent duplicates (either direction)
	var existing models.CardReference
	if err := database.DB.Where(
		"(source_card_id = ? AND target_card_id = ?) OR (source_card_id = ? AND target_card_id = ?)",
		cardID, targetCard.ID, targetCard.ID, cardID,
	).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "cards are already linked"})
		return
	}

	newRef := models.CardReference{
		SourceCardID: uint(cardID),
		TargetCardID: targetCard.ID,
	}
	if err := database.DB.Create(&newRef).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link"})
		return
	}

	// Return the enriched entry right away so the UI can append it
	var r struct {
		ID          uint
		CardNumber  int
		KeyPrefix   string
		Title       string
		Priority    string
		Closed      bool
		ColumnName  string
		ProjectSlug string
		ProjectName string
	}
	database.DB.Table("cards").
		Select("cards.id, cards.card_number, projects.key_prefix, cards.title, cards.priority, cards.closed, columns.name as column_name, projects.slug as project_slug, projects.name as project_name").
		Joins("JOIN columns ON columns.id = cards.column_id").
		Joins("JOIN projects ON projects.id = cards.project_id").
		Where("cards.id = ?", targetCard.ID).
		Scan(&r)

	c.JSON(http.StatusCreated, cardRefEntry{
		RefID:       newRef.ID,
		ID:          r.ID,
		CardNumber:  r.CardNumber,
		KeyPrefix:   r.KeyPrefix,
		Title:       r.Title,
		Priority:    r.Priority,
		Closed:      r.Closed,
		ColumnName:  r.ColumnName,
		ProjectSlug: r.ProjectSlug,
		ProjectName: r.ProjectName,
	})
}

// DeleteCardRef removes a card reference by its ID.
func DeleteCardRef(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}
	refID, err := strconv.ParseUint(c.Param("refId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ref id"})
		return
	}

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Verify the reference involves this card
	var ref models.CardReference
	if err := database.DB.Where(
		"id = ? AND (source_card_id = ? OR target_card_id = ?)",
		refID, cardID, cardID,
	).First(&ref).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	database.DB.Delete(&ref)
	c.Status(http.StatusNoContent)
}
