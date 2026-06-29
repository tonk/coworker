# Changelog

## v0.12.37 — 2026-06-29

### Added
- **Customer hide/unhide** — customers can be hidden (filtered from list views and sidebar by default) and unhidden via Admin → Customers. A "Show hidden" checkbox controls visibility in the admin table. Projects belonging to hidden customers are excluded from project lists.
- **Customer `is_hidden` i18n** — all 12 frontend locales translated for hide/unhide and show hidden actions.

### Fixed
- **Admin customer hide/unhide no longer zeroes counts** — `toggleHideCustomer` now updates `is_hidden` in-place on the `CustomerListItem` instead of replacing the whole object with the bare `Customer` response, preserving `contract_count`, `project_count`, `my_role`, and other list-only fields.

## v0.12.36 — 2026-06-29

### Added
- **Close board / project closed state** — projects can now be closed (hidden from sidebar and project lists) and reopened from Project Settings → Danger Zone or Admin Panel → Projects.
- **Admin show/hide closed projects** — "Show closed" toggle in Admin → Projects table filters the list; closed projects show a "Closed" badge.
- **Closed board banner** — a yellow banner appears on closed boards with a link to reopen from Project Settings.
- **Closed project i18n** — all 12 frontend locales translated for close/reopen board actions and closed status labels.

### Changed
- **ListProjects filters closed** — `GET /api/v1/projects` excludes `is_closed = true` by default; use `?include_closed=true` to include them.
- **ListStarredProjects filters closed** — starred/favorite endpoints also exclude closed projects.
- **AdminListProjects closed param** — supports `?closed=true` (only closed) and `?closed=hide` (only open).
- **UpdateProject no longer requires customer_id** — partial updates (e.g. toggling `is_closed`) no longer need to include `customer_id`.
- **Close/reopen moved to Project Settings** — the close button is in the Danger Zone (project owners/admins only) rather than the board toolbar.
- **User & admin guides updated** — PDF guides document the new close/reopen feature.

### Documentation
- **CLAUDE.md** — project closed state section documenting API filtering and admin options.
- **User guide** — close/reopen documented in Project Settings with sidebar visibility note.
- **Admin guide** — close/reopen, show/hide closed, and status badges documented.

## v0.12.35 — 2026-06-29

### Added
- **Time macro server sync** — macro libraries are stored per user on the server (`GET/PUT /api/v1/time-entries/macro-library`) with a browser cache fallback; seed data includes sample macros for demo users.
- **Macro run popout** — bottom-bar *⚡ Macro* opens a run panel (pick macro, choose *Start on*, preview, run); top-bar *⚡* opens the editor next to contract rates.
- **Macro i18n** — macro UI strings translated in all 12 frontend languages.

### Fixed
- **Macro JSON export (Tauri)** — desktop app uses `triggerDownload` so export works on Linux WebKit.
- **Macro time parsing** — `HH:MM` values in macro templates parse correctly even when the user prefers decimal hour notation.
- **Time sheet comment i18n** — row-comment add/edit/placeholder strings and `common.double_click_edit` are translated in every locale.

### Changed
- **Macro editor UX** — run and edit are separate; macro dropdown sorted alphabetically; time/distance popups use accent-coloured borders.
- **User guide** — time macros section documents server sync, run popout, and the top-bar editor.

### Documentation
- **Website** — release blog post for v0.12.35.

## v0.12.34 — 2026-06-29

### Changed
- **Time macro apply options** — choose how many weekdays to fill (1–7, default 5) and toggle *Alternating A/B pattern* off for the same values every day; when off the editor shows a single hours/start/end/distance column per row; when on, pattern A and B columns return; apply settings are saved per macro.

### Documentation
- **User guide** — time macros section documents apply-to-days, alternating toggle, and per-day vs A/B layout.
- **Website** — release blog posts for v0.12.33 and v0.12.34.

## v0.12.33 — 2026-06-29

### Added
- **Time tracking macros** — reusable weekly templates (multiple named macros, variable rows, day 1/day 2 presets with optional start/end times and distance); apply to the first *N* days of the week with alternating day-1/day-2 pattern; import/export as JSON; stored per browser.
- **Show/hide closed** — Kanban board toolbar toggle hides closed cards per column (with count badge); ticket list and inbox views get a matching toggle alongside the spam filter; preference persisted in `localStorage`.

### Fixed
- **Breadcrumbs** — correct labels and casing for tickets, inbox, epics, charts, news, invoices, and time-tracking routes; intermediate segments for nested ticket URLs.

### Changed
- **Time tracking sheet editing** — double-click the customer/project or activity cell to enter row edit mode (same as ✎ in the Actions column).
- **Time tracking copy/cut/paste** — `Ctrl+C` copies the full cell (not partial text); `Ctrl+X` cuts; rectangular selections work across rows; clipboard persists across week navigation; paste supports multi-cell blocks and cross-week day columns; today's column shows selection highlight correctly.

### Documentation
- **User guide** — documents row editing, cell selection, copy/cut/paste, and time macros on the Log Time tab.

## v0.12.32 — 2026-06-25

### Fixed
- **Time tracking copy/paste** — copying a cell now includes distance, start/end time, and holiday flag; all fields are pasted and correctly restored on undo.

### Changed
- **Holiday country flags** — emoji flags in the browser holidays dropdown are restored.

## v0.12.31 — 2026-06-19

### Added
- **Hideable sidebar** — hover the sidebar edge to reveal a collapse button; a thin reveal strip brings it back. State is persisted across sessions.

### Fixed
- **Desktop context menu** — the WebKit right-click menu (Back, Reload, Inspect Element) is suppressed in the desktop app. Browser builds are unaffected.

## v0.12.30 — 2026-06-19

### Added
- **Global invoices view** — admins now have a dedicated *Invoices* tab in the main navigation that lists invoices across all customers with status filtering and overdue highlighting.
- **Ansible collection modules** — new `customer_contact`, `invoice`, and `invoice_template` modules for automating helpdesk and billing resources via the `ansilabnl.warmdesk` collection.
- **Molecule test suite** — the Ansible collection ships a full Molecule scenario (`extensions/molecule/default/`) that spins up a live WarmDesk server and exercises create / idempotency / update / absent lifecycle tests for every module.

### Fixed
- **Invoices view access control** — the global invoices list is now restricted to admin users only; non-admin users no longer see the navigation entry.

### Changed
- **Account Settings layout** — the settings page is split into clearly labelled tabs (Profile, Security, Notifications, API Keys) to reduce scrolling and make individual sections easier to find.
- **Text selection in dark and black themes** — selected text now uses a theme-aware highlight colour instead of the browser default, which was unreadable against dark backgrounds.

### Documentation
- **API reference** — sections 8–10 added for Invoices, Invoice Templates, and Customer Contacts, including all endpoints, request/response schemas, and Bruno collection requests.
- **Admin and user guides** — expanded Invoices chapter (credit notes, payment methods, overdue tracking), new Contacts sub-section, new Invoice Templates section.

## v0.12.21 — 2026-06-18

### Fixed
- **Card delete confirmation** — the confirmation dialog no longer appears behind the open card detail modal.

## v0.12.20 — 2026-06-18

### Added
- **Black theme in quick switcher** — the pure black (AMOLED) theme is now available directly from the header theme menu alongside Light, Dark, and System, without needing to open Settings.

## v0.12.19 — 2026-06-17

### Added
- **Travel distance per time entry** — each cell in the weekly time sheet now has a distance field (km or miles). Click the ⇆ icon in the cell to set the distance for that entry.
- **Distance unit preference** — user profile settings now include a km / miles selector. The chosen unit is used across all views and exports.
- **Distance in PDF export** — when *Show distance* is checked in the PDF export options, distance appears as a column in every entry row, group subtotal row, and grand total row so customers can be billed per trip.
- **Distance in XLSX export** — distance column added to the spreadsheet export with per-entry values, group subtotals, and grand total.
- **Distance totals** — the weekly time sheet footer and the report tab both show the total distance travelled when any entry has a distance logged.

## v0.12.18 — 2026-06-12

### Added
- **`warmdesk-timetracking.desktop`** — deb and rpm packages now install a dedicated timetracking launcher entry with `Icon=warmdesk-timetracking`, making it easy to pin a timetracking-branded shortcut alongside the main WarmDesk entry.

### Changed
- **Admin settings in timetracking mode** — Scrum Story Points, New project defaults, Email (SMTP), and Incoming Mail (IMAP) sections are now hidden when the server runs in `--mode=timetracking`.
- **User settings in timetracking mode** — Email Notifications and Personal API Keys are now hidden in timetracking mode.

### Fixed
- **Blog post sort order** — posts sharing the same date now sort with the newest release first (filename descending as tiebreaker) instead of oldest first.

## v0.12.17 — 2026-06-12

### Fixed
- **Logo switching in timetracking mode (web UI)** — all 13 logo `<img>` elements in the frontend (header, login, register, forgot/reset password, connect, about, news modals, and print headers) now switch to the time-tracking variants when the server runs in `--mode=timetracking`. Previously only the Go backend's static-file remapping handled this, which had no effect in dev mode or on surfaces that bypassed the NoRoute handler.

### Changed
- **deb/rpm icon tree** — both `warmdesk` and `warmdesk-timetracking` icon sets are now fully installed into the system icon tree: scalable SVG plus PNGs at 32×32, 128×128, 256×256, and 512×512. A custom `.desktop` file can reference `Icon=warmdesk-timetracking` without the icon being buried inside the binary.

## v0.12.16 — 2026-06-12

### Added
- **Time-tracking logo set** — new clock-on-desk icon (`timetracking.svg`) and full wordmark (`timetracking-full.svg`) in the same geometric green style as the WarmDesk logo.
- **`--mode=timetracking` logo branding** — when the server runs in time-tracking-only mode, the time-tracking logos replace the default WarmDesk logos everywhere: web UI, login screen, and exported PDFs (when no custom company logo is configured).

### Changed
- **Admin guide** — new *Time-tracking-only mode* sub-section documents `app_mode`, its three setting methods, and all effects.
- **User guide** — NOTE at the top of the Time Tracking section explains that boards/chat/helpdesk may be absent in time-tracking-only deployments.

## v0.12.15 — 2026-06-12

### Added
- **Website release blog** — `release-0.12.14` post documents guide PDF downloads, copy-week safety, and time-grid polish.

### Changed
- **Destructive confirm dialogs** — delete, revoke, purge, leave/remove, and similar actions use `{ destructive: true }` for a red *Delete* button app-wide; sprint start/complete and Copy Previous Week keep a primary *Yes* label.
- **Versioned guide PDF filenames** — avatar menu and API downloads use `WarmDesk-user-guide-vX.Y.Z.pdf` and `WarmDesk-admin-guide-vX.Y.Z.pdf` matching the running release.
- **`scripts/release bump`** — runs `make sync-doc-revisions` so guide `:revnumber:` / `:revdate:` track `CHANGELOG.md` automatically.
- **Admin guide PDF layout** — remaining long configuration, backup, and shell example blocks reformatted for page width.
- **User guide** — icons table documents Copy Previous Week confirmation, undo (50-action stack), and muted undeclarable-time styling.

### Fixed
- **CI PDF build** — install `asciidoctor-pdf` with `sudo gem install` in server workflows to avoid permission errors.

### Removed
- *(none)*

## v0.12.14 — 2026-06-12

### Added
- **Guide PDF downloads** — user and admin guide PDFs are built on every server release, embedded in the binary, and downloadable from the avatar menu → **Downloads** (admin guide for admins only).
- **Branded documentation PDFs** — `make docs-pdf-guides` produces PDFs with the WarmDesk title page, logo, revision metadata, and bundled fonts.
- **Weekly working hours total** — User Settings shows the total net hours per week below the working hours grid.
- **Copy Previous Week confirmation** — prompts before adding rows when the grid is not empty; the copy action is undoable.

### Changed
- **Confirm dialog** — the primary button defaults to *Yes*; destructive flows can pass `{ destructive: true }` for a red *Delete* label.
- **Time-tracking undeclarable styling** — grid totals, report undeclarable lines, and project undeclarable badges use muted text instead of red.
- **Admin guide PDF layout** — long configuration and database example blocks reformatted to fit the page width.
- **User guide** — icons & indicators overview table; avatar menu documents the Downloads submenu.

### Removed
- **`warmdesk-blog.adoc`** — marketing blog document removed from the docs tree.

## v0.12.13 — 2026-06-11

### Added
- **Working hours start & end time** — the Working Hours section in User Settings now has a Start and End column per day instead of a single start time. The calculated net hours are shown in a third column.
- **Lunch break setting** — a single "Lunch break" field (default 30 minutes, 0–120 range) is subtracted from each day's net hours calculation.

### Changed
- **Docs converted to AsciiDoc** — `user-guide` and `admin-guide` are now `.adoc` files, enabling proper `asciidoctor-pdf` output and native GitHub rendering from the same source.

## v0.12.12 — 2026-06-11

### Fixed
- **Fill-from-slots: overnight-continuation days now get correct start/end times** — multi-day overnight slots (e.g. a Friday standby running until Monday 07:00) cover Saturday and Sunday via their overflow extension. Previously those continuation days got no start/end time stored, so the dot indicator was missing and the time popup showed the 09:00–17:00 placeholder. They now correctly store 00:00–00:00 (full-day coverage).

### Changed
- **Seed: standby entries pre-filled for current week** — `warmdesk-seed --reset` now creates Mon–Sun standby entries for the Acme Phase 1 contract so the time-tracking grid immediately demonstrates the rate column, dot indicators, and special-rate slots.

## v0.12.11 — 2026-06-11

### Added
- **Rate column in the weekly time log** — the Log Time grid now shows the contract's hourly rate (e.g. "45 €/h") for each row. A ✦ badge indicates that the contract has time-slot rate overrides; click the ₡ toolbar button for the full breakdown.
- **Syslog output** — the backend now writes all log output to the system syslog (`LOG_DAEMON`, tag `warmdesk`) in addition to stderr, making it easy to capture server and auth events (logins, failures, MFA activity) via standard log aggregation tools.

### Changed
- **Startup message in time-tracking mode** — when started with `--mode=timetracking`, the server now prints `Starting WarmDesk - Time Tracking <version>` instead of the generic startup line.

## v0.12.10 — 2026-06-10

### Added
- **Contract time slot duration summary** — the contract time slot editor now shows the duration each slot occupies directly after the day selector. For slots active on a single day the duration is shown once; for slots spanning multiple days each active day is listed individually (e.g. Mon: 8h Tue: 8h Fri: 8h) followed by the weekly total.
- **Contract selection in time grid** — the add/edit row dialogs in the time-tracking grid now include a contract dropdown. The selected contract is stored on each time entry and displayed in the row in the primary colour. A ⏱ "Fill from slots" button pre-fills the entire week's hours based on the contract's time slot definitions.

## v0.12.9 — 2026-06-10

### Fixed
- **Timetracking mode — Board report tab hidden** — the Board tab in the Time Tracking view is no longer shown when the server runs in `--mode=timetracking`, as boards are unavailable in that mode.
- **Timetracking mode — browser tab title** — the browser tab now reads "WarmDesk - Time Tracking" instead of "WarmDesk" when running in `--mode=timetracking`. The title updates immediately on load and correctly prefixes the `●` unread indicator when notifications are present.

## v0.12.8 — 2026-06-10

### Added
- **`--port` CLI flag** — override the configured port per instance, allowing two processes to share a single config file (e.g. `./warmdesk --config warmdesk.yaml` + `./warmdesk --config warmdesk.yaml --port 8081 --mode=timetracking`).
- **SQLite WAL mode** — `PRAGMA journal_mode=WAL` and `PRAGMA busy_timeout=5000` are applied on startup for SQLite databases, allowing two WarmDesk instances to safely share the same database file without "database is locked" errors.
- **Login screen mode indicator** — when running with `--mode=timetracking` the login page shows a "Time tracking only" pill below the sign-in heading (localised in all 12 supported languages).

### Fixed
- **Tauri: timetracking mode not detected after server switch** — `app_mode` is now re-fetched in `ConnectView` immediately after a new server URL is saved, and again in `LoginView` on mount, so the correct UI mode is applied even when switching from a full server to a time-tracking-only server in the desktop app.

## v0.12.7 — 2026-06-10

### Added
- **Multi-day contract time slot selection** — the day-type dropdown on contract time slots is replaced by a compact row of seven toggle buttons (Mo Tu We Th Fr Sa Su). Any combination of days can be selected in a single slot; the selection normalises to the `all`, `weekdays`, or `weekends` preset when it matches exactly. This makes schedules like "Mon–Thu 19:00 → 07:00 + Fri 19:00 → Mon 07:00" expressible in two slots instead of five.
- **`--mode=timetracking` server flag** — start the server with `--mode=timetracking` (or `APP_MODE=timetracking`) to run a stripped-down instance that exposes only time-tracking routes (customers, contracts, time entries, reports). Boards, chat, helpdesk, sprints, and all associated admin configuration are disabled at the API level. The frontend detects the mode from `GET /version` before first render and hides the sidebar, search bar, non-time-tracking nav links, and irrelevant admin tabs. The landing page redirects to `/time-tracking`. Designed for a shared-database setup alongside a full WarmDesk instance.

