BINARY    := warmdesk
DIST_DIR  := dist
DIST_ARM64 := $(DIST_DIR)/arm64
BACKEND   := backend
FRONTEND  := frontend
VERSION   := $(shell git describe --tags --always --match 'v*' 2>/dev/null || echo "dev")
ARCHIVE   := warmdesk-$(VERSION).tar.gz
.PHONY: all build build-frontend embed-web build-backend build-arm64 build-backend-arm64 clean dev-backend dev-frontend run package stamp-desktop-version appimage appimage-arm64 _fetch-gst-plugin deb deb-arm64 rpm rpm-arm64 dmg windows-installer windows-portable docs-commit sync-doc-revisions docs-pdf docs-pdf-guides docs-pdf-assets test test-backend test-frontend screenshots screenshots-dev help

help:
	@echo "WarmDesk $(VERSION)"
	@echo ""
	@echo "Development"
	@echo "  dev-backend          Start the Go backend in development mode (localhost:8080)"
	@echo "  dev-frontend         Start the Vite dev server (localhost:5173, proxies /api to :8080)"
	@echo ""
	@echo "Server builds  →  dist/"
	@echo "  build                Build frontend + backend (x86_64) into dist/"
	@echo "  build-arm64          Build frontend + backend (arm64)  into dist/arm64/"
	@echo "  run                  build then run the production binary locally"
	@echo "  package              build then create a dist tarball (warmdesk-<version>.tar.gz)"
	@echo "  clean                Remove dist/, build artifacts and Tauri target directory"
	@echo "  docs-commit          Stage docs/website files and create a commit (set MSG=... and BODY=...)"
	@echo "  sync-doc-revisions   Set :revnumber:/:revdate: in docs/*.adoc from CHANGELOG.md"
	@echo "  docs-pdf-guides      Build user-guide.pdf and admin-guide.pdf (included in make build)"
	@echo "  docs-pdf             Alias for docs-pdf-guides"
	@echo "  docs-pdf-assets      Regenerate docs/pdf-theme/warmdesk-title-logo.png from SVG"
	@echo ""
	@echo "Linux desktop  (requires Rust + webkit2gtk4.1-devel, gtk3-devel, librsvg2-devel, openssl-devel)"
	@echo "  appimage             AppImage (x86_64)"
	@echo "  appimage-arm64       AppImage (arm64)  — also needs gcc-aarch64-linux-gnu + arm64 webkit2gtk"
	@echo "  deb                  Debian/Ubuntu .deb package (x86_64)"
	@echo "  deb-arm64            Debian/Ubuntu .deb package (arm64)"
	@echo "  rpm                  Fedora/RHEL .rpm package (x86_64)"
	@echo "  rpm-arm64            Fedora/RHEL .rpm package (arm64)"
	@echo ""
	@echo "macOS desktop  (must run on macOS — requires Rust + Xcode command line tools)"
	@echo "  dmg                  Universal DMG (Intel + Apple Silicon)"
	@echo ""
	@echo "Windows desktop  (must run on Windows — requires Rust + WebView2)"
	@echo "  windows-installer    NSIS installer (.exe)"
	@echo "  windows-portable     Portable zip (no installation needed)"
	@echo ""
	@echo "Tests"
	@echo "  test                 Run all backend + frontend tests"
	@echo "  test-backend         Run Go tests only"
	@echo "  test-frontend        Run Vitest tests only"
	@echo ""
	@echo "Screenshots  (requires Go, Node.js, chromium)"
	@echo "  screenshots          Full run: seed DB, start servers, capture screenshots, clean up"
	@echo "  screenshots-dev      Run Playwright against already-running servers"

# Build everything into dist/
all: build

