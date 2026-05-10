/**
 * Optional text chat column during voice/video calls (1:1 WebRTC or LiveKit group).
 * Single shared toggled state so ActiveCallBar can show CallChatSidebar.
 */
import { ref } from 'vue'

const KEY = 'warmdesk_call_chat_open'

const showCallChat = ref(
  typeof localStorage !== 'undefined' && localStorage.getItem(KEY) === '1'
)

function persist() {
  try {
    localStorage.setItem(KEY, showCallChat.value ? '1' : '0')
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

  return { showCallChat, toggleCallChat, setCallChat }
}
