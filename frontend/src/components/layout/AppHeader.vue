<template>
  <header class="app-header">
    <div class="header-left">
      <RouterLink to="/" class="logo">
        <img src="/logo-full.svg" alt="WarmDesk" class="logo-img" />
      </RouterLink>
      <slot name="breadcrumb" />
    </div>
    <div v-if="!systemStore.isTimetrackingMode" class="header-center">
      <GlobalSearch />
    </div>
    <nav class="header-nav" aria-label="Main navigation">
      <RouterLink v-if="!systemStore.isTimetrackingMode" to="/" class="header-nav-link" :class="{ active: route.path === '/' }">{{ $t('nav.dashboard') }}</RouterLink>
      <RouterLink to="/news" class="header-nav-link" :class="{ active: route.path === '/news' }">{{ $t('nav.news') }}</RouterLink>
      <RouterLink v-if="!systemStore.isTimetrackingMode" to="/chats" class="header-nav-link" :class="{ active: route.path === '/chats' }">
        {{ $t('nav.messages') }}
        <span v-if="notificationsStore.hasUnread" class="nav-unread-dot" aria-hidden="true"></span>
        <span v-if="notificationsStore.hasUnread" class="sr-only">({{ $t('sidebar.unread_messages') }})</span>
      </RouterLink>
      <RouterLink v-if="!systemStore.isTimetrackingMode && (auth.timeTrackingEnabled || auth.canViewReports)" to="/time-tracking" class="header-nav-link" :class="{ active: route.path.startsWith('/time-tracking') }">{{ $t('timeTracking.nav') }}</RouterLink>
    </nav>
    <div class="header-right">
      <span class="presence-count" v-if="presenceCount > 0" :title="`${presenceCount} ${$t('presence.online')}`">
        <span class="presence-dot"></span>{{ presenceCount }}
      </span>
      <div class="lang-switcher">
        <select class="form-input lang-select" :value="locale" @change="onLocaleChange" :aria-label="$t('common.language')">
          <option value="en">EN</option>
          <option value="nl">NL</option>
          <option value="de">DE</option>
          <option value="es">ES</option>
          <option value="fr">FR</option>
          <option value="da">DA</option>
          <option value="sv">SV</option>
          <option value="nb">NB</option>
          <option value="fi">FI</option>
          <option value="is">IS</option>
          <option value="pt">PT</option>
          <option value="it">IT</option>
        </select>
      </div>
      <div class="theme-switcher" ref="themeRef">
        <button
          class="btn-icon"
          @click.stop="themeOpen = !themeOpen"
          :aria-label="$t('settings.theme')"
          :aria-expanded="themeOpen"
          aria-haspopup="listbox"
        >
          <!-- sun: light mode -->
          <svg v-if="theme === 'light'" aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="5"/>
            <line x1="12" y1="1" x2="12" y2="3"/>
            <line x1="12" y1="21" x2="12" y2="23"/>
            <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
            <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
            <line x1="1" y1="12" x2="3" y2="12"/>
            <line x1="21" y1="12" x2="23" y2="12"/>
            <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
            <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
          </svg>
          <!-- moon: dark mode -->
          <svg v-else-if="theme === 'dark'" aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
          <!-- monitor: system theme -->
          <svg v-else aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="3" width="20" height="14" rx="2"/>
            <polyline points="8 21 12 17 16 21"/>
          </svg>
        </button>
        <div class="dropdown theme-dropdown" v-if="themeOpen" role="listbox" :aria-label="$t('settings.theme')">
          <div class="dropdown-item" role="option" :aria-selected="theme === 'light'" :class="{ 'dropdown-item-active': theme === 'light' }" @click="selectTheme('light')">
            <svg aria-hidden="true" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="5"/>
              <line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/>
              <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
              <line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/>
              <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
            </svg>
            {{ $t('settings.theme_light') }}
          </div>
          <div class="dropdown-item" role="option" :aria-selected="theme === 'dark'" :class="{ 'dropdown-item-active': theme === 'dark' }" @click="selectTheme('dark')">
            <svg aria-hidden="true" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
            </svg>
            {{ $t('settings.theme_dark') }}
          </div>
          <div class="dropdown-item" role="option" :aria-selected="theme === 'system'" :class="{ 'dropdown-item-active': theme === 'system' }" @click="selectTheme('system')">
            <svg aria-hidden="true" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="3" width="20" height="14" rx="2"/>
              <polyline points="8 21 12 17 16 21"/>
            </svg>
            {{ $t('settings.theme_system') }}
          </div>
        </div>
      </div>
      <button
        type="button"
        class="btn-icon"
        @click="showHelp = true"
        :aria-label="$t('help.button')"
      >
        <svg aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/>
          <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
          <line x1="12" y1="17" x2="12.01" y2="17"/>
        </svg>
      </button>
      <div class="user-menu" v-if="auth.user" ref="menuRef">
        <button
          class="avatar-btn"
          @click.stop="menuOpen = !menuOpen"
          :aria-expanded="menuOpen"
          aria-haspopup="menu"
          :aria-label="$t('nav.user_menu')"
        >
          <div class="avatar">
            <img v-if="userAvatar" :src="userAvatar" :alt="initials" class="avatar-img" @error="avatarErr = true" />
            <span v-else aria-hidden="true">{{ initials }}</span>
          </div>
        </button>
        <div class="dropdown" v-if="menuOpen" role="menu" @keydown="handleMenuKeyDown">
          <div v-if="!systemStore.isTimetrackingMode" class="dropdown-item" role="menuitem" @click="navigate('/')">{{ $t('nav.dashboard') }}</div>
          <div class="dropdown-item" role="menuitem" @click="navigate('/news')">{{ $t('nav.news') }}</div>
          <div class="dropdown-item" role="menuitem" @click="navigate('/settings')">{{ $t('nav.settings') }}</div>
          <div class="dropdown-item" role="menuitem" v-if="auth.isAdmin" @click="navigate('/admin')">{{ $t('nav.admin') }}</div>
          <div class="dropdown-divider" role="separator"></div>
          <div v-if="!systemStore.isTimetrackingMode" class="dropdown-item" role="menuitem" @click="navigate('/chats')">
            {{ $t('nav.messages') }}
            <span v-if="notificationsStore.hasUnread" class="msg-unread-dot" aria-hidden="true"></span>
            <span v-if="notificationsStore.hasUnread" class="sr-only">({{ $t('sidebar.unread_messages') }})</span>
          </div>
          <div class="dropdown-item" role="menuitem" v-if="auth.timeTrackingEnabled || auth.canViewReports" @click="navigate('/time-tracking')">{{ $t('timeTracking.nav') }}</div>
          <div class="dropdown-divider" role="separator"></div>
          <div class="dropdown-item" role="menuitem" @click="openShortcuts">{{ $t('nav.keyboard_shortcuts') }}</div>
          <div
            class="dropdown-submenu"
            :class="{ 'dropdown-submenu-open': downloadsOpen }"
          >
            <button
              ref="downloadsTriggerRef"
              type="button"
              class="dropdown-item dropdown-submenu-trigger"
              role="menuitem"
              aria-haspopup="menu"
              :aria-expanded="downloadsOpen"
              :aria-label="$t('nav.downloads')"
              @click.stop="toggleDownloads"
            >
              <span>{{ $t('nav.downloads') }}</span>
              <svg class="dropdown-submenu-chevron" aria-hidden="true" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="15 18 9 12 15 6"/>
              </svg>
            </button>
            <div
              ref="downloadsPanelRef"
              class="dropdown-submenu-panel"
              role="menu"
              :aria-label="$t('nav.downloads')"
              @keydown="handleDownloadsKeyDown"
            >
              <button type="button" class="dropdown-item" role="menuitem" @click="downloadUserGuide">{{ $t('nav.user_guide') }}</button>
              <button type="button" class="dropdown-item" role="menuitem" v-if="auth.isAdmin" @click="downloadAdminGuide">{{ $t('nav.admin_guide') }}</button>
            </div>
          </div>
          <div class="dropdown-item" role="menuitem" @click="openAbout">{{ $t('nav.about') }}</div>
          <div class="dropdown-item dropdown-item-danger" role="menuitem" @click="handleLogout">{{ $t('nav.logout') }}</div>
        </div>
      </div>
    </div>
  </header>
  <AboutModal v-if="showAbout" @close="showAbout = false" />
  <HelpPanelModal v-if="showHelp" @close="showHelp = false" />
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { RouterLink, useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'
import { setLocale } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { useNotificationsStore } from '@/stores/notifications'
import { avatarUrl } from '@/composables/useAvatar'
import { fetchBinary, triggerDownload } from '@/api/client'
import { useUIStore } from '@/stores/ui'
import GlobalSearch from '@/components/common/GlobalSearch.vue'
import AboutModal from '@/components/common/AboutModal.vue'
import HelpPanelModal from '@/components/common/HelpPanelModal.vue'

const props = defineProps({ presenceCount: { type: Number, default: 0 } })
const emit = defineEmits(['open-shortcuts'])

const auth = useAuthStore()
const systemStore = useSystemStore()
const router = useRouter()
const route = useRoute()
const { locale, t } = useI18n()
const { theme, setTheme } = useTheme()
const notificationsStore = useNotificationsStore()
const ui = useUIStore()
const menuOpen = ref(false)
const downloadsOpen = ref(false)
const themeOpen = ref(false)
const showAbout = ref(false)
const showHelp = ref(false)
const menuRef = ref(null)
const themeRef = ref(null)
const downloadsTriggerRef = ref(null)
const downloadsPanelRef = ref(null)
const avatarErr = ref(false)

const userAvatar = computed(() => avatarErr.value ? null : avatarUrl(auth.user))

function selectTheme(value) {
  setTheme(value)
  themeOpen.value = false
  if (auth.isLoggedIn) auth.updateProfile({ theme: value })
}

const initials = computed(() => {
  const name = auth.user?.display_name || auth.user?.username || '?'
  return name.slice(0, 2).toUpperCase()
})

async function onLocaleChange(e) {
  const lang = e.target.value
  await setLocale(lang)
  if (auth.isLoggedIn) auth.updateProfile({ locale: lang })
}

function closeMenu() {
  menuOpen.value = false
  downloadsOpen.value = false
}

function navigate(path) {
  closeMenu()
  router.push(path)
}

function openShortcuts() {
  closeMenu()
  emit('open-shortcuts')
}

function openAbout() {
  closeMenu()
  showAbout.value = true
}

function closeDownloads() {
  downloadsOpen.value = false
}

function toggleDownloads() {
  downloadsOpen.value = !downloadsOpen.value
  if (downloadsOpen.value) {
    nextTick(() => {
      const first = downloadsPanelRef.value?.querySelector('[role="menuitem"]')
      first?.focus()
    })
  }
}

function openDownloads() {
  downloadsOpen.value = true
  nextTick(() => {
    const first = downloadsPanelRef.value?.querySelector('[role="menuitem"]')
    first?.focus()
  })
}

function guideDownloadFilename(slug) {
  const ver = __APP_VERSION__.startsWith('v') ? __APP_VERSION__ : `v${__APP_VERSION__}`
  return `WarmDesk-${slug}-${ver}.pdf`
}

async function downloadGuide(path, slug) {
  closeMenu()
  try {
    const data = await fetchBinary(path)
    await triggerDownload(data, guideDownloadFilename(slug), 'application/pdf')
  } catch {
    ui.error(t('nav.guide_download_error'))
  }
}

async function downloadUserGuide() {
  await downloadGuide('/docs/user-guide.pdf', 'user-guide')
}

async function downloadAdminGuide() {
  await downloadGuide('/docs/admin-guide.pdf', 'admin-guide')
}

function handleDownloadsKeyDown(e) {
  if (e.key === 'ArrowLeft' || e.key === 'Escape') {
    e.preventDefault()
    e.stopPropagation()
    closeDownloads()
    downloadsTriggerRef.value?.focus()
  }
}

function handleMenuKeyDown(e) {
  if (e.target !== downloadsTriggerRef.value) return
  if (e.key === 'ArrowRight' || e.key === 'Enter' || e.key === ' ') {
    e.preventDefault()
    openDownloads()
  }
}

function handleLogout() {
  closeMenu()
  auth.logout()
  router.push('/login')
}

function handleClick(e) {
  if (menuRef.value && !menuRef.value.contains(e.target)) closeMenu()
  if (themeRef.value && !themeRef.value.contains(e.target)) themeOpen.value = false
}

function handleKeyDown(e) {
  if (e.key === 'Escape') {
    if (downloadsOpen.value) {
      closeDownloads()
      downloadsTriggerRef.value?.focus()
      return
    }
    closeMenu()
    themeOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClick)
  document.addEventListener('keydown', handleKeyDown)
  document.documentElement.lang = locale.value
})
onBeforeUnmount(() => {
  document.removeEventListener('click', handleClick)
  document.removeEventListener('keydown', handleKeyDown)
})
</script>

