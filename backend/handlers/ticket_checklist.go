package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

var errChecklistTemplateEmpty = errors.New("checklist template has no items")

const errChecklistIncomplete = "complete all checklist items before closing"

// attachTicketChecklist loads checklist items onto a ticket for JSON responses.
func attachTicketChecklist(ticket *models.Ticket) {
	var items []models.TicketChecklistItem
	database.DB.Where("ticket_id = ?", ticket.ID).Order("position asc, id asc").Find(&items)
	if items == nil {
		items = []models.TicketChecklistItem{}
	}
	ticket.ChecklistItems = items
}

// ticketChecklistBlocksClose returns true when the ticket has incomplete checklist items.
func ticketChecklistBlocksClose(ticketID uint) bool {
	var count int64
	database.DB.Model(&models.TicketChecklistItem{}).
		Where("ticket_id = ? AND is_completed = ?", ticketID, false).
		Count(&count)
	return count > 0
}

func statusRequiresChecklistComplete(status string) bool {
	return status == "closed" || status == "pending_close"
}

// AdminListTicketChecklistTemplates GET /api/v1/admin/ticket-checklist-templates
func AdminListTicketChecklistTemplates(c *gin.Context) {
	var templates []models.TicketChecklistTemplate
	database.DB.Order("sort_order asc, id asc").Find(&templates)
	if templates == nil {
		templates = []models.TicketChecklistTemplate{}
	}
	c.JSON(http.StatusOK, templates)
}

// ListTicketChecklistTemplates GET /api/v1/ticket-checklist-templates
func ListTicketChecklistTemplates(c *gin.Context) {
	var templates []models.TicketChecklistTemplate
	database.DB.Where("is_active = ?", true).Order("sort_order asc, id asc").Find(&templates)
	if templates == nil {
		templates = []models.TicketChecklistTemplate{}
	}
	c.JSON(http.StatusOK, templates)
}

// AdminCreateTicketChecklistTemplate POST /api/v1/admin/ticket-checklist-templates
func AdminCreateTicketChecklistTemplate(c *gin.Context) {
	var req struct {
		Name        string                              `json:"name" binding:"required"`
		Description string                              `json:"description"`
		Items       models.TicketChecklistTemplateItems `json:"items"`
		IsActive    *bool                               `json:"is_active"`
		SortOrder   int                                 `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.Items == nil {
		req.Items = models.TicketChecklistTemplateItems{}
	}
	tmpl := models.TicketChecklistTemplate{
		Name:        req.Name,
		Description: req.Description,
		Items:       req.Items,
		IsActive:    isActive,
		SortOrder:   req.SortOrder,
	}
	if err := database.DB.Create(&tmpl).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create checklist template"})
		return
	}
	c.JSON(http.StatusCreated, tmpl)
}

// AdminUpdateTicketChecklistTemplate PUT /api/v1/admin/ticket-checklist-templates/:id
func AdminUpdateTicketChecklistTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var tmpl models.TicketChecklistTemplate
	if err := database.DB.First(&tmpl, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "checklist template not found"})
		return
	}
	var req struct {
		Name        *string                              `json:"name"`
		Description *string                              `json:"description"`
		Items       *models.TicketChecklistTemplateItems `json:"items"`
		IsActive    *bool                                `json:"is_active"`
		SortOrder   *int                                 `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Items != nil {
		updates["items"] = *req.Items
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if len(updates) > 0 {
		database.DB.Model(&tmpl).Updates(updates)
	}
	database.DB.First(&tmpl, id)
	c.JSON(http.StatusOK, tmpl)
}

// AdminDeleteTicketChecklistTemplate DELETE /api/v1/admin/ticket-checklist-templates/:id
func AdminDeleteTicketChecklistTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var tmpl models.TicketChecklistTemplate
	if err := database.DB.First(&tmpl, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "checklist template not found"})
		return
	}
	database.DB.Delete(&tmpl)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func applyChecklistTemplateToTicket(ticket *models.Ticket, templateID uint) error {
	if ticket.ChecklistTemplateID != nil {
		return services.ErrAlreadyExists
	}
	var tmpl models.TicketChecklistTemplate
	if err := database.DB.Where("id = ? AND is_active = ?", templateID, true).First(&tmpl).Error; err != nil {
		return services.ErrNotFound
	}
	if len(tmpl.Items) == 0 {
		return errChecklistTemplateEmpty
	}
	tx := database.DB.Begin()
	tid := templateID
	if err := tx.Model(ticket).Update("checklist_template_id", tid).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i, body := range tmpl.Items {
		if body == "" {
			continue
		}
		item := models.TicketChecklistItem{
			TicketID: ticket.ID,
			Body:     body,
			Position: float64(i+1) * 1000,
		}
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	ticket.ChecklistTemplateID = &tid
	return nil
}

// ApplyTicketChecklistTemplate POST /api/v1/customers/:customerId/tickets/:ticketId/checklist/templates/:templateId
func ApplyTicketChecklistTemplate(c *gin.Context) {
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
	templateID, err := strconv.ParseUint(c.Param("templateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
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
	if err := applyChecklistTemplateToTicket(&ticket, uint(templateID)); err != nil {
		if err == services.ErrAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "checklist already applied"})
			return
		}
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "checklist template not found"})
			return
		}
		if err == errChecklistTemplateEmpty {
			c.JSON(http.StatusBadRequest, gin.H{"error": "checklist template has no items"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to apply checklist"})
		return
	}
	database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "checklist_applied"})
	database.DB.Preload("CreatedBy").Preload("AssignedTo").Preload("Owner").Preload("Group").Preload("Tags").Preload("SlaPolicy").First(&ticket, ticket.ID)
	attachTicketChecklist(&ticket)
	c.JSON(http.StatusOK, ticket)
}

