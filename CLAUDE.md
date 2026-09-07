# WarmDesk — Developer Guide for Claude

Cursor loads the same content automatically via [.cursor/rules/warmdesk-developer-guide.mdc](.cursor/rules/warmdesk-developer-guide.mdc). Update both files when you change this documentation.

WarmDesk is a self-hosted project management tool (Kanban boards, team chat with one-to-one WebRTC voice & video, LiveKit-based group calls for conversations with 3+ members, discussions, time reporting). It has a Go backend and a Vue 3 frontend; both live in this repository. A Tauri wrapper produces native desktop apps from the same frontend code.

---

## Development

```bash
# Backend — runs on http://localhost:8080
cd backend
go run .

# Frontend — runs on http://localhost:5173, proxies /api to :8080
cd frontend
npm install
npm run dev
```

No database setup required: SQLite is the default and the file (`warmdesk.db`) is created automatically.

---

## Build

### Prerequisites

| Tool | Min version | Notes |
|---|---|---|
| Go | 1.26 | Install from go.dev/dl; add `/usr/local/go/bin` to `PATH` |
| Node.js | 20 LTS | Via NodeSource or nvm; npm is bundled |
| Rust + Cargo | 1.85 | Already present on most systems; needed for AppImage/desktop |

AppImage additionally requires these system libraries (Debian/Ubuntu):
```bash
sudo apt install libwebkit2gtk-4.1-dev libgtk-3-dev librsvg2-dev \
                 libayatana-appindicator3-dev libssl-dev \
                 patchelf squashfs-tools \
                 gstreamer1.0-tools gstreamer1.0-plugins-base \
                 gstreamer1.0-plugins-good gstreamer1.0-plugins-bad \
                 gstreamer1.0-pulseaudio
```
The `gstreamer1.0-*` packages are not linked at build time — `make appimage`'s
`_bundle-gst-appimage` step copies them from the build host straight into the
AppImage after `tauri build` (`linuxdeploy-plugin-gstreamer` doesn't reliably
invoke, so this project post-processes the AppImage directly instead). Without
`plugins-base`/`plugins-good`/`pulseaudio`, camera/microphone selection in the
built app silently shows no devices — WebKitGTK's bundled GStreamer core has no
`v4l2src`/`pulsesrc`/`autoaudiosrc` elements to enumerate with. Without
**`plugins-bad`** specifically (which provides the `webrtcbin` element),
`RTCPeerConnection` is **entirely undefined** in the built app — calls fail
with `ReferenceError: Can't find variable: RTCPeerConnection` even though
device selection itself still works fine, since that only needs the
`good`/`base`/`pulseaudio` plugins. `_bundle-gst-appimage` prints a warning if
no `webrtc*.so` plugin ends up bundled, precisely to catch this. Must be built
on Ubuntu 24.04 so the copied plugins are ABI-compatible with the webkit2gtk
GStreamer core also bundled from this host (e.g. Fedora's GStreamer is a
different major version and cannot be mixed in).

### Targets

```bash
make build          # builds frontend + backend → dist/ (includes user + admin guide PDFs)
make docs-pdf-guides  # user-guide.pdf + admin-guide.pdf only (also run by make build)
make docs-pdf       # alias for docs-pdf-guides
make run            # build then run
make appimage       # Linux AppImage (requires Rust + system libs above)
make dmg            # macOS universal DMG
make windows-installer  # Windows NSIS installer
```

Guide PDFs are embedded in the server binary (`staticweb/files/docs/`) and copied to `dist/docs/`.
Authenticated users download the user and admin guides via the avatar menu → **Downloads** (`GET /api/v1/docs/user-guide.pdf` and `/docs/admin-guide.pdf`).

Production:
```bash
cd dist
WEB_DIR=./web ./warmdesk
```

### Testing on Windows

Full per-OS toolchain install steps live in `docs/dev-setup.md`. Notes specific to standing up a Windows test environment:

- **VM**: on Fedora/RHEL, `qemu-kvm` + `virt-manager` (native KVM) outperforms VirtualBox — `sudo dnf install qemu-kvm libvirt virt-manager && sudo systemctl enable --now libvirtd`, then create a VM from a Windows ISO or import Microsoft's free [Windows 11 dev environment VM](https://developer.microsoft.com/en-us/windows/downloads/virtual-machines/) (90-day eval, no product key needed).
- **Rust/Tauri builds require the MSVC C++ Build Tools** (`Microsoft.VisualStudio.2022.BuildTools` with the `VCTools` workload) — `rustup` offers to install this automatically if it's missing.
- **WebView2** is required to run the desktop app (`npm run tauri:dev` or a built installer) — pre-installed on Windows 11 and modern Windows 10; otherwise `winget install Microsoft.EdgeWebView2Runtime`.
- **`make` must be run from Git Bash, not PowerShell/cmd.** GnuWin32 `make` needs a POSIX `sh` on `PATH` to execute the Makefile's Unix-shell-syntax recipes; without one it silently falls back to `cmd.exe`, which mangles the `git describe` version stamp into the literal string `"dev"` — Tauri then rejects that as invalid semver (`tauri.conf.json > version must be a semver string`). Symptom: `The system cannot find the path specified.` near the start of the build. Fix: launch `make` from a Git Bash terminal (Git for Windows ships its own `sh.exe`) — see `docs/dev-setup.md` for the full writeup, or use WSL instead.

---

## Configuration

Config is loaded in priority order: CLI flag `--config` → `CONFIG_FILE` env var → `warmdesk.yaml` in CWD → built-in defaults. Every YAML key has a matching environment variable override.

Key settings (`warmdesk.yaml.example` has full documentation):

| Setting | Env var | Default |
|---|---|---|
| `port` | `PORT` | `8080` |
| `db_driver` | `DB_DRIVER` | `sqlite` |
| `db_dsn` | `DB_DSN` | `./warmdesk.db` |
| `db_log` | `DB_LOG` | `info` (silent / error / warn / info) |
| `db_tls_mode` | `DB_TLS_MODE` | `disable` (see Database TLS section) |
| `db_tls_ca_cert` / `db_tls_cert` / `db_tls_key` | `DB_TLS_*` | *(empty)* |
| `jwt_secret` | `JWT_SECRET` | *(server refuses to start at default)* |
| `web_dir` | `WEB_DIR` | *(empty — falls back to embedded frontend in `staticweb`)* |
| `upload_dir` | `UPLOAD_DIR` | `./uploads` |
| `max_upload_mb` | `MAX_UPLOAD_MB` | `25` |
| `redis_url` | `REDIS_URL` | *(optional — enables horizontal scaling)* |
| `allowed_origins` | `ALLOWED_ORIGINS` | `http://localhost:5173` — `*` blocked in `release` mode |
| `trusted_proxies` | `TRUSTED_PROXIES` | *(empty — trust no proxy headers; comma-separated CIDRs/IPs)* |
| `tls_cert` / `tls_key` | `TLS_CERT` / `TLS_KEY` | *(empty — set both to enable HTTPS directly)* |
| `base_url` | `BASE_URL` | *(empty — used in Swagger host and email links)* |
| `default_locale` | `DEFAULT_LOCALE` | `en` |
| `api_log` | `API_LOG` | `true` |
| `gin_mode` | `GIN_MODE` | `release` — set to `debug` for local development |
| `app_mode` | `APP_MODE` | *(empty — full WarmDesk); set to `timetracking` to run as a time-tracking-only server: boards, chat, and helpdesk routes are disabled, and the time-tracking logos replace the default WarmDesk logos everywhere (web UI, PDFs). Also available as `--mode=timetracking` CLI flag.* |
| `instance_mode` | `INSTANCE_MODE` | *(empty — production); set to `test` to serve the orange, "TEST"-ribboned logo variants everywhere (web UI, PDFs) instead of the default green ones, so a test/staging instance is never mistaken for production. Combines independently with `app_mode` — e.g. `timetracking` + `test` serves the orange time-tracking logo. Also available as `--instance-mode=test` CLI flag.* |
| `smtp.host` / `.port` / `.from` / `.username` / `.password` / `.use_tls` | — | port `587`, rest empty (overridable via system settings) |
| `livekit_url` / `livekit_api_key` / `livekit_api_secret` / `livekit_room_prefix` | `LIVEKIT_*` | *(empty — required for voice/video calls)* |
| `oauth2.google_client_id` / `.google_client_secret` / `.office_client_id` / `.office_client_secret` | — | *(empty — required for IMAP OAuth2 auth)* |