build: build-frontend docs-pdf-guides embed-web build-backend
	@cp warmdesk.yaml.example $(DIST_DIR)/warmdesk.yaml.example
	@cp warmdesk-migrate.yaml.example $(DIST_DIR)/warmdesk-migrate.yaml.example
	@cp -r deploy $(DIST_DIR)/deploy
	@cp deploy/get_warmdesk $(DIST_DIR)/get_warmdesk
	@cp INSTALL.md $(DIST_DIR)/INSTALL.md
	@cp README.md $(DIST_DIR)/README.md
	@rm -rf $(DIST_DIR)/docs && cp -r docs $(DIST_DIR)/docs
	@echo "Build complete. Output: $(DIST_DIR)/"

build-frontend:
	@echo "Building frontend..."
	cd $(FRONTEND) && npm install && npm run build

# Copy the Vite build output into the embed source directory so go build -tags embed
# picks it up and bakes the frontend into the warmdesk binary.
embed-web:
	@echo "Embedding web files into binary..."
	@find $(BACKEND)/staticweb/files -mindepth 1 -not -name 'placeholder' -delete 2>/dev/null || true
	cp -r $(FRONTEND)/dist/. $(BACKEND)/staticweb/files/
	@mkdir -p $(BACKEND)/staticweb/files/docs
	@cp docs/user-guide.pdf docs/admin-guide.pdf $(BACKEND)/staticweb/files/docs/

build-backend:
	@echo "Building backend..."
	mkdir -p $(DIST_DIR)
	cd $(BACKEND) && go build -tags embed -ldflags="-s -w -X main.version=$(VERSION)" -o ../$(DIST_DIR)/$(BINARY) .
	cd $(BACKEND) && go build -ldflags="-s -w -X main.version=$(VERSION)" -o ../$(DIST_DIR)/$(BINARY)-seed ./cmd/seed
	cd $(BACKEND) && go build -ldflags="-s -w" -o ../$(DIST_DIR)/$(BINARY)-export ./cmd/export
	cd $(BACKEND) && go build -ldflags="-s -w" -o ../$(DIST_DIR)/$(BINARY)-import ./cmd/importer
	cd $(BACKEND) && go build -ldflags="-s -w -X main.version=$(VERSION)" -o ../$(DIST_DIR)/$(BINARY)-training ./cmd/training
	cd $(BACKEND) && go build -ldflags="-s -w" -o ../$(DIST_DIR)/$(BINARY)-db-convert ./cmd/db-convert

# Build everything for linux/arm64 (server + embedded web assets).
# No C cross-compiler required — the SQLite driver (glebarez/sqlite) is pure Go.
build-arm64: build-frontend docs-pdf-guides embed-web build-backend-arm64
	@cp warmdesk.yaml.example $(DIST_ARM64)/warmdesk.yaml.example
	@cp warmdesk-migrate.yaml.example $(DIST_ARM64)/warmdesk-migrate.yaml.example
	@cp -r deploy $(DIST_ARM64)/deploy
	@cp deploy/get_warmdesk $(DIST_ARM64)/get_warmdesk
	@cp INSTALL.md $(DIST_ARM64)/INSTALL.md
	@cp README.md $(DIST_ARM64)/README.md
	@rm -rf $(DIST_ARM64)/docs && cp -r docs $(DIST_ARM64)/docs
	@echo "ARM64 server build complete. Output: $(DIST_ARM64)/"

build-backend-arm64:
	@echo "Building backend (linux/arm64)..."
	mkdir -p $(DIST_ARM64)
	cd $(BACKEND) && GOOS=linux GOARCH=arm64 go build -tags embed -ldflags="-s -w -X main.version=$(VERSION)" -o ../$(DIST_ARM64)/$(BINARY) .
	cd $(BACKEND) && GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o ../$(DIST_ARM64)/$(BINARY)-seed ./cmd/seed
	cd $(BACKEND) && GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ../$(DIST_ARM64)/$(BINARY)-export ./cmd/export
	cd $(BACKEND) && GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ../$(DIST_ARM64)/$(BINARY)-import ./cmd/importer
	cd $(BACKEND) && GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o ../$(DIST_ARM64)/$(BINARY)-training ./cmd/training
	cd $(BACKEND) && GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ../$(DIST_ARM64)/$(BINARY)-db-convert ./cmd/db-convert


