package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// GetTimeEntryReportPDF renders a time entry report as PDF.
// Admin and timetracking roles may pass ?user_id= to render another user's report.
func GetTimeEntryReportPDF(c *gin.Context) {
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

	// Font preference comes from the requesting user (admin/viewer), not the target.
	var requestingUser models.User
	fontFamily := "FreeSans"
	if err := database.DB.First(&requestingUser, userID).Error; err == nil {
		fontFamily = pdfFontFamily(requestingUser.Font)
	}
	if fam, ok := pdfFontFromParam(c.Query("font")); ok {
		fontFamily = fam
	}
	tr := pdfI18nFromLang(c.Query("lang"))

	// Employee name shown in the PDF header — resolved from targetUserID.
	var employeeName string
	if targetUserID == 0 {
		employeeName = tr.AllEmployees
	} else if targetUserID == userID {
		employeeName = requestingUser.DisplayName
		if employeeName == "" {
			employeeName = requestingUser.Username
		}
	} else {
		var targetUser models.User
		if err := database.DB.First(&targetUser, targetUserID).Error; err == nil {
			employeeName = targetUser.DisplayName
			if employeeName == "" {
				employeeName = targetUser.Username
			}
		}
	}

	// ── Build PDF ─────────────────────────────────────────────────────────────
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AliasNbPages("{nb}")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(true, 20)

	for fam, files := range pdfFonts {
		pdf.AddUTF8FontFromBytes(fam, "", mustFont(files[0]))
		pdf.AddUTF8FontFromBytes(fam, "B", mustFont(files[1]))
	}

	company := report.CompanyName
	if company == "" {
		company = "WarmDesk"
	}
	title := tr.TimeReport + " — " + report.PeriodLabel
	pdf.SetTitle(title, true)
	pdf.SetAuthor(employeeName, true)
	pdf.SetSubject(company+" — "+title, true)
	pdf.SetCreator(company, true)

	ff := fontFamily
	emp := employeeName
	rpt := report
	showPageNumbers := c.DefaultQuery("show_page_numbers", "1") != "0"
	pdf.SetFooterFunc(func() {
		if !showPageNumbers {
			return
		}
		pdf.SetY(-12)
		pdf.SetFont(ff, "", 8)
		setTxt(pdf, clrMuted)
		label := "WarmDesk — " + rpt.PeriodLabel
		if rpt.CompanyName != "" {
			label = rpt.CompanyName + " — " + rpt.PeriodLabel
		}
		pdf.CellFormat(pdfBodyW/2, 5, label, "", 0, "L", false, 0, "")
		pdf.CellFormat(pdfBodyW/2, 5,
			fmt.Sprintf("%s %d / {nb}", tr.Page, pdf.PageNo()), "", 0, "R", false, 0, "")
		_ = emp
	})

	// ── Determine paging mode before building the header func ────────────────
	groupBy := c.DefaultQuery("group_by", "period")
	pageBreakPerCustomer := groupBy == "customer" && c.Query("page_break") == "customer"
	showAbbr := c.Query("show_abbr") == "1"
	showCosts := c.Query("show_costs") == "1"
	showUndeclarable := c.DefaultQuery("show_undeclarable", "1") != "0"

	// Build contract info map for cost display (includes time slots for slot-aware costing).
	var contractInfos map[uint]projectContractInfo
	if showCosts {
		var allEntries []models.TimeEntry
		for _, g := range report.Groups {
			allEntries = append(allEntries, g.Entries...)
		}
		contractInfos = contractInfoMap(allEntries)
	}

	// Pre-load logo once so the header closure doesn't hit disk on every page.
	var cachedLogoBytes []byte
	var cachedLogoExt string
	var cachedLogoOK bool
	if b, ext, ok := resolveLogoBytes(report.CompanyLogo); ok {
		cachedLogoBytes, cachedLogoExt, cachedLogoOK = b, ext, true
	}

	// drawDocHeader renders the logo / company / title / period block and
	// leaves the cursor just below the rule line, ready for the table header.
	drawDocHeader := func() {
		logoY := pdfMargin
		logoLoaded := false
		if cachedLogoOK {
			logoLoaded = renderLogoIntoPDF(pdf, cachedLogoBytes, cachedLogoExt, pdfMargin, logoY)
		}
		textX := pdfMargin
		if logoLoaded {
			textX = pdfMargin + 32
		}
		textW := pdfPageW - pdfMargin - textX
		pdf.SetXY(textX, logoY)
		if report.CompanyName != "" {
			pdf.SetFont(fontFamily, "B", 13)
			setTxt(pdf, clrPrimary)
			pdf.CellFormat(textW, 7, report.CompanyName, "", 2, "L", false, 0, "")
			pdf.SetX(textX)
		}
		pdf.SetFont(fontFamily, "B", 16)
		setTxt(pdf, clrText)
		pdf.CellFormat(textW, 9, tr.TimeReport, "", 2, "L", false, 0, "")
		pdf.SetX(textX)
		pdf.SetFont(fontFamily, "", 10)
		setTxt(pdf, clrMuted)
		pdf.CellFormat(textW, 6, pdfTranslateLabel(report.PeriodLabel, tr), "", 2, "L", false, 0, "")
		pdf.SetX(textX)
		pdf.SetFont(fontFamily, "", 9)
		pdf.CellFormat(textW, 5, employeeName, "", 2, "L", false, 0, "")
		ruleY := pdf.GetY() + 2
		if logoLoaded && ruleY < logoY+22 {
			ruleY = logoY + 22
		}
		setDraw(pdf, clrPrimary)
		pdf.SetLineWidth(0.4)
		pdf.Line(pdfMargin, ruleY, pdfMargin+pdfBodyW, ruleY)
		pdf.SetY(ruleY + 4)
	}

	// ── Column widths (mm) ────────────────────────────────────────────────────
	// When abbreviations are shown the date column is wider; description shrinks
	// by the same amount so the total stays constant.
	// colAbbr is a fixed sub-cell inside colDate so dates always align regardless
	// of whether the language uses 2- or 3-character day abbreviations.
	const colAbbr = 9.0 // wide enough for 3-char abbreviations in any language
	var (
		colDate = 25.0
		colCust = 45.0
		colProj = 45.0
		colDesc = 50.0
		colCost = 0.0
	)
	if showCosts {
		colDate = 22.0
		colCust = 38.0
		colProj = 38.0
		colDesc = 47.0
		colCost = 20.0
	}
	if showAbbr {
		colDate = colAbbr + 25.0
		if showCosts {
			colDesc = 37.0
		} else {
			colDesc = 40.0
		}
	}
	colHours := pdfBodyW - colDate - colCust - colProj - colDesc - colCost
	nonHourW := colDate + colCust + colProj + colDesc + colCost

	// ── Table header ──────────────────────────────────────────────────────────
	drawTableHeader := func() {
		setFill(pdf, clrPrimary)
		setTxt(pdf, rgb{255, 255, 255})
		pdf.SetFont(fontFamily, "B", 8)
		const rowH = 6.0
		pdf.CellFormat(colDate, rowH, tr.Date, "0", 0, "L", true, 0, "")
		pdf.CellFormat(colCust, rowH, tr.Customer, "0", 0, "L", true, 0, "")
		pdf.CellFormat(colProj, rowH, tr.Project, "0", 0, "L", true, 0, "")
		pdf.CellFormat(colDesc, rowH, tr.Activity, "0", 0, "L", true, 0, "")
		if showCosts {
			pdf.CellFormat(colCost, rowH, tr.Cost, "0", 0, "R", true, 0, "")
		}
		pdf.CellFormat(colHours, rowH, tr.Hours, "0", 1, "R", true, 0, "")
		setTxt(pdf, clrText)
	}

	pdf.AddPage()

	// Page 1 — draw header manually. SetHeaderFunc is registered below, AFTER
	// this page, so it does not fire here and there is no duplication.
	drawDocHeader()
	drawTableHeader()

	// For page-per-customer breaks the header func replays the full header so
	// every customer page looks identical to the first. For continuous reports
	// only the column-header bar repeats on auto page breaks.
	if pageBreakPerCustomer {
		pdf.SetHeaderFunc(func() {
			drawDocHeader()
			drawTableHeader()
		})
	} else {
		pdf.SetHeaderFunc(drawTableHeader)
	}

	// ── Group sections ────────────────────────────────────────────────────────

	var activeGroups []timeEntryGroup
	for _, g := range report.Groups {
		if len(g.Entries) > 0 {
			activeGroups = append(activeGroups, g)
		}
	}

	const rowH = 5.5
	altRow := false

	clrAlt := rgb{245, 247, 250}
	clrWhite := rgb{255, 255, 255}

	for gi, grp := range activeGroups {
		// Group header bar
		grpY := pdf.GetY()
		setFill(pdf, rgb{230, 235, 242})
		setDraw(pdf, rgb{200, 210, 220})
		pdf.SetLineWidth(0.1)
		pdf.Rect(pdfMargin, grpY, pdfBodyW, 5.5, "FD")
		pdf.SetFont(fontFamily, "B", 8.5)
		setTxt(pdf, clrText)
		pdf.CellFormat(pdfBodyW-20, 5.5, pdfTranslateLabel(grp.Label, tr), "", 0, "L", false, 0, "")
		pdf.SetFont(fontFamily, "B", 8.5)
		setTxt(pdf, clrPrimary)
		pdf.CellFormat(20, 5.5, fmtDecimalH(grp.DeclarableMinutes), "", 1, "R", false, 0, "")

		altRow = false
		for _, e := range grp.Entries {
			fillClr := clrWhite
			if altRow {
				fillClr = clrAlt
			}
			altRow = !altRow

			setFill(pdf, fillClr)
			setDraw(pdf, rgb{220, 225, 230})
			pdf.SetLineWidth(0.1)
			pdf.SetFont(fontFamily, "", 8)
			setTxt(pdf, clrText)

			customerName := ""
			if e.Customer != nil {
				customerName = e.Customer.Name
			}
			projectName := ""
			if e.Project != nil {
				projectName = e.Project.Name
			}

			// Determine if we have a time range to show.
			timeRange := ""
			if e.StartTime != nil && *e.StartTime != "" && e.EndTime != nil && *e.EndTime != "" {
				timeRange = *e.StartTime + "–" + *e.EndTime
			}
			entryRowH := rowH
			if timeRange != "" {
				entryRowH = 9.5
			}

			dateCell := e.Date.Format("2006-01-02")
			rowStartX := pdf.GetX()
			rowStartY := pdf.GetY()

			if showAbbr {
				abbrIdx := (int(e.Date.Weekday()) + 6) % 7
				dateAlign := "L"
				if timeRange != "" {
					dateAlign = "LT"
				}
				pdf.CellFormat(colAbbr, entryRowH, tr.DaysAbbr[abbrIdx], "B", 0, dateAlign, true, 0, "")
				pdf.CellFormat(colDate-colAbbr, entryRowH, dateCell, "B", 0, dateAlign, true, 0, "")
			} else {
				dateAlign := "L"
				if timeRange != "" {
					dateAlign = "LT"
				}
				pdf.CellFormat(colDate, entryRowH, dateCell, "B", 0, dateAlign, true, 0, "")
			}

			// Overlay time range in small font below the date.
			if timeRange != "" {
				afterDateX := pdf.GetX()
				timeX := rowStartX
				if showAbbr {
					timeX = rowStartX + colAbbr
				}
				pdf.SetXY(timeX, rowStartY+5.5)
				pdf.SetFont(fontFamily, "", 6.5)
				setTxt(pdf, clrMuted)
				pdf.CellFormat(colDate-(timeX-rowStartX), 3.5, timeRange, "", 0, "L", false, 0, "")
				pdf.SetXY(afterDateX, rowStartY)
				pdf.SetFont(fontFamily, "", 8)
				setTxt(pdf, clrText)
			}

			pdf.CellFormat(colCust, entryRowH, customerName, "B", 0, "L", true, 0, "")
			pdf.CellFormat(colProj, entryRowH, projectName, "B", 0, "L", true, 0, "")
			pdf.CellFormat(colDesc, entryRowH, truncate(e.Description, 32), "B", 0, "L", true, 0, "")
			if showCosts {
				costStr := ""
				var cost float64
				var cur string
				if e.ProjectID != nil {
					if info, ok := contractInfos[*e.ProjectID]; ok {
						cost, cur = entrySlotCost(e, info)
					}
				}
				if cost > 0 {
					costStr = fmtCost(cost) + " " + cur
				}
				pdf.CellFormat(colCost, entryRowH, costStr, "B", 0, "R", true, 0, "")
			}
			pdf.CellFormat(colHours, entryRowH, fmtDecimalH(pdfEntryDeclarable(e)), "B", 1, "R", true, 0, "")

			// ── Per-entry time-slot sub-rows ─────────────────────────────────
			// Only shown when the entry has a time range AND the project has slots.
			if timeRange != "" && e.ProjectID != nil {
				if info, ok := contractInfos[*e.ProjectID]; ok && len(info.TimeSlots) > 0 {
					entryStartMins := parseHHMM(*e.StartTime)
					entryEndMins := parseHHMM(*e.EndTime)
					if entryStartMins >= 0 && entryEndMins >= 0 && entryStartMins != entryEndMins {
						_, slotSegs := entrySlotBreakdown(e, info)
						stdSegs := standardTimeSegments(entryStartMins, entryTimelineEnd(entryStartMins, entryEndMins), slotSegs)

						type subRow struct {
							startMins    int
							timeRangeStr string
							label        string
							minutes      int
							cost         float64
							currency     string
						}
						var subRows []subRow

						for _, seg := range stdSegs {
							c := 0.0
							if info.BaseRate != nil {
								c = float64(seg.Minutes) / 60.0 * *info.BaseRate
							}
							subRows = append(subRows, subRow{
								startMins:    seg.Start,
								timeRangeStr: fmtWallClockPDF(seg.Start) + "–" + fmtWallClockPDF(seg.End),
								label:        tr.Standard,
								minutes:      seg.Minutes,
								cost:         c,
								currency:     info.Currency,
							})
						}
						for _, s := range slotSegs {
							var c float64
							switch {
							case s.HourlyRate != nil:
								c = float64(s.Minutes) / 60.0 * *s.HourlyRate
							case s.MultiplicationFactor != nil && info.BaseRate != nil:
								c = float64(s.Minutes) / 60.0 * *info.BaseRate * *s.MultiplicationFactor
							case info.BaseRate != nil:
								c = float64(s.Minutes) / 60.0 * *info.BaseRate
							}
							lbl := s.Label
							if lbl == "" {
								lbl = "—"
							}
							subRows = append(subRows, subRow{
								startMins:    s.OverlapStart,
								timeRangeStr: fmtWallClockPDF(s.OverlapStart) + "–" + fmtWallClockPDF(s.OverlapEnd),
								label:        lbl,
								minutes:      s.Minutes,
								cost:         c,
								currency:     info.Currency,
							})
						}
						sort.Slice(subRows, func(i, j int) bool {
							return subRows[i].startMins < subRows[j].startMins
						})

						const subH = 4.5
						for _, sr := range subRows {
							setFill(pdf, rgb{245, 248, 253})
							setDraw(pdf, rgb{220, 225, 230})
							pdf.SetLineWidth(0.1)
							pdf.SetFont(fontFamily, "", 7)
							setTxt(pdf, clrMuted)
							if showAbbr {
								pdf.CellFormat(colAbbr, subH, "", "B", 0, "L", true, 0, "")
								pdf.CellFormat(colDate-colAbbr, subH, sr.timeRangeStr, "B", 0, "L", true, 0, "")
							} else {
								pdf.CellFormat(colDate, subH, sr.timeRangeStr, "B", 0, "L", true, 0, "")
							}
							pdf.CellFormat(colCust, subH, "", "B", 0, "L", true, 0, "")
							pdf.CellFormat(colProj, subH, "", "B", 0, "L", true, 0, "")
							pdf.CellFormat(colDesc, subH, "  "+sr.label, "B", 0, "L", true, 0, "")
							if showCosts {
								costStr := ""
								if sr.cost > 0 {
									costStr = fmtCost(sr.cost) + " " + sr.currency
								}
								pdf.CellFormat(colCost, subH, costStr, "B", 0, "R", true, 0, "")
							}
							pdf.CellFormat(colHours, subH, fmtDecimalH(sr.minutes), "B", 1, "R", true, 0, "")
						}
					}
				}
			}
		}

		// Group subtotal
		grpCost := 0.0
		grpCurrency := ""
		if showCosts {
			for _, e := range grp.Entries {
				if e.ProjectID != nil {
					if info, ok := contractInfos[*e.ProjectID]; ok {
						c, cur := entrySlotCost(e, info)
						grpCost += c
						if cur != "" {
							grpCurrency = cur
						}
					}
				}
			}
		}
		pdf.SetFont(fontFamily, "B", 8)
		setTxt(pdf, clrMuted)
		setFill(pdf, rgb{238, 242, 248})
		pdf.CellFormat(nonHourW-colCost, rowH, "  "+pdfTranslateLabel(grp.Label, tr)+" "+tr.Total, "0", 0, "R", true, 0, "")
		if showCosts {
			costStr := ""
			if grpCost > 0 {
				costStr = fmtCost(grpCost) + " " + grpCurrency
			}
			pdf.CellFormat(colCost, rowH, costStr, "0", 0, "R", true, 0, "")
		}
		setTxt(pdf, clrPrimary)
		pdf.CellFormat(colHours, rowH, fmtDecimalH(grp.DeclarableMinutes), "0", 1, "R", true, 0, "")

		// Per-group undeclarable line (customer grouping only)
		if showUndeclarable && groupBy == "customer" && grp.UndeclarableMinutes > 0 {
			setFill(pdf, rgb{252, 245, 245})
			setTxt(pdf, clrMuted)
			pdf.SetFont(fontFamily, "", 8)
			pdf.CellFormat(nonHourW-colCost, rowH, "  "+tr.Undeclarable, "0", 0, "R", true, 0, "")
			if showCosts {
				pdf.CellFormat(colCost, rowH, "", "0", 0, "R", true, 0, "")
			}
			setTxt(pdf, rgb{180, 80, 80})
			pdf.CellFormat(colHours, rowH, "−"+fmtDecimalH(grp.UndeclarableMinutes), "0", 1, "R", true, 0, "")
		}

		if pageBreakPerCustomer && gi < len(activeGroups)-1 {
			pdf.AddPage()
		} else {
			pdf.Ln(2)
		}
	}

	// ── Grand total (omitted for page-per-customer — each customer already has
	// its own subtotal and a grand total on the last page would be confusing) ──
	if !pageBreakPerCustomer {
		grandMinutes := report.TotalMinutes
		if report.UndeclarableMinutes > 0 {
			grandMinutes = report.DeclarableMinutes
		}
		totalCost := 0.0
		totalCurrency := ""
		if showCosts {
			for _, g := range report.Groups {
				for _, e := range g.Entries {
					if e.ProjectID == nil {
						continue
					}
					if info, ok := contractInfos[*e.ProjectID]; ok {
						c, cur := entrySlotCost(e, info)
						totalCost += c
						if cur != "" {
							totalCurrency = cur
						}
					}
				}
			}
		}
		setFill(pdf, clrPrimary)
		setTxt(pdf, rgb{255, 255, 255})
		pdf.SetFont(fontFamily, "B", 9)
		pdf.CellFormat(nonHourW-colCost, rowH+1, "  "+tr.Total, "0", 0, "L", true, 0, "")
		if showCosts {
			costStr := ""
			if totalCost > 0 {
				costStr = fmtCost(totalCost) + " " + totalCurrency
			}
			pdf.CellFormat(colCost, rowH+1, costStr, "0", 0, "R", true, 0, "")
		}
		pdf.CellFormat(colHours, rowH+1, fmtDecimalH(grandMinutes), "0", 1, "R", true, 0, "")

		if showUndeclarable && groupBy == "customer" && report.UndeclarableMinutes > 0 {
			setFill(pdf, rgb{252, 245, 245})
			setTxt(pdf, clrMuted)
			pdf.SetFont(fontFamily, "", 8.5)
			pdf.CellFormat(nonHourW-colCost, rowH, "  "+tr.Undeclarable, "0", 0, "L", true, 0, "")
			if showCosts {
				pdf.CellFormat(colCost, rowH, "", "0", 0, "R", true, 0, "")
			}
			setTxt(pdf, rgb{180, 80, 80})
			pdf.CellFormat(colHours, rowH, "−"+fmtDecimalH(report.UndeclarableMinutes), "0", 1, "R", true, 0, "")
		}
	}

	// ── Output ────────────────────────────────────────────────────────────────
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pdf generation failed"})
		return
	}

	filename := "time-registration-" + report.PeriodLabel + ".pdf"
	if c.Query("base64") == "1" {
		c.JSON(http.StatusOK, gin.H{"data": base64.StdEncoding.EncodeToString(buf.Bytes())})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// pdfEntryDeclarable returns the declarable minutes for a single entry.
func pdfEntryDeclarable(e models.TimeEntry) int {
	if e.Project == nil || e.Project.UndeclarableMinutes <= 0 {
		return e.Minutes
	}
	d := e.Minutes - e.Project.UndeclarableMinutes
	if d < 0 {
		return 0
	}
	return d
}

// fmtDecimalH formats minutes as decimal hours ("8.00", "7.50").
func fmtDecimalH(minutes int) string {
	if minutes == 0 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", float64(minutes)/60.0)
}

// truncate clips a string to max n runes.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// projectContractInfo holds the base rate, currency and time slots for a project's contract.
type projectContractInfo struct {
	BaseRate  *float64
	Currency  string
	TimeSlots []models.ContractTimeSlot
}

// contractInfoMap builds a projectContractInfo for every project referenced by the entries.
func contractInfoMap(entries []models.TimeEntry) map[uint]projectContractInfo {
	projIDs := make(map[uint]struct{})
	for _, e := range entries {
		if e.ProjectID != nil {
			projIDs[*e.ProjectID] = struct{}{}
		}
	}
	if len(projIDs) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(projIDs))
	for id := range projIDs {
		ids = append(ids, id)
	}
	var projects []models.Project
	database.DB.Where("id IN ?", ids).Find(&projects)
	result := make(map[uint]projectContractInfo)
	for _, p := range projects {
		if p.ContractID == nil {
			continue
		}
		var c models.Contract
		if err := database.DB.Preload("TimeSlots").First(&c, *p.ContractID).Error; err != nil {
			continue
		}
		info := projectContractInfo{Currency: c.Currency, TimeSlots: c.TimeSlots}
		if c.PricePerHour != nil {
			info.BaseRate = c.PricePerHour
		}
		result[p.ID] = info
	}
	return result
}