## v0.12.6 — 2026-06-10

### Fixed
- **Contract time slot week preview** — slots after the first now always show their week preview. Equal start/end times (e.g. `07:00 → 07:00`) are treated as a 24-hour cycle and render correctly. Each slot's preview bar uses a distinct colour (primary → amber → emerald → violet → red → teal) so multiple slots are easy to tell apart.

## v0.12.5 — 2026-06-10

### Added
- **Contract Rates Overview** — a ₡ button in the Customers page header and the Time Tracking toolbar opens a modal listing every contract time slot across all accessible customers, grouped by customer → contract, so standby rates are visible at a glance without navigating into each customer.
- New API endpoint `GET /api/v1/customers/rates` backing the above; returns customers with contracts that have ≥1 time slot, filtered by the same access control as `GET /customers`.

### Fixed
- **Database schema documentation** — `contracts` table entry had stale columns (`budget_hours`, `status`, `deleted_at`) and was missing `price_per_hour` and `currency`; `customers` table was missing `logo_url`, `position`, `time_tracking_only`, `created_by_id`; `customer_favorites` and `customer_access` showed a surrogate `id` PK that does not exist; `contract_time_slots` table was absent entirely. All corrected.

## v0.12.4 — 2026-06-10

### Added
- **Per-user audit trail** — the admin Edit User panel now has a Login History button showing a full audit log: successful and failed login attempts (password, passkey, MFA), logoffs, password changes, MFA events (challenge, verify, trusted-device skip, disable), passkey registration/deletion, API key lifecycle, email changes, and all admin actions on the account. Each event records timestamp, IP address, client string, and — for admin-on-behalf actions — the actor who performed it.

### Fixed
- **RBAC** — `customer`-role users were able to update and delete inbox tickets; both handlers now enforce `requireNotCustomerRole`. Any authenticated user could read contracts for any customer (`ListContracts` IDOR); the endpoint now requires `requireCustomerAccess`. Project-scoped API keys could be created or deleted by any project member; the required role is now `owner`.

### Changed
- **Ansible collection v0.5.0** — new `epic`, `sprint`, and `ticket_checklist_template` modules; `user` module gains `state: restore` (un-soft-delete) and `state: purge` (permanent removal with FK cleanup).

## v0.12.3 — 2026-06-09

### Fixed
- **Time-tracking time-field input** — typing `:` manually in a start/end time field (popup, standby form, or grid cell) no longer produces a doubled separator (`19::0`); the explicit colon is absorbed cleanly. Entering a bare hour value (e.g. `20`) and pressing Enter is now accepted as `20:00` across all three input surfaces.

## v0.12.2 — 2026-06-06

### Added
- **Row-level comments on time tracking grid** — add a comment to any row in the weekly timesheet via the 💬 button; comments are persisted per ISO week alongside the row order, loaded on week navigation, and rendered below the row label in PDF exports. Demo seed populates comments for Ton Kersten and Alex Admin.

### Fixed
- **Grid PDF comment row misalignment** — `SetX` → `SetXY` in `buildWeekGridPDF` so overlay `CellFormat` calls no longer push day cells to the wrong Y position, fixing broken grid lines, fills, and cell values.
- **Call button shown when user is offline** — the 1:1 call button in Direct Messages is now disabled with reduced opacity when the other user is offline, with a "User is offline" tooltip.

## v0.12.1 — 2026-06-06

### Added
- **MFA remember-devices admin policy** — admins can disable trusted devices, allow 1-week trust only, or allow 1 week or 1 month (default). Tightening the policy revokes trusts that no longer comply. Passkey login and the Tauri desktop app honour the same policy.
- **Screenshot 24 — undeclarable grid alignment** — Playwright captures the weekly timesheet grid with undeclarable deductions for documentation; `screenshots.sh` re-seeds with `--reset` so dates stay current.

### Fixed
- **Undeclarable time alignment** — red undeclarable amounts in day cells right-align with entered time; row totals now show declarable time (matching footer totals); `-` prefix used consistently in the UI and PDF exports.
- **Timesheet accessibility** — day cells with undeclarable time expose the deduction to screen readers via `aria-describedby`.
- **Page help** — time-tracking sheet help resolves array-valued i18n keys; modal body scrolls when content overflows; undeclarable-time help explains declarable vs logged totals.
- **E2E login** — auth cookies no longer force `Secure` on plain HTTP in release mode (browsers were rejecting them, breaking form login); Playwright clears stale MFA-trust storage and falls back to API login on rate limit.
- **Reverse-proxy auth cookies** — `cookieSecure()` now keys off direct TLS or `X-Forwarded-Proto: https` only; deploy templates and `warmdesk.yaml.example` document the requirement.

## v0.12.0 — 2026-06-06

### Added
- **MFA trusted devices** — after completing an MFA challenge, users can choose to trust their device for 1 week or 1 month. Subsequent logins from that device skip the TOTP prompt. The Settings page (Security → Trusted Devices) lists all active trusted devices with last-used and expiry dates, and provides individual and bulk revoke buttons. Logging out automatically revokes the trust record for that device.

### Fixed
- **Grid PDF date format** — the week-grid period label and year-grid print date now follow the user's date/time format setting instead of always using DD-MM(-YYYY).
- **Grid PDF time notation** — cell values, totals, and undeclarable rows in week, month, and year grid exports now use the user's time notation setting (decimal or hh:mm) instead of always rendering decimal hours.
- **Inline help icon positioning** — help icon correctly placed next to the undeclarable badge; service accounts hidden from the sidebar member list.

## v0.11.0 — 2026-06-05

### Added
- **Prometheus metrics integration** — a new `metrics` global role allows dedicated read-only scraper accounts that can only access `GET /api/v1/metrics`. The `Authorization: ApiKey <key>` header is now accepted so Prometheus `scrape_configs` work natively with `authorization: {type: ApiKey, credentials: …}` without needing arbitrary header support. The metrics endpoint records each scrape's timestamp and outcome, exposed as `warmdesk_metrics_last_access_timestamp_seconds` and `warmdesk_metrics_last_access_success` gauges. The last-access timestamp in the admin panel now respects the user's date/time format setting.
- **Admin API key management** — admins can create, list, and revoke API keys on behalf of any user directly from the edit-user modal, making it straightforward to issue keys for service accounts.
- **Soft-delete and restore users** — deleted users are soft-deleted and can be restored from the admin panel ("Show deleted" toggle). Permanent purge nullifies FK references on content records (tickets, cards, messages) while removing personal and membership rows.
- **Improved in-app contextual help** — page-level help entries for Project Settings, System Settings (admin), and Customers & Contracts have been rewritten with full task coverage across all tabs and sub-sections. Eleven new inline field hints (ⓘ) added to the most complex fields: IMAP poll interval, IMAP auth method, IMAP OAuth2 provider, backup schedule, password minimum length, webhook type, webhook secret, and all four contract time-slot fields. All changes available in all 12 supported languages.

### Fixed
- **HelpIcon popover clipping** — help popovers are now teleported to `<body>` to prevent clipping by sidebar overflow; an `align` prop controls left/right alignment so popovers near screen edges stay fully visible.

## v0.10.41 — 2026-06-04

### Added
- **Desktop app profiles** — the Tauri desktop app now supports multiple named profiles, each with its own isolated localStorage, login session, and preferences. Useful when connecting to several WarmDesk servers (e.g. one per customer). Profiles are managed entirely from the command line:
  - `--list-profiles` — list all profiles; the default is marked with `*`
  - `--create-profile <name> [--label <label>]` — create a new profile
  - `--set-default <name>` — change the default profile
  - `--delete-profile <name>` — remove a profile entry
  - `--profile <name>` — launch with the named profile
  - The window title shows **"WarmDesk — \<label\>"** for non-default profiles so multiple instances can be told apart at a glance.
  - Profile configuration is stored in `profiles.json` in the platform config directory; profile data lives under `<data dir>/profiles/<name>/`.
- **Extended in-app contextual help** — the header "?" button now provides page-level help on the Kanban board, inbox, and all six project settings tabs (general, members, labels, API keys, webhooks, deleted cards). Charts view switches help context per chart type (velocity, burndown, CFD, sprint report, epic burndown, release burndown, cycle time, lead time). Time tracking gains three inline field hints: undeclarable time, report group-by, and the export button. All new help content is available in all 12 supported languages.

## v0.10.40 — 2026-06-03

### Added
- **Undeclarable time shown inline in the weekly timesheet grid** — when a project has undeclarable minutes configured, each day cell now shows the undeclarable amount in red below the logged time, and the footer total row shows net declarable time with the undeclarable deducted and highlighted in red.

### Changed
- **Grid PDF exports show declarable vs. undeclarable** — the week, month, and year grid PDFs now display declarable time as the primary value in each cell and total column. A red undeclarable row appears below the totals row. Per-cell undeclarable amounts are shown in the week PDF via absolute positioning, keeping the cell borders intact.

## v0.10.39 — 2026-06-02

### Added
- **Spam badge on Show Spam button** — a red count badge appears on the "Show Spam" button in ticket list and inbox views whenever spam tickets are hidden, so agents know spam is waiting without having to toggle the filter.
- **Custom accent colour** — a rainbow colour-picker swatch in Settings lets users choose any colour as their accent, beyond the four built-in presets; the colour is applied live via CSS custom properties.
- **Ticket list sections in card and list views** — both the customer ticket view and the inbox now show explicit per-status sections (New / Open / Pending / Pending reminders / Pending close / Closed) in card and list view, matching the group view layout; section dividers use the user's accent colour.

### Fixed
- **Show Spam toggle re-fetched unnecessarily** — toggling the spam filter in ticket list and inbox now filters client-side without an extra network request; the full ticket list (including spam) is loaded once per page load.
- **Spam tickets missing from seed** — the demo seed now includes two spam tickets (a phishing email and an SEO pitch) so the spam workflow can be demonstrated.

### Changed
- **Admin panel tab grids** — SLA Policies, Macros, and Checklists tabs now use the same grid styling (borders, header background, cell padding) as the Time Tracking tab.
- **"Add Project" / "Add Customer" renamed to "New Project" / "New Customer"** across all 12 locale files for consistency with the rest of the UI.
- **Dark-mode ticket badge colours** — type, priority, and status badges in the customer ticket list now have proper dark-mode colour overrides (parity with the inbox).

## v0.10.38 — 2026-06-02

### Added
- **Persist empty timesheet rows** — blank rows in the Log Time grid are no longer removed when all hours are cleared; row layout is saved per ISO week and restored when you return to that week.
- **Copy Previous Week includes empty rows** — the copy action brings over the previous week's full row layout (including blank separator rows) in the same order, not just rows that had time entries.
- **Standby time ranges in week grid PDF** — cells with start/end times show them below the hours in a smaller indigo label; overnight standby shifts render as separate two-day spans with start time on the first day, end time on the second, and a connector line between them.

### Fixed
- **Week grid PDF layout broken by time labels** — drawing time annotations no longer disrupts the PDF cursor, so day columns stay continuous and the total column aligns correctly.
- **Week grid PDF merged consecutive standby shifts** — back-to-back overnight shifts (e.g. Tue 19:00→Wed 07:00 and Wed 19:00→Thu 07:00) no longer collapse into one long span; each shift gets its own line and readable start/end labels.

## v0.10.37 — 2026-06-01

### Added
- **Epics for Scrum projects** — colour-coded milestones that group cards across sprints. Manage epics in the new ⚡ Epics view (create, reorder, expand to see cards inline); assign a card to an epic from the card detail Epic dropdown. Cards show a colour bar and badge on the board and backlog. Epics have Open / Done status.
- **Epic Burndown chart** — new chart tab showing remaining cards and story points per day from epic creation to today, with an ideal burn-down line. Select any epic from the dropdown.
- **Sprint Report chart** — new chart tab showing completed vs not-completed cards for any sprint, with story-point totals and a completion percentage. Replaces manual post-sprint counting.
- **Control Chart enhancement** — the Cycle Time scatter plot now overlays a 7-day rolling average line, making trend analysis easier.
- **Backlog drag-to-reorder** — cards in the Product Backlog can be reordered by dragging the ⠿ handle; new position is persisted immediately.
- **Sprint list drag-to-reorder** — sprints in the Backlog view can be reordered by dragging; a △▽ sort button toggles ascending / descending sort by sprint ID, disabling drag while active.
- **Drag-to-reorder in admin editors** — macro actions and checklist template items in the Admin panel can now be reordered by dragging the ⠿ handle.
- **Click-to-edit names in all admin lists** — clicking the name in Users, Groups, Projects, News, and Time Tracking project/customer tables opens the edit form directly, matching the existing Macros and Checklist Templates behaviour.
- **Platform-specific update download URL** — the desktop app now detects its installation method (AppImage, .deb, .rpm, portable tar.gz, DMG, Windows) and links to the matching binary in the update banner. Linux detection uses `/etc/os-release` OS family plus `dpkg --search` / `rpm -qf` ownership queries; covers all distros from the Ansible OS\_FAMILY\_MAP.
- **Seed: epics with card assignments** — `product-platform` and `api-platform` demo projects now have five epics each with cards pre-assigned and back-dated `created_at` values so the Epic Burndown chart shows meaningful history.
- **Blog post** — `website/content/blog/threaded-replies-and-scrum-epics.adoc` covers the new threaded replies and Epics features.
- **Docs: Scrum Board section** — `docs/user-guide.md` gains a new §4b covering Epics, Product Backlog, Sprints, Sprint Board, and Sprint Report.

### Fixed
- **Epic badge lost on board refresh** — `GetProject` was missing `Preload("Columns.Cards.Epic")`; the colour bar and badge disappeared after a page refresh even though the data was present after a card save.
- **Epics page "Board" link** — used a non-existent i18n key `board.board`; replaced with the hardcoded string used by every other toolbar.
- **Sort button invisible** — the sprint sort button had `opacity: 0.35` at the button level combined with `opacity: 0.25` on the span, making it effectively invisible. Removed button-level opacity; inactive triangles now render at 40 % muted.
- **Admin: TicketChecklistTemplatesTab used native `confirm()`** — replaced with `ui.confirm()` to match every other admin section.
- **Admin: button style inconsistency** — Users, Projects, Groups, and Customers used `btn-secondary` / `btn-danger`; now unified to `btn-ghost` / `btn-ghost btn-danger` matching all other admin tabs.
- **Admin: Time Tracking add buttons below table** — moved to the tab toolbar above the table and changed to `btn-primary`, matching all other "New" buttons.
- **Admin: News empty state was a `<div>` outside the table** — replaced with a `<tr><td colspan="5">` inside tbody, matching every other admin table.
- **Ticket detail max-width** — was 900 px (≈62 % of a 1536 px screen); raised to 1 400 px so the Apply Macro and Apply Checklist dropdowns are no longer cut off.
- **Sort indicators** — `BoardColumn` used `↑`/`↓` arrows and `NewsView` used SVG chevrons; both now use `△`/`▽` outline triangles matching every other sort indicator.
- **Button/action label capitalisation** — 28 English i18n strings corrected to standard title case (e.g. "Add item" → "Add Item", "Log time" → "Log Time").
- **Docs: ticket checklist section** — described "Apply Template" as a button in the checklist section; corrected to the "Apply Checklist" dropdown in the ticket header; documented one-template-per-ticket limit, per-item delete, and drag-to-reorder.
- **Docs: ticket messages section** — "Reply" table row implied a single compose area; split into new-message vs inline threaded reply; added Threaded Replies and private-inheritance sections.

### Changed
- **Linux installation-method detection** — now reads `/etc/os-release` to identify the OS family (Debian/RedHat), then runs `dpkg --search` or `rpm -qf` on the canonical exe path to confirm package ownership. Previously relied on the exe path starting with `/usr/lib/` which was incorrect for most package managers. Covers all distros in the Ansible OS\_FAMILY\_MAP including Alma Linux, Rocky, Amazon Linux 2, Oracle Linux, Kylin, MIRACLE, EuroLinux, and others.

## v0.10.36 — 2026-06-01

### Added
- **Threaded ticket messages** — messages can now be replied to at any depth, building a nested thread tree. The reply form appears inline directly below the message being replied to. Nesting is unlimited: the backend loads all messages flat and builds the tree in-memory instead of relying on fixed-depth GORM Preload chains.
- **Private reply inheritance** — replying to a private (internal) message automatically marks the reply as private, keeping internal notes contained.

### Fixed
- **Deep nesting persistence** — replies deeper than 1 level were lost on page refresh because the forward-iteration tree builder snapshot-copied children before their own subtrees were populated. Fixed by building the tree bottom-up (reverse iteration) so every child's subtree is fully assembled before it's copied into its parent.

## v0.10.35 — 2026-05-31

