# Changelog

All notable changes to WarmDesk are documented here.

## v0.7.8 — 2026-04-19

### Added
- **Backup email notifications** — Admin → Backup / Restore tab now has a toggle to send an email after every backup (manual or scheduled); configure a recipient address; the email includes date/time, success or failure status, and a list of all available backups with file sizes and dates
- **Backup Prometheus metrics** — three new gauges on `GET /api/v1/metrics`: `warmdesk_backup_last_run_timestamp_seconds` (Unix timestamp of the last backup attempt), `warmdesk_backup_last_success` (1 = success, 0 = failure, −1 = never run), and `warmdesk_backup_files_total` (count of backup files currently on disk)
- **HTML emails** — all outbound emails (backup notifications, @mention alerts, card assignments, direct message notifications, password reset, SMTP test) are now sent as `multipart/alternative` with a plain-text fallback and a styled HTML version; the HTML email uses a shared template with a blue header showing the company name and logo, a footer with the WarmDesk icon, version number, and instance URL
- **Dark-mode email support** — HTML emails include `prefers-color-scheme: dark` media-query overrides so the password-reset button and other elements remain readable in Apple Mail, iOS Mail, Samsung Mail, and Outlook for Mac when the user has dark mode enabled; a visible border is added to the button as a fallback for clients that strip background colours entirely

### Fixed
- **Email footer double `v`** — version tag (e.g. `v0.7.7`) was displayed as `vv0.7.7`; leading `v` is now stripped before rendering

## v0.7.7 — 2026-04-17

### Fixed
- **Desktop packages `.desktop` template** — `desktopTemplate` must be set under `bundle.linux.deb` and `bundle.linux.rpm` individually in Tauri 2, not under `bundle.linux` directly; previously caused a build error when bundling `.deb` and `.rpm`
- **Desktop app category** — `"Office"` is not in Tauri's accepted category list; changed to `"Productivity"` which maps to the `Office;` XDG desktop category

## v0.7.6 — 2026-04-17

### Fixed
- **`warmdesk-seed --reset` crash on fresh database** — seed tool failed with "no such table: projects" when run against a brand-new (or just-wiped) database; the startup migration guard now checks whether the projects table exists before attempting the `key_prefix` column backfill, so fresh databases initialise cleanly
- **Desktop app `.desktop` file incomplete** — generated `.desktop` file was missing `GenericName`, `Comment`, `Keywords`, and had an empty `Categories` field; fixed by adding a custom Tauri desktop template and wiring `shortDescription` and `category` in `tauri.conf.json`

## v0.7.5 — 2026-04-17

### Added
- **Download backup** — each backup in Admin → Backup / Restore now has a Download button that streams the file directly to the browser; useful for offsite storage or transferring a backup to another server
- **Repository layout docs** — new `docs/repository-layout.md` explains every directory in the repo

## v0.7.4 — 2026-04-17

### Added
- **`make deb` and `make rpm`** — build Debian and Fedora packages for the WarmDesk desktop client using Tauri's built-in bundlers; `.desktop` entry and icons are included automatically; requires `dpkg` (deb) or `rpm-build` (rpm) plus the same Rust/webkit dependencies as `make appimage`

### Fixed
- **Ansible collection — namespace renamed to `ansilabnl`** — was previously `ansilab`; all module, plugin, and import references updated
- **Ansible collection — Galaxy upload requirements** — added `README.md` and `meta/runtime.yml` (with `requires_ansible: ">=2.14"`) which Ansible Galaxy requires before accepting an upload
- **Ansible collection — remaining YAML parse errors** — fixed six modules still referencing the non-existent `auth` doc fragment (→ `connection`); fixed `Note:`, `B(...):`, and `following types:` colon patterns that YAML misread as keys; removed `{...}` literal in the `webhook` module description that triggered YAML flow-mapping parsing
- **Backup file list** — long filenames now truncate with an ellipsis (full name visible on hover); Size column widened so values no longer wrap; unit corrected from `KB` to `kB` (SI prefix)

## v0.7.3 — 2026-04-17

### Added
- **Backup scheduler** — Admin → Backup / Restore tab now includes a built-in scheduler; choose an interval (every 6 h, 8 h, 12 h, or once a day), set how many backups to retain, and WarmDesk creates backups automatically server-side — no cron job needed; last run time and next scheduled run are shown in the UI; old backups are pruned automatically once the retention limit is reached
- **Backup start time** — optional start time (HH:MM) for the backup scheduler; when set, backups run at fixed time-of-day slots (e.g. start 02:00 + every 6 h → 02:00, 08:00, 14:00, 20:00) instead of as an offset from the last run, preventing backups from drifting into peak hours

### Fixed
- **nginx WebSocket config** — dedicated `location ~ ^/api/v1/ws` block with hardcoded `Upgrade`/`Connection` headers; fixes WebSocket failures with nginx 1.25+ and HTTP/2 where `$http_upgrade` is empty for RFC 8441 extended CONNECT requests; updated `listen` directive to `listen 443 ssl; http2 on;` (required syntax from nginx 1.25+)
- **Ansible collection — DOCUMENTATION parse errors** — replaced colon-after-word patterns in RST description strings that YAML interpreted as key-value pairs; added missing `doc_fragments/connection.py` fragment; fixed three modules that referenced a non-existent `auth` fragment; corrected the `card` module example that was missing `card_number`

## v0.7.2 — 2026-04-16

### Added
- **Backup / Restore tab in Admin panel** — new dedicated tab next to Users, Projects, and Settings; create a timestamped backup (`warmdesk_db_YYYYmmdd_HHMM`) stored in `./backups/`; list all available backups with filename, size, and creation date; restore from any backup (SQLite: live close / copy / reinit — no server restart needed; PostgreSQL: `psql`; MySQL: `mysql`); delete individual backup files
- **`backup` global role** — dedicated role for automated backup accounts; users with this role can only call `POST /api/v1/backup`; intended for cron jobs and CI scripts: create a user, assign the `backup` role, generate an API key, then `curl -X POST .../api/v1/backup -H "X-API-Key: ..."` on a schedule
- **Backup and restore logging** — server logs every successful backup and restore with filename, database driver, triggering user ID, and client IP
- **Show and restore deleted projects** — Admin → Projects now has a "Show deleted" toggle; deleted projects appear with a Deleted badge and a Restore button that recovers them from soft-delete
- **All 12 languages in user settings and PDF export** — the per-user locale selector in User Settings and the PDF language selector in the Report view now list all 12 supported languages (previously only 5 were available in those two places)

