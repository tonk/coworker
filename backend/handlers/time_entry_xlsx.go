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

// contractRateMap builds a map of projectID → (price_per_hour, currency) for
// projects that have a contract with a price set.
func contractRateMap(entries []models.TimeEntry) map[uint]struct {
	Rate     float64
	Currency string
} {
	infos := contractInfoMap(entries)
	if len(infos) == 0 {
		return nil
	}
	result := make(map[uint]struct {
		Rate     float64
		Currency string
	})
	for id, info := range infos {
		if info.BaseRate == nil {
			continue
		}
		result[id] = struct {
			Rate     float64
			Currency string
		}{*info.BaseRate, info.Currency}
	}
	return result
}

// entryCost computes the slot-aware cost for a time entry.
func entryCost(e models.TimeEntry, infos map[uint]projectContractInfo) (float64, string) {
	if e.ProjectID == nil {
		return 0, ""
	}
	info, ok := infos[*e.ProjectID]
	if !ok {
		return 0, ""
	}
	return entrySlotCost(e, info)
}

// fmtCost formats a cost value with 2 decimal places.
func fmtCost(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

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

	// Collect all entries and build contract info (rates + time slots)
	var allEntries []models.TimeEntry
	for _, g := range report.Groups {
		allEntries = append(allEntries, g.Entries...)
	}
	contractInfos := contractInfoMap(allEntries)
	baseRates := contractRateMap(allEntries)
	tr := pdfI18nFromLang(c.Query("lang"))

	f := excelize.NewFile()
	defer f.Close()

	sheetName := report.PeriodLabel
	if len([]rune(sheetName)) > 31 {
		sheetName = string([]rune(sheetName)[:31])
	}
	f.SetSheetName("Sheet1", sheetName)

	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	moneyFmt, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true},
		NumFmt: 4, // "#,##0.00"
	})
	subFmt, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Italic: true, Color: "666666"},
	})

	name := report.CompanyName
	if name == "" {
		name = "WarmDesk"
	}

	row := 1
	f.SetCellValue(sheetName, cell(1, row), name+" — Time Registration")
	row++
	f.SetCellValue(sheetName, cell(1, row), report.PeriodLabel)
	row += 2

	costCol := 8
	headers := []string{"Date", "Customer", "Project", "Activity", "Hours", "Currency", "Rate", "Cost"}
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
		f.SetCellStyle(sheetName, cell(1, row), cell(costCol, row), bold)
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
			cost, currency := entryCost(e, contractInfos)
			rate := 0.0
			if e.ProjectID != nil {
				if r, ok := baseRates[*e.ProjectID]; ok {
					rate = r.Rate
				}
			}
			f.SetCellValue(sheetName, cell(1, row), e.Date.Format("2006-01-02"))
			f.SetCellValue(sheetName, cell(2, row), customerName)
			f.SetCellValue(sheetName, cell(3, row), projectName)
			f.SetCellValue(sheetName, cell(4, row), e.Description)
			f.SetCellValue(sheetName, cell(5, row), decimalHours(e.Minutes))
			if currency != "" {
				f.SetCellValue(sheetName, cell(6, row), currency)
			}
			if rate > 0 {
				f.SetCellValue(sheetName, cell(7, row), rate)
			}
			if cost > 0 {
				f.SetCellValue(sheetName, cell(8, row), cost)
			}
			row++

			if e.ProjectID != nil {
				if info, ok := contractInfos[*e.ProjectID]; ok {
					for _, seg := range entryCostSegments(e, info, tr.Standard) {
						effRate := entryCostSegmentRate(seg)
						f.SetCellValue(sheetName, cell(1, row), seg.TimeRange)
						f.SetCellValue(sheetName, cell(4, row), "  "+seg.Label)
						f.SetCellValue(sheetName, cell(5, row), decimalHours(seg.Minutes))
						if seg.Currency != "" {
							f.SetCellValue(sheetName, cell(6, row), seg.Currency)
						}
						if effRate > 0 {
							f.SetCellValue(sheetName, cell(7, row), effRate)
						}
						if seg.Cost > 0 {
							f.SetCellValue(sheetName, cell(8, row), seg.Cost)
						}
						f.SetCellStyle(sheetName, cell(1, row), cell(8, row), subFmt)
						row++
					}
				}
			}
		}

		// Group subtotal — hours + cost
		grpCost := 0.0
		grpCurrency := ""
		for _, e := range grp.Entries {
			c, cur := entryCost(e, contractInfos)
			grpCost += c
			if cur != "" {
				grpCurrency = cur
			}
		}
		f.SetCellValue(sheetName, cell(4, row), "Total")
		f.SetCellValue(sheetName, cell(5, row), decimalHours(grp.TotalMinutes))
		if grpCurrency != "" {
			f.SetCellValue(sheetName, cell(6, row), grpCurrency)
		}
		if grpCost > 0 {
			f.SetCellValue(sheetName, cell(8, row), grpCost)
		}
		f.SetCellStyle(sheetName, cell(1, row), cell(8, row), bold)
		row += 2
	}

	// Grand total
	totalCost := 0.0
	totalCurrency := ""
	for _, g := range report.Groups {
		for _, e := range g.Entries {
			c, cur := entryCost(e, contractInfos)
			totalCost += c
			if cur != "" {
				totalCurrency = cur
			}
		}
	}
	f.SetCellValue(sheetName, cell(4, row), "Total")
	f.SetCellValue(sheetName, cell(5, row), decimalHours(report.TotalMinutes))
	if totalCurrency != "" {
		f.SetCellValue(sheetName, cell(6, row), totalCurrency)
	}
	if totalCost > 0 {
		f.SetCellValue(sheetName, cell(8, row), totalCost)
	}
	f.SetCellStyle(sheetName, cell(1, row), cell(8, row), moneyFmt)

	if report.UndeclarableMinutes > 0 {
		row++
		f.SetCellValue(sheetName, cell(4, row), "Undeclarable")
		f.SetCellValue(sheetName, cell(5, row), decimalHours(report.UndeclarableMinutes))
		row++
		f.SetCellValue(sheetName, cell(4, row), "Declarable")
		f.SetCellValue(sheetName, cell(5, row), decimalHours(report.DeclarableMinutes))
		f.SetCellStyle(sheetName, cell(1, row), cell(8, row), bold)
	}

	f.SetColWidth(sheetName, "A", "A", 14)
	f.SetColWidth(sheetName, "B", "B", 25)
	f.SetColWidth(sheetName, "C", "C", 25)
	f.SetColWidth(sheetName, "D", "D", 35)
	f.SetColWidth(sheetName, "E", "E", 10)
	f.SetColWidth(sheetName, "F", "F", 10)
	f.SetColWidth(sheetName, "G", "G", 10)
	f.SetColWidth(sheetName, "H", "H", 14)

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

	var weekStart time.Time
	if sd := c.Query("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			weekStart = t
		}
	}
	if weekStart.IsZero() {
		now := time.Now()
		year := intOrDefault(c.Query("year"), now.Year())
		week := intOrDefault(c.Query("week"), isoWeek(now))
		weekStart = isoWeekStart(year, week)
	}
	weekEnd := weekStart.AddDate(0, 0, 6)

	days := make([]time.Time, 7)
	for i := range days {
		days[i] = weekStart.AddDate(0, 0, i)
	}
	startDayIdx := int(weekStart.Weekday()) // 0=Sun, 1=Mon … used to map entry dates to columns

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
		// Offset from the week's start day (works for both Monday- and Sunday-start weeks)
		dayIdx := (int(e.Date.Weekday()) - startDayIdx + 7) % 7
		rowMap[key].minutes[dayIdx] += e.Minutes
	}

	f := excelize.NewFile()
	defer f.Close()
	isoYear, isoWeekNum := weekStart.AddDate(0, 0, 3).ISOWeek() // Thursday → stable ISO week/year
	weekLabel := fmt.Sprintf("Week %d · %d", isoWeekNum, isoYear)
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

	filename := fmt.Sprintf("time-tracking-week%d-%d.xlsx", isoWeekNum, isoYear)
	if c.Query("base64") == "1" {
		c.JSON(http.StatusOK, gin.H{"data": base64.StdEncoding.EncodeToString(buf.Bytes())})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, xlsxMIME, buf.Bytes())
}
