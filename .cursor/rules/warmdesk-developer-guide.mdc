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
                 patchelf squashfs-tools
```

### Targets

```bash
make build          # builds frontend + backend → dist/
make run            # build then run
make appimage       # Linux AppImage (requires Rust + system libs above)
make dmg            # macOS universal DMG
make windows-installer  # Windows NSIS installer
```

Production:
```bash
cd dist
WEB_DIR=./web ./warmdesk
```

---

## Repository layout

```
backend/
  main.go            # entry point — config, DB, services, router
  cmd/               # auxiliary binaries (e.g. seed)
  config/            # Config struct + env var / YAML loading
  database/          # GORM init, AutoMigrate for all models
  docs/              # Swagger-generated API docs (do not hand-edit)
  handlers/          # One file per feature area (card.go, report.go, …)
  middleware/        # auth, api_key, cors, ratelimit, ip_allowlist, security_headers
  migrate/           # one-off data migration helpers (separate from AutoMigrate)
  models/            # GORM model structs (board.go, user.go, project.go, …)
  router/            # Single router.go — all routes in one place
  services/          # Business logic (auth, email, project helpers, ordering, git)
  staticweb/         # Embeds the built frontend (dist/web) into the Go binary
  ws/                # WebSocket hub + client + pub/sub (memory & Redis)

frontend/
  src/
    api/             # Axios wrappers, one file per domain (projects.js, reports.js, …)
    components/      # Reusable Vue components (board/, call/, chat/, common/, layout/)
    composables/     # useTheme, useWebSocket, useDateFormat, useAvatar, useLiveKitGroupCall, …
    i18n/            # en.json + nl, de, fr, es, da, sv, nb, fi, is, pt, it — all keys must be mirrored
    router/          # index.js — all routes + auth guards
    stores/          # Pinia stores (auth, board, chat, project, sprint, topics, ui, …)
    styles/          # Global CSS custom properties (light/dark theme vars)
    utils/           # Small framework-agnostic helpers
    views/           # Page-level Vue components

