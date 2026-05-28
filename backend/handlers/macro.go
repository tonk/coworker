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
	"gorm.io/gorm"
)

// AdminListMacros GET /api/v1/admin/macros
func AdminListMacros(c *gin.Context) {
	var macros []models.Macro
	database.DB.Order("sort_order asc, id asc").Find(&macros)
	if macros == nil {
		macros = []models.Macro{}
	}
	c.JSON(http.StatusOK, macros)
}

// ListMacros GET /api/v1/macros — active macros visible to all helpdesk users
func ListMacros(c *gin.Context) {
	var macros []models.Macro
	database.DB.Where("is_active = ?", true).Order("sort_order asc, id asc").Find(&macros)
	if macros == nil {
		macros = []models.Macro{}
	}
	c.JSON(http.StatusOK, macros)
}

// AdminCreateMacro POST /api/v1/admin/macros
func AdminCreateMacro(c *gin.Context) {
	var req struct {
		Name        string              `json:"name" binding:"required"`
		Description string              `json:"description"`
		Actions     models.MacroActions `json:"actions"`
		IsActive    *bool               `json:"is_active"`
		SortOrder   int                 `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.Actions == nil {
		req.Actions = models.MacroActions{}
	}
	macro := models.Macro{
		Name:        req.Name,
		Description: req.Description,
		Actions:     req.Actions,
		IsActive:    isActive,
		SortOrder:   req.SortOrder,
	}
	if err := database.DB.Create(&macro).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create macro"})
		return
	}
	c.JSON(http.StatusCreated, macro)
}

// AdminUpdateMacro PUT /api/v1/admin/macros/:id
func AdminUpdateMacro(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var macro models.Macro
	if err := database.DB.First(&macro, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "macro not found"})
		return
	}
	var req struct {
		Name        *string             `json:"name"`
		Description *string             `json:"description"`
		Actions     models.MacroActions `json:"actions"`
		IsActive    *bool               `json:"is_active"`
		SortOrder   *int                `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Actions != nil {
		updates["actions"] = req.Actions
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	database.DB.Model(&macro).Updates(updates)
	database.DB.First(&macro, macro.ID)
	c.JSON(http.StatusOK, macro)
}

// AdminDeleteMacro DELETE /api/v1/admin/macros/:id
func AdminDeleteMacro(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var macro models.Macro
	if err := database.DB.First(&macro, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "macro not found"})
		return
	}
	database.DB.Delete(&macro)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ApplyMacro POST /api/v1/customers/:customerId/tickets/:ticketId/macros/:macroId
func ApplyMacro(c *gin.Context) {
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
	macroID, err := strconv.ParseUint(c.Param("macroId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid macro id"})
		return
	}
	if err := requireCustomerAccess(uint(customerID), userID, role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var ticket models.Ticket
	if err := database.DB.Preload("CreatedBy").Where("id = ? AND customer_id = ?", ticketID, customerID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	var macro models.Macro
	if err := database.DB.First(&macro, uint(macroID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "macro not found"})
		return
	}
	applyMacroActions(c, &ticket, &macro, userID)
}

// ApplyInboxMacro POST /api/v1/tickets/inbox/:ticketId/macros/:macroId
func ApplyInboxMacro(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	macroID, err := strconv.ParseUint(c.Param("macroId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid macro id"})
		return
	}
	var ticket models.Ticket
	if err := database.DB.Preload("CreatedBy").Where("id = ? AND customer_id IS NULL", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	var macro models.Macro
	if err := database.DB.First(&macro, uint(macroID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "macro not found"})
		return
	}
	applyMacroActions(c, &ticket, &macro, userID)
}

var validMacroStatuses = map[string]bool{
	"new": true, "open": true, "pending": true, "pending_close": true, "closed": true,
}
var validMacroPriorities = map[string]bool{
	"low": true, "medium": true, "high": true, "critical": true,
}
var validMacroTypes = map[string]bool{
	"incident": true, "problem": true, "service_request": true, "change_request": true,
}