## v0.7.1 — 2026-04-16

### Added
- **Key prefix visible in Admin → Projects** — each project row now shows its card prefix (e.g. `AAP`) next to the slug, making duplicate or missing prefixes immediately visible

### Fixed
- **Database upgrade fails with UNIQUE constraint on `key_prefix`** — three-part fix for databases upgrading from before v0.7.0:
  1. When the `key_prefix` column does not exist yet, it is added via `ALTER TABLE` (without the unique index) so data can be populated before AutoMigrate creates the constraint
  2. Projects with an empty prefix are now assigned generated prefixes before AutoMigrate runs, not after
  3. Soft-deleted projects are included in both deduplication passes — the unique index applies to all rows in the table, not just active ones

## v0.7.0 — 2026-04-16

### Added
- **7 new UI languages** — Danish (Dansk), Swedish (Svenska), Norwegian Bokmål (Norsk), Finnish (Suomi), Icelandic (Íslenska), Portuguese (Português), and Italian (Italiano); all ~350 translation keys covered; all 12 languages are selectable in the header language picker, the system-default locale setting, and the per-user locale setting in Admin → Edit User
- **About modal** — a new **About** item in the user navigation dropdown opens a modal showing the frontend version, server version (fetched live), project description, repository link, license, and copyright notice
- **Customer access control** — non-admin users only see customers they are explicitly assigned to (strict allowlist — no assignment means not visible); global admins always see all customers
- **Per-customer roles** — each CustomerAccess row carries a role: **Member** (read-only visibility) or **Admin** (can edit customer details, contracts, and manage the member list); configurable from Admin → Users → Edit User → Customer Access and from the Customer Detail page → Members
- **Customer member management API** — `GET/PUT /customers/:id/members` lets global admins and customer-admins manage customer visibility and roles; self-lockout protection prevents a non-global-admin from removing their own admin row
- **Unique card prefix enforcement** — `key_prefix` (e.g. `PRJ` in `PRJ-42`) is now unique across all projects so card codes are unambiguous for GitHub and webhook integrations; the auto-generator appends a numeric suffix when the base is taken (`WAR`, `WAR2`, `WAR3`); duplicate prefixes in existing databases are deduplicated on startup before AutoMigrate runs; the prefix cannot be changed after creation
- **Ansible `customer_member` module** — new `ansilab.warmdesk.customer_member` module manages `GET/PUT /customers/:id/members`; parameters: `customer` (name), `username`, `role` (member/admin), `state` (present/absent); full-list sync via PUT; resolves username to user\_id via `GET /users`; check\_mode aware
- **Ansible `user` module `customer_roles` parameter** — a dict `{customer_name: role}` that performs a full sync of a user's customer assignments; pass `{}` to clear all; customer names are resolved to IDs via `GET /customers`

### Changed
- **Resizable Edit User modal** — the Admin → Edit User modal can now be resized by dragging any corner or edge; header and footer remain pinned while the body scrolls
- **Training seeder unique prefixes** — training projects now receive unique card prefixes (`EDA00`, `EDA01`, …) to satisfy the new uniqueness constraint
- **Git integration regex** — the card-reference pattern updated from `[A-Z]{2,8}` to `[A-Z][A-Z0-9]{0,9}` to match digit-suffixed prefixes like `WAR2`

## v0.6.9 — 2026-04-15

### Added
- **PDF language selector** — a new **PDF Language** dropdown in the report export bar lets you choose the language used for all labels in the exported PDF (column headers, subtotals, grand total, footer) independently of your UI language; **Auto** follows the current interface language; manual options are English, Nederlands, Deutsch, Français, and Español; selection is remembered across sessions — useful when the UI is in one language but the report is intended for someone using another
- **Per-project subtotal pill badge in PDF** — each project header bar in the exported PDF now shows the project total time as a white rounded pill on the right, matching the on-screen report

### Fixed
- **PDF export crashes with HTTP 500 when company logo is a PNG with transparency** — PNG images with an alpha channel are now composited over a white background before embedding in the PDF; gofpdf cannot handle RGBA PNGs and previously set an internal error that caused the entire export to fail
- **WebSocket close-1005 messages logged as errors** — close code 1005 ("no status received") is sent by browsers on normal navigation and tab close; it is no longer logged as an unexpected error

## v0.6.8 — 2026-04-15

### Added
- **Configurable card prefix** — the short identifier used in card references (e.g. `PRJ-42`) can now be set when creating a project; it auto-generates from the project name (same algorithm as before) but can be freely edited to any 1–10 uppercase letter/digit string; a live preview (e.g. `PRJ-1`) is shown next to the field; input is forced to uppercase as you type; the prefix cannot be changed after the project is created; available in both the dashboard **New Project** modal and the admin **Projects** panel

## v0.6.7 — 2026-04-15

### Fixed
- **Windows desktop app login-screen typing lag (further reduced)** — three additional mitigations applied: (1) WebView2's built-in password-reveal button is now hidden via `::-ms-reveal { display: none }`, removing a synchronous IPC round-trip that fired each time the password field gained focus; (2) `spellcheck`, `autocorrect`, and `autocapitalize` are disabled on both credential inputs, preventing the Windows Spell Check service IPC from running on every word boundary; (3) WebView2 autofill is disabled at the engine level via `ICoreWebView2Settings4::SetIsGeneralAutofillEnabled(false)` and `SetIsPasswordAutosaveEnabled(false)`, cutting the per-keystroke credential-manager IPC from the renderer to the browser process; some residual lag remains under investigation

### Changed
- **INSTALL.md desktop build prerequisites expanded** — section 14 now lists per-platform requirements in full: Linux requires Ubuntu 24.04 (older HarfBuzz bundled by Ubuntu 22.04 breaks font rendering on Fedora 43), Rust via rustup, and the appropriate system libraries; macOS requires Xcode Command Line Tools, Rust, and both architecture targets for a universal binary; Windows requires Go, Node.js, Rust via rustup-init.exe, and notes that WebView2 is pre-installed and NSIS is downloaded automatically

## v0.6.6 — 2026-04-14

### Added
- **Forgotten password** — users can click **Forgot password?** on the login page and receive a one-time reset link by email; the link is valid for one hour; requires SMTP to be configured by an administrator
- **Password policy** — administrators can configure minimum password length, and require uppercase, lowercase, digit, and/or special characters under **Admin → Settings → Password Policy**; policy is enforced on registration, password change, and password reset; active requirements are shown to users beneath the password field
- **Avatar image upload** — users can upload an avatar image directly in User Settings instead of supplying a URL; the image is stored on the server like any other attachment

