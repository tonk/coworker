<template>
  <div class="editor-wrapper" @keydown.escape.stop>
    <textarea ref="editorEl"></textarea>
    <!-- Mention dropdown — positioned over the editor -->
    <MentionDropdown
      v-if="mentionUsers.length"
      :users="mentionUsers"
      :active-index="mentionIndex"
      :style="mentionPos"
      class="editor-mention-dropdown"
      @pick="editorPickMention"
      @update:activeIndex="mentionIndex = $event"
    />
    <!-- Emoji picker — anchored to the emoji toolbar button -->
    <InlineEmojiPicker
      v-if="emojiOpen"
      :initial-search="emojiQuery || ''"
      :style="emojiStart ? emojiPos : {}"
      :class="['editor-emoji-picker', { 'is-autocomplete': emojiStart }]"
      @pick="onEmojiPick"
      @escape="onEmojiEscape"
      @close="emojiOpen = false"
    />
  </div>
</template>

<script setup>
import { ref, watch, computed, onMounted, onBeforeUnmount } from 'vue'
import EasyMDE from 'easymde'
import 'easymde/dist/easymde.min.css'
import MentionDropdown from '@/components/common/MentionDropdown.vue'
import InlineEmojiPicker from '@/components/common/InlineEmojiPicker.vue'
import { detectEmoticon } from '@/utils/emoticons'
import { useAuthStore } from '@/stores/auth'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: 'Write using Markdown...' },
  minHeight: { type: String, default: '150px' },
  users: { type: Array, default: () => [] },
})
const emit = defineEmits(['update:modelValue'])

const auth = useAuthStore()
const editorEl = ref(null)
let mde = null

// ── Emoji ─────────────────────────────────────────────────────────────────────
const emojiOpen = ref(false)
const emojiQuery = ref(null)
const emojiStart = ref(null)
const emojiPos   = ref({ top: '0px', left: '0px' })

function onEmojiPick(emoji) {
  if (!mde) return
  if (emojiStart.value) {
    mde.codemirror.replaceRange(emoji, emojiStart.value, mde.codemirror.getCursor())
  } else {
    mde.codemirror.replaceSelection(emoji)
  }
  mde.codemirror.focus()
  emojiOpen.value = false
  emojiQuery.value = null
}

function onEmojiEscape() {
  emojiOpen.value = false
  emojiQuery.value = null
  mde?.codemirror?.focus()
}

// ── Mention state ─────────────────────────────────────────────────────────────
const mentionQuery = ref(null)
const mentionStart = ref(null)   // CodeMirror { line, ch } of the leading @
const mentionIndex = ref(0)
const mentionPos   = ref({ top: '0px', left: '0px' })

const mentionUsers = computed(() => {
  if (mentionQuery.value === null) return []
  const q = (mentionQuery.value || '').toLowerCase()
  const normalized = (props.users || [])
    .map(u => u?.user || u?.User || u)
    .filter(Boolean)
    .map(u => ({
      ...u,
      id: u.id ?? u.user_id ?? u.UserID,
      username: u.username ?? u.Username ?? '',
      display_name: u.display_name ?? u.displayName ?? u.DisplayName ?? '',
    }))
    .filter(u => !!u.username)

  return normalized.filter(u => {
    const username = String(u.username).toLowerCase()
    const displayName = String(u.display_name || '').toLowerCase()
    return username.startsWith(q) || displayName.startsWith(q)
  }).slice(0, 8)
})

function detectTriggers() {
  if (!mde) return
  const cm     = mde.codemirror
  const cursor = cm.getCursor()
  const line   = cm.getLine(cursor.line) || ''
  const before = line.slice(0, cursor.ch)
  const cursorPx = cm.cursorCoords(cursor, 'page')
  const anchorTop = cursorPx.bottom - window.scrollY + 4
  const anchorLeft = cursorPx.left - window.scrollX

  const mentionMatch = before.match(/(^|[^\w])@([\w\s]*)$/)
  const emojiMatch   = before.match(/(^|[^\w]):(\w*)$/)

  if (mentionMatch) {
    mentionQuery.value = (mentionMatch[2] || '').replace(/\s+/g, '')
    const offset = mentionMatch[1].length
    mentionStart.value = { line: cursor.line, ch: cursor.ch - mentionMatch[0].length + offset }
    mentionIndex.value = 0
    mentionPos.value = { top: `${anchorTop}px`, left: `${anchorLeft}px` }
    emojiQuery.value = null
    emojiStart.value = null
  } else if (emojiMatch) {
    emojiQuery.value = emojiMatch[2]
    const offset = emojiMatch[1].length
    emojiStart.value = { line: cursor.line, ch: cursor.ch - emojiMatch[0].length + offset }
    emojiPos.value = { top: `${anchorTop}px`, left: `${anchorLeft}px` }
    emojiOpen.value = true
    mentionQuery.value = null
    mentionStart.value = null
  } else {
    mentionQuery.value = null
    emojiQuery.value = null
    if (!emojiMatch && emojiStart.value) emojiOpen.value = false
  }
}