---

## Architecture decisions

### Database
- **GORM AutoMigrate runs on every startup** — no separate migration files. Adding a new field to a model struct is all that is needed; the column appears on next boot.
- Supported drivers: `sqlite`, `postgres`, `mysql`. The driver string goes in `db_driver`, the DSN in `db_dsn`.
- Card numbering (`PRJ-1`, `PRJ-2`, …) is maintained by an atomic `card_counter` increment on the `projects` table.

### Backup / restore
`handlers/backup.go` creates backups as a `warmdesk_backup_<timestamp>_<hex>.tar.gz` in `./backups/`, bundling the database dump (SQLite `VACUUM INTO`, `pg_dump`, or `mysqldump` depending on `db_driver`) together with a copy of `upload_dir` (attachments, avatars, customer logos, company branding) under a `db/` and `uploads/` path inside the archive. Bundling the upload directory can be disabled via the `backup_include_uploads` system setting (default `true`) for deployments with very large upload directories.

`POST /api/v1/admin/system/backups/upload` (`AdminUploadBackup`) accepts a `.tar.gz`/`.db`/`.sql` file — e.g. one downloaded from a different WarmDesk server — and saves it into `./backups/` under a normalised name (`validateBackupArchive`/`validateSQLiteFile` do a light content sanity check) so it's listed and restorable alongside locally created backups.

`AdminRestoreBackup` takes a `mode`: `"replace"` (default) or `"merge"`.
- **replace** extracts the archive to a temp dir, restores the database, then replaces `upload_dir` wholesale with the archived copy (`replaceDir` — removes existing contents first). Backups created before archive-bundling was added (`warmdesk_db_<timestamp>_<hex>.db`/`.sql`, database only) are still listable, downloadable, and restorable — `isBackupFile`/`backupSortKey` handle both filename formats so old and new backups sort chronologically together in the admin UI and during pruning (`backup_keep`).
- **merge** adds the backup's data to what's already there instead of wiping it: uploaded files are copied in without deleting existing ones (`mergeDir` — collision-free since filenames are random hex), and (SQLite only) `mergeSQLiteDatabase` `ATTACH`es the backup's database file and runs `INSERT OR IGNORE` per common table/column, so rows are added only when their primary key doesn't already exist locally. This is a genuine limitation for cross-server merges: two independently-run servers have no shared ID space, so a row whose ID collides with an unrelated existing row is silently skipped, never overwritten. Postgres/MySQL report `db_merge_unsupported` for the database portion but still merge uploads.

Triggered manually via `POST /api/v1/admin/system/backup` (admin) or `POST /api/v1/backup` (admin or `backup`-role API key, for CI/CD automation — see `middleware.BackupAuth()`), or automatically by `StartBackupScheduler()` per the `backup_schedule`/`backup_start_time` settings.

### Authentication
- **JWT access token**: 15 min expiry, HS256. Claims: `UserID`, `Username`, `GlobalRole`.
- **JWT refresh token**: 7 day expiry. Auto-refreshes silently on 401.
- **Browser transport**: tokens are issued as **httpOnly + SameSite=Strict cookies** (`access_token`, `refresh_token`) by `setAuthCookies` in `handlers/auth.go`. JavaScript never sees the token; the browser attaches it to every request automatically. On startup the SPA calls `authStore.initSession()` → `GET /me` to hydrate user state from the cookie.
- **Tauri transport**: there is no httpOnly cookie jar in the WebView, so tokens are kept in `sessionStorage` and attached as an `Authorization: Bearer …` header by `api/client.js`. This is the security limitation described later in this document.
- **MFA / TOTP**: optional per-user. When enabled, password and passkey login return `{mfa_required: true, mfa_token}` (a 5-minute purpose-restricted JWT from `IssueMFAToken`) and the frontend posts the 6-digit code back to complete login. TOTP secret generation/verification lives in `services/auth_service.go` (`GenerateTOTPSecret`, `VerifyTOTP`).
- **MFA trusted devices**: after a successful MFA challenge users may elect to skip MFA for 7 or 30 days (subject to the admin `mfa_remember_devices` policy: `disabled`, `week`, `week_month`). Browser clients store the trust token in an httpOnly `mfa_trust` cookie; Tauri clients receive `mfa_trust_token` in the MFA verify response and persist it in `localStorage`, sending it back via the `X-MFA-Trust` header on login. Tightening the admin policy revokes incompatible trust records in `mfa_trusted_devices`.
- **WebSocket tickets**: 30-second purpose-`"ws"` JWTs from `IssueWSTicket`, used by Tauri so the long-lived access token never appears in the WebSocket URL. Browser clients rely on the cookie and don't need a ticket.
- **Media tickets**: 5-minute purpose-`"media"` JWTs from `IssueMediaTicket`, used to grant attachment downloads to `<img>`/`<video>` elements that can't send the `Authorization` header.
- **API keys**: SHA-256 hash stored in DB. Auth via `X-API-Key` header or `?api_key=` query param. Used for the Ticket API (CI/CD automation).
- **Passwords**: hashed with bcrypt, cost factor 12 (pinned in `services/auth_service.go`).
- Middleware sets context keys consumed by handlers: `middleware.GetUserID(c)`, `middleware.GetGlobalRole(c)`, `middleware.GetUsername(c)`. `middleware.AdminOnly()` gates admin-only routes; `middleware.MetricsAuth()` and `BackupAuth()` protect the metrics and backup endpoints; `middleware.BlockCustomerRole()` is applied to project, chat, and time-tracking route groups to restrict the `customer` global role to helpdesk access only.

### Security hardening

**Startup checks** — the server refuses to start if any of these conditions is true:
- `jwt_secret` is still `"change-me-in-production"` (the default).
- `jwt_secret` is shorter than 32 characters.
- `gin_mode` is `"release"` and `allowed_origins` contains `"*"`.
- `db_driver` is `"mysql"` and `db_dsn` doesn't include `parseTime=true` (checked via `go-sql-driver/mysql`'s `ParseDSN` in `database.Init`) — without it, `DATETIME` columns fail to scan into Go's `time.Time`, so every read of a row with a timestamp column silently fails while writes keep succeeding, which otherwise surfaces as confusing "invalid credentials" / "user not found" errors that have nothing to do with the actual data.

