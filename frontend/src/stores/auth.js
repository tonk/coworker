import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { setLocale } from '@/i18n'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))
  const accessToken = ref(localStorage.getItem('access_token') || null)

  const isLoggedIn = computed(() => !!accessToken.value && !!user.value)
  const isAdmin = computed(() => user.value?.global_role === 'admin')

  // ── Idle session timeout ─────────────────────────────────────────────────
  let idleTimer = null

  function startIdleTimer(timeoutMinutes) {
    stopIdleTimer()
    if (!timeoutMinutes || timeoutMinutes <= 0) return
    idleTimer = setTimeout(() => {
      logout()
      window.location.href = '/login'
    }, timeoutMinutes * 60 * 1000)
  }

  function resetIdleTimer(timeoutMinutes) {
    if (!timeoutMinutes || timeoutMinutes <= 0) return
    startIdleTimer(timeoutMinutes)
  }

  function stopIdleTimer() {
    if (idleTimer) {
      clearTimeout(idleTimer)
      idleTimer = null
    }
  }

  async function login(login, password) {
    const { data } = await authApi.login({ login, password })
    setTokens(data.access_token, data.refresh_token)
    await fetchMe()
  }

  async function register(payload) {
    const { data } = await authApi.register(payload)
    setTokens(data.access_token, data.refresh_token)
    await fetchMe()
  }

  async function fetchMe() {
    const { data } = await authApi.me()
    user.value = data
    localStorage.setItem('user', JSON.stringify(data))
    if (data.locale) setLocale(data.locale)
  }

  function setTokens(access, refresh) {
    accessToken.value = access
    localStorage.setItem('access_token', access)
    localStorage.setItem('refresh_token', refresh)
  }

  function logout() {
    stopIdleTimer()
    user.value = null
    accessToken.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
  }

  async function updateProfile(data) {
    const { data: updated } = await authApi.updateMe(data)
    user.value = updated
    localStorage.setItem('user', JSON.stringify(updated))
    if (data.locale) setLocale(data.locale)
  }

  return { user, accessToken, isLoggedIn, isAdmin, login, register, logout, fetchMe, updateProfile, startIdleTimer, resetIdleTimer, stopIdleTimer }
})
