package handlers

import (
	"cmp"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

var keyPrefixRe = regexp.MustCompile(`^[A-Z0-9]{1,10}$`)

// labelPalette is cycled through when assigning colors to default labels.
var labelPalette = []string{
	"#ef4444", // red
	"#3b82f6", // blue
	"#8b5cf6", // purple
	"#10b981", // green
	"#f59e0b", // amber
	"#06b6d4", // cyan
	"#ec4899", // pink
	"#84cc16", // lime
}

type labelDef struct {
	Name  string
	Color string
}

// getDefaultLabelDefs reads the configured initial label names from system settings
// and pairs them with colors from the palette.
func getDefaultLabelDefs() []labelDef {
	all := loadAllSettings()
	raw := all[settingDefaultLabels]
	var defs []labelDef
	for i, line := range strings.Split(raw, "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		color := labelPalette[i%len(labelPalette)]
		defs = append(defs, labelDef{Name: name, Color: color})
	}
	return defs
}

// getDefaultColumnNames reads the configured initial column names from system settings.
func getDefaultColumnNames() []string {
	all := loadAllSettings()
	raw := all[settingDefaultColumns]
	var names []string
	for _, line := range strings.Split(raw, "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return []string{"Backlog"}
	}
	return names
}

type ProjectListItem struct {
	models.Project
	OpenCardCount int64 `json:"open_card_count"`
}

func projectsWithCounts(projects []models.Project) []ProjectListItem {
	result := make([]ProjectListItem, len(projects))
	for i, p := range projects {
		var count int64
		database.DB.Model(&models.Card{}).Where("project_id = ? AND closed = false", p.ID).Count(&count)
		result[i] = ProjectListItem{Project: p, OpenCardCount: count}
	}
	return result
}

// ListProjects godoc
// @Summary      List projects accessible to the current user
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Param        include_closed query string false "Set to 'true' to include closed projects"
// @Success      200 {array}  models.Project
// @Router       /projects [get]
func ListProjects(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)
	includeClosed := c.Query("include_closed") == "true"

	// Admins and global viewers see all non-deleted, non-time-tracking-only projects
	if globalRole == "admin" || globalRole == "viewer" {
		var projects []models.Project
		q := database.DB.Preload("Customer").Preload("Contract").Where("deleted_at IS NULL AND time_tracking_only = false")
		if !includeClosed {
			q = q.Where("is_closed = false")
		}
		q.Order("position asc, id asc").Find(&projects)
		c.JSON(http.StatusOK, projectsWithCounts(projects))
		return
	}

	var members []models.ProjectMember
	database.DB.Preload("Project.Customer").Preload("Project.Contract").Where("user_id = ?", userID).Find(&members)

	seen := make(map[uint]struct{})
	projects := make([]models.Project, 0, len(members))
	for _, m := range members {
		if m.Project.DeletedAt.Valid || m.Project.TimeTrackingOnly {
			continue
		}
		if !includeClosed && m.Project.IsClosed {
			continue
		}
		seen[m.Project.ID] = struct{}{}
		projects = append(projects, m.Project)
	}

	// Also include projects accessible via group membership
	var gpRows []struct{ ProjectID uint }
	database.DB.Raw(`
		SELECT DISTINCT gpa.project_id
		FROM group_project_accesses gpa
		JOIN group_members gm ON gm.group_id = gpa.group_id
		WHERE gm.user_id = ?`, userID).Scan(&gpRows)
	if len(gpRows) > 0 {
		extra := make([]uint, 0, len(gpRows))
		for _, r := range gpRows {
			if _, ok := seen[r.ProjectID]; !ok {
				extra = append(extra, r.ProjectID)
			}
		}
		if len(extra) > 0 {
			var gProjects []models.Project
			q := database.DB.Preload("Customer").Preload("Contract").
				Where("id IN ? AND deleted_at IS NULL", extra)
			if !includeClosed {
				q = q.Where("is_closed = false")
			}
			q.Find(&gProjects)
			projects = append(projects, gProjects...)
		}
	}

	slices.SortFunc(projects, func(a, b models.Project) int {
		if n := cmp.Compare(a.Position, b.Position); n != 0 {
			return n
		}
		return cmp.Compare(a.ID, b.ID)
	})
	c.JSON(http.StatusOK, projectsWithCounts(projects))
}