**Response headers** (`middleware/security_headers.go`) — every response includes:
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `X-Frame-Options: SAMEORIGIN` and `frame-ancestors 'none'` in CSP
- `Content-Security-Policy` — allows `unsafe-inline` and `unsafe-eval` in `script-src` because Vite injects an inline bootstrap script and vue-i18n uses `Function()` for compiled message functions; restricts everything else to same-origin

**CORS** (`middleware/cors.go`) — wildcard origin (`*`) is blocked both at startup (release mode) and at middleware level; a request arriving with a wildcard config receives no `Access-Control-Allow-Origin` header at all.

**Rate limiting** (`middleware/ratelimit.go`) — applied to:
- Auth endpoints (login, register, password reset)
- Message-send endpoints: `POST /direct-messages/:userId` and `POST /conversations/:id/messages` (60 req/min per IP via `MessageRateLimit()`)

`AuthRateLimit()` (10 requests / 15 min per IP) is shared across `/auth/login`, `/auth/refresh`, `/auth/mfa/verify`, and `/auth/passkey/login/{begin,finish}` — a single passkey login attempt costs 2 requests (begin + finish), 3 if MFA is also required. Every successful authentication (`Login`, `MFAVerify`, `PasskeyLoginFinish`, `Refresh`) calls `middleware.ClearAuthRateLimit(c)` to wipe that IP's bucket, so a legitimate correct login is never blocked by, or left stuck behind, its own earlier failed attempts. A brute-force attacker never reaches this call since it only fires after successful authentication, so this doesn't weaken the limit's effectiveness against guessing attacks.

**Attachment access control** (`handlers/attachment.go`) — `DownloadAttachment` calls `checkAttachmentAccess` after loading the record. It walks the ownership chain (`card` → project membership, `card_comment` → card → project membership, `chat_message` → project membership, `conv_message` → conversation membership) and returns 403 if the requesting user has no access. Admins bypass all checks.

**Sensitive data in logs** — the password-reset audit log entry includes only the first 8 characters of the token followed by `...`; the full token is never written to logs.

### File uploads
MIME type is detected server-side from the first 512 bytes of the saved file (`net/http.DetectContentType`) — the client-supplied `Content-Type` header is ignored.

### System settings
Settings (SMTP, locale defaults, company branding, session timeout, …) are stored as key/value rows in `system_settings`. They are read at request time via `loadAllSettings()` so changes take effect **without a restart**. `handlers/system.go` owns all setting keys as package-level constants.

### WebSocket
- One project `Hub` per project, created on first connection and destroyed when empty (`ws/hub.go`).
- Per-user notification hubs are stored in the same map under `userID | 0x80000000` so they don't collide with project IDs. Use `ws.GetOrCreateUserHub(userID)` and the convenience helpers `ws.BroadcastToUser(userID, msg)`, `ws.IsUserOnline(userID)`, `ws.GetAllOnlineUsers()`.
- Messages are JSON `{type, payload}`. Handlers call `ws.BroadcastToProject(projectID, msg)` for board/chat/topic events.
- For horizontal scaling, set `redis_url`: broadcasts route through a Redis pub/sub channel instead of in-process memory (`ws/pubsub_redis.go`).

### Frontend state
- **Pinia stores** own all shared state; components read from stores and call store actions. Stores live in `frontend/src/stores/`: `auth`, `board`, `chat`, `customers`, `notifications`, `project`, `sidebar`, `sprint`, `system`, `topics`, `ui`.
- **`board.js`** owns columns + cards and applies WebSocket updates (`board.card.*`, `board.column.*`, checklist events). Drag-and-drop reordering is implemented in the components themselves and pushed via the API.
- **`useWebSocket.js`** establishes one connection per project view and routes messages to the appropriate store by type prefix (`board.*` → `boardStore`, `sprint.*` → `sprintStore`, `chat.*` → `chatStore`, `topic.*` → `topicsStore`, `presence.*` is handled inline).

---

## Backend conventions

### Project access control
```go
project, err := services.GetProjectBySlug(slug)   // 404 if not found
services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member")
// role levels: "viewer" < "member" < "owner" (admins bypass all checks)
```

### Project closed state
Projects have an `IsClosed` field (`is_closed` in JSON). A closed board is hidden from the sidebar (both "All Projects" and "Starred") and excluded from `GET /api/v1/projects` by default. It remains accessible via direct URL and can be reopened from **Project Settings → General → Danger Zone** or the **Admin Panel → Projects** table. Only project owners/admins and global admins can close/reopen a project.

- `ListProjects` (`handlers/project.go`) adds `WHERE is_closed = false` unless `?include_closed=true`.
- `ListStarredProjects` (`handlers/starred.go`) filters out closed projects.
- `AdminListProjects` (`handlers/admin_project.go`) supports `?closed=true` (only closed) and `?closed=hide` (only open).

### Adding a new route
1. Add handler function to the appropriate `handlers/*.go` file (or a new file).
2. Register the route in `router/router.go` under the correct group (`protected`, `admin`, `projects`, etc.).

### Adding a new model field
1. Add the field to the struct in `models/`.
2. If it needs a non-zero default, add `gorm:"default:..."` tag.
3. AutoMigrate picks it up on next startup.
4. Update the relevant handler(s) to accept / return the field.

---

## Frontend conventions

### API calls
All HTTP calls go through `src/api/client.js` (Axios with token refresh). Each domain has its own API file in `src/api/`.

### i18n
**All language files must be kept in sync.** When adding a key to `en.json`, add the same key (with a translated value or a placeholder) to all other language files (`nl`, `de`, `fr`, `es`, `da`, `sv`, `nb`, `fi`, `is`, `pt`, `it`). Keys are namespaced by feature: `board.*`, `report.*`, `admin.*`, `common.*`, etc.

### Theming
Theme is controlled by a `data-theme` attribute on `<html>`. CSS custom properties (`--color-primary`, `--color-surface`, `--color-text`, etc.) are defined in `src/styles/` for both light and dark. Never use hard-coded colour values in components.

### Component patterns
- **Modals** use `<BaseModal>` with a `#footer` slot for action buttons.
- **Toast notifications** go through `useUIStore().success(msg)` / `.error(msg)`.
- **Locked / read-only state** in `CardDetail.vue` is controlled by `locked` ref; viewer-role users see plain text instead of inputs.

---

## File uploads

Files are stored in `upload_dir` (default `./uploads`) with randomised hex names. The original filename and MIME type are recorded in the `attachments` table. Ownership is by `owner_type` + `owner_id` (`card`, `card_comment`, `chat_message`, `conv_message`, `ticket`, `ticket_message`). Images are served inline; other files are forced as downloads.

`DownloadAttachment` verifies project membership or conversation participation before serving the file (see Security hardening above). The `Content-Disposition` filename is escaped to prevent header injection.

