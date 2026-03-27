# Coworker

A self-hosted, multi-user Kanban board application with real-time
collaboration, direct messaging, and a ticket API.

## Experiment

This is an experiment, and a biggie :-)

I haven't written a single line of code, I only created, and changed the
`what.md` and asked Claude Code to generate the APP.

## Features

- **Kanban boards** — columns, cards, drag-and-drop reorder, labels, priorities, due dates, assignees, watchers, markdown descriptions and comments
- **Card sorting** — sort column cards by date, assignee, or priority (ascending / descending)
- **Comment replies** — reply to any comment; replies are visually indented
- **Multi-project** — each project has its own board, members, and chat
- **Role-based access** — global roles (admin / user) and per-project roles (owner / member / viewer)
- **Real-time** — board changes, card moves, and chat messages sync instantly across all connected users via WebSocket
- **Internal chat** — per-project team chat and direct messages between users
- **Unread DM notifications** — pulsing indicator in the sidebar and header when there are unread direct messages
- **Sidebar** — starred projects, live online-users list, auto-refreshes when users are added or removed
- **Starred project indicator** — ★ shown next to Project Settings in the board toolbar when a project is starred
- **Dark / light / system theme** — defaults to light
- **Multi-language** — English, Dutch (Nederlands), German (Deutsch), Spanish (Español), French (Français)
- **User settings** — display name, avatar, email, locale, theme, date/time format, timezone, password change
- **Admin panel** — manage all users (create, edit, assign projects, disable, delete) and all projects; toggle public registration on/off; configure global defaults (theme, locale, date format, timezone, font)
- **Ticket API** — create cards, add comments, and move cards via API key (for CI/CD pipelines and external integrations)
- **Database support** — SQLite (zero configuration), PostgreSQL, MySQL/MariaDB

## Quick Start

### Development

```bash
# Terminal 1 — backend (Go)
cd backend
go run .

# Terminal 2 — frontend (Vue 3 + Vite)
cd frontend
npm install
npm run dev
```

Open **http://localhost:5173** in your browser.

### Production build

```bash
make build
cd dist
WEB_DIR=./web ./coworker
```

Open **http://localhost:8080**.

## Configuration

Copy the example config file and edit it:

```bash
cp coworker.yaml.example coworker.yaml
```

Settings can also be provided as environment variables, which always take precedence over the config file. Key options:

| Option | Env var | Default | Description |
|--------|---------|---------|-------------|
| `port` | `PORT` | `8080` | HTTP listen port |
| `db_driver` | `DB_DRIVER` | `sqlite` | `sqlite` / `postgres` / `mysql` |
| `db_dsn` | `DB_DSN` | `./coworker.db` | Database connection string |
| `jwt_secret` | `JWT_SECRET` | *(change this)* | Secret for signing JWT tokens |
| `allowed_origins` | `ALLOWED_ORIGINS` | `http://localhost:8080` | CORS allowed origins |
| `default_locale` | `DEFAULT_LOCALE` | `en` | Default UI language for new users |
| `gin_mode` | `GIN_MODE` | `debug` | `debug` or `release` |
| `api_log` | `API_LOG` | `true` | Log incoming HTTP requests |
| `db_log` | `DB_LOG` | `info` | DB query log level: `silent` / `error` / `warn` / `info` |

See [INSTALL.md](INSTALL.md) for full options and deployment instructions.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25, Gin, GORM, gorilla/websocket |
| Frontend | Vue 3, Vite, Pinia, vue-router, vue-i18n, EasyMDE |
| Database | SQLite / PostgreSQL / MySQL |
| Auth | JWT (access + refresh tokens), bcrypt |

## Ticket API

Automate ticket management from CI/CD pipelines or external tools using API keys.

Generate an API key under **Project Settings → API Keys**, then use it with any of the endpoints below.

```
POST  /api/v1/ticket/{slug}/cards                    — create a card
POST  /api/v1/ticket/{slug}/cards/{id}/comments      — add a comment
PATCH /api/v1/ticket/{slug}/cards/{id}/move          — move to a column
```

Pass the key in the `X-API-Key` header or as `?api_key=` query parameter.

## Installation

See [INSTALL.md](INSTALL.md) for full instructions including:
- Building from source
- Running as a systemd service
- Nginx and Apache reverse proxy configuration
- PostgreSQL / MySQL setup
- First admin account setup
