// Seed populates the database with demo projects, users, cards, comments, and more.
//
// Usage (run from the backend/ directory):
//
//	go run ./cmd/seed
//	go run ./cmd/seed --config /path/to/warmdesk.yaml
//	go run ./cmd/seed --reset   # drop all demo data first, then re-seed
//
// The script is idempotent: it exits early when it detects that the demo data
// is already present (looks for username "demo.admin").
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
	"github.com/tonk/warmdesk/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

// ─── helpers ────────────────────────────────────────────────────────────────

func must(err error) {
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
}

func hashPassword(plain string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	must(err)
	return string(h)
}

func ptr[T any](v T) *T { return &v }

func days(n int) *time.Time {
	t := time.Now().UTC().AddDate(0, 0, n).Truncate(24 * time.Hour)
	return &t
}

// ─── demo data definitions ──────────────────────────────────────────────────

type seedProject struct {
	name      string
	slug      string
	prefix    string
	color     string
	avatar    string
	desc      string
	boardType string // "kanban" (default) or "scrum"
	columns   []string
	labels    []struct{ name, color string }
}

var demoProjects = []seedProject{
	{
		name:      "Website Redesign",
		slug:      "website-redesign",
		prefix:    "WEB",
		color:     "#6366f1",
		avatar:    "https://api.dicebear.com/9.x/shapes/svg?seed=Website-Redesign&backgroundColor=dbeafe,e9d5ff",
		boardType: "kanban",
		desc:      "Full redesign of the marketing website — new brand, new stack, new speed.",
		columns:   []string{"Backlog", "In Progress", "Review", "Done"},
		labels: []struct{ name, color string }{
			{"Bug", "#ef4444"}, {"Feature", "#3b82f6"},
			{"Design", "#8b5cf6"}, {"Content", "#10b981"},
		},
	},
	{
		name:      "Mobile App v2",
		slug:      "mobile-app-v2",
		prefix:    "MOB",
		color:     "#f59e0b",
		avatar:    "https://api.dicebear.com/9.x/shapes/svg?seed=Mobile-App-v2&backgroundColor=fef3c7,fee2e2",
		boardType: "kanban",
		desc:      "Ground-up rewrite of the iOS and Android apps with a shared React Native core.",
		columns:   []string{"Ideas", "Development", "Testing", "Released"},
		labels: []struct{ name, color string }{
			{"Bug", "#ef4444"}, {"Enhancement", "#3b82f6"},
			{"iOS", "#0ea5e9"}, {"Android", "#22c55e"},
		},
	},
	{
		name:      "DevOps & Infrastructure",
		slug:      "devops-infra",
		prefix:    "INF",
		color:     "#10b981",
		avatar:    "https://api.dicebear.com/9.x/shapes/svg?seed=DevOps-Infrastructure&backgroundColor=dcfce7,bfdbfe",
		boardType: "kanban",
		desc:      "Kubernetes migration, monitoring, backups, and everything that keeps the lights on.",
		columns:   []string{"Todo", "In Progress", "Done"},
		labels: []struct{ name, color string }{
			{"Critical", "#ef4444"}, {"Enhancement", "#3b82f6"},
			{"Monitoring", "#f59e0b"}, {"Security", "#8b5cf6"},
		},
	},
	{
		name:      "Product Platform",
		slug:      "product-platform",
		prefix:    "PLT",
		color:     "#8b5cf6",
		avatar:    "https://api.dicebear.com/9.x/shapes/svg?seed=Product-Platform&backgroundColor=e9d5ff,dbeafe",
		boardType: "scrum",
		desc:      "The core SaaS platform — built sprint by sprint with a cross-functional team.",
		columns:   []string{"To Do", "In Progress", "In Review", "Done"},
		labels: []struct{ name, color string }{
			{"Bug", "#ef4444"}, {"Feature", "#3b82f6"},
			{"Enhancement", "#f59e0b"}, {"Tech Debt", "#6b7280"},
		},
	},
	{
		name:      "API Platform",
		slug:      "api-platform",
		prefix:    "API",
		color:     "#06b6d4",
		avatar:    "https://api.dicebear.com/9.x/shapes/svg?seed=API-Platform&backgroundColor=cffafe,e0f2fe",
		boardType: "scrum",
		desc:      "Developer-facing REST API — designed API-first, built sprint by sprint with a focus on stability and developer experience.",
		columns:   []string{"To Do", "In Progress", "In Review", "Done"},
		labels: []struct{ name, color string }{
			{"Bug", "#ef4444"}, {"Feature", "#3b82f6"},
			{"Enhancement", "#f59e0b"}, {"Security", "#8b5cf6"},
		},
	},
	{
		name:      "Marketing Campaigns",
		slug:      "marketing",
		prefix:    "MKT",
		color:     "#f43f5e",
		avatar:    "https://api.dicebear.com/9.x/shapes/svg?seed=Marketing-Campaigns&backgroundColor=fce7f3,fef3c7",
		boardType: "kanban",
		desc:      "Campaign planning, content production, and launch coordination for the marketing team.",
		columns:   []string{"Ideas", "Planned", "In Progress", "Published"},
		labels: []struct{ name, color string }{
			{"Campaign", "#f43f5e"}, {"Content", "#10b981"},
			{"Social", "#0ea5e9"}, {"Email", "#f59e0b"},
		},
	},
}