### Changed
- **Default minimum password length raised to 12** — new installations default to a minimum of 12 characters instead of 8; existing deployments are not affected until the setting is explicitly saved

### Fixed
- **Seed: tonk user now has starred projects and customers** — the persistent `tonk` admin account is pre-seeded with all three demo projects and Acme Corporation + Globex Systems starred, matching the experience for the `demo.admin` account
- **INSTALL.md Go requirement corrected** — the installation manual now states Go 1.25 (matching `go.mod`); it previously incorrectly listed 1.22
- **INSTALL.md first-admin instructions corrected** — the first registered user is automatically promoted to admin; the manual previously gave incorrect instructions to update the database manually

## v0.6.4 — 2026-04-09

### Fixed
- **Windows desktop app typing lag on login screen** — the global `keydown` listener for Ctrl+zoom is now registered as a passive listener in the Tauri desktop app; WebView2 required a synchronous IPC round-trip for every keystroke when the listener was non-passive, causing visible lag when typing credentials; the fix does not affect browser builds
- **Sidebar drag-to-reorder broken on Linux desktop app** — replaced the HTML5 Drag-and-Drop API with a pointer-events implementation (`pointerdown` / `pointermove` / `pointerup`); WebKitGTK's DnD support is incomplete so section and item reordering was silently broken on the Linux AppImage; the new approach works on all platforms

## v0.6.3 — 2026-04-09

### Added
- **Per-user accent colour** — users can choose an accent colour (blue, red, green, or orange) in Account Settings; affects buttons, links, active highlights, and focus rings throughout the UI; saved per user in the database and applied on login
- **Drag-to-reorder sidebar sections** — all sidebar sections can be reordered by dragging their grab handle; new default order is Starred Projects → All Projects → Favourite Customers → All Customers → Favourites → People → Chats; order persisted in localStorage

### Changed
- **Sidebar section spacing** — increased whitespace between sidebar sections for improved readability
- **Sidebar item indentation** — items in All Projects, All Customers, People, and Chats sections are indented to visually group them under their section header; empty-state messages follow the same indent

## v0.6.2 — 2026-04-08

### Fixed
- **Deleted projects reappear in admin list** — the admin project list no longer returns soft-deleted projects; previously `Unscoped()` caused deleted projects to reappear whenever the admin navigated back to the Projects tab
- **Deleted project stays visible in sidebar** — deleting a project via the admin interface now immediately removes it from the sidebar "All Projects" list without requiring a page refresh
- **Seed `--reset` fails with unique constraint error** — `--reset` now collects soft-deleted demo projects (via `Unscoped`) before wiping; previously, projects deleted through the admin UI were missed and their slugs remained in the database, causing a UNIQUE constraint failure on re-seed

## v0.6.1 — 2026-04-08

### Added
- **Sidebar drag-to-reorder starred projects and customers** — starred projects and starred customers in the sidebar can be dragged to a custom order; order is persisted in localStorage
- **All Customers section in sidebar** — a new collapsible "All Customers" section lists every customer alphabetically (starred ones first with a ★ badge); collapsed by default

### Fixed
- **Card date pickers don't open** — Start Date and Due Date pickers in the card detail now use an overlay `<input type="date">` hidden behind the calendar icon button; avoids the browser `NotAllowedError` thrown by `showPicker()` on `display:none` elements
- **Contract editor date fields ignore configured date format** — Start Date and End Date in the contract editor now use the same text-input + overlay-picker pattern as card dates; they display and parse dates according to the user's configured date format, with a calendar icon and a clear button

## v0.6.0 — 2026-04-08

### Added
- **Gantt chart per project** — a Gantt view (frappe-gantt) is accessible from the board toolbar via the 📅 button; cards with a start date and/or due date appear as bars; bars are colour-coded by priority; clicking a bar opens the card detail; dragging a bar updates the card's start and due dates; Day / Week / Month view modes
- **Start date on cards** — cards now have a `start_date` field alongside the existing `due_date`; editable in the card detail via the same date-picker style as due date; stored in the database and returned in all card API responses
- **Cross-references between cards** — cards can be linked to each other from the card detail "Linked Cards" section; links are bidirectional; shown with card number, project key, title, priority, and column; clicking a linked card opens it in a nested detail modal; cross-project references supported
- **Demo data extended** — the seed tool now includes start dates and due dates on all demo cards, sub-cards on selected cards, and cross-reference links between cards; all three demo projects now populate the Gantt view

### Fixed
- **"Edit customer" shows empty fields** — clicking the edit button in Customer Detail now correctly pre-populates the form before opening it
- **Switching customers in the sidebar does not update the overview** — the Customer Detail view now watches the route parameter and reloads when the customer changes

## v0.5.3 — 2026-04-07

### Added
- **Customer / Contract / Project hierarchy** — customers are top-level entities; contracts sit under a customer; projects can be linked to a customer and optionally to a contract within that customer
- **Customers page** (`/customers`) — grid of customer tiles showing name, description, logo, contract count, and project count; star/unstar favourites; admins and project owners can create, edit, and delete customers and contracts
- **Customer detail page** (`/customers/:id`) — customer header with inline name editing; contracts listed with their projects grouped beneath; unassigned projects shown separately; contract date ranges displayed
- **Customers section in sidebar** — starred customers displayed with star/unstar toggle; link to the full customers list
- **Customer / Contract in Project Settings** — General tab gains Customer and Contract dropdowns; saving links the project to the selected customer and contract (or clears the link)
- **Customer badge on dashboard tiles** — project tiles show the customer name when a project is linked to a customer
- **Sub-cards** — cards can have child cards (one level deep); sub-cards are created and managed inside the parent card's detail view and are hidden from the board column; parent cards show a progress pill (done / total) on the board; each sub-card gets its own card number and full detail (assignees, labels, comments, etc.); clicking the open button in the sub-card list opens the sub-card in a nested detail modal

### Fixed
- **Linux desktop app blank screen on Fedora / Wayland** — at startup the app now automatically detects `libwayland-client.so` at well-known paths (Fedora `/usr/lib64`, Ubuntu `/usr/lib/x86_64-linux-gnu`, ARM64, and others) and re-execs itself once with `LD_PRELOAD` set; a sentinel env var (`WARMDESK_PRELOAD_DONE`) prevents infinite loops; no manual `LD_PRELOAD` configuration is required

