You are performing a WarmDesk release. The version to release is: $ARGUMENTS

Follow these steps in order:

## 0. Verify the version number follows semver

WarmDesk follows real semver discipline as of v0.13.0 (see the "Versioning" section in CLAUDE.md). Before proceeding, classify every commit since the last tag (`git log --oneline $(git describe --tags --abbrev=0)..HEAD`) as one of:

- **fix** — bug fix only, no new capability
- **feat** — new backwards-compatible feature or additive surface (new endpoint, new config key with a default, new UI capability)
- **breaking** — removed/renamed config or API surface, a DB change that isn't safely reversible, or a changed response shape

Then confirm `$ARGUMENTS` matches what that classification requires, within the current `0.MINOR.PATCH` range:
- All commits are **fix**-only → PATCH bump (e.g. `0.13.0` → `0.13.1`)
- Any commit is **feat** → MINOR bump, PATCH resets to 0 (e.g. `0.13.4` → `0.14.0`)
- Any commit is **breaking** → this needs a deliberate decision (MAJOR / the `1.0.0` milestone), not a routine bump — stop and confirm with the user rather than guessing

If `$ARGUMENTS` doesn't match the required bump, say so and confirm the correct version before continuing — don't silently proceed with a mismatched number.

Releases at or before v0.12.42 used a flat incrementing counter regardless of change type; don't use that range as a precedent for what kind of bump a given change "usually" gets.

## 1. Gather changes
Run `git log --oneline $(git describe --tags --abbrev=0)..HEAD` to see all commits since the last tag. Use these to write the release notes.

## 2. Update CHANGELOG.md
Add a new section at the top (below the `# Changelog` heading) in this format:

```
## v{version} — {today's date YYYY-MM-DD}

### Added
- ...

### Fixed
- ...

### Changed
- ...
```

Only include sections that have entries. Be specific and user-facing in the descriptions — same style as existing entries.

## 3. Update README.md
- Update the `## Latest release (v{old_version})` heading to `## Latest release (v{version})` and replace the bullet points beneath it with the highlights of this release (one to three short bullets, user-facing, same style as the existing entries).
- If any new features were added, update the **Features** list to reflect them.
- If the seed tool changed, update the **Load demo data** section.
- No other sections need changing for a routine release.

## 4. Update what.md
Append any new features or changes as bullet points at the end of the file, matching the imperative style already used there.

## 5. Update documentation
- If any changes where made, that need an update on the documentation,
  update the documentation
- If a change touches anything a multi-instance deployment depends on —
  shared config keys (`db_driver`/`db_dsn`, `redis_url`, `upload_dir`,
  `jwt_secret`, `trusted_proxies`), the WebSocket/Redis pub-sub layer, the
  auth rate limiter, or the backup scheduler — update
  `deploy/multi-instance/README.md` (and its templates) to match.

## 6. Update website version
In `website/hugo.toml`, update three values:
- Under `[params]`: `warmdesk_version = "v{version}"` and `release_date = "{today's date YYYY-MM-DD}"`
- Under `[markup.asciidocext.attributes]`: `"warmdesk-version" = "v{version}"`

The params feed the homepage release strip; the AsciiDoc attribute feeds the install docs code blocks.

## 7. Bump Ansible collection version
In `ansible/galaxy.yml`, increment the `version` field by one patch level (e.g. `0.3.1` → `0.3.2`).
Only do this if any commits since the last tag touched files under `ansible/`.

## 8. Sync doc revision headers
Run `./scripts/sync-doc-revisions.sh` (or `make sync-doc-revisions`) to update `:revnumber:`/`:revdate:` in `docs/admin-guide.adoc` and `docs/user-guide.adoc` from the CHANGELOG entry just added. Do this *after* step 2 (CHANGELOG.md) — the script reads the latest release header from it. Easy to forget since it's a script step rather than a direct edit like the others; skipping it leaves both guides' PDFs claiming the previous release's version.

## 9. Commit and tag
```bash
git add CHANGELOG.md README.md what.md website/hugo.toml ansible/galaxy.yml docs/admin-guide.adoc docs/user-guide.adoc
git commit -m "chore: release v{version} — CHANGELOG, README, what.md\n\nCo-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
git tag -a v{version} -m "Release v{version}"
```

## 10. Push
```bash
git push && git push --tags
```

Report what was pushed when done.