Avatars, customer/company logos, and project/group avatars are not `attachments` rows — they're plain files referenced directly by a URL column on their owning record (`User.AvatarURL`, `Customer.LogoURL`, etc.). `backend/cmd/gc-uploads` is a standalone CLI that scans `upload_dir` for files no longer referenced by any user/customer/project/group/system-setting/attachment row and lists them (dry run by default) or deletes them with `--delete` — useful for reclaiming orphans left behind by replace-without-delete bugs in earlier releases (fixed in v0.16.4 for new replacements going forward).

---

## Helpdesk (ticketing)

The helpdesk module is gated by the `helpdesk_enabled` user flag (default `false`). Admins always have access. The feature middleware is `middleware.RequireFeature("helpdesk_enabled")`.

Users with the `customer` global role bypass the `helpdesk_enabled` flag and go straight to ticket access — but are blocked from everything else (boards, chat, time tracking) by `middleware.BlockCustomerRole()` on those route groups.

### Access control

**Customer visibility is a strict allowlist.** Non-admin users with no `CustomerAccess` rows (direct or via group) see no customers at all — there is no "see everything" fallback for unprivileged users. Global admins always see all customers.

`getAccessibleCustomerRoles(userID)` in `handlers/customer.go` builds the effective role map for a user by combining direct `CustomerAccess` rows with `GroupCustomerAccess` rows (highest role wins). This function is used by `ListCustomers`, `GetCustomer`, and `requireCustomerAccess`.

`requireCustomerAccess` in `handlers/ticket.go` calls `getAccessibleCustomerRoles` and returns `ErrForbidden` if the customer is not in the result — so both direct and group-based assignments grant ticket access. The `customer` global role passes through this check (they still need a `CustomerAccess` row) but is then restricted to read and comment operations by `requireNotCustomerRole()`, which is applied to every write handler (create/update/delete ticket, tags, links, macros, spam, checklists). Customer-role users cannot mark messages as private.

`ListCustomerMembers` (`handlers/customer_access.go`) only returns **direct** `CustomerAccess` rows — not group-based access. The ticket assignee dropdown in the UI depends on this endpoint, so every customer that uses tickets must have at least one direct `CustomerAccess` row per user who should appear as an assignee.

### SLA policies

`MatchSlaPolicy(priority)` in `handlers/sla.go` finds an active policy for the given priority (exact match or catch-all). `ComputeSlaDeadlines` computes the response and resolution deadlines from `time.Now()`. Both are called in `CreateTicket`. `refreshSlaBreachStatus` updates `SlaResponseBreached` and `SlaResolutionBreached` on every `GetTicket` / `ListTickets` call so the UI always reflects real-time breach state without a separate cron job.

SLA policies are managed by admins via `GET/POST/PUT/DELETE /api/v1/admin/sla-policies` (no feature flag — always visible to admins).

### Pending reminder

A ticket in `"pending"` status may have a `ReminderAt` timestamp. `ListTickets` orders pending tickets with a due reminder (`reminder_at <= now`) to the top. The frontend stores only a `YYYY-MM-DD` date string and normalises it to `T12:00:00Z` UTC before sending to the API.

### Pending close

A ticket in `"pending_close"` status has a `CloseAt` timestamp. `autoClosePendingTickets()` (called inline from `ListTickets`/`GetTicket`) closes any ticket whose `close_at` has passed.

### Date title prefix

When a reminder date (`pending`) or close date (`pending_close`) is set or changed, the ticket title is automatically prefixed with `[YYYY-mm-dd]`. Clearing the date removes the prefix. The logic lives in `titleWithDatePrefix()` in `TicketDetailView.vue` and fires on status change (if a date already exists) and on date update.

### Macros

Macros are reusable action sequences that agents can apply to tickets in one click. They are managed by admins and applied by any helpdesk user.

**Model** (`models/macro.go`): `Name`, `Description`, `Actions` (JSON array of `MacroAction{Type, Value}`), `IsActive`, `SortOrder`.

**Action types:**

| Type | Effect |
|---|---|
| `set_status` | Sets ticket status (`new`, `open`, `pending`, `pending_close`, `closed`) |
| `set_priority` | Sets priority (`low`, `medium`, `high`, `critical`) |
| `set_type` | Sets type (`incident`, `problem`, `service_request`, `change_request`) |
| `add_tag` | Adds a tag (idempotent via `FirstOrCreate`) |
| `add_message` | Appends a message body; supports placeholders `{email}`, `{fname}`, `{name}`, `{subject}`, `{ticket_id}`, `{agent}`, `{agent_fname}` |

**Frontend:** `components/admin/MacrosTab.vue` (admin CRUD with drag-and-drop placeholder insertion). Apply dropdown lives in `TicketDetailView.vue`.

**Apply response:** `{ticket, macro_messages}` — `macro_messages` is a list of expanded message bodies for `add_message` actions; the frontend POSTs them as ticket messages.

### Private messages

`TicketMessage` has an `IsPrivate bool` field (GORM default `false`). A private message is an **internal note** — it is:

- **Not emailed** to the ticket's original sender (`sendEmailReply` is skipped; `first_response_at` is not set).
- **Hidden from `customer`-role users** — `GetTicket` adds `WHERE is_private = false OR is_private IS NULL` to the messages preload when the requester has the `customer` global role.
- Shown in the UI with an amber highlight and a 🔒 badge.

`CreateTicketMessage` accepts `{ "body": "...", "is_private": true }`. If a `customer`-role user sends `is_private: true` it is silently reset to `false` server-side.

### Spam marking

Marking a ticket as spam closes it and hides it from the default list view (filtered by `include_spam=true` query param).

- `POST /api/v1/customers/:customerId/tickets/:ticketId/spam` — sets `is_spam=true`, `status=closed`
- `DELETE /api/v1/customers/:customerId/tickets/:ticketId/spam` — sets `is_spam=false`, `status=open`
- `POST/DELETE /api/v1/tickets/inbox/:ticketId/spam` — same for inbox tickets

**Model field:** `IsSpam bool` (default `false`). `ListTickets` excludes spam by default; pass `?include_spam=true` to include them.

### Routes

All ticket routes live under `/api/v1/customers/:customerId/tickets` and require `RequireFeature("helpdesk_enabled")`. SLA policy routes live under `/api/v1/admin/sla-policies` and require `AdminOnly`.

### Frontend

- **`components/common/DatePicker.vue`** — custom calendar picker; does **not** use a native `<input type="date">`. Respects `auth.user.date_time_format` and `auth.user.week_start`. Emits `update:modelValue` with a `YYYY-MM-DD` string (or `null` on clear).
- **`components/common/DashboardNews.vue`** — manages dismissed news IDs in `localStorage` under key `dashboard_news_dismissed_ids`.

### User setting: `dashboard_default`

Stored on `User.DashboardDefault` (GORM default `"boards"`). Values: `"boards"` | `"tickets"`. Set via `PUT /auth/me`. In `DashboardView.vue`, if `dashboard_default === "tickets"` and `helpdeskEnabled`, the component immediately redirects to the first starred customer's tickets page (or first customer, or `/customers` as fallback). The setting is exposed in **Settings → Dashboard shows**.

---

## IMAP polling — how processed mail is handled

