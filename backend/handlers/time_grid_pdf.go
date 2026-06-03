package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// ── Color palette for grid PDFs ───────────────────────────────────────────────

var (
	gClrPrimary    = rgb{99, 102, 241}
	gClrText       = rgb{30, 30, 46}
	gClrMuted      = rgb{100, 116, 139}
	gClrAlt        = rgb{245, 247, 250}
	gClrHdrFill    = rgb{238, 242, 248}
	gClrQtrFill    = rgb{224, 231, 255}
	gClrTotFill    = rgb{199, 210, 254}
	gClrWhite      = rgb{255, 255, 255}
	gClrHoliday    = rgb{255, 247, 205} // light amber — holiday cells
	gClrWeekend    = rgb{232, 233, 240} // light lavender-gray — weekend data cells
	gClrWeekendHdr = rgb{218, 220, 232} // slightly darker — weekend column headers
	gClrTimeRange  = rgb{79, 70, 229}   // indigo — start/end times in week grid cells
	gClrDanger     = rgb{229, 62, 62}   // red — undeclarable time
)

// ── Landscape A4 page geometry (mm) ──────────────────────────────────────────

const (
	gPageW  = 297.0
	gPageH  = 210.0
	gMargin = 10.0
	gBodyW  = gPageW - 2*gMargin
)

// ── gridEntry holds one row of grid data ─────────────────────────────────────

type gridEntry struct {
	label    string
	cells    []int  // minutes per column (len == numCols)
	holidays []bool // true if column contains at least one is_holiday entry
	undecl   []int  // undeclarable minutes per column
}

// weekGridCell holds per-day data for the week PDF including optional time ranges.
type weekGridCell struct {
	minutes      int
	undeclMins   int
	holiday      bool
	singleRange  string // same-day start–end (e.g. "09:00–17:00")
	eveningStart string // overnight shift begins (e.g. "19:00", ends 23:59)
	morningEnd   string // overnight shift ends (e.g. "07:00", started 00:00)
}

type weekGridRow struct {
	label string
	cells []weekGridCell
}

// weekCellSpan marks a multi-day time range rendered as start/end labels with a connector line.
type weekCellSpan struct {
	startCol, endCol   int
	startTime, endTime string
}

// ── gridFmt formats minutes as decimal hours, or blank for zero ──────────────

func gridFmt(minutes int) string {
	if minutes == 0 {
		return ""
	}
	return fmtDecimalH(minutes)
}

// entryUndeclMins returns the undeclarable portion of an entry's minutes.
func entryUndeclMins(e models.TimeEntry) int {
	if e.Project == nil || e.Project.UndeclarableMinutes <= 0 {
		return 0
	}
	if e.Minutes < e.Project.UndeclarableMinutes {
		return e.Minutes
	}
	return e.Project.UndeclarableMinutes
}

// ── Data fetching helper ──────────────────────────────────────────────────────

// fetchGridEntries queries time entries for targetUserID (0 = all users) in [from, to).
func fetchGridEntries(targetUserID uint, from, to time.Time) []models.TimeEntry {
	q := database.DB.Preload("Customer").Preload("Project").
		Where("date >= ? AND date < ?", from, to).
		Order("date, id")
	if targetUserID > 0 {
		q = q.Where("user_id = ?", targetUserID)
	}
	var entries []models.TimeEntry
	q.Find(&entries)
	return entries
}

// ── Row building helper ───────────────────────────────────────────────────────

// buildGridRows groups entries by label and accumulates minutes per column.
func buildGridRows(entries []models.TimeEntry, dayFn func(models.TimeEntry) int, numCols int) []gridEntry {
	type key = string
	order := []key{}
	rows := map[key]*gridEntry{}

	for _, e := range entries {
		col := dayFn(e)
		if col < 0 || col >= numCols {
			continue
		}
		var label string
		if e.Customer != nil && e.Customer.Name != "" {
			if e.Project != nil {
				label = e.Customer.Name + " / " + e.Project.Name
			} else {
				label = e.Customer.Name
			}
		} else if e.Project != nil {
			label = e.Project.Name
		} else if e.Description != "" {
			label = e.Description
		} else {
			label = "—"
		}

		if _, exists := rows[label]; !exists {
			rows[label] = &gridEntry{
				label:    label,
				cells:    make([]int, numCols),
				holidays: make([]bool, numCols),
				undecl:   make([]int, numCols),
			}
			order = append(order, label)
		}
		rows[label].cells[col] += e.Minutes
		rows[label].undecl[col] += entryUndeclMins(e)
		if e.IsHoliday {
			rows[label].holidays[col] = true
		}
	}

	// Sort alphabetically.
	sort.Strings(order)
	result := make([]gridEntry, 0, len(order))
	for _, k := range order {
		result = append(result, *rows[k])
	}
	return result
}

func timeEntryRowKey(e models.TimeEntry) string {
	custKey, projKey := "", ""
	if e.CustomerID != nil {
		custKey = strconv.FormatUint(uint64(*e.CustomerID), 10)
	}
	if e.ProjectID != nil {
		projKey = strconv.FormatUint(uint64(*e.ProjectID), 10)
	}
	return custKey + "|" + projKey + "|" + e.Description
}

func timeEntryRowLabel(e models.TimeEntry) string {
	if e.Customer != nil && e.Customer.Name != "" {
		if e.Project != nil {
			return e.Customer.Name + " / " + e.Project.Name
		}
		return e.Customer.Name
	}
	if e.Project != nil {
		return e.Project.Name
	}
	if e.Description != "" {
		return e.Description
	}
	return "—"
}

