package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
	"gorm.io/gorm"
)

// ListTickets GET /api/v1/customers/:customerId/tickets
func ListTickets(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetGlobalRole(c)
	customerID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	if err := requireCustomerAccess(uint(customerID), userID, role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var tickets []models.Ticket
	database.DB.Where("customer_id = ?", customerID).
		Preload("CreatedBy").
		Preload("AssignedTo").
		Preload("Owner").
		Preload("Group").
		Preload("Tags").
		Preload("SlaPolicy").
		Order("CASE WHEN status = 'pending' AND reminder_at IS NOT NULL AND reminder_at <= datetime('now') THEN 0 ELSE 1 END, created_at desc").
		Find(&tickets)
	if tickets == nil {
		tickets = []models.Ticket{}
	}
	for i := range tickets {
		refreshSlaBreachStatus(&tickets[i])
	}
	c.JSON(http.StatusOK, tickets)
}

// GetTicket GET /api/v1/customers/:customerId/tickets/:ticketId
func GetTicket(c *gin.Context) {
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
	if err := database.DB.Where("id = ? AND customer_id = ?", ticketID, customerID).
		Preload("CreatedBy").
		Preload("AssignedTo").
		Preload("Owner").
		Preload("Group").
		Preload("Tags").
		Preload("SlaPolicy").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc")
		}).
		Preload("Messages.User").
		First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	refreshSlaBreachStatus(&ticket)

	// Load description-level attachments (owner_type="ticket")
	if am := LoadAttachments("ticket", []uint{ticket.ID}); len(am[ticket.ID]) > 0 {
		ticket.Attachments = am[ticket.ID]
	}

	// Load per-message attachments
	msgIDs := make([]uint, len(ticket.Messages))
	for i, m := range ticket.Messages {
		msgIDs[i] = m.ID
	}
	attachMap := LoadAttachments("ticket_message", msgIDs)
	for i := range ticket.Messages {
		ticket.Messages[i].Attachments = attachMap[ticket.Messages[i].ID]
		if ticket.Messages[i].Attachments == nil {
			ticket.Messages[i].Attachments = []models.Attachment{}
		}
	}

	c.JSON(http.StatusOK, ticket)
}

// CreateTicket POST /api/v1/customers/:customerId/tickets
func CreateTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetGlobalRole(c)
	customerID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	if err := requireCustomerAccess(uint(customerID), userID, role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Priority    string `json:"priority"`
		AssignedTo  *uint  `json:"assigned_to_id"`
		OwnerID     *uint  `json:"owner_id"`
		GroupID     *uint  `json:"group_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	validTypes := map[string]bool{"incident": true, "problem": true, "service_request": true, "change_request": true}
	if req.Type == "" || !validTypes[req.Type] {
		req.Type = "incident"
	}

	validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
	if req.Priority != "" && !validPriorities[req.Priority] {
		req.Priority = "medium"
	} else if req.Priority == "" {
		req.Priority = "medium"
	}

	ticket := models.Ticket{
		CustomerID:   uint(customerID),
		Title:        req.Title,
		Description:  req.Description,
		Type:         req.Type,
		Priority:     req.Priority,
		Status:       "open",
		CreatedByID:  userID,
		AssignedToID: req.AssignedTo,
		OwnerID:      req.OwnerID,
		GroupID:      req.GroupID,
	}

	if policy := MatchSlaPolicy(req.Priority); policy != nil {
		now := time.Now()
		respDeadline, resDeadline := ComputeSlaDeadlines(policy, now)
		ticket.SlaPolicyID = &policy.ID
		ticket.SlaResponseDeadline = respDeadline
		ticket.SlaResolutionDeadline = resDeadline
	}

	if err := database.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create ticket"})
		return
	}

	database.DB.Preload("CreatedBy").Preload("AssignedTo").Preload("Owner").Preload("Group").Preload("Tags").Preload("SlaPolicy").First(&ticket, ticket.ID)
	c.JSON(http.StatusCreated, ticket)
}

// UpdateTicket PUT /api/v1/customers/:customerId/tickets/:ticketId
func UpdateTicket(c *gin.Context) {
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
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		Type        *string    `json:"type"`
		Status      *string    `json:"status"`
		Priority    *string    `json:"priority"`
		AssignedTo  *uint      `json:"assigned_to_id"`
		OwnerID     *uint      `json:"owner_id"`
		GroupID     *uint      `json:"group_id"`
		ReminderAt  *time.Time `json:"reminder_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updates := map[string]interface{}{}
	priorityChanged := false
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Type != nil {
		validTypes := map[string]bool{"incident": true, "problem": true, "service_request": true, "change_request": true}
		if validTypes[*req.Type] {
			updates["type"] = *req.Type
		}
	}
	if req.Status != nil {
		validStatuses := map[string]bool{"open": true, "in_progress": true, "resolved": true, "closed": true, "pending": true}
		if validStatuses[*req.Status] {
			updates["status"] = *req.Status
		}
	}
	if req.ReminderAt != nil {
		updates["reminder_at"] = req.ReminderAt
	}
	if req.Priority != nil {
		validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
		if validPriorities[*req.Priority] {
			updates["priority"] = *req.Priority
			priorityChanged = true
		}
	}
	if req.AssignedTo != nil {
		if *req.AssignedTo == 0 {
			updates["assigned_to_id"] = nil
		} else {
			updates["assigned_to_id"] = *req.AssignedTo
		}
	}
	if req.OwnerID != nil {
		if *req.OwnerID == 0 {
			updates["owner_id"] = nil
		} else {
			updates["owner_id"] = *req.OwnerID
		}
	}
	if req.GroupID != nil {
		if *req.GroupID == 0 {
			updates["group_id"] = nil
		} else {
			updates["group_id"] = *req.GroupID
		}
	}

	if priorityChanged {
		now := time.Now()
		policy := MatchSlaPolicy(*req.Priority)
		if policy != nil {
			updates["sla_policy_id"] = policy.ID
			respDeadline, resDeadline := ComputeSlaDeadlines(policy, now)
			if respDeadline != nil {
				updates["sla_response_deadline"] = respDeadline
			}
			if resDeadline != nil {
				updates["sla_resolution_deadline"] = resDeadline
			}
			updates["sla_response_breached"] = false
			updates["sla_resolution_breached"] = false
		} else {
			updates["sla_policy_id"] = nil
			updates["sla_response_deadline"] = nil
			updates["sla_resolution_deadline"] = nil
			updates["sla_response_breached"] = false
			updates["sla_resolution_breached"] = false
		}
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&ticket).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update ticket"})
			return
		}
	}

	database.DB.Preload("CreatedBy").Preload("AssignedTo").Preload("Owner").Preload("Group").Preload("Tags").Preload("SlaPolicy").First(&ticket, ticket.ID)
	c.JSON(http.StatusOK, ticket)
}

