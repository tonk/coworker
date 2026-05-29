#!/usr/bin/env python3
"""Inject bulk time entries into WarmDesk.

Usage:
  python scripts/inject_time.py \\
    --url http://localhost:8080 \\
    --user demo.admin \\
    --password demo1234 \\
    --customer "Acme Corp" \\
    --project "Website Redesign" \\
    --minutes 480 \\
    --start 2026-05-01 \\
    --end 2026-05-31 \\
    --weekends    # optional: include Sat/Sun
"""

import argparse
import sys
import time
from datetime import date, datetime, timedelta
from getpass import getpass

import requests


def parse_args():
    p = argparse.ArgumentParser(description="Inject bulk time entries into WarmDesk")
    p.add_argument("--url", required=True, help="WarmDesk base URL (e.g. http://localhost:8080)")
    p.add_argument("--user", required=True, help="Username or email")
    p.add_argument("--password", nargs="?", default=None, const="__prompt__", help="Password (will prompt if omitted)")
    p.add_argument("--customer", default=None, help="Customer name (fuzzy matched) or ID")
    p.add_argument("--project", default=None, help="Project name (fuzzy matched) or ID")
    p.add_argument("--minutes", required=True, type=int, help="Minutes per day")
    p.add_argument("--start", required=True, help="Start date YYYY-MM-DD")
    p.add_argument("--end", required=True, help="End date YYYY-MM-DD (inclusive)")
    p.add_argument("--weekends", action="store_true", help="Include Sat/Sun")
    p.add_argument("--description", default="", help="Default description for each entry")
    p.add_argument("--dry-run", action="store_true", help="Print what would be done without creating")
    return p.parse_args()


def api(verb, url, headers=None, json=None):
    f = getattr(requests, verb.lower(), None)
    if f is None:
        raise ValueError(f"Unsupported HTTP verb: {verb}")
    r = f(url, headers=headers, json=json)
    try:
        body = r.json()
    except Exception:
        body = r.text
    if not r.ok:
        msg = body.get("error", body) if isinstance(body, dict) else body
        raise SystemExit(f"HTTP {r.status_code} on {verb.upper()} {url}: {msg}")
    return body


def resolve_customer(customers, name_or_id):
    """Return (id, name) by matching on ID (if numeric) or case-insensitive substring."""
    try:
        cid = int(name_or_id)
        for c in customers:
            if c["id"] == cid:
                return cid, c["name"]
    except ValueError:
        pass
    lower = name_or_id.lower()
    matches = [c for c in customers if lower in c["name"].lower()]
    if len(matches) == 1:
        return matches[0]["id"], matches[0]["name"]
    if len(matches) > 1:
        print("Multiple customers match. Use the numeric ID:")
        for c in matches:
            print(f"  {c['id']:>5}  {c['name']}")
        sys.exit(1)
    raise SystemExit(f"Customer '{name_or_id}' not found. Use one of --customer <ID> to search by ID.")


def resolve_project(projects, name_or_id):
    """Return (id, name) by matching on ID (if numeric) or case-insensitive substring."""
    try:
        pid = int(name_or_id)
        for p in projects:
            if p["id"] == pid:
                return pid, p["name"]
    except ValueError:
        pass
    lower = name_or_id.lower()
    matches = [p for p in projects if lower in p["name"].lower()]
    if len(matches) == 1:
        return matches[0]["id"], matches[0]["name"]
    if len(matches) > 1:
        print("Multiple projects match. Use the numeric ID:")
        for p in matches:
            print(f"  {p['id']:>5}  {p['name']}")
        sys.exit(1)
    raise SystemExit(f"Project '{name_or_id}' not found. Use one of --project <ID> to search by ID.")


def daterange(start, end, weekends):
    for n in range((end - start).days + 1):
        d = start + timedelta(days=n)
        if not weekends and d.weekday() >= 5:
            continue
        yield d


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
            })
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