func loadWeekGridRowOrder(userID uint, weekStart time.Time) []string {
	if userID == 0 {
		return nil
	}
	year, week := weekStart.ISOWeek()
	var weekOrder models.TimeEntryWeekRowOrder
	if err := database.DB.Where("user_id = ? AND year = ? AND week = ?", userID, year, week).First(&weekOrder).Error; err != nil {
		return nil
	}
	var keys []string
	if weekOrder.OrderedKeys == "" || json.Unmarshal([]byte(weekOrder.OrderedKeys), &keys) != nil || keys == nil {
		return nil
	}
	return keys
}

func buildWeekGridRows(entries []models.TimeEntry, dayFn func(models.TimeEntry) int, savedOrder []string) []weekGridRow {
	type key = string
	order := []key{}
	rows := map[key]*weekGridRow{}

	for _, e := range entries {
		col := dayFn(e)
		if col < 0 || col >= 7 {
			continue
		}
		k := timeEntryRowKey(e)
		if _, exists := rows[k]; !exists {
			rows[k] = &weekGridRow{
				label: timeEntryRowLabel(e),
				cells: make([]weekGridCell, 7),
			}
			order = append(order, k)
		}
		cell := &rows[k].cells[col]
		cell.minutes += e.Minutes
		cell.undeclMins += entryUndeclMins(e)
		if e.IsHoliday {
			cell.holiday = true
		}
		st, et := "", ""
		if e.StartTime != nil {
			st = *e.StartTime
		}
		if e.EndTime != nil {
			et = *e.EndTime
		}
		switch {
		case st != "" && st != "00:00" && et == "23:59":
			cell.eveningStart = st
		case (st == "" || st == "00:00") && et != "" && et != "23:59":
			cell.morningEnd = et
		case st != "" && et != "":
			cell.singleRange = st + "–" + et
		}
	}

	if len(savedOrder) > 0 {
		seen := map[key]bool{}
		merged := make([]key, 0, len(savedOrder))
		for _, k := range savedOrder {
			if rows[k] != nil && !seen[k] {
				merged = append(merged, k)
				seen[k] = true
			}
		}
		for _, k := range order {
			if !seen[k] {
				merged = append(merged, k)
			}
		}
		order = merged
	} else {
		sort.Strings(order)
	}

	result := make([]weekGridRow, 0, len(order))
	for _, k := range order {
		result = append(result, *rows[k])
	}
	return result
}

func findWeekCellSpans(cells []weekGridCell) []weekCellSpan {
	var spans []weekCellSpan
	for i := 0; i < len(cells)-1; i++ {
		if cells[i].eveningStart == "" || cells[i+1].morningEnd == "" {
			continue
		}
		spans = append(spans, weekCellSpan{
			startCol:  i,
			endCol:    i + 1,
			startTime: cells[i].eveningStart,
			endTime:   cells[i+1].morningEnd,
		})
	}
	return spans
}

func weekCellLeft(col int, wLabel, wDay float64) float64 {
	return gMargin + wLabel + float64(col)*wDay
}

// ── Main handler ──────────────────────────────────────────────────────────────

// GetTimeSheetGridPDF renders the time-sheet grid as a landscape A4 PDF.
// Route: GET /api/v1/time-entries/grid/pdf
// Query params: grid=week|month|year, year, month, week, start_date, user_id, font, lang
func GetTimeSheetGridPDF(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	// Resolve requesting user (for font preference).
	var requestingUser models.User
	database.DB.First(&requestingUser, userID)

	// Resolve target user (for data).
	targetUserID := userID
	if globalRole == "admin" || requestingUser.TimeTrackingViewer {
		if s := c.Query("user_id"); s != "" {
			if id, err := strconv.ParseUint(s, 10, 64); err == nil {
				targetUserID = uint(id)
			}
		}
	}

	// Font preference: explicit param wins, then profile, then default.
	fontFamily := "FreeSans"
	if requestingUser.Font != "" {
		fontFamily = pdfFontFamily(requestingUser.Font)
	}
	if fam, ok := pdfFontFromParam(c.Query("font")); ok {
		fontFamily = fam
	}

	tr := pdfI18nFromLang(c.Query("lang"))

	// Resolve employee name.
	var employeeName string
	if targetUserID == 0 {
		employeeName = tr.AllEmployees
	} else if targetUserID == userID {
		employeeName = requestingUser.DisplayName
		if employeeName == "" {
			employeeName = requestingUser.Username
		}
	} else {
		var tu models.User
		if err := database.DB.First(&tu, targetUserID).Error; err == nil {
			employeeName = tu.DisplayName
			if employeeName == "" {
				employeeName = tu.Username
			}
		}
	}

	// Company settings.
	settings := loadAllSettings()
	companyName := settings[settingCompanyName]
	companyLogo := settings[settingCompanyLogo]
	if companyName == "" {
		companyName = "WarmDesk"
	}

	// Dispatch to the correct grid.
	gridType := c.DefaultQuery("grid", "week")

	var buf bytes.Buffer
	var filename string
	var genErr error

	switch gridType {
	case "month":
		year, month := resolveYearMonth(c)
		buf, filename, genErr = buildMonthGridPDF(fontFamily, tr, companyName, companyLogo, employeeName, year, month, targetUserID)
	case "year":
		year := resolveYear(c)
		buf, filename, genErr = buildYearGridPDF(fontFamily, tr, companyName, companyLogo, employeeName, year, targetUserID)
	default: // "week"
		start := resolveWeekStart(c)
		buf, filename, genErr = buildWeekGridPDF(fontFamily, tr, companyName, companyLogo, employeeName, start, targetUserID)
	}

	if genErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pdf generation failed"})
		return
	}

	if c.Query("base64") == "1" {
		c.JSON(http.StatusOK, gin.H{"data": base64.StdEncoding.EncodeToString(buf.Bytes())})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// ── Period resolution helpers ─────────────────────────────────────────────────

func resolveYear(c *gin.Context) int {
	if s := c.Query("year"); s != "" {
		if y, err := strconv.Atoi(s); err == nil {
			return y
		}
	}
	return time.Now().Year()
}

func resolveYearMonth(c *gin.Context) (int, int) {
	year := resolveYear(c)
	month := int(time.Now().Month())
	if s := c.Query("month"); s != "" {
		if m, err := strconv.Atoi(s); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}
	return year, month
}

func resolveWeekStart(c *gin.Context) time.Time {
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			return t
		}
	}
	now := time.Now()
	year, week := now.ISOWeek()
	if yStr := c.Query("year"); yStr != "" {
		if y, err := strconv.Atoi(yStr); err == nil {
			year = y
		}
	}
	if wStr := c.Query("week"); wStr != "" {
		if w, err := strconv.Atoi(wStr); err == nil {
			week = w
		}
	}
	return isoWeekStart(year, week)
}