function editorPickMention(user) {
  if (!mde || !mentionStart.value) return
  const cm   = mde.codemirror
  const text = '@' + user.username + ' '
  cm.replaceRange(text, mentionStart.value, cm.getCursor())
  mentionQuery.value = null
  cm.focus()
}

function handleCmKeydown(cm, e) {
  if (mentionQuery.value !== null && mentionUsers.value.length) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      mentionIndex.value = (mentionIndex.value + 1) % mentionUsers.value.length
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      mentionIndex.value = (mentionIndex.value - 1 + mentionUsers.value.length) % mentionUsers.value.length
    } else if (e.key === 'Enter' || e.key === 'Tab') {
      e.preventDefault()
      editorPickMention(mentionUsers.value[mentionIndex.value])
    } else if (e.key === 'Escape') {
      mentionQuery.value = null
    }
  } else if (emojiQuery.value !== null && emojiOpen.value) {
    if (e.key === 'Escape') {
      emojiOpen.value = false
      emojiQuery.value = null
    }
  }
}

// ── EasyMDE setup ─────────────────────────────────────────────────────────────
onMounted(() => {
  mde = new EasyMDE({
    element: editorEl.value,
    initialValue: props.modelValue,
    placeholder: props.placeholder,
    spellChecker: false,
    nativeSpellcheck: true,
    autofocus: false,
    minHeight: props.minHeight,
    toolbar: [
      'bold', 'italic', 'strikethrough', '|',
      'heading', 'quote', 'code', '|',
      'unordered-list', 'ordered-list', '|',
      'link', 'image', '|',
      {
        name: 'emoji',
        action: () => { emojiOpen.value = !emojiOpen.value; emojiStart.value = null },
        className: 'emoji-toolbar-btn',
        title: 'Emoji',
        text: '😊',
      },
      '|',
      'preview', 'side-by-side', 'fullscreen', '|',
      'guide'
    ]
  })

  // Enable browser spellcheck using the user's locale
  const inputEl = mde.codemirror.getInputField()
  inputEl.setAttribute('spellcheck', 'true')
  inputEl.setAttribute('lang', auth.user?.locale || 'en')

  mde.codemirror.on('change', (cm, changeObj) => {
    // Only replace on regular character input (not undo/redo/paste)
    if (changeObj.origin === '+input') {
      const cursor = cm.getCursor()
      const line   = cm.getLine(cursor.line) || ''
      const before = line.slice(0, cursor.ch)
      const hit    = detectEmoticon(before)
      if (hit) {
        const { pattern, emoji } = hit
        const from = { line: cursor.line, ch: cursor.ch - pattern.length }
        cm.replaceRange(emoji, from, cursor)
        return // skip emit/detectMention — the replacement triggers another change event
      }
    }
    emit('update:modelValue', mde.value())
    detectTriggers()
  })

  mde.codemirror.on('cursorActivity', detectTriggers)
  mde.codemirror.on('keydown', handleCmKeydown)
})

watch(() => props.modelValue, (val) => {
  if (mde && mde.value() !== val) mde.value(val)
})

onBeforeUnmount(() => {
  mde?.codemirror.off('change')
  mde?.codemirror.off('cursorActivity', detectTriggers)
  mde?.codemirror.off('keydown', handleCmKeydown)
  mde?.cleanup()      // removes EasyMDE's document-level keydown listener
  mde?.toTextArea()
  mde = null
})
</script>

<style scoped>
.editor-wrapper {
  width: 100%;
  position: relative;
}

/* Override EasyMDE toolbar emoji button — suppress the icon font, show text emoji */
:deep(.emoji-toolbar-btn) {
  font-style: normal !important;
  font-size: 16px !important;
  line-height: 1 !important;
}
:deep(.emoji-toolbar-btn::before) {
  content: none !important;
}

.editor-mention-dropdown {
  position: fixed !important;
  z-index: 1100;
}

.editor-emoji-picker {
  position: absolute !important;
  top: 36px;   /* just below the toolbar */
  right: 0;
  bottom: auto !important;
  left: auto !important;
  z-index: 500;
}

/* When triggered by ':' autocomplete, follow the cursor coordinates */
.editor-emoji-picker.is-autocomplete {
  position: fixed !important;
  top: auto !important;
  right: auto !important;
  bottom: auto !important;
  left: auto !important;
  z-index: 1100;
  /* Reset any absolute offsets if necessary, though style binding will override top/left */
  margin: 0;
}
</style>