const minutesPerDay = 24 * 60

// timelineInterval is a half-open [start, end) range on a timeline where 0 is midnight on the entry date.
type timelineInterval struct {
	start int
	end   int
}

// entryTimelineEnd returns the entry end position on the entry-date timeline (may exceed minutesPerDay).
func entryTimelineEnd(startMins, endMins int) int {
	if endMins > startMins {
		return endMins
	}
	return minutesPerDay + endMins
}

// entryTimelineIntervals converts entry start/end times into timeline intervals relative to the entry date.
func entryTimelineIntervals(startMins, endMins int) []timelineInterval {
	if startMins < 0 || endMins < 0 || startMins == endMins {
		return nil
	}
	if endMins > startMins {
		return []timelineInterval{{startMins, endMins}}
	}
	return []timelineInterval{
		{startMins, minutesPerDay},
		{minutesPerDay, minutesPerDay + endMins},
	}
}

// slotEndDayOffset returns how many calendar days after the anchor day the end time falls.
func slotEndDayOffset(slot models.ContractTimeSlot) int {
	slotStart := parseHHMM(slot.StartTime)
	slotEnd := parseHHMM(slot.EndTime)
	if slotStart < 0 || slotEnd < 0 {
		return 0
	}
	if slotEnd > slotStart {
		return 0
	}
	if slot.EndDayOffset > 0 {
		return slot.EndDayOffset
	}
	return 1
}

