package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// renderInvoicePDF generates an invoice PDF and returns the raw bytes.
// lang is a BCP-47 language tag (e.g. "en", "nl"). Other display options
// (font, distance_unit) use the requesting user's DB preferences.
func renderInvoicePDF(invoice models.Invoice, lang string) ([]byte, error) {
	// Reuse the full PDF-generation logic by synthesising a gin context that
	// writes into a buffer.  The simplest approach is to call the same helper
	// code.  Here we duplicate just the parameters needed and call
	// buildInvoicePDF, which is the refactored core of GetInvoicePDF.
	return buildInvoicePDF(invoice, lang, "FreeSans", "km")
}

// buildInvoicePDF contains all PDF-building logic and returns the raw PDF bytes.
// It is the shared core called by both GetInvoicePDF and renderInvoicePDF.
func buildInvoicePDF(invoice models.Invoice, lang, fontFamily, distanceUnit string) ([]byte, error) {
	var lineItems []models.InvoiceLineItem
	_ = json.Unmarshal([]byte(invoice.LineItems), &lineItems)

	tr := pdfI18nFromLang(lang)

	settings := loadAllSettings()
	companyName := settings[settingCompanyName]
	if companyName == "" {
		companyName = "WarmDesk"
	}
	companyLogo := settings[settingCompanyLogo]
	companyAddress := settings[settingCompanyAddress]
	companyCity := settings[settingCompanyCity]
	companyPostal := settings[settingCompanyPostalCode]
	companyCountry := settings[settingCompanyCountry]
	companyVAT := settings[settingCompanyVATNumber]
	companyCOC := settings[settingCompanyCOCNumber]
	companyIBAN := settings[settingCompanyIBAN]
	companyBIC := settings[settingCompanyBIC]
	companyTerms := settings[settingCompanyPaymentTerms]
	vatExempt := settings[settingInvoiceVATExempt] == "true"

	// ── Build PDF ────────────────────────────────────────────────────────────
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AliasNbPages("{nb}")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(true, 25)

	for fam, files := range pdfFonts {
		pdf.AddUTF8FontFromBytes(fam, "", mustFont(files[0]))
		pdf.AddUTF8FontFromBytes(fam, "B", mustFont(files[1]))
	}

	invoiceTitle := tr.Invoice + " " + invoice.InvoiceNumber
	pdf.SetTitle(invoiceTitle, true)
	pdf.SetAuthor(companyName, true)
	pdf.SetSubject(companyName+" — "+invoiceTitle, true)
	pdf.SetCreator(companyName, true)

	ff := fontFamily
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont(ff, "", 8)
		setTxt(pdf, clrMuted)
		pdf.CellFormat(pdfBodyW/2, 5, companyName+" — "+invoiceTitle, "", 0, "L", false, 0, "")
		pdf.CellFormat(pdfBodyW/2, 5,
			fmt.Sprintf("%s %d / {nb}", tr.Page, pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()

	// ── Header: logo left, invoice metadata right ─────────────────────────────
	logoLoaded := false
	if rawBytes, ext, ok := resolveLogoBytes(companyLogo); ok {
		logoLoaded = renderLogoIntoPDF(pdf, rawBytes, ext, pdfMargin, pdfMargin)
	}

	// Company block — below or beside logo.
	textX := pdfMargin
	if logoLoaded {
		textX = pdfMargin + 32
	}
	headerTopY := pdfMargin
	textW := (pdfBodyW / 2) - 4

	pdf.SetXY(textX, headerTopY)
	pdf.SetFont(fontFamily, "B", 13)
	setTxt(pdf, clrPrimary)
	pdf.CellFormat(textW, 7, companyName, "", 2, "L", false, 0, "")
	pdf.SetX(textX)
	pdf.SetFont(fontFamily, "", 9)
	setTxt(pdf, clrText)
	if companyAddress != "" {
		pdf.CellFormat(textW, 5, companyAddress, "", 2, "L", false, 0, "")
		pdf.SetX(textX)
	}
	if companyPostal != "" || companyCity != "" {
		pdf.CellFormat(textW, 5, companyPostal+" "+companyCity, "", 2, "L", false, 0, "")
		pdf.SetX(textX)
	}
	if companyCountry != "" {
		pdf.CellFormat(textW, 5, companyCountry, "", 2, "L", false, 0, "")
		pdf.SetX(textX)
	}
	if companyVAT != "" {
		setTxt(pdf, clrMuted)
		pdf.CellFormat(textW, 5, "VAT: "+companyVAT, "", 2, "L", false, 0, "")
		pdf.SetX(textX)
	}
	if companyCOC != "" {
		pdf.CellFormat(textW, 5, "COC: "+companyCOC, "", 2, "L", false, 0, "")
		pdf.SetX(textX)
	}

	// Invoice metadata block (right column, top-aligned).
	metaX := pdfMargin + pdfBodyW/2 + 4
	metaW := pdfBodyW/2 - 4
	pdf.SetXY(metaX, headerTopY)
	pdf.SetFont(fontFamily, "B", 18)
	setTxt(pdf, clrText)
	pdf.CellFormat(metaW, 9, tr.Invoice, "", 2, "R", false, 0, "")
	pdf.SetX(metaX)
	pdf.SetFont(fontFamily, "B", 10)
	setTxt(pdf, clrPrimary)
	pdf.CellFormat(metaW, 6, invoice.InvoiceNumber, "", 2, "R", false, 0, "")
	pdf.SetX(metaX)
	pdf.SetFont(fontFamily, "", 9)
	setTxt(pdf, clrMuted)
	pdf.CellFormat(metaW, 5, invoiceDateLabel(invoice.CreatedAt, tr), "", 2, "R", false, 0, "")
	if invoice.DueDate != nil {
		pdf.SetX(metaX)
		setTxt(pdf, clrText)
		pdf.SetFont(fontFamily, "B", 9)
		pdf.CellFormat(metaW, 5, invoiceDueDateLabel(tr)+" "+invoiceDateLabel(*invoice.DueDate, tr), "", 2, "R", false, 0, "")
	}

	// ── Rule ────────────────────────────────────────────────────────────────
	ruleY := pdf.GetY() + 4
	leftY := headerTopY + 40 // ensure enough room for company address block
	if logoLoaded && leftY < headerTopY+22 {
		leftY = headerTopY + 22
	}
	companyBottomY := pdf.GetY()
	if companyBottomY > leftY {
		leftY = companyBottomY
	}
	if ruleY < leftY+4 {
		ruleY = leftY + 4
	}
	setDraw(pdf, clrPrimary)
	pdf.SetLineWidth(0.4)
	pdf.Line(pdfMargin, ruleY, pdfMargin+pdfBodyW, ruleY)
	pdf.SetY(ruleY + 5)

	// ── Bill-To block ───────────────────────────────────────────────────────
	if invoice.Customer != nil {
		cust := invoice.Customer
		pdf.SetFont(fontFamily, "B", 8)
		setTxt(pdf, clrMuted)
		pdf.CellFormat(pdfBodyW/2, 5, invoiceBillToLabel(tr), "", 2, "L", false, 0, "")
		pdf.SetX(pdfMargin)
		pdf.SetFont(fontFamily, "B", 10)
		setTxt(pdf, clrText)
		pdf.CellFormat(pdfBodyW/2, 6, cust.Name, "", 2, "L", false, 0, "")
		pdf.SetX(pdfMargin)
		pdf.SetFont(fontFamily, "", 9)
		setTxt(pdf, clrText)
		if cust.BillingStreet != "" {
			pdf.CellFormat(pdfBodyW/2, 5, cust.BillingStreet, "", 2, "L", false, 0, "")
			pdf.SetX(pdfMargin)
		}
		if cust.BillingPostalCode != "" || cust.BillingCity != "" {
			pdf.CellFormat(pdfBodyW/2, 5, cust.BillingPostalCode+" "+cust.BillingCity, "", 2, "L", false, 0, "")
			pdf.SetX(pdfMargin)
		}
		if cust.BillingCountry != "" {
			pdf.CellFormat(pdfBodyW/2, 5, cust.BillingCountry, "", 2, "L", false, 0, "")
			pdf.SetX(pdfMargin)
		}
		setTxt(pdf, clrMuted)
		if cust.VATNumber != "" {
			pdf.CellFormat(pdfBodyW/2, 5, "VAT: "+cust.VATNumber, "", 2, "L", false, 0, "")
			pdf.SetX(pdfMargin)
		}
		if cust.POReference != "" {
			pdf.CellFormat(pdfBodyW/2, 5, "PO: "+cust.POReference, "", 2, "L", false, 0, "")
			pdf.SetX(pdfMargin)
		}
		pdf.Ln(4)
	}

	// ── Period label ─────────────────────────────────────────────────────────
	periodStr := invoiceDateLabel(invoice.PeriodStart, tr) + " – " + invoiceDateLabel(invoice.PeriodEnd, tr)
	pdf.SetFont(fontFamily, "", 9)
	setTxt(pdf, clrMuted)
	pdf.CellFormat(pdfBodyW, 5, invoicePeriodLabel(tr)+": "+periodStr, "", 2, "L", false, 0, "")
	pdf.Ln(3)

	// ── Line items table ──────────────────────────────────────────────────────
	const (
		colDate   = 22.0
		colProj   = 38.0
		colDesc   = 0.0 // filled dynamically
		colHours  = 18.0
		colDist   = 16.0
		colRate   = 18.0
		colAmount = 22.0
	)
	hasDistance := false
	for _, li := range lineItems {
		if li.Distance > 0 {
			hasDistance = true
			break
		}
	}
	distColW := 0.0
	if hasDistance {
		distColW = colDist
	}
	descW := pdfBodyW - colDate - colProj - colHours - distColW - colRate - colAmount

	// Table header.
	setFill(pdf, clrTblHdr)
	setTxt(pdf, clrText)
	pdf.SetFont(fontFamily, "B", 8)
	pdf.SetX(pdfMargin)
	pdf.CellFormat(colDate, pdfRowH, tr.Date, "", 0, "L", true, 0, "")
	pdf.CellFormat(colProj, pdfRowH, tr.Project, "", 0, "L", true, 0, "")
	pdf.CellFormat(descW, pdfRowH, tr.Activity, "", 0, "L", true, 0, "")
	pdf.CellFormat(colHours, pdfRowH, tr.Hours, "", 0, "R", true, 0, "")
	if hasDistance {
		pdf.CellFormat(colDist, pdfRowH, distanceUnit, "", 0, "R", true, 0, "")
	}
	pdf.CellFormat(colRate, pdfRowH, tr.Rate, "", 0, "R", true, 0, "")
	pdf.CellFormat(colAmount, pdfRowH, tr.Cost, "B", 2, "R", true, 0, "")

	// Data rows.
	for i, li := range lineItems {
		alt := i%2 == 1
		if alt {
			setFill(pdf, clrAltRow)
		} else {
			setFill(pdf, clrWhite)
		}
		setTxt(pdf, clrText)
		pdf.SetFont(fontFamily, "", 8)
		pdf.SetX(pdfMargin)
		pdf.CellFormat(colDate, pdfRowH, li.Date, "", 0, "L", alt, 0, "")
		pdf.CellFormat(colProj, pdfRowH, truncStr(li.ProjectName, 22), "", 0, "L", alt, 0, "")
		pdf.CellFormat(descW, pdfRowH, truncStr(li.Description, 40), "", 0, "L", alt, 0, "")
		pdf.CellFormat(colHours, pdfRowH, fmtMinutes(li.Minutes), "", 0, "R", alt, 0, "")
		if hasDistance {
			distStr := ""
			if li.Distance > 0 {
				distStr = fmt.Sprintf("%.1f", li.Distance)
			}
			pdf.CellFormat(colDist, pdfRowH, distStr, "", 0, "R", alt, 0, "")
		}
		rateStr := ""
		if li.HourlyRate > 0 {
			rateStr = fmt.Sprintf("%.2f", li.HourlyRate)
		} else if li.Distance > 0 && li.PricePerKm > 0 {
			rateStr = fmt.Sprintf("%.2f", li.PricePerKm)
		}
		pdf.CellFormat(colRate, pdfRowH, rateStr, "", 0, "R", alt, 0, "")
		amtStr := fmt.Sprintf("%s %.2f", invoice.Currency, li.Amount)
		pdf.CellFormat(colAmount, pdfRowH, amtStr, "", 2, "R", alt, 0, "")
	}

	pdf.Ln(2)

	// ── Totals block ──────────────────────────────────────────────────────────
	totalsX := pdfMargin + pdfBodyW - colAmount - colRate
	totalsLabelW := colRate
	totalsValW := colAmount

	setFill(pdf, clrTotal)
	setTxt(pdf, clrText)
	pdf.SetFont(fontFamily, "", 9)
	pdf.SetX(totalsX)
	pdf.CellFormat(totalsLabelW, pdfRowH, tr.Subtotal, "", 0, "L", true, 0, "")
	pdf.CellFormat(totalsValW, pdfRowH, fmt.Sprintf("%s %.2f", invoice.Currency, invoice.Subtotal), "", 2, "R", true, 0, "")

	if vatExempt {
		// Show exemption note below the subtotal, spanning the full body width.
		pdf.SetX(pdfMargin)
		setTxt(pdf, clrMuted)
		pdf.SetFont(fontFamily, "I", 8)
		exemptNote := tr.VATExemptNote
		if exemptNote == "" {
			exemptNote = "VAT exempt under the small business scheme"
		}
		pdf.CellFormat(pdfBodyW, pdfRowH, exemptNote, "", 2, "L", false, 0, "")
		pdf.SetFont(fontFamily, "", 9)
		pdf.SetX(totalsX)
	} else if invoice.VATRate > 0 {
		pdf.SetX(totalsX)
		setTxt(pdf, clrMuted)
		vatLabel := fmt.Sprintf("VAT %.0f%%", invoice.VATRate)
		pdf.CellFormat(totalsLabelW, pdfRowH, vatLabel, "", 0, "L", true, 0, "")
		pdf.CellFormat(totalsValW, pdfRowH, fmt.Sprintf("%s %.2f", invoice.Currency, invoice.VATAmount), "", 2, "R", true, 0, "")
	}

	// Grand total row — primary background.
	pdf.SetX(totalsX)
	setFill(pdf, clrPrimary)
	setTxt(pdf, clrWhite)
	pdf.SetFont(fontFamily, "B", 10)
	pdf.CellFormat(totalsLabelW, pdfRowH+1, tr.Total, "", 0, "L", true, 0, "")
	pdf.CellFormat(totalsValW, pdfRowH+1, fmt.Sprintf("%s %.2f", invoice.Currency, invoice.Total), "", 2, "R", true, 0, "")

	pdf.Ln(6)

	// ── Payment details block ────────────────────────────────────────────────
	if companyIBAN != "" || companyBIC != "" || companyTerms != "" {
		setDraw(pdf, clrPrimary)
		pdf.SetLineWidth(0.3)
		pdf.Line(pdfMargin, pdf.GetY(), pdfMargin+pdfBodyW, pdf.GetY())
		pdf.Ln(4)

		pdf.SetFont(fontFamily, "B", 8)
		setTxt(pdf, clrMuted)
		pdf.CellFormat(pdfBodyW, 5, invoicePaymentLabel(tr), "", 2, "L", false, 0, "")
		pdf.SetX(pdfMargin)
		pdf.SetFont(fontFamily, "", 9)
		setTxt(pdf, clrText)
		if companyIBAN != "" {
			pdf.CellFormat(pdfBodyW, 5, "IBAN: "+companyIBAN, "", 2, "L", false, 0, "")
			pdf.SetX(pdfMargin)
		}
		if companyBIC != "" {
			pdf.CellFormat(pdfBodyW, 5, "BIC: "+companyBIC, "", 2, "L", false, 0, "")
			pdf.SetX(pdfMargin)
		}
		if companyTerms != "" {
			pdf.SetFont(fontFamily, "", 9)
			setTxt(pdf, clrMuted)
			pdf.CellFormat(pdfBodyW, 5, invoicePaymentTermsStr(companyTerms, tr), "", 2, "L", false, 0, "")
		}
	}

	// ── Notes ────────────────────────────────────────────────────────────────
	if invoice.Notes != "" {
		pdf.Ln(4)
		pdf.SetFont(fontFamily, "B", 8)
		setTxt(pdf, clrMuted)
		pdf.CellFormat(pdfBodyW, 5, invoiceNotesLabel(tr), "", 2, "L", false, 0, "")
		pdf.SetX(pdfMargin)
		pdf.SetFont(fontFamily, "", 9)
		setTxt(pdf, clrText)
		pdf.MultiCell(pdfBodyW, 5, invoice.Notes, "", "L", false)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetInvoicePDF renders an invoice as a PDF and streams it to the client.
func GetInvoicePDF(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	invoiceID, err := strconv.ParseUint(c.Param("invoiceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice id"})
		return
	}
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)
	if err := requireCustomerAccess(uint(custID), userID, globalRole); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var invoice models.Invoice
	if err := database.DB.Preload("Customer").First(&invoice, invoiceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if invoice.CustomerID != uint(custID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// Requesting user's font preference.
	var requestingUser models.User
	fontFamily := "FreeSans"
	if err := database.DB.First(&requestingUser, userID).Error; err == nil {
		fontFamily = pdfFontFamily(requestingUser.Font)
	}
	if fam, ok := pdfFontFromParam(c.Query("font")); ok {
		fontFamily = fam
	}
	lang := c.Query("lang")
	distanceUnit := c.DefaultQuery("distance_unit", "km")

	pdfBytes, err := buildInvoicePDF(invoice, lang, fontFamily, distanceUnit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}

	// ── Stream PDF ───────────────────────────────────────────────────────────
	filename := fmt.Sprintf("invoice-%s.pdf", invoice.InvoiceNumber)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `inline; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// invoiceDateLabel formats a time.Time for the invoice using the locale's DMY setting.
func invoiceDateLabel(t time.Time, tr pdfI18n) string {
	if tr.DMY {
		return fmt.Sprintf("%02d %s %d", t.Day(), tr.MonthsAbbr[t.Month()-1], t.Year())
	}
	return fmt.Sprintf("%s %02d, %d", tr.MonthsAbbr[t.Month()-1], t.Day(), t.Year())
}

// invoiceBillToLabel returns the "Bill To" label for the current locale.
func invoiceBillToLabel(tr pdfI18n) string {
	if tr.BillTo != "" {
		return tr.BillTo
	}
	return "Bill To"
}

func invoicePeriodLabel(tr pdfI18n) string {
	if tr.Period != "" {
		return tr.Period
	}
	return "Period"
}

func invoiceDueDateLabel(tr pdfI18n) string {
	if tr.Due != "" {
		return tr.Due
	}
	return "Due"
}

func invoicePaymentLabel(tr pdfI18n) string {
	if tr.PaymentDetails != "" {
		return tr.PaymentDetails
	}
	return "Payment Details"
}

func invoiceNotesLabel(tr pdfI18n) string {
	if tr.Notes != "" {
		return tr.Notes
	}
	return "Notes"
}

// invoicePaymentTermsStr formats a raw payment-terms value (e.g. "30") as a
// human-readable string ("Net 30 days") using the locale template.
func invoicePaymentTermsStr(terms string, tr pdfI18n) string {
	tmpl := tr.PaymentTermsDays
	if tmpl == "" {
		tmpl = "Net %s days"
	}
	return fmt.Sprintf(tmpl, terms)
}
