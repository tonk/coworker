<template>
  <header class="app-header">
    <div class="header-left">
      <RouterLink to="/" class="logo">
        <img src="/logo-full.svg" alt="WarmDesk" class="logo-img" />
      </RouterLink>
      <slot name="breadcrumb" />
    </div>
    <div class="header-center">
      <GlobalSearch />
    </div>
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
        <div class="dropdown" v-if="menuOpen" role="menu">
          <div class="dropdown-item" role="menuitem" @click="navigate('/')">{{ $t('nav.dashboard') }}</div>
          <div class="dropdown-item" role="menuitem" @click="navigate('/news')">{{ $t('nav.news') }}</div>
          <div class="dropdown-item" role="menuitem" @click="navigate('/settings')">{{ $t('nav.settings') }}</div>
          <div class="dropdown-item" role="menuitem" v-if="auth.isAdmin" @click="navigate('/admin')">{{ $t('nav.admin') }}</div>
          <div class="dropdown-divider" role="separator"></div>
          <div class="dropdown-item" role="menuitem" @click="navigate('/chats')">
            {{ $t('nav.messages') }}
            <span v-if="notificationsStore.hasUnread" class="msg-unread-dot" aria-hidden="true"></span>
            <span v-if="notificationsStore.hasUnread" class="sr-only">({{ $t('sidebar.unread_messages') }})</span>
          </div>
          <div class="dropdown-item" role="menuitem" v-if="auth.canViewReports" @click="navigate('/reports')">{{ $t('report.nav') }}</div>
          <div class="dropdown-item" role="menuitem" v-if="auth.timeTrackingEnabled" @click="navigate('/time-tracking')">{{ $t('timeTracking.nav') }}</div>
          <div class="dropdown-divider" role="separator"></div>
          <div class="dropdown-item" role="menuitem" @click="openShortcuts">{{ $t('nav.keyboard_shortcuts') }}</div>
          <div class="dropdown-item" role="menuitem" @click="openAbout">{{ $t('nav.about') }}</div>
          <div class="dropdown-item dropdown-item-danger" role="menuitem" @click="handleLogout">{{ $t('nav.logout') }}</div>
        </div>
      </div>
    </div>
  </header>
  <AboutModal v-if="showAbout" @close="showAbout = false" />
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { setLocale } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { useNotificationsStore } from '@/stores/notifications'
import { avatarUrl } from '@/composables/useAvatar'
import GlobalSearch from '@/components/common/GlobalSearch.vue'
import AboutModal from '@/components/common/AboutModal.vue'

const props = defineProps({ presenceCount: { type: Number, default: 0 } })
const emit = defineEmits(['open-shortcuts'])

const auth = useAuthStore()
const router = useRouter()
const { locale } = useI18n()
const { theme, setTheme } = useTheme()
const notificationsStore = useNotificationsStore()
const menuOpen = ref(false)
const themeOpen = ref(false)
const showAbout = ref(false)
const menuRef = ref(null)
const themeRef = ref(null)
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

function onLocaleChange(e) {
  const lang = e.target.value
  setLocale(lang)
  document.documentElement.lang = lang
  if (auth.isLoggedIn) auth.updateProfile({ locale: lang })
}

function navigate(path) {
  menuOpen.value = false
  router.push(path)
}

function openShortcuts() {
  menuOpen.value = false
  emit('open-shortcuts')
}

function openAbout() {
  menuOpen.value = false
  showAbout.value = true
}

function handleLogout() {
  menuOpen.value = false
  auth.logout()
  router.push('/login')
}

function handleClick(e) {
  if (menuRef.value && !menuRef.value.contains(e.target)) menuOpen.value = false
  if (themeRef.value && !themeRef.value.contains(e.target)) themeOpen.value = false
}

function handleKeyDown(e) {
  if (e.key === 'Escape') {
    menuOpen.value = false
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
  min-width: 160px;
  z-index: 200;
}

.dropdown-item {
  padding: 10px 16px;
  cursor: pointer;
  font-size: 14px;
  color: var(--color-text);
}
.dropdown-item:hover { background: var(--color-bg); }
.dropdown-item { display: flex; align-items: center; gap: 6px; }

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
</style>
