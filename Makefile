BINARY   := coworker
DIST_DIR := dist
BACKEND  := backend
FRONTEND := frontend
VERSION  := $(shell git -C $(BACKEND) describe --tags --always 2>/dev/null || echo "dev")
ARCHIVE  := coworker-$(VERSION).tar.gz

.PHONY: all build build-frontend build-backend clean dev-backend dev-frontend run package appimage dmg windows-installer

# Build everything into dist/
all: build

build: build-frontend build-backend
	@cp coworker.yaml.example $(DIST_DIR)/coworker.yaml.example
	@cp -r deploy $(DIST_DIR)/deploy
	@cp INSTALL.md $(DIST_DIR)/INSTALL.md
	@cp README.md $(DIST_DIR)/README.md
	@echo "Build complete. Output: $(DIST_DIR)/"

build-frontend:
	@echo "Building frontend..."
	cd $(FRONTEND) && npm install && npm run build
	mkdir -p $(DIST_DIR)/web
	cp -r $(FRONTEND)/dist/. $(DIST_DIR)/web/

build-backend:
	@echo "Building backend..."
	mkdir -p $(DIST_DIR)
	cd $(BACKEND) && go build -ldflags="-s -w" -o ../$(DIST_DIR)/$(BINARY) .

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

# Build the Tauri desktop client as an AppImage (Linux).
# Requires: Rust, webkit2gtk4.1-devel, gtk3-devel, librsvg2-devel, openssl-devel
# NO_STRIP=true works around linuxdeploy's bundled strip being too old for newer glibc.
appimage:
	@echo "Building Coworker desktop app (AppImage)..."
	cd $(FRONTEND) && NO_STRIP=true npm run tauri:build -- --bundles appimage
	@echo "AppImage: $(FRONTEND)/src-tauri/target/release/bundle/appimage/Coworker_*_amd64.AppImage"

# Build the Tauri desktop client as a macOS DMG (universal: Intel + Apple Silicon).
# Must be run on macOS. Requires: Rust, Xcode command line tools.
dmg:
	@echo "Building Coworker desktop app (macOS DMG)..."
	rustup target add aarch64-apple-darwin x86_64-apple-darwin 2>/dev/null || true
	cd $(FRONTEND) && npm run tauri:build -- --bundles dmg --target universal-apple-darwin
	@echo "DMG: $(FRONTEND)/src-tauri/target/universal-apple-darwin/release/bundle/dmg/Coworker_*.dmg"

# Build the Tauri desktop client as a Windows NSIS installer.
# Must be run on Windows. Requires: Rust, WebView2 (pre-installed on Windows 11).
windows-installer:
	@echo "Building Coworker desktop app (Windows installer)..."
	cd $(FRONTEND) && npm run tauri:build -- --bundles nsis
	@echo "Installer: $(FRONTEND)/src-tauri/target/release/bundle/nsis/Coworker_*_x64-setup.exe"

# Remove build artifacts
clean:
	rm -rf $(DIST_DIR) coworker-*.tar.gz
	rm -rf $(FRONTEND)/dist $(FRONTEND)/src-tauri/target

# Build and run production binary locally
run: build
	cd $(DIST_DIR) && WEB_DIR=./web ./$(BINARY)
