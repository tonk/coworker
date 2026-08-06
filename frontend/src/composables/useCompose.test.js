import { describe, it, expect } from 'vitest'
import { ref as vueRef } from 'vue'
import { useCompose } from './useCompose'

function makeHarness(initial = '') {
  const value = { text: initial }
  const el = { selectionStart: initial.length, selectionEnd: initial.length, focus: () => {} }
  const textareaEl = vueRef(el)
  // mentionQuery isn't exposed directly by useCompose(); a non-empty user
  // list lets mentionUsers.value act as a proxy — the computed returns []
  // outright while mentionQuery is null, and only filters by prefix once a
  // mention is actually active (empty-string query still matches everyone).
  const users = vueRef([{ id: 1, username: 'alice', display_name: 'Alice' }])
  const compose = useCompose({
    textareaEl,
    getValue: () => value.text,
    setValue: (v) => { value.text = v },
    users,
  })
  return { compose, el, value }
}

describe('useCompose emoji trigger', () => {
  it('does not open on a bare colon', () => {
    const { compose, el, value } = makeHarness()
    value.text = 'Note:'
    el.selectionStart = el.selectionEnd = value.text.length
    compose.onTextareaInput()
    expect(compose.emojiQuery.value).toBeNull()
  })

  it('opens once at least one character follows the colon', () => {
    const { compose, el, value } = makeHarness()
    value.text = 'Note:s'
    el.selectionStart = el.selectionEnd = value.text.length
    compose.onTextareaInput()
    expect(compose.emojiQuery.value).toBe('s')
  })

  it('still opens the mention picker on a bare @ (unaffected by the emoji fix)', () => {
    const { compose, el, value } = makeHarness()
    value.text = 'Hello @'
    el.selectionStart = el.selectionEnd = value.text.length
    compose.onTextareaInput()
    expect(compose.mentionUsers.value.length).toBe(1)
  })
})