// ─── main ────────────────────────────────────────────────────────────────────

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	configPath := flag.String("config", "", "path to warmdesk.yaml (optional)")
	reset := flag.Bool("reset", false, "remove existing demo data before seeding")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	cfg := config.Load(*configPath)
	cfg.DBLog = "silent" // keep seed output readable
	must(database.Init(cfg))
	db := database.DB

	// Guard: already seeded?
	var existing models.User
	if err := db.Where("username = ?", "demo.admin").First(&existing).Error; err == nil {
		if !*reset {
			fmt.Println("✓ Demo data already present (username 'demo.admin' found).")
			fmt.Println("  Run with --reset to wipe and re-seed.")
			return
		}
		fmt.Println("⚠  --reset: removing existing demo data…")
		removeDemoData(db)
	}

	fmt.Println("🌱 Seeding demo data…")
	fmt.Println()

	// ── 0. System settings ────────────────────────────────────────────────────
	fmt.Println("→ Configuring system settings…")
	defaultColumns := "Backlog\nIn Progress\nTest & Review\nTo Production"
	if r := db.Model(&models.SystemSetting{}).Where("key = ?", "default_columns").Update("value", defaultColumns); r.RowsAffected == 0 {
		must(db.Create(&models.SystemSetting{Key: "default_columns", Value: defaultColumns}).Error)
	}
	defaultLabels := "Bug\nFeature\nDesign\nContent"
	if r := db.Model(&models.SystemSetting{}).Where("key = ?", "default_labels").Update("value", defaultLabels); r.RowsAffected == 0 {
		must(db.Create(&models.SystemSetting{Key: "default_labels", Value: defaultLabels}).Error)
	}
	if r := db.Model(&models.SystemSetting{}).Where("key = ?", "company_name").Update("value", "WarmDesk Company"); r.RowsAffected == 0 {
		must(db.Create(&models.SystemSetting{Key: "company_name", Value: "WarmDesk Company"}).Error)
	}
	if r := db.Model(&models.SystemSetting{}).Where("key = ?", "company_logo").Update("value", "/logo.svg"); r.RowsAffected == 0 {
		must(db.Create(&models.SystemSetting{Key: "company_logo", Value: "/logo.svg"}).Error)
	}
	if r := db.Model(&models.SystemSetting{}).Where("key = ?", "login_branding_enabled").Update("value", "true"); r.RowsAffected == 0 {
		must(db.Create(&models.SystemSetting{Key: "login_branding_enabled", Value: "true"}).Error)
	}

	// ── 1. Users ──────────────────────────────────────────────────────────────
	fmt.Println("→ Creating users…")

	// Ton Kersten — real system admin, excluded from --reset
	var tonk models.User
	if err := db.Where("username = ?", "tonk").First(&tonk).Error; err != nil {
		tonk = models.User{
			Email: "tonk@smartowl.nl", Username: "tonk",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "admin",
			FirstName: "Ton", LastName: "Kersten", DisplayName: "Ton Kersten",
			IsActive: true, EmailNotifications: true,
			HelpdeskEnabled: true,
		}
		must(db.Create(&tonk).Error)
		fmt.Println("   Created system admin: tonk (tonk@smartowl.nl)")
	} else {
		fmt.Println("   System admin 'tonk' already exists — skipping")
	}

	users := map[string]*models.User{
		"admin": {
			Email: "admin@demo.example", Username: "demo.admin",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "admin",
			FirstName: "Alex", LastName: "Admin", DisplayName: "Alex Admin",
			IsActive: true, EmailNotifications: true,
			AvatarURL: "https://api.dicebear.com/9.x/avataaars/svg?seed=admin",
			HelpdeskEnabled: true,
		},
		"sarah": {
			Email: "sarah@demo.example", Username: "demo.sarah",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Sarah", LastName: "Chen", DisplayName: "Sarah Chen",
			IsActive: true, EmailNotifications: true,
			AvatarURL: "https://api.dicebear.com/9.x/avataaars/svg?seed=sarah",
			HelpdeskEnabled: true,
		},
		"marc": {
			Email: "marc@demo.example", Username: "demo.marc",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Marc", LastName: "Dubois", DisplayName: "Marc Dubois",
			IsActive: true, EmailNotifications: true,
			AvatarURL: "https://api.dicebear.com/9.x/avataaars/svg?seed=marc",
			HelpdeskEnabled: true,
		},
		"lisa": {
			Email: "lisa@demo.example", Username: "demo.lisa",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Lisa", LastName: "Park", DisplayName: "Lisa Park",
			IsActive: true, EmailNotifications: true,
			AvatarURL: "https://api.dicebear.com/9.x/avataaars/svg?seed=lisa",
			HelpdeskEnabled: true,
		},
		"priya": {
			Email: "priya@demo.example", Username: "demo.priya",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Priya", LastName: "Nair", DisplayName: "Priya Nair",
			IsActive: true, EmailNotifications: true,
			AvatarURL: "https://api.dicebear.com/9.x/avataaars/svg?seed=priya",
			HelpdeskEnabled: true,
		},
		"james": {
			Email: "james@demo.example", Username: "demo.james",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "James", LastName: "O'Brien", DisplayName: "James O'Brien",
			IsActive: true, EmailNotifications: true,
			AvatarURL: "https://api.dicebear.com/9.x/avataaars/svg?seed=james",
			HelpdeskEnabled: true,
		},
		"elena": {
			Email: "elena@demo.example", Username: "demo.elena",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Elena", LastName: "Kovač", DisplayName: "Elena Kovač",
			IsActive: true, EmailNotifications: true,
			AvatarURL: "https://api.dicebear.com/9.x/avataaars/svg?seed=elena",
			HelpdeskEnabled: true,
		},
		"raj": {
			Email: "raj@demo.example", Username: "demo.raj",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Raj", LastName: "Sharma", DisplayName: "Raj Sharma",
			IsActive: true, EmailNotifications: false,
			AvatarURL: "https://api.dicebear.com/9.x/avataaars/svg?seed=raj",
			HelpdeskEnabled: true,
		},
		"viewer": {
			Email: "viewer@demo.example", Username: "demo.viewer",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "viewer",
			FirstName: "Victor", LastName: "Viewer", DisplayName: "Victor Viewer",
			IsActive: true, EmailNotifications: false,
			AvatarURL: "https://api.dicebear.com/9.x/avataaars/svg?seed=viewer",
			HelpdeskEnabled: true,
		},
	}

	for _, u := range users {
		must(db.Create(u).Error)
	}
	// Include the real system admin in the lookup map so seed code can reference 'tonk' in convs
	users["tonk"] = &tonk
	fmt.Printf("   Created %d demo users (password for all: demo1234)\n", len(users))

	// ── 2. Projects + columns + labels + members ──────────────────────────────
	fmt.Println("→ Creating projects…")

	type projectData struct {
		project *models.Project
		cols    map[string]*models.Column
		labels  map[string]*models.Label
	}
	projects := map[string]*projectData{}

	projectMembers := map[string][]struct {
		user string
		role string
	}{
		"website-redesign": {
			{"admin", "owner"}, {"sarah", "admin"}, {"marc", "member"},
			{"priya", "member"}, {"james", "member"}, {"viewer", "viewer"},
		},
		"mobile-app-v2": {
			{"sarah", "owner"}, {"marc", "admin"}, {"lisa", "member"},
			{"elena", "member"}, {"priya", "member"}, {"viewer", "viewer"},
		},
		"devops-infra": {
			{"marc", "owner"}, {"lisa", "admin"}, {"admin", "member"},
			{"james", "member"}, {"raj", "member"}, {"viewer", "viewer"},
		},
		"product-platform": {
			{"priya", "owner"}, {"sarah", "admin"}, {"james", "member"},
			{"raj", "member"}, {"elena", "member"}, {"viewer", "viewer"},
		},
		"api-platform": {
			{"elena", "owner"}, {"raj", "admin"}, {"james", "member"},
			{"priya", "member"}, {"sarah", "member"}, {"viewer", "viewer"},
		},
		"marketing": {
			{"lisa", "owner"}, {"admin", "admin"}, {"sarah", "member"},
			{"james", "member"}, {"viewer", "viewer"},
		},
	}

	for _, sp := range demoProjects {
		boardType := sp.boardType
		if boardType == "" {
			boardType = "kanban"
		}
		p := &models.Project{
			Name: sp.name, Slug: sp.slug, KeyPrefix: sp.prefix,
			Color: sp.color, Avatar: sp.avatar, Description: sp.desc, BoardType: boardType,
			CreatedByID: users["admin"].ID,
		}
		must(db.Create(p).Error)

		pd := &projectData{
			project: p,
			cols:    map[string]*models.Column{},
			labels:  map[string]*models.Label{},
		}

		// Columns
		for i, colName := range sp.columns {
			col := &models.Column{
				ProjectID: p.ID, Name: colName,
				Position: float64((i + 1) * 1000),
			}
			must(db.Create(col).Error)
			pd.cols[colName] = col
		}

		// Labels
		for _, l := range sp.labels {
			lbl := &models.Label{ProjectID: p.ID, Name: l.name, Color: l.color}
			must(db.Create(lbl).Error)
			pd.labels[l.name] = lbl
		}

		// Members
		for _, m := range projectMembers[sp.slug] {
			must(db.Create(&models.ProjectMember{
				ProjectID: p.ID, UserID: users[m.user].ID, Role: m.role,
			}).Error)
		}

		projects[sp.slug] = pd
	}
	fmt.Printf("   Created %d projects\n", len(projects))

	// ── 2b. Starred projects ──────────────────────────────────────────────────
	fmt.Println("→ Starring projects…")

	starredProjectRules := []struct {
		user    string
		project string
	}{
		// Admin has all projects in the sidebar
		{"admin", "website-redesign"},
		{"admin", "mobile-app-v2"},
		{"admin", "devops-infra"},
		{"admin", "product-platform"},
		{"admin", "marketing"},
		{"admin", "api-platform"},
		// Sarah: website owner, follows mobile and product-platform
		{"sarah", "website-redesign"},
		{"sarah", "mobile-app-v2"},
		{"sarah", "product-platform"},
		// Marc: mobile owner, follows devops-infra
		{"marc", "mobile-app-v2"},
		{"marc", "devops-infra"},
		// Lisa: devops owner, follows website-redesign and marketing
		{"lisa", "devops-infra"},
		{"lisa", "website-redesign"},
		{"lisa", "marketing"},
		// Elena: API Platform owner
		{"elena", "api-platform"},
		// Raj: API Platform admin
		{"raj", "api-platform"},
		// Priya: product-platform owner
		{"priya", "product-platform"},
	}

	for _, r := range starredProjectRules {
		pd := projects[r.project]
		must(db.Create(&models.StarredProject{
			UserID:    users[r.user].ID,
			ProjectID: pd.project.ID,
		}).Error)
	}
	// Tonk is a persistent user not in the users map — star all projects
	for slug := range projects {
		must(db.Create(&models.StarredProject{
			UserID:    tonk.ID,
			ProjectID: projects[slug].project.ID,
		}).Error)
	}
	fmt.Printf("   Starred %d project–user pairs\n", len(starredProjectRules)+len(projects))

	// ── 3. Cards ──────────────────────────────────────────────────────────────
	fmt.Println("→ Creating cards…")

	type cardSpec struct {
		title       string
		description string
		col         string
		priority    string
		labels      []string
		tags        []string
		assignee    string // user key or ""
		startInDays *int
		dueInDays   *int
		timeMin     int    // time_spent_minutes
		storyPoints int    // 0 = not set
		sprintName  string // sprint name to link this card to (scrum projects)
		checklist   []struct {
			body string
			done bool
		}
		comments []struct {
			author string
			body   string
		}
		subCards      []cardSpec
		closed        bool
		closedAtDays  *int // negative = days in the past; sets closed_at
		createdAtDays *int // negative = days in the past; sets created_at for CFD chart spread
	}

	webCards := []cardSpec{
		// Backlog
		{
			title: "Redesign homepage hero section", col: "Backlog",
			description: "The current hero section feels dated and no longer reflects our new brand positioning. We need a bold, conversion-focused hero with an updated headline, sub-headline, primary CTA button, and a supporting visual.\n\n**Acceptance criteria:**\n- Hero headline updated per brand copywriting guidelines\n- Responsive across mobile, tablet, and desktop\n- Primary CTA links to the sign-up flow\n- Visual asset (illustration or screenshot) provided by design team\n- Lighthouse Performance score unaffected",
			priority: "high", labels: []string{"Feature", "Design"},
			tags: []string{"homepage", "design"},
		},
		{
			title: "Add cookie consent banner", col: "Backlog",
			description: "GDPR compliance requires an explicit consent banner for analytics and marketing cookies. The banner should appear on first visit, remember the user's choice, and block tracking scripts until consent is given.\n\n**Acceptance criteria:**\n- Banner shown once on first visit; not shown on return visits\n- Three options: Accept All, Reject Non-Essential, Manage Preferences\n- Choice persisted in `localStorage` for 365 days\n- Analytics scripts blocked until consent is granted\n- Fully keyboard-navigable with correct ARIA roles",
			priority: "medium", labels: []string{"Feature"},
			assignee: "marc", startInDays: ptr(7), dueInDays: ptr(14),
		},
		{
			title: "Write new About page copy", col: "Backlog",
			description: "Our About page hasn't been updated since the company's founding and no longer reflects the current team, mission, or culture. Fresh copy is needed before the redesign launches.\n\n**Deliverables:**\n- Mission and values section (150–200 words)\n- Team bios for all 8 core team members (photo + 2–3 sentences each)\n- Company timeline / milestone section\n- Culture section highlighting our remote-first approach\n\nSource material and brand voice guidelines are in the shared Google Doc (link in comments).",
			priority: "none", labels: []string{"Content"},
			assignee: "sarah",
		},
		// In Progress
		{
			title: "Implement dark mode toggle", col: "In Progress",
			description: "Add a user-accessible toggle to switch between light and dark themes. The preference should be persisted so it survives page reloads, and the system preference (`prefers-color-scheme`) should be respected on first visit.\n\n**Acceptance criteria:**\n- Toggle visible in the site header on all pages\n- Respects `prefers-color-scheme` as the initial default\n- User's choice stored in `localStorage`\n- Smooth CSS `transition` on colour changes (no hard flash)\n- All pages, components, and third-party embeds covered\n- No flash of wrong theme on initial page load (SSR-safe)",
			priority: "medium", labels: []string{"Feature", "Design"},
			assignee: "sarah", timeMin: 180, startInDays: ptr(-3), dueInDays: ptr(5),
			checklist: []struct {
				body string
				done bool
			}{
				{"Research CSS variables approach", true},
				{"Implement toggle button in header", true},
				{"Persist preference to localStorage", false},
				{"Test on Safari and Firefox", false},
			},
			comments: []struct {
				author string
				body   string
			}{
				{"marc", "I suggest using `prefers-color-scheme` as the initial default — saves a flash of wrong theme on first load."},
				{"sarah", "Good call! Adding that to the checklist."},
			},
			subCards: []cardSpec{
				{title: "Define CSS custom properties for dark palette", assignee: "sarah", priority: "medium"},
				{title: "Add toggle button to site header", assignee: "sarah", priority: "medium"},
				{title: "Persist theme choice in localStorage", assignee: "marc", priority: "low"},
				{title: "QA dark mode on Safari and Firefox", assignee: "priya", priority: "low"},
			},
		},
		{
			title: "Fix mobile navigation overflow on small screens", col: "In Progress",
			description: "On viewports narrower than ~390 px the hamburger menu button clips behind the logo, making it unreachable on iPhone SE and older Android devices.\n\n**Steps to reproduce:**\n1. Open the site in Chrome DevTools at a 375 px viewport width\n2. Observe the nav — the ☰ icon overlaps the logo\n\n**Root cause:** `overflow: hidden` is missing on the `.nav-container` element.\n\n**Fix:** Add `overflow: hidden` to `.nav-container` and verify on iPhone SE, iPhone 12 mini, and Pixel 4a viewports before closing.",
			priority: "high", labels: []string{"Bug"},
			assignee: "marc", timeMin: 90, startInDays: ptr(0), dueInDays: ptr(2),
			comments: []struct {
				author string
				body   string
			}{
				{"sarah", "Reproduced on iPhone SE (375 px). The hamburger menu clips behind the logo."},
				{"marc", "On it — looks like `overflow: hidden` is missing on the nav container."},
			},
		},
		{
			title: "Update brand colour palette across all components", col: "In Progress",
			description: "Marketing has finalised the new brand palette (indigo primary, teal accent, updated neutrals). All components must use the new CSS custom properties before the redesign launches.\n\n**Scope:**\n- Update `--color-primary`, `--color-accent`, and neutral tokens in `:root`\n- Audit every component for hard-coded colour values and replace them\n- Update button, badge, link, alert, and form element styles\n- Verify all colour combinations meet WCAG AA contrast ratios in both light and dark modes",
			priority: "low", labels: []string{"Design"},
			assignee: "sarah", timeMin: 120,
		},
		// Review
		{
			title: "Optimise image loading with lazy + WebP", col: "Review",
			description: "The homepage serves 14 images as PNG/JPEG. Converting to WebP with lazy loading reduced LCP from 3.8 s to 1.2 s in local testing.\n\n**Changes made:**\n- All content images converted to WebP with JPEG/PNG `<picture>` fallback\n- `loading=\"lazy\"` added to all below-the-fold images\n- Explicit `width` and `height` attributes set on all `<img>` elements to prevent layout shift (CLS)\n- Hero image compressed to ≤ 150 KB at 2× resolution\n\n**Needs sign-off on:** WebP fallback strategy for browsers without WebP support (< 3% of our traffic).",
			priority: "medium", labels: []string{"Feature"},
			assignee: "sarah", timeMin: 240, startInDays: ptr(-5), dueInDays: ptr(1),
			comments: []struct {
				author string
				body   string
			}{
				{"marc", "LCP went from 3.8 s to 1.2 s in local tests — great improvement!"},
				{"sarah", "Waiting for sign-off on the WebP fallback strategy for older browsers."},
			},
		},
		{
			title: "Accessibility audit and ARIA fixes", col: "Review",
			description: "A third-party accessibility audit identified 23 issues. All critical and high-severity findings must be resolved before launch.\n\n**Critical (must fix):**\n- Missing `alt` text on 11 images\n- Form inputs not associated with `<label>` elements\n- Focus indicator not visible on custom button components\n\n**High severity:**\n- Colour contrast below 4.5:1 on muted body text\n- Navigation landmark roles missing (`<nav>`, `<main>`, `<header>`)\n- Modal dialogs not trapping keyboard focus\n\nFull audit report is in `/docs/a11y-audit-2024.pdf`.",
			priority: "high", labels: []string{"Feature"},
			assignee: "marc", timeMin: 300, startInDays: ptr(-7), dueInDays: ptr(3),
		},
		// Done
		{
			title: "Set up GitHub Actions CI/CD pipeline", col: "Done",
			description: "Automated the full CI/CD pipeline with GitHub Actions. Every PR gets lint, build, and a Vercel preview deploy; merges to `main` go straight to production.\n\n**Pipeline stages:**\n1. Lint — ESLint + Prettier check\n2. Build — Vite production build\n3. Preview deploy to Vercel (PRs only)\n4. Lighthouse CI score gate (Performance ≥ 85)\n5. Production deploy on merge to `main`\n\nSecrets are stored in repository Settings → Secrets. See `.github/workflows/ci.yml` for the full pipeline definition.",
			priority: "high", labels: []string{"Feature"},
			assignee: "admin", timeMin: 360, startInDays: ptr(-16), dueInDays: ptr(-10),
			closed: true, closedAtDays: ptr(-10), createdAtDays: ptr(-16),
		},
		{
			title: "Migrate DNS and SSL to new hosting provider", col: "Done",
			description: "Migrated all DNS records from GoDaddy to Cloudflare and provisioned Let's Encrypt SSL certificates via the new hosting provider. Achieved zero downtime with a staged TTL reduction.\n\n**Steps taken:**\n- Reduced DNS TTL to 60 s a full 48 hours before cutover\n- Verified all A, CNAME, MX, and TXT records in Cloudflare before switching nameservers\n- SSL auto-renewal configured (90-day certs, renew at 30 days)\n- Old hosting account cancelled — monthly saving: €120",
			priority: "critical", labels: []string{"Feature"},
			assignee: "admin", timeMin: 480, startInDays: ptr(-9), dueInDays: ptr(-5),
			closed: true, closedAtDays: ptr(-5), createdAtDays: ptr(-9),
		},
		{
			title: "Create component library documentation", col: "Done",
			description: "Documented all 34 shared UI components in Storybook with props, usage examples, and visual variants. The docs are published at `design.example.com` and linked from the project README.\n\n**Coverage:**\n- All button variants (primary, secondary, ghost, danger; sm / md / lg)\n- Form inputs: text, select, textarea, checkbox, radio, toggle\n- Overlay components: Modal, Toast, Tooltip, Dropdown\n- Data display: Badge, Avatar, Spinner, ProgressBar\n- Layout: Sidebar, Toolbar, PageHeader, Divider",
			priority: "low", labels: []string{"Design", "Content"},
			assignee: "sarah", timeMin: 210, startInDays: ptr(-10), dueInDays: ptr(-3),
			closed: true, closedAtDays: ptr(-3), createdAtDays: ptr(-10),
		},
		{
			title: "Audit and fix all broken links", col: "Done",
			description: "A full site crawl with Screaming Frog found 17 broken internal links and 4 broken external links. All have been corrected.\n\n**Fixed:**\n- 8 blog post internal links pointing to renamed URL slugs\n- 4 footer links pointing to deprecated product pages\n- 3 resource download links (files moved to new CDN path)\n- 2 external links updated to current URLs\n- 4 external links removed (target pages no longer exist)\n\nCrawl report archived in `/docs/link-audit-2024.csv`.",
			priority: "medium", labels: []string{"Bug"},
			assignee: "marc", timeMin: 60, startInDays: ptr(-9), dueInDays: ptr(-7),
			closed: true, closedAtDays: ptr(-7), createdAtDays: ptr(-9),
		},
	}

	mobCards := []cardSpec{
		// Ideas
		{
			title: "Offline mode with sync queue", col: "Ideas",
			description: "Users in low-connectivity areas (trains, rural locations) report data loss when the app loses connection mid-action. Implement a local sync queue that persists operations and replays them automatically when connectivity is restored.\n\n**Proposed scope:**\n- Queue: create, update, and delete operations on tasks and notes\n- Persistence: SQLite via React Native async storage\n- Conflict resolution: last-write-wins with a merge prompt for conflicting edits\n- UI: subtle offline banner; sync progress indicator on reconnect\n\n**Out of scope for v1:** Real-time collaboration conflict resolution.",
			priority: "high", labels: []string{"Enhancement"},
			tags: []string{"offline", "ux"},
		},
		{
			title: "Push notification preferences screen", col: "Ideas",
			description: "Users have no granular control over which push notifications they receive, leading to high opt-out rates. We need a preferences screen where they can toggle categories individually.\n\n**Proposed categories:**\n- Task due-date reminders\n- Card assignment notifications\n- Comment `@mention` alerts\n- Team activity digest\n- Product updates / marketing (off by default)\n\n**Tech note:** Preferences must be stored server-side so they sync across a user's devices.",
			priority: "medium", labels: []string{"Enhancement"},
		},
		{
			title: "Biometric login (Face ID / fingerprint)", col: "Ideas",
			description: "After the initial password login, allow users to authenticate with Face ID (iOS) or fingerprint (Android) for all subsequent sessions. The biometric check wraps the existing JWT refresh flow.\n\n**Proposed flow:**\n1. User logs in with email/password → JWT stored in the secure device keychain\n2. On next app launch, biometric prompt appears\n3. Success → silently fetch a new access token using the stored refresh token\n4. Failure (3 attempts) → fall back to password login\n\n**Library:** `expo-local-authentication` (already in our Expo managed workflow).",
			priority: "medium", labels: []string{"Enhancement"},
			tags: []string{"security", "auth"},
		},
		// Development
		{
			title: "User profile screen", col: "Development",
			description: "Build a screen where users can view and edit their display name, bio, and profile photo. Changes persist to the backend via `PATCH /api/v1/profile`.\n\n**Fields:**\n- Display name (required, max 60 chars)\n- Bio (optional, max 200 chars)\n- Avatar: pick from camera roll or take a new photo, cropped to 1:1 square\n- Email address (read-only, with a link to the change-email flow)\n\n**Validation rules:**\n- Warn user about unsaved changes on back navigation\n- Show live character count for the bio field\n- Avatar upload progress shown inline",
			priority: "high", labels: []string{"Enhancement", "iOS", "Android"},
			assignee: "sarah", timeMin: 300, startInDays: ptr(-7), dueInDays: ptr(7),
			checklist: []struct {
				body string
				done bool
			}{
				{"Approve design mockup", true},
				{"Build form components", true},
				{"Connect to profile API", true},
				{"Add avatar upload", false},
				{"Write E2E tests", false},
			},
			comments: []struct {
				author string
				body   string
			}{
				{"marc", "The avatar cropper might need a third-party library — I found `react-easy-crop`, looks solid."},
				{"sarah", "Added to the checklist. Let's decide before Thursday."},
				{"lisa", "I can test on Android once the form components are done."},
			},
			subCards: []cardSpec{
				{title: "Design mockup sign-off", assignee: "sarah", priority: "high"},
				{title: "Build display name and bio form fields", assignee: "sarah", priority: "high"},
				{title: "Implement avatar upload with crop", assignee: "marc", priority: "medium"},
				{title: "Connect form to PATCH /profile API", assignee: "sarah", priority: "high"},
				{title: "E2E tests for profile save flow", assignee: "lisa", priority: "medium"},
			},
		},
		{
			title: "Settings / preferences screen", col: "Development",
			description: "A consolidated settings screen covering account, notifications, appearance, and privacy.\n\n**Sections:**\n- **Account:** change password, linked social accounts, delete account\n- **Notifications:** granular push-notification toggles (see separate card)\n- **Appearance:** theme (System / Light / Dark), text size\n- **Privacy:** analytics opt-out, data export request\n\nDesign spec linked in Figma (see comments). Settings should be persisted both locally and synced to the user profile on the server so they apply on all devices.",
			priority: "medium", labels: []string{"Enhancement", "iOS", "Android"},
			assignee: "marc", timeMin: 180, startInDays: ptr(0), dueInDays: ptr(10),
		},
		{
			title: "Dark mode support", col: "Development",
			description: "Follow the device's system Appearance setting by default, with an in-app override available in Settings → Appearance.\n\n**Implementation:**\n- Use the React Native `useColorScheme` hook to read the system preference\n- Maintain a `ThemeContext` that all components consume\n- All colours defined as theme tokens — no hard-coded hex values anywhere\n- Test on iOS Dark Mode and Android Dark Theme in both states\n\n**Edge cases to handle:**\n- `KeyboardAvoidingView` (background must match theme)\n- Native share sheet (system-controlled — nothing to do)\n- Push notification previews (plain text only)",
			priority: "low", labels: []string{"Enhancement"},
			assignee: "lisa", timeMin: 240, startInDays: ptr(-5), dueInDays: ptr(14),
		},
		{
			title: "App crashes on empty conversation list", col: "Development",
			description: "**Severity: Critical — reproducible in production.**\n\nWhen a brand-new user opens the Conversations tab before sending or receiving any messages, the app crashes immediately with a null pointer exception.\n\n**Stack trace:**\n```\nTypeError: Cannot read property 'length' of null\n  at ConversationListScreen.tsx:42\n```\n\n**Fix:** Add a null guard before accessing the conversation array:\n```ts\nconst count = (conversations ?? []).length\n```\n\n**Regression test:** Add a Detox test that renders `ConversationListScreen` with a null and an empty conversation list — both must render the empty state without crashing.",
			priority: "critical", labels: []string{"Bug"},
			assignee: "marc", timeMin: 45, startInDays: ptr(0), dueInDays: ptr(1),
			comments: []struct {
				author string
				body   string
			}{
				{"lisa", "Stack trace attached. Null check missing in `ConversationListScreen.tsx` line 42."},
				{"marc", "Fix is one line — pushing a hotfix build now."},
			},
		},
		// Testing
		{
			title: "Integration tests for authentication flow", col: "Testing",
			description: "The authentication flow — register, verify email, login, token refresh, logout — has no automated test coverage. Any regression here would block all users.\n\n**Test cases:**\n- Register with valid email/password → verify email → login succeeds\n- Register with duplicate email → correct error shown\n- Login with wrong password → lockout after 5 failed attempts\n- Token refresh succeeds and the new access token is valid\n- Logout clears all tokens and redirects to the login screen\n\n**Framework:** Detox (already configured in the `e2e/` directory).",
			priority: "high", labels: []string{"Enhancement", "iOS", "Android"},
			assignee: "lisa", timeMin: 360, startInDays: ptr(-4), dueInDays: ptr(3),
		},
		{
			title: "Performance profiling on low-end Android devices", col: "Testing",
			description: "Users on budget Android devices (Moto G Play, Galaxy A03) report janky scrolling in the main feed. Profile the app on real hardware and address the top three findings.\n\n**Target:** Consistent 60 fps scrolling on a 2021-era mid-range Android device.\n\n**Known suspects:**\n- Feed renders ~200 items without `FlatList` virtualisation\n- Each list item makes a separate avatar image request (no batching)\n- Expensive `elevation` shadow styles on older Android GPU\n\n**Tooling:** Android Studio CPU Profiler + React Native Perf Monitor running simultaneously on a physical Moto G device.",
			priority: "medium", labels: []string{"Enhancement", "Android"},
			assignee: "sarah", timeMin: 150, startInDays: ptr(-3), dueInDays: ptr(7),
			comments: []struct {
				author string
				body   string
			}{
				{"marc", "Focus on the feed screen — renders ~200 items without virtualisation right now."},
			},
		},
		// Released
		{
			title: "Initial 1.0 app release", col: "Released",
			description: "Successfully shipped the v1.0 app to both the Apple App Store and Google Play Store. All items on the pre-launch checklist completed.\n\n**Launch results:**\n- 4.6 ★ average rating in the first week (iOS: 4.8, Android: 4.4)\n- 1,200 installs in the first 48 hours\n- Zero critical crashes in the first 72 hours\n\n**Post-launch actions taken:**\n- App Store Optimisation (ASO) pass scheduled for week 3\n- Review response workflow set up and documented\n- Crash monitoring configured in Sentry",
			priority: "critical", labels: []string{"Enhancement", "iOS", "Android"},
			assignee: "sarah", timeMin: 1200, startInDays: ptr(-60), dueInDays: ptr(-30),
			closed: true, closedAtDays: ptr(-30), createdAtDays: ptr(-60),
		},
		{
			title: "Bug fix release 1.0.1", col: "Released",
			description: "Hotfix release addressing two production issues found after the 1.0 launch.\n\n**Fixes included:**\n- Crash on empty conversation list (MOB-7) — null guard added to `ConversationListScreen.tsx`\n- Profile photo not updating in the UI after a successful upload — cache-busting timestamp now appended to the avatar URL\n\nRelease notes published to App Store and Google Play. The JS bundle fix was pushed OTA via Expo EAS Update for immediate delivery without a full store review.",
			priority: "high", labels: []string{"Bug"},
			assignee: "marc", timeMin: 120, startInDays: ptr(-18), dueInDays: ptr(-14),
			closed: true, closedAtDays: ptr(-14), createdAtDays: ptr(-18),
		},
	}

	infCards := []cardSpec{
		// Todo
		{
			title: "Set up Kubernetes cluster on cloud provider", col: "Todo",
			description: "Migrate our VM-based deployment to a managed Kubernetes cluster to improve scalability, resilience, and deployment velocity.\n\n**Requirements:**\n- Managed control plane (EKS / GKE / AKS — to be decided, see sub-tasks)\n- Auto-scaling for the API and worker node pools\n- Namespace isolation per environment: `dev`, `staging`, `prod`\n- RBAC: dev team gets read access; ops team gets write\n- Ingress controller with TLS termination (cert-manager + Let's Encrypt)\n\n**Success criteria:** Zero-downtime rolling deploys for all services within 30 s of a push to `main`.",
			priority: "high", labels: []string{"Enhancement"},
			startInDays: ptr(3), dueInDays: ptr(21),
			checklist: []struct {
				body string
				done bool
			}{
				{"Choose cloud provider (GKE / EKS / AKS)", false},
				{"Design namespace and RBAC structure", false},
				{"Configure networking and ingress", false},
				{"Set up auto-scaling policies", false},
				{"Disaster recovery runbook", false},
			},
			subCards: []cardSpec{
				{title: "Evaluate GKE vs EKS pricing and SLAs", assignee: "marc", priority: "high"},
				{title: "Design namespace and RBAC structure", assignee: "lisa", priority: "high"},
				{title: "Configure ingress controller and TLS termination", assignee: "marc", priority: "medium"},
				{title: "Set up cluster auto-scaler", assignee: "lisa", priority: "medium"},
				{title: "Write disaster recovery runbook", assignee: "raj", priority: "low"},
			},
		},
		{
			title: "Add Prometheus + Grafana monitoring stack", col: "Todo",
			description: "We have no metrics observability. Deploy Prometheus for collection and Grafana for dashboards and alerting.\n\n**Dashboards required (MVP):**\n- HTTP request rate, error rate, and latency (RED metrics)\n- Database connection pool utilisation\n- Pod CPU and memory utilisation\n- Disk I/O and free space\n\n**Alerting:** PagerDuty integration for P1/P2 alerts:\n- Error rate > 5% over 5 minutes\n- Pod restart count > 3 in 5 minutes\n- Disk usage > 85%",
			priority: "medium", labels: []string{"Monitoring"},
			assignee: "lisa", startInDays: ptr(7), dueInDays: ptr(28),
		},
		{
			title: "Quarterly security audit", col: "Todo",
			description: "**Due this week.** Perform the Q2 security audit required by our SOC 2 Type II commitments.\n\n**Checklist:**\n- [ ] Review IAM user list — remove all inactive accounts\n- [ ] Rotate service account credentials older than 90 days\n- [ ] Review security group and firewall rules for over-permissive entries\n- [ ] Verify all S3 / GCS buckets are private (no public ACLs)\n- [ ] Review Snyk dependency vulnerability report — address all critical findings\n- [ ] Audit access logs for anomalies over the past 30 days\n\nFindings must be documented in the security register. Any P0/P1 issues are escalated immediately.",
			priority: "critical", labels: []string{"Security"},
			startInDays: ptr(1), dueInDays: ptr(7),
		},
		// In Progress
		{
			title: "Migrate primary database to PostgreSQL", col: "In Progress",
			description: "Our production database is SQLite, causing write-contention issues under load. This card tracks the migration to PostgreSQL 16 on RDS.\n\n**Migration plan:**\n1. Provision RDS PostgreSQL 16 instance (Multi-AZ for prod)\n2. Run `pgloader` to migrate existing data\n3. Shadow-write period: 1 week writing to both, reading from Postgres\n4. Verify row counts and checksums across both databases\n5. Promote Postgres as primary; demote SQLite to read-only standby\n6. Remove SQLite dependency after 2 weeks of stable operation\n\n**Rollback:** SQLite remains available as a read-only fallback throughout the cutover window.",
			priority: "high", labels: []string{"Enhancement"},
			assignee: "marc", timeMin: 480, startInDays: ptr(-10), dueInDays: ptr(4),
			comments: []struct {
				author string
				body   string
			}{
				{"admin", "Remember to keep the SQLite DB as a read-only fallback for at least two weeks after cutover."},
				{"marc", "Agreed. I've scripted the data export — schema diff is smaller than expected."},
				{"lisa", "I'll keep an eye on slow-query logs during the transition."},
			},
		},
		{
			title: "Automate database backups with off-site retention", col: "In Progress",
			description: "We currently have no automated backup strategy — a critical gap for our RTO/RPO commitments.\n\n**Requirements:**\n- Daily automated backups, retained for 30 days\n- Weekly backups retained for 1 year\n- All backups encrypted at rest (AES-256) and in transit (TLS)\n- Off-site storage in a different cloud region or provider\n- Monthly restore drills to verify backup integrity\n- Alert via PagerDuty if the backup job fails or exceeds 2× its expected duration",
			priority: "high", labels: []string{"Monitoring"},
			assignee: "lisa", timeMin: 180, startInDays: ptr(-2), dueInDays: ptr(6),
		},
		{
			title: "Renew and automate SSL certificate rotation", col: "In Progress",
			description: "Our wildcard certificate expires in 8 days. Automate renewal via cert-manager + Let's Encrypt so this never requires manual intervention again.\n\n**Actions:**\n1. Install cert-manager in the cluster\n2. Create a `ClusterIssuer` for the Let's Encrypt production environment\n3. Annotate all ingress resources with `cert-manager.io/cluster-issuer`\n4. Verify auto-renewal fires at 30 days before expiry\n5. Add a Grafana alert at 14 days remaining as a safety net\n\n**Immediate stopgap:** Manually renew the current cert today while automation is being wired up.",
			priority: "critical", labels: []string{"Security"},
			assignee: "marc", timeMin: 60, startInDays: ptr(-1), dueInDays: ptr(2),
		},
		// Done
		{
			title: "Set up private Docker registry", col: "Done",
			description: "Moved all container images from Docker Hub (public) to a private AWS ECR registry to eliminate rate limits, reduce pull latency, and prevent accidental public image exposure.\n\n**Configuration:**\n- Registry: AWS ECR `warmdesk` repository with image scanning enabled\n- Lifecycle policy: retain last 10 tagged images; delete untagged images after 1 day\n- CI/CD pipelines updated to push to and pull from ECR\n- Kubernetes `imagePullSecrets` configured per namespace\n- All team members granted ECR read access via IAM role",
			priority: "medium", labels: []string{"Enhancement"},
			assignee: "marc", timeMin: 240, startInDays: ptr(-14), dueInDays: ptr(-10),
			closed: true, closedAtDays: ptr(-10), createdAtDays: ptr(-14),
		},
		{
			title: "Configure Nginx load balancer with health checks", col: "Done",
			description: "Configured Nginx as a layer-7 load balancer with active health checks. Failed API server instances are automatically removed from rotation within 10 seconds.\n\n**Configuration highlights:**\n- Upstream block with 3 API server addresses\n- Active health check: `GET /api/v1/health` every 5 s, 2 s timeout\n- Passive health check: 3 consecutive failures within 30 s marks the upstream down\n- Sticky sessions disabled — all API instances are stateless\n- Connection timeout: 30 s; read timeout: 60 s\n- Gzip compression enabled for JSON and static asset responses",
			priority: "high", labels: []string{"Enhancement"},
			assignee: "lisa", timeMin: 300, startInDays: ptr(-13), dueInDays: ptr(-8),
			closed: true, closedAtDays: ptr(-8), createdAtDays: ptr(-13),
		},
		{
			title: "Deploy staging environment", col: "Done",
			description: "Provisioned a full-replica staging environment mirroring production: same region, same instance types, same configuration — with a separate, isolated database.\n\n**Environment overview:**\n- API: 2× t3.medium (auto-scaled, matching prod)\n- DB: RDS t3.small PostgreSQL 16 (single-AZ, cost-optimised)\n- Redis: ElastiCache t3.micro\n- DNS: `staging.example.com` (internal, VPN access only)\n\n**Deployment:** Staging is updated automatically on every merge to the `develop` branch via GitHub Actions.\n\n**Access:** Requires VPN. Credentials are in the 1Password vault under \"Staging Access\".",
			priority: "high", labels: []string{"Enhancement"},
			assignee: "marc", timeMin: 360, startInDays: ptr(-21), dueInDays: ptr(-15),
			closed: true, closedAtDays: ptr(-15), createdAtDays: ptr(-21),
		},
	}

	pltCards := []cardSpec{
		// Sprint 1 — Discovery (completed, -70 to -56): all in Done
		{title: "Product market research and user interviews", col: "Done", priority: "high",
			description: "Conduct structured user interviews and market research to validate the product concept and inform the roadmap before development begins.\n\n**Deliverables:**\n- 12 user interviews across three personas: solo founder, team lead, enterprise PM\n- Competitor matrix covering Jira, Linear, Notion, and Asana\n- Problem statement and opportunity sizing document\n- Priority feature list ranked by user pain severity\n\nInterview recordings and consent forms are in `/docs/research/`.",
			labels: []string{"Feature"}, assignee: "priya", timeMin: 300,
			storyPoints: 5, sprintName: "Sprint 1 — Discovery",
			startInDays: ptr(-70), dueInDays: ptr(-63),
			closed: true, closedAtDays: ptr(-59), createdAtDays: ptr(-70)},
		{title: "Competitor analysis and positioning", col: "Done", priority: "medium",
			description: "Analyse the competitive landscape to identify differentiation opportunities and define our go-to-market positioning.\n\n**Competitors analysed:**\n- Jira — enterprise-grade, complex, expensive\n- Linear — developer-focused, fast, opinionated\n- Asana — team-friendly, broad feature set\n- Notion — flexible but not purpose-built for project management\n\n**Output:** A positioning document with our chosen niche (\"fast, beautiful PM tool for growing dev teams\") and three key differentiators. Approved by the founders.",
			labels: []string{"Feature"}, assignee: "james", timeMin: 180,
			storyPoints: 3, sprintName: "Sprint 1 — Discovery",
			startInDays: ptr(-69), dueInDays: ptr(-62),
			closed: true, closedAtDays: ptr(-58), createdAtDays: ptr(-69)},
		{title: "Technical feasibility study", col: "Done", priority: "high",
			description: "Assess technical risk areas before committing to the architecture. Focus on real-time collaboration, data volume at scale, and cross-platform parity.\n\n**Key questions answered:**\n- Can WebSockets handle 500 concurrent connections on a single Go server? ✓ Yes (tested with k6)\n- What is the read/write ratio at 10 k cards per project? ✓ ~20:1 — reads dominate\n- Is Go + Vue + SQLite/Postgres maintainable by a team of five? ✓ Yes\n\n**Output:** Architecture Decision Records (ADRs) for the top three technical risks. Stored in `/docs/architecture/`.",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 240,
			storyPoints: 3, sprintName: "Sprint 1 — Discovery",
			startInDays: ptr(-68), dueInDays: ptr(-61),
			closed: true, closedAtDays: ptr(-58), createdAtDays: ptr(-68)},
		{title: "Initial wireframes and UX sketches", col: "Done", priority: "medium",
			description: "Create low-fidelity wireframes for the core product flows to gather feedback before investing in high-fidelity design work.\n\n**Screens required:**\n- Kanban board — column and card layout\n- Card detail drawer / modal\n- Sprint backlog with drag-to-assign\n- Reports: velocity and burndown charts\n- User settings page\n\nWireframes will be presented at the Monday design review. Feedback incorporated in the Figma file before sprint 2 begins.",
			labels: []string{"Feature"}, assignee: "sarah", timeMin: 360,
			storyPoints: 5, sprintName: "Sprint 1 — Discovery",
			startInDays: ptr(-67), dueInDays: ptr(-60),
			closed: true, closedAtDays: ptr(-57), createdAtDays: ptr(-67)},
		{title: "Project charter and team kick-off", col: "Done", priority: "high",
			description: "Formally kick off the product-platform project with a written charter agreed by all stakeholders and a team alignment session.\n\n**Charter covers:**\n- Project goals, success metrics, and explicit out-of-scope items\n- Team roles and decision-making process (RACI matrix)\n- Development cadence: 2-week sprints, Wednesday retrospectives\n- Communication norms: async-first, daily stand-up in Slack at 09:30\n- Initial risk register (5 risks identified and owned)\n\nKick-off deck and charter PDF shared in `/docs/project-charter-v1.pdf`.",
			labels: []string{"Feature"}, assignee: "priya", timeMin: 60,
			storyPoints: 2, sprintName: "Sprint 1 — Discovery",
			startInDays: ptr(-66), dueInDays: ptr(-59),
			closed: true, closedAtDays: ptr(-57), createdAtDays: ptr(-66)},

		// Sprint 2 — Auth & Security (completed, -56 to -42): all in Done
		{title: "OAuth 2.0 provider integration", col: "Done", priority: "critical",
			description: "Implement sign-in with Google and GitHub using the OAuth 2.0 Authorization Code flow. Users can link social accounts to an existing email/password account.\n\n**Providers shipped:**\n- Google (covers ~70 % of target users)\n- GitHub (priority for developer users)\n\n**Flows implemented:**\n- New user: OAuth → account creation → onboarding wizard\n- Existing user: OAuth → linked to existing account (matched by email)\n- Disconnect: UI to unlink a provider (requires a password to be set first)\n\nMicrosoft/Azure AD is deferred to the enterprise tier.",
			labels: []string{"Feature"}, assignee: "priya", timeMin: 480,
			storyPoints: 5, sprintName: "Sprint 2 — Auth & Security",
			startInDays: ptr(-56), dueInDays: ptr(-49),
			closed: true, closedAtDays: ptr(-45), createdAtDays: ptr(-56)},
		{title: "JWT token management and refresh", col: "Done", priority: "high",
			description: "Implement the full JWT token lifecycle: short-lived access tokens, longer-lived refresh tokens, silent refresh on 401, and secure storage.\n\n**Access token:** HS256, 15-minute TTL. Claims: `user_id`, `username`, `role`.\n\n**Refresh token:** Opaque UUID stored in the DB, 7-day TTL, single-use with rotation on each refresh.\n\n**Frontend behaviour:** Axios interceptor catches 401 → calls `/auth/refresh` → retries the original request transparently. The user is only logged out if the refresh token has also expired or been revoked.\n\n**Security:** Refresh tokens are invalidated on logout and on every password change.",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 300,
			storyPoints: 5, sprintName: "Sprint 2 — Auth & Security",
			startInDays: ptr(-55), dueInDays: ptr(-48),
			closed: true, closedAtDays: ptr(-44), createdAtDays: ptr(-55)},
		{title: "Bcrypt password hashing", col: "Done", priority: "high",
			description: "Ensure all passwords are stored as bcrypt hashes with a cost factor tuned to our server hardware.\n\n**Implementation:**\n- Cost factor: 12 (≈ 250 ms per hash on the CI server — acceptable for login latency)\n- `bcrypt.GenerateFromPassword` called at registration and on every password change\n- `bcrypt.CompareHashAndPassword` used at login (constant-time comparison built in)\n\n**Migration:** Any plain-text passwords in the dev database are transparently upgraded on the user's next login.\n\n**Tests:** Verify a cost-12 hash is accepted; verify that hashes from other cost factors are also accepted (bcrypt embeds the cost in the hash string).",
			labels: []string{"Feature"}, assignee: "james", timeMin: 120,
			storyPoints: 3, sprintName: "Sprint 2 — Auth & Security",
			startInDays: ptr(-54), dueInDays: ptr(-48),
			closed: true, closedAtDays: ptr(-44), createdAtDays: ptr(-54)},
		{title: "Admin dashboard scaffolding", col: "Done", priority: "medium",
			description: "Build the structural shell of the admin dashboard: navigation, layout, and the five core admin sections. Individual sections will be detailed in subsequent cards.\n\n**Sections (MVP):**\n- User management — list, invite, deactivate users\n- Project overview — all projects with member counts and activity\n- System settings — SMTP, storage, branding, locale\n- Audit log viewer — searchable event feed\n- Health summary — DB status, queue depth, server uptime\n\n**Access control:** `GlobalRole = \"admin\"` only. Middleware guard on all `/api/v1/admin/*` routes.",
			labels: []string{"Feature"}, assignee: "sarah", timeMin: 360,
			storyPoints: 5, sprintName: "Sprint 2 — Auth & Security",
			startInDays: ptr(-53), dueInDays: ptr(-46),
			closed: true, closedAtDays: ptr(-43), createdAtDays: ptr(-53)},
		{title: "User invitation and onboarding flow", col: "Done", priority: "high",
			description: "Allow existing users to invite colleagues by email. Invited users receive a tokenised signup link that pre-fills their email and skips the separate email verification step.\n\n**Flow:**\n1. Admin / project owner enters one or more email addresses → invitation records created\n2. Email sent with a signed, 72-hour invite token\n3. Recipient clicks the link → registration form with email pre-filled\n4. On submit → account created, token invalidated, user added to the project\n5. Expired token → friendly error with a \"request a new invite\" option\n\n**Edge case:** If the invited user already has an account, the link logs them in and adds them to the project directly without creating a duplicate.",
			labels: []string{"Feature"}, assignee: "priya", timeMin: 240,
			storyPoints: 3, sprintName: "Sprint 2 — Auth & Security",
			startInDays: ptr(-52), dueInDays: ptr(-45),
			closed: true, closedAtDays: ptr(-43), createdAtDays: ptr(-52)},

		// Sprint 3 — Foundation (completed, -28 to -14): all in Done
		{title: "Set up monorepo and CI/CD pipeline", col: "Done", priority: "high",
			description: "Configure the repository as a Go + Vue monorepo with a unified CI/CD pipeline that builds and deploys both services together.\n\n**Repository structure:**\n```\n/backend   → Go API service\n/frontend  → Vue 3 + Vite SPA\n/deploy    → Docker, nginx, systemd templates\nMakefile   → build, test, lint targets\n```\n\n**CI (GitHub Actions):**\n- Go: `go vet`, `staticcheck`, build verification\n- Vue: ESLint, Vite production build\n- Docker: build + push to ECR on every green merge to `main`\n\n**CD:** Auto-deploy to staging on every green build on `main`.",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 240,
			storyPoints: 3, sprintName: "Sprint 3 — Foundation",
			startInDays: ptr(-28), dueInDays: ptr(-21),
			closed: true, closedAtDays: ptr(-17), createdAtDays: ptr(-28)},
		{title: "Design system and component library", col: "Done", priority: "high",
			description: "Establish a reusable component library as the foundation for all UI work. Every component must support both light and dark themes through CSS custom properties.\n\n**Core components shipped:**\n- Buttons: primary, secondary, ghost, danger — sm / md / lg sizes\n- Form inputs: text, select, textarea, checkbox, radio, toggle\n- Overlay: BaseModal, Toast, Tooltip, Dropdown\n- Data display: Badge, Avatar, Spinner, ProgressBar\n- Layout: Sidebar, Toolbar, PageHeader, Divider\n\n**Theme tokens:** `--color-primary`, `--color-surface`, `--color-text`, `--color-border`, `--color-danger`, `--color-success`, `--color-warning`.",
			labels: []string{"Feature"}, assignee: "sarah", timeMin: 480,
			storyPoints: 5, sprintName: "Sprint 3 — Foundation",
			startInDays: ptr(-27), dueInDays: ptr(-20),
			closed: true, closedAtDays: ptr(-16), createdAtDays: ptr(-27)},
		{title: "User registration and login flow", col: "Done", priority: "critical",
			description: "Implement the full email/password authentication flow: registration with email verification, login with rate limiting, and the full password reset flow.\n\n**Registration:**\n- Email + password (minimum 8 characters, must include a number)\n- Email verification link valid for 24 hours\n- Welcome email sent on first successful login\n\n**Login:**\n- Rate limit: 10 attempts per 15 minutes per IP address\n- Returns a JWT access + refresh token pair on success\n\n**Password reset:**\n- Request reset → email with a 1-hour signed token\n- Reset form → all existing refresh tokens invalidated on success",
			labels: []string{"Feature"}, assignee: "priya", timeMin: 360,
			storyPoints: 5, sprintName: "Sprint 3 — Foundation",
			startInDays: ptr(-26), dueInDays: ptr(-19),
			closed: true, closedAtDays: ptr(-15), createdAtDays: ptr(-26),
			comments: []struct {
				author string
				body   string
			}{
				{"sarah", "Let's go with JWT + refresh tokens from the start — much easier to scale later."},
				{"priya", "Agreed. Done. Refresh token is stored httpOnly to prevent XSS."},
			}},
		{title: "Database schema v1 with migrations", col: "Done", priority: "high",
			description: "Design and implement the initial database schema covering users, projects, columns, cards, and memberships. Use GORM AutoMigrate so adding new fields requires only a struct change.\n\n**Core tables:**\n- `users` — authentication and profile data\n- `projects` — board type, slug, key prefix, settings\n- `project_members` — user-to-project with role\n- `columns` — ordered columns per project\n- `cards` — all card metadata including story points\n- `card_labels`, `card_tags`, `card_assignees` — join tables\n\n**Indexing strategy:** Foreign keys indexed; `(project_id, position)` composite index for ordered card queries.",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 180,
			storyPoints: 3, sprintName: "Sprint 3 — Foundation",
			startInDays: ptr(-25), dueInDays: ptr(-18),
			closed: true, closedAtDays: ptr(-16), createdAtDays: ptr(-25)},
		{title: "Deployment to staging (Docker + nginx)", col: "Done", priority: "medium",
			description: "Package the backend and frontend into Docker images and deploy to the staging environment behind an Nginx reverse proxy.\n\n**Backend image:** Multi-stage build — Go compiled in `golang:1.22-alpine`, binary runs in a `scratch` base (final image ~8 MB).\n\n**Frontend:** Vite production build output served as static files by Nginx.\n\n**Nginx configuration:**\n- `/ → static frontend files`\n- `/api/ → proxy_pass to Go backend on port 8080`\n- SSL termination via Let's Encrypt certificate\n- Gzip compression enabled for JS, CSS, and JSON responses",
			labels: []string{"Feature"}, assignee: "james", timeMin: 120,
			storyPoints: 2, sprintName: "Sprint 3 — Foundation",
			startInDays: ptr(-24), dueInDays: ptr(-17),
			closed: true, closedAtDays: ptr(-15), createdAtDays: ptr(-24)},

		// Sprint 4 — Core Features (active, -14 to +7): mix of Done and In Progress
		{title: "User dashboard overview screen", col: "Done", priority: "high",
			description: "The first screen users see after logging in: a personalised overview of assigned cards, active sprint progress, and recent team activity.\n\n**Widgets:**\n- My open cards — filtered to the current user, ordered by due date\n- Active sprint progress bar — completed points out of total sprint points\n- Recent activity feed — last 10 events across all projects the user is a member of\n- Quick-create button — opens the card creation modal with the current project pre-selected\n\n**Performance target:** All widgets loaded in parallel; skeleton loaders shown during fetch; total load time < 500 ms on a 10 Mbps connection.",
			labels: []string{"Feature"}, assignee: "sarah", timeMin: 300,
			storyPoints: 5, sprintName: "Sprint 4 — Core Features",
			startInDays: ptr(-13), dueInDays: ptr(1),
			closed: true, closedAtDays: ptr(-5), createdAtDays: ptr(-13),
			checklist: []struct {
				body string
				done bool
			}{
				{"Design approved in Figma", true},
				{"API endpoints wired up", true},
				{"Responsive layout", true},
				{"Unit tests written", true},
			}},
		{title: "Email notification service", col: "In Progress", priority: "medium",
			description: "Send transactional emails for key events: card assignment, comment mention, due-date reminder, and sprint completion.\n\n**Event → trigger mapping:**\n- Card assigned to you → immediate email\n- `@mention` in a comment → immediate email\n- Card due in 24 hours → daily digest email at 08:00 in the user's local timezone\n- Sprint completed → summary email to all sprint participants\n\n**Provider:** SMTP (configurable via `warmdesk.yaml`). Templates use Go `text/template` with a plain-text fallback for clients that block HTML.\n\n**User control:** Per-event toggles available in the notification settings page.",
			labels: []string{"Feature"}, assignee: "james", timeMin: 120,
			storyPoints: 3, sprintName: "Sprint 4 — Core Features",
			startInDays: ptr(-10), dueInDays: ptr(3), createdAtDays: ptr(-13)},
		{title: "Full-text search endpoint", col: "In Progress", priority: "high",
			description: "Implement a search endpoint that queries card titles and descriptions across all projects the current user is a member of.\n\n**Endpoint:** `GET /api/v1/search?q=<query>&limit=20`\n\n**Response fields per result:** card ref, title, project name, column name, assignee display name, closed status.\n\n**Implementation:** SQLite FTS5 virtual table for fast full-text search without an external service. For PostgreSQL deployments, use `tsvector` with a `GIN` index.\n\n**Limits:** Maximum 20 results per request; maximum query length 200 characters; cards in soft-deleted projects excluded.",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 180,
			storyPoints: 5, sprintName: "Sprint 4 — Core Features",
			startInDays: ptr(-9), dueInDays: ptr(4), createdAtDays: ptr(-12)},
		{title: "Role-based access control (RBAC)", col: "Done", priority: "high",
			description: "Implement the full RBAC model: global roles that span all projects and per-project roles for fine-grained access control.\n\n**Global roles:**\n- `admin` — access to all projects, system settings, and user management\n- `user` — standard team member\n- `viewer` — read-only access across all projects\n\n**Per-project roles:**\n- `owner` — full control including the ability to delete the project\n- `admin` — manage members and project settings\n- `member` — create, edit, and move cards\n- `viewer` — read-only access, can comment but not edit\n\n**Middleware:** `RequireProjectRole(minRole)` helper applied on every project-scoped route.",
			labels: []string{"Feature"}, assignee: "priya", timeMin: 240,
			storyPoints: 3, sprintName: "Sprint 4 — Core Features",
			startInDays: ptr(-12), dueInDays: ptr(2),
			closed: true, closedAtDays: ptr(-6), createdAtDays: ptr(-14),
			comments: []struct {
				author string
				body   string
			}{
				{"elena", "PR looks good overall — one concern about the middleware order. Discussed in PR comments."},
				{"priya", "Fixed! Ready for re-review."},
			}},
		{title: "API rate limiting middleware", col: "In Progress", priority: "medium",
			description: "Protect the API from abuse with per-IP and per-user rate limiting using a token bucket algorithm.\n\n**Limits:**\n- Unauthenticated requests: 60 per minute per IP\n- Authenticated requests: 300 per minute per user\n- Auth endpoints (login, register, password reset): 10 per 15 minutes per IP\n\n**Response on limit exceeded:** `429 Too Many Requests` with a `Retry-After` header and JSON body `{\"error\": \"rate limit exceeded\"}`.\n\n**Note:** The current implementation is per-instance in-memory. For multi-instance deployments, a Redis backend is required — documented in CLAUDE.md under horizontal scaling limitations.",
			labels: []string{"Enhancement"}, assignee: "james", timeMin: 60,
			storyPoints: 2, sprintName: "Sprint 4 — Core Features",
			startInDays: ptr(-5), dueInDays: ptr(6), createdAtDays: ptr(-13)},

		// Sprint 5 — Polish & Launch (planning): all in To Do
		{title: "Stripe payment integration", col: "To Do", priority: "critical",
			description: "Integrate Stripe Billing for subscription management so users can subscribe to a paid plan, manage their billing details, and cancel at any time.\n\n**Plans:**\n- Free: 1 project, 3 members\n- Pro: unlimited projects, 15 members ($12/month)\n- Team: unlimited projects + members, SSO ($49/month)\n\n**Flows to implement:**\n- Checkout: Stripe Checkout hosted page (avoids PCI scope)\n- Billing portal: Stripe Customer Portal for plan changes, invoice history, cancellation\n- Webhooks: `customer.subscription.updated/deleted` → update the user's `plan` field in the DB\n\n**Note:** Use Stripe test mode exclusively until QA sign-off.",
			labels: []string{"Feature"}, storyPoints: 8, sprintName: "Sprint 5 — Polish & Launch",
			dueInDays: ptr(21)},
		{title: "Analytics event tracking", col: "To Do", priority: "high",
			description: "Instrument the frontend with event tracking to understand feature adoption, funnel drop-off, and engagement patterns.\n\n**Events to track (MVP):**\n- `card.created`, `card.moved`, `card.closed`\n- `sprint.started`, `sprint.completed`\n- `project.created`\n- `report.viewed` (with report type as a property)\n- `user.invited`, `user.activated`\n\n**Provider:** Mixpanel (GDPR-compliant, EU data residency option). All tracking is gated on cookie consent.\n\n**PII rule:** Never send card titles, user names, or email addresses as event properties.",
			labels: []string{"Feature"}, storyPoints: 5, sprintName: "Sprint 5 — Polish & Launch"},
		{title: "Performance audit and optimizations", col: "To Do", priority: "medium",
			description: "Run a full performance audit across backend and frontend before the public launch and address the top findings.\n\n**Backend:**\n- Profile the 5 slowest API endpoints under 100 concurrent users using k6\n- Add missing DB indexes identified by `EXPLAIN ANALYSE`\n- Enable HTTP caching headers for read-heavy, rarely-changing endpoints\n\n**Frontend:**\n- Lighthouse audit on all 5 key pages (Performance target ≥ 90)\n- Bundle analysis: identify and eliminate unused dependencies\n- Route-level code splitting (lazy-load all non-critical views)\n\n**Success criteria:** P99 API latency ≤ 200 ms; Lighthouse Performance score ≥ 90 on the dashboard.",
			labels: []string{"Tech Debt"}, storyPoints: 3, sprintName: "Sprint 5 — Polish & Launch"},
		{title: "Launch checklist and public docs", col: "To Do", priority: "medium",
			description: "Create the go-live checklist and ensure all public-facing documentation is ready before the launch date.\n\n**Checklist:**\n- [ ] Status page configured and linked from the footer\n- [ ] Privacy Policy and Terms of Service reviewed by legal\n- [ ] GDPR Data Processing Agreement available for EU customers\n- [ ] Help centre articles covering the top 10 user journeys\n- [ ] Changelog published for v1.0\n- [ ] Support email alias configured and tested\n- [ ] Social media accounts updated with launch announcement copy ready to post",
			labels: []string{"Feature"}, storyPoints: 2, sprintName: "Sprint 5 — Polish & Launch"},

		// Backlog (no sprint)
		{title: "Multi-language support (i18n)", col: "To Do", priority: "low",
			description: "Add internationalisation support so the UI can be localised into additional languages without any code changes.\n\n**Implementation:**\n- `vue-i18n` v9 with a flat JSON key structure (`feature.key_name`)\n- All UI strings extracted to `en.json` as the base language file\n- Language selector in user settings and on the login page\n- RTL support (Arabic, Hebrew) via `dir=\"rtl\"` on `<html>`\n\n**First-wave languages** (> 5 % of our analytics): French, German, Dutch, Spanish.\n\n**Note:** The backend's 5 most common validation error messages also need localised versions.",
			labels: []string{"Enhancement"}, tags: []string{"i18n", "ux"}},
		{title: "Mobile app wrapper (Capacitor)", col: "To Do", priority: "low",
			description: "Wrap the Vue frontend in a Capacitor shell to produce iOS and Android apps from the same codebase, enabling native device features.\n\n**Scope:**\n- Basic Capacitor project setup with `@capacitor/ios` and `@capacitor/android`\n- Push notifications via `@capacitor/push-notifications`\n- Deep link handling (`warmdesk://` URL scheme)\n- App icon and splash screen assets configured for both platforms\n\n**Out of scope:** Offline mode (separate card), biometric auth (separate card), App Store submission process.",
			labels: []string{"Feature"}, tags: []string{"mobile"}},
		{title: "Fix sorting bug on dashboard table", col: "To Do", priority: "high",
			description: "The \"My Cards\" table sorts by due date incorrectly — cards without a due date appear at the top when sorting ascending, but should always appear at the bottom.\n\n**Steps to reproduce:**\n1. Create three cards: one with no due date, one due tomorrow, one due next week\n2. Sort the table by \"Due date\" ascending\n3. **Expected:** tomorrow → next week → (no due date)\n4. **Actual:** (no due date) → tomorrow → next week\n\n**Fix:** Update the sort comparator to push `null` due dates to the end regardless of sort direction:\n```js\nif (!a.due_date) return 1\nif (!b.due_date) return -1\n```",
			labels: []string{"Bug"}},
	}

	mktCards := []cardSpec{
		// Ideas
		{title: "Summer product launch campaign", col: "Ideas", priority: "high",
			description: "Plan and execute a multi-channel marketing campaign for the Q3 product launch, targeting SMB decision-makers across Europe and North America.\n\n**Proposed channels:**\n- Email: 3-part drip sequence to existing leads\n- LinkedIn: 4 sponsored posts targeting IT Directors and Heads of Product\n- Blog: 2 long-form articles + 1 customer case study\n- Webinar: Live product demo + Q&A (target: 500 registrants)\n\n**Budget:** €15,000 total. ROI target: 50 qualified leads at < €300 cost-per-lead.\n\n**Timeline:** All assets must be ready 3 weeks before the launch date. Legal review required for all ad copy.",
			labels: []string{"Campaign"}, tags: []string{"launch", "summer"}},
		{title: "Customer testimonial video series", col: "Ideas", priority: "medium",
			description: "Film a series of 3–4 short testimonial videos (60–90 seconds each) featuring real customers explaining how they use the product and what results they've achieved.\n\n**Target customers:**\n- A fast-growing startup (relatability for SMB prospects)\n- A mid-market company (credibility for enterprise prospects)\n- An agency or consultancy (shows product versatility)\n\n**Production approach:**\n- Remote recording via Riverside.fm (4K webcam footage)\n- Light editing: intro/outro branding, captions, background music\n- Deliverables: YouTube-optimised cut + vertical social cut (30 s)",
			labels: []string{"Content"}},
		// Planned
		{title: "Q2 newsletter redesign", col: "Planned", priority: "high",
			description: "Redesign the quarterly newsletter template to match the updated brand guidelines and improve click-through rates. Our current CTR is 1.8 %; the industry benchmark is 3.5 %.\n\n**Changes planned:**\n- New header design with updated logo and colour palette\n- Single-column layout for better mobile rendering\n- CTA buttons replacing plain text hyperlinks\n- Personalisation token in subject line (`Hi {{first_name}}`)\n- Improved plain-text fallback\n\n**Validation:** A/B test new vs. old template on a 20/80 split for the Q2 send. Report results within one week of sending.",
			labels: []string{"Email"}, assignee: "lisa",
			startInDays: ptr(3), dueInDays: ptr(14)},
		{title: "Influencer partnership outreach", col: "Planned", priority: "medium",
			description: "Identify and reach out to 20 micro-influencers in the project-management and developer-tools space for potential partnerships.\n\n**Selection criteria:**\n- 5k–100k followers on LinkedIn or YouTube\n- Audience profile: developers, PMs, startup founders\n- Engagement rate > 3 %\n- Location: EU or North America\n\n**Outreach approach:**\n- Personalised emails (no copy-paste templates) that reference their recent content\n- Offer: complimentary Pro account + revenue share for verified referrals\n- Track all responses in CRM under campaign tag \"Q3 Influencer 2024\"\n\n**Target:** 5 active partnerships signed by end of quarter.",
			labels: []string{"Campaign"}, assignee: "james",
			startInDays: ptr(5), dueInDays: ptr(21)},
		// In Progress
		{title: "Blog: 10 productivity tips for remote teams", col: "In Progress", priority: "medium",
			description: "Write and publish a long-form SEO blog post targeting the keyword \"productivity tips for remote teams\" (2,400 monthly searches, KD 32).\n\n**Outline:**\n1. Intro — why remote work demands different habits\n2. Tips 1–5: async communication, deep work blocks, meeting hygiene, documentation culture, time-zone etiquette\n3. Tips 6–10: the right tooling stack, structured onboarding, team rituals, visibility without surveillance, remote wellbeing\n4. CTA: link to relevant product features in tips 7 and 10\n\n**SEO requirements:**\n- Target keyword in title, first 100 words, and 2 subheadings\n- 1,500–2,000 words total\n- 3 internal links + 5 external links to authority sources",
			labels: []string{"Content"}, assignee: "sarah", timeMin: 90,
			startInDays: ptr(-5), dueInDays: ptr(3),
			comments: []struct {
				author string
				body   string
			}{
				{"lisa", "Can you work in a mention of our new features in tips 7 and 10?"},
				{"sarah", "On it — draft ready for review tomorrow."},
			}},
		{title: "LinkedIn ad campaign — EMEA region", col: "In Progress", priority: "high",
			description: "Run a LinkedIn Sponsored Content campaign targeting IT decision-makers across key EMEA markets to drive trial sign-ups.\n\n**Targeting:**\n- Job titles: Head of Product, Engineering Manager, CTO, IT Director\n- Company size: 50–500 employees\n- Geography: UK, DE, FR, NL, SE\n\n**Ad formats:**\n- Single image ad (3 creative variants for A/B testing)\n- Lead Gen Form attached (reduces friction versus an external landing page)\n\n**Budget:** €5,000 over 4 weeks. Target: 80 leads at < €65 cost-per-lead.\n\n**Reporting:** Weekly performance summary shared every Monday in the #marketing-reports Slack channel.",
			labels: []string{"Campaign"}, assignee: "james", timeMin: 120,
			startInDays: ptr(-7), dueInDays: ptr(2)},
		{title: "Trade show booth materials", col: "In Progress", priority: "critical",
			description: "Design and produce all physical and digital materials for the upcoming trade show. Shipping deadline is 5 working days before the show opens.\n\n**Physical materials:**\n- 3× roller banners (2 m × 0.8 m, printer-ready bleed-safe PDF)\n- 500× A5 double-sided flyers (matte laminate finish)\n- 200× branded notebooks + 200× branded pens\n- Demo station: 13\" MacBook Pro pre-loaded with the staging demo environment\n\n**Digital:**\n- 65\" TV loop presentation (10 slides, 15 s per slide, no text-heavy slides)\n- QR code linking to a landing page with show-specific UTM parameters",
			labels: []string{"Campaign"}, assignee: "lisa", timeMin: 180,
			startInDays: ptr(-3), dueInDays: ptr(3),
			checklist: []struct {
				body string
				done bool
			}{
				{"Banner design approved", true},
				{"Flyers printed", false},
				{"Demo laptop prepared", false},
				{"Shipping arranged", false},
			}},
		// Published
		{title: "Q1 newsletter — Spring edition", col: "Published", priority: "medium",
			description: "Produced and sent the Q1 newsletter to our 8,400-subscriber list. Results significantly above industry benchmarks.\n\n**Performance:**\n- Open rate: 31.4 % (industry avg: 22 %)\n- Click-through rate: 2.9 % (previous: 1.8 % — +61 % improvement)\n- Unsubscribes: 12 (0.14 % — well below the 0.5 % warning threshold)\n- Best-performing link: \"New features roundup\" (CTR 4.1 %)\n\n**Key learnings:** The single-column layout and personalised subject line drove the CTR improvement. Both changes will be carried forward into the Q2 template redesign.",
			labels: []string{"Email"}, assignee: "lisa", timeMin: 120,
			startInDays: ptr(-20), dueInDays: ptr(-12),
			closed: true, closedAtDays: ptr(-12), createdAtDays: ptr(-20)},
		{title: "Product Hunt launch announcement", col: "Published", priority: "high",
			description: "Coordinated and executed the Product Hunt launch for v1.0, achieving #3 Product of the Day.\n\n**Results:**\n- 847 upvotes on launch day\n- #3 Product of the Day / #12 Product of the Week\n- 1,240 new sign-ups attributed to Product Hunt in the first 72 hours\n- 34 reviews with an average rating of 4.8 ★\n\n**What worked:**\n- Hunter with 2,400 followers secured 2 weeks in advance\n- Launch timed at 00:01 PST to maximise the full voting window\n- 6-hour moderated comment thread with founder responses to every comment\n- Email blast to the waitlist timed to coincide with the launch going live",
			labels: []string{"Campaign"}, assignee: "admin", timeMin: 60,
			startInDays: ptr(-30), dueInDays: ptr(-25),
			closed: true, closedAtDays: ptr(-25), createdAtDays: ptr(-30)},
		{title: "Year in review blog post", col: "Published", priority: "low",
			description: "Annual review post summarising product milestones, team growth, and key metrics for the year.\n\n**Sections covered:**\n- By the numbers: total users, projects created, cards completed\n- Top 5 product features shipped during the year\n- Team growth: new hires, office expansion, culture highlights\n- Community: events attended, partnerships formed, open-source contributions\n- What's coming next: roadmap teaser for the year ahead\n\n**Amplification:**\n- Sent to all newsletter subscribers on day of publish\n- LinkedIn post + Twitter/X highlight thread\n- Pinned to the top of the company blog throughout January",
			labels: []string{"Content"}, assignee: "sarah", timeMin: 150,
			startInDays: ptr(-45), dueInDays: ptr(-38),
			closed: true, closedAtDays: ptr(-38), createdAtDays: ptr(-45)},
	}

	apiCards := []cardSpec{
		// Sprint 1 — Bootstrap (completed, -84 to -70): all in Done
		{title: "Design API schema and OpenAPI spec", col: "Done", priority: "high",
			description: "Define the complete API surface before writing any implementation code. The OpenAPI 3.1 spec will be the single source of truth for generating server stubs, SDK clients, and documentation.\n\n**Deliverables:**\n- `openapi.yaml` with all resource schemas, endpoints, and error envelope definitions\n- Authentication section documenting the API key flow (header and query param)\n- At least 3 example values per schema property\n- Reviewed and signed off by all stakeholders\n\n**Tooling:** Stoplight Studio for editing; Redocly CLI for local preview. Spec committed to `/api/openapi.yaml`.",
			labels: []string{"Feature"}, assignee: "elena", timeMin: 480,
			storyPoints: 5, sprintName: "Sprint 1 — Bootstrap",
			startInDays: ptr(-84), dueInDays: ptr(-77),
			closed: true, closedAtDays: ptr(-73), createdAtDays: ptr(-84)},
		{title: "Set up Go project structure with CI/CD", col: "Done", priority: "high",
			description: "Bootstrap the Go project with a clean package layout, linting configuration, and a GitHub Actions CI pipeline.\n\n**Package layout:**\n```\n/cmd/server   entry point\n/handlers     HTTP handlers\n/middleware   auth, rate limit, CORS\n/models       GORM model structs\n/services     business logic\n/database     DB init and migrations\n```\n\n**CI pipeline:**\n- `go vet` + `staticcheck` + `golangci-lint`\n- Build verification on Go 1.22 and 1.23\n- Docker build + push to ECR on every merge to `main`",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 300,
			storyPoints: 3, sprintName: "Sprint 1 — Bootstrap",
			startInDays: ptr(-83), dueInDays: ptr(-76),
			closed: true, closedAtDays: ptr(-73), createdAtDays: ptr(-83)},
		{title: "Implement health-check and metrics endpoints", col: "Done", priority: "medium",
			description: "Expose operational endpoints for load balancers and monitoring systems.\n\n**Endpoints:**\n- `GET /health` — returns `200 OK` with `{\"status\": \"ok\", \"db\": \"ok\"}`. Returns `503` if the database is unreachable.\n- `GET /metrics` — Prometheus text-format metrics\n\n**Metrics exposed:** `http_requests_total`, `http_request_duration_seconds` (histogram), `db_connections_active`.\n\n**Access control:** `/metrics` restricted to the internal network via an IP allowlist middleware. `/health` is public and used by the load balancer's health check.",
			labels: []string{"Feature"}, assignee: "james", timeMin: 180,
			storyPoints: 3, sprintName: "Sprint 1 — Bootstrap",
			startInDays: ptr(-82), dueInDays: ptr(-75),
			closed: true, closedAtDays: ptr(-72), createdAtDays: ptr(-82)},
		{title: "Docker containerisation and registry setup", col: "Done", priority: "high",
			description: "Package the API as a minimal Docker image and push it to the private ECR registry.\n\n**Dockerfile approach:** Multi-stage build — Go compiled with `CGO_ENABLED=0` in a `golang:1.22-alpine` builder stage; the resulting binary runs in a `scratch` base image. Final image size: ~8 MB.\n\n**Registry:** AWS ECR `api-platform` repository. Lifecycle policy: retain last 10 tagged images; delete untagged images after 1 day.\n\n**CI integration:** Image built and pushed on every green build on `main`.",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 240,
			storyPoints: 5, sprintName: "Sprint 1 — Bootstrap",
			startInDays: ptr(-81), dueInDays: ptr(-74),
			closed: true, closedAtDays: ptr(-72), createdAtDays: ptr(-81)},
		{title: "Error response standardisation (RFC 7807)", col: "Done", priority: "medium",
			description: "Standardise all API error responses to follow RFC 7807 (Problem Details for HTTP APIs) so clients receive consistent, machine-readable errors.\n\n**Standard error format:**\n```json\n{\n  \"type\": \"https://api.example.com/errors/validation-error\",\n  \"title\": \"Validation Error\",\n  \"status\": 422,\n  \"detail\": \"The 'email' field must be a valid email address.\",\n  \"instance\": \"/api/v1/users\"\n}\n```\n\n**Error types defined:** `validation-error`, `not-found`, `forbidden`, `rate-limited`, `internal-error`.\n\n**Note:** This is a breaking change for any existing error consumers — coordinate rollout accordingly.",
			labels: []string{"Enhancement"}, assignee: "elena", timeMin: 120,
			storyPoints: 2, sprintName: "Sprint 1 — Bootstrap",
			startInDays: ptr(-80), dueInDays: ptr(-73),
			closed: true, closedAtDays: ptr(-71), createdAtDays: ptr(-80)},

		// Sprint 2 — Core API (completed, -70 to -56): all in Done
		{title: "Implement resource CRUD endpoints", col: "Done", priority: "high",
			description: "Implement the full Create, Read, Update, Delete endpoint set for the primary API resource.\n\n**Endpoints:**\n- `POST /api/v1/resources` — create\n- `GET /api/v1/resources` — list (with pagination)\n- `GET /api/v1/resources/:id` — retrieve by ID\n- `PUT /api/v1/resources/:id` — full replacement update\n- `PATCH /api/v1/resources/:id` — partial update\n- `DELETE /api/v1/resources/:id` — soft delete\n\n**Validation:** All inputs validated with `go-playground/validator` struct tags. Missing required fields return `422` with field-level error details per the RFC 7807 standard.",
			labels: []string{"Feature"}, assignee: "elena", timeMin: 480,
			storyPoints: 5, sprintName: "Sprint 2 — Core API",
			startInDays: ptr(-70), dueInDays: ptr(-63),
			closed: true, closedAtDays: ptr(-59), createdAtDays: ptr(-70)},
		{title: "Pagination and filtering support", col: "Done", priority: "high",
			description: "Add offset-based pagination and field filtering to all list endpoints.\n\n**Pagination parameters:** `?page=1&per_page=20` (default), max `per_page=100`.\n\n**Filtering parameters:**\n- `?status=active` — filter by status field\n- `?created_after=2024-01-01` — date range filter\n- `?q=search+term` — full-text search across title and description\n\n**Standard response envelope:**\n```json\n{\"data\": [...], \"meta\": {\"page\": 1, \"per_page\": 20, \"total\": 143}}\n```\n\nCursor-based pagination is deferred to a later sprint.",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 360,
			storyPoints: 5, sprintName: "Sprint 2 — Core API",
			startInDays: ptr(-69), dueInDays: ptr(-62),
			closed: true, closedAtDays: ptr(-59), createdAtDays: ptr(-69)},
		{title: "Request validation middleware", col: "Done", priority: "medium",
			description: "Centralise request validation so all handlers receive clean, validated data and produce consistent error responses for any invalid input.\n\n**Approach:**\n- Middleware binds and validates the request body using `go-playground/validator` struct tags\n- Path and query parameters validated inline within handlers\n- Validation failures return `422 Unprocessable Entity` with an array of field-level errors:\n\n```json\n{\"errors\": [{\"field\": \"email\", \"message\": \"must be a valid email\"}]}\n```\n\n**Custom validators registered:** `api_key_format`, `slug` (alphanumeric + hyphens only), `date_range` (start must be ≤ end).",
			labels: []string{"Enhancement"}, assignee: "james", timeMin: 240,
			storyPoints: 3, sprintName: "Sprint 2 — Core API",
			startInDays: ptr(-68), dueInDays: ptr(-61),
			closed: true, closedAtDays: ptr(-58), createdAtDays: ptr(-68)},
		{title: "Database migrations with versioning", col: "Done", priority: "high",
			description: "Implement a versioned migration system so schema changes can be applied, rolled back, and tracked consistently across all environments.\n\n**Tooling:** `golang-migrate` with sequential numbered SQL files.\n\n**File structure:** `/migrations/NNNN_name.up.sql` and `.down.sql`.\n\n**Workflow:**\n- CI: migrations run automatically before integration tests\n- Production: migrations run as a pre-deploy job in the pipeline\n\n**Rules:**\n- Never modify a migration that has been applied to any environment\n- Every `up` migration must have a corresponding `down`\n- Long-running index creation must use `CONCURRENTLY` to avoid table locks",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 300,
			storyPoints: 5, sprintName: "Sprint 2 — Core API",
			startInDays: ptr(-67), dueInDays: ptr(-60),
			closed: true, closedAtDays: ptr(-58), createdAtDays: ptr(-67)},
		{title: "Structured logging with correlation IDs", col: "Done", priority: "medium",
			description: "Replace ad-hoc `fmt.Println` calls with structured JSON logging and add correlation IDs to trace requests end-to-end.\n\n**Library:** `zerolog` (zero-allocation, structured).\n\n**Correlation ID flow:**\n1. Ingress middleware generates a UUID if `X-Request-ID` header is absent\n2. ID propagated through `context.Context` to all downstream calls\n3. Included in every log line as `request_id` field\n4. Returned to the client in the `X-Request-ID` response header\n\n**Log levels:** `debug` in development, `info` in production, `error` automatically logged for all 5xx responses.",
			labels: []string{"Enhancement"}, assignee: "elena", timeMin: 180,
			storyPoints: 3, sprintName: "Sprint 2 — Core API",
			startInDays: ptr(-66), dueInDays: ptr(-59),
			closed: true, closedAtDays: ptr(-57), createdAtDays: ptr(-66)},

		// Sprint 3 — Auth & Docs (completed, -56 to -42): all in Done
		{title: "API key authentication and rotation", col: "Done", priority: "critical",
			description: "Implement API key authentication as the primary mechanism for machine-to-machine integrations.\n\n**Key format:** `wdk_` prefix + 32 bytes of base64url-encoded random data (e.g. `wdk_abc123...`).\n\n**Storage:** Only the SHA-256 hash is stored in the database — the raw key is shown exactly once at creation.\n\n**Authentication:** `Authorization: Bearer wdk_...` header or `?api_key=` query parameter.\n\n**Rotation:** Users can create a replacement key before revoking the old one (zero-downtime rotation). The old key remains valid for 24 hours after a replacement is activated.",
			labels: []string{"Security"}, assignee: "elena", timeMin: 480,
			storyPoints: 5, sprintName: "Sprint 3 — Auth & Docs",
			startInDays: ptr(-56), dueInDays: ptr(-49),
			closed: true, closedAtDays: ptr(-45), createdAtDays: ptr(-56)},
		{title: "Rate limiting per API key", col: "Done", priority: "high",
			description: "Enforce rate limits at the individual API key level to prevent abuse and ensure fair use across all tenants.\n\n**Limits:**\n- Default: 1,000 requests per minute per API key\n- Burst: up to 100 requests per second (token bucket algorithm)\n- Auth endpoints (key creation, rotation): 10 requests per 15 minutes per IP\n\n**Response headers on every request:**\n```\nX-RateLimit-Limit: 1000\nX-RateLimit-Remaining: 847\nX-RateLimit-Reset: 1712345678\n```\n\n**429 response:** includes a `Retry-After` header with seconds until the limit resets.\n\n**Storage:** Redis for distributed enforcement; in-memory fallback for single-instance deployments.",
			labels: []string{"Security"}, assignee: "raj", timeMin: 360,
			storyPoints: 5, sprintName: "Sprint 3 — Auth & Docs",
			startInDays: ptr(-55), dueInDays: ptr(-48),
			closed: true, closedAtDays: ptr(-44), createdAtDays: ptr(-55)},
		{title: "Interactive API documentation (Swagger UI)", col: "Done", priority: "medium",
			description: "Serve Swagger UI at `/docs` so developers can explore and test all API endpoints directly in the browser without any client setup.\n\n**Implementation:**\n- The `openapi.yaml` spec is embedded in the binary using `//go:embed`\n- Swagger UI assets served from `/docs` (loaded from CDN)\n- `/docs/openapi.yaml` serves the raw spec for tooling integration\n\n**Auth in the UI:** The `ApiKey` security scheme is pre-configured so developers can paste their key and immediately make authenticated test calls.\n\n**Access:** Public in staging; in production `/docs` is accessible only to authenticated users.",
			labels: []string{"Feature"}, assignee: "james", timeMin: 240,
			storyPoints: 3, sprintName: "Sprint 3 — Auth & Docs",
			startInDays: ptr(-54), dueInDays: ptr(-47),
			closed: true, closedAtDays: ptr(-44), createdAtDays: ptr(-54)},
		{title: "Webhook delivery with retry logic", col: "Done", priority: "high",
			description: "Allow API consumers to register webhook URLs that receive signed event payloads when resources are created, updated, or deleted.\n\n**Event types:** `resource.created`, `resource.updated`, `resource.deleted`.\n\n**Delivery details:**\n- Signed with HMAC-SHA256 (`X-Webhook-Signature` header)\n- Per-delivery timeout: 5 seconds\n- Retry schedule: immediate → 1 min → 5 min → 30 min → 2 hr → 24 hr (6 total attempts)\n- Dead-lettered after 6 failures; account owner notified by email\n\n**Dashboard:** Users can view delivery history and manually re-trigger any failed delivery.",
			labels: []string{"Feature"}, assignee: "priya", timeMin: 300,
			storyPoints: 3, sprintName: "Sprint 3 — Auth & Docs",
			startInDays: ptr(-53), dueInDays: ptr(-46),
			closed: true, closedAtDays: ptr(-43), createdAtDays: ptr(-53)},
		{title: "Audit log for all API mutations", col: "Done", priority: "medium",
			description: "Record every state-changing API call to an append-only audit log for security, compliance, and post-incident debugging.\n\n**Fields logged per event:**\n- Actor: user ID or API key ID\n- Action: HTTP method + path\n- Resource type and ID\n- Before/after state diff (for UPDATE operations)\n- Timestamp, source IP address, correlation request ID\n\n**Access:** Audit log queryable by account owners via `GET /api/v1/audit-log`.\n\n**Retention:** 90 days queryable online; 2 years in cold storage (S3 + Glacier).\n\n**Integrity constraint:** The audit table has no UPDATE or DELETE — append-only enforced at the database level.",
			labels: []string{"Security"}, assignee: "elena", timeMin: 120,
			storyPoints: 2, sprintName: "Sprint 3 — Auth & Docs",
			startInDays: ptr(-52), dueInDays: ptr(-45),
			closed: true, closedAtDays: ptr(-43), createdAtDays: ptr(-52)},

		// Sprint 4 — SDKs & Testing (completed, -42 to -28): all in Done
		{title: "Python SDK for the public API", col: "Done", priority: "high",
			description: "Publish an official Python SDK to PyPI so developers can integrate with the API without writing raw HTTP client code.\n\n**Package name:** `warmdesk-python` on PyPI.\n\n**Features:**\n- Fully typed with PEP 484 type hints and `py.typed` marker\n- Sync and async clients (`requests` + `httpx`)\n- Automatic retry with exponential backoff on 429 and 5xx responses\n- `list_all()` pagination helper that auto-iterates through all pages\n- Full test suite (pytest) with VCR cassettes for offline testing\n\n**Docs:** Auto-generated via `pdoc`, published to `docs.example.com/python-sdk`.",
			labels: []string{"Feature"}, assignee: "elena", timeMin: 480,
			storyPoints: 5, sprintName: "Sprint 4 — SDKs & Testing",
			startInDays: ptr(-42), dueInDays: ptr(-35),
			closed: true, closedAtDays: ptr(-31), createdAtDays: ptr(-42)},
		{title: "JavaScript / TypeScript SDK", col: "Done", priority: "high",
			description: "Publish an official JavaScript / TypeScript SDK to npm for use in browsers, Node.js, and edge runtimes.\n\n**Package name:** `@warmdesk/sdk` on npm.\n\n**Features:**\n- Written in TypeScript; ships with `.d.ts` type declarations\n- Dual ESM + CJS build output\n- Compatible with Node.js 18+, modern browsers, and Cloudflare Workers\n- Automatic retry with exponential backoff\n- `async for await (const page of client.resources.list())` async iterator for pagination\n\n**Bundle size target:** < 15 KB minified + gzipped with no runtime dependencies.",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 420,
			storyPoints: 5, sprintName: "Sprint 4 — SDKs & Testing",
			startInDays: ptr(-41), dueInDays: ptr(-34),
			closed: true, closedAtDays: ptr(-30), createdAtDays: ptr(-41)},
		{title: "End-to-end integration test suite", col: "Done", priority: "high",
			description: "Build a comprehensive integration test suite that runs against a live test database (not mocks) to catch contract breakages before each release.\n\n**Coverage:**\n- All CRUD endpoints: happy path and common error cases\n- Authentication: valid key, expired key, revoked key, missing key\n- Rate limiting: verify a `429` response is returned after the limit is exceeded\n- Webhooks: delivery attempt and retry behaviour\n- Pagination: first page, last page, and cursor boundary\n\n**Tooling:** Go `testing` package + `testcontainers-go` to manage the PostgreSQL lifecycle in CI.\n\n**CI:** Integration tests run on every PR against a fresh, isolated database container.",
			labels: []string{"Enhancement"}, assignee: "priya", timeMin: 360,
			storyPoints: 3, sprintName: "Sprint 4 — SDKs & Testing",
			startInDays: ptr(-40), dueInDays: ptr(-33),
			closed: true, closedAtDays: ptr(-30), createdAtDays: ptr(-40)},
		{title: "Load testing and performance benchmarks", col: "Done", priority: "medium",
			description: "Establish performance baselines and verify the API meets SLA targets under realistic load before the public launch.\n\n**Tool:** k6 with scenarios modelled on production traffic patterns.\n\n**Test scenarios:**\n- Baseline: 100 VUs, 5-minute steady state\n- Spike: ramp to 1,000 VUs in 30 seconds, hold 1 minute, ramp down\n- Soak: 200 VUs for 30 minutes\n\n**SLA targets:**\n- P50 latency ≤ 50 ms\n- P99 latency ≤ 200 ms\n- Error rate < 0.1 % at 500 RPS\n\nResults stored in `/load-tests/results/` and summarised in the pre-launch readiness report.",
			labels: []string{"Enhancement"}, assignee: "james", timeMin: 300,
			storyPoints: 3, sprintName: "Sprint 4 — SDKs & Testing",
			startInDays: ptr(-39), dueInDays: ptr(-32),
			closed: true, closedAtDays: ptr(-29), createdAtDays: ptr(-39)},
		{title: "API changelog and versioning policy", col: "Done", priority: "low",
			description: "Document the API versioning policy and publish the first changelog entry so integrators have clear expectations about future compatibility.\n\n**Versioning policy:**\n- Major version in the URL path (`/api/v1/`, `/api/v2/`)\n- Minor and patch changes are non-breaking (additive only)\n- Breaking changes require a new major version with a minimum 6-month parallel support period\n- Deprecation signalled via response headers: `Deprecation: true` and `Sunset: <date>`\n\n**Changelog format:** Keep A Changelog (`CHANGELOG.md` in the repo + `GET /api/changelog` endpoint returning JSON for programmatic consumption).",
			labels: []string{"Feature"}, assignee: "elena", timeMin: 120,
			storyPoints: 2, sprintName: "Sprint 4 — SDKs & Testing",
			startInDays: ptr(-38), dueInDays: ptr(-31),
			closed: true, closedAtDays: ptr(-29), createdAtDays: ptr(-38)},

		// Sprint 5 — Stabilisation (active, -14 to +7): mix of Done and In Progress
		{title: "Deprecate v0 endpoints and notify users", col: "Done", priority: "high",
			description: "The v0 API has been superseded by v1. This card covers the formal deprecation communication and the planned removal after the sunset period.\n\n**Deprecation timeline:**\n- Today: `Deprecation: true` and `Sunset: <date+90days>` headers added to all v0 responses\n- Day 30: Email to all API key holders with recorded v0 usage in the past 7 days\n- Day 60: Second warning email + `api.v0.sunset_approaching` webhook event\n- Day 90: v0 endpoints removed; any remaining requests receive `410 Gone` with a migration link\n\n**Migration guide:** Published at `/docs/migrate-v0-to-v1`.",
			labels: []string{"Enhancement"}, assignee: "elena", timeMin: 180,
			storyPoints: 3, sprintName: "Sprint 5 — Stabilisation",
			startInDays: ptr(-14), dueInDays: ptr(-8),
			closed: true, closedAtDays: ptr(-9), createdAtDays: ptr(-14)},
		{title: "Fix response envelope inconsistency bug", col: "Done", priority: "critical",
			description: "Some list endpoints return the array directly while others wrap it in `{\"data\": [...]}`. This inconsistency breaks the SDK auto-pagination logic.\n\n**Affected endpoints:**\n- `GET /api/v0/webhooks` — returns `[...]` directly instead of `{\"data\": [...]}`\n- `GET /api/v0/audit-log` — returns `{\"logs\": [...]}` instead of `{\"data\": [...]}`\n\n**Fix:** Wrap all list responses in the standard `{\"data\": [...], \"meta\": {...}}` envelope.\n\n**Coordination required:** This is a breaking change for v0 consumers. Notify affected integrators and pair the fix with SDK patch releases.",
			labels: []string{"Bug"}, assignee: "raj", timeMin: 120,
			storyPoints: 2, sprintName: "Sprint 5 — Stabilisation",
			startInDays: ptr(-13), dueInDays: ptr(-7),
			closed: true, closedAtDays: ptr(-10), createdAtDays: ptr(-13)},
		{title: "Add batch API endpoints for bulk operations", col: "In Review", priority: "high",
			description: "Allow API consumers to perform multiple operations in a single HTTP request, reducing round-trips for bulk workflows.\n\n**Endpoints:**\n- `POST /api/v1/resources/batch` — create up to 100 resources\n- `PATCH /api/v1/resources/batch` — update up to 100 resources by ID\n- `DELETE /api/v1/resources/batch` — delete up to 100 resources by ID\n\n**Request format:**\n```json\n{\"operations\": [{\"id\": \"123\", \"data\": {...}}, ...]}\n```\n\n**Response:** `207 Multi-Status` with a per-item status code so partial failures are clearly surfaced rather than returning a single pass/fail for the whole batch.",
			labels: []string{"Feature"}, assignee: "priya", timeMin: 300,
			storyPoints: 5, sprintName: "Sprint 5 — Stabilisation",
			startInDays: ptr(-12), dueInDays: ptr(3), createdAtDays: ptr(-12),
			comments: []struct {
				author string
				body   string
			}{
				{"elena", "Design looks solid. Make sure the response envelope is consistent with single-resource endpoints."},
				{"priya", "Good catch — I'll normalise the wrapper before merging."},
			}},
		{title: "Improve error messages with actionable hints", col: "In Progress", priority: "medium",
			description: "Current error messages are terse and unhelpful. Adding actionable hints will let developers self-serve common errors without opening a support ticket.\n\n**Before:** `{\"error\": \"validation failed\"}`\n\n**After:**\n```json\n{\n  \"title\": \"Validation Error\",\n  \"detail\": \"The 'created_after' field must be an ISO 8601 date (e.g. '2024-03-15').\",\n  \"docs\": \"https://api.example.com/docs/errors/validation-error\"\n}\n```\n\n**Priority focus areas:** Date format errors, pagination parameter validation, missing required field messages, and rate limit error explanations.",
			labels: []string{"Enhancement"}, assignee: "james", timeMin: 90,
			storyPoints: 3, sprintName: "Sprint 5 — Stabilisation",
			startInDays: ptr(-10), dueInDays: ptr(4), createdAtDays: ptr(-10)},
		{title: "Support cursor-based pagination", col: "In Progress", priority: "high",
			description: "Add cursor-based pagination as an alternative to offset pagination for large, frequently-updated collections where offset pagination produces inconsistent results.\n\n**Query parameters:** `?cursor=<opaque_token>&per_page=20`\n\n**Response format:**\n```json\n{\n  \"data\": [...],\n  \"meta\": {\"next_cursor\": \"eyJpZCI6IDEyM30=\", \"has_more\": true}\n}\n```\n\n**Cursor encoding:** Base64-encoded JSON `{\"id\": <last_id>, \"created_at\": \"<iso8601>\"}` — opaque and URL-safe.\n\n**Known limitation:** Cursor pagination does not support random-page access or total result counts — document this clearly in the API reference.",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 120,
			storyPoints: 5, sprintName: "Sprint 5 — Stabilisation",
			startInDays: ptr(-9), dueInDays: ptr(5), createdAtDays: ptr(-9)},

		// Sprint 6 — GA Launch (planning): all in To Do
		{title: "Public developer portal and onboarding docs", col: "To Do", priority: "high",
			description: "Build the developer portal that serves as the front door for all API consumers — docs, quickstart, API reference, and community resources.\n\n**Sections:**\n- Quickstart guide (5 minutes to first successful API call)\n- Authentication and key management guide\n- Full API reference (Redoc rendering of the OpenAPI spec)\n- SDK guides with code examples for Python, JavaScript/TypeScript, and cURL\n- Changelog and release history\n- Status page embed\n- Community Slack invite link\n\n**Tooling:** Mintlify for the docs site; auto-redeployed on spec changes via GitHub Actions.",
			labels: []string{"Feature"}, storyPoints: 8, sprintName: "Sprint 6 — GA Launch",
			dueInDays: ptr(28)},
		{title: "API key self-service management console", col: "To Do", priority: "high",
			description: "Let users create, name, and revoke their own API keys from the account settings page — no support ticket required.\n\n**Features:**\n- Create a key: set a human-readable name and an optional expiry date\n- Key list: shows name, last-used timestamp, creation date, and expiry\n- Revoke a key: immediate effect with a confirmation dialog\n- Copy-to-clipboard: key value displayed once at creation only\n\n**Security constraints:**\n- Raw key value is never stored and cannot be retrieved after the creation screen is closed\n- Support staff cannot view raw key values in the admin panel\n- Rate limit on key creation: maximum 10 active keys per account",
			labels: []string{"Feature"}, storyPoints: 5, sprintName: "Sprint 6 — GA Launch"},
		{title: "SLA monitoring and status-page integration", col: "To Do", priority: "medium",
			description: "Publish a public status page backed by uptime monitoring so API consumers can check service health and subscribe to incident notifications without contacting support.\n\n**Monitoring:** UptimeRobot (or equivalent) checking `/health` every 30 seconds from 3 geographic regions.\n\n**Status page:** Hosted at `status.api.example.com`. Components tracked: API, Webhooks, Dashboard, Authentication.\n\n**Incident workflow:**\n1. Monitor triggers → incident automatically created on the status page\n2. On-call engineer updates the incident with impact assessment and timeline\n3. Resolution → post-mortem published within 48 hours\n\n**Subscriptions:** Email and RSS feeds available for consumers.",
			labels: []string{"Enhancement"}, storyPoints: 3, sprintName: "Sprint 6 — GA Launch"},
		{title: "Launch announcement and developer blog post", col: "To Do", priority: "low",
			description: "Write the GA launch announcement and coordinate publication across all channels on the launch day.\n\n**Blog post outline:**\n1. Why we built this API and what problem it solves for developers\n2. Key features: authentication, rate limiting, SDKs, and webhooks\n3. Performance numbers from load testing\n4. A quickstart code snippet (30 lines to working integration)\n5. What's coming in v1.1 — a roadmap teaser\n\n**Publication channels:**\n- Company engineering blog\n- Hacker News (Show HN post)\n- Dev.to cross-post\n- LinkedIn and Twitter/X thread\n- SDK README banners updated to reflect stable release",
			labels: []string{"Feature"}, storyPoints: 2, sprintName: "Sprint 6 — GA Launch"},

		// Backlog (no sprint)
		{title: "GraphQL gateway layer", col: "To Do", priority: "medium",
			description: "Add a GraphQL gateway in front of the REST API so clients can fetch exactly the data they need in a single round-trip, eliminating over-fetching.\n\n**Implementation options under evaluation:**\n- `gqlgen` — Go-native, type-safe, fits our existing stack (preferred)\n- GraphQL Mesh — adapter layer over the OpenAPI spec (faster to ship, less control)\n\n**Endpoint:** `POST /graphql` alongside the existing REST endpoints — a complement, not a replacement.\n\n**MVP schema:** Resource, User, Webhook, and AuditLog types with standard CRUD queries and mutations.\n\n**Known trade-off:** GraphQL caching is significantly more complex than REST caching — document this clearly for consumers.",
			labels: []string{"Enhancement"}, tags: []string{"graphql", "api"}},
		{title: "Multi-region failover support", col: "To Do", priority: "high",
			description: "Enable the API to survive a complete cloud region outage with a Recovery Time Objective (RTO) of less than 15 minutes.\n\n**Target architecture:**\n- Active-passive: primary in `eu-west-1`, standby in `eu-central-1`\n- Database: RDS Multi-AZ + cross-region read replica promoted on failover\n- DNS failover: Route 53 health-check-based routing\n- Redis: ElastiCache Global Datastore for cross-region replication\n\n**Runbook:** Failover procedure documented in `/ops/runbooks/region-failover.md`. Drilled quarterly.\n\n**RTO target:** < 15 minutes (DNS propagation + application warmup). **RPO target:** < 1 minute (replication lag).",
			labels: []string{"Security"}},
		{title: "Fix null pointer in optional field serialisation", col: "To Do", priority: "high",
			description: "When a resource is created with an optional field explicitly set to `null`, the JSON serialiser panics with a nil pointer dereference instead of emitting `\"field\": null`.\n\n**Steps to reproduce:**\n```bash\ncurl -X POST /api/v1/resources \\\n  -H 'Content-Type: application/json' \\\n  -d '{\"optional_field\": null}'\n# Returns: 500 Internal Server Error\n```\n\n**Root cause:** The custom `MarshalJSON` method dereferences the `*string` pointer without a nil check.\n\n**Fix:**\n```go\nif v.OptionalField == nil {\n    return json.Marshal(nil)\n}\nreturn json.Marshal(*v.OptionalField)\n```",
			labels: []string{"Bug"}},
	}

	projectCards := map[string][]cardSpec{
		"website-redesign": webCards,
		"mobile-app-v2":    mobCards,
		"devops-infra":     infCards,
		"product-platform": pltCards,
		"api-platform":     apiCards,
		"marketing":        mktCards,
	}

	createdCards := map[string]*models.Card{} // key: "slug/title"

	totalCards := 0
	for slug, specs := range projectCards {
		pd := projects[slug]
		for i, spec := range specs {
			col := pd.cols[spec.col]
			if col == nil {
				log.Fatalf("seed: unknown column %q in project %s", spec.col, slug)
			}

			// Increment project card counter
			must(db.Model(&models.Project{}).Where("id = ?", pd.project.ID).
				UpdateColumn("card_counter", gorm.Expr("card_counter + 1")).Error)
			var proj models.Project
			must(db.Select("card_counter").First(&proj, pd.project.ID).Error)

			var assigneeID *uint
			if spec.assignee != "" {
				assigneeID = &users[spec.assignee].ID
			}

			priority := spec.priority
			if priority == "" {
				priority = "none"
			}

			var startDate, dueDate *time.Time
			if spec.startInDays != nil {
				startDate = days(*spec.startInDays)
			}
			if spec.dueInDays != nil {
				dueDate = days(*spec.dueInDays)
			}

			var sp *int
			if spec.storyPoints > 0 {
				v := spec.storyPoints
				sp = &v
			}
			var closedAt *time.Time
			if spec.closed && spec.closedAtDays != nil {
				t := time.Now().UTC().AddDate(0, 0, *spec.closedAtDays)
				closedAt = &t
			}
			card := &models.Card{
				ColumnID:         col.ID,
				ProjectID:        pd.project.ID,
				Title:            spec.title,
				Description:      spec.description,
				Priority:         priority,
				AssigneeID:       assigneeID,
				CreatedByID:      users["admin"].ID,
				Position:         float64((i + 1) * 1000),
				CardNumber:       proj.CardCounter,
				StartDate:        startDate,
				DueDate:          dueDate,
				TimeSpentMinutes: spec.timeMin,
				StoryPoints:      sp,
				Closed:           spec.closed,
				ClosedAt:         closedAt,
			}
			must(db.Create(card).Error)
			cardHist := &models.CardHistory{CardID: card.ID, UserID: users["admin"].ID, EventType: "created"}
			must(db.Create(cardHist).Error)
			if spec.createdAtDays != nil {
				createdAt := time.Now().UTC().AddDate(0, 0, *spec.createdAtDays).Truncate(24 * time.Hour)
				must(db.Model(card).UpdateColumn("created_at", createdAt).Error)
				must(db.Model(cardHist).UpdateColumn("created_at", createdAt).Error)
			}
			createdCards[slug+"/"+spec.title] = card

			// Labels
			for _, lname := range spec.labels {
				if lbl, ok := pd.labels[lname]; ok {
					must(db.Exec("INSERT INTO card_labels (card_id, label_id) VALUES (?, ?)", card.ID, lbl.ID).Error)
				}
			}

			// Tags
			for _, tag := range spec.tags {
				must(db.Create(&models.CardTag{CardID: card.ID, Name: tag}).Error)
			}

			// Multi-assignees (mirror the single assignee for display consistency)
			if assigneeID != nil {
				must(db.Exec("INSERT OR IGNORE INTO card_assignees (card_id, user_id) VALUES (?, ?)", card.ID, *assigneeID).Error)
			}

			// Checklist
			for j, item := range spec.checklist {
				must(db.Create(&models.CardChecklistItem{
					CardID:      card.ID,
					Body:        item.body,
					IsCompleted: item.done,
					Position:    float64((j + 1) * 1000),
				}).Error)
			}

			// Comments
			for _, c := range spec.comments {
				author := users[c.author]
				if author == nil {
					author = users["admin"]
				}
				must(db.Create(&models.CardComment{
					CardID: card.ID, UserID: author.ID, Body: c.body,
				}).Error)
				must(db.Create(&models.CardHistory{CardID: card.ID, UserID: author.ID, EventType: "comment_added"}).Error)
			}

			// Sub-cards
			for si, sub := range spec.subCards {
				must(db.Model(&models.Project{}).Where("id = ?", pd.project.ID).
					UpdateColumn("card_counter", gorm.Expr("card_counter + 1")).Error)
				var subProj models.Project
				must(db.Select("card_counter").First(&subProj, pd.project.ID).Error)

				var subAssigneeID *uint
				if sub.assignee != "" {
					subAssigneeID = &users[sub.assignee].ID
				}
				subPriority := sub.priority
				if subPriority == "" {
					subPriority = "none"
				}
				subCard := &models.Card{
					ColumnID:     col.ID,
					ProjectID:    pd.project.ID,
					ParentCardID: &card.ID,
					Title:        sub.title,
					Priority:     subPriority,
					AssigneeID:   subAssigneeID,
					CreatedByID:  users["admin"].ID,
					Position:     float64((si + 1) * 1000),
					CardNumber:   subProj.CardCounter,
				}
				must(db.Create(subCard).Error)
				must(db.Create(&models.CardHistory{CardID: subCard.ID, UserID: users["admin"].ID, EventType: "created"}).Error)
				if subAssigneeID != nil {
					must(db.Exec("INSERT OR IGNORE INTO card_assignees (card_id, user_id) VALUES (?, ?)", subCard.ID, *subAssigneeID).Error)
				}
				totalCards++
			}

			totalCards++
		}
	}
	fmt.Printf("   Created %d cards\n", totalCards)

	// ── 3b. Sprints (scrum projects) ─────────────────────────────────────────
	fmt.Println("→ Creating sprints for scrum projects…")

	type sprintSpec struct {
		name      string
		goal      string
		status    string // "planning" | "active" | "completed"
		startDays int
		endDays   int
		hasDate   bool
	}

	scrumSprints := map[string][]sprintSpec{
		"product-platform": {
			{
				name: "Sprint 1 — Discovery", status: "completed",
				goal:      "Validate the product concept, define architecture, and align the team.",
				startDays: -70, endDays: -56, hasDate: true,
			},
			{
				name: "Sprint 2 — Auth & Security", status: "completed",
				goal:      "Implement authentication, authorisation, and the admin dashboard.",
				startDays: -56, endDays: -42, hasDate: true,
			},
			{
				name: "Sprint 3 — Foundation", status: "completed",
				goal:      "Establish the core codebase, CI/CD pipeline, and authentication layer.",
				startDays: -28, endDays: -14, hasDate: true,
			},
			{
				name: "Sprint 4 — Core Features", status: "active",
				goal:      "Ship user dashboard, email notifications, search, and RBAC.",
				startDays: -14, endDays: 7, hasDate: true,
			},
			{
				name: "Sprint 5 — Polish & Launch", status: "planning",
				goal:      "Payment integration, analytics, performance, and public launch.",
				hasDate:   false,
			},
		},
		"api-platform": {
			{
				name: "Sprint 1 — Bootstrap", status: "completed",
				goal:      "Set up project structure, CI/CD, containerisation, and baseline endpoints.",
				startDays: -84, endDays: -70, hasDate: true,
			},
			{
				name: "Sprint 2 — Core API", status: "completed",
				goal:      "Deliver CRUD endpoints, pagination, request validation, and structured logging.",
				startDays: -70, endDays: -56, hasDate: true,
			},
			{
				name: "Sprint 3 — Auth & Docs", status: "completed",
				goal:      "Ship API key auth, rate limiting, Swagger UI, webhooks, and audit log.",
				startDays: -56, endDays: -42, hasDate: true,
			},
			{
				name: "Sprint 4 — SDKs & Testing", status: "completed",
				goal:      "Publish Python and JS/TS SDKs, integration tests, and performance benchmarks.",
				startDays: -42, endDays: -28, hasDate: true,
			},
			{
				name: "Sprint 5 — Stabilisation", status: "active",
				goal:      "Deprecate v0, fix known issues, land batch endpoints and cursor pagination.",
				startDays: -14, endDays: 7, hasDate: true,
			},
			{
				name: "Sprint 6 — GA Launch", status: "planning",
				goal:      "Developer portal, self-service key management, SLA monitoring, and launch announcement.",
				hasDate:   false,
			},
		},
	}

	// Build a lookup: project slug + card title → card ID
	type sprintCardRef struct {
		sprintName string
		cardID     uint
	}
	sprintCardRefs := map[string][]sprintCardRef{} // key: project slug

	for slug, specs := range projectCards {
		for _, spec := range specs {
			if spec.sprintName == "" {
				continue
			}
			key := slug + "/" + spec.title
			if c, ok := createdCards[key]; ok {
				sprintCardRefs[slug] = append(sprintCardRefs[slug], sprintCardRef{
					sprintName: spec.sprintName,
					cardID:     c.ID,
				})
			}
		}
	}

	totalSprints := 0
	for slug, sprints := range scrumSprints {
		pd := projects[slug]
		for _, ss := range sprints {
			sprint := &models.Sprint{
				ProjectID: pd.project.ID,
				Name:      ss.name,
				Goal:      ss.goal,
				Status:    ss.status,
			}
			if ss.hasDate {
				start := time.Now().UTC().AddDate(0, 0, ss.startDays).Truncate(24 * time.Hour)
				end := time.Now().UTC().AddDate(0, 0, ss.endDays).Truncate(24 * time.Hour)
				sprint.StartDate = &start
				sprint.EndDate = &end
			}
			must(db.Create(sprint).Error)

			// Link cards to this sprint
			for i, ref := range sprintCardRefs[slug] {
				if ref.sprintName != ss.name {
					continue
				}
				must(db.Create(&models.SprintCard{
					SprintID: sprint.ID,
					CardID:   ref.cardID,
					Position: float64((i + 1) * 1000),
				}).Error)
			}
			totalSprints++
		}
	}
	fmt.Printf("   Created %d sprints\n", totalSprints)

	// ── 3c. Card history (column moves — feeds the CFD chart) ─────────────────
	fmt.Println("→ Creating card history for CFD chart…")

	type histSpec struct {
		slug    string
		title   string
		fromCol string
		toCol   string
		daysAgo int
	}

	histSpecs := []histSpec{
		// ── Product Platform — Sprint 1 Discovery (created -70 to -66, closed -59 to -57) ──
		{"product-platform", "Product market research and user interviews", "To Do", "In Progress", 67},
		{"product-platform", "Product market research and user interviews", "In Progress", "In Review", 62},
		{"product-platform", "Product market research and user interviews", "In Review", "Done", 59},
		{"product-platform", "Competitor analysis and positioning", "To Do", "In Progress", 66},
		{"product-platform", "Competitor analysis and positioning", "In Progress", "In Review", 61},
		{"product-platform", "Competitor analysis and positioning", "In Review", "Done", 58},
		{"product-platform", "Technical feasibility study", "To Do", "In Progress", 65},
		{"product-platform", "Technical feasibility study", "In Progress", "In Review", 60},
		{"product-platform", "Technical feasibility study", "In Review", "Done", 58},
		{"product-platform", "Initial wireframes and UX sketches", "To Do", "In Progress", 64},
		{"product-platform", "Initial wireframes and UX sketches", "In Progress", "In Review", 59},
		{"product-platform", "Initial wireframes and UX sketches", "In Review", "Done", 57},
		{"product-platform", "Project charter and team kick-off", "To Do", "In Progress", 63},
		{"product-platform", "Project charter and team kick-off", "In Progress", "In Review", 58},
		{"product-platform", "Project charter and team kick-off", "In Review", "Done", 57},

		// ── Product Platform — Sprint 2 Auth & Security (created -56 to -52, closed -45 to -43) ──
		{"product-platform", "OAuth 2.0 provider integration", "To Do", "In Progress", 53},
		{"product-platform", "OAuth 2.0 provider integration", "In Progress", "In Review", 48},
		{"product-platform", "OAuth 2.0 provider integration", "In Review", "Done", 45},
		{"product-platform", "JWT token management and refresh", "To Do", "In Progress", 52},
		{"product-platform", "JWT token management and refresh", "In Progress", "In Review", 47},
		{"product-platform", "JWT token management and refresh", "In Review", "Done", 44},
		{"product-platform", "Bcrypt password hashing", "To Do", "In Progress", 51},
		{"product-platform", "Bcrypt password hashing", "In Progress", "In Review", 46},
		{"product-platform", "Bcrypt password hashing", "In Review", "Done", 44},
		{"product-platform", "Admin dashboard scaffolding", "To Do", "In Progress", 50},
		{"product-platform", "Admin dashboard scaffolding", "In Progress", "In Review", 45},
		{"product-platform", "Admin dashboard scaffolding", "In Review", "Done", 43},
		{"product-platform", "User invitation and onboarding flow", "To Do", "In Progress", 49},
		{"product-platform", "User invitation and onboarding flow", "In Progress", "In Review", 44},
		{"product-platform", "User invitation and onboarding flow", "In Review", "Done", 43},

		// ── Product Platform — Sprint 3 Foundation (created -28 to -24, closed -17 to -15) ──
		{"product-platform", "Set up monorepo and CI/CD pipeline", "To Do", "In Progress", 25},
		{"product-platform", "Set up monorepo and CI/CD pipeline", "In Progress", "In Review", 20},
		{"product-platform", "Set up monorepo and CI/CD pipeline", "In Review", "Done", 17},
		{"product-platform", "Design system and component library", "To Do", "In Progress", 24},
		{"product-platform", "Design system and component library", "In Progress", "In Review", 19},
		{"product-platform", "Design system and component library", "In Review", "Done", 16},
		{"product-platform", "User registration and login flow", "To Do", "In Progress", 23},
		{"product-platform", "User registration and login flow", "In Progress", "In Review", 18},
		{"product-platform", "User registration and login flow", "In Review", "Done", 15},
		{"product-platform", "Database schema v1 with migrations", "To Do", "In Progress", 22},
		{"product-platform", "Database schema v1 with migrations", "In Progress", "In Review", 17},
		{"product-platform", "Database schema v1 with migrations", "In Review", "Done", 16},
		{"product-platform", "Deployment to staging (Docker + nginx)", "To Do", "In Progress", 21},
		{"product-platform", "Deployment to staging (Docker + nginx)", "In Progress", "In Review", 16},
		{"product-platform", "Deployment to staging (Docker + nginx)", "In Review", "Done", 15},

		// ── Product Platform — Sprint 4 Core Features (active): Done and In Progress ──
		{"product-platform", "User dashboard overview screen", "To Do", "In Progress", 11},
		{"product-platform", "User dashboard overview screen", "In Progress", "In Review", 8},
		{"product-platform", "User dashboard overview screen", "In Review", "Done", 5},
		{"product-platform", "Role-based access control (RBAC)", "To Do", "In Progress", 12},
		{"product-platform", "Role-based access control (RBAC)", "In Progress", "In Review", 8},
		{"product-platform", "Role-based access control (RBAC)", "In Review", "Done", 6},
		{"product-platform", "Email notification service", "To Do", "In Progress", 10},
		{"product-platform", "Full-text search endpoint", "To Do", "In Progress", 9},
		{"product-platform", "API rate limiting middleware", "To Do", "In Progress", 5},

		// ── API Platform — Sprint 1 Bootstrap (created -84 to -80, closed -73 to -71) ──
		{"api-platform", "Design API schema and OpenAPI spec", "To Do", "In Progress", 81},
		{"api-platform", "Design API schema and OpenAPI spec", "In Progress", "In Review", 76},
		{"api-platform", "Design API schema and OpenAPI spec", "In Review", "Done", 73},
		{"api-platform", "Set up Go project structure with CI/CD", "To Do", "In Progress", 80},
		{"api-platform", "Set up Go project structure with CI/CD", "In Progress", "In Review", 75},
		{"api-platform", "Set up Go project structure with CI/CD", "In Review", "Done", 73},
		{"api-platform", "Implement health-check and metrics endpoints", "To Do", "In Progress", 79},
		{"api-platform", "Implement health-check and metrics endpoints", "In Progress", "In Review", 74},
		{"api-platform", "Implement health-check and metrics endpoints", "In Review", "Done", 72},
		{"api-platform", "Docker containerisation and registry setup", "To Do", "In Progress", 78},
		{"api-platform", "Docker containerisation and registry setup", "In Progress", "In Review", 73},
		{"api-platform", "Docker containerisation and registry setup", "In Review", "Done", 72},
		{"api-platform", "Error response standardisation (RFC 7807)", "To Do", "In Progress", 77},
		{"api-platform", "Error response standardisation (RFC 7807)", "In Progress", "In Review", 72},
		{"api-platform", "Error response standardisation (RFC 7807)", "In Review", "Done", 71},

		// ── API Platform — Sprint 2 Core API (created -70 to -66, closed -59 to -57) ──
		{"api-platform", "Implement resource CRUD endpoints", "To Do", "In Progress", 67},
		{"api-platform", "Implement resource CRUD endpoints", "In Progress", "In Review", 62},
		{"api-platform", "Implement resource CRUD endpoints", "In Review", "Done", 59},
		{"api-platform", "Pagination and filtering support", "To Do", "In Progress", 66},
		{"api-platform", "Pagination and filtering support", "In Progress", "In Review", 61},
		{"api-platform", "Pagination and filtering support", "In Review", "Done", 59},
		{"api-platform", "Request validation middleware", "To Do", "In Progress", 65},
		{"api-platform", "Request validation middleware", "In Progress", "In Review", 60},
		{"api-platform", "Request validation middleware", "In Review", "Done", 58},
		{"api-platform", "Database migrations with versioning", "To Do", "In Progress", 64},
		{"api-platform", "Database migrations with versioning", "In Progress", "In Review", 59},
		{"api-platform", "Database migrations with versioning", "In Review", "Done", 58},
		{"api-platform", "Structured logging with correlation IDs", "To Do", "In Progress", 63},
		{"api-platform", "Structured logging with correlation IDs", "In Progress", "In Review", 58},
		{"api-platform", "Structured logging with correlation IDs", "In Review", "Done", 57},

		// ── API Platform — Sprint 3 Auth & Docs (created -56 to -52, closed -45 to -43) ──
		{"api-platform", "API key authentication and rotation", "To Do", "In Progress", 53},
		{"api-platform", "API key authentication and rotation", "In Progress", "In Review", 48},
		{"api-platform", "API key authentication and rotation", "In Review", "Done", 45},
		{"api-platform", "Rate limiting per API key", "To Do", "In Progress", 52},
		{"api-platform", "Rate limiting per API key", "In Progress", "In Review", 47},
		{"api-platform", "Rate limiting per API key", "In Review", "Done", 44},
		{"api-platform", "Interactive API documentation (Swagger UI)", "To Do", "In Progress", 51},
		{"api-platform", "Interactive API documentation (Swagger UI)", "In Progress", "In Review", 46},
		{"api-platform", "Interactive API documentation (Swagger UI)", "In Review", "Done", 44},
		{"api-platform", "Webhook delivery with retry logic", "To Do", "In Progress", 50},
		{"api-platform", "Webhook delivery with retry logic", "In Progress", "In Review", 45},
		{"api-platform", "Webhook delivery with retry logic", "In Review", "Done", 43},
		{"api-platform", "Audit log for all API mutations", "To Do", "In Progress", 49},
		{"api-platform", "Audit log for all API mutations", "In Progress", "In Review", 44},
		{"api-platform", "Audit log for all API mutations", "In Review", "Done", 43},

		// ── API Platform — Sprint 4 SDKs & Testing (created -42 to -38, closed -31 to -29) ──
		{"api-platform", "Python SDK for the public API", "To Do", "In Progress", 39},
		{"api-platform", "Python SDK for the public API", "In Progress", "In Review", 34},
		{"api-platform", "Python SDK for the public API", "In Review", "Done", 31},
		{"api-platform", "JavaScript / TypeScript SDK", "To Do", "In Progress", 38},
		{"api-platform", "JavaScript / TypeScript SDK", "In Progress", "In Review", 33},
		{"api-platform", "JavaScript / TypeScript SDK", "In Review", "Done", 30},
		{"api-platform", "End-to-end integration test suite", "To Do", "In Progress", 37},
		{"api-platform", "End-to-end integration test suite", "In Progress", "In Review", 32},
		{"api-platform", "End-to-end integration test suite", "In Review", "Done", 30},
		{"api-platform", "Load testing and performance benchmarks", "To Do", "In Progress", 36},
		{"api-platform", "Load testing and performance benchmarks", "In Progress", "In Review", 31},
		{"api-platform", "Load testing and performance benchmarks", "In Review", "Done", 29},
		{"api-platform", "API changelog and versioning policy", "To Do", "In Progress", 35},
		{"api-platform", "API changelog and versioning policy", "In Progress", "In Review", 30},
		{"api-platform", "API changelog and versioning policy", "In Review", "Done", 29},

		// ── API Platform — Sprint 5 Stabilisation (active): Done and In Progress ──
		{"api-platform", "Deprecate v0 endpoints and notify users", "To Do", "In Progress", 12},
		{"api-platform", "Deprecate v0 endpoints and notify users", "In Progress", "In Review", 11},
		{"api-platform", "Deprecate v0 endpoints and notify users", "In Review", "Done", 9},
		{"api-platform", "Fix response envelope inconsistency bug", "To Do", "In Progress", 12},
		{"api-platform", "Fix response envelope inconsistency bug", "In Progress", "Done", 10},
		{"api-platform", "Add batch API endpoints for bulk operations", "To Do", "In Progress", 10},
		{"api-platform", "Add batch API endpoints for bulk operations", "In Progress", "In Review", 7},
		{"api-platform", "Improve error messages with actionable hints", "To Do", "In Progress", 8},
		{"api-platform", "Support cursor-based pagination", "To Do", "In Progress", 7},
	}

	totalHistory := 0
	for _, hs := range histSpecs {
		cardKey := hs.slug + "/" + hs.title
		card, ok := createdCards[cardKey]
		if !ok {
			fmt.Printf("   ⚠ skipping history for %q (card not found)\n", cardKey)
			continue
		}
		pd := projects[hs.slug]
		fromCol := pd.cols[hs.fromCol]
		toCol := pd.cols[hs.toCol]
		if fromCol == nil || toCol == nil {
			fmt.Printf("   ⚠ skipping history move %s→%s for %q (column not found)\n", hs.fromCol, hs.toCol, cardKey)
			continue
		}
		moveTime := time.Now().UTC().AddDate(0, 0, -hs.daysAgo)
		hist := &models.CardHistory{
			CardID:       card.ID,
			UserID:       users["admin"].ID,
			EventType:    "column_move",
			FromColumnID: fromCol.ID,
			ToColumnID:   toCol.ID,
		}
		must(db.Create(hist).Error)
		must(db.Model(hist).UpdateColumn("created_at", moveTime).Error)
		totalHistory++
	}
	fmt.Printf("   Created %d card history records\n", totalHistory)

	// ── 3d. Releases ──────────────────────────────────────────────────────────
	fmt.Println("→ Creating releases…")

	type releaseSpec struct {
		project    string
		name       string
		goal       string
		targetDays int
		sprints    []string
	}

	releaseSpecs := []releaseSpec{
		{
			project:    "product-platform",
			name:       "v1.0 — Platform Launch",
			goal:       "Ship the complete SaaS platform — from auth through payments — to production.",
			targetDays: 14,
			sprints: []string{
				"Sprint 1 — Discovery",
				"Sprint 2 — Auth & Security",
				"Sprint 3 — Foundation",
				"Sprint 4 — Core Features",
				"Sprint 5 — Polish & Launch",
			},
		},
		{
			project:    "api-platform",
			name:       "v1.0 — Public API",
			goal:       "Release the stable, documented public API with SDKs to developers.",
			targetDays: 15,
			sprints: []string{
				"Sprint 1 — Bootstrap",
				"Sprint 2 — Core API",
				"Sprint 3 — Auth & Docs",
				"Sprint 4 — SDKs & Testing",
				"Sprint 5 — Stabilisation",
			},
		},
		{
			project:    "api-platform",
			name:       "v1.1 — GA Launch",
			goal:       "Developer portal, self-service key management, and general availability announcement.",
			targetDays: 45,
			sprints:    []string{"Sprint 6 — GA Launch"},
		},
	}

	totalReleases := 0
	for _, rs := range releaseSpecs {
		pd := projects[rs.project]
		targetDate := time.Now().UTC().AddDate(0, 0, rs.targetDays).Truncate(24 * time.Hour)
		release := &models.Release{
			ProjectID:  pd.project.ID,
			Name:       rs.name,
			Goal:       rs.goal,
			TargetDate: &targetDate,
		}
		must(db.Create(release).Error)
		for _, sprintName := range rs.sprints {
			var sprint models.Sprint
			if err := db.Where("project_id = ? AND name = ?", pd.project.ID, sprintName).First(&sprint).Error; err == nil {
				must(db.Create(&models.ReleaseSprint{ReleaseID: release.ID, SprintID: sprint.ID}).Error)
			}
		}
		totalReleases++
	}
	fmt.Printf("   Created %d releases\n", totalReleases)

	// ── 3e. Card cross-references ─────────────────────────────────────────────
	fmt.Println("→ Creating card cross-references…")

	cardLinks := [][2]string{
		// Same feature implemented on two platforms
		{"website-redesign/Implement dark mode toggle", "mobile-app-v2/Dark mode support"},
		// Both contribute to Lighthouse score / site quality
		{"website-redesign/Accessibility audit and ARIA fixes", "website-redesign/Optimise image loading with lazy + WebP"},
		// Auth tests cover the profile screen flow
		{"mobile-app-v2/User profile screen", "mobile-app-v2/Integration tests for authentication flow"},
		// K8s cluster is a prerequisite for the DB migration
		{"devops-infra/Set up Kubernetes cluster on cloud provider", "devops-infra/Migrate primary database to PostgreSQL"},
		// Monitoring stack runs on the new K8s cluster
		{"devops-infra/Add Prometheus + Grafana monitoring stack", "devops-infra/Set up Kubernetes cluster on cloud provider"},
		// Backups depend on the completed DB migration
		{"devops-infra/Automate database backups with off-site retention", "devops-infra/Migrate primary database to PostgreSQL"},
	}

	totalRefs := 0
	for _, pair := range cardLinks {
		src, ok1 := createdCards[pair[0]]
		tgt, ok2 := createdCards[pair[1]]
		if !ok1 || !ok2 {
			fmt.Printf("   ⚠ skipping ref %q ↔ %q (card not found)\n", pair[0], pair[1])
			continue
		}
		must(db.Create(&models.CardReference{SourceCardID: src.ID, TargetCardID: tgt.ID}).Error)
		totalRefs++
	}
	fmt.Printf("   Created %d card cross-references\n", totalRefs)

	// ── 4. Topics ─────────────────────────────────────────────────────────────
	fmt.Println("→ Creating topics…")

	type topicSpec struct {
		project string
		author  string
		title   string
		body    string
		pinned  bool
		replies []struct{ author, body string }
	}

	topicSpecs := []topicSpec{
		{
			project: "website-redesign",
			author:  "admin",
			title:   "Q4 Design Direction",
			pinned:  true,
			body: `Hi everyone 👋

After last week's brand workshop I want to summarise the three pillars we agreed on:

1. **Clarity over cleverness** — every page should answer the visitor's question in under 5 seconds.
2. **Performance as a feature** — target a Lighthouse score ≥ 90 on all pages.
3. **Accessible by default** — WCAG AA minimum, aiming for AAA on primary flows.

Designs go into Figma first; no dev work starts without a reviewed mockup.
Questions or push-back? Reply here!`,
			replies: []struct{ author, body string }{
				{"sarah", "Fully on board with the performance target. I'll set up Lighthouse CI so every PR gets a score automatically."},
				{"marc", "On the accessibility point — should we bring in an external auditor mid-project or rely on our own review?"},
				{"admin", "@demo.marc Good question. Let's do an internal pass first, then budget for one external review before launch."},
			},
		},
		{
			project: "mobile-app-v2",
			author:  "sarah",
			title:   "API integration strategy — REST vs GraphQL",
			body: `We need to decide how the mobile app talks to the backend before we go deeper into development.

**Option A — REST (current approach)**
- Proven, simple, easy to cache
- We already have the endpoints; just needs mobile-friendly pagination

**Option B — GraphQL**
- Fetch exactly what you need (big win on mobile bandwidth)
- Requires a new gateway layer; non-trivial migration

My preference is to stick with REST for v2 and add a thin BFF (Backend for Frontend) layer to shape responses for mobile. Thoughts?`,
			replies: []struct{ author, body string }{
				{"marc", "I'd vote REST + BFF too. GraphQL would be ideal but the migration cost isn't justified for v2."},
				{"lisa", "Agree. We can always add a GraphQL layer for v3 once we know our real query patterns from production data."},
				{"sarah", "Settled then — REST + BFF it is. I'll open a card for the BFF scaffolding."},
			},
		},
		{
			project: "devops-infra",
			author:  "marc",
			title:   "PostgreSQL migration — go/no-go checklist",
			pinned:  true,
			body: `Before we cut over to Postgres in production, everyone needs to sign off on this list:

- [ ] Schema migration tested on a production-size data copy
- [ ] All queries verified for Postgres compatibility (no SQLite-isms)
- [ ] Read-replica in place for reporting queries
- [ ] Rollback procedure documented and rehearsed
- [ ] Monitoring dashboards updated for Postgres metrics
- [ ] On-call runbook updated

I'll move us to "go" once all boxes are checked. Please comment here with your sign-off.`,
			replies: []struct{ author, body string }{
				{"lisa", "Monitoring dashboards are live in Grafana. Signing off on that item ✓"},
				{"admin", "Rollback procedure is documented in the wiki and tested in staging. ✓"},
				{"marc", "Schema migration tested — 4.2 M rows, completed in 8 minutes. Well within our maintenance window. ✓"},
			},
		},
		// ── Additional website-redesign topics ────────────────────────────────
		{
			project: "website-redesign",
			author:  "sarah",
			title:   "Font and typography choices",
			body: `I've been evaluating typefaces for the redesign. Here are my top three candidates:

1. **Inter** — clean, very legible at small sizes, used widely in SaaS products. Free via Google Fonts.
2. **Geist** — Vercel's typeface, feels modern and technical. Good for a developer-adjacent audience.
3. **DM Sans** — friendly and approachable, good contrast with a geometric headline font.

I'm leaning toward **Inter for body** and **Sora for headings** — the contrast between them works really well in the mockups.

Thoughts? Any strong opinions before I finalise the Figma components?`,
			replies: []struct{ author, body string }{
				{"marc", "Inter is a safe and excellent choice — I've used it on three projects and it always looks sharp. Sora pairs nicely."},
				{"admin", "Agreed on Inter. Whatever we pick, make sure it's self-hosted or loaded with `font-display: swap` so it doesn't block rendering."},
				{"sarah", "Good point — I'll self-host both via Fontsource so we have full control over loading strategy. Will update the Figma file this week."},
			},
		},
		{
			project: "website-redesign",
			author:  "marc",
			title:   "Cookie consent and GDPR compliance",
			body: `Legal asked us to review our cookie consent implementation before launch. A few things we need to address:

**Current state**
- We drop a _ga analytics cookie on page load with no prior consent — this is non-compliant in the EU.
- No consent management platform (CMP) is in place.

**Proposed approach**
1. Add a lightweight CMP (I'm looking at **Klaro** — open source, no SaaS fees).
2. Gate analytics and marketing scripts behind consent categories.
3. Add a "Cookie settings" link in the footer.
4. Store consent choice in localStorage so the banner only appears once.

This needs to be done before we go live. I'll open a card for it.`,
			replies: []struct{ author, body string }{
				{"admin", "Klaro looks good — I've seen it used well before. Make sure the default state for analytics is **opt-out**, not opt-in, to be safe."},
				{"sarah", "Agreed. Also worth checking if we need a cookie policy page — some regulators want a dedicated URL linked from the banner."},
				{"marc", "I'll draft both the implementation card and a short cookie policy page. Will share for review before merging."},
			},
		},
		// ── Additional mobile-app-v2 topics ───────────────────────────────────
		{
			project: "mobile-app-v2",
			author:  "lisa",
			title:   "Testing strategy for v2",
			body: `Now that the main features are taking shape, we should agree on a testing strategy before we hit Testing column on the board.

**What I'm proposing:**

| Layer | Tool | Owner |
|---|---|---|
| Unit tests | Jest + React Native Testing Library | All devs |
| Integration | Detox (E2E on simulator) | Lisa |
| Manual exploratory | Checklist per feature | Rotating |
| Performance | Flashlight (Android) + Instruments (iOS) | Sarah |

**Devices to cover as a minimum:**
- iPhone SE (small screen)
- iPhone 15 (latest)
- Pixel 7 (Android 13)
- Samsung Galaxy A54 (mid-range Android)

Does this look reasonable? Anything missing?`,
			replies: []struct{ author, body string }{
				{"sarah", "Looks solid. I'd add a tablet pass (iPad mini at minimum) since our analytics show ~12% of users are on tablets."},
				{"marc", "Detox can be flaky on CI — worth adding retry logic in the pipeline. Happy to set that up."},
				{"lisa", "Good points both. I'll update the strategy doc and open a board card for the CI Detox config."},
			},
		},
		{
			project: "mobile-app-v2",
			author:  "marc",
			title:   "App store submission checklist",
			pinned:  true,
			body: `Tracking what we need before we can submit to the App Store and Google Play.

**App Store (iOS)**
- [ ] App icon (all required sizes via Xcode asset catalogue)
- [ ] Screenshots for 6.7" and 5.5" displays
- [ ] Privacy policy URL
- [ ] Age rating questionnaire filled in
- [ ] In-app purchase declarations (none for v1 — confirm)
- [ ] TestFlight beta round completed

**Google Play**
- [ ] Feature graphic (1024 × 500)
- [ ] Screenshots for phone and 7" tablet
- [ ] Privacy policy URL
- [ ] Data safety form completed
- [ ] Internal testing track approved before production rollout

Tag me here when your section is done.`,
			replies: []struct{ author, body string }{
				{"sarah", "Privacy policy is drafted — waiting for legal sign-off. Should have it by end of week."},
				{"lisa", "TestFlight build is up. Three external testers invited. Feedback so far: login flow is smooth, settings screen needs larger tap targets."},
				{"marc", "Noted on tap targets — I'll fix that before the next build. Good progress everyone 🚀"},
			},
		},
		// ── Additional devops-infra topics ────────────────────────────────────
		{
			project: "devops-infra",
			author:  "lisa",
			title:   "Incident review — staging outage 2026-03-21",
			body: `**Summary**
Staging was down for 47 minutes on 2026-03-21 between 14:12 and 14:59 UTC due to a misconfigured Nginx upstream after the load balancer update.

**Timeline**
- 14:12 — deploy of Nginx config v2.4 triggered automatically via CI
- 14:15 — first alerts fired (5xx rate > 5%)
- 14:22 — Marc acknowledged alert and began investigation
- 14:51 — root cause identified: upstream block pointed to old container name
- 14:59 — config corrected and reloaded, traffic restored

**Root cause**
The container name changed as part of the Docker Compose refactor but the Nginx template was not updated.

**Action items**
- [ ] Add integration test that validates Nginx upstream names match running containers
- [ ] Add staging smoke test to CI pipeline (runs after every deploy)
- [ ] Update runbook with "check upstream names" as first step in 5xx incidents`,
			replies: []struct{ author, body string }{
				{"marc", "I've opened cards for the integration test and smoke test. Both are in the Todo column."},
				{"admin", "Good write-up. Let's also add a 5-minute grace period to the alert so we don't page on brief deploy blips — 47 minutes is a real incident, a 30-second blip during a rolling restart is not."},
				{"lisa", "Agreed. I'll update the alert threshold in Grafana. Will post the updated config here for review."},
			},
		},
		{
			project: "devops-infra",
			author:  "admin",
			title:   "On-call rota — Q2 2026",
			body: `Setting up the on-call schedule for Q2. We'll use a weekly rotation.

| Week | Primary | Secondary |
|---|---|---|
| Apr 1–7 | Marc | Lisa |
| Apr 8–14 | Lisa | Alex |
| Apr 15–21 | Alex | Marc |
| Apr 22–28 | Marc | Lisa |
| May 1–7 | Lisa | Alex |
| May 8–14 | Alex | Marc |

**Expectations**
- Primary is first responder; target acknowledgement within 15 minutes during business hours, 30 minutes outside.
- Secondary is backup if primary is unreachable.
- Swap requests: post here at least 48 hours in advance and confirm with the person covering for you.

Pagerduty schedules will be updated to match this by Friday.`,
			replies: []struct{ author, body string }{
				{"marc", "Works for me. Can I swap Apr 22–28 with Lisa? I have a conference that week."},
				{"lisa", "Fine by me — I'll take Apr 22–28 primary, Marc takes May 1–7 primary. Alex stays as secondary both weeks."},
				{"admin", "Updated the table and will sync Pagerduty. Thanks for coordinating quickly."},
			},
		},
	}

	totalTopics := 0
	for _, ts := range topicSpecs {
		pd := projects[ts.project]
		topic := &models.Topic{
			ProjectID: pd.project.ID,
			UserID:    users[ts.author].ID,
			Title:     ts.title,
			Body:      ts.body,
			IsPinned:  ts.pinned,
		}
		must(db.Create(topic).Error)

		for _, r := range ts.replies {
			author := users[r.author]
			if author == nil {
				author = users["admin"]
			}
			must(db.Create(&models.TopicReply{
				TopicID: topic.ID, UserID: author.ID, Body: r.body,
			}).Error)
		}
		totalTopics++
	}
	fmt.Printf("   Created %d topics\n", totalTopics)

	// ── 5. Conversations (DMs + group chat) ──────────────────────────────────
	fmt.Println("→ Creating conversations…")

	type msgSpec struct {
		author string
		body   string
		ago    time.Duration // how long ago the message was sent
	}
	type convSpec struct {
		members  []string // user keys; first member is "created by"
		isGroup  bool
		name     string // only for group chats
		messages []msgSpec
	}

	now := time.Now()

	convSpecs := []convSpec{
		// 1-on-1: Alex ↔ Sarah
		{
			members: []string{"admin", "sarah"},
			messages: []msgSpec{
				{"admin", "Hey Sarah, quick question — are you planning to push the dark mode PR today or should I move the card to Review?", 72 * time.Hour},
				{"sarah", "Hey! Yes, I'll open the PR this afternoon once I've sorted the Safari flicker bug.", 71*time.Hour + 45*time.Minute},
				{"admin", "Perfect, no rush. Let me know if you need a second pair of eyes on the CSS.", 71*time.Hour + 30*time.Minute},
				{"sarah", "Will do. Also — did you see Marc's comment about `prefers-color-scheme`? Good catch from him.", 71*time.Hour + 10*time.Minute},
				{"admin", "Yeah, I replied there. Totally agree, saves a flash on first load. Classic progressive enhancement.", 71 * time.Hour},
				{"sarah", "PR is up! Assigned you as reviewer 🎉", 48 * time.Hour},
				{"admin", "Reviewed — two minor nits but nothing blocking. LGTM overall 👍", 47*time.Hour + 30*time.Minute},
				{"sarah", "Thanks, fixed both. Merging now.", 47 * time.Hour},
			},
		},
		// 1-on-1: Marc ↔ Lisa
		{
			members: []string{"marc", "lisa"},
			messages: []msgSpec{
				{"lisa", "Marc, I spotted something weird in the Postgres migration script — it's not handling NULL values in the `description` column correctly.", 36 * time.Hour},
				{"marc", "Oh no — can you paste the exact row that's failing?", 35*time.Hour + 50*time.Minute},
				{"lisa", "It's any card where description was never set. The column is NOT NULL in the Postgres schema but nullable in SQLite, so rows come over as empty string and then the constraint fires.", 35*time.Hour + 40*time.Minute},
				{"marc", "Good catch. I'll add a COALESCE in the export query to default to '' and change the Postgres column to allow NULL. Give me 20 min.", 35*time.Hour + 30*time.Minute},
				{"lisa", "Sounds good. I'll keep an eye on the slow-query log in the meantime.", 35*time.Hour + 25*time.Minute},
				{"marc", "Fixed! New script is in the `migration/` branch. Can you run a test against the staging copy?", 35 * time.Hour},
				{"lisa", "Running now… ✅ All 4.2 M rows migrated cleanly. Zero errors.", 34*time.Hour + 30*time.Minute},
				{"marc", "Legend. I'll update the checklist topic and schedule the prod window for Saturday at 02:00.", 34*time.Hour + 15*time.Minute},
			},
		},
		// 1-on-1: Sarah ↔ Lisa
		{
			members: []string{"sarah", "lisa"},
			messages: []msgSpec{
				{"lisa", "Sarah, do you have five minutes for a quick call? The Android dark mode is looking off on the Pixel 7 emulator.", 24 * time.Hour},
				{"sarah", "Sure! Give me 10 min — in a standup right now.", 23*time.Hour + 55*time.Minute},
				{"sarah", "Ready when you are.", 23*time.Hour + 45*time.Minute},
				{"lisa", "I think it's the `StatusBar` style not switching — it stays light even when the theme flips to dark.", 23*time.Hour + 30*time.Minute},
				{"sarah", "Ah yes! You need to call `StatusBar.setStyle(Style.Dark)` explicitly inside the `ionViewWillEnter` hook — the automatic detection doesn't work on Capacitor 5.", 23*time.Hour + 20*time.Minute},
				{"lisa", "That's exactly it! Works perfectly now. Thanks for the quick fix 🙌", 23*time.Hour + 10*time.Minute},
				{"sarah", "No problem. I'll add a note to the mobile dev guide so the next person doesn't hit the same thing.", 23 * time.Hour},
			},
		},
		// 1-on-1: Admin ↔ Marc
		{
			members: []string{"admin", "marc"},
			messages: []msgSpec{
				{"admin", "Marc — heads up, the quarterly security audit is due next week. Have you started the vulnerability scan?", 5 * 24 * time.Hour},
				{"marc", "Not yet, I've been deep in the Postgres migration. Can I start it Thursday?", 5*24*time.Hour - 30*time.Minute},
				{"admin", "Thursday is fine, just make sure it's done before Friday EOD. Legal needs the report by Monday.", 5*24*time.Hour - 1*time.Hour},
				{"marc", "Understood. I'll use the same toolchain as last quarter — nmap + OWASP ZAP + a manual headers review.", 5*24*time.Hour - 2*time.Hour},
				{"admin", "Perfect. Ping me if you find anything critical.", 5*24*time.Hour - 2*time.Hour - 15*time.Minute},
				{"marc", "Scan complete — no critical findings, two mediums. Both are missing security headers (X-Frame-Options and Referrer-Policy).", 2 * 24 * time.Hour},
				{"admin", "Easy fixes. Can you open cards for both and assign to yourself?", 2*24*time.Hour - 10*time.Minute},
				{"marc", "Done. Cards INF-22 and INF-23. Should have patches deployed by tomorrow.", 2*24*time.Hour - 20*time.Minute},
			},
		},
		// 1-on-1: Tonk ↔ Priya (friendly)
		{
			members: []string{"tonk", "priya"},
			messages: []msgSpec{
				{"tonk", "Morning Priya — just wanted to say your demo yesterday was brilliant. The new onboarding flow is so much clearer now.", 26 * time.Hour},
				{"priya", "Thanks Ton! Really appreciate that — the team had great feedback. I cleaned up the edge-case on the invite link you mentioned.", 25*time.Hour + 40*time.Minute},
				// Priya shares a small Go helper
				{"priya", "FYI — here's the invite expiry helper I used:\n```go\nfunc InviteExpired(expiry time.Time) bool {\n\treturn time.Now().After(expiry)\n}\n```", 25*time.Hour + 30*time.Minute},
				// Tonk responds with a follow-up snippet
				{"tonk", "Nice helper — consider this variant that notifies when expired:\n```go\nfunc NotifyIfExpired(expiry time.Time, userID int) {\n\tif InviteExpired(expiry) {\n\t\t// send notification to userID\n\t}\n}\n```", 25*time.Hour + 10*time.Minute},
				// Priya shares repo link for preview
				{"priya", "Also — here's the repo if you want to inspect the code: https://github.com/tonk/warmdesk.git", 25*time.Hour + 5*time.Minute},
				{"priya", "Perfect, thanks. Also — if you have a minute later, could you glance at the accessibility labels? Want to make sure we're shipping inclusive defaults.", 25 * time.Hour},
				{"tonk", "Absolutely. I'll take a look after lunch. Great work on this, Priya — really great to see this land. 🙌", 24*time.Hour + 50*time.Minute},
			},
		},

		// Group chat: Website Redesign team
		{
			members: []string{"admin", "sarah", "marc"},
			isGroup: true,
			name:    "Website Redesign Team",
			messages: []msgSpec{
				{"admin", "Morning everyone! Quick sync on the redesign — where are we blocking?", 3 * 24 * time.Hour},
				{"sarah", "The image optimisation PR is in review and should be merged today. LCP is looking great 🚀", 3*24*time.Hour - 15*time.Minute},
				{"marc", "I'm finishing up the mobile nav fix. Should be done by EOD.", 3*24*time.Hour - 20*time.Minute},
				{"admin", "Great. After those land, the only thing blocking staging deploy is the accessibility audit. Marc, are you planning to pick that up?", 3*24*time.Hour - 30*time.Minute},
				{"marc", "Yes, I've blocked out Thursday afternoon for it.", 3*24*time.Hour - 35*time.Minute},
				{"sarah", "I can pair on the ARIA side if you want — I've done a few of these before.", 3*24*time.Hour - 40*time.Minute},
				{"marc", "That would be really helpful actually, thanks Sarah 🙏", 3*24*time.Hour - 45*time.Minute},
				{"admin", "Awesome. I'll push the staging deploy for Friday then. Any blockers I should know about?", 2 * 24 * time.Hour},
				{"sarah", "None from my side. Image PR just got merged 🎉", 2*24*time.Hour - 5*time.Minute},
				{"marc", "All clear. Nav fix is deployed to staging already.", 2*24*time.Hour - 10*time.Minute},
				{"admin", "Perfect. Friday deploy is on. I'll send a calendar invite for the staging review.", 2*24*time.Hour - 15*time.Minute},
			},
		},
	}

	totalConvs := 0
	totalConvMsgs := 0
	for _, cs := range convSpecs {
		conv := &models.Conversation{
			Name:        cs.name,
			IsGroup:     cs.isGroup,
			CreatedByID: users[cs.members[0]].ID,
		}
		must(db.Create(conv).Error)

		for _, key := range cs.members {
			must(db.Create(&models.ConversationMember{
				ConversationID: conv.ID,
				UserID:         users[key].ID,
				JoinedAt:       now.Add(-7 * 24 * time.Hour),
			}).Error)
		}

		for _, ms := range cs.messages {
			sentAt := now.Add(-ms.ago)
			msg := &models.ConversationMessage{
				ConversationID: conv.ID,
				SenderID:       users[ms.author].ID,
				Body:           ms.body,
			}
			must(db.Create(msg).Error)
			// Set realistic created_at timestamps
			must(db.Model(msg).Updates(map[string]interface{}{
				"created_at": sentAt,
				"updated_at": sentAt,
			}).Error)
			totalConvMsgs++
		}

		// Bump updated_at to the last message time so conversations sort correctly
		if len(cs.messages) > 0 {
			lastMsgTime := now.Add(-cs.messages[len(cs.messages)-1].ago)
			must(db.Model(conv).Update("updated_at", lastMsgTime).Error)
		}

		totalConvs++
	}
	fmt.Printf("   Created %d conversations (%d messages)\n", totalConvs, totalConvMsgs)

	// ── 6. Customers & Contracts ──────────────────────────────────────────────
	fmt.Println("→ Creating customers and contracts…")

	type timeSlotSpec struct {
		label        string
		from         string
		to           string
		days         string
		factor       float64
		endDayOffset int
	}
	type contractSpec struct {
		name         string
		desc         string
		startDays    int // relative to today
		endDays      int // 0 = no end date
		projects     []string
		pricePerHour *float64
		timeSlots    []timeSlotSpec
	}
	type customerSpec struct {
		name      string
		desc      string
		logoURL   string
		contracts []contractSpec
		projects  []string // unassigned projects (no contract)
	}

	acmePricePerHour := 110.0
	acmeStandbySlots := []timeSlotSpec{
		{label: "Standby - week", from: "19:00", to: "07:00", days: "weekdays", factor: 1.5, endDayOffset: 1},
		{label: "Standby - weekend", from: "19:00", to: "07:00", days: "friday", factor: 2.0, endDayOffset: 3},
	}

	demoCustomers := []customerSpec{
		{
			name:    "Acme Corporation",
			desc:    "Long-running client delivering internal tooling and web platforms.",
			logoURL: "https://api.dicebear.com/9.x/shapes/svg?seed=Acme-Corp&backgroundColor=ecfeff,bfdbfe,e9d5ff",
			contracts: []contractSpec{
				{
					name:         "Phase 1 — Marketing Site",
					desc:         "Full redesign and relaunch of the corporate marketing website.",
					startDays:    -180,
					endDays:      90,
					projects:     []string{"website-redesign"},
					pricePerHour: &acmePricePerHour,
					timeSlots:    acmeStandbySlots,
				},
				{
					name:         "Phase 2 — Mobile Apps",
					desc:         "iOS and Android companion apps for the new platform.",
					startDays:    -60,
					endDays:      180,
					projects:     []string{"mobile-app-v2"},
					pricePerHour: &acmePricePerHour,
					timeSlots:    acmeStandbySlots,
				},
			},
		},
		{
			name:    "Globex Systems",
			desc:    "Infrastructure-heavy client focused on cloud reliability and security.",
			logoURL: "https://api.dicebear.com/9.x/shapes/svg?seed=Globex-Systems&backgroundColor=fef3c7,fde68a,fee2e2",
			contracts: []contractSpec{
				{
					name:      "Managed DevOps 2025",
					desc:      "Kubernetes migration, monitoring setup, and on-call engineering support.",
					startDays: -90,
					endDays:   275,
					projects:  []string{"devops-infra"},
				},
			},
		},
		{
			name:     "Initech Ltd",
			desc:     "Prospective client — evaluation phase, no active contract yet.",
			logoURL:  "https://api.dicebear.com/9.x/shapes/svg?seed=Initech-Ltd&backgroundColor=d1fae5,ddd6fe,fce7f3",
			projects: []string{}, // no projects yet
		},
	}

	var demoCustomerIDs []uint
	for _, cs := range demoCustomers {
		cust := &models.Customer{Name: cs.name, Description: cs.desc, LogoURL: cs.logoURL}
		must(db.Create(cust).Error)
		demoCustomerIDs = append(demoCustomerIDs, cust.ID)

		// Star customers for relevant users so the sidebar shows favourites
		switch cs.name {
		case "Acme Corporation":
			// Tonk, Admin, Sarah (website owner), and Marc (mobile owner) work on Acme contracts
			must(db.Create(&models.CustomerFavorite{UserID: tonk.ID, CustomerID: cust.ID}).Error)
			must(db.Create(&models.CustomerFavorite{UserID: users["admin"].ID, CustomerID: cust.ID}).Error)
			must(db.Create(&models.CustomerFavorite{UserID: users["sarah"].ID, CustomerID: cust.ID}).Error)
			must(db.Create(&models.CustomerFavorite{UserID: users["marc"].ID, CustomerID: cust.ID}).Error)
		case "Globex Systems":
			// Tonk, Marc and Lisa own the DevOps contract for Globex
			must(db.Create(&models.CustomerFavorite{UserID: tonk.ID, CustomerID: cust.ID}).Error)
			must(db.Create(&models.CustomerFavorite{UserID: users["marc"].ID, CustomerID: cust.ID}).Error)
			must(db.Create(&models.CustomerFavorite{UserID: users["lisa"].ID, CustomerID: cust.ID}).Error)
		}

		for _, conSpec := range cs.contracts {
			start := time.Now().UTC().AddDate(0, 0, conSpec.startDays).Truncate(24 * time.Hour)
			con := &models.Contract{
				CustomerID:   cust.ID,
				Name:         conSpec.name,
				Description:  conSpec.desc,
				StartDate:    &start,
				PricePerHour: conSpec.pricePerHour,
			}
			if conSpec.endDays != 0 {
				end := time.Now().UTC().AddDate(0, 0, conSpec.endDays).Truncate(24 * time.Hour)
				con.EndDate = &end
			}
			must(db.Create(con).Error)

			for _, slotSpec := range conSpec.timeSlots {
				factor := slotSpec.factor
				must(db.Create(&models.ContractTimeSlot{
					ContractID:           con.ID,
					Label:                slotSpec.label,
					StartTime:            slotSpec.from,
					EndTime:              slotSpec.to,
					DayType:              slotSpec.days,
					EndDayOffset:         slotSpec.endDayOffset,
					MultiplicationFactor: &factor,
				}).Error)
			}

			for _, slug := range conSpec.projects {
				must(db.Model(&models.Project{}).Where("slug = ?", slug).
					Updates(map[string]interface{}{"customer_id": cust.ID, "contract_id": con.ID}).Error)
			}
		}

		for _, slug := range cs.projects {
			must(db.Model(&models.Project{}).Where("slug = ?", slug).
				Updates(map[string]interface{}{"customer_id": cust.ID}).Error)
		}
	}
	fmt.Printf("   Created %d customers with contracts\n", len(demoCustomers))

	// ── 7. SLA Policies ───────────────────────────────────────────────────────
	fmt.Println("→ Creating SLA policies…")

	slaPolicies := []models.SlaPolicy{
		{Name: "Critical — 1h response / 4h resolution", ResponseTimeMinutes: 60, ResolutionTimeMinutes: 240, PriorityFilter: "critical", IsActive: true},
		{Name: "High — 2h response / 8h resolution", ResponseTimeMinutes: 120, ResolutionTimeMinutes: 480, PriorityFilter: "high", IsActive: true},
		{Name: "Standard — 4h response / 24h resolution", ResponseTimeMinutes: 240, ResolutionTimeMinutes: 1440, PriorityFilter: "medium,low", IsActive: true},
	}
	for i := range slaPolicies {
		must(db.Create(&slaPolicies[i]).Error)
	}
	fmt.Printf("   Created %d SLA policies\n", len(slaPolicies))

	// Build a priority → SLA policy lookup for assigning SLA to tickets
	slaByPriority := map[string]*models.SlaPolicy{}
	for i := range slaPolicies {
		if slaPolicies[i].PriorityFilter == "" {
			// Match-all policy — store for every priority (lowest wins via existing insertion)
			for _, p := range []string{"low", "medium", "high", "critical"} {
				if _, ok := slaByPriority[p]; !ok {
					slaByPriority[p] = &slaPolicies[i]
				}
			}
		} else {
			for _, p := range strings.Split(slaPolicies[i].PriorityFilter, ",") {
				pri := strings.TrimSpace(p)
				if _, ok := slaByPriority[pri]; !ok {
					slaByPriority[pri] = &slaPolicies[i]
				}
			}
		}
	}

	// ── 8. Tickets ────────────────────────────────────────────────────────────
	fmt.Println("→ Creating tickets…")

	type ticketMsgSpec struct {
		userKey  string
		body     string
		hoursAgo int // hours before the ticket was created (positive = before)
	}
	type ticketSeed struct {
		customerIdx  int
		subject      string
		description  string
		ticketType   string
		status       string
		priority     string
		createdByKey string
		assignedToKey string
		createdAgo   time.Duration // how long ago the ticket was created
		messages     []ticketMsgSpec
	}

	// Build a lookup from customer name → demoCustomerIDs index
	customerIdx := map[string]int{
		"Acme Corporation":     0,
		"Globex Systems":       1,
		"Initech Ltd":          2,
	}

	ticketSeeds := []ticketSeed{
		{
			customerIdx:  customerIdx["Acme Corporation"],
			subject:      "Login page returns 500 on Safari",
			description:  "Users on Safari are seeing a 500 error when trying to log in. This started after yesterday's deployment. Works fine on Chrome and Firefox.\n\n**Steps to reproduce:**\n1. Open Safari (any version)\n2. Navigate to https://app.example.com/login\n3. Enter valid credentials\n4. Click \"Sign In\"\n\n**Actual result:** White screen with \"500 Internal Server Error\"\n**Expected result:** Redirect to dashboard",
			ticketType:   "incident",
			status:       "new",
			priority:     "critical",
			createdByKey: "sarah",
			assignedToKey: "marc",
			createdAgo:   3 * time.Hour,
			messages: []ticketMsgSpec{
				{"marc", "Can reproduce in Safari 18.0. The error seems to be in the session cookie parsing — Safari strips some characters that Chrome tolerates. Looking into a fix now.", 2},
				{"sarah", "Thanks Marc. Let me know if you need me to set up a Safari remote debugging session.", 1},
			},
		},
		{
			customerIdx:  customerIdx["Globex Systems"],
			subject:      "Kubernetes cluster node drain failing",
			description:  "Our `kubectl drain` command is failing on node `gke-prod-4` with the following error:\n\n```\nerror: unable to drain node \"gke-prod-4\" ...\n```\n\nThis is blocking a scheduled security patch rollout.",
			ticketType:   "incident",
			status:       "open",
			priority:     "high",
			createdByKey: "lisa",
			assignedToKey: "marc",
			createdAgo:   24 * time.Hour,
			messages: []ticketMsgSpec{
				{"marc", "Looked at the cluster. The issue is a PodDisruptionBudget on the monitoring stack that's set to minAvailable: 100%. Patching the PDB to allow disruption and will retry the drain.", 20},
				{"lisa", "Good find. Do we need to coordinate with Globex's on-call team before proceeding?", 18},
				{"marc", "Already confirmed with their lead — they're fine with a brief window of degraded monitoring coverage. Proceeding with the patch.", 16},
			},
		},
		{
			customerIdx:  customerIdx["Acme Corporation"],
			subject:      "Invoice #2024-0042 shows wrong VAT amount",
			description:  "The VAT rate applied to invoice #2024-0042 is 21% but should be 9% for the services rendered in Q1.\n\n**Invoice details:**\n- Invoice #: 2024-0042\n- Amount: €12,450.00\n- VAT applied: €2,614.50 (21%)\n- Expected VAT: €1,120.50 (9%)",
			ticketType:   "service_request",
			status:       "pending_close",
			priority:     "medium",
			createdByKey: "admin",
			assignedToKey: "priya",
			createdAgo:   168 * time.Hour, // 7 days ago
			messages: []ticketMsgSpec{
				{"priya", "Confirmed the VAT error. The billing system incorrectly applied the default rate instead of the reduced rate that Acme qualifies for. I've corrected the invoice and it will be re-sent today.", 120},
				{"admin", "Great, thanks Priya. Please CC Sarah on the re-send so she's in the loop.", 118},
				{"priya", "Done. Invoice re-issued and sent to both Acme Finance and Sarah. Closing this out.", 96},
			},
		},
		{
			customerIdx:  customerIdx["Initech Ltd"],
			subject:      "SSO integration — SAML metadata endpoint",
			description:  "We need the SAML metadata endpoint URL so Initech can configure their identity provider for SSO.\n\nCan someone provide:\n1. The metadata endpoint\n2. The supported NameID formats\n3. Whether SP-initiated or IdP-initiated SSO is supported",
			ticketType:   "service_request",
			status:       "new",
			priority:     "low",
			createdByKey: "admin",
			assignedToKey: "elena",
			createdAgo:   48 * time.Hour,
			messages:     nil,
		},
		{
			customerIdx:  customerIdx["Acme Corporation"],
			subject:      "Rate limiting causing 429 errors on API",
			description:  "Our backend integration with Acme is seeing 429 Too Many Requests errors on the `/api/v2/orders` endpoint. We're sending ~200 req/min but the documented limit is 300 req/min.\n\n**Question:** Is there a per-IP or per-token limit that's lower than the documented rate?",
			ticketType:   "problem",
			status:       "pending_close",
			priority:     "high",
			createdByKey: "sarah",
			assignedToKey: "raj",
			createdAgo:   336 * time.Hour, // 14 days ago
			messages: []ticketMsgSpec{
				{"raj", "Investigated. The rate limiter was using a cluster-wide counter that wasn't resetting properly on token refresh. I've applied a fix and verified with a load test.", 312},
				{"sarah", "Can you share the test results?", 310},
				{"raj", "Attached in the ticket. 300 req/min sustained with 0% error rate for 30 minutes.", 308},
			},
		},
		{
			customerIdx:  customerIdx["Globex Systems"],
			subject:      "Request: Read-only dashboard access for intern",
			description:  "Globex's new intern, Alex Chen, needs read-only access to the DevOps monitoring dashboard. No write permissions required.\n\n**Details:**\n- Name: Alex Chen\n- Email: alex.chen@globex.com\n- Access level: Read-only (view only, no edit/delete)\n- Scope: devops-infra project only\n\nPlease provision and notify.",
			ticketType:   "change_request",
			status:       "new",
			priority:     "low",
			createdByKey: "lisa",
			assignedToKey: "admin",
			createdAgo:   12 * time.Hour,
			messages:     nil,
		},
	}

	createdTickets := map[string]*models.Ticket{}
	for _, ts := range ticketSeeds {
		custID := demoCustomerIDs[ts.customerIdx]
		createdBy := users[ts.createdByKey]
		assignedTo := users[ts.assignedToKey]

	createdAt := now.Add(-ts.createdAgo)
	ticket := models.Ticket{
		CustomerID:   &custID,
		Title:        ts.subject,
		Description:  ts.description,
		Type:         ts.ticketType,
		Status:       ts.status,
		Priority:     ts.priority,
		CreatedByID:  createdBy.ID,
		AssignedToID: &assignedTo.ID,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	}
	// Assign SLA policy matching the ticket's priority
	if policy, ok := slaByPriority[ts.priority]; ok {
		ticket.SlaPolicyID = &policy.ID
		if policy.ResponseTimeMinutes > 0 {
			d := createdAt.Add(time.Duration(policy.ResponseTimeMinutes) * time.Minute)
			ticket.SlaResponseDeadline = &d
		}
		if policy.ResolutionTimeMinutes > 0 {
			d := createdAt.Add(time.Duration(policy.ResolutionTimeMinutes) * time.Minute)
			ticket.SlaResolutionDeadline = &d
		}
	}
	must(db.Create(&ticket).Error)
	createdTickets[ts.subject] = &ticket

		var firstResponseAt *time.Time
		for _, m := range ts.messages {
			author := users[m.userKey]
			msgCreatedAt := ticket.CreatedAt.Add(time.Duration(-m.hoursAgo) * time.Hour)
			must(db.Create(&models.TicketMessage{
				TicketID:  ticket.ID,
				UserID:    author.ID,
				Body:      m.body,
				CreatedAt: msgCreatedAt,
				UpdatedAt: msgCreatedAt,
			}).Error)
			// Track earliest agent reply for first_response_at
			if author.ID != createdBy.ID {
				if firstResponseAt == nil || msgCreatedAt.Before(*firstResponseAt) {
					firstResponseAt = &msgCreatedAt
				}
			}
		}
		if firstResponseAt != nil {
			db.Model(&ticket).Update("first_response_at", firstResponseAt)
		}
	}
	fmt.Printf("   Created %d tickets with messages\n", len(ticketSeeds))

	// ── 8b. Ticket–Card links ─────────────────────────────────────────────────
	fmt.Println("→ Creating ticket–card links…")

	type ticketCardLinkSeed struct {
		ticketSubject string
		cardKey       string // "slug/title"
	}
	ticketCardLinks := []ticketCardLinkSeed{
		{"Login page returns 500 on Safari", "website-redesign/Accessibility audit and ARIA fixes"},
		{"Rate limiting causing 429 errors on API", "mobile-app-v2/Integration tests for authentication flow"},
		{"Kubernetes cluster node drain failing", "devops-infra/Set up Kubernetes cluster on cloud provider"},
		{"SSO integration — SAML metadata endpoint", "devops-infra/Renew and automate SSL certificate rotation"},
		{"Invoice #2024-0042 shows wrong VAT amount", "website-redesign/Update brand colour palette across all components"},
	}
	totalLinks := 0
	for _, spec := range ticketCardLinks {
		ticket, ok1 := createdTickets[spec.ticketSubject]
		card, ok2 := createdCards[spec.cardKey]
		if !ok1 || !ok2 {
			fmt.Printf("   ⚠ skipping ticket–card link %q ↔ %q (ticket=%v card=%v)\n", spec.ticketSubject, spec.cardKey, ok1, ok2)
			continue
		}
		must(db.Create(&models.TicketCardLink{
			TicketID:    ticket.ID,
			CardID:      card.ID,
			CreatedByID: users["admin"].ID,
		}).Error)
		totalLinks++
	}
	fmt.Printf("   Created %d ticket–card links\n", totalLinks)

	// ── 8c. Inbox tickets (no customer) ───────────────────────────────────────
	fmt.Println("→ Creating inbox tickets…")

	type inboxTicketSeed struct {
		subject      string
		description  string
		ticketType   string
		status       string
		priority     string
		createdByKey string
		assignedToKey string // may be empty
		createdAgo   time.Duration
		messages     []ticketMsgSpec
	}

	inboxSeeds := []inboxTicketSeed{
		{
			subject:      "Cannot access the admin panel after update",
			description:  "Hi,\n\nSince the update rolled out this morning I can no longer log into the admin panel. I get \"403 Forbidden\" immediately after entering my credentials. Other users on my team are not affected.\n\nBrowser: Chrome 124\nOS: Windows 11\n\nPlease advise.\n\nRegards,\nT. Bergmann",
			ticketType:   "incident",
			status:       "open",
			priority:     "high",
			createdByKey: "admin",
			createdAgo:   2 * time.Hour,
			messages: []ticketMsgSpec{
				{"sarah", "Thank you for reaching out. Could you tell us which role your account has and whether any permission changes were made before the update?", 1},
			},
		},
		{
			subject:      "Question about pricing for additional seats",
			description:  "Hello,\n\nWe are currently on the Team plan (10 seats) and are planning to expand to about 25 users over the next quarter. Could you send us a quote for the additional 15 seats and let us know if volume pricing applies?\n\nWe would also like to know the minimum contract term.\n\nThanks,\nM. Okonkwo\nProcurement Manager",
			ticketType:   "service_request",
			status:       "open",
			priority:     "low",
			createdByKey: "admin",
			createdAgo:   18 * time.Hour,
			messages:     nil,
		},
		{
			subject:      "Data export stuck at 0% for 3 hours",
			description:  "I triggered a full data export from Settings → Export at 09:14 this morning. The progress bar has shown 0% ever since. The export job ID is `exp_7f3a91c`. No error is shown.\n\nThis is urgent — I need the export for an audit tomorrow morning.",
			ticketType:   "incident",
			status:       "open",
			priority:     "critical",
			createdByKey: "admin",
			assignedToKey: "marc",
			createdAgo:   3*time.Hour + 20*time.Minute,
			messages: []ticketMsgSpec{
				{"marc", "I can see the export job in the queue. It appears to be blocked waiting on a database lock. I'm investigating and will have an update within 30 minutes.", 2},
			},
		},
	}

	for _, is := range inboxSeeds {
		createdBy := users[is.createdByKey]
		createdAt := now.Add(-is.createdAgo)
		iTicket := models.Ticket{
			CustomerID:  nil, // inbox — no customer assigned
			Title:       is.subject,
			Description: is.description,
			Type:        is.ticketType,
			Status:      is.status,
			Priority:    is.priority,
			CreatedByID: createdBy.ID,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
		}
		if is.assignedToKey != "" {
			iTicket.AssignedToID = &users[is.assignedToKey].ID
		}
		if policy, ok := slaByPriority[is.priority]; ok {
			iTicket.SlaPolicyID = &policy.ID
			if policy.ResponseTimeMinutes > 0 {
				d := createdAt.Add(time.Duration(policy.ResponseTimeMinutes) * time.Minute)
				iTicket.SlaResponseDeadline = &d
			}
			if policy.ResolutionTimeMinutes > 0 {
				d := createdAt.Add(time.Duration(policy.ResolutionTimeMinutes) * time.Minute)
				iTicket.SlaResolutionDeadline = &d
			}
		}
		must(db.Create(&iTicket).Error)

		var firstResponseAt *time.Time
		for _, m := range is.messages {
			author := users[m.userKey]
			msgCreatedAt := iTicket.CreatedAt.Add(time.Duration(-m.hoursAgo) * time.Hour)
			must(db.Create(&models.TicketMessage{
				TicketID:  iTicket.ID,
				UserID:    author.ID,
				Body:      m.body,
				CreatedAt: msgCreatedAt,
				UpdatedAt: msgCreatedAt,
			}).Error)
			if author.ID != createdBy.ID {
				if firstResponseAt == nil || msgCreatedAt.Before(*firstResponseAt) {
					firstResponseAt = &msgCreatedAt
				}
			}
		}
		if firstResponseAt != nil {
			db.Model(&iTicket).Update("first_response_at", firstResponseAt)
		}
	}
	fmt.Printf("   Created %d inbox tickets\n", len(inboxSeeds))

	// ── 9. Groups ─────────────────────────────────────────────────────────────
	fmt.Println("→ Creating groups…")

	type groupProjectSpec struct{ slug, role string }
	type groupCustomerSpec struct{ name, role string }
	type groupSpec struct {
		name        string
		description string
		avatar      string
		members     []string // user keys
		projects    []groupProjectSpec
		customers   []groupCustomerSpec
	}

	demoGroupSpecs := []groupSpec{
		{
			name:        "Frontend Team",
			description: "Web and mobile designers and developers.",
			avatar:      "https://api.dicebear.com/9.x/shapes/svg?seed=Frontend-Team&backgroundColor=dbeafe,e9d5ff",
			members:     []string{"sarah", "marc", "priya", "james"},
			projects: []groupProjectSpec{
				{"website-redesign", "member"},
				{"mobile-app-v2", "member"},
				{"marketing", "viewer"},
			},
		},
		{
			name:        "DevOps Team",
			description: "Infrastructure engineers and site reliability engineers.",
			avatar:      "https://api.dicebear.com/9.x/shapes/svg?seed=DevOps-Team&backgroundColor=dcfce7,bfdbfe",
			members:     []string{"marc", "lisa", "raj"},
			projects: []groupProjectSpec{
				{"devops-infra", "owner"},
				{"product-platform", "member"},
				{"api-platform", "member"},
			},
		},
		{
			name:        "Acme Stakeholders",
			description: "Client-facing team for the Acme Corporation account.",
			avatar:      "https://api.dicebear.com/9.x/shapes/svg?seed=Acme-Stakeholders&backgroundColor=fef3c7,fce7f3",
			members:     []string{"admin", "sarah", "marc"},
			customers: []groupCustomerSpec{
				{"Acme Corporation", "viewer"},
			},
		},
	}

	for _, gs := range demoGroupSpecs {
		g := &models.UserGroup{Name: gs.name, Description: gs.description, Avatar: gs.avatar}
		must(db.Create(g).Error)

		conv := models.Conversation{Name: gs.name, Avatar: gs.avatar, IsGroup: true, CreatedByID: 0}
		must(db.Create(&conv).Error)
		must(db.Model(g).Update("conversation_id", conv.ID).Error)

		now := time.Now()
		for _, key := range gs.members {
			if u, ok := users[key]; ok {
				must(db.Create(&models.GroupMember{GroupID: g.ID, UserID: u.ID}).Error)
				must(db.Create(&models.ConversationMember{
					ConversationID: conv.ID,
					UserID:         u.ID,
					JoinedAt:       now,
				}).Error)
			}
		}
		for _, pa := range gs.projects {
			if pd, ok := projects[pa.slug]; ok {
				must(db.Create(&models.GroupProjectAccess{
					GroupID:   g.ID,
					ProjectID: pd.project.ID,
					Role:      pa.role,
				}).Error)
			}
		}
		for _, ca := range gs.customers {
			var cust models.Customer
			if db.Where("name = ?", ca.name).First(&cust).Error == nil {
				must(db.Create(&models.GroupCustomerAccess{
					GroupID:    g.ID,
					CustomerID: cust.ID,
					Role:       ca.role,
				}).Error)
			}
		}
	}
	fmt.Printf("   Created %d groups\n", len(demoGroupSpecs))

	// ── 10a. Customer direct members (required for ticket assignee dropdown) ──
	// ListCustomerMembers only returns direct CustomerAccess rows (not group
	// access), so each customer needs explicit member rows for reassignment to
	// work in the UI.
	fmt.Println("→ Creating customer member access…")
	type custMemberSpec struct {
		customerName string
		userKey      string
		role         string
	}
	custMemberSpecs := []custMemberSpec{
		// Acme Corporation — admin, sarah, marc own this account
		{"Acme Corporation", "admin", "admin"},
		{"Acme Corporation", "sarah", "member"},
		{"Acme Corporation", "marc", "member"},
		// Globex Systems — DevOps team (marc, lisa, raj) + admin
		{"Globex Systems", "admin", "admin"},
		{"Globex Systems", "marc", "member"},
		{"Globex Systems", "lisa", "member"},
		{"Globex Systems", "raj", "member"},
		// Initech Ltd — admin, sarah, priya
		{"Initech Ltd", "admin", "admin"},
		{"Initech Ltd", "sarah", "member"},
		{"Initech Ltd", "priya", "member"},
	}
	for _, cm := range custMemberSpecs {
		var cust models.Customer
		if db.Where("name = ?", cm.customerName).First(&cust).Error != nil {
			continue
		}
		u, ok := users[cm.userKey]
		if !ok {
			continue
		}
		must(db.Create(&models.CustomerAccess{
			CustomerID: cust.ID,
			UserID:     u.ID,
			Role:       cm.role,
		}).Error)
	}
	fmt.Printf("   Created %d customer member rows\n", len(custMemberSpecs))

	// ── 10. Time entries ──────────────────────────────────────────────────────
	fmt.Println("→ Creating time entries…")

	// Enable time tracking for the users who will have seeded log entries.
	for _, key := range []string{"tonk", "admin", "sarah", "marc", "lisa"} {
		if u, ok := users[key]; ok {
			must(db.Model(u).Update("time_tracking_enabled", true).Error)
		}
	}

	// Resolve customer IDs by name for FK references.
	custIDByName := map[string]uint{}
	for _, name := range []string{"Acme Corporation", "Globex Systems", "Initech Ltd"} {
		var c models.Customer
		if db.Where("name = ?", name).First(&c).Error == nil {
			custIDByName[name] = c.ID
		}
	}

	type teSpec struct {
		user     string // key in users map (or "tonk")
		customer string // customer name, "" = none
		project  string // project slug,   "" = none
		daysAgo  int
		minutes  int
		desc     string
	}

	pUint := func(v uint) *uint { return &v }

	teSpecs := []teSpec{
		// ── Ton Kersten (system admin) ─────────────────────────────────────────
		{"tonk", "Acme Corporation", "website-redesign", 14, 240, "Sprint planning and backlog grooming"},
		{"tonk", "Acme Corporation", "website-redesign", 13, 390, "Architecture review and stakeholder call"},
		{"tonk", "Acme Corporation", "mobile-app-v2",    12, 300, "API design session with Marc"},
		{"tonk", "Globex Systems",   "devops-infra",     11, 480, "Kubernetes migration kick-off"},
		{"tonk", "Globex Systems",   "devops-infra",     10, 360, "Infrastructure review and documentation"},
		{"tonk", "Acme Corporation", "website-redesign",  7, 480, "Design system implementation"},
		{"tonk", "Acme Corporation", "mobile-app-v2",     6, 240, "Mobile API review"},
		{"tonk", "Globex Systems",   "devops-infra",      5, 480, "CI/CD pipeline setup"},
		{"tonk", "Acme Corporation", "website-redesign",  4, 300, "Code review and QA"},
		{"tonk", "Acme Corporation", "website-redesign",  3, 120, "Client demo preparation"},
		{"tonk", "Acme Corporation", "website-redesign",  2, 480, "Sprint review and retrospective"},
		{"tonk", "Globex Systems",   "devops-infra",      1, 360, "On-call handover and monitoring setup"},

		// ── Sarah Chen — website redesign ──────────────────────────────────────
		{"sarah", "Acme Corporation", "website-redesign", 14, 480, "Component library setup"},
		{"sarah", "Acme Corporation", "website-redesign", 13, 360, "Design implementation: hero section"},
		{"sarah", "Acme Corporation", "website-redesign", 12, 480, "Responsive layout work"},
		{"sarah", "Acme Corporation", "website-redesign", 11, 300, "Cross-browser testing"},
		{"sarah", "Acme Corporation", "website-redesign", 10, 420, "Accessibility audit and fixes"},
		{"sarah", "Acme Corporation", "website-redesign",  7, 480, "Navigation and routing refactor"},
		{"sarah", "Acme Corporation", "website-redesign",  6, 360, "CMS integration"},
		{"sarah", "Acme Corporation", "website-redesign",  5, 480, "Performance optimisation"},
		{"sarah", "Acme Corporation", "website-redesign",  4, 240, "Content migration"},
		{"sarah", "Acme Corporation", "website-redesign",  3, 420, "UAT support and bug fixes"},
		{"sarah", "Acme Corporation", "website-redesign",  2, 480, "Pre-launch checklist"},
		{"sarah", "Acme Corporation", "website-redesign",  1, 300, "Go-live support"},

		// ── Marc Dubois — mobile app + devops ──────────────────────────────────
		{"marc", "Acme Corporation", "mobile-app-v2", 14, 480, "React Native project setup"},
		{"marc", "Acme Corporation", "mobile-app-v2", 13, 360, "Authentication flow implementation"},
		{"marc", "Acme Corporation", "mobile-app-v2", 12, 480, "Push notification integration"},
		{"marc", "Globex Systems",   "devops-infra",  11, 240, "Terraform modules for Globex"},
		{"marc", "Acme Corporation", "mobile-app-v2", 10, 420, "Offline sync implementation"},
		{"marc", "Acme Corporation", "mobile-app-v2",  7, 480, "App store submission preparation"},
		{"marc", "Acme Corporation", "mobile-app-v2",  6, 300, "Beta testing and feedback"},
		{"marc", "Globex Systems",   "devops-infra",   5, 480, "Monitoring dashboard setup"},
		{"marc", "Acme Corporation", "mobile-app-v2",  4, 360, "Bug fixes from beta"},
		{"marc", "Acme Corporation", "mobile-app-v2",  3, 240, "Performance profiling"},
		{"marc", "Globex Systems",   "devops-infra",   2, 480, "Alerting rules configuration"},
		{"marc", "Acme Corporation", "mobile-app-v2",  1, 420, "App submission and review"},

		// ── Lisa Park — devops ─────────────────────────────────────────────────
		{"lisa", "Globex Systems", "devops-infra", 14, 480, "Kubernetes cluster audit"},
		{"lisa", "Globex Systems", "devops-infra", 13, 420, "Node upgrade and security patching"},
		{"lisa", "Globex Systems", "devops-infra", 12, 360, "Helm chart updates"},
		{"lisa", "Globex Systems", "devops-infra", 11, 480, "Service mesh configuration"},
		{"lisa", "Globex Systems", "devops-infra", 10, 300, "Load testing and capacity planning"},
		{"lisa", "Globex Systems", "devops-infra",  7, 480, "Disaster recovery drill"},
		{"lisa", "Globex Systems", "devops-infra",  6, 360, "Cost optimisation review"},
		{"lisa", "Globex Systems", "devops-infra",  5, 480, "Log aggregation setup"},
		{"lisa", "Globex Systems", "devops-infra",  4, 240, "Runbook documentation"},
		{"lisa", "Globex Systems", "devops-infra",  3, 420, "Incident post-mortem"},
		{"lisa", "Globex Systems", "devops-infra",  2, 480, "Database backup automation"},
		{"lisa", "Globex Systems", "devops-infra",  1, 360, "Weekly ops review"},

		// ── Alex Admin — cross-project management ──────────────────────────────
		{"admin", "Acme Corporation", "website-redesign", 14, 120, "Project kick-off meeting"},
		{"admin", "Acme Corporation", "mobile-app-v2",    13, 180, "Requirements refinement"},
		{"admin", "Acme Corporation", "",                 12, 120, "Client status update call"},
		{"admin", "Globex Systems",   "devops-infra",     11, 120, "Contract review meeting"},
		{"admin", "Acme Corporation", "website-redesign", 10, 180, "Sprint 3 review"},
		{"admin", "Acme Corporation", "mobile-app-v2",     7, 120, "Sprint planning"},
		{"admin", "Globex Systems",   "devops-infra",      6, 180, "Quarterly business review"},
		{"admin", "Acme Corporation", "website-redesign",  5, 120, "Stakeholder presentation"},
		{"admin", "",                 "",                  4, 120, "Internal team sync"},
		{"admin", "Acme Corporation", "mobile-app-v2",     3, 180, "Beta sign-off meeting"},
		{"admin", "Globex Systems",   "devops-infra",      2, 120, "SLA review"},
		{"admin", "Acme Corporation", "website-redesign",  1, 180, "Launch readiness check"},
	}

	totalTimeEntries := 0
	for _, s := range teSpecs {
		u, ok := users[s.user]
		if !ok {
			continue
		}
		date := time.Now().UTC().AddDate(0, 0, -s.daysAgo).Truncate(24 * time.Hour)
		entry := models.TimeEntry{
			UserID:      u.ID,
			Date:        date,
			Minutes:     s.minutes,
			Description: s.desc,
		}
		if cid, ok := custIDByName[s.customer]; ok {
			entry.CustomerID = pUint(cid)
		}
		if s.project != "" {
			if pd, ok := projects[s.project]; ok {
				entry.ProjectID = pUint(pd.project.ID)
			}
		}
		must(db.Create(&entry).Error)
		totalTimeEntries++
	}
	fmt.Printf("   Created %d time entries for 5 users\n", totalTimeEntries)

	// ── 10b. Personal TT-only projects and customers for tonk ─────────────────
	fmt.Println("→ Creating personal TT-only projects and customers for tonk…")

	// TT-only customers owned by tonk
	type ttCustomerSpec struct {
		name string
		desc string
	}
	ttCustomers := []ttCustomerSpec{
		{name: "Smart Owl Consulting", desc: "Internal employer — used for travel, holidays, and internal hours."},
		{name: "Personal", desc: "Purely personal time: study, training, and other non-billable activities."},
	}
	ttCustIDByName := map[string]uint{}
	for _, cs := range ttCustomers {
		var cust models.Customer
		if err := db.Where("name = ? AND time_tracking_only = ? AND created_by_id = ?", cs.name, true, tonk.ID).First(&cust).Error; err != nil {
			cust = models.Customer{
				Name:             cs.name,
				Description:      cs.desc,
				TimeTrackingOnly: true,
				CreatedByID:      &tonk.ID,
			}
			must(db.Create(&cust).Error)
			must(db.Create(&models.CustomerFavorite{UserID: tonk.ID, CustomerID: cust.ID}).Error)
		}
		ttCustIDByName[cs.name] = cust.ID
	}

	// TT-only projects owned by tonk
	type ttProjectSpec struct {
		name         string
		slug         string
		color        string
		undeclarable int // minutes per entry that cannot be declared
		customer     string
	}
	ttProjectSpecs := []ttProjectSpec{
		{name: "Travel", slug: "travel-tt", color: "#f59e0b", undeclarable: 45, customer: "Smart Owl Consulting"},
		{name: "Holidays", slug: "holidays-tt", color: "#10b981", undeclarable: 480, customer: "Smart Owl Consulting"},
		{name: "Study & Training", slug: "study-tt", color: "#6366f1", undeclarable: 0, customer: "Personal"},
		{name: "Internal", slug: "internal-tt", color: "#8b5cf6", undeclarable: 60, customer: "Smart Owl Consulting"},
	}
	ttProjIDBySlug := map[string]uint{}
	for _, ps := range ttProjectSpecs {
		var proj models.Project
		if err := db.Where("slug = ?", ps.slug).First(&proj).Error; err != nil {
			proj = models.Project{
				Name:                ps.name,
				Slug:                ps.slug,
				KeyPrefix:           services.GenerateKeyPrefix(ps.name),
				Color:               ps.color,
				TimeTrackingOnly:    true,
				UndeclarableMinutes: ps.undeclarable,
				CreatedByID:         tonk.ID,
			}
			must(db.Create(&proj).Error)
		}
		ttProjIDBySlug[ps.slug] = proj.ID
	}

	// Time entries for tonk's personal projects
	type ttTeSpec struct {
		slug     string
		customer string
		daysAgo  int
		minutes  int
		desc     string
	}
	ttEntries := []ttTeSpec{
		// Travel
		{slug: "travel-tt", customer: "Smart Owl Consulting", daysAgo: 1, minutes: 120, desc: "Travel to customer site"},
		{slug: "travel-tt", customer: "Smart Owl Consulting", daysAgo: 3, minutes: 90, desc: "Travel to Amsterdam office"},
		{slug: "travel-tt", customer: "Smart Owl Consulting", daysAgo: 8, minutes: 120, desc: "Travel to and from training location"},
		{slug: "travel-tt", customer: "Smart Owl Consulting", daysAgo: 15, minutes: 90, desc: "Travel to customer workshop"},
		// Holidays
		{slug: "holidays-tt", customer: "Smart Owl Consulting", daysAgo: 5, minutes: 480, desc: "Annual leave"},
		{slug: "holidays-tt", customer: "Smart Owl Consulting", daysAgo: 6, minutes: 480, desc: "Annual leave"},
		{slug: "holidays-tt", customer: "Smart Owl Consulting", daysAgo: 20, minutes: 480, desc: "Bank holiday"},
		// Study & Training
		{slug: "study-tt", customer: "Personal", daysAgo: 2, minutes: 180, desc: "Go concurrency patterns — self-study"},
		{slug: "study-tt", customer: "Personal", daysAgo: 9, minutes: 240, desc: "Cloud-native training course"},
		{slug: "study-tt", customer: "Personal", daysAgo: 14, minutes: 120, desc: "Read: The Pragmatic Programmer"},
		// Internal
		{slug: "internal-tt", customer: "Smart Owl Consulting", daysAgo: 1, minutes: 60, desc: "Team standup and weekly sync"},
		{slug: "internal-tt", customer: "Smart Owl Consulting", daysAgo: 4, minutes: 90, desc: "Internal retrospective meeting"},
		{slug: "internal-tt", customer: "Smart Owl Consulting", daysAgo: 7, minutes: 60, desc: "Department all-hands"},
		{slug: "internal-tt", customer: "Smart Owl Consulting", daysAgo: 11, minutes: 120, desc: "Quarterly planning session"},
	}
	for _, te := range ttEntries {
		date := time.Now().UTC().AddDate(0, 0, -te.daysAgo).Truncate(24 * time.Hour)
		entry := models.TimeEntry{
			UserID:      tonk.ID,
			Date:        date,
			Minutes:     te.minutes,
			Description: te.desc,
		}
		if cid, ok := ttCustIDByName[te.customer]; ok {
			entry.CustomerID = pUint(cid)
		}
		if pid, ok := ttProjIDBySlug[te.slug]; ok {
			entry.ProjectID = pUint(pid)
		}
		must(db.Create(&entry).Error)
	}
	fmt.Printf("   Created %d TT-only customers, %d TT-only projects, %d personal time entries for tonk\n",
		len(ttCustomers), len(ttProjectSpecs), len(ttEntries))

	// ── 11. News items ────────────────────────────────────────────────────────
	fmt.Println("→ Creating news items…")

	type newsSpec struct {
		title        string
		text         string
		sidebarColor string
		startDate    *time.Time
		endDate      *time.Time
		active       bool
		showOnLogin  bool
	}

	newsSpecs := []newsSpec{
		{
			title:        "Welcome to WarmDesk 🎉",
			sidebarColor: "#6366f1",
			active:       true,
			showOnLogin:  true,
			text: `We're excited to have you on board! WarmDesk brings all your project management tools into one place — Kanban boards, sprint planning, team chat, and voice & video calls.

**Getting started:**

- Browse your projects on the dashboard
- Open a board and drag cards between columns
- Start a conversation with a colleague in **Chats**
- Check the **Admin** panel to invite your team

If you have questions, reach out to your administrator. Happy shipping! 🚀`,
		},
		{
			title:        "New feature: Gantt chart view",
			sidebarColor: "#22c55e",
			active:       true,
			startDate:    days(-7),
			text: `You can now visualise your project timeline as a **Gantt chart**.

Open any project and click **Gantt** in the sidebar to see cards plotted by start and due date. Drag the bars to reschedule, or resize them to adjust duration.

**Tips:**
- Assign a *start date* and *due date* to a card for it to appear on the chart
- Use the zoom controls to switch between day, week, and month views
- Cards without dates are listed in the panel on the right

Available for all project types — Kanban and Scrum alike.`,
		},
		{
			title:        "Scheduled maintenance — Sunday 02:00–04:00 UTC",
			sidebarColor: "#f59e0b",
			active:       true,
			startDate:    days(-1),
			endDate:      days(3),
			text: `We will perform routine infrastructure maintenance this **Sunday between 02:00 and 04:00 UTC**.

During this window the service may be briefly unavailable. We expect total downtime to be under 10 minutes.

**What's happening:**
- Database engine upgrade
- TLS certificate renewal
- Dependency security patches

No data loss is expected. We recommend saving any open drafts before the window begins.

Apologies for the inconvenience — this is necessary to keep WarmDesk fast and secure.`,
		},
		{
			title:        "Q2 retrospective highlights",
			sidebarColor: "#8b5cf6",
			active:       true,
			startDate:    days(-14),
			text: `The Q2 all-hands retrospective is behind us. Here's a summary of what the team surfaced:

**What went well:**
- Shipped 3 major features ahead of schedule
- Customer satisfaction score up 12 points to 4.6 / 5
- Zero critical incidents in production

**What we're improving:**
- Clearer sprint goals before planning sessions start
- Faster PR review turnaround (target: < 24 h)
- Better async update culture to reduce meeting load

Full notes and action items have been added to the *Team Processes* board. Thanks to everyone who participated!`,
		},
		{
			title:        "Security reminder: enable MFA on your account",
			sidebarColor: "#ef4444",
			active:       true,
			text: `Protecting your account with **multi-factor authentication (MFA)** takes less than two minutes and significantly reduces the risk of unauthorised access.

**To enable MFA:**

1. Go to **Settings → Security**
2. Click *Set up authenticator app*
3. Scan the QR code with Google Authenticator, Authy, or any TOTP app
4. Enter the 6-digit code to confirm

Once enabled, you'll be asked for the code each time you log in from a new device.

If you have any trouble, contact your administrator.`,
		},
		{
			title:        "Team lunch — Friday 12:30",
			sidebarColor: "#06b6d4",
			active:       true,
			startDate:    days(-2),
			endDate:      days(5),
			text: `Join us for a team lunch this **Friday at 12:30** in the main meeting room (or the garden if the weather holds ☀️).

The agenda is deliberately light — good food, good company, and a chance to catch up outside of Slack and standups.

**Please RSVP** by Thursday EOD so we can get the catering right. Reply in the *#general* chat or ping @demo.admin directly.

Hope to see you there!`,
		},
	}

	totalNewsItems := 0
	for _, s := range newsSpecs {
		item := models.NewsItem{
			Title:        s.title,
			Text:         s.text,
			SidebarColor: s.sidebarColor,
			Active:       s.active,
			StartDate:    s.startDate,
			EndDate:      s.endDate,
			ShowOnLogin:  s.showOnLogin,
		}
		must(db.Create(&item).Error)
		totalNewsItems++
	}
	fmt.Printf("   Created %d news items\n", totalNewsItems)

	// ── 12. Summary ──────────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("✅ Demo data seeded successfully!")
	fmt.Println()
	fmt.Println("  Accounts (password: demo1234)")
	fmt.Println("  ┌─────────────────────┬─────────────────────┬────────┐")
	fmt.Println("  │ Username            │ Display name        │ Role   │")
	fmt.Println("  ├─────────────────────┼─────────────────────┼────────┤")
	fmt.Println("  │ tonk                │ Ton Kersten         │ admin  │  ← system admin (not reset)")
	fmt.Println("  │ demo.admin          │ Alex Admin          │ admin  │")
	fmt.Println("  │ demo.sarah          │ Sarah Chen          │ user   │  ← project admin: website-redesign")
	fmt.Println("  │ demo.marc           │ Marc Dubois         │ user   │  ← project admin: mobile-app-v2")
	fmt.Println("  │ demo.lisa           │ Lisa Park           │ user   │  ← project admin: devops-infra")
	fmt.Println("  │ demo.priya          │ Priya Nair          │ user   │")
	fmt.Println("  │ demo.james          │ James O'Brien       │ user   │")
	fmt.Println("  │ demo.elena          │ Elena Kovač         │ user   │")
	fmt.Println("  │ demo.raj            │ Raj Sharma          │ user   │")
	fmt.Println("  │ demo.viewer         │ Victor Viewer       │ viewer │")
	fmt.Println("  └─────────────────────┴─────────────────────┴────────┘")
	fmt.Println()
	fmt.Printf("  Projects      : %d (%d kanban, %d scrum)\n", len(demoProjects), len(demoProjects)-2, 2)
	fmt.Printf("  Cards         : %d\n", totalCards)
	fmt.Printf("  Topics        : %d\n", totalTopics)
	fmt.Printf("  Conversations : %d (%d messages)\n", totalConvs, totalConvMsgs)
	fmt.Printf("  Sprints       : %d (PLT: 3 completed, 1 active, 1 planning • API: 4 completed, 1 active, 1 planning)\n", totalSprints)
	fmt.Printf("  Customers     : %d (Acme Corporation, Globex Systems, Initech Ltd)\n", len(demoCustomers))
	fmt.Printf("  Groups        : %d (Frontend Team, DevOps Team, Acme Stakeholders)\n", len(demoGroupSpecs))
	fmt.Printf("  Time entries  : %d (tonk, demo.admin, demo.sarah, demo.marc, demo.lisa)\n", totalTimeEntries)
	fmt.Printf("  News items    : %d\n", totalNewsItems)
	fmt.Println()
	fmt.Println("  Starred projects")
	fmt.Println("  ┌─────────────────────┬──────────────────────────────────────────────────────────────┐")
	fmt.Println("  │ Username            │ Starred projects                                             │")
	fmt.Println("  ├─────────────────────┼──────────────────────────────────────────────────────────────┤")
	fmt.Println("  │ tonk                │ Website Redesign, Mobile App v2, DevOps & Infra              │")
	fmt.Println("  │ demo.admin          │ Website Redesign, Mobile App v2, DevOps & Infra              │")
	fmt.Println("  │ demo.sarah          │ Website Redesign, Mobile App v2                              │")
	fmt.Println("  │ demo.marc           │ Mobile App v2, DevOps & Infra                                │")
	fmt.Println("  │ demo.lisa           │ DevOps & Infra, Website Redesign                             │")
	fmt.Println("  └─────────────────────┴──────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  Starred customers")
	fmt.Println("  ┌─────────────────────┬──────────────────────────────────────────────────────────────┐")
	fmt.Println("  │ Username            │ Starred customers                                            │")
	fmt.Println("  ├─────────────────────┼──────────────────────────────────────────────────────────────┤")
	fmt.Println("  │ tonk                │ Acme Corporation, Globex Systems                             │")
	fmt.Println("  │ demo.admin          │ Acme Corporation                                             │")
	fmt.Println("  │ demo.sarah          │ Acme Corporation                                             │")
	fmt.Println("  │ demo.marc           │ Acme Corporation, Globex Systems                             │")
	fmt.Println("  │ demo.lisa           │ Globex Systems                                               │")
	fmt.Println("  └─────────────────────┴──────────────────────────────────────────────────────────────┘")
	fmt.Println()
	// Additional dynamic summary box
	fmt.Println()
	fmt.Println("  Detailed summary:")

	// Projects (name + type)
	var projectsList []models.Project
	if err := db.Find(&projectsList).Error; err == nil && len(projectsList) > 0 {
		fmt.Println()
		fmt.Println("  Projects:")
		for _, p := range projectsList {
			bt := p.BoardType
			if bt == "" {
				bt = "kanban"
			}
			fmt.Printf("    - %s (%s)\n", p.Name, bt)
		}
	}

	// Conversations (who <-> who or group members)
	var convs []models.Conversation
	if err := db.Preload("Members.User").Find(&convs).Error; err == nil && len(convs) > 0 {
		fmt.Println()
		fmt.Println("  Conversations:")
		for _, c := range convs {
			var names []string
			for _, m := range c.Members {
				if m.User.DisplayName != "" {
					names = append(names, m.User.DisplayName)
				} else {
					names = append(names, m.User.Username)
				}
			}
			if len(names) == 2 && !c.IsGroup {
				fmt.Printf("    - %s <-> %s\n", names[0], names[1])
			} else if c.Name != "" {
				fmt.Printf("    - %s (group: %s)\n", c.Name, joinNames(names))
			} else {
				fmt.Printf("    - Group (%d): %s\n", len(names), joinNames(names))
			}
		}
	}

	// Customers
	var custs []models.Customer
	if err := db.Find(&custs).Error; err == nil && len(custs) > 0 {
		fmt.Println()
		fmt.Println("  Customers:")
		for _, c := range custs {
			fmt.Printf("    - %s\n", c.Name)
		}
	}

	// Groups and members
	var groups []models.UserGroup
	if err := db.Find(&groups).Error; err == nil && len(groups) > 0 {
		fmt.Println()
		fmt.Println("  Groups:")
		for _, g := range groups {
			var memberNames []string
			db.Table("users").Select("users.display_name").Joins("join group_members gm on gm.user_id = users.id").Where("gm.group_id = ?", g.ID).Pluck("display_name", &memberNames)
			fmt.Printf("    - %s: %s\n", g.Name, joinNames(memberNames))
		}
	}

	fmt.Println()
	fmt.Println("  Start the server and log in at http://localhost:8080")
	fmt.Println()
	fmt.Println("  Repository: https://github.com/tonk/warmdesk.git")
	fmt.Println("  Preview:")
	fmt.Println("    A self-hosted, multi-user project management tool with Kanban and Scrum boards,")
	fmt.Println("    real-time collaboration, direct messaging, time tracking, and a ticket API.")
}