// ApplyInboxTicketChecklistTemplate POST /api/v1/tickets/inbox/:ticketId/checklist/templates/:templateId
func ApplyInboxTicketChecklistTemplate(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	templateID, err := strconv.ParseUint(c.Param("templateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}
	var ticket models.Ticket
	if err := database.DB.Where("id = ? AND customer_id IS NULL", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	if err := applyChecklistTemplateToTicket(&ticket, uint(templateID)); err != nil {
		if err == services.ErrAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "checklist already applied"})
			return
		}
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "checklist template not found"})
			return
		}
		if err == errChecklistTemplateEmpty {
			c.JSON(http.StatusBadRequest, gin.H{"error": "checklist template has no items"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to apply checklist"})
		return
	}
	database.DB.Create(&models.TicketHistory{TicketID: ticket.ID, UserID: userID, EventType: "checklist_applied"})
	database.DB.Preload("CreatedBy").Preload("AssignedTo").Preload("Owner").Preload("Group").Preload("Tags").Preload("SlaPolicy").First(&ticket, ticket.ID)
	attachTicketChecklist(&ticket)
	c.JSON(http.StatusOK, ticket)
}

func updateTicketChecklistItem(c *gin.Context, ticketID uint, customerID *uint) {
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var ticket models.Ticket
	q := database.DB.Where("id = ?", ticketID)
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	} else {
		q = q.Where("customer_id IS NULL")
	}
	if err := q.First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	var item models.TicketChecklistItem
	if err := database.DB.Where("id = ? AND ticket_id = ?", itemID, ticketID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	var req struct {
		IsCompleted *bool `json:"is_completed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.IsCompleted != nil {
		database.DB.Model(&item).Update("is_completed", *req.IsCompleted)
	}
	database.DB.First(&item, item.ID)
	c.JSON(http.StatusOK, item)
}

// UpdateTicketChecklistItem PUT /api/v1/customers/:customerId/tickets/:ticketId/checklist/:itemId
func UpdateTicketChecklistItem(c *gin.Context) {
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
	cid := uint(customerID)
	updateTicketChecklistItem(c, uint(ticketID), &cid)
}

// UpdateInboxTicketChecklistItem PUT /api/v1/tickets/inbox/:ticketId/checklist/:itemId
func UpdateInboxTicketChecklistItem(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	updateTicketChecklistItem(c, uint(ticketID), nil)
}

func deleteTicketChecklistItem(c *gin.Context, ticketID uint, customerID *uint) {
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var ticket models.Ticket
	q := database.DB.Where("id = ?", ticketID)
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	} else {
		q = q.Where("customer_id IS NULL")
	}
	if err := q.First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	var item models.TicketChecklistItem
	if err := database.DB.Where("id = ? AND ticket_id = ?", itemID, ticketID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	database.DB.Delete(&item)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// DeleteTicketChecklistItem DELETE /api/v1/customers/:customerId/tickets/:ticketId/checklist/:itemId
func DeleteTicketChecklistItem(c *gin.Context) {
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
	cid := uint(customerID)
	deleteTicketChecklistItem(c, uint(ticketID), &cid)
}

// DeleteInboxTicketChecklistItem DELETE /api/v1/tickets/inbox/:ticketId/checklist/:itemId
func DeleteInboxTicketChecklistItem(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	deleteTicketChecklistItem(c, uint(ticketID), nil)
}

func reorderTicketChecklistItems(c *gin.Context, ticketID uint, customerID *uint) {
	var ticket models.Ticket
	q := database.DB.Where("id = ?", ticketID)
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	} else {
		q = q.Where("customer_id IS NULL")
	}
	if err := q.First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	var req []struct {
		ID       uint    `json:"id"`
		Position float64 `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	for _, r := range req {
		database.DB.Model(&models.TicketChecklistItem{}).
			Where("id = ? AND ticket_id = ?", r.ID, ticketID).
			Update("position", r.Position)
	}
	var items []models.TicketChecklistItem
	database.DB.Where("ticket_id = ?", ticketID).Order("position asc, id asc").Find(&items)
	if items == nil {
		items = []models.TicketChecklistItem{}
	}
	c.JSON(http.StatusOK, items)
}

// ReorderTicketChecklistItems PATCH /api/v1/customers/:customerId/tickets/:ticketId/checklist/reorder
func ReorderTicketChecklistItems(c *gin.Context) {
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
	cid := uint(customerID)
	reorderTicketChecklistItems(c, uint(ticketID), &cid)
}

// ReorderInboxTicketChecklistItems PATCH /api/v1/tickets/inbox/:ticketId/checklist/reorder
func ReorderInboxTicketChecklistItems(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}
	reorderTicketChecklistItems(c, uint(ticketID), nil)
}
