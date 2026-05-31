package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

// ListTicketCardLinks GET /api/v1/customers/:customerId/tickets/:ticketId/cards
// Returns linked cards enriched with project info for display.
func ListTicketCardLinks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetGlobalRole(c)
	customerID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	if err := requireCustomerAccess(uint(customerID), userID, role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var links []struct {
		LinkID         uint   `json:"link_id"`
		CardID         uint   `json:"card_id"`
		CardNumber     int    `json:"card_number"`
		Title          string `json:"title"`
		ProjectSlug    string `json:"project_slug"`
		ProjectKey     string `json:"project_key"`
		ColumnID       uint   `json:"column_id"`
	}
	if err := database.DB.
		Table("ticket_card_links").
		Select("ticket_card_links.id as link_id, cards.id as card_id, cards.card_number, cards.title, projects.slug as project_slug, projects.key_prefix as project_key, cards.column_id").
		Joins("JOIN cards ON cards.id = ticket_card_links.card_id").
		Joins("JOIN projects ON projects.id = cards.project_id").
		Where("ticket_card_links.ticket_id = ? AND cards.deleted_at IS NULL", ticketID).
		Scan(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list card links"})
		return
	}
	if links == nil {
		links = []struct {
			LinkID         uint   `json:"link_id"`
			CardID         uint   `json:"card_id"`
			CardNumber     int    `json:"card_number"`
			Title          string `json:"title"`
			ProjectSlug    string `json:"project_slug"`
			ProjectKey     string `json:"project_key"`
			ColumnID       uint   `json:"column_id"`
		}{}
	}
	c.JSON(http.StatusOK, links)
}

// CreateTicketCardLink POST /api/v1/customers/:customerId/tickets/:ticketId/cards
// Links a ticket to a card. Body: { card_id } or { ref: "PRJ-123" }.
func CreateTicketCardLink(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetGlobalRole(c)
	customerID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	if err := requireCustomerAccess(uint(customerID), userID, role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if err := requireNotCustomerRole(role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		CardID *uint  `json:"card_id"`
		Ref    string `json:"ref"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "card_id or ref required"})
		return
	}

	var cardID uint
	if req.CardID != nil {
		cardID = *req.CardID
	} else if req.Ref != "" {
		sep := strings.LastIndex(req.Ref, "-")
		if sep <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ref format, use PROJECT-123"})
			return
		}
		prefix := strings.ToUpper(req.Ref[:sep])
		number, err := strconv.Atoi(req.Ref[sep+1:])
		if err != nil || number < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ref"})
			return
		}
		var result struct{ ID uint }
		if err := database.DB.
			Table("cards").
			Select("cards.id").
			Joins("JOIN projects ON projects.id = cards.project_id").
			Where("projects.key_prefix = ? AND cards.card_number = ? AND cards.deleted_at IS NULL", prefix, number).
			Scan(&result).Error; err != nil || result.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		cardID = result.ID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "card_id or ref required"})
		return
	}

	// Verify the user has project access to the card
	var projectID uint
	if err := database.DB.Model(&models.Card{}).Select("project_id").First(&projectID, cardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}
	if err := services.RequireProjectRole(projectID, userID, role, "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	link := models.TicketCardLink{
		TicketID:    uint(ticketID),
		CardID:      cardID,
		CreatedByID: userID,
	}
	if err := database.DB.Create(&link).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "Duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "ticket is already linked to this card"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link"})
		return
	}
	c.JSON(http.StatusCreated, link)
	database.DB.Create(&models.TicketHistory{TicketID: uint(ticketID), UserID: userID, EventType: "card_linked"})
}

// DeleteTicketCardLink DELETE /api/v1/customers/:customerId/tickets/:ticketId/cards/:linkId
func DeleteTicketCardLink(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetGlobalRole(c)
	customerID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	linkID, err := strconv.ParseUint(c.Param("linkId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link id"})
		return
	}

	if err := requireCustomerAccess(uint(customerID), userID, role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if err := requireNotCustomerRole(role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var link models.TicketCardLink
	if err := database.DB.Where("id = ? AND ticket_id = ?", linkID, ticketID).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	database.DB.Delete(&link)
	database.DB.Create(&models.TicketHistory{TicketID: uint(ticketID), UserID: userID, EventType: "card_unlinked"})
	c.JSON(http.StatusOK, gin.H{"message": "link removed"})
}

// ListCardTicketLinks GET /api/v1/projects/:slug/cards/:cardId/tickets
// Returns linked tickets from the card side.
func ListCardTicketLinks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetGlobalRole(c)
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
	if err := services.RequireProjectRole(project.ID, userID, role, "viewer"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var links []struct {
		LinkID     uint   `json:"link_id"`
		TicketID   uint   `json:"ticket_id"`
		Title      string `json:"title"`
		Status     string `json:"status"`
		Priority   string `json:"priority"`
		CustomerID uint   `json:"customer_id"`
	}
	if err := database.DB.
		Table("ticket_card_links").
		Select("ticket_card_links.id as link_id, tickets.id as ticket_id, tickets.title, tickets.status, tickets.priority, tickets.customer_id").
		Joins("JOIN tickets ON tickets.id = ticket_card_links.ticket_id").
		Where("ticket_card_links.card_id = ?", cardID).
		Scan(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list ticket links"})
		return
	}
	if links == nil {
		links = []struct {
			LinkID     uint   `json:"link_id"`
			TicketID   uint   `json:"ticket_id"`
			Title      string `json:"title"`
			Status     string `json:"status"`
			Priority   string `json:"priority"`
			CustomerID uint   `json:"customer_id"`
		}{}
	}
	c.JSON(http.StatusOK, links)
}