// slotTimelineIntervals returns slot coverage on the entry-date timeline, including overnight and multi-day spans.
func slotTimelineIntervals(slot models.ContractTimeSlot, entryDate time.Time) []timelineInterval {
	slotStart := parseHHMM(slot.StartTime)
	slotEnd := parseHHMM(slot.EndTime)
	if slotStart < 0 || slotEnd < 0 || slotStart == slotEnd {
		return nil
	}

	endOffset := slotEndDayOffset(slot)
	var out []timelineInterval
	for dayOffset := -endOffset; dayOffset <= 1; dayOffset++ {
		day := entryDate.AddDate(0, 0, dayOffset)
		if !dayTypeMatches(slot.DayType, day.Weekday()) {
			continue
		}
		base := dayOffset * minutesPerDay
		if slotEnd > slotStart {
			out = append(out, timelineInterval{base + slotStart, base + slotEnd})
			continue
		}
		out = append(out, timelineInterval{base + slotStart, base + endOffset*minutesPerDay + slotEnd})
	}
	return out
}

// intervalOverlap returns the overlap of two timeline intervals.
func intervalOverlap(a, b timelineInterval) (timelineInterval, bool) {
	start := a.start
	if b.start > start {
		start = b.start
	}
	end := a.end
	if b.end < end {
		end = b.end
	}
	if end <= start {
		return timelineInterval{}, false
	}
	return timelineInterval{start, end}, true
}

