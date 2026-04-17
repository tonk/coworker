Create an application that has all these features and requirements

- Do not use Docker or Podman
- Written in Golang, minumum version 1.25
- Webbased
- Kanban board
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
