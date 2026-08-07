package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
	"gorm.io/gorm/clause"
)

// parseNullableTimestamp distinguishes an omitted JSON key (raw is empty —
// caller wants to leave the field untouched) from an explicit "null" (caller
// wants to clear it) or a real RFC3339 timestamp — a distinction a plain
// *time.Time field can't make, since both an absent key and an explicit null
// unmarshal to a nil pointer. ok is false when the key was absent or held an
// unparseable value; t is nil when the caller asked to clear the field.
func parseNullableTimestamp(raw json.RawMessage) (t *time.Time, ok bool) {
	if len(raw) == 0 {
		return nil, false
	}
	if string(raw) == "null" {
		return nil, true
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return nil, false
	}
	return &parsed, true
}

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

	autoCloseTickets()
	var tickets []models.Ticket
	q := database.DB.Where("customer_id = ?", customerID)
	if c.Query("include_spam") != "true" {
		q = q.Where("is_spam = false OR is_spam IS NULL")
	}
	// datetime('now') is SQLite-specific — MySQL/PostgreSQL have no such
	// function. GORM's Order() takes a raw string with no parameter binding,
	// so the current time is formatted here instead, as a portable ISO-ish
	// literal every supported driver compares against a DATETIME column correctly.
	nowLiteral := time.Now().UTC().Format("2006-01-02 15:04:05")
	q.Preload("CreatedBy").
		Preload("AssignedTo").
		Preload("Owner").
		Preload("Group").
		Preload("Tags").
		Preload("SlaPolicy").
		Order(fmt.Sprintf("CASE WHEN status = 'pending' AND reminder_at IS NOT NULL AND reminder_at <= '%s' THEN 0 ELSE 1 END, created_at desc", nowLiteral)).
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

	autoCloseTickets()
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id = ?", ticketID, customerID).
		Preload("CreatedBy").
		Preload("AssignedTo").
		Preload("Owner").
		Preload("Group").
		Preload("Tags").
		Preload("SlaPolicy").
		First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	refreshSlaBreachStatus(&ticket)

	// Load description-level attachments (owner_type="ticket")
	if am := LoadAttachments("ticket", []uint{ticket.ID}); len(am[ticket.ID]) > 0 {
		ticket.Attachments = am[ticket.ID]
	}

	// Load all messages flat and build nested tree for arbitrary depth
	var allMessages []models.TicketMessage
	q := database.DB.Where("ticket_id = ?", ticket.ID).Order("created_at asc")
	if role == "customer" {
		q = q.Where("is_private = false OR is_private IS NULL")
	}
	if err := q.Preload("User").Find(&allMessages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load messages"})
		return
	}

	msgIDs := make([]uint, len(allMessages))
	for i, m := range allMessages {
		msgIDs[i] = m.ID
	}
	attachMap := LoadAttachments("ticket_message", msgIDs)

	messageMap := make(map[uint]*models.TicketMessage)
	for i := range allMessages {
		messageMap[allMessages[i].ID] = &allMessages[i]
		allMessages[i].Attachments = attachMap[allMessages[i].ID]
		if allMessages[i].Attachments == nil {
			allMessages[i].Attachments = []models.Attachment{}
		}
	}

	for i := len(allMessages) - 1; i >= 0; i-- {
		if allMessages[i].ParentID != nil {
			if parent, ok := messageMap[*allMessages[i].ParentID]; ok {
				if parent.Replies == nil {
					parent.Replies = []models.TicketMessage{}
				}
				parent.Replies = append(parent.Replies, allMessages[i])
			}
		}
	}

	ticket.Messages = nil
	for i := range allMessages {
		if allMessages[i].ParentID == nil {
			ticket.Messages = append(ticket.Messages, allMessages[i])
		}
	}

	attachTicketChecklist(&ticket)

	// Record / refresh the view timestamp for the current user.
	database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ticket_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"viewed_at"}),
	}).Create(&models.TicketView{TicketID: ticket.ID, UserID: userID, ViewedAt: time.Now()})

	c.JSON(http.StatusOK, ticket)
}