// joinNames helper prints a comma-separated list or '-' when empty
func joinNames(arr []string) string {
	if len(arr) == 0 {
		return "-"
	}
	s := arr[0]
	for i := 1; i < len(arr); i++ {
		s += ", " + arr[i]
	}
	return s
}

// removeDemoData deletes all records created by the seed (identified by the
// demo users and projects), then the users themselves,
func removeDemoData(db *gorm.DB) {
	demoUsernames := []string{"demo.admin", "demo.sarah", "demo.marc", "demo.lisa", "demo.viewer", "demo.priya", "demo.james", "demo.elena", "demo.raj"}
	demoSlugs := []string{"website-redesign", "mobile-app-v2", "devops-infra", "product-platform", "api-platform", "marketing", "travel-tt", "holidays-tt", "study-tt", "internal-tt"}
	demoCustomerNames := []string{"Acme Corporation", "Globex Systems", "Initech Ltd", "Smart Owl Consulting", "Personal"}

	// Collect user IDs
	var userIDs []uint
	db.Model(&models.User{}).Where("username IN ?", demoUsernames).Pluck("id", &userIDs)

	// Collect project IDs
	var projectIDs []uint
	db.Unscoped().Model(&models.Project{}).Where("slug IN ?", demoSlugs).Pluck("id", &projectIDs)

	if len(projectIDs) > 0 {
		// Collect card IDs
		var cardIDs []uint
		db.Model(&models.Card{}).Where("project_id IN ?", projectIDs).Pluck("id", &cardIDs)

		// Sprints
		var sprintIDs []uint
		db.Model(&models.Sprint{}).Where("project_id IN ?", projectIDs).Pluck("id", &sprintIDs)
		if len(sprintIDs) > 0 {
			db.Where("sprint_id IN ?", sprintIDs).Delete(&models.SprintCard{})
			db.Unscoped().Where("id IN ?", sprintIDs).Delete(&models.Sprint{})
		}

		if len(cardIDs) > 0 {
			db.Where("card_id IN ?", cardIDs).Delete(&models.CardComment{})
			db.Where("card_id IN ?", cardIDs).Delete(&models.CardChecklistItem{})
			db.Where("card_id IN ?", cardIDs).Delete(&models.CardTag{})
			db.Exec("DELETE FROM card_labels WHERE card_id IN ?", cardIDs)
			db.Exec("DELETE FROM card_assignees WHERE card_id IN ?", cardIDs)
			db.Exec("DELETE FROM sprint_cards WHERE card_id IN ?", cardIDs)
			db.Where("source_card_id IN ? OR target_card_id IN ?", cardIDs, cardIDs).Delete(&models.CardReference{})
		}

		db.Unscoped().Where("project_id IN ?", projectIDs).Delete(&models.Card{})
		db.Unscoped().Where("project_id IN ?", projectIDs).Delete(&models.Column{})
		db.Unscoped().Where("project_id IN ?", projectIDs).Delete(&models.Label{})
		db.Where("project_id IN ?", projectIDs).Delete(&models.StarredProject{})
		db.Where("project_id IN ?", projectIDs).Delete(&models.ProjectMember{})

		// Topics
		var topicIDs []uint
		db.Model(&models.Topic{}).Where("project_id IN ?", projectIDs).Pluck("id", &topicIDs)
		if len(topicIDs) > 0 {
			db.Unscoped().Where("topic_id IN ?", topicIDs).Delete(&models.TopicReply{})
		}
		db.Unscoped().Where("project_id IN ?", projectIDs).Delete(&models.Topic{})
		// Delete time entries referencing these projects (covers tonk's TT entries)
		db.Where("project_id IN ?", projectIDs).Delete(&models.TimeEntry{})
		db.Unscoped().Where("id IN ?", projectIDs).Delete(&models.Project{})
	}

	// Conversations created by or involving demo users
	if len(userIDs) > 0 {
		var convIDs []uint
		db.Model(&models.ConversationMember{}).
			Where("user_id IN ?", userIDs).
			Pluck("conversation_id", &convIDs)
		if len(convIDs) > 0 {
			db.Unscoped().Where("conversation_id IN ?", convIDs).Delete(&models.ConversationMessage{})
			db.Where("conversation_id IN ?", convIDs).Delete(&models.ConversationMember{})
			db.Unscoped().Where("id IN ?", convIDs).Delete(&models.Conversation{})
		}
		db.Where("user_id IN ?", userIDs).Delete(&models.TimeEntry{})
		db.Unscoped().Where("id IN ?", userIDs).Delete(&models.User{})
	}

	// Groups
	demoGroupNames := []string{"Frontend Team", "DevOps Team", "Acme Stakeholders"}
	var groupIDs []uint
	db.Model(&models.UserGroup{}).Where("name IN ?", demoGroupNames).Pluck("id", &groupIDs)
	if len(groupIDs) > 0 {
		db.Where("group_id IN ?", groupIDs).Delete(&models.GroupMember{})
		db.Where("group_id IN ?", groupIDs).Delete(&models.GroupProjectAccess{})
		db.Where("group_id IN ?", groupIDs).Delete(&models.GroupCustomerAccess{})
		db.Where("id IN ?", groupIDs).Delete(&models.UserGroup{})
	}

	// Tickets and SLA policies
	var custIDs []uint
	db.Model(&models.Customer{}).Where("name IN ?", demoCustomerNames).Pluck("id", &custIDs)
	if len(custIDs) > 0 {
		var ticketIDs []uint
		db.Model(&models.Ticket{}).Where("customer_id IN ?", custIDs).Pluck("id", &ticketIDs)
		if len(ticketIDs) > 0 {
			db.Unscoped().Where("ticket_id IN ?", ticketIDs).Delete(&models.TicketMessage{})
			db.Unscoped().Where("ticket_id IN ?", ticketIDs).Delete(&models.TicketTag{})
			db.Unscoped().Where("source_ticket_id IN ? OR target_ticket_id IN ?", ticketIDs, ticketIDs).Delete(&models.TicketLink{})
			db.Unscoped().Where("ticket_id IN ?", ticketIDs).Delete(&models.TicketCardLink{})
			db.Unscoped().Where("id IN ?", ticketIDs).Delete(&models.Ticket{})
		}
	}
	// Inbox tickets (customer_id IS NULL — matched by title)
	demoInboxTitles := []string{
		"Cannot access the admin panel after update",
		"Question about pricing for additional seats",
		"Data export stuck at 0% for 3 hours",
	}
	var inboxIDs []uint
	db.Model(&models.Ticket{}).Where("customer_id IS NULL AND title IN ?", demoInboxTitles).Pluck("id", &inboxIDs)
	if len(inboxIDs) > 0 {
		db.Unscoped().Where("ticket_id IN ?", inboxIDs).Delete(&models.TicketMessage{})
		db.Unscoped().Where("ticket_id IN ?", inboxIDs).Delete(&models.TicketTag{})
		db.Unscoped().Where("source_ticket_id IN ? OR target_ticket_id IN ?", inboxIDs, inboxIDs).Delete(&models.TicketLink{})
		db.Unscoped().Where("id IN ?", inboxIDs).Delete(&models.Ticket{})
	}
	demoSlaNames := []string{"Critical — 1h response / 4h resolution", "High — 2h response / 8h resolution", "Standard — 4h response / 24h resolution"}
	db.Unscoped().Where("name IN ?", demoSlaNames).Delete(&models.SlaPolicy{})

	// Customers and contracts
	db.Model(&models.Customer{}).Where("name IN ?", demoCustomerNames).Pluck("id", &custIDs)
	if len(custIDs) > 0 {
		db.Where("customer_id IN ?", custIDs).Delete(&models.CustomerFavorite{})
		db.Where("customer_id IN ?", custIDs).Delete(&models.Contract{})
		db.Where("id IN ?", custIDs).Delete(&models.Customer{})
	}

	// News items (matched by title so re-seeding is clean)
	demoNewsTitles := []string{
		"Welcome to WarmDesk 🎉",
		"New feature: Gantt chart view",
		"Scheduled maintenance — Sunday 02:00–04:00 UTC",
		"Q2 retrospective highlights",
		"Security reminder: enable MFA on your account",
		"Team lunch — Friday 12:30",
	}
	db.Unscoped().Where("title IN ?", demoNewsTitles).Delete(&models.NewsItem{})

	fmt.Println("   Done.")
}
