<template>
  <aside class="app-sidebar">

    <!-- Starred Projects -->
    <section class="sidebar-section">
      <button class="section-header" @click="toggle('starred')">
        <span class="section-title">{{ $t('sidebar.starred') }}</span>
        <span class="chevron" :class="{ open: open.starred }">›</span>
      </button>
      <div v-show="open.starred" class="section-body">
        <div v-if="!sidebarStore.starredProjects.length" class="section-empty">
          {{ $t('sidebar.no_starred') }}
        </div>
        <nav class="sidebar-nav">
          <RouterLink
            v-for="project in sidebarStore.starredProjects"
            :key="project.id"
            :to="`/projects/${project.slug}`"
            class="sidebar-link"
          >
            <span class="project-dot" :style="{ background: project.color || '#6366f1' }"></span>
            <span class="link-text">{{ project.name }}</span>
          </RouterLink>
        </nav>
      </div>
    </section>

    <!-- All Projects -->
    <section class="sidebar-section">
      <button class="section-header" @click="toggle('projects')">
        <span class="section-title">{{ $t('sidebar.all_projects') }}</span>
        <span class="chevron" :class="{ open: open.projects }">›</span>
      </button>
      <div v-show="open.projects" class="section-body">
        <div v-if="!sortedProjects.length" class="section-empty">
          {{ $t('sidebar.no_projects') }}
        </div>
        <nav class="sidebar-nav">
          <RouterLink
            v-for="project in sortedProjects"
            :key="project.id"
            :to="`/projects/${project.slug}`"
            class="sidebar-link"
          >
            <span class="project-dot" :style="{ background: project.color || '#6366f1' }"></span>
            <span class="link-text">{{ project.name }}</span>
            <span v-if="project.starred" class="star-mark">★</span>
          </RouterLink>
        </nav>
      </div>
    </section>

    <!-- Favorite People -->
    <section class="sidebar-section">
      <button class="section-header" @click="toggle('favorites')">
        <span class="section-title">{{ $t('sidebar.favorites') }}</span>
        <span class="chevron" :class="{ open: open.favorites }">›</span>
      </button>
      <div v-show="open.favorites" class="section-body">
        <div v-if="!favoritedUsers.length" class="section-empty">
          {{ $t('sidebar.no_favorites') }}
        </div>
        <div class="user-list">
          <RouterLink
            v-for="user in favoritedUsers"
            :key="user.id"
            :to="{ name: 'messages', query: { user: user.id } }"
            class="user-row"
          >
            <span class="presence-dot" :class="{ online: isOnline(user.id) }" :title="isOnline(user.id) ? $t('sidebar.online') : $t('sidebar.offline')"></span>
            <span class="user-row-name">{{ user.display_name || user.username }}</span>
            <button class="fav-btn fav-btn-active" @click.prevent="unfavorite(user)" :title="$t('sidebar.unfavorite')">★</button>
          </RouterLink>
        </div>
      </div>
    </section>

    <!-- All People -->
    <section class="sidebar-section">
      <button class="section-header" @click="toggle('people')">
        <span class="section-title">{{ $t('sidebar.users') }}</span>
        <span class="badge-count" v-if="onlineCount">{{ onlineCount }}</span>
        <span v-if="notificationsStore.hasUnread" class="unread-dot" title="Unread messages"></span>
        <span class="chevron" :class="{ open: open.people }">›</span>
      </button>
      <div v-show="open.people" class="section-body">
        <div class="user-list">
          <RouterLink
            v-for="user in sortedUsers"
            :key="user.id"
            :to="{ name: 'messages', query: { user: user.id } }"
            class="user-row"
          >
            <span class="presence-dot" :class="{ online: user.online }" :title="user.online ? $t('sidebar.online') : $t('sidebar.offline')"></span>
            <span class="user-row-name">{{ user.display_name || user.username }}</span>
            <button
              class="fav-btn"
              :class="{ 'fav-btn-active': sidebarStore.isFavorite(user.id) }"
              @click.prevent="toggleFavorite(user)"
              :title="sidebarStore.isFavorite(user.id) ? $t('sidebar.unfavorite') : $t('sidebar.favorite')"
            >{{ sidebarStore.isFavorite(user.id) ? '★' : '☆' }}</button>
          </RouterLink>
          <div v-if="!sortedUsers.length" class="section-empty">
            {{ $t('sidebar.no_users') }}
          </div>
        </div>
      </div>
    </section>

  </aside>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useSidebarStore } from '@/stores/sidebar'
import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'

const sidebarStore = useSidebarStore()
const auth = useAuthStore()
const notificationsStore = useNotificationsStore()

// Collapse state — persisted in localStorage
const STORAGE_KEY = 'sidebar_open'
const defaults = { starred: true, projects: true, favorites: true, people: true }
const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || 'null') || defaults
const open = ref({ ...defaults, ...saved })

function toggle(section) {
  open.value[section] = !open.value[section]
  localStorage.setItem(STORAGE_KEY, JSON.stringify(open.value))
}

const onlineIds = computed(() => new Set(sidebarStore.chatUsers.map(u => u.id)))