## v0.5.2 — 2026-04-07

### Added
- **Prometheus metrics endpoint** — `GET /api/v1/metrics` exposes project, column, and card counts in Prometheus text format; protected by a new `metrics` global role (assignable in Admin → Users); admins also have access
- **Typing indicator in chat** — animated three-dot indicator appears above the compose area showing who is currently typing in project chat; auto-clears 4 seconds after the last keystroke
- **@mention autocomplete in card description and comments** — the `@username` mention dropdown (already available in chat) now works in the card detail description field and comment box; Escape dismisses the dropdown without closing the card
- **Project ordering** — admins can drag project tiles on the dashboard to set a custom display order; order is persisted to the database and respected for all users

### Fixed
- **Forgejo webhook shows Gitea logo** — the card Git Links section now correctly detects Forgejo events from `X-Forgejo-Event` headers and shows the Forgejo badge (blue) rather than the Gitea badge (green)
- **Git links in desktop app do nothing when clicked** — external links in the card Git Links section and the update banner "View release notes" link now open in the system browser when running as the desktop app (via `tauri-plugin-opener`); previously nothing happened
- **Escape closes dirty card accidentally** — pressing Escape now only closes the card modal when there are no unsaved changes; if there are changes the card stays open
- **Cancel with unsaved changes loses work silently** — clicking Cancel (or the ✕ button, or backdrop) on a card with unsaved changes now shows a "Save / Discard / Back" confirmation panel instead of closing immediately

## v0.5.1 — 2026-04-03

### Fixed
- **Webhook setup URL shows `tauri://` in desktop app** — the payload URL displayed in Project Settings → Webhooks for GitHub, GitLab, and Gitea was derived from `window.location.origin`, which is `tauri://localhost` inside the desktop app; it now uses the configured server URL so the correct `http(s)://` address is shown

## v0.5.0 — 2026-04-03

### Added
- **Server version in footer** — after login, the footer shows both the client version and the server version (`WarmDesk vX.Y.Z · server vX.Y.Z`); fetched from the new public `GET /api/v1/version` endpoint

### Fixed
- **`make appimage` / `make dmg` broken by non-semver git tags** — `git describe` was picking up arbitrary tags (e.g. `works_on_win_and_linux`) and producing a version string that Tauri rejects; the Makefile now passes `--match 'v*'` so only version tags are considered

## v0.4.12 — 2026-04-03

### Fixed
- **Linux desktop app COLRv1 crash (final fix)** — `font-variant-emoji: text` added globally to CSS forces text presentation of emoji, bypassing the COLRv1 colour-font rendering path in Skia entirely; webkit2gtk 2.50.x on Fedora 43 has a bounds-check assertion failure (`colrv1_configure_skpaint`) when rendering COLRv1 emoji; env vars and hardware-acceleration settings cannot prevent it because the crash is in the CPU font-rendering path

## v0.4.11 — 2026-04-03

### Fixed
- **Linux desktop app COLRv1 crash (attempt)** — added `GDK_RENDERING=image` to force GDK software rendering; did not prevent the crash (Skia font rendering is unaffected by GDK rendering mode)

## v0.4.10 — 2026-04-03

### Added
- **FreeFont support** — FreeSans, FreeSerif, and FreeMono are now available as selectable fonts; the woff files are served from the same origin (no external font CDN required)
- **Linux `.desktop` file** — `deploy/warmdesk.desktop` for system-wide installation; documented in `INSTALL.md`

### Fixed
- **Linux desktop app COLRv1 crash (attempt)** — disabled WebKit hardware acceleration (`HardwareAccelerationPolicy::Never`) via the `with_webview` API; also sets `WEBKIT_DISABLE_DMABUF_RENDERER=1` to avoid a DMA-BUF renderer blank window on many GPU configurations; the COLRv1 crash was not fully resolved until v0.4.12

### Changed
- **Fonts now self-hosted** — Inter, Roboto, Open Sans, and Source Code Pro are bundled via `@fontsource` npm packages instead of loading from Google Fonts; eliminates the external network dependency and makes the font setting work in air-gapped and desktop (Tauri) deployments
- **Ctrl+Scroll zoom** — mouse wheel zoom (Ctrl+Scroll) added alongside existing Ctrl+/Ctrl-/Ctrl+0 keyboard shortcuts
- **Windows code signing temporarily disabled** — SignPath signing steps are commented out in the release workflow until the signing certificate is renewed

## v0.4.8 — 2026-04-03

### Added
- **Project-scoped API keys** — keys created in Project Settings → API Keys are now scoped to that project only; a project key is rejected on requests for any other project; ideal for CI/CD pipelines
- **Personal API keys** — a new API Keys section in User Settings lets you create personal keys with full access across all your projects; use these for scripts and tools that span multiple projects
- **API keys on all endpoints** — API keys (both personal and project-scoped) now authenticate all protected endpoints, not just the Ticket API
- **Swagger UI base URL** — new `base_url` config setting (env: `BASE_URL`) sets the host shown in the Swagger UI so "Try it out" calls reach the correct server; documented in `warmdesk.yaml.example`, `README.md`, and `INSTALL.md`
- **Code Signing Policy** — added required SignPath Foundation code signing policy section to `README.md`

### Fixed
- **Font family setting had no effect** — selected fonts (Inter, Roboto, Open Sans, Source Code Pro) are now loaded from Google Fonts on demand; the CSS variable was previously set to a bare name with no corresponding font loaded
- **Open Sans and Source Code Pro showed wrong font** — the Google Fonts lookup was keyed by plain name but option values are full CSS stacks (`'Open Sans', sans-serif`); now extracts the font name from the CSS value before lookup
- **Font size setting had no effect** — `button`, `input`, `textarea`, and `select` elements had a hardcoded `font-size: 14px` that overrode the `--user-font-size` variable; changed to `inherit`

### Changed
- **Swagger UI** — interactive API documentation available at `/swagger/index.html`; documented in `docs/api.md`

## v0.4.7 — 2026-04-03

### Fixed
- **Windows release CI** — the version-stamping step in the release workflow now runs under `shell: bash` (Git Bash) instead of the default PowerShell; PowerShell was interpreting the regex character class `[^"]*` in the inline Node.js script as an array index expression and aborting with a parse error

## v0.4.6 — 2026-04-03

