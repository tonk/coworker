import { defineStore } from 'pinia'
import { ref } from 'vue'
import { projectsApi } from '@/api/projects'
import client from '@/api/client'

export const useSidebarStore = defineStore('sidebar', () => {
  const starredProjects = ref([])
  const allProjects = ref([])
  const allUsers = ref([])
  const chatUsers = ref([])
  const favoriteUsers = ref([])

  async function fetchStarred() {
    try {
      const { data } = await projectsApi.listStarred()
      starredProjects.value = data || []
    } catch {}
  }

  async function fetchAllProjects() {
    try {
      const { data } = await projectsApi.list()
      allProjects.value = data || []
    } catch {}
  }

  async function fetchAllUsers() {
    try {
      const { data } = await client.get('/users')
      allUsers.value = data || []
    } catch {}
  }

  async function fetchChatUsers() {
    try {
      const { data } = await client.get('/online-users')
      chatUsers.value = data || []
    } catch {}
  }

  async function fetchFavoriteUsers() {
    try {
      const { data } = await client.get('/favorite-users')
      favoriteUsers.value = data || []
    } catch {}
  }

  async function addFavoriteUser(userId) {
    await client.post(`/favorite-users/${userId}`)
    await fetchFavoriteUsers()
  }

  async function removeFavoriteUser(userId) {
    await client.delete(`/favorite-users/${userId}`)
    favoriteUsers.value = favoriteUsers.value.filter(u => u.id !== userId)
  }

  function isFavorite(userId) {
    return favoriteUsers.value.some(u => u.id === userId)
  }

  async function starProject(slug) {
    await projectsApi.starProject(slug)
    await fetchStarred()
  }

  async function unstarProject(slug) {
    await projectsApi.unstarProject(slug)
    starredProjects.value = starredProjects.value.filter(p => p.slug !== slug)
  }

  function isStarred(slug) {
    return starredProjects.value.some(p => p.slug === slug)
  }

  return {
    starredProjects, allProjects, allUsers, chatUsers, favoriteUsers,
    fetchStarred, fetchAllProjects, fetchAllUsers, fetchChatUsers,
    fetchFavoriteUsers, addFavoriteUser, removeFavoriteUser, isFavorite,
    starProject, unstarProject, isStarred
  }
})
