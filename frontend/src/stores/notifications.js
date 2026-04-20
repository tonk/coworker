import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { messagesApi } from '@/api/messages'
import { useAuthStore } from '@/stores/auth'

const STORAGE_KEY = 'conv_last_seen'

function loadSeen() {
  try { return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}') } catch { return {} }
}

export const useNotificationsStore = defineStore('notifications', () => {
  // Per-conversation last-seen timestamps: { [convId]: ms }
  const convLastSeen = ref(loadSeen())
  const conversations = ref([])

  const auth = useAuthStore()

  const hasUnread = computed(() =>
    conversations.value.some(c => isConvUnread(c))
  )

  async function checkUnread() {
    if (!auth.isLoggedIn) return
    try {
      const { data } = await messagesApi.getConversations()
      conversations.value = data || []
      // Seed a baseline for conversations never opened on this client so they
      // don't permanently show as unread. Only messages arriving after this
      // point will trigger the indicator.
      let changed = false
      for (const c of conversations.value) {
        if (convLastSeen.value[c.id] == null) {
          convLastSeen.value[c.id] = new Date(c.updated_at).getTime()
          changed = true
        }
      }
      if (changed) localStorage.setItem(STORAGE_KEY, JSON.stringify(convLastSeen.value))
    } catch {}
  }

  // Call when the user opens a conversation or sends a message in it
  function markConvSeen(convId) {
    convLastSeen.value[convId] = Date.now()
    localStorage.setItem(STORAGE_KEY, JSON.stringify(convLastSeen.value))
  }

  // Legacy: mark all current conversations as seen at once
  function markSeen() {
    const now = Date.now()
    for (const c of conversations.value) {
      convLastSeen.value[c.id] = now
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(convLastSeen.value))
  }

  function isConvUnread(conv) {
    const seen = convLastSeen.value[conv.id] || 0
    return new Date(conv.updated_at).getTime() > seen
  }

  return { hasUnread, conversations, isConvUnread, checkUnread, markSeen, markConvSeen }
})
