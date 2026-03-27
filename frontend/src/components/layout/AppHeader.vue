<template>
  <header class="app-header">
    <div class="header-left">
      <RouterLink to="/" class="logo">
        <img src="/logo.svg" alt="Coworker" class="logo-img" />
      </RouterLink>
      <slot name="breadcrumb" />
    </div>
    <div class="header-right">
      <span class="presence-count" v-if="presenceCount > 0" :title="`${presenceCount} ${$t('presence.online')}`">
        <span class="presence-dot"></span>{{ presenceCount }}
      </span>
      <div class="lang-switcher">
        <select class="form-input lang-select" :value="locale" @change="onLocaleChange">
          <option value="en">EN</option>
          <option value="nl">NL</option>
          <option value="de">DE</option>
          <option value="es">ES</option>
          <option value="fr">FR</option>
        </select>
      </div>
      <button class="btn-icon" @click="cycleTheme" :title="$t('settings.theme')">
        <span v-if="theme === 'light'">☀️</span>
        <span v-else-if="theme === 'dark'">🌙</span>
        <span v-else>🖥</span>
      </button>
      <div class="user-menu" v-if="auth.user" @click="menuOpen = !menuOpen" ref="menuRef">
        <div class="avatar">
          <img v-if="userAvatar" :src="userAvatar" :alt="initials" class="avatar-img" @error="avatarErr = true" />
          <span v-else>{{ initials }}</span>
        </div>
        <div class="dropdown" v-if="menuOpen">
          <div class="dropdown-item" @click="router.push('/')">{{ $t('nav.dashboard') }}</div>
          <div class="dropdown-item" @click="router.push('/settings')">{{ $t('nav.settings') }}</div>
          <div class="dropdown-item" v-if="auth.isAdmin" @click="router.push('/admin')">{{ $t('nav.admin') }}</div>
          <div class="dropdown-divider"></div>
          <div class="dropdown-item" @click="router.push('/messages')">
            {{ $t('nav.messages') }}
            <span v-if="notificationsStore.hasUnread" class="msg-unread-dot"></span>
          </div>
          <div class="dropdown-divider"></div>
          <div class="dropdown-item dropdown-item-danger" @click="handleLogout">{{ $t('nav.logout') }}</div>
        </div>
      </div>
    </div>
  </header>
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

const props = defineProps({ presenceCount: { type: Number, default: 0 } })

const auth = useAuthStore()
const router = useRouter()
const { locale } = useI18n()
const { theme, setTheme } = useTheme()
const notificationsStore = useNotificationsStore()
const menuOpen = ref(false)
const menuRef = ref(null)
const avatarErr = ref(false)

const userAvatar = computed(() => avatarErr.value ? null : avatarUrl(auth.user))

const themes = ['light', 'dark', 'system']
function cycleTheme() {
  const idx = themes.indexOf(theme.value)
  setTheme(themes[(idx + 1) % themes.length])
  if (auth.isLoggedIn) auth.updateProfile({ theme: theme.value })
}

const initials = computed(() => {
  const name = auth.user?.display_name || auth.user?.username || '?'
  return name.slice(0, 2).toUpperCase()
})

function onLocaleChange(e) {
  setLocale(e.target.value)
  if (auth.isLoggedIn) auth.updateProfile({ locale: e.target.value })
}

function handleLogout() {
  auth.logout()
  router.push('/login')
}

function handleClick(e) {
  if (menuRef.value && !menuRef.value.contains(e.target)) menuOpen.value = false
}

onMounted(() => document.addEventListener('click', handleClick))
onBeforeUnmount(() => document.removeEventListener('click', handleClick))
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
.header-right { display: flex; align-items: center; gap: 12px; }

.logo { text-decoration: none; display: flex; align-items: center; }
.logo-img { height: 28px; width: auto; display: block; }

.lang-select { width: 60px; padding: 4px 6px; font-size: 12px; }

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
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
  background: var(--color-danger, #ef4444);
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
</style>
