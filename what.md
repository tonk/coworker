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
- Fix Escape key in card comment editor closing the card modal
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
- Due date field replaced from browser date picker to text input following user's configured date format, with clear button
- Spellcheck in card description, comments, and title (plain textarea replaces CodeMirror for editing; markdown preview unchanged)
- Auth tokens moved to sessionStorage so closing the browser ends the session
- Fix project switching in sidebar not reloading board content (watch route slug; useWebSocket accepts reactive ref)
- Due date calendar picker: hidden native `<input type="date">` triggered by a calendar icon button (📅) in the card detail; preserves configured date format display while allowing picker-based input
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
- Fix Escape key closing card modal even when a mention dropdown is open: check e.defaultPrevented before closing
- Fix Cancel button discarding unsaved card changes silently: show Save / Discard / Back confirmation panel when the card is dirty
- Escape key closes card modal when there are no unsaved changes
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
- Add HSTS and Content-Security-Policy headers to nginx and Apache reverse-proxy templates
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
- Show a ← [parent card title] back link at the top of sub-card and linked-card nested modals to return to the originating card
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
