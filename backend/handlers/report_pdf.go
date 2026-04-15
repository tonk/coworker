package handlers

import (
	"bytes"
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

//go:embed fonts
var fontEmbedFS embed.FS

var reportCfg *config.Config

// InitReport stores the config reference for PDF generation.
func InitReport(cfg *config.Config) {
	reportCfg = cfg
}

// mustFont reads a font file from the embedded FS and panics on failure.
// Only called at server start-up when fonts are registered.
func mustFont(name string) []byte {
	b, err := fontEmbedFS.ReadFile("fonts/" + name)
	if err != nil {
		panic("warmdesk: embedded font missing: " + name + ": " + err.Error())
	}
	return b
}

// pdfFontFamily maps a user's CSS font preference to a FreeFont family name.
func pdfFontFamily(fontPref string) string {
	fp := strings.ToLower(fontPref)
	switch {
	case strings.Contains(fp, "mono") || strings.Contains(fp, "code"):
		return "FreeMono"
	case strings.Contains(fp, "serif") || strings.Contains(fp, "georgia"):
		return "FreeSerif"
	default:
		return "FreeSans"
	}
}

// fmtMinutes formats a minute count as H:MM.
func fmtMinutes(m int) string {
	return fmt.Sprintf("%d:%02d", m/60, m%60)
}

// truncStr clips a string to maxRunes, appending "…" if needed.
func truncStr(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes-1]) + "…"
}

// Page geometry (mm, A4 portrait).
const (
	pdfPageW  = 210.0
	pdfMargin = 15.0
	pdfBodyW  = pdfPageW - 2*pdfMargin
	pdfRowH   = 6.5
	pdfColRef = 20.0
	pdfColTit = 82.0
	pdfColAss = 48.0
	pdfColTim = 30.0
)

// Colors.
type rgb struct{ r, g, b int }

var (
	clrPrimary = rgb{99, 102, 241}
	clrTblHdr  = rgb{232, 232, 240}
	clrAltRow  = rgb{248, 248, 252}
	clrMuted   = rgb{100, 116, 139}
	clrTotal   = rgb{240, 240, 248}
	clrWhite   = rgb{255, 255, 255}
	clrText    = rgb{30, 30, 46}
)

func setFill(pdf *gofpdf.Fpdf, c rgb) { pdf.SetFillColor(c.r, c.g, c.b) }
func setDraw(pdf *gofpdf.Fpdf, c rgb) { pdf.SetDrawColor(c.r, c.g, c.b) }
func setTxt(pdf *gofpdf.Fpdf, c rgb)  { pdf.SetTextColor(c.r, c.g, c.b) }