// DeleteTicket DELETE /api/v1/customers/:customerId/tickets/:ticketId
func DeleteTicket(c *gin.Context) {
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

	database.DB.Where("ticket_id = ?", ticket.ID).Delete(&models.TicketMessage{})
	database.DB.Where("ticket_id = ?", ticket.ID).Delete(&models.TicketTag{})
	database.DB.Where("source_ticket_id = ? OR target_ticket_id = ?", ticket.ID, ticket.ID).Delete(&models.TicketLink{})
	database.DB.Delete(&ticket)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// CreateTicketMessage POST /api/v1/customers/:customerId/tickets/:ticketId/messages
func CreateTicketMessage(c *gin.Context) {
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
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}

	msg := models.TicketMessage{
		TicketID: ticket.ID,
		UserID:   userID,
		Body:     req.Body,
	}
	if err := database.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message"})
		return
	}

	if ticket.FirstResponseAt == nil && userID != ticket.CreatedByID && ticket.SlaResponseDeadline != nil {
		now := time.Now()
		database.DB.Model(&ticket).Update("first_response_at", now)
	}

	database.DB.Preload("User").First(&msg, msg.ID)
	msg.Attachments = []models.Attachment{}
	c.JSON(http.StatusCreated, msg)
}

// refreshSlaBreachStatus updates breach flags on a ticket if deadlines have passed.
func refreshSlaBreachStatus(ticket *models.Ticket) {
	now := time.Now()
	changed := false
	if ticket.SlaResponseDeadline != nil && !ticket.SlaResponseBreached && now.After(*ticket.SlaResponseDeadline) {
		if ticket.FirstResponseAt == nil || ticket.FirstResponseAt.After(*ticket.SlaResponseDeadline) {
			ticket.SlaResponseBreached = true
			database.DB.Model(ticket).Update("sla_response_breached", true)
			changed = true
		}
	}
	if ticket.SlaResolutionDeadline != nil && !ticket.SlaResolutionBreached && now.After(*ticket.SlaResolutionDeadline) {
		if ticket.Status != "resolved" && ticket.Status != "closed" {
			ticket.SlaResolutionBreached = true
			database.DB.Model(ticket).Update("sla_resolution_breached", true)
			changed = true
		}
	}
	_ = changed
}

// requireCustomerAccess checks that the user has member or admin access to the customer.
// Admins bypass all checks.
func requireCustomerAccess(customerID, userID uint, role string) error {
	if role == "admin" {
		return nil
	}
	var access models.CustomerAccess
	if err := database.DB.Where("customer_id = ? AND user_id = ?", customerID, userID).First(&access).Error; err != nil {
		return services.ErrForbidden
	}
	return nil
}