### Added
- **Desktop app CLI flags** — `--version` / `-V` prints the app version and exits; `--maximized` starts the window maximised
- **Database TLS** — PostgreSQL and MySQL connections can now be encrypted via `db_tls_mode` (`disable` / `require` / `verify-ca` / `verify-full`), `db_tls_ca_cert`, `db_tls_cert`, `db_tls_key`; matching `DB_TLS_*` env vars; mutual TLS (client certificate) supported
- **Server TLS** — WarmDesk can now serve HTTPS directly without a reverse proxy; set `tls_cert` and `tls_key` (or `TLS_CERT` / `TLS_KEY` env vars) to enable; falls back to plain HTTP when either is absent

### Fixed
- **Linux desktop app network error** — webkit2gtk 4.1 treats `tauri://localhost` as a secure context and blocks `http://` requests as mixed content (same restriction as Windows WebView2); the fetch proxy now routes all HTTP/HTTPS requests through `tauri-plugin-http` on all Tauri platforms, while non-HTTP requests (internal `tauri://` scheme loads) continue to use native WebKit fetch — this also resolves the previously-reported blank screen caused by routing all requests through the plugin
- **Desktop app icons contained old Coworker branding** — all icon files (`32x32.png`, `128x128.png`, `128x128@2x.png`, `icon.png`, `icon.ico`, `icon.icns`) regenerated from the current WarmDesk SVG logo

### Changed
- **Desktop app version stamping** — `Cargo.toml` is now stamped with the git tag version alongside `tauri.conf.json`; `make appimage` / `make dmg` / `make windows-installer` stamp both files automatically before building so local builds report the correct version
- **AppImage build dependencies documented** — `INSTALL.md` gains a desktop app prerequisites section listing the required system libraries for Fedora/RHEL and Ubuntu/Debian, plus Rust installation instructions

## v0.4.5 — 2026-04-03

### Added
- **Database TLS** — PostgreSQL and MySQL connections can now be encrypted and verified via four new settings (`db_tls_mode`, `db_tls_ca_cert`, `db_tls_cert`, `db_tls_key`) with matching `DB_TLS_*` environment variables; modes: `disable` (default), `require` (encrypt without cert verification), `verify-ca`, `verify-full`; mutual TLS (client certificate) is also supported
- **Server URL change from login page (desktop app)** — the current server URL is shown at the bottom of the login screen with a "Change" link that navigates back to the Connect screen; no need to reinstall or clear local storage to point the app at a different server
- **Version number on Connect screen** — the app version is now shown on the Connect screen in addition to the login page
- **`ALLOWED_ORIGINS=*` wildcard support** — setting `allowed_origins` to `*` now correctly allows requests from any origin; previously `*` was treated as a literal string and had no effect

### Fixed
- **Windows desktop app login 403** — a combination of root causes all resolved: `http://tauri.localhost` (the actual Windows Tauri origin) was missing from the hard-coded CORS allow-list (only `https://tauri.localhost` was listed); HTTP/2 negotiation with `tauri-plugin-http` was rejected by some servers; some reverse proxies blocked the non-browser `reqwest` User-Agent on POST endpoints; error messages returned as a plain string body were not parsed correctly and showed as a generic failure
- **Desktop app fetch patch applied too early** — `window.fetch` is now patched via a synchronous inline script in `index.html` before any ES module loads, preventing a race condition where the first API request fired before the patch was in place

### Changed
- **CI: manual desktop build workflows** — split into per-platform jobs (Linux AppImage, macOS DMG, Windows installer); a manual server build workflow added; PowerShell-based version stamping replaced with a Node.js script that works on all platforms

## v0.4.4 — 2026-04-02

### Fixed
- **Linux desktop app blank screen (regression in v0.4.3)** — the `tauri-plugin-http` fetch patch was applied on all platforms; on Linux (WebKitGTK, `tauri://` origin) this caused a blank screen on startup; the patch is now scoped to Windows only where the mixed-content restriction actually applies
- **Windows desktop app still could not connect (v0.4.3 partial fix)** — the plugin import was fire-and-forget; Vue mounted and fired the first API request before `window.fetch` was patched; the app now awaits the import before mounting so Axios sees the patched fetch from the very first request

## v0.4.3 — 2026-04-02

### Fixed
- **Windows desktop app cannot connect to server** — the `@tauri-apps/plugin-http` JavaScript package was missing; the Rust crate was present but without the JS counterpart `window.fetch` was never patched, so WebView2 made all HTTP requests itself and blocked them as mixed content (`https://tauri.localhost` → `http://server`); installing the package and importing it at startup routes every request through the native Rust HTTP client
- **Axios requests in desktop app bypassed `tauri-plugin-http`** — Axios defaults to `XMLHttpRequest`, which is not intercepted by the plugin; the desktop app now uses the `fetch` adapter so Axios requests also go through the native HTTP client
- **GitHub Actions Go module cache failing** — `setup-go` was searching for `go.sum` in the repo root; path corrected to `backend/go.sum`

### Changed
- **GitHub Actions now runs on Node.js 24** — opted in via `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24` ahead of the June 2026 forced migration

## v0.4.2 — 2026-04-02

### Added
- **Resizable sidebar** — drag the inner edge of the sidebar to set your preferred width (150px–480px); works whether the sidebar is positioned left or right; width is persisted across sessions
- **App zoom** — `Ctrl +` / `Ctrl -` zoom the entire interface in or out in 10% steps (50%–200%); `Ctrl 0` resets to 100%; zoom level is persisted across sessions

## v0.4.1 — 2026-04-02

### Fixed
- **`logo-full.svg` not served in production** — the backend SPA catch-all was returning `index.html` for any path not explicitly registered; `logo-full.svg` is now registered as a static route in the router alongside `logo.svg`

### Changed
- **Migration tool config** — YAML section key renamed from `coworker:` to `warmdesk:`, environment variable overrides renamed from `COWORKER_URL` / `COWORKER_USERNAME` / `COWORKER_PASSWORD` / `COWORKER_PROJECT` to `WARMDESK_URL` / `WARMDESK_USERNAME` / `WARMDESK_PASSWORD` / `WARMDESK_PROJECT`, and the default config filename changed from `coworker-migrate.yaml` to `warmdesk-migrate.yaml`; Go type `CoworkerConfig` renamed to `WarmDeskConfig`; internal priority-map variable names updated to match
- **Header logo** — the app header now uses the full WarmDesk logo (`logo-full.svg`) instead of the icon-only mark
- **Documentation** — admin guide gains a Migration Tools section (§16) covering `warmdesk-export` / `warmdesk-import` usage, config, env vars, and column mapping; user guide corrects the header description and replaces the outdated EasyMDE editor reference with the plain-textarea reality; API reference fixes the API key format example; INSTALL.md lists all four distribution binaries
- **`.gitignore`** — `.claude/` directory excluded from version control

