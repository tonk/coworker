# Releasing WarmDesk

This document describes how we cut a versioned release. The repo root script **`./release`** automates the mechanical version bumps; **you still author** `CHANGELOG.md` and the bullet list under **Latest release** in `README.md`.

## What a release involves

| Step | What | Automated? |
|------|------|----------------|
| 1 | Add a `## vX.Y.Z — YYYY-MM-DD` section to `CHANGELOG.md` with user-facing notes | No — write by hand |
| 2 | Update **Latest release** bullets in `README.md` (short highlights for readers) | No — write by hand |
| 3 | Bump version strings in `website/hugo.toml`, Tauri `Cargo.toml` / `tauri.conf.json`, and the README heading line | Yes — `./release bump` |
| 4 | Refresh `frontend/src-tauri/Cargo.lock` after Tauri version change | Yes — `cargo metadata` inside `./release` |
| 5 | Commit, annotated tag `vX.Y.Z`, push `main` and the tag | Manual, or `./release publish` |

The **Git tag** drives the build: `make` and CI use `git describe --tags --match 'v*'` for embedding the version in the Go binary (`-ldflags "-X main.version=..."`).

Files touched by **`./release bump`**:

- `website/hugo.toml` — `params.warmdesk_version`, `params.release_date` (today’s date in UTC unless `RELEASE_DATE` is set), and `markup.asciidocext.attributes` **`warmdesk-version`** (AsciiDoc attribute `{warmdesk-version}` in the theme)
- `frontend/src-tauri/tauri.conf.json` — `"version"` (semver **without** the `v` prefix)
- `frontend/src-tauri/Cargo.toml` — `version = "x.y.z"`
- `README.md` — **only** the line `## Latest release (vX.Y.Z)`; bullets below are yours

Not modified by the script (by design):

- `CHANGELOG.md`
- `frontend/package.json` — npm package version is independent (frontend embed version comes from git / build)
- Go source — uses linker flags from git describe at build time

## Commands reference

### Show help

```bash
./release help
```

### Bump versions only (usual workflow)

After editing `CHANGELOG.md` and README bullets:

```bash
./release bump v0.9.40
# or equivalently:
./release v0.9.40
```

Dry run (no file writes):

```bash
./release --dry-run bump v0.9.40
```

Use a fixed date for the website (e.g. when documenting retroactively):

```bash
RELEASE_DATE=2026-05-08 ./release bump v0.9.35
```

### Sanity-check version strings

```bash
./release check
./release check v0.9.40
```

### One-shot: bump + commit + tag + push

Use only when you are ready to commit **everything** in the working tree (including your CHANGELOG/README edits):

```bash
./release publish v0.9.40
```

This runs `bump`, then `git add -A`, `git commit -m "Release v0.9.40"`, annotated tag `v0.9.40`, `git push origin main`, and `git push origin v0.9.40`.

If you prefer a manual commit message or staged files only, run **`./release bump`** and then git yourself.

## Equivalent manual commands (historical)

These are roughly what the automation does; you do **not** need to run them if you use `./release bump`:

```bash
# From repository root; example tag v0.9.40, date 2026-05-10
V=v0.9.40
D=2026-05-10
BARE=0.9.40

# Edit website/hugo.toml (warmdesk_version, release_date, warmdesk-version attribute)
# Edit frontend/src-tauri/tauri.conf.json and Cargo.toml version to $BARE
# Edit README.md heading to "## Latest release ($V)"

( cd frontend/src-tauri && cargo metadata --format-version 1 > /dev/null )

git add -A
git commit -m "Release $V"
git tag -a "$V" -m "WarmDesk $V"
git push origin main
git push origin "$V"
```

## CI / GitHub Actions

Pushing tag **`v*`** triggers the release workflow (artifacts, signing, GitHub Release page, etc.—see `.github/workflows/release.yml`). Ensure `CHANGELOG.md` and README are committed **before** the tag push.

## Checklist

- [ ] `CHANGELOG.md` section for this tag
- [ ] `README.md` bullets under Latest release
- [ ] `./release bump vX.Y.Z` (or `publish` if you want one shot)
- [ ] `git diff` review
- [ ] Tag on `main` with matching `vX.Y.Z`
- [ ] Push branch + tag; confirm workflow success
