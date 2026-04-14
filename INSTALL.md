# WarmDesk — Installation Manual

## Requirements

| Component | Requirement |
|-----------|-------------|
| Go | 1.25 or later |
| Node.js | 20 or later |
| GCC | Required for SQLite (not needed for MySQL/PostgreSQL) |

---

## 1. Install Prerequisites

### Go

Download and install from https://go.dev/dl/

```bash
# Verify
go version
```

### Node.js

Download and install from https://nodejs.org/ (LTS recommended)

```bash
# Verify
node --version
npm --version
```

### GCC (for SQLite only)

- **Ubuntu / Debian**: `sudo apt install gcc`
- **RHEL / Fedora**: `sudo dnf install gcc`
- **macOS**: `xcode-select --install`
- **Windows**: Install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or use WSL

### Rust + Tauri CLI (desktop app builds only)

Required only when building the AppImage, DMG, or Windows installer.

```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"

# Install Tauri CLI
cargo install tauri-cli --version '^2'
```

---

## 2. Build

```bash
git clone <repo-url>
cd warmdesk
make build
```

Output is placed in `dist/`:

```
dist/
  warmdesk               # server binary (Linux/macOS) or warmdesk.exe (Windows)
  warmdesk-seed          # demo data seeder
  warmdesk-export        # migration: WarmDesk → Jira / Trello / OpenProject / Ryver
  warmdesk-import        # migration: Jira / Trello / OpenProject / Ryver → WarmDesk
  web/                   # compiled frontend assets
  warmdesk.yaml.example  # annotated server config template
  warmdesk-migrate.yaml.example  # migration tool config template
  deploy/                # systemd / nginx / Apache templates
  docs/                  # user, API, and admin documentation
```

---

## 3. Configure

WarmDesk looks for a `warmdesk.yaml` file in its working directory.
Copy the example and edit it:

```bash
cp warmdesk.yaml.example dist/warmdesk.yaml
# Edit dist/warmdesk.yaml with your database, secret, and domain settings
```

You can also specify a config file path on the command line — useful when running
multiple instances or keeping configs outside the working directory:

```bash
./warmdesk --config /etc/warmdesk/production.yaml
```

Priority order (highest wins): CLI `--config` flag → `CONFIG_FILE` env var → `warmdesk.yaml` in working directory → built-in defaults.

Alternatively, use environment variables — they always override any config file.

---

## 4. Run

```bash
cd dist

# With config file (recommended)
WEB_DIR=./web ./warmdesk

# Or with environment variables only
PORT=8080 \
DB_DRIVER=sqlite \
DB_DSN=./warmdesk.db \
JWT_SECRET=your-secret-key \
ALLOWED_ORIGINS=https://yourdomain.com \
WEB_DIR=./web \
./warmdesk
```

Open the application at **http://localhost:8080** (or your configured port).

---

## 5. Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP port |
| `DB_DRIVER` | `sqlite` | `sqlite`, `mysql`, or `postgres` |
| `DB_DSN` | `./warmdesk.db` | Database connection string / file path |
| `DB_TLS_MODE` | *(off)* | `disable` / `require` / `verify-ca` / `verify-full` |
| `DB_TLS_CA_CERT` | *(empty)* | Path to CA certificate file |
| `DB_TLS_CERT` | *(empty)* | Path to client certificate (mTLS) |
| `DB_TLS_KEY` | *(empty)* | Path to client private key (mTLS) |
| `TLS_CERT` | *(empty)* | Path to server TLS certificate (enables HTTPS when set with `TLS_KEY`) |
| `TLS_KEY` | *(empty)* | Path to server TLS private key |
| `JWT_SECRET` | `change-me-in-production` | Token signing secret — **always change this** |
| `ALLOWED_ORIGINS` | `http://localhost:5173` | CORS allowed origins (`*` for any) |
| `WEB_DIR` | *(empty)* | Path to built frontend files (required in production) |
| `BASE_URL` | *(empty)* | Public base URL (e.g. `https://desk.example.com`) — sets the host shown in Swagger UI |

---

## 6. Database Options

### SQLite (default — zero configuration)

```bash
DB_DRIVER=sqlite
DB_DSN=./warmdesk.db
```

### PostgreSQL

```bash
DB_DRIVER=postgres
DB_DSN="host=localhost user=warmdesk password=secret dbname=warmdesk port=5432 sslmode=disable"
```

### MySQL / MariaDB

```bash
DB_DRIVER=mysql
DB_DSN="warmdesk:secret@tcp(localhost:3306)/warmdesk?charset=utf8mb4&parseTime=True&loc=Local"
```

The schema is created automatically on first start via GORM's AutoMigrate.

---

## 7. Running as a System Service (Linux)

A ready-to-use service file is provided at `deploy/warmdesk.service`.

