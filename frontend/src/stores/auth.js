import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { setLocale } from '@/i18n'
import client from '@/api/client'
import { setMfaTrustToken, clearMfaTrustToken } from '@/api/mfaTrust'

// In Tauri (desktop) mode tokens are stored in sessionStorage and sent via the
// Authorization header. In browser mode the server issues httpOnly cookies and
// no token ever touches JavaScript — the browser attaches them automatically.
const isTauri = !!window.__TAURI_INTERNALS__

export const useAuthStore = defineStore('auth', () => {
  const user = ref(isTauri ? JSON.parse(sessionStorage.getItem('user') || 'null') : null)
  const accessToken = ref(isTauri ? (sessionStorage.getItem('access_token') || null) : null)
  const pendingMFAToken = ref(null)
  const mfaSetupRequired = ref(false)

  // Seed axios Authorization header from stored token (Tauri only).
  if (isTauri && accessToken.value) {
    client.defaults.headers.common.Authorization = `Bearer ${accessToken.value}`
  }

  // isLoggedIn is driven by the user profile ref only.
  // In Tauri mode user comes from sessionStorage; in browser mode it is fetched
  // from the server on startup via initSession() before the app mounts.
  const isLoggedIn = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.global_role === 'admin')
  const canViewReports = computed(() => isAdmin.value || !!user.value?.can_view_reports)
  const timeTrackingEnabled = computed(() => !!user.value?.time_tracking_enabled || !!user.value?.time_tracking_viewer)
  const boardEnabled = computed(() => isAdmin.value || user.value?.board_enabled !== false)
  const chatEnabled = computed(() => isAdmin.value || user.value?.chat_enabled !== false)
  const helpdeskEnabled = computed(() => isAdmin.value || !!user.value?.helpdesk_enabled)

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

  // ── Token storage (Tauri only) ───────────────────────────────────────────
  function setTokens(access, refresh) {
    if (!isTauri) return
    accessToken.value = access
    sessionStorage.setItem('access_token', access)
    sessionStorage.setItem('refresh_token', refresh)
    client.defaults.headers.common.Authorization = `Bearer ${access}`
  }

  // ── Session restoration ──────────────────────────────────────────────────
  // Called once before app mount in browser mode to hydrate user state from
  // the httpOnly cookie without exposing the token to JavaScript.
  async function initSession() {
    if (isTauri) return
    try {
      await fetchMe()
    } catch {
      user.value = null
    }
  }

  // ── Auth actions ─────────────────────────────────────────────────────────
  async function login(login, password) {
    const { data } = await authApi.login({ login, password })
    if (data.mfa_required) {
      pendingMFAToken.value = data.mfa_token
      return { mfa_required: true }
    }
    setTokens(data.access_token, data.refresh_token)
    mfaSetupRequired.value = !!data.mfa_setup_required
    await fetchMe()
    return { password_expired: !!data.password_expired, must_change_password: !!data.must_change_password }
  }

  async function verifyMFA(code, rememberDays = 0) {
    const { data } = await authApi.verifyMFA(pendingMFAToken.value, code, rememberDays)
    pendingMFAToken.value = null
    setTokens(data.access_token, data.refresh_token)
    if (data.mfa_trust_token) setMfaTrustToken(data.mfa_trust_token)
    await fetchMe()
    return { must_change_password: !!data.must_change_password }
  }

  async function register(payload) {
    const { data } = await authApi.register(payload)
    setTokens(data.access_token, data.refresh_token)
    await fetchMe()
  }

  async function fetchMe() {
    const { data } = await authApi.me()
    user.value = data
    if (isTauri) sessionStorage.setItem('user', JSON.stringify(data))
    if (data.locale) await setLocale(data.locale)
  }

  function logout() {
    stopIdleTimer()
    user.value = null
    accessToken.value = null
    pendingMFAToken.value = null
    mfaSetupRequired.value = false
    // Tell the server to expire the httpOnly cookies. Fire-and-forget so the
    // UI clears immediately even if the network is slow.
    authApi.logout().catch(() => {})
    clearMfaTrustToken()
    if (isTauri) {
      sessionStorage.removeItem('access_token')
      sessionStorage.removeItem('refresh_token')
      sessionStorage.removeItem('user')
      delete client.defaults.headers.common.Authorization
    }
  }

  async function updateProfile(data) {
    const { data: updated } = await authApi.updateMe(data)
    user.value = updated
    if (isTauri) sessionStorage.setItem('user', JSON.stringify(updated))
    if (data.locale) await setLocale(data.locale)
  }

  return {
    user, accessToken, isLoggedIn, isAdmin, canViewReports, timeTrackingEnabled,
    boardEnabled, chatEnabled, helpdeskEnabled,
    pendingMFAToken, mfaSetupRequired,
    login, verifyMFA, register, logout, fetchMe, updateProfile, initSession, setTokens,
    startIdleTimer, resetIdleTimer, stopIdleTimer,
  }
})
