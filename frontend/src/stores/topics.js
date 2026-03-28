import { defineStore } from 'pinia'
import { ref } from 'vue'
import { topicsApi } from '@/api/topics'

export const useTopicsStore = defineStore('topics', () => {
  const topics = ref([])
  const loading = ref(false)

  async function loadTopics(slug) {
    loading.value = true
    try {
      const { data } = await topicsApi.list(slug)
      topics.value = data || []
    } finally {
      loading.value = false
    }
  }

  function handleWsEvent(type, payload) {
    switch (type) {
      case 'topic.created':
        if (!topics.value.find(t => t.id === payload.id)) {
          topics.value = [payload, ...topics.value]
        }
        break
      case 'topic.updated': {
        const idx = topics.value.findIndex(t => t.id === payload.id)
        if (idx !== -1) topics.value[idx] = { ...topics.value[idx], ...payload }
        break
      }
      case 'topic.deleted':
        topics.value = topics.value.filter(t => t.id !== payload.id)
        break
    }
  }

  function reset() {
    topics.value = []
  }

  return { topics, loading, loadTopics, handleWsEvent, reset }
})
