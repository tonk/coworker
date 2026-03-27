import { defineStore } from 'pinia'
import { ref } from 'vue'
import { projectsApi } from '@/api/projects'

export const useBoardStore = defineStore('board', () => {
  const columns = ref([])
  const loading = ref(false)
  const projectSlug = ref(null)

  async function loadBoard(slug) {
    loading.value = true
    projectSlug.value = slug
    try {
      const { data } = await projectsApi.get(slug)
      columns.value = (data.columns || []).sort((a, b) => a.position - b.position).map(col => ({
        ...col,
        cards: (col.cards || []).sort((a, b) => a.position - b.position)
      }))
    } finally {
      loading.value = false
    }
  }

  function reset() {
    columns.value = []
    projectSlug.value = null
  }

  // Column mutations
  function addColumn(col) {
    columns.value.push({ ...col, cards: [] })
  }

  function updateColumn(col) {
    const idx = columns.value.findIndex(c => c.id === col.id)
    if (idx !== -1) columns.value[idx] = { ...columns.value[idx], ...col }
  }

  function removeColumn(columnId) {
    columns.value = columns.value.filter(c => c.id !== columnId)
  }

  function reorderColumns(items) {
    items.forEach(({ id, position }) => {
      const col = columns.value.find(c => c.id === id)
      if (col) col.position = position
    })
    columns.value.sort((a, b) => a.position - b.position)
  }

  // Card mutations
  function addCard(card) {
    const col = columns.value.find(c => c.id === card.column_id)
    if (col) {
      col.cards.push(card)
      col.cards.sort((a, b) => a.position - b.position)
    }
  }

  function updateCard(card) {
    for (const col of columns.value) {
      const idx = col.cards.findIndex(c => c.id === card.id)
      if (idx !== -1) {
        col.cards[idx] = { ...col.cards[idx], ...card }
        return
      }
    }
  }

  function moveCard({ card_id, from_column_id, to_column_id, position }) {
    const fromCol = columns.value.find(c => c.id === from_column_id)
    const toCol = columns.value.find(c => c.id === to_column_id)
    if (!fromCol || !toCol) return

    const cardIdx = fromCol.cards.findIndex(c => c.id === card_id)
    if (cardIdx === -1) return

    const [card] = fromCol.cards.splice(cardIdx, 1)
    card.column_id = to_column_id
    card.position = position
    toCol.cards.push(card)
    toCol.cards.sort((a, b) => a.position - b.position)
  }

  function removeCard({ card_id, column_id }) {
    const col = columns.value.find(c => c.id === column_id)
    if (col) col.cards = col.cards.filter(c => c.id !== card_id)
  }

  // API actions
  async function createColumn(name, extra = {}) {
    const { data } = await projectsApi.createColumn(projectSlug.value, { name, ...extra })
    // addColumn is intentionally omitted here: the backend broadcasts board.column.created
    // via WebSocket to all clients including the creator, so handleWsEvent does the insert.
    return data
  }

  async function createCard(columnId, payload) {
    const { data } = await projectsApi.createCard(projectSlug.value, columnId, payload)
    // addCard is intentionally omitted here: the backend broadcasts board.card.created
    // via WebSocket to all clients including the creator, so handleWsEvent does the insert.
    return data
  }

  async function deleteCard(cardId, columnId) {
    await projectsApi.deleteCard(projectSlug.value, cardId)
    removeCard({ card_id: cardId, column_id: columnId })
  }

  async function updateCardData(cardId, payload) {
    const { data } = await projectsApi.updateCard(projectSlug.value, cardId, payload)
    updateCard(data)
    return data
  }

  // WebSocket event handlers
  function handleWsEvent(type, payload) {
    switch (type) {
      case 'board.card.created': addCard(payload); break
      case 'board.card.updated': updateCard(payload); break
      case 'board.card.moved': moveCard(payload); break
      case 'board.card.deleted': removeCard(payload); break
      case 'board.column.created': addColumn(payload); break
      case 'board.column.updated': updateColumn(payload); break
      case 'board.column.deleted': removeColumn(payload.column_id); break
      case 'board.columns.reordered': reorderColumns(payload); break
    }
  }

  return {
    columns, loading, projectSlug,
    loadBoard, reset,
    addColumn, updateColumn, removeColumn, reorderColumns,
    addCard, updateCard, moveCard, removeCard,
    createColumn, createCard, deleteCard, updateCardData,
    handleWsEvent
  }
})
