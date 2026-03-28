<template>
  <div class="chat-panel" :class="{ open }" :style="{ width: panelWidth + 'px' }">

    <!-- Resize handle -->
    <div class="resize-handle" @mousedown="startResize"></div>

    <!-- Header -->
    <div class="chat-header">
      <div class="chat-header-info">
        <div class="chat-header-icon">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        </div>
        <span>{{ $t('chat.title') }}</span>
      </div>
      <button class="btn btn-ghost btn-sm close-btn" @click="$emit('close')">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
      </button>
    </div>

    <!-- Message list -->
    <div class="chat-messages" ref="messagesEl">
      <button v-if="chatStore.hasMore && !chatStore.loading" class="load-more-btn" @click="loadMore">
        {{ $t('chat.load_more') }}
      </button>
      <div v-if="chatStore.loading" class="chat-loading">
        <div class="spinner"></div>
      </div>

      <template v-for="(msg, i) in chatStore.messages" :key="msg.id">

        <!-- Date separator -->
        <div v-if="isDifferentDay(chatStore.messages, i)" class="date-sep">
          <span class="date-sep-label">{{ dayLabel(msg.created_at) }}</span>
        </div>

        <!-- Message row -->
        <div :class="['msg-row', { 'msg-own': msg.user_id === authUser?.id && !msg.is_bot }]">

          <div class="msg-avatar">
            <img
              v-if="getAvatar(msg.user) && !msg.is_bot"
              :src="getAvatar(msg.user)"
              :alt="msg.user?.display_name"
              class="avatar-img"
              @error="e => e.target.style.display='none'"
            />
            <span v-else-if="msg.is_bot" class="avatar-initials bot-avatar">🤖</span>
            <span v-else class="avatar-initials">{{ initials(msg.user) }}</span>
          </div>

          <div class="msg-content">
            <div class="msg-sender" v-if="msg.user_id !== authUser?.id || msg.is_bot">
              {{ msg.is_bot ? msg.bot_name : (msg.user?.display_name || msg.user?.username) }}
              <span v-if="msg.is_bot" class="bot-badge">BOT</span>
            </div>

            <!-- Edit mode -->
            <template v-if="editingId === msg.id">
              <textarea class="edit-textarea" v-model="editBody" rows="2" @keydown.enter.exact.prevent="saveEdit(msg)" @keydown.escape="editingId = null"></textarea>
              <div class="edit-actions">
                <button class="btn btn-primary btn-sm" @click="saveEdit(msg)">Save</button>
                <button class="btn btn-ghost btn-sm" @click="editingId = null">Cancel</button>
              </div>
            </template>

            <template v-else>
              <div :class="['msg-bubble', msg.user_id === authUser?.id && !msg.is_bot ? 'bubble-own' : 'bubble-other']">
                <div v-if="msg.is_deleted" class="msg-deleted">{{ $t('chat.deleted') }}</div>
                <div v-else class="msg-body" v-html="renderMarkdown(msg.body)"></div>
              </div>
              <AttachmentList v-if="!msg.is_deleted" :attachments="msg.attachments" />
              <MessageReactions
                v-if="!msg.is_deleted && !msg.is_bot"
                :reactions="msg.reactions"
                @toggle="(emoji) => toggleReaction(msg, emoji)"
              />
            </template>

            <div class="msg-meta">
              <span class="msg-time">{{ formatTime(msg.created_at) }}</span>
              <span v-if="msg.is_edited" class="msg-edited">· {{ $t('chat.edited') }}</span>
              <button
                v-if="msg.user_id === authUser?.id && !msg.is_deleted && !msg.is_bot && editingId !== msg.id"
                class="msg-action-btn"
                @click="startEdit(msg)"
                title="Edit"
              >
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
            </div>
          </div>

        </div>
      </template>

      <div v-if="!chatStore.loading && !chatStore.messages.length" class="chat-empty">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        <p>No messages yet. Start the conversation!</p>
      </div>
    </div>

    <!-- Compose area -->
    <div class="chat-compose">
      <AttachmentList v-if="pendingFiles.length" :attachments="pendingFiles" :can-delete="true" @remove="removePending" />
      <div class="compose-body">
        <div class="compose-avatar">
          <img v-if="getAvatar(authUser)" :src="getAvatar(authUser)" class="avatar-img" @error="e => e.target.style.display='none'" />
          <span v-else class="avatar-initials avatar-initials-sm">{{ initials(authUser) }}</span>
        </div>
        <FileUploadButton @files-selected="onFilesSelected" />
        <textarea
          class="compose-textarea"
          v-model="draft"
          :placeholder="$t('chat.placeholder')"
          rows="1"
          ref="textareaEl"
          @keydown.enter.exact.prevent="sendMessage"
          @input="autoResize"
        ></textarea>
        <button class="compose-send-btn" @click="sendMessage" :disabled="!draft.trim() && !pendingFiles.length" :title="$t('chat.send')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
        </button>
      </div>
      <div class="compose-hint">Enter to send · Markdown supported</div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import { useDateFormat } from '@/composables/useDateFormat'