### Added
- **Customer portal role** — new `customer` global role lets end-customers log in and view/comment on tickets for their assigned customers; blocked from boards, chat, and time tracking; private (internal) notes are hidden from them; customer-role users cannot create, update, or delete tickets or apply macros/tags.
- **Private ticket messages** — a **Private** checkbox on the reply form marks a comment as an internal note: not emailed to the ticket's original sender, hidden from users with the `customer` role, and displayed with an amber border and 🔒 badge.
- **Expanded Prometheus metrics** — `GET /api/v1/metrics` now also exposes `warmdesk_users_total{role,active}`, `warmdesk_customers_total`, `warmdesk_tickets_total{status}`, `warmdesk_tickets_by_priority_total{priority}`, `warmdesk_sla_breaches_total{type}`, and `warmdesk_ticket_messages_total{visibility}` in addition to the existing project/card/backup metrics.
- **Prometheus scrape config** — `docs/prometheus.yml` is a ready-made scrape config for WarmDesk (basic auth, 30 s interval, instance relabel).
- **Grafana dashboard** — `docs/grafana-dashboard.json` is a pre-built dashboard covering all metrics across four rows: Projects & Cards, Helpdesk, Users, and Backup Health.
- **Seed: metrics and customer-portal demo accounts** — `demo.metrics` (metrics role, for Prometheus scraping), `demo.cust1` (Alice Porter, customer role, assigned to Acme Corporation), and `demo.cust2` (Bob Mason, customer role, assigned to Globex Systems and Initech Ltd).

## v0.10.34 — 2026-05-30

### Added
- **Ticket viewers** — a row of avatars at the bottom of every ticket (including inbox tickets) shows who has viewed it and when; hover an avatar for the full name and last-viewed timestamp.
- **Admin search boxes** — live search input on the Users, Groups, Customers, and Projects tabs in the Admin panel; filters all visible rows as you type.
- **"Show inactive" toggle** — Admin → Users now hides inactive users by default; a "Show inactive" checkbox restores them, matching the "Show deleted" pattern on the Projects tab.
- **Seed: 8 previously empty feature areas** — the demo seed now populates project chat messages, ticket tags, ticket-to-ticket links, ticket checklist items, ticket history trails, demo file attachment records, emoji reactions on messages, and project webhooks.

### Fixed
- **Date/time format consistency** — five places were showing raw ISO timestamps or using native `<input type="date/time">` elements that ignored the user's format setting: board report "Updated" column, time-tracking personal report date column, Charts view release target-date picker, and contract slot start/end time inputs. All now respect the user's format via `formatDate`, `DatePicker`, or format-aware text inputs.
- **Backup schedule start time** — the "Start time" input was a native `<input type="time">` that bypassed the user's 12 h/24 h preference; replaced with a text input using parse-and-reformat logic on blur.
- **Grid PDF customer/project label column** — label columns were truncated too aggressively in all three grid types. Year and week now use the full page width; month computes label width dynamically from the actual days in the month; truncation limits recalibrated from measured font metrics.
- **SLA policies form** — replaced cramped inline table-row edit (six fields on one line) with a card-above-table layout matching the Macros tab: labeled rows, sensible input widths, and a Save/Cancel footer.
- **Admin "Create User" / "Create Group" renamed** — button labels now read "New User" and "New Group", consistent with "New Customer" and "New Project"; updated across all 12 locales.
- **Admin panel hardcoded column headers** — "Status", "Filename", and "Size" are now translated via `$t()` in all 12 locales.
- **Board report "Updated" column** — was using `formatDateTime`; corrected to `formatDate`.
- **Seed `--reset` missing TicketView cleanup** — `ticket_views` rows are now deleted for both customer and inbox tickets on reset.

### Changed
- **Grid PDF month: no phantom day columns** — months shorter than 31 days no longer render empty greyed columns; the freed space goes to the label column.

## v0.10.33 — 2026-05-29

### Fixed
- **Settings view blank screen** — bare `@` in the `req_special` translation string (`!@#$…`) was treated as a vue-i18n linked-message prefix, causing `SyntaxError: 10` when SettingsView loaded. Escaped to `{'@'}` in all 12 language files.

## v0.10.32 — 2026-05-29

### Added
- **Grid PDF period picker** — the Grid PDF button now opens a panel with a segmented Week / Month / Year control and context-sensitive date selectors (week number + year, month name + year, or year alone), so any period can be exported without navigating to it first.
- **`scripts/inject_time.py`** — CLI utility to bulk-create time entries over a date range with optional customer/project lookup, MFA support, dry-run mode, and automatic token refresh.

### Fixed
- **Grid PDF vertical grid lines** — data rows now draw left and right cell borders across all three grid types (week, month, year); the totals row also gains a horizontal border below it.
- **PDF export options panel direction** — the Export options dropdown in the Report tab now opens downward instead of upward, preventing it from being clipped at the top of the viewport.
- **Timesheet new-row column shift** — the inline new-row editor was missing its leading drag-handle cell, causing all columns to shift left compared to the header.
- **Token refresh in inject_time.py** — switched from a count-based refresh (every 100 entries) to a time-based check (every 13 minutes) so the access token never expires on slow APIs.

## v0.10.31 — 2026-05-29

### Added
- **Time tracking Grid PDF export** — new landscape A4 PDF grid for week (7 day columns with day abbreviation and date), month (31 day columns at compact 5.5 pt, days beyond month end greyed), and year (12 month + 4 quarter + yearly total columns). Accessible via **Grid PDF ▸** dropdown in the Log Time tab export bar. All three views honour the selected PDF font and language.
- **Holiday indicators in grid PDFs** — cells with a holiday entry are highlighted in amber; 0-minute holidays show `•` so the day is marked even when no hours were logged.
- **Weekend column highlights in month grid PDF** — Saturday and Sunday columns are tinted light gray in both the column header and all data rows.

### Fixed
- **Dutch "no customer" translation** — changed from "Geen klant" to "Zonder klant" in all three contexts (project dropdown, time tracking dropdown, inbox display).
- **PDF customer subtotal label** — the customer subtotal row in the report PDF now reads `<Customer> - Total` instead of `<Customer> Total`.
- **Grid PDF dropdown opens upward** — the Grid PDF and PDF Options dropdowns in the export bar now open upward instead of being hidden below the viewport.
- **Time range popup** — the start/end time picker now correctly flips downward for top rows by measuring position relative to the scroll container, not the viewport.
- **Year grid header overlap** — the column header row no longer overlaps the first data row (was using `Ln(0)` instead of `ln=1` on the last header cell).
- **Month grid header label** — month PDF header now shows "mei 2026" instead of "mei Jaar 2026".
- **Month grid day column offset** — day index computation normalised to UTC midnight to prevent timezone-related column misplacement.
- **Grid PDF row label** — falls back to the activity description when no customer or project is linked.

## v0.10.30 — 2026-05-29

### Added
- **`get_warmdesk` update script** — included in every server tarball; run as root to download and install the latest release automatically, with architecture auto-detection (`amd64` / `arm64`) and a `--force` flag.

### Fixed
- **Time range popup clipping** — the start/end time picker now flips to open downward when the cell is in one of the top rows, preventing it from being cut off above the viewport.

### Changed
- **Documentation sync** — admin guide now covers the helpdesk module, IMAP incoming mail, and the `get_warmdesk` update script; user guide adds a full Helpdesk section, fixes the Time Reports route, and documents all three Time Tracking tabs; `release.md` reflects the Claude Code skill workflow; `api.md` corrects the card prefix rules (1–10 chars/digits) and adds the three helpdesk Bruno folders.

## v0.10.29 — 2026-05-29

### Added
- **Today column highlight** — the current day's column in the time tracking grid is tinted with the primary colour so today is immediately obvious.
- **IMAP processed-folder docs** — the behaviour of moving processed emails to a `Processed` mailbox is now documented in CLAUDE.md and `warmdesk.yaml.example`, including the configurable `processed_mailbox` key.

### Fixed
- **Customer access — no rows means no access** — non-admin users with no `CustomerAccess` rows (direct or via group) now correctly see no customers; the stale model comment that documented the opposite behaviour has been corrected. `requireCustomerAccess` now also honours group-based access via `GroupCustomerAccess`.
- **Time tracking footer colspan** — the Total, Undeclarable, and Declarable footer rows used `colspan="3"` but the table has four fixed columns, shifting all day-column totals one position to the left.
- **Time tracking tab bar** — the Log Time / Report / Board tabs are now always right-aligned in the header bar; the ⚙ manage-projects button floats to the far right.
- **Today column CSS specificity** — the `th.c-day-today` background was overridden by the lower-specificity `.tt-head th` rule; fixed with `.tt-head th.c-day-today` and a box-shadow overlay for data cells.
- **Backup "Start time" alignment** — the Start time input in Scheduled Backups now bottom-aligns with the Interval dropdown.
- **Time tracking export controls alignment** — PDF font/language selects now bottom-align with the Export XLSX / Export PDF buttons.

## v0.10.28 — 2026-05-28

### Added
- **Ticket checklist templates** — admins define reusable ordered checklists in Admin → Checklists; agents apply a template to a ticket in one click; closing or moving to pending-close is blocked until every item is checked off; progress bar with n/m counter and drag-to-reorder shown in the ticket detail.
- **Bruno & Postman collections expanded** — added coverage for macros, checklist templates, inbox tickets, spam, move, history, and raw-email endpoints.

### Fixed
- **Ansible macro module** — YAML documentation scanner error on `ansible-galaxy` import caused by unquoted colon-space patterns in description bullets; now quoted correctly.
- **Build** — `make build` now removes `dist/docs/` before copying, preventing a nested `dist/docs/docs/` directory on incremental builds.

## v0.10.27 — 2026-05-28

### Added
- **Macro system** — create reusable macros that set status, priority, type, add tags, and pre-fill a reply message in one click; apply from a dropdown in the ticket detail view.
- **Macro placeholder expansion** — `add_message` actions support `{email}`, `{fname}`, `{name}`, `{subject}`, `{ticket_id}`, `{agent}`, `{agent_fname}` placeholders; clickable insertion chips in the macro editor.
- **Macro demo data** — seed program includes six ready-made helpdesk macros (Acknowledge & Investigate, Request More Information, Escalate to Critical, Resolved — Pending Close, Close & Thank, Mark as Duplicate).
- **Spam handling** — mark any ticket as spam to close it and hide it from the list; red SPAM badge on spam tickets; Show/Hide Spam toggle in ticket list and inbox headers; Not Spam button restores the ticket to open.
- **Inbox card/group/list views** — the Inbox now has the same three-view layout as customer ticket lists, with independent localStorage persistence.
- **Larger reply box** — the comment box is now 8 rows tall (min 180 px) with a Cancel button that clears the draft and any pending attachments.

### Fixed
- **Macro add_message pre-fill** — applying a macro pre-fills the reply box with the expanded text instead of posting it immediately, so agents can review and edit before sending.
- **Macro action row layout** — type select constrained to 160 px; value inputs use `min-width: 0` so the row no longer overflows the form card.

## v0.10.26 — 2026-05-28

### Added
- **Helpdesk inbox** — unassigned email-created tickets queue in a dedicated Inbox view with its own sidebar badge showing `unread/total` counts; inbox list refreshes live when new mail arrives.
- **IMAP polling** — automatically creates inbox tickets from incoming email; outbound agent replies are sent back to the customer by email.
- **IMAP OAuth2** — authenticate to Gmail and Office 365 via XOAUTH2/OAUTHBEARER; refresh tokens are stored and auto-renewed before each poll.
- **Reply threading** — incoming email replies are matched to existing tickets via `In-Reply-To`, `References` headers, subject `[#N]` tag, or `X-WarmDesk-Ticket-Id` header; closed/resolved tickets are automatically reopened on customer reply.
- **Email indicators** — tickets and messages show the sender's name and address; an ✉ badge marks messages that triggered an outbound email reply.
- **Real-time inbox refresh** — inbox counter and ticket list update instantly via WebSocket when IMAP delivers new mail; open ticket detail refreshes incoming replies in-place without clearing a reply draft.
- **Move ticket between customers** — reassign an unassigned inbox ticket to a customer, or move an existing ticket to a different customer.

### Fixed
- **IMAP test connection** — "Test connection" now uses the current form values instead of the last saved settings.
- **Inbox badge format** — badge consistently shows `unread/total`.

## v0.10.25 — 2026-05-27

### Added
- **Board report tab** — the project-board time report is now a third "Board" tab inside Time Tracking; the separate Reports menu entry is retired and `/reports` redirects automatically.
- **PDF export options** — toggle page numbers and undeclarable time on/off directly from the export options panel.

### Changed
- **Export options panel** — PDF font and language selects moved into the export options popover, keeping the filter bar on a single line at HD resolution.

## v0.10.24 — 2026-05-27

### Fixed
- **Time tracking row drag-and-drop** — reordering rows would stop working after navigating to a different week; Sortable is now re-attached whenever the table body is recreated.

## v0.10.23 — 2026-05-27

### Changed
- **Website "What's new" strip** — homepage now shows the current release highlights between the hero and features sections; updated each release via `hugo.toml`.

## v0.10.22 — 2026-05-27

### Added
- **Leave / remove conversation** — hover over any conversation in the chat list to reveal a ✕ button. For 1-on-1 chats this deletes the conversation for both parties; for group chats it removes only you from the group.

### Fixed
- **Week picker and holidays dropdown not closing on Escape** — pressing Escape now closes the week calendar picker and the holiday country dropdown in the time tracking view.

## v0.10.21 — 2026-05-27

### Fixed
- **Backspace in time tracking cells** — Backspace now deletes one character at a time when
  the cursor is positioned within a cell value; it only clears the whole cell when all text is
  selected (the initial state on focus), matching standard spreadsheet behaviour.
- **Sidebar resize handle blocked** — `overflow-y: auto` on `.app-sidebar` was clipping the
  3 px of the resize handle that extended outside the element and placing the scrollbar directly
  over it. The scrollable content is now wrapped in an inner `.sidebar-scroll` div; the handle
  sits outside the scroll container and is always accessible.

### Changed
- **Ticket group count badge** — styled to match the board column card-count pill
  (`--color-primary` background, `9999px` radius, `11px/600`).
- **UI consistency pass** — resolved a wide range of visual inconsistencies introduced by
  multiple contributors over time:
  - All `toLocaleDateString()` calls replaced with `useDateFormat()` (chat day separators,
    chart axes, calendar aria-labels) so they respect the user's date format preference.
  - Backlog card count badge now matches the board column pill style.
  - Call accept/decline buttons unified across 1:1 and group call overlays (`--color-success`
    / `--color-danger`).
  - `.toast-success` now uses `--color-success`; `.toast-mention` uses `--color-primary`.
  - Board type badge (`Kanban`/`Scrum`) colour scheme synced between Dashboard and Project
    Settings views.
  - SLA badge padding/radius unified across ticket list and ticket detail.
  - `btn-warning` moved to global `main.css`; per-view duplicates in BacklogView and
    SprintBoardView removed.
  - Redundant per-view `.btn` / `.btn-primary` / `.btn-danger` overrides removed from
    CustomerDetailView and CustomersView.
  - Page `<h1>` standardised to `22px` across all views; canonical `.page-header` rule added
    to `main.css`.
  - All 37 `window.confirm()` calls replaced with `await ui.confirm()` backed by a new themed
    `ConfirmDialog.vue` component (accessible, keyboard-trapped, Escape to cancel).

## v0.10.20 — 2026-05-28

### Added
- **Duplicate row in time tracking** — ⧉ button duplicates a row below the original with
  auto-incrementing `(copy)` / `(copy 2)` suffix.
- **Drag handles in time tracking** — grab the `⠿` handle to manually reorder rows; order
  persisted in `localStorage` across sessions.
- **Stable row order** — rows sorted by creation order (min entry ID) so they don't jump
  on week load; editing a duplicated row's description/project preserves its position
  instead of jumping to the bottom.
- **`common.duplicate` i18n key** added to all 12 locale files.

### Fixed
- **Time entry rows jumping** — `entryRows` computed now sorts by minimum entry ID
  ascending for stable creation order instead of following the API's `date desc` order.
- **Duplicated rows moving to bottom on edit** — `confirmEditRow` renames the key in
  `_keyOrder` at its current position instead of leaving the old key and letting the
  new key land at the end.
- **Row order reset on week change** — `_keyOrder` now saves to and restores from
  `localStorage`, preserving the user's arrangement across weeks.

## v0.10.19 — 2026-05-27

### Added
- **Ticket status redesign** — merged resolved/closed into `pending_close` (date picker + auto-close),
  re-added `closed` as an immediate final status.
- **View modes** — three ticket list layouts: Card (grid), Group (grouped by status with cards or
  sortable table per group), and List (full-width sortable table), persisted in `localStorage`.
- **Ticket history** — activity timeline tracking all status changes, comments, links, tag
  changes, and date updates, shown as a collapsible timeline in ticket detail.
- **Date title prefix** — setting a pending reminder or pending close date automatically
  prefixes the ticket title with `[YYYY-mm-dd]`; clearing the date removes it.
- **Sidebar ticket counts** — customer sidebar now shows badge counts for new/open, pending,
  pending_close, and closed tickets.
- **Sort on grouped tables** — group view's per-section tables have sortable columns.
- **Pending close auto-close** — tickets with `close_at` in the past are automatically closed
  on list/detail load.

