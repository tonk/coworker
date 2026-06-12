package handlers

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/fs"
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
var reportWebFS fs.FS

// InitReport stores the config and web filesystem reference for PDF generation.
func InitReport(cfg *config.Config, webFS fs.FS) {
	reportCfg = cfg
	reportWebFS = webFS
}

// resolveLogoBytes reads the raw bytes and file extension for a logo URL.
// Handles /uploads/ paths (uploaded files) and /path static web assets.
// Returns (data, ext, ok).
func resolveLogoBytes(logoURL string) ([]byte, string, bool) {
	if logoURL == "" {
		return nil, "", false
	}
	ext := strings.ToLower(filepath.Ext(logoURL))

	if strings.HasPrefix(logoURL, "/uploads/") {
		uploadDir := "./uploads"
		if reportCfg != nil && reportCfg.UploadDir != "" {
			uploadDir = reportCfg.UploadDir
		}
		data, err := os.ReadFile(filepath.Join(uploadDir, strings.TrimPrefix(logoURL, "/uploads/")))
		if err != nil {
			return nil, "", false
		}
		return data, ext, true
	}

	// Static web asset (e.g. /logo.svg served by the frontend)
	rel := strings.TrimPrefix(logoURL, "/")
	if rel == "" {
		return nil, "", false
	}
	// In timetracking mode, substitute the time-tracking logo variants so PDFs
	// get the correct branding without requiring a custom company logo to be set.
	if serverAppMode == "timetracking" {
		switch rel {
		case "logo.svg":
			rel, ext = "timetracking.svg", ".svg"
		case "logo-full.svg":
			rel, ext = "timetracking-full.svg", ".svg"
		}
	}
	// Try WebDir on disk first
	if reportCfg != nil && reportCfg.WebDir != "" {
		if data, err := os.ReadFile(filepath.Join(reportCfg.WebDir, rel)); err == nil {
			return data, ext, true
		}
	}
	// Fall back to embedded web FS
	if reportWebFS != nil {
		if data, err := fs.ReadFile(reportWebFS, rel); err == nil {
			return data, ext, true
		}
	}
	return nil, "", false
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
	Customer     string // e.g. "Customer"
	Project      string // e.g. "Project"
	Activity     string // e.g. "Activity"
	Hours        string // e.g. "Hours"
	Total        string // e.g. "Total"
	AllEmployees string // e.g. "All Employees"
	Date         string // e.g. "Date"
	Undeclarable string // e.g. "Undeclarable"
	Declarable   string // e.g. "Declarable"
	// Period-label words used in board and time-entry PDFs
	AllTime    string // e.g. "All Time"
	YearPrefix string // e.g. "Year" (followed by the year number)
	WeekLabel  string // e.g. "Week" (followed by week number)
	// Month and day names — used by pdfTranslateLabel to localise period labels.
	// Index order: months 0=January…11=December; days 0=Monday…6=Sunday.
	MonthsFull [12]string
	MonthsAbbr [12]string
	DaysFull   [7]string
	DaysAbbr   [7]string // Mon=0 … Sun=6 (same order as DaysFull)
	// Costs / invoice
	Rate     string // e.g. "Rate"
	Cost     string // e.g. "Cost"
	Invoice  string // e.g. "Invoice"
	Standard string // e.g. "Standard" (base rate for time-slot breakdowns)
	// DMY signals that dates should be rendered in day-month order ("17 mei")
	// rather than the American month-day order ("May 17") that Go's time.Format
	// always produces. Set to true for every non-English locale.
	DMY bool
}