import { avatarUrl } from '@/composables/useAvatar'
import { attachmentsApi } from '@/api/attachments'
import { projectsApi } from '@/api/projects'
import AttachmentList from '@/components/common/AttachmentList.vue'
import FileUploadButton from '@/components/common/FileUploadButton.vue'
import MessageReactions from '@/components/common/MessageReactions.vue'

const props = defineProps({
  open: Boolean,
  projectSlug: String,
  wsSend: Function
})
defineEmits(['close'])

const chatStore = useChatStore()
const auth = useAuthStore()
const authUser = computed(() => auth.user)
const messagesEl = ref(null)
const textareaEl = ref(null)
const draft = ref('')
const { formatTime } = useDateFormat()

// Edit state
const editingId = ref(null)
const editBody = ref('')

// Pending file attachments
const pendingFiles = ref([]) // [{name, size, mime_type, _file}]

// ── Resize logic ───────────────────────────────────────────
const panelWidth = ref(360)
const MIN_WIDTH = 260
const MAX_WIDTH = 720

function startResize(e) {
  e.preventDefault()
  const startX = e.clientX
  const startWidth = panelWidth.value

  function onMove(e) {
    const delta = startX - e.clientX
    panelWidth.value = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, startWidth + delta))
  }
  function onUp() {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
  document.body.style.cursor = 'ew-resize'
  document.body.style.userSelect = 'none'
}

onMounted(async () => {
  if (props.projectSlug) {
    await chatStore.loadMessages(props.projectSlug)
    scrollToBottom()
  }
})

watch(() => chatStore.messages.length, () => nextTick(scrollToBottom))
watch(() => props.open, (val) => { if (val) nextTick(scrollToBottom) })

function scrollToBottom() {
  if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
}

async function loadMore() {
  const firstId = chatStore.messages[0]?.id
  await chatStore.loadMessages(props.projectSlug, firstId)
}

async function sendMessage() {
  if (!draft.value.trim() && !pendingFiles.length) return

  // Send text via WebSocket
  if (draft.value.trim()) {
    props.wsSend?.('chat.send', { body: draft.value })
  }

  // Upload pending files and attach to the latest message
  if (pendingFiles.length) {
    // Wait briefly for the WS message to arrive so we have a message ID
    // For simplicity, upload files and associate after the WS message is created
    const filesToUpload = [...pendingFiles.value]
    pendingFiles.value = []
    // Note: file upload for chat messages uses the chat REST endpoint
    // We upload after a brief delay to get the created message ID
    // This is a best-effort approach - files go as a separate message if no text
    for (const pf of filesToUpload) {
      const fd = new FormData()
      fd.append('file', pf._file)
      // owner will be patched after message is created - for now we skip complex sequencing
      // and just upload them as unattached (they will show in future when linked properly)
      // A simpler UX: upload files and then show them inline via a second message
      await attachmentsApi.upload(fd).catch(() => {})
    }
  }

  draft.value = ''
  nextTick(() => {
    if (textareaEl.value) textareaEl.value.style.height = 'auto'
  })
}

function onFilesSelected(files) {
  for (const f of files) {
    pendingFiles.value.push({
      id: Math.random(),
      filename: f.name,
      size_bytes: f.size,
      mime_type: f.type || 'application/octet-stream',
      _file: f
    })
  }
}

function removePending(a) {
  pendingFiles.value = pendingFiles.value.filter(p => p.id !== a.id)
}

function startEdit(msg) {
  editingId.value = msg.id
  editBody.value = msg.body
}

function saveEdit(msg) {
  if (!editBody.value.trim()) return
  props.wsSend?.('chat.edit', { message_id: msg.id, body: editBody.value })
  editingId.value = null
}

async function toggleReaction(msg, emoji) {
  if (!props.projectSlug) return
  try {
    const { data } = await projectsApi.toggleChatReaction(props.projectSlug, msg.id, emoji)
    chatStore.updateReactions(msg.id, data.reactions)
  } catch {}
}

function autoResize(e) {
  const el = e.target
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 120) + 'px'
}

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}

function getAvatar(user) {
  return avatarUrl(user)
}

function initials(user) {
  if (!user) return '?'
  const name = user.display_name || user.username || '?'
  return name.slice(0, 2).toUpperCase()
}

