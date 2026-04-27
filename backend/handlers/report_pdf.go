package handlers

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
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

// pdfFonts maps font family names to their embedded TTF file names (regular, bold).
var pdfFonts = map[string][2]string{
	"Inter":            {"Inter-Regular.ttf", "Inter-Bold.ttf"},
	"Roboto":           {"Roboto-Regular.ttf", "Roboto-Bold.ttf"},
	"OpenSans":         {"OpenSans-Regular.ttf", "OpenSans-Bold.ttf"},
	"SourceCodePro":    {"SourceCodePro-Regular.ttf", "SourceCodePro-Bold.ttf"},
	"FreeSans":         {"FreeSans.ttf", "FreeSansBold.ttf"},
	"FreeSerif":        {"FreeSerif.ttf", "FreeSerifBold.ttf"},
	"FreeMono":         {"FreeMono.ttf", "FreeMonoBold.ttf"},
}

// pdfFontFromParam maps the ?font= query value to a registered family name.
func pdfFontFromParam(param string) (string, bool) {
	switch param {
	case "inter":
		return "Inter", true
	case "roboto":
		return "Roboto", true
	case "opensans":
		return "OpenSans", true
	case "sourcecode":
		return "SourceCodePro", true
	case "freesans":
		return "FreeSans", true
	case "freeserif":
		return "FreeSerif", true
	case "freemono":
		return "FreeMono", true
	}
	return "", false
}

// pdfFontFamily maps a user's CSS font preference to an embedded font family.
func pdfFontFamily(fontPref string) string {
	fp := strings.ToLower(fontPref)
	switch {
	case strings.Contains(fp, "inter"):
		return "Inter"
	case strings.Contains(fp, "roboto"):
		return "Roboto"
	case strings.Contains(fp, "open sans"):
		return "OpenSans"
	case strings.Contains(fp, "source code") || strings.Contains(fp, "code pro"):
		return "SourceCodePro"
	case strings.Contains(fp, "freemono") || (strings.Contains(fp, "mono") && strings.Contains(fp, "free")):
		return "FreeMono"
	case strings.Contains(fp, "freeserif") || (strings.Contains(fp, "serif") && strings.Contains(fp, "free")):
		return "FreeSerif"
	case strings.Contains(fp, "freesans") || (strings.Contains(fp, "sans") && strings.Contains(fp, "free")):
		return "FreeSans"
	case strings.Contains(fp, "mono") || strings.Contains(fp, "code"):
		return "FreeMono"
	case strings.Contains(fp, "serif") || strings.Contains(fp, "georgia"):
		return "FreeSerif"
	default:
		return "FreeSans"
	}
}

// pdfI18n holds translated strings used directly in the PDF.
type pdfI18n struct {
	TimeReport string // e.g. "Time Report"
	ColRef     string // e.g. "Ref"
	ColTask    string // e.g. "Task"
	ColAssign  string // e.g. "Assignees"
	ColTime    string // e.g. "Time"
	Subtotal   string // e.g. "Subtotal:"
	GrandTotal string // e.g. "Grand Total:"
	Page       string // e.g. "Page"
	Generated  string // e.g. "Generated:"
	// Fields used by the personal time-entry PDF
	Customer string // e.g. "Customer"
	Project  string // e.g. "Project"
	Activity string // e.g. "Activity"
	Hours    string // e.g. "Hours"
	Total    string // e.g. "Total"
}