// GetTicketViewers returns all users who have viewed the ticket, ordered by most recent first.
func GetTicketViewers(c *gin.Context) {
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
	var viewers []models.TicketView
	database.DB.Where("ticket_id = ?", ticketID).
		Preload("User").
		Order("viewed_at desc").
		Find(&viewers)
	c.JSON(http.StatusOK, viewers)
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
	if err := requireNotCustomerRole(role); err != nil {
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

	cid := uint(customerID)
	ticket := models.Ticket{
		CustomerID:   &cid,
		Title:        req.Title,
		Description:  req.Description,
		Type:         req.Type,
		Priority:     req.Priority,
		Status:       "new",
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
	database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "created"})
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
	if err := requireNotCustomerRole(role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id = ?", ticketID, customerID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	var req struct {
		Title       *string         `json:"title"`
		Description *string         `json:"description"`
		Type        *string         `json:"type"`
		Status      *string         `json:"status"`
		Priority    *string         `json:"priority"`
		AssignedTo  *uint           `json:"assigned_to_id"`
		OwnerID     *uint           `json:"owner_id"`
		GroupID     *uint           `json:"group_id"`
		ReminderAt  json.RawMessage `json:"reminder_at"`
		CloseAt     json.RawMessage `json:"close_at"`
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
		validStatuses := map[string]bool{"new": true, "open": true, "pending": true, "pending_close": true, "closed": true}
		if validStatuses[*req.Status] {
			if statusRequiresChecklistComplete(*req.Status) && ticketChecklistBlocksClose(ticket.ID) {
				c.JSON(http.StatusBadRequest, gin.H{"error": errChecklistIncomplete})
				return
			}
			updates["status"] = *req.Status
		}
	}
	if t, ok := parseNullableTimestamp(req.ReminderAt); ok {
		updates["reminder_at"] = t // nil clears
	}
	if t, ok := parseNullableTimestamp(req.CloseAt); ok {
		updates["close_at"] = t // nil clears
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

	// Log history for relevant changes
	if req.Title != nil && *req.Title != ticket.Title {
		database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "title_changed", Detail: *req.Title})
	}
	if req.Status != nil && *req.Status != ticket.Status {
		database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "status_changed", Detail: *req.Status})
	}
	if req.Priority != nil && *req.Priority != ticket.Priority {
		database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "priority_changed", Detail: *req.Priority})
	}
	if req.Type != nil && *req.Type != ticket.Type {
		database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "type_changed", Detail: *req.Type})
	}
	if req.AssignedTo != nil {
		oldAssignee := ticket.AssignedToID
		newVal := *req.AssignedTo
		if newVal == 0 {
			newVal = 0
		}
		if oldAssignee == nil || *oldAssignee != newVal {
			detail := "unassigned"
			if newVal != 0 {
				var u models.User
				database.DB.First(&u, newVal)
				detail = u.DisplayName
				if detail == "" {
					detail = u.Username
				}
			}
			database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "assignee_changed", Detail: detail})
		}
	}
	if req.OwnerID != nil {
		newVal := *req.OwnerID
		if newVal == 0 {
			newVal = 0
		}
		if ticket.OwnerID == nil || *ticket.OwnerID != newVal {
			detail := "unassigned"
			if newVal != 0 {
				var u models.User
				database.DB.First(&u, newVal)
				detail = u.DisplayName
				if detail == "" {
					detail = u.Username
				}
			}
			database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "owner_changed", Detail: detail})
		}
	}
	if req.GroupID != nil {
		newVal := *req.GroupID
		if newVal == 0 {
			newVal = 0
		}
		if ticket.GroupID == nil || *ticket.GroupID != newVal {
			detail := "unassigned"
			if newVal != 0 {
				var g models.UserGroup
				database.DB.First(&g, newVal)
				detail = g.Name
			}
			database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "group_changed", Detail: detail})
		}
	}
	if req.ReminderAt != nil || req.CloseAt != nil {
		database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "dates_changed", Detail: "reminder / close"})
	}

	database.DB.Preload("CreatedBy").Preload("AssignedTo").Preload("Owner").Preload("Group").Preload("Tags").Preload("SlaPolicy").First(&ticket, ticket.ID)
	attachTicketChecklist(&ticket)
	c.JSON(http.StatusOK, ticket)
}

