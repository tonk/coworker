# Repository layout

```
warmdesk/
├── ansible/
├── backend/
├── deploy/
├── docs/
├── frontend/
└── screenshots/
```

## Root level

| Directory | Purpose |
|---|---|
| `ansible/` | Ansible collection (`ansilabnl.warmdesk`) for automating WarmDesk via playbooks |
| `backend/` | Go server — REST API, WebSocket, database, business logic |
| `deploy/` | Deployment templates — systemd service, nginx/Apache config, `.desktop` entry, `get_warmdesk` server update script, `update_warmdesk_client` desktop client update script |
| `docs/` | Developer and user-facing documentation (admin guide, API docs, user guide, setup guide) |
| `frontend/` | Vue 3 web app + Tauri desktop wrapper |
| `screenshots/` | 24 reference screenshots used in README.md and the Hugo website (`website/Makefile` copies them at build time) |

## `ansible/`

| Directory | Purpose |
|---|---|
| `meta/` | `runtime.yml` — Galaxy requirement (`requires_ansible`) |
| `plugins/doc_fragments/` | Shared doc fragment (`connection.py`) — auth params reused by all modules |
| `plugins/inventory/` | Dynamic inventory plugin — builds Ansible inventory from WarmDesk projects/users |
| `plugins/lookup/` | Lookup plugins — fetch cards, projects, customers, etc. in playbooks |
| `plugins/modules/` | Ansible modules — manage users, projects, cards, customers, etc. |
| `plugins/module_utils/` | Shared Python helpers — HTTP client, auth, name resolvers |

## `backend/`

| Directory | Purpose |
|---|---|
| `cmd/export` | Standalone export tool — migrate projects to Jira, Trello, OpenProject, or Ryver |
| `cmd/importer` | Standalone import tool — migrate from Jira, Trello, OpenProject, or Ryver |
| `cmd/seed` | Demo data seeder — populates a database with realistic sample content |
| `cmd/training` | Training/lab data setup tool |
| `config/` | Config struct + YAML/env loading |
| `database/` | GORM init and AutoMigrate |
| `docs/` | Swagger/OpenAPI spec (served at `/api/docs`) |
| `handlers/` | One file per feature area (cards, users, reports, backup, …) |
| `handlers/fonts/` | Embedded fonts for server-side PDF generation |
| `i18n/locales/` | Server-side translation strings (PDF export language support) |
| `middleware/` | JWT auth, admin-only guard, API key auth, CORS |
| `migrate/` | Pre-AutoMigrate data fixups (e.g. backfilling key prefixes) |
| `models/` | GORM model structs (User, Project, Card, …) |
| `router/` | All routes registered in one place |
| `services/` | Business logic shared across handlers (auth, ordering, project helpers, IMAP polling, XOAUTH2) |
| `staticweb/` | Embeds the compiled frontend into the Go binary (files placed here at build time) |
| `ws/` | WebSocket hub — real-time broadcast, in-memory and Redis pub/sub |

## `frontend/`

| Directory | Purpose |
|---|---|
| `public/fonts/` | Web fonts served as static assets |
| `src/api/` | Axios wrappers, one file per domain |
| `src/components/` | Reusable Vue components, grouped by feature area |
| `src/composables/` | Vue composables — `useTheme`, `useWebSocket`, `useDateFormat`, etc. |
| `src/i18n/` | 12 language JSON files |
| `src/router/` | Vue Router — all routes + auth guards |
| `src/stores/` | Pinia stores — board, auth, chat, project, UI, … |
| `src/styles/` | Global CSS custom properties (light/dark theme vars) |
| `src/types/` | TypeScript type definitions |
| `src/utils/` | Shared utility functions |
| `src/views/` | Page-level Vue components (one per route) |
| `src-tauri/capabilities/` | Tauri 2 permission declarations (what the app is allowed to do) |
| `src-tauri/gen/schemas/` | Auto-generated Tauri capability JSON schemas |
| `src-tauri/icons/` | App icons in all sizes (PNG, ICO, ICNS) |
| `src-tauri/src/` | Rust entry point for the Tauri desktop app (minimal — mostly glue code) |