### Fixed
- **Vue i18n SyntaxError in 11 locale files** — unescaped `{'@'}` in `req_special`
  translations caused a Vue compile error.

## v0.10.18 — 2026-05-26

### Fixed
- **Settings page crash in Linux RPM** — `loadPersonalKeys()` and
  `revokePersonalKey()` in the user settings page lacked error handling; an
  unhandled promise rejection crashed the WebKit GTK WebView on Linux desktop
  builds.
- **Clipboard API crash in settings and project settings** — `navigator.clipboard`
  calls in the copy-key buttons had no guard; WebKit GTK on Linux does not always
  expose the clipboard API, causing a `TypeError` that took down the page. Now
  falls back to a "copy manually" error message.
- **Locale not saved for 7 languages** — selecting Danish, Swedish, Norwegian,
  Finnish, Icelandic, Portuguese, or Italian in personal settings silently did
  nothing (backend validation was incomplete). All 12 supported locales are now
  accepted.
- **Black theme not persisted** — selecting the *Black* theme in personal settings
  was silently discarded by the backend; it is now accepted and saved alongside
  light / dark / system.

## v0.10.17 — 2026-05-26

### Added
- **Helpdesk — full ticketing module** — tickets with type (incident / problem /
  service request / change request), priority, status (open / in-progress /
  resolved / closed / pending), assignee, owner, internal messages with file
  attachments, tags, linked tickets, linked board cards, and a pending-reminder
  date stored per ticket.
- **SLA policies** — admin-configurable SLA policies with response and resolution
  time limits and optional priority filters. Policies auto-apply to new tickets
  based on priority; breach status is recomputed and stored on every fetch.
- **Pending reminder — custom date picker** — replaced the native
  `<input type="date">` with a fully themed `DatePicker.vue` component that
  respects the user's date format and week-start preferences and works correctly
  in all themes (light / dark / black).
- **Dashboard default setting** — users can set their home page to *Boards* or
  *Tickets* in personal settings. The ticket option redirects to the first
  starred (or first listed) customer's ticket list.
- **News on ticket pages** — the dashboard news widget now also appears at the
  top of the ticket list page so helpdesk-first users see announcements without
  visiting the boards dashboard. The dismissed-items list is shared between both
  views via `localStorage`.
- **Demo seed — customer member access** — added direct `CustomerAccess` rows for
  all three demo customers (Acme Corporation, Globex Systems, Initech Ltd) so the
  ticket assignee dropdown is populated immediately after seeding.
- **Playwright screenshots — tickets** — screenshots 21 (ticket list) and 22
  (ticket detail) added to the automated capture suite and README table.

### Fixed
- **Markdown list indentation** — `ul` / `ol` inside `.markdown-body` were not
  indented because styles were defined only in `AdminView.vue`'s scoped block.
  Moved to global `main.css` so all views (tickets, news, dashboard) benefit.
- **Ticket list page title** — showed "Tickets — Tickets"; now shows only the
  customer name (e.g. "Acme Corporation").
- **Customer name in ticket list** — `GET /customers/:id` returns
  `{ customer: {...}, contracts: [...] }`, but the frontend was reading `.name`
  from the wrapper object; fixed to unpack `data.customer`.

### Ansible (collection v0.4.0)
- **New module `ticket`** — create, update, and delete helpdesk tickets; customer
  is resolved by name; supports check mode.
- **New module `sla_policy`** — manage SLA policies via the admin API; name is
  used as the idempotency key; supports check mode.
- **New module `user_access`** — manage user feature flags (board, chat,
  helpdesk, time tracking) and global role via the admin API; supports check mode.
- Collection description updated to mention helpdesk / ticketing.

## v0.10.16 — 2025-05-25

### Fixed
- **CI test suite — e2e/ directory excluded from vitest** — Playwright
  screenshot specs were being picked up by vitest, causing a false test
  failure. Excluded `e2e/**` from the vitest runner.

## v0.10.15 — 2025-05-25

### Added
- **Admin panel — Time Tracking tab** — manage time-tracking-only projects and
  customers directly from the Admin panel with full CRUD (name, colour,
  undeclarable minutes).
- **Compact time-slot display on contract tiles** — slots are shown as a 🕒 N
  badge with hover popup instead of an inline list, keeping the contract tile
  clean.
- **Passkey login on Tauri desktop app** — WebAuthn authentication now works in
  the native desktop build.
- **Playwright screenshot automation** — 18 reference screenshots captured
  automatically for documentation and release notes.
- **Delete/Backspace key to clear time tracking cells** — press Delete or
  Backspace on a selected cell to clear its value without reaching for the
  mouse.

### Fixed
- **Missing i18n translations** — added missing translation keys across all 12
  languages.

### Documentation
- **Time Tracking section in user guide** — documents the weekly grid, managing
  time-tracking-only projects/customers via the ⚙ gear button, and the
  "Global (created by admin)" concept.
- **Contract time slots in user guide** — documents the compact badge display
  and hover popup.
- **Admin guide — Time Tracking subsection** — documents both the Admin panel
  tab and the ⚙ manage modal.

## v0.10.13 — 2026-05-22

### Added
- **Overnight and multi-day contract time slots** — slots can span midnight (e.g. 19:00–07:00) and extend across calendar days via `end_day_offset` (e.g. Friday 19:00 → Monday 07:00). Day types now include `weekends` and individual weekday anchors (`monday`–`sunday`).
- **Contract slot week preview** — the contract edit dialog shows a visual strip of the current week with each slot highlighted on the days and times it covers.
- **Resizable contract edit modal** — the contract form can be resized for easier editing of long slot lists.
- **Standby shift logging** — a ⏳ action on each time-tracking row opens a dialog to log a multi-day shift (e.g. weekend standby Fri 19:00 → Mon 07:00); the entry is split automatically across calendar days with correct start/end times per day.
- **Time-tracking cell selection** — arrow keys move the selection; Shift+arrow or Shift+click extends a rectangular range (spreadsheet-style).
- **Multi-cell paste** — Ctrl+V (or Shift+Ins) pastes the copied cell into all selected cells at once.
- **Time-tracking undo** — Ctrl+Z reverts the last saved cell change, paste, holiday toggle, time-range edit, or standby shift; an **Undo** button appears in the bottom bar when there is something to undo.
- **Keyboard shortcuts — Time Tracking section** — the shortcuts modal (`?`) documents selection, copy/paste, undo, and Escape to clear the clipboard.
- **XLSX slot-aware billing** — time-tracking XLSX export adds per-slot sub-rows with label, hours, and cost when entries carry start/end times and the contract defines slots.

### Changed
- **PDF/XLSX slot matching** — overnight and multi-day contract slots are matched on a timeline so entries crossing midnight or spanning several days bill against the correct tiers.
- **Time cell input** — cells accept only valid time characters (digits, colon, or decimal separator); invalid text is blocked at input and paste.
- **Acme demo seed** — sample contract includes weekday and weekend standby slots (×1.5 / ×2.0) for trying the new billing and logging flows.

### Fixed
- **Ctrl+Z on Firefox** — undo is intercepted via `beforeinput` and a window-level capture listener so Firefox no longer swallows the shortcut when a cell is focused.
- **Undo focus scope** — Ctrl+Z works anywhere on the time-tracking page, not only when a cell input still has focus.

## v0.10.12 — 2026-05-22

### Added
- **Contract time slots** — contracts on the Customer detail page now support named time-of-day rate tiers (e.g. "Evening", "Weekend") with configurable start/end time, day type (all / weekdays / Saturday / Sunday), hourly rate, and multiplication factor. The PDF and XLSX time-tracking exports apply matching slots automatically when a time entry carries a start/end time.
- **Contract hourly rate and currency** — contracts now store a base `price_per_hour` and `currency`; nine currency options are available (EUR, USD, GBP, CHF, SEK, NOK, DKK, PLN, CZK).
- **Time entry start/end time** — each cell in the time-tracking grid can optionally record a wall-clock start and end time (HH:MM). A ⏱ button on each filled cell opens a compact popup; setting both times auto-fills the duration. A small dot indicator shows cells that already have a time range stored.
- **PDF inline time slot breakdown** — when a time entry has a start/end time and the project's contract defines time slots, the PDF export renders per-entry sub-rows immediately below each entry, sorted chronologically, showing the exact HH:MM–HH:MM overlap with each slot (and any standard-rate gap), its label, hours, and cost.
- **Time-tracking cell copy/paste** — Ctrl+C on any filled cell copies its full contents (duration, start/end time, holiday flag) to an internal clipboard. Ctrl+V pastes into any other cell; the source cell is highlighted with a dashed outline. Escape clears the clipboard.
- **Deleted cards restore** — project owners and system admins can view, permanently delete, or restore soft-deleted cards from a new "Deleted cards" tab in Project Settings. Restore records an event in card history.
- **Working hours per user** — users can configure expected start times for each day of the week (Mon–Sun) in User Settings. The time-tracking view uses these to flag days where logged time exceeds the configured limit.
- **Conditional visibility in card detail** — each section of the card detail modal (labels, attachments, checklist, linked cards, watchers, etc.) can be individually shown or hidden from a sections menu (⋮). Visibility state persists in `localStorage`.
- **Time notation preference** — users can choose between decimal (8.5 h) and HH:MM (8:00) display format for durations; the setting is applied consistently across the time-tracking UI and settings forms.

### Changed
- **Sidebar drag-handle tooltips** — drag handles on starred projects and customers in the sidebar now show a tooltip, improving discoverability of the reorder feature.
- **Seeder idempotency** — the seed program (`cmd/seed`) can now be re-run safely without creating duplicate data.
- **`--version` flags** — `cmd/seed` and `cmd/training` now accept a `--version` flag and print their version string.

### Fixed
- **Ansible `user_options` module** — self-service calls now use `/auth/me` instead of the admin endpoint, fixing permission errors for non-admin automation users.

## v0.10.11 — 2026-05-22

### Added
- **Ansible collection — `user_options` module** — new `ansilabnl.warmdesk.user_options` module manages per-user preferences (locale, timezone, font, sidebar position, accent colour, time notation, week start, time tracking toggle, theme, breadcrumbs, email notifications) via the admin API. Idempotent — only sends a PUT when values differ.
- **Admin API — new user preference fields** — `PUT /api/v1/admin/users/:id` now accepts `time_tracking_enabled`, `theme`, `show_breadcrumbs`, and `email_notifications`, matching the existing user-facing `PUT /api/v1/auth/me` endpoint.

### Changed
- **Ansible collection — `card_comment` option renamed** — `time_spent_minutes` → `time_spent` for consistency with the `card` module. The API JSON key remains `time_spent_minutes`.

## v0.10.10 — 2026-05-22

### Added
- **Ansible collection — `time_spent` option on card module** — `ansilabnl.warmdesk.card` now accepts `time_spent` (int, minutes) to set the total time logged against a card. Idempotent: only issues a PUT when the value differs from the server.

## v0.10.9 — 2026-05-22

### Security
- **CVE patches** — `golang.org/x/net` updated to v0.55.0 and `golang.org/x/sys` to v0.45.0
- **CSP hardened** — `object-src 'none'` added to Content-Security-Policy, blocking Flash and other plugin embedding

### Accessibility (WCAG 2.1 AA)
- **Board cards keyboard-accessible** — card divs now carry `role="button"`, `tabindex="0"`, and Enter/Space handlers so cards can be opened without a mouse
- **i18n aria-labels** — hardcoded English `aria-label` strings in BoardColumn, CardDetail, and BoardCard replaced with translated `$t()` equivalents; three new keys added to all 12 locales (`board.sort_select_aria`, `board.sort_asc_action`, `board.sort_desc_action`)
- **A11yStatusModal focus management** — added focus trap (Tab/Shift+Tab cycling), Escape-to-close, and focus-restore on unmount
- **ProjectSettings colour input** — `aria-label` added to the label colour picker

## v0.10.8 — 2026-05-22

### Security
- **JWT purpose segregation** — refresh and MFA tokens now carry a `purpose` claim; dedicated `ValidateRefreshToken` / `ValidateMFAToken` helpers reject any token whose purpose doesn't match, preventing token-type confusion attacks
- **SVG upload blocked** — SVG files are no longer accepted on image upload endpoints, eliminating a stored-XSS vector via inline `<script>` in SVG
- **Accurate MIME sniffing** — replaced `net/http.DetectContentType` with `gabriel-vasile/mimetype` for attachment serving and conversation-avatar uploads, preventing MIME-type spoofing
- **Attachment ownership enforced before parsing** — the attachment handler now verifies project membership before opening the file, preventing path-based enumeration
- **SSRF / DNS-rebinding fix** — the media-proxy handler resolves the upstream hostname once and dials by IP, preventing DNS-rebinding attacks
- **Webhook tokens hashed** — tokens are now stored as SHA-256 hashes (`token_hash` column); existing plaintext tokens are migrated automatically on startup
- **Auth rate limiting extended** — `/auth/refresh` and `/auth/passkey/login/begin` are now covered by the `AuthRateLimit` middleware
- **Group-call invite restricted** — users can only be invited to a group call via a conversation they already share with the inviter
- **Reduced user data exposure** — non-admin callers of `/users` receive a trimmed response (id, username, display name, avatar, online status) rather than the full user struct
- **Secure cookie flag improved** — derived from TLS state or `X-Forwarded-Proto` rather than hard-coded; `X-WarmDesk-Client` header is stripped of control characters; non-upload avatar URLs are rejected on profile update

### Changed
- **Remote DB without TLS warns on startup** — a log warning is emitted when `db_driver` is `postgres` or `mysql` and `db_tls_mode` is `disable`
- **Backup error messages sanitised** — `pg_dump` / `mysqldump` stderr is no longer forwarded to the API response

### Accessibility (WCAG 2.1 AA)
- **Tab panels** — `role=tablist/tab/tabpanel`, `aria-selected`, `aria-controls`, `aria-labelledby` wired up in ProjectSettings, Backlog, Admin, and DirectMessages views
- **Icon-only buttons** — `aria-label` added to every icon-only button across all views; `aria-pressed` on toggle buttons (star/favourite, layout picker, notification bell, view-mode, accent swatches)
- **Form label linkage** — `for`/`id` pairs added to all form inputs in Settings, Admin (system settings + all modals), CustomerDetail, Customers, Topics, and DirectMessages
- **Hover-only elements keyboard-visible** — `:focus-visible` / `:focus-within` rules ensure emoji triggers, drag handles, and message action buttons are reachable by keyboard
- **Dialog accessibility** — `aria-modal="true"` and focus-restore on `AboutModal` and `NewsWelcomeModal`
- **About box** — website URL (`tonk.github.io/warmdesk`) added; `about.website` i18n key added to all 12 locales

## v0.10.7 — 2026-05-21

### Added
- **Maximize button on card detail** — a maximize/restore button (⤢) appears next to the close button on every card; one click fills the full viewport, a second click returns to the previous size; useful for presenting cards to trainees or reviewing long descriptions

### Fixed
- **Card modal clipped when zoomed in** — at high browser zoom levels the card detail modal was larger than the viewport with no way to scroll; the modal backdrop is now the scroll container so the full card is always reachable

## v0.10.6 — 2026-05-21

### Added
- **Card activity history panel** — a new *Card History* button in the card detail footer opens a full activity timeline showing who did what and when: card created, comments added, title/description/priority/assignee/start-date/due-date changes, column moves, and open/close events; all events are logged server-side with timestamps and user attribution
- **Ansible collection — `card_comment` module** — `ansilabnl.warmdesk.card_comment` creates, updates, or deletes comments on a card identified by project slug and card number; idempotent update via `comment_id`; supports `time_spent_minutes`

### Fixed
- **Media proxy panic on IPv6 upstreams** — `GET /api/v1/media/proxy` panicked with a nil pointer dereference when the upstream hostname (e.g. `api.dicebear.com`) resolved to an IPv6 address; IPv6 literals are now correctly bracketed in the dial-target URL and the request-construction error is handled instead of silently discarded

### Changed
- **Vite dev server now proxies `/uploads`** — project avatar images were missing in development mode because the Vite proxy only forwarded `/api`; `/uploads` is now forwarded to the backend so avatars render correctly during local development

## v0.10.5 — 2026-05-21

### Added
- **Ansible collection — `closed` parameter for `card` module** — `ansilabnl.warmdesk.card` now accepts `closed: true/false` to open or close a card; idempotent (only updates when state differs)

## v0.10.4 — 2026-05-21

### Added
- **Ansible collection — `card_ref` return value** — `ansilabnl.warmdesk.card` now returns `card.card_ref` (e.g. `GF00-4`) so the result can be passed directly as `card_number` in follow-up tasks

### Fixed
- **Ansible collection — assignee lookup with project-scoped API keys** — `GET /api/v1/users` is blocked (403) for project-scoped keys; the resolver now falls back to `GET /projects/{slug}/members` automatically; both username and email address are accepted as the assignee value

## v0.10.3 — 2026-05-21

### Added
- **Week start day preference** — users can now choose whether their time-tracking week starts on Monday (ISO default) or Sunday via Settings; the preference affects the sheet view, week-picker calendar, current-week detection, and XLSX/PDF exports

