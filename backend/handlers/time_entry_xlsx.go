package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/xuri/excelize/v2"
)

// GetTimeEntryReportXLSX streams the grouped time-entry report as an XLSX file.
func GetTimeEntryReportXLSX(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var uFlag models.User
	database.DB.Select("time_tracking_viewer").First(&uFlag, userID)

	targetUserID := userID
	if globalRole == "admin" || uFlag.TimeTrackingViewer {
		if targetStr := c.Query("user_id"); targetStr != "" {
			if id, err := strconv.ParseUint(targetStr, 10, 64); err == nil {
				targetUserID = uint(id)
			}
		}
	}

	report, status, msg := assembleTimeEntryReport(c, targetUserID)
	if status != 0 {
		c.JSON(status, gin.H{"error": msg})
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := report.PeriodLabel
	if len([]rune(sheetName)) > 31 {
		sheetName = string([]rune(sheetName)[:31])
	}
	f.SetSheetName("Sheet1", sheetName)

	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	name := report.CompanyName
	if name == "" {
		name = "WarmDesk"
	}

	row := 1
	f.SetCellValue(sheetName, cell(1, row), name+" — Time Registration")
	row++
	f.SetCellValue(sheetName, cell(1, row), report.PeriodLabel)
	row += 2

	headers := []string{"Date", "Customer", "Project", "Activity", "Hours"}
	for col, h := range headers {
		cr := cell(col+1, row)
		f.SetCellValue(sheetName, cr, h)
		f.SetCellStyle(sheetName, cr, cr, bold)
	}
	row++

	for _, grp := range report.Groups {
		if len(grp.Entries) == 0 {
			continue
		}
		f.SetCellValue(sheetName, cell(1, row), grp.Label)
		f.SetCellStyle(sheetName, cell(1, row), cell(5, row), bold)
		row++

		for _, e := range grp.Entries {
			customerName := ""
			if e.Customer != nil {
				customerName = e.Customer.Name
			}
			projectName := ""
			if e.Project != nil {
				projectName = e.Project.Name
			}
			f.SetCellValue(sheetName, cell(1, row), e.Date.Format("2006-01-02"))
			f.SetCellValue(sheetName, cell(2, row), customerName)
			f.SetCellValue(sheetName, cell(3, row), projectName)
			f.SetCellValue(sheetName, cell(4, row), e.Description)
			f.SetCellValue(sheetName, cell(5, row), decimalHours(e.Minutes))
			row++
		}

		f.SetCellValue(sheetName, cell(4, row), "Total")
		f.SetCellValue(sheetName, cell(5, row), decimalHours(grp.TotalMinutes))
		f.SetCellStyle(sheetName, cell(1, row), cell(5, row), bold)
		row += 2
	}

	f.SetCellValue(sheetName, cell(4, row), "Total")
	f.SetCellValue(sheetName, cell(5, row), decimalHours(report.TotalMinutes))
	f.SetCellStyle(sheetName, cell(1, row), cell(5, row), bold)

	if report.UndeclarableMinutes > 0 {
		row++
		f.SetCellValue(sheetName, cell(4, row), "Undeclarable")
		f.SetCellValue(sheetName, cell(5, row), decimalHours(report.UndeclarableMinutes))
		row++
		f.SetCellValue(sheetName, cell(4, row), "Declarable")
		f.SetCellValue(sheetName, cell(5, row), decimalHours(report.DeclarableMinutes))
		f.SetCellStyle(sheetName, cell(1, row), cell(5, row), bold)
	}

	f.SetColWidth(sheetName, "A", "A", 14)
	f.SetColWidth(sheetName, "B", "B", 25)
	f.SetColWidth(sheetName, "C", "C", 25)
	f.SetColWidth(sheetName, "D", "D", 35)
	f.SetColWidth(sheetName, "E", "E", 10)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "xlsx generation failed"})
		return
	}

	if c.Query("base64") == "1" {
		c.JSON(http.StatusOK, gin.H{"data": base64.StdEncoding.EncodeToString(buf.Bytes())})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+xlsxFilename("time-tracking", report.PeriodLabel)+`"`)
	c.Data(http.StatusOK, xlsxMIME, buf.Bytes())
}