// pdfTranslations provides label sets for each supported language code.
// Strings are derived from the same source as the frontend i18n files.
var pdfTranslations = map[string]pdfI18n{
	"en": {
		TimeReport: "Time Report",
		ColRef:     "Ref",
		ColTask:    "Task",
		ColAssign:  "Assignees",
		ColTime:    "Time",
		Subtotal:   "Subtotal:",
		GrandTotal: "Grand Total:",
		Page:       "Page",
		Generated:  "Generated:",
		Customer:   "Customer",
		Project:    "Project",
		Activity:   "Activity",
		Hours:      "Hours",
		Total:      "Total",
	},
	"nl": {
		TimeReport: "Tijdrapport",
		ColRef:     "Ref",
		ColTask:    "Taak",
		ColAssign:  "Toegewezen",
		ColTime:    "Tijd",
		Subtotal:   "Subtotaal:",
		GrandTotal: "Eindtotaal:",
		Page:       "Pagina",
		Generated:  "Gegenereerd:",
		Customer:   "Klant",
		Project:    "Project",
		Activity:   "Activiteit",
		Hours:      "Uren",
		Total:      "Totaal",
	},
	"de": {
		TimeReport: "Zeitbericht",
		ColRef:     "Ref",
		ColTask:    "Aufgabe",
		ColAssign:  "Bearbeiter",
		ColTime:    "Zeit",
		Subtotal:   "Zwischensumme:",
		GrandTotal: "Gesamtsumme:",
		Page:       "Seite",
		Generated:  "Erstellt:",
		Customer:   "Kunde",
		Project:    "Projekt",
		Activity:   "Aktivität",
		Hours:      "Stunden",
		Total:      "Gesamt",
	},
	"fr": {
		TimeReport: "Rapport de temps",
		ColRef:     "Réf",
		ColTask:    "Tâche",
		ColAssign:  "Assignés",
		ColTime:    "Temps",
		Subtotal:   "Sous-total :",
		GrandTotal: "Total général :",
		Page:       "Page",
		Generated:  "Généré le :",
		Customer:   "Client",
		Project:    "Projet",
		Activity:   "Activité",
		Hours:      "Heures",
		Total:      "Total",
	},
	"es": {
		TimeReport: "Informe de tiempo",
		ColRef:     "Ref",
		ColTask:    "Tarea",
		ColAssign:  "Asignados",
		ColTime:    "Tiempo",
		Subtotal:   "Subtotal:",
		GrandTotal: "Total general:",
		Page:       "Página",
		Generated:  "Generado el:",
		Customer:   "Cliente",
		Project:    "Proyecto",
		Activity:   "Actividad",
		Hours:      "Horas",
		Total:      "Total",
	},
}

// pdfI18nFromLang returns the translation set for the given BCP-47 language tag.
// Falls back to English for unknown or empty codes.
func pdfI18nFromLang(lang string) pdfI18n {
	// Accept tags like "nl-NL" as well as plain "nl".
	code := strings.ToLower(lang)
	if idx := strings.IndexByte(code, '-'); idx > 0 {
		code = code[:idx]
	}
	if tr, ok := pdfTranslations[code]; ok {
		return tr
	}
	return pdfTranslations["en"]
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

// svgToPNG rasterizes an SVG file to a PNG held in a bytes.Buffer.
// heightMM is the desired output height in millimetres; width is derived from
// the SVG's own aspect ratio. We render at 3× (≈ 216 dpi for A4) for crisp output.
func svgToPNG(svgPath string, heightMM float64) (bytes.Buffer, error) {
	f, err := os.Open(svgPath)
	if err != nil {
		return bytes.Buffer{}, err
	}
	defer f.Close()

	icon, err := oksvg.ReadIconStream(f)
	if err != nil {
		return bytes.Buffer{}, err
	}

	// Target pixel height at 3× (72 dpi base × 3 = 216 dpi).
	const scale = 3.0
	const dpi = 72.0
	const mmPerInch = 25.4
	pxH := int(heightMM / mmPerInch * dpi * scale)

	// Derive width from the SVG viewport aspect ratio.
	vw, vh := icon.ViewBox.W, icon.ViewBox.H
	if vh == 0 {
		vh = 1
	}
	pxW := int(float64(pxH) * vw / vh)
	if pxW < 1 {
		pxW = 1
	}

	icon.SetTarget(0, 0, float64(pxW), float64(pxH))

	img := image.NewRGBA(image.Rect(0, 0, pxW, pxH))
	// Fill with transparent white so the background is white in the PDF.
	for i := range img.Pix {
		img.Pix[i] = 255
	}

	scanner := rasterx.NewScannerGV(pxW, pxH, img, img.Bounds())
	raster := rasterx.NewDasher(pxW, pxH, scanner)
	icon.Draw(raster, 1.0)

	// Set fully opaque alpha for all pixels (compose over white).
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] == 0 {
			// Transparent pixel — already white from the fill above, make opaque.
			img.Pix[i] = 255
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return bytes.Buffer{}, err
	}
	return buf, nil
}

