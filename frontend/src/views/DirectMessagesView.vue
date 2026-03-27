<template>
  <div class="dm-layout">

    <!-- ── Sidebar ──────────────────────────────────────── -->
    <aside class="dm-sidebar">

      <div class="dm-sidebar-header">
        <h2>{{ $t('dm.title') }}</h2>
        <button class="new-chat-btn" @click="toggleNewConv" :class="{ active: showNewConv }" title="New conversation">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        </button>
      </div>

      <!-- New conversation panel: multi-select users -->
      <div v-if="showNewConv" class="new-conv-panel">
        <div class="search-wrap">
          <svg class="search-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input
            class="search-input"
            v-model="userSearch"
            :placeholder="$t('dm.search_users')"
            @input="filterUsers"
            autofocus
          />
        </div>

        <!-- Selected user chips -->
        <div v-if="selectedUsers.length" class="selected-chips">
          <span v-for="u in selectedUsers" :key="u.id" class="chip">
            {{ u.display_name || u.username }}
            <button class="chip-remove" @click="toggleUser(u)">×</button>
          </span>
        </div>

        <!-- Group name field (only when 2+ users selected) -->
        <input
          v-if="selectedUsers.length > 1"
          class="group-name-input"
          v-model="newGroupName"
          placeholder="Group name (optional)"
        />

        <div class="user-search-results">
          <div
            v-for="u in filteredUsers"
            :key="u.id"
            :class="['user-result', { selected: isSelected(u) }]"
            @click="toggleUser(u)"
          >
            <div class="conv-avatar" :style="avatarBg(u)">
              <img v-if="getAvatar(u)" :src="getAvatar(u)" class="avatar-img" @error="e => e.target.style.display='none'" />
              <span v-else class="avatar-initials">{{ initials(u) }}</span>
            </div>
            <div class="conv-info">
              <div class="conv-name">{{ u.display_name || u.username }}</div>
              <div class="conv-handle">@{{ u.username }}</div>
            </div>
            <svg v-if="isSelected(u)" class="check-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
          </div>
          <div v-if="!filteredUsers.length" class="search-empty">No users found</div>
        </div>

        <button
          class="start-conv-btn"
          :disabled="!selectedUsers.length"
          @click="startConversation"
        >
          {{ selectedUsers.length > 1 ? 'Start group chat' : 'Open chat' }}
        </button>
      </div>

      <!-- Conversation list -->
      <div class="conv-list">
        <div
          v-for="conv in conversations"
          :key="conv.id"
          :class="['conv-item', { active: activeConvId === conv.id }]"
          @click="openConversation(conv)"
        >
          <!-- Avatar: stacked for groups, single for 1-on-1 -->
          <div class="conv-avatar-wrap">
            <template v-if="conv.is_group">
              <div class="group-avatar">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
              </div>
            </template>
            <template v-else>
              <div class="conv-avatar" :style="avatarBg(otherMember(conv))">
                <img v-if="getAvatar(otherMember(conv))" :src="getAvatar(otherMember(conv))" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else class="avatar-initials">{{ initials(otherMember(conv)) }}</span>
              </div>
            </template>
          </div>

          <div class="conv-info">
            <div class="conv-name">{{ convDisplayName(conv) }}</div>
            <div class="conv-handle">
              {{ conv.is_group ? memberList(conv) : ('@' + (otherMember(conv)?.username || '')) }}
            </div>
          </div>
          <div v-if="activeConvId === conv.id" class="conv-active-dot"></div>
        </div>

        <div v-if="!conversations.length && !showNewConv" class="conv-empty">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          <p>{{ $t('dm.no_conversations') }}</p>
          <button class="btn btn-primary btn-sm" @click="showNewConv = true">Start a chat</button>
        </div>
      </div>

    </aside>

    <!-- ── Chat main area ────────────────────────────────── -->
    <main class="dm-main">

      <!-- Empty state -->
      <div v-if="!activeConv" class="dm-empty">
        <div class="dm-empty-icon">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        </div>
        <h3>{{ $t('dm.select_conversation') }}</h3>
        <p>Choose a conversation from the list or start a new one.</p>
      </div>

      <template v-else>

        <!-- Chat header -->
        <div class="dm-chat-header">
          <div class="conv-avatar-wrap">
            <template v-if="activeConv.is_group">
              <div class="group-avatar group-avatar-md">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
              </div>
            </template>
            <template v-else>
              <div class="conv-avatar conv-avatar-md" :style="avatarBg(otherMember(activeConv))">
                <img v-if="getAvatar(otherMember(activeConv))" :src="getAvatar(otherMember(activeConv))" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else class="avatar-initials">{{ initials(otherMember(activeConv)) }}</span>
              </div>
            </template>
          </div>
          <div class="dm-header-info">
            <div class="dm-header-name">{{ convDisplayName(activeConv) }}</div>
            <div class="dm-header-handle">{{ memberList(activeConv) }}</div>
          </div>
          <!-- Add member button for group chats -->
          <button v-if="activeConv.is_group" class="add-member-btn" @click="showAddMember = !showAddMember" title="Add member">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><line x1="20" y1="8" x2="20" y2="14"/><line x1="23" y1="11" x2="17" y2="11"/></svg>
          </button>
        </div>

        <!-- Add member dropdown -->
        <div v-if="showAddMember" class="add-member-panel">
          <div class="search-wrap">
            <input class="search-input" v-model="addMemberSearch" placeholder="Search users…" @input="filterAddMembers" autofocus />
          </div>
          <div class="user-search-results">
            <div v-for="u in filteredAddMembers" :key="u.id" class="user-result" @click="addMember(u)">
              <div class="conv-avatar" :style="avatarBg(u)">
                <img v-if="getAvatar(u)" :src="getAvatar(u)" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else class="avatar-initials">{{ initials(u) }}</span>
              </div>
              <div class="conv-info">
                <div class="conv-name">{{ u.display_name || u.username }}</div>
              </div>
            </div>
            <div v-if="!filteredAddMembers.length" class="search-empty">No users to add</div>
          </div>
        </div>

        <!-- Messages -->
        <div class="dm-messages" ref="messagesEl">

          <template v-for="(msg, i) in messages" :key="msg.id">

            <div v-if="isDifferentDay(messages, i)" class="date-sep">
              <span class="date-sep-label">{{ dayLabel(msg.created_at) }}</span>
            </div>

            <div :class="['msg-row', { 'msg-own': msg.sender_id === auth.user?.id }]">

              <div class="msg-avatar" :style="avatarBg(msg.sender)">
                <img v-if="getAvatar(msg.sender)" :src="getAvatar(msg.sender)" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else class="avatar-initials">{{ initials(msg.sender) }}</span>
              </div>

              <div class="msg-content">
                <div class="msg-sender" v-if="msg.sender_id !== auth.user?.id">
                  {{ msg.sender?.display_name || msg.sender?.username }}
                </div>
                <div :class="['msg-bubble', msg.sender_id === auth.user?.id ? 'bubble-own' : 'bubble-other']">
                  <span v-if="msg.is_deleted" class="msg-deleted">{{ $t('chat.deleted') }}</span>
                  <div v-else class="msg-body" v-html="renderMarkdown(msg.body)"></div>
                  <span v-if="msg.is_edited && !msg.is_deleted" class="msg-edited"> · {{ $t('chat.edited') }}</span>
                </div>
                <div class="msg-time">
                  {{ formatTime(msg.created_at) }}
                  <button
                    v-if="msg.sender_id === auth.user?.id && !msg.is_deleted"
                    class="delete-btn"
                    @click="deleteMsg(msg)"
                    title="Delete"
                  >
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                </div>
              </div>

            </div>

          </template>

          <div v-if="!messages.length" class="messages-empty">
            <p>No messages yet. Say hello! 👋</p>
          </div>

        </div>

        <!-- Compose -->
        <div class="dm-compose">
          <div class="compose-body">
            <div class="compose-avatar" :style="avatarBg(auth.user)">
              <img v-if="getAvatar(auth.user)" :src="getAvatar(auth.user)" class="avatar-img" @error="e => e.target.style.display='none'" />
              <span v-else class="avatar-initials avatar-initials-sm">{{ initials(auth.user) }}</span>
            </div>
            <textarea
              class="compose-textarea"
              v-model="newMessage"
              :placeholder="$t('chat.placeholder')"
              rows="1"
              :disabled="sending"
              ref="textareaEl"
              @keydown.enter.exact.prevent="send"
              @input="autoResize"
            ></textarea>
            <button class="compose-send-btn" @click="send" :disabled="!newMessage.trim() || sending" :title="$t('chat.send')">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
            </button>
          </div>
          <div class="compose-hint">Press Enter to send</div>
        </div>

      </template>
    </main>

  </div>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { messagesApi } from '@/api/messages'
