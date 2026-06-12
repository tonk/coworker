# Releasing WarmDesk

## Overview

Releases are cut from the `main` branch. Every release:

1. Updates `CHANGELOG.md` with a new `## vX.Y.Z — YYYY-MM-DD` section.
2. Updates `README.md` — the **Latest release** heading and bullet highlights.
3. Appends new-feature bullets to `what.md`.
4. Bumps version strings in `website/hugo.toml` (and the AsciiDoc attribute).
5. Runs `make sync-doc-revisions` so `docs/*.adoc` `:revnumber:` / `:revdate:` match `CHANGELOG.md`.
6. Bumps `ansible/galaxy.yml` version if anything under `ansible/` changed since the last tag.
7. Commits with message `chore: release vX.Y.Z — CHANGELOG, README, what.md`.
8. Creates an annotated tag `vX.Y.Z` and pushes branch + tag.

The **Git tag** drives the build: `make` and CI use
`git describe --tags --match 'v*'` to embed the version string in the Go
binary via `-ldflags "-X main.version=..."`.

---

## Primary workflow — Claude Code `/release` skill

The recommended approach is to invoke the `/release` skill in Claude Code. Run:

```
/release v0.10.30
```

Claude Code will:
- Collect all commits since the last tag (`git log --oneline vX.Y.W..HEAD`).
- Write a user-facing `CHANGELOG.md` entry (Added / Fixed / Changed sections).
- Update the **Latest release** heading and bullets in `README.md`.
- Append bullets to `what.md`.
- Update `website/hugo.toml` (`warmdesk_version`, `release_date`, `warmdesk-version`).
- Bump `ansible/galaxy.yml` if Ansible files changed.
- Commit, tag, and push.

---

## Manual workflow — `./release` script

The repo root `./release` script automates the mechanical version bumps.
Use it when releasing outside Claude Code.

### Commands reference

```bash
# Show help
./release help

# Bump version strings only (usual workflow)
# Edit CHANGELOG.md and README.md bullets first, then:
./release bump v0.10.30

# Dry run (no file writes)
./release --dry-run bump v0.10.30

# One-shot: bump + commit + tag + push
./release publish v0.10.30
```

**Files touched by `./release bump`:**

| File | What changes |
|------|-------------|
| `website/hugo.toml` | `params.warmdesk_version`, `params.release_date`, `markup.asciidocext.attributes["warmdesk-version"]` |
| `frontend/src-tauri/tauri.conf.json` | `"version"` (semver without `v` prefix) |
| `frontend/src-tauri/Cargo.toml` | `version = "x.y.z"` |
| `README.md` | Only the `## Latest release (vX.Y.Z)` heading line |

**Not modified by the script** (update manually):

- `CHANGELOG.md` — write by hand before running the script
- `what.md` — append new-feature bullets by hand
- `frontend/package.json` — independent; not used for versioning
- Go source — version comes from linker flags at build time

### Manual git commands (reference)

```bash
V=v0.10.30
D=$(date -u +%Y-%m-%d)

# After editing CHANGELOG.md, README.md bullets, and what.md:
./release bump $V

git add CHANGELOG.md README.md what.md website/hugo.toml ansible/galaxy.yml
git commit -m "chore: release $V — CHANGELOG, README, what.md"
git tag -a "$V" -m "Release $V"
git push origin main
git push origin "$V"
```

---

## CI / GitHub Actions

Pushing a `v*` tag triggers the release workflow (see
`.github/workflows/release.yml`), which builds all artefacts (Linux amd64 +
arm64 tarballs, AppImage, deb, rpm, Windows installer, macOS DMG), signs them,
and publishes the GitHub Release page.

Ensure `CHANGELOG.md` and `README.md` are committed **before** pushing the tag.

---

## Checklist

- [ ] `CHANGELOG.md` — new section with user-facing notes
- [ ] `README.md` — **Latest release** bullets updated
- [ ] `what.md` — new-feature bullets appended
- [ ] `website/hugo.toml` — version and date bumped
- [ ] `make sync-doc-revisions` — `:revnumber:` / `:revdate:` in `docs/*.adoc`
- [ ] `make docs-pdf-guides` — user/admin guide PDFs (included in `make build`; `make docs-pdf` is the same)
- [ ] Avatar menu → **Downloads** — verify user and admin guide PDFs download in browser and Tauri
- [ ] `ansible/galaxy.yml` — version bumped (if Ansible files changed)
- [ ] Commit message: `chore: release vX.Y.Z — CHANGELOG, README, what.md`
- [ ] Annotated tag `vX.Y.Z` on `main`
- [ ] Branch and tag pushed; CI workflow succeeds
