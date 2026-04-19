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
	"time"

	"github.com/tonk/warmdesk/config"
	"github.com/tonk/warmdesk/database"
	"github.com/tonk/warmdesk/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
		boardType: "scrum",
		desc:      "The core SaaS platform — built sprint by sprint with a cross-functional team.",
		columns:   []string{"To Do", "In Progress", "In Review", "Done"},
		labels: []struct{ name, color string }{
			{"Bug", "#ef4444"}, {"Feature", "#3b82f6"},
			{"Enhancement", "#f59e0b"}, {"Tech Debt", "#6b7280"},
		},
	},
	{
		name:      "Marketing Campaigns",
		slug:      "marketing",
		prefix:    "MKT",
		color:     "#f43f5e",
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
	configPath := flag.String("config", "", "path to warmdesk.yaml (optional)")
	reset := flag.Bool("reset", false, "remove existing demo data before seeding")
	flag.Parse()

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
		},
		"sarah": {
			Email: "sarah@demo.example", Username: "demo.sarah",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Sarah", LastName: "Chen", DisplayName: "Sarah Chen",
			IsActive: true, EmailNotifications: true,
		},
		"marc": {
			Email: "marc@demo.example", Username: "demo.marc",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Marc", LastName: "Dubois", DisplayName: "Marc Dubois",
			IsActive: true, EmailNotifications: true,
		},
		"lisa": {
			Email: "lisa@demo.example", Username: "demo.lisa",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Lisa", LastName: "Park", DisplayName: "Lisa Park",
			IsActive: true, EmailNotifications: true,
		},
		"viewer": {
			Email: "viewer@demo.example", Username: "demo.viewer",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "viewer",
			FirstName: "Victor", LastName: "Viewer", DisplayName: "Victor Viewer",
			IsActive: true, EmailNotifications: false,
		},
		// Additional demo users
		"priya": {
			Email: "priya@demo.example", Username: "demo.priya",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Priya", LastName: "Nair", DisplayName: "Priya Nair",
			IsActive: true, EmailNotifications: true,
		},
		"james": {
			Email: "james@demo.example", Username: "demo.james",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "James", LastName: "O'Brien", DisplayName: "James O'Brien",
			IsActive: true, EmailNotifications: true,
		},
		"elena": {
			Email: "elena@demo.example", Username: "demo.elena",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Elena", LastName: "Kovač", DisplayName: "Elena Kovač",
			IsActive: true, EmailNotifications: true,
		},
		"raj": {
			Email: "raj@demo.example", Username: "demo.raj",
			PasswordHash: hashPassword("demo1234"), GlobalRole: "user",
			FirstName: "Raj", LastName: "Sharma", DisplayName: "Raj Sharma",
			IsActive: true, EmailNotifications: false,
		},
	}

	for _, u := range users {
		must(db.Create(u).Error)
	}
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
			Color: sp.color, Description: sp.desc, BoardType: boardType,
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
		col         string
		priority    string
		labels      []string
		tags        []string
		assignee    string // user key or ""
		startInDays *int
		dueInDays   *int
		timeMin     int // time_spent_minutes
		storyPoints int // 0 = not set
		sprintName  string // sprint name to link this card to (scrum projects)
		checklist   []struct{ body string; done bool }
		comments    []struct{ author string; body string }
		subCards    []cardSpec
	}

	webCards := []cardSpec{
		// Backlog
		{
			title: "Redesign homepage hero section", col: "Backlog",
			priority: "high", labels: []string{"Feature", "Design"},
			tags: []string{"homepage", "design"},
		},
		{
			title: "Add cookie consent banner", col: "Backlog",
			priority: "medium", labels: []string{"Feature"},
			assignee: "marc", startInDays: ptr(7), dueInDays: ptr(14),
		},
		{
			title: "Write new About page copy", col: "Backlog",
			priority: "none", labels: []string{"Content"},
			assignee: "sarah",
		},
		// In Progress
		{
			title: "Implement dark mode toggle", col: "In Progress",
			priority: "medium", labels: []string{"Feature", "Design"},
			assignee: "sarah", timeMin: 180, startInDays: ptr(-3), dueInDays: ptr(5),
			checklist: []struct{ body string; done bool }{
				{"Research CSS variables approach", true},
				{"Implement toggle button in header", true},
				{"Persist preference to localStorage", false},
				{"Test on Safari and Firefox", false},
			},
			comments: []struct{ author string; body string }{
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
			priority: "high", labels: []string{"Bug"},
			assignee: "marc", timeMin: 90, startInDays: ptr(0), dueInDays: ptr(2),
			comments: []struct{ author string; body string }{
				{"sarah", "Reproduced on iPhone SE (375 px). The hamburger menu clips behind the logo."},
				{"marc", "On it — looks like `overflow: hidden` is missing on the nav container."},
			},
		},
		{
			title: "Update brand colour palette across all components", col: "In Progress",
			priority: "low", labels: []string{"Design"},
			assignee: "sarah", timeMin: 120,
		},
		// Review
		{
			title: "Optimise image loading with lazy + WebP", col: "Review",
			priority: "medium", labels: []string{"Feature"},
			assignee: "sarah", timeMin: 240, startInDays: ptr(-5), dueInDays: ptr(1),
			comments: []struct{ author string; body string }{
				{"marc", "LCP went from 3.8 s to 1.2 s in local tests — great improvement!"},
				{"sarah", "Waiting for sign-off on the WebP fallback strategy for older browsers."},
			},
		},
		{
			title: "Accessibility audit and ARIA fixes", col: "Review",
			priority: "high", labels: []string{"Feature"},
			assignee: "marc", timeMin: 300, startInDays: ptr(-7), dueInDays: ptr(3),
		},
		// Done
		{
			title: "Set up GitHub Actions CI/CD pipeline", col: "Done",
			priority: "high", labels: []string{"Feature"},
			assignee: "admin", timeMin: 360, startInDays: ptr(-16), dueInDays: ptr(-10),
		},
		{
			title: "Migrate DNS and SSL to new hosting provider", col: "Done",
			priority: "critical", labels: []string{"Feature"},
			assignee: "admin", timeMin: 480, startInDays: ptr(-9), dueInDays: ptr(-5),
		},
		{
			title: "Create component library documentation", col: "Done",
			priority: "low", labels: []string{"Design", "Content"},
			assignee: "sarah", timeMin: 210, startInDays: ptr(-10), dueInDays: ptr(-3),
		},
		{
			title: "Audit and fix all broken links", col: "Done",
			priority: "medium", labels: []string{"Bug"},
			assignee: "marc", timeMin: 60, startInDays: ptr(-9), dueInDays: ptr(-7),
		},
	}

	mobCards := []cardSpec{
		// Ideas
		{
			title: "Offline mode with sync queue", col: "Ideas",
			priority: "high", labels: []string{"Enhancement"},
			tags: []string{"offline", "ux"},
		},
		{
			title: "Push notification preferences screen", col: "Ideas",
			priority: "medium", labels: []string{"Enhancement"},
		},
		{
			title: "Biometric login (Face ID / fingerprint)", col: "Ideas",
			priority: "medium", labels: []string{"Enhancement"},
			tags: []string{"security", "auth"},
		},
		// Development
		{
			title: "User profile screen", col: "Development",
			priority: "high", labels: []string{"Enhancement", "iOS", "Android"},
			assignee: "sarah", timeMin: 300, startInDays: ptr(-7), dueInDays: ptr(7),
			checklist: []struct{ body string; done bool }{
				{"Approve design mockup", true},
				{"Build form components", true},
				{"Connect to profile API", true},
				{"Add avatar upload", false},
				{"Write E2E tests", false},
			},
			comments: []struct{ author string; body string }{
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
			priority: "medium", labels: []string{"Enhancement", "iOS", "Android"},
			assignee: "marc", timeMin: 180, startInDays: ptr(0), dueInDays: ptr(10),
		},
		{
			title: "Dark mode support", col: "Development",
			priority: "low", labels: []string{"Enhancement"},
			assignee: "lisa", timeMin: 240, startInDays: ptr(-5), dueInDays: ptr(14),
		},
		{
			title: "App crashes on empty conversation list", col: "Development",
			priority: "critical", labels: []string{"Bug"},
			assignee: "marc", timeMin: 45, startInDays: ptr(0), dueInDays: ptr(1),
			comments: []struct{ author string; body string }{
				{"lisa", "Stack trace attached. Null check missing in `ConversationListScreen.tsx` line 42."},
				{"marc", "Fix is one line — pushing a hotfix build now."},
			},
		},
		// Testing
		{
			title: "Integration tests for authentication flow", col: "Testing",
			priority: "high", labels: []string{"Enhancement", "iOS", "Android"},
			assignee: "lisa", timeMin: 360, startInDays: ptr(-4), dueInDays: ptr(3),
		},
		{
			title: "Performance profiling on low-end Android devices", col: "Testing",
			priority: "medium", labels: []string{"Enhancement", "Android"},
			assignee: "sarah", timeMin: 150, startInDays: ptr(-3), dueInDays: ptr(7),
			comments: []struct{ author string; body string }{
				{"marc", "Focus on the feed screen — renders ~200 items without virtualisation right now."},
			},
		},
		// Released
		{
			title: "Initial 1.0 app release", col: "Released",
			priority: "critical", labels: []string{"Enhancement", "iOS", "Android"},
			assignee: "sarah", timeMin: 1200, startInDays: ptr(-60), dueInDays: ptr(-30),
		},
		{
			title: "Bug fix release 1.0.1", col: "Released",
			priority: "high", labels: []string{"Bug"},
			assignee: "marc", timeMin: 120, startInDays: ptr(-18), dueInDays: ptr(-14),
		},
	}

	infCards := []cardSpec{
		// Todo
		{
			title: "Set up Kubernetes cluster on cloud provider", col: "Todo",
			priority: "high", labels: []string{"Enhancement"},
			startInDays: ptr(3), dueInDays: ptr(21),
			checklist: []struct{ body string; done bool }{
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
			priority: "medium", labels: []string{"Monitoring"},
			assignee: "lisa", startInDays: ptr(7), dueInDays: ptr(28),
		},
		{
			title: "Quarterly security audit", col: "Todo",
			priority: "critical", labels: []string{"Security"},
			startInDays: ptr(1), dueInDays: ptr(7),
		},
		// In Progress
		{
			title: "Migrate primary database to PostgreSQL", col: "In Progress",
			priority: "high", labels: []string{"Enhancement"},
			assignee: "marc", timeMin: 480, startInDays: ptr(-10), dueInDays: ptr(4),
			comments: []struct{ author string; body string }{
				{"admin", "Remember to keep the SQLite DB as a read-only fallback for at least two weeks after cutover."},
				{"marc", "Agreed. I've scripted the data export — schema diff is smaller than expected."},
				{"lisa", "I'll keep an eye on slow-query logs during the transition."},
			},
		},
		{
			title: "Automate database backups with off-site retention", col: "In Progress",
			priority: "high", labels: []string{"Monitoring"},
			assignee: "lisa", timeMin: 180, startInDays: ptr(-2), dueInDays: ptr(6),
		},
		{
			title: "Renew and automate SSL certificate rotation", col: "In Progress",
			priority: "critical", labels: []string{"Security"},
			assignee: "marc", timeMin: 60, startInDays: ptr(-1), dueInDays: ptr(2),
		},
		// Done
		{
			title: "Set up private Docker registry", col: "Done",
			priority: "medium", labels: []string{"Enhancement"},
			assignee: "marc", timeMin: 240, startInDays: ptr(-14), dueInDays: ptr(-10),
		},
		{
			title: "Configure Nginx load balancer with health checks", col: "Done",
			priority: "high", labels: []string{"Enhancement"},
			assignee: "lisa", timeMin: 300, startInDays: ptr(-13), dueInDays: ptr(-8),
		},
		{
			title: "Deploy staging environment", col: "Done",
			priority: "high", labels: []string{"Enhancement"},
			assignee: "marc", timeMin: 360, startInDays: ptr(-21), dueInDays: ptr(-15),
		},
	}

	pltCards := []cardSpec{
		// Sprint 1 — Foundation (completed): all in Done
		{title: "Set up monorepo and CI/CD pipeline", col: "Done", priority: "high",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 240,
			storyPoints: 3, sprintName: "Sprint 1 — Foundation",
			startInDays: ptr(-28), dueInDays: ptr(-21)},
		{title: "Design system and component library", col: "Done", priority: "high",
			labels: []string{"Feature"}, assignee: "sarah", timeMin: 480,
			storyPoints: 5, sprintName: "Sprint 1 — Foundation",
			startInDays: ptr(-27), dueInDays: ptr(-20)},
		{title: "User registration and login flow", col: "Done", priority: "critical",
			labels: []string{"Feature"}, assignee: "priya", timeMin: 360,
			storyPoints: 5, sprintName: "Sprint 1 — Foundation",
			startInDays: ptr(-26), dueInDays: ptr(-19),
			comments: []struct{ author string; body string }{
				{"sarah", "Let's go with JWT + refresh tokens from the start — much easier to scale later."},
				{"priya", "Agreed. Done. Refresh token is stored httpOnly to prevent XSS."},
			}},
		{title: "Database schema v1 with migrations", col: "Done", priority: "high",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 180,
			storyPoints: 3, sprintName: "Sprint 1 — Foundation",
			startInDays: ptr(-25), dueInDays: ptr(-18)},
		{title: "Deployment to staging (Docker + nginx)", col: "Done", priority: "medium",
			labels: []string{"Feature"}, assignee: "james", timeMin: 120,
			storyPoints: 2, sprintName: "Sprint 1 — Foundation",
			startInDays: ptr(-24), dueInDays: ptr(-17)},

		// Sprint 2 — Core Features (active): In Progress / In Review
		{title: "User dashboard overview screen", col: "In Review", priority: "high",
			labels: []string{"Feature"}, assignee: "sarah", timeMin: 300,
			storyPoints: 5, sprintName: "Sprint 2 — Core Features",
			startInDays: ptr(-13), dueInDays: ptr(1),
			checklist: []struct{ body string; done bool }{
				{"Design approved in Figma", true},
				{"API endpoints wired up", true},
				{"Responsive layout", true},
				{"Unit tests written", false},
			}},
		{title: "Email notification service", col: "In Progress", priority: "medium",
			labels: []string{"Feature"}, assignee: "james", timeMin: 120,
			storyPoints: 3, sprintName: "Sprint 2 — Core Features",
			startInDays: ptr(-10), dueInDays: ptr(3)},
		{title: "Full-text search endpoint", col: "In Progress", priority: "high",
			labels: []string{"Feature"}, assignee: "raj", timeMin: 180,
			storyPoints: 5, sprintName: "Sprint 2 — Core Features",
			startInDays: ptr(-9), dueInDays: ptr(4)},
		{title: "Role-based access control (RBAC)", col: "In Review", priority: "high",
			labels: []string{"Feature"}, assignee: "priya", timeMin: 240,
			storyPoints: 3, sprintName: "Sprint 2 — Core Features",
			startInDays: ptr(-12), dueInDays: ptr(2),
			comments: []struct{ author string; body string }{
				{"elena", "PR looks good overall — one concern about the middleware order. Discussed in PR comments."},
				{"priya", "Fixed! Ready for re-review."},
			}},
		{title: "API rate limiting middleware", col: "In Progress", priority: "medium",
			labels: []string{"Enhancement"}, assignee: "james", timeMin: 60,
			storyPoints: 2, sprintName: "Sprint 2 — Core Features",
			startInDays: ptr(-5), dueInDays: ptr(6)},

		// Sprint 3 — Polish (planning): all in To Do
		{title: "Stripe payment integration", col: "To Do", priority: "critical",
			labels: []string{"Feature"}, storyPoints: 8, sprintName: "Sprint 3 — Polish & Launch",
			dueInDays: ptr(21)},
		{title: "Analytics event tracking", col: "To Do", priority: "high",
			labels: []string{"Feature"}, storyPoints: 5, sprintName: "Sprint 3 — Polish & Launch"},
		{title: "Performance audit and optimizations", col: "To Do", priority: "medium",
			labels: []string{"Tech Debt"}, storyPoints: 3, sprintName: "Sprint 3 — Polish & Launch"},
		{title: "Launch checklist and public docs", col: "To Do", priority: "medium",
			labels: []string{"Feature"}, storyPoints: 2, sprintName: "Sprint 3 — Polish & Launch"},

		// Backlog (no sprint)
		{title: "Multi-language support (i18n)", col: "To Do", priority: "low",
			labels: []string{"Enhancement"}, tags: []string{"i18n", "ux"}},
		{title: "Mobile app wrapper (Capacitor)", col: "To Do", priority: "low",
			labels: []string{"Feature"}, tags: []string{"mobile"}},
		{title: "Fix sorting bug on dashboard table", col: "To Do", priority: "high",
			labels: []string{"Bug"}},
	}

	mktCards := []cardSpec{
		// Ideas
		{title: "Summer product launch campaign", col: "Ideas", priority: "high",
			labels: []string{"Campaign"}, tags: []string{"launch", "summer"}},
		{title: "Customer testimonial video series", col: "Ideas", priority: "medium",
			labels: []string{"Content"}},
		// Planned
		{title: "Q2 newsletter redesign", col: "Planned", priority: "high",
			labels: []string{"Email"}, assignee: "lisa",
			startInDays: ptr(3), dueInDays: ptr(14)},
		{title: "Influencer partnership outreach", col: "Planned", priority: "medium",
			labels: []string{"Campaign"}, assignee: "james",
			startInDays: ptr(5), dueInDays: ptr(21)},
		// In Progress
		{title: "Blog: 10 productivity tips for remote teams", col: "In Progress", priority: "medium",
			labels: []string{"Content"}, assignee: "sarah", timeMin: 90,
			startInDays: ptr(-5), dueInDays: ptr(3),
			comments: []struct{ author string; body string }{
				{"lisa", "Can you work in a mention of our new features in tips 7 and 10?"},
				{"sarah", "On it — draft ready for review tomorrow."},
			}},
		{title: "LinkedIn ad campaign — EMEA region", col: "In Progress", priority: "high",
			labels: []string{"Campaign"}, assignee: "james", timeMin: 120,
			startInDays: ptr(-7), dueInDays: ptr(2)},
		{title: "Trade show booth materials", col: "In Progress", priority: "critical",
			labels: []string{"Campaign"}, assignee: "lisa", timeMin: 180,
			startInDays: ptr(-3), dueInDays: ptr(3),
			checklist: []struct{ body string; done bool }{
				{"Banner design approved", true},
				{"Flyers printed", false},
				{"Demo laptop prepared", false},
				{"Shipping arranged", false},
			}},
		// Published
		{title: "Q1 newsletter — Spring edition", col: "Published", priority: "medium",
			labels: []string{"Email"}, assignee: "lisa", timeMin: 120,
			startInDays: ptr(-20), dueInDays: ptr(-12)},
		{title: "Product Hunt launch announcement", col: "Published", priority: "high",
			labels: []string{"Campaign"}, assignee: "admin", timeMin: 60,
			startInDays: ptr(-30), dueInDays: ptr(-25)},
		{title: "Year in review blog post", col: "Published", priority: "low",
			labels: []string{"Content"}, assignee: "sarah", timeMin: 150,
			startInDays: ptr(-45), dueInDays: ptr(-38)},
	}

	projectCards := map[string][]cardSpec{
		"website-redesign":  webCards,
		"mobile-app-v2":     mobCards,
		"devops-infra":      infCards,
		"product-platform":  pltCards,
		"marketing":         mktCards,
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
			card := &models.Card{
				ColumnID:         col.ID,
				ProjectID:        pd.project.ID,
				Title:            spec.title,
				Priority:         priority,
				AssigneeID:       assigneeID,
				CreatedByID:      users["admin"].ID,
				Position:         float64((i + 1) * 1000),
				CardNumber:       proj.CardCounter,
				StartDate:        startDate,
				DueDate:          dueDate,
				TimeSpentMinutes: spec.timeMin,
				StoryPoints:      sp,
			}
			must(db.Create(card).Error)
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
				name: "Sprint 1 — Foundation", status: "completed",
				goal:      "Establish the core codebase, CI/CD pipeline, and authentication layer.",
				startDays: -28, endDays: -14, hasDate: true,
			},
			{
				name: "Sprint 2 — Core Features", status: "active",
				goal:      "Ship user dashboard, email notifications, search, and RBAC.",
				startDays: -14, endDays: 7, hasDate: true,
			},
			{
				name: "Sprint 3 — Polish & Launch", status: "planning",
				goal:      "Payment integration, analytics, performance, and public launch.",
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

	// ── 3c. Card cross-references ─────────────────────────────────────────────
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
		// Group chat: Website Redesign team
		{
			members: []string{"admin", "sarah", "marc"},
			isGroup:  true,
			name:     "Website Redesign Team",
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

	type contractSpec struct {
		name      string
		desc      string
		startDays int  // relative to today
		endDays   int  // 0 = no end date
		projects  []string // project slugs to link
	}
	type customerSpec struct {
		name      string
		desc      string
		logoURL   string
		contracts []contractSpec
		projects  []string // unassigned projects (no contract)
	}

	demoCustomers := []customerSpec{
		{
			name: "Acme Corporation",
			desc: "Long-running client delivering internal tooling and web platforms.",
			contracts: []contractSpec{
				{
					name:      "Phase 1 — Marketing Site",
					desc:      "Full redesign and relaunch of the corporate marketing website.",
					startDays: -180,
					endDays:   90,
					projects:  []string{"website-redesign"},
				},
				{
					name:      "Phase 2 — Mobile Apps",
					desc:      "iOS and Android companion apps for the new platform.",
					startDays: -60,
					endDays:   180,
					projects:  []string{"mobile-app-v2"},
				},
			},
		},
		{
			name: "Globex Systems",
			desc: "Infrastructure-heavy client focused on cloud reliability and security.",
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
				CustomerID:  cust.ID,
				Name:        conSpec.name,
				Description: conSpec.desc,
				StartDate:   &start,
			}
			if conSpec.endDays != 0 {
				end := time.Now().UTC().AddDate(0, 0, conSpec.endDays).Truncate(24 * time.Hour)
				con.EndDate = &end
			}
			must(db.Create(con).Error)

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

	// ── 7. Summary ────────────────────────────────────────────────────────────
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
	fmt.Printf("  Projects      : %d (%d kanban, %d scrum)\n", len(demoProjects), len(demoProjects)-1, 1)
	fmt.Printf("  Cards         : %d\n", totalCards)
	fmt.Printf("  Topics        : %d\n", totalTopics)
	fmt.Printf("  Conversations : %d (%d messages)\n", totalConvs, totalConvMsgs)
	fmt.Printf("  Sprints       : %d (Product Platform — 1 completed, 1 active, 1 planning)\n", totalSprints)
	fmt.Printf("  Customers     : %d (Acme Corporation, Globex Systems, Initech Ltd)\n", len(demoCustomers))
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
	fmt.Println("  Start the server and log in at http://localhost:8080")
}

// removeDemoData deletes all records created by the seed (identified by the
// demo users and projects), then the users themselves.
func removeDemoData(db *gorm.DB) {
	demoUsernames := []string{"demo.admin", "demo.sarah", "demo.marc", "demo.lisa", "demo.viewer", "demo.priya", "demo.james", "demo.elena", "demo.raj"}
	demoSlugs := []string{"website-redesign", "mobile-app-v2", "devops-infra", "product-platform", "marketing"}
	demoCustomerNames := []string{"Acme Corporation", "Globex Systems", "Initech Ltd"}

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
		db.Unscoped().Where("id IN ?", userIDs).Delete(&models.User{})
	}

	// Customers and contracts
	var custIDs []uint
	db.Model(&models.Customer{}).Where("name IN ?", demoCustomerNames).Pluck("id", &custIDs)
	if len(custIDs) > 0 {
		db.Where("customer_id IN ?", custIDs).Delete(&models.CustomerFavorite{})
		db.Where("customer_id IN ?", custIDs).Delete(&models.Contract{})
		db.Where("id IN ?", custIDs).Delete(&models.Customer{})
	}

	fmt.Println("   Done.")
}
