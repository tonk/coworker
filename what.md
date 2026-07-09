Create an application that has all these features and requirements

- Do not use Docker or Podman
- Written in Golang, minumum version 1.25
- Webbased
- Kanban board
- Scrum board
  * Board type (Kanban or Scrum) is chosen at project creation and cannot be changed
  * Scrum projects have a Backlog view (two-panel sprint planner with drag-and-drop, sprint CRUD, goal/date editing, velocity chart)
  * Scrum projects have a Sprint Board view (filtered to active sprint cards)
  * Sprint lifecycle: planning → active → completed
  * Completing a sprint returns unfinished cards to the backlog
  * Story points per card (admin toggle to enable)
- Multi user
- Multi project
- Multi language support
  * English
  * Dutch
  * German
  * French
  * Spanish
- Database support for
  * SQLlite
  * MySQL
  * PostgreSQL
- Database configuration through config file
- Allow another config file as an option on the command line
- Role based access control
- Internal chat function
- Group chat between members
- Allow chat to be scaled horizontally
- All text editors must support Github markdown
- Allow for dark, light and follow-system interface
- Add settings per user
  * First name
  * Last name
  * Email address
  * Avatar, with Gravatar support
  * Change password
  * Date / time format
    ** Use IЅO timeformat as the default
  * Timezone
    ** Use UTC as the default
  * Interface font and font size
- Allow Admin settings for
  * For all users
    ** Change all user settings
    ** Change all project settings
  * Create new users
  * Disable users
  * Delete users
  * Switch the register on the login page on/off
    ** Remove "Don't have an account? Register" from the login page if
       the setting is off
  * Global system settings, can be overruled by the user
    ** Date / time format
       ** Use IЅO timeformat as the default
    * * Timezone
      ** Use UTC as the default
    ** Interface color, dark, light or follow-system interface
    ** Interface font and font size
    ** Default language setting, English by default
  * Add, change and remove projects
- Keep track of last login
- Keep track of last user settings change
- Enable a sidebar
  * Collapsible list of starred projects
  * Collapsible list of all projects, starred project marked and at the top
  * Collapsible list of all users, in chat users marked and at the top
  * Make the users clickable and when clicked, open an direct chat box to them
- Allow sidebar to be moved to the left or right
- In the projects
  * Allow the column names to be edited
  * Allow the columns to be moved
  * Allow project to be starred for the sidebar
  * In the "Invite members" box, give a pull down with all active users
- When a new project is created, automatically create a default column
  called "Inbox"
- In the ticket overview, place the users avatar at the top right of the
  box
- In the tickets
  * Keep history of column changes
  * In the "Edit Card" box
    ** Allow the box to be resized horizontal and vertical
    ** Add "Save" button
    ** Start the date picker with the current date
    ** "Due date" field must follow user date / time settings
    ** "Due date" must be saved with the card
- Enable an API to control all tickets
  * Add
  * Comment
  * Move to other column
- Add a footer to the UI
  * Application name and version number
    ** Align left
  * Logged users full name
    ** Align right
- In the tickets
  * Allow multiple assignees per card
  * Allow watching a card to receive notifications
  * Allow sorting cards within a column by due date, assignee or priority
  * Add a checklist to a card
  * Add a "Time Spent" field (hours and minutes) to log effort on a card
- Add topics (threaded discussions) per project
  * Create, edit and delete topics
  * Reply to topics with markdown support
- Add a viewer role to role based access control
  * Viewers can read but not create or modify content
- Favourite people
  * Mark users as favourites for quick access
- Direct message notifications
  * Notify users of new direct messages
- Group direct messages
  * Start a group chat between multiple users
- Admin can assign users to projects directly
- Allow admin to reset user passwords
- Build a desktop app using Tauri
  * Distribute as AppImage (Linux), DMG (macOS) and installer (Windows)
- SMTP email configuration through the admin web interface
  * Username and password are optional (support auth-less relay servers)
  * Settings take effect without a server restart
- Time reports
  * Generate a time report filtered by period (all time, year, month, ISO week)
    and optionally by project
  * Export to PDF (print-optimised layout)
  * Export to Excel (XLSX)
  * Report header shows configurable company name and logo
- Company branding
  * Admin can set a company name and logo (URL or uploaded image)
  * Used in report headers
- Demo seed tool
  * A standalone `warmdesk-seed` binary included in the distribution
  * Populates the database with demo users (admin, members, viewer), projects,
    cards, checklists, comments, time entries, and discussion topics
  * Idempotent — safe to run multiple times; supports --reset flag

- Add README.md with explanation of WarmDesk
- Add systemd example service files
- Add Nginx example configuration file
- Add Apache example configuration file
- Add a decent .gitignore file
- Build production version
- Create installation manual
- Create distribution package containing everything needed
- Create a Github actions file to build a new release when a new tag is
  pushed in the main branch
- Add CLAUDE.md developer guide for AI-assisted development
- Show version number on the login page
- Frontend version number must follow the git tag automatically (no manual updates)
- Time-tracking project dropdown always shows all TT-only projects when a customer is selected, in addition to the customer's own regular projects
- Time-tracking report can be grouped by Period, Customer, Project, or Customer & Project; grouping applies to PDF and XLSX exports
- All projects visible to admins in the sidebar
- PDF time report must print only the report content (no sidebar, header, or footer)
- Time in reports displayed as H:MM (hours unbounded, minutes zero-padded)
- Default initial project column renamed from "Inbox" to "Backlog"
- Configurable initial columns per new project (admin setting, one name per line)
- Delete empty columns from the board view
- Assignee filter on time reports (select one, multiple, or all assignees)
- Direct message history loads from database when opening a conversation
- Remove a member from a group chat
- Demo conversations in the seed tool (4 DMs + 1 group chat with realistic message history)
- Screenshots of all main views, referenced in the README
- Project admin role (column management) separate from member and owner
- Board toolbar shows project name; settings gear visible only to admins
- Upload an avatar image for group chats
- Auto-delete empty group chat when last non-creator member is removed
- Start a group chat from a project team in Direct Messages (Teams tab)
- More demo users in seed (Priya, James, Elena, Raj) with project admin roles
- Persistent system admin (tonk) in seed, not affected by --reset
- Emoji picker in all chat inputs and card editors (EasyMDE toolbar button)
- @mention autocomplete in all chat inputs and card editors with dropdown navigation
- Real-time popup notifications for @mentions when the user is online; email for offline
- Chats sidebar section with per-conversation unread indicators
- SMTP test email in admin settings
- "Direct Messages" renamed to "Chats" throughout; /messages redirects to /chats
- Team Chat removed from project board page
- Card comments now trigger @mention notifications (email + real-time WS)
- Auto-replace text emoticons (e.g. :-) ;-) <3) with emoji in all editors and chat inputs
- Fix Escape key in card comment editor closing the card modal (first pass: stop propagation from the textarea)
- Fix unread indicator showing for messages the current user sent
- Git platform integration: connect GitHub, GitLab, Gitea, or Forgejo via webhooks
  - Push / PR / issue events post formatted messages to project chat
  - Card references (e.g. PRJ-42) in commit messages and PR/issue titles auto-link to cards
  - Linked events appear in a Git Links section on the card detail
