package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

func ListReleases(c *gin.Context) {
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

	var releases []models.Release
	database.DB.Where("project_id = ?", project.ID).Order("created_at asc").Find(&releases)

	for i := range releases {
		var sprintIDs []uint
		database.DB.Model(&models.ReleaseSprint{}).
			Where("release_id = ?", releases[i].ID).
			Pluck("sprint_id", &sprintIDs)
		if len(sprintIDs) > 0 {
			database.DB.Where("id IN ? AND deleted_at IS NULL", sprintIDs).Find(&releases[i].Sprints)
			for j := range releases[i].Sprints {
				populateReleaseSprintPoints(&releases[i].Sprints[j])
			}
		} else {
			releases[i].Sprints = []models.Sprint{}
		}
	}

	c.JSON(http.StatusOK, releases)
}

func populateReleaseSprintPoints(s *models.Sprint) {
	var totals struct {
		Total     int
		Completed int
		Count     int
	}
	database.DB.Raw(`
		SELECT
			COALESCE(SUM(c.story_points), 0) AS total,
			COALESCE(SUM(CASE WHEN c.closed = true THEN c.story_points ELSE 0 END), 0) AS completed,
			COUNT(*) AS count
		FROM sprint_cards sc
		JOIN cards c ON c.id = sc.card_id
		WHERE sc.sprint_id = ? AND c.deleted_at IS NULL
	`, s.ID).Scan(&totals)
	s.TotalPoints = totals.Total
	s.CompletedPoints = totals.Completed
	s.CardCount = totals.Count
}

func CreateRelease(c *gin.Context) {
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
		Name       string          `json:"name" binding:"required"`
		Goal       string          `json:"goal"`
		TargetDate json.RawMessage `json:"target_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	release := models.Release{
		ProjectID:  project.ID,
		Name:       req.Name,
		Goal:       req.Goal,
		TargetDate: parseReleaseDate(req.TargetDate),
	}
	database.DB.Create(&release)
	release.Sprints = []models.Sprint{}
	c.JSON(http.StatusCreated, release)
}

func UpdateRelease(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	releaseID, err := strconv.ParseUint(c.Param("releaseId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid release id"})
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

	var release models.Release
	if err := database.DB.Where("id = ? AND project_id = ?", releaseID, project.ID).First(&release).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "release not found"})
		return
	}

	var req struct {
		Name       string          `json:"name"`
		Goal       string          `json:"goal"`
		TargetDate json.RawMessage `json:"target_date"`
	}
	c.ShouldBindJSON(&req)
	if req.Name != "" {
		release.Name = req.Name
	}
	release.Goal = req.Goal
	if len(req.TargetDate) > 0 {
		release.TargetDate = parseReleaseDate(req.TargetDate)
	}
	database.DB.Save(&release)
	c.JSON(http.StatusOK, release)
}

func DeleteRelease(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	releaseID, err := strconv.ParseUint(c.Param("releaseId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid release id"})
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

	var release models.Release
	if err := database.DB.Where("id = ? AND project_id = ?", releaseID, project.ID).First(&release).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "release not found"})
		return
	}
	database.DB.Where("release_id = ?", release.ID).Delete(&models.ReleaseSprint{})
	database.DB.Delete(&release)
	c.JSON(http.StatusNoContent, nil)
}

func AddSprintToRelease(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	releaseID, err := strconv.ParseUint(c.Param("releaseId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid release id"})
		return
	}
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

	var release models.Release
	if err := database.DB.Where("id = ? AND project_id = ?", releaseID, project.ID).First(&release).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "release not found"})
		return
	}
	var sprint models.Sprint
	if err := database.DB.Where("id = ? AND project_id = ?", sprintID, project.ID).First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprint not found"})
		return
	}

	rs := models.ReleaseSprint{ReleaseID: uint(releaseID), SprintID: uint(sprintID)}
	database.DB.FirstOrCreate(&rs, rs)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func RemoveSprintFromRelease(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	releaseID, err := strconv.ParseUint(c.Param("releaseId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid release id"})
		return
	}
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

	database.DB.Where("release_id = ? AND sprint_id = ?", releaseID, sprintID).Delete(&models.ReleaseSprint{})
	c.JSON(http.StatusNoContent, nil)
}

func parseReleaseDate(raw json.RawMessage) *time.Time {
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
