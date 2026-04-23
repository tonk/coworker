/**
 * useCompose — emoji insertion and @mention autocomplete for plain <textarea> elements.
 *
 * Usage:
 *   const { mentionUsers, mentionIndex, insertText, onTextareaInput, onTextareaKeydown, pickMention }
 *     = useCompose({ textareaEl, getValue, setValue, users })
 *
 *   - textareaEl : ref to the <textarea> DOM element
 *   - getValue   : () => string  — read the current model value
 *   - setValue   : (s) => void   — write a new model value
 *   - users      : ref/computed of user objects { id, username, display_name }
 */
import { ref, computed, nextTick } from 'vue'
import { detectEmoticon, detectEmojiShortcode } from '@/utils/emoticons'

export function useCompose({ textareaEl, getValue, setValue, users }) {
  const mentionQuery = ref(null)   // null = no active mention; string = partial after @
  const mentionStart = ref(0)      // character offset of the leading @
  const mentionIndex = ref(0)      // keyboard-highlighted row

  // Emoji-by-name query (e.g. :smile)
  const emojiQuery = ref(null)     // null = no active emoji query; string = partial after :
  const emojiStart = ref(0)

  const mentionUsers = computed(() => {
    if (mentionQuery.value === null) return []
    const q = (mentionQuery.value || '').toLowerCase()
    return (users.value || []).filter(u =>
      u.username.toLowerCase().startsWith(q) ||
      (u.display_name || '').toLowerCase().startsWith(q)
    ).slice(0, 8)
  })

  // Insert arbitrary text at the textarea cursor (emoji, completed mention, etc.)
  function insertText(text) {
    const el = textareaEl.value
    if (!el) return
    const start = el.selectionStart
    const end   = el.selectionEnd
    const val   = getValue()
    setValue(val.slice(0, start) + text + val.slice(end))
    nextTick(() => {
      el.selectionStart = el.selectionEnd = start + [...text].length
      el.focus()
    })
  }

  // Call this from the textarea's @input handler
  function onTextareaInput() {
    const el = textareaEl.value
    if (!el) return
    const pos    = el.selectionStart
    const val    = getValue()
    const before = val.slice(0, pos)

    // Emoticon replacement
    const hit = detectEmoticon(before)
    if (hit) {
      const { pattern, emoji } = hit
      const newVal = val.slice(0, pos - pattern.length) + emoji + val.slice(pos)
      setValue(newVal)
      nextTick(() => {
        el.selectionStart = el.selectionEnd = pos - pattern.length + [...emoji].length
        el.focus()
      })
      mentionQuery.value = null
      emojiQuery.value = null
      return
    }

    // :shortcode: replacement (e.g. :fingerscrossed:)
    const short = detectEmojiShortcode(before)
    if (short) {
      const { pattern, emoji } = short
      const newVal = val.slice(0, pos - pattern.length) + emoji + val.slice(pos)
      setValue(newVal)
      nextTick(() => {
        el.selectionStart = el.selectionEnd = pos - pattern.length + [...emoji].length
        el.focus()
      })
      mentionQuery.value = null
      emojiQuery.value = null
      return
    }

    // Mention detection
    const m = before.match(/@(\w*)$/)
    if (m) {
      mentionQuery.value = m[1]
      mentionStart.value = pos - m[0].length
      mentionIndex.value = 0
      emojiQuery.value = null
      return
    } else {
      mentionQuery.value = null
    }

    // Emoji-name detection (start with ':' and then letters/numbers/_+-)
    const em = before.match(/:([a-z0-9_+\-]*)$/i)
    if (em) {
      emojiQuery.value = em[1]
      emojiStart.value = pos - em[0].length
    } else {
      emojiQuery.value = null
    }
  }

  // Call this from the textarea's @keydown handler (before default handling)
  // Returns true if the key was consumed (caller should suppress default + stop).
  function onTextareaKeydown(e) {
    // Mention navigation
    if (mentionQuery.value !== null && mentionUsers.value.length) {
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        mentionIndex.value = (mentionIndex.value + 1) % mentionUsers.value.length
        return true
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault()
        mentionIndex.value = (mentionIndex.value - 1 + mentionUsers.value.length) % mentionUsers.value.length
        return true
      }
      if (e.key === 'Enter' || e.key === 'Tab') {
        e.preventDefault()
        pickMention(mentionUsers.value[mentionIndex.value])
        return true
      }
      if (e.key === 'Escape') {
        mentionQuery.value = null
        return true
      }
      return false
    }

    // Emoji query keyboard handling: Escape closes, Enter/Tab do nothing here
    if (emojiQuery.value !== null) {
      if (e.key === 'Escape') {
        emojiQuery.value = null
        return true
      }
      // Let Enter/Tab insert normally — selection from picker is via mouse for now
      return false
    }

    return false
  }

  function pickMention(user) {
    const el  = textareaEl.value
    const pos = el?.selectionStart ?? (getValue().length)
    const val = getValue()
    const mention = '@' + user.username + ' '
    setValue(val.slice(0, mentionStart.value) + mention + val.slice(pos))
    mentionQuery.value = null
    nextTick(() => {
      if (el) {
        el.selectionStart = el.selectionEnd = mentionStart.value + mention.length
        el.focus()
      }
    })
  }

  function pickEmoji(emoji) {
    const el = textareaEl.value
    const val = getValue()
    const pos = el?.selectionStart ?? val.length

    if (emojiQuery.value !== null) {
      const insert = emoji + ' '
      setValue(val.slice(0, emojiStart.value) + insert + val.slice(pos))
      emojiQuery.value = null
      nextTick(() => {
        if (el) {
          const nextPos = emojiStart.value + [...insert].length
          el.selectionStart = el.selectionEnd = nextPos
          el.focus()
        }
      })
      return
    }

    insertText(emoji)
  }

  return { mentionUsers, mentionIndex, insertText, onTextareaInput, onTextareaKeydown, pickMention, emojiQuery, pickEmoji }
}
