package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

type ticketLinkEntry struct {
	LinkID       uint   `json:"link_id"`
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Priority     string `json:"priority"`
	CustomerID   uint   `json:"customer_id"`
}

// ListTicketLinks GET /api/v1/customers/:customerId/tickets/:ticketId/links
func ListTicketLinks(c *gin.Context) {
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

	var links []models.TicketLink
	database.DB.Where("source_ticket_id = ? OR target_ticket_id = ?", ticketID, ticketID).Find(&links)

	out := make([]ticketLinkEntry, 0, len(links))
	for _, l := range links {
		var linkedTicketID uint
		linkID := l.ID
		if l.SourceTicketID == uint(ticketID) {
			linkedTicketID = l.TargetTicketID
		} else {
			linkedTicketID = l.SourceTicketID
		}

		var t models.Ticket
		if err := database.DB.Select("id", "title", "status", "priority", "customer_id").First(&t, linkedTicketID).Error; err != nil {
			continue
		}
		out = append(out, ticketLinkEntry{
			LinkID:     linkID,
			ID:         t.ID,
			Title:      t.Title,
			Status:     t.Status,
			Priority:   t.Priority,
			CustomerID: t.CustomerID,
		})
	}

	c.JSON(http.StatusOK, out)
}

// CreateTicketLink POST /api/v1/customers/:customerId/tickets/:ticketId/links
func CreateTicketLink(c *gin.Context) {
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

	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id = ?", ticketID, customerID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	var req struct {
		TargetTicketID uint `json:"target_ticket_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_ticket_id is required"})
		return
	}

	if req.TargetTicketID == uint(ticketID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot link ticket to itself"})
		return
	}

	var target models.Ticket
	if err := database.DB.First(&target, req.TargetTicketID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target ticket not found"})
		return
	}

	// Prevent duplicate links (check both directions)
	var count int64
	database.DB.Model(&models.TicketLink{}).
		Where("(source_ticket_id = ? AND target_ticket_id = ?) OR (source_ticket_id = ? AND target_ticket_id = ?)",
			ticketID, req.TargetTicketID, req.TargetTicketID, ticketID).
		Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "tickets are already linked"})
		return
	}

	link := models.TicketLink{
		SourceTicketID: uint(ticketID),
		TargetTicketID: req.TargetTicketID,
	}
	if err := database.DB.Create(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link"})
		return
	}

	c.JSON(http.StatusCreated, ticketLinkEntry{
		LinkID:     link.ID,
		ID:         target.ID,
		Title:      target.Title,
		Status:     target.Status,
		Priority:   target.Priority,
		CustomerID: target.CustomerID,
	})
	database.DB.Create(&models.TicketHistory{TicketID: uint(ticketID), UserID: userID, EventType: "ticket_linked", Detail: target.Title})
}

// DeleteTicketLink DELETE /api/v1/customers/:customerId/tickets/:ticketId/links/:linkId
func DeleteTicketLink(c *gin.Context) {
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

	database.DB.Where("id = ? AND (source_ticket_id = ? OR target_ticket_id = ?)", linkID, ticketID, ticketID).Delete(&models.TicketLink{})
	database.DB.Create(&models.TicketHistory{TicketID: uint(ticketID), UserID: userID, EventType: "ticket_unlinked"})
	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}
