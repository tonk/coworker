# Release (Copilot CLI style)

This document converts the original release instructions into concise, runnable commands and a Copilot-friendly workflow.

Usage (manual):

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

6. Commit and tag (use this template)

```bash
git add CHANGELOG.md README.md what.md

git commit -m "chore: release v{version} — CHANGELOG, README, what.md\n\nCo-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"

git tag -a v{version} -m "Release v{version}"
```

Note: commits created by you must include the Copilot co-author trailer above when this tool assisted with the release.

7. Push

```bash
git push && git push --tags
```

Optional: Use the Copilot CLI "release" skill (automates the steps above):

- copilot release {version}

When using the Copilot CLI skill, review the generated CHANGELOG entry and commit message before pushing.

Reporting: After pushing, run `git log --oneline $(git describe --tags --abbrev=0)..HEAD` and paste the output into the release notes summary.

That's it — follow these steps to produce a consistent, Copilot-friendly release.