## v0.4.0 — 2026-04-02

### Added
- **`warmdesk-export`** — standalone binary that reads a WarmDesk project (columns, cards, checklists, comments, labels, tags, time entries, attachments, topics and replies) and pushes it to Jira, Trello, OpenProject, or Ryver
- **`warmdesk-import`** — standalone binary that reads a project from Jira, Trello, OpenProject, or Ryver and creates it in WarmDesk
- **`warmdesk-migrate.yaml.example`** — documented config file covering all four platforms; credentials can be supplied via the file, environment variables, or interactive prompts
- **Column mapping** — `column_map` in the config translates WarmDesk column names to/from platform-specific status/list names; unmapped columns are passed through unchanged
- Both migration binaries are built by `make build` and included in the distribution archive alongside `warmdesk-seed`

### Changed
- **Product renamed to WarmDesk** — all binaries, config files, documentation, and the application UI now use the WarmDesk name and logo; Go module path updated to `github.com/tonk/warmdesk`
- Config example file renamed from `coworker.yaml.example` to `warmdesk.yaml.example`
- Default database file is now `warmdesk.db`
- Distribution archive is now `warmdesk-{version}.tar.gz`
- Service template renamed to `deploy/warmdesk.service`

## v0.3.3 — 2026-04-02

### Added
- **`warmdesk-export`** — standalone binary that reads a WarmDesk project (columns, cards, checklists, comments, labels, tags, time entries, attachments, topics and replies) and pushes it to Jira, Trello, OpenProject, or Ryver; supports `--config FILE` and `--dry-run`
- **`warmdesk-import`** — standalone binary that reads a project from Jira, Trello, OpenProject, or Ryver and creates it in WarmDesk; same flags and config format
- **`warmdesk-migrate.yaml.example`** — documented config file covering all four platforms; credentials can be supplied via the file, environment variables (`WARMDESK_URL`, `WARMDESK_USERNAME`, `WARMDESK_PASSWORD`, `WARMDESK_PROJECT`, `PLATFORM_API_TOKEN`, `PLATFORM_API_KEY`), or interactive prompts
- **Column mapping** — `column_map` in the config translates WarmDesk column names to/from platform-specific status/list names; unmapped columns are passed through unchanged
- Both binaries are built by `make build` and included in the distribution archive alongside `warmdesk-seed`

### Platform notes
- **Jira**: issues created via REST API v3; descriptions and comments in Atlassian Document Format; checklist items as Subtasks; time via worklogs; column mapped via workflow transitions
- **Trello**: lists created on the board as needed; checklists native; time posted as a comment; labels created per card
- **OpenProject**: work packages via API v3 HAL+JSON; checklist items as child work packages; time entries posted; status/priority/type resolved by name at export time
- **Ryver**: tasks posted to a team workroom via the OData API; columns encoded as tags; topics exported as forum posts; falls back to topic post if the task API is unavailable

## v0.3.2 — 2026-04-02

### Fixed
- **Desktop app cannot connect to server** — `tauri-plugin-http` was never installed, so `globalThis.fetch` fell back to the native WebView browser fetch which is subject to CORS; on Windows the Tauri app origin (`https://tauri.localhost`) was not in the server's `ALLOWED_ORIGINS`, blocking every API call and the ConnectView probe; added `tauri-plugin-http` which patches `globalThis.fetch` with a native HTTP client that bypasses CORS entirely
- **Blank screen on Linux desktop app** — WebKitGTK's DMA-BUF renderer silently fails on many GPU configurations (Intel/AMD integrated, NVIDIA with certain drivers, VMs, some Wayland compositors), leaving the window completely blank; `WEBKIT_DISABLE_DMABUF_RENDERER=1` is now set automatically on Linux before the Tauri runtime starts to force the reliable compositing fallback; users can override by setting the variable themselves before launching

### Changed
- **CI: Node.js upgraded to 24** in the GitHub Actions release workflow (Node 20 actions were deprecated)

## v0.3.1 — 2026-04-02

### Fixed
- **Code blocks unreadable in dark mode** — inline code had a hard-coded `background: #f1f5f9` (the same near-white as dark-mode text), making code invisible; background is now `var(--color-border)` with an explicit `color: var(--color-text)`; fenced code blocks (`pre`) now use `var(--color-bg)` / `var(--color-text)` with a border; `pre code` resets the background to transparent so the outer block colour wins

## v0.3.0 — 2026-04-02

### Added
- **Close / reopen cards** — a Close Card button in the card detail footer marks a card as closed; closed cards appear on the board with a strikethrough title and reduced opacity and can be reopened at any time; closed cards are included in time reports with a "Closed" badge and strikethrough in the title column
- **Closed cards in time reports** — the report response now carries a `closed` flag per card; closed cards are visually distinguished in the report table (strikethrough + red "Closed" badge) without being excluded from totals
- **Copy card** — a "Copy Card" button in the card detail footer duplicates the card (title, description, priority, due date, labels, tags) in the same column; the copy is appended below the original with "(copy)" appended to the title; board updates in real time for all connected users
- **Transfer card** — a "Transfer…" panel in the card detail lets you copy or move a card to any project you have access to; choose a destination project and column, then click "Copy Here" or "Move Here"; labels and assignees are intentionally not copied (they are project-specific); the originating project board updates instantly when a card is moved away
- **Open card count in Admin → Projects** — the projects table now shows an "Open Cards" column with the number of non-closed cards per project
- **Open card count on project tiles** — the dashboard project grid shows the open card count below each project description

