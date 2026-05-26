package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"gorm.io/gorm"
)

// ListTimeEntries returns time entries for the authenticated user, optionally
// filtered by a date range (?from=YYYY-MM-DD&to=YYYY-MM-DD).
// Admin and timetracking roles may pass ?user_id= to view another user's entries
// (0 or omitted returns all users for those roles).
func ListTimeEntries(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var u models.User
	database.DB.Select("time_tracking_viewer").First(&u, userID)

	q := database.DB.
		Preload("Customer").
		Preload("Project").
		Preload("User").
		Order("date desc, id desc")

	if globalRole == "admin" || u.TimeTrackingViewer {
		if targetStr := c.Query("user_id"); targetStr != "" {
			if targetID, err := strconv.ParseUint(targetStr, 10, 64); err == nil && targetID > 0 {
				q = q.Where("user_id = ?", targetID)
			}
			// targetID == 0 → no user filter (all users)
		} else {
			q = q.Where("user_id = ?", userID)
		}
	} else {
		q = q.Where("user_id = ?", userID)
	}

	if ticketStr := c.Query("ticket_id"); ticketStr != "" {
		if ticketID, err := strconv.ParseUint(ticketStr, 10, 64); err == nil && ticketID > 0 {
			q = q.Where("ticket_id = ?", ticketID)
		}
	}
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			q = q.Where("date >= ?", t)
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			q = q.Where("date <= ?", t.Add(24*time.Hour-time.Second))
		}
	}

	var entries []models.TimeEntry
	if err := q.Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// checkContractNotExpired verifies that no contract associated with the given
// customer or project has an end_date before the entry date.
func checkContractNotExpired(db *gorm.DB, customerID, projectID *uint, entryDate time.Time) error {
	if projectID != nil {
		var project models.Project
		if err := db.First(&project, *projectID).Error; err == nil && project.ContractID != nil {
			var contract models.Contract
			if err := db.First(&contract, *project.ContractID).Error; err == nil && contract.EndDate != nil {
				if entryDate.After(*contract.EndDate) {
					return fmt.Errorf("the contract for this project expired on %s", contract.EndDate.Format("2006-01-02"))
				}
			}
		}
	}
	if customerID != nil {
		var contracts []models.Contract
		db.Where("customer_id = ? AND end_date IS NOT NULL AND end_date < ?", *customerID, entryDate).Find(&contracts)
		if len(contracts) > 0 {
			return fmt.Errorf("a contract for this customer expired on %s", contracts[0].EndDate.Format("2006-01-02"))
		}
	}
	return nil
}

// CreateTimeEntry logs a new time entry for the authenticated user.
func CreateTimeEntry(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		CustomerID  *uint   `json:"customer_id"`
		ProjectID   *uint   `json:"project_id"`
		TicketID    *uint   `json:"ticket_id"`
		Date        string  `json:"date"`
		Minutes     int     `json:"minutes"`
		Description string  `json:"description"`
		IsHoliday   bool    `json:"is_holiday"`
		StartTime   *string `json:"start_time"`
		EndTime     *string `json:"end_time"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.Minutes < 0 || (req.Minutes == 0 && !req.IsHoliday) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "minutes must be positive"})
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	// If a ticket_id is provided and customer_id is not, auto-populate from the ticket.
	if req.TicketID != nil && req.CustomerID == nil {
		var ticket models.Ticket
		if database.DB.Select("customer_id").First(&ticket, *req.TicketID).Error == nil {
			req.CustomerID = &ticket.CustomerID
		}
	}

	if err := checkContractNotExpired(database.DB, req.CustomerID, req.ProjectID, date); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry := models.TimeEntry{
		UserID:      userID,
		CustomerID:  req.CustomerID,
		ProjectID:   req.ProjectID,
		TicketID:    req.TicketID,
		Date:        date,
		Minutes:     req.Minutes,
		Description: req.Description,
		IsHoliday:   req.IsHoliday,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}
	if err := database.DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	database.DB.Preload("Customer").Preload("Project").Preload("User").First(&entry, entry.ID)
	c.JSON(http.StatusCreated, entry)
}

// UpdateTimeEntry updates a time entry owned by the authenticated user.
func UpdateTimeEntry(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var entry models.TimeEntry
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&entry).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var req struct {
		CustomerID  *uint   `json:"customer_id"`
		ProjectID   *uint   `json:"project_id"`
		Date        string  `json:"date"`
		Minutes     int     `json:"minutes"`
		Description string  `json:"description"`
		IsHoliday   bool    `json:"is_holiday"`
		StartTime   *string `json:"start_time"`
		EndTime     *string `json:"end_time"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.Minutes < 0 || (req.Minutes == 0 && !req.IsHoliday) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "minutes must be positive"})
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	if err := checkContractNotExpired(database.DB, req.CustomerID, req.ProjectID, date); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry.CustomerID = req.CustomerID
	entry.ProjectID = req.ProjectID
	entry.Date = date
	entry.Minutes = req.Minutes
	entry.Description = req.Description
	entry.IsHoliday = req.IsHoliday
	entry.StartTime = req.StartTime
	entry.EndTime = req.EndTime

	// Explicitly clear nullable FK columns so zeroing them is persisted.
	if req.CustomerID == nil {
		database.DB.Model(&entry).Update("customer_id", nil)
	}
	if req.ProjectID == nil {
		database.DB.Model(&entry).Update("project_id", nil)
	}
	if req.StartTime == nil {
		database.DB.Model(&entry).Update("start_time", nil)
	}
	if req.EndTime == nil {
		database.DB.Model(&entry).Update("end_time", nil)
	}
	if err := database.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	database.DB.Preload("Customer").Preload("Project").First(&entry, entry.ID)
	c.JSON(http.StatusOK, entry)
}