import { useNotificationsStore } from '@/stores/notifications'
import { useDateFormat } from '@/composables/useDateFormat'
import { avatarUrl } from '@/composables/useAvatar'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const route = useRoute()
const auth = useAuthStore()
const notificationsStore = useNotificationsStore()
const { formatTime } = useDateFormat()
const ui = useUIStore()

const conversations = ref([])
const allUsers = ref([])
const filteredUsers = ref([])
const userSearch = ref('')
const showNewConv = ref(false)
const selectedUsers = ref([])
const newGroupName = ref('')

const activeConv = ref(null)
const activeConvId = ref(null)
const messages = ref([])
const newMessage = ref('')
const sending = ref(false)
const messagesEl = ref(null)
const textareaEl = ref(null)
let pollTimer = null

// Add member panel
const showAddMember = ref(false)
const addMemberSearch = ref('')
const filteredAddMembers = ref([])

onMounted(async () => {
  notificationsStore.markSeen()
  try {
    const [convRes, userRes] = await Promise.all([
      messagesApi.getConversations(),
      messagesApi.listUsers()
    ])
    conversations.value = convRes.data || []
    allUsers.value = (userRes.data || []).filter(u => u.id !== auth.user?.id)
    filteredUsers.value = allUsers.value
  } catch {}

  const targetId = route.query.user ? Number(route.query.user) : null
  if (targetId) await openOrCreateDM(targetId)
})

