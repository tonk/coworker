import { defineStore } from 'pinia'
import { ref } from 'vue'
import { epicsApi } from '@/api/epics'

export const useEpicsStore = defineStore('epics', () => {
  const epics = ref([])
  const loading = ref(false)
  const projectSlug = ref(null)

  async function loadEpics(slug) {
    projectSlug.value = slug
    loading.value = true
    try {
      const { data } = await epicsApi.list(slug)
      epics.value = data || []
    } finally {
      loading.value = false
    }
  }

  async function createEpic(data) {
    const { data: epic } = await epicsApi.create(projectSlug.value, data)
    epics.value.push(epic)
    return epic
  }

  async function updateEpic(id, data) {
    const { data: epic } = await epicsApi.update(projectSlug.value, id, data)
    const idx = epics.value.findIndex(e => e.id === id)
    if (idx !== -1) epics.value[idx] = epic
    return epic
  }

  async function deleteEpic(id) {
    await epicsApi.delete(projectSlug.value, id)
    epics.value = epics.value.filter(e => e.id !== id)
  }

  async function reorderEpics(items) {
    await epicsApi.reorder(projectSlug.value, items)
  }

  function reset() {
    epics.value = []
    projectSlug.value = null
  }

  return { epics, loading, projectSlug, loadEpics, createEpic, updateEpic, deleteEpic, reorderEpics, reset }
})
