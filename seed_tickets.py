#!/usr/bin/env python3
"""
Seed bulk tickets for a customer. Creates 100 tickets with varied statuses,
priorities, and types. Pending tickets get reminder dates (some overdue,
some due soon, some far out).

Usage:
  python3 seed_tickets.py  [--url http://localhost:8080] [--count 100]
                           [--customer 1] [--user admin] [--pass demo1234]

The customer ID is shown when you run it the first time.
"""

import argparse
import json
import subprocess
import sys
from datetime import datetime, timedelta
from collections import Counter

COOKIE_JAR = "/tmp/wd-ticket-seed-cookies.txt"
URL = "http://localhost:8080"


def api(method, path, data=None):
    """Run a curl command and return parsed JSON."""
    args = [
        "curl", "-s", "-b", COOKIE_JAR,
        "-X", method,
        f"{URL}{path}",
        "-H", "Content-Type: application/json",
    ]
    if data is not None:
        args += ["-d", json.dumps(data)]
    result = subprocess.run(args, capture_output=True, text=True)
    try:
        return json.loads(result.stdout)
    except json.JSONDecodeError:
        print(f"  ⚠  Non-JSON response from {method} {path}: {result.stdout[:200]}")
        return None


def login(login, password):
    subprocess.run(["curl", "-s", "-c", COOKIE_JAR, "-X", "POST",
                    f"{URL}/api/v1/auth/login",
                    "-H", "Content-Type: application/json",
                    "-d", json.dumps({"login": login, "password": password})],
                   capture_output=True)
    # Verify by listing customers
    customers = api("GET", "/api/v1/customers")
    if isinstance(customers, list):
        print(f"  Logged in as {login}. Available customers:")
        for c in customers:
            print(f"    {c['id']:>5}  {c['name']}")
        return customers
    elif isinstance(customers, dict) and "error" in customers:
        print(f"  Login failed: {customers['error']}")
        sys.exit(1)
    else:
        print(f"  Unexpected response: {str(customers)[:200]}")
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="Seed bulk tickets for a customer")
    parser.add_argument("--url", default="http://localhost:8080", help="Base URL")
    parser.add_argument("--count", type=int, default=100, help="Number of tickets to create")
    parser.add_argument("--customer", type=int, required=True, help="Customer ID")
    parser.add_argument("--user", default="demo.admin", help="Login username")
    parser.add_argument("--password", default="demo1234", help="Login password")
    args = parser.parse_args()

    global URL
    URL = args.url.rstrip("/")

    print(f"WarmDesk Ticket Seeder")
    print(f"  Server  : {URL}")
    print(f"  Customer: {args.customer}")
    print(f"  Count   : {args.count}")
    print()

    # Login
    print("1. Authenticating…")
    login(args.user, args.password)

    # Create tickets
    print(f"\n2. Creating {args.count} tickets…")
    types = ["incident", "problem", "service_request", "change_request"]
    priorities = ["low", "medium", "high", "critical"]
    statuses = ["new", "open", "pending", "pending_close", "closed"]

    for i in range(1, args.count + 1):
        t = types[(i - 1) % len(types)]
        p = priorities[(i - 1) % len(priorities)]
        ticket = api("POST", f"/api/v1/customers/{args.customer}/tickets", {
            "title": f"Seed ticket #{i} — {t} ({p})",
            "type": t,
            "priority": p,
        })
        if ticket is None or isinstance(ticket, dict) and ticket.get("error"):
            print(f"  ✗ Failed at ticket #{i}: {ticket}")
            sys.exit(1)
        if i % 25 == 0:
            print(f"  … {i} created")

    print(f"  ✓ {args.count} tickets created\n")

    # Fetch all ticket IDs
    print("3. Fetching ticket list…")
    all_tickets = api("GET", f"/api/v1/customers/{args.customer}/tickets")
    if not isinstance(all_tickets, list):
        print(f"  ✗ Failed to fetch tickets: {all_tickets}")
        sys.exit(1)
    print(f"  Got {len(all_tickets)} tickets\n")

    # Update statuses and set reminders
    print("4. Assigning statuses and reminder dates…")
    tomorrow = (datetime.now() + timedelta(days=1)).strftime("%Y-%m-%d")
    next_week = (datetime.now() + timedelta(days=7)).strftime("%Y-%m-%d")
    yesterday = (datetime.now() - timedelta(days=1)).strftime("%Y-%m-%d")

    # Sort by ID to make order predictable
    ids = sorted(t["id"] for t in all_tickets)
    updated = 0
    pending_count = 0

    for idx, tid in enumerate(ids):
        st = statuses[idx % len(statuses)]

        api("PUT", f"/api/v1/customers/{args.customer}/tickets/{tid}", {
            "status": st,
        })
        updated += 1

        if st == "pending":
            if idx % 3 == 0:
                d = tomorrow
            elif idx % 3 == 1:
                d = next_week
            else:
                d = yesterday
            api("PUT", f"/api/v1/customers/{args.customer}/tickets/{tid}", {
                "reminder_at": f"{d}T12:00:00Z",
            })
            pending_count += 1

        if (idx + 1) % 50 == 0:
            print(f"  … {idx + 1} updated")

    print(f"  ✓ {updated} statuses updated, {pending_count} have reminder dates\n")

    # Summary
    final = api("GET", f"/api/v1/customers/{args.customer}/tickets")
    if isinstance(final, list):
        c = Counter(t["status"] for t in final)
        print("5. Final distribution:")
        for status in statuses:
            print(f"    {status:>16}: {c.get(status, 0)}")
        pending_warn = sum(
            1 for t in final
            if t["status"] == "pending" and t.get("reminder_at")
        )
        print(f"    {'pending with reminder':>16}: {pending_warn}")
    else:
        print("  Could not fetch final state")


if __name__ == "__main__":
    main()