// ── PDF initialisation helper ─────────────────────────────────────────────────

func newLandscapePDF(fontFamily string) *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AliasNbPages("{nb}")
	pdf.SetMargins(gMargin, gMargin, gMargin)
	pdf.SetAutoPageBreak(true, 15)

	for fam, files := range pdfFonts {
		pdf.AddUTF8FontFromBytes(fam, "", mustFont(files[0]))
		pdf.AddUTF8FontFromBytes(fam, "B", mustFont(files[1]))
	}
	return pdf
}

// drawGridDocHeader draws the two-line header used by all three grids.
// line1Left (large bold primary), line1Right (small muted), line2 (medium bold text).
// Returns the Y position just below the rule.
func drawGridDocHeader(pdf *gofpdf.Fpdf, ff, companyLogo, line1Left, line1Right, line2 string) float64 {
	// Optionally render logo.
	logoY := gMargin
	logoLoaded := false
	if rawBytes, ext, ok := resolveLogoBytes(companyLogo); ok {
		logoLoaded = renderLogoIntoPDF(pdf, rawBytes, ext, gMargin, logoY)
	}

	textX := gMargin
	if logoLoaded {
		textX = gMargin + 22
	}
	textW := gPageW - gMargin - textX

	pdf.SetXY(textX, logoY)

	// Line 1 — company/title left + employee right.
	pdf.SetFont(ff, "B", 13)
	setTxt(pdf, gClrPrimary)
	pdf.CellFormat(textW*0.7, 7, line1Left, "", 0, "L", false, 0, "")
	pdf.SetFont(ff, "", 9)
	setTxt(pdf, gClrMuted)
	pdf.CellFormat(textW*0.3, 7, line1Right, "", 1, "R", false, 0, "")
	pdf.SetX(textX)

	// Line 2 — period label.
	pdf.SetFont(ff, "B", 10)
	setTxt(pdf, gClrText)
	pdf.CellFormat(textW, 6, line2, "", 1, "L", false, 0, "")

	ruleY := pdf.GetY() + 2
	if logoLoaded && ruleY < logoY+22 {
		ruleY = logoY + 22
	}
	setDraw(pdf, gClrPrimary)
	pdf.SetLineWidth(0.4)
	pdf.Line(gMargin, ruleY, gPageW-gMargin, ruleY)
	return ruleY + 4
}

// setGridFooter registers a footer function that renders page number and period.
func setGridFooter(pdf *gofpdf.Fpdf, ff, companyName, periodLabel string, tr pdfI18n) {
	pdf.SetFooterFunc(func() {
		pdf.SetY(-10)
		pdf.SetFont(ff, "", 7)
		setTxt(pdf, gClrMuted)
		pdf.CellFormat(gBodyW/2, 5, companyName+" — "+periodLabel, "", 0, "L", false, 0, "")
		pdf.CellFormat(gBodyW/2, 5,
			fmt.Sprintf("%s %d / {nb}", tr.Page, pdf.PageNo()), "", 0, "R", false, 0, "")
	})
}

// ── Week grid ─────────────────────────────────────────────────────────────────

func weekRowHasTimes(r weekGridRow) bool {
	for _, c := range r.cells {
		if c.singleRange != "" || c.eveningStart != "" || c.morningEnd != "" {
			return true
		}
	}
	return false
}

func weekRowHasUndecl(r weekGridRow) bool {
	for _, c := range r.cells {
		if c.undeclMins > 0 {
			return true
		}
	}
	return false
}

func drawWeekGridTimeAt(pdf *gofpdf.Fpdf, ff, txt string, x, y float64) {
	if txt == "" {
		return
	}
	pdf.SetFont(ff, "", 5.5)
	setTxt(pdf, gClrTimeRange)
	tw := pdf.GetStringWidth(txt)
	pdf.Text(x-tw/2, y, txt)
}