After the IMAP poller creates a ticket from an incoming email it **moves** the message to a separate folder so it is never processed twice. The destination defaults to `"Processed"` and is configurable via `processed_mailbox` in `warmdesk.yaml`. The folder is created automatically if it does not exist.

**The poller only watches the source mailbox** (default `INBOX`, configurable via `mailbox`). Emails in the `Processed` folder are never re-scanned — this is intentional. If a user cannot find a message in their inbox, it was successfully picked up and will be in `Processed`.

If the IMAP server does not support `MOVE` (RFC 6851), the service falls back to marking the message as `\Seen` in place rather than relocating it.

Key config keys (all under `smtp:` / system settings in the admin UI):

| Key | Default | Notes |
|---|---|---|
| `mailbox` | `INBOX` | Source folder the poller reads |
| `processed_mailbox` | `Processed` | Destination after a ticket is created |

---

## IMAP OAuth2

The IMAP polling service supports OAuth2 authentication via the XOAUTH2 SASL mechanism (with OAUTHBEARER as fallback). This requires registering an OAuth 2.0 application with the email provider.

### Configuration (`warmdesk.yaml`)

Application-level client credentials are configured in the YAML file (not the database):

```yaml
oauth2:
  google_client_id: "xxxx.apps.googleusercontent.com"
  google_client_secret: "GOCSPX-xxxx"
  office_client_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  office_client_secret: "xxxx~xxxx"
```

**Google (Gmail):**
1. [Google Cloud Console](https://console.cloud.google.com/apis/credentials) → Create project → OAuth consent screen (External, scope `https://mail.google.com/`)
2. Credentials → Create OAuth 2.0 Client ID → Web application
3. Authorized Redirect URI: `https://your-domain/api/v1/admin/imap/oauth2/callback`

**Office 365 (Outlook):**
1. [Azure Portal → App registrations](https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps) → New registration
2. Accounts in any organizational directory + personal Microsoft accounts
3. Redirect URI: Web → `https://your-domain/api/v1/admin/imap/oauth2/callback`
4. API permissions → Microsoft Graph → Delegated → `IMAP.AccessAsUser.All` + `offline_access`

### UI flow (Admin → Settings → Incoming Mail)

1. Set **Authentication** to **OAuth2**
2. Select the **Email Provider** (Google or Office 365)
3. Click **Authorize** — a popup opens to the provider's consent screen
4. After granting access, the popup closes and tokens are stored in the database
5. The polling service automatically refreshes the access token via the stored refresh token before it expires

Tokens are stored in `system_settings`. Access tokens are masked in admin API responses (returned as `imap_access_token_set: "true"`). Token refresh happens automatically in `RefreshIMAPOAuth2Token()` up to 5 minutes before expiry.

---

## Migration tools (`backend/migrate/`, `cmd/importer`, `cmd/export`)

`warmdesk-import`/`warmdesk-export` move projects to/from Jira, Trello, OpenProject, or Ryver — see the admin guide's "Migration Tools" section for user-facing config. Ryver is the deepest-implemented platform; its real OData API differs from what you'd guess in several ways that cost real debugging time to discover:

- **Entity set names are lowercase** (`workrooms`, `posts`, `tasks`, `taskBoards`, `taskComments`) even though bound action names stay PascalCase (`Chat.PostMessage()`, `TaskBoard.Create()`). A team only has tasks once a task board is explicitly set up (`GET workrooms(id)/board` 404s if not); tasks then live under `taskBoards(id)/tasks`, never queried directly by workroom.
- **Columns are task board categories, not tags.** Each task's `category` nav property is the WarmDesk-column equivalent; tags are a separate, unrelated concept that now maps to WarmDesk Labels (case-insensitive dedup against existing labels, including auto-created project defaults, before creating a new one — same pattern used for reusing default columns).
- **Nested `$expand` on a collection query silently truncates** to a small page size independent of the outer collection's own (50-item) page cap — task attachments and task-comment attachments are each fetched as their own dedicated paginated query (`ryverFetchAttachments`), never via inline expand on the parent.
- **Timestamp formats vary per entity** and don't parse with one layout: task/topic dates use `2021-09-22T15:07:53+0000`; chat message `when` uses fractional seconds with a literal `Z` (`2022-06-01T08:01:27.588852Z`, parsed with `time.RFC3339Nano`). A task is "done" via a nullable `completeDate`, not an `isComplete` field (which doesn't exist).
- **Chat messages only carry a bare author user ID** (`from.id`), unlike every other entity which inlines a `createUser` object — resolve authors via one org-wide `GET /users` lookup instead of a request per message.
- **DMs are unreachable from a Custom Integration token.** `Chat.History()` on a `users(id)` entity is implicitly scoped to "the caller's own conversation with that user" — an integration has no personal DMs of its own. Only real per-user Basic Auth credentials can recover a given person's DM history, and only for threads they were part of. Not implemented.

**Author attribution on import** (`authorHeaders`/`mintTempAPIKey` in `warmdesk.go`): `user_map` resolves an external display name to a WarmDesk account, then a short-lived personal API key is minted for that user (the same admin capability behind Admin → Users → API Keys — requires `warmdesk.username` to be a global admin), used to post the comment/message/topic-reply as them, then revoked once the whole import finishes. Falls back to a `*[Name]* ...` text prefix under the importer's own account when unresolvable. `include.cards`/`include.chat` gate which categories `ImportFromRyver` fetches at all.

**Chat backfill needed new backend surface**: project chat previously had no REST create path, only the WebSocket `TypeChatSend` handler, which always stamps `time.Now()`. `POST /projects/:slug/chat/messages` (`handlers/chat.go`, `CreateChatMessage`) adds an optional `created_at` override — when set, the message is treated as a backfill and skips the WS broadcast/mention notification, since flooding online users with years-old imported history would be worse than silence. `CreateComment`/`CreateTopicReply` similarly no longer require a non-empty `body`, since a comment/message can legitimately be just an attachment (the web UI's own compose boxes still block empty submissions client-side, so this doesn't open a new footgun there).

**`card_comment` attachments were a dead end until this release**: `UploadAttachment` always accepted and stored them (`owner_type: "card_comment"` was already valid), but nothing ever called `LoadAttachments("card_comment", ...)` on read — not `GetCard`, not `ListComments`, not `CreateComment`. Fixed via a shared `attachComments()` helper in `handlers/card_comment.go`, called from all of those plus `UpdateComment`.

**Project chat's own UI was deleted as dead code** (`ChatPanel.vue`, in a prior cleanup, after an earlier release removed its slide-in panel from the board page without ever shipping the promised replacement page) — only the backend, the `chat` Pinia store, and unread-badge plumbing (`useProjectChatUnread`) survived. Revived as a "Project Chat" entry in the Topics page sidebar (`TopicsView.vue`) rather than a new dedicated route, reusing `FileUploadButton`/`AttachmentList` from the DM compose pattern. No message-edit or reaction UI yet — those only exist over WebSocket, not REST.

---

## Time reporting

All reporting lives in **`TimeTrackingView.vue`** under `/time-tracking`, which has three tabs:

