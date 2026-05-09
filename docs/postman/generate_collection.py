#!/usr/bin/env python3
"""Regenerate Postman collection + environments from docs/bruno/*.bru."""

from __future__ import annotations

import json
import re
from pathlib import Path

REPO = Path(__file__).resolve().parents[2]
BRUNO_ROOT = REPO / "docs" / "bruno"
OUT_DIR = REPO / "docs" / "postman"


def parse_block(lines: list[str], start_i: int) -> tuple[str, int]:
    """Parse a `keyword { ... }` block; return inner text and line after closing `}`."""
    first = lines[start_i].rstrip()
    if "{" not in first:
        raise ValueError(f"expected opening brace: {first!r}")
    keyword, rest = first.split("{", 1)
    keyword = keyword.strip()
    buf: list[str] = []
    depth = first.count("{") - first.count("}")
    rest = rest.rstrip()
    if rest:
        buf.append(rest)
    i = start_i + 1
    while i < len(lines):
        lin = lines[i]
        opens = lin.count("{")
        closes = lin.count("}")
        pending = depth + opens - closes
        if pending == 0 and lin.strip() == "}":
            return "\n".join(buf), i + 1
        buf.append(lin)
        depth = pending
        i += 1
    raise ValueError(f"unclosed block starting line {start_i + 1} ({keyword})")


def parse_bru(text: str) -> dict[str, str]:
    lines = text.splitlines()
    blocks: dict[str, str] = {}
    i = 0
    while i < len(lines):
        stripped = lines[i].strip()
        if not stripped:
            i += 1
            continue
        if not stripped.endswith("{"):
            i += 1
            continue
        key = stripped.split("{", 1)[0].strip()
        body, ni = parse_block(lines, i)
        blocks[key] = body
        i = ni
    return blocks


def meta_name(seq_body: str) -> tuple[str, int]:
    m = re.search(r"^\s*name:\s*(.+)$", seq_body, re.MULTILINE)
    name = m.group(1).strip() if m else "Unnamed"
    sq = re.search(r"^\s*seq:\s*(\d+)", seq_body, re.MULTILINE)
    seq_n = int(sq.group(1)) if sq else 999
    return name, seq_n


def parse_http_block(body: str) -> tuple[str, str, str]:
    url_m = re.search(r"^\s*url:\s*(.+)$", body, re.MULTILINE)
    auth_m = re.search(r"^\s*auth:\s*(\S+)", body, re.MULTILINE)
    btype_m = re.search(r"^\s*body:\s*(\S+)", body, re.MULTILINE)
    url = url_m.group(1).strip() if url_m else ""
    auth = auth_m.group(1).strip() if auth_m else "none"
    bmode = btype_m.group(1).strip() if btype_m else "none"
    return url, auth, bmode


def key_value_simple(body: str) -> dict[str, str]:
    out: dict[str, str] = {}
    for ln in body.splitlines():
        ln = ln.strip()
        if not ln or ln.startswith("//"):
            continue
        if ":" not in ln:
            continue
        k, v = ln.split(":", 1)
        out[k.strip()] = v.strip()
    return out


def postman_url(raw: str, query: dict[str, str] | None) -> dict:
    uobj: dict = {"raw": raw}
    if query:
        arr = []
        for k, v in query.items():
            item = {"key": k, "value": v}
            if v == "":
                item["disabled"] = True
            arr.append(item)
        uobj["query"] = arr
    return uobj


def build_request(
    name: str,
    method: str,
    url: str,
    auth_type: str,
    headers: dict[str, str],
    query: dict[str, str] | None,
    body_raw: str | None,
    body_mode: str,
    docs: str,
    test_lines: list[str] | None,
) -> dict:
    req_body: dict = {
        "name": name,
        "request": {
            "method": method.upper(),
            "header": [{"key": k, "value": v, "type": "text"} for k, v in headers.items()],
            "url": postman_url(url, query),
        },
    }

    doc = docs.strip()
    if doc:
        req_body["request"]["description"] = doc

    if auth_type in ("bearer", "inherit"):
        req_body["request"]["auth"] = {
            "type": "bearer",
            "bearer": [{"key": "token", "value": "{{token}}", "type": "string"}],
        }
    else:
        req_body["request"]["auth"] = {"type": "noauth"}

    if body_mode == "json" and body_raw is not None:
        req_body["request"]["body"] = {
            "mode": "raw",
            "raw": body_raw.strip(),
            "options": {"raw": {"language": "json"}},
        }

    if test_lines:
        req_body["event"] = [
            {
                "listen": "test",
                "script": {"exec": test_lines, "type": "text/javascript"},
            }
        ]

    return req_body


LOGIN_TEST_SCRIPT = [
    "if (pm.response.code === 200 && pm.response.json().access_token) {",
    "    pm.environment.set('token', pm.response.json().access_token);",
    "}",
]