def main():
    args = parse_args()

    # Resolve password
    password = args.password
    if password is None or password == "__prompt__":
        try:
            password = getpass("Password: ")
        except (EOFError, KeyboardInterrupt):
            raise SystemExit("Aborted.")

    # -------------------------------------------------------------- Date range
    start = datetime.strptime(args.start, "%Y-%m-%d").date()
    end = datetime.strptime(args.end, "%Y-%m-%d").date()
    if start > end:
        raise SystemExit(f"Start date {start} is after end date {end}.")

    days = list(daterange(start, end, args.weekends))
    total_minutes = len(days) * args.minutes
    print(f"Days:       {len(days)} day(s) ({start} → {end})")
    print(f"Minutes:    {args.minutes}/day = {total_minutes} total")
    print(f"Customer:   {args.customer or '(none)'}")
    print(f"Project:    {args.project or '(none)'}")
    print(f"Description: {args.description or '(none)'}")
    print()

    if args.dry_run:
        for d in days:
            print(f"  [DRY-RUN] Would create: {d}  {args.minutes}min  {args.customer or '—'}  {args.project or '—'}")
        print(f"\nDry run complete — no entries created.")
        return

    base = args.url.rstrip("/")
    api_base = f"{base}/api/v1"

    # ------------------------------------------------------------------ Auth
    print(f"Logging in as {args.user} …")
    tokens = api("post", f"{api_base}/auth/login", json={
        "login": args.user,
        "password": password,
    })
    if tokens.get("mfa_required"):
        print("  MFA required.")
        access_token, refresh_token = mfa_verify(api_base, tokens["mfa_token"])
    else:
        access_token = tokens["access_token"]
        refresh_token = tokens["refresh_token"]
    headers = {
        "Authorization": f"Bearer {access_token}",
        "Content-Type": "application/json",
    }
    print("  OK\n")

    # ---------------------------------------------------------------- Lookups
    customer_id = None
    customer_name = None
    if args.customer:
        print(f"Looking up customer '{args.customer}' …")
        try:
            customers = api("get", f"{api_base}/time-tracking-customers", headers=headers)
            customer_id, customer_name = resolve_customer(customers, args.customer)
        except SystemExit:
            customers = api("get", f"{api_base}/customers", headers=headers)
            customer_id, customer_name = resolve_customer(customers, args.customer)
        print(f"  → {customer_name} (id={customer_id})\n")

    project_id = None
    project_name = None
    if args.project:
        print(f"Looking up project '{args.project}' …")
        try:
            projects = api("get", f"{api_base}/time-tracking-projects", headers=headers)
        except SystemExit:
            projects = api("get", f"{api_base}/projects", headers=headers)
        project_id, project_name = resolve_project(projects, args.project)
        print(f"  → {project_name} (id={project_id})\n")

    # -------------------------------------------------------- Create entries
    created = 0
    token_obtained_at = time.monotonic()
    print("Creating time entries …")
    for d in days:
        # Refresh token if it is approaching expiry (access tokens expire in 15 min).
        if created > 0 and (time.monotonic() - token_obtained_at) > 780:
            print("  Refreshing token …")
            tokens = api("post", f"{api_base}/auth/refresh", headers=headers, json={
                "refresh_token": refresh_token,
            })
            access_token = tokens["access_token"]
            refresh_token = tokens["refresh_token"]
            headers["Authorization"] = f"Bearer {access_token}"
            token_obtained_at = time.monotonic()

        payload = {
            "date": d.isoformat(),
            "minutes": args.minutes,
            "description": args.description or None,
        }
        if customer_id is not None:
            payload["customer_id"] = customer_id
        if project_id is not None:
            payload["project_id"] = project_id

        err = None
        try:
            api("post", f"{api_base}/time-entries", headers=headers, json=payload)
            created += 1
        except SystemExit as e:
            err = str(e)

        status = "✓" if err is None else "✗"
        extra = f"  {err}" if err else ""
        print(f"  {status} {d}  {args.minutes}min{extra}")

    print(f"\nDone — {created}/{len(days)} time entries created.")


if __name__ == "__main__":
    main()
