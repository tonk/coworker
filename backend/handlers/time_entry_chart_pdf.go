package handlers

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// chartMaxSlices matches the frontend's REPORT_CHART_MAX_SLICES — the top N
// activities are drawn individually, the rest fold into a single "Other" slice.
const chartMaxSlices = 7

// pdfChartPalette is the same fixed categorical hue order used by the
// frontend's --chart-cat-1..7 CSS variables, so a given activity's color is
// the same whether viewed on screen or in the exported PDF.
var pdfChartPalette = []rgb{
	{42, 120, 214},  // blue
	{27, 175, 122},  // aqua
	{237, 161, 0},   // yellow
	{0, 131, 0},     // green
	{74, 58, 167},   // violet
	{227, 73, 72},   // red
	{232, 123, 164}, // magenta
}

var pdfChartOtherColor = rgb{148, 163, 184}

func pdfChartColorAt(i int, isOther bool) rgb {
	if isOther {
		return pdfChartOtherColor
	}
	return pdfChartPalette[i%len(pdfChartPalette)]
}

// chartSlice is one bar/pie slice — a single activity's total.
type chartSlice struct {
	Label   string
	Minutes int
	IsOther bool
}

// activityBreakdown aggregates entries by activity (description), sorted
// descending, folding everything past chartMaxSlices into a single "Other"
// slice. Mirrors the frontend's reportActivityBreakdown computed.
func activityBreakdown(entries []models.TimeEntry, basis, noActivityLabel, otherLabel string) []chartSlice {
	totals := map[string]int{}
	var order []string
	for _, e := range entries {
		label := strings.TrimSpace(e.Description)
		if label == "" {
			label = noActivityLabel
		}
		minutes := e.Minutes
		switch basis {
		case "total":
			// minutes already set to e.Minutes
		case "undeclarable":
			minutes = pdfEntryUndecl(e)
		default:
			minutes = pdfEntryDeclarable(e)
		}
		if minutes <= 0 {
			continue
		}
		if _, ok := totals[label]; !ok {
			order = append(order, label)
		}
		totals[label] += minutes
	}

	list := make([]chartSlice, len(order))
	for i, l := range order {
		list[i] = chartSlice{Label: l, Minutes: totals[l]}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Minutes > list[j].Minutes })

	if len(list) <= chartMaxSlices+1 {
		return list
	}
	out := append([]chartSlice{}, list[:chartMaxSlices]...)
	otherSum := 0
	for _, s := range list[chartMaxSlices:] {
		otherSum += s.Minutes
	}
	out = append(out, chartSlice{Label: otherLabel, Minutes: otherSum, IsOther: true})
	return out
}

// chartSeries is one activity's values across every period, for the stacked chart.
type chartSeries struct {
	Label   string
	IsOther bool
	Data    []int // minutes per period, aligned with the periods slice
}

// stackedBreakdown buckets entries into period groups (always time-based —
// year→month, month→week, week→day, regardless of the report's own
// group_by) and returns one series per top activity plus an "Other" catch-all.
// Mirrors the frontend's reportStackedBreakdown computed.
func stackedBreakdown(groups []timeEntryGroup, basis, noActivityLabel, otherLabel string) ([]string, []chartSeries) {
	periods := make([]string, len(groups))
	for i, g := range groups {
		periods[i] = g.Label
	}

	grand := map[string]int{}
	var grandOrder []string
	perPeriod := make([]map[string]int, len(groups))
	for gi, g := range groups {
		m := map[string]int{}
		for _, e := range g.Entries {
			label := strings.TrimSpace(e.Description)
			if label == "" {
				label = noActivityLabel
			}
			minutes := e.Minutes
			switch basis {
			case "total":
				// minutes already set to e.Minutes
			case "undeclarable":
				minutes = pdfEntryUndecl(e)
			default:
				minutes = pdfEntryDeclarable(e)
			}
			if minutes <= 0 {
				continue
			}
			if _, ok := grand[label]; !ok {
				grandOrder = append(grandOrder, label)
			}
			grand[label] += minutes
			m[label] += minutes
		}
		perPeriod[gi] = m
	}

	ranked := make([]chartSlice, len(grandOrder))
	for i, l := range grandOrder {
		ranked[i] = chartSlice{Label: l, Minutes: grand[l]}
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].Minutes > ranked[j].Minutes })

	top := map[string]bool{}
	n := chartMaxSlices
	if n > len(ranked) {
		n = len(ranked)
	}
	seriesLabels := make([]string, 0, n+1)
	for i := 0; i < n; i++ {
		seriesLabels = append(seriesLabels, ranked[i].Label)
		top[ranked[i].Label] = true
	}
	hasOther := len(ranked) > chartMaxSlices
	if hasOther {
		seriesLabels = append(seriesLabels, otherLabel)
	}

	series := make([]chartSeries, len(seriesLabels))
	for si, label := range seriesLabels {
		isOther := hasOther && si == len(seriesLabels)-1
		data := make([]int, len(groups))
		for gi := range groups {
			if isOther {
				sum := 0
				for l, m := range perPeriod[gi] {
					if !top[l] {
						sum += m
					}
				}
				data[gi] = sum
			} else {
				data[gi] = perPeriod[gi][label]
			}
		}
		series[si] = chartSeries{Label: label, IsOther: isOther, Data: data}
	}
	return periods, series
}