def convert_bru(path: Path) -> dict:
    txt = path.read_text(encoding="utf-8")
    blocks = parse_bru(txt)

    meta_body = blocks.get("meta", "")
    name, seq_n = meta_name(meta_body)

    http_key = None
    for cand in ("get", "post", "put", "patch", "delete"):
        if cand in blocks:
            http_key = cand
            break
    if http_key is None:
        raise ValueError(f"no HTTP method in {path}")

    url, auth_type, body_mode = parse_http_block(blocks[http_key])
    headers = key_value_simple(blocks.get("headers", ""))
    query_kv = (
        key_value_simple(blocks["params:query"]) if "params:query" in blocks else None
    )
    body_raw = blocks["body:json"].strip() if "body:json" in blocks else None
    docs = blocks.get("docs", "").strip()

    test_lines = None
    if "script:post-response" in blocks:
        scr = blocks["script:post-response"]
        if "access_token" in scr and ("bru.setVar" in scr or "bru" in scr):
            test_lines = LOGIN_TEST_SCRIPT

    item = build_request(
        name, http_key, url, auth_type, headers, query_kv, body_raw, body_mode, docs, test_lines
    )
    item["_seq"] = seq_n
    return item


def main() -> None:
    folders: dict[str, list] = {}

    for bru_path in sorted(BRUNO_ROOT.rglob("*.bru")):
        if bru_path.parts[-2] == "environments":
            continue
        rel = bru_path.relative_to(BRUNO_ROOT)
        folder = rel.parts[0] if len(rel.parts) > 1 else "root"

        folders.setdefault(folder, []).append(convert_bru(bru_path))

    for fk in folders:
        folders[fk].sort(key=lambda x: (x["_seq"], x["name"]))
        for it in folders[fk]:
            it.pop("_seq", None)

    collection_items = [{"name": fk, "item": folders[fk]} for fk in sorted(folders.keys())]

    collection = {
        "info": {
            "name": "WarmDesk API",
            "description": (
                "Postman mirror of docs/bruno. Import the matching environment "
                "(WarmDesk Local or WarmDesk Production), run Login — the test "
                "script saves access_token as `token` on the active environment "
                "(same behaviour as the Bruno Login post-response script)."
            ),
            "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
        },
        "variable": [{"key": "baseUrl", "value": "http://localhost:8080"}],
        "item": collection_items,
    }

    OUT_DIR.mkdir(parents=True, exist_ok=True)
    (OUT_DIR / "WarmDesk-API.postman_collection.json").write_text(
        json.dumps(collection, indent=2), encoding="utf-8"
    )

    local_env = {
        "name": "WarmDesk Local",
        "values": [
            {"key": "baseUrl", "value": "http://localhost:8080", "enabled": True},
            {"key": "username", "value": "tonk", "enabled": True},
            {"key": "password", "value": "demo1234", "enabled": True, "type": "secret"},
            {"key": "token", "value": "", "enabled": True, "type": "secret"},
            {"key": "refreshToken", "value": "", "enabled": True, "type": "secret"},
            {"key": "apiKey", "value": "", "enabled": True, "type": "secret"},
            {"key": "projectSlug", "value": "website-redesign", "enabled": True},
            {"key": "columnId", "value": "1", "enabled": True},
            {"key": "cardId", "value": "1", "enabled": True},
            {"key": "sprintId", "value": "1", "enabled": True},
            {"key": "releaseId", "value": "1", "enabled": True},
            {"key": "userId", "value": "2", "enabled": True},
            {"key": "labelId", "value": "1", "enabled": True},
            {"key": "topicId", "value": "1", "enabled": True},
        ],
    }

    prod_env = {
        "name": "WarmDesk Production",
        "values": [
            {"key": "baseUrl", "value": "https://warmdesk.smartowl.nl", "enabled": True},
            {"key": "username", "value": "tonk", "enabled": True},
            {"key": "password", "value": "", "enabled": True, "type": "secret"},
            {"key": "token", "value": "", "enabled": True, "type": "secret"},
            {"key": "refreshToken", "value": "", "enabled": True, "type": "secret"},
            {"key": "apiKey", "value": "", "enabled": True, "type": "secret"},
            {"key": "projectSlug", "value": "", "enabled": True},
            {"key": "columnId", "value": "", "enabled": True},
            {"key": "cardId", "value": "", "enabled": True},
            {"key": "sprintId", "value": "", "enabled": True},
            {"key": "releaseId", "value": "", "enabled": True},
            {"key": "userId", "value": "", "enabled": True},
            {"key": "labelId", "value": "", "enabled": True},
            {"key": "topicId", "value": "", "enabled": True},
        ],
    }

    (OUT_DIR / "WarmDesk-Local.postman_environment.json").write_text(
        json.dumps(local_env, indent=2), encoding="utf-8"
    )
    (OUT_DIR / "WarmDesk-Production.postman_environment.json").write_text(
        json.dumps(prod_env, indent=2), encoding="utf-8"
    )

    n = sum(len(folders[k]) for k in folders)
    print(f"Wrote {OUT_DIR} ({n} requests in {len(folders)} folders)")


if __name__ == "__main__":
    main()