### Fixed
- **Date format on board cards** — due dates were rendered using the UTC date from the ISO timestamp, causing an off-by-one in negative-UTC timezones; the date portion is now sliced before formatting so it matches the user's local calendar date
- **Due date picker ignored configured format** — `<input type="date">` always displays in the OS/browser locale regardless of user settings; replaced with a plain text input that parses and displays dates using the user's configured format (e.g. `DD/MM/YYYY`); a clear button appears when a date is set
- **Spellcheck in card description** — EasyMDE/CodeMirror renders text in its own span-based DOM layer so the browser's native spellchecker cannot reach it regardless of settings; the description editor is now a plain `<textarea>` (markdown is still rendered in preview/read-only mode)
- **Spellcheck in card comments** — same root cause as description; the comment editor is now a plain `<textarea>` with `spellcheck="true"` and the user's locale set as the `lang` attribute
- **Spellcheck on card title** — added `spellcheck="true"` and `lang` to the title input field
- **Session lost on browser close** — auth tokens (access + refresh) and the cached user object were stored in `localStorage`, surviving browser restarts; moved to `sessionStorage` so closing the browser or tab ends the session as expected
- **Project switching in sidebar not updating board** — Vue Router reuses the `BoardView` component when navigating between projects so `onMounted` never fires again; fixed by watching `route.params.slug` and reloading board data, project info, members, and the WebSocket connection when the slug changes; `useWebSocket` now accepts a reactive ref so the connection URL updates correctly
- **Board cards showing light background in dark mode** — `.board-card` had a hard-coded `background: #fff`; replaced with `var(--color-surface)` so it respects the active theme; priority badge colours now also have explicit `[data-theme="dark"]` overrides
- **Report date/time not following configured format** — the "Generated" timestamp and card update dates in the time report used `toLocaleString`, producing browser-locale formatting regardless of user settings; now uses the `useDateFormat` composable so the output matches the user's configured date/time format
- **Report URL printed at bottom of page** — browsers print the page URL in the margin area by default; suppressed via `@page { margin: 0 }` and explicit empty `@top-*` and `@bottom-*` margin box rules
- **PDF export missing pages** — `.app-shell-body { overflow: hidden }` clipped the print output to the visible viewport, truncating multi-page reports; overridden with `overflow: visible; height: auto` in `@media print`
- **Print header duplicated/cut off across pages** — the `position: fixed` per-page header was positioned relative to the CSS content area, overlapping content on pages 2 and onwards; replaced with native CSS `@page` margin boxes: the WarmDesk logo appears inline at the top of page 1, and "WarmDesk" text + page number (`n / total`) appear in the top margin on subsequent pages via `@page @top-left` and `@page @top-right`

## v0.2.10 — 2026-03-29

### Fixed
- **SMTP port saved as number** — `<input type="number">` causes Vue to send the port as a JSON number; the Go struct expected a string and rejected it with an unmarshal error; frontend now coerces to string before sending and the backend field accepts `json.Number` so either format works

## v0.2.9 — 2026-03-29

### Fixed
- **Admin Settings tab blank** — `@` in the SMTP test email placeholder was parsed by vue-i18n as a linked-message prefix, throwing `Invalid linked format` on first render and wiping the admin panel; escaped with `{'@'}` in all five language files
- **JWT token lost on LocalStorage eviction** — access token is now also kept in the axios default header so API calls succeed even if another tab or the browser clears LocalStorage between requests
- **Admin settings errors hidden** — the `loadSettings` error handler was a silent `catch {}`; errors are now shown as toast notifications
- **SMTP password placeholder always shown** — `!!data.smtp_password_set` evaluated a non-empty string `"false"` as truthy; fixed with strict `=== 'true'` comparison
- **Reports menu hidden for admins with stale session** — cached user objects without `can_view_reports` no longer hide the Reports link for admins

## v0.2.8 — 2026-03-29

### Added
- **Webhook URL with live token** — after creating or regenerating a webhook, the setup docs now show the full ready-to-paste URL with the real token already substituted in; falls back to `<token>` placeholder when no token is in view

### Changed
- **Reports access restricted** — time report generation is now limited to project admins/owners and system admins; regular members and viewers no longer see the Reports menu item and are redirected if they navigate directly to `/reports`

## v0.2.7 — 2026-03-29

### Added
- **Git platform integration** — connect GitHub, GitLab, Gitea, or Forgejo via webhooks; push / PR / issue events post formatted messages to the project chat, and any card reference (e.g. `PRJ-42`) in a commit message or PR / issue title automatically creates a link in the card detail; links show platform badge, type (commit / pull request / issue), short reference, title, and open / closed / merged status
- **GitHub webhook** — new webhook type with HMAC-SHA256 signature verification; handles `push`, `pull_request`, `issues`, `create`, `delete`, and `ping` events
- **GitLab webhook** — new webhook type with `X-Gitlab-Token` validation; handles Push Hook, Merge Request Hook, and Issue Hook events
- **Gitea / Forgejo card links** — existing Gitea webhook now also creates card links from commit messages, PR titles, and issue titles (chat posting was already supported)
- **Documentation** — three new Markdown documents shipped with every release in `docs/`:
  - `docs/user-guide.md` — end-user walkthrough of all features
  - `docs/api.md` — Ticket API and all webhook integration reference
  - `docs/admin-guide.md` — installation, configuration, SMTP, scaling, backup, and security checklist

## v0.2.6 — 2026-03-29

### Fixed
- **Windows build** — Tauri v2 removed `zip` as a valid `--bundles` value on Windows (only `msi` and `nsis` are supported); the CI workflow and `make windows-portable` target now build with `--bundles nsis` and create the portable zip from the compiled binary using PowerShell's `Compress-Archive`

## v0.2.5 — 2026-03-29

### Added
- **Emoji picker** — a full emoji picker (8 categories + search) is now available in all chat inputs (project chat, direct messages) and card editors (EasyMDE toolbar button); emojis insert at the cursor position
- **@mention autocomplete** — typing `@` in any chat input or card editor shows a dropdown of matching project members; use arrow keys to navigate, Enter/Tab to complete; mentions also work in card comments
- **Real-time mention notifications** — when a user is @mentioned and is currently online, a purple popup notification appears immediately with the sender's name, context (project chat / card comment / direct message), and a preview of the message; offline users still receive an email
- **Chats sidebar section** — the sidebar now has a collapsible "Chats" section showing the 8 most recently active conversations; each entry shows an unread indicator (pulsing red dot) when there are new messages since the conversation was last viewed
- **SMTP test email** — the admin SMTP settings page has a new "Send Test Email" field; enter any address and click Send to verify that the SMTP configuration works without leaving the admin panel

### Fixed
- **SMTP settings not saving on fresh install** — GORM `Save()` with a non-zero string primary key only issues an UPDATE, silently failing on a new database; replaced all system-setting saves with a proper upsert using `clause.OnConflict`
- **Admin error messages hidden** — the SMTP save error catch block was missing the error parameter, showing a generic fallback message instead of the real server error; now shows the actual API error message
- **Card comments missing @mention notifications** — `CreateComment` was not calling `NotifyMentions`; mentions in card comments now trigger both real-time WS notifications and emails