// DeleteTimeEntry removes a time entry owned by the authenticated user.
func DeleteTimeEntry(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result := database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.TimeEntry{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// timeEntryGroup is a single period bucket in the report.
type timeEntryGroup struct {
	Label               string             `json:"label"`
	Entries             []models.TimeEntry `json:"entries"`
	TotalMinutes        int                `json:"total_minutes"`
	UndeclarableMinutes int                `json:"undeclarable_minutes"`
	DeclarableMinutes   int                `json:"declarable_minutes"`
}

// TimeEntryReportResponse is the shape returned by GetTimeEntryReport.
type TimeEntryReportResponse struct {
	Period              string           `json:"period"`
	PeriodLabel         string           `json:"period_label"`
	Groups              []timeEntryGroup `json:"groups"`
	TotalMinutes        int              `json:"total_minutes"`
	UndeclarableMinutes int              `json:"undeclarable_minutes"`
	DeclarableMinutes   int              `json:"declarable_minutes"`
	CompanyName         string           `json:"company_name"`
	CompanyLogo         string           `json:"company_logo"`
}

// assembleTimeEntryReport builds the report data from query parameters.
// targetUserID == 0 means all users (admin/timetracking only).
// Returns (report, httpStatus, errMsg) — status is 0 on success.
func assembleTimeEntryReport(c *gin.Context, targetUserID uint) (*TimeEntryReportResponse, int, string) {
	now := time.Now()

	period := c.DefaultQuery("period", "month")
	year := intOrDefault(c.Query("year"), now.Year())
	month := intOrDefault(c.Query("month"), int(now.Month()))
	week := intOrDefault(c.Query("week"), isoWeek(now))

	var from, to time.Time
	var periodLabel string

	switch period {
	case "week":
		if sd := c.Query("start_date"); sd != "" {
			if t, err := time.Parse("2006-01-02", sd); err == nil {
				from = t
			}
		}
		if from.IsZero() {
			from = isoWeekStart(year, week)
		}
		to = from.AddDate(0, 0, 7)
		periodLabel = from.Format("Jan 2") + " – " + to.AddDate(0, 0, -1).Format("Jan 2, 2006")
	case "year":
		from = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(1, 0, 0)
		periodLabel = strconv.Itoa(year)
	default:
		period = "month"
		from = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 1, 0)
		periodLabel = from.Format("January 2006")
	}

	var entries []models.TimeEntry
	q := database.DB.
		Preload("Customer").
		Preload("Project").
		Where("date >= ? AND date < ?", from, to).
		Order("date asc, id asc")
	if targetUserID > 0 {
		q = q.Where("user_id = ?", targetUserID)
	}
	if err := q.Find(&entries).Error; err != nil {
		return nil, http.StatusInternalServerError, "internal error"
	}

	groupBy := c.DefaultQuery("group_by", "period")
	var groups []timeEntryGroup
	switch groupBy {
	case "customer":
		groups = buildGroupsByCustomer(entries)
	case "project":
		groups = buildGroupsByProject(entries)
	case "customer_project":
		groups = buildGroupsByCustomerProject(entries)
	default:
		groups = buildGroups(period, from, to, entries)
	}

	total, undeclarable := 0, 0
	for _, g := range groups {
		total += g.TotalMinutes
		undeclarable += g.UndeclarableMinutes
	}
	declarable := total - undeclarable
	if declarable < 0 {
		declarable = 0
	}

	settings := loadAllSettings()
	return &TimeEntryReportResponse{
		Period:              period,
		PeriodLabel:         periodLabel,
		Groups:              groups,
		TotalMinutes:        total,
		UndeclarableMinutes: undeclarable,
		DeclarableMinutes:   declarable,
		CompanyName:         settings["company_name"],
		CompanyLogo:         settings["company_logo"],
	}, 0, ""
}

// GetTimeEntryReport returns a grouped report for the authenticated user.
//
// Query params:
//
//	period=week|month|year  (default: month)
//	year=2026               (default: current year)
//	month=4                 (1-12, used when period=month)
//	week=17                 (ISO week, used when period=week)
//
// Grouping inside the period:
//
//	period=year  → group by month
//	period=month → group by ISO week
//	period=week  → group by day
func GetTimeEntryReport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var u models.User
	database.DB.Select("time_tracking_viewer").First(&u, userID)

	targetUserID := userID
	if globalRole == "admin" || u.TimeTrackingViewer {
		if targetStr := c.Query("user_id"); targetStr != "" {
			if id, err := strconv.ParseUint(targetStr, 10, 64); err == nil {
				targetUserID = uint(id) // 0 means all users
			}
		}
	}

	report, status, msg := assembleTimeEntryReport(c, targetUserID)
	if status != 0 {
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, report)
}