func buildWeekGridPDF(ff string, tr pdfI18n, companyName, companyLogo, employeeName string, weekStart time.Time, targetUserID uint) (bytes.Buffer, string, error) {
	// Column widths: 74 label + 7×25 days + 28 total = 277mm (full body width).
	const (
		wLabel    = 74.0
		wDay      = 25.0
		wTotal    = 28.0
		baseRowH  = 6.5
		tallRowH  = 10.0
		timeBaseY = 1.8 // mm above cell bottom for time labels / connector
	)

	// Period label.
	_, isoWeek := weekStart.ISOWeek()
	year := weekStart.Year()
	weekEnd := weekStart.AddDate(0, 0, 6)
	periodLabel := fmt.Sprintf("%s %d  •  %s %d  (%02d-%02d – %02d-%02d)",
		tr.YearPrefix, year,
		tr.WeekLabel, isoWeek,
		weekStart.Day(), int(weekStart.Month()),
		weekEnd.Day(), int(weekEnd.Month()),
	)

	// Fetch data.
	entries := fetchGridEntries(targetUserID, weekStart, weekStart.AddDate(0, 0, 7))
	dayFn := func(e models.TimeEntry) int {
		diff := int(e.Date.Sub(weekStart).Hours()) / 24
		if diff < 0 || diff >= 7 {
			return -1
		}
		return diff
	}
	savedOrder := loadWeekGridRowOrder(targetUserID, weekStart)
	rows := buildWeekGridRows(entries, dayFn, savedOrder)

	// Column totals.
	colTotals := make([]int, 7)
	grandTotal := 0
	for _, r := range rows {
		for i, c := range r.cells {
			colTotals[i] += c.minutes
			grandTotal += c.minutes
		}
	}

	// Undeclarable totals.
	colUndecl := make([]int, 7)
	grandUndecl := 0
	for _, r := range rows {
		for ci, c := range r.cells {
			colUndecl[ci] += c.undeclMins
			grandUndecl += c.undeclMins
		}
	}
	grandDecl := grandTotal - grandUndecl
	if grandDecl < 0 {
		grandDecl = 0
	}

	pdf := newLandscapePDF(ff)
	setGridFooter(pdf, ff, companyName, periodLabel, tr)

	// Day-column headers: abbr (Mon=0..Sun=6) + day-of-month.
	dayHeaders := make([]string, 7)
	for i := 0; i < 7; i++ {
		d := weekStart.AddDate(0, 0, i)
		abbrIdx := (int(d.Weekday()) + 6) % 7 // Mon=0..Sun=6
		dayHeaders[i] = tr.DaysAbbr[abbrIdx] + "\n" + fmt.Sprintf("%d", d.Day())
	}

	// drawHeader draws the full doc header and then the column header row.
	// Returns the Y after column header row.
	line1Left := companyName
	drawHeader := func() float64 {
		y := drawGridDocHeader(pdf, ff, companyLogo, line1Left, employeeName, periodLabel)
		pdf.SetY(y)
		return drawWeekColHeader(pdf, ff, tr, dayHeaders, wLabel, wDay, wTotal, baseRowH*1.4)
	}

	pdf.AddPage()
	tableY := drawHeader()
	pdf.SetY(tableY)

	pdf.SetHeaderFunc(func() {
		drawHeader()
	})

	// Data rows.
	for i, r := range rows {
		hasTimes := weekRowHasTimes(r)
		hasUndecl := weekRowHasUndecl(r)
		rowH := baseRowH
		if hasTimes && hasUndecl {
			rowH = 13.0
		} else if hasTimes || hasUndecl {
			rowH = tallRowH
		}
		if pdf.GetY() > gPageH-gMargin-rowH*3 {
			pdf.AddPage()
			tableY = drawHeader()
			pdf.SetY(tableY)
		}

		// Pre-compute row totals before drawing.
		rowTotal := 0
		rowUndecl := 0
		for _, c := range r.cells {
			rowTotal += c.minutes
			rowUndecl += c.undeclMins
		}
		rowDeclarable := rowTotal - rowUndecl
		if rowDeclarable < 0 {
			rowDeclarable = 0
		}

		rowY := pdf.GetY()
		alt := i%2 == 1
		normalFill := gClrWhite
		if alt {
			normalFill = gClrAlt
		}
		spans := findWeekCellSpans(r.cells)

		setTxt(pdf, gClrText)
		pdf.SetFont(ff, "", 8)
		pdf.SetX(gMargin)
		setFill(pdf, normalFill)
		pdf.CellFormat(wLabel, rowH, truncate(r.label, 36), "LRB", 0, "L", true, 0, "")
		for _, c := range r.cells {
			if c.holiday {
				setFill(pdf, gClrHoliday)
			} else {
				setFill(pdf, normalFill)
			}
			txt := gridFmt(c.minutes)
			if c.holiday && c.minutes == 0 {
				txt = "•"
			}
			pdf.CellFormat(wDay, rowH, txt, "LRB", 0, "C", true, 0, "")
		}

		// Time labels and connector lines use absolute coordinates so the table
		// row layout (continuous cells) is not disturbed.
		timeY := rowY + rowH - timeBaseY
		lineY := timeY - 0.6
		for ci, c := range r.cells {
			cellX := weekCellLeft(ci, wLabel, wDay)
			if c.singleRange != "" {
				drawWeekGridTimeAt(pdf, ff, c.singleRange, cellX+wDay/2, timeY)
			}
			if c.morningEnd != "" {
				drawWeekGridTimeAt(pdf, ff, c.morningEnd, cellX+wDay*0.28, timeY)
			}
			if c.eveningStart != "" {
				drawWeekGridTimeAt(pdf, ff, c.eveningStart, cellX+wDay*0.72, timeY)
			}
		}
		setDraw(pdf, gClrTimeRange)
		pdf.SetLineWidth(0.35)
		for _, sp := range spans {
			x1 := weekCellLeft(sp.startCol, wLabel, wDay) + wDay*0.72
			x2 := weekCellLeft(sp.endCol, wLabel, wDay) + wDay*0.28
			pdf.Line(x1, lineY, x2, lineY)
		}

		// Per-cell undeclarable labels (absolute positioning, below main time value).
		if hasUndecl {
			undeclY := rowY + baseRowH + 0.8
			undeclH := 3.5
			pdf.SetFont(ff, "", 6.5)
			setTxt(pdf, gClrDanger)
			for ci, c := range r.cells {
				if c.undeclMins > 0 {
					cellX := weekCellLeft(ci, wLabel, wDay)
					pdf.SetXY(cellX, undeclY)
					pdf.CellFormat(wDay, undeclH, "−"+gridFmt(c.undeclMins), "", 0, "C", false, 0, "")
				}
			}
			if rowUndecl > 0 {
				pdf.SetXY(gMargin+wLabel+7*wDay, undeclY)
				pdf.CellFormat(wTotal, undeclH, "−"+gridFmt(rowUndecl), "", 0, "C", false, 0, "")
			}
		}

		setTxt(pdf, gClrText)
		pdf.SetXY(gMargin+wLabel+7*wDay, rowY)
		pdf.SetFont(ff, "B", 8)
		setFill(pdf, gClrTotFill)
		pdf.CellFormat(wTotal, rowH, gridFmt(rowDeclarable), "LRB", 1, "C", true, 0, "")
	}

	// Totals row.
	setFill(pdf, gClrTotFill)
	setTxt(pdf, gClrText)
	pdf.SetFont(ff, "B", 8)
	pdf.SetX(gMargin)
	pdf.CellFormat(wLabel, baseRowH, tr.Total, "1", 0, "L", true, 0, "")
	for i, m := range colTotals {
		decl := m - colUndecl[i]
		if decl < 0 {
			decl = 0
		}
		pdf.CellFormat(wDay, baseRowH, gridFmt(decl), "1", 0, "C", true, 0, "")
	}
	setFill(pdf, gClrPrimary)
	setTxt(pdf, gClrWhite)
	pdf.CellFormat(wTotal, baseRowH, gridFmt(grandDecl), "1", 1, "C", true, 0, "")

	// Undeclarable row below totals (when present).
	if grandUndecl > 0 {
		setFill(pdf, gClrTotFill)
		setTxt(pdf, gClrDanger)
		pdf.SetFont(ff, "", 7)
		pdf.SetX(gMargin)
		pdf.CellFormat(wLabel, baseRowH*0.8, "  "+tr.Undeclarable, "B", 0, "L", true, 0, "")
		for _, u := range colUndecl {
			txt := ""
			if u > 0 {
				txt = "−" + gridFmt(u)
			}
			pdf.CellFormat(wDay, baseRowH*0.8, txt, "B", 0, "C", true, 0, "")
		}
		setFill(pdf, gClrPrimary)
		pdf.CellFormat(wTotal, baseRowH*0.8, "−"+gridFmt(grandUndecl), "B", 1, "C", true, 0, "")
		setTxt(pdf, gClrText)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return bytes.Buffer{}, "", err
	}
	filename := fmt.Sprintf("timesheet-week-%d-%02d.pdf", year, isoWeek)
	return buf, filename, nil
}