function isOnline(userId) {
  return onlineIds.value.has(userId)
}

const onlineCount = computed(() => {
  return sidebarStore.allUsers.filter(u => u.id !== auth.user?.id && isOnline(u.id)).length
})

// All projects sorted: starred first (marked), then the rest
const sortedProjects = computed(() => {
  const starredSet = new Set(sidebarStore.starredProjects.map(p => p.id))
  const starred = sidebarStore.allProjects
    .filter(p => starredSet.has(p.id))
    .map(p => ({ ...p, starred: true }))
  const rest = sidebarStore.allProjects
    .filter(p => !starredSet.has(p.id))
    .map(p => ({ ...p, starred: false }))
  return [...starred, ...rest]
})

// Favorites section — only favorited users, enriched with online status
const favoritedUsers = computed(() => {
  const favIds = new Set(sidebarStore.favoriteUsers.map(u => u.id))
  return sidebarStore.allUsers
    .filter(u => favIds.has(u.id))
    .map(u => ({ ...u, online: isOnline(u.id) }))
})

// All users: online first, then offline; exclude self
const sortedUsers = computed(() => {
  const others = sidebarStore.allUsers.filter(u => u.id !== auth.user?.id)
  const online = others.filter(u => isOnline(u.id)).map(u => ({ ...u, online: true }))
  const offline = others.filter(u => !isOnline(u.id)).map(u => ({ ...u, online: false }))
  return [...online, ...offline]
})

async function toggleFavorite(user) {
  if (sidebarStore.isFavorite(user.id)) {
    await sidebarStore.removeFavoriteUser(user.id)
  } else {
    await sidebarStore.addFavoriteUser(user.id)
  }
}

async function unfavorite(user) {
  await sidebarStore.removeFavoriteUser(user.id)
}

let pollInterval = null

onMounted(() => {
  sidebarStore.fetchStarred()
  sidebarStore.fetchAllProjects()
  sidebarStore.fetchAllUsers()
  sidebarStore.fetchChatUsers()
  sidebarStore.fetchFavoriteUsers()
  notificationsStore.checkUnread()
  pollInterval = setInterval(() => {
    sidebarStore.fetchAllUsers()
    sidebarStore.fetchChatUsers()
    notificationsStore.checkUnread()
  }, 30_000)
})

onUnmounted(() => {
  clearInterval(pollInterval)
})
</script>

<style scoped>
.app-sidebar {
  width: 220px;
  flex-shrink: 0;
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  padding: 12px 0;
}

.sidebar-section {
  margin-bottom: 4px;
}

.section-header {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 6px 12px 6px 16px;
  background: transparent;
  border: none;
  cursor: pointer;
  gap: 6px;
  text-align: left;
}
.section-header:hover { background: var(--color-bg); }

.section-title {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .05em;
  color: var(--color-text-muted);
  flex: 1;
}

.chevron {
  font-size: 14px;
  color: var(--color-text-muted);
  line-height: 1;
  transform: rotate(90deg);
  transition: transform .15s;
  display: inline-block;
}
.chevron.open {
  transform: rotate(-90deg);
}

.unread-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-danger, #ef4444);
  flex-shrink: 0;
  animation: pulse 1.4s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.75); }
}

.badge-count {
  font-size: 11px;
  font-weight: 600;
  background: var(--color-success);
  color: #fff;
  border-radius: 9999px;
  padding: 0 5px;
  line-height: 16px;
}

.section-body { }

.section-empty {
  padding: 4px 16px;
  font-size: 12px;
  color: var(--color-text-muted);
}

.sidebar-nav { display: flex; flex-direction: column; }

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  font-size: 13px;
  color: var(--color-text);
  text-decoration: none;
  transition: background .1s;
}
.sidebar-link:hover { background: var(--color-bg); text-decoration: none; }
.sidebar-link.router-link-active { background: var(--color-bg); color: var(--color-primary); font-weight: 600; }

.project-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.link-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.star-mark {
  font-size: 11px;
  color: var(--color-warning, #f59e0b);
  flex-shrink: 0;
}

/* User rows (favorites + people) */
.user-list { display: flex; flex-direction: column; }

.user-row {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 5px 10px 5px 16px;
  font-size: 13px;
  text-decoration: none;
  color: var(--color-text);
  cursor: pointer;
}
.user-row:hover { background: var(--color-bg); }
.user-row.router-link-active { background: var(--color-bg); color: var(--color-primary); }

.presence-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-border);
  flex-shrink: 0;
  transition: background .2s;
}
.presence-dot.online {
  background: var(--color-success);
}

.user-row-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.fav-btn {
  flex-shrink: 0;
  background: none;
  border: none;
  padding: 0 2px;
  font-size: 13px;
  color: var(--color-text-muted);
  cursor: pointer;
  opacity: 0;
  line-height: 1;
  transition: opacity .1s, color .1s;
}
.user-row:hover .fav-btn { opacity: 1; }
.fav-btn.fav-btn-active {
  color: #f59e0b;
  opacity: 1;
}
.fav-btn:hover { color: #f59e0b; }
</style>