func buildGroups(period string, from, to time.Time, entries []models.TimeEntry) []timeEntryGroup {
	type key struct{ label string; order int }

	buckets := map[string]*timeEntryGroup{}
	var order []string

	label := func(e models.TimeEntry) string {
		switch period {
		case "week":
			return e.Date.Format("Monday, January 2")
		case "year":
			return e.Date.Format("January 2006")
		default: // month
			y, w := e.Date.ISOWeek()
			wStart := isoWeekStart(y, w)
			wEnd := wStart.AddDate(0, 0, 6)
			return "Week " + strconv.Itoa(w) + " (" + wStart.Format("Jan 2") + "–" + wEnd.Format("Jan 2") + ")"
		}
	}

	for i := range entries {
		l := label(entries[i])
		if _, ok := buckets[l]; !ok {
			buckets[l] = &timeEntryGroup{Label: l}
			order = append(order, l)
		}
		b := buckets[l]
		b.Entries = append(b.Entries, entries[i])
		b.TotalMinutes += entries[i].Minutes
		b.UndeclarableMinutes += projUndecl(entries[i])
	}

	// Fill empty buckets for periods with no entries so the report always
	// shows the full range (e.g. all 12 months for a year report).
	switch period {
	case "year":
		for m := from; m.Before(to); m = m.AddDate(0, 1, 0) {
			l := m.Format("January 2006")
			if _, ok := buckets[l]; !ok {
				buckets[l] = &timeEntryGroup{Label: l}
				order = append(order, l)
			}
		}
	case "week":
		for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
			l := d.Format("Monday, January 2")
			if _, ok := buckets[l]; !ok {
				buckets[l] = &timeEntryGroup{Label: l}
				order = append(order, l)
			}
		}
	}

	// Deduplicate order list while preserving first-seen ordering.
	seen := map[string]bool{}
	var deduped []string
	for _, l := range order {
		if !seen[l] {
			seen[l] = true
			deduped = append(deduped, l)
		}
	}

	result := make([]timeEntryGroup, 0, len(deduped))
	for _, l := range deduped {
		b := buckets[l]
		if b.Entries == nil {
			b.Entries = []models.TimeEntry{}
		}
		result = append(result, *b)
	}
	return setDeclarable(result)
}

func intOrDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

func isoWeek(t time.Time) int {
	_, w := t.ISOWeek()
	return w
}

// custKey encodes a nullable customer_id: 0 == no customer.
// projKey encodes a nullable project_id: 0 == no project.
// Using integer IDs as map keys avoids false splits when GORM's Preload
// populates the relation struct for only the first entry sharing a given ID.

func custIDKey(e models.TimeEntry) uint {
	if e.CustomerID != nil {
		return *e.CustomerID
	}
	return 0
}

func projIDKey(e models.TimeEntry) uint {
	if e.ProjectID != nil {
		return *e.ProjectID
	}
	return 0
}

