# WarmDesk — Administrator Guide

## Contents

1. [Installation](#1-installation)
2. [Configuration Reference](#2-configuration-reference)
3. [Database Setup](#3-database-setup)
4. [Running as a Service](#4-running-as-a-service)
5. [Reverse Proxy](#5-reverse-proxy)
6. [First Admin Account](#6-first-admin-account)
7. [Admin Panel](#7-admin-panel)
8. [SMTP Email](#8-smtp-email)
9. [Company Branding](#9-company-branding)
10. [System Settings](#10-system-settings)
11. [Password Policy](#11-password-policy)
12. [Horizontal Scaling](#12-horizontal-scaling)
13. [Desktop Apps](#13-desktop-apps)
14. [Updates](#14-updates)
15. [Backup and Recovery](#15-backup-and-recovery)
16. [Demo Data](#16-demo-data)
17. [Migration Tools](#17-migration-tools)
18. [Security Checklist](#18-security-checklist)

---

## 1. Installation

For full build-from-source and quick-start instructions see
[INSTALL.md](../INSTALL.md). This guide assumes the binary is already running
and focuses on configuration and operations.

### Requirements at runtime

| Component | Minimum |
|-----------|---------|
| OS | Linux, macOS, or Windows |
| CPU | 1 core |
| RAM | 128 MB (SQLite) / 256 MB (PostgreSQL / MySQL) |
| Disk | 50 MB for the binary + your database and uploaded files |
| Network | Outbound SMTP (optional); inbound HTTP on your chosen port |

No external runtime dependencies — Go produces a single statically-linked
binary (except for SQLite, which requires `glibc` / `musl`).

---

## 2. Configuration Reference

Configuration is loaded in priority order (highest wins):

1. CLI flag `--config /path/to/file.yaml`
2. Environment variable `CONFIG_FILE=/path/to/file.yaml`
3. `warmdesk.yaml` in the current working directory
4. Built-in defaults

Every YAML key has a matching environment variable. Environment variables
always override the YAML file.

### Full configuration reference

```yaml
# ── Server ────────────────────────────────────────────────────────────────────
port: 8080                        # PORT — HTTP listen port
base_url: "https://desk.example.com"  # BASE_URL — public URL; sets Swagger UI host
allowed_origins: "https://app.example.com"  # ALLOWED_ORIGINS — CORS origins

# ── Security ──────────────────────────────────────────────────────────────────
jwt_secret: "change-me-in-production"  # JWT_SECRET — HS256 signing key
                                        # Generate: openssl rand -hex 32

# ── Web assets ────────────────────────────────────────────────────────────────
web_dir: "./web"                  # WEB_DIR — compiled frontend (required in prod)

# ── Database ──────────────────────────────────────────────────────────────────
db_driver: "sqlite"               # DB_DRIVER — sqlite | postgres | mysql
db_dsn: "./warmdesk.db"           # DB_DSN — file path or connection string
db_log: "warn"                    # DB_LOG — silent | error | warn | info

# ── File uploads ──────────────────────────────────────────────────────────────
upload_dir: "./uploads"           # UPLOAD_DIR — where attachments are stored
max_upload_mb: 25                 # MAX_UPLOAD_MB — per-file upload limit

# ── Group video calls (optional, LiveKit SFU) ────────────────────────────────
livekit_url: "wss://livekit.example.com"  # LIVEKIT_URL — LiveKit websocket URL
livekit_api_key: "APIxxxxxxxxxxxxxxxx"    # LIVEKIT_API_KEY — server API key
livekit_api_secret: "your-secret"         # LIVEKIT_API_SECRET — matching secret

# ── Logging ───────────────────────────────────────────────────────────────────
gin_mode: "release"               # GIN_MODE — debug | release
api_log: false                    # API_LOG — log every HTTP request

# ── Redis (optional — for horizontal scaling) ──────────────────────────────────
redis_url: ""                     # REDIS_URL — e.g. redis://localhost:6379
                                  # Leave empty to use in-process pub/sub

# ── Locale defaults (overridden per user) ─────────────────────────────────────
default_locale: "en"              # DEFAULT_LOCALE — en | nl | de | fr | es | da | sv | nb | fi | is | pt | it
```

### Generating a strong JWT secret

```bash
openssl rand -hex 32
# or
python3 -c "import secrets; print(secrets.token_hex(32))"
```

Never use the built-in default `change-me-in-production` in any environment
that is accessible from a network.

---

## 3. Database Setup

### SQLite (default — recommended for single-server installs)

No setup needed. WarmDesk creates the file automatically.

```yaml
db_driver: sqlite
db_dsn: /var/lib/warmdesk/warmdesk.db
```

Ensure the directory is writable by the process user and is on a volume that is
included in your backup.

### PostgreSQL

```bash
# Create database and user
psql -U postgres -c "CREATE USER warmdesk WITH PASSWORD 'secret';"
psql -U postgres -c "CREATE DATABASE warmdesk OWNER warmdesk;"
```

```yaml
db_driver: postgres
db_dsn: "host=localhost user=warmdesk password=secret dbname=warmdesk port=5432 sslmode=require"
```

### MySQL / MariaDB

```bash
mysql -u root -p -e "CREATE DATABASE warmdesk CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p -e "CREATE USER 'warmdesk'@'localhost' IDENTIFIED BY 'secret';"
mysql -u root -p -e "GRANT ALL PRIVILEGES ON warmdesk.* TO 'warmdesk'@'localhost';"
```

```yaml
db_driver: mysql
db_dsn: "warmdesk:secret@tcp(localhost:3306)/warmdesk?charset=utf8mb4&parseTime=True&loc=Local"
```

### Database TLS

> **Strongly advised:** if your database server is not on the same host as WarmDesk,
> enable TLS. Without it, all queries — including passwords and session data — travel
> in plaintext between the application and the database. Use `verify-full` in
> production; `disable` (the default) is only appropriate when the database socket
> is a local Unix socket or a loopback connection on the same machine.

Both PostgreSQL and MySQL support encrypted connections. Use the `db_tls_*`
settings (or their `DB_TLS_*` env var equivalents) instead of embedding TLS
parameters in the DSN directly.

| Setting | Env var | Description |
|---------|---------|-------------|
| `db_tls_mode` | `DB_TLS_MODE` | `disable` (default) / `require` / `verify-ca` / `verify-full` |
| `db_tls_ca_cert` | `DB_TLS_CA_CERT` | Path to CA certificate file |
| `db_tls_cert` | `DB_TLS_CERT` | Path to client certificate (mTLS, optional) |
| `db_tls_key` | `DB_TLS_KEY` | Path to client key (mTLS, optional) |

```yaml
# Encrypt and fully verify the server certificate
db_tls_mode: "verify-full"
db_tls_ca_cert: "/etc/ssl/warmdesk/ca.pem"

# Additionally authenticate with a client certificate (mTLS)
db_tls_cert: "/etc/ssl/warmdesk/client.crt"
db_tls_key:  "/etc/ssl/warmdesk/client.key"
```

`require` encrypts the connection but skips certificate verification (useful
for self-signed certs in development). `verify-ca` checks the certificate
chain; `verify-full` also checks that the hostname matches the certificate CN.

### Schema migration

WarmDesk runs **GORM AutoMigrate** on every startup. New columns and tables are
created automatically; existing data is never destroyed. There are no separate
migration files to run.

---

## 4. Running as a Service

### systemd (Linux — recommended)

A ready-made unit file is at `deploy/warmdesk.service`. Edit it before
installing — at minimum set `JWT_SECRET` and `ALLOWED_ORIGINS`.

```ini
[Service]
Environment="JWT_SECRET=your-secret-here"
Environment="ALLOWED_ORIGINS=https://warmdesk.example.com"
Environment="DB_DSN=/var/lib/warmdesk/warmdesk.db"
```

```bash
sudo useradd -r -s /bin/false -d /opt/warmdesk warmdesk
sudo mkdir -p /opt/warmdesk/{data,uploads}
sudo cp -r dist/. /opt/warmdesk/
sudo chown -R warmdesk:warmdesk /opt/warmdesk

sudo cp deploy/warmdesk.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now warmdesk
sudo journalctl -u warmdesk -f   # follow logs
```

---

## 5. Reverse Proxy

Always run WarmDesk behind a reverse proxy in production. Ready-made configs
are in `deploy/`.

### Nginx (`deploy/nginx.conf`)

Key configuration points:

```nginx
# Increase timeouts for WebSocket connections
proxy_read_timeout 3600s;
proxy_send_timeout 3600s;

# WebSocket upgrade
proxy_http_version 1.1;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";

# Forward real IP
proxy_set_header X-Real-IP $remote_addr;
```

```bash
sudo cp deploy/nginx.conf /etc/nginx/sites-available/warmdesk
# Edit: replace yourdomain.com and SSL certificate paths
sudo ln -s /etc/nginx/sites-available/warmdesk /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

### Apache (`deploy/apache.conf`)

```bash
sudo a2enmod proxy proxy_http proxy_wstunnel ssl headers rewrite
sudo cp deploy/apache.conf /etc/apache2/sites-available/warmdesk.conf
# Edit: replace yourdomain.com and SSL certificate paths
sudo a2ensite warmdesk
sudo systemctl reload apache2
```

The template uses a dedicated `<Location "/api/v1/ws">` block with `ProxyPass ws://…` and `ProxyTimeout 86400` to keep WebSocket connections (board updates, chat, video-call signalling) alive for up to 24 hours without a forced reconnect. All other traffic is handled by the catch-all `ProxyPass / http://…` below it.

### Server TLS

WarmDesk can serve HTTPS directly without a reverse proxy. Set both `tls_cert`
and `tls_key` (or their `TLS_CERT` / `TLS_KEY` env vars) to enable it:

```yaml
tls_cert: "/etc/ssl/warmdesk/server.crt"
tls_key:  "/etc/ssl/warmdesk/server.key"
```

When either value is absent the server starts in plain HTTP mode. This is the
right choice when TLS is terminated upstream by nginx, Apache, or a load
balancer — no extra configuration is needed in that case.

> **Note:** if you enable server TLS, update `ALLOWED_ORIGINS` and any reverse
> proxy config to use `https://` URLs accordingly.

---

### CORS

Set `ALLOWED_ORIGINS` to the **exact** origin that users access WarmDesk from
(including scheme and port). A mismatch will prevent the frontend from making
API calls.

```bash
# Single domain
ALLOWED_ORIGINS=https://warmdesk.example.com

# Multiple domains (comma-separated)
ALLOWED_ORIGINS=https://warmdesk.example.com,https://warmdesk.internal

# Allow any origin (useful when the desktop app connects to an internal server
# that is not exposed to the internet)
ALLOWED_ORIGINS=*
```

Desktop app origins (`tauri://localhost`, `https://tauri.localhost`,
`http://tauri.localhost`) are always allowed automatically regardless of this
setting — no extra configuration is needed for the native clients.

---

## 6. First Admin Account

Register normally through the web interface. The **first account registered
on a fresh database is automatically made a global admin** — no database
manipulation required.

Once one admin exists, you can promote further users through
**Admin → Users → Edit** in the web interface.

If public registration is not desired, disable it after creating the first
admin: **Admin → Settings → Allow public registration → Off**.

> **Recovery:** if you ever need to promote an account via the database
> directly (e.g. after accidentally demoting the only admin):
>
> **SQLite**
> ```bash
> sqlite3 /var/lib/warmdesk/warmdesk.db \
>   "UPDATE users SET global_role='admin' WHERE username='yourname';"
> ```
>
> **PostgreSQL**
> ```sql
> UPDATE users SET global_role = 'admin' WHERE username = 'yourname';
> ```
>
> **MySQL**
> ```bash
> mysql -u warmdesk -p -e "UPDATE users SET global_role='admin' WHERE username='yourname';" warmdesk
> ```

---

## 7. Admin Panel

Access the admin panel via the **Admin** link in the navigation (visible to
admin users only).

### Users

| Action | Where |
|--------|-------|
| Create a user | Admin → Users → Create User |
| Edit name, email, role | Admin → Users → (click user) |
| Reset a password | Admin → Users → Edit → Change Password |
| Disable / enable | Admin → Users → Edit → Enabled toggle |
| Assign to projects | Admin → Users → (click user) → Projects tab |
| Assign to customers | Admin → Users → (click user) → Customer Access picker |
| Delete a user | Admin → Users → Edit → Delete (permanent) |

**Global roles**

| Role | Access |
|------|--------|
| `user` | Can use the application; sees only their own projects |
| `admin` | Full access to all projects, all admin panel features |
| `viewer` | Read-only access; cannot create or modify anything |
| `metrics` | Can only call `GET /api/v1/metrics`; no access to any other endpoint |
| `backup` | Can only call `POST /api/v1/backup`; no access to any other endpoint |

The `metrics` role is intended for Prometheus scraper accounts. Create a dedicated user, set their role to `metrics`, generate an API key in User Settings, and configure Prometheus to send `Authorization: Bearer <token>` (or `?api_key=<key>`) with each scrape request.

The `backup` role is intended for automated backup scripts and cron jobs. See [section 15](#15-backup-and-recovery) for setup instructions.

### Groups

Groups let you manage project and customer access for a set of users in one place instead of assigning each user individually.

| Action | Where |
|--------|-------|
| Create a group | Admin → Groups → Create Group |
| Edit name / description | Admin → Groups → (click group) |
| Add / remove members | Admin → Groups → (click group) → Members tab |
| Grant project access | Admin → Groups → (click group) → Project Access tab |
| Grant customer access | Admin → Groups → (click group) → Customer Access tab |
| Delete a group | Admin → Groups → (click group) → Delete |

**How group access works**

Group access is **additive**: a user's effective role on a project or customer is the highest role they hold, whether that comes from a direct assignment or from any group they belong to. Removing a user from a group immediately revokes any access they received exclusively through that group.

Global admins bypass all access checks and are unaffected by group assignments.

**Project roles via group**

| Role | Effect |
|------|--------|
| `viewer` | Read-only access to the project board and cards |
| `member` | Can create and edit cards, post comments |
| `owner` | Full project permissions including settings and member management |

**Customer roles via group**

| Role | Effect |
|------|--------|
| `member` | Can view the customer and its contracts |
| `admin` | Can edit the customer, create/edit contracts, manage members |

### Customer access control

Non-admin users only see the customers they are **explicitly assigned to**. A
user with no customer assignments sees an empty customer list.

Assign customers in **Admin → Users → Edit User → Customer Access**. Click a
customer chip to toggle access; when a chip is selected, click the small **M**
or **A** badge inside it to set the role:

| Role | Access |
|------|--------|
| **M** (Member) | Can see the customer and its contracts and projects |
| **A** (Admin) | All member permissions plus: edit customer details, create/edit contracts, manage the customer's member list |

Global admins bypass all customer access checks and always see every customer.

**Customer-admin self-service**

A user with the Admin role on a customer can also manage that customer's member
list directly from the **Customer Detail** page → **Members** section, without
needing access to the Admin panel. They can add users, change roles, and remove
members. They cannot remove themselves from the customer (prevents self-lockout).

### Prometheus metrics

`GET /api/v1/metrics` returns project, column, and card counts in Prometheus text format. Accessible to `admin` and `metrics` roles.

Example Prometheus scrape config:

```yaml
scrape_configs:
  - job_name: warmdesk
    static_configs:
      - targets: ['warmdesk.example.com']
    metrics_path: /api/v1/metrics
    scheme: https
    authorization:
      credentials: <api-key>
```

Metrics exposed:

| Metric | Labels | Description |
|--------|--------|-------------|
| `warmdesk_projects_total` | — | Active (non-archived) project count |
| `warmdesk_columns_total` | `project`, `project_name` | Column count per project |
| `warmdesk_cards_total` | `project`, `column`, `status` | Card count per column; `status` is `open` or `closed` |

### Customers

Admins create and manage customer organisations via **Admin → Customers**.

| Action | Where |
|--------|-------|
| Create a customer | Admin → Customers → Create Customer |
| Edit name / description / logo | Admin → Customers → (click Edit) |
| Delete a customer | Admin → Customers → (click Delete) |

Customer access (which users or groups can see a customer) is managed separately under **Admin → Users** and **Admin → Groups** — see [Customer access control](#customer-access-control) below.

### Projects

Admins can create, rename, archive, and delete any project regardless of
project membership. Access via **Admin → Projects**.

When creating a project the **Card Prefix** field sets the short identifier
used in all card references (e.g. `PRJ-42`). It is auto-generated from the
project name but can be freely edited before saving — 1–10 uppercase letters
or digits. The prefix must be **unique** across all projects; if the
auto-generated value is already taken a numeric suffix is appended
automatically (`WAR`, `WAR2`, `WAR3`, …). The prefix **cannot be changed
after the project is created** — existing card codes in commit messages and
external integrations would become invalid.

### System settings

All settings under **Admin → Settings** take effect **immediately without a
server restart**. They are stored in the database and loaded at request time.

---

## 8. SMTP Email

Email is used for:
- @mention notifications when the mentioned user is offline
- Password reset links (when a user clicks "Forgot password?" on the login page)

### Configuring SMTP

Go to **Admin → Settings → Email** and fill in:

| Field | Notes |
|-------|-------|
| SMTP Host | Hostname of your mail server, e.g. `smtp.gmail.com` |
| SMTP Port | Typically `587` (STARTTLS), `465` (TLS), or `25` (relay) |
| Username | Often your email address; leave empty for relay servers |
| Password | Leave empty for relay servers that don't require auth |
| From address | The `From:` header, e.g. `warmdesk@example.com` |
| From name | Display name, e.g. `WarmDesk` |

Click **Save** and then use **Send Test Email** to verify the configuration
before going live. Enter any email address in the test field and click **Send
Test** — a test message is delivered immediately.

### Gmail example

```
Host:     smtp.gmail.com
Port:     587
Username: youraddress@gmail.com
Password: (App Password — not your Google account password)
From:     youraddress@gmail.com
```

You must enable 2-factor authentication on the Google account and generate an
**App Password** for WarmDesk. Standard account passwords will not work.

### Auth-less relay (common in corporate environments)

Leave Username and Password empty. The mail server must be configured to accept
connections from the WarmDesk server's IP without authentication.

---

## 9. Company Branding

Go to **Admin → Settings → Branding** to set:

| Setting | Notes |
|---------|-------|
| Company name | Appears in the login screen branding panel and in the report header / PDF footer |
| Company logo (light background) | Upload an image file (JPG, PNG, GIF, WebP, or SVG) for use on light-themed screens; displayed in the login screen branding panel, the on-screen report header, and in exported PDFs. Click **Clear** to remove. |
| Company logo (dark background) | Optional second logo shown when the login page is in dark mode; if omitted, the light logo is used on both themes. |
| Show company branding on the login screen | When enabled, the login page splits into two panels: the left panel shows the company logo and name; the right panel shows the login form. Requires at least one of company name or a logo to be set. The login page switches between the light and dark logo automatically as the user toggles the login theme. |

Changes take effect immediately — the login screen and reports both reflect the new values without a restart.

---

## 10. System Settings

### Session timeout

**Admin → Settings → Session → Idle Timeout (minutes)**

Default is `60` minutes. Set to `0` to disable the timeout entirely (sessions
last until the refresh token expires — 7 days). Fractional values are not
supported; enter a whole number of minutes.

The timer resets on any user interaction (navigation, clicks, API calls). When
the timeout expires the user is redirected to the login page.

### Default initial columns

**Admin → Settings → New Project Defaults → Initial Columns**

Enter one column name per line. These columns are created automatically whenever
a new project is made. The built-in default is:

```
Backlog
In Progress
Test & Review
To Production
```

### Default initial labels

**Admin → Settings → New Project Defaults → Initial Labels**

Enter one label name per line. These labels are created automatically for every
new project. The built-in default is:

```
Bug
Feature
Design
Content
```

> **Note:** Changes to Initial Columns and Initial Labels only affect
> **new** projects created after saving. Existing projects are not modified.
>
> Click the **Save** button that appears below the textareas to persist your
> changes before switching to another settings tab.

### Public registration

**Admin → Settings → Allow public registration**

When off, the Register link disappears from the login page. Users can only be
created by administrators via the admin panel.

### Scrum Story Points

**Admin → Settings → Scrum Story Points**

When enabled, a **Story Points** number field appears on every card detail panel
and a compact **SP badge** is shown on the card tile on the board. Users can
enter any whole number; leaving the field empty means no estimate has been given.

This setting takes effect immediately for all open sessions without a page
reload.

### Group video calls (LiveKit)

WarmDesk uses LiveKit for group video calls in DM group conversations with 3+ members.

Set these server values in `warmdesk.yaml` (or env vars) and restart the service:

```yaml
livekit_url: "wss://livekit.example.com"
livekit_api_key: "APIxxxxxxxxxxxxxxxx"
livekit_api_secret: "your-secret"
```

Without these values, 1:1 WebRTC calls still work, but group video shows an in-app "not configured" banner.

### Global defaults (overridden per user)

Admins set global defaults for:
- Date / time format
- Timezone
- Theme (light / dark / system)
- Language
- Font and font size

Individual users can override any of these in their own User Settings.

---

## 11. Password Policy

**Admin → Settings → Password Policy**

Configure the requirements that all new passwords must satisfy — whether set during registration, changed by the user in Settings, or reset via email.

| Setting | Notes |
|---------|-------|
| **Minimum length** | Minimum number of characters; default is 12; the floor is 8 and cannot be set lower |
| **Require uppercase** | At least one uppercase letter (A–Z) |
| **Require lowercase** | At least one lowercase letter (a–z) |
| **Require digit** | At least one digit (0–9) |
| **Require special character** | At least one character from `!@#$%^&*()_+-=[]{}|;':",./<>?` |

Click **Save** below the checkboxes to apply changes. The policy is enforced immediately; existing user passwords are not affected until they are next changed.

The active requirements are displayed to users beneath the new-password field in their Settings page and on the password reset form.

---

## 12. Horizontal Scaling

WarmDesk uses WebSocket connections for real-time updates. In a single-instance
setup, connections are managed in memory. When running multiple instances behind
a load balancer, each instance has its own connection pool — a message broadcast
by one instance is not seen by clients connected to another.

### Redis pub/sub

Enable Redis to route broadcasts across all instances:

```yaml
redis_url: redis://localhost:6379
```

or

```bash
REDIS_URL=redis://username:password@redis-host:6379/0
```

When `redis_url` is set, WarmDesk subscribes to a Redis channel and all
`BroadcastToProject` calls publish to that channel. Every instance receives the
message and delivers it to its own connected clients.

### Load balancer requirements

WebSocket connections require **sticky sessions** (a.k.a. session affinity) at
the load balancer. Without sticky sessions a client's HTTP upgrade request and
subsequent WebSocket frames may reach different instances and fail.

With Redis enabled, sticky sessions are not strictly required for correctness,
but they reduce Redis traffic.

### Redis configuration

WarmDesk uses a single pub/sub channel per subscription scope. A minimal Redis
install with default settings works. No persistence (AOF/RDB) is required for
the pub/sub use case.

```bash
# Test connectivity
redis-cli -h redis-host ping
```

---

## 13. Desktop Apps

WarmDesk ships Tauri-based desktop apps that wrap the web frontend and connect
to a WarmDesk server. The apps are standalone — they do not bundle the server.

Users configure the server URL in the app's **Connect** screen on first launch.
The URL is saved locally and can be changed at any time via the **Change** link
shown next to the server URL on the login page.

### Command-line flags

| Flag | Description |
|------|-------------|
| `--version`, `-V` | Print the app version and exit |
| `--maximized` | Start the window maximised |

```bash
# Examples (Linux AppImage)
./WarmDesk.AppImage --version
./WarmDesk.AppImage --maximized
```

### Distributing desktop apps

Pre-built desktop apps are attached to each GitHub release:

| Platform | File | Notes |
|----------|------|-------|
| Linux | `WarmDesk-vX.Y.Z-x86_64.AppImage` | Portable; no installation required |
| Linux | `WarmDesk-vX.Y.Z-amd64.deb` | Debian/Ubuntu package |
| Linux | `WarmDesk-vX.Y.Z-x86_64.rpm` | RPM package |
| Windows | `WarmDesk-vX.Y.Z-x64-setup.exe` | NSIS installer |
| Windows | `WarmDesk-vX.Y.Z-x64-portable.zip` | Extract and run `WarmDesk.exe` |
| macOS | `WarmDesk-vX.Y.Z-universal.dmg` | Universal binary (Intel + Apple Silicon) |

Each file has a companion `.asc` detached GPG signature. To verify a download:

```bash
# Import the WarmDesk release key once
gpg --import signing-key.asc          # key is in the root of the repository

# Verify any downloaded file
gpg --verify WarmDesk-vX.Y.Z-x86_64.AppImage.asc WarmDesk-vX.Y.Z-x86_64.AppImage
```

A `Good signature from "WarmDesk Releases"` message confirms the file is unmodified and was signed by the official release key.

### Building desktop apps from source

See [INSTALL.md](../INSTALL.md) — the `make appimage`, `make dmg`, and
`make windows-installer` targets.

---

## 14. Updates

```bash
# Pull latest source
git pull

# Rebuild
make build

# Restart the service
sudo systemctl restart warmdesk
```

AutoMigrate runs on startup and applies any schema changes automatically. No
manual migration step is needed.

### Zero-downtime update (advanced)

1. Build the new binary on a staging machine.
2. Copy the binary and `web/` directory to the server.
3. Send `SIGTERM` to the running process (systemd restart handles this).
4. The process finishes in-flight requests before exiting.

---

## 15. Backup and Recovery

### What to back up

| Item | Location | Frequency |
|------|----------|-----------|
| Database | `./backups/` (via admin panel) or raw DB file | Daily or more |
| Uploads | `upload_dir` (default `./uploads/`) | Daily or more |
| Config | `warmdesk.yaml` | On change |

### Via the Admin panel

The **Admin → Backup / Restore** tab provides full backup management without touching the server directly.

**Scheduled backups**

At the top of the tab you can configure automatic backups:

| Field | Description |
|-------|-------------|
| **Interval** | Disabled / Every 6 h (4×/day) / Every 8 h (3×/day) / Every 12 h (2×/day) / Once a day |
| **Start time** | Optional HH:MM anchor for slot-based scheduling (e.g. `02:00`). When set, backups run at fixed time-of-day slots derived from the start time and the chosen interval — for example, `02:00` + every 6 h produces runs at 02:00, 08:00, 14:00, and 20:00. Leave empty to use the default behaviour (interval counted from the last run). |
| **Keep last (backups)** | Maximum number of backup files to retain on disk. Oldest files are pruned automatically after each scheduled run. Default: 10. |

The scheduler runs server-side — no cron job or external tool needed. It checks every 5 minutes whether a backup is due and creates a new file when a slot has been missed. The last run time and next scheduled run are displayed below the settings.

> **Tip:** set the start time to an off-peak hour (e.g. `02:00`) to ensure backups never run during business hours.

**Create a manual backup**

Click **Create Backup**. WarmDesk creates a timestamped file in `./backups/` next to the server binary:

```
./backups/warmdesk_db_20260416_1430_a3f9.db   ← SQLite
./backups/warmdesk_db_20260416_1430_a3f9.sql  ← PostgreSQL / MySQL
```

The filename format is `warmdesk_db_YYYYMMDD_HHMM_<4-hex>.db/.sql`. The four-character hex suffix is random and exists solely to prevent collisions when two backups are triggered within the same minute (e.g. a scheduled backup and a manual one firing simultaneously).

For SQLite, `VACUUM INTO` is used — an atomic online copy that requires no downtime. For PostgreSQL, `pg_dump --clean --if-exists` is used. For MySQL, `mysqldump` is used.

**List, restore, and delete backups**

All files in `./backups/` are listed in the tab with filename, size, and creation date, newest first.

- **Restore** — replaces the live database with the selected backup. SQLite restores are live (the connection pool is closed, the file is replaced, and the connection is reopened — no server restart needed). A confirmation prompt is shown before proceeding.
- **Download** — downloads the backup file directly to your browser. Useful for offsite storage or transferring backups to another server.
- **Delete** — removes the backup file from disk after confirmation.

Every backup and restore operation is logged to the server log with the filename, database driver, user ID, and client IP.

### Automated backups (cron / CI)

Use the `backup` global role to create a service account for automated backups:

1. In **Admin → Users**, create a user (e.g. `backup-bot`) and set their role to **Backup**.
2. Log in as that user and generate an API key under **User Settings → API Keys**.
3. Call `POST /api/v1/backup` on a schedule:

```bash
# Cron — daily at 02:00
0 2 * * * curl -sf -X POST https://desk.example.com/api/v1/backup \
               -H "X-API-Key: cwk_your_key_here"
```

The `backup` role cannot access any other endpoint.

### Server-side backup (manual)

You can also back up the database directly on the server without going through the API.

**SQLite**

```bash
# Hot copy using VACUUM INTO — safe while the server is running
sqlite3 /var/lib/warmdesk/warmdesk.db \
  ".backup /backup/warmdesk-$(date +%Y%m%d).db"
```

**PostgreSQL**

```bash
pg_dump -U warmdesk warmdesk | gzip > /backup/warmdesk-$(date +%Y%m%d).sql.gz
```

**Restoring manually**

```bash
# SQLite — stop the service, replace the file, restart
sudo systemctl stop warmdesk
cp /backup/warmdesk-20260416.db /var/lib/warmdesk/warmdesk.db
sudo systemctl start warmdesk

# PostgreSQL
gunzip -c /backup/warmdesk-20260416.sql.gz | psql -U warmdesk warmdesk
```

---

## 16. Demo Data

The `warmdesk-seed` binary ships alongside `warmdesk` and populates the
database with realistic demo content for evaluation and testing.

```bash
cd dist
./warmdesk-seed           # seed (idempotent — safe to run multiple times)
./warmdesk-seed --reset   # wipe all demo data and re-seed
```

**Demo accounts** (password for all: `demo1234`)

| Username | Display name | Role | Notes |
|----------|--------------|------|-------|
| `tonk` | Ton Kersten | admin | Persistent — not removed by `--reset` |
| `demo.admin` | Alex Admin | admin | |
| `demo.sarah` | Sarah Chen | user | Project admin: Website Redesign |
| `demo.marc` | Marc Dubois | user | Project admin: Mobile App v2 |
| `demo.lisa` | Lisa Park | user | Project admin: DevOps & Infra |
| `demo.priya` | Priya Nair | user | |
| `demo.james` | James O'Brien | user | |
| `demo.elena` | Elena Kovač | user | |
| `demo.raj` | Raj Sharma | user | |
| `demo.viewer` | Victor Viewer | viewer | Read-only demo account |

**Demo content**

- 3 projects: Website Redesign, Mobile App v2, DevOps & Infra
- Multiple columns per project with realistic cards including checklists, labels, priorities, start dates, due dates, time entries, and card cross-references
- Threaded topics per project
- 4 direct message conversations and 1 group chat with realistic history
- 3 customers (Acme Corporation, Globex Systems, Initech Ltd) with contracts linked to projects
- 3 groups: Frontend Team, DevOps Team, Acme Stakeholders (with members and project/customer access pre-configured)
- Favorite projects pre-set for each demo user (e.g. admin has all three; owners have their own project plus one adjacent)
- Favorite customers pre-set (Acme starred for admin, sarah, marc; Globex for marc, lisa)

### Training seeder

The `warmdesk-training` binary sets up isolated training environments — one
customer, contract, project, and user per slot.

```bash
# Create trainer (guru00) + 5 trainees (guru01…guru05), password base "Training"
./warmdesk-training 5 Training

# Remove all guru** training data from the database
./warmdesk-training --reset
```

**Training slots**

| Slot | Username | Password | Customer | Role |
|------|----------|----------|----------|------|
| 00 | `guru00` | `Training00` | All training customers | Customer-admin (sees all) |
| 01 | `guru01` | `Training01` | Ansible Laboratory 01 | Member (restricted to own) |
| … | … | … | … | … |

- Each trainee sees **only** their own customer.
- The trainer (guru00) has customer-admin access to every training customer so
  they can observe and assist all trainees.
- Users are seeded with a DiceBear avatar so they are visually distinct.
- The seeder is idempotent: re-running applies any missing access rows and
  avatars without duplicating data.

---

## 17. Migration Tools

`warmdesk-export` and `warmdesk-import` are standalone binaries (shipped in
`dist/` alongside the main server) for moving projects between WarmDesk and
other platforms.

| Binary | Direction |
|--------|-----------|
| `warmdesk-export` | WarmDesk → Jira, Trello, OpenProject, or Ryver |
| `warmdesk-import` | Jira, Trello, OpenProject, or Ryver → WarmDesk |

Both binaries export / import: columns, cards (with title, description,
priority, due date, labels, tags, assignees, checklist, comments, time entries,
attachments), and threaded topics.

### Configuration

Copy `warmdesk-migrate.yaml.example` to `warmdesk-migrate.yaml` and fill in
the connection details. Sensitive values can be supplied as environment
variables instead:

```
WARMDESK_URL        URL of your WarmDesk instance
WARMDESK_USERNAME   WarmDesk account username
WARMDESK_PASSWORD   WarmDesk account password
WARMDESK_PROJECT    Project slug (visible in the URL)
PLATFORM_API_TOKEN  API token for the target platform
PLATFORM_API_KEY    API key for the target platform (Trello)
```

Any field still missing after reading the file and environment is prompted
interactively.

### Column mapping

The `column_map` section in the config translates WarmDesk column names to the
corresponding status / list names on the target platform (and back on import):

```yaml
column_map:
  Backlog: "To Do"
  "In Progress": "In Progress"
  "Test & Review": "In Review"
  "To Production": Done
```

Columns not listed are passed through unchanged.

### Usage

```bash
cd dist

# Export a WarmDesk project to Jira
./warmdesk-export --config warmdesk-migrate.yaml

# Import a Jira project into WarmDesk
./warmdesk-import --config warmdesk-migrate.yaml

# Preview what would happen without making any changes
./warmdesk-export --dry-run
./warmdesk-import --dry-run
```

---

## 18. Security Checklist

Before exposing WarmDesk to the internet:

- [ ] Changed `JWT_SECRET` to a randomly generated 32-byte hex string
- [ ] Set `ALLOWED_ORIGINS` to the exact production domain
- [ ] Running behind HTTPS (TLS termination at the reverse proxy)
- [ ] `GIN_MODE=release` (suppresses debug output)
- [ ] `API_LOG=false` (or piped to a log file, not stdout)
- [ ] Database credentials are strong and not the defaults
- [ ] Uploads directory (`upload_dir`) is outside the web root
- [ ] Firewall allows inbound traffic on port 80/443 only; WarmDesk's port
      (8080) is not directly exposed
- [ ] Systemd service runs as a non-root dedicated user (`warmdesk`)
- [ ] Backup schedule is in place for the database and uploads
- [ ] Public registration disabled (`Allow public registration = off`) if only
      known users should access the instance
- [ ] Password policy configured (**Admin → Settings → Password Policy**) with
      a minimum length of at least 12 and one or more character-class requirements
- [ ] SMTP credentials (if used) are an app-specific password, not a primary
      account password