// Sidebar click navigation
watch(() => route.query.user, async (userId) => {
  if (!userId) return
  if (!allUsers.value.length) {
    try {
      const { data } = await messagesApi.listUsers()
      allUsers.value = (data || []).filter(u => u.id !== auth.user?.id)
      filteredUsers.value = allUsers.value
    } catch {}
  }
  await openOrCreateDM(Number(userId))
})

// Open an existing 1-on-1 with this user, or create one
async function openOrCreateDM(userId) {
  if (!userId) return
  const existing = conversations.value.find(c =>
    !c.is_group &&
    c.members?.some(m => m.user_id === userId)
  )
  if (existing) {
    await openConversation(existing)
    return
  }
  // Create a new 1-on-1 conversation
  try {
    const { data } = await messagesApi.createConversation({ user_ids: [userId] })
    if (!conversations.value.find(c => c.id === data.id)) {
      conversations.value.unshift(data)
    }
    await openConversation(data)
  } catch {
    ui.error('Could not open conversation')
  }
}

function toggleNewConv() {
  showNewConv.value = !showNewConv.value
  if (!showNewConv.value) {
    selectedUsers.value = []
    newGroupName.value = ''
    userSearch.value = ''
    filteredUsers.value = allUsers.value
  }
}

function filterUsers() {
  const q = userSearch.value.toLowerCase()
  filteredUsers.value = allUsers.value.filter(u =>
    u.username.toLowerCase().includes(q) ||
    (u.display_name || '').toLowerCase().includes(q)
  )
}