// drawWeekColHeader renders the column header row and returns the Y position below it.
func drawWeekColHeader(pdf *gofpdf.Fpdf, ff string, tr pdfI18n, dayHeaders []string, wLabel, wDay, wTotal, hdrH float64) float64 {
	setFill(pdf, gClrHdrFill)
	setTxt(pdf, gClrText)
	pdf.SetFont(ff, "B", 7.5)
	pdf.SetX(gMargin)

	// Multi-line day headers need a taller row; use MultiCell approach with fixed cell height.
	// We emit each header as two lines manually.
	startY := pdf.GetY()
	lineH := hdrH / 2

	// Label column.
	pdf.CellFormat(wLabel, hdrH, tr.Customer+" / "+tr.Project, "1", 0, "L", true, 0, "")
	// Day columns.
	for _, hdr := range dayHeaders {
		parts := strings.SplitN(hdr, "\n", 2)
		x := pdf.GetX()
		y := pdf.GetY()
		pdf.CellFormat(wDay, hdrH, "", "1", 0, "C", true, 0, "")
		pdf.SetXY(x, y)
		pdf.CellFormat(wDay, lineH, parts[0], "", 0, "C", false, 0, "")
		pdf.SetXY(x, y+lineH)
		txt2 := ""
		if len(parts) > 1 {
			txt2 = parts[1]
		}
		pdf.CellFormat(wDay, lineH, txt2, "", 0, "C", false, 0, "")
		pdf.SetXY(x+wDay, y)
	}
	// Total column.
	pdf.CellFormat(wTotal, hdrH, tr.Total, "1", 1, "C", true, 0, "")
	_ = startY
	return pdf.GetY()
}

// ── Month grid ────────────────────────────────────────────────────────────────