// mergeTimelineIntervals merges overlapping or adjacent timeline intervals.
func mergeTimelineIntervals(ints []timelineInterval) []timelineInterval {
	if len(ints) == 0 {
		return nil
	}
	sorted := make([]timelineInterval, len(ints))
	copy(sorted, ints)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].start < sorted[j].start })

	merged := []timelineInterval{sorted[0]}
	for _, iv := range sorted[1:] {
		last := &merged[len(merged)-1]
		if iv.start <= last.end {
			if iv.end > last.end {
				last.end = iv.end
			}
			continue
		}
		merged = append(merged, iv)
	}
	return merged
}

// parseHHMM parses a "HH:MM" string into minutes since midnight. Returns -1 on error.
func parseHHMM(s string) int {
	if len(s) != 5 || s[2] != ':' {
		return -1
	}
	h, err1 := strconv.Atoi(s[:2])
	m, err2 := strconv.Atoi(s[3:])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return -1
	}
	return h*60 + m
}

// dayTypeMatches returns true when the slot's DayType applies to the given weekday.
func dayTypeMatches(dayType string, weekday time.Weekday) bool {
	switch dayType {
	case "all", "":
		return true
	case "weekdays":
		return weekday >= time.Monday && weekday <= time.Friday
	case "weekends":
		return weekday == time.Saturday || weekday == time.Sunday
	case "monday":
		return weekday == time.Monday
	case "tuesday":
		return weekday == time.Tuesday
	case "wednesday":
		return weekday == time.Wednesday
	case "thursday":
		return weekday == time.Thursday
	case "friday":
		return weekday == time.Friday
	case "saturday":
		return weekday == time.Saturday
	case "sunday":
		return weekday == time.Sunday
	}
	return false
}