function isSelected(u) {
  return selectedUsers.value.some(s => s.id === u.id)
}

function toggleUser(u) {
  if (isSelected(u)) {
    selectedUsers.value = selectedUsers.value.filter(s => s.id !== u.id)
  } else {
    selectedUsers.value.push(u)
  }
}

async function startConversation() {
  if (!selectedUsers.value.length) return
  const userIds = selectedUsers.value.map(u => u.id)
  try {
    const { data } = await messagesApi.createConversation({
      user_ids: userIds,
      name: newGroupName.value.trim() || ''
    })
    if (!conversations.value.find(c => c.id === data.id)) {
      conversations.value.unshift(data)
    }
    showNewConv.value = false
    selectedUsers.value = []
    newGroupName.value = ''
    userSearch.value = ''
    filteredUsers.value = allUsers.value
    await openConversation(data)
  } catch {
    ui.error('Could not create conversation')
  }
}

async function openConversation(conv) {
  activeConv.value = conv
  activeConvId.value = conv.id
  showAddMember.value = false
  // Stop polling the previous conversation
  clearInterval(pollTimer)
  await fetchMessages(true)
  // Poll for new messages from other participants every 5 s
  pollTimer = setInterval(fetchMessages, 5_000)
}

async function fetchMessages(initial = false) {
  if (!activeConvId.value) return
  try {
    const { data } = await messagesApi.getMessages(activeConvId.value)
    const incoming = data || []
    const atBottom = initial || isAtBottom()
    messages.value = incoming
    if (atBottom) {
      await nextTick()
      scrollToBottom()
    }
  } catch {
    if (initial) ui.error('Failed to load messages')
  }
}

function isAtBottom() {
  if (!messagesEl.value) return true
  const el = messagesEl.value
  return el.scrollHeight - el.scrollTop - el.clientHeight < 60
}

onUnmounted(() => clearInterval(pollTimer))

async function send() {
  const body = newMessage.value.trim()
  if (!body || !activeConv.value) return
  sending.value = true
  try {
    const { data } = await messagesApi.sendConvMessage(activeConv.value.id, { body })
    messages.value.push(data)
    newMessage.value = ''
    if (textareaEl.value) textareaEl.value.style.height = 'auto'
    // Bump this conversation to the top
    const idx = conversations.value.findIndex(c => c.id === activeConv.value.id)
    if (idx > 0) {
      const [c] = conversations.value.splice(idx, 1)
      conversations.value.unshift(c)
    }
    await nextTick()
    scrollToBottom()
    // Refresh to pick up any concurrent messages from others
    await fetchMessages()
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to send message')
  } finally {
    sending.value = false
  }
}

async function deleteMsg(msg) {
  try {
    await messagesApi.deleteConvMessage(activeConv.value.id, msg.id)
    msg.is_deleted = true
  } catch {
    ui.error('Failed to delete message')
  }
}

// Add member to active group conversation
function filterAddMembers() {
  const q = addMemberSearch.value.toLowerCase()
  const memberIds = new Set(activeConv.value?.members?.map(m => m.user_id) || [])
  filteredAddMembers.value = allUsers.value.filter(u =>
    !memberIds.has(u.id) &&
    (u.username.toLowerCase().includes(q) || (u.display_name || '').toLowerCase().includes(q))
  )
}

watch(showAddMember, (v) => {
  if (v) {
    addMemberSearch.value = ''
    filterAddMembers()
  }
})

async function addMember(user) {
  try {
    await messagesApi.addMember(activeConv.value.id, { user_id: user.id })
    // Refresh conversation to get updated member list
    const { data } = await messagesApi.getConversations()
    conversations.value = data || []
    activeConv.value = conversations.value.find(c => c.id === activeConvId.value) || activeConv.value
    showAddMember.value = false
  } catch {
    ui.error('Failed to add member')
  }
}