- Documentation: user guide, API reference, and admin guide shipped with every release in docs/
- Reports restricted to project admins/owners and system admins; hidden from regular members and viewers
- Webhook setup docs show the real token in the URL immediately after creating a webhook
- Fix admin Settings tab blank (vue-i18n @ in SMTP placeholder)
- Fix JWT token lost on LocalStorage eviction (keep token in axios default header)
- Fix admin settings errors silently swallowed (now shown as toast)
- Fix SMTP password placeholder always shown due to truthy string "false"
- Fix Reports menu hidden for admins with stale cached user object
- Close / reopen cards: Close Card button in card detail; closed cards shown with strikethrough and muted opacity on board; closed cards appear in time reports with a "Closed" badge
- Fix due date on board cards showing wrong date in negative-UTC timezones (UTC vs local date slice)
- Due date field: replace native browser date picker with a plain text input that follows the user's configured date format, plus a clear button
- Spellcheck in card description, comments, and title (plain textarea replaces CodeMirror for editing; markdown preview unchanged)
- Auth tokens moved to sessionStorage so closing the browser ends the session
- Fix project switching in sidebar not reloading board content (watch route slug; useWebSocket accepts reactive ref)
- Due date calendar picker: add a calendar icon button (📅) alongside the text input (previous item) that triggers a hidden native `<input type="date">`; preserves the configured-format text display while also allowing picker-based input
- Fix default labels not automatically added to new projects created via Admin → Projects (AdminCreateProject was missing the getDefaultLabelDefs() seeding loop)
- Fix custom default columns not applied to new projects: replaced unreliable @change on textarea (destroyed by v-if before event fired) with an explicit Save button in the Project Defaults settings section
- Fix saveSetting not updating existing rows: replaced GORM clause.OnConflict upsert (silently failed to UPDATE for string PKs in SQLite) with explicit UPDATE + RowsAffected == 0 → CREATE pattern
- Demo seed tool now configures default system settings: 4 columns (Backlog, In Progress, Test & Review, To Production) and 4 labels (Bug, Feature, Design, Content)
- Fix board cards showing light background in dark mode (hard-coded #fff replaced with var(--color-surface)); priority badge colours now have explicit dark-mode overrides
- Open card count shown in Admin → Projects table and on dashboard project tiles
- Copy card: duplicate a card within the same column via "Copy Card" button in the card detail footer; title gets "(copy)" suffix; labels and tags are copied; board broadcasts in real time
- Transfer card: copy or move a card to any project via a "Transfer…" panel in card detail; choose destination project and column, then "Copy Here" or "Move Here"; labels/assignees are not copied (project-specific); board updates immediately for all connected users
- Fix report date/time not following configured format (was using toLocaleString; now uses useDateFormat composable)
- Fix report URL printed at bottom of page (suppressed via @page margin rules)
- Fix PDF export missing pages (overflow: hidden on shell body clipped print output; overridden in @media print)
- Fix print header duplicated/cut off across pages (position: fixed replaced with @page margin boxes); WarmDesk logo on page 1; "WarmDesk" + page number (n / total) in top margin on pages 2+
- Fix code blocks unreadable in dark mode: inline code background changed from hard-coded #f1f5f9 to var(--color-border) with explicit text colour; fenced code blocks (pre) styled with var(--color-bg)/var(--color-text) and a border; pre code resets background to transparent
- Fix desktop app cannot connect to server: add tauri-plugin-http so globalThis.fetch uses a native HTTP client that bypasses CORS (Windows Tauri origin https://tauri.localhost was blocked by server CORS policy)
- Fix blank screen on Linux desktop app: set WEBKIT_DISABLE_DMABUF_RENDERER=1 before Tauri starts to work around silent WebKitGTK DMA-BUF renderer failure on many GPU configurations
- CI: upgrade Node.js to 24 in GitHub Actions release workflow
- Add coworker-export: standalone binary to export a Coworker project to Jira, Trello, OpenProject, or Ryver (columns, cards, checklists, comments, labels, tags, time, attachments, topics)
- Add coworker-import: standalone binary to import a project from Jira, Trello, OpenProject, or Ryver into Coworker
- Config file (coworker-migrate.yaml) with column mapping, credential env var overrides, and interactive prompts for missing fields
- Both binaries included in the server distribution build
- Rename product from Coworker to WarmDesk: all binaries, config files, documentation, logos, and the application UI renamed; Go module path updated to github.com/tonk/warmdesk
- Rename migration tool config: YAML section key `coworker:` → `warmdesk:`, env vars `COWORKER_*` → `WARMDESK_*`, default config filename `coworker-migrate.yaml` → `warmdesk-migrate.yaml`, Go type `CoworkerConfig` → `WarmDeskConfig`
- Show full WarmDesk logo (logo-full.svg) in the app header instead of the icon-only mark
- Fix logo-full.svg (and any other non-listed static asset) returning index.html in production: register it explicitly as a static route in the backend router
- Update docs: add Migration Tools section to admin guide, fix header and editor descriptions in user guide, correct API key format example, list all dist binaries in INSTALL.md
- Exclude .claude/ directory from version control via .gitignore
- Resizable sidebar: drag the inner edge to set a custom width (150–480px), persisted in localStorage; handle moves to the correct edge when the sidebar is on the right
- App-wide zoom via Ctrl+/Ctrl-/Ctrl+0 (50%–200%, 10% steps), persisted in localStorage and restored on next load
- Fix Windows desktop app connection: install @tauri-apps/plugin-http JS package so window.fetch is patched at startup and all requests go through the native Rust HTTP client instead of WebView2 (which blocked them as mixed content)
- Fix desktop app Axios requests bypassing tauri-plugin-http: switch to fetch adapter in Tauri so Axios calls are also intercepted by the native client
- Fix GitHub Actions Go cache: point setup-go cache-dependency-path to backend/go.sum
- Opt GitHub Actions into Node.js 24 via FORCE_JAVASCRIPT_ACTIONS_TO_NODE24
- Fix Linux desktop app blank screen regression (v0.4.3 applied tauri-plugin-http fetch patch on all platforms; now Windows-only)
- Fix Windows desktop app connection: await plugin-http import before Vue mounts so Axios sees the patched fetch from the first request
- Fix Windows desktop app login 403: add `http://tauri.localhost` to CORS allow-list (actual Windows Tauri origin); disable HTTP/2 in tauri-plugin-http; send browser User-Agent to avoid WAF blocks; parse plain-string error response bodies
- Fix `allowed_origins: *` wildcard not working (was treated as literal string)
- Allow server URL to be changed from the login page in the desktop app ("Change" link next to current server)
- Show version number on the Connect screen in the desktop app
- Install window.fetch proxy via inline script in index.html so Tauri HTTP patch is active before any ES module fires a request
- CI: split manual desktop build into per-platform workflows; add manual server build; replace PowerShell version stamping with Node.js
- Database TLS for PostgreSQL and MySQL: db_tls_mode (disable/require/verify-ca/verify-full), db_tls_ca_cert, db_tls_cert, db_tls_key with DB_TLS_* env var overrides; mTLS (client certificate) supported
- Server TLS: set tls_cert and tls_key (or TLS_CERT/TLS_KEY env vars) to serve HTTPS directly without a reverse proxy
- Regenerate all desktop app icons from WarmDesk SVG (removed old Coworker branding from 32x32, 128x128, 128x128@2x, icon.png, icon.ico, icon.icns)
- Desktop app CLI flags: --version/-V prints version and exits; --maximized starts window maximised
- Fix Linux desktop app network error with webkit2gtk 4.1: route all HTTP/HTTPS fetch calls through tauri-plugin-http on all Tauri platforms; non-HTTP requests fall back to native fetch (also fixes blank screen from routing all requests through the plugin)
- Stamp Cargo.toml version from git tag alongside tauri.conf.json; make appimage/dmg/windows-installer targets stamp both files automatically
- Document AppImage build prerequisites (system libraries for Fedora and Ubuntu; Rust install) in INSTALL.md
- Fix Windows release CI: run version-stamping Node.js script under bash instead of PowerShell (PowerShell parsed the regex character class `[^"]*` as an array index and aborted)
- Project-scoped API keys: keys created in Project Settings → API Keys are locked to that project and rejected on any other; personal API keys in User Settings give full access across all projects
- Accept API keys on all authenticated endpoints, not just the Ticket API (X-API-Key header or ?api_key= query param)
- Add base_url config setting (BASE_URL env var) to set the correct host in Swagger UI so "Try it out" calls reach the right server
- Fix font family setting having no effect: load selected fonts (Inter, Roboto, Open Sans, Source Code Pro) from Google Fonts on demand
- Fix Open Sans and Source Code Pro showing wrong font: extract font name from CSS font-family stack before Google Fonts lookup
- Fix font size setting having no effect: change hardcoded font-size: 14px on button/input/textarea/select to inherit
- Code signing policy section added to README.md as required by SignPath Foundation OSS programme
- Bundle Inter, Roboto, Open Sans, Source Code Pro via @fontsource (no Google Fonts CDN); FreeSans, FreeSerif, FreeMono served from /fonts/ (woff files copied from FreeFont project)
- Fix Linux desktop app COLRv1 crash in webkit2gtk/Skia: set HardwareAccelerationPolicy::Never via with_webview API; also set WEBKIT_DISABLE_DMABUF_RENDERER=1 to prevent blank window on many GPU configurations
- Add Linux .desktop file (deploy/warmdesk.desktop) for system-wide installation; document installation steps in INSTALL.md
- Add Ctrl+Scroll mouse wheel zoom (alongside existing Ctrl+/Ctrl-/Ctrl+0 keyboard shortcuts)
- Temporarily disable Windows code signing in release CI (SignPath signing steps commented out)
- Show server version in footer alongside client version (fetched from new public GET /api/v1/version endpoint)
- Fix make appimage/dmg broken by non-semver git tags: pass --match 'v*' to git describe in Makefile
- Fix webhook setup URL showing tauri://localhost in desktop app: use configured server URL instead of window.location.origin for GitHub, GitLab, and Gitea payload URL display
- Fix Forgejo webhook showing Gitea logo on card Git Links: detect platform from X-Forgejo-Event header instead of webhook name
- Fix git links and release notes banner doing nothing when clicked in desktop app: open in system browser via tauri-plugin-opener
- Fix Escape key closing card modal even when a mention dropdown is open: check e.defaultPrevented before closing (second pass)
- Fix Cancel button discarding unsaved card changes silently: show Save / Discard / Back confirmation panel when the card is dirty; Escape also shows this panel when dirty, and closes immediately when the card is clean
- Typing indicator in project chat: animated three-dot indicator shows who is typing; auto-clears after 4 seconds of inactivity
- @mention autocomplete in card description and comment fields (same dropdown as in project chat)
- Prometheus metrics endpoint GET /api/v1/metrics: project count, column count per project, open/closed card count per column; protected by a new metrics global role
- Admin UI: metrics role option added to the role selector in user list and create user form
- Project ordering: admins can drag project tiles on the dashboard to reorder them; order persisted to database and shown to all users
- Fix Linux desktop app blank screen on Fedora / Wayland: at startup, automatically detect libwayland-client.so at well-known paths (Fedora /usr/lib64, Ubuntu /usr/lib/x86_64-linux-gnu, ARM64, etc.) and re-exec the binary once with LD_PRELOAD set; a sentinel env var (WARMDESK_PRELOAD_DONE) prevents infinite loops; no manual LD_PRELOAD configuration required
- Add Customer/Contract/Project hierarchy: customers are top-level entities; contracts sit under customers; projects can be linked to a customer and optionally to a contract within that customer; manage from a dedicated Customers view and from Project Settings
- Add Customers page (/customers): grid of customer tiles showing name, description, logo, contract count, project count; star/unstar favourites; admins and project owners can create/edit customers and contracts
- Add Customer detail page (/customers/:id): shows customer info, contracts with their projects grouped beneath, and unassigned projects; admins can create/edit/delete contracts; inline name editing for managers
- Add Customers section to sidebar: starred customers shown with star/unstar toggle; link to all customers
- Link projects to customers and contracts via Project Settings General tab; customer and contract dropdowns are filtered and linked
- Show customer name on dashboard project tiles when a project is linked to a customer
- Add sub-cards: cards can have child cards (one level deep) visible only in the parent card's detail view; parent cards show a progress pill (done/total) on the board; sub-cards are hidden from the board column; sub-cards have full card detail (comments, assignees, labels, etc.) accessible via open button; adding a sub-card creates a card number in the same project
- Add start_date field to cards: editable in card detail alongside due_date; stored in database; returned in all card API responses
- Add cross-references between cards: bidirectional links shown in card detail "Linked Cards" section with card number, project key, title, priority, and column; clicking opens linked card in nested detail modal; cross-project references supported
- Add Gantt chart per project: frappe-gantt based view accessible from board toolbar; cards with start/due dates shown as colour-coded priority bars; click to open card detail; drag to update dates; Day/Week/Month view modes
- Fix "Edit customer" showing empty fields: openEdit() now pre-populates form before opening dialog
- Fix switching customers in sidebar not updating overview: added watch(custId) to CustomerDetailView to reload on route param change
- Extend demo seed: start dates, due dates, sub-cards, and cross-reference links added to all demo card sets
- Sidebar drag-to-reorder starred projects and customers: drag handles appear on hover; custom order persisted in localStorage
- All Customers section in sidebar: collapsible list of every customer, starred ones first with a ★ badge; collapsed by default
- Fix card Start Date and Due Date pickers not opening: use overlay input[type=date] behind calendar icon button instead of showPicker() on hidden element
- Fix contract editor date fields ignoring configured date format: Start Date and End Date now use the text-input + overlay-picker pattern with format-aware display and parsing
- Fix deleted projects reappearing in admin list: remove Unscoped() from AdminListProjects so soft-deleted projects are excluded
- Fix deleted project remaining visible in sidebar until refresh: admin delete now immediately filters the project from sidebarStore.allProjects
- Fix seed --reset failing with UNIQUE constraint error when demo projects were previously soft-deleted: collect project IDs via Unscoped() before wiping
- Per-user accent colour setting (blue, red, green, orange): saved to user profile, applied on login, affects buttons, links, and active highlights across the full UI
- Drag-to-reorder sidebar sections: grab handle on each section header; default order is Starred Projects, All Projects, Favourite Customers, All Customers, Favourites, People, Chats; order persisted in localStorage
- Increase whitespace between sidebar sections for readability
- Indent items in All Projects, All Customers, People, and Chats sidebar sections; apply the same indent to empty-state messages throughout the sidebar
- Fix Windows desktop app typing lag during login: register the Ctrl+zoom keydown listener as passive in Tauri so WebView2 skips the synchronous IPC round-trip on every keystroke
- Fix sidebar drag-to-reorder broken on Linux desktop app: replace HTML5 Drag-and-Drop API with pointer events (pointerdown/pointermove/pointerup) because WebKitGTK's DnD support is incomplete; new implementation works on all platforms
- Forgotten password: Forgot password? link on login page sends a one-time reset link by email valid for one hour; requires SMTP; always responds 200 to prevent account enumeration
- Password policy: admin-configurable minimum length (floor 8, default 12), require uppercase, lowercase, digit, special character; enforced on registration, password change, and reset; active requirements shown to users beneath the password field
- Avatar image upload in User Settings: upload an image file directly instead of providing a URL; stored as an attachment on the server
- Raise default password minimum length from 8 to 12 in system settings defaults
- Seed: star all three demo projects and Acme Corporation + Globex Systems for the persistent tonk admin user
- Fix INSTALL.md Go requirement (1.22 → 1.25) and first-admin instructions (first registered user is auto-promoted, no DB update needed)
- Reduce Windows desktop app login-screen typing lag: hide WebView2 password-reveal button (`::-ms-reveal`), disable spellcheck/autocorrect/autocapitalize on credential inputs, and disable WebView2 autofill IPC via ICoreWebView2Settings4 in Tauri Rust setup
- Expand INSTALL.md desktop build prerequisites with per-platform details (Linux: Ubuntu 24.04 required for correct HarfBuzz; macOS: Xcode CLI tools + universal Rust targets; Windows: Go + Node.js + rustup-init.exe + WebView2 pre-installed)
- Allow the card prefix (e.g. PRJ in PRJ-42) to be set when creating a project; auto-generated from the name but fully editable; 1–10 uppercase letters or digits; live preview shown; forced uppercase on input; cannot be changed after creation
- PDF language selector in report export: choose output language (EN/NL/DE/FR/ES) independently of the UI language; Auto follows current locale; persisted in localStorage
- Per-project subtotal shown as white pill badge on the right of each project header bar in exported PDFs, matching the on-screen report
- Fix PDF export HTTP 500 when company logo is a PNG with alpha channel: composite over white before embedding
- Suppress WebSocket close-1005 log noise: normal browser navigation/tab-close event, not an error
- Customer access control: non-admin users only see customers they are explicitly assigned to (strict allowlist — no assignment means not visible, even if the user has no access rows at all); global admins always see everything
- Per-customer admin role: CustomerAccess rows carry a role field ("member" = read-only visibility, "admin" = can also manage contracts and the customer's member list); role is independent of the global role
- Customer member management: GET/PUT /customers/:id/members lets global admins and customer-admins manage who can see a customer and at what role; accessible from a Members section in the customer detail page; self-lockout protection (non-global-admin cannot remove their own admin row)
- Admin UI: customer chip picker in Create/Edit User now shows an M/A badge on each selected chip (click to toggle between member and admin); role is persisted alongside the assignment
- Training seeder (warmdesk-training): CustomerAccess rows created for each trainee (role member, restricted to their own customer); guru00 receives admin-role CustomerAccess for every training customer so they can see and manage all of them without global admin rights; idempotency path applies the same corrections when users already exist; DiceBear 9.x avataaars avatar seeded from username for each guru user
- Enforce unique card prefix (key_prefix) across all projects so card codes like PRJ-42 are unambiguous for GitHub/webhook integrations; uniqueIndex added to the model; auto-generator appends a numeric suffix when the base is taken (WAR → WAR2 → WAR3); duplicate prefixes in existing databases are deduplicated on startup before AutoMigrate runs; prefix is immutable after creation (changing it would invalidate existing commit messages and external refs); git integration regex updated to match digit-suffixed prefixes
- Ansible collection customer_member module: new ansilab.warmdesk.customer_member module manages GET/PUT /customers/:id/members; parameters: customer (name), username, role (member/admin), state (present/absent); full-list sync via PUT; resolves username to user_id via GET /users for new members; check mode aware
- Ansible collection user module: added customer_roles parameter — a dict of {customer_name: role} that performs a full sync of a user's customer assignments via PUT /admin/users/:id/customers; omit to leave assignments unchanged, pass {} to clear all; customer names resolved to IDs via GET /customers; unresolved names emit a warning; works in both CREATE and UPDATE paths
- About modal: nav dropdown "About" item opens a modal showing the frontend version (from __APP_VERSION__ Vite global), server version (fetched from GET /api/v1/version), description, repository link, license, and copyright; the modal fetches the server version itself on mount so it works without wiring through the store
- Resizable Edit User modal in Admin panel: the Edit User modal has resize: both enabled via a new .modal-resizable CSS class; flex column layout keeps the header and footer pinned while the body scrolls; min-width 400px, min-height 300px, max 95vw/90vh
- 7 new UI languages: Danish (da), Swedish (sv), Norwegian Bokmål (nb), Finnish (fi), Icelandic (is), Portuguese (pt), Italian (it); all ~350 translation keys covered; all languages available in the header language selector, the system default locale selector, and the per-user locale selector in Admin → Edit User
- All 12 languages now also available in the per-user locale selector in User Settings and the PDF language selector in the Report view (previously only 5 languages were shown there)
- Show deleted projects in Admin → Projects: a "Show deleted" toggle lists soft-deleted projects with a Deleted badge; a Restore button recovers a project from soft-delete
- Database backup in Admin panel: new Backup / Restore tab (4th tab, alongside Users, Projects, Settings); Create Backup button makes a timestamped copy in ./backups/ using VACUUM INTO for SQLite (atomic, no downtime) or pg_dump/mysqldump for Postgres/MySQL; backup list shows filename, size, and date with Restore and Delete actions; SQLite restore is live (close connection pool, overwrite file, reinit GORM — no server restart); every backup and restore is logged with filename, driver, user ID, and client IP
- Backup role: new global_role value "backup" (alongside admin/user/viewer/metrics); BackupAuth middleware allows admin + backup; dedicated endpoint POST /api/v1/backup triggers a backup via API key for cron jobs and automation scripts; backup role appears in the create-user and users-table role dropdowns in Admin
- Built-in backup scheduler: Admin → Backup / Restore tab; choose interval (every 6 h / 8 h / 12 h / 24 h); configurable retention count; runs server-side without cron; last run and next scheduled run displayed in UI; oldest files pruned automatically when retention limit is reached
- Backup start time: optional HH:MM anchor for the backup scheduler; when set, backups run at fixed time-of-day slots (e.g. start 02:00 + every 6 h → runs at 02:00, 08:00, 14:00, 20:00) instead of drifting with each run
- Package as .deb (Debian/Ubuntu) and .rpm (Fedora/RHEL): make deb and make rpm targets using Tauri's built-in bundlers; .desktop entry and icons included automatically
- Rename Ansible collection namespace from ansilab to ansilabnl
- Ansible Galaxy compliance: add README.md and meta/runtime.yml (requires_ansible: ">=2.14")
- Fix all remaining Ansible DOCUMENTATION YAML parse errors across all modules
- Backup file list: truncate long filenames with ellipsis, widen Size column, correct unit to kB
- Download backup: each backup in Admin → Backup / Restore has a Download button that streams the file to the browser for offsite storage or transfer
- Fix warmdesk-seed --reset crash on fresh database: startup key_prefix migration guard now checks HasTable before attempting ALTER TABLE, so a brand-new or just-wiped database initialises cleanly without "no such table: projects"
- Fix desktop app .desktop file incomplete: add custom Tauri desktop template (warmdesk.desktop) with GenericName, Comment, Keywords, StartupNotify; wire shortDescription and category in tauri.conf.json so Categories and Comment are populated in the installed .desktop entry
- Fix desktopTemplate placement in tauri.conf.json: must be under bundle.linux.deb and bundle.linux.rpm individually in Tauri 2, not bundle.linux directly
- Fix desktop app category: "Office" is not in Tauri's accepted list; changed to "Productivity" (maps to Office; XDG category in the .desktop file)
- Send an email notification after every backup (manual or scheduled): toggle and recipient address in Admin → Backup / Restore; email contains date/time, success/failure status, and a list of all available backups with sizes and dates
- Add backup metrics to Prometheus endpoint: warmdesk_backup_last_run_timestamp_seconds, warmdesk_backup_last_success (1/0/−1), warmdesk_backup_files_total
- Send all outbound emails as HTML with a shared branded template: blue header with company name and logo, content area, footer with WarmDesk icon, version, and instance URL; plain-text fallback included for all emails
- Add dark-mode email support: prefers-color-scheme media queries keep the password-reset button and key elements readable in Apple Mail, iOS Mail, Samsung Mail, and Outlook for Mac in dark mode; border fallback for clients that strip background colours
- Fix email footer showing vv0.7.7 (double v): strip leading v from version tag before rendering
- Fix Gin "trusted all proxies" warning: call SetTrustedProxies(nil) at startup so proxy headers are not trusted by default
- Config file auto-discovery accepts both .yaml and .yml extensions (warmdesk.yaml tried first, falls back to warmdesk.yml)
- Embed company logo as base64 data URI in outbound emails: uploaded logos read from disk, external URLs fetched at send time; works without base_url and survives email clients that block remote images
- Fix company logo not shown in emails when stored as a data URI: pass data: values through directly instead of discarding them
- Add Scrum Story Points: admin toggle enables a story points field on every card detail and a SP badge on card tiles; takes effect immediately without a page reload
- Add customer filter on the project dashboard: selector appears when projects span multiple customers; drag-to-reorder suppressed while filtered
- Auth audit logging: login, logout, failed login, registration, password change/reset, MFA enable/disable/verify events written to server log with user ID, username, and client IP
- Admin permanently delete project: hard-delete a soft-deleted project and all its data from Admin → Projects (deleted tab); safe — only available after soft-delete
- Four selectable chat layouts (Bubble, Comfortable, Compact, Cozy) in project chat panel and DM view; choice persisted in localStorage
- In-app new-message toast: dismissible pop-up in chat when a message arrives from someone else; controlled by bell toggle in chat header; persisted in localStorage
- OS desktop notifications: native OS notification when a message arrives and the window is not focused; uses browser Notifications API; respects the bell toggle
- Document-title unread indicator: tab title shows ● WarmDesk when there are unread messages, reverts when all read
- Fix backup email filename invisible on light backgrounds (missing text colour on filename column)
- Sprint date fields now respect user date format (overlay date-picker pattern)
- Fix false-positive unread DM indicator on every login (conversations seeded to current updated_at on first load)
- Fix unread dot reappearing for actively-read conversations (markConvSeen called on every poll when user is at bottom)
- Fix Admin permanently delete button doing nothing (missing useI18n import in AdminView.vue)
- Remove 480px max-width cap on DM message bubbles so wide content (ASCII tables, code) uses full available width
- Sidebar unread check split to dedicated 5-second interval; presence/user-list poll remains at 10 seconds
- Native desktop notifications in Tauri app via tauri-plugin-notification (libnotify on Linux); browser falls back to Web Notifications API
- trusted_proxies config setting (TRUSTED_PROXIES env var): comma-separated IPs/CIDRs for reverse-proxy trust so auth logs show real client IPs; documented in warmdesk.yaml.example
- Grouped chat layout (fifth option alongside Bubble/Comfortable/Compact/Cozy): consecutive messages from the same sender within 5 minutes collapse avatar and name to first message only, Discord/Mattermost style
- Paste images into chat: Ctrl+V in the compose textarea pastes clipboard images as pending attachments; local object-URL preview shown before send; works in DM view and project chat panel
- Inline image display: attached images render at up to 520×480 px directly in the message; attachment URLs include ?token= so <img> tags authenticate without custom headers
- Compose textarea retains keyboard focus after sending a message
- Card references in chat: typing #PRJ-42 renders a clickable badge; click navigates to the card; Ctrl/middle-click opens a new tab; backend resolves prefix+number via GET /api/v1/cards/resolve/:ref
- Fix Tauri .desktop file empty Name/Comment/StartupWMClass: Tauri v2 does not substitute {{app_name}} or {{short_description}} placeholders; values hardcoded in template
- Ansible collection project module supports board_type (kanban or scrum) at creation time; collection bumped to v0.2.0
- Board silently refreshes on WebSocket reconnect so cards created via API during a dropped connection appear automatically without a manual reload
- Fix favicon: replace CoWorker indigo Kanban-columns icon with WarmDesk desk-and-cup motif in green brand colours
- Prevent admin from deleting their own account: backend returns 403 and the Delete button is disabled in Admin → Users for the logged-in user's own row
- Ansible collection: new api_key lookup plugin lists personal or project-scoped API keys enriched with username and display_name; filter by key name or return all
- Hide Deactivate and Delete buttons for the logged-in admin's own row in Admin → Users (previously Delete was disabled but still visible; Deactivate was fully functional)
- Swagger UI: /swagger and /swagger/ redirect to /swagger/index.html so typing index.html is no longer needed
- Fix Swagger startup panic: remove /swagger/ route that conflicted with Gin's /*any wildcard; gin-swagger handles the trailing-slash redirect internally
- Restore Gravatar avatars with an admin toggle to enable/disable them; fall back to DiceBear initials avatars when disabled
- Rate-limit auth endpoints: 10 login/MFA attempts per 15 minutes, 5 registrations per hour, 5 password resets per 30 minutes
- Add security response headers (X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, Referrer-Policy)
- Write structured audit log lines for all authentication events
- Validate uploaded image files against magic bytes to reject MIME-type spoofing
- Fix login-page redirect loop: token-refresh interceptor no longer reloads the page when already on /login
- Pass PGPASSWORD via environment variable for pg_dump instead of embedding credentials in the DSN argument
- Add random hex suffix to backup filenames to prevent same-minute overwrites
- Issue httpOnly SameSite=Strict cookies for browser auth flows; add POST /auth/logout to clear them server-side
- Accept access-token cookie for WebSocket upgrades in browser mode
- Enforce allowed_origins list in the WebSocket upgrader
- Raise search minimum to 3 characters and strip SQL LIKE wildcards from queries
- Remove ?api_key= query param support; require X-API-Key header for API key authentication
- Log a warning when allowed_origins is set to wildcard (*)
- Fix Windows CI build: remove FORCE_JAVASCRIPT_ACTIONS_TO_NODE24 from all GitHub Actions workflows
- Add per-user time tracking module: optional toggle in User Settings; weekly grid (Exact Online-style) with customer/project/activity rows and day columns; week navigation; add, edit, and delete rows; day and row totals; grand total
- Time tracking report tab: period selector (week/month/year), grouped entries, subtotals, grand total
- Export weekly timesheet and report to XLSX (SheetJS) and PDF (backend gofpdf, all 12 languages)
- Move "Time Spent" from card form to per-comment: optional hours/minutes input when writing a comment; card shows running total; auto-creates a TimeEntry (today's date, card's project and customer, first sentence as description) when user has time tracking enabled
- Highlight overdue board cards with a subtle orange background and border tint (distinct from the red date text)
- Enforce customer on every project: required at creation (create-project dialog and API) and at update (project settings); blank option removed from project settings customer selector
- Re-check server and GitHub versions every hour while logged in (previously only on startup)
- Vite manualChunks bundle split: Vue core, i18n, markdown, HTTP libs moved to named chunks; main JS bundle reduced from ~504 kB to ~298 kB
- Demo seed: add 60 time entries across 5 users spanning 2 weeks; enable time tracking for all five active demo users
- Refuse to start if jwt_secret is still the default value or if allowed_origins contains wildcard in release mode
- Add HSTS and Content-Security-Policy headers to nginx and Apache reverse-proxy templates (for deployments behind a reverse proxy)
- Detect uploaded file MIME type from file bytes server-side; ignore client-supplied Content-Type
- Pin bcrypt password hashing cost to 12 (raised from library default of 10)
- Default gin_mode to release; require explicit opt-in for debug mode
- Add API IP allowlist system setting: restrict access to comma-separated IPs or CIDR ranges from the admin panel
- Apply IP allowlist to Swagger UI routes alongside the API
- Replace ?token= query param in WebSocket URLs and attachment image URLs with short-lived purpose-limited tickets (30 s WS ticket, 5 min media ticket) so the long-lived JWT never appears in server logs or browser history
- Sanitise handler error responses: replace raw err.Error() output with generic messages to prevent leaking internal paths or driver details
- Add drag-to-reorder for checklist items via ⠿ handle; new position saved immediately via a dedicated PATCH reorder endpoint
- Broadcast checklist changes (add, tick, edit, delete, reorder) over WebSocket so all users viewing the same card see updates in real time
- Show a ← [parent card title] back link at the top of sub-card and linked-card nested modals to return to the originating card
- Add ARM64 server build targets (make build-arm64) and ARM64 Tauri desktop targets (appimage-arm64, deb-arm64, rpm-arm64)
- Add make help target listing all build targets grouped by category
- Fix version label alignment on the login page
- Show WarmDesk wordmark next to logo on the login screen
- Fix company logo URL not resolving to absolute path in Tauri desktop client
- Admin Users table: add Last Password Change column and dedicated MFA column (moved from Status badge)
- Admin Settings: password_change_period_days policy; login flags expired passwords and redirects to Settings with a warning banner
- People list in DM/chat: non-admin users with explicit customer assignments only see colleagues in shared customers
- Admin tables (Users, Groups, Customers, Projects): sortable name column with ↑/↓ toggle
- Edit User modal in Admin: group assignment chip picker alongside existing project and customer pickers; shows and saves current group memberships
- Reaction pills in chat and DMs: hover tooltip shows names of all users who reacted ("You, Alice, Bob")
- Fix column drag-to-reorder for project owners/admins: Sortable is now enabled reactively once member list loads, not just at board init time
- F5 and Ctrl+F5 reload the page in the Tauri desktop client
- Fix IP allowlist: only apply to X-API-Key requests, not browser/JWT sessions; prevents admin self-lockout
- Fix training seeder --reset: was matching wrong group name pattern, leaving orphaned groups that crashed re-runs
- Fix training seeder group creation to be idempotent (FirstOrCreate)
- Expand training seeder character list: 21 new Monty Python / Arthurian characters, two duplicates removed
- Add Black theme (true #000 background) for OLED screens; available alongside Light, Dark, and System in Settings
- Fix dark mode native select dropdowns: declare color-scheme on :root and [data-theme="dark"] so browser-native dropdowns render correctly
- Fix stray text insertion cursor on UI chrome: user-select: none on body, restored only for inputs and textareas
- Fix blinking caret in login branding panel: re-focus login input after async branding loads and layout switches from plain to split
- Increase login branding panel logo max size from 180x120px to 240x180px for more visual presence with breathing room retained
- Demo seeder now seeds company_name, company_logo, and login_branding_enabled so the branded login panel is active after seeding
- About modal now mentions both Kanban and Scrum boards in the description
- Add ⋮ sections visibility menu to card detail: toggle Labels, Tags, Attachments, Checklist, Sub-cards, Linked Cards, Watchers on/off; sections hidden by default when empty; can only be turned off when empty; preferences persisted in localStorage
- Show avatars for both primary assignee and extra assignees on board card tiles (stacked, max 3 + overflow counter)
- Rename "Assignees" chip picker to "Extra Assignees" to distinguish from primary assignee dropdown
- Filter primary assignee out of Extra Assignees chip list; auto-remove from extra assignees when selected as primary
- Fix clearing primary assignee (selecting "—") not persisting on Save: backend was skipping null assignee_id; changed to json.RawMessage with explicit null handling matching the date field pattern
- Upgrade Vite 5 → 8 and @vitejs/plugin-vue 5 → 6; rename rollupOptions → rolldownOptions in vite.config.js
- Fix invalid :deep() usage in global CSS and double-:deep() chains in scoped components; eliminates Vite 8 / lightningcss warnings
- Bump postcss to 8.5.12 (resolves moderate XSS advisory GHSA-qx2v-qp2m-jg93)
- Fix WebSocket connections rejected for same-origin clients in production: allow connections from browsers where Origin matches the backend host
- Fix new cards not appearing on board until page refresh: board store adds the card immediately after API response; WS broadcast deduplication prevents double insertion
- Add separate light/dark company logos for login screen branding: admin sets two logos (light background and dark background); login page switches logo live when theme changes, including OS preference changes in system mode
- Show company logo and name in the Time Tracking report header (matches main Time Report style)
- Include company logo and name in time-tracking PDF export; resolve static frontend assets (e.g. /logo.svg) in addition to uploaded files
- Fix PDF export employee name: show the target employee's name (not the exporter's) when an admin exports for a specific person; show translated "All Employees" label when exporting across all users
- Fix sidebar indentation: "All Projects" and "All Customers" items now indent to the same level as starred/favourite items
- Fix time-tracking weekly sheet in dark/black themes: replace all hardcoded light colours with CSS custom properties
- Align time-tracking report columns across groups using table-layout: fixed and shared colgroup widths
- Fix backup email HTML rendering: row background colours moved from <tr> to each <td>/<th> with background-color so alternating-row shading renders in all email clients
- Git issue linking on cards: optional ExternalIssueURL and ExternalIssueRef fields on each card; reference auto-filled from URL path (/issues/, /pull/, /merge_requests/); toggled via card ⋮ sections menu
- Group team chat: every user group has a linked Conversation (IsGroup=true) created on save; membership kept in sync; rename/avatar propagated; deletion cleans up conversation; one-time startup migration for existing groups
- Sort card ⋮ sections menu options alphabetically by translated label
- PDF font and language selects in Time Tracking: weekly sheet export bar and Report tab filter bar both show PDF Font and PDF Language dropdowns (same options as main Time Report); persisted in localStorage; passed to backend PDF generator
- 1:1 audio and video calls via WebRTC: call button in DM header; peer-to-peer with STUN; signaling over existing WebSocket; audio-only fallback when camera unavailable; incoming call overlay with ringtone, accept/decline; full-screen video overlay with remote video + mirrored self-preview PiP + gradient controls bar; slim audio bar with live timer for voice-only calls
- Call settings dropdown: chevron next to call button opens floating panel with mic selector + live input-level bar, camera selector + live mirrored preview, speaker selector + test-tone button; device preferences persisted in localStorage; speaker section hidden when setSinkId unsupported (Linux Tauri)
- Fix ringtone playing when chatting: IncomingCallOverlay used onMounted to start ringtone but is always mounted in App.vue; changed to watch on call phase so ringtone only plays when phase becomes 'ringing'
- Demo seeder creates group conversations at seed time: each group gets its linked Conversation and ConversationMember rows during seeding, not relying on the startup migration
- Online presence in DM view: green dot on 1:1 conversation avatars + "Online/Offline" status line in chat header; uses sidebar store chatUsers (10s poll)
- Fix call hang-up race condition: startCall and acceptCall now check phase after await _getMedia so cancelling during permission prompt correctly aborts the call; 45s auto-cancel timeout added for unanswered outgoing calls
- Fix WebSocket reconnect robustness: Tauri mode retries after ticket-fetch failure; browser mode silently refreshes token before each reconnect to survive 15-minute JWT expiry
- Show error toast on call failure: "unavailable" (callee offline) and "no_mic" (microphone denied) errors now surface a toast instead of the call bar silently disappearing
- Fix chat layout: chat panel and Direct Messages view fill available height correctly with input anchored to bottom
- Fix nginx WebSocket Connection header: quote "upgrade" value per HTTP spec in provided nginx template
- Reduce presence poll interval from 10 s to 5 s for faster online/offline updates; extract as named constant
- Fix blank screen under CSP: add unsafe-eval to script-src in nginx/Apache templates; vue-i18n runtime message compiler requires new Function() and was blocked by the previous policy
- Fix update checker blocked by CSP: add https://api.github.com to connect-src in nginx/Apache templates
- Show toast when callee declines an outgoing call: "{name} declined the call" error toast appears instead of call bar silently disappearing; all 12 locales covered
- Fix call error toasts dropped by phase race: watch errorMsg directly instead of phase so declined/unavailable/no_mic toasts always fire even when ring timeout races the signal
- Add video toggle button to audio call bar: camera on/off button appears in the slim bottom bar when the call has a video track; matches mute button style and uses same toggleCamera action
- Fix call error toasts still not appearing: useI18n and useUIStore were missing imports in ActiveCallBar.vue so every toast attempt threw a silent runtime error; both now properly imported
- Add LiveKit token endpoint: GET /api/v1/conversations/:id/livekit-token issues a signed LiveKit JWT for the conversation room; membership-checked; returns 503 when livekit_api_key/secret not configured
- Add LiveKit config fields: livekit_url, livekit_api_key, livekit_api_secret with LIVEKIT_URL/LIVEKIT_API_KEY/LIVEKIT_API_SECRET env var overrides; documented in warmdesk.yaml.example
- Add in-call attendee invite: a + button in the bottom controls bar of any active video or audio call opens a searchable user picker with multi-select; selected users receive a real-time invite popup (IncomingGroupCallOverlay) and join the LiveKit group room with one click; backend adds invitees to the conversation and relays call.group_invite via personal WebSocket
- Upgrade 1:1 call to group when inviting: clicking + during a 1:1 WebRTC call sends the group invite to selected users and to the existing call partner, ends the WebRTC call, and joins the LiveKit room automatically
- Fix Apache WebSocket proxy timeout: replace RewriteRule-based WebSocket handling with a dedicated <Location "/api/v1/ws"> block using ProxyPass ws:// and ProxyTimeout 86400; prevents silent 5-minute disconnects under default Apache settings
- Fix backup email size unit: display kB (correct SI prefix, lowercase k) instead of KB; add white-space:nowrap to size and date columns so narrow email clients do not wrap "548.0 kB" across two lines
- Add "All Customers" collapsible section to sidebar: lists every customer with starred ones sorted first and marked; mirrors the existing "All Projects" section
- Add drag-to-reorder for starred customers in sidebar via ⠿ handle (pointer events — works on Linux/WebKitGTK); order persisted in localStorage
- Add livekit_room_prefix config option (LIVEKIT_ROOM_PREFIX env var): prepends a prefix to every LiveKit room name to avoid collisions when sharing a server across multiple WarmDesk instances
- Harden external image proxy: resolve DNS once and verify every IP is a publicly routable address, then dial by IP directly (prevents DNS-rebinding attacks); re-encode query strings so reserved characters in parameters (e.g. DiceBear seed names with spaces or commas) do not produce malformed requests; validate Content-Type header and reject non-image responses
- Rate-limit external image proxy endpoint: 200 requests per 10 minutes per IP to prevent unauthenticated bandwidth abuse
- Tauri desktop: load external images (Gravatar, DiceBear, etc.) directly instead of via the same-origin proxy; WebKit's tauri:// origin treats proxied HTTP responses as mixed content and blocks them
- Enable video call button in all group chats regardless of member count; remove the previous 3-member minimum
- Fix Linux AppImage camera and microphone: bundle all GStreamer 1.24 plugins and set GST_PLUGIN_SYSTEM_PATH_1_0 and GST_REGISTRY so the AppImage's plugins take precedence over incompatible system GStreamer installations on Fedora and other non-Ubuntu distributions
- Use camera icon for 1:1 call button to match the group chat button
- Fix backup email filename wrapping and raw timestamp: long filenames no longer wrap mid-word; backup date displays as readable date/time instead of a Unix timestamp
- Fix backup email filename wrapping in HTML viewers: wrap filenames in <nobr> so w3m, lynx, and Mutt's HTML viewer no longer break the filename mid-hash at narrow terminal widths
- Accept 'Name <address>' format in SMTP From setting: smtp_from now accepts RFC 5322 display-name form in addition to plain addresses; envelope address extracted automatically for smtp.SendMail
- Use monospace font for all columns in the Available backups email table: Size and Date columns now match the Filename column; header row also monospace
- GPG-sign all release artifacts: every file attached to a GitHub release has a companion detached armoured .asc signature; signing is conditional on GPG_SIGNING_KEY and GPG_KEY_ID secrets being set in the repository
- Fix GPG signing in headless CI: configure pinentry-mode loopback and restart gpg-agent after import so signing works without a TTY
- Fix GPG signing batchmode passphrase error: pass --passphrase from GPG_PASSPHRASE secret so keys with a passphrase work without interactive prompting
- Fix external avatar 502 errors in web app: media proxy returns 302 redirect to original URL when server cannot reach upstream host instead of 502; browsers follow redirect and load image directly
- Fix CSP font-src blocking inlined fonts: add data: to font-src in nginx and Apache deploy templates so Vite-bundled woff2/woff data-URI fonts are allowed
- Fix CSP img-src blocking external avatars: add https: to img-src in nginx and Apache deploy templates so Gravatar, DiceBear, and other external avatar images load when the proxy redirects to them
- Fix customer logo proxy 400 when hostname resolves to internal IP: return 302 redirect to original URL instead of 400 when resolveAndVerify rejects a host; browser loads the image directly, server-side SSRF protection still holds
- Show "no one online" error for group video call: clicking the group call button when no other members are online now shows a toast message instead of silently joining an empty room
- App: replace theme cycle button in app navbar with a Light / Dark / System dropdown; active option highlighted; persisted to user profile
- Close user menu immediately when any item is selected instead of requiring an extra click elsewhere to dismiss it
- Fix PDF and XLSX report export in the Tauri desktop app: open a native OS save dialog using tauri-plugin-dialog and tauri-plugin-fs; the previous anchor-download approach was silently ignored by the WebView
- Close all popups and overlays with Escape: About dialog, Call Settings panel, Emoji reaction picker, and incoming call / group call notifications; the call chat sidebar shows an inline discard-draft confirmation before closing when a message has been typed
- Website: add Light / Dark / System theme switcher to marketing website navbar; dark mode CSS variables applied throughout; system mode follows OS preference; choice persisted in localStorage

- Allow users and admins to create time-tracking-only projects (no board created) for use in the weekly timesheet
- Allow users and admins to create time-tracking-only customers (no CRM record created) that can be associated with time-tracking-only projects
- Manage time-tracking projects and customers via a tabbed modal (⚙ button) in the time-tracking view with full CRUD support
- All UI must be WCAG 2.1 AA compliant
- Group time-tracking report entries by Period, Customer, Project, or Customer & Project via a selector in the report filters; grouping is applied in PDF and XLSX exports
- Copy previous week rows to the current week in the Log Time tab with a single button; only the row definitions are copied (no hours), and copying is scoped to the displayed week
- Add a Today button to the week navigator to jump back to the current week in one click
- Sort Log Time rows by Customer/Project or Activity by clicking the column header; clicking again toggles direction; the day columns always follow their row
- Add a "Remember me" checkbox to the login page that saves the email/username in the browser's local storage and restores it on the next visit
- Add passkey (WebAuthn) sign-in: register passkeys (Touch ID, Windows Hello, hardware security keys) in User Settings and sign in passwordlessly from the login page using discoverable credentials; browser-only (Tauri desktop excluded)
- Add time notation user setting: choose between decimal (8.25) and HH:MM (8:15) for the weekly timesheet; preference saved to user profile; all calculations remain in minutes
- Fix time-tracking PDF: "Date" column header now translated in all 12 languages; month names in period labels translated; DMY date order used for all non-English locales
- Fix desktop app time-tracking PDF and XLSX export: all four export buttons now open a native OS save dialog via tauri-plugin-dialog and tauri-plugin-fs; previously silently dropped (XLSX) or errored (PDF)
- Fix desktop app PDF/XLSX export error: switch binary download endpoints from responseType:'blob' to responseType:'arraybuffer' to avoid unreliable response.blob() in Tauri's HTTP plugin
- Fix desktop app PDF and XLSX export in Linux AppImage: WebKit GTK2 throws TypeError on all response body methods (arrayBuffer, text, blob, getReader) for ReadableStream-backed Responses produced by tauri-plugin-http; bypass WebKit entirely with a new fetch_binary_b64 Rust command that fetches via reqwest and returns base64; JS decodes with atob()
- Fix desktop app save dialog default path: open in user home directory instead of AppImage mount path
- Desktop app remembers last export directory: save dialog opens in the last-used export folder (stored in localStorage); falls back to home directory on first use
- Fix attachment IDOR: verify project membership or conversation participation before serving any attachment download
- Fix Content-Disposition header injection: escape quotes and backslashes in attachment filenames before inserting into the header
- Add HSTS and Content-Security-Policy response headers to every Go server response (complements the proxy-template headers above; ensures headers are present when WarmDesk is accessed directly without a reverse proxy)
- Block CORS wildcard origin at middleware level in addition to startup check
- Enforce minimum 32-character length for jwt_secret at startup
- Rate-limit direct-message and conversation message-send endpoints (60 req/min per IP)
- Mask password-reset token in audit log: log only the first 8 characters followed by ellipsis
- Make all icon-only buttons WCAG 2.1 AA compliant with aria-label across board, card detail, chat, call, time-tracking, dashboard, admin, and layout components
- Add aria-label to all unlabelled inputs and selects
- Add aria-hidden to decorative SVGs throughout the frontend
- Add full ARIA tablist/tab/tabpanel roles to time-tracking mode tabs and chat layout picker
- Make hover-only interactive elements keyboard-accessible via focus-visible and focus-within CSS rules
- Replace hard-coded colour values with CSS custom properties in frontend components
- Fix time-tracking sheet row sorting: rows no longer re-sort while entering time; order is stable and only changes when the user clicks a column header
- Sort personal time-tracking projects alphabetically by name in the manage modal
- Add quick-nav strip to the app header with links to Dashboard, News, Chats, Reports, and Time Tracking; active page highlighted; permission-gated items hidden for users without access; strip hidden below 960px viewport width
- Add undeclarable minutes field to time-tracking-only projects: set in the ⚙ manage modal as HH:MM; automatically subtracted from totals in the weekly sheet (row badge, sheet footer), report tab, PDF, and XLSX export; result is never below zero
- Show undeclarable and declarable rows in time-tracking report when grouped by Customer: per-customer undeclarable line below each group subtotal, grand total undeclarable line at the bottom
- Add PDF page-break per customer option: checkbox in export controls (visible when group_by=customer); each customer page repeats the full document header (logo, company, title, period, employee); grand total omitted in page-per-customer mode
- Add day-of-week abbreviation option to PDF time-tracking export: prepend localised day abbreviation (Mon–Sun) before each date; abbreviation rendered in a fixed-width sub-cell so dates stay aligned for all 12 supported languages
- Group PDF export options in a dropdown button (Export options) with all settings persisted to localStorage; day-of-week toggle and per-customer page-break toggle shown together
- Add personal TT-only seed data for tonk: customers Smart Owl Consulting and Personal, projects Travel (45 min undeclarable), Holidays (480 min), Study & Training, Internal (60 min), with 14 sample time entries; --reset cleans them up
- Fix time-tracking week header date mismatch in UTC+ timezones: use local date components instead of toISOString() so column dates and day abbreviations always agree
- Add public holidays to time tracking: dropdown in the week navigation bar lists 12 countries by flag and native name; clicking a country adds 0-minute holiday entries for the current year; skip existing entries and report added vs. skipped count
- Highlight holiday entries in the sheet: amber background on day column headers (with glowing dot), amber row background, and stronger amber on the specific day cell
- Fix backup download in Tauri desktop app: use fetchBinary + triggerDownload so the native save dialog opens correctly on Linux
- Leave time entry cells blank when value is 0:00 to reduce visual noise; placeholder still shows on focus
- Replace emoji flag characters in holidays dropdown with CSS gradient badges approximating each country's flag colours; works on Linux without a colour emoji font
- Add calendar date-picker to week label in time tracking: clicking "Week 14 2026" opens a month calendar popover with week numbers; click any day or week number to jump to that week
- Fix time entry cell placeholder rendering on Tauri/GTK WebKit: hide placeholder by default, show faintly on focus only
- Update all dependencies to latest: Go 1.26, Vue 3.5.34, Pinia 3, Vue Router 5, vue-i18n 11, marked 18, Vite 8, Tauri 2.11, and all npm packages
- Speed up AppImage builds with sccache (Rust compilation cache) and mold linker
- Fix CI build pipeline broken after dependency upgrade: install clang and mold on GitHub Actions runners, suppress sccache wrapper (not installed on runners) with RUSTC_WRAPPER="" in all client and release workflows
- Add week start day preference to time tracking: users can choose Monday or Sunday as the first day of the week in Settings; drives the sheet view, week-picker calendar, current-week detection, and XLSX/PDF exports
- Fix week start preference not persisting: wire week_start into the PUT /auth/me handler so the setting is actually saved
- Add card_ref to ansilabnl.warmdesk.card return value: the full card reference (e.g. GF00-4) is now returned so it can be passed as card_number in follow-up tasks
- Fix Ansible assignee lookup with project-scoped API keys: fall back to GET /projects/{slug}/members on 403; accept username or email address as assignee value
- Add closed parameter to ansilabnl.warmdesk.card Ansible module: set closed: true/false to open or close a card; idempotent
- Add card activity history panel to card detail: full audit timeline of created, commented, field changes, column moves, and open/close events with timestamps and user attribution
- Add ansilabnl.warmdesk.card_comment Ansible module: create, update, or delete card comments identified by project slug and card number
- Fix media proxy panic when upstream hostname resolves to an IPv6 address: bracket IPv6 literals in dial-target URL and handle request-construction errors
- Proxy /uploads through Vite dev server so project avatar images load correctly during local development
- Add maximize/restore button to card detail modal: fills the full viewport on click; second click restores previous size
- Fix card modal clipped at high browser zoom: make modal backdrop the scroll container so the full card is always reachable
- Ansible card module: added `time_spent` option to log time (in minutes) against a card
- Ansible card_comment module: renamed `time_spent_minutes` to `time_spent` for consistency
- Ansible collection: new `user_options` module for managing per-user preferences
- Admin API: `PUT /api/v1/admin/users/:id` now accepts `time_tracking_enabled`, `theme`, `show_breadcrumbs`, `email_notifications`
- Add Time Tracking tab to Admin panel for managing time-tracking-only projects and customers with full CRUD
- Replace inline time-slot list on contract tiles with a compact 🕒 N badge and hover popup
- Enable passkey login on Tauri desktop app
- Add Playwright screenshot automation: 18 reference screenshots for documentation
- Add Delete/Backspace key to clear selected time tracking cells
- Add Time Tracking section to user guide and Admin Time Tracking subsection to admin guide
- Document contract time-slot compact badge display and hover popup in user guide
- Fix CI test suite: exclude e2e/ directory from vitest runner so Playwright screenshot specs don't cause false failures
- Add helpdesk ticketing module: tickets with type (incident/problem/service request/change request), priority, status (open/in-progress/resolved/closed/pending), assignee, owner, internal messages with attachments, tags, linked tickets, linked board cards, and a pending reminder date
- Add SLA policies: admin-configurable response and resolution time limits with optional priority filter; auto-applied to new tickets; breach status tracked per ticket
- Custom themed DatePicker component for date fields (respects user date format and week-start, works in all themes)
- User setting: dashboard default (boards or tickets); ticket option redirects to first starred customer's ticket list
- News widget shown on ticket list page so helpdesk-first users don't miss announcements
- Demo seed: add direct CustomerAccess entries for all demo customers so ticket assignee dropdown is populated
- Playwright screenshots 21 (ticket list) and 22 (ticket detail) added to automated capture suite
- Fix markdown list indentation missing in ticket/news/dashboard views (moved .markdown-body ul/ol styles to global CSS)
- Ansible collection v0.4.0: new ticket, sla_policy, and user_access modules
- Add time logging on tickets: users can log time entries directly on a ticket (gated by time_tracking_enabled), with a project selector populated from the customer's contracts and unassigned projects
- Fix Playwright screenshot suite: dismiss welcome/news backdrop before every screenshot (not just before clicks), change test.describe.serial to test.describe so one failure does not abort the whole suite, add dismissWelcome() helper used consistently across all 20 tests
- Update website homepage: add Helpdesk & Ticketing feature card, add ticket list and ticket detail screenshots to the gallery, update hero subtitle to mention helpdesk ticketing with SLA
- Add Hugo blog post about the helpdesk ticketing system with screenshots
- Fix release workflow: replace softprops/action-gh-release with gh CLI commands (gh release create/upload/edit) to eliminate dependency on codeload.github.com; add workflow_dispatch trigger with tag input for manual re-runs; add RELEASE_TAG env for consistent tag resolution across push and dispatch events
- Fix settings page crash in Linux RPM (WebKit GTK): wrap loadPersonalKeys() and revokePersonalKey() in try/catch — unhandled promise rejections crash the WebKit GTK WebView; guard navigator.clipboard calls in copyPersonalKey() (SettingsView) and copyKey()/copyWebhookToken() (ProjectSettingsView) with fallback error message for when clipboard API is unavailable
- Fix locale not saved for 7 languages: backend validLocales map only accepted en/nl/de/fr/es; now accepts all 12 supported locales (da/sv/nb/fi/is/pt/it)
- Fix black theme not persisted: backend theme validation now accepts 'black' alongside light/dark/system
- Add helpdesk ticket and SLA policy endpoints to Bruno and Postman collections: 14 Bruno requests in docs/bruno/helpdesk/ (list, create, get, update, delete ticket; add message; add/remove tags; list/create/delete ticket links; list/create/delete card links) and 4 in docs/bruno/helpdesk-admin/ (SLA policy CRUD); both Postman environment files updated with customerId, ticketId, tagId, linkId, slaPolicyId variables
- Add duplicate row button in time tracking: ⧉ clones a row with auto-incrementing (copy)/(copy 2) suffix, inserted directly below the original
- Add drag handles to time tracking rows: grab the ⠿ handle to manually reorder rows; order persisted in localStorage and restored across sessions
- Fix time tracking rows jumping: entryRows computed now sorts by minimum entry ID ascending (creation order) instead of following the API's date-desc ordering; editing a duplicated row's description/project preserves its position in _keyOrder instead of jumping to the bottom
- Fix Backspace in time tracking cells: only clears the cell when all text is selected (initial focus state); otherwise deletes one character, matching spreadsheet behaviour
- Fix sidebar resize handle blocked by overflow clipping: move overflow-y:auto from .app-sidebar to an inner .sidebar-scroll wrapper so the handle at right:-3px is never clipped and the scrollbar no longer covers it
- Style ticket group count badge to match board column card-count pill (--color-primary, 9999px radius, 11px/600)
- UI consistency pass: replace all toLocaleDateString() calls with useDateFormat() in chat, DM view, charts, and time-tracking calendar; unify call accept/decline button colours (--color-success/--color-danger) across 1:1 and group overlays; fix toast-success to use --color-success; sync board-type badge colours between Dashboard and Project Settings; unify SLA badge padding; move btn-warning to global CSS; remove redundant per-view .btn overrides; standardise page h1 to 22px; replace all 37 window.confirm() calls with a themed accessible ConfirmDialog
- Add leave/remove conversation from chat list: hover any conversation to reveal a ✕ button; for 1-on-1 chats deletes the conversation for both parties, for group chats removes only the current user; backend DELETE /conversations/:id endpoint with WS broadcasts to remaining members
- Fix Escape key not closing week picker and holidays dropdown in time tracking: add onNavEscapeKey document handler that closes both popups on Escape
- Add "What's new" strip to website homepage: a tinted band between hero and features showing current release highlights pulled from hugo.toml release_highlights param; updated each release alongside warmdesk_version and release_date
- Fix time tracking row drag-and-drop stopping after week navigation: watch tbodyEl template ref and destroy/recreate Sortable whenever Vue replaces the tbody element on week load
- Merge Board report into Time Tracking as third tab: extract ReportView into BoardReportPanel.vue, embed as "Board" tab in TimeTrackingView, redirect /reports → /time-tracking?tab=board-report, remove duplicate Reports nav entry from AppHeader
- Add PDF export toggles for page numbers and undeclarable time: show_page_numbers and show_undeclarable query params (default on), checkboxes in export options panel, persisted in localStorage
- Move PDF font and language selects into export options popover to prevent filter bar wrapping at HD resolution
- Add helpdesk inbox queue for unassigned email-created tickets: /tickets/inbox route with sidebar badge showing unread/total count
- Add IMAP polling service: poll a configured mailbox on a configurable interval, create inbox tickets from incoming email, move processed messages to a configurable "Processed" mailbox
- Add IMAP outbound reply: when an agent sends a message on a ticket that originated from email, send the reply back to the customer via SMTP
- Add IMAP OAuth2 authentication: support XOAUTH2 and OAUTHBEARER SASL mechanisms for Gmail and Office 365 mailboxes; store and auto-renew refresh tokens; admin UI in Settings → Incoming Mail
- Add reply threading: match incoming email replies to existing tickets via In-Reply-To, References headers, subject [#N] tag, or X-WarmDesk-Ticket-Id header; reopen closed/resolved tickets on customer reply
- Add email indicators: show sender name and address on inbox tickets; show ✉ badge on messages that triggered an outbound email reply
- Add move ticket between customers: reassign an inbox ticket to a customer, or move an existing ticket to a different customer from the ticket detail view
- Fix IMAP test connection to use current form values instead of last saved settings
- Add real-time inbox refresh: broadcast ticket.created and ticket.message.added WebSocket events from the IMAP service; InboxView reloads its list and TicketDetailView refreshes incoming replies in-place without clearing the reply draft
- Add macro system for helpdesk tickets: create reusable macros with set_status, set_priority, set_type, add_tag, add_message actions; apply from dropdown in ticket detail; admin CRUD in Settings → Macros tab
- Add macro placeholder expansion: add_message actions support {email}, {fname}, {name}, {subject}, {ticket_id}, {agent}, {agent_fname} with clickable insertion chips in the macro editor; applied macros pre-fill the reply box instead of posting immediately
- Add spam handling: mark ticket as spam sets is_spam=true and closes it; spam tickets hidden from lists by default; Show/Hide Spam toggle in ticket list and inbox headers; Not Spam restores status to open
- Add card/group/list views to Inbox matching the customer ticket list layout, with independent localStorage persistence
- Enlarge reply box to 8 rows (min 180 px) and add a Cancel button to clear the draft and pending attachments
- Add ticket checklist templates for helpdesk tickets: admins define reusable ordered item lists in Admin → Checklists tab; agents apply a template to a ticket in one click; tickets are blocked from closing or moving to pending_close until all items are checked off; progress bar with n/m counter and drag-to-reorder shown in ticket detail
- Fix Ansible macro module YAML scanner error on galaxy import: quote description bullets containing colon-space patterns
- Highlight today's column in the time tracking grid with a primary-colour tint on the header and a box-shadow overlay on data cells
- Fix customer access strict allowlist: non-admin users with no CustomerAccess rows see no customers; requireCustomerAccess now checks group-based access via getAccessibleCustomerRoles
- Fix time tracking footer colspan from 3 to 4 so day totals in Total/Undeclarable/Declarable rows align correctly with the day column headers
- Fix time tracking tab bar: keep Log Time/Report/Board tabs right-aligned with the gear icon at the far right via margin-left:auto on the tab group
- Fix backup Start time input alignment and time tracking export controls bottom-alignment
- Add get_warmdesk update script to server tarball: auto-detects architecture (amd64/arm64), downloads latest release from GitHub, stops/starts systemd service, uses $(id -u) root check and then/do on own lines
- Fix time range popup clipping: flip to open downward when trigger cell is within 200px of the top of the viewport
- Sync admin guide, user guide, release.md, repository-layout.md, and api.md with current feature set
- Add Grid PDF export for time tracking: landscape A4 grids for week (7 day columns), month (31 day columns, 5.5pt font, greyed past-month days), and year (12 month + 4 quarter columns); Grid PDF dropdown in Log Time export bar
- Add holiday cell indicators to grid PDFs: amber background, bullet '•' for 0-minute holiday entries
- Add weekend column highlights to month grid PDF: light gray Sat/Sun columns in header and data rows
- Fix Dutch 'no customer' translation from 'Geen klant' to 'Zonder klant' in all three i18n contexts
- Change PDF customer subtotal row label format from '<Customer> Total' to '<Customer> - Total'
- Fix grid PDF and PDF Options dropdowns opening upward instead of downward
- Fix time range popup flip to use scroll container position instead of viewport position
- Fix year grid column header row overlapping first data row (Ln(0) → ln=1 on last header cell)
- Fix month grid header label removing 'Jaar' prefix: now shows 'mei 2026' not 'mei Jaar 2026'
- Fix month grid day column index normalisation to UTC midnight to prevent timezone-related offset
- Fix grid PDF row label to fall back to activity description when no customer or project is linked
- Replace Grid PDF dropdown with a period-picker panel: segmented Week/Month/Year control with date selectors (week number + year, month name + year, or year alone) initialised from the current view
- Add vertical grid lines to all data rows in grid PDFs and a horizontal border below the totals row across week, month, and year grids
- Fix timesheet new-row editor missing drag-handle cell causing column shift when adding a row
- Fix PDF export options panel in Report tab opening upward and being clipped; now opens downward
- Add scripts/inject_time.py: bulk time-entry injection CLI with customer/project lookup, MFA support, dry-run mode, and time-based token refresh
- Fix settings view blank screen: escape bare @ in req_special i18n string to {'@'} in all 12 languages so vue-i18n does not treat it as a linked-message reference
- Fix grid PDF customer/project label columns: widen year (60→74 mm) and week (60→74 mm) using available page space; month computes label width dynamically from days in the month; remove phantom empty columns for short months; recalibrate truncation limits from measured font metrics
- Add ticket viewers: avatar row at the bottom of every ticket (customer and inbox) showing who viewed it and when; upsert via ON CONFLICT on (ticket_id, user_id) so each user appears once with their last view time
- Fix date/time format consistency: replace native <input type="date/time"> and raw ISO interpolations in board report, time-tracking personal report, Charts release picker, and contract slot time inputs with useDateFormat()-based equivalents
- Fix backup schedule start time: replace native <input type="time"> with text input using parse-and-reformat logic; also fix backupLastRun/backupNextRun computed names that were mismatched with template references
- Redesign SLA policies admin form: replace cramped inline table-row edit with card-above-table layout matching the Macros tab (labeled rows, explicit column widths, Save/Cancel footer)
- Add live search boxes to Admin Users, Groups, Customers, and Projects tabs; add "Show inactive" toggle to Users tab (inactive hidden by default)
- Rename "Create User"/"Create Group" buttons to "New User"/"New Group" across all 12 locales to match "New Customer"/"New Project"
- Fix admin panel hardcoded Status, Filename, Size column headers: add common.status/filename/file_size i18n keys to all 12 locales
- Seed 8 previously empty feature areas: project chat messages, ticket tags, ticket-to-ticket links, ticket checklist items applied to tickets, ticket history entries, demo attachment records, emoji reactions on messages, project webhooks
- Fix seed --reset to delete TicketView rows for both customer and inbox tickets
- Add customer global role: view and comment on assigned customers' tickets only; blocked from boards, chat, and time tracking
- Add private ticket messages: internal notes not emailed to the ticket sender and hidden from customer-role users; shown with amber highlight and lock badge
- Expand Prometheus metrics: add warmdesk_users_total, warmdesk_customers_total, warmdesk_tickets_total, warmdesk_tickets_by_priority_total, warmdesk_sla_breaches_total, warmdesk_ticket_messages_total
- Add docs/prometheus.yml Prometheus scrape config and docs/grafana-dashboard.json Grafana dashboard covering all metrics
- Add seed demo accounts for metrics role (demo.metrics) and customer portal (demo.cust1 Alice Porter, demo.cust2 Bob Mason)
- Add Epics layer to Scrum projects: colour-coded milestones that group cards across sprints, with CRUD API, EpicsView (create/edit/reorder/expand), card detail Epic selector, board colour bar and badge, backlog epic filter
- Add Epic Burndown chart: remaining cards and story points per day from epic creation to today with ideal line; select any epic from dropdown
- Add Sprint Report chart: post-sprint completed vs not-completed breakdown with story-point totals and completion percentage
- Enhance Cycle Time (Control Chart) with 7-day rolling average line
- Add drag-to-reorder for Product Backlog cards (persisted via PATCH /backlog/reorder)
- Add drag-to-reorder for sprint list with ascending/descending sort toggle (△▽ button)
- Add drag-to-reorder for macro actions and checklist template items in admin editors
- Add click-to-edit names in all admin panel list views (Users, Groups, Projects, News, TT projects/customers)
- Detect desktop app installation method (AppImage/deb/rpm/portable/dmg/windows) using /etc/os-release OS family + dpkg/rpm ownership query; show platform-specific download URL in update banner
- Widen ticket detail view max-width from 900 px to 1400 px
- Consolidate sort indicators to outline triangles (△/▽) throughout the UI
- Fix 28 English i18n button/action labels to standard title case
- Unify admin panel button styles to btn-ghost/btn-ghost btn-danger throughout
- Persist empty timesheet rows in the Log Time grid per ISO week; do not remove blank rows when all hours are cleared
- Copy Previous Week includes empty rows from the previous week in the same order
- Week grid PDF shows start/end times on cells; overnight standby shifts render as separate two-day spans with connector lines
- Move Time Tracking add buttons to toolbar; fix News empty state; fix TicketChecklistTemplatesTab to use ui.confirm()
- Apply consistent grid styling (borders, header background, cell padding) to admin SLA Policies, Macros, and Checklists tabs matching the Time Tracking tab
- Rename "Add Project" / "Add Customer" to "New Project" / "New Customer" across all 12 locales for UI consistency
- Add spam count badge on Show Spam button in ticket list and inbox views; always fetch full ticket list including spam and filter client-side
- Add two demo spam tickets to the seed (phishing email, SEO pitch)
- Show explicit per-status sections (New / Open / Pending / Closed) in card and list views for both inbox and customer ticket pages; section dividers use accent colour
- Add custom accent colour picker (rainbow swatch) in Settings; supports any hex colour beyond the four presets
- Show undeclarable time inline in weekly timesheet grid cells: red amount below logged time, footer total row shows net declarable with undeclarable deducted in red
- Show declarable vs. undeclarable in week/month/year grid PDF exports: declarable as primary value per cell, red undeclarable sub-row below totals; per-cell undeclarable labels in week PDF
- Add multi-profile support to the Tauri desktop app: each profile has its own isolated WebView data directory (localStorage, cookies, session); manage profiles with --create-profile <name> [--label <label>], --list-profiles, --set-default <name>, --delete-profile <name>, --profile <name>; window title shows "WarmDesk — <label>" for non-default profiles
- Extend in-app contextual help to Kanban board, helpdesk inbox, all six project settings tabs (general, members, labels, API keys, webhooks, deleted cards), and all chart types (velocity, burndown, burnup, CFD, sprint report, epic burndown, release burndown, cycle time, lead time); add inline HelpIcon field hints for undeclarable time, report group-by, and export in the time tracking view
- Add metrics global role for read-only Prometheus scraper accounts; accept Authorization: ApiKey <key> header so Prometheus scrape_configs work natively without a reverse proxy
- Expose warmdesk_metrics_last_access_timestamp_seconds and warmdesk_metrics_last_access_success gauges; record scrape history in system_settings; display last-access time using the user's date/time format in the admin panel
- Add admin API key management: create, list, and revoke API keys for any user from the edit-user modal; service accounts no longer need admin access to get a key
- Add soft-delete and restore users: deleted users are soft-deleted and recoverable via a "Show deleted" toggle; permanent purge nullifies FK references on content records and removes personal/membership rows
- Improve in-app contextual help for Project Settings, System Settings, and Customers & Contracts with full tab coverage and eleven new inline field hints; teleport HelpIcon popovers to body to prevent sidebar clipping
- Add MFA trusted devices: after successful TOTP verification, users can trust their device for 7 or 30 days; subsequent logins skip the MFA challenge; trust tokens are stored as SHA-256 hashes with httpOnly SameSite=Strict cookies; Settings shows all active trusted devices with last-used and expiry dates and supports individual and bulk revoke; logout automatically revokes the trust record for the current device
- Fix grid PDF exports to respect the user's date/time format setting in the week period label and year print date
- Fix grid PDF exports to respect the user's time notation setting (decimal or hh:mm) for all cell values, totals, and undeclarable rows
- Add MFA remember-devices admin policy (disabled / 1 week / 1 week or 1 month) with automatic trust purge when tightened; passkey login and Tauri honour the policy
- Fix undeclarable timesheet alignment: day-cell deductions right-align with entered time; row totals show declarable time; consistent `-` prefix in UI and PDF exports; screen-reader labels via aria-describedby
- Fix page help for time-tracking sheet (array i18n keys, modal scroll, declarable vs logged totals); add screenshot 24 for undeclarable grid documentation
- Fix auth cookies: Secure flag only on direct TLS or X-Forwarded-Proto: https (fixes form login on HTTP and behind reverse proxies); document requirement in deploy templates

- Fix time-tracking input fields: absorb a manually typed colon when the auto-colon is already present (prevents 19::0 double-separator); accept bare hour values (e.g. 20) as 20:00 on Enter in the start/end popup, standby form, and grid cell

- Add per-user audit trail: admin Edit User panel shows full login/activity history (login ok/fail, logout, password change/reset, MFA events, passkey and API key lifecycle, email change, admin-on-behalf actions); each entry records timestamp, IP, client, and actor; LoginHistory model gains ActorID/ActorUsername fields
- Fix RBAC: block customer-role users from UpdateInboxTicket and DeleteInboxTicket; fix ListContracts IDOR (now requires requireCustomerAccess); raise project API key create/delete from member to owner role
- Ansible collection v0.5.0: add epic, sprint, ticket_checklist_template modules; user module gains state=restore (un-soft-delete) and state=purge (permanent removal with FK cleanup)
- Show contract hourly rate in the weekly time log grid: display the base rate (e.g. "45 €/h") per row with a ✦ badge when time-slot overrides are present; contracts are eagerly loaded on week load
- Route all backend log output to syslog (LOG_DAEMON, tag "warmdesk") in addition to stderr so auth events and server messages are captured by standard log aggregation tools
- Change server startup message to "Starting WarmDesk - Time Tracking <version>" when running in --mode=timetracking
- Fix fill-from-slots: multi-day overnight slots (e.g. Fri 19:00→Mon 07:00 with end_day_offset=3) now store start_time="00:00"/end_time="00:00" on continuation days (Sat/Sun), showing the dot indicator and correct shift times in the popup instead of the 09:00–17:00 placeholder
- Seed standby entries for current week: warmdesk-seed --reset pre-fills Mon–Sun with Acme Phase 1 standby time entries (Mon–Fri 19:00→07:00, Sat/Sun 00:00–00:00) so the rate column and time-slot features are immediately visible
- Replace working-hours single start-time input with start+end time pair per day: three-column grid (Start / End / Hours) with calculated net hours displayed live; existing start-time values are preserved
- Add lunch break field (default 30 minutes, 0–120 range) to working hours settings; subtracted from each day's net hours calculation
- Convert user-guide and admin-guide documentation from Markdown to AsciiDoc for asciidoctor-pdf support and native GitHub rendering
- Ship user and admin guide PDFs in the server distribution; download from avatar menu → Downloads submenu (keyboard-accessible flyout)
- Branded PDF documentation theme with title page, logo, revision metadata, and bundled fonts (make docs-pdf-guides)
- Show weekly net working hours total below the working hours grid in User Settings
- Confirm before Copy Previous Week when the timesheet grid is not empty; make the copy undoable
- Confirm dialog primary button defaults to Yes; destructive actions use optional red Delete label
- Use muted styling for undeclarable time in the timesheet grid, report tab, and project manage modal
- Remove warmdesk-blog.adoc from docs tree
- Apply destructive confirm styling (red Delete button) to delete, revoke, purge, and leave/remove flows app-wide
- Version guide PDF downloads as WarmDesk-user-guide-vX.Y.Z.pdf and WarmDesk-admin-guide-vX.Y.Z.pdf
- Run make sync-doc-revisions automatically from scripts/release bump
- Reformat remaining admin guide code blocks for PDF page width; document copy-week confirm/undo and muted undeclarable styling in user guide
- Add website release blog post for v0.12.14
- Move release helper to scripts/release and seed_tickets.py to scripts/
- Fix CI asciidoctor-pdf install with sudo gem install in server workflows
- Add time-tracking logo set (clock-on-desk icon and full wordmark)
- Remap logos to time-tracking variants when running with --mode=timetracking
- Document time-tracking-only mode in admin guide, user guide, and config reference
- Switch all frontend logo elements to timetracking variants via systemStore.logoSrc/logoFullSrc
- Install warmdesk and warmdesk-timetracking icon sets (SVG + 4 PNG sizes) into deb/rpm system icon tree
- Add warmdesk-timetracking.desktop launcher entry to deb/rpm packages
- Hide SMTP, IMAP, project defaults, and Scrum Story Points in admin settings when in timetracking mode
- Hide Email Notifications and Personal API Keys in user settings when in timetracking mode
- Add black (pure #000000 / AMOLED) theme option to the header quick theme switcher
- Fix card delete confirmation dialog appearing behind the open card detail modal
- Restore emoji flags in browser holidays dropdown
- Copy time tracking cell including distance, start/end time, and holiday flag; paste and undo restore all fields
- Add time tracking macros: named templates, variable rows, day 1/day 2 presets (hours, start/end, distance), apply to first N days with alternating pattern, JSON import/export
- Add Ctrl+X cut for time tracking cells; improve rectangular copy/paste across rows and weeks; double-click customer/project or activity to edit row
- Add show/hide closed toggle on Kanban board (per-column, persisted) and ticket list/inbox views
- Fix breadcrumbs for tickets, inbox, epics, charts, news, invoices, and time-tracking routes
- Simplify time macro apply: choose apply-to-days (default 5), alternating A/B toggle (off by default), single-column per-day grid when not alternating; settings saved per macro
- Add website release blog posts for v0.12.33 and v0.12.34
- Sync time macro libraries per user on the server; separate run popout (bottom bar) from editor (top bar); macro i18n in all languages; fix row-comment and double-click-edit translations
- Add website release blog post for v0.12.35
- Add system tray icon to the Tauri desktop app with unread-notification badge (cards, tickets, chat), show/hide window menu item, and quit item; add close-to-tray behavior and Settings → Interface toggles for both
- Add custom date range option to the time-tracking report period selector, opening a Start/End date picker instead of only week/month/year presets
- Show distance and undeclarable (unbillable) time per row in the time-tracking report and its PDF export, in addition to group/grand totals; add matching PDF export options
- Show the undeclarable/billable time breakdown for every report grouping, not only when grouped by customer
- Add Table/Chart toggle to the time-tracking report with bar, pie, and stacked-bar chart types, a declarable/total time basis switch, and a server-side chart PDF export
- Extract report date-range resolution into a single resolveReportEntries() function shared by the JSON, table PDF/XLSX, and chart PDF exports
- Place the custom date range modal's Start/End date pickers side by side instead of stacked
- Fix dead-code warning on Windows builds for a helper only used on Linux/macOS
- Show the customer instead of a duplicate activity name in the bold line of time-tracking chart tooltips
- Audit the user and admin guides against the codebase and correct outdated labels, wrong defaults, missing features, and incorrect claims about login history, credit notes, and SLA matching
- Fix systemd unit ReadWritePaths to include the uploads directory alongside data
- Net credit notes into the invoice list's revenue summary totals instead of excluding them
- Add a per-user "require password change on next login" flag, settable from the admin New User and Edit User forms, enforced across password/passkey/MFA login and cleared automatically on next password change
- Add Generate/Copy password buttons to the admin New User and Edit User password fields
- Backfill 84 i18n keys missing from all 11 non-English locales and translate remaining untranslated English strings across every locale
- Show a passkey-registered indicator in Settings → Profile, and a "Logged in as {username} (ID: {id})" hover tooltip on the header avatar and footer display name
- Add admin visibility and management for user passkeys: a registered-count badge and Revoke passkeys button in Edit User, backed by new GET/DELETE /admin/users/:id/passkeys endpoints
- Adopt real semver (PATCH/MINOR/MAJOR) versioning discipline starting at v0.13.0, replacing the flat incrementing counter used through v0.12.42
- Add an "Undeclarable" time basis option to the time-tracking report chart view, alongside the existing Declarable and Total options, in both the on-screen chart and the chart PDF export
- Add a Passkey column to the Admin Users table showing at a glance which users have a passkey registered; shorten "Last Password Change" to "Last PWD Change" to make room for it
- Clear the auth rate limit bucket for an IP after a fully successful login (password, passkey, MFA, or refresh) so earlier failed attempts no longer risk locking out a user who just authenticated correctly
- Fix passkey login errors being masked by a bogus token-refresh retry: the 401-retry interceptor now excludes /auth/passkey/login so the real failure reason is shown instead of a generic "missing refresh token" message
- Fix bar and stacked-bar chart hover offset when the app is zoomed in or out: chart canvases now cancel the app-wide CSS zoom locally so Chart.js's hover hit-testing always operates in a consistent, unzoomed coordinate space
- Fix time-tracking report pie chart cramped into a 420px square: 1.6:1 aspect ratio and a 640px-wide wrapper let the circle and legend fill the available space
- Fix admin Settings tab hard-capped narrower than every other admin tab: restructured into a responsive card grid that fills the full admin container width; also fixes paired fields (Postal Code+City, VAT+CoC number) stacking vertically instead of side-by-side in both admin Settings and the Customer edit modal
- Fix time-tracking report charts grouping near-duplicate activity descriptions (case, spacing, punctuation variants) as separate activities and inflating the "Other" slice: activities are now grouped by a normalized key, displaying whichever raw spelling has the most logged minutes
- Raise the time-tracking report chart category limit from 7 to 10 before falling into "Other"; added 3 new colorblind-safe colors per theme (light/dark/black), validated so adjacent chart slices stay distinguishable under protanopia/deuteranopia simulation
