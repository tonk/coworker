- ~~Bugfix:
  On Fedora 43 this setting must be done to not show the blank screen:
  LD_PRELOAD=/usr/lib64/libwayland-client.so
  Client was built on Ubuntu 24.04. Could be that this library is
  LD_PRELOAD=/usr/lib/libwayland-client.so on Ubuntu
  Fix this, once and for all~~ **Done in v0.5.3**

- ~~In the client, an issue created in Forgejo is displayed on the card, but
  when clicked, nothing happens. The link should be opened in a browser.
  This works in the browser version.~~ **Done in v0.5.2**

- ~~When a Forgejo webhook is configured and triggered, the card shows the
  Gitea logo. This is not really an error, but can be better.
  Show the Forgejo logo when the event comes from Forgejo.~~ **Done in v0.5.2**

- ~~In the project overview, allow projects to be moved, change the order
  Drag-to-reorder on dashboard (admin only)~~ **Done in v0.5.2**

- ~~After zooming in or out, save the zoom setting for the next session~~ **Already implemented (localStorage)**

- ~~Currently the theme is blue, also allow, per user, a red, green,
  orange, black and white theme (Save buttons, highlight lines in the
  sidebar, etc). This, of course, needs to be a user setting~~ **Done in v0.6.3 (blue, red, green, orange — black and white skipped)**

- ~~In the sidebar, allow for re-ordering the starred customers and
  projects~~ **Done in v0.6.1**

- ~~In the sidebar show all customers~~ **Done in v0.6.1**

- ~~In the cards, the date-pickers for the "Start Date" and "Due Date"
  don't work~~ **Done in v0.6.1**

- ~~In the contract editor the date fields should follow the configured
  date format~~ **Done in v0.6.1**

- ~~In the client, when a new version is available and the "Release notes"
  are clicked, nothing happens.
  This works in the browser version.~~ **Done in v0.5.2**

- ~~In the chat show an indication when someone else in the chat is typing
  and who it is~~ **Done in v0.5.2**

- ~~Add a `/api/v1/metrics`, non public, API endpoint for Prometheus
  monitoring, and create a user type that is only allowed to read these
  metrics. User type should be "metrics"~~ **Done in v0.5.2**

- ~~Re-add '@..' mentions in cards, the description and comments, for users
  and for teams, this worked before. This does work in the chat~~ **Done in v0.5.2**

- ~~Add cross reference between cards~~ **Done in v0.6.0**

- ~~Pressing "Esc" in an open card, should close the card, if no changes
  where made~~ **Done in v0.5.2**

- ~~Pressing "Cancel" with changes in the card should show a pop-up with
  "Save" and "No save"~~ **Done in v0.5.2**

- Investigate (partially improved, residual lag remains):
  When starting the Windows client it takes rather long before
  it responds, it does show the login screen, but typing lags behind and
  is very slow, for userid/password entry. After logging in, performance
  is OK.
  Fixes applied so far:
  * Passive keydown listener for zoom handler (eliminated username-field lag)
  * `::-ms-reveal { display: none }` (suppressed WebView2 reveal-button IPC on focus)
  * `spellcheck="false" autocorrect="off" autocapitalize="off"` on both inputs
  * `ICoreWebView2Settings4::SetIsGeneralAutofillEnabled(false)` +
    `SetIsPasswordAutosaveEnabled(false)` via Rust `with_webview` (disabled
    autofill IPC at the WebView2 engine level)
  Still some lag — root cause not yet identified.

- ~~Change projects to add an extra layer
  so that we can have "Customer / Projects" and between the "Customer / Projects"
  there should be the option for "Contract". So:
  * "Customer / Projects"
  * "Customer / Contract / Projects"~~ **Done in v0.5.3**

- ~~Also add 'Customers' and "Favorite customers" to the sidebar
  with, of course, the possibility to "Star" and "Un-star"  customer.~~ **Done in v0.5.3**

- ~~Add indicator to card, to show that sub-cards are present~~ **Already implemented (BoardCard.vue)**

- ~~"Edit customer" shows empty fields~~ **Fixed: openEdit() was never called**

- ~~Switching customers in the sidebar doesn't update the overview on the
  right hand side~~ **Fixed: added watch(custId) to CustomerDetailView**

- ~~Add customers and contracts, demo data, to `warmdesk-seed`~~ **Done**

