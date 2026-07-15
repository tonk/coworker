# Pre-release testing checklist

Covers everything changed since `v0.12.41` (commits `4e302a4`..`dfee727`), ahead of the next release.
Not committed to the repo by default — this is a working checklist for manual QA. Delete it, or ask to have it committed, once testing is done.

---

## 1. Time tracking — chart tooltips

Area: `frontend/src/views/TimeTrackingView.vue`, Time Tracking → Report tab → Chart view.

- [ ] Bar chart: hover a bar — bold line shows the customer, second line shows `activity: time` with the color swatch (no duplicate activity name)
- [ ] Pie chart: same check on a pie slice
- [ ] Stacked bar chart: hover a segment — bold line shows the customer for that activity, **not** the period (period is already on the x-axis)
- [ ] Activity that has entries from more than one customer — bold line shows a comma-joined list of customers, not just one
- [ ] Entry with no customer set — bold line shows "No customer" (or your locale's translation) rather than blank
- [ ] "Other" bucket slice (when more than 7 activities exist) — tooltip doesn't error, customer line is blank/sensible

## 2. Invoices — credit notes & revenue summary

Area: `frontend/src/views/InvoicesView.vue`, Time Tracking → Invoices tab.

- [ ] Create and send an invoice — confirm it appears in **Total Invoiced** and **Outstanding**
- [ ] Issue a credit note against that sent invoice — confirm **Total Invoiced** and **Outstanding** both drop back down by the credited amount (previously they didn't move at all)
- [ ] Mark an invoice **Paid**, then credit it — confirm **Paid** total drops accordingly
- [ ] Confirm the *original* invoice's status is unchanged after crediting (still `Sent`/`Paid`, not flipped to `Credit note`) and it still shows the "↩ Credit note issued" badge
- [ ] Overdue invoice — credit it, confirm **Overdue** total also nets down
- [ ] Global invoice list filters (by customer, by status) still work correctly after netting change

## 3. Admin — require password change on next login

Area: `backend/handlers/{auth,passkey,user}.go`, `frontend/src/views/{AdminView,LoginView,SettingsView}.vue`.

- [ ] **New User**: check "Require password change on next login", create the user, log in as them with a normal password → redirected to Settings → Security with the banner, can't just navigate away and ignore it (banner is there but confirm at least it's visible immediately)
- [ ] Same, but the new user logs in via **passkey** (register a passkey for them first, or test with an existing passkey user flagged via Edit User) → same redirect/banner
- [ ] Same, but the user has **MFA enabled** → complete the MFA challenge, then confirm the redirect/banner still fires after MFA verify (this path required a separate fix — worth extra attention)
- [ ] **Edit User**: check the box on an *existing* user who's already logged in before → their *next* login (not current session) is forced
- [ ] After the flagged user changes their password (self-service, via Settings → Security) → flag clears; logging out and back in no longer shows the banner
- [ ] Flag also clears when the password is changed via **admin-forced reset** (Edit User → set a new password) — confirm this specifically, since it's a different code path than self-service
- [ ] Flag also clears via the **"Forgot password" email reset link** flow
- [ ] Uncheck the box in Edit User for a currently-flagged user → requirement is lifted, next login is normal
- [ ] Regular login (flag never set) is completely unaffected — no banner, no redirect

## 4. Admin — password generator

Area: `frontend/src/views/AdminView.vue`, New User / Edit User modals.

- [ ] New User form: password field is visibly narrower with 🎲 and 📋 buttons at the right edge, aligned to the same height as the input (not shorter/taller)
- [ ] Click 🎲 — field fills with a password ≥30 characters (or longer if admin's configured minimum exceeds 30), containing at least one uppercase, lowercase, digit, and special character
- [ ] Click 📋 — password is copied to the clipboard (paste somewhere to confirm); success toast shown
- [ ] Click 📋 with an empty password field — error toast shown, nothing copied
- [ ] Hover each button — native tooltip text appears; tab to each button with keyboard — screen reader / accessibility tree shows a real label (not just the emoji)
- [ ] Same four checks repeated in the **Edit User** form
- [ ] Generated password is accepted by the `minlength="8"` client-side validation and by the backend on submit
- [ ] ⚠️ Known gap, not fixed in this cycle: a password an admin **types manually** in these forms is *not* checked against the configured password policy (only an 8-char minimum). Confirm this is still the case and decide if it should block release.

## 5. Deployment — systemd service

Area: `deploy/warmdesk.service`.

- [ ] Fresh install following `docs/admin-guide.adoc`'s systemd steps exactly (`mkdir -p /opt/warmdesk/{data,uploads}`, copy the service file, start it)
- [ ] Upload a file (card attachment) — confirm it succeeds and lands in `/opt/warmdesk/uploads`, with `ProtectSystem=strict` still in effect
- [ ] Confirm the service still starts cleanly and the database path (`/opt/warmdesk/data/warmdesk.db`) is unaffected

## 6. Localization — all 11 locales

Area: `frontend/src/i18n/*.json` (large backfill + translation sweep across `nl`, `de`, `fr`, `es`, `da`, `sv`, `nb`, `fi`, `is`, `pt`, `it`).

For **each** of the 11 locales, switch the UI language and spot-check:

- [ ] Admin → Users → New/Edit User modal, including the new "Require password change on next login" checkbox and Generate/Copy button tooltips
- [ ] Admin → Settings → SLA Policies tab (previously 100% untranslated)
- [ ] Admin → Settings → Feature Access toggles (Board / Chat / Helpdesk / Time tracking labels)
- [ ] A ticket detail view — status labels, priority labels, type labels, "Linked Tickets", message composer placeholder, reminder/close-date controls (previously ~35% English)
- [ ] Time tracking ticket time-logging area (Log time, Duration, Note, Total — previously all English)
- [ ] No raw `key.path` strings visible anywhere in the areas above (sign of a missed/mistyped key)
- [ ] No obviously-broken layout from longer translated strings (German/Finnish especially — check the password field + button row, and any tight admin-panel columns)
- [ ] Login page password-expiry/must-change-password banner text renders correctly (this flow was dead code before this cycle — first real end-to-end test of the translated string)

## 7. Website / release artifacts

- [ ] `cd website && hugo server -D`, confirm the new release blog post renders and the homepage "What's new" strip shows the updated highlights
- [ ] Confirm `website/hugo.toml`'s `warmdesk-version` attribute matches the actual release tag once cut
- [ ] Update banner (`useUpdateCheck.js`) — use the DevTools sessionStorage injection trick from `CLAUDE.md` to confirm the banner still triggers correctly against the new version number

## 8a. Time tracking — week-switching row stability (v0.14.3–v0.14.6)

Area: `frontend/src/views/TimeTrackingView.vue` (`loadWeek()`, `_scheduleSaveOrder`/`_flushSaveOrder`, `loadTimeEntryRowOrderKeys` in `backend/handlers/time_entry.go`).

- [ ] Rapidly click the week next/prev arrows several times in a row — no week ever shows rows/entries left over from a different week, and every cell's date matches the displayed week
- [ ] Visit a week you've never customized the row order for — it starts with only its own actual logged entries, not the row layout from whatever week you used last
- [ ] Rearrange rows in one week, then jump to a different, unrelated week and back — the rearranged order is still there and no rows bled into the other week
- [ ] **Delete a stray/empty row**, then immediately switch to another week and back within ~1 second — confirm the deleted row does **not** reappear (this is the specific regression reported and fixed in v0.14.6)
- [ ] Delete a row, then immediately navigate away from Time Tracking entirely (e.g. to the Boards view) before returning — confirm the deletion still persisted (tests the `onUnmounted` flush)
- [ ] Add/edit a row-level comment, switch weeks quickly a few times, come back — comment is preserved on the correct week only
- [ ] If you have older data from before v0.14.5: any week previously affected by stray/duplicated rows should now show only rows with real logged time (verify against the cleanup SQL applied to production, if relevant to your data)

## 8b. Admin — backup list, download, restore

Area: `frontend/src/views/AdminView.vue` (backup tab), `backend/handlers/backup.go`.

- [ ] Admin → Settings/Backup tab → backup list shows a real date/time in the date column for every existing backup (was always blank before v0.14.6)
- [ ] Click **Download** on an existing backup — file downloads successfully with a sensible filename (was failing with "Failed to download backup" before v0.14.6)
- [ ] Click **Restore** on a backup — confirmation dialog appears (destructive warning), confirming restores the database; cancelling the dialog does nothing (button previously did nothing at all, silently)
- [ ] Create a fresh backup, confirm it appears in the list immediately with a correct date, and can itself be downloaded
- [ ] Delete a backup — still works as before (unaffected by this fix, but confirm no regression)

## 9. Regression / build health

- [ ] `cd backend && go build ./... && go vet ./... && go test ./...`
- [ ] `cd frontend && npm run build && npx vitest run`
- [ ] `make docs-pdf-guides` builds both guide PDFs without errors
- [ ] Full manual smoke test of a board + chat + time tracking + helpdesk flow in **English**, to confirm nothing in the large i18n file rewrite broke key lookup for the default locale