// slotMinutes describes how many minutes of an entry fall within a specific time slot.
type slotMinutes struct {
	Label                string
	Minutes              int
	OverlapStart         int // minutes on the entry-date timeline
	OverlapEnd           int // minutes on the entry-date timeline
	HourlyRate           *float64
	MultiplicationFactor *float64
}

// timeSegment represents a contiguous block of standard (non-slot) time.
type timeSegment struct {
	Start   int
	End     int
	Minutes int
}

// fmtWallClockPDF formats a timeline minute value as "HH:MM" on a 24-hour clock.
func fmtWallClockPDF(mins int) string {
	mins %= minutesPerDay
	if mins < 0 {
		mins += minutesPerDay
	}
	return fmt.Sprintf("%02d:%02d", mins/60, mins%60)
}

// standardTimeSegments returns the entry time ranges NOT covered by any slot,
// sorted chronologically.
func standardTimeSegments(entryStart, entryEnd int, slots []slotMinutes) []timeSegment {
	if len(slots) == 0 {
		return []timeSegment{{entryStart, entryEnd, entryEnd - entryStart}}
	}
	sorted := make([]slotMinutes, len(slots))
	copy(sorted, slots)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].OverlapStart < sorted[j].OverlapStart
	})
	var segs []timeSegment
	pos := entryStart
	for _, s := range sorted {
		if s.OverlapStart > pos {
			segs = append(segs, timeSegment{pos, s.OverlapStart, s.OverlapStart - pos})
		}
		if s.OverlapEnd > pos {
			pos = s.OverlapEnd
		}
	}
	if pos < entryEnd {
		segs = append(segs, timeSegment{pos, entryEnd, entryEnd - pos})
	}
	return segs
}