frontend/src-tauri/  # Rust/Tauri config (minimal — mostly tauri.conf.json)
deploy/              # systemd / nginx / apache templates
```

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
| `smtp.host` / `.port` / `.from` / `.username` / `.password` / `.use_tls` | — | port `587`, rest empty (overridable via system settings) |
| `livekit_url` / `livekit_api_key` / `livekit_api_secret` / `livekit_room_prefix` | `LIVEKIT_*` | *(empty — required for voice/video calls)* |
| `oauth2.google_client_id` / `.google_client_secret` / `.office_client_id` / `.office_client_secret` | — | *(empty — required for IMAP OAuth2 auth)* |

---

## Architecture decisions

### Database
- **GORM AutoMigrate runs on every startup** — no separate migration files. Adding a new field to a model struct is all that is needed; the column appears on next boot.
- Supported drivers: `sqlite`, `postgres`, `mysql`. The driver string goes in `db_driver`, the DSN in `db_dsn`.
- Card numbering (`PRJ-1`, `PRJ-2`, …) is maintained by an atomic `card_counter` increment on the `projects` table.

### Authentication
- **JWT access token**: 15 min expiry, HS256. Claims: `UserID`, `Username`, `GlobalRole`.
- **JWT refresh token**: 7 day expiry. Auto-refreshes silently on 401.
- **Browser transport**: tokens are issued as **httpOnly + SameSite=Strict cookies** (`access_token`, `refresh_token`) by `setAuthCookies` in `handlers/auth.go`. JavaScript never sees the token; the browser attaches it to every request automatically. On startup the SPA calls `authStore.initSession()` → `GET /me` to hydrate user state from the cookie.
- **Tauri transport**: there is no httpOnly cookie jar in the WebView, so tokens are kept in `sessionStorage` and attached as an `Authorization: Bearer …` header by `api/client.js`. This is the security limitation described later in this document.
- **MFA / TOTP**: optional per-user. When enabled, login returns `{mfa_required: true, mfa_token}` (a 5-minute purpose-restricted JWT from `IssueMFAToken`) and the frontend posts the 6-digit code back to complete login. TOTP secret generation/verification lives in `services/auth_service.go` (`GenerateTOTPSecret`, `VerifyTOTP`).
- **WebSocket tickets**: 30-second purpose-`"ws"` JWTs from `IssueWSTicket`, used by Tauri so the long-lived access token never appears in the WebSocket URL. Browser clients rely on the cookie and don't need a ticket.
- **Media tickets**: 5-minute purpose-`"media"` JWTs from `IssueMediaTicket`, used to grant attachment downloads to `<img>`/`<video>` elements that can't send the `Authorization` header.
- **API keys**: SHA-256 hash stored in DB. Auth via `X-API-Key` header or `?api_key=` query param. Used for the Ticket API (CI/CD automation).
- **Passwords**: hashed with bcrypt, cost factor 12 (pinned in `services/auth_service.go`).
- Middleware sets context keys consumed by handlers: `middleware.GetUserID(c)`, `middleware.GetGlobalRole(c)`, `middleware.GetUsername(c)`. `middleware.AdminOnly()` gates admin-only routes; `middleware.MetricsAuth()` and `BackupAuth()` protect the metrics and backup endpoints.

### Security hardening

**Startup checks** — the server refuses to start if any of these conditions is true:
- `jwt_secret` is still `"change-me-in-production"` (the default).
- `jwt_secret` is shorter than 32 characters.
- `gin_mode` is `"release"` and `allowed_origins` contains `"*"`.

**Response headers** (`middleware/security_headers.go`) — every response includes:
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `X-Frame-Options: SAMEORIGIN` and `frame-ancestors 'none'` in CSP
- `Content-Security-Policy` — allows `unsafe-inline` and `unsafe-eval` in `script-src` because Vite injects an inline bootstrap script and vue-i18n uses `Function()` for compiled message functions; restricts everything else to same-origin

**CORS** (`middleware/cors.go`) — wildcard origin (`*`) is blocked both at startup (release mode) and at middleware level; a request arriving with a wildcard config receives no `Access-Control-Allow-Origin` header at all.

**Rate limiting** (`middleware/ratelimit.go`) — applied to:
- Auth endpoints (login, register, password reset)
- Message-send endpoints: `POST /direct-messages/:userId` and `POST /conversations/:id/messages` (60 req/min per IP via `MessageRateLimit()`)

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

### Handlers
Every handler follows the same pattern:
```go
func DoThing(c *gin.Context) {
    userID := middleware.GetUserID(c)
    // 1. Parse & validate path params
    // 2. Load project, check membership with services.RequireProjectRole(...)
    // 3. Bind JSON body
    // 4. DB operation
    // 5. Broadcast WS event if needed
    // 6. Return JSON
}
```

Error responses always use `gin.H{"error": "..."}`:
```go
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid column id"})
```

### Project access control
```go
project, err := services.GetProjectBySlug(slug)   // 404 if not found
services.RequireProjectRole(project.ID, userID, middleware.GetGlobalRole(c), "member")
// role levels: "viewer" < "member" < "owner" (admins bypass all checks)
```

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
All HTTP calls go through `src/api/client.js` (Axios with token refresh). Each domain has its own API file:
```js
// src/api/projects.js
export const projectsApi = {
  updateCard: (slug, id, data) => client.put(`/projects/${slug}/cards/${id}`, data),
}
```

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

---

## Helpdesk (ticketing)

The helpdesk module is gated by the `helpdesk_enabled` user flag (default `false`). Admins always have access. The feature middleware is `middleware.RequireFeature("helpdesk_enabled")`.

### Models (`models/ticket.go`, `models/sla.go`)

| Model | Key fields |
|---|---|
| `Ticket` | `CustomerID`, `Title`, `Description`, `Type` (incident/problem/service_request/change_request), `Priority` (low/medium/high/critical), `Status` (new/open/pending/pending_close/closed), `AssignedToID`, `OwnerID`, `GroupID`, `ReminderAt *time.Time`, `CloseAt *time.Time`, `IsSpam bool`, `SlaPolicyID`, `SlaResponseDeadline`, `SlaResolutionDeadline`, `SlaResponseBreached bool`, `SlaResolutionBreached bool`, `FirstResponseAt *time.Time` |
| `TicketMessage` | `TicketID`, `UserID`, `Body`, `EmailSent bool` — internal messages on a ticket; `EmailSent` is `true` when the message triggered an outbound email reply |
| `TicketTag` | `TicketID`, `Name` — free-form tags |
| `TicketLink` | `TicketID`, `LinkedTicketID`, `LinkType` — links between tickets |
| `TicketCardLink` | `TicketID`, `CardID` — links from a ticket to a board card |
| `SlaPolicy` | `Name`, `ResponseTimeMinutes`, `ResolutionTimeMinutes`, `PriorityFilter` (comma-separated), `IsActive` |

### Access control

**Customer visibility is a strict allowlist.** Non-admin users with no `CustomerAccess` rows (direct or via group) see no customers at all — there is no "see everything" fallback for unprivileged users. Global admins always see all customers.

`getAccessibleCustomerRoles(userID)` in `handlers/customer.go` builds the effective role map for a user by combining direct `CustomerAccess` rows with `GroupCustomerAccess` rows (highest role wins). This function is used by `ListCustomers`, `GetCustomer`, and `requireCustomerAccess`.

`requireCustomerAccess` in `handlers/ticket.go` calls `getAccessibleCustomerRoles` and returns `ErrForbidden` if the customer is not in the result — so both direct and group-based assignments grant ticket access.

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

**Routes:**
- `GET/POST /api/v1/admin/macros` — list all / create (admin only)
- `PUT/DELETE /api/v1/admin/macros/:id` — update / delete (admin only)
- `GET /api/v1/macros` — list active macros (all authenticated helpdesk users)
- `POST /api/v1/customers/:customerId/tickets/:ticketId/macros/:macroId` — apply to ticket
- `POST /api/v1/tickets/inbox/:ticketId/macros/:macroId` — apply to inbox ticket

**Frontend:** `components/admin/MacrosTab.vue` (admin CRUD with drag-and-drop placeholder insertion). Apply dropdown lives in `TicketDetailView.vue`.

**Apply response:** `{ticket, macro_messages}` — `macro_messages` is a list of expanded message bodies for `add_message` actions; the frontend POSTs them as ticket messages.

### Spam marking

Marking a ticket as spam closes it and hides it from the default list view (filtered by `include_spam=true` query param).

- `POST /api/v1/customers/:customerId/tickets/:ticketId/spam` — sets `is_spam=true`, `status=closed`
- `DELETE /api/v1/customers/:customerId/tickets/:ticketId/spam` — sets `is_spam=false`, `status=open`
- `POST/DELETE /api/v1/tickets/inbox/:ticketId/spam` — same for inbox tickets

**Model field:** `IsSpam bool` (default `false`). `ListTickets` excludes spam by default; pass `?include_spam=true` to include them.

### Routes

All ticket routes live under `/api/v1/customers/:customerId/tickets` and require `RequireFeature("helpdesk_enabled")`. SLA policy routes live under `/api/v1/admin/sla-policies` and require `AdminOnly`.

### Frontend

- **`TicketListView.vue`** — lists tickets for one customer; shows `DashboardNews` at the top so news is visible to helpdesk-first users.
- **`TicketDetailView.vue`** — full ticket detail with inline title editing, status/priority/type/assignee/group dropdowns, tag management, SLA card, linked tickets/cards panel, internal messages with attachments, the pending reminder `DatePicker`, macro apply dropdown, and spam mark/unmark controls. The original email body is rendered as plain text with `white-space: pre-wrap` (selectable via `.selectable` class to override `body { user-select: none }`). Messages from agents that triggered an email reply show an ✉ badge (`email_sent` field). Clicking the ticket number (`#123`) copies `Ticket#123` to the clipboard.
- **`components/admin/MacrosTab.vue`** — admin UI for macro CRUD with drag-and-drop placeholder insertion buttons.
- **`components/admin/SlaPoliciesTab.vue`** — admin UI for CRUD on SLA policies.
- **`components/common/DatePicker.vue`** — custom calendar picker that uses CSS custom properties for theming and respects `auth.user.date_time_format` and `auth.user.week_start`. Emits `update:modelValue` with a `YYYY-MM-DD` string (or `null` on clear). Does **not** use a native `<input type="date">` at all.
- **`components/common/DashboardNews.vue`** — self-contained news widget that fetches active news, manages dismissed IDs in `localStorage` (key `dashboard_news_dismissed_ids`), and renders the widget grid. Used in `DashboardView`, `CustomersView`, and `TicketListView`.
- **`stores/tickets.js`** — Pinia store for ticket list state.
- **`api/tickets.js`** — Axios wrappers for all ticket endpoints.
- **`api/macros.js`** — Axios wrappers for macro CRUD and apply endpoints.
- **`api/sla.js`** — Axios wrappers for SLA policy admin endpoints.

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