# Run in development mode (two terminals needed)
dev-backend:
	cd $(BACKEND) && go run .

dev-frontend:
	cd $(FRONTEND) && npm run dev

# Create distribution archive (run after build)
package: build
	@echo "Creating distribution archive $(ARCHIVE)..."
	@tar -czf $(ARCHIVE) -C $(DIST_DIR) .
	@echo "Distribution package: $(ARCHIVE)"

# Stamp version into tauri.conf.json and Cargo.toml from the current git tag.
stamp-desktop-version:
	@node -e "\
		const fs = require('fs');\
		const ver = '$(VERSION)'.replace(/^v/, '');\
		const tp = '$(FRONTEND)/src-tauri/tauri.conf.json';\
		const tc = JSON.parse(fs.readFileSync(tp, 'utf8'));\
		tc.version = ver;\
		fs.writeFileSync(tp, JSON.stringify(tc, null, 2) + '\n');\
		const cp = '$(FRONTEND)/src-tauri/Cargo.toml';\
		let cargo = fs.readFileSync(cp, 'utf8');\
		cargo = cargo.replace(/^version = \"[^\"]*\"/m, 'version = \"' + ver + '\"');\
		fs.writeFileSync(cp, cargo);\
		console.log('Stamped desktop version:', ver);\
	"

# URL for the linuxdeploy GStreamer plugin script (bundles audio/video capture plugins).
LINUXDEPLOY_PLUGIN_GST_URL := https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gstreamer/master/linuxdeploy-plugin-gstreamer.sh
# Cached to /tmp so repeated builds in one session skip the download.
LINUXDEPLOY_PLUGIN_GST     := /tmp/linuxdeploy-plugin-gstreamer

# Fetch linuxdeploy-plugin-gstreamer if not already present.
# linuxdeploy auto-discovers executables named linuxdeploy-plugin-* in PATH and
# runs them, so placing the script here and prepending /tmp to PATH is enough.
.PHONY: _fetch-gst-plugin
_fetch-gst-plugin:
	@if [ ! -x $(LINUXDEPLOY_PLUGIN_GST) ]; then \
		echo "  Downloading linuxdeploy-plugin-gstreamer..."; \
		curl -fsSL -o $(LINUXDEPLOY_PLUGIN_GST) "$(LINUXDEPLOY_PLUGIN_GST_URL)" \
		&& chmod +x $(LINUXDEPLOY_PLUGIN_GST) \
		|| { echo "ERROR: failed to download linuxdeploy-plugin-gstreamer"; exit 1; }; \
	fi

# Build the Tauri desktop client as an AppImage (Linux).
# Requires: Rust, webkit2gtk4.1-devel, gtk3-devel, librsvg2-devel, openssl-devel
# Must be built on Ubuntu 24.04: the AppImage bundles webkit2gtk + GStreamer core
# from the build host.  linuxdeploy-plugin-gstreamer then adds the matching
# GStreamer plugins so getUserMedia / camera-mic selection works on all distros
# (Fedora's GStreamer is a different major version and cannot be mixed in).
# NO_STRIP=true works around linuxdeploy's bundled strip being too old for newer glibc.
appimage: stamp-desktop-version _fetch-gst-plugin
	@echo "Building WarmDesk desktop app (AppImage)..."
	cd $(FRONTEND) && NO_STRIP=true PATH="$$PATH:/tmp" npm run tauri:build -- --bundles appimage
	@echo "AppImage: $(FRONTEND)/src-tauri/target/release/bundle/appimage/WarmDesk_*_amd64.AppImage"