// GetTimeReportPDF generates the time report as a PDF file.
func GetTimeReportPDF(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var user models.User
	database.DB.First(&user, userID)

	if !userCanViewReports(userID, globalRole, user.TimeTrackingViewer) {
		c.JSON(http.StatusForbidden, gin.H{"error": "reports are only available to project admins and system admins"})
		return
	}

	report, status, errMsg := assembleTimeReport(c, userID, globalRole, user.TimeTrackingViewer)
	if status != 0 {
		c.JSON(status, gin.H{"error": errMsg})
		return
	}

	// Resolve font: explicit ?font= param wins, then user profile preference.
	fontFamily := "FreeSans"
	if user.Font != "" {
		fontFamily = pdfFontFamily(user.Font)
	}
	if fam, ok := pdfFontFromParam(c.Query("font")); ok {
		fontFamily = fam
	}

	// Resolve PDF language: ?lang= param, default English.
	tr := pdfI18nFromLang(c.Query("lang"))

	// ── Build PDF ────────────────────────────────────────────────

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AliasNbPages("{nb}")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(true, 20)

	// Register all font families from embedded bytes.
	// AddUTF8Font in gofpdf v1 ignores SetFontLoader, so we use AddUTF8FontFromBytes.
	for fam, files := range pdfFonts {
		pdf.AddUTF8FontFromBytes(fam, "", mustFont(files[0]))
		pdf.AddUTF8FontFromBytes(fam, "B", mustFont(files[1]))
	}

	// PDF metadata.
	title := tr.TimeReport + " — " + report.PeriodLabel
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
	trr := tr
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont(ff, "", 8)
		setTxt(pdf, clrMuted)
		label := "WarmDesk — " + trr.TimeReport
		if rpt.CompanyName != "" {
			label = rpt.CompanyName + " — " + trr.TimeReport
		}
		pdf.CellFormat(pdfBodyW/2, 5, label, "", 0, "L", false, 0, "")
		pdf.CellFormat(pdfBodyW/2, 5,
			fmt.Sprintf("%s %d / {nb}", trr.Page, pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()

	// ── Page header ──────────────────────────────────────────────

	// Try to load company logo (uploaded files only; skip external URLs).
	// SVGs are rasterized to PNG in memory; raster formats are streamed directly.
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

		if ext == ".svg" {
			// Rasterize SVG → PNG at 2× resolution for a crisp 18 mm logo.
			if pngBuf, err := svgToPNG(logoPath, 18); err == nil {
				opts := gofpdf.ImageOptions{ImageType: "PNG"}
				pdf.RegisterImageOptionsReader("_logo", opts, &pngBuf)
				pdf.ImageOptions("_logo", pdfMargin, logoY, 0, 18, false, opts, 0, "")
				logoLoaded = true
			}
		} else {
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
				if rawBytes, err := os.ReadFile(logoPath); err == nil {
					var imgReader *bytes.Reader
					if imgType == "PNG" {
						// Composite over white to strip any alpha channel, which
						// gofpdf cannot handle and causes an internal error.
						if src, decErr := png.Decode(bytes.NewReader(rawBytes)); decErr == nil {
							bounds := src.Bounds()
							dst := image.NewRGBA(bounds)
							draw.Draw(dst, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
							draw.Draw(dst, bounds, src, bounds.Min, draw.Over)
							var pngBuf bytes.Buffer
							if encErr := png.Encode(&pngBuf, dst); encErr == nil {
								imgReader = bytes.NewReader(pngBuf.Bytes())
							}
						}
					}
					if imgReader == nil {
						imgReader = bytes.NewReader(rawBytes)
					}
					opts := gofpdf.ImageOptions{ImageType: imgType}
					pdf.RegisterImageOptionsReader("_logo", opts, imgReader)
					if pdf.Error() == nil {
						pdf.ImageOptions("_logo", pdfMargin, logoY, 0, 18, false, opts, 0, "")
						logoLoaded = true
					}
				}
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
	pdf.CellFormat(textW, 9, tr.TimeReport, "", 2, "L", false, 0, "")
	pdf.SetX(textX)

	pdf.SetFont(fontFamily, "", 10)
	setTxt(pdf, clrMuted)
	pdf.CellFormat(textW, 5.5, report.PeriodLabel, "", 2, "L", false, 0, "")
	pdf.SetX(textX)

	pdf.SetFont(fontFamily, "", 8)
	pdf.CellFormat(textW, 4.5, tr.Generated+" "+report.GeneratedAt, "", 2, "L", false, 0, "")

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
		// Section header bar — primary-coloured background with project name on
		// the left and a white pill badge showing the total on the right.
		const hdrH = 7.0
		hdrY := pdf.GetY()

		// Full-width primary bar.
		setFill(pdf, clrPrimary)
		setDraw(pdf, clrPrimary)
		pdf.Rect(pdfMargin, hdrY, pdfBodyW, hdrH, "F")

		// Pill badge: measure the time string, draw a white rounded rect.
		pdf.SetFont(fontFamily, "B", 8)
		timeStr := fmtMinutes(proj.TotalMinutes)
		const badgePadX = 4.0
		const badgeH = 4.5
		badgeW := pdf.GetStringWidth(timeStr) + 2*badgePadX
		badgeX := pdfMargin + pdfBodyW - badgeW - 2.5
		badgeY := hdrY + (hdrH-badgeH)/2
		pdf.SetFillColor(255, 255, 255)
		pdf.SetDrawColor(255, 255, 255)
		pdf.RoundedRect(badgeX, badgeY, badgeW, badgeH, badgeH/2, "1234", "F")
		setTxt(pdf, clrPrimary)
		pdf.SetXY(badgeX, badgeY)
		pdf.CellFormat(badgeW, badgeH, timeStr, "", 0, "C", false, 0, "")

		// Project name (left side of bar, white text).
		setTxt(pdf, clrWhite)
		pdf.SetFont(fontFamily, "B", 9.5)
		nameW := pdfBodyW - badgeW - 6
		pdf.SetXY(pdfMargin, hdrY)
		pdf.CellFormat(nameW, hdrH, "  "+proj.ProjectName, "", 0, "L", false, 0, "")

		// Restore draw colour and advance past the header row.
		setDraw(pdf, clrPrimary)
		pdf.SetY(hdrY + hdrH)

		// Column header row.
		setFill(pdf, clrTblHdr)
		setTxt(pdf, clrText)
		pdf.SetFont(fontFamily, "B", 8)
		pdf.SetX(pdfMargin)
		pdf.CellFormat(pdfColRef, pdfRowH, tr.ColRef, "", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColTit, pdfRowH, tr.ColTask, "", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColAss, pdfRowH, tr.ColAssign, "", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColTim, pdfRowH, tr.ColTime, "B", 2, "R", true, 0, "")

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
		pdf.CellFormat(pdfColRef+pdfColTit+pdfColAss, pdfRowH, tr.Subtotal, "T", 0, "R", true, 0, "")
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
	pdf.CellFormat(pdfColRef+pdfColTit+pdfColAss, 8, tr.GrandTotal, "", 0, "R", true, 0, "")
	pdf.CellFormat(pdfColTim, 8, fmtMinutes(report.TotalMinutes), "", 2, "R", true, 0, "")

	// ── Stream output ────────────────────────────────────────────

	filename := "time-report"
	if report.PeriodLabel != "" && report.PeriodLabel != "All Time" {
		filename += "-" + strings.ReplaceAll(report.PeriodLabel, " ", "-")
	}
	filename += ".pdf"

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		log.Printf("pdf generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pdf generation failed"})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