// MoveTicket PUT /api/v1/customers/:customerId/tickets/:ticketId/move
// Reassigns a ticket from one customer to another. Requires member access on both.
func MoveTicket(c *gin.Context) {
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
		TargetCustomerID uint `json:"target_customer_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.TargetCustomerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_customer_id is required"})
		return
	}
	if req.TargetCustomerID == uint(customerID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ticket is already assigned to that customer"})
		return
	}
	if err := requireCustomerAccess(req.TargetCustomerID, userID, role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "no access to target customer"})
		return
	}

	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id = ?", ticketID, customerID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	if err := database.DB.Model(&ticket).Update("customer_id", req.TargetCustomerID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to move ticket"})
		return
	}

	var targetCustomer models.Customer
	database.DB.First(&targetCustomer, req.TargetCustomerID)
	database.DB.Create(&models.TicketHistory{
		TicketID:  ticket.ID,
		UserID:    userID,
		EventType: "customer_moved",
		Detail:    targetCustomer.Name,
	})

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
	if err := requireNotCustomerRole(role); err != nil {
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
	database.DB.Where("ticket_id = ?", ticket.ID).Delete(&models.TicketChecklistItem{})
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
		Body      string `json:"body" binding:"required"`
		IsPrivate bool   `json:"is_private"`
		ParentID  *uint  `json:"parent_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}
	// Customer-role users cannot post internal/private notes.
	if req.IsPrivate && role == "customer" {
		req.IsPrivate = false
	}

	// Validate parent message belongs to the same ticket if set.
	if req.ParentID != nil {
		var parent models.TicketMessage
		if err := database.DB.Where("id = ? AND ticket_id = ?", *req.ParentID, ticket.ID).First(&parent).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parent message not found"})
			return
		}
	}

	msg := models.TicketMessage{
		TicketID:  ticket.ID,
		UserID:    userID,
		Body:      req.Body,
		IsPrivate: req.IsPrivate,
		ParentID:  req.ParentID,
	}
	if err := database.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message"})
		return
	}

	if !msg.IsPrivate && ticket.FirstResponseAt == nil && userID != ticket.CreatedByID && ticket.SlaResponseDeadline != nil {
		now := time.Now()
		database.DB.Model(&ticket).Update("first_response_at", now)
	}

	database.DB.Preload("User").First(&msg, msg.ID)
	if !msg.IsPrivate {
		sendEmailReply(&ticket, msg.Body)
		if ticket.FromEmail != nil && *ticket.FromEmail != "" {
			database.DB.Model(&msg).Update("email_sent", true)
			msg.EmailSent = true
		}
	}
	msg.Attachments = []models.Attachment{}
	database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "comment_added"})
	c.JSON(http.StatusCreated, msg)
}