<style scoped>
.app-header {
  height: 56px;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left { display: flex; align-items: center; gap: 16px; }
.header-center { flex: 1; display: flex; justify-content: center; padding: 0 16px; }
.header-right { display: flex; align-items: center; gap: 12px; }

.logo { text-decoration: none; display: flex; align-items: center; }
.logo-img { height: 28px; width: auto; display: block; }

.lang-select { width: 60px; padding: 4px 6px; font-size: 12px; }

.avatar-btn {
  background: transparent;
  border: none;
  padding: 0;
  cursor: pointer;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.avatar-btn:focus-visible { outline-offset: 3px; }

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--color-primary);
  color: var(--color-text-on-primary, #fff);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  overflow: hidden;
}
.avatar-img { width: 100%; height: 100%; object-fit: cover; border-radius: 50%; }

.user-menu { position: relative; }

.dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-md);
  min-width: 280px;
  width: max-content;
  max-width: min(360px, calc(100vw - 24px));
  z-index: 200;
}

.dropdown-item {
  padding: 10px 16px;
  cursor: pointer;
  font-size: 14px;
  color: var(--color-text);
  white-space: nowrap;
  width: 100%;
  border: none;
  background: transparent;
  text-align: left;
  font-family: inherit;
}
.dropdown-item:hover { background: var(--color-bg); }
.dropdown-item { display: flex; align-items: center; gap: 6px; }

