package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
)

type velocitySprintData struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	Status          string     `json:"status"`
	TotalPoints     int        `json:"total_points"`
	CompletedPoints int        `json:"completed_points"`
	TotalCards      int        `json:"total_cards"`
	CompletedCards  int        `json:"completed_cards"`
}

type burndownDay struct {
	Date      string  `json:"date"`
	Remaining float64 `json:"remaining"`
	Ideal     float64 `json:"ideal"`
}

type burnupDay struct {
	Date      string  `json:"date"`
	Completed float64 `json:"completed"`
	Total     float64 `json:"total"`
}

type sprintCardRow struct {
	StoryPoints int
	ClosedAt    *time.Time
}

type cfdColumnSeries struct {
	ColumnID uint   `json:"column_id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Data     []int  `json:"data"`
}

type cycleTimeCard struct {
	ID        uint      `json:"id"`
	CardRef   string    `json:"card_ref"`
	Title     string    `json:"title"`
	ClosedAt  time.Time `json:"closed_at"`
	CreatedAt time.Time `json:"created_at"`
	DaysOpen  float64   `json:"days_open"`
}

func requireScrumProject(c *gin.Context) (*models.Project, bool) {
	project, ok := requireProjectAccess(c)
	if !ok {
		return nil, false
	}
	if project.BoardType != "scrum" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not a scrum project"})
		return nil, false
	}
	return project, true
}

func requireProjectAccess(c *gin.Context) (*models.Project, bool) {
	userID := middleware.GetUserID(c)
	slug := c.Param("projectSlug")
	project, err := services.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return nil, false
	}
	if err := services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "viewer"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return nil, false
	}
	return project, true
}

// GetVelocityChart returns story point and card velocity for all sprints in a project.
func GetVelocityChart(c *gin.Context) {
	project, ok := requireScrumProject(c)
	if !ok {
		return
	}

	var sprints []models.Sprint
	database.DB.Where("project_id = ? AND deleted_at IS NULL", project.ID).Order("created_at asc").Find(&sprints)

	result := make([]velocitySprintData, 0, len(sprints))
	for _, s := range sprints {
		var totals struct {
			Total          int
			Completed      int
			TotalCards     int
			CompletedCards int
		}
		database.DB.Raw(`
			SELECT
				COALESCE(SUM(c.story_points), 0) AS total,
				COALESCE(SUM(CASE WHEN c.closed = true THEN c.story_points ELSE 0 END), 0) AS completed,
				COUNT(*) AS total_cards,
				COALESCE(SUM(CASE WHEN c.closed = true THEN 1 ELSE 0 END), 0) AS completed_cards
			FROM sprint_cards sc
			JOIN cards c ON c.id = sc.card_id
			WHERE sc.sprint_id = ? AND c.deleted_at IS NULL
		`, s.ID).Scan(&totals)

		result = append(result, velocitySprintData{
			ID:              s.ID,
			Name:            s.Name,
			StartDate:       s.StartDate,
			EndDate:         s.EndDate,
			Status:          s.Status,
			TotalPoints:     totals.Total,
			CompletedPoints: totals.Completed,
			TotalCards:      totals.TotalCards,
			CompletedCards:  totals.CompletedCards,
		})
	}

	c.JSON(http.StatusOK, gin.H{"sprints": result})
}

// GetBurndownChart returns daily remaining story points for a sprint.
func GetBurndownChart(c *gin.Context) {
	project, ok := requireScrumProject(c)
	if !ok {
		return
	}

	sprintID, err := strconv.ParseUint(c.Param("sprintId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sprint id"})
		return
	}

	sprint, cards, totalPoints, ok2 := loadSprintCards(c, project.ID, sprintID)
	if !ok2 {
		return
	}

	start := sprint.StartDate.Truncate(24 * time.Hour)
	end := sprint.EndDate.Truncate(24 * time.Hour)
	now := time.Now().UTC().Truncate(24 * time.Hour)
	if now.Before(end) {
		end = now
	}

	totalDays := sprint.EndDate.Sub(*sprint.StartDate).Hours() / 24

	var data []burndownDay
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		remaining := totalPoints
		for _, cr := range cards {
			if cr.ClosedAt != nil && !cr.ClosedAt.UTC().Truncate(24*time.Hour).After(d) {
				remaining -= cr.StoryPoints
			}
		}
		dayIndex := d.Sub(start).Hours() / 24
		ideal := float64(totalPoints) * (1 - dayIndex/totalDays)
		if ideal < 0 {
			ideal = 0
		}
		data = append(data, burndownDay{
			Date:      d.Format("2006-01-02"),
			Remaining: float64(remaining),
			Ideal:     math.Round(ideal*10) / 10,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"sprint": sprintSummary(sprint, totalPoints),
		"data":   data,
	})
}

// GetBurnupChart returns daily completed story points for a sprint.
func GetBurnupChart(c *gin.Context) {
	project, ok := requireScrumProject(c)
	if !ok {
		return
	}

	sprintID, err := strconv.ParseUint(c.Param("sprintId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sprint id"})
		return
	}

	sprint, cards, totalPoints, ok2 := loadSprintCards(c, project.ID, sprintID)
	if !ok2 {
		return
	}

	start := sprint.StartDate.Truncate(24 * time.Hour)
	end := sprint.EndDate.Truncate(24 * time.Hour)
	now := time.Now().UTC().Truncate(24 * time.Hour)
	if now.Before(end) {
		end = now
	}

	var data []burnupDay
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		completed := 0
		for _, cr := range cards {
			if cr.ClosedAt != nil && !cr.ClosedAt.UTC().Truncate(24*time.Hour).After(d) {
				completed += cr.StoryPoints
			}
		}
		data = append(data, burnupDay{
			Date:      d.Format("2006-01-02"),
			Completed: float64(completed),
			Total:     float64(totalPoints),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"sprint": sprintSummary(sprint, totalPoints),
		"data":   data,
	})
}

// GetCFDChart returns daily card counts per column (Cumulative Flow Diagram).
func GetCFDChart(c *gin.Context) {
	project, ok := requireProjectAccess(c)
	if !ok {
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))
	if days <= 0 || days > 365 {
		days = 90
	}

	var columns []models.Column
	database.DB.Where("project_id = ? AND deleted_at IS NULL", project.ID).
		Order("position asc").Find(&columns)
	if len(columns) == 0 {
		c.JSON(http.StatusOK, gin.H{"labels": []string{}, "series": []interface{}{}})
		return
	}

	type cardBasic struct {
		ID        uint
		ColumnID  uint
		CreatedAt time.Time
	}
	var cards []cardBasic
	database.DB.Model(&models.Card{}).
		Select("id, column_id, created_at").
		Where("project_id = ? AND deleted_at IS NULL AND parent_card_id IS NULL", project.ID).
		Find(&cards)
	if len(cards) == 0 {
		c.JSON(http.StatusOK, gin.H{"labels": []string{}, "series": []interface{}{}})
		return
	}

	cardIDs := make([]uint, len(cards))
	for i, card := range cards {
		cardIDs[i] = card.ID
	}

	type histRow struct {
		CardID       uint
		FromColumnID uint
		ToColumnID   uint
		CreatedAt    time.Time
	}
	var history []histRow
	database.DB.Model(&models.CardHistory{}).
		Select("card_id, from_column_id, to_column_id, created_at").
		Where("card_id IN ?", cardIDs).
		Order("card_id asc, created_at asc").
		Find(&history)

	histByCard := make(map[uint][]histRow, len(cards))
	for _, h := range history {
		histByCard[h.CardID] = append(histByCard[h.CardID], h)
	}

	type cardInfo struct {
		InitialColumn uint
		CreatedAt     time.Time
	}
	cardInfoMap := make(map[uint]cardInfo, len(cards))
	for _, card := range cards {
		initialCol := card.ColumnID
		if hist := histByCard[card.ID]; len(hist) > 0 {
			initialCol = hist[0].FromColumnID
		}
		cardInfoMap[card.ID] = cardInfo{InitialColumn: initialCol, CreatedAt: card.CreatedAt}
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	start := now.AddDate(0, 0, -days+1)

	earliest := now
	for _, card := range cards {
		if d := card.CreatedAt.UTC().Truncate(24 * time.Hour); d.Before(earliest) {
			earliest = d
		}
	}
	if earliest.After(start) {
		start = earliest
	}

	colIndex := make(map[uint]int, len(columns))
	for i, col := range columns {
		colIndex[col.ID] = i
	}

	var labels []string
	var dateRange []time.Time
	for d := start; !d.After(now); d = d.AddDate(0, 0, 1) {
		labels = append(labels, d.Format("2006-01-02"))
		dateRange = append(dateRange, d)
	}
	if len(dateRange) == 0 {
		c.JSON(http.StatusOK, gin.H{"labels": []string{}, "series": []interface{}{}})
		return
	}

	counts := make([][]int, len(dateRange))
	for i := range counts {
		counts[i] = make([]int, len(columns))
	}

	for _, card := range cards {
		info := cardInfoMap[card.ID]
		hist := histByCard[card.ID]
		cardCreatedDay := info.CreatedAt.UTC().Truncate(24 * time.Hour)

		for dayIdx, day := range dateRange {
			if cardCreatedDay.After(day) {
				continue
			}
			colID := info.InitialColumn
			for _, h := range hist {
				if !h.CreatedAt.UTC().Truncate(24 * time.Hour).After(day) {
					colID = h.ToColumnID
				} else {
					break
				}
			}
			if idx, ok := colIndex[colID]; ok {
				counts[dayIdx][idx]++
			}
		}
	}

	series := make([]cfdColumnSeries, len(columns))
	for i, col := range columns {
		data := make([]int, len(dateRange))
		for dayIdx := range dateRange {
			data[dayIdx] = counts[dayIdx][i]
		}
		series[i] = cfdColumnSeries{
			ColumnID: col.ID,
			Name:     col.Name,
			Color:    col.Color,
			Data:     data,
		}
	}

	c.JSON(http.StatusOK, gin.H{"labels": labels, "series": series})
}

// GetCycleTimeChart returns cycle time data for all closed cards in the project.
func GetCycleTimeChart(c *gin.Context) {
	project, ok := requireProjectAccess(c)
	if !ok {
		return
	}

	type cardRow struct {
		ID         uint
		CardNumber int
		Title      string
		ClosedAt   *time.Time
		CreatedAt  time.Time
	}
	var rows []cardRow
	database.DB.Model(&models.Card{}).
		Select("id, card_number, title, closed_at, created_at").
		Where("project_id = ? AND deleted_at IS NULL AND closed = true AND closed_at IS NOT NULL AND parent_card_id IS NULL", project.ID).
		Order("closed_at asc").
		Find(&rows)

	result := make([]cycleTimeCard, 0, len(rows))
	for _, row := range rows {
		if row.ClosedAt == nil {
			continue
		}
		days := row.ClosedAt.Sub(row.CreatedAt).Hours() / 24
		if days < 0 {
			days = 0
		}
		result = append(result, cycleTimeCard{
			ID:        row.ID,
			CardRef:   fmt.Sprintf("%s-%d", project.KeyPrefix, row.CardNumber),
			Title:     row.Title,
			ClosedAt:  *row.ClosedAt,
			CreatedAt: row.CreatedAt,
			DaysOpen:  math.Round(days*10) / 10,
		})
	}

	c.JSON(http.StatusOK, gin.H{"cards": result})
}

// GetReleaseBurndownChart returns daily remaining story points across all sprints in a release.
func GetReleaseBurndownChart(c *gin.Context) {
	project, ok := requireScrumProject(c)
	if !ok {
		return
	}

	releaseID, err := strconv.ParseUint(c.Param("releaseId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid release id"})
		return
	}

	var release models.Release
	if err := database.DB.Where("id = ? AND project_id = ?", releaseID, project.ID).First(&release).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "release not found"})
		return
	}

	var sprintIDs []uint
	database.DB.Model(&models.ReleaseSprint{}).
		Where("release_id = ?", release.ID).
		Pluck("sprint_id", &sprintIDs)
	if len(sprintIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"release": gin.H{"id": release.ID, "name": release.Name}, "data": []interface{}{}})
		return
	}

	var sprints []models.Sprint
	database.DB.Where("id IN ? AND deleted_at IS NULL", sprintIDs).Order("start_date asc").Find(&sprints)

	var start, end *time.Time
	for _, s := range sprints {
		if s.StartDate == nil || s.EndDate == nil {
			continue
		}
		if start == nil || s.StartDate.Before(*start) {
			start = s.StartDate
		}
		if end == nil || s.EndDate.After(*end) {
			end = s.EndDate
		}
	}
	if start == nil || end == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sprints have no start/end dates"})
		return
	}

	var cards []sprintCardRow
	database.DB.Raw(`
		SELECT c.story_points, c.closed_at
		FROM sprint_cards sc
		JOIN cards c ON c.id = sc.card_id
		WHERE sc.sprint_id IN ? AND c.deleted_at IS NULL AND c.story_points > 0
	`, sprintIDs).Scan(&cards)

	totalPoints := 0
	for _, card := range cards {
		totalPoints += card.StoryPoints
	}

	startDay := start.UTC().Truncate(24 * time.Hour)
	endDay := end.UTC().Truncate(24 * time.Hour)
	now := time.Now().UTC().Truncate(24 * time.Hour)
	if now.Before(endDay) {
		endDay = now
	}
	totalDays := end.Sub(*start).Hours() / 24

	type releaseDay struct {
		Date      string  `json:"date"`
		Remaining float64 `json:"remaining"`
		Ideal     float64 `json:"ideal"`
	}
	var data []releaseDay
	for d := startDay; !d.After(endDay); d = d.AddDate(0, 0, 1) {
		remaining := totalPoints
		for _, card := range cards {
			if card.ClosedAt != nil && !card.ClosedAt.UTC().Truncate(24*time.Hour).After(d) {
				remaining -= card.StoryPoints
			}
		}
		dayIndex := d.Sub(startDay).Hours() / 24
		ideal := float64(totalPoints) * (1 - dayIndex/totalDays)
		if ideal < 0 {
			ideal = 0
		}
		data = append(data, releaseDay{
			Date:      d.Format("2006-01-02"),
			Remaining: float64(remaining),
			Ideal:     math.Round(ideal*10) / 10,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"release": gin.H{
			"id":           release.ID,
			"name":         release.Name,
			"target_date":  release.TargetDate,
			"total_points": totalPoints,
		},
		"data": data,
	})
}

type throughputWeek struct {
	WeekStart string `json:"week_start"`
	Count     int    `json:"count"`
}

// GetThroughputChart returns weekly closed-card counts for any project type.
func GetThroughputChart(c *gin.Context) {
	project, ok := requireProjectAccess(c)
	if !ok {
		return
	}

	weeks, _ := strconv.Atoi(c.DefaultQuery("weeks", "12"))
	if weeks <= 0 || weeks > 52 {
		weeks = 12
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -weeks*7)

	type closedRow struct {
		ClosedAt time.Time
	}
	var rows []closedRow
	database.DB.Model(&models.Card{}).
		Select("closed_at").
		Where("project_id = ? AND deleted_at IS NULL AND closed = true AND closed_at IS NOT NULL AND closed_at >= ? AND parent_card_id IS NULL", project.ID, cutoff).
		Order("closed_at asc").
		Find(&rows)

	// Build Monday-aligned week buckets
	now := time.Now().UTC()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	thisMonday := now.Truncate(24 * time.Hour).AddDate(0, 0, -(weekday - 1))

	result := make([]throughputWeek, weeks)
	for i := 0; i < weeks; i++ {
		weekStart := thisMonday.AddDate(0, 0, -(weeks-1-i)*7)
		result[i] = throughputWeek{WeekStart: weekStart.Format("2006-01-02"), Count: 0}
	}

	for _, row := range rows {
		t := row.ClosedAt.UTC()
		wd := int(t.Weekday())
		if wd == 0 {
			wd = 7
		}
		monday := t.Truncate(24 * time.Hour).AddDate(0, 0, -(wd - 1))
		key := monday.Format("2006-01-02")
		for i := range result {
			if result[i].WeekStart == key {
				result[i].Count++
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"weeks": result})
}

func loadSprintCards(c *gin.Context, projectID uint, sprintID uint64) (*models.Sprint, []sprintCardRow, int, bool) {
	var sprint models.Sprint
	if err := database.DB.Where("id = ? AND project_id = ?", sprintID, projectID).First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprint not found"})
		return nil, nil, 0, false
	}
	if sprint.StartDate == nil || sprint.EndDate == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sprint has no start/end date"})
		return nil, nil, 0, false
	}

	var rows []sprintCardRow
	database.DB.Raw(`
		SELECT c.story_points, c.closed_at
		FROM sprint_cards sc
		JOIN cards c ON c.id = sc.card_id
		WHERE sc.sprint_id = ? AND c.deleted_at IS NULL AND c.story_points > 0
	`, sprintID).Scan(&rows)

	total := 0
	for _, r := range rows {
		total += r.StoryPoints
	}
	return &sprint, rows, total, true
}

func sprintSummary(s *models.Sprint, totalPoints int) gin.H {
	return gin.H{
		"id":           s.ID,
		"name":         s.Name,
		"start_date":   s.StartDate,
		"end_date":     s.EndDate,
		"status":       s.Status,
		"total_points": totalPoints,
	}
}

// ── Sprint Report ─────────────────────────────────────────────────────────────

type sprintReportCard struct {
	ID          uint    `json:"id"`
	CardRef     string  `json:"card_ref"`
	Title       string  `json:"title"`
	StoryPoints *int    `json:"story_points"`
	Assignee    string  `json:"assignee"`
	ColumnName  string  `json:"column_name"`
	Priority    string  `json:"priority"`
}

// GetSprintReport GET /projects/:projectSlug/sprints/:sprintId/report
func GetSprintReport(c *gin.Context) {
	project, ok := requireScrumProject(c)
	if !ok {
		return
	}
	sprintID, err := strconv.ParseUint(c.Param("sprintId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sprint id"})
		return
	}
	var sprint models.Sprint
	if err := database.DB.Where("id = ? AND project_id = ?", sprintID, project.ID).First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprint not found"})
		return
	}

	var sprintCardIDs []uint
	database.DB.Model(&models.SprintCard{}).Where("sprint_id = ?", sprintID).Pluck("card_id", &sprintCardIDs)

	if len(sprintCardIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"sprint": sprint, "completed": []sprintReportCard{}, "incomplete": []sprintReportCard{},
			"summary": gin.H{"committed_count": 0, "completed_count": 0, "committed_points": 0, "completed_points": 0},
		})
		return
	}

	type cardRow struct {
		ID           uint
		CardNumber   int
		Title        string
		StoryPoints  *int
		Priority     string
		Closed       bool
		ColumnName   string
		AssigneeName string
	}
	var rows []cardRow
	database.DB.Model(&models.Card{}).
		Select("cards.id, cards.card_number, cards.title, cards.story_points, cards.priority, cards.closed, columns.name as column_name, COALESCE(NULLIF(users.display_name,''), users.username, '') as assignee_name").
		Joins("LEFT JOIN columns ON columns.id = cards.column_id").
		Joins("LEFT JOIN users ON users.id = cards.assignee_id").
		Where("cards.id IN ? AND cards.deleted_at IS NULL", sprintCardIDs).
		Find(&rows)

	completed := []sprintReportCard{}
	incomplete := []sprintReportCard{}
	committedPoints, completedPoints := 0, 0

	for _, row := range rows {
		card := sprintReportCard{
			ID:          row.ID,
			CardRef:     fmt.Sprintf("%s-%d", project.KeyPrefix, row.CardNumber),
			Title:       row.Title,
			StoryPoints: row.StoryPoints,
			Assignee:    row.AssigneeName,
			ColumnName:  row.ColumnName,
			Priority:    row.Priority,
		}
		if row.StoryPoints != nil {
			committedPoints += *row.StoryPoints
		}
		if row.Closed {
			completed = append(completed, card)
			if row.StoryPoints != nil {
				completedPoints += *row.StoryPoints
			}
		} else {
			incomplete = append(incomplete, card)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"sprint": sprint, "completed": completed, "incomplete": incomplete,
		"summary": gin.H{
			"committed_count":  len(rows),
			"completed_count":  len(completed),
			"committed_points": committedPoints,
			"completed_points": completedPoints,
		},
	})
}

// ── Epic Burndown ─────────────────────────────────────────────────────────────

// GetEpicBurndown GET /projects/:projectSlug/epics/:epicId/burndown
func GetEpicBurndown(c *gin.Context) {
	project, ok := requireProjectAccess(c)
	if !ok {
		return
	}
	epicID, err := strconv.ParseUint(c.Param("epicId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid epic id"})
		return
	}
	var epic models.Epic
	if err := database.DB.Where("id = ? AND project_id = ?", epicID, project.ID).First(&epic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "epic not found"})
		return
	}

	type cardData struct {
		StoryPoints *int
		ClosedAt    *time.Time
	}
	var cards []cardData
	database.DB.Model(&models.Card{}).
		Select("story_points, closed_at").
		Where("epic_id = ? AND parent_card_id IS NULL AND deleted_at IS NULL", epicID).
		Find(&cards)

	if len(cards) == 0 {
		c.JSON(http.StatusOK, gin.H{"epic": gin.H{"id": epic.ID, "name": epic.Name, "color": epic.Color}, "data": []interface{}{}})
		return
	}

	totalCards := len(cards)
	totalPoints := 0
	for _, card := range cards {
		if card.StoryPoints != nil {
			totalPoints += *card.StoryPoints
		}
	}

	startDay := epic.CreatedAt.UTC().Truncate(24 * time.Hour)
	endDay := time.Now().UTC().Truncate(24 * time.Hour)
	totalDays := endDay.Sub(startDay).Hours() / 24
	if totalDays < 1 {
		totalDays = 1
	}

	type epicBurnDay struct {
		Date            string  `json:"date"`
		RemainingCards  int     `json:"remaining_cards"`
		RemainingPoints int     `json:"remaining_points"`
		IdealRemaining  float64 `json:"ideal_remaining"`
	}

	var data []epicBurnDay
	for d := startDay; !d.After(endDay); d = d.AddDate(0, 0, 1) {
		closedCards, closedPoints := 0, 0
		for _, card := range cards {
			if card.ClosedAt != nil && !card.ClosedAt.UTC().Truncate(24*time.Hour).After(d) {
				closedCards++
				if card.StoryPoints != nil {
					closedPoints += *card.StoryPoints
				}
			}
		}
		dayIndex := d.Sub(startDay).Hours() / 24
		idealRemaining := float64(totalCards) * (1 - dayIndex/totalDays)
		if idealRemaining < 0 {
			idealRemaining = 0
		}
		data = append(data, epicBurnDay{
			Date:            d.Format("2006-01-02"),
			RemainingCards:  totalCards - closedCards,
			RemainingPoints: totalPoints - closedPoints,
			IdealRemaining:  math.Round(idealRemaining*10) / 10,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"epic": gin.H{"id": epic.ID, "name": epic.Name, "color": epic.Color, "total_cards": totalCards, "total_points": totalPoints},
		"data": data,
	})
}
