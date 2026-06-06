#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────────
# WarmDesk — Automated screenshot updater
#
# Seeds the database, starts the backend + frontend, runs Playwright
# screenshot tests, then shuts everything down.
#
# Usage:
#   ./e2e/screenshots.sh                  # full run
#   JWT_SECRET=xxx ./e2e/screenshots.sh   # with custom secret
#   SKIP_SEED=1   ./e2e/screenshots.sh    # skip re-seeding (faster; dates may drift)
#   SKIP_SERVERS=1 ./e2e/screenshots.sh   # servers already running
# ──────────────────────────────────────────────────────────────────────
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FRONTEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKEND_DIR="$(cd "$FRONTEND_DIR/../backend" && pwd)"
export JWT_SECRET="${JWT_SECRET:-change-me-in-production-but-this-is-ok-for-e2e}"

cleanup() {
  if [ -n "${BACKEND_PID:-}" ]; then
    echo "Shutting down backend…"
    kill "$BACKEND_PID" 2>/dev/null || true
    wait "$BACKEND_PID" 2>/dev/null || true
  fi
  if [ -n "${FRONTEND_PID:-}" ]; then
    echo "Shutting down frontend…"
    kill "$FRONTEND_PID" 2>/dev/null || true
    wait "$FRONTEND_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

# ── 1. Seed ──────────────────────────────────────────────────────────
if [ -z "${SKIP_SEED:-}" ]; then
  echo "=== Seeding database ==="
  (cd "$BACKEND_DIR" && go run ./cmd/seed --reset)
fi

# ── 2. Start backend ────────────────────────────────────────────────
BACKEND_UP=false
FRONTEND_UP=false
curl -s http://localhost:8080/api/v1/version > /dev/null 2>&1 && BACKEND_UP=true
curl -s http://localhost:5173 > /dev/null 2>&1 && FRONTEND_UP=true

if [ "${SKIP_SERVERS:-false}" = "true" ]; then
  if [ "$BACKEND_UP" != "true" ]; then echo "ERROR: SKIP_SERVERS=1 but backend :8080 is down" >&2; exit 1; fi
  if [ "$FRONTEND_UP" != "true" ]; then echo "ERROR: SKIP_SERVERS=1 but frontend :5173 is down" >&2; exit 1; fi
elif [ "$BACKEND_UP" = "true" ] && [ "$FRONTEND_UP" = "true" ]; then
  echo "=== Both servers already running — reusing ==="
else
  if [ "$BACKEND_UP" = "true" ]; then
    echo "=== Backend already running → http://localhost:8080 ==="
  else
    echo "=== Starting backend  → http://localhost:8080 ==="
    (cd "$BACKEND_DIR" && GIN_MODE=debug exec go run .) &
    BACKEND_PID=$!
    echo "Waiting for backend…"
    while ! curl -s http://localhost:8080/api/v1/version > /dev/null 2>&1; do sleep 1; done
    echo "  → backend is up"
  fi

  if [ "$FRONTEND_UP" = "true" ]; then
    echo "=== Frontend already running → http://localhost:5173 ==="
  else
    echo "=== Starting frontend → http://localhost:5173 ==="
    (cd "$FRONTEND_DIR" && exec npm run dev) &
    FRONTEND_PID=$!
    echo "Waiting for frontend…"
    while ! curl -s http://localhost:5173 > /dev/null 2>&1; do sleep 1; done
    echo "  → frontend is up"
  fi
fi

# ── 3. Run screenshots ──────────────────────────────────────────────
echo "=== Running Playwright screenshots ==="
cd "$FRONTEND_DIR"
npx playwright test --config=e2e/playwright.config.js

echo ""
echo "Done — screenshots written to ../screenshots/"