### Changed
- **"Direct Messages" renamed to "Chats"** — navigation item, page title, and all UI labels updated; the old `/messages` route redirects to `/chats`
- **Team Chat removed from project board** — the slide-in chat panel on the board page has been removed; project chat is accessible via dedicated project pages

## v0.2.4 — 2026-03-29

### Added
- **Project teams in Direct Messages** — new "Teams" tab in the new-conversation panel lists all projects the user belongs to; clicking a project pre-fills all its members and the project name as the group name, ready to start a team chat with one click
- **Project admin role** — new `admin` role between `member` and `owner`; project admins can create, rename, reorder, and delete columns; regular members cannot; board toolbar shows settings gear only to project admins and global admins
- **Group chat avatar** — group conversations can have a custom avatar image; click the group icon in the chat header to upload one
- **Auto-delete empty group chat** — when removing the last non-creator member from a group chat that has no messages, the conversation is deleted automatically and all participants are notified
- **Persistent system admin in seed** — `warmdesk-seed` now creates `tonk` (Ton Kersten) as a system admin account that is never removed by `--reset`
- **More demo users in seed** — four additional demo users (Priya Nair, James O'Brien, Elena Kovač, Raj Sharma) are created; project admin roles are demonstrated across the three demo projects

### Fixed
- **Report assignee dropdown z-index** — placeholder text was visible through the open dropdown; fixed by establishing a stacking context on the filters row

### Changed
- **Board toolbar** — project name replaces the "Project Settings" text link; the settings gear icon is only shown to users who can manage the project

## v0.2.3 — 2026-03-29

### Added
- **Assignee filter on time reports** — the report page now has a multi-select dropdown to filter by one, several, or all assignees; selected names are shown as a summary label; passed to the backend as a comma-separated `assignees` query param
- **Direct message history** — opening a conversation (including via a sidebar user link) now immediately loads all stored messages from the database; history persists across sessions
- **Remove member from group chat** — any group member can remove another member via the × chip next to their name in the chat header; removal is confirmed and broadcast to remaining members via WebSocket
- **Demo conversations in seed** — `warmdesk-seed` now creates 5 conversations with 42 realistic messages (4 one-on-one DMs: Alex↔Sarah, Marc↔Lisa, Sarah↔Lisa, Alex↔Marc; plus a "Website Redesign Team" group chat) with historically-spread timestamps
- **Screenshots in README** — a 2-column screenshot grid has been added to the README covering all main views

### Fixed
- **DM sidebar navigation race condition** — clicking a user in the sidebar while conversations were still loading could create a new blank conversation instead of opening the existing one; the watch handler now waits for both conversations and users to be loaded before calling `openOrCreateDM`

## v0.2.2 — 2026-03-29

### Added
- **Configurable initial columns** — admin can define which columns are created automatically when a new project is made (Admin → Settings → New Project Defaults); one column name per line; defaults to "Backlog"
- **Delete empty column** — a trash icon appears on any column that has no cards; clicking it asks for confirmation and removes the column

### Fixed
- **Version number on login page** — app version is now shown below the login card, matching the footer
- **Frontend version follows git tag** — `__APP_VERSION__` is now derived from `git describe --tags --always` at build time instead of the static `package.json` version; the update-available banner no longer appears falsely after a release
- **Admin sidebar shows all projects** — admins now see all projects in the sidebar, not only the ones they were explicitly added to as a member
- **PDF report shows only the report** — the browser print dialog now hides the sidebar, header, and footer so only the time report content is printed
- **Time format in reports** — changed from "1h 30m" to `H:MM` (e.g. `1:30`, `100:05`); hours are unbounded, minutes are always zero-padded to two digits

### Changed
- Default initial column renamed from "Inbox" to "Backlog"

## v0.2.0 — 2026-03-29

### Added
- **Time Spent on cards** — log hours and minutes directly on a card; stored as `time_spent_minutes` and shown in the card detail dialog
- **Time Report** — new `/reports` page that generates a time overview grouped by project, filterable by period (all time / year / month / ISO week) and by project
- **Export to PDF** — print-optimised layout with company logo and period header; uses the browser's native print-to-PDF
- **Export to Excel (XLSX)** — downloads a formatted spreadsheet via SheetJS; includes ref, title, assignees, date, and time columns with subtotals per project and a grand total
- **Company branding** — admin can set a company name and logo (URL or uploaded image) under Admin → Settings → Branding; both appear on generated reports
- **Demo seed tool** — `warmdesk-seed` binary (included in the distribution) populates the database with four demo users, three projects, 32 cards with labels/assignees/checklists/comments/time, and three discussion topics; run with `--reset` to wipe and re-seed; idempotent on repeated runs
- **CLAUDE.md** — developer guide for AI-assisted development: architecture decisions, conventions, and how to add routes, models, and settings
- **Configurable idle session timeout** — admin setting (default 60 minutes); users are automatically logged out after the configured period of inactivity; set to 0 to disable
- **Update check** — on login the server is compared against the latest GitHub release; a dismissable banner is shown when a newer version is available (web and desktop)

### Fixed
- **SMTP settings could not be saved** — the save button shared a function with all auto-saving dropdowns (theme, timezone, etc.), causing SMTP fields to be sent in every general-settings request and potentially overwriting saved values; SMTP now has its own dedicated save
- **SMTP username and password made optional** — all SMTP credential fields are now pointer types in the backend; omitting them from a request leaves the stored value untouched, allowing auth-less SMTP relay configurations

### Changed
- `warmdesk-seed` is built alongside the main binary by `make build-backend` and included in distribution archives
- System settings handler splits SMTP saves from general settings saves to prevent cross-contamination

## 2026-03-28

### Added
- Tauri desktop app — distributable as AppImage (Linux), DMG (macOS), and installer (Windows)
- Topics — threaded discussions per project with markdown support and replies
- Checklists on cards
- Multiple assignees per card
- Viewer role — read-only access at project and global level
- Favourite people — mark users for quick access
- Card watchers — subscribe to card activity notifications
- Card sorting within columns by due date, assignee, or priority
- Direct message notifications
- Group direct messages
- Admin can assign users to projects directly
- Admin can reset user passwords

### Fixed
- Topics view was rendering its own header, causing duplicate search bar, language selector, and avatar
- Adding a new card showed it twice until page refresh (duplicate WebSocket event handling)
- Logo and favicon not served correctly
- Build artifacts (AppImage, DMG, Windows installer, Rust target/) excluded from git via .gitignore

### Changed
- Group DMs, markdown in chat, i18n expansion, and UI polish