// GetTimeEntrySheetXLSX streams the weekly pivot timesheet as an XLSX file.
// Rows are customer/project/activity combos; columns are the 7 days of the week.
func GetTimeEntrySheetXLSX(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var uFlag models.User
	database.DB.Select("time_tracking_viewer").First(&uFlag, userID)

	targetUserID := userID
	if globalRole == "admin" || uFlag.TimeTrackingViewer {
		if targetStr := c.Query("user_id"); targetStr != "" {
			if id, err := strconv.ParseUint(targetStr, 10, 64); err == nil {
				targetUserID = uint(id)
			}
		}
	}

	now := time.Now()
	year := intOrDefault(c.Query("year"), now.Year())
	week := intOrDefault(c.Query("week"), isoWeek(now))

	weekStart := isoWeekStart(year, week)
	weekEnd := weekStart.AddDate(0, 0, 6)

	days := make([]time.Time, 7)
	for i := range days {
		days[i] = weekStart.AddDate(0, 0, i)
	}

	q := database.DB.
		Preload("Customer").
		Preload("Project").
		Where("date >= ? AND date <= ?", weekStart, weekEnd.Add(24*time.Hour-time.Second)).
		Order("date asc, id asc")
	if targetUserID > 0 {
		q = q.Where("user_id = ?", targetUserID)
	}
	var entries []models.TimeEntry
	if err := q.Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	type sheetRow struct {
		customerName string
		projectName  string
		description  string
		minutes      [7]int
	}
	rowMap := map[string]*sheetRow{}
	var rowOrder []string

	for _, e := range entries {
		custKey := ""
		if e.CustomerID != nil {
			custKey = strconv.FormatUint(uint64(*e.CustomerID), 10)
		}
		projKey := ""
		if e.ProjectID != nil {
			projKey = strconv.FormatUint(uint64(*e.ProjectID), 10)
		}
		key := custKey + "|" + projKey + "|" + e.Description
		if _, ok := rowMap[key]; !ok {
			cname, pname := "", ""
			if e.Customer != nil {
				cname = e.Customer.Name
			}
			if e.Project != nil {
				pname = e.Project.Name
			}
			rowMap[key] = &sheetRow{customerName: cname, projectName: pname, description: e.Description}
			rowOrder = append(rowOrder, key)
		}
		// (Weekday()+6)%7 maps Mon=0 … Sun=6
		dayIdx := (int(e.Date.Weekday()) + 6) % 7
		rowMap[key].minutes[dayIdx] += e.Minutes
	}

	f := excelize.NewFile()
	defer f.Close()
	weekLabel := fmt.Sprintf("Week %d · %d", week, year)
	f.SetSheetName("Sheet1", weekLabel)

	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	settings := loadAllSettings()
	name := settings["company_name"]
	if name == "" {
		name = "WarmDesk"
	}

	row := 1
	f.SetCellValue(weekLabel, cell(1, row), name+" — Time Registration")
	row++
	f.SetCellValue(weekLabel, cell(1, row), weekLabel)
	row += 2

	// Header: Customer/Project, Activity, Mon dd/mm … Sun dd/mm, Total
	colCount := 2 + 7 + 1 // info + 7 days + total
	hdr := make([]string, colCount)
	hdr[0] = "Customer / Project"
	hdr[1] = "Activity"
	for i, d := range days {
		hdr[2+i] = d.Format("Mon") + " " + d.Format("01/02")
	}
	hdr[9] = "Total"
	for col, h := range hdr {
		cr := cell(col+1, row)
		f.SetCellValue(weekLabel, cr, h)
		f.SetCellStyle(weekLabel, cr, cr, bold)
	}
	row++

	for _, key := range rowOrder {
		r := rowMap[key]
		info := r.projectName
		if r.customerName != "" {
			info = r.customerName + " / " + r.projectName
		}
		f.SetCellValue(weekLabel, cell(1, row), info)
		f.SetCellValue(weekLabel, cell(2, row), r.description)
		total := 0
		for i, m := range r.minutes {
			if m > 0 {
				f.SetCellValue(weekLabel, cell(3+i, row), decimalHours(m))
			}
			total += m
		}
		if total > 0 {
			f.SetCellValue(weekLabel, cell(10, row), decimalHours(total))
		}
		row++
	}

	// Totals row
	f.SetCellValue(weekLabel, cell(1, row), "Total")
	f.SetCellStyle(weekLabel, cell(1, row), cell(1, row), bold)
	grandTotal := 0
	for i, d := range days {
		dayTotal := 0
		for _, e := range entries {
			if e.Date.Format("2006-01-02") == d.Format("2006-01-02") {
				dayTotal += e.Minutes
			}
		}
		if dayTotal > 0 {
			f.SetCellValue(weekLabel, cell(3+i, row), decimalHours(dayTotal))
		}
		grandTotal += dayTotal
	}
	if grandTotal > 0 {
		f.SetCellValue(weekLabel, cell(10, row), decimalHours(grandTotal))
	}

	f.SetColWidth(weekLabel, "A", "A", 35)
	f.SetColWidth(weekLabel, "B", "B", 28)
	for i := 0; i < 8; i++ {
		colName, _ := excelize.ColumnNumberToName(3 + i)
		f.SetColWidth(weekLabel, colName, colName, 9)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "xlsx generation failed"})
		return
	}

	filename := fmt.Sprintf("time-tracking-week%d-%d.xlsx", week, year)
	if c.Query("base64") == "1" {
		c.JSON(http.StatusOK, gin.H{"data": base64.StdEncoding.EncodeToString(buf.Bytes())})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, xlsxMIME, buf.Bytes())
}
