# WarmDesk — API Reference

## Contents

1. [Authentication](#1-authentication)
2. [Ticket API](#2-ticket-api)
3. [Helpdesk — Ticket Messages](#3-helpdesk--ticket-messages)
4. [Generic Webhook](#4-generic-webhook)
5. [Git Platform Webhooks](#5-git-platform-webhooks)
   - [Gitea / Forgejo](#51-gitea--forgejo)
   - [GitHub](#52-github)
   - [GitLab](#53-gitlab)
6. [Card References](#6-card-references)
7. [Response Formats](#7-response-formats)

See also: [Interactive API (Swagger UI)](#interactive-api-swagger-ui) · [Bruno Collection](#bruno-collection)

---

## Interactive API (Swagger UI)

WarmDesk ships with a full Swagger UI at:

```
http://<your-server>:8080/swagger/index.html
```

The interactive documentation lists every endpoint, shows request/response
schemas, and lets you try requests directly from the browser. A valid JWT
Bearer token (obtained from `POST /api/v1/auth/login`) can be entered via the
**Authorize** button to authenticate requests.

---

## Bruno Collection

A [Bruno](https://www.usebruno.com/) API collection covering all major endpoints
is included at `docs/bruno/`. Bruno is an open-source API client that stores
collections as plain text files — no account or cloud sync required.

**Opening the collection:**

1. Install Bruno from [usebruno.com](https://www.usebruno.com/)
2. Open Bruno → **Open Collection** → select the `docs/bruno/` directory
3. Select the **local** or **production** environment from the top-right dropdown
4. In the left sidebar open the **auth** folder, click **Login**, and press
   **Send** — the `token` variable is set automatically and all other requests
   will use it immediately

**Environments:**

| Environment | Base URL |
|---|---|
| `local` | `http://localhost:8080` — pre-filled with demo credentials |
| `production` | Your production instance URL |

**Folders in the collection:**

| Folder | Contents |
|---|---|
| `auth` | Login, register, refresh, current user, API keys |
| `system` | Server version, public settings |
| `projects` | CRUD, members, labels, star/unstar |
| `columns` | CRUD |
| `cards` | CRUD, move, comments, checklist, labels, history, cross-references |
| `scrum` | Backlog, sprints (create/start/complete), releases |
| `charts` | Velocity, burndown, burnup, CFD, cycle time, throughput, release burndown |
| `topics` | List, create, get, reply |
| `reports` | Time report, time entries |
| `helpdesk` | Customer ticket list, ticket detail, messages, tags, links, macros |
| `helpdesk-admin` | SLA policies, macros, checklist templates |
| `helpdesk-inbox` | Inbox list, inbox ticket detail, move, spam |
| `admin` | User management, system settings, backup |
| `ticket-api` | Create card, add comment, move card (API key auth) |

---

## 1. Authentication

### JWT (browser / SPA)

Most endpoints require a valid JWT access token passed as a Bearer token in the
`Authorization` header:

```
Authorization: Bearer <access_token>
```

Tokens are issued by `POST /api/v1/auth/login` and expire after 15 minutes.
The frontend refreshes them silently via `POST /api/v1/auth/refresh` using the
7-day refresh token.

### API Keys (automation / CI-CD)

API keys are personal (per user, not per project). Generate one in the UI under
**Project Settings → API Keys**, or via the API while authenticated with a JWT:

```bash
curl -X POST http://localhost:8080/api/v1/auth/api-keys \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-ci-key"}'
```

The response includes the full key (prefixed `cwk_...`) — **it is only shown
once**. Subsequent list calls only return the key prefix.

Pass the key in one of two ways:

```
X-API-Key: <key>
```
or as a query parameter:
```
GET /api/v1/projects/my-project/cards/1?api_key=<key>
```

API keys work on all authenticated endpoints, not just the Ticket API.

---

## 2. Ticket API

The Ticket API lets CI/CD pipelines and external tools create cards, add
comments, and move cards without a user account. All endpoints sit under
`/api/v1/ticket/` and require API key authentication.

### Create a card

```
POST /api/v1/ticket/{projectSlug}/cards
```

**Body**

```json
{
  "title":       "Deploy v1.2.3 to production",
  "description": "Automated deploy triggered by tag v1.2.3",
  "column_id":   5,
  "priority":    "high"
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `title` | string | yes | Card title |
| `description` | string | no | Markdown body |
| `column_id` | number | no | Target column; defaults to first column |
| `priority` | string | no | `none` / `low` / `medium` / `high` / `critical` |

**Response** `201 Created`

```json
{
  "id":          42,
  "card_number": 17,
  "title":       "Deploy v1.2.3 to production",
  "column_id":   5,
  "project_id":  3
}
```

### Add a comment

```
POST /api/v1/ticket/{projectSlug}/cards/{cardId}/comments
```

**Body**

```json
{
  "body": "Pipeline passed all tests. Deployment started."
}
```

**Response** `201 Created` — the created comment object.

### Move a card to a column

```
PATCH /api/v1/ticket/{projectSlug}/cards/{cardId}/move
```

**Body**

```json
{
  "column_id": 8,
  "position":  1000
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `column_id` | number | yes | Target column ID |
| `position` | number | no | Sort order within the column; omit to append at the end |

**Response** `200 OK`

```json
{ "ok": true }
```

### Example: full CI pipeline workflow

```bash
API_KEY="cwk_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6"
BASE="https://warmdesk.example.com/api/v1/ticket"
PROJECT="my-project"

# 1. Create a deploy card
CARD=$(curl -s -X POST "$BASE/$PROJECT/cards" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"title":"Deploy v1.2.3","priority":"high"}')

CARD_ID=$(echo $CARD | jq .id)

# 2. Move it to "In Progress"
curl -s -X PATCH "$BASE/$PROJECT/cards/$CARD_ID/move" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"column_id": 5}'

# 3. After tests pass, comment and move to Done
curl -s -X POST "$BASE/$PROJECT/cards/$CARD_ID/comments" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"body":"All tests passed. Deployed to production."}'

curl -s -X PATCH "$BASE/$PROJECT/cards/$CARD_ID/move" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"column_id": 8}'
```

---

## 3. Helpdesk — Ticket Messages

All ticket message endpoints are under `/api/v1/customers/:customerId/tickets/:ticketId/messages` and require the `helpdesk_enabled` feature flag (or the `customer` global role) and customer access.

### Post a message

```
POST /api/v1/customers/{customerId}/tickets/{ticketId}/messages
```

**Body**

```json
{
  "body":       "We have reproduced the issue and are working on a fix.",
  "is_private": false
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `body` | string | yes | Message text; Markdown supported |
| `is_private` | bool | no | Default `false`. When `true`, the message is an internal note: not emailed to the ticket's original sender and hidden from users with the `customer` global role. Users with the `customer` global role cannot set this field — it is silently reset to `false`. |

**Response** `201 Created` — the created message object.

```json
{
  "id": 42,
  "ticket_id": 7,
  "user_id": 3,
  "body": "We have reproduced the issue and are working on a fix.",
  "from_name": "",
  "email_sent": true,
  "is_private": false,
  "created_at": "2025-06-01T10:15:00Z",
  "updated_at": "2025-06-01T10:15:00Z",
  "user": { "id": 3, "display_name": "Sarah Chen", "username": "demo.sarah" },
  "attachments": []
}
```

`email_sent` is `true` when an email reply was dispatched to the ticket's original sender. It is always `false` for private messages.

### Inbox messages

The same body schema applies to inbox (unassigned) tickets:

```
POST /api/v1/tickets/inbox/{ticketId}/messages
```

### Access by role

| Global role | Can read messages | Can post messages | Can post private messages |
|---|---|---|---|
| `admin` | ✓ all | ✓ | ✓ |
| `user` / `viewer` | ✓ all | ✓ | ✓ |
| `customer` | ✓ public only | ✓ | ✗ (silently downgraded) |
| `metrics` / `backup` | ✗ | ✗ | ✗ |

---

## 4. Generic Webhook

The generic webhook accepts plain JSON and posts a formatted message to the
project chat. Use it for any custom automation that doesn't fit a specific git
platform.

### Setup

1. In **Project Settings → Webhooks**, click **Create Webhook**, set the type
   to **Generic (plain JSON)**, and give it a name.
2. Copy the token shown once on creation.

### Sending a message

```
POST /api/v1/webhooks/{token}
Content-Type: application/json
```

**Body**

```json
{
  "text":     "Build #42 passed in 2m 13s",
  "username": "CI Bot"
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `text` | string | yes | Message body; Markdown is supported |
| `username` | string | no | Bot display name; defaults to the webhook name |

**Response** `201 Created`

```json
{ "ok": true }
```

### Example: send from a shell script

```bash
curl -s -X POST https://warmdesk.example.com/api/v1/webhooks/TOKEN \
  -H "Content-Type: application/json" \
  -d '{"text":"**Build passed** — [view run](https://ci.example.com/build/42)","username":"GitHub Actions"}'
```

---

## 5. Git Platform Webhooks

Git platform webhooks do two things simultaneously:

1. **Post a formatted chat message** to the project chat whenever a push,
   pull request, merge request, or issue event arrives.
2. **Create card links** when a commit message, PR title, or issue title
   contains a card reference (e.g. `PRJ-42`). The linked event appears in the
   **Git Links** section of the card detail.

### 5.1 Gitea / Forgejo

**Endpoint**

```
POST /api/v1/gitea-webhook/{token}
```

**Setup**

1. Create a webhook of type **Gitea / Forgejo** in Project Settings.
2. In your Gitea or Forgejo repository, go to
   **Settings → Webhooks → Add Webhook → Gitea** and set:
   - Target URL: `https://warmdesk.example.com/api/v1/gitea-webhook/{token}`
   - Content type: `application/json`
   - Secret: leave empty (the token in the URL is sufficient)
   - Trigger: select all events, or at minimum: *Push*, *Issues*,
     *Pull Request*, *Issue Comment*, *Pull Request Review Comment*,
     *Create*, *Delete*, *Release*

**Supported events**

| `X-Gitea-Event` | Chat message content |
|-----------------|----------------------|
| `push` | Pusher, branch, commit list (up to 5) |
| `issues` | Author, action (opened/closed/etc.), issue title |
| `issue_comment` | Commenter, issue reference, comment preview |
| `pull_request` | Author, action, PR title |
| `pull_request_review_comment` | Reviewer, PR title, comment preview |
| `create` | Author, ref type, ref name |
| `delete` | Author, ref type, ref name |
| `release` | Author, release name |
| `fork` | Forker |

**Signature verification**

Gitea/Forgejo optionally sign payloads with HMAC-SHA256 in the
`X-Gitea-Signature` / `X-Forgejo-Signature` header. If no signature header is
present the request is still accepted (the URL token authenticates it). If a
signature is present it is verified against the webhook token.

---

### 5.2 GitHub

**Endpoint**

```
POST /api/v1/github-webhook/{token}
```

**Setup**

1. Create a webhook of type **GitHub** in Project Settings.
2. In your GitHub repository, go to
   **Settings → Webhooks → Add webhook** and set:
   - Payload URL: `https://warmdesk.example.com/api/v1/github-webhook/{token}`
   - Content type: `application/json`
   - Secret: leave empty (or enter any string — currently not validated)
   - Events: choose **Let me select individual events** and enable at minimum:
     - Pushes
     - Pull requests
     - Issues

**Supported events**

| `X-GitHub-Event` | Chat message content |
|------------------|----------------------|
| `push` | Pusher, branch, commit list (up to 5) |
| `pull_request` | Author, action (opened / closed / merged), PR title |
| `issues` | Author, action (opened / closed / etc.), issue title |
| `create` | Author, ref type, ref name |
| `delete` | Author, ref type, ref name |
| `ping` | Acknowledged silently (no chat message) |

**Card link extraction**

Card links are created from:
- Each commit message in a `push` event
- The PR title in a `pull_request` event
- The issue title in an `issues` event

The PR status is set to `merged` if `pull_request.merged` is `true`; otherwise
`open` or `closed` reflects `pull_request.state`.

---

### 5.3 GitLab

**Endpoint**

```
POST /api/v1/gitlab-webhook/{token}
```

**Setup**

1. Create a webhook of type **GitLab** in Project Settings.
2. In your GitLab repository, go to **Settings → Webhooks** and set:
   - URL: `https://warmdesk.example.com/api/v1/gitlab-webhook/{token}`
   - Secret token: leave empty, or set it to the WarmDesk webhook token for
     extra validation (GitLab sends it in `X-Gitlab-Token`)
   - Trigger: enable at minimum **Push events**, **Merge request events**, and
     **Issues events**

**Supported events**

| `object_kind` | Chat message content |
|---------------|----------------------|
| `push` | Pusher, branch, commit list (up to 5) |
| `merge_request` | Author, action (open / close / merge), MR title |
| `issue` | Author, action (open / close / reopen), issue title |

**Card link extraction**

Card links are created from:
- Each commit message in a `push` event
- The merge request title (`object_attributes.title`) in a `merge_request` event
- The issue title (`object_attributes.title`) in an `issue` event

The merge request reference uses the internal IID (`!42`); issues use `#42`.

---

## 6. Card References

A card reference is a string in the format `PREFIX-NUMBER` where:

- `PREFIX` is the project's 1–10 uppercase letter or digit key (e.g. `PRJ`, `WEB`, `API`)
- `NUMBER` is the card's sequential number within the project (e.g. `42`)

Examples: `PRJ-1`, `WEBAPP-99`, `API-200`

The prefix is auto-generated from the project name when the project is created
(first letters of each word, padded to 3 characters). It is visible in the card
reference badge at the top of the card detail.

### Using references in git

Include a card reference anywhere in a commit message, PR title, or issue title
and WarmDesk will create a link automatically when the webhook event arrives:

```
git commit -m "Fix login redirect loop — closes PRJ-42"
git commit -m "WIP: PRJ-55 add pagination"
```

Multiple references in the same message are all linked:

```
git commit -m "PRJ-10 PRJ-11: refactor auth middleware"
```

### Card link data

Each linked event stores:

| Field | Contents |
|-------|----------|
| `platform` | `github`, `gitlab`, `gitea`, or `forgejo` |
| `link_type` | `commit`, `pr`, or `issue` |
| `reference` | Commit SHA (full) or PR/issue number |
| `title` | Commit first line or PR/issue title |
| `url` | Direct link to the event on the platform |
| `author` | Display name or username of the author |
| `status` | `open`, `closed`, or `merged` |
| `repo_name` | `owner/repo` full path |

---

## 7. Response Formats

### Success

All successful responses return JSON. Creation responses use `201 Created`;
queries and updates use `200 OK`.

### Errors

All error responses return JSON with an `error` key:

```json
{ "error": "project not found" }
```

Common status codes:

| Code | Meaning |
|------|---------|
| `400` | Bad request — missing or invalid field |
| `401` | Unauthorized — missing or invalid token / API key |
| `403` | Forbidden — insufficient role |
| `404` | Not found |
| `500` | Internal server error |

### Dates

All dates and timestamps are returned as ISO 8601 strings in UTC:
`2026-03-29T14:05:00Z`