### Fixed
- **Week start preference not saving** — `week_start` was missing from the `PUT /auth/me` handler, so the setting silently reverted to Monday on every page load

## v0.10.2 — 2026-05-21

### Fixed
- **CI build pipeline broken after dependency upgrade** — `sccache`, `clang`, and `mold` were referenced by `.cargo/config.toml` but not installed on GitHub Actions runners; all three client build workflows and the release workflow now install `clang`/`mold` via apt and suppress the `sccache` wrapper with `RUSTC_WRAPPER=""`

### Changed
- **All Go dependencies updated to latest**

## v0.10.1 — 2026-05-20

### Fixed
- **Time entry cells blank on Tauri desktop** — placeholder text (`0:00`) was rendering at full opacity in GTK WebKit, indistinguishable from real values; placeholder is now transparent by default and only appears faintly on focus

### Changed
- **Go upgraded to 1.26** — `go.mod` and CI workflows updated
- **All frontend dependencies updated to latest** — Vue 3.5.34, Pinia 3, Vue Router 5, vue-i18n 11, marked 18, Vite 8.0.13, Tauri 2.11.2, axios 1.16.1, dompurify 3.4.5, livekit-client 2.19.0, and more
- **Faster AppImage builds** — `sccache` caches compiled Rust objects between builds; `mold` replaces the GNU linker for the link step; first build unchanged, subsequent builds significantly faster

## v0.9.60 — 2026-05-20

### Added
- **Calendar date-picker on week label** — clicking *Week 14 2026* in the time-tracking navigation bar opens a month calendar popover; click any day to jump to its week, or click a week number in the left column to jump directly to that ISO week; month navigation arrows browse forward and backward; the current week is highlighted in primary colour, today is bold; closes on outside click

### Fixed
- **Empty time cells show blank instead of 0:00** — time entry cells with no logged time (or a 0-minute holiday stub) are now left blank for a quieter visual; the `0:00` placeholder still appears on focus
- **Holiday dropdown flag badges** — emoji flag characters rendered as letter boxes on Linux without a colour emoji font; replaced with CSS gradient badges approximating each country's flag colours (gradient stripes for tricolors, simplified patterns for Nordic-cross and Union Jack)

## v0.9.59 — 2026-05-20

### Added
- **Public holidays in time tracking** — an *Add holidays* dropdown in the week navigation bar lists 12 countries (UK, Netherlands, Germany, France, Spain, Denmark, Sweden, Norway, Finland, Iceland, Portugal, Italy) by flag and native name; clicking a country creates 0-minute holiday entries for the selected year with all national public holidays; entries already on a date are skipped and the result is reported as added / skipped
- **Holiday highlighting in the sheet** — day column headers with a holiday show an amber background and a glowing dot; the holiday row has a distinct amber background; the specific day cell within the row is highlighted more strongly

### Fixed
- **Backup download in Tauri desktop app** — the *Download* button in the Admin → Backup tab now opens the native save dialog correctly; previously it did nothing because it used the WebKit blob API which is broken on Linux GTK WebKit

## v0.9.58 — 2026-05-20

### Fixed
- **Time-tracking week header date mismatch** — in UTC+ timezones the column dates were one day behind the displayed day abbreviation because `Date.toISOString()` returns UTC time; dates are now derived from local date components so the header, date keys, and stored entries all agree

## v0.9.57 — 2026-05-20

### Added
- **Day-of-week abbreviation in PDF export** — a new *Show day of week before date* toggle in the PDF export options panel prepends the localised day abbreviation (Mon–Sun in the report language) to each date column entry; the abbreviation is rendered in a fixed-width cell so dates stay aligned regardless of 2- or 3-character abbreviations across all 12 supported languages
- **Personal TT-only projects and customers in seed data** — the demo seed now creates TT-only customers (Smart Owl Consulting, Personal) and projects (Travel with 45 min undeclarable, Holidays with 480 min, Study & Training, Internal with 60 min) for the `tonk` user, with 14 sample time entries; `--reset` cleans them up correctly

### Changed
- **PDF export controls replaced with dropdown** — the bare *New page per customer* checkbox is now grouped inside an *Export options* dropdown button alongside the new day-of-week toggle; all settings are persisted to `localStorage` between sessions

## v0.9.56 — 2026-05-20

### Added
- **Undeclarable time per time-tracking project** — each time-tracking-only project can now have an *undeclarable minutes* value set in the ⚙ manage modal (displayed as HH:MM); this represents time that cannot be billed (e.g. travel, holidays); in the weekly sheet the row total shows an ↓ badge with the undeclarable portion and the footer shows an undeclarable row (red) and declarable row (green) when any undeclarable time exists
- **Undeclarable time in reports** — the time-tracking report tab subtracts undeclarable time from each entry and group total; when grouping by Customer an *Undeclarable* line is shown per customer and a total undeclarable row appears at the bottom; declarable time is shown as the primary value throughout
- **Undeclarable time in PDF export** — the PDF time-tracking report applies the same undeclarable/declarable logic as the on-screen report; when grouping by Customer an undeclarable line is added below each customer subtotal; the grand total shows declarable time
- **PDF page-break per customer** — a *New page per customer* checkbox appears in the PDF export controls when the report is grouped by Customer; when checked each customer gets its own page with the full document header (logo, company name, report title, period, employee name) repeated at the top; the grand total is omitted in this mode as each customer's page already contains its own subtotal
- **Undeclarable time in XLSX export** — the XLSX time-tracking report export shows undeclarable and declarable totals below the grand total row when undeclarable time is present

## v0.9.55 — 2026-05-19

### Added
- **Quick-nav strip in the header** — Dashboard, News, Chats, Reports, and Time Tracking links appear directly in the header bar between the search field and the right-side controls; the active page is highlighted; Reports and Time Tracking are only shown to users who have access; the strip is hidden automatically on viewports narrower than 960 px (the user-menu dropdown remains the fallback on small screens); the unread-messages dot appears on the Chats link when there are unread direct messages

## v0.9.54 — 2026-05-19

### Fixed
- **Time-tracking sheet — rows no longer re-sort while entering time** — entering or saving a time value no longer causes all rows to jump to new positions; the sort order is now stable and only changes when the user explicitly clicks a column header to sort
- **Time-tracking projects list — sorted alphabetically** — the personal (time-tracking-only) projects list in the ⚙ manage modal is now sorted alphabetically by name; the order updates automatically when a project is added, renamed, or removed

## v0.9.53 — 2026-05-19

### Security
- **Fix attachment IDOR** — `GET /api/v1/attachments/:id` now verifies that the requesting user is a member of the parent project (for card, comment, and chat attachments) or a participant in the conversation (for direct-message attachments) before serving the file; previously any authenticated user could download any attachment by ID
- **Fix `Content-Disposition` header injection** — user-supplied attachment filenames are now escaped before being inserted into the `Content-Disposition` header, preventing header injection via filenames containing quotes or backslashes
- **Add HSTS and CSP response headers** — `Strict-Transport-Security` (1 year, includeSubDomains) and `Content-Security-Policy` (same-origin scripts/styles/fonts, `frame-ancestors 'none'`, `form-action 'self'`) are now sent on every response by `middleware/security_headers.go`
- **Block CORS wildcard at middleware level** — a wildcard `allowed_origins` value is now rejected at request time in addition to the existing startup check; no `Access-Control-Allow-Origin` header is set when a wildcard config is detected
- **Enforce minimum JWT secret length** — the server refuses to start if `jwt_secret` is shorter than 32 characters
- **Rate-limit message-send endpoints** — `POST /direct-messages/:userId` and `POST /conversations/:id/messages` are now covered by a 60 req/min rate limiter via `middleware.MessageRateLimit()`
- **Mask password-reset token in logs** — the audit log entry for a sent password-reset email now includes only the first 8 characters of the token followed by `...` instead of the full URL

### Changed
- **WCAG 2.1 AA compliance** — all icon-only buttons, unlabelled inputs, and selects across the board, card detail, chat, call, time-tracking, dashboard, admin, and layout components now have `aria-label`; decorative SVGs have `aria-hidden="true"`; custom tab patterns in `TimeTrackingView` and `ChatPanel` use full `tablist`/`tab`/`tabpanel` ARIA roles; hover-only interactive elements (sidebar star button, chat message actions) are now keyboard-accessible via `:focus-visible` / `:focus-within` CSS rules; hard-coded colour values replaced with CSS custom properties

## v0.9.52 — 2026-05-18

### Fixed
- **Desktop app — PDF and XLSX export** — all export buttons now work in the Linux AppImage; previous attempts using `response.arrayBuffer()`, `response.text()`, and `ReadableStream` readers all failed because WebKit GTK2 throws `TypeError("Type error")` on any body-read method of a `ReadableStream`-backed `Response` (the form tauri-plugin-http uses); fixed by routing binary downloads through a new `fetch_binary_b64` Rust command that fetches via reqwest entirely outside WebKit and returns base64, decoded in JS with `atob()`

### Changed
- **Desktop app — save dialog opens in home directory** — the native file-save dialog for PDF/XLSX exports now defaults to the user's home directory instead of the AppImage mount path
- **Desktop app — save dialog remembers last export directory** — after each successful export the chosen directory is stored in `localStorage`; the next export opens the dialog in the same location

## v0.9.51 — 2026-05-18

### Changed
- **Export error toast now shows the actual error message** — instead of a generic "Export failed." the toast displays the real JavaScript exception, making it possible to diagnose remaining export issues in the desktop app without DevTools access

## v0.9.50 — 2026-05-18

### Fixed
- **Desktop app — PDF and XLSX export error** — Axios's `responseType: 'blob'` calls `response.blob()` on the fetch Response, which is unreliable in Tauri's HTTP plugin for binary responses; switched all binary download endpoints to `responseType: 'arraybuffer'` (`response.arrayBuffer()`) which the plugin handles correctly; also removed a redundant Blob→arrayBuffer conversion in the Tauri write path

## v0.9.49 — 2026-05-18

### Fixed
- **Desktop app — time-tracking PDF and XLSX export** — all four export buttons (weekly sheet PDF/XLSX and report tab PDF/XLSX) now open a native OS save dialog in the Tauri desktop app; previously the buttons did nothing (XLSX) or threw an error (PDF) because the browser-anchor download approach is silently ignored by the WebView

## v0.9.48 — 2026-05-18

### Added
- **Time notation setting** — users can choose between decimal notation (e.g. 8.25) and HH:MM notation (e.g. 8:15) for the weekly timesheet; the preference is saved to the user profile and applied to all display and input; all calculations remain in minutes internally

### Fixed
- **PDF time report — "Date" column not translated** — the date column header in the time-tracking PDF export now uses the correct translated label in all 12 supported languages
- **PDF time report — month names not translated** — month names and abbreviated month names in period labels (e.g. "Apr 28 – May 4, 2026") are now fully localised in all 12 supported languages
- **PDF time report — DMY date order for European languages** — period labels and group headers now use day-month order ("17 mei") instead of the English month-day order ("mei 17") for all non-English locales

## v0.9.47 — 2026-05-17

### Added
- **"Remember me" on login** — checking the box saves the email/username to the browser's local storage and pre-fills it on the next visit; unchecking removes the saved value
- **Passkey sign-in** — browser clients can register passkeys (Touch ID, Windows Hello, hardware security keys) in User Settings → Passkeys and sign in from the login page without a password; uses discoverable credentials so no username entry is needed before the prompt; the Tauri desktop app is excluded (platform authenticator support is too inconsistent across Linux/macOS/Windows WebViews)

## v0.9.46 — 2026-05-13

### Added
- **Sortable columns in Log Time** — clicking the Customer/Project or Activity column header sorts all rows by that column; clicking again reverses direction; the day-hour cells always follow their row correctly

### Fixed
- **"Customer & Project" grouping showed only the first day** — the report grouped entries by relation name rather than by ID; GORM only populates the preloaded relation on the first occurrence of a given ID, so later entries with the same customer or project had a nil relation and were bucketed separately; grouping now uses numeric IDs as map keys and reads the label once from the first populated entry
- **Copied rows persisted across all weeks** — rows copied from the previous week survived navigation to other weeks because `localRows` was only cleaned up via a filter that kept anything not yet in the fetched entries; `loadWeek` now always resets `localRows` at the start, so copied rows belong exclusively to the week where the copy was triggered

## v0.9.45 — 2026-05-13

### Added
- **Time-tracking report grouping** — a new "Group by" selector in the time-tracking report tab lets users group entries by Period (existing behaviour), Customer, Project, or Customer & Project; the selected grouping is preserved in PDF and XLSX exports; customer+project pairs are shown as composite group headers (e.g. "Acme Corp › Travel time")

## v0.9.44 — 2026-05-13

### Changed
- **TT projects always available when a customer is selected** — in the time-tracking row form, selecting any customer now keeps all time-tracking-only projects (e.g. "Travel time") visible in the project dropdown, plus the customer's own regular projects; previously only exact customer-linked projects were shown

## v0.9.43 — 2026-05-13

### Added
- **Time-tracking-only projects** — users and admins can create lightweight projects used exclusively for time tracking; no Kanban/Scrum board, columns, or members are created; managed via the new ⚙ button in the time-tracking top bar; personal projects are visible only to their creator, admin-created projects are visible to everyone
- **Time-tracking-only customers** — same concept for customers; personal time-tracking customers are not added to the CRM customer list; managed in the same modal under the Customers tab
- **Manage modal with tabs** — the ⚙ gear button in the time-tracking view opens a Projects / Customers tabbed modal for full CRUD (add, rename, recolour, delete) of time-tracking-only entities
- **WCAG 2.1 AA enforcement** — accessibility requirement added to `CLAUDE.md`; the new manage modal is fully compliant: `aria-modal` dialog, labelled inputs, proper tab ARIA roles, `aria-label` on icon buttons, focus management on open

### Fixed
- **Time-tracking customer not selectable in time-tracking view** — new-row customer dropdown was still bound to the regular customers list; all customer dropdowns now use the merged regular + TT customers list
- **Time-tracking project not selectable after choosing a TT customer** — selecting a time-tracking customer now shows TT projects that are unassigned or explicitly linked to that customer
- **Customer / project column shows "—" after saving a TT entry** — row name resolution in `confirmNewRow` and `confirmEditRow` searched only the regular customers list; both now search the full merged list

## v0.9.42 — 2026-05-13

### Added
- **News read/unread toggle** — every item in the News overview now has a toggle button; marking an item "read" hides it from the dashboard, marking it "unread" restores it; replaces the old conditional "Mark as unread" button that only appeared on already-dismissed items
- **Welcome message on login** — news items can be flagged "Show as welcome message on login" in the admin panel; flagged items appear in a modal overlay immediately after login, paginated if there are multiple; each item is shown only once per user per browser (seen state stored in localStorage keyed by user ID)
- **Sidebar color for news tiles** — admin news form has a color picker (preset swatches + custom color input) that controls the left-border color of dashboard and News-view tiles
- **News overview page** — `/news` route accessible from the user menu shows all news items with an Active / All filter, sort by creation date / start date / end date / title (ascending or descending), and status badges (Inactive, Expired, Scheduled, Read)
- **Demo news items in seed** — `go run ./cmd/seed` now creates six realistic news items covering welcome, feature announcement, maintenance window, retrospective, security reminder, and team event; the welcome item is marked `show_on_login`

### Fixed
- **Dashboard news not appearing after DB reset** — dismissed item IDs (stored in localStorage) are now pruned against the live API response on each page load; stale IDs for items that no longer exist are removed so new items with recycled IDs are not incorrectly hidden

## v0.9.41 — 2026-05-12

### Fixed
- **User menu closes on selection** — selecting any item from the top-right user menu now closes the dropdown immediately instead of requiring an extra click elsewhere
- **PDF and XLSX export in desktop app** — report exports now open a native OS save dialog in the Tauri desktop app (using `tauri-plugin-dialog` + `tauri-plugin-fs`); previously the download was silently dropped because the WebView has no built-in download manager
- **Escape closes all popups and overlays** — all modals, dropdowns, and overlays now respond to Escape: About dialog, Call Settings panel, Emoji reaction picker, and incoming call / group call notifications; the call chat sidebar shows an inline discard-draft confirmation before closing when a message is in progress

## v0.9.40 — 2026-05-12

### Added
- **`Alt+A` — accessibility status modal** — press `Alt+A` anywhere in the app to open the WCAG 2.1 AA compliance status overlay; also listed in the keyboard shortcuts reference

### Fixed
- **WCAG AA color contrast** — light theme `--color-text-muted` raised from `#64748b` (4.34:1) to `#475569` (6.91:1); all text in all three themes now meets the 4.5:1 AA minimum

## v0.9.39 — 2026-05-12

### Added
- **Markdown in news** — admin news editor has a Write/Preview tab pair; dashboard news tiles render body text as Markdown (using `marked` + `DOMPurify`)
- **Ansible `news` module** — `ansilabnl.warmdesk.news` manages news items (create, update, delete) with full idempotency on `title` and ISO-8601 date normalisation
- **`from_vars.py` news support** — `news_items:` list in `warmdesk_vars.yml` is processed as Phase 6 alongside existing resource types