.dropdown-submenu { position: relative; }
.dropdown-submenu-trigger {
  justify-content: space-between;
}
.dropdown-submenu-trigger:focus-visible { outline: 2px solid var(--color-primary); outline-offset: -2px; }
.dropdown-item:focus-visible { outline: 2px solid var(--color-primary); outline-offset: -2px; }
.dropdown-submenu-chevron {
  flex-shrink: 0;
  color: var(--color-text-muted);
  transition: transform 0.15s;
}
.dropdown-submenu-open .dropdown-submenu-chevron,
.dropdown-submenu:focus-within .dropdown-submenu-chevron {
  transform: rotate(-90deg);
}
.dropdown-submenu-panel {
  display: none;
  position: absolute;
  top: 0;
  right: calc(100% - 2px);
  min-width: 280px;
  width: max-content;
  max-width: min(360px, calc(100vw - 24px));
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-md);
  z-index: 210;
}
.dropdown-submenu:hover .dropdown-submenu-panel,
.dropdown-submenu:focus-within .dropdown-submenu-panel,
.dropdown-submenu-open .dropdown-submenu-panel {
  display: block;
}

.msg-unread-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-danger);
  flex-shrink: 0;
  margin-left: auto;
  animation: hdr-pulse 1.4s ease-in-out infinite;
}
@keyframes hdr-pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.75); }
}
.dropdown-item-danger { color: var(--color-danger); }
.dropdown-divider { height: 1px; background: var(--color-border); }