function scrollToBottom() {
  if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
}


function autoResize(e) {
  const el = e.target
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 120) + 'px'
}

// ── Helpers ──────────────────────────────────────────────

function getAvatar(user) {
  return avatarUrl(user)
}

function initials(u) {
  if (!u) return '?'
  const name = u.display_name || u.username || '?'
  return name.slice(0, 2).toUpperCase()
}

const AVATAR_COLORS = ['#6366f1','#8b5cf6','#ec4899','#f59e0b','#10b981','#3b82f6','#ef4444']
function avatarBg(u) {
  const idx = (u?.username?.charCodeAt(0) || 0) % AVATAR_COLORS.length
  return { background: AVATAR_COLORS[idx] }
}

// Returns the other participant in a 1-on-1 conversation
function otherMember(conv) {
  if (!conv?.members) return null
  const m = conv.members.find(m => m.user_id !== auth.user?.id)
  return m?.user || null
}

// Human-readable conversation name
function convDisplayName(conv) {
  if (!conv) return ''
  if (conv.name) return conv.name
  if (conv.is_group) {
    return conv.members
      ?.filter(m => m.user_id !== auth.user?.id)
      .map(m => m.user?.display_name || m.user?.username)
      .join(', ') || 'Group'
  }
  const other = otherMember(conv)
  return other?.display_name || other?.username || 'Unknown'
}

// Short member list for subtitle
function memberList(conv) {
  if (!conv?.members) return ''
  return conv.members
    .filter(m => m.user_id !== auth.user?.id)
    .map(m => m.user?.display_name || m.user?.username)
    .join(', ')
}

// Date grouping
function isDifferentDay(msgs, index) {
  if (index === 0) return true
  const curr = new Date(msgs[index].created_at)
  const prev = new Date(msgs[index - 1].created_at)
  return curr.getFullYear() !== prev.getFullYear() ||
    curr.getMonth() !== prev.getMonth() ||
    curr.getDate() !== prev.getDate()
}

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}

function dayLabel(dateStr) {
  const d = new Date(dateStr)
  const now = new Date()
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  const sameDay = (a, b) =>
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  if (sameDay(d, now)) return 'Today'
  if (sameDay(d, yesterday)) return 'Yesterday'
  return d.toLocaleDateString(undefined, { weekday: 'long', month: 'short', day: 'numeric' })
}
</script>

<style scoped>
/* ── Layout ──────────────────────────────────────────────── */
.dm-layout { flex: 1; display: flex; overflow: hidden; height: 100%; }

/* ── Sidebar ─────────────────────────────────────────────── */
.dm-sidebar {
  width: 280px;
  flex-shrink: 0;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dm-sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 54px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.dm-sidebar-header h2 { font-size: 15px; font-weight: 700; }

.new-chat-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  transition: all .15s;
}
.new-chat-btn:hover, .new-chat-btn.active {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: #fff;
}

