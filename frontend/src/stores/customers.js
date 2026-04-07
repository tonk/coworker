import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { customersApi } from '@/api/customers'

export const useCustomersStore = defineStore('customers', () => {
  const customers = ref([])

  async function fetchCustomers() {
    try {
      const { data } = await customersApi.list()
      customers.value = data || []
    } catch {}
  }

  async function addFavorite(id) {
    await customersApi.addFavorite(id)
    const c = customers.value.find(c => c.id === id)
    if (c) c.is_favorite = true
  }

  async function removeFavorite(id) {
    await customersApi.removeFavorite(id)
    const c = customers.value.find(c => c.id === id)
    if (c) c.is_favorite = false
  }

  async function toggleFavorite(id) {
    const c = customers.value.find(c => c.id === id)
    if (!c) return
    if (c.is_favorite) {
      await removeFavorite(id)
    } else {
      await addFavorite(id)
    }
  }

  function isFavorite(id) {
    return customers.value.some(c => c.id === id && c.is_favorite)
  }

  const starredCustomers = computed(() =>
    customers.value.filter(c => c.is_favorite)
  )

  return {
    customers, starredCustomers,
    fetchCustomers, addFavorite, removeFavorite, toggleFavorite, isFavorite,
  }
})