func applyMacroActions(c *gin.Context, ticket *models.Ticket, macro *models.Macro, userID uint) {
	updates := map[string]any{}
	var pendingMessages []string

	for _, action := range macro.Actions {
		switch action.Type {
		case "set_status":
			if validMacroStatuses[action.Value] {
				updates["status"] = action.Value
			}
		case "set_priority":
			if validMacroPriorities[action.Value] {
				updates["priority"] = action.Value
			}
		case "set_type":
			if validMacroTypes[action.Value] {
				updates["type"] = action.Value
			}
		case "add_message":
			if action.Value != "" {
				body := expandPlaceholders(action.Value, ticket, userID)
				pendingMessages = append(pendingMessages, body)
			}
		case "add_tag":
			if action.Value != "" {
				tag := models.TicketTag{TicketID: ticket.ID, Name: action.Value}
				database.DB.FirstOrCreate(&tag, models.TicketTag{TicketID: ticket.ID, Name: action.Value})
			}
		}
	}

	if len(updates) > 0 {
		database.DB.Model(ticket).Updates(updates)
	}

	var updated models.Ticket
	database.DB.Where("id = ?", ticket.ID).
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
		First(&updated)
	refreshSlaBreachStatus(&updated)

	// Load message attachments
	msgIDs := make([]uint, len(updated.Messages))
	for i, m := range updated.Messages {
		msgIDs[i] = m.ID
	}
	attachMap := LoadAttachments("ticket_message", msgIDs)
	for i := range updated.Messages {
		updated.Messages[i].Attachments = attachMap[updated.Messages[i].ID]
		if updated.Messages[i].Attachments == nil {
			updated.Messages[i].Attachments = []models.Attachment{}
		}
	}

	if pendingMessages == nil {
		pendingMessages = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"ticket": updated, "macro_messages": pendingMessages})
}

// expandPlaceholders replaces template variables in macro message bodies.
// Supported: {email}, {fname}, {name}, {subject}, {ticket_id}, {agent}, {agent_fname}
func expandPlaceholders(tmpl string, ticket *models.Ticket, agentID uint) string {
	fromEmail := ""
	if ticket.FromEmail != nil {
		fromEmail = *ticket.FromEmail
	}
	senderName, senderFname := parseSenderInfo(fromEmail, &ticket.CreatedBy)

	var agent models.User
	agentName, agentFname := "", ""
	if err := database.DB.First(&agent, agentID).Error; err == nil {
		agentName = agent.DisplayName
		if agentName == "" {
			agentName = strings.TrimSpace(agent.FirstName + " " + agent.LastName)
		}
		if agentName == "" {
			agentName = agent.Username
		}
		agentFname = agent.FirstName
		if agentFname == "" {
			if parts := strings.Fields(agentName); len(parts) > 0 {
				agentFname = parts[0]
			}
		}
	}

	r := strings.NewReplacer(
		"{email}", fromEmail,
		"{fname}", senderFname,
		"{name}", senderName,
		"{subject}", ticket.Title,
		"{ticket_id}", fmt.Sprintf("#%d", ticket.ID),
		"{agent}", agentName,
		"{agent_fname}", agentFname,
	)
	return r.Replace(tmpl)
}

// parseSenderInfo extracts display name and first name from a From-style email
// string (e.g. "John Doe <john@example.com>") or falls back to the ticket creator.
func parseSenderInfo(fromEmail string, createdBy *models.User) (name, fname string) { //nolint:unparam
	if fromEmail != "" {
		if idx := strings.Index(fromEmail, "<"); idx > 0 {
			display := strings.TrimSpace(fromEmail[:idx])
			if display != "" {
				parts := strings.Fields(display)
				return display, parts[0]
			}
		}
		// plain address — use local part as first name
		if idx := strings.Index(fromEmail, "@"); idx > 0 {
			return fromEmail, fromEmail[:idx]
		}
		return fromEmail, fromEmail
	}
	if createdBy != nil {
		display := createdBy.DisplayName
		if display == "" {
			display = strings.TrimSpace(createdBy.FirstName + " " + createdBy.LastName)
		}
		if display == "" {
			display = createdBy.Username
		}
		fn := createdBy.FirstName
		if fn == "" {
			if parts := strings.Fields(display); len(parts) > 0 {
				fn = parts[0]
			}
		}
		return display, fn
	}
	return "", ""
}
