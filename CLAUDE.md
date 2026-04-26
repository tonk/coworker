# WarmDesk — Developer Guide for Claude

WarmDesk is a self-hosted project management tool (Kanban boards, team chat, discussions, time reporting). It has a Go backend and a Vue 3 frontend; both live in this repository. A Tauri wrapper produces native desktop apps from the same frontend code.

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

```bash
make build          # builds frontend + backend → dist/
make run            # build then run
make appimage       # Linux AppImage (requires Rust + system libs)
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
  config/            # Config struct + env var / YAML loading
  database/          # GORM init, AutoMigrate for all models
  handlers/          # One file per feature area (card.go, report.go, …)
  middleware/        # Auth (JWT), AdminOnly, APIKeyAuth, CORS
  models/            # GORM model structs (board.go, user.go, project.go, …)
  router/            # Single router.go — all routes in one place
  services/          # Business logic (auth, email, project helpers, ordering)
  ws/                # WebSocket hub + client + pub/sub (memory & Redis)

frontend/
  src/
    api/             # Axios wrappers, one file per domain (projects.js, reports.js, …)
    components/      # Reusable Vue components (board/, chat/, common/, layout/)
    composables/     # useTheme, useWebSocket, useDateFormat, useAvatar, …
    i18n/            # en.json + nl, de, fr, es, da, sv, nb, fi, is, pt, it — all keys must be mirrored
    router/          # index.js — all routes + auth guards
    stores/          # Pinia stores (auth, board, chat, project, ui, …)
    styles/          # Global CSS custom properties (light/dark theme vars)
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
| `jwt_secret` | `JWT_SECRET` | *(change in prod)* |
| `web_dir` | `WEB_DIR` | `dist/web` |
| `upload_dir` | `UPLOAD_DIR` | `./uploads` |
| `max_upload_mb` | `MAX_UPLOAD_MB` | `25` |
| `redis_url` | `REDIS_URL` | *(optional)* |
| `allowed_origins` | `ALLOWED_ORIGINS` | `http://localhost:5173` |

---

## Architecture decisions

### Database
- **GORM AutoMigrate runs on every startup** — no separate migration files. Adding a new field to a model struct is all that is needed; the column appears on next boot.
- Supported drivers: `sqlite`, `postgres`, `mysql`. The driver string goes in `db_driver`, the DSN in `db_dsn`.
- Card numbering (`PRJ-1`, `PRJ-2`, …) is maintained by an atomic `card_counter` increment on the `projects` table.

### Authentication
- **JWT access token**: 15 min expiry, HS256. Claims: `UserID`, `Username`, `GlobalRole`.
- **JWT refresh token**: 7 day expiry. Frontend auto-refreshes silently on 401.
- **API keys**: SHA-256 hash stored in DB. Auth via `X-API-Key` header or `?api_key=` query param. Used for the Ticket API (CI/CD automation).
- Middleware sets context keys consumed by handlers: `middleware.GetUserID(c)`, `middleware.GetGlobalRole(c)`.

### System settings
Settings (SMTP, locale defaults, company branding, session timeout, …) are stored as key/value rows in `system_settings`. They are read at request time via `loadAllSettings()` so changes take effect **without a restart**. `handlers/system.go` owns all setting keys as package-level constants.

### WebSocket
One `Hub` per project, created on first connection, destroyed when empty. Messages are JSON `{type, payload}`. Handlers call `ws.BroadcastToProject(projectID, msg)`. For horizontal scaling, replace the in-memory pub/sub with Redis (`redis_url` in config).

### Frontend state
- **Pinia stores** own all shared state; components read from stores and call store actions.
- **`board.js`** is the most complex store: it owns columns + cards, applies WebSocket updates, and handles drag-drop reordering.
- **`useWebSocket.js`** establishes one connection per project view and routes messages to the appropriate store by type prefix (`board.*`, `chat.*`, `topic.*`, `presence.*`).

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

Files are stored in `upload_dir` (default `./uploads`) with randomised hex names. The original filename and MIME type are recorded in the `attachments` table. Ownership is by `owner_type` + `owner_id` (`card`, `card_comment`, `chat_message`, `conv_message`). Images are served inline; other files are forced as downloads.

---

## Time reporting

`GET /api/v1/reports/time` returns JSON grouped by project. The frontend (`ReportView.vue`) renders the table and exports:
- **PDF**: `window.print()` + `@media print` CSS (filters hidden, table styled)
- **XLSX**: SheetJS (`xlsx` npm package, dynamically imported)

---

## No tests

There are currently no automated tests (no `*_test.go` files, no Vitest/Jest config). When adding features, verify manually; do not add a testing framework unless explicitly asked.

---

## Deployment notes

- `deploy/` has ready-made templates for systemd, nginx (with SSL), and Apache.
- For multi-instance deployments set `redis_url` — this routes WebSocket broadcasts through Redis pub/sub instead of in-process memory.
- First-run with an empty DB: register the first user normally; promote to admin with a direct DB update or via another admin account.