# Build the Tauri desktop client as an AppImage (Linux, arm64).
# Requires (host): Rust aarch64-unknown-linux-gnu target, gcc-aarch64-linux-gnu cross-compiler,
#   arm64 webkit2gtk + GTK3 dev libraries (via sysroot or running natively on an arm64 host).
# Fedora/RHEL: dnf install gcc-aarch64-linux-gnu
# Debian/Ubuntu: apt install gcc-aarch64-linux-gnu
# The linker is configured in frontend/src-tauri/.cargo/config.toml.
appimage-arm64: stamp-desktop-version _fetch-gst-plugin
	@echo "Building WarmDesk desktop app (AppImage, arm64)..."
	rustup target add aarch64-unknown-linux-gnu 2>/dev/null || true
	cd $(FRONTEND) && NO_STRIP=true PATH="$$PATH:/tmp" npm run tauri:build -- --target aarch64-unknown-linux-gnu --bundles appimage
	@echo "AppImage: $(FRONTEND)/src-tauri/target/aarch64-unknown-linux-gnu/release/bundle/appimage/WarmDesk_*_aarch64.AppImage"

# Build the Tauri desktop client as a .deb package (Debian/Ubuntu).
# Requires: Rust, dpkg, webkit2gtk4.1-devel, gtk3-devel, librsvg2-devel, openssl-devel
deb: stamp-desktop-version
	@echo "Building WarmDesk desktop app (.deb)..."
	cd $(FRONTEND) && npm run tauri:build -- --bundles deb
	@echo "Package: $(FRONTEND)/src-tauri/target/release/bundle/deb/"

# Build the Tauri desktop client as a .deb package (Debian/Ubuntu, arm64).
# Same cross-compilation prerequisites as appimage-arm64.
deb-arm64: stamp-desktop-version
	@echo "Building WarmDesk desktop app (.deb, arm64)..."
	rustup target add aarch64-unknown-linux-gnu 2>/dev/null || true
	cd $(FRONTEND) && npm run tauri:build -- --target aarch64-unknown-linux-gnu --bundles deb
	@echo "Package: $(FRONTEND)/src-tauri/target/aarch64-unknown-linux-gnu/release/bundle/deb/"

# Build the Tauri desktop client as an .rpm package (Fedora/RHEL).
# Requires: Rust, rpm-build, webkit2gtk4.1-devel, gtk3-devel, librsvg2-devel, openssl-devel
rpm: stamp-desktop-version
	@echo "Building WarmDesk desktop app (.rpm)..."
	cd $(FRONTEND) && npm run tauri:build -- --bundles rpm
	@echo "Package: $(FRONTEND)/src-tauri/target/release/bundle/rpm/"

# Build the Tauri desktop client as an .rpm package (Fedora/RHEL, arm64).
# Same cross-compilation prerequisites as appimage-arm64.
rpm-arm64: stamp-desktop-version
	@echo "Building WarmDesk desktop app (.rpm, arm64)..."
	rustup target add aarch64-unknown-linux-gnu 2>/dev/null || true
	cd $(FRONTEND) && npm run tauri:build -- --target aarch64-unknown-linux-gnu --bundles rpm
	@echo "Package: $(FRONTEND)/src-tauri/target/aarch64-unknown-linux-gnu/release/bundle/rpm/"

# Build the Tauri desktop client as a macOS DMG (universal: Intel + Apple Silicon).
# Must be run on macOS. Requires: Rust, Xcode command line tools.
dmg: stamp-desktop-version
	@echo "Building WarmDesk desktop app (macOS DMG)..."
	rustup target add aarch64-apple-darwin x86_64-apple-darwin 2>/dev/null || true
	cd $(FRONTEND) && npm run tauri:build -- --bundles dmg --target universal-apple-darwin
	@echo "DMG: $(FRONTEND)/src-tauri/target/universal-apple-darwin/release/bundle/dmg/WarmDesk_*.dmg"

# Build the Tauri desktop client as a Windows NSIS installer.
# Must be run on Windows. Requires: Rust, WebView2 (pre-installed on Windows 11).
windows-installer: stamp-desktop-version
	@echo "Building WarmDesk desktop app (Windows installer)..."
	cd $(FRONTEND) && npm run tauri:build -- --bundles nsis
	@echo "Installer: $(FRONTEND)/src-tauri/target/release/bundle/nsis/WarmDesk_*_x64-setup.exe"

