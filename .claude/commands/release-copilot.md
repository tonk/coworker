# Release (Copilot CLI style)

This document converts the original release instructions into concise, runnable commands and a Copilot-friendly workflow.

Usage (manual):

0. Verify the version number follows semver (see "Versioning" in CLAUDE.md, effective v0.13.0)
   - Classify every commit since the last tag as fix / feat / breaking.
   - All fixes → PATCH bump. Any feat → MINOR bump (PATCH resets to 0). Any breaking change → stop and confirm with the user; don't guess a MAJOR/1.0.0 bump on your own.
   - Releases at or before v0.12.42 used a flat incrementing counter regardless of change type — don't treat that range as precedent.

1. Gather changes
   git log --oneline $(git describe --tags --abbrev=0)..HEAD
   Use the output to write user-facing release notes.

2. Update CHANGELOG.md
   Add a new section below `# Changelog` in this exact format (omit empty sections):

```
## v{version} — {YYYY-MM-DD}

### Added
- ...

### Fixed
- ...

### Changed
- ...
```

3. Update README.md
- Update the Features or Load demo data sections if relevant.

4. Update what.md
- Append bullet points describing new features/changes in imperative style.

5. Update documentation
- Edit any docs that require updates.
- If a change touches anything a multi-instance deployment depends on — shared config keys (`db_driver`/`db_dsn`, `redis_url`, `upload_dir`, `jwt_secret`, `trusted_proxies`), the WebSocket/Redis pub-sub layer, the auth rate limiter, or the backup scheduler — update `deploy/multi-instance/README.md` (and its templates) to match.

6. Update website version
- In `website/hugo.toml`, update `warmdesk_version = "v{version}"` and `release_date = "{YYYY-MM-DD}"` under `[params]`, and `"warmdesk-version" = "v{version}"` under `[markup.asciidocext.attributes]`.

7. Bump Ansible collection version
- In `ansible/galaxy.yml`, increment `version` by one patch level (e.g. `0.3.1` → `0.3.2`). Only if commits since the last tag touched files under `ansible/`.

8. Sync doc revision headers
- Run `./scripts/sync-doc-revisions.sh` (or `make sync-doc-revisions`) *after* step 2, so it picks up the CHANGELOG entry just added. Updates `:revnumber:`/`:revdate:` in `docs/admin-guide.adoc` and `docs/user-guide.adoc`. Easy to skip since it's a script step, not a direct edit — skipping it leaves both guides' PDFs claiming the previous release's version.

9. Commit and tag (use this template)

```bash
git add CHANGELOG.md README.md what.md website/hugo.toml ansible/galaxy.yml docs/admin-guide.adoc docs/user-guide.adoc

git commit -m "chore: release v{version} — CHANGELOG, README, what.md\n\nCo-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"

git tag -a v{version} -m "Release v{version}"
```

Note: commits created by you must include the Copilot co-author trailer above when this tool assisted with the release.

10. Push

```bash
git push && git push --tags
```

Optional: Use the Copilot CLI "release" skill (automates the steps above):

- copilot release {version}

When using the Copilot CLI skill, review the generated CHANGELOG entry and commit message before pushing.

Reporting: After pushing, run `git log --oneline $(git describe --tags --abbrev=0)..HEAD` and paste the output into the release notes summary.

That's it — follow these steps to produce a consistent, Copilot-friendly release.