```bash
# Create a dedicated user
sudo useradd -r -s /bin/false warmdesk

# Copy files
sudo mkdir -p /opt/warmdesk/data
sudo cp -r dist/. /opt/warmdesk/
sudo chown -R warmdesk:warmdesk /opt/warmdesk

# Edit the service file to set your JWT_SECRET and domain, then install
sudo cp deploy/warmdesk.service /etc/systemd/system/warmdesk.service
sudo systemctl daemon-reload
sudo systemctl enable --now warmdesk
sudo systemctl status warmdesk
```

---

## 8. Reverse Proxy

A ready-to-use configuration for each web server is provided in the `deploy/` directory.
Both configurations handle HTTP→HTTPS redirect, SSL termination, and WebSocket proxying.

### Nginx (`deploy/nginx.conf`)

```bash
sudo cp deploy/nginx.conf /etc/nginx/sites-available/warmdesk
# Edit the file: replace yourdomain.com and update SSL paths
sudo ln -s /etc/nginx/sites-available/warmdesk /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

Obtain a free SSL certificate (if needed):
```bash
sudo certbot --nginx -d yourdomain.com
```

### Apache (`deploy/apache.conf`)

```bash
# Enable required modules
sudo a2enmod proxy proxy_http proxy_wstunnel ssl headers rewrite

sudo cp deploy/apache.conf /etc/apache2/sites-available/warmdesk.conf
# Edit the file: replace yourdomain.com and update SSL paths
sudo a2ensite warmdesk
sudo systemctl reload apache2
```

Obtain a free SSL certificate (if needed):
```bash
sudo certbot --apache -d yourdomain.com
```

Set `ALLOWED_ORIGINS=https://yourdomain.com` in the systemd service environment.

---

## 9. Desktop Entry (Linux)

A `.desktop` file is provided at `deploy/warmdesk.desktop` so WarmDesk
appears in application menus on GNOME, KDE, and other freedesktop-compatible
desktops.

```bash
# Install the desktop file
sudo cp deploy/warmdesk.desktop /usr/share/applications/

# Install the icon
sudo mkdir -p /usr/share/icons/hicolor/scalable/apps
sudo cp frontend/public/logo.svg /usr/share/icons/hicolor/scalable/apps/warmdesk.svg

# Refresh caches
sudo update-desktop-database
sudo gtk-update-icon-cache /usr/share/icons/hicolor
```

---

## 10. First Admin Account

The **first account registered on a fresh database is automatically made a
global admin** — no database manipulation required. Register normally through
the web interface.

Once one admin exists, further users can be promoted through
**Admin → Users → Edit** in the web interface.

> **Recovery:** if you ever need to promote an account directly (e.g. after
> accidentally demoting the only admin):
>
> **SQLite**
> ```bash
> sqlite3 /opt/warmdesk/data/warmdesk.db \
>   "UPDATE users SET global_role='admin' WHERE username='yourname';"
> ```
>
> **PostgreSQL / MySQL**
> ```sql
> UPDATE users SET global_role = 'admin' WHERE username = 'yourname';
> ```

---

## 11. Development Mode

Run backend and frontend separately with hot-reloading:

```bash
# Terminal 1 — backend API server on :8080
make dev-backend

# Terminal 2 — frontend dev server on :5173
make dev-frontend
```

Open **http://localhost:5173** during development.

---

## 12. Updating

```bash
git pull
make build
# restart the service
sudo systemctl restart warmdesk
```

---

## 13. Distribution Package

To create a portable archive for deployment on another machine:

```bash
make build
tar -czf warmdesk-$(date +%Y%m%d).tar.gz -C dist .
```

Extract on the target machine:

```bash
tar -xzf warmdesk-*.tar.gz -C /opt/warmdesk
```

Then follow steps 3–7 above.

---

## 14. Desktop App Builds

The desktop apps are Tauri 2 wrappers around the same frontend. They require
Rust and the system libraries listed below in addition to the normal
prerequisites.

### System libraries

**Ubuntu / Debian**
```bash
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev libssl-dev \
  libsoup-3.0-dev libglib2.0-dev librsvg2-dev squashfs-tools
```

**RHEL / Fedora**
```bash
sudo dnf install gtk3-devel webkit2gtk4.1-devel openssl-devel \
  libsoup3-devel glib2-devel librsvg2-devel squashfs-tools
```

**macOS** — no extra libraries needed beyond Xcode Command Line Tools.

**Windows** — no extra libraries needed; the NSIS installer tool is
downloaded automatically by Tauri.

### Build

```bash
make appimage          # Linux — WarmDesk-vX.Y.Z-x86_64.AppImage
make dmg               # macOS — WarmDesk-vX.Y.Z-universal.dmg
make windows-installer # Windows — WarmDesk-vX.Y.Z-x64-setup.exe
```

Output is placed in `frontend/src-tauri/target/release/bundle/`.
