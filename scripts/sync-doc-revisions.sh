#!/usr/bin/env bash
# Sync :revnumber: and :revdate: in docs/*.adoc from the latest CHANGELOG.md release header.
# Run at release time after updating CHANGELOG (or via: make sync-doc-revisions).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CHANGELOG="${ROOT}/CHANGELOG.md"

if [[ ! -f "${CHANGELOG}" ]]; then
  echo "CHANGELOG.md not found at ${CHANGELOG}" >&2
  exit 1
fi

header="$(grep -m1 '^## v' "${CHANGELOG}" || true)"
if [[ -z "${header}" ]]; then
  echo "No release header (## vX.Y.Z — YYYY-MM-DD) found in CHANGELOG.md" >&2
  exit 1
fi

revnumber="$(sed -n 's/^## \(v[^ ]*\).*/\1/p' <<<"${header}")"
revdate="$(sed -n 's/^## v[^ ]* — \([0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\}\).*/\1/p' <<<"${header}")"

if [[ -z "${revnumber}" || -z "${revdate}" ]]; then
  echo "Could not parse version/date from: ${header}" >&2
  exit 1
fi

shopt -s nullglob
files=("${ROOT}"/docs/*.adoc)
if [[ ${#files[@]} -eq 0 ]]; then
  echo "No docs/*.adoc files found" >&2
  exit 1
fi

for f in "${files[@]}"; do
  tmp="$(mktemp)"
  awk -v rev="${revnumber}" -v date="${revdate}" '
    /^:revnumber:/ { next }
    /^:revdate:/ { next }
    {
      print
      if (!inserted && /^= /) {
        print ":revnumber: " rev
        print ":revdate: " date
        inserted = 1
      }
    }
    END {
      if (!inserted) exit 1
    }
  ' "${f}" > "${tmp}" || {
    rm -f "${tmp}"
    echo "Could not insert revision attributes in ${f} (no document title found)" >&2
    exit 1
  }
  mv "${tmp}" "${f}"
  echo "Updated $(basename "${f}") → ${revnumber}, ${revdate}"
done
