import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { projectsApi } from '@/api/projects'

export const useSprintStore = defineStore('sprint', () => {
  const sprints = ref([])
  const backlog = ref([])
  const loading = ref(false)
  const projectSlug = ref(null)

  const activeSprint = computed(() => sprints.value.find(s => s.status === 'active') || null)
  const planningSprints = computed(() => sprints.value.filter(s => s.status === 'planning'))
  const completedSprints = computed(() => sprints.value.filter(s => s.status === 'completed'))

  async function loadSprints(slug) {
    projectSlug.value = slug
    loading.value = true
    try {
      const { data } = await projectsApi.listSprints(slug)
      sprints.value = data
    } finally {
      loading.value = false
    }
  }

  async function loadBacklog(slug) {
    const { data } = await projectsApi.listBacklog(slug)
    backlog.value = data
  }

  async function createSprint(data) {
    const res = await projectsApi.createSprint(projectSlug.value, data)
    sprints.value.push(res.data)
    return res.data
  }

  async function updateSprint(sprintId, data) {
    const res = await projectsApi.updateSprint(projectSlug.value, sprintId, data)
    const idx = sprints.value.findIndex(s => s.id === sprintId)
    if (idx !== -1) sprints.value[idx] = res.data
    return res.data
  }

  async function deleteSprint(sprintId) {
    await projectsApi.deleteSprint(projectSlug.value, sprintId)
    sprints.value = sprints.value.filter(s => s.id !== sprintId)
  }

  async function startSprint(sprintId) {
    const { data } = await projectsApi.startSprint(projectSlug.value, sprintId)
    const idx = sprints.value.findIndex(s => s.id === sprintId)
    if (idx !== -1) sprints.value[idx] = data
    return data
  }

  async function completeSprint(sprintId) {
    const { data } = await projectsApi.completeSprint(projectSlug.value, sprintId)
    const idx = sprints.value.findIndex(s => s.id === sprintId)
    if (idx !== -1) sprints.value[idx] = data
    // Reload backlog since unfinished cards return here
    await loadBacklog(projectSlug.value)
    return data
  }

  async function addCardToSprint(sprintId, cardId) {
    await projectsApi.addCardToSprint(projectSlug.value, sprintId, cardId)
    // Update sprint card_ids locally
    const sprint = sprints.value.find(s => s.id === sprintId)
    if (sprint && !sprint.card_ids.includes(cardId)) {
      sprint.card_ids.push(cardId)
      sprint.card_count++
    }
    // Remove from backlog locally
    backlog.value = backlog.value.filter(c => c.id !== cardId)
  }

  async function reorderBacklog(items) {
    await projectsApi.reorderBacklog(projectSlug.value, items)
  }

  async function reorderSprints(items) {
    await projectsApi.reorderSprints(projectSlug.value, items)
  }

  async function removeCardFromSprint(sprintId, cardId) {
    await projectsApi.removeCardFromSprint(projectSlug.value, sprintId, cardId)
    const sprint = sprints.value.find(s => s.id === sprintId)
    if (sprint) {
      sprint.card_ids = sprint.card_ids.filter(id => id !== cardId)
      sprint.card_count = Math.max(0, sprint.card_count - 1)
    }
    // Reload backlog to show the returned card
    await loadBacklog(projectSlug.value)
  }

  // WebSocket event handlers
  function handleWsEvent(type, payload) {
    switch (type) {
      case 'sprint.created':
        if (!sprints.value.some(s => s.id === payload.id)) sprints.value.push(payload)
        break
      case 'sprint.updated':
      case 'sprint.started':
      case 'sprint.completed': {
        const idx = sprints.value.findIndex(s => s.id === payload.id)
        if (idx !== -1) sprints.value[idx] = payload
        break
      }
      case 'sprint.deleted':
        sprints.value = sprints.value.filter(s => s.id !== payload.sprint_id)
        break
      case 'sprint.card.added': {
        const s = sprints.value.find(s => s.id === payload.sprint_id)
        if (s && !s.card_ids.includes(payload.card_id)) {
          s.card_ids.push(payload.card_id)
          s.card_count++
        }
        break
      }
      case 'sprint.card.removed': {
        const s = sprints.value.find(s => s.id === payload.sprint_id)
        if (s) {
          s.card_ids = s.card_ids.filter(id => id !== payload.card_id)
          s.card_count = Math.max(0, s.card_count - 1)
        }
        break
      }
    }
  }

  function reset() {
    sprints.value = []
    backlog.value = []
    projectSlug.value = null
  }

  return {
    sprints, backlog, loading, projectSlug,
    activeSprint, planningSprints, completedSprints,
    loadSprints, loadBacklog, createSprint, updateSprint, deleteSprint,
    startSprint, completeSprint, addCardToSprint, removeCardFromSprint,
    reorderBacklog, reorderSprints, handleWsEvent, reset,
  }
})