// pdfTranslations provides label sets for each supported language code.
// Strings are derived from the same source as the frontend i18n files.
var pdfTranslations = map[string]pdfI18n{
	"en": {
		TimeReport:   "Time Report",
		ColRef:       "Ref",
		ColTask:      "Task",
		ColAssign:    "Assignees",
		ColTime:      "Time",
		Subtotal:     "Subtotal:",
		GrandTotal:   "Grand Total:",
		Page:         "Page",
		Generated:    "Generated:",
		Customer:     "Customer",
		Project:      "Project",
		Activity:     "Activity",
		Hours:        "Hours",
		Total:        "Total",
		AllEmployees: "All Employees",
		Date:         "Date",
		Undeclarable: "Undeclarable",
		Declarable:   "Declarable",
		AllTime:      "All Time",
		YearPrefix:   "Year",
		WeekLabel:    "Week",
		Rate:         "Rate",
		Cost:         "Cost",
		Invoice:      "Invoice",
		Standard:     "Standard",
		MonthsFull:   [12]string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
		MonthsAbbr:   [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
		DaysFull:     [7]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
		DaysAbbr:     [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
	},
	"nl": {
		TimeReport:   "Tijdrapport",
		ColRef:       "Ref",
		ColTask:      "Taak",
		ColAssign:    "Toegewezen",
		ColTime:      "Tijd",
		Subtotal:     "Subtotaal:",
		GrandTotal:   "Eindtotaal:",
		Page:         "Pagina",
		Generated:    "Gegenereerd:",
		Customer:     "Klant",
		Project:      "Project",
		Activity:     "Activiteit",
		Hours:        "Uren",
		Total:        "Totaal",
		AllEmployees: "Alle medewerkers",
		Date:         "Datum",
		Undeclarable: "Niet declarabel",
		Declarable:   "Declarabel",
		AllTime:      "Gehele periode",
		YearPrefix:   "Jaar",
		WeekLabel:    "Week",
		Rate:         "Tarief",
		Cost:         "Kosten",
		Invoice:      "Factuur",
		Standard:     "Standaard",
		MonthsFull:   [12]string{"januari", "februari", "maart", "april", "mei", "juni", "juli", "augustus", "september", "oktober", "november", "december"},
		MonthsAbbr:   [12]string{"jan", "feb", "mrt", "apr", "mei", "jun", "jul", "aug", "sep", "okt", "nov", "dec"},
		DaysFull:     [7]string{"maandag", "dinsdag", "woensdag", "donderdag", "vrijdag", "zaterdag", "zondag"},
		DaysAbbr:     [7]string{"ma", "di", "wo", "do", "vr", "za", "zo"},
		DMY:          true,
	},
	"de": {
		TimeReport:   "Zeitbericht",
		ColRef:       "Ref",
		ColTask:      "Aufgabe",
		ColAssign:    "Bearbeiter",
		ColTime:      "Zeit",
		Subtotal:     "Zwischensumme:",
		GrandTotal:   "Gesamtsumme:",
		Page:         "Seite",
		Generated:    "Erstellt:",
		Customer:     "Kunde",
		Project:      "Projekt",
		Activity:     "Aktivität",
		Hours:        "Stunden",
		Total:        "Gesamt",
		AllEmployees: "Alle Mitarbeiter",
		Date:         "Datum",
		Undeclarable: "Nicht abrechenbar",
		Declarable:   "Abrechenbar",
		AllTime:      "Gesamter Zeitraum",
		YearPrefix:   "Jahr",
		WeekLabel:    "Woche",
		Rate:         "Satz",
		Cost:         "Kosten",
		Invoice:      "Rechnung",
		Standard:     "Standard",
		MonthsFull:   [12]string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"},
		MonthsAbbr:   [12]string{"Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"},
		DaysFull:     [7]string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"},
		DaysAbbr:     [7]string{"Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"},
		DMY:          true,
	},
	"fr": {
		TimeReport:   "Rapport de temps",
		ColRef:       "Réf",
		ColTask:      "Tâche",
		ColAssign:    "Assignés",
		ColTime:      "Temps",
		Subtotal:     "Sous-total :",
		GrandTotal:   "Total général :",
		Page:         "Page",
		Generated:    "Généré le :",
		Customer:     "Client",
		Project:      "Projet",
		Activity:     "Activité",
		Hours:        "Heures",
		Total:        "Total",
		AllEmployees: "Tous les employés",
		Date:         "Date",
		Undeclarable: "Non facturable",
		Declarable:   "Facturable",
		AllTime:      "Toute la période",
		YearPrefix:   "Année",
		WeekLabel:    "Semaine",
		Rate:         "Tarif",
		Cost:         "Coût",
		Invoice:      "Facture",
		Standard:     "Standard",
		MonthsFull:   [12]string{"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"},
		MonthsAbbr:   [12]string{"janv.", "févr.", "mars", "avr.", "mai", "juin", "juil.", "août", "sept.", "oct.", "nov.", "déc."},
		DaysFull:     [7]string{"lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi", "dimanche"},
		DaysAbbr:     [7]string{"lun", "mar", "mer", "jeu", "ven", "sam", "dim"},
		DMY:          true,
	},
	"es": {
		TimeReport:   "Informe de tiempo",
		ColRef:       "Ref",
		ColTask:      "Tarea",
		ColAssign:    "Asignados",
		ColTime:      "Tiempo",
		Subtotal:     "Subtotal:",
		GrandTotal:   "Total general:",
		Page:         "Página",
		Generated:    "Generado el:",
		Customer:     "Cliente",
		Project:      "Proyecto",
		Activity:     "Actividad",
		Hours:        "Horas",
		Total:        "Total",
		AllEmployees: "Todos los empleados",
		Date:         "Fecha",
		Undeclarable: "No declarable",
		Declarable:   "Declarable",
		AllTime:      "Todo el período",
		YearPrefix:   "Año",
		WeekLabel:    "Semana",
		Rate:         "Tarifa",
		Cost:         "Coste",
		Invoice:      "Factura",
		Standard:     "Estándar",
		MonthsFull:   [12]string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"},
		MonthsAbbr:   [12]string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"},
		DaysFull:     [7]string{"lunes", "martes", "miércoles", "jueves", "viernes", "sábado", "domingo"},
		DaysAbbr:     [7]string{"lun", "mar", "mié", "jue", "vie", "sáb", "dom"},
		DMY:          true,
	},
	"da": {
		TimeReport:   "Tidsrapport",
		ColRef:       "Ref",
		ColTask:      "Opgave",
		ColAssign:    "Tildelt",
		ColTime:      "Tid",
		Subtotal:     "Subtotal:",
		GrandTotal:   "Samlet total:",
		Page:         "Side",
		Generated:    "Genereret:",
		Customer:     "Kunde",
		Project:      "Projekt",
		Activity:     "Aktivitet",
		Hours:        "Timer",
		Total:        "Total",
		AllEmployees: "Alle medarbejdere",
		Date:         "Dato",
		Undeclarable: "Ikke fakturerbar",
		Declarable:   "Fakturerbar",
		AllTime:      "Hele perioden",
		YearPrefix:   "År",
		WeekLabel:    "Uge",
		Rate:         "Sats",
		Cost:         "Omkostning",
		Invoice:      "Faktura",
		Standard:     "Standard",
		MonthsFull:   [12]string{"januar", "februar", "marts", "april", "maj", "juni", "juli", "august", "september", "oktober", "november", "december"},
		MonthsAbbr:   [12]string{"jan", "feb", "mar", "apr", "maj", "jun", "jul", "aug", "sep", "okt", "nov", "dec"},
		DaysFull:     [7]string{"mandag", "tirsdag", "onsdag", "torsdag", "fredag", "lørdag", "søndag"},
		DaysAbbr:     [7]string{"man", "tir", "ons", "tor", "fre", "lør", "søn"},
		DMY:          true,
	},
	"sv": {
		TimeReport:   "Tidsrapport",
		ColRef:       "Ref",
		ColTask:      "Uppgift",
		ColAssign:    "Tilldelad",
		ColTime:      "Tid",
		Subtotal:     "Delsumma:",
		GrandTotal:   "Totalsumma:",
		Page:         "Sida",
		Generated:    "Genererad:",
		Customer:     "Kund",
		Project:      "Projekt",
		Activity:     "Aktivitet",
		Hours:        "Timmar",
		Total:        "Totalt",
		AllEmployees: "Alla anställda",
		Date:         "Datum",
		Undeclarable: "Ej fakturerbar",
		Declarable:   "Fakturerbar",
		AllTime:      "Hela perioden",
		YearPrefix:   "År",
		WeekLabel:    "Vecka",
		Rate:         "Pris",
		Cost:         "Kostnad",
		Invoice:      "Faktura",
		Standard:     "Standard",
		MonthsFull:   [12]string{"januari", "februari", "mars", "april", "maj", "juni", "juli", "augusti", "september", "oktober", "november", "december"},
		MonthsAbbr:   [12]string{"jan", "feb", "mar", "apr", "maj", "jun", "jul", "aug", "sep", "okt", "nov", "dec"},
		DaysFull:     [7]string{"måndag", "tisdag", "onsdag", "torsdag", "fredag", "lördag", "söndag"},
		DaysAbbr:     [7]string{"mån", "tis", "ons", "tor", "fre", "lör", "sön"},
		DMY:          true,
	},
	"nb": {
		TimeReport:   "Tidsrapport",
		ColRef:       "Ref",
		ColTask:      "Oppgave",
		ColAssign:    "Tildelt",
		ColTime:      "Tid",
		Subtotal:     "Delsum:",
		GrandTotal:   "Totalsum:",
		Page:         "Side",
		Generated:    "Generert:",
		Customer:     "Kunde",
		Project:      "Prosjekt",
		Activity:     "Aktivitet",
		Hours:        "Timer",
		Total:        "Totalt",
		AllEmployees: "Alle ansatte",
		Date:         "Dato",
		Undeclarable: "Ikke fakturerbar",
		Declarable:   "Fakturerbar",
		AllTime:      "Hele perioden",
		YearPrefix:   "År",
		WeekLabel:    "Uke",
		Rate:         "Sats",
		Cost:         "Kostnad",
		Invoice:      "Faktura",
		Standard:     "Standard",
		MonthsFull:   [12]string{"januar", "februar", "mars", "april", "mai", "juni", "juli", "august", "september", "oktober", "november", "desember"},
		MonthsAbbr:   [12]string{"jan", "feb", "mar", "apr", "mai", "jun", "jul", "aug", "sep", "okt", "nov", "des"},
		DaysFull:     [7]string{"mandag", "tirsdag", "onsdag", "torsdag", "fredag", "lørdag", "søndag"},
		DaysAbbr:     [7]string{"man", "tir", "ons", "tor", "fre", "lør", "søn"},
		DMY:          true,
	},
	"fi": {
		TimeReport:   "Tuntiraportti",
		ColRef:       "Viite",
		ColTask:      "Tehtävä",
		ColAssign:    "Vastuuhenkilöt",
		ColTime:      "Aika",
		Subtotal:     "Välisumma:",
		GrandTotal:   "Kokonaissumma:",
		Page:         "Sivu",
		Generated:    "Luotu:",
		Customer:     "Asiakas",
		Project:      "Projekti",
		Activity:     "Toiminto",
		Hours:        "Tunnit",
		Total:        "Yhteensä",
		AllEmployees: "Kaikki työntekijät",
		Date:         "Päivämäärä",
		Undeclarable: "Ei laskutettava",
		Declarable:   "Laskutettava",
		AllTime:      "Koko aika",
		YearPrefix:   "Vuosi",
		WeekLabel:    "Viikko",
		Rate:         "Tuntihinta",
		Cost:         "Kustannus",
		Invoice:      "Lasku",
		Standard:     "Perusmäärä",
		MonthsFull:   [12]string{"tammikuu", "helmikuu", "maaliskuu", "huhtikuu", "toukokuu", "kesäkuu", "heinäkuu", "elokuu", "syyskuu", "lokakuu", "marraskuu", "joulukuu"},
		MonthsAbbr:   [12]string{"tammi", "helmi", "maalis", "huhti", "touko", "kesä", "heinä", "elo", "syys", "loka", "marras", "joulu"},
		DaysFull:     [7]string{"maanantai", "tiistai", "keskiviikko", "torstai", "perjantai", "lauantai", "sunnuntai"},
		DaysAbbr:     [7]string{"ma", "ti", "ke", "to", "pe", "la", "su"},
		DMY:          true,
	},
	"is": {
		TimeReport:   "Tímaskýrsla",
		ColRef:       "Tilvísun",
		ColTask:      "Verkefni",
		ColAssign:    "Úthlutað til",
		ColTime:      "Tími",
		Subtotal:     "Millisamtala:",
		GrandTotal:   "Heildarsamtala:",
		Page:         "Síða",
		Generated:    "Útbúið:",
		Customer:     "Viðskiptavinur",
		Project:      "Verkefni",
		Activity:     "Verkþáttur",
		Hours:        "Klukkustundir",
		Total:        "Samtals",
		AllEmployees: "Allir starfsmenn",
		Date:         "Dagsetning",
		Undeclarable: "Ekki reikningshæft",
		Declarable:   "Reikningshæft",
		AllTime:      "Allur tíminn",
		YearPrefix:   "Ár",
		WeekLabel:    "Vika",
		Rate:         "Gjald",
		Cost:         "Kostnaður",
		Invoice:      "Reikningur",
		Standard:     "Staðlað",
		MonthsFull:   [12]string{"janúar", "febrúar", "mars", "apríl", "maí", "júní", "júlí", "ágúst", "september", "október", "nóvember", "desember"},
		MonthsAbbr:   [12]string{"jan", "feb", "mar", "apr", "maí", "jún", "júl", "ágú", "sep", "okt", "nóv", "des"},
		DaysFull:     [7]string{"mánudagur", "þriðjudagur", "miðvikudagur", "fimmtudagur", "föstudagur", "laugardagur", "sunnudagur"},
		DaysAbbr:     [7]string{"mán", "þri", "mið", "fim", "fös", "lau", "sun"},
		DMY:          true,
	},
	"pt": {
		TimeReport:   "Relatório de tempo",
		ColRef:       "Ref",
		ColTask:      "Tarefa",
		ColAssign:    "Atribuídos",
		ColTime:      "Tempo",
		Subtotal:     "Subtotal:",
		GrandTotal:   "Total geral:",
		Page:         "Página",
		Generated:    "Gerado em:",
		Customer:     "Cliente",
		Project:      "Projeto",
		Activity:     "Atividade",
		Hours:        "Horas",
		Total:        "Total",
		AllEmployees: "Todos os funcionários",
		Date:         "Data",
		Undeclarable: "Não declarável",
		Declarable:   "Declarável",
		AllTime:      "Todo o período",
		YearPrefix:   "Ano",
		WeekLabel:    "Semana",
		Rate:         "Taxa",
		Cost:         "Custo",
		Invoice:      "Fatura",
		Standard:     "Padrão",
		MonthsFull:   [12]string{"janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"},
		MonthsAbbr:   [12]string{"jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"},
		DaysFull:     [7]string{"segunda-feira", "terça-feira", "quarta-feira", "quinta-feira", "sexta-feira", "sábado", "domingo"},
		DaysAbbr:     [7]string{"seg", "ter", "qua", "qui", "sex", "sáb", "dom"},
		DMY:          true,
	},
	"it": {
		TimeReport:   "Rapporto ore",
		ColRef:       "Rif",
		ColTask:      "Attività",
		ColAssign:    "Assegnatari",
		ColTime:      "Tempo",
		Subtotal:     "Subtotale:",
		GrandTotal:   "Totale generale:",
		Page:         "Pagina",
		Generated:    "Generato il:",
		Customer:     "Cliente",
		Project:      "Progetto",
		Activity:     "Attività",
		Hours:        "Ore",
		Total:        "Totale",
		AllEmployees: "Tutti i dipendenti",
		Date:         "Data",
		Undeclarable: "Non dichiarabile",
		Declarable:   "Dichiarabile",
		AllTime:      "Tutto il periodo",
		YearPrefix:   "Anno",
		WeekLabel:    "Settimana",
		Rate:         "Tariffa",
		Cost:         "Costo",
		Invoice:      "Fattura",
		Standard:     "Standard",
		MonthsFull:   [12]string{"gennaio", "febbraio", "marzo", "aprile", "maggio", "giugno", "luglio", "agosto", "settembre", "ottobre", "novembre", "dicembre"},
		MonthsAbbr:   [12]string{"gen", "feb", "mar", "apr", "mag", "giu", "lug", "ago", "set", "ott", "nov", "dic"},
		DaysFull:     [7]string{"lunedì", "martedì", "mercoledì", "giovedì", "venerdì", "sabato", "domenica"},
		DaysAbbr:     [7]string{"lun", "mar", "mer", "gio", "ven", "sab", "dom"},
		DMY:          true,
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

// enMonthsForReorder lists English month names in the order they should be
// tried when scanning for "MonthName DayNumber" patterns. Full names come
// first so that e.g. "September" is matched before "Sep".
var enMonthsForReorder = []string{
	"September", "November", "December", "February",
	"October", "January", "August", "March", "April", "June", "July",
	// "May" is identical in full and abbreviated forms.
	"May",
	// Abbreviated forms that differ from the full name.
	"Jan", "Feb", "Mar", "Apr", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

// reorderMonthDay rewrites every "MonthName DD" occurrence in s to "DD MonthName"
// (DMY order). It operates on English source strings produced by Go's
// time.Format, so only ASCII boundaries are needed. Four-digit year numbers
// (YYYY) are never reordered because the digit-count check stops at two digits.
func reorderMonthDay(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		matched := false
		for _, month := range enMonthsForReorder {
			mlen := len(month)
			// Need "Month " at position i.
			if !strings.HasPrefix(s[i:], month+" ") {
				continue
			}
			// Not preceded by an ASCII letter (avoids matching inside a word).
			if i > 0 && ((s[i-1] >= 'a' && s[i-1] <= 'z') || (s[i-1] >= 'A' && s[i-1] <= 'Z')) {
				continue
			}
			// Collect 1-2 digit characters after "Month ".
			j := i + mlen + 1
			k := j
			for k < len(s) && s[k] >= '0' && s[k] <= '9' {
				k++
			}
			digits := k - j
			if digits == 0 || digits > 2 {
				// No digits, or too many (e.g. a year "2026") — not a day number.
				break
			}
			// Must not be followed by another digit (guards against matching "20" in "2026").
			if k < len(s) && s[k] >= '0' && s[k] <= '9' {
				break
			}
			// Reorder: emit "DD Month" instead of "Month DD".
			b.WriteString(s[j:k]) // day digits
			b.WriteByte(' ')
			b.WriteString(month)
			i = k
			matched = true
			break
		}
		if !matched {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}

// replaceWord replaces each occurrence of old in s that is not immediately
// preceded or followed by an ASCII letter (whole-word match). This prevents
// replacing prefixes inside already-translated words (e.g. "Jan" inside
// "Januar"). Source strings from Go's time.Format are always ASCII, so
// byte-level boundary checks are sufficient.
func replaceWord(s, old, rep string) string {
	if old == "" || old == rep {
		return s
	}
	n := len(old)
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], old) {
			prevLetter := i > 0 && ((s[i-1] >= 'a' && s[i-1] <= 'z') || (s[i-1] >= 'A' && s[i-1] <= 'Z'))
			nextLetter := (i+n) < len(s) && ((s[i+n] >= 'a' && s[i+n] <= 'z') || (s[i+n] >= 'A' && s[i+n] <= 'Z'))
			if !prevLetter && !nextLetter {
				b.WriteString(rep)
				i += n
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// pdfTranslateLabel rewrites English month/day names and period keywords in a
// label string produced by Go's time.Format (which is always English). It is
// safe to call on customer/project names because those are never identical to
// English month or day names as whole words.
func pdfTranslateLabel(s string, tr pdfI18n) string {
	// Replace period keywords (exact, case-sensitive).
	if tr.AllTime != "" {
		s = strings.ReplaceAll(s, "All Time", tr.AllTime)
	}
	if tr.YearPrefix != "" && tr.YearPrefix != "Year" {
		s = replaceWord(s, "Year", tr.YearPrefix)
	}
	if tr.WeekLabel != "" && tr.WeekLabel != "Week" {
		s = replaceWord(s, "Week", tr.WeekLabel)
	}
	// Reorder "MonthName DD" → "DD MonthName" for DMY locales before translating names.
	if tr.DMY {
		s = reorderMonthDay(s)
	}
	// Full month names (longest first to avoid prefix clashes like Sep/September).
	enFull := [12]string{
		"September", "November", "December", "February",
		"October", "January", "August", "March",
		"April", "June", "July", "May",
	}
	trFull := [12]string{
		tr.MonthsFull[8], tr.MonthsFull[10], tr.MonthsFull[11], tr.MonthsFull[1],
		tr.MonthsFull[9], tr.MonthsFull[0], tr.MonthsFull[7], tr.MonthsFull[2],
		tr.MonthsFull[3], tr.MonthsFull[5], tr.MonthsFull[6], tr.MonthsFull[4],
	}
	for i, en := range enFull {
		if trFull[i] != "" && trFull[i] != en {
			s = replaceWord(s, en, trFull[i])
		}
	}
	// Abbreviated month names (same length-descending order where needed).
	enAbbr := [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	for i, en := range enAbbr {
		if tr.MonthsAbbr[i] != "" && tr.MonthsAbbr[i] != en {
			s = replaceWord(s, en, tr.MonthsAbbr[i])
		}
	}
	// Full day names.
	enDays := [7]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for i, en := range enDays {
		if tr.DaysFull[i] != "" && tr.DaysFull[i] != en {
			s = replaceWord(s, en, tr.DaysFull[i])
		}
	}
	return s
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

// renderLogoIntoPDF places the logo image at (x, y) in the PDF, 18 mm tall.
// rawBytes is the file content, ext is the lowercase file extension.
// Returns true if the image was placed successfully.
func renderLogoIntoPDF(pdf *gofpdf.Fpdf, rawBytes []byte, ext string, x, y float64) bool {
	if ext == ".svg" {
		pngBuf, err := svgToPNGFromBytes(rawBytes, 18)
		if err != nil {
			return false
		}
		opts := gofpdf.ImageOptions{ImageType: "PNG"}
		pdf.RegisterImageOptionsReader("_logo", opts, &pngBuf)
		pdf.ImageOptions("_logo", x, y, 0, 18, false, opts, 0, "")
		return pdf.Error() == nil
	}
	imgType := ""
	switch ext {
	case ".jpg", ".jpeg":
		imgType = "JPG"
	case ".png":
		imgType = "PNG"
	case ".gif":
		imgType = "GIF"
	}
	if imgType == "" {
		return false
	}
	var imgReader *bytes.Reader
	if imgType == "PNG" {
		// Composite over white to strip alpha — gofpdf cannot handle transparent PNGs.
		if src, err := png.Decode(bytes.NewReader(rawBytes)); err == nil {
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
	if pdf.Error() != nil {
		return false
	}
	pdf.ImageOptions("_logo", x, y, 0, 18, false, opts, 0, "")
	return pdf.Error() == nil
}

// svgToPNGFromBytes rasterizes SVG bytes to a PNG held in a bytes.Buffer.
func svgToPNGFromBytes(data []byte, heightMM float64) (bytes.Buffer, error) {
	return svgRasterize(bytes.NewReader(data), heightMM)
}

func svgRasterize(r interface{ Read([]byte) (int, error) }, heightMM float64) (bytes.Buffer, error) {
	icon, err := oksvg.ReadIconStream(r)
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

	// Translate any English month/period words in the period label.
	localPeriodLabel := pdfTranslateLabel(report.PeriodLabel, tr)

	// PDF metadata.
	title := tr.TimeReport + " — " + localPeriodLabel
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

	// Try to load company logo. SVGs are rasterized to PNG; raster formats streamed directly.
	logoY := pdfMargin
	logoLoaded := false
	if rawBytes, ext, ok := resolveLogoBytes(report.CompanyLogo); ok {
		logoLoaded = renderLogoIntoPDF(pdf, rawBytes, ext, pdfMargin, logoY)
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
	pdf.CellFormat(textW, 5.5, localPeriodLabel, "", 2, "L", false, 0, "")
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

	if c.Query("base64") == "1" {
		c.JSON(http.StatusOK, gin.H{"data": base64.StdEncoding.EncodeToString(buf.Bytes())})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
