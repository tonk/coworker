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
	"github.com/tonk/warmdesk/ws"
)

// populateSprint fills computed fields (card_ids, counts, story points) on a sprint.
func populateSprint(s *models.Sprint) {
	var sprintCards []models.SprintCard
	database.DB.Where("sprint_id = ?", s.ID).Order("position asc").Find(&sprintCards)

	s.CardIDs = make([]uint, len(sprintCards))
	cardIDs := make([]uint, len(sprintCards))
	for i, sc := range sprintCards {
		s.CardIDs[i] = sc.CardID
		cardIDs[i] = sc.CardID
	}
	s.CardCount = len(cardIDs)

	if len(cardIDs) == 0 {
		return
	}

	type spStat struct {
		Total     int
		Completed int
	}
	var stat spStat
	database.DB.Model(&models.Card{}).
		Select("COALESCE(SUM(story_points), 0) as total, COALESCE(SUM(CASE WHEN closed = true AND story_points IS NOT NULL THEN story_points ELSE 0 END), 0) as completed").
		Where("id IN ?", cardIDs).
		Scan(&stat)
	s.TotalPoints = stat.Total
	s.CompletedPoints = stat.Completed
}

// ListSprints GET /projects/:projectSlug/sprints
func ListSprints(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "viewer"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var sprints []models.Sprint
	database.DB.Where("project_id = ?", project.ID).Order("created_at asc").Find(&sprints)
	for i := range sprints {
		populateSprint(&sprints[i])
	}
	c.JSON(http.StatusOK, sprints)
}

