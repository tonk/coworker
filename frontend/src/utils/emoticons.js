import { gemoji } from 'gemoji'

// Text emoticons → emoji replacements.
// Sorted longest-first so greedy matching prefers e.g. ":-)" over ":)".
export const EMOTICONS = [
  ["O:-)",  "😇"],
  [":'-)",  "😂"],
  [":'-(", "😢"],
  [">:-(", "😠"],
  [":-)",  "😊"],
  [":-D",  "😄"],
  [":-P",  "😛"],
  [":-p",  "😛"],
  [":-O",  "😮"],
  [":-o",  "😮"],
  [":-*",  "😘"],
  [":-|",  "😐"],
  [":-/",  "😕"],
  [":-X",  "🤐"],
  [":-x",  "🤐"],
  [";-)",  "😉"],
  ["B-)",  "😎"],
  ["O:)",  "😇"],
  ["</3",  "💔"],
  [">:(",  "😠"],
  [":'(",  "😢"],
  [":)",   "😊"],
  [":(",   "😢"],
  [":D",   "😄"],
  [":P",   "😛"],
  [":p",   "😛"],
  [":O",   "😮"],
  [":o",   "😮"],
  [":*",   "😘"],
  [":|",   "😐"],
  [":/",   "😕"],
  [":X",   "🤐"],
  [":x",   "🤐"],
  [";)",   "😉"],
  ["<3",   "❤️"],
]

// Shared quick-reaction defaults for chat hover toolbars.
export const QUICK_REACTION_EMOJIS = ['👍', '👎', '😂', '🤣', '✅', '❌', '🤘']

// Full GitHub-style :shortcode: support via gemoji.
// We also add underscore-less aliases (e.g. :fingerscrossed:) for convenience.
export const EMOJI_SHORTCODES = Object.create(null)
for (const entry of gemoji) {
  const emoji = entry.emoji
  for (const name of entry.names || []) {
    const key = String(name || '').toLowerCase()
    if (!key) continue
    EMOJI_SHORTCODES[key] = emoji
    EMOJI_SHORTCODES[key.replace(/_/g, '')] = emoji
  }
}

/**
 * Check if `text` ends with a known emoticon that is either at the start of
 * the string or preceded by a whitespace character.
 * Returns { pattern, emoji } if found, otherwise null.
 */
export function detectEmoticon(text) {
  for (const [pattern, emoji] of EMOTICONS) {
    if (!text.endsWith(pattern)) continue
    const before = text[text.length - pattern.length - 1]
    if (before !== undefined && before !== ' ' && before !== '\n' && before !== '\t') continue
    return { pattern, emoji }
  }
  return null
}

/**
 * Check if `text` ends with a known :shortcode: token.
 * Returns { pattern, emoji } if found, otherwise null.
 */
export function detectEmojiShortcode(text) {
  const m = text.match(/:([a-z0-9_+\-]+):$/i)
  if (!m) return null
  const pattern = m[0]
  const name = (m[1] || '').toLowerCase()
  const emoji = EMOJI_SHORTCODES[name]
  if (!emoji) return null
  const before = text[text.length - pattern.length - 1]
  if (before !== undefined && before !== ' ' && before !== '\n' && before !== '\t') return null
  return { pattern, emoji }
}
