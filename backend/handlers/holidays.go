package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/middleware"
	"github.com/tonk/warmdesk/models"
)

// ── Easter (Gregorian, anonymous algorithm) ───────────────────────────────

func easterSunday(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := (h+l-7*m+114)%31 + 1
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// ── Helpers ───────────────────────────────────────────────────────────────

func hd(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func easterOffset(easter time.Time, days int) time.Time {
	return easter.AddDate(0, 0, days)
}

// nthWeekday returns the nth (1-based) or last (-1) occurrence of wd in a month.
func nthWeekday(year int, month time.Month, wd time.Weekday, n int) time.Time {
	if n > 0 {
		d := hd(year, month, 1)
		for d.Weekday() != wd {
			d = d.AddDate(0, 0, 1)
		}
		return d.AddDate(0, 0, (n-1)*7)
	}
	// Last occurrence
	d := hd(year, month+1, 1).AddDate(0, 0, -1) // last day of month
	for d.Weekday() != wd {
		d = d.AddDate(0, 0, -1)
	}
	return d
}

// firstWeekdayOnOrAfter returns the first occurrence of wd on or after t.
func firstWeekdayOnOrAfter(t time.Time, wd time.Weekday) time.Time {
	for t.Weekday() != wd {
		t = t.AddDate(0, 0, 1)
	}
	return t
}

// ── Holiday definitions per locale ────────────────────────────────────────

type holidayDef struct {
	Date time.Time
	Name string
}

func nlHolidays(year int, easter time.Time) []holidayDef {
	kingsDay := hd(year, time.April, 27)
	if kingsDay.Weekday() == time.Sunday {
		kingsDay = hd(year, time.April, 26)
	}
	return []holidayDef{
		{hd(year, time.January, 1), "Nieuwjaarsdag"},
		{easterOffset(easter, -2), "Goede Vrijdag"},
		{easter, "Eerste Paasdag"},
		{easterOffset(easter, 1), "Tweede Paasdag"},
		{kingsDay, "Koningsdag"},
		{hd(year, time.May, 5), "Bevrijdingsdag"},
		{easterOffset(easter, 39), "Hemelvaartsdag"},
		{easterOffset(easter, 49), "Eerste Pinksterdag"},
		{easterOffset(easter, 50), "Tweede Pinksterdag"},
		{hd(year, time.December, 25), "Eerste Kerstdag"},
		{hd(year, time.December, 26), "Tweede Kerstdag"},
	}
}

func deHolidays(year int, easter time.Time) []holidayDef {
	return []holidayDef{
		{hd(year, time.January, 1), "Neujahr"},
		{easterOffset(easter, -2), "Karfreitag"},
		{easter, "Ostersonntag"},
		{easterOffset(easter, 1), "Ostermontag"},
		{hd(year, time.May, 1), "Tag der Arbeit"},
		{easterOffset(easter, 39), "Christi Himmelfahrt"},
		{easterOffset(easter, 49), "Pfingstsonntag"},
		{easterOffset(easter, 50), "Pfingstmontag"},
		{easterOffset(easter, 60), "Fronleichnam"},
		{hd(year, time.October, 3), "Tag der Deutschen Einheit"},
		{hd(year, time.December, 25), "Erster Weihnachtstag"},
		{hd(year, time.December, 26), "Zweiter Weihnachtstag"},
	}
}

func frHolidays(year int, easter time.Time) []holidayDef {
	return []holidayDef{
		{hd(year, time.January, 1), "Jour de l'An"},
		{easterOffset(easter, 1), "Lundi de Pâques"},
		{hd(year, time.May, 1), "Fête du Travail"},
		{hd(year, time.May, 8), "Victoire 1945"},
		{easterOffset(easter, 39), "Ascension"},
		{easterOffset(easter, 50), "Lundi de Pentecôte"},
		{hd(year, time.July, 14), "Fête Nationale"},
		{hd(year, time.August, 15), "Assomption"},
		{hd(year, time.November, 1), "Toussaint"},
		{hd(year, time.November, 11), "Armistice"},
		{hd(year, time.December, 25), "Noël"},
	}
}

func esHolidays(year int, easter time.Time) []holidayDef {
	return []holidayDef{
		{hd(year, time.January, 1), "Año Nuevo"},
		{hd(year, time.January, 6), "Epifanía del Señor"},
		{easterOffset(easter, -2), "Viernes Santo"},
		{hd(year, time.May, 1), "Día del Trabajo"},
		{hd(year, time.August, 15), "Asunción de la Virgen"},
		{hd(year, time.October, 12), "Fiesta Nacional de España"},
		{hd(year, time.November, 1), "Todos los Santos"},
		{hd(year, time.December, 6), "Día de la Constitución"},
		{hd(year, time.December, 8), "Inmaculada Concepción"},
		{hd(year, time.December, 25), "Navidad"},
	}
}

func daHolidays(year int, easter time.Time) []holidayDef {
	return []holidayDef{
		{hd(year, time.January, 1), "Nytårsdag"},
		{easterOffset(easter, -3), "Skærtorsdag"},
		{easterOffset(easter, -2), "Langfredag"},
		{easter, "Påskedag"},
		{easterOffset(easter, 1), "Anden påskedag"},
		{easterOffset(easter, 39), "Kristi himmelfartsdag"},
		{easterOffset(easter, 49), "Pinsedag"},
		{easterOffset(easter, 50), "Anden pinsedag"},
		{hd(year, time.June, 5), "Grundlovsdag"},
		{hd(year, time.December, 24), "Juleaftensdag"},
		{hd(year, time.December, 25), "Juledag"},
		{hd(year, time.December, 26), "Anden juledag"},
	}
}

func svHolidays(year int, easter time.Time) []holidayDef {
	// Midsommarafton: Friday between June 19–25
	midsummerEve := firstWeekdayOnOrAfter(hd(year, time.June, 19), time.Friday)
	// Alla helgons dag: Saturday between Oct 31 – Nov 6
	allSaints := firstWeekdayOnOrAfter(hd(year, time.October, 31), time.Saturday)
	return []holidayDef{
		{hd(year, time.January, 1), "Nyårsdagen"},
		{hd(year, time.January, 6), "Trettondedag jul"},
		{easterOffset(easter, -2), "Långfredag"},
		{easter, "Påskdagen"},
		{easterOffset(easter, 1), "Annandag påsk"},
		{hd(year, time.May, 1), "Första maj"},
		{easterOffset(easter, 39), "Kristi himmelsfärdsdag"},
		{hd(year, time.June, 6), "Sveriges nationaldag"},
		{midsummerEve, "Midsommarafton"},
		{midsummerEve.AddDate(0, 0, 1), "Midsommardagen"},
		{allSaints, "Alla helgons dag"},
		{hd(year, time.December, 24), "Julafton"},
		{hd(year, time.December, 25), "Juldagen"},
		{hd(year, time.December, 26), "Annandag jul"},
	}
}

func nbHolidays(year int, easter time.Time) []holidayDef {
	return []holidayDef{
		{hd(year, time.January, 1), "Nyttårsdag"},
		{easterOffset(easter, -3), "Skjærtorsdag"},
		{easterOffset(easter, -2), "Langfredag"},
		{easter, "Første påskedag"},
		{easterOffset(easter, 1), "Andre påskedag"},
		{hd(year, time.May, 1), "Arbeidernes dag"},
		{hd(year, time.May, 17), "Grunnlovsdagen"},
		{easterOffset(easter, 39), "Kristi himmelfartsdag"},
		{easterOffset(easter, 49), "Første pinsedag"},
		{easterOffset(easter, 50), "Andre pinsedag"},
		{hd(year, time.December, 25), "Første juledag"},
		{hd(year, time.December, 26), "Andre juledag"},
	}
}

func fiHolidays(year int, easter time.Time) []holidayDef {
	// Juhannusaatto: Friday between June 19–25
	midsummerEve := firstWeekdayOnOrAfter(hd(year, time.June, 19), time.Friday)
	// Pyhäinpäivä: Saturday between Oct 31 – Nov 6
	allSaints := firstWeekdayOnOrAfter(hd(year, time.October, 31), time.Saturday)
	return []holidayDef{
		{hd(year, time.January, 1), "Uudenvuodenpäivä"},
		{hd(year, time.January, 6), "Loppiainen"},
		{easterOffset(easter, -2), "Pitkäperjantai"},
		{easter, "Pääsiäispäivä"},
		{easterOffset(easter, 1), "Toinen pääsiäispäivä"},
		{hd(year, time.May, 1), "Vappu"},
		{easterOffset(easter, 39), "Helatorstai"},
		{midsummerEve, "Juhannusaatto"},
		{midsummerEve.AddDate(0, 0, 1), "Juhannuspäivä"},
		{allSaints, "Pyhäinpäivä"},
		{hd(year, time.December, 6), "Itsenäisyyspäivä"},
		{hd(year, time.December, 24), "Jouluaatto"},
		{hd(year, time.December, 25), "Joulupäivä"},
		{hd(year, time.December, 26), "Tapaninpäivä"},
	}
}

func isHolidays(year int, easter time.Time) []holidayDef {
	// Sumardagurinn fyrsti: first Thursday on or after April 19
	firstSummer := firstWeekdayOnOrAfter(hd(year, time.April, 19), time.Thursday)
	// Frídagur verslunarmanna: first Monday in August
	commerceDay := nthWeekday(year, time.August, time.Monday, 1)
	return []holidayDef{
		{hd(year, time.January, 1), "Nýársdagur"},
		{easterOffset(easter, -3), "Skírdagur"},
		{easterOffset(easter, -2), "Föstudagurinn langi"},
		{easter, "Páskadagur"},
		{easterOffset(easter, 1), "Annar í páskum"},
		{firstSummer, "Sumardagurinn fyrsti"},
		{hd(year, time.May, 1), "Verkalýðsdagurinn"},
		{easterOffset(easter, 39), "Uppstigningardagur"},
		{easterOffset(easter, 49), "Hvítasunnudagur"},
		{easterOffset(easter, 50), "Annar í hvítasunnu"},
		{hd(year, time.June, 17), "Þjóðhátíðardagurinn"},
		{commerceDay, "Frídagur verslunarmanna"},
		{hd(year, time.December, 24), "Aðfangadagur Jóla"},
		{hd(year, time.December, 25), "Jóladagur"},
		{hd(year, time.December, 26), "Annar í jólum"},
		{hd(year, time.December, 31), "Gamlársdagur"},
	}
}

func ptHolidays(year int, easter time.Time) []holidayDef {
	return []holidayDef{
		{hd(year, time.January, 1), "Ano Novo"},
		{easterOffset(easter, -2), "Sexta-Feira Santa"},
		{easter, "Páscoa"},
		{hd(year, time.April, 25), "Dia da Liberdade"},
		{hd(year, time.May, 1), "Dia do Trabalhador"},
		{easterOffset(easter, 60), "Corpus Christi"},
		{hd(year, time.June, 10), "Dia de Portugal"},
		{hd(year, time.August, 15), "Assunção de Nossa Senhora"},
		{hd(year, time.October, 5), "Implantação da República"},
		{hd(year, time.November, 1), "Dia de Todos os Santos"},
		{hd(year, time.December, 1), "Restauração da Independência"},
		{hd(year, time.December, 8), "Imaculada Conceição"},
		{hd(year, time.December, 25), "Natal"},
	}
}

func itHolidays(year int, easter time.Time) []holidayDef {
	return []holidayDef{
		{hd(year, time.January, 1), "Capodanno"},
		{hd(year, time.January, 6), "Epifania"},
		{easter, "Domenica di Pasqua"},
		{easterOffset(easter, 1), "Lunedì dell'Angelo"},
		{hd(year, time.April, 25), "Festa della Liberazione"},
		{hd(year, time.May, 1), "Festa del Lavoro"},
		{hd(year, time.June, 2), "Festa della Repubblica"},
		{hd(year, time.August, 15), "Ferragosto"},
		{hd(year, time.November, 1), "Tutti i Santi"},
		{hd(year, time.December, 8), "Immacolata Concezione"},
		{hd(year, time.December, 25), "Natale"},
		{hd(year, time.December, 26), "Santo Stefano"},
	}
}

func enHolidays(year int, easter time.Time) []holidayDef {
	// England & Wales bank holidays
	earlyMay := nthWeekday(year, time.May, time.Monday, 1)
	springBH := nthWeekday(year, time.May, time.Monday, -1)
	summerBH := nthWeekday(year, time.August, time.Monday, -1)
	return []holidayDef{
		{hd(year, time.January, 1), "New Year's Day"},
		{easterOffset(easter, -2), "Good Friday"},
		{easterOffset(easter, 1), "Easter Monday"},
		{earlyMay, "Early May Bank Holiday"},
		{springBH, "Spring Bank Holiday"},
		{summerBH, "Summer Bank Holiday"},
		{hd(year, time.December, 25), "Christmas Day"},
		{hd(year, time.December, 26), "Boxing Day"},
	}
}

var holidaysByLocale = map[string]func(int, time.Time) []holidayDef{
	"nl": nlHolidays,
	"de": deHolidays,
	"fr": frHolidays,
	"es": esHolidays,
	"da": daHolidays,
	"sv": svHolidays,
	"nb": nbHolidays,
	"fi": fiHolidays,
	"is": isHolidays,
	"pt": ptHolidays,
	"it": itHolidays,
	"en": enHolidays,
}

// ── Handler ───────────────────────────────────────────────────────────────

type addHolidaysRequest struct {
	Year   int    `json:"year" binding:"required,min=2000,max=2100"`
	Locale string `json:"locale" binding:"required"`
	UserID *uint  `json:"user_id,omitempty"`
}

type holidayResult struct {
	Date  string `json:"date"`
	Name  string `json:"name"`
	Added bool   `json:"added"`
}

type addHolidaysResponse struct {
	Added    int             `json:"added"`
	Skipped  int             `json:"skipped"`
	Holidays []holidayResult `json:"holidays"`
}

func AddHolidays(c *gin.Context) {
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetGlobalRole(c)

	var req addHolidaysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	gen, ok := holidaysByLocale[req.Locale]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported locale"})
		return
	}

	targetUserID := userID
	if req.UserID != nil && globalRole == "admin" {
		targetUserID = *req.UserID
	}

	easter := easterSunday(req.Year)
	holidays := gen(req.Year, easter)

	var results []holidayResult
	added, skipped := 0, 0

	for _, h := range holidays {
		var count int64
		database.DB.Model(&models.TimeEntry{}).
			Where("user_id = ? AND date = ?", targetUserID, h.Date).
			Count(&count)

		if count > 0 {
			results = append(results, holidayResult{Date: h.Date.Format("2006-01-02"), Name: h.Name, Added: false})
			skipped++
			continue
		}

		entry := models.TimeEntry{
			UserID:      targetUserID,
			Date:        h.Date,
			Minutes:     0,
			Description: h.Name,
			IsHoliday:   true,
		}
		if err := database.DB.Create(&entry).Error; err == nil {
			results = append(results, holidayResult{Date: h.Date.Format("2006-01-02"), Name: h.Name, Added: true})
			added++
		}
	}

	c.JSON(http.StatusOK, addHolidaysResponse{Added: added, Skipped: skipped, Holidays: results})
}
