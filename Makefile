BINARY   := warmdesk
DIST_DIR := dist
BACKEND  := backend
FRONTEND := frontend
VERSION  := $(shell git describe --tags --always --match 'v*' 2>/dev/null || echo "dev")
ARCHIVE  := warmdesk-$(VERSION).tar.gz
.PHONY: all build build-frontend embed-web build-backend clean dev-backend dev-frontend run package stamp-desktop-version appimage deb rpm dmg windows-installer windows-portable

# Build everything into dist/
all: build

build: build-frontend embed-web build-backend
	@cp warmdesk.yaml.example $(DIST_DIR)/warmdesk.yaml.example
	@cp warmdesk-migrate.yaml.example $(DIST_DIR)/warmdesk-migrate.yaml.example
	@cp -r deploy $(DIST_DIR)/deploy
	@cp INSTALL.md $(DIST_DIR)/INSTALL.md
	@cp README.md $(DIST_DIR)/README.md
	@cp -r docs $(DIST_DIR)/docs
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

build-backend:
	@echo "Building backend..."
	mkdir -p $(DIST_DIR)
	cd $(BACKEND) && go build -tags embed -ldflags="-s -w -X main.version=$(VERSION)" -o ../$(DIST_DIR)/$(BINARY) .
	cd $(BACKEND) && go build -ldflags="-s -w" -o ../$(DIST_DIR)/$(BINARY)-seed ./cmd/seed
	cd $(BACKEND) && go build -ldflags="-s -w" -o ../$(DIST_DIR)/$(BINARY)-export ./cmd/export
	cd $(BACKEND) && go build -ldflags="-s -w" -o ../$(DIST_DIR)/$(BINARY)-import ./cmd/importer
	cd $(BACKEND) && go build -ldflags="-s -w" -o ../$(DIST_DIR)/$(BINARY)-training ./cmd/training


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

# Build the Tauri desktop client as an AppImage (Linux).
# Requires: Rust, webkit2gtk4.1-devel, gtk3-devel, librsvg2-devel, openssl-devel
# NO_STRIP=true works around linuxdeploy's bundled strip being too old for newer glibc.
appimage: stamp-desktop-version
	@echo "Building WarmDesk desktop app (AppImage)..."
	cd $(FRONTEND) && NO_STRIP=true npm run tauri:build -- --bundles appimage
	@echo "AppImage: $(FRONTEND)/src-tauri/target/release/bundle/appimage/WarmDesk_*_amd64.AppImage"

# Build the Tauri desktop client as a .deb package (Debian/Ubuntu).
# Requires: Rust, dpkg, webkit2gtk4.1-devel, gtk3-devel, librsvg2-devel, openssl-devel
deb: stamp-desktop-version
	@echo "Building WarmDesk desktop app (.deb)..."
	cd $(FRONTEND) && npm run tauri:build -- --bundles deb
	@echo "Package: $(FRONTEND)/src-tauri/target/release/bundle/deb/"

# Build the Tauri desktop client as an .rpm package (Fedora/RHEL).
# Requires: Rust, rpm-build, webkit2gtk4.1-devel, gtk3-devel, librsvg2-devel, openssl-devel
rpm: stamp-desktop-version
	@echo "Building WarmDesk desktop app (.rpm)..."
	cd $(FRONTEND) && npm run tauri:build -- --bundles rpm
	@echo "Package: $(FRONTEND)/src-tauri/target/release/bundle/rpm/"

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

# Remove build artifacts
clean:
	rm -rf $(DIST_DIR) warmdesk-*.tar.gz
	rm -rf $(FRONTEND)/dist $(FRONTEND)/src-tauri/target

# Build and run production binary locally (web UI is embedded in the binary)
run: build
	cd $(DIST_DIR) && ./$(BINARY)
