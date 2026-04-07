import { defineStore } from 'pinia'
import { ref } from 'vue'
import { projectsApi } from '@/api/projects'

export const useChatStore = defineStore('chat', () => {
  const messages = ref([])
  const loading = ref(false)
  const hasMore = ref(true)
  const typingUsers = ref([])   // [{user_id, username, display_name, _timer}]

  function setTyping(user) {
    const existing = typingUsers.value.find(u => u.user_id === user.user_id)
    if (existing) {
      clearTimeout(existing._timer)
      existing._timer = setTimeout(() => clearTyping(user.user_id), 4000)
    } else {
      const entry = { ...user, _timer: setTimeout(() => clearTyping(user.user_id), 4000) }
      typingUsers.value.push(entry)
    }
  }

  function clearTyping(userId) {
    typingUsers.value = typingUsers.value.filter(u => u.user_id !== userId)
  }

  async function loadMessages(slug, before = null) {
    loading.value = true
    try {
      const params = { limit: 50 }
      if (before) params.before = before
      const { data } = await projectsApi.listMessages(slug, params)
      if (before) {
        messages.value = [...data, ...messages.value]
      } else {
        messages.value = data
      }
      hasMore.value = data.length === 50
    } finally {
      loading.value = false
    }
  }

  function addMessage(msg) {
    messages.value.push(msg)
  }

  function updateMessage({ id, body, is_edited }) {
    const msg = messages.value.find(m => m.id === id)
    if (msg) {
      msg.body = body
      msg.is_edited = is_edited
    }
  }

  function removeMessage({ id }) {
    const msg = messages.value.find(m => m.id === id)
    if (msg) msg.is_deleted = true
  }

  function updateReactions(messageId, reactions) {
    const msg = messages.value.find(m => m.id === messageId)
    if (msg) msg.reactions = reactions
  }

  function reset() {
    messages.value = []
    hasMore.value = true
  }

  function handleWsEvent(type, payload) {
    switch (type) {
      case 'chat.message.created': addMessage(payload); break
      case 'chat.message.updated': updateMessage(payload); break
      case 'chat.message.deleted': removeMessage(payload); break
      case 'chat.reaction.updated': updateReactions(payload.message_id, payload.reactions); break
      case 'chat.user.typing': setTyping(payload); break
    }
  }

  return { messages, loading, hasMore, typingUsers, loadMessages, addMessage, updateMessage, removeMessage, updateReactions, reset, handleWsEvent }
})