func buildMonthGridPDF(ff string, tr pdfI18n, companyName, companyLogo, employeeName string, year, month int, targetUserID uint) (bytes.Buffer, string, error) {
	const (
		wDay   = 7.0
		wTotal = 14.0
		rowH   = 5.5
		hdrH   = 6.0
		cellFS = 5.5
	)

	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	// Label column fills all space not taken by day columns and the total column.
	wLabel := gBodyW - float64(daysInMonth)*wDay - wTotal
	// Approx char capacity at cellFS pt (FreeSans ~1.4 mm/char at 5.5 pt).
	labelMaxChars := int(wLabel / 1.4)
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := firstDay.AddDate(0, 1, 0)

	periodLabel := fmt.Sprintf("%s  %d", tr.MonthsFull[month-1], year)

	entries := fetchGridEntries(targetUserID, firstDay, nextMonth)
	dayFn := func(e models.TimeEntry) int {
		// Normalise to UTC midnight before comparing to avoid timezone edge cases
		// where a stored date near midnight appears as the wrong calendar day.
		entryDay := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, time.UTC)
		diff := int(entryDay.Sub(firstDay).Hours()) / 24
		if diff < 0 || diff >= 31 {
			return -1
		}
		return diff
	}
	rows := buildGridRows(entries, dayFn, daysInMonth)

	colTotals := make([]int, daysInMonth)
	grandTotal := 0
	for _, r := range rows {
		for i, m := range r.cells {
			colTotals[i] += m
			grandTotal += m
		}
	}

	// Undeclarable totals.
	colUndecl := make([]int, daysInMonth)
	grandUndecl := 0
	for _, r := range rows {
		for i, u := range r.undecl {
			colUndecl[i] += u
			grandUndecl += u
		}
	}
	grandDecl := grandTotal - grandUndecl
	if grandDecl < 0 {
		grandDecl = 0
	}

	pdf := newLandscapePDF(ff)
	setGridFooter(pdf, ff, companyName, periodLabel, tr)

	drawHeader := func() float64 {
		line2 := fmt.Sprintf("%s %d", tr.MonthsFull[month-1], year)
		y := drawGridDocHeader(pdf, ff, companyLogo, companyName, employeeName, line2)
		pdf.SetY(y)
		return drawMonthColHeader(pdf, ff, tr, year, month, daysInMonth, wLabel, wDay, wTotal, hdrH)
	}

	pdf.AddPage()
	tableY := drawHeader()
	pdf.SetY(tableY)
	pdf.SetHeaderFunc(func() {
		drawHeader()
	})

	// Data rows.
	for i, r := range rows {
		if pdf.GetY() > gPageH-gMargin-rowH*3 {
			pdf.AddPage()
			tableY = drawHeader()
			pdf.SetY(tableY)
		}
		alt := i%2 == 1
		normalFill := gClrWhite
		if alt {
			normalFill = gClrAlt
		}
		// Pre-compute row totals before drawing.
		rowTotal := 0
		rowUndecl := 0
		for _, m := range r.cells {
			rowTotal += m
		}
		for _, u := range r.undecl {
			rowUndecl += u
		}
		rowDeclarable := rowTotal - rowUndecl
		if rowDeclarable < 0 {
			rowDeclarable = 0
		}

		setTxt(pdf, gClrText)
		pdf.SetFont(ff, "", cellFS)
		pdf.SetX(gMargin)
		setFill(pdf, normalFill)
		pdf.CellFormat(wLabel, rowH, truncate(r.label, labelMaxChars), "LRB", 0, "L", true, 0, "")
		for d := 0; d < daysInMonth; d++ {
			m := r.cells[d]
			u := r.undecl[d]
			decl := m - u
			if decl < 0 {
				decl = 0
			}
			isHol := d < len(r.holidays) && r.holidays[d]
			isWkd := isWeekendDay(year, month, d+1)
			switch {
			case isHol:
				setFill(pdf, gClrHoliday)
			case isWkd:
				setFill(pdf, gClrWeekend)
			default:
				setFill(pdf, normalFill)
			}
			txt := gridFmt(decl)
			if isHol && m == 0 {
				txt = "•"
			}
			pdf.CellFormat(wDay, rowH, txt, "LRB", 0, "C", true, 0, "")
		}
		pdf.SetFont(ff, "B", cellFS)
		setFill(pdf, gClrTotFill)
		pdf.CellFormat(wTotal, rowH, gridFmt(rowDeclarable), "LRB", 1, "C", true, 0, "")
	}

	// Totals row.
	setTxt(pdf, gClrText)
	pdf.SetFont(ff, "B", cellFS)
	pdf.SetX(gMargin)
	setFill(pdf, gClrTotFill)
	pdf.CellFormat(wLabel, rowH, tr.Total, "1", 0, "L", true, 0, "")
	for d := 0; d < daysInMonth; d++ {
		decl := colTotals[d] - colUndecl[d]
		if decl < 0 {
			decl = 0
		}
		if isWeekendDay(year, month, d+1) {
			setFill(pdf, gClrWeekend)
		} else {
			setFill(pdf, gClrTotFill)
		}
		pdf.CellFormat(wDay, rowH, gridFmt(decl), "1", 0, "C", true, 0, "")
	}
	setFill(pdf, gClrPrimary)
	setTxt(pdf, gClrWhite)
	pdf.CellFormat(wTotal, rowH, gridFmt(grandDecl), "1", 1, "C", true, 0, "")

	// Undeclarable row.
	if grandUndecl > 0 {
		undeclRowH := rowH * 0.8
		setFill(pdf, gClrTotFill)
		setTxt(pdf, gClrDanger)
		pdf.SetFont(ff, "", cellFS-0.5)
		pdf.SetX(gMargin)
		pdf.CellFormat(wLabel, undeclRowH, "  "+tr.Undeclarable, "B", 0, "L", true, 0, "")
		for d := 0; d < daysInMonth; d++ {
			u := colUndecl[d]
			txt := ""
			if u > 0 {
				txt = "−" + gridFmt(u)
			}
			if isWeekendDay(year, month, d+1) {
				setFill(pdf, gClrWeekend)
			} else {
				setFill(pdf, gClrTotFill)
			}
			pdf.CellFormat(wDay, undeclRowH, txt, "B", 0, "C", true, 0, "")
		}
		setFill(pdf, gClrPrimary)
		pdf.CellFormat(wTotal, undeclRowH, "−"+gridFmt(grandUndecl), "B", 1, "C", true, 0, "")
		setTxt(pdf, gClrText)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return bytes.Buffer{}, "", err
	}
	filename := fmt.Sprintf("timesheet-month-%d-%02d.pdf", year, month)
	return buf, filename, nil
}

// drawMonthColHeader renders the column headers for the month grid.
func drawMonthColHeader(pdf *gofpdf.Fpdf, ff string, tr pdfI18n, year, month, daysInMonth int, wLabel, wDay, wTotal, hdrH float64) float64 {
	_ = tr
	setTxt(pdf, gClrText)
	pdf.SetFont(ff, "B", 5.5)
	pdf.SetX(gMargin)
	setFill(pdf, gClrHdrFill)
	pdf.CellFormat(wLabel, hdrH, tr.Customer+" / "+tr.Project, "1", 0, "L", true, 0, "")
	for d := 1; d <= daysInMonth; d++ {
		if isWeekendDay(year, month, d) {
			setFill(pdf, gClrWeekendHdr)
		} else {
			setFill(pdf, gClrHdrFill)
		}
		pdf.CellFormat(wDay, hdrH, fmt.Sprintf("%d", d), "1", 0, "C", true, 0, "")
	}
	setFill(pdf, gClrTotFill)
	pdf.CellFormat(wTotal, hdrH, tr.Total, "1", 1, "C", true, 0, "")
	return pdf.GetY()
}

