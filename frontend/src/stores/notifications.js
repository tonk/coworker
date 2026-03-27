import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { messagesApi } from '@/api/messages'
import { useAuthStore } from '@/stores/auth'

const STORAGE_KEY = 'messages_last_seen_at'

export const useNotificationsStore = defineStore('notifications', () => {
  // Timestamp (ms) when the user last had the messages view open
  const lastSeenAt = ref(parseInt(localStorage.getItem(STORAGE_KEY) || '0', 10))
  const conversations = ref([])

  const auth = useAuthStore()

  const hasUnread = computed(() =>
    conversations.value.some(c => {
      const updatedAt = new Date(c.updated_at).getTime()
      return updatedAt > lastSeenAt.value
    })
  )

  async function checkUnread() {
    if (!auth.isLoggedIn) return
    try {
      const { data } = await messagesApi.getConversations()
      conversations.value = data || []
    } catch {}
  }

  function markSeen() {
    lastSeenAt.value = Date.now()
    localStorage.setItem(STORAGE_KEY, String(lastSeenAt.value))
  }

  return { hasUnread, checkUnread, markSeen }
})