// entrySlotBreakdown splits an entry's minutes into standard and per-slot buckets.
// Returns standard (non-slot) minutes and a slice of per-slot breakdowns.
// Returns the full entry minutes as standard when no start/end time is set.
func entrySlotBreakdown(entry models.TimeEntry, info projectContractInfo) (int, []slotMinutes) {
	if entry.StartTime == nil || entry.EndTime == nil {
		return entry.Minutes, nil
	}
	startMins := parseHHMM(*entry.StartTime)
	endMins := parseHHMM(*entry.EndTime)
	if startMins < 0 || endMins < 0 {
		return entry.Minutes, nil
	}
	entryInts := entryTimelineIntervals(startMins, endMins)
	if len(entryInts) == 0 {
		return entry.Minutes, nil
	}

	var slots []slotMinutes
	slotCovered := 0
	for _, slot := range info.TimeSlots {
		var overlaps []timelineInterval
		for _, slotInt := range slotTimelineIntervals(slot, entry.Date) {
			for _, entryInt := range entryInts {
				if ov, ok := intervalOverlap(entryInt, slotInt); ok {
					overlaps = append(overlaps, ov)
				}
			}
		}
		for _, ov := range mergeTimelineIntervals(overlaps) {
			minutes := ov.end - ov.start
			slots = append(slots, slotMinutes{
				Label:                slot.Label,
				Minutes:              minutes,
				OverlapStart:         ov.start,
				OverlapEnd:           ov.end,
				HourlyRate:           slot.HourlyRate,
				MultiplicationFactor: slot.MultiplicationFactor,
			})
			slotCovered += minutes
		}
	}

	standard := entry.Minutes - slotCovered
	if standard < 0 {
		standard = 0
	}
	return standard, slots
}

