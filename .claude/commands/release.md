You are performing a WarmDesk release. The version to release is: $ARGUMENTS

Follow these steps in order:

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
- If any new features were added, update the **Features** list to reflect them.
- If the seed tool changed, update the **Load demo data** section.
- No other sections need changing for a routine release.

## 4. Update what.md
Append any new features or changes as bullet points at the end of the file, matching the imperative style already used there.

## 5. Update documentation
- If any changes where made, that need an update on the documentation,
  update the documentation

## 6. Commit and tag
```bash
git add CHANGELOG.md README.md what.md
git commit -m "chore: release v{version} — CHANGELOG, README, what.md\n\nCo-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
git tag -a v{version} -m "Release v{version}"
```

## 7. Push
```bash
git push && git push --tags
```

Report what was pushed when done.
