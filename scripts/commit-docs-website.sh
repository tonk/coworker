#!/usr/bin/env bash
set -euo pipefail

MSG="${1:-}"
BODY="${2:-}"

if [[ -z "${MSG}" ]]; then
  MSG="docs(website): update docs, site copy, and blog content"
fi

if [[ -z "${BODY}" ]]; then
  BODY="Refresh product/docs content and publish website documentation updates with a short blog announcement."
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Not inside a git repository."
  exit 1
fi

# Stage docs + website content related files.
git add docs/ website/content/ README.md CHANGELOG.md 2>/dev/null || true

# Include top-level docs files that frequently accompany website/doc updates.
for f in INSTALL.md TODO.md; do
  if [[ -f "$f" ]]; then
    git add "$f"
  fi
done

if git diff --cached --quiet; then
  echo "No staged docs/website changes found. Nothing to commit."
  exit 0
fi

git commit -m "$(cat <<EOF
${MSG}

${BODY}
EOF
)"

git status --short