- ~~Implement sub-cards in the cards~~ **Done in v0.5.3**

- ~~Implement a Gantt chart per project~~ **Done in v0.6.0**

- ~~Change the Github workflow to produce a single binary server,
  `warmdesk`, without the `web` directory.~~ **Done — both build-server.yml and release.yml now embed the frontend with `-tags embed`; no web/ directory shipped**

- ~~Move creation of Customers to the "Admin Panel", just after groups
  No more creation of customers in the sidebar. So, only Admins can
  create new customers.
  The contracts move with this as well~~ **Done — Admin Panel now has a Customers tab (after Groups); customer creation restricted to admins only**

- ~~Project Settings -> Members: Also show the assigned groups and their
  Role~~ **Done — groups with access shown below direct members in Project Settings → Members**

- ~~Customer Settings -> Members: Also show the assigned groups and their
  Role~~ **Done — groups with access shown below direct members in Customer Detail → Members**

- ~~Pasting a picture in the chat, in the client (Linux) doesn't work
  It does in the web interface.
  When I paste it in the web interface, it is displayed in the client.~~ **Done in v0.8.10**

- ~~When pasting or typing a URL in the chat, generate a preview, if the
  site can be reached.~~ **Done in v0.8.10**

- ~~Enable syntax highlighting in the chat, when code is entered.~~ **Done in v0.8.10**

- ~~In the chat fix code colors, as they look bad, right now (weird
  background on the text). This happens in `Bubble`, `Comfortable` and
  `Cozy`.~~ **Done in v0.8.10**

- ~~With the backup/restore function, add an option to send an email after
  the backup is done. This has to have, of course, a tick-box to switch
  this functionality on and off and a box to add the e-mail address
  where the mails should be send.
  This mail should contain:
    - Date / time
    - Backup successful or not
    - List of all available backups~~ **Done**

- ~~Also add to the backup, API metrics for:
    - Date / time of last backup
    - Backup successful or not
    - List of all available backups~~ **Done**

- ~~In the "Edit Sprint" card the date fields should follow the configured
    the users date format setting.~~ **Done**

- ~~The Admin panel -> Projects also shows all deleted projects and an
  option to restore them. Would it be possible to remove them
  completely, so add an extra option.
  I can imagine that this will create problems, so investigate first
  if problems could arise.~~ **Done — "Permanently Delete" button added next to Restore; purges all project data including cards, columns, labels, chat, sprints and their associations**

- ~~In the client, when a new chat arrives and the client doesn't have
  focus, show a notification that a new message has arrived (through the
  OS notification system, if possible)~~ **Done in v0.8.12**

- ~~In the chat and edit boxes, if typing a ':' show a popup with
  emojis and let the user pick one, or continue typing.~~ **Done in v0.8.11**

- ~~In the chat, when I type, or select, an emoji, it is displayed
  differently in the chat history.~~ **Done in v0.8.11**

- ~~The backup scheduler "Start time" should follow configured date/time
  formats~~ **Done in v0.8.11**

- ~~In the "Edit User" box, show the users avatar at the top left corner.~~ **Done in v0.8.11**

- ~~In the chat, currently the avatar is just below the message. Place it to the left, or above the message.~~ **Done in v0.8.11**

- ~~In the "Invite Member" form of a project, change the initials blob to
  the users avatar. Also show the members avatar next to their name in
  the members list.~~ **Done in v0.8.11**

- ~~In the Admin -> Users list, also show the users avatar.~~ **Done in v0.8.11**

- ~~In the Admin -> Customers list, also show the customers avatar.~~ **Done in v0.8.11**

- ~~In the Admin -> Users list, add three columns:
  - Last successful login
  - Last password change
  - MFA enabled~~ **Done — last_login already shown; added password_changed_at column and dedicated MFA column (moved from Status badge)**

- ~~In the Admin -> Settings add a setting to set a required password change
  period, 0 is disabled.~~ **Done — password_change_period_days setting added; login response includes password_expired flag and redirects to settings**

- ~~In the "People" list, only show people that are a member of the
  customers that I am also a member of.~~ **Done — ListAllUsers filters by shared customer access (direct + group); admins and users without explicit customer access still see everyone**

- ~~In the chat, place the time of the chat behind the username~~ **Done**