// GetTimeEntryReportChartPDF renders the report's activity chart (bar, pie,
// or stacked-by-period) as a standalone PDF. Admin and timetracking viewer
// roles may pass ?user_id= to render another user's chart.
func GetTimeEntryReportChartPDF(c *gin.Context) {
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

	period, from, to, periodLabel, entries, status, msg := resolveReportEntries(c, targetUserID)
	if status != 0 {
		c.JSON(status, gin.H{"error": msg})
		return
	}

	var requestingUser models.User
	fontFamily := "FreeSans"
	if err := database.DB.First(&requestingUser, userID).Error; err == nil {
		fontFamily = pdfFontFamily(requestingUser.Font)
	}
	if fam, ok := pdfFontFromParam(c.Query("font")); ok {
		fontFamily = fam
	}
	tr := pdfI18nFromLang(c.Query("lang"))

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

	chartType := c.DefaultQuery("chart_type", "bar")
	basis := c.DefaultQuery("chart_basis", "declarable")

	settings := loadAllSettings()
	companyName := settings["company_name"]
	companyLogo := settings["company_logo"]

	// ── Build PDF ─────────────────────────────────────────────────────────────
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AliasNbPages("{nb}")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(true, 20)

	for fam, files := range pdfFonts {
		pdf.AddUTF8FontFromBytes(fam, "", mustFont(files[0]))
		pdf.AddUTF8FontFromBytes(fam, "B", mustFont(files[1]))
	}

	company := companyName
	if company == "" {
		company = "WarmDesk"
	}
	title := tr.ChartTitle + " — " + periodLabel
	pdf.SetTitle(title, true)
	pdf.SetAuthor(employeeName, true)
	pdf.SetSubject(company+" — "+title, true)
	pdf.SetCreator(company, true)

	showPageNumbers := c.DefaultQuery("show_page_numbers", "1") != "0"
	pdf.SetFooterFunc(func() {
		if !showPageNumbers {
			return
		}
		pdf.SetY(-12)
		pdf.SetFont(fontFamily, "", 8)
		setTxt(pdf, clrMuted)
		label := "WarmDesk — " + periodLabel
		if companyName != "" {
			label = companyName + " — " + periodLabel
		}
		pdf.CellFormat(pdfBodyW/2, 5, label, "", 0, "L", false, 0, "")
		pdf.CellFormat(pdfBodyW/2, 5,
			fmt.Sprintf("%s %d / {nb}", tr.Page, pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	var cachedLogoBytes []byte
	var cachedLogoExt string
	var cachedLogoOK bool
	if b, ext, ok := resolveLogoBytes(companyLogo); ok {
		cachedLogoBytes, cachedLogoExt, cachedLogoOK = b, ext, true
	}

	pdf.AddPage()

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
	if companyName != "" {
		pdf.SetFont(fontFamily, "B", 13)
		setTxt(pdf, clrPrimary)
		pdf.CellFormat(textW, 7, companyName, "", 2, "L", false, 0, "")
		pdf.SetX(textX)
	}
	pdf.SetFont(fontFamily, "B", 16)
	setTxt(pdf, clrText)
	pdf.CellFormat(textW, 9, tr.ChartTitle, "", 2, "L", false, 0, "")
	pdf.SetX(textX)
	pdf.SetFont(fontFamily, "", 10)
	setTxt(pdf, clrMuted)
	pdf.CellFormat(textW, 6, pdfTranslateLabel(periodLabel, tr), "", 2, "L", false, 0, "")
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
	pdf.SetY(ruleY + 8)

	switch chartType {
	case "pie":
		slices := activityBreakdown(entries, basis, tr.Activity, tr.ChartOther)
		drawPieChart(pdf, fontFamily, tr, slices)
	case "stacked":
		groups := buildGroups(period, from, to, entries)
		periods, series := stackedBreakdown(groups, basis, tr.Activity, tr.ChartOther)
		drawStackedBarChart(pdf, fontFamily, tr, periods, series)
	default:
		slices := activityBreakdown(entries, basis, tr.Activity, tr.ChartOther)
		drawBarChart(pdf, fontFamily, tr, slices)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pdf generation failed"})
		return
	}

	filename := "time-chart-" + periodLabel + ".pdf"
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// drawBarChart renders one horizontal bar per activity, longest first.
func drawBarChart(pdf *gofpdf.Fpdf, fontFamily string, tr pdfI18n, slices []chartSlice) {
	if len(slices) == 0 {
		drawChartEmptyState(pdf, fontFamily, tr)
		return
	}
	const labelW = 55.0
	const valueW = 20.0
	const rowH = 8.0
	barMaxW := pdfBodyW - labelW - valueW

	maxMinutes := 0
	for _, s := range slices {
		if s.Minutes > maxMinutes {
			maxMinutes = s.Minutes
		}
	}
	if maxMinutes == 0 {
		maxMinutes = 1
	}

	for i, s := range slices {
		y := pdf.GetY()
		pdf.SetFont(fontFamily, "", 8.5)
		setTxt(pdf, clrText)
		pdf.SetXY(pdfMargin, y+1)
		pdf.CellFormat(labelW-2, rowH-2, truncate(s.Label, 26), "", 0, "L", false, 0, "")

		barW := barMaxW * float64(s.Minutes) / float64(maxMinutes)
		if barW < 1 && s.Minutes > 0 {
			barW = 1
		}
		clr := pdfChartColorAt(i, s.IsOther)
		setFill(pdf, clr)
		pdf.Rect(pdfMargin+labelW, y+1, barW, rowH-3, "F")

		pdf.SetXY(pdfMargin+labelW+barMaxW, y+1)
		pdf.SetFont(fontFamily, "B", 8.5)
		setTxt(pdf, clrMuted)
		pdf.CellFormat(valueW, rowH-2, fmtDecimalH(s.Minutes), "", 0, "R", false, 0, "")

		pdf.SetY(y + rowH)
	}
}

// drawPieChart renders a pie (as a polygon-approximated sector fan) with a
// color-swatch legend to the right, matching the on-screen legend layout.
func drawPieChart(pdf *gofpdf.Fpdf, fontFamily string, tr pdfI18n, slices []chartSlice) {
	if len(slices) == 0 {
		drawChartEmptyState(pdf, fontFamily, tr)
		return
	}
	total := 0
	for _, s := range slices {
		total += s.Minutes
	}
	if total == 0 {
		drawChartEmptyState(pdf, fontFamily, tr)
		return
	}

	const radius = 32.0
	cx := pdfMargin + radius + 5
	cy := pdf.GetY() + radius

	startAngle := -90.0 // 12 o'clock, sweeping clockwise
	for i, s := range slices {
		sweep := 360.0 * float64(s.Minutes) / float64(total)
		drawPieSlice(pdf, cx, cy, radius, startAngle, sweep, pdfChartColorAt(i, s.IsOther))
		startAngle += sweep
	}

	// Legend to the right of the pie.
	legendX := cx + radius + 12
	legendY := cy - radius
	pdf.SetFont(fontFamily, "", 8.5)
	for i, s := range slices {
		swY := legendY + float64(i)*6.5
		clr := pdfChartColorAt(i, s.IsOther)
		setFill(pdf, clr)
		setDraw(pdf, clr)
		pdf.Rect(legendX, swY, 4, 4, "FD")
		setTxt(pdf, clrText)
		pdf.SetXY(legendX+6, swY-1)
		pct := 100 * float64(s.Minutes) / float64(total)
		pdf.CellFormat(pdfPageW-pdfMargin-(legendX+6), 6, fmt.Sprintf("%s (%.0f%%)", truncate(s.Label, 28), pct), "", 0, "L", false, 0, "")
	}

	pdf.SetY(cy + radius + 10)
}

// drawPieSlice fills one pie sector as a polygon fan approximating the arc.
// startDeg is the sector's start angle in degrees (0 = 3 o'clock, clockwise),
// sweepDeg is how many degrees the sector spans.
func drawPieSlice(pdf *gofpdf.Fpdf, cx, cy, radius, startDeg, sweepDeg float64, clr rgb) {
	if sweepDeg <= 0 {
		return
	}
	const maxSegments = 60
	segments := int(math.Ceil(sweepDeg / 6.0))
	if segments < 1 {
		segments = 1
	}
	if segments > maxSegments {
		segments = maxSegments
	}
	points := make([]gofpdf.PointType, 0, segments+2)
	points = append(points, gofpdf.PointType{X: cx, Y: cy})
	for i := 0; i <= segments; i++ {
		deg := startDeg + sweepDeg*float64(i)/float64(segments)
		rad := deg * math.Pi / 180
		points = append(points, gofpdf.PointType{
			X: cx + radius*math.Cos(rad),
			Y: cy + radius*math.Sin(rad),
		})
	}
	setFill(pdf, clr)
	setDraw(pdf, rgb{255, 255, 255})
	pdf.SetLineWidth(0.3)
	pdf.Polygon(points, "FD")
}

// drawStackedBarChart renders one horizontal bar per period, each split into
// colored segments by activity, with a color-swatch legend above the bars.
func drawStackedBarChart(pdf *gofpdf.Fpdf, fontFamily string, tr pdfI18n, periods []string, series []chartSeries) {
	if len(periods) == 0 || len(series) == 0 {
		drawChartEmptyState(pdf, fontFamily, tr)
		return
	}

	// Legend — wraps to a new line when it would overflow the body width.
	pdf.SetFont(fontFamily, "", 8)
	x := pdfMargin
	y := pdf.GetY()
	const swatchH = 4.0
	for i, s := range series {
		label := truncate(s.Label, 22)
		w := 6 + pdf.GetStringWidth(label) + 6
		if x+w > pdfMargin+pdfBodyW {
			x = pdfMargin
			y += 6
		}
		clr := pdfChartColorAt(i, s.IsOther)
		setFill(pdf, clr)
		setDraw(pdf, clr)
		pdf.Rect(x, y+1, swatchH, swatchH, "FD")
		setTxt(pdf, clrText)
		pdf.SetXY(x+swatchH+1.5, y)
		pdf.CellFormat(pdf.GetStringWidth(label)+2, swatchH+2, label, "", 0, "L", false, 0, "")
		x += w
	}
	pdf.SetY(y + 10)

	const labelW = 45.0
	const rowH = 8.0
	barMaxW := pdfBodyW - labelW

	periodTotals := make([]int, len(periods))
	maxTotal := 0
	for pi := range periods {
		sum := 0
		for _, s := range series {
			sum += s.Data[pi]
		}
		periodTotals[pi] = sum
		if sum > maxTotal {
			maxTotal = sum
		}
	}
	if maxTotal == 0 {
		maxTotal = 1
	}

	for pi, label := range periods {
		rowY := pdf.GetY()
		pdf.SetFont(fontFamily, "", 8)
		setTxt(pdf, clrText)
		pdf.SetXY(pdfMargin, rowY+1)
		pdf.CellFormat(labelW-2, rowH-2, truncate(pdfTranslateLabel(label, tr), 24), "", 0, "L", false, 0, "")

		segX := pdfMargin + labelW
		for si, s := range series {
			v := s.Data[pi]
			if v <= 0 {
				continue
			}
			segW := barMaxW * float64(v) / float64(maxTotal)
			setFill(pdf, pdfChartColorAt(si, s.IsOther))
			pdf.Rect(segX, rowY+1, segW, rowH-3, "F")
			segX += segW
		}

		pdf.SetXY(segX+2, rowY+1)
		pdf.SetFont(fontFamily, "B", 8)
		setTxt(pdf, clrMuted)
		pdf.CellFormat(20, rowH-2, fmtDecimalH(periodTotals[pi]), "", 0, "L", false, 0, "")

		pdf.SetY(rowY + rowH)
	}
}

// drawChartEmptyState renders a centered "no data" message in place of a chart.
func drawChartEmptyState(pdf *gofpdf.Fpdf, fontFamily string, tr pdfI18n) {
	pdf.SetFont(fontFamily, "", 10)
	setTxt(pdf, clrMuted)
	pdf.CellFormat(pdfBodyW, 20, tr.ChartTitle, "", 1, "C", false, 0, "")
}