func isWeekendDay(year, month, day int) bool {
	w := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Weekday()
	return w == time.Saturday || w == time.Sunday
}

// ── Year grid ─────────────────────────────────────────────────────────────────

// yearColDef describes a single data column in the year grid.
type yearColDef struct {
	label   string
	width   float64
	isQtr   bool
	isTot   bool
	monthOr int // 0-based month index, or -1 for non-month cols
}

func buildYearGridPDF(ff string, tr pdfI18n, companyName, companyLogo, employeeName string, year int, targetUserID uint) (bytes.Buffer, string, error) {
	// Column layout: 74 label + months+quarters + 15 total = 277mm (full body width).
	// Jan(12) Feb(12) Mar(12) Q1(11) Apr(12) May(12) Jun(12) Q2(11)
	// Jul(12) Aug(12) Sep(12) Q3(11) Oct(12) Nov(12) Dec(12) Q4(11) Total(15)
	const (
		wLabel   = 74.0
		wMonth   = 12.0
		wQtr     = 11.0
		wYrTotal = 15.0
		rowH     = 6.0
		hdrH     = 6.5
	)

	cols := []yearColDef{
		{label: "", width: wMonth, monthOr: 0},  // Jan
		{label: "", width: wMonth, monthOr: 1},  // Feb
		{label: "", width: wMonth, monthOr: 2},  // Mar
		{label: "Q1", width: wQtr, isQtr: true, monthOr: -1},
		{label: "", width: wMonth, monthOr: 3},  // Apr
		{label: "", width: wMonth, monthOr: 4},  // May
		{label: "", width: wMonth, monthOr: 5},  // Jun
		{label: "Q2", width: wQtr, isQtr: true, monthOr: -1},
		{label: "", width: wMonth, monthOr: 6},  // Jul
		{label: "", width: wMonth, monthOr: 7},  // Aug
		{label: "", width: wMonth, monthOr: 8},  // Sep
		{label: "Q3", width: wQtr, isQtr: true, monthOr: -1},
		{label: "", width: wMonth, monthOr: 9},  // Oct
		{label: "", width: wMonth, monthOr: 10}, // Nov
		{label: "", width: wMonth, monthOr: 11}, // Dec
		{label: "Q4", width: wQtr, isQtr: true, monthOr: -1},
		{label: "", width: wYrTotal, isTot: true, monthOr: -1},
	}
	// Fill month labels from translation.
	for i := range cols {
		if cols[i].monthOr >= 0 {
			cols[i].label = tr.MonthsAbbr[cols[i].monthOr]
		}
	}
	// Last column = Total.
	cols[len(cols)-1].label = tr.Total

	// Fetch all entries for the year.
	from := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	entries := fetchGridEntries(targetUserID, from, to)

	// Build month-based grid (12 month columns).
	monthFn := func(e models.TimeEntry) int {
		return int(e.Date.Month()) - 1
	}
	monthRows := buildGridRows(entries, monthFn, 12)

	// Convert to full column set (months + quarters + total).
	toFullRow := func(monthCells []int) []int {
		out := make([]int, len(cols))
		for i, c := range cols {
			if c.monthOr >= 0 && c.monthOr < 12 {
				out[i] = monthCells[c.monthOr]
			}
		}
		// Quarter totals.
		qtrMonths := [4][3]int{{0, 1, 2}, {4, 5, 6}, {8, 9, 10}, {12, 13, 14}}
		qtrCols := []int{3, 7, 11, 15}
		for qi, qc := range qtrCols {
			for _, mc := range qtrMonths[qi] {
				out[qc] += out[mc]
			}
		}
		// Year total.
		for m := 0; m < 12; m++ {
			out[len(cols)-1] += monthCells[m]
		}
		return out
	}

	// toFullUndecl converts 12 monthly undeclarable values to the full column set.
	toFullUndecl := func(monthUndecl []int) []int {
		out := make([]int, len(cols))
		for i, c := range cols {
			if c.monthOr >= 0 && c.monthOr < 12 {
				out[i] = monthUndecl[c.monthOr]
			}
		}
		// Quarter totals.
		qtrMonths := [4][3]int{{0, 1, 2}, {4, 5, 6}, {8, 9, 10}, {12, 13, 14}}
		qtrCols := []int{3, 7, 11, 15}
		for qi, qc := range qtrCols {
			for _, mc := range qtrMonths[qi] {
				out[qc] += out[mc]
			}
		}
		// Year total.
		for m := 0; m < 12; m++ {
			out[len(cols)-1] += monthUndecl[m]
		}
		return out
	}

	// Build full rows.
	type fullRow struct {
		label  string
		cells  []int
		undecl []int
	}
	var fullRows []fullRow
	for _, r := range monthRows {
		fullRows = append(fullRows, fullRow{
			label:  r.label,
			cells:  toFullRow(r.cells),
			undecl: toFullUndecl(r.undecl),
		})
	}

	// Column totals.
	colTotals := make([]int, len(cols))
	for _, r := range fullRows {
		for i, m := range r.cells {
			colTotals[i] += m
		}
	}
	// Recalculate quarter and year totals from raw month totals.
	qtrMonthsIdx := [4][3]int{{0, 1, 2}, {4, 5, 6}, {8, 9, 10}, {12, 13, 14}}
	qtrColsIdx := []int{3, 7, 11, 15}
	for qi, qc := range qtrColsIdx {
		colTotals[qc] = 0
		for _, mc := range qtrMonthsIdx[qi] {
			colTotals[qc] += colTotals[mc]
		}
	}
	colTotals[len(cols)-1] = 0
	for i, c := range cols {
		if c.monthOr >= 0 {
			colTotals[len(cols)-1] += colTotals[i]
		}
	}

	// Undeclarable column totals.
	colUndecl := make([]int, len(cols))
	for _, r := range fullRows {
		for i, u := range r.undecl {
			colUndecl[i] += u
		}
	}
	grandUndecl := colUndecl[len(cols)-1] // year total column

	periodLabel := fmt.Sprintf("%s %d", tr.YearPrefix, year)
	printDate := time.Now().Format("02-01-2006")

	line1Left := fmt.Sprintf("Time registration year overview  %d", year)
	line2 := fmt.Sprintf("Print date: %s", printDate)

	pdf := newLandscapePDF(ff)
	setGridFooter(pdf, ff, companyName, periodLabel, tr)

	drawHeader := func() float64 {
		y := drawGridDocHeader(pdf, ff, companyLogo, line1Left, employeeName, line2)
		pdf.SetY(y)
		return drawYearColHeader(pdf, ff, cols, wLabel, hdrH)
	}

	pdf.AddPage()
	tableY := drawHeader()
	pdf.SetY(tableY)
	pdf.SetHeaderFunc(func() {
		drawHeader()
	})

	// Data rows.
	for i, r := range fullRows {
		if pdf.GetY() > gPageH-gMargin-rowH*3 {
			pdf.AddPage()
			tableY = drawHeader()
			pdf.SetY(tableY)
		}
		alt := i%2 == 1
		if alt {
			setFill(pdf, gClrAlt)
		} else {
			setFill(pdf, gClrWhite)
		}
		setTxt(pdf, gClrText)
		pdf.SetFont(ff, "", 7)
		pdf.SetX(gMargin)
		pdf.CellFormat(wLabel, rowH, truncate(r.label, 42), "LRB", 0, "L", alt, 0, "")
		for ci, c := range cols {
			val := r.cells[ci]
			u := r.undecl[ci]
			decl := val - u
			if decl < 0 {
				decl = 0
			}
			txt := gridFmt(decl)
			if c.isQtr {
				setFill(pdf, gClrQtrFill)
				pdf.SetFont(ff, "B", 7)
				pdf.CellFormat(c.width, rowH, txt, "LRB", 0, "C", true, 0, "")
				if alt {
					setFill(pdf, gClrAlt)
				} else {
					setFill(pdf, gClrWhite)
				}
				pdf.SetFont(ff, "", 7)
			} else if c.isTot {
				setFill(pdf, gClrTotFill)
				pdf.SetFont(ff, "B", 7)
				pdf.CellFormat(c.width, rowH, txt, "LRB", 1, "C", true, 0, "")
				if alt {
					setFill(pdf, gClrAlt)
				} else {
					setFill(pdf, gClrWhite)
				}
				pdf.SetFont(ff, "", 7)
			} else {
				pdf.CellFormat(c.width, rowH, txt, "LRB", 0, "C", alt, 0, "")
			}
		}
	}

	// Totals row.
	setFill(pdf, gClrTotFill)
	setTxt(pdf, gClrText)
	pdf.SetFont(ff, "B", 7.5)
	pdf.SetX(gMargin)
	pdf.CellFormat(wLabel, rowH, tr.Total, "1", 0, "L", true, 0, "")
	for ci, c := range cols {
		val := colTotals[ci]
		u := colUndecl[ci]
		decl := val - u
		if decl < 0 {
			decl = 0
		}
		txt := gridFmt(decl)
		if c.isQtr {
			setFill(pdf, gClrQtrFill)
		} else if c.isTot {
			setFill(pdf, gClrPrimary)
			setTxt(pdf, gClrWhite)
		} else {
			setFill(pdf, gClrTotFill)
		}
		nl := 0
		if c.isTot {
			nl = 1
		}
		pdf.CellFormat(c.width, rowH, txt, "1", nl, "C", true, 0, "")
	}

	// Undeclarable row below totals (when present).
	if grandUndecl > 0 {
		undeclRowH := rowH * 0.8
		setFill(pdf, gClrTotFill)
		setTxt(pdf, gClrDanger)
		pdf.SetFont(ff, "", 6.5)
		pdf.SetX(gMargin)
		pdf.CellFormat(wLabel, undeclRowH, "  "+tr.Undeclarable, "B", 0, "L", true, 0, "")
		for ci, c := range cols {
			u := colUndecl[ci]
			txt := ""
			if u > 0 {
				txt = "−" + gridFmt(u)
			}
			if c.isQtr {
				setFill(pdf, gClrQtrFill)
			} else if c.isTot {
				setFill(pdf, gClrPrimary)
			} else {
				setFill(pdf, gClrTotFill)
			}
			nl := 0
			if c.isTot {
				nl = 1
			}
			pdf.CellFormat(c.width, undeclRowH, txt, "B", nl, "C", true, 0, "")
		}
		setTxt(pdf, gClrText)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return bytes.Buffer{}, "", err
	}
	filename := fmt.Sprintf("timesheet-year-%d.pdf", year)
	return buf, filename, nil
}

// drawYearColHeader renders the column header row for the year grid.
func drawYearColHeader(pdf *gofpdf.Fpdf, ff string, cols []yearColDef, wLabel, hdrH float64) float64 {
	pdf.SetFont(ff, "B", 7)
	setTxt(pdf, gClrText)
	pdf.SetX(gMargin)
	setFill(pdf, gClrHdrFill)
	pdf.CellFormat(wLabel, hdrH, "", "1", 0, "L", true, 0, "")
	for i, c := range cols {
		if c.isQtr {
			setFill(pdf, gClrQtrFill)
		} else if c.isTot {
			setFill(pdf, gClrTotFill)
		} else {
			setFill(pdf, gClrHdrFill)
		}
		ln := 0
		if i == len(cols)-1 {
			ln = 1
		}
		pdf.CellFormat(c.width, hdrH, c.label, "1", ln, "C", true, 0, "")
	}
	return pdf.GetY()
}

