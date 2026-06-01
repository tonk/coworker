#!/usr/bin/env python3
"""Bulk-export time-sheet grid PDFs from WarmDesk for a given user.

Usage:
  # Export every week of 2026 for user "jane"
  python scripts/export_grid_pdf.py \\
    --url http://localhost:8080 \\
    --user demo.admin --password demo1234 \\
    --target jane \\
    --grid week --year 2026 \\
    --out ./pdfs

  # Export every month of Q1 2026
  python scripts/export_grid_pdf.py \\
    --url http://localhost:8080 \\
    --user demo.admin \\
    --target jane \\
    --grid month --year 2026 --month-start 1 --month-end 3 \\
    --out ./pdfs

  # Single month
  python scripts/export_grid_pdf.py \\
    --url http://localhost:8080 \\
    --user demo.admin \\
    --target jane \\
    --grid month --year 2026 --month 5 \\
    --out ./pdfs

  # Single week (ISO week)
  python scripts/export_grid_pdf.py \\
    --url http://localhost:8080 \\
    --user demo.admin \\
    --target jane \\
    --grid week --year 2026 --week 22 \\
    --out ./pdfs

  # Single week by start date
  python scripts/export_grid_pdf.py \\
    --url http://localhost:8080 \\
    --user demo.admin \\
    --target jane \\
    --grid week --start-date 2026-06-01 \\
    --out ./pdfs

  # All months in a whole year
  python scripts/export_grid_pdf.py \\
    --url http://localhost:8080 \\
    --user demo.admin \\
    --target jane \\
    --grid month --year 2026 \\
    --out ./pdfs
"""

import argparse
import os
import string
import sys
import time
from datetime import date, datetime, timedelta
from getpass import getpass

import requests


def parse_args():
    p = argparse.ArgumentParser(description="Bulk-export grid PDFs from WarmDesk")
    p.add_argument("--url", required=True, help="WarmDesk base URL (e.g. http://localhost:8080)")
    p.add_argument("--user", help="Your WarmDesk username or email (must be admin or time_tracking_viewer)")
    p.add_argument("--password", nargs="?", default=None, const="__prompt__", help="Password (prompted if omitted)")
    p.add_argument("--refresh-token", default=None, help="Use an existing refresh token instead of logging in (skips rate limiter)")
    p.add_argument("--get-token", action="store_true", help="Login and print the refresh token to stdout, then exit")
    p.add_argument("--target", help="Target username to export (or numeric user ID)")
    p.add_argument("--grid", default="week", choices=["week", "month", "year"], help="Grid type")
    p.add_argument("--year", type=int, default=None, help="Year (default: current)")
    p.add_argument("--month", type=int, default=None, help="Single month (1-12) for month grid")
    p.add_argument("--month-start", type=int, default=None, help="First month for range export (1-12)")
    p.add_argument("--month-end", type=int, default=None, help="Last month for range export (1-12, inclusive)")
    p.add_argument("--week", type=int, default=None, help="Single ISO week number for week grid")
    p.add_argument("--week-start", type=int, default=None, help="First ISO week for range export")
    p.add_argument("--week-end", type=int, default=None, help="Last ISO week for range export (inclusive)")
    p.add_argument("--start-date", type=str, default=None, help="Start date YYYY-MM-DD (week grid, uses ISO week of this date)")
    p.add_argument("--out", "-o", default=".", help="Output directory (default: current dir)")
    p.add_argument("--dry-run", action="store_true", help="Print what would be downloaded without making API calls")
    p.add_argument("--overwrite", action="store_true", help="Overwrite existing PDFs")
    p.add_argument("--font", default=None, help="PDF font override (e.g. 'dejavu', 'freesans', 'courier')")
    p.add_argument("--lang", default=None, help="PDF language override (e.g. 'en', 'nl', 'de')")
    p.add_argument("--filename-template", default=None,
                    help='Output filename template with {grid}, {year}, {month}, {week} placeholders. '
                         'Default: grid-{grid}_{year}_{month:02d}.pdf / grid-{grid}_{year}_w{week}.pdf')
    return p.parse_args()


def api(verb, url, headers=None, json=None, params=None):
    f = getattr(requests, verb.lower(), None)
    if f is None:
        raise ValueError(f"Unsupported HTTP verb: {verb}")
    r = f(url, headers=headers, json=json, params=params)
    if not r.ok:
        try:
            body = r.json()
        except Exception:
            body = r.text
        msg = body.get("error", body) if isinstance(body, dict) else body
        raise SystemExit(f"HTTP {r.status_code} on {verb.upper()} {url}: {msg}")
    return r