func custLabel(e models.TimeEntry) string {
	if e.Customer != nil && e.Customer.Name != "" {
		return e.Customer.Name
	}
	return "(no customer)"
}

func projLabel(e models.TimeEntry) string {
	if e.Project != nil && e.Project.Name != "" {
		return e.Project.Name
	}
	return "(no project)"
}

func buildGroupsByCustomer(entries []models.TimeEntry) []timeEntryGroup {
	buckets := map[uint]*timeEntryGroup{}
	var order []uint
	for i := range entries {
		e := entries[i]
		k := custIDKey(e)
		if _, ok := buckets[k]; !ok {
			buckets[k] = &timeEntryGroup{Label: custLabel(e)}
			order = append(order, k)
		}
		b := buckets[k]
		b.Entries = append(b.Entries, e)
		b.TotalMinutes += e.Minutes
		b.UndeclarableMinutes += projUndecl(e)
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := order[i], order[j]
		if a == 0 { return false }
		if b == 0 { return true }
		return strings.ToLower(buckets[a].Label) < strings.ToLower(buckets[b].Label)
	})
	result := make([]timeEntryGroup, 0, len(order))
	for _, k := range order {
		b := buckets[k]
		if b.Entries == nil { b.Entries = []models.TimeEntry{} }
		result = append(result, *b)
	}
	return setDeclarable(result)
}

func buildGroupsByProject(entries []models.TimeEntry) []timeEntryGroup {
	buckets := map[uint]*timeEntryGroup{}
	var order []uint
	for i := range entries {
		e := entries[i]
		k := projIDKey(e)
		if _, ok := buckets[k]; !ok {
			buckets[k] = &timeEntryGroup{Label: projLabel(e)}
			order = append(order, k)
		}
		b := buckets[k]
		b.Entries = append(b.Entries, e)
		b.TotalMinutes += e.Minutes
		b.UndeclarableMinutes += projUndecl(e)
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := order[i], order[j]
		if a == 0 { return false }
		if b == 0 { return true }
		return strings.ToLower(buckets[a].Label) < strings.ToLower(buckets[b].Label)
	})
	result := make([]timeEntryGroup, 0, len(order))
	for _, k := range order {
		b := buckets[k]
		if b.Entries == nil { b.Entries = []models.TimeEntry{} }
		result = append(result, *b)
	}
	return setDeclarable(result)
}

func buildGroupsByCustomerProject(entries []models.TimeEntry) []timeEntryGroup {
	type cpKey struct{ cid, pid uint }
	buckets := map[cpKey]*timeEntryGroup{}
	var order []cpKey
	for i := range entries {
		e := entries[i]
		k := cpKey{custIDKey(e), projIDKey(e)}
		if _, ok := buckets[k]; !ok {
			buckets[k] = &timeEntryGroup{Label: custLabel(e) + " › " + projLabel(e)}
			order = append(order, k)
		}
		b := buckets[k]
		b.Entries = append(b.Entries, e)
		b.TotalMinutes += e.Minutes
		b.UndeclarableMinutes += projUndecl(e)
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := order[i], order[j]
		la, lb := buckets[a].Label, buckets[b].Label
		if a.cid == 0 && b.cid != 0 { return false }
		if b.cid == 0 && a.cid != 0 { return true }
		if a.pid == 0 && b.pid != 0 { return false }
		if b.pid == 0 && a.pid != 0 { return true }
		return strings.ToLower(la) < strings.ToLower(lb)
	})
	result := make([]timeEntryGroup, 0, len(order))
	for _, k := range order {
		b := buckets[k]
		if b.Entries == nil { b.Entries = []models.TimeEntry{} }
		result = append(result, *b)
	}
	return setDeclarable(result)
}

// projUndecl returns how many minutes of a single entry are undeclarable,
// derived from the project's UndeclarableMinutes setting (never exceeds entry.Minutes).
func projUndecl(e models.TimeEntry) int {
	if e.Project == nil || e.Project.UndeclarableMinutes <= 0 {
		return 0
	}
	if e.Project.UndeclarableMinutes >= e.Minutes {
		return e.Minutes
	}
	return e.Project.UndeclarableMinutes
}

// setDeclarable computes DeclarableMinutes = max(0, Total - Undeclarable) for each group.
func setDeclarable(groups []timeEntryGroup) []timeEntryGroup {
	for i := range groups {
		d := groups[i].TotalMinutes - groups[i].UndeclarableMinutes
		if d < 0 {
			d = 0
		}
		groups[i].DeclarableMinutes = d
	}
	return groups
}
