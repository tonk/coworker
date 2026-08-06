<template>
  <aside class="call-chat-sidebar" :class="{ 'call-chat-sidebar--fixed': fixed }" :style="{ width: callChatWidth + 'px' }" @click.stop @keydown.esc.stop="onEsc">
    <div class="call-chat-resize-handle" @mousedown="startResize"></div>
    <div class="call-chat-head">
      <span class="call-chat-title">{{ $t('call.chat_while_in_call') }}</span>
      <button type="button" class="call-chat-close" aria-label="Close sidebar" :title="$t('call.hide_chat')" @click="$emit('close')">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>
    </div>
    <div class="call-chat-msgs" ref="msgsEl">
      <div v-if="loading" class="call-chat-loading">{{ $t('common.loading') }}</div>
      <template v-else>
        <div v-for="msg in messages" :key="msg.id" :class="['cc-row', { 'cc-own': msg.sender_id === auth.user?.id }]">
          <div class="cc-bubble">
            <span v-if="msg.sender_id !== auth.user?.id" class="cc-name">{{ msg.sender?.display_name || msg.sender?.username }}</span>
            <div v-if="msg.is_deleted" class="cc-deleted">{{ $t('chat.deleted') }}</div>
            <div v-else class="cc-body" v-html="renderMarkdown(msg.body)"></div>
          </div>
        </div>
      </template>
    </div>
    <div v-if="showDiscardWarning" class="call-chat-discard-warn">
      <span>{{ $t('chat.discard_draft') }}</span>
      <div class="call-chat-discard-actions">
        <button type="button" class="btn btn-danger btn-sm" @click="$emit('close')">{{ $t('common.discard') }}</button>
        <button type="button" class="btn btn-secondary btn-sm" @click="showDiscardWarning = false">{{ $t('common.cancel') }}</button>
      </div>
    </div>
    <div class="call-chat-compose">
      <textarea
        v-model="draft"
        rows="2"
        class="call-chat-input"
        aria-label="Message"
        :placeholder="$t('chat.placeholder')"
        :disabled="sending"
        spellcheck="true"
        :lang="auth.user?.locale || 'en'"
        @keydown.enter.exact.prevent="send"
      />
      <button type="button" class="btn btn-primary btn-sm call-chat-send" :disabled="!draft.trim() || sending" @click="send">
        {{ $t('chat.send') }}
      </button>
    </div>
  </aside>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { messagesApi } from '@/api/messages'
import { renderMarkdown } from '@/composables/useCardRef'
import { useCallChatPanel } from '@/composables/useCallChatPanel'

const props = defineProps({
  conversationId: { type: Number, required: true },
  /** When true, pin to the right viewport edge (over page content); when false, share a flex row with video */
  fixed: { type: Boolean, default: false },
})

const emit = defineEmits(['close'])

const { callChatWidth, setCallChatWidth } = useCallChatPanel()
const auth = useAuthStore()
const messages = ref([])
const loading = ref(true)
const draft = ref('')
const sending = ref(false)
const msgsEl = ref(null)
const showDiscardWarning = ref(false)

function onEsc() {
  if (draft.value.trim()) {
    showDiscardWarning.value = true
  } else {
    emit('close')
  }
}

let pollTimer = null

async function load() {
  if (!props.conversationId) return
  try {
    const { data } = await messagesApi.getMessages(props.conversationId)
    messages.value = Array.isArray(data) ? data : []
  } catch {
    messages.value = []
  } finally {
    loading.value = false
    nextTick(scrollBottom)
  }
}

function scrollBottom() {
  const el = msgsEl.value
  if (el) el.scrollTop = el.scrollHeight
}

async function send() {
  const body = draft.value.trim()
  if (!body || sending.value) return
  sending.value = true
  try {
    const { data } = await messagesApi.sendConvMessage(props.conversationId, { body })
    if (data) messages.value.push(data)
    draft.value = ''
    nextTick(scrollBottom)
  } catch {
    // ignore
  } finally {
    sending.value = false
  }
}

watch(
  () => props.conversationId,
  () => {
    loading.value = true
    messages.value = []
    void load()
  }
)

// ── Resize (drag the left edge — this panel is always docked to the right) ──
let resizing = false
let startX = 0
let startWidth = 0

function startResize(e) {
  resizing = true
  startX = e.clientX
  startWidth = callChatWidth.value
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

function onResize(e) {
  if (!resizing) return
  setCallChatWidth(startWidth - (e.clientX - startX))
}

function stopResize() {
  if (!resizing) return
  resizing = false
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}

onMounted(() => {
  void load()
  pollTimer = setInterval(load, 5_000)
})

onUnmounted(() => {
  stopResize()
  clearInterval(pollTimer)
})
</script>

<style scoped>
.call-chat-sidebar {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  position: relative;
  background: var(--color-surface);
  border-left: 1px solid var(--color-border);
  color: var(--color-text);
  box-sizing: border-box;
  z-index: 520;
}

.call-chat-resize-handle {
  position: absolute;
  top: 0;
  left: -3px;
  width: 6px;
  height: 100%;
  cursor: col-resize;
  z-index: 10;
}
.call-chat-resize-handle:hover,
.call-chat-resize-handle:active {
  background: var(--color-primary);
  opacity: 0.4;
}
.call-chat-sidebar--fixed {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 510;
  border-left: 1px solid var(--color-border);
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.15);
}

.call-chat-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.call-chat-title {
  font-size: 13px;
  font-weight: 600;
}
.call-chat-close {
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.call-chat-close:hover {
  background: var(--color-bg);
  color: var(--color-text);
}

.call-chat-msgs {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 10px 12px;
  font-size: 13px;
}
.call-chat-loading {
  color: var(--color-text-muted);
  font-size: 12px;
}
.cc-row {
  margin-bottom: 8px;
  display: flex;
}
.cc-row.cc-own {
  justify-content: flex-end;
}
.cc-bubble {
  max-width: 92%;
  padding: 6px 10px;
  border-radius: 10px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
}
.cc-own .cc-bubble {
  background: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 22%, var(--color-bg));
  border-color: color-mix(in srgb, var(--color-primary) 40%, var(--color-border));
}
.cc-name {
  display: block;
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  margin-bottom: 2px;
}
.cc-deleted {
  font-style: italic;
  color: var(--color-text-muted);
  font-size: 12px;
}
.cc-body :deep(p) {
  margin: 0;
}
.cc-body :deep(p + p) {
  margin-top: 4px;
}

.call-chat-discard-warn {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 12px;
  background: color-mix(in srgb, var(--color-danger, #ef4444) 10%, var(--color-surface));
  border-top: 1px solid color-mix(in srgb, var(--color-danger, #ef4444) 30%, var(--color-border));
  font-size: 13px;
  color: var(--color-text);
}
.call-chat-discard-actions {
  display: flex;
  gap: 8px;
}

.call-chat-compose {
  flex-shrink: 0;
  padding: 10px 12px;
  border-top: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.call-chat-input {
  width: 100%;
  box-sizing: border-box;
  resize: none;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  color: var(--color-text);
  font-family: inherit;
  font-size: 13px;
}
.call-chat-send {
  align-self: flex-end;
}
</style>
