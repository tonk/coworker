import { defineStore } from 'pinia'
import { ref } from 'vue'
import { ticketsApi } from '@/api/tickets'

export const useTicketsStore = defineStore('tickets', () => {
  const tickets = ref([])

  async function fetchTickets(customerId) {
    try {
      const { data } = await ticketsApi.list(customerId)
      tickets.value = data || []
    } catch {}
  }

  async function createTicket(customerId, data) {
    const { data: ticket } = await ticketsApi.create(customerId, data)
    tickets.value.unshift(ticket)
    return ticket
  }

  async function updateTicket(customerId, id, data) {
    const { data: ticket } = await ticketsApi.update(customerId, id, data)
    const idx = tickets.value.findIndex(t => t.id === id)
    if (idx >= 0) tickets.value[idx] = ticket
    return ticket
  }

  async function deleteTicket(customerId, id) {
    await ticketsApi.delete(customerId, id)
    tickets.value = tickets.value.filter(t => t.id !== id)
  }

  return {
    tickets,
    fetchTickets,
    createTicket,
    updateTicket,
    deleteTicket,
  }
})