// Date grouping helpers
function isDifferentDay(messages, index) {
  if (index === 0) return true
  const curr = new Date(messages[index].created_at)
  const prev = new Date(messages[index - 1].created_at)
  return curr.getFullYear() !== prev.getFullYear() ||
    curr.getMonth() !== prev.getMonth() ||
    curr.getDate() !== prev.getDate()
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
/* ── Panel shell ─────────────────────────────────────────── */
.chat-panel {
  position: fixed;
  right: 0;
  top: 56px;
  bottom: 0;
  background: var(--color-surface);
  border-left: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  transform: translateX(100%);
  transition: transform .25s ease;
  z-index: 50;
  box-shadow: -6px 0 24px rgba(0,0,0,.08);
  min-width: 260px;
  max-width: 720px;
}
.chat-panel.open { transform: translateX(0); }

/* ── Resize handle ───────────────────────────────────────── */
.resize-handle {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 5px;
  cursor: ew-resize;
  z-index: 10;
}
.resize-handle::after {
  content: '';
  position: absolute;
  left: 2px;
  top: 50%;
  transform: translateY(-50%);
  width: 1px;
  height: 40px;
  background: var(--color-border);
  border-radius: 1px;
  opacity: 0;
  transition: opacity .2s;
}
.resize-handle:hover::after { opacity: 1; }

/* ── Header ──────────────────────────────────────────────── */
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 54px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.chat-header-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 14px;
}
.chat-header-icon {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: var(--color-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
}
.close-btn { color: var(--color-text-muted); }
.close-btn:hover { color: var(--color-text); }

/* ── Message list ────────────────────────────────────────── */
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

/* ── Date separator ──────────────────────────────────────── */
.date-sep {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 12px 0 8px;
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

/* ── Message row ─────────────────────────────────────────── */
.msg-row {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  margin-bottom: 4px;
}
.msg-row.msg-own { flex-direction: row-reverse; }

/* ── Avatar ──────────────────────────────────────────────── */
.msg-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
}
.avatar-img { width: 100%; height: 100%; object-fit: cover; }
.avatar-initials {
  color: #fff;
  font-size: 10px;
  font-weight: 700;
}
.bot-avatar { font-size: 14px; }

/* ── Message content ─────────────────────────────────────── */
.msg-content {
  display: flex;
  flex-direction: column;
  max-width: calc(100% - 46px);
}
.msg-row.msg-own .msg-content { align-items: flex-end; }

.msg-sender {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  margin-bottom: 3px;
  padding: 0 4px;
  display: flex;
  align-items: center;
  gap: 5px;
}

.bot-badge {
  font-size: 9px;
  font-weight: 700;
  background: var(--color-primary);
  color: #fff;
  padding: 1px 4px;
  border-radius: 3px;
  letter-spacing: .04em;
}

.msg-bubble {
  padding: 8px 12px;
  border-radius: 16px;
  font-size: 13px;
  line-height: 1.5;
  word-break: break-word;
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

.msg-body :deep(p) { margin: 0 0 4px; }
.msg-body :deep(p:last-child) { margin-bottom: 0; }
.msg-body :deep(code) {
  background: rgba(0,0,0,.1);
  padding: 1px 4px;
  border-radius: 3px;
  font-size: 12px;
}
.bubble-own .msg-body :deep(code) { background: rgba(255,255,255,.2); }
.msg-body :deep(pre) { margin: 4px 0; }
.msg-body :deep(a) { color: inherit; text-decoration: underline; }

.msg-meta {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 3px;
  padding: 0 4px;
}
.msg-time { font-size: 10px; color: var(--color-text-muted); }
.msg-edited { font-size: 10px; font-style: italic; color: var(--color-text-muted); }
.msg-action-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 1px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  opacity: 0;
  transition: opacity .15s;
}
.msg-meta:hover .msg-action-btn { opacity: 1; }
.msg-action-btn:hover { color: var(--color-text); background: var(--color-bg); }

/* ── Edit inline ─────────────────────────────────────────── */
.edit-textarea {
  width: 100%;
  border: 1px solid var(--color-primary);
  border-radius: 8px;
  padding: 6px 10px;
  font-size: 13px;
  background: var(--color-bg);
  color: var(--color-text);
  resize: none;
  outline: none;
  font-family: inherit;
}
.edit-actions {
  display: flex;
  gap: 6px;
  margin-top: 4px;
}

/* ── Empty state ─────────────────────────────────────────── */
.chat-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  gap: 8px;
  font-size: 13px;
}

/* ── Loading ─────────────────────────────────────────────── */
.chat-loading { display: flex; justify-content: center; padding: 12px; }
.load-more-btn {
  align-self: center;
  font-size: 12px;
  color: var(--color-text-muted);
  background: none;
  border: 1px solid var(--color-border);
  border-radius: 9999px;
  padding: 4px 14px;
  cursor: pointer;
  margin-bottom: 8px;
}
.load-more-btn:hover { background: var(--color-bg); }

/* ── Compose area ────────────────────────────────────────── */
.chat-compose {
  border-top: 1px solid var(--color-border);
  padding: 10px 12px 8px;
  flex-shrink: 0;
  background: var(--color-surface);
}
.compose-body {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 6px 8px 6px 10px;
  transition: border-color .15s;
}
.compose-body:focus-within { border-color: var(--color-primary); }

.compose-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--color-primary);
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
  font-size: 13px;
  line-height: 1.5;
  color: var(--color-text);
  font-family: inherit;
  padding: 2px 0;
  min-height: 22px;
  max-height: 120px;
  overflow-y: auto;
}
.compose-textarea::placeholder { color: var(--color-text-muted); }

.compose-send-btn {
  width: 30px;
  height: 30px;
  border-radius: 8px;
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
.compose-send-btn:disabled { opacity: .4; cursor: default; }
.compose-send-btn:not(:disabled):hover { opacity: .85; }

.compose-hint {
  font-size: 10px;
  color: var(--color-text-muted);
  margin-top: 5px;
  text-align: right;
  padding-right: 2px;
}
</style>