.presence-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-success);
  margin-right: 4px;
}
.presence-count { font-size: 13px; color: var(--color-text-muted); }

.btn-icon {
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 18px;
  padding: 4px;
  border-radius: var(--radius-sm);
  line-height: 1;
}
.btn-icon:hover { background: var(--color-bg); }

.theme-switcher { position: relative; }
.theme-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 140px;
}
.dropdown-item-active {
  color: var(--color-primary);
  font-weight: 600;
}

/* ── Header nav strip ────────────────────────────────────────────────── */
.header-nav {
  display: flex;
  align-items: center;
  gap: 2px;
  margin-right: 8px;
}
@media (max-width: 960px) {
  .header-nav { display: none; }
}

.header-nav-link {
  position: relative;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-muted);
  text-decoration: none;
  white-space: nowrap;
  transition: background 0.15s, color 0.15s;
}
.header-nav-link:hover {
  background: var(--color-bg);
  color: var(--color-text);
}
.header-nav-link.active {
  background: var(--color-primary-subtle, color-mix(in srgb, var(--color-primary) 12%, transparent));
  color: var(--color-primary);
  font-weight: 600;
}
.header-nav-link:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 1px;
}

.nav-unread-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--color-danger);
  flex-shrink: 0;
  animation: hdr-pulse 1.4s ease-in-out infinite;
}
</style>