// CreateProject godoc
// @Summary      Create a new project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body map[string]string true "Project name, description, color"
// @Success      201 {object} models.Project
// @Failure      400 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Router       /projects [post]
func CreateProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if middleware.GetGlobalRole(c) == "viewer" {
		c.JSON(http.StatusForbidden, gin.H{"error": "viewers cannot create projects"})
		return
	}
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=200"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Avatar      string `json:"avatar"`
		KeyPrefix   string `json:"key_prefix"`
		BoardType   string `json:"board_type"`
		CustomerID  *uint  `json:"customer_id"`
		ContractID  *uint  `json:"contract_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.CustomerID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer is required"})
		return
	}

	keyPrefix := strings.ToUpper(strings.TrimSpace(req.KeyPrefix))
	if keyPrefix != "" {
		if !keyPrefixRe.MatchString(keyPrefix) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "card prefix must be 1–10 uppercase letters or digits"})
			return
		}
		var count int64
		database.DB.Model(&models.Project{}).Where("key_prefix = ?", keyPrefix).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "card prefix is already used by another project"})
			return
		}
	} else {
		keyPrefix = services.GenerateKeyPrefix(req.Name)
	}

	boardType := req.BoardType
	if boardType != "kanban" && boardType != "scrum" {
		boardType = "kanban"
	}

	project := models.Project{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Avatar:      req.Avatar,
		Slug:        services.GenerateSlug(req.Name),
		KeyPrefix:   keyPrefix,
		BoardType:   boardType,
		CreatedByID: userID,
		CustomerID:  req.CustomerID,
		ContractID:  req.ContractID,
	}

	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Creator becomes owner
	member := models.ProjectMember{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "owner",
		InvitedBy: userID,
	}
	database.DB.Create(&member)

	// Default columns from system settings
	for i, name := range getDefaultColumnNames() {
		database.DB.Create(&models.Column{ProjectID: project.ID, Name: name, Position: float64((i + 1) * 1000)})
	}

	// Default labels from system settings
	for _, def := range getDefaultLabelDefs() {
		database.DB.Create(&models.Label{ProjectID: project.ID, Name: def.Name, Color: def.Color})
	}

	c.JSON(http.StatusCreated, project)
}

// GetProject godoc
// @Summary      Get a project with its columns, cards, labels and members
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Success      200 {object} models.Project
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /projects/{projectSlug} [get]
func GetProject(c *gin.Context) {
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

	database.DB.
		Preload("Columns").
		Preload("Columns.Cards", "parent_card_id IS NULL").
		Preload("Columns.Cards.Assignee").
		Preload("Columns.Cards.Labels").
		Preload("Columns.Cards.Tags").
		Preload("Columns.Cards.Epic").
		Preload("Labels").
		Preload("Members.User").
		Preload("Customer").
		Preload("Contract").
		First(project, project.ID)

	// Populate sub-card counts for all board cards in one query
	var allCardIDs []uint
	for _, col := range project.Columns {
		for _, card := range col.Cards {
			allCardIDs = append(allCardIDs, card.ID)
		}
	}
	if len(allCardIDs) > 0 {
		type subStat struct {
			ParentCardID uint
			Total        int64
			Done         int64
		}
		var stats []subStat
		database.DB.Model(&models.Card{}).
			Select("parent_card_id, count(*) as total, sum(case when closed = true then 1 else 0 end) as done").
			Where("parent_card_id IN ?", allCardIDs).
			Group("parent_card_id").
			Scan(&stats)
		statMap := make(map[uint]subStat, len(stats))
		for _, s := range stats {
			statMap[s.ParentCardID] = s
		}
		for ci := range project.Columns {
			for i := range project.Columns[ci].Cards {
				if s, ok := statMap[project.Columns[ci].Cards[i].ID]; ok {
					project.Columns[ci].Cards[i].SubCardCount = int(s.Total)
					project.Columns[ci].Cards[i].SubCardsDone = int(s.Done)
				}
			}
		}
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject godoc
// @Summary      Update a project (owner/admin only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Param        body body map[string]interface{} true "Fields to update"
// @Success      200 {object} models.Project
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /projects/{projectSlug} [put]
func UpdateProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "owner"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Color       string  `json:"color"`
		Avatar      *string `json:"avatar"`
		IsArchived  *bool   `json:"is_archived"`
		IsClosed    *bool   `json:"is_closed"`
		BoardType   string  `json:"board_type"`
		CustomerID  *uint   `json:"customer_id"`
		ContractID  *uint   `json:"contract_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Color != "" {
		updates["color"] = req.Color
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.IsArchived != nil {
		updates["is_archived"] = *req.IsArchived
	}
	if req.IsClosed != nil {
		updates["is_closed"] = *req.IsClosed
	}
	// board_type is immutable after project creation
	if req.CustomerID != nil {
		updates["customer_id"] = req.CustomerID
	}
	if req.ContractID != nil {
		updates["contract_id"] = req.ContractID
	}

	database.DB.Model(project).Updates(updates)
	database.DB.First(project, project.ID)
	c.JSON(http.StatusOK, project)
}

// DeleteProject godoc
// @Summary      Delete a project (owner/admin only)
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Param        projectSlug path string true "Project slug"
// @Success      204
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /projects/{projectSlug} [delete]
func DeleteProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")

	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	globalRole := middleware.GetGlobalRole(c)
	if globalRole != "admin" {
		if err := services.RequireProjectRole(project.ID, userID, globalRole, "owner"); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	database.DB.Delete(project)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ReorderProjects PATCH /api/v1/projects/reorder — update display order (admin only).
func ReorderProjects(c *gin.Context) {
	if middleware.GetGlobalRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}

	var req []struct {
		ID       uint `json:"id"`
		Position int  `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	for _, item := range req {
		database.DB.Model(&models.Project{}).Where("id = ?", item.ID).Update("position", item.Position)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
