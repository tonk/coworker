# Changelog

All notable changes to Coworker are documented here.

## Unreleased

### Added
- Configurable idle session timeout (admin setting, default 1 hour); users are automatically logged out after the configured period of inactivity
- Update check against GitHub releases on login; a dismissable banner is shown when a newer version is available (web and desktop)

## 2026-03-28

### Added
- Tauri desktop app — distributable as AppImage (Linux), DMG (macOS), and installer (Windows)
- Topics — threaded discussions per project with markdown support and replies
- Checklists on cards
- Multiple assignees per card
- Viewer role — read-only access at project and global level
- Favourite people — mark users for quick access
- Card watchers — subscribe to card activity notifications
- Card sorting within columns by due date, assignee, or priority
- Direct message notifications
- Group direct messages
- Admin can assign users to projects directly
- Admin can reset user passwords

### Fixed
- Topics view was rendering its own header, causing duplicate search bar, language selector, and avatar
- Adding a new card showed it twice until page refresh (duplicate WebSocket event handling)
- Logo and favicon not served correctly
- Build artifacts (AppImage, DMG, Windows installer, Rust target/) excluded from git via .gitignore

### Changed
- Group DMs, markdown in chat, i18n expansion, and UI polish
