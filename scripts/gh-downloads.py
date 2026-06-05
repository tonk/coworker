#!/usr/bin/env python3
"""Show download counts per GitHub release."""

import json
import sys
import urllib.request
from urllib.error import HTTPError

REPO = "tonk/warmdesk"
API  = f"https://api.github.com/repos/{REPO}/releases?per_page=100"


def fetch(url: str) -> list:
    req = urllib.request.Request(url, headers={"User-Agent": "gh-downloads/1.0",
                                                "Accept": "application/vnd.github+json"})
    try:
        with urllib.request.urlopen(req) as r:
            return json.loads(r.read())
    except HTTPError as e:
        sys.exit(f"GitHub API error {e.code}: {e.reason}")


def main() -> None:
    repo = sys.argv[1] if len(sys.argv) > 1 else REPO
    url  = f"https://api.github.com/repos/{repo}/releases?per_page=100"

    releases = fetch(url)
    if not releases:
        print("No releases found.")
        return

    grand_total = 0
    col = 52  # asset name column width

    for rel in releases:
        tag        = rel["tag_name"]
        published  = rel.get("published_at", "")[:10]
        prerelease = "  [pre-release]" if rel["prerelease"] else ""
        assets     = rel.get("assets", [])
        total      = sum(a["download_count"] for a in assets)
        grand_total += total

        print(f"\n{tag}  ({published}){prerelease}  —  {total:,} total")
        print("  " + "-" * (col + 12))

        if not assets:
            print("  (no assets)")
        else:
            for a in sorted(assets, key=lambda x: -x["download_count"]):
                name  = a["name"]
                count = a["download_count"]
                bar   = "#" * min(count // max(1, total // 20), 20)
                print(f"  {name:<{col}}  {count:>7,}  {bar}")

    print(f"\n{'=' * (col + 16)}")
    print(f"  {'GRAND TOTAL':<{col}}  {grand_total:>7,}")


if __name__ == "__main__":
    main()