/* New conversation panel */
.new-conv-panel {
  border-bottom: 1px solid var(--color-border);
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.search-wrap {
  position: relative;
  display: flex;
  align-items: center;
}
.search-icon {
  position: absolute;
  left: 9px;
  color: var(--color-text-muted);
  pointer-events: none;
}
.search-input {
  width: 100%;
  padding: 7px 10px 7px 28px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 13px;
  outline: none;
  color: var(--color-text);
}
.search-input:focus { border-color: var(--color-primary); }

/* Selected user chips */
.selected-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 9999px;
  padding: 2px 8px 2px 10px;
  font-size: 12px;
  font-weight: 500;
}
.chip-remove {
  background: none;
  border: none;
  color: rgba(255,255,255,.8);
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  padding: 0;
}
.chip-remove:hover { color: #fff; }

.group-name-input {
  width: 100%;
  padding: 6px 10px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 13px;
  outline: none;
  color: var(--color-text);
  box-sizing: border-box;
}
.group-name-input:focus { border-color: var(--color-primary); }

.user-search-results { max-height: 180px; overflow-y: auto; }
.user-result {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 6px;
  border-radius: 8px;
  cursor: pointer;
  transition: background .1s;
}
.user-result:hover { background: var(--color-bg); }
.user-result.selected { background: color-mix(in srgb, var(--color-primary) 10%, transparent); }
.check-icon { color: var(--color-primary); flex-shrink: 0; margin-left: auto; }
.search-empty { padding: 12px 6px; font-size: 13px; color: var(--color-text-muted); text-align: center; }

.start-conv-btn {
  width: 100%;
  padding: 8px;
  background: var(--color-primary);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity .15s;
}
.start-conv-btn:disabled { opacity: .35; cursor: default; }
.start-conv-btn:not(:disabled):hover { opacity: .88; }

/* Conversation list */
.conv-list { flex: 1; overflow-y: auto; padding: 6px; }

.conv-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  border-radius: 10px;
  cursor: pointer;
  position: relative;
  transition: background .1s;
}
.conv-item:hover { background: var(--color-bg); }
.conv-item.active { background: color-mix(in srgb, var(--color-primary) 12%, transparent); }

.conv-active-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--color-primary);
  margin-left: auto;
  flex-shrink: 0;
}

/* Avatars */
.conv-avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
}
.conv-avatar-md { width: 42px; height: 42px; }
.conv-avatar-wrap { flex-shrink: 0; }

.group-avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--color-text-muted);
}
.group-avatar-md { width: 42px; height: 42px; }