// entrySlotCost computes the cost for an entry using time-slot-aware rates.
func entrySlotCost(entry models.TimeEntry, info projectContractInfo) (float64, string) {
	if info.BaseRate == nil && len(info.TimeSlots) == 0 {
		return 0, ""
	}
	standard, slots := entrySlotBreakdown(entry, info)
	total := 0.0
	if info.BaseRate != nil && standard > 0 {
		total += float64(standard) / 60.0 * *info.BaseRate
	}
	for _, s := range slots {
		switch {
		case s.HourlyRate != nil:
			total += float64(s.Minutes) / 60.0 * *s.HourlyRate
		case s.MultiplicationFactor != nil && info.BaseRate != nil:
			total += float64(s.Minutes) / 60.0 * *info.BaseRate * *s.MultiplicationFactor
		case info.BaseRate != nil:
			total += float64(s.Minutes) / 60.0 * *info.BaseRate
		}
	}
	return total, info.Currency
}

// entryCostSegment is one billable slice of a time entry (standard rate or a matching slot).
type entryCostSegment struct {
	StartMins int
	TimeRange string
	Label     string
	Minutes   int
	Cost      float64
	Currency  string
}

// entryCostSegmentRate returns the effective hourly rate for a segment.
func entryCostSegmentRate(seg entryCostSegment) float64 {
	if seg.Minutes <= 0 || seg.Cost <= 0 {
		return 0
	}
	return seg.Cost / (float64(seg.Minutes) / 60.0)
}