def mfa_verify(api_base, mfa_token):
    """Prompt for 6-digit MFA code and verify. Returns (access_token, refresh_token) or exits."""
    for attempt in range(3):
        try:
            code = getpass("MFA code (6 digits): ")
        except (EOFError, KeyboardInterrupt):
            raise SystemExit("Aborted.")
        code = code.strip()
        if not code.isdigit() or len(code) != 6:
            print("  Invalid code — must be 6 digits.")
            continue
        try:
            resp = api("post", f"{api_base}/auth/mfa/verify", json={
                "mfa_token": mfa_token,
                "code": code,
            }).json()
            return resp["access_token"], resp["refresh_token"]
        except SystemExit as e:
            msg = str(e)
            if "invalid_code" in msg or "unprocessable" in msg.lower():
                print(f"  Wrong code ({2 - attempt} attempt(s) left).")
            elif "mfa_session_expired" in msg or "session_expired" in msg:
                raise SystemExit("MFA session expired. Run the script again.")
            else:
                raise
    raise SystemExit("Too many failed MFA attempts.")


def resolve_user_id(api_base, headers, target):
    """Resolve target username/user-id to numeric user ID."""
    if target.isdigit():
        r = api("get", f"{api_base}/admin/users/{target}", headers=headers)
        u = r.json()
        return u["id"], u.get("display_name") or u["username"]

    # Search by username
    r = api("get", f"{api_base}/users", headers=headers)
    users = r.json()
    for u in users:
        if u.get("username", "").lower() == target.lower():
            return u["id"], u.get("display_name") or u["username"]
        if u.get("email", "").lower() == target.lower():
            return u["id"], u.get("display_name") or u["username"]
    # Try admin endpoint
    r = api("get", f"{api_base}/admin/users", headers=headers)
    users = r.json()
    for u in users:
        if u.get("username", "").lower() == target.lower():
            return u["id"], u.get("display_name") or u["username"]
        if u.get("email", "").lower() == target.lower():
            return u["id"], u.get("display_name") or u["username"]
    raise SystemExit(f"User '{target}' not found.")


def iso_week_start(year, week):
    """Return the Monday of the given ISO year/week."""
    jan4 = date(year, 1, 4)
    first_mon = jan4 - timedelta(days=jan4.weekday())
    return first_mon + timedelta(weeks=week - 1)


class _Fmt(string.Formatter):
    """Formatter that renders None values as empty string instead of crashing."""
    def format_field(self, value, format_spec):
        if value is None:
            return ""
        return super().format_field(value, format_spec)
_fmt = _Fmt()

def download_grid_pdf(api_base, headers, target_user_id, grid_type, params, out_dir, font, lang, dry_run, overwrite, filename_template=None):
    """Download a single grid PDF and save to disk. Returns the file path or None if skipped."""
    query = {
        "grid": grid_type,
        "user_id": str(target_user_id),
    }
    query.update(params)

    if font:
        query["font"] = font
    if lang:
        query["lang"] = lang

    # Build filename context — only keys present in the query
    ctx = {"grid": grid_type}
    if query.get("year"):
        ctx["year"] = int(query["year"])
    if query.get("month"):
        ctx["month"] = int(query["month"])
    if query.get("week"):
        ctx["week"] = int(query["week"])
    if filename_template:
        filename = _fmt.format(filename_template, **ctx)
    else:
        parts = [f"grid-{grid_type}"]
        if "year" in query:
            parts.append(query["year"])
        if "month" in query:
            parts.append(f"{int(query['month']):02d}")
        if "week" in query:
            parts.append(f"w{query['week']}")
        if "start_date" in query:
            parts.append(query["start_date"])
        if font:
            parts.append(f"font-{font}")
        if lang:
            parts.append(f"lang-{lang}")
        filename = "_".join(parts) + ".pdf"
    filepath = os.path.join(out_dir, filename)

    if not overwrite and os.path.exists(filepath):
        print(f"  SKIP  {filename} (exists)")
        return None

    if dry_run:
        print(f"  DRY-RUN  {filename}")
        return filepath

    r = api("get", f"{api_base}/time-entries/grid/pdf", headers=headers, params=query)
    with open(filepath, "wb") as f:
        f.write(r.content)
    print(f"  OK    {filename}")
    return filepath


