/**
 * Optional text chat column during voice/video calls (1:1 WebRTC or LiveKit group).
 * Single shared toggled state so ActiveCallBar can show CallChatSidebar.
 */
import { ref } from 'vue'

const KEY = 'warmdesk_call_chat_open'
const WIDTH_KEY = 'warmdesk_call_chat_width'
const MIN_WIDTH = 260
const MAX_WIDTH = 600
const DEFAULT_WIDTH = 340

const showCallChat = ref(
  typeof localStorage !== 'undefined' && localStorage.getItem(KEY) === '1'
)

const callChatWidth = ref(
  Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, parseInt(localStorage.getItem(WIDTH_KEY)) || DEFAULT_WIDTH))
)

function persist() {
  try {
    localStorage.setItem(KEY, showCallChat.value ? '1' : '0')
  } catch {}
}

function setCallChatWidth(width) {
  callChatWidth.value = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, width))
  try {
    localStorage.setItem(WIDTH_KEY, String(callChatWidth.value))
  } catch {}
}

export function useCallChatPanel() {
  function toggleCallChat() {
    showCallChat.value = !showCallChat.value
    persist()
  }

  function setCallChat(open) {
    showCallChat.value = !!open
    persist()
  }

  return { showCallChat, toggleCallChat, setCallChat, callChatWidth, setCallChatWidth }
}