### Backend implementation

| File | Role |
|---|---|
| `handlers/oauth2.go` | OAuth2 auth URL generation, callback (code → token exchange), status, disconnect, `RefreshIMAPOAuth2Token()` |
| `services/xoauth2.go` | Custom XOAUTH2 SASL client (`user=...\x01auth=Bearer ...\x01\x01`) |
| `services/imap_service.go` | `connectAndLogin` calls `c.Authenticate()` with XOAUTH2 (then OAUTHBEARER fallback) when `cfg.AuthMechanism == "oauth2"`; `poll` calls token refresher before connecting |
| `config/config.go` | `IMAPConfig` auth fields + `OAuth2Config` struct for client credentials |
| `handlers/system.go` | `settingIMAP*` keys for auth mechanism, provider, access token, refresh token, expiry |

Tokens are stored in `system_settings` (same table as passwords). Access tokens are masked in admin API responses (returned as `imap_access_token_set: "true"`). Token refresh happens automatically in `RefreshIMAPOAuth2Token()` up to 5 minutes before expiry.

### Routes

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/v1/admin/imap/oauth2/auth-url?provider=google\|office365` | Returns the OAuth2 authorization URL |
| `GET` | `/api/v1/admin/imap/oauth2/callback` | Handles the provider callback (code → token exchange) |
| `GET` | `/api/v1/admin/imap/oauth2/status` | Returns connection status |
| `POST` | `/api/v1/admin/imap/oauth2/disconnect` | Clears stored tokens, reverts to plain auth |

---

## Time reporting

All reporting lives in **`TimeTrackingView.vue`** under `/time-tracking`, which has three tabs:

- **Log Time** — weekly time sheet (requires `time_tracking_enabled`).
- **Report** — personal time-tracking report. Uses `/api/v1/time-entries/*` endpoints (requires `time_tracking_enabled`).
- **Board** — project-board time report extracted into `BoardReportPanel.vue`. `GET /api/v1/reports/time` returns JSON; the table is rendered in Vue. Exports call the backend PDF/XLSX endpoints (requires `can_view_reports`).

The `/reports` route redirects to `/time-tracking?tab=board-report`. `ReportView.vue` is no longer a standalone route.

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

### Backend — test index

| Package | File | What it tests | Tests |
|---------|------|---------------|-------|
| `services` | `auth_service_test.go` | JWT issuance/validation (access, refresh, MFA, WS, media, passkey), bcrypt hash/verify, TOTP generate/verify, expired/invalid signature | 13 |
| `services` | `email_service_test.go` | `envelopeAddress` (plain, display-name, invalid), `ExtractMentions` (none, single, multiple, dupes, underscore) | 8 |
| `services` | `email_template_test.go` | `WrapHTML`/`WrapText` with fallback, custom branding, version-strip, empty company, instance URL | 8 |
| `services` | `order_service_test.go` | `MidPosition`, `PositionAfter` — ordering slot calculation with various neighbours | 10 |
| `services` | `project_service_test.go` | `keyPrefixBase` (11 cases), `slugify` (11 cases), `itoa` (8 cases) | 30 subtests |
| `middleware` | `cors_test.go` | `ParseOrigins` — empty, single, multiple, trimming, skips empty, Tauri origins always included | 6 |
| `ws` | `ws_test.go` | `memoryPubSub` (IsLocal, Publish, Subscribe/cancel), `InitPubSub`, `StartPubSubListener`, all ~40 message type constants, struct constructors | 4 funcs |
| `config` | `config_test.go` | `defaults()` — all ~25 default field values asserted | 1 |
| `cmd/training` | `main_test.go` | `getName()` with index/rollover, `projectAvatarURL`, `groupAvatarURL`, constants | 6 |

The `testutil` package (`backend/testutil/db.go`) provides `SetupTestDB()` which returns an in-memory SQLite database with all models migrated — use it in tests that need a real database.

### Frontend — test index

| File | What it tests | Tests |
|------|---------------|-------|
| `composables/useUpdateCheck.test.js` | `isNewer` (semver comparison, leading v, dev), `pickAsset` (platform/arch matching, Tauri detection) | 11 |
| `composables/useDateFormat.test.js` | `pad`, `applyFormat` (Date, ISO string, 12h, midnight, invalid), `dateOnlyFmt` | 10 |
| `utils/shiftTimeEntries.test.js` | `parseWallClock`, `fmtWallClock`, `wallClockSpanMinutes`, `addDaysISO`, `splitShiftIntoDayEntries`, `weekendStandbyDefaults` | 13 |
| `utils/contractSlotPreview.test.js` | `parseSlotHHMM`, `slotDayTypeMatches`, `slotCoverageOnWeekday`, `slotPreviewReady`, `buildSlotPreviewDays`, `formatSlotPreviewTime` | 18 |
| `utils/emoticons.test.js` | EMOTICONS sort order, QUICK_REACTION_EMOJIS, EMOJI_SHORTCODES underscore aliases, `detectEmoticon`, `detectEmojiShortcode` | 9 |

**Total: 141+ tests** (61 backend + 5 test files / 80 frontend across 5 test files)

Config lives in `frontend/vitest.config.ts`. Component tests use `@vue/test-utils` `mount()`/`shallowMount()`, store tests use `@pinia/testing`. Pure functions are exported for direct testing (e.g. `isNewer`, `pickAsset` from `useUpdateCheck.js`, `pad`/`applyFormat`/`dateOnlyFmt` from `useDateFormat.js`).

### E2E screenshots (Playwright)

Playwright-driven screenshot specs live in `frontend/e2e/`. They produce the 20 reference PNGs under `screenshots/` used for documentation/README.

```bash
# Full run — seed DB, start servers, capture screenshots, clean up
make screenshots
# or: cd frontend && npm run screenshots

# Run against already-running servers (faster for iteration)
cd frontend && npm run screenshots:dev
```

**Prerequisites:** Go, Node.js, Chrome/Chromium (installed automatically by Playwright). The first full run will install the Playwright browser binary via `npx playwright install chromium`.

The spec (`frontend/e2e/screenshots.spec.js`) logs in as `demo.admin` / `demo1234` and captures each view. Auth state is saved/reused via Playwright `storageState`.

| Screenshot | Route / action |
|---|---|
| `01-login.png` | `/login` page |
| `02-dashboard.png` | Dashboard after login |
| `03-board.png` | `/projects/website-redesign` |
| `04-card-detail.png` | Click first card on board |
| `05-topics.png` | `/projects/website-redesign/topics`, open first topic |
| `06-messages.png` | `/chats` (direct messages) |
| `07-report.png` | `/time-tracking` → Board tab |
| `08-admin-users.png` | `/admin` (users tab) |
| `09-admin-settings.png` | `/admin` → Settings tab |
| `10-user-settings.png` | `/settings` |
| `13-gant.png` | `/projects/website-redesign/gantt` |
| `14-cumulative.png` | `/projects/product-platform/charts` → Cumulative tab |
| `15-scrum-backlog.png` | `/projects/product-platform/backlog` |
| `16-scrum-throughput.png` | `/projects/product-platform/charts` → Throughput tab |
| `17-scrum-burndown.png` | `/projects/product-platform/charts` → Burndown tab + select sprint |
| `18-scrum-burnup.png` | `/projects/product-platform/charts` → Burnup tab + select sprint |
| `19-scrum-release.png` | `/projects/product-platform/charts` → Release tab |
| `20-standby-shift.png` | `/time-tracking` |
| `21-ticket-list.png` | `/customers/{id}/tickets` — ticket list for first customer |
| `22-ticket-detail.png` | `/customers/{id}/tickets/{ticketId}` — click first ticket |

**Screenshots 11–12 (chat reactions)** require interaction in the embedded chat panel and are not yet automated — capture manually or run with `DEBUG=pw:api` to verify selectors.

**Adding a new screenshot:** add a new test in `frontend/e2e/screenshots.spec.js` following the existing pattern (create context from `AUTH_FILE`, navigate, wait, screenshot). To update all reference PNGs, run `make screenshots`. For individual updates, point your test at the specific view and use Playwright's `--update-snapshots` if using `toHaveScreenshot` assertions.

---

## Deployment notes

- `deploy/` has ready-made templates for systemd, nginx (with SSL), and Apache.
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