def main():
    args = parse_args()

    base = args.url.rstrip("/")
    api_base = f"{base}/api/v1"

    if not args.get_token and not args.target:
        raise SystemExit("--target is required when not using --get-token")

    out_dir = os.path.abspath(args.out)
    os.makedirs(out_dir, exist_ok=True)

    # ── Auth ────────────────────────────────────────────────────────────────
    if args.refresh_token:
        print(f"Using refresh token …")
        r = api("post", f"{api_base}/auth/refresh", json={"refresh_token": args.refresh_token})
        tokens = r.json()
        access_token = tokens["access_token"]
        refresh_token = tokens["refresh_token"]
        print("  OK\n")
    else:
        if not args.user:
            raise SystemExit("--user is required when not using --refresh-token")
        password = args.password
        if password is None or password == "__prompt__":
            try:
                password = getpass("Password: ")
            except (EOFError, KeyboardInterrupt):
                raise SystemExit("Aborted.")
        print(f"Logging in as {args.user} …")
        tokens = api("post", f"{api_base}/auth/login", json={
            "login": args.user,
            "password": password,
        }).json()
        if tokens.get("mfa_required"):
            print("  MFA required.")
            access_token, refresh_token = mfa_verify(api_base, tokens["mfa_token"])
        else:
            access_token = tokens["access_token"]
            refresh_token = tokens["refresh_token"]
        print("  OK\n")
    headers = {
        "Authorization": f"Bearer {access_token}",
    }

    if args.get_token:
        print(refresh_token)
        return

    # ── Resolve target user ─────────────────────────────────────────────────
    print(f"Looking up user '{args.target}' …")
    target_user_id, target_name = resolve_user_id(api_base, headers, args.target)
    print(f"  → {target_name} (id={target_user_id})\n")

    now = datetime.now()
    year = args.year or now.year
    today = now.date()

    # ── Build period list ───────────────────────────────────────────────────
    periods = []

    if args.grid == "year":
        # Single year grid
        periods.append({"year": str(year)})

    elif args.grid == "month":
        if args.month is not None:
            months = [args.month]
        elif args.month_start is not None and args.month_end is not None:
            months = list(range(args.month_start, args.month_end + 1))
        else:
            months = list(range(1, 13))
        for m in months:
            periods.append({"year": str(year), "month": str(m)})

    elif args.grid == "week":
        if args.week is not None:
            weeks = [args.week]
        elif args.week_start is not None and args.week_end is not None:
            weeks = list(range(args.week_start, args.week_end + 1))
        elif args.start_date is not None:
            d = datetime.strptime(args.start_date, "%Y-%m-%d").date()
            iso_year, iso_week, _ = d.isocalendar()
            periods.append({"year": str(iso_year), "week": str(iso_week)})
            year = iso_year
        else:
            # All weeks of the year
            weeks_list = []
            # ISO week 1 of the year
            d = iso_week_start(year, 1)
            while d.year <= year:
                iso_y, iso_w, _ = d.isocalendar()
                if iso_y == year:
                    weeks_list.append(iso_w)
                d += timedelta(weeks=1)
            weeks = weeks_list
        if args.week is not None or (args.week_start is not None and args.week_end is not None):
            weeks_list = list(range(min(weeks), max(weeks) + 1)) if isinstance(weeks, list) else weeks
            for w in weeks_list:
                periods.append({"year": str(year), "week": str(w)})

    count = len(periods)
    print(f"Exporting {count} {args.grid} grid PDF(s) for {target_name} → {out_dir}\n")

    # ── Download ────────────────────────────────────────────────────────────
    token_obtained_at = time.monotonic()
    downloaded = 0
    skipped = 0

    for i, params in enumerate(periods, 1):
        # Refresh token if nearing expiry (access tokens expire in 15 min).
        if (time.monotonic() - token_obtained_at) > 780:
            print("  Refreshing token …")
            r = api("post", f"{api_base}/auth/refresh", headers=headers, json={
                "refresh_token": refresh_token,
            })
            data = r.json()
            access_token = data["access_token"]
            refresh_token = data["refresh_token"]
            headers["Authorization"] = f"Bearer {access_token}"
            token_obtained_at = time.monotonic()

        print(f"[{i}/{count}] ", end="")
        result = download_grid_pdf(
            api_base, headers, target_user_id, args.grid, params,
            out_dir, args.font, args.lang, args.dry_run, args.overwrite,
            args.filename_template,
        )
        if result is None:
            skipped += 1
        else:
            downloaded += 1

    print(f"\nDone — {downloaded} downloaded, {skipped} skipped (out of {count}).")


if __name__ == "__main__":
    main()
