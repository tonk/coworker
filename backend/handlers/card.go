package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
	"github.com/tonk/warmdesk/ws"
	"gorm.io/gorm"
)

// ListCards godoc
// @Summary      List cards in a column
// @Tags         cards
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Param        columnId path int true "Column ID"
// @Success      200 {array}  models.Card
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /projects/{projectSlug}/columns/{columnId}/cards [get]
func ListCards(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	colID, err := strconv.ParseUint(c.Param("columnId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid column id"})
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

	var cards []models.Card
	database.DB.Preload("Labels").Preload("Assignee").Preload("Tags").
		Where("column_id = ? AND project_id = ? AND parent_card_id IS NULL", colID, project.ID).
		Order("position asc").Find(&cards)

	// Populate sub-card counts so the board can show progress pills
	for i := range cards {
		var total, done int64
		database.DB.Model(&models.Card{}).Where("parent_card_id = ?", cards[i].ID).Count(&total)
		if total > 0 {
			database.DB.Model(&models.Card{}).Where("parent_card_id = ? AND closed = true", cards[i].ID).Count(&done)
		}
		cards[i].SubCardCount = int(total)
		cards[i].SubCardsDone = int(done)
	}
	c.JSON(http.StatusOK, cards)
}

// CreateCard godoc
// @Summary      Create a new card in a column
// @Tags         cards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Param        columnId path int true "Column ID"
// @Param        body body map[string]interface{} true "Card details (title required)"
// @Success      201 {object} models.Card
// @Failure      400 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Router       /projects/{projectSlug}/columns/{columnId}/cards [post]
func CreateCard(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	colID, err := strconv.ParseUint(c.Param("columnId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid column id"})
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

	var req struct {
		Title       string          `json:"title" binding:"required"`
		Description string          `json:"description"`
		Priority    string          `json:"priority"`
		StartDate   json.RawMessage `json:"start_date"` // "YYYY-MM-DD" string or null
		DueDate     json.RawMessage `json:"due_date"`   // "YYYY-MM-DD" string or null
		AssigneeID  *uint           `json:"assignee_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var maxPos float64
	database.DB.Model(&models.Card{}).Where("column_id = ?", colID).Select("COALESCE(MAX(position), 0)").Scan(&maxPos)

	priority := req.Priority
	if priority == "" {
		priority = "none"
	}

	parseDate := func(raw json.RawMessage) *time.Time {
		if len(raw) == 0 || string(raw) == "null" {
			return nil
		}
		var s string
		if err := json.Unmarshal(raw, &s); err == nil && s != "" {
			if t, err := time.Parse("2006-01-02", s); err == nil {
				return &t
			}
		}
		return nil
	}

	// Atomically increment the project's card counter
	database.DB.Model(&models.Project{}).Where("id = ?", project.ID).
		UpdateColumn("card_counter", gorm.Expr("card_counter + 1"))
	var updatedProject models.Project
	database.DB.Select("card_counter").First(&updatedProject, project.ID)

	card := models.Card{
		ColumnID:    uint(colID),
		ProjectID:   project.ID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    priority,
		StartDate:   parseDate(req.StartDate),
		DueDate:     parseDate(req.DueDate),
		AssigneeID:  req.AssigneeID,
		CreatedByID: userID,
		Position:    maxPos + 1000,
		CardNumber:  updatedProject.CardCounter,
	}
	database.DB.Create(&card)
	database.DB.Preload("Labels").Preload("Assignee").Preload("Tags").First(&card, card.ID)

	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardCardCreated, Payload: card})
	c.JSON(http.StatusCreated, card)
}

// GetCard godoc
// @Summary      Get a single card with full details
// @Tags         cards
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Param        cardId path int true "Card ID"
// @Success      200 {object} models.Card
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /projects/{projectSlug}/cards/{cardId} [get]
func GetCard(c *gin.Context) {
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

	var card models.Card
	if err := database.DB.Preload("Labels").Preload("Assignee").Preload("Assignees").Preload("Watchers").Preload("Comments.User").Preload("Tags").Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}
	if am := LoadAttachments("card", []uint{card.ID}); len(am[card.ID]) > 0 {
		card.Attachments = am[card.ID]
	}
	// Populate sub-card counts
	var subTotal, subDone int64
	database.DB.Model(&models.Card{}).Where("parent_card_id = ?", card.ID).Count(&subTotal)
	if subTotal > 0 {
		database.DB.Model(&models.Card{}).Where("parent_card_id = ? AND closed = true", card.ID).Count(&subDone)
	}
	card.SubCardCount = int(subTotal)
	card.SubCardsDone = int(subDone)
	c.JSON(http.StatusOK, card)
}

// UpdateCard godoc
// @Summary      Update a card's fields
// @Tags         cards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Param        cardId path int true "Card ID"
// @Param        body body map[string]interface{} true "Fields to update"
// @Success      200 {object} models.Card
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /projects/{projectSlug}/cards/{cardId} [put]
func UpdateCard(c *gin.Context) {
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
		Title            string          `json:"title"`
		Description      string          `json:"description"`
		Priority         string          `json:"priority"`
		StartDate        json.RawMessage `json:"start_date"` // "YYYY-MM-DD" string or null
		DueDate          json.RawMessage `json:"due_date"`   // "YYYY-MM-DD" string or null
		AssigneeID       *uint           `json:"assignee_id"`
		TimeSpentMinutes *int            `json:"time_spent_minutes"`
		StoryPoints      *int            `json:"story_points"`
		Closed           *bool           `json:"closed"`
	}
	c.ShouldBindJSON(&req)

	parseDate := func(raw json.RawMessage) (interface{}, bool) {
		if len(raw) == 0 {
			return nil, false
		}
		if string(raw) == "null" {
			return nil, true
		}
		var s string
		if err := json.Unmarshal(raw, &s); err == nil && s != "" {
			if t, err := time.Parse("2006-01-02", s); err == nil {
				return t, true
			}
		}
		return nil, false
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Priority != "" {
		updates["priority"] = req.Priority
	}
	if v, ok := parseDate(req.StartDate); ok {
		updates["start_date"] = v
	}
	if v, ok := parseDate(req.DueDate); ok {
		updates["due_date"] = v
	}
	if req.AssigneeID != nil {
		updates["assignee_id"] = req.AssigneeID
	}
	if req.TimeSpentMinutes != nil {
		updates["time_spent_minutes"] = *req.TimeSpentMinutes
	}
	if req.StoryPoints != nil {
		updates["story_points"] = req.StoryPoints
	}
	if req.Closed != nil {
		updates["closed"] = *req.Closed
	}

	database.DB.Model(&card).Updates(updates)
	database.DB.Preload("Labels").Preload("Assignee").Preload("Tags").First(&card, card.ID)

	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardCardUpdated, Payload: card})
	c.JSON(http.StatusOK, card)
}

// DeleteCard godoc
// @Summary      Delete a card
// @Tags         cards
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Param        cardId path int true "Card ID"
// @Success      204
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /projects/{projectSlug}/cards/{cardId} [delete]
func DeleteCard(c *gin.Context) {
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

	database.DB.Delete(&card)
	ws.BroadcastToProject(project.ID, ws.Message{
		Type:    ws.TypeBoardCardDeleted,
		Payload: map[string]uint{"card_id": uint(cardID), "column_id": card.ColumnID},
	})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// MoveCard godoc
// @Summary      Move a card to a different column or position
// @Tags         cards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Param        cardId path int true "Card ID"
// @Param        body body map[string]interface{} true "column_id and position"
// @Success      200 {object} models.Card
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /projects/{projectSlug}/cards/{cardId}/move [patch]
func MoveCard(c *gin.Context) {
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

	var req struct {
		ColumnID uint    `json:"column_id" binding:"required"`
		Position float64 `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var card models.Card
	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	oldColumnID := card.ColumnID
	database.DB.Model(&card).Updates(map[string]interface{}{
		"column_id": req.ColumnID,
		"position":  req.Position,
	})

	if oldColumnID != req.ColumnID {
		database.DB.Create(&models.CardHistory{
			CardID:       card.ID,
			UserID:       userID,
			FromColumnID: oldColumnID,
			ToColumnID:   req.ColumnID,
		})
	}

	ws.BroadcastToProject(project.ID, ws.Message{
		Type: ws.TypeBoardCardMoved,
		Payload: map[string]interface{}{
			"card_id":        card.ID,
			"from_column_id": oldColumnID,
			"to_column_id":   req.ColumnID,
			"position":       req.Position,
		},
	})
	c.JSON(http.StatusOK, card)
}

func GetCardHistory(c *gin.Context) {
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

	var history []models.CardHistory
	database.DB.Preload("User").Preload("FromColumn").Preload("ToColumn").
		Where("card_id = ?", cardID).
		Order("created_at desc").
		Find(&history)
	c.JSON(http.StatusOK, history)
}

func ReorderCards(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	colID, err := strconv.ParseUint(c.Param("columnId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid column id"})
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

	var req []struct {
		ID       uint    `json:"id"`
		Position float64 `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, item := range req {
		database.DB.Model(&models.Card{}).Where("id = ? AND column_id = ?", item.ID, colID).Update("position", item.Position)
	}

	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

func AssignLabel(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, _ := strconv.ParseUint(c.Param("cardId"), 10, 64)
	labelID, _ := strconv.ParseUint(c.Param("labelId"), 10, 64)

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	cl := models.CardLabel{CardID: uint(cardID), LabelID: uint(labelID)}
	database.DB.FirstOrCreate(&cl, cl)
	c.JSON(http.StatusOK, gin.H{"message": "assigned"})
}

func RemoveLabel(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, _ := strconv.ParseUint(c.Param("cardId"), 10, 64)
	labelID, _ := strconv.ParseUint(c.Param("labelId"), 10, 64)

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	database.DB.Where("card_id = ? AND label_id = ?", cardID, labelID).Delete(&models.CardLabel{})
	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}

// CopyCard creates a duplicate of the card within the same project and column.
func CopyCard(c *gin.Context) {
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

	var original models.Card
	if err := database.DB.Preload("Labels").Preload("Tags").Where("id = ? AND project_id = ?", cardID, project.ID).First(&original).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	var maxPos float64
	database.DB.Model(&models.Card{}).Where("column_id = ?", original.ColumnID).Select("COALESCE(MAX(position), 0)").Scan(&maxPos)

	database.DB.Model(&models.Project{}).Where("id = ?", project.ID).
		UpdateColumn("card_counter", gorm.Expr("card_counter + 1"))
	var updatedProject models.Project
	database.DB.Select("card_counter").First(&updatedProject, project.ID)

	newCard := models.Card{
		ColumnID:         original.ColumnID,
		ProjectID:        original.ProjectID,
		Title:            original.Title + " (copy)",
		Description:      original.Description,
		Priority:         original.Priority,
		DueDate:          original.DueDate,
		AssigneeID:       original.AssigneeID,
		CreatedByID:      userID,
		Position:         maxPos + 1000,
		CardNumber:       updatedProject.CardCounter,
		TimeSpentMinutes: 0,
	}
	database.DB.Create(&newCard)

	for _, label := range original.Labels {
		database.DB.Create(&models.CardLabel{CardID: newCard.ID, LabelID: label.ID})
	}
	for _, tag := range original.Tags {
		database.DB.Create(&models.CardTag{CardID: newCard.ID, Name: tag.Name})
	}

	database.DB.Preload("Labels").Preload("Assignee").Preload("Tags").First(&newCard, newCard.ID)
	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeBoardCardCreated, Payload: newCard})
	c.JSON(http.StatusCreated, newCard)
}

// TransferCard copies or moves a card to a column in another project.
func TransferCard(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	cardID, err := strconv.ParseUint(c.Param("cardId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

	var req struct {
		TargetProjectSlug string `json:"target_project_slug" binding:"required"`
		ColumnID          uint   `json:"column_id" binding:"required"`
		Action            string `json:"action" binding:"required"` // "copy" or "move"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Action != "copy" && req.Action != "move" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action must be 'copy' or 'move'"})
		return
	}

	sourceProject, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "source project not found"})
		return
	}
	if err := services.RequireProjectRole(sourceProject.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	targetProject, err := services.GetProjectBySlug(req.TargetProjectSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target project not found"})
		return
	}
	if err := services.RequireProjectRole(targetProject.ID, userID, middleware.GetGlobalRole(c), "member"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden in target project"})
		return
	}

	// Verify target column belongs to target project
	var targetColumn models.Column
	if err := database.DB.Where("id = ? AND project_id = ?", req.ColumnID, targetProject.ID).First(&targetColumn).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "column not found in target project"})
		return
	}

	var original models.Card
	if err := database.DB.Preload("Tags").Where("id = ? AND project_id = ?", cardID, sourceProject.ID).First(&original).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	var maxPos float64
	database.DB.Model(&models.Card{}).Where("column_id = ?", req.ColumnID).Select("COALESCE(MAX(position), 0)").Scan(&maxPos)

	database.DB.Model(&models.Project{}).Where("id = ?", targetProject.ID).
		UpdateColumn("card_counter", gorm.Expr("card_counter + 1"))
	var updatedProject models.Project
	database.DB.Select("card_counter").First(&updatedProject, targetProject.ID)

	newCard := models.Card{
		ColumnID:    req.ColumnID,
		ProjectID:   targetProject.ID,
		Title:       original.Title,
		Description: original.Description,
		Priority:    original.Priority,
		DueDate:     original.DueDate,
		CreatedByID: userID,
		Position:    maxPos + 1000,
		CardNumber:  updatedProject.CardCounter,
	}
	database.DB.Create(&newCard)

	for _, tag := range original.Tags {
		database.DB.Create(&models.CardTag{CardID: newCard.ID, Name: tag.Name})
	}

	database.DB.Preload("Labels").Preload("Assignee").Preload("Tags").First(&newCard, newCard.ID)
	ws.BroadcastToProject(targetProject.ID, ws.Message{Type: ws.TypeBoardCardCreated, Payload: newCard})

	if req.Action == "move" {
		database.DB.Delete(&original)
		ws.BroadcastToProject(sourceProject.ID, ws.Message{
			Type:    ws.TypeBoardCardDeleted,
			Payload: map[string]uint{"card_id": original.ID, "column_id": original.ColumnID},
		})
	}

	c.JSON(http.StatusCreated, newCard)
}

func UpdateAssignee(c *gin.Context) {
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

	var req struct {
		UserID *uint `json:"user_id"`
	}
	c.ShouldBindJSON(&req)

	database.DB.Model(&models.Card{}).Where("id = ? AND project_id = ?", cardID, project.ID).Update("assignee_id", req.UserID)

	if notifSvc != nil && req.UserID != nil {
		var card models.Card
		var assignee, assigner models.User
		database.DB.First(&card, cardID)
		database.DB.First(&assignee, *req.UserID)
		database.DB.First(&assigner, userID)
		go notifSvc.NotifyCardAssignment(card, assignee, assigner)
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// ResolveCardRef GET /api/v1/cards/resolve/:ref
// Resolves a card reference like "PRJ-42" to its project slug and card ID.
func ResolveCardRef(c *gin.Context) {
	ref := c.Param("ref")
	sep := strings.LastIndex(ref, "-")
	if sep <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ref"})
		return
	}
	prefix := strings.ToUpper(ref[:sep])
	number, err := strconv.Atoi(ref[sep+1:])
	if err != nil || number < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ref"})
		return
	}

	var result struct {
		ID          uint   `json:"id"`
		CardNumber  int    `json:"card_number"`
		KeyPrefix   string `json:"key_prefix"`
		ProjectSlug string `json:"project_slug"`
		Title       string `json:"title"`
	}
	if err := database.DB.
		Table("cards").
		Select("cards.id, cards.card_number, projects.key_prefix, projects.slug as project_slug, cards.title").
		Joins("JOIN projects ON projects.id = cards.project_id").
		Where("projects.key_prefix = ? AND cards.card_number = ? AND cards.deleted_at IS NULL", prefix, number).
		Scan(&result).Error; err != nil || result.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}
