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
8. [Invoices](#8-invoices)
9. [Invoice Templates](#9-invoice-templates)
10. [Customer Contacts](#10-customer-contacts)

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
| `customers` | Invoice CRUD, invoice send/credit-note/PDF, customer contacts |
| `invoices` | Global invoice list |
| `invoice-templates` | List, create, update, delete invoice templates (admin) |
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

---

## 8. Invoices

Invoices are linked to customers. All endpoints require authentication and the
user must have access to the customer (`CustomerAccess` row, group-based access,
or the global `admin` role).

### List invoices for a customer

```
GET /api/v1/customers/{customerId}/invoices
```

Returns all invoices for the customer, newest first.

**Response** `200 OK`

```json
[
  {
    "id": 1,
    "customer_id": 3,
    "invoice_number": "INV-0001",
    "period_start": "2026-05-01",
    "period_end": "2026-05-31",
    "due_date": "2026-06-15",
    "status": "sent",
    "currency": "EUR",
    "vat_rate": 21.0,
    "subtotal": 1200.00,
    "vat_amount": 252.00,
    "total": 1452.00,
    "notes": "",
    "payment_method": "",
    "credited_invoice_id": null,
    "created_at": "2026-06-01T08:00:00Z",
    "updated_at": "2026-06-01T08:00:00Z"
  }
]
```

### Get a single invoice

```
GET /api/v1/customers/{customerId}/invoices/{invoiceId}
```

Returns the invoice with its full line items array.

### Create an invoice

```
POST /api/v1/customers/{customerId}/invoices
```

**Body**

```json
{
  "period_start":    "2026-05-01",
  "period_end":      "2026-05-31",
  "due_date":        "2026-06-15",
  "vat_rate":        21.0,
  "currency":        "EUR",
  "notes":           "May services",
  "line_items": [
    { "description": "Backend development", "quantity": 20, "unit_price": 95.00 },
    { "description": "DevOps consulting",   "quantity": 4,  "unit_price": 110.00 }
  ]
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `period_start` | string (date) | yes | `YYYY-MM-DD` |
| `period_end` | string (date) | yes | `YYYY-MM-DD` |
| `due_date` | string (date) | no | `YYYY-MM-DD` |
| `vat_rate` | float | no | Default 0 |
| `currency` | string | no | ISO 4217 code, e.g. `EUR` |
| `notes` | string | no | Free text; appears on the PDF |
| `line_items` | array | no | See line item fields below |

**Line item fields**

| Field | Type | Notes |
|-------|------|-------|
| `description` | string | Row label |
| `quantity` | float | Number of units |
| `unit_price` | float | Price per unit (exc. VAT) |

The invoice is created in `draft` status.

**Response** `201 Created` — the created invoice object.

### Update an invoice

```
PUT /api/v1/customers/{customerId}/invoices/{invoiceId}
```

Accepts the same body as create. All fields are optional (patch-style). Only
`draft` invoices may be edited — attempting to modify a `sent` or `paid` invoice
returns `403 Forbidden`.

Additional field available on update:

| Field | Type | Notes |
|-------|------|-------|
| `status` | string | `draft` / `sent` / `paid` / `credit_note` |
| `payment_method` | string | `bank` / `card` / `cash` / `other` — recorded when marking as paid |

### Delete an invoice

```
DELETE /api/v1/customers/{customerId}/invoices/{invoiceId}
```

Only `draft` invoices may be deleted. Returns `204 No Content` on success.

### Send an invoice (mark as sent)

```
POST /api/v1/customers/{customerId}/invoices/{invoiceId}/send
```

No body required. Changes the invoice status to `sent`.

**Response** `200 OK` — the updated invoice object.

### Create a credit note

```
POST /api/v1/customers/{customerId}/invoices/{invoiceId}/credit-note
```

Creates a new invoice that credits the original. The new invoice has:
- `status` set to `credit_note`
- `credited_invoice_id` set to the original invoice's ID
- Line items with negated amounts
- The original invoice is marked void (status set to `credit_note` as well)

No body required.

**Response** `201 Created` — the new credit-note invoice object.

### Download invoice PDF

```
GET /api/v1/customers/{customerId}/invoices/{invoiceId}/pdf?lang=en
```

Returns the PDF as `application/pdf`. The `lang` query parameter selects the
PDF language (default `en`; supported: `en`, `nl`, `de`, `fr`, `es`, `da`,
`sv`, `nb`, `fi`, `is`, `pt`, `it`).

### Global invoice list

```
GET /api/v1/invoices
```

Returns invoices across all customers the requesting user has access to.

**Query parameters**

| Parameter | Type | Notes |
|-----------|------|-------|
| `customer_id` | number | Filter to a single customer |
| `status` | string | Filter by status: `draft`, `sent`, `paid`, `credit_note` |

**Response** `200 OK` — array of invoice objects (same schema as above), each
with a nested `customer` object `{ id, name }`.

---

## 9. Invoice Templates

Invoice templates let admins define reusable sets of line items and default
settings that users can apply when creating a new invoice.

### List templates

```
GET /api/v1/invoice-templates
```

Returns all active templates. No admin role required — any authenticated user
may read templates.

**Response** `200 OK`

```json
[
  {
    "id": 1,
    "name": "Monthly Support",
    "line_items": "[{\"description\":\"Monthly support\",\"quantity\":1,\"unit_price\":500}]",
    "default_vat_rate": 21.0,
    "default_currency": "EUR",
    "notes": "Monthly retainer — 8 h included",
    "created_at": "2026-05-01T00:00:00Z",
    "updated_at": "2026-05-01T00:00:00Z"
  }
]
```

`line_items` is a JSON-encoded string. Parse it to get the array.

### Create a template (admin only)

```
POST /api/v1/admin/invoice-templates
```

**Body**

```json
{
  "name":              "Monthly Support",
  "line_items":        [{ "description": "Monthly support", "quantity": 1, "unit_price": 500 }],
  "default_vat_rate":  21.0,
  "default_currency":  "EUR",
  "notes":             "Monthly retainer — 8 h included"
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `name` | string | yes | Unique display name |
| `line_items` | array | no | Pre-defined line items (same schema as invoice line items) |
| `default_vat_rate` | float | no | Pre-fills the VAT rate field |
| `default_currency` | string | no | Pre-fills the currency field |
| `notes` | string | no | Default notes text for the invoice |

**Response** `201 Created` — the created template object.

### Update a template (admin only)

```
PUT /api/v1/admin/invoice-templates/{id}
```

Same body as create; all fields optional.

**Response** `200 OK` — the updated template object.

### Delete a template (admin only)

```
DELETE /api/v1/admin/invoice-templates/{id}
```

**Response** `204 No Content`

---

## 10. Customer Contacts

Contact persons are stored per customer and appear in the customer detail view.
They are informational — they are not WarmDesk user accounts.

### List contacts

```
GET /api/v1/customers/{customerId}/contacts
```

**Response** `200 OK`

```json
[
  {
    "id": 1,
    "customer_id": 3,
    "name": "Alice Bakker",
    "email": "alice@example.com",
    "phone": "+31 20 123 4567",
    "department": "Procurement",
    "is_primary": true,
    "created_at": "2026-05-01T00:00:00Z",
    "updated_at": "2026-05-01T00:00:00Z"
  }
]
```

### Create a contact

```
POST /api/v1/customers/{customerId}/contacts
```

**Body**

```json
{
  "name":       "Alice Bakker",
  "email":      "alice@example.com",
  "phone":      "+31 20 123 4567",
  "department": "Procurement",
  "is_primary": true
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `name` | string | yes | Contact full name |
| `email` | string | no | Email address |
| `phone` | string | no | Phone number |
| `department` | string | no | Department or job title at the customer |
| `is_primary` | bool | no | Whether this is the primary contact; setting to `true` clears the flag on all other contacts for this customer |

**Response** `201 Created` — the created contact object.

### Update a contact

```
PUT /api/v1/customers/{customerId}/contacts/{contactId}
```

Same body as create; all fields optional.

**Response** `200 OK` — the updated contact object.

### Delete a contact

```
DELETE /api/v1/customers/{customerId}/contacts/{contactId}
```

**Response** `204 No Content`