# Build the Tauri desktop client as a portable Windows zip — extract and run, no installation.
# Must be run on Windows. Requires: Rust, WebView2 (pre-installed on Windows 11).
# WebView2 is pre-installed on Windows 10 (2018+) and Windows 11.
windows-portable: stamp-desktop-version
	@echo "Building WarmDesk desktop app (Windows portable zip)..."
	cd $(FRONTEND) && npm run tauri:build -- --bundles nsis
	powershell -Command "Compress-Archive -Path '$(FRONTEND)/src-tauri/target/release/WarmDesk.exe' -DestinationPath 'WarmDesk-portable.zip' -Force"
	@echo "Portable zip: WarmDesk-portable.zip"

# Run all tests
test: test-backend test-frontend

# Regenerate all screenshots using Playwright (seeds DB, starts servers, runs specs)
screenshots:
	@echo "Regenerating screenshots..."
	cd $(FRONTEND) && npm run screenshots

# Run Playwright screenshot specs against already-running servers (dev mode)
screenshots-dev:
	@echo "Running Playwright screenshots (servers must be running)..."
	cd $(FRONTEND) && npm run screenshots:dev

# Run Go tests
test-backend:
	@echo "Running Go tests..."
	cd $(BACKEND) && go test ./...

# Run frontend tests
test-frontend:
	@echo "Running frontend tests..."
	cd $(FRONTEND) && npm test

# Remove build artifacts
clean:
	rm -rf $(DIST_DIR) warmdesk-*.tar.gz
	rm -rf $(FRONTEND)/dist $(FRONTEND)/src-tauri/target

# Build and run production binary locally (web UI is embedded in the binary)
run: build
	cd $(DIST_DIR) && ./$(BINARY)

# Sync :revnumber: / :revdate: in docs/*.adoc from the latest CHANGELOG.md release header.
sync-doc-revisions:
	@./scripts/sync-doc-revisions.sh

# Stage docs/website updates and create a formatted commit.
# Usage:
#   make docs-commit MSG="docs: update group video docs" BODY="Add LiveKit docs and blog item."
docs-commit: sync-doc-revisions
	@./scripts/commit-docs-website.sh "$(MSG)" "$(BODY)"

DOCS_GUIDE_PDF_SOURCES := user-guide admin-guide

# Regenerate the title-page logo PNG (source: docs/pdf-theme/warmdesk-title-logo.svg).
docs-pdf-assets:
	@command -v inkscape >/dev/null 2>&1 || { echo "inkscape is required to build docs/pdf-theme/warmdesk-title-logo.png"; exit 1; }
	@mkdir -p docs/pdf-theme/fonts
	cp backend/handlers/fonts/FreeSans.ttf backend/handlers/fonts/FreeSansBold.ttf \
		backend/handlers/fonts/SourceCodePro-Regular.ttf backend/handlers/fonts/SourceCodePro-Bold.ttf \
		docs/pdf-theme/fonts/
	inkscape docs/pdf-theme/warmdesk-title-logo.svg \
		--export-type=png \
		--export-filename=docs/pdf-theme/warmdesk-title-logo.png \
		-w 900

# Build user/admin guide PDFs (included in server distribution).
docs-pdf-guides: sync-doc-revisions docs-pdf-assets
	@command -v asciidoctor-pdf >/dev/null 2>&1 || { echo "asciidoctor-pdf is required (gem install asciidoctor-pdf)"; exit 1; }
	@set -e; for src in $(DOCS_GUIDE_PDF_SOURCES); do \
		echo "Building docs/$${src}.pdf …"; \
		cd docs && asciidoctor-pdf -o "$${src}.pdf" "$${src}.adoc" && cd ..; \
	done

# Build all branded PDFs (same as docs-pdf-guides).
docs-pdf: docs-pdf-guides