### Fixed
- **News "not found" on edit** — `NewsItem` model now serialises `id` (lowercase) instead of `ID`; corrected a GORM model embedding issue that caused `PUT /admin/news/undefined`

## v0.9.38 — 2026-05-12

### Added
- **Admin News management** — admins can create, edit, and delete news items (title, body text, show-from / show-until dates, active toggle) via a new "News" tab in the admin panel
- **Dynamic dashboard news tiles** — active news items (within their date window) are shown as dismissible "What's new" tiles on the dashboard, fetched from the API; dismissed items stored per-ID in localStorage
- **WarmDesk logo on news tiles** — the app logo appears in the header row of each dashboard news widget
- **`DateTimeInput` component** — reusable text input that formats and parses dates using the user's configured `date_time_format`, so date fields follow admin settings instead of browser locale
- **Website: WCAG highlight dismissible** — the accessibility highlight on the website homepage can be closed with an × button or toggled with `Alt+A`; dismissed state persists in localStorage

### Fixed
- **News date display** — admin news table and dashboard widgets now use `useDateFormat` (respects `date_time_format` setting) instead of `toLocaleDateString`
- **News form date pickers** — replaced `<input type="datetime-local">` with `DateTimeInput` so the show-from/show-until fields honour the configured date format

## v0.9.37 — 2026-05-12

### Added
- **WCAG 2.1 AA accessibility pass** — skip-to-content link, focus trap in all modals with focus return, `role="alertdialog"` on call overlays with auto-focus on accept button, `role="log"` on chat message lists, `role="alert"` on form errors, full combobox ARIA on global search, `aria-expanded`/`aria-controls` on all sidebar sections and menus
- **Keyboard shortcuts modal** — press `?` or use the user menu to open a two-column reference of all shortcuts grouped by context (Global, Navigation & Search, Dialogs, Board & Cards, Chat & Messages)
- **`Ctrl/⌘ + K`** — focuses the global search bar from anywhere in the app
- **Dashboard widget tiles** — a dismissible "What's New" blog tile and a persistent WCAG 2.1 AA compliance status tile on the home page
- **Heading hierarchy** — every view now has exactly one `<h1>`; Board, Backlog, SprintBoard, Charts, DM, Topics, Report, Gantt, and Time Tracking all corrected
- **Alt text** — all decorative images (`aria-hidden="true" alt=""`) and meaningful images given descriptive labels across call overlays, mention dropdown, link preview card, and video cells
- **i18n** — `shortcuts.*` (25 keys) and `dashboard.*` (3 keys) namespaces added to all 12 language files

### Fixed
- **Date picker inputs** in card detail had no accessible label; now carry `aria-label` for Start Date and Due Date
- **External issue ref** input in card detail labelled with `board.external_issue` i18n key

## v0.9.36 — 2026-05-10

### Added
- **Screen sharing in video calls** — share your display in one-to-one (WebRTC) and group (LiveKit) calls; remote sees the shared content while your camera preview stays in the picture-in-picture when applicable
- **Optional text chat during calls** — toggle a right-hand conversation panel (same DM or group as the call) to type without leaving the call UI; preference can persist via local storage
- **Website: seed program blog post** — documents `warmdesk-seed` / `go run ./cmd/seed`, idempotency, `--reset`, and example terminal output (`website/content/blog/warmdesk-seed-program.adoc`)

### Changed
- **Frontend bundle** — `livekit-client` is loaded on demand when joining a group call so the main call UI chunk stays smaller and build chunk-size warnings are avoided

### Fixed
- **Call control bar** — the bottom control strip wraps on narrow layouts (e.g. with the chat sidebar open) so the camera toggle and other buttons are no longer clipped off-screen

## v0.9.35 — 2026-05-08

### Fixed
- **Scrum chart rendering** — Burndown, Burnup, Cumulative Flow, and Release Burndown charts now render correctly; a Vue timing bug left the canvas element out of the DOM when Chart.js tried to draw

### Changed
- **Demo seed descriptions** — all ~101 seeded demo cards now include rich Markdown descriptions with context, acceptance criteria, technical notes, and reproduction steps for bugs

### Documentation
- **New screenshots** — Gantt chart, Scrum backlog, CFD, sprint throughput, burndown, burnup, and release burndown screenshots added to the website homepage, blog post, user guide, and README

## v0.9.34 — 2026-05-06

### Added
- **Kanban charts** — the 📊 Charts button is now available on all board types; Kanban projects get four dedicated charts: CFD (Cumulative Flow Diagram), Cycle Time, Lead Time, and Throughput (weekly closed-card count)

## v0.9.33 — 2026-05-06

### Changed
- **Theme switcher in app** — replaced the single-click cycle button in the top navbar with a Light / Dark / System dropdown; the active selection is highlighted and choice is persisted to the user profile
- **Theme switcher on website** — added a Light / Dark / System dropdown to the marketing website navbar; selection persisted in `localStorage`; dark mode CSS variables applied throughout; system mode follows the OS preference automatically

## v0.9.32 — 2026-05-06

### Fixed
- **Group call with no one online** — clicking the group video call button now shows "No other members are currently online" instead of silently joining an empty room when you are the only member present

## v0.9.31 — 2026-05-06

### Fixed
- **Customer logo proxy 400 on same-network hosts** — media proxy now returns 302 redirect (instead of 400) when the target hostname resolves to a private/internal IP; the browser loads the image directly while the server-side SSRF guard still prevents the server from fetching internal hosts

## v0.9.30 — 2026-05-06

### Fixed
- **External avatars in web app** — media proxy now returns a 302 redirect to the original URL instead of 502 when the server cannot reach the upstream image host (e.g. outbound firewall); browsers follow the redirect and load the image directly
- **CSP font-src blocking inlined fonts** — nginx and Apache deploy templates now include `data:` in `font-src` so Vite-bundled woff2/woff data-URI fonts load correctly
- **CSP img-src blocking redirected avatars** — nginx and Apache deploy templates now include `https:` in `img-src` so external avatar images (Gravatar, DiceBear, custom URLs) load correctly when the proxy redirects to them

## v0.9.29 — 2026-05-06

### Fixed
- **GPG signing: passphrase in batch mode** — signing no longer fails with "Sorry, we are in batchmode - can't get input" when the key has a passphrase; the passphrase is now read from the `GPG_PASSPHRASE` repository secret and passed directly to every signing call

## v0.9.28 — 2026-05-06

### Fixed
- **GPG signing in CI** — release artifact signing no longer fails with "Inappropriate ioctl for device"; the signing steps now configure `pinentry-mode loopback` and restart gpg-agent before signing so the headless runner does not need a TTY

## v0.9.27 — 2026-05-06

### Added
- **GPG-signed release artifacts** — every file attached to a GitHub release (server tarballs, AppImage, deb, rpm, dmg, Windows installer and portable zip) now has a companion detached armoured signature (`.asc`); verify with `gpg --verify file.asc file` after importing `signing-key.asc` from the repository

## v0.9.26 — 2026-05-06

### Added
- **SMTP From: display-name support** — the "From Address" setting now accepts both a plain address (`noreply@example.com`) and RFC 5322 display-name form (`WarmDesk <noreply@example.com>`); the admin input field placeholder is updated to reflect this

### Fixed
- **Backup email: monospace font throughout** — the Size and Date columns in the "Available backups" table now use the same fixed-width font as the Filename column; the header row is also monospace

## v0.9.25 — 2026-05-06