.avatar-img { width: 100%; height: 100%; object-fit: cover; }
.avatar-initials { color: #fff; font-size: 13px; font-weight: 700; }

.conv-info { flex: 1; min-width: 0; }
.conv-name { font-size: 14px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.conv-handle { font-size: 11px; color: var(--color-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.conv-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  gap: 10px;
  color: var(--color-text-muted);
  text-align: center;
}
.conv-empty p { font-size: 13px; margin: 0; }

/* ── Main chat area ───────────────────────────────────────── */
.dm-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.dm-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  gap: 10px;
  text-align: center;
  padding: 40px;
}
.dm-empty-icon {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4px;
}
.dm-empty h3 { font-size: 16px; font-weight: 600; margin: 0; color: var(--color-text); }
.dm-empty p { font-size: 13px; margin: 0; }

/* Chat header */
.dm-chat-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 20px;
  height: 54px;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
}
.dm-header-info { flex: 1; min-width: 0; }
.dm-header-name { font-size: 15px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.dm-header-handle { font-size: 12px; color: var(--color-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.add-member-btn {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  flex-shrink: 0;
  transition: all .15s;
}
.add-member-btn:hover { background: var(--color-primary); border-color: var(--color-primary); color: #fff; }

/* Add member panel */
.add-member-panel {
  border-bottom: 1px solid var(--color-border);
  padding: 10px 20px;
  background: var(--color-surface);
  flex-shrink: 0;
}
.add-member-panel .search-input { padding-left: 10px; }

/* Messages */
.dm-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.messages-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  font-size: 14px;
}

/* Date separator */
.date-sep {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 14px 0 10px;
}
.date-sep::before,
.date-sep::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--color-border);
}
.date-sep-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: .06em;
  white-space: nowrap;
  padding: 0 4px;
}

/* Message rows */
.msg-row {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  margin-bottom: 4px;
}
.msg-row.msg-own { flex-direction: row-reverse; }

.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
}
.msg-avatar .avatar-initials { color: #fff; font-size: 11px; font-weight: 700; }

.msg-content {
  display: flex;
  flex-direction: column;
  max-width: calc(100% - 48px);
}
.msg-row.msg-own .msg-content { align-items: flex-end; }

.msg-sender {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  margin-bottom: 3px;
  padding: 0 4px;
}

.msg-bubble {
  padding: 9px 14px;
  border-radius: 18px;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
  max-width: 480px;
}
.bubble-other {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-bottom-left-radius: 4px;
  color: var(--color-text);
}
.bubble-own {
  background: var(--color-primary);
  color: #fff;
  border-bottom-right-radius: 4px;
}
.msg-deleted { font-style: italic; opacity: .6; }
.msg-edited { font-size: 11px; opacity: .7; }

/* Markdown body inside bubbles */
.msg-body { line-height: 1.5; }
.msg-body :deep(p) { margin: 0 0 4px; }
.msg-body :deep(p:last-child) { margin-bottom: 0; }
.msg-body :deep(strong) { font-weight: 700; }
.msg-body :deep(em) { font-style: italic; }
.msg-body :deep(code) {
  font-family: ui-monospace, monospace;
  font-size: 12px;
  background: rgba(0,0,0,.12);
  border-radius: 4px;
  padding: 1px 5px;
}
.bubble-own .msg-body :deep(code) { background: rgba(255,255,255,.2); }
.msg-body :deep(pre) {
  background: rgba(0,0,0,.15);
  border-radius: 8px;
  padding: 10px 12px;
  overflow-x: auto;
  margin: 6px 0;
}
.bubble-own .msg-body :deep(pre) { background: rgba(255,255,255,.15); }
.msg-body :deep(pre code) { background: none; padding: 0; font-size: 12px; }
.msg-body :deep(ul), .msg-body :deep(ol) { margin: 4px 0; padding-left: 18px; }
.msg-body :deep(li) { margin: 2px 0; }
.msg-body :deep(blockquote) {
  border-left: 3px solid rgba(0,0,0,.2);
  margin: 4px 0;
  padding: 2px 10px;
  opacity: .85;
}
.bubble-own .msg-body :deep(blockquote) { border-left-color: rgba(255,255,255,.4); }
.msg-body :deep(a) { color: inherit; text-decoration: underline; }
.msg-body :deep(h1), .msg-body :deep(h2), .msg-body :deep(h3) { font-size: 1em; font-weight: 700; margin: 4px 0; }
.msg-body :deep(hr) { border: none; border-top: 1px solid rgba(0,0,0,.15); margin: 6px 0; }
.bubble-own .msg-body :deep(hr) { border-top-color: rgba(255,255,255,.3); }

.msg-time {
  font-size: 10px;
  color: var(--color-text-muted);
  margin-top: 3px;
  padding: 0 4px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.delete-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 1px;
  display: flex;
  align-items: center;
  opacity: 0;
  transition: opacity .15s;
}
.msg-time:hover .delete-btn { opacity: 1; }
.delete-btn:hover { color: var(--color-danger); }

/* Compose */
.dm-compose {
  border-top: 1px solid var(--color-border);
  padding: 10px 20px 8px;
  flex-shrink: 0;
  background: var(--color-surface);
}
.compose-body {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 14px;
  padding: 7px 10px 7px 12px;
  transition: border-color .15s;
}
.compose-body:focus-within { border-color: var(--color-primary); }

.compose-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 2px;
}
.avatar-initials-sm { color: #fff; font-size: 9px; font-weight: 700; }

.compose-textarea {
  flex: 1;
  border: none;
  background: transparent;
  resize: none;
  outline: none;
  font-size: 14px;
  line-height: 1.5;
  color: var(--color-text);
  font-family: inherit;
  padding: 2px 0;
  min-height: 24px;
  max-height: 120px;
  overflow-y: auto;
}
.compose-textarea::placeholder { color: var(--color-text-muted); }

.compose-send-btn {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  background: var(--color-primary);
  color: #fff;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: opacity .15s;
}
.compose-send-btn:disabled { opacity: .35; cursor: default; }
.compose-send-btn:not(:disabled):hover { opacity: .85; }

.compose-hint {
  font-size: 10px;
  color: var(--color-text-muted);
  margin-top: 5px;
  text-align: right;
  padding-right: 2px;
}
</style>