// entryCostSegments splits an entry into standard and slot segments for export breakdowns.
// Returns nil when the entry has no start/end time or the contract has no time slots.
func entryCostSegments(entry models.TimeEntry, info projectContractInfo, standardLabel string) []entryCostSegment {
	if entry.StartTime == nil || entry.EndTime == nil || len(info.TimeSlots) == 0 {
		return nil
	}
	startMins := parseHHMM(*entry.StartTime)
	endMins := parseHHMM(*entry.EndTime)
	if startMins < 0 || endMins < 0 || startMins == endMins {
		return nil
	}

	_, slotSegs := entrySlotBreakdown(entry, info)
	stdSegs := standardTimeSegments(startMins, entryTimelineEnd(startMins, endMins), slotSegs)
	if len(stdSegs) == 0 && len(slotSegs) == 0 {
		return nil
	}

	var out []entryCostSegment
	for _, seg := range stdSegs {
		cost := 0.0
		if info.BaseRate != nil {
			cost = float64(seg.Minutes) / 60.0 * *info.BaseRate
		}
		out = append(out, entryCostSegment{
			StartMins: seg.Start,
			TimeRange: fmtWallClockPDF(seg.Start) + "–" + fmtWallClockPDF(seg.End),
			Label:     standardLabel,
			Minutes:   seg.Minutes,
			Cost:      cost,
			Currency:  info.Currency,
		})
	}
	for _, s := range slotSegs {
		cost := 0.0
		switch {
		case s.HourlyRate != nil:
			cost = float64(s.Minutes) / 60.0 * *s.HourlyRate
		case s.MultiplicationFactor != nil && info.BaseRate != nil:
			cost = float64(s.Minutes) / 60.0 * *info.BaseRate * *s.MultiplicationFactor
		case info.BaseRate != nil:
			cost = float64(s.Minutes) / 60.0 * *info.BaseRate
		}
		lbl := s.Label
		if lbl == "" {
			lbl = "—"
		}
		out = append(out, entryCostSegment{
			StartMins: s.OverlapStart,
			TimeRange: fmtWallClockPDF(s.OverlapStart) + "–" + fmtWallClockPDF(s.OverlapEnd),
			Label:     lbl,
			Minutes:   s.Minutes,
			Cost:      cost,
			Currency:  info.Currency,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StartMins < out[j].StartMins
	})
	return out
}