// CreateSprint POST /projects/:projectSlug/sprints
func CreateSprint(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

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
		Name      string     `json:"name" binding:"required"`
		Goal      string     `json:"goal"`
		StartDate *time.Time `json:"start_date"`
		EndDate   *time.Time `json:"end_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	sprint := models.Sprint{
		ProjectID: project.ID,
		Name:      req.Name,
		Goal:      req.Goal,
		Status:    "planning",
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}
	database.DB.Create(&sprint)
	populateSprint(&sprint)
	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeSprintCreated, Payload: sprint})
	c.JSON(http.StatusCreated, sprint)
}

// UpdateSprint PUT /projects/:projectSlug/sprints/:sprintId
func UpdateSprint(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	sprintID, err := strconv.ParseUint(c.Param("sprintId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sprint id"})
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

	var sprint models.Sprint
	if err := database.DB.Where("id = ? AND project_id = ?", sprintID, project.ID).First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprint not found"})
		return
	}

	var req struct {
		Name      string     `json:"name"`
		Goal      string     `json:"goal"`
		StartDate *time.Time `json:"start_date"`
		EndDate   *time.Time `json:"end_date"`
	}
	c.ShouldBindJSON(&req)

	updates := map[string]interface{}{
		"goal":       req.Goal,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}

	database.DB.Model(&sprint).Updates(updates)
	database.DB.First(&sprint, sprint.ID)
	populateSprint(&sprint)
	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeSprintUpdated, Payload: sprint})
	c.JSON(http.StatusOK, sprint)
}

// DeleteSprint DELETE /projects/:projectSlug/sprints/:sprintId
func DeleteSprint(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	sprintID, err := strconv.ParseUint(c.Param("sprintId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sprint id"})
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

	var sprint models.Sprint
	if err := database.DB.Where("id = ? AND project_id = ?", sprintID, project.ID).First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprint not found"})
		return
	}
	if sprint.Status == "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete an active sprint; complete it first"})
		return
	}

	database.DB.Where("sprint_id = ?", sprint.ID).Delete(&models.SprintCard{})
	database.DB.Delete(&sprint)
	ws.BroadcastToProject(project.ID, ws.Message{
		Type:    ws.TypeSprintDeleted,
		Payload: gin.H{"sprint_id": sprint.ID},
	})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// StartSprint POST /projects/:projectSlug/sprints/:sprintId/start
func StartSprint(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	sprintID, err := strconv.ParseUint(c.Param("sprintId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sprint id"})
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

	var activeCount int64
	database.DB.Model(&models.Sprint{}).Where("project_id = ? AND status = 'active'", project.ID).Count(&activeCount)
	if activeCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "another sprint is already active"})
		return
	}

	var sprint models.Sprint
	if err := database.DB.Where("id = ? AND project_id = ?", sprintID, project.ID).First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprint not found"})
		return
	}
	if sprint.Status != "planning" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only planning sprints can be started"})
		return
	}

	database.DB.Model(&sprint).Update("status", "active")
	database.DB.First(&sprint, sprint.ID)
	populateSprint(&sprint)
	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeSprintStarted, Payload: sprint})
	c.JSON(http.StatusOK, sprint)
}

// CompleteSprint POST /projects/:projectSlug/sprints/:sprintId/complete
// Marks the sprint completed; unfinished (not closed) cards return to the backlog.
func CompleteSprint(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	sprintID, err := strconv.ParseUint(c.Param("sprintId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sprint id"})
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

	var sprint models.Sprint
	if err := database.DB.Where("id = ? AND project_id = ?", sprintID, project.ID).First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprint not found"})
		return
	}
	if sprint.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only active sprints can be completed"})
		return
	}

	// Find card IDs in this sprint that are not closed — remove them (back to backlog)
	var unclosedIDs []uint
	database.DB.Table("sprint_cards").
		Joins("JOIN cards ON cards.id = sprint_cards.card_id").
		Where("sprint_cards.sprint_id = ? AND cards.closed = false AND cards.deleted_at IS NULL", sprintID).
		Pluck("sprint_cards.card_id", &unclosedIDs)

	if len(unclosedIDs) > 0 {
		database.DB.Where("sprint_id = ? AND card_id IN ?", sprintID, unclosedIDs).Delete(&models.SprintCard{})
	}

	database.DB.Model(&sprint).Update("status", "completed")
	database.DB.First(&sprint, sprint.ID)
	populateSprint(&sprint)
	ws.BroadcastToProject(project.ID, ws.Message{Type: ws.TypeSprintCompleted, Payload: sprint})
	c.JSON(http.StatusOK, sprint)
}

// ListBacklog GET /projects/:projectSlug/backlog
// Returns open cards not assigned to any planning or active sprint.
func ListBacklog(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "viewer"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Collect card IDs that are in any planning or active sprint for this project
	var sprintedIDs []uint
	database.DB.Table("sprint_cards").
		Joins("JOIN sprints ON sprints.id = sprint_cards.sprint_id").
		Where("sprints.project_id = ? AND sprints.status IN ? AND sprints.deleted_at IS NULL", project.ID, []string{"planning", "active"}).
		Pluck("sprint_cards.card_id", &sprintedIDs)

	q := database.DB.Preload("Labels").Preload("Assignee").Preload("Tags").
		Where("project_id = ? AND parent_card_id IS NULL AND closed = false", project.ID)
	if len(sprintedIDs) > 0 {
		q = q.Not("id IN ?", sprintedIDs)
	}

	var cards []models.Card
	q.Order("position asc").Find(&cards)

	// Populate sub-card counts
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

// AddCardToSprint POST /projects/:projectSlug/sprints/:sprintId/cards/:cardId
func AddCardToSprint(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	sprintID, err := strconv.ParseUint(c.Param("sprintId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sprint id"})
		return
	}
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

	var sprint models.Sprint
	if err := database.DB.Where("id = ? AND project_id = ?", sprintID, project.ID).First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprint not found"})
		return
	}
	if sprint.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot add cards to a completed sprint"})
		return
	}

	if err := database.DB.Where("id = ? AND project_id = ?", cardID, project.ID).First(&models.Card{}).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	// Remove the card from any other non-completed sprint in this project first
	database.DB.Exec(`DELETE FROM sprint_cards WHERE card_id = ? AND sprint_id IN (
		SELECT id FROM sprints WHERE project_id = ? AND status != 'completed' AND deleted_at IS NULL
	)`, cardID, project.ID)

	var maxPos float64
	database.DB.Table("sprint_cards").Where("sprint_id = ?", sprintID).Select("COALESCE(MAX(position), 0)").Scan(&maxPos)

	sc := models.SprintCard{SprintID: uint(sprintID), CardID: uint(cardID), Position: maxPos + 1000}
	database.DB.Create(&sc)

	ws.BroadcastToProject(project.ID, ws.Message{
		Type:    ws.TypeSprintCardAdded,
		Payload: gin.H{"sprint_id": sprint.ID, "card_id": cardID},
	})
	c.JSON(http.StatusOK, gin.H{"message": "added"})
}

// RemoveCardFromSprint DELETE /projects/:projectSlug/sprints/:sprintId/cards/:cardId
func RemoveCardFromSprint(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	sprintID, err := strconv.ParseUint(c.Param("sprintId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sprint id"})
		return
	}
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

	var sprint models.Sprint
	if err := database.DB.Where("id = ? AND project_id = ?", sprintID, project.ID).First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprint not found"})
		return
	}

	database.DB.Where("sprint_id = ? AND card_id = ?", sprintID, cardID).Delete(&models.SprintCard{})

	ws.BroadcastToProject(project.ID, ws.Message{
		Type:    ws.TypeSprintCardRemoved,
		Payload: gin.H{"sprint_id": sprint.ID, "card_id": cardID},
	})
	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}

// ReorderSprintCards PATCH /projects/:projectSlug/sprints/:sprintId/cards/reorder
func ReorderSprintCards(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	sprintID, err := strconv.ParseUint(c.Param("sprintId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sprint id"})
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
		CardID   uint    `json:"card_id"`
		Position float64 `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	for _, item := range req {
		database.DB.Model(&models.SprintCard{}).
			Where("sprint_id = ? AND card_id = ?", sprintID, item.CardID).
			Update("position", item.Position)
	}

	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
