<template>
  <div v-if="auth.isLoggedIn" class="app-shell">
    <UpdateBanner v-if="updateAvailable" :latest-version="latestVersion" :release-url="releaseUrl" />
    <AppHeader class="app-shell-header" />
    <div class="app-shell-body" :class="sidebarPos === 'right' ? 'sidebar-right' : 'sidebar-left'">
      <AppSidebar />
      <div class="app-shell-content">
        <RouterView />
        <footer class="app-footer">
          <span class="footer-left">Coworker v{{ appVersion }}</span>
          <span class="footer-right">{{ userFullName }}</span>
        </footer>
      </div>
    </div>
  </div>
  <RouterView v-else />
  <ToastContainer />
</template>

<script setup>
import { computed, watch, onMounted, onUnmounted } from 'vue'

const appVersion = __APP_VERSION__
import { RouterView } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import ToastContainer from '@/components/common/ToastContainer.vue'
import UpdateBanner from '@/components/common/UpdateBanner.vue'
import { applyUserPreferences } from '@/composables/useUserPreferences'
import { useUpdateCheck } from '@/composables/useUpdateCheck'

const auth = useAuthStore()
const systemStore = useSystemStore()

const { updateAvailable, latestVersion, releaseUrl, check: checkForUpdate } = useUpdateCheck()
// Run once when the user is logged in
watch(() => auth.isLoggedIn, (loggedIn) => {
  if (loggedIn) checkForUpdate(appVersion)
}, { immediate: true })

const sidebarPos = computed(() => auth.user?.sidebar_position || localStorage.getItem('sidebar_position') || 'left')

const userFullName = computed(() => {
  const u = auth.user
  if (!u) return ''
  const full = [u.first_name, u.last_name].filter(Boolean).join(' ')
  return full || u.display_name || u.username || ''
})

watch(() => auth.user, (user) => {
  if (user) applyUserPreferences(user)
}, { immediate: true })

// ── Idle session timeout ─────────────────────────────────────────────────────
const ACTIVITY_EVENTS = ['mousemove', 'mousedown', 'keydown', 'touchstart', 'scroll']

function onActivity() {
  if (auth.isLoggedIn) auth.resetIdleTimer(systemStore.sessionTimeoutMinutes)
}

watch([() => auth.isLoggedIn, () => systemStore.sessionTimeoutMinutes], ([loggedIn, timeout]) => {
  if (loggedIn && timeout > 0) {
    auth.startIdleTimer(timeout)
  } else {
    auth.stopIdleTimer()
  }
}, { immediate: true })

onMounted(() => {
  ACTIVITY_EVENTS.forEach(e => window.addEventListener(e, onActivity, { passive: true }))
})

onUnmounted(() => {
  ACTIVITY_EVENTS.forEach(e => window.removeEventListener(e, onActivity))
  auth.stopIdleTimer()
})
</script>

<style>
.app-shell {
  height: 100%;
  display: flex;
  flex-direction: column;
  font-family: var(--user-font, var(--font-family));
  font-size: var(--user-font-size, 14px);
}

.app-shell-header {
  flex-shrink: 0;
  position: sticky;
  top: 0;
  z-index: 100;
}

.app-shell-body {
  flex: 1;
  min-height: 0;
  display: flex;
  overflow: hidden;
}

.app-shell-body.sidebar-right {
  flex-direction: row-reverse;
}

.app-shell-content {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  display: flex;
  flex-direction: column;
}

.app-footer {
  margin-top: auto;
  padding: 8px 24px;
  font-size: 11px;
  color: var(--color-text-muted);
  border-top: 1px solid var(--color-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.footer-left { text-align: left; }
.footer-right { text-align: right; }
</style>