// GetTimeReportPDF generates the time report as a PDF file.
func GetTimeReportPDF(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	if !userCanViewReports(userID, globalRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "reports are only available to project admins and system admins"})
		return
	}

	report, status, errMsg := assembleTimeReport(c, userID, globalRole)
	if status != 0 {
		c.JSON(status, gin.H{"error": errMsg})
		return
	}

	// Resolve user's font preference.
	var user models.User
	fontFamily := "FreeSans"
	if err := database.DB.First(&user, userID).Error; err == nil {
		fontFamily = pdfFontFamily(user.Font)
	}

	// ── Build PDF ────────────────────────────────────────────────

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AliasNbPages("{nb}")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(true, 20)

	// Register FreeFont families (regular + bold) directly from embedded bytes.
	// AddUTF8Font in gofpdf v1 does not use SetFontLoader, so we use
	// AddUTF8FontFromBytes which accepts raw TTF bytes.
	for _, fam := range []string{"FreeSans", "FreeSerif", "FreeMono"} {
		pdf.AddUTF8FontFromBytes(fam, "", mustFont(fam+".ttf"))
		pdf.AddUTF8FontFromBytes(fam, "B", mustFont(fam+"Bold.ttf"))
	}

	// PDF metadata.
	title := "Time Report — " + report.PeriodLabel
	author := user.DisplayName
	if author == "" {
		author = user.Username
	}
	company := report.CompanyName
	if company == "" {
		company = "WarmDesk"
	}
	pdf.SetTitle(title, true)
	pdf.SetAuthor(author, true)
	pdf.SetSubject(company+" — "+title, true)
	pdf.SetCreator(company, true)

	// Footer — repeated on every page.
	ff := fontFamily
	rpt := report
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont(ff, "", 8)
		setTxt(pdf, clrMuted)
		label := "WarmDesk — Time Report"
		if rpt.CompanyName != "" {
			label = rpt.CompanyName + " — Time Report"
		}
		pdf.CellFormat(pdfBodyW/2, 5, label, "", 0, "L", false, 0, "")
		pdf.CellFormat(pdfBodyW/2, 5,
			fmt.Sprintf("Page %d / {nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()

	// ── Page header ──────────────────────────────────────────────

	// Try to load company logo (uploaded files only; skip SVG and external URLs).
	logoY := pdfMargin
	logoLoaded := false
	if report.CompanyLogo != "" && strings.HasPrefix(report.CompanyLogo, "/uploads/") {
		uploadDir := "./uploads"
		if reportCfg != nil && reportCfg.UploadDir != "" {
			uploadDir = reportCfg.UploadDir
		}
		storedName := strings.TrimPrefix(report.CompanyLogo, "/uploads/")
		logoPath := filepath.Join(uploadDir, storedName)
		ext := strings.ToLower(filepath.Ext(logoPath))
		imgType := ""
		switch ext {
		case ".jpg", ".jpeg":
			imgType = "JPG"
		case ".png":
			imgType = "PNG"
		case ".gif":
			imgType = "GIF"
		case ".webp":
			imgType = "WEBP"
		}
		if imgType != "" {
			if f, err := os.Open(logoPath); err == nil {
				opts := gofpdf.ImageOptions{ImageType: imgType, ReadDpi: false}
				pdf.RegisterImageOptionsReader("_logo", opts, f)
				f.Close()
				pdf.ImageOptions("_logo", pdfMargin, logoY, 0, 18, false, opts, 0, "")
				logoLoaded = true
			}
		}
	}

	// Text block — shifted right when a logo is present.
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

	pdf.SetFont(fontFamily, "B", 17)
	setTxt(pdf, clrText)
	pdf.CellFormat(textW, 9, "Time Report", "", 2, "L", false, 0, "")
	pdf.SetX(textX)

	pdf.SetFont(fontFamily, "", 10)
	setTxt(pdf, clrMuted)
	pdf.CellFormat(textW, 5.5, report.PeriodLabel, "", 2, "L", false, 0, "")
	pdf.SetX(textX)

	pdf.SetFont(fontFamily, "", 8)
	pdf.CellFormat(textW, 4.5, "Generated: "+report.GeneratedAt, "", 2, "L", false, 0, "")

	// Horizontal rule below header.
	ruleY := pdf.GetY() + 2
	if logoLoaded && ruleY < logoY+22 {
		ruleY = logoY + 22
	}
	setDraw(pdf, clrPrimary)
	pdf.SetLineWidth(0.4)
	pdf.Line(pdfMargin, ruleY, pdfPageW-pdfMargin, ruleY)
	pdf.SetY(ruleY + 5)

	// ── Project sections ─────────────────────────────────────────

	for _, proj := range report.Projects {
		// Section header bar.
		setFill(pdf, clrPrimary)
		setTxt(pdf, clrWhite)
		pdf.SetFont(fontFamily, "B", 9.5)
		pdf.SetX(pdfMargin)
		pdf.CellFormat(pdfBodyW, 7, "  "+proj.ProjectName, "", 2, "L", true, 0, "")

		// Column header row.
		setFill(pdf, clrTblHdr)
		setTxt(pdf, clrText)
		pdf.SetFont(fontFamily, "B", 8)
		pdf.SetX(pdfMargin)
		pdf.CellFormat(pdfColRef, pdfRowH, "Ref", "", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColTit, pdfRowH, "Title", "", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColAss, pdfRowH, "Assignees", "", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColTim, pdfRowH, "Time", "B", 2, "R", true, 0, "")

		// Card rows.
		for i, card := range proj.Cards {
			alt := i%2 == 1
			if alt {
				setFill(pdf, clrAltRow)
			} else {
				setFill(pdf, clrWhite)
			}
			setTxt(pdf, clrText)
			pdf.SetFont(fontFamily, "", 8)
			pdf.SetX(pdfMargin)
			pdf.CellFormat(pdfColRef, pdfRowH, card.CardRef, "", 0, "L", alt, 0, "")
			pdf.CellFormat(pdfColTit, pdfRowH, truncStr(card.Title, 46), "", 0, "L", alt, 0, "")
			pdf.CellFormat(pdfColAss, pdfRowH, truncStr(strings.Join(card.Assignees, ", "), 28), "", 0, "L", alt, 0, "")
			pdf.CellFormat(pdfColTim, pdfRowH, fmtMinutes(card.TimeSpentMinutes), "", 2, "R", alt, 0, "")
		}

		// Project subtotal.
		setFill(pdf, clrTotal)
		setTxt(pdf, clrText)
		pdf.SetFont(fontFamily, "B", 8)
		pdf.SetX(pdfMargin)
		pdf.CellFormat(pdfColRef+pdfColTit+pdfColAss, pdfRowH, "Subtotal:", "T", 0, "R", true, 0, "")
		pdf.CellFormat(pdfColTim, pdfRowH, fmtMinutes(proj.TotalMinutes), "T", 2, "R", true, 0, "")

		pdf.Ln(4)
	}

	// ── Grand total ──────────────────────────────────────────────

	setDraw(pdf, clrPrimary)
	pdf.SetLineWidth(0.3)
	pdf.Line(pdfMargin, pdf.GetY(), pdfPageW-pdfMargin, pdf.GetY())
	pdf.Ln(1)

	setFill(pdf, clrPrimary)
	setTxt(pdf, clrWhite)
	pdf.SetFont(fontFamily, "B", 9.5)
	pdf.SetX(pdfMargin)
	pdf.CellFormat(pdfColRef+pdfColTit+pdfColAss, 8, "Grand Total:", "", 0, "R", true, 0, "")
	pdf.CellFormat(pdfColTim, 8, fmtMinutes(report.TotalMinutes), "", 2, "R", true, 0, "")

	// ── Stream output ────────────────────────────────────────────

	filename := "time-report"
	if report.PeriodLabel != "" && report.PeriodLabel != "All Time" {
		filename += "-" + strings.ReplaceAll(report.PeriodLabel, " ", "-")
	}
	filename += ".pdf"

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pdf generation failed: " + err.Error()})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
