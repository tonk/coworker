package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
	"github.com/xuri/excelize/v2"
)

// GetTimeReportXLSX streams the project-board time report as an XLSX file.
func GetTimeReportXLSX(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var u models.User
	database.DB.Select("time_tracking_viewer").First(&u, userID)

	if !userCanViewReports(userID, globalRole, u.TimeTrackingViewer) {
		c.JSON(http.StatusForbidden, gin.H{"error": "reports are only available to project admins and system admins"})
		return
	}

	report, status, errMsg := assembleTimeReport(c, userID, globalRole, u.TimeTrackingViewer)
	if status != 0 {
		c.JSON(status, gin.H{"error": errMsg})
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Time Report"
	f.SetSheetName("Sheet1", sheet)

	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	title := report.CompanyName
	if title == "" {
		title = "Time Report"
	}

	row := 1
	f.SetCellValue(sheet, cell(1, row), title+" — "+report.PeriodLabel)
	row++
	f.SetCellValue(sheet, cell(1, row), "Generated: "+report.GeneratedAt)
	row += 2

	headers := []string{"Project", "Task", "Ref", "Assignees", "Date", "Time (min)", "Time (h)"}
	for col, h := range headers {
		cr := cell(col+1, row)
		f.SetCellValue(sheet, cr, h)
		f.SetCellStyle(sheet, cr, cr, bold)
	}
	row++

	for _, proj := range report.Projects {
		for _, card := range proj.Cards {
			f.SetCellValue(sheet, cell(1, row), proj.ProjectName)
			f.SetCellValue(sheet, cell(2, row), card.Title)
			f.SetCellValue(sheet, cell(3, row), card.CardRef)
			f.SetCellValue(sheet, cell(4, row), strings.Join(card.Assignees, ", "))
			f.SetCellValue(sheet, cell(5, row), card.UpdatedAt)
			f.SetCellValue(sheet, cell(6, row), card.TimeSpentMinutes)
			f.SetCellValue(sheet, cell(7, row), decimalHours(card.TimeSpentMinutes))
			row++
		}
		f.SetCellValue(sheet, cell(5, row), "Subtotal")
		f.SetCellValue(sheet, cell(6, row), proj.TotalMinutes)
		f.SetCellValue(sheet, cell(7, row), decimalHours(proj.TotalMinutes))
		f.SetCellStyle(sheet, cell(1, row), cell(7, row), bold)
		row += 2
	}

	f.SetCellValue(sheet, cell(5, row), "Grand Total")
	f.SetCellValue(sheet, cell(6, row), report.TotalMinutes)
	f.SetCellValue(sheet, cell(7, row), decimalHours(report.TotalMinutes))
	f.SetCellStyle(sheet, cell(1, row), cell(7, row), bold)

	f.SetColWidth(sheet, "A", "A", 25)
	f.SetColWidth(sheet, "B", "B", 45)
	f.SetColWidth(sheet, "C", "C", 10)
	f.SetColWidth(sheet, "D", "D", 30)
	f.SetColWidth(sheet, "E", "E", 12)
	f.SetColWidth(sheet, "F", "F", 12)
	f.SetColWidth(sheet, "G", "G", 10)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "xlsx generation failed"})
		return
	}

	filename := "time-report-" + strings.ToLower(strings.ReplaceAll(report.PeriodLabel, " ", "-")) + ".xlsx"
	if c.Query("base64") == "1" {
		c.JSON(http.StatusOK, gin.H{"data": base64.StdEncoding.EncodeToString(buf.Bytes())})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, xlsxMIME, buf.Bytes())
}

// cell converts 1-based column and row to an Excel cell address (e.g. col=1,row=2 → "A2").
func cell(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

// decimalHours converts minutes to a rounded 2-decimal float (e.g. 90 → 1.50).
func decimalHours(minutes int) float64 {
	return math.Round(float64(minutes)/60*100) / 100
}

const xlsxMIME = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

func xlsxFilename(prefix, periodLabel string) string {
	slug := strings.ToLower(strings.ReplaceAll(periodLabel, " ", "-"))
	slug = strings.ReplaceAll(slug, "/", "-")
	return fmt.Sprintf("%s-%s.xlsx", prefix, slug)
}
