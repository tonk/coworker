package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
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
	pdf.SetFooterFunc(func() {
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
	)
	if showAbbr {
		colDate = colAbbr + 25.0
		colDesc = 40.0
	}
	colHours := pdfBodyW - colDate - colCust - colProj - colDesc

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

			dateCell := e.Date.Format("2006-01-02")
			if showAbbr {
				// Render abbreviation in a fixed-width sub-cell so the date
				// column stays aligned regardless of 2- vs 3-char abbr width.
				abbrIdx := (int(e.Date.Weekday()) + 6) % 7
				pdf.CellFormat(colAbbr, rowH, tr.DaysAbbr[abbrIdx], "B", 0, "L", true, 0, "")
				pdf.CellFormat(colDate-colAbbr, rowH, dateCell, "B", 0, "L", true, 0, "")
			} else {
				pdf.CellFormat(colDate, rowH, dateCell, "B", 0, "L", true, 0, "")
			}
			pdf.CellFormat(colCust, rowH, customerName, "B", 0, "L", true, 0, "")
			pdf.CellFormat(colProj, rowH, projectName, "B", 0, "L", true, 0, "")
			pdf.CellFormat(colDesc, rowH, truncate(e.Description, 32), "B", 0, "L", true, 0, "")
			pdf.CellFormat(colHours, rowH, fmtDecimalH(pdfEntryDeclarable(e)), "B", 1, "R", true, 0, "")
		}

		// Group subtotal
		pdf.SetFont(fontFamily, "B", 8)
		setTxt(pdf, clrMuted)
		setFill(pdf, rgb{238, 242, 248})
		pdf.CellFormat(pdfBodyW-colHours, rowH, "  "+pdfTranslateLabel(grp.Label, tr)+" "+tr.Total, "0", 0, "R", true, 0, "")
		setTxt(pdf, clrPrimary)
		pdf.CellFormat(colHours, rowH, fmtDecimalH(grp.DeclarableMinutes), "0", 1, "R", true, 0, "")

		// Per-group undeclarable line (customer grouping only)
		if groupBy == "customer" && grp.UndeclarableMinutes > 0 {
			setFill(pdf, rgb{252, 245, 245})
			setTxt(pdf, clrMuted)
			pdf.SetFont(fontFamily, "", 8)
			pdf.CellFormat(pdfBodyW-colHours, rowH, "  "+tr.Undeclarable, "0", 0, "R", true, 0, "")
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
		setFill(pdf, clrPrimary)
		setTxt(pdf, rgb{255, 255, 255})
		pdf.SetFont(fontFamily, "B", 9)
		pdf.CellFormat(pdfBodyW-colHours, rowH+1, "  "+tr.Total, "0", 0, "L", true, 0, "")
		pdf.CellFormat(colHours, rowH+1, fmtDecimalH(grandMinutes), "0", 1, "R", true, 0, "")

		if groupBy == "customer" && report.UndeclarableMinutes > 0 {
			setFill(pdf, rgb{252, 245, 245})
			setTxt(pdf, clrMuted)
			pdf.SetFont(fontFamily, "", 8.5)
			pdf.CellFormat(pdfBodyW-colHours, rowH, "  "+tr.Undeclarable, "0", 0, "L", true, 0, "")
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