- ~~In the chat, the response popup should only be shown when you click on
  the message~~ **Done**

- ~~Make sure you can't call someone who isn't online~~ **Done in v0.12.2**
- ~~Add row-level comments to the time tracking grid (with PDF display)~~ **Done in v0.12.2**

---

## Group video/audio calls with LiveKit

Right now WarmDesk can only call one person at a time (1-on-1). This works by sending
your voice and video directly from your computer to the other person's computer — no
middleman needed.

For **group calls** (3+ people at the same time) that direct approach falls apart: everyone
would need to send their video to everyone else simultaneously, which hammers your upload
speed and melts slower computers. The solution is a **media server** — a computer in the
middle that receives everyone's streams and re-sends only what each person needs to see.

**LiveKit** is a free, open-source media server that does exactly this. Here is what you
would need to get it running with WarmDesk.

---

### The server (hardware / VM)

You need a separate machine — either a cheap VPS (virtual server you rent) or a physical
box you already own. Minimum requirements:

| What         | Minimum (≤ 10 people per call)  | Comfortable (≤ 30 people per call) |
|---|---|---|
| CPU          | 2 cores                          | 4 cores                            |
| RAM          | 2 GB                             | 4 GB                               |
| Disk         | 10 GB (LiveKit itself is tiny)   | 10 GB                              |
| Network      | 100 Mbps upload **dedicated**    | 500 Mbps upload                    |
| OS           | Ubuntu 22.04 LTS (recommended)  | Ubuntu 22.04 LTS                   |

> **The network is the bottleneck, not the CPU.** A cheap €5/month VPS with a shared 1 Gbps
> port from Hetzner, DigitalOcean, or Vultr works fine for small teams.

This machine does **not** have to be the same machine that runs WarmDesk. In fact it is
better if it is not — LiveKit is chatty on the network and you do not want it competing
with your web traffic.

---

### What needs to be installed on that server

1. **LiveKit server** — one binary, runs as a systemd service. Download from
   `https://github.com/livekit/livekit/releases` or run it as a Docker container.

2. **A domain name pointing at that server** — LiveKit must be reachable over HTTPS
   (encrypted). You cannot use a bare IP address because browsers refuse unencrypted
   camera/microphone streams. Something like `livekit.yourcompany.com` is fine.

3. **An SSL certificate** — free from Let's Encrypt (`certbot`). Takes 5 minutes to set up.

4. **Open firewall ports:**
   - TCP 443 — secure web traffic (HTTPS / WebSocket)
   - TCP 7881 — fallback for video when UDP is blocked
   - UDP 50000–60000 — the actual video/audio streams (this range is configurable)

---

### What needs to change inside WarmDesk

This is developer work, not server work. A rough list of what has to be built:

**Backend (Go):**
- Add a LiveKit API key and secret to `warmdesk.yaml` (two new config fields).
- Add one new API endpoint — something like `POST /api/v1/projects/:slug/livekit-token`.
  When a user wants to join a group call, they ask WarmDesk for a short-lived *token*
  (a signed ticket that proves who they are). WarmDesk generates this using the LiveKit
  Go SDK and hands it back. No token = no entry to the call room.

**Frontend (Vue):**
- Replace (or extend) the current `useWebRTCCall.js` composable with the
  **LiveKit JavaScript SDK** (`@livekit/components-react` or the lower-level
  `livekit-client` npm package).
- Build a group-call UI: a grid of video tiles, one per participant, that appears
  either as a full-screen overlay or as a floating panel.
- Wire up the existing "start call" button so that, for group/channel calls, it joins
  a LiveKit room instead of dialling a single person directly.

The existing 1-on-1 call system (`useWebRTCCall.js`) can stay as-is for direct
person-to-person calls — it does not need LiveKit and costs nothing to keep.

---

### Summary in one paragraph

Rent a small Linux VPS (2 CPU, 2 GB RAM, good network), point a subdomain at it, get a
free SSL certificate, and install the LiveKit binary. Open the right firewall ports. Add
the LiveKit address and a secret key to WarmDesk's config file. Then a developer needs
to write roughly 200–400 lines of Go + Vue code to generate call tokens on the backend
and show a multi-person video grid on the frontend. That is it — no licences, no monthly
fees beyond the VPS cost (~€5–15/month depending on provider).