### Fixed
- **Backup email: filename wrapping in HTML viewers** — filenames in the backup list and the "File" detail row are now wrapped in `<nobr>` so HTML-to-text renderers (w3m, lynx, Mutt's HTML viewer) no longer break the filename mid-hash at narrow terminal widths

## v0.9.24 — 2026-05-06

### Added
- **Group video call in all group chats** — the video call button is now available in any group conversation regardless of member count; the previous restriction requiring 3 or more members has been removed

### Fixed
- **Linux desktop: camera and microphone in AppImage** — the Linux AppImage now bundles the complete set of GStreamer 1.24 plugins and sets `GST_PLUGIN_SYSTEM_PATH_1_0` and `GST_REGISTRY` so camera and microphone work correctly on all distributions including Fedora and openSUSE, without requiring a matching system GStreamer installation
- **1:1 call button icon** — the call button in one-on-one conversations now shows a camera icon, consistent with the group chat button
- **Backup email: filename wrapping and raw timestamp** — long backup filenames no longer wrap mid-word in narrow email clients; the backup date now shows as a readable date/time instead of a raw Unix timestamp

## v0.9.23 — 2026-05-04

### Added
- **"All Customers" sidebar section** — a collapsible section lists all customers below the Favourite Customers panel, with starred customers sorted first and marked; mirrors the existing "All Projects" section
- **Starred customers drag-to-reorder** — favourite customers in the sidebar can be reordered by dragging (pointer events — works on Linux/WebKitGTK)
- **LiveKit room prefix** — new `livekit_room_prefix` config option (env: `LIVEKIT_ROOM_PREFIX`) prepends a prefix to every room name; useful when sharing a LiveKit server across multiple WarmDesk instances
- **Media proxy rate limiting** — the external image proxy endpoint is now rate-limited to 200 requests per 10 minutes per IP to prevent unauthenticated bandwidth abuse

### Fixed
- **Media proxy SSRF protection** — the image proxy now resolves DNS once, verifies every resolved IP is a publicly routable address, and dials by IP directly; a subsequent DNS change cannot redirect the connection to an internal host (DNS-rebinding defence)
- **Media proxy query-string re-encoding** — URLs with spaces, commas, or other reserved characters in query parameters (e.g. DiceBear seed names) are now re-encoded before fetching, preventing malformed upstream requests
- **Media proxy content-type validation** — the proxy now rejects upstream responses whose `Content-Type` is not an image, preventing use as a general-purpose data-exfiltration relay
- **Tauri desktop: external images bypass same-origin proxy** — in the desktop app, external image URLs (Gravatar, DiceBear, etc.) are loaded directly instead of via the same-origin proxy; WebKit's `tauri://` origin treats the proxied HTTP responses as mixed content and blocks them

## v0.9.22 — 2026-05-04

### Added
- **In-call attendee invites** — a **+** button in the bottom controls bar of any active call opens a searchable user picker; selected users receive a real-time invite popup and join the LiveKit group room with one click; works in both group calls and 1:1 calls
- **1:1 → group call upgrade** — inviting someone from a 1:1 WebRTC call automatically upgrades the session to a LiveKit group room; the existing call partner also receives the join popup so no one is dropped

### Fixed
- **Apache WebSocket proxy timeout** — `deploy/apache.conf` now uses a dedicated `<Location "/api/v1/ws">` block with `ProxyTimeout 86400` (24 hours); the previous catch-all approach left WebSocket connections subject to Apache's 300-second default, silently disconnecting board updates, chat, and video-call sessions
- **Backup email: unit and layout** — file sizes now display as `kB` (correct SI prefix) instead of `KB`; size and date columns have `white-space:nowrap` so narrow email clients no longer wrap `548.0 kB` across two lines

## v0.9.20 — 2026-05-01

### Added
- **Video toggle button in the audio call bar** — a camera on/off button now appears in the slim bottom bar (shown during outgoing calls and audio-only calls) whenever the call has a video track, matching the style of the mute button
- **LiveKit token endpoint** — `GET /api/v1/conversations/:id/livekit-token` generates a signed access token for a LiveKit room; foundation for future group video/audio calls; returns 503 when LiveKit is not configured so existing behaviour is unchanged
- **LiveKit config fields** — `livekit_url`, `livekit_api_key`, and `livekit_api_secret` added to config with env var overrides (`LIVEKIT_URL`, `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET`) and full documentation in `warmdesk.yaml.example`

### Fixed
- **Call error toasts finally working** — despite the v0.9.19 watcher fix, the "declined", "unavailable", and "microphone denied" toasts were still never shown because `useI18n` and `useUIStore` were missing from `ActiveCallBar.vue`; every toast attempt threw a silent runtime error; both are now properly imported and initialised

## v0.9.19 — 2026-04-30

### Fixed
- **Call error toasts now always appear** — the "declined", "unavailable", and "microphone denied" toasts were silently dropped when a race between the 45-second ring timeout and an incoming signal meant the call phase was already `ended` before the error was set; the trigger now watches the error state directly so the toast fires every time

## v0.9.18 — 2026-04-30

### Added
- **"Declined" toast on outgoing calls** — when the person you are calling declines, an error toast now appears with their name ("Alice declined the call") instead of the call bar silently disappearing; translated in all 12 supported languages

## v0.9.17 — 2026-04-30

### Fixed
- **Blank screen under CSP** — added `'unsafe-eval'` to `script-src` in the nginx and Apache templates; vue-i18n's runtime message compiler requires `new Function()` and was blocked by the previous policy, causing a hard crash on every page load
- **Update checker blocked by CSP** — added `https://api.github.com` to `connect-src` so the in-app version check can reach the GitHub Releases API

## v0.9.16 — 2026-04-30

### Fixed
- **Blank screen under strict CSP** — vue-i18n's runtime JIT compiler used `new Function()`, triggering a CSP `unsafe-eval` violation that broke all translated views; messages are now pre-compiled at build time via `@intlify/unplugin-vue-i18n` and the runtime-only vue-i18n bundle is used, so no eval is needed at runtime

## v0.9.15 — 2026-04-30

### Fixed
- **WebSocket reconnect robustness** — in Tauri mode, a ticket-fetch failure now schedules a retry so the user WebSocket reconnects after transient errors; in browser mode, a silent token refresh is attempted before each reconnect to prevent a 15-minute JWT expiry from stranding the user
- **Call failure toast** — when an outgoing call fails because the callee is offline ("unavailable") or microphone access was denied ("no_mic"), an error toast is now shown instead of the call bar silently disappearing
- **Chat layout** — fixed layout issues in the chat panel and Direct Messages view so message lists fill available height correctly and the input area stays anchored to the bottom
- **nginx WebSocket header** — `Connection` header value is now correctly quoted (`"upgrade"`) in the provided nginx template, matching the HTTP spec requirement
- **Online presence poll interval** — reduced the sidebar presence poll from 10 s to 5 s so online/offline status updates appear more quickly; interval extracted as a named constant

## v0.9.14 — 2026-04-30

### Added
- **Online presence in Direct Messages** — 1:1 conversations now show a green presence dot overlaid on the avatar in the conversation list, and an "Online / Offline" status line below the name in the chat header; updates every 10 seconds alongside the sidebar

### Fixed
- **Call hang-up when other side doesn't answer** — two bugs fixed: (1) a race condition where cancelling a call during media acquisition caused `startCall` to continue after the hang-up and re-send the offer; (2) no auto-cancel timeout — unanswered outgoing calls now automatically hang up after 45 seconds
- **Same race condition in `acceptCall`** — if the caller hung up while the callee's media permission dialog was open, `acceptCall` would continue and send an answer to a dead call; a phase guard now aborts cleanly

## v0.9.13 — 2026-04-30

### Added
- **1:1 audio and video calls** — call any user directly from the Direct Messages header; calls use WebRTC peer-to-peer (STUN servers) with signaling over the existing WebSocket connection; audio-only fallback when camera is unavailable; call states: calling → ringing → active → ended
- **Incoming call overlay** — fixed overlay in the bottom-right corner shows the caller's avatar and name, with Accept and Decline buttons; a camera icon indicates an incoming video call; a ringtone plays for incoming calls
- **Video call overlay** — when a video call is active the UI switches to a full-screen overlay: remote video fills the frame, a mirrored self-preview appears in the bottom-right corner, and a gradient controls bar offers mute, camera toggle, and end-call buttons; audio-only calls show a slim bottom bar with a live duration timer
- **Call settings dropdown** — a chevron next to the call button in the DM header opens a floating settings panel with three sections: microphone (device selector with live input-level bar), camera (device selector with mirrored live preview), and speaker (device selector with a test-tone button; hidden when `setSinkId` is not supported, e.g. Linux Tauri); selections are persisted in localStorage and applied to subsequent calls

### Fixed
- **Ringtone playing when chatting** — the incoming-call ringtone was started unconditionally when the overlay component mounted (always present in `App.vue`); changed to a `watch` on call phase so it only plays when the phase transitions to `ringing`

## v0.9.12 — 2026-04-30

### Added
- **Git issue linking on cards** — each card can store an external issue URL and a short reference (e.g. `#42`). The URL is entered in the card detail; the reference is auto-filled from the URL path (`/issues/`, `/pull/`, `/merge_requests/`) but can be edited manually. The fields are saved on every card save. The section is optional and toggled through the card ⋮ sections menu.
- **Group team chat** — every user group now has a dedicated group conversation automatically created with it. Members added or removed from the group are kept in sync with the conversation. Group renames and avatar changes propagate to the linked conversation. Deleting a group removes the conversation and all its messages. Existing groups at startup receive a linked conversation via a one-time migration.
- **PDF font and language selects in Time Tracking** — the weekly timesheet export bar and the Report tab filter bar both now show PDF Font and PDF Language dropdowns (same options as in the main Time Report view). Selections are persisted in localStorage separately from the report view and passed to the backend PDF generator.

### Fixed
- **Backup email HTML rendering** — table row background colours were set on `<tr>` elements, which most email clients ignore; moved to `background-color` on each `<td>` and `<th>` individually so alternating-row shading renders correctly.

### Changed
- **Card ⋮ sections menu sorted alphabetically** — options are now sorted by their translated label in the active UI language.

## v0.9.11 — 2026-04-29

### Added
- **Company logo and name in time-tracking reports** — the Time Tracking → Report tab now shows the company logo and name in the report header, matching the style of the main Time Report.
- **Company logo and name in time-tracking PDF export** — the PDF export of the time-tracking sheet and report now includes the company logo and name in the document header. Static frontend assets (e.g. `/logo.svg`) are resolved correctly in addition to uploaded logos.

### Fixed
- **PDF export shows the correct employee name** — when an admin or time-tracking viewer exports a PDF for a specific employee, the header now shows that employee's name instead of the exporter's own name. Exporting for "All Employees" shows a translated label.
- **Sidebar indentation inconsistency** — items in "All Projects" and "All Customers" are now indented to the same level as items in "Favourite Projects" and "Favourite Customers".
- **Time-tracking table unreadable in dark theme** — all hardcoded light colours in the weekly timesheet table (backgrounds, borders, text) have been replaced with CSS custom properties so the table renders correctly in dark and black themes.

### Changed
- **Time-tracking report columns align across groups** — each week/month group in the Report tab now uses `table-layout: fixed` with shared column widths, so Date, Customer, Project, Activity and Time columns line up consistently across all groups.

## v0.9.10 — 2026-04-29

### Added
- **Separate light/dark branding logos for the login screen** — Admin Settings → Branding now accepts a second company logo for dark backgrounds. The login page automatically shows the appropriate logo based on the active login theme; when the theme is set to "System", the logo switches live as the OS dark/light preference changes.

### Fixed
- **New cards appear on the board immediately** — creating a card via the "Add card" form now adds it to the board instantly, without requiring a page refresh.
- **WebSocket connections no longer rejected in production** — same-origin browser clients (frontend served by the backend) were rejected by the WebSocket upgrader because the backend host was not in the `allowed_origins` list. Same-origin connections are now always permitted.

## v0.9.9 — 2026-04-29

### Added
- **Card sections visibility menu** — a ⋮ button in the top-right corner of the card detail lets users toggle seven sections on/off (Labels, Tags, Attachments, Checklist, Sub-cards, Linked Cards, Watchers). Sections are hidden by default when empty and revealed only when content is added. Preferences are saved per-browser.
- **All assignees shown on card tiles** — the board card now shows avatars for both the primary assignee and all extra assignees (stacked, up to 3 + overflow counter).

### Fixed
- **Clearing the primary assignee now persists** — selecting "—" in the Assignee dropdown and pressing Save correctly removes the assignee from the card; previously the null value was silently ignored by the backend.
- **Primary assignee excluded from Extra Assignees list** — the selected primary assignee is no longer shown as a toggle chip in the Extra Assignees section; selecting someone as primary also removes them from extra assignees automatically.

### Changed
- **"Assignees" renamed to "Extra Assignees"** — the multi-assignee chip picker below the primary assignee dropdown is now labelled "Extra Assignees" in all 12 supported languages to distinguish it from the primary assignee field.
- **Upgraded Vite 5 → 8** — build tooling updated to Vite 8 (Rolldown bundler); `@vitejs/plugin-vue` upgraded to v6; `rollupOptions` renamed to `rolldownOptions` in config. Build output and chunk structure are unchanged.
- **Fixed invalid `:deep()` usage in global CSS** — removed Vue-specific `:deep()` combinator from global stylesheet and corrected double-`:deep()` chains in scoped component styles; eliminates lightningcss warnings introduced by Vite 8.
- **npm audit: postcss bumped to 8.5.12** — resolves moderate XSS advisory (GHSA-qx2v-qp2m-jg93).

## v0.9.8 — 2026-04-28

### Added
- **Black theme** — new "Black" option in Settings uses a true `#000` background and matching dark chrome for OLED/pure-black setups.
- **Demo seeder configures company branding** — the seed tool now writes `company_name`, `company_logo`, and `login_branding_enabled` so the branded login panel is active out of the box after seeding.

### Fixed
- **Dark mode native `<select>` dropdowns** — `color-scheme` is now declared on `:root` and `[data-theme="dark"]` so browser-native dropdowns render in the correct colour scheme rather than always appearing in light mode.
- **Stray text insertion cursor on UI chrome** — clicking non-editable areas of the interface no longer shows a text-insertion caret; `user-select: none` is set globally and restored only on inputs and text areas.
- **Blinking caret in login branding panel** — when the branding API response arrives and the layout switches from plain to split, the login input is now re-focused so the browser caret no longer appears inside the brand panel.

### Changed
- **Login branding panel logo size** — maximum size increased from 180 × 120 px to 240 × 180 px so the company logo has more visual presence while retaining breathing room.
- **About modal** — description now explicitly mentions both Kanban and Scrum boards.

## v0.9.7 — 2026-04-28

### Added
- **F5 / Ctrl+F5 reload in Tauri desktop client** — pressing F5 (with or without modifiers) now reloads the page in the desktop app, matching the behaviour of the web browser.

### Fixed
- **IP allowlist now only restricts API-key requests** — previously the allowlist blocked all traffic including browser sessions, making it impossible to recover from a misconfigured setting via the web interface; now only `X-API-Key`-authenticated requests are subject to the restriction.
- **Training seeder `--reset` now removes groups correctly** — the cleanup query was matching `"Holy Grail __"` but groups are named `"Shrubbery Bringing XX"`; orphaned groups caused a UNIQUE constraint crash on re-runs.
- **Training seeder group creation is idempotent** — a `FirstOrCreate` replaces the bare `Create` so a partial previous run cannot block a fresh seed attempt.
- **Training seeder expanded character list** — 21 new Monty Python / Arthurian characters added; two duplicates removed (Bedivere The Wise / Sir Bedevere, French Taunter / The French Taunter).

## v0.9.6 — 2026-04-28

### Fixed
- **Column drag-to-reorder now works for project owners and admins** — the Sortable instance was initialised with `disabled: true` because project membership hadn't loaded yet at setup time; a watcher on `canManageColumns` now enables it as soon as the member list resolves.

## v0.9.5 — 2026-04-28

### Added
- **Admin Users table: Last Password Change column** — shows when each user last changed their password; tracked on self-service change, password reset, and admin-forced reset.
- **Admin Users table: dedicated MFA column** — MFA status moved from the Status badge to its own column for clarity.
- **Password change period policy** — new Admin → Settings → Password Policy field sets a maximum password age in days (0 = disabled); on login, users whose password has expired are redirected to Settings and shown a banner prompting them to change it.
- **People list filtered by customer membership** — in the DM / People selector, non-admin users who have explicit customer assignments only see colleagues who share at least one customer with them; users with no customer restrictions continue to see everyone.
- **Sortable name column in Admin tables** — Users, Groups, Customers, and Projects lists all have a clickable name header (↑/↓) that toggles ascending / descending alphabetical order; sorting is client-side.
- **Assign groups in Edit User modal** — the Edit User modal now includes a group chip picker (purple) alongside the existing project and customer pickers; the user's current group memberships are shown and can be changed in one save.
- **Reaction tooltip showing reactor names** — hovering over a reaction pill in project chat or DMs shows a styled tooltip with the names of all users who sent that reaction ("You, Alice, Bob"); replaces the previous plain emoji+count tooltip.
- **WarmDesk wordmark on login screen** — company name displayed next to the logo on the login page.

### Fixed
- **Company logo URL resolved to absolute path in Tauri desktop client** — relative `/uploads/…` logo URLs are now resolved against the configured server URL so the logo loads correctly in the desktop app.
- **Back button for sub-cards and linked cards** — a ← [parent card title] breadcrumb is shown at the top of nested card modals; clicking it returns to the originating card.

## v0.9.4 — 2026-04-27

### Added
- **Checklist drag-to-reorder** — drag the ⠿ handle on any checklist item to reorder it; new position is saved immediately.
- **Checklist real-time sync** — checklist changes (add, tick, edit, delete, reorder) now broadcast via WebSocket and are reflected instantly for all users viewing the same card.
- **Back navigation in nested card modals** — opening a sub-card or a linked card now shows a **← [parent card title]** breadcrumb at the top of the modal; clicking it closes the nested view and returns to the originating card.
- **ARM64 server builds** — `make build-arm64` cross-compiles all server binaries for linux/arm64 into `dist/arm64/` (no C cross-compiler needed).
- **ARM64 desktop builds** — `make appimage-arm64`, `make deb-arm64`, and `make rpm-arm64` targets build Tauri desktop packages for aarch64.
- **`make help`** — lists all Makefile targets grouped by category with short descriptions.

### Fixed
- **Login page version label** — version text is now correctly centred on the login screen.

## v0.9.3 — 2026-04-27

### Added
- **API IP allowlist** — new system setting (`allowed_ips`) restricts API and Swagger UI access to a comma-separated list of IP addresses or CIDR ranges; empty (default) allows all.
- **Short-lived WS and media tickets** — Tauri desktop clients no longer put the long-lived JWT in WebSocket URLs or attachment `<img src>` URLs; a 30-second WS ticket and a 5-minute media ticket are issued per-connection instead.

### Security
- **Startup safety checks** — server refuses to start if `jwt_secret` is still the default value, or if `allowed_origins` contains `*` in release mode.
- **HSTS and CSP headers** — added `Strict-Transport-Security` and `Content-Security-Policy` to the nginx and Apache reverse-proxy templates.
- **Server-side MIME detection** — uploaded file MIME type is now detected from the first 512 bytes of the saved file; the client-supplied `Content-Type` header is ignored.
- **bcrypt cost pinned at 12** — password hashing cost raised from the library default (10) to 12.
- **Wildcard CORS blocked in production** — `allowed_origins: "*"` is rejected at startup when `gin_mode` is `release`.
- **`gin_mode` defaults to `release`** — debug mode must now be explicitly opted in; the server previously defaulted to debug.
- **Swagger UI gated by IP allowlist** — the `/swagger/*` routes are now subject to the same `allowed_ips` check as the API.
- **Error messages sanitised** — handler JSON responses no longer include raw `err.Error()` text that could expose internal paths or driver details.

### Changed
- **Backup filenames** include a 4-character random hex suffix to prevent collisions when two backups are triggered within the same minute.
- **Documentation corrections** — language count corrected to 12 (was 5/6 in several places); artifact filenames use consistent `<version>` placeholder; various stale references updated.

## v0.9.1 — 2026-04-26

### Added
- **Time tracking: view all users** — admins and users with the `time_tracking_viewer` flag can switch between individual users or "All employees" using a dropdown in the time tracking header; the sheet is read-only when viewing another user.
- **Ansible `from_vars` module** — new `ansilabnl.warmdesk.from_vars` module reads one or more YAML variable files and provisions all WarmDesk resources (users, customers, contracts, projects, columns, labels, cards, groups, system settings) in dependency-correct order regardless of declaration order in the file.

### Changed
- **Ansible `project` module** — `customer` is now a required parameter when `state=present`, matching the server-side enforcement introduced in v0.9.0.
- **Session timeout setting** — the "minutes (0 = disabled)" label no longer wraps; input field and row width increased.

### Fixed
- **Login screen inside app shell** — after a server restart and page refresh the login form was rendered inside the sidebar/header shell instead of as a standalone page; fixed by initialising the auth session before installing Vue Router so the navigation guard sees the correct state on first load.

### Security
- **TLS minimum version** — database TLS connections now enforce a minimum of TLS 1.2 (`tls.VersionTLS12`).
- **Gravatar hash** — switched from MD5 to SHA-256 for Gravatar URL generation; Gravatar has supported SHA-256 since 2024.
- **nginx H2C smuggling** — WebSocket proxy location now only accepts `Upgrade: websocket`; general proxy location no longer forwards `Upgrade`/`Connection` headers, preventing h2c upgrade smuggling.
- **Training seed** — replaced `math/rand` with `math/rand/v2` (auto-seeded, no deprecated `rand.Seed` call).

## v0.9.0 — 2026-04-24

### Added
- **Per-user time tracking module** — optional time registration feature (toggle in User Settings); when enabled, a weekly grid view (Exact Online-style) lets users log hours per customer/project/activity row across the seven days of any week; week navigation, row add/edit/delete, and day/row/grand totals included.
- **Time tracking report tab** — period selector (week / month / year) with grouped entries, subtotals, and grand total; accessible from the same Time Tracking view.
- **Time tracking exports** — weekly timesheet and report can be exported to XLSX (SheetJS, client-side) and PDF (backend, gofpdf); all 12 UI languages supported in PDF output.
- **Comment-level time logging** — "Time Spent" field removed from the card form and moved to each comment; users with time tracking enabled see an optional hours/minutes input when writing a comment; the card header shows the running total across all comments.
- **Auto time entry from comments** — when a user with time tracking enabled logs time in a card comment, a `TimeEntry` is created automatically: today's date, the card's project, the project's customer (if set), and the first sentence of the comment as description; editable in the Time Tracking view.
- **Overdue card highlight** — board cards whose due date has passed now show a subtle orange background and border tint, distinct from the red overdue text on the date badge.
- **Customer required on projects** — every project must be linked to a customer; the create-project dialog and project settings no longer allow saving without one; enforced server-side as well.
- **Demo seed: time entries** — 60 realistic time entries (5 users, 2 weeks back) added to the demo dataset; time tracking enabled for the five active demo users.

### Changed
- **Hourly version check** — the client previously checked GitHub/server versions only on login; it now re-checks every hour while the user is logged in.
- **Frontend bundle split** — Vite `manualChunks` splits Vue core, vue-i18n, marked/highlight.js, and axios/DOMPurify into named chunks; the main `index.js` bundle dropped from ~504 kB to ~298 kB.

## v0.8.15.2 — 2026-04-24

### Fixed
- **Windows desktop build** — Tauri requires strict semver (`MAJOR.MINOR.PATCH`); four-part tag names such as `v0.8.15.1` were rejected. Version-stamp steps in the release and manual Windows workflows now truncate any fourth component before writing `tauri.conf.json` and `Cargo.toml`.

## v0.8.15.1 — 2026-04-24

### Fixed
- **Windows CI build** — removed `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true` from all five GitHub Actions workflows; the flag was forcing the Actions runner infrastructure itself to use Node 24, breaking the `actions/setup-node` step. The `node-version: '24'` setting already ensures Node 24 is used for build steps.

## v0.8.15 — 2026-04-24

### Added
- **Gravatar avatars restored** — user profile pictures are fetched from Gravatar again. A new toggle in Admin → System Settings enables or disables Gravatar; when disabled, initials-based DiceBear avatars are used instead.
- **Rate limiting on auth endpoints** — login and MFA verify are limited to 10 attempts per 15 minutes per IP; registration to 5 per hour; password reset to 5 per 30 minutes.
- **Security response headers** — `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, and `Referrer-Policy` headers are now sent on every response.
- **Auth audit log** — structured log lines are written for all authentication events (login, logout, register, MFA verify/enable/disable, password change, password reset).
- **Image upload MIME validation** — uploaded images are validated against magic bytes so files with a mismatched Content-Type are rejected.

### Fixed
- **Login redirect loop on startup** — the token-refresh interceptor no longer redirects to `/login` when the page is already `/login`, preventing an infinite reload loop on first visit with no active session.
- **PostgreSQL backup DSN leak** — `pg_dump` is now invoked with `PGPASSWORD` set as an environment variable instead of embedding credentials in the DSN argument; error messages no longer expose connection details.
- **Backup filename collisions** — backup filenames now include a random 4-byte hex suffix to prevent overwrites when two backups are triggered within the same minute.

### Changed
- **Browser auth uses httpOnly cookies** — login, register, refresh, and MFA flows now issue `httpOnly SameSite=Strict` cookies for browser clients so tokens no longer touch JavaScript. A new `POST /auth/logout` endpoint clears the cookies server-side.
- **WebSocket auth accepts cookies** — the WebSocket upgrade handler now accepts the access-token cookie (browser mode) in addition to the `?token=` query param (Tauri/API mode).
- **WebSocket origin validation** — the WebSocket upgrader now enforces the same `allowed_origins` list as the HTTP CORS middleware instead of allowing all origins unconditionally.
- **Search input hardened** — minimum query length raised to 3 characters; SQL LIKE wildcards (`%`, `_`) are stripped before the query reaches the database.
- **API key auth header-only** — `X-API-Key` is the sole accepted method; the `?api_key=` query parameter is no longer supported.
- **CORS wildcard warning** — a warning is logged at startup when `allowed_origins` is set to `*`.

## v0.8.14 — 2026-04-23

### Changed
- **Hover reaction toolbar** — incoming/group messages in project chat and DMs now show quick reaction buttons on hover plus a `+` action that opens the full inline emoji picker.
- **Quick-reaction configuration** — hover quick reactions are now sourced from one shared frontend constant (`QUICK_REACTION_EMOJIS`) to keep project chat and DM behavior in sync.
- **User docs and screenshots refreshed** — chat reaction documentation now reflects the hover toolbar + full picker flow, with updated screenshots in both the User Guide and README.

### Fixed
- **Reaction emoji whitelist mismatch** — backend reaction validation now allows the newly exposed quick reactions (`🤣`, `✅`, `❌`, `🤘`) so all hover-toolbar emojis can be selected successfully.

## v0.8.13 — 2026-04-23

### Added
- **Project and group avatars** — projects and groups now support avatar URLs in the backend and admin/project settings UI, including upload and clear actions.
- **Default project/group avatars in demo generators** — both `warmdesk-seed` and `warmdesk-training` now generate avatars for projects and groups.
- **Global breadcrumbs bar** — top-level breadcrumb navigation with back/forward controls was added for faster movement through views.

### Changed
- **Avatar rendering across the UI** — project/group/customer/user/conversation avatars are now shown consistently in key surfaces, including sidebar sections, project toolbars/cards, admin tables, customer detail, and group lists.
- **Breadcrumb preference** — users can now show or hide breadcrumbs from settings; the preference is persisted per user.
- **Emoji shortcode support in compose** — chat inputs now support Slack/GitHub-style `:shortcode:` replacements using the full `gemoji` dataset (including aliases such as `:fingerscrossed:`).

### Fixed
- **Emoji picker keyboard UX** — when only one emoji match remains, `Enter` selects it; pressing `Escape` closes the picker and restores focus to the originating input/editor.
- **Chat focus continuity** — compose/edit focus handling was tightened around emoji picker close/send flows in project chat, direct messages, and card editors.

## v0.8.12 — 2026-04-23

### Added
- **Emoji autocomplete in edit mode** — the inline emoji picker (triggered by `:`) now works when editing existing messages, matching the behavior of the main chat compose area.
- **Desktop app URL override flag** — the Tauri client now accepts `--url=<http(s)://host:port>` (or `--url <...>`) to override the configured server URL for the current launch only; the override is runtime-only and is not persisted to local storage.
- **Desktop CLI help output** — `--help` / `-h` now prints supported startup flags (`--help`, `--version`, `--maximized`, and `--url`).

### Changed
- **Unfocused-message notifications** — chat and direct-message notification checks now use both focus and visibility state, so OS notifications trigger more reliably when the desktop client is backgrounded/minimized.

## v0.8.11 — 2026-04-22

### Added
- Desktop notifications for direct messages and unread counts.
- Inline emoji picker triggered by ':' in the compose textarea, with initial-search support for fast emoji insertion.
- Demo seeder: detailed summary box listing projects, conversations, customers, and groups.
- **Customer avatars in Admin** — customer avatars are now displayed in the Admin → Customers list view for better visual identification.
- **User avatars in Admin** — user avatars (Gravatar) are now displayed in the Admin → Users list view.
- **User avatar in Edit User modal** — the user's avatar is now shown at the top left of the Edit User dialog for a better overview.
- **Chat avatar placement** — user avatars in the chat are now positioned at the top-left (or top-right for own messages) of the message bubble.

### Fixed
- Restore syntax token colors inside chat code blocks while removing per-line highlights so tokens keep color without boxed backgrounds.
- Fix color and avatar-size issues in chat messages.

### Changed
- Code block styling: uniform background across tokens, darker bubble tint for readability, and white background for own bubbles to improve contrast.
- Removed per-token background highlights to avoid line-highlighting; tokens now inherit color only.
- Updated TODO list.

All notable changes to WarmDesk are documented here.

## v0.8.10 — 2026-04-21

### Added
- **URL link previews in chat** — the first URL in any chat or DM message automatically fetches Open Graph metadata (title, description, thumbnail, site name) and renders a compact preview card below the message; previews are cached per session so repeated URLs cost only one request; the backend endpoint includes SSRF protection (private IP ranges and hostnames are blocked, 5 s timeout, 64 KB response limit)
- **Syntax highlighting in chat** — fenced code blocks (` ```lang ... ``` `) in chat messages and DMs are now highlighted using highlight.js; supported languages include Bash, CSS, Go, INI/TOML, JavaScript/TypeScript, JSON, Python, Rust, SQL, XML/HTML, and YAML; dark mode uses a matching dark code theme

### Fixed
- **Dark-mode code blocks unreadable** — inline `code` elements used `rgba(0,0,0,.08)` which is near-invisible on the dark background; code block `pre` used `--color-bg` (`#0f172a`) which mismatched the highlight.js dark theme background (`#1e2130`); both are now corrected with explicit dark-mode overrides
- **Paste images on Linux/Tauri** — `e.clipboardData.items` is empty in WebKitGTK (Linux desktop app); the paste handler in both project chat and DMs now falls back to `navigator.clipboard.read()` when running inside Tauri, so Ctrl+V image paste works on Linux

### Changed
- **Sidebar section labels renamed** for consistency: "Starred Projects" → "Favorite Projects", "Customers" (starred) → "Favorite Customers", "Favorites" (people) → "Favorite People"
- **Customer creation moved to Admin Panel** — the "Customers…" shortcut link in the sidebar is removed; customers are now created exclusively via **Admin → Customers**

## v0.8.9 — 2026-04-21

### Added
- **User groups** — admins can create named groups (`Admin → Groups`) and assign members, project roles, and customer roles to a group; group access is additive: a user gains the highest role they hold either directly or through any group they belong to
- **`Admin → Groups` tab** — full CRUD for groups in the Admin panel (positioned after Users); each group detail panel shows members, project access, and customer access with add/remove controls
- **Demo seeder groups** — `warmdesk-seed` now creates three demo groups: _Frontend Team_ (sarah, marc, priya, james; access to Website Redesign, Mobile App v2, Marketing), _DevOps Team_ (marc, lisa, raj; access to DevOps & Infra, Product Platform), and _Acme Stakeholders_ (admin, sarah, marc; viewer access to Acme Corporation)
- **Ansible `ansilabnl.warmdesk.group` module** — manage groups, members, project access, and customer access idempotently; supports declarative reconciliation (provide full desired lists to add/remove in one pass) and check mode; collection bumped to v0.3.0

### Fixed
- **Group members not persisted** — `AdminAddGroupMember` (and Project/Customer variants) used GORM `FirstOrCreate` on composite-primary-key models without an auto-increment ID, which silently failed to create the row; replaced with an explicit `Count` + `Create` pattern

## v0.8.7 — 2026-04-20

### Fixed
- **Swagger UI redirect caused startup panic** — the `/swagger/` route registered in v0.8.6 conflicted with Gin's `/*any` wildcard; removed the trailing-slash route (gin-swagger handles that redirect internally); `/swagger` → `/swagger/index.html` still works correctly

## v0.8.6 — 2026-04-20

### Added
- **`api_key` lookup in Ansible collection** — new `ansilabnl.warmdesk.api_key` lookup plugin lists personal or project-scoped API keys and enriches each result with `username` and `display_name` so you can see who owns each key; filter by key name via `_terms` or omit to return all keys

### Fixed
- **Admin Deactivate/Delete buttons visible for own account** — both buttons are now hidden (not just disabled) for the logged-in admin's own row in Admin → Users
- **Swagger UI requires typing `index.html`** — `/swagger` and `/swagger/` now redirect to `/swagger/index.html`

## v0.8.5 — 2026-04-20

### Added
- **`board_type` in Ansible collection** — the `ansilabnl.warmdesk.project` module now accepts a `board_type` parameter (`kanban` or `scrum`) at creation time; matches the API's immutability rule (silently ignored on update); collection bumped to v0.2.0

### Fixed
- **Board auto-refresh on WebSocket reconnect** — if the WebSocket dropped while the board was open (network hiccup, proxy timeout), the hub was torn down and any card created via the API during that window was silently lost; the board now performs a silent background refresh whenever the WebSocket reconnects, recovering any missed events without a visible reload
- **WarmDesk favicon showed CoWorker branding** — `favicon.svg` used the indigo colour scheme and Kanban-columns icon inherited from CoWorker; replaced with the WarmDesk desk-and-cup motif using the green brand palette (`#1D9E75` background)
- **Admin could delete their own account** — the `DELETE /api/v1/admin/users/:id` endpoint now returns `403 Forbidden` when the requesting admin targets their own user ID; the Delete button in Admin → Users is also disabled for the logged-in user's own row

## v0.8.4 — 2026-04-20

### Added
- **Card references in chat** — typing `#PRJ-42` (key prefix + card number) in any chat message or DM renders an inline badge that links directly to that card; clicking navigates to the project board and opens the card detail panel; Ctrl+click or middle-click opens the card in a new browser tab; the backend resolves the reference via a new `GET /api/v1/cards/resolve/:ref` endpoint that looks up the card by key prefix and number

### Fixed
- **Tauri `.desktop` file missing Name and Comment** — `{{app_name}}` and `{{short_description}}` placeholders in the Tauri desktop template were not substituted by Tauri v2; values are now hardcoded directly in the template; `StartupWMClass`, `Categories`, and `Keywords` also corrected

## v0.8.3 — 2026-04-20

### Added
- **Grouped chat layout** — a fifth chat layout (Discord/Mattermost-style) available in the layout picker of both the project chat panel and the DM view; consecutive messages from the same sender within a 5-minute window collapse so the avatar and name appear only on the first message of each group, giving a clean threaded feel
- **Paste images into chat** — pressing Ctrl+V (or ⌘V) while the compose textarea is focused pastes any clipboard image directly as a pending attachment; the image previews immediately in the compose area using a local object URL, then uploads and appears inline in the chat when the message is sent; works in both the DM view and project chat panel

### Fixed
- **Inline image display** — attached images now render at full natural size (up to 520×480 px) directly in the chat message instead of showing as a small thumbnail or opaque chip; images in received messages load correctly via a `?token=` query-parameter added to attachment URLs so browser `<img>` tags can authenticate without custom headers
- **Compose textarea loses focus after send** — keyboard focus now returns to the compose bar immediately after a message is sent, so the user can keep typing without clicking back in
- **Wording** — README and `.desktop` files updated from "Kanban boards" to "Kanban and Scrum boards"

## v0.8.2 — 2026-04-20

### Added
- **Native desktop notifications** — the Tauri desktop app now uses `tauri-plugin-notification` for OS-level notifications (libnotify on Linux, system notifications on macOS/Windows); the browser build continues to use the Web Notifications API; both paths respect the bell toggle and only fire when the window is not focused
- **`trusted_proxies` config setting** — comma-separated list of reverse-proxy IPs/CIDRs that WarmDesk should trust for the `X-Forwarded-For` header; when set, `c.ClientIP()` returns the real client IP instead of the proxy address, so auth log entries show the originating IP; configurable via `trusted_proxies:` in `warmdesk.yaml` or the `TRUSTED_PROXIES` environment variable; documented in `warmdesk.yaml.example`

## v0.8.1 — 2026-04-20

### Added
- **Auth audit logging** — login, logout, failed login, registration, password change, password reset (request + confirm), MFA enable/disable, and MFA verification are now written to the server log with the event, user ID, username, and originating IP address
- **Permanently delete projects** — Admin → Projects (deleted tab) now has a **Permanently Delete** button next to Restore; hard-deletes the project and all dependent data (cards, columns, labels, sprints, topics, chat messages, attachments, API keys, etc.) in the correct dependency order; only available for already-soft-deleted projects
- **Four chat layouts** — both the project chat panel and direct-message view now offer four selectable layouts via an icon picker in the chat header: **Bubble** (original WhatsApp-style), **Comfortable** (Slack-style, all messages left-aligned), **Compact** (IRC/terminal, no avatars, inline timestamp + name), and **Cozy** (document-style, no bubble background, own messages accented with a left border); choice is persisted in `localStorage` and shared across all chat views
- **In-app new-message toast** — a dismissible pop-up slides up from the bottom of the chat area when a new message arrives from someone else; shows sender name and a plain-text preview (code blocks replaced with `[code]`); controlled by a bell toggle button in the chat header; setting is persisted in `localStorage`
- **OS desktop notifications** — when the browser window is not focused, new messages trigger a native OS notification (browser Notifications API); permission is requested automatically on first load (or when the bell is enabled); respects the same bell toggle as the in-app toast; works in browser and Tauri desktop app
- **Document-title unread indicator** — the browser tab / taskbar title shows `● WarmDesk` when there are unread direct messages or unread project chat messages, and reverts to `WarmDesk` when all are read

### Fixed
- **Backup email filename invisible on light backgrounds** — the filename column in backup notification emails had no explicit text colour, rendering as white-on-white in some email clients; now explicitly set to `#333`
- **Sprint date fields respect user date format** — the start and end date inputs in the Backlog sprint editor now use the same overlay date-picker pattern as card due dates, displaying dates in the user's preferred format while storing ISO values internally
- **False-positive unread DM indicator on login** — conversations that had never been opened on the current device defaulted their last-seen timestamp to epoch (1970), making every conversation appear unread on every login; they are now seeded to their current `updated_at` on first load so only genuinely new messages trigger the indicator
- **Unread dot persists after reading messages** — when a new message arrived in an open conversation via the 5-second poll, the sidebar unread dot could reappear because `markConvSeen` was only called on conversation open, not on each poll; it is now also called whenever the user is scrolled to the bottom (i.e. reading the latest messages)
- **Admin "Permanently Delete" button did nothing** — `useI18n` was never imported in `AdminView.vue`, causing a silent `ReferenceError` inside the confirm dialog before any action was taken

### Changed
- **Chat bubble max-width removed** — message bubbles in the DM view were capped at 480 px; the cap is removed so wide content (ASCII tables, long code lines) uses the full available width with horizontal scroll inside the bubble
- **Sidebar unread check interval** — `checkUnread` now polls every 5 seconds (previously 30) via a dedicated interval, independent of the presence/user-list poll (which remains at 10 seconds); unread indicators now feel near-instant

## v0.8.0 — 2026-04-19

### Added
- **Scrum support** — projects now have a board type (Kanban or Scrum) set at creation time; Scrum projects unlock two new views: **Backlog** (two-panel sprint planner with drag-and-drop card assignment, sprint CRUD, goal and date editing, and a velocity chart of completed sprints) and **Sprint Board** (board view filtered to the cards in the active sprint); sprint lifecycle: planning → active → completed; completing a sprint returns unfinished cards to the backlog automatically; story points field on cards (enabled via Admin → Settings → Scrum Story Points); SP badge on card tiles and a done/total summary in the sprint board toolbar
- **Direct add-card UX** — clicking **+ Add card** on a column now opens the full card detail form immediately (no intermediate title-only modal); only the fields that make sense before a card exists are shown (title, description, priority, dates, time, single assignee, story points); clicking **Create** saves the card and closes the panel
- **Board type locked after creation** — board type (Kanban / Scrum) is chosen in the **New Project** dialog and cannot be changed later; the Project Settings general tab now shows it as a read-only badge

### Changed
- **Demo data** — seeder (`go run ./cmd/seed`) now creates five projects: three Kanban (Website Redesign, Mobile App v2, DevOps & Infrastructure) and two new ones — **Product Platform** (Scrum, prefix `PLT`) with a completed sprint, an active sprint, and a planning sprint fully populated with cards and story points; **Marketing Campaigns** (Kanban, prefix `MKT`) with cards across Ideas / Planned / In Progress / Published columns

## v0.7.12 — 2026-04-19

### Added
- **Scrum Story Points** — new admin toggle (**Admin → Settings → Scrum Story Points**); when enabled, a Story Points number field appears on every card detail panel and a coloured SP badge is shown on the card tile on the board; the setting takes effect immediately for all open sessions without a page reload
- **Customer filter on the project dashboard** — when projects span multiple customers, a customer selector appears in the dashboard header; choose a specific customer to show only their projects, or **All Customers** to see everything; drag-to-reorder is suppressed while a filter is active

## v0.7.11 — 2026-04-19

### Fixed
- **Company logo stored as data URI not shown in emails** — when the logo in Admin → Settings is pasted or saved as a `data:…` URI (rather than an uploaded file or external URL), it was silently discarded by the email branding logic; it is now passed through directly

## v0.7.10 — 2026-04-19

### Fixed
- **Company logo embedded in emails** — the company logo is now inlined as a base64 data URI in all outbound emails (same technique as the WarmDesk footer icon); uploaded logos are read from disk, external URL logos are fetched at send time; no longer requires `base_url` to be configured and no longer broken by email clients that block remote images

## v0.7.9 — 2026-04-19

### Fixed
- **Gin "trusted all proxies" warning** — `SetTrustedProxies(nil)` is now called at startup so Gin no longer trusts all proxy headers by default; set `["127.0.0.1"]` in code if WarmDesk sits behind a local reverse proxy and you need real client IPs from `X-Forwarded-For`
- **Config file discovery accepts `.yml`** — when no explicit path is given (via `--config` or `CONFIG_FILE`), WarmDesk now tries `warmdesk.yaml` first and falls back to `warmdesk.yml`; previously only `.yaml` was recognised

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
