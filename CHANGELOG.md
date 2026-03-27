# Changelog

## Unreleased

### Features

#### Internationalization
- Added German (`de`), French (`fr`), and Spanish (`es`) translation files with full coverage of all 191 UI keys
- Registered all three new locales in `i18n/index.js` alongside the existing English and Dutch
- Extended locale validation in backend (`auth.go`, `system.go`) to accept `de`, `fr`, and `es`
- Added **Default Language** setting to the global system settings (backend + Admin UI), allowing admins to control the default locale for newly registered users
- New users now inherit the system-configured default locale instead of always getting `en`

#### Chat UI — modern redesign
- Rewrote `ChatPanel.vue` with a Limitless-style message layout: chat bubbles, own messages right-aligned, received messages left-aligned
- Added avatar images with initials fallback for each participant
- Added date separators between message groups (e.g. "Today", "Yesterday", full date for older messages)
- Replaced EasyMDE editor in the chat compose area with a plain auto-resizing textarea
- Made the chat panel **resizable**: drag the left edge to adjust width between 260 px and 720 px
- Applied the same bubble/date-separator/avatar pattern to `DirectMessagesView.vue`
- Added a conversation sidebar with an active indicator and empty state to `DirectMessagesView.vue`
- Two lines added to router.go — one for `/favicon.svg` and one for `/logo.svg`.

#### Direct Messages — sidebar navigation
- Clicking a username in the app sidebar now correctly opens a direct message conversation
- The sidebar filters out the currently logged-in user so you cannot click on yourself
- The `watch` on `route.query.user` is now async and re-fetches the user list if it hasn't loaded yet, fixing a race condition on fresh navigation

#### Project members — multi-select invite
- The "Invite Members" user picker in Project Settings is now a searchable multi-select checkbox list
- Selecting multiple users and clicking Invite sends individual invitations for each selected user in sequence
- A selection counter shows how many users are currently selected

#### Group / multi-user Direct Messages
- Introduced three new backend models: `Conversation`, `ConversationMember`, `ConversationMessage` — auto-migrated on startup
- New REST API under `/api/v1/conversations`: create, list, get messages, send, delete message, add member
- Creating a 1-on-1 conversation deduplicates: opening the same partner twice reuses the existing thread
- **New conversation panel** in `DirectMessagesView.vue` supports selecting multiple users with chip UI; shows an optional group name field when 2+ users are selected
- Group conversations display a people icon in the sidebar; 1-on-1 conversations show the other user's avatar
- Chat header shows all member names as subtitle; group chats have an **Add Member** button with an inline user search
- Sender name and avatar are shown for every participant, making group threads readable
- Messages are persisted to the database and polled every 5 seconds so all participants see new messages without a manual refresh; scroll position is preserved when reading history
- After sending a message a fresh fetch runs immediately to pick up concurrent messages
- GitHub Markdown rendered in all conversation messages using `marked` + `DOMPurify` (same as project chat and card comments); own-message bubble overrides keep code blocks and blockquotes readable on the primary colour background

### Bug Fixes
- `User.LastSeenAt` field renamed to `LastLoginAt` (JSON: `last_login_at`) to match the requirements; updated in model, login handler, Settings view, and Admin view
- Fixed one remaining `last_seen_at` reference in `AdminView.vue` that was missed in the initial rename