// UpdateTicketMessage PATCH /api/v1/customers/:customerId/tickets/:ticketId/messages/:msgId
func UpdateTicketMessage(c *gin.Context) {
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
	msgID, err := strconv.ParseUint(c.Param("msgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
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

	var msg models.TicketMessage
	if err := database.DB.Where("id = ? AND ticket_id = ?", msgID, ticket.ID).First(&msg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	var req struct {
		IsPrivate *bool `json:"is_private"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Customer-role users cannot make messages private.
	if req.IsPrivate != nil && *req.IsPrivate && role == "customer" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	updates := map[string]interface{}{}
	if req.IsPrivate != nil {
		updates["is_private"] = *req.IsPrivate
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&msg).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update message"})
			return
		}
	}
	database.DB.Preload("User").First(&msg, msg.ID)
	database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "comment_updated"})
	c.JSON(http.StatusOK, msg)
}

// GetTicketHistory GET /api/v1/customers/:customerId/tickets/:ticketId/history
func GetTicketHistory(c *gin.Context) {
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

	var history []models.TicketHistory
	database.DB.Preload("User").
		Where("ticket_id = ?", ticketID).
		Order("created_at desc").
		Find(&history)
	if history == nil {
		history = []models.TicketHistory{}
	}
	c.JSON(http.StatusOK, history)
}

// autoCloseTickets closes any pending_close tickets whose close_at has passed.
func autoCloseTickets() {
	// datetime('now') is SQLite-specific — MySQL/PostgreSQL have no such
	// function. Passing time.Now() as a bound parameter compares correctly
	// against the close_at column on every supported driver.
	var tickets []models.Ticket
	database.DB.Where("status = 'pending_close' AND close_at IS NOT NULL AND close_at <= ?", time.Now()).Find(&tickets)
	for _, t := range tickets {
		if ticketChecklistBlocksClose(t.ID) {
			continue
		}
		database.DB.Model(&t).Update("status", "closed")
	}
	// Clear close_at on already-closed tickets that still have it set
	database.DB.Model(&models.Ticket{}).
		Where("status = 'closed' AND close_at IS NOT NULL").
		Update("close_at", nil)
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
		if ticket.Status != "pending_close" && ticket.Status != "closed" {
			ticket.SlaResolutionBreached = true
			database.DB.Model(ticket).Update("sla_resolution_breached", true)
			changed = true
		}
	}
	_ = changed
}

// requireCustomerAccess checks that the user has member or admin access to the customer,
// honouring both direct CustomerAccess rows and group-based GroupCustomerAccess rows.
// Admins bypass all checks. Customer-role users are checked against their CustomerAccess rows.
func requireCustomerAccess(customerID, userID uint, role string) error {
	if role == "admin" {
		return nil
	}
	accessible := getAccessibleCustomerRoles(userID)
	if _, ok := accessible[customerID]; !ok {
		return services.ErrForbidden
	}
	return nil
}

// requireNotCustomerRole returns ErrForbidden for "customer" global-role users.
// Applied to ticket write operations (create, update, delete, tags, macros, spam, etc.)
// that customer-portal users are not permitted to perform.
func requireNotCustomerRole(role string) error {
	if role == "customer" {
		return services.ErrForbidden
	}
	return nil
}

// ListInboxTickets GET /api/v1/tickets/inbox
func ListInboxTickets(c *gin.Context) {
	autoCloseTickets()
	var tickets []models.Ticket
	q := database.DB.Where("customer_id IS NULL")
	if c.Query("include_spam") != "true" {
		q = q.Where("is_spam = false OR is_spam IS NULL")
	}
	q.Preload("CreatedBy").
		Preload("AssignedTo").
		Preload("Owner").
		Preload("Group").
		Preload("Tags").
		Preload("SlaPolicy").
		Order("created_at desc").
		Find(&tickets)
	if tickets == nil {
		tickets = []models.Ticket{}
	}
	for i := range tickets {
		refreshSlaBreachStatus(&tickets[i])
	}
	c.JSON(http.StatusOK, tickets)
}

// GetInboxTicket GET /api/v1/tickets/inbox/:ticketId
func GetInboxTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	autoCloseTickets()
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id IS NULL", ticketID).
		Preload("CreatedBy").
		Preload("AssignedTo").
		Preload("Owner").
		Preload("Group").
		Preload("Tags").
		Preload("SlaPolicy").
		First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	refreshSlaBreachStatus(&ticket)
	if am := LoadAttachments("ticket", []uint{ticket.ID}); len(am[ticket.ID]) > 0 {
		ticket.Attachments = am[ticket.ID]
	}

	var allMessages []models.TicketMessage
	if err := database.DB.Where("ticket_id = ?", ticket.ID).Order("created_at asc").
		Preload("User").Find(&allMessages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load messages"})
		return
	}

	msgIDs := make([]uint, len(allMessages))
	for i, m := range allMessages {
		msgIDs[i] = m.ID
	}
	attachMap := LoadAttachments("ticket_message", msgIDs)

	messageMap := make(map[uint]*models.TicketMessage)
	for i := range allMessages {
		messageMap[allMessages[i].ID] = &allMessages[i]
		allMessages[i].Attachments = attachMap[allMessages[i].ID]
		if allMessages[i].Attachments == nil {
			allMessages[i].Attachments = []models.Attachment{}
		}
	}

	for i := len(allMessages) - 1; i >= 0; i-- {
		if allMessages[i].ParentID != nil {
			if parent, ok := messageMap[*allMessages[i].ParentID]; ok {
				if parent.Replies == nil {
					parent.Replies = []models.TicketMessage{}
				}
				parent.Replies = append(parent.Replies, allMessages[i])
			}
		}
	}

	ticket.Messages = nil
	for i := range allMessages {
		if allMessages[i].ParentID == nil {
			ticket.Messages = append(ticket.Messages, allMessages[i])
		}
	}
	attachTicketChecklist(&ticket)

	database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ticket_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"viewed_at"}),
	}).Create(&models.TicketView{TicketID: ticket.ID, UserID: userID, ViewedAt: time.Now()})

	c.JSON(http.StatusOK, ticket)
}

// GetInboxTicketViewers returns all users who viewed an inbox ticket, ordered by most recent first.
func GetInboxTicketViewers(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	var viewers []models.TicketView
	database.DB.Where("ticket_id = ?", ticketID).
		Preload("User").
		Order("viewed_at desc").
		Find(&viewers)
	c.JSON(http.StatusOK, viewers)
}

// UpdateInboxTicket PUT /api/v1/tickets/inbox/:ticketId
// Supports all standard ticket fields plus customer_id to move to a customer.
func UpdateInboxTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetGlobalRole(c)
	if err := requireNotCustomerRole(role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id IS NULL", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	var req struct {
		Title       *string         `json:"title"`
		Description *string         `json:"description"`
		Type        *string         `json:"type"`
		Status      *string         `json:"status"`
		Priority    *string         `json:"priority"`
		AssignedTo  *uint           `json:"assigned_to_id"`
		OwnerID     *uint           `json:"owner_id"`
		GroupID     *uint           `json:"group_id"`
		ReminderAt  json.RawMessage `json:"reminder_at"`
		CloseAt     json.RawMessage `json:"close_at"`
		CustomerID  *uint           `json:"customer_id"`
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
		validStatuses := map[string]bool{"new": true, "open": true, "pending": true, "pending_close": true, "closed": true}
		if validStatuses[*req.Status] {
			if statusRequiresChecklistComplete(*req.Status) && ticketChecklistBlocksClose(ticket.ID) {
				c.JSON(http.StatusBadRequest, gin.H{"error": errChecklistIncomplete})
				return
			}
			updates["status"] = *req.Status
		}
	}
	if t, ok := parseNullableTimestamp(req.ReminderAt); ok {
		updates["reminder_at"] = t // nil clears
	}
	if t, ok := parseNullableTimestamp(req.CloseAt); ok {
		updates["close_at"] = t // nil clears
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
	if req.CustomerID != nil && *req.CustomerID > 0 {
		updates["customer_id"] = *req.CustomerID
		database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "customer_assigned", Detail: strconv.FormatUint(uint64(*req.CustomerID), 10)})
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

	if req.Status != nil && *req.Status != ticket.Status {
		database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "status_changed", Detail: *req.Status})
	}
	if req.Priority != nil && *req.Priority != ticket.Priority {
		database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "priority_changed", Detail: *req.Priority})
	}

	database.DB.Where("id = ?", ticket.ID).
		Preload("CreatedBy").Preload("AssignedTo").Preload("Owner").Preload("Group").Preload("Tags").Preload("SlaPolicy").
		First(&ticket)
	refreshSlaBreachStatus(&ticket)
	attachTicketChecklist(&ticket)
	c.JSON(http.StatusOK, ticket)
}

// CreateInboxTicketMessage POST /api/v1/tickets/inbox/:ticketId/messages
func CreateInboxTicketMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id IS NULL", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	var req struct {
		Body      string `json:"body" binding:"required"`
		IsPrivate bool   `json:"is_private"`
		ParentID  *uint  `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}
	if req.ParentID != nil {
		var parent models.TicketMessage
		if err := database.DB.Where("id = ? AND ticket_id = ?", *req.ParentID, ticket.ID).First(&parent).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parent message not found"})
			return
		}
	}
	msg := models.TicketMessage{
		TicketID:  ticket.ID,
		UserID:    userID,
		Body:      req.Body,
		IsPrivate: req.IsPrivate,
		ParentID:  req.ParentID,
	}
	if err := database.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message"})
		return
	}
	if !msg.IsPrivate && ticket.FirstResponseAt == nil {
		now := time.Now()
		database.DB.Model(&ticket).Update("first_response_at", now)
	}
	database.DB.Preload("User").First(&msg, msg.ID)
	if !msg.IsPrivate {
		sendEmailReply(&ticket, msg.Body)
	}
	c.JSON(http.StatusCreated, msg)
}

// UpdateInboxTicketMessage PATCH /api/v1/tickets/inbox/:ticketId/messages/:msgId
func UpdateInboxTicketMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	msgID, err := strconv.ParseUint(c.Param("msgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id IS NULL", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	var msg models.TicketMessage
	if err := database.DB.Where("id = ? AND ticket_id = ?", msgID, ticket.ID).First(&msg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}
	var req struct {
		IsPrivate *bool `json:"is_private"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updates := map[string]interface{}{}
	if req.IsPrivate != nil {
		updates["is_private"] = *req.IsPrivate
	}
	if len(updates) > 0 {
		if err := database.DB.Model(&msg).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update message"})
			return
		}
	}
	database.DB.Preload("User").First(&msg, msg.ID)
	database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "comment_updated"})
	c.JSON(http.StatusOK, msg)
}

// sendEmailReply sends a reply to the ticket's original sender when SMTP is configured.
// Sets proper threading headers so the customer's reply is threaded back to the same ticket.
func sendEmailReply(ticket *models.Ticket, replyBody string) {
	if ticket.FromEmail == nil || *ticket.FromEmail == "" {
		return
	}
	svc := services.GetEmailService()
	if svc == nil {
		return
	}
	subject := "Re: [#" + strconv.Itoa(int(ticket.ID)) + "] " + ticket.Title
	inReplyTo := ""
	if ticket.EmailMessageID != nil {
		inReplyTo = *ticket.EmailMessageID
	}
	go svc.SendReply(*ticket.FromEmail, subject, replyBody, ticket.ID, inReplyTo, "") //nolint:errcheck
}

// CreateInboxTicket POST /api/v1/tickets/inbox
func CreateInboxTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Priority    string `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	ticketType := req.Type
	if ticketType == "" {
		ticketType = "service_request"
	}
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	ticket := models.Ticket{
		Title:       req.Title,
		Description: req.Description,
		Type:        ticketType,
		Priority:    priority,
		Status:      "new",
		CreatedByID: userID,
	}
	if err := database.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create ticket"})
		return
	}
	database.DB.Preload("CreatedBy").First(&ticket, ticket.ID)
	c.JSON(http.StatusCreated, ticket)
}

// DeleteInboxTicket DELETE /api/v1/tickets/inbox/:ticketId
func DeleteInboxTicket(c *gin.Context) {
	role := middleware.GetGlobalRole(c)
	if err := requireNotCustomerRole(role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id IS NULL", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	database.DB.Where("ticket_id = ?", ticket.ID).Delete(&models.TicketMessage{})
	database.DB.Where("ticket_id = ?", ticket.ID).Delete(&models.TicketTag{})
	database.DB.Where("ticket_id = ?", ticket.ID).Delete(&models.TicketChecklistItem{})
	database.DB.Where("source_ticket_id = ? OR target_ticket_id = ?", ticket.ID, ticket.ID).Delete(&models.TicketLink{})
	database.DB.Delete(&ticket)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetTicketRawEmail GET /api/v1/customers/:customerId/tickets/:ticketId/raw-email
func GetTicketRawEmail(c *gin.Context) {
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

	var rawEmail *string
	database.DB.Model(&models.Ticket{}).
		Select("raw_email").
		Where("id = ? AND customer_id = ?", ticketID, customerID).
		Scan(&rawEmail)

	if rawEmail == nil || *rawEmail == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no raw email available"})
		return
	}

	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(*rawEmail))
}

// MarkSpam POST /api/v1/customers/:customerId/tickets/:ticketId/spam
func MarkSpam(c *gin.Context) {
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
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id = ?", ticketID, customerID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	database.DB.Model(&ticket).Updates(map[string]any{"is_spam": true, "status": "closed"})
	database.DB.First(&ticket, ticket.ID)
	c.JSON(http.StatusOK, ticket)
}

// UnmarkSpam DELETE /api/v1/customers/:customerId/tickets/:ticketId/spam
func UnmarkSpam(c *gin.Context) {
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
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id = ?", ticketID, customerID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	database.DB.Model(&ticket).Updates(map[string]any{"is_spam": false, "status": "open"})
	database.DB.First(&ticket, ticket.ID)
	c.JSON(http.StatusOK, ticket)
}

// MarkInboxSpam POST /api/v1/tickets/inbox/:ticketId/spam
func MarkInboxSpam(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id IS NULL", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	database.DB.Model(&ticket).Updates(map[string]any{"is_spam": true, "status": "closed"})
	database.DB.First(&ticket, ticket.ID)
	c.JSON(http.StatusOK, ticket)
}

// UnmarkInboxSpam DELETE /api/v1/tickets/inbox/:ticketId/spam
func UnmarkInboxSpam(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id IS NULL", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	database.DB.Model(&ticket).Updates(map[string]any{"is_spam": false, "status": "open"})
	database.DB.First(&ticket, ticket.ID)
	c.JSON(http.StatusOK, ticket)
}