- **Log Time** — weekly time sheet (requires `time_tracking_enabled`). Has two sub-views, `logTimeView` ref: `'table'` (the weekly grid) or `'calendar'` (see Calendar view below), toggled via a button next to the macro editor. Defaults to `User.TimeTrackingViewDefault` (`"table"` | `"calendar"`, settable in Settings → General, same pattern as `DashboardDefault`).
- **Report** — personal time-tracking report. Uses `/api/v1/time-entries/*` endpoints (requires `time_tracking_enabled`).
- **Board** — project-board time report extracted into `BoardReportPanel.vue`. `GET /api/v1/reports/time` returns JSON; the table is rendered in Vue. Exports call the backend PDF/XLSX endpoints (requires `can_view_reports`).

The `/reports` route redirects to `/time-tracking?tab=board-report`. `ReportView.vue` is no longer a standalone route.

### Calendar view

An alternate renderer for the Log Time tab's data — no new backend endpoints, it reuses the existing `/time-entries` CRUD. Components live in `frontend/src/components/timetracking/`:

- **`TimeTrackingCalendar.vue`** — owns the `ContextMenu` and `TimeEntryModal` instances, zoom state (`pxPerHour`, persisted in `localStorage` under `tt_calendar_zoom`), and per-customer color assignment (`assignCustomerColors`, `frontend/src/utils/calendarColors.js`).
- **`TimeTrackingCalendarWeekGrid.vue`** / **`TimeTrackingCalendarDay.vue`** / **`TimeTrackingCalendarBlock.vue`** — the week grid, one day column each, and the draggable/resizable time blocks. Pixel/time conversions are in `frontend/src/utils/calendarLayout.js` (`topOffsetPx`, `heightPx`, `pxToWallClock`).
- Drag-move and drag-resize use native Pointer Events with `setPointerCapture`, not SortableJS (SortableJS is list-reorder only; the calendar needs free 2D positioning + resize, which it has no concept of). Day-column boundaries for cross-day drag detection are measured **live** via `getColumnRects()` at drop time, not cached — the calendar panel is `display:none` when the table view is active (`v-show`), so a cached measurement taken at mount would freeze every column at a zero-size rect.
- Every drag/resize has a keyboard-operable equivalent: `TimeEntryModal.vue` (reachable via Tab + Enter on a block) exposes date/start/end time directly, mirroring `CardDetail.vue`'s "Transfer card" panel pattern (a separate explicit affordance instead of trying to make the drag gesture itself keyboard-operable).
- Click-and-drag on an empty slot opens the create modal pre-filled with the dragged range (a plain click defaults to a 1-hour span); right-click gives an edit/delete (existing entry) or create (empty slot) menu via `frontend/src/components/common/ContextMenu.vue` (generic, reusable elsewhere).
- **Click-to-create position must not be computed from the scrolled day column's own `getBoundingClientRect()`** (nor from a pointer event's `offsetY`, which is computed relative to that same rect) — on the Tauri Linux desktop build, WebKitGTK has been observed returning a stale `getBoundingClientRect().top` for a day column once `.cal-scroll` (`TimeTrackingCalendarWeekGrid.vue`) has been scrolled, one that doesn't reflect how far the view has actually scrolled. Using it (directly, or via `offsetY`) throws the click position off by roughly the scrolled distance — reproducible at any zoom level, not just when zoomed in. `TimeTrackingCalendarDay.vue`'s `pxFromEvent()` instead combines `.cal-scroll`'s own `getBoundingClientRect()` (stable — it doesn't move when its own content scrolls) with its `scrollTop` (a plain numeric DOM property, unaffected by the same bug) to compute the click offset within the day column, sidestepping the bad geometry query entirely.
- **`clientY`/`getBoundingClientRect()` can also disagree with the CSS pixel space `pxPerHour` is defined in, by a constant multiple** — observed on at least one real Tauri Linux desktop machine (not reproducible in a plain, non-HiDPI test environment, so it's suspected to be tied to the display's OS-level scale factor), present even unscrolled and at any zoom level, unlike the staleness bug above. A raw `(clientY - scrollRect.top)` is off by this ratio in every direction, so the click position drifts by a fixed *fraction* of the offset from the day's top — smaller in absolute minutes at higher zoom, larger when zoomed out, which is what makes it easy to mistake for a scroll or zoom-specific bug. `pxFromEvent()`'s `visualScale()` recovers the live ratio at click time from two `.cal-hour-label` elements (positioned by a pure `top: N*pxPerHour` inline style with no geometry query in their own layout, so they aren't subject to either bug) and divides it out of the `(clientY - scrollRect.top)` term — rather than assuming that ratio is always `1`, which would only hold on displays that don't trigger it.
- `.cal-grid-body` (`TimeTrackingCalendarWeekGrid.vue`) has a `GRID_TOP_INSET_PX` (`calendarLayout.js`, currently `4`) top padding so the `00:00` label/gridline stands visually clear of the border separating the day-header row from the scrollable grid, instead of sitting right on it. `pxFromEvent()` subtracts this same constant back out — it's a deliberate, known offset (unlike the two bugs above), so it's safe to hardcode identically on both sides rather than measuring it live.

### Customer colors

`Customer.Color` (`backend/models/customer.go`, `gorm:"size:7"`, same convention as `Project.Color`) is optional and editable from three places that must be kept in sync: `CustomersView.vue`/`CustomerDetailView.vue` (regular customers), `AdminView.vue`'s Customers tab and its separate Time Tracking tab (both regular and time-tracking-only customers), and `TimeTrackingView.vue`'s manage-projects modal (time-tracking-only customers only). `assignCustomerColors()` (`frontend/src/utils/calendarColors.js`) resolves the calendar's per-block color: a customer's own `Color` wins; customers without one get the first palette color not already used by another customer (explicit or auto-assigned), falling back to a hash once the 10-color palette is exhausted.

### Customer locations and travel distance auto-fill

`CustomerLocation` (`backend/models/customer_location.go`) is a customer-scoped, one-to-many child model (address line 1/2, city, postal code, region, country, phone, contact name/email/phone, `TravelDistance *float64`) with full CRUD (`backend/handlers/customer_location.go`, routes under `/customers/:customerId/locations`, same `requireCustomerAccess`/`canManageCustomer` pattern as `CustomerContact`). Managed from the customer detail view's **Locations** tab.

`TimeEntry.Distance` predates this feature (see Contract.PricePerKm below) but had no default-value mechanism — `CustomerLocation.TravelDistance` fills that gap purely on the frontend, with no new backend field or FK linking an entry to a location: picking a location just copies its `travel_distance` into the distance input, which stays a normal editable value from then on.

- **`TimeTrackingView.vue`** (weekly grid): `locationsByCustomer`/`loadLocationsForCustomer` (lazy-loaded per customer, mirrors `contractsByCustomer`) backs a **Location** `<select>` in the add-row form, the edit-row form, and the per-cell distance popup — shown only when the row's customer has a location with `travel_distance` set. `rowLocationDistance` (keyed by `row.key`) remembers the last-picked distance for a row so newly created day-cells (`onCellBlur`'s create branch) get it as their default `distance`, without needing to reopen the popup each time.
- Re-picking a row's location also retroactively fixes entries: `confirmEditRow` (rename/reassign) and the popup's `applyDistPopupLocation` both push the newly picked distance to every *other* entry this week on that row that **already has a non-null `distance`** — days deliberately left blank are never touched.
- **`TimeEntryModal.vue`** (calendar view) has the same **Location** select tied to `form.customer_id`, unconditionally overwriting `form.distance` on pick (no "already has a value" guard — there's no sibling-entry concept in a single-entry modal).
- **Macro editor** (`TimeTrackingView.vue`'s `macroEditorOpen` modal, see Time macros below) has the same pattern per row: `day1_location_id` (single-day mode) or `day1_location_id`/`day2_location_id` (alternating A/B mode) back a **Location** `<select>` next to each distance input, populated the same way via `macroLocationsForRow`/`loadLocationsForCustomer`. `onMacroLocationChange` copies the picked location's `travel_distance` into the matching `day1_distance`/`day2_distance` field, which then applies like any other macro-defined distance when the macro is run (`applyMacroTemplate` reads `day1_distance`/`day2_distance` directly — it has no separate awareness of the location, only of the distance value that picking one filled in). The location id itself is saved as part of the macro template (client-side only — `backend/models/time_entry.go`'s `TimeMacroLibrary.Payload` is an opaque JSON blob with no server-side schema), so a saved macro remembers its per-row location across sessions, not just for the current run.
- `backend/handlers/time_entry.go`'s `UpdateTimeEntry`/`CreateTimeEntry` treat an **omitted** `distance` (or `contract_id`/`start_time`/`end_time`) key as an explicit *clear to NULL*, not "leave unchanged" — `ShouldBindJSON` can't distinguish the two for a `*float64`. Every write path in `TimeTrackingView.vue` that updates an existing entry (`onCellBlur`, `confirmEditRow`, `applyDistPopup`/`clearDistPopup`) must therefore always resend the entry's current value for these fields, not just the ones it's actually changing, or it silently wipes them.

### Undeclarable time

Time-tracking-only projects (`Project.TimeTrackingOnly = true`) carry an `UndeclarableMinutes int` field (GORM default 0). This represents time that cannot be billed — e.g. a "Travel" project with 45 undeclarable minutes means every entry logs the full time but only `max(0, entry.minutes − 45)` counts as declarable.

- **Per-entry declarable** = `max(0, entry.minutes − project.undeclarable_minutes)` (capped so it never goes negative).
- **Per-group totals** — `TotalMinutes`, `UndeclarableMinutes`, `DeclarableMinutes` are all carried in `timeEntryGroup` (handlers/time_entry.go) and in `TimeEntryReportResponse`.
- **On-screen report** — group headers and entry rows show declarable time; when `group_by=customer` an undeclarable line appears below each customer subtotal and below the grand total.
- **PDF** (`handlers/time_entry_pdf.go`) — same display logic; when `?page_break=customer` each customer gets its own page with the full document header (`drawDocHeader` closure) repeated via `SetHeaderFunc` registered *after* the first page so it fires only on explicit `AddPage()` calls; grand total is omitted in page-per-customer mode.
- **XLSX** (`handlers/time_entry_xlsx.go`) — undeclarable and declarable rows appended after the grand total when `UndeclarableMinutes > 0`.
- **`pdfI18n`** (`handlers/report_pdf.go`) — `Undeclarable` and `Declarable` string fields, populated for all 12 languages.

All PDF and XLSX files are **generated server-side** (Go/excelize for XLSX, Go/gofpdf for PDF) and downloaded as binary. The frontend never builds documents itself.

Binary downloads go through `fetchBinary` (`src/api/client.js`):
- **Browser**: Axios `GET` with `responseType: 'arraybuffer'` — returns `ArrayBuffer`.
- **Tauri**: invokes the `fetch_binary_b64` Rust command (`src-tauri/src/lib.rs`) which fetches via reqwest and returns base64. JavaScript decodes with `atob()`. This bypasses WebKit GTK2's broken `ReadableStream`-backed `Response` body methods (`arrayBuffer()`, `text()`, etc. all throw `TypeError` on Linux).

Saving in Tauri uses `@tauri-apps/plugin-dialog` `save()` + `@tauri-apps/plugin-fs` `writeFile()`. The last-used export directory is persisted in `localStorage` under `warmdesk_last_export_dir`; falls back to `homeDir()` on first use.

---

## Update banner

`useUpdateCheck.js` fetches `https://api.github.com/repos/tonk/warmdesk/releases/latest` on login and every hour thereafter. If the latest tag is semver-greater than `__APP_VERSION__` (set at Vite build time via `git describe`), an `UpdateBanner` appears at the top of the page with a "Download" button that picks the right binary by platform + architecture (`navigator.platform`). The download link only appears in the Tauri desktop app (`window.__TAURI_INTERNALS__`); the web interface shows "View release notes" only.

The CSP in `backend/middleware/security_headers.go` must list `https://api.github.com` in `connect-src` for this to work.

### Testing locally

Inject a fake cache entry in DevTools console, then reload:

```js
sessionStorage.removeItem('update_dismissed')
sessionStorage.setItem('update_check', JSON.stringify({
  tag: 'v99.99.99',
  url: 'https://github.com/tonk/warmdesk/releases/tag/v99.99.99',
  assets: [
    {name: 'WarmDesk-v99.99.99-x86_64.AppImage',  browser_download_url: 'https://github.com/tonk/warmdesk/releases/download/v99.99.99/WarmDesk-v99.99.99-x86_64.AppImage'},
    {name: 'WarmDesk-v99.99.99-universal.dmg',     browser_download_url: 'https://github.com/tonk/warmdesk/releases/download/v99.99.99/WarmDesk-v99.99.99-universal.dmg'},
    {name: 'WarmDesk-v99.99.99-x64-portable.zip',  browser_download_url: 'https://github.com/tonk/warmdesk/releases/download/v99.99.99/WarmDesk-v99.99.99-x64-portable.zip'},
    {name: 'warmdesk-v99.99.99-linux-amd64.tar.gz', browser_download_url: 'https://github.com/tonk/warmdesk/releases/download/v99.99.99/warmdesk-v99.99.99-linux-amd64.tar.gz'}
  ],
  expires: Date.now() + 3600000
}))
location.reload()
```

Both the `update_dismissed` key (which hides the banner) and the `update_check` key (which caches the GitHub API response) live in `sessionStorage` and reset on tab close.

---

## Versioning

As of **v0.13.0**, WarmDesk follows real semver discipline within the `0.MINOR.PATCH` range:

- **PATCH** — bug fixes only. No new features, no additive config/API surface.
- **MINOR** — new backwards-compatible features (new endpoint, new config key with a sane default, new UI capability). Resets PATCH to 0.
- **MAJOR** — reserved for the eventual `1.0.0` stability milestone, or a genuine breaking change before then (removed/renamed config key or API field, a DB change that isn't safely reversible, a changed response shape).

When preparing a release, classify every commit since the last tag as fix / feat / breaking first, then pick the version bump accordingly — don't just increment the last number. See `.claude/commands/release.md` for the full release checklist.

Releases through **v0.12.42** used a flat incrementing counter on the last digit regardless of change type — that range is not semver-accurate, so don't infer change-type history from old tags.

---

## Tests

### Running

```bash
make test             # backend + frontend
make test-backend     # go test ./...
make test-frontend    # vitest run

# Backend — single package or test
cd backend
go test ./...                    # all packages
go test ./services/ -v -run TestMidPosition   # single test by name

# Frontend
cd frontend
npm test              # vitest run
npm run test:watch    # vitest watch mode
```

Test dependencies: `github.com/stretchr/testify` (Go assertions), `vitest` / `@vue/test-utils` / `jsdom` (frontend).

The `testutil` package (`backend/testutil/db.go`) provides `SetupTestDB()` which returns an in-memory SQLite database with all models migrated — use it in tests that need a real database.

Config lives in `frontend/vitest.config.ts`. Component tests use `@vue/test-utils` `mount()`/`shallowMount()`, store tests use `@pinia/testing`. Pure functions are exported for direct testing.

### E2E screenshots (Playwright)

Playwright-driven screenshot specs live in `frontend/e2e/`. They produce the 24 reference PNGs under `screenshots/` used for documentation/README.

```bash
# Full run — seed DB, start servers, capture screenshots, clean up
make screenshots
# or: cd frontend && npm run screenshots

# Run against already-running servers (faster for iteration)
cd frontend && npm run screenshots:dev
```

**Prerequisites:** Go, Node.js, Chrome/Chromium (installed automatically by Playwright). The first full run will install the Playwright browser binary via `npx playwright install chromium`.

The spec (`frontend/e2e/screenshots.spec.js`) logs in via the API (cookie session) as `demo.admin` / `demo1234` and captures each view. Auth state is saved/reused via Playwright `storageState`. Screenshot 24 logs in as `tonk` and captures the undeclarable-time grid alignment.

`screenshots.sh` runs `go run ./cmd/seed --reset` so time-entry dates stay relative to the current week.

**Screenshots 11–12 (chat reactions)** require interaction in the embedded chat panel and are not yet automated — capture manually or run with `DEBUG=pw:api` to verify selectors.

**Adding a new screenshot:** add a new test in `frontend/e2e/screenshots.spec.js` following the existing pattern (create context from `AUTH_FILE`, navigate, wait, screenshot). To update all reference PNGs, run `make screenshots`. For individual updates, point your test at the specific view and use Playwright's `--update-snapshots` if using `toHaveScreenshot` assertions.

---

## Deployment notes

- `deploy/` has ready-made templates for systemd, nginx (with SSL), and Apache. The nginx/Apache configs set `X-Forwarded-Proto` so httpOnly auth and MFA-trust cookies get the `Secure` flag when TLS terminates at the proxy.
- For multi-instance deployments set `redis_url` — this routes WebSocket broadcasts through Redis pub/sub instead of in-process memory.
- First-run with an empty DB: register the first user normally—they receive `admin` automatically; use a direct DB update only to recover if every admin is lost.

### Tauri desktop — known security limitation

This applies **only to the Tauri desktop build**. Browser clients use httpOnly cookies (see Authentication above) and are not affected.

In the desktop app (Tauri), JWT tokens are stored in the WebView's `sessionStorage` because the WebView has no httpOnly cookie jar shared with the Go backend. `sessionStorage` is readable by any JavaScript running in the WebView, so a malicious npm dependency or an XSS in the app itself could exfiltrate the token.

Mitigations already in place:
- CSP (`Content-Security-Policy` in the proxy templates) blocks external script sources.
- `withGlobalTauri: false` in `tauri.conf.json` avoids polluting the global namespace.
- WebSocket connections use a 30-second `ws` ticket (`IssueWSTicket`) instead of the access token in the URL.

**Proper fix (not yet implemented):** move token storage into Rust `tauri::State`, intercept all API requests at the Rust/reqwest layer to inject the `Authorization` header, and never expose the raw token to JavaScript. This requires replacing the Axios-based `api/client.js` Tauri path with `invoke('api_request', …)` calls routed through a custom Rust command — a significant refactor of the HTTP client layer.

### Tauri desktop — WebKit binary download workaround

On Linux, WebKit GTK2 constructs `Response` objects with a `ReadableStream` body (via tauri-plugin-http / reqwest). All body-reading methods on such a `Response` (`arrayBuffer()`, `text()`, `blob()`, `body.getReader()`) throw `TypeError("Type error")` — this affects Axios's fetch adapter regardless of `responseType`.

**Workaround:** `fetchBinary` in `src/api/client.js` detects Tauri and calls `invoke('fetch_binary_b64', { url, headers })`. The `fetch_binary_b64` command in `src-tauri/src/lib.rs` performs the HTTP request using reqwest directly (no WebKit involved), encodes the response bytes as base64, and returns the string to JS. JavaScript decodes with `atob()` — no WebKit HTTP API needed at any point.

---

### Database TLS — strongly advised for remote databases

`db_tls_mode` defaults to `"disable"`. This is acceptable only when the database runs on the same host (Unix socket or loopback). For any remote PostgreSQL or MySQL instance, **set `db_tls_mode: "verify-full"`** (plus `db_tls_ca_cert`) so the connection is encrypted and the server certificate is verified. Without TLS, credentials and query data travel in plaintext.

---

### Horizontal scaling limitations

When running more than one WarmDesk instance behind a load balancer, two subsystems are **in-memory only** and do not share state across instances:

| Subsystem | Limitation | Fix |
|---|---|---|
| **WebSocket broadcasts** | Board/chat updates only reach clients connected to the same instance | Set `redis_url` — broadcasts are routed through Redis pub/sub |
| **Rate limiter** (`middleware/ratelimit.go`) | Login/register/reset attempt counts are per-instance; an attacker can hit multiple instances to bypass the limit | Set `redis_url` — currently no Redis-backed rate limiter is implemented; this is a known gap for multi-instance deployments |

---

## Accessibility

**All frontend changes must be WCAG 2.1 AA compliant.** Check every new or modified UI element against the following before considering it done:

- **Icon-only buttons** — always add `aria-label` (not just `title`)
- **Modals / dialogs** — `role="dialog"`, `aria-modal="true"`, `aria-labelledby` pointing to a heading element (`<h2>`/`<h3>`); Escape closes the dialog; focus moves to the first focusable element on open
- **Inputs** — always paired with a `<label>` (use a `.sr-only` visually-hidden label when space is tight); `placeholder` alone is not a label
- **Color inputs** — need both a `<label>` and an `aria-label`
- **Custom tabs** — `role="tablist"` on the container, `role="tab"` + `aria-selected` + `aria-controls` on each tab, `role="tabpanel"` + `aria-labelledby` on each panel
- **Hover-only visibility** — elements revealed only via `opacity`/`visibility` on a parent hover are invisible to keyboard users; keep the hover CSS but also add a `:focus-visible` or `:focus-within` rule that sets the same visible state (e.g. `.fav-btn:focus-visible { opacity: 1; }`, `.msg-hover-actions:focus-within { opacity: 1; }`)
- **Decorative elements** — add `aria-hidden="true"` to purely visual icons and decorative spans
- **Every interactive element** must have a programmatic accessible name: visible text, `aria-label`, or `aria-labelledby`

Key success criteria: 1.3.1 Info and Relationships, 2.1.1 Keyboard, 2.4.3 Focus Order, 3.3.2 Labels or Instructions, 4.1.2 Name / Role / Value.
