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

    <!-- All Users -->
    <section class="sidebar-section">
      <button class="section-header" @click="toggle('chat')">
        <span class="section-title">{{ $t('sidebar.users') }}</span>
        <span class="badge-count" v-if="sidebarStore.chatUsers.length">{{ sidebarStore.chatUsers.length }}</span>
        <span v-if="notificationsStore.hasUnread" class="unread-dot" title="Unread messages"></span>
        <span class="chevron" :class="{ open: open.chat }">›</span>
      </button>
      <div v-show="open.chat" class="section-body">
        <div class="user-list">
          <RouterLink
            v-for="user in sortedUsers"
            :key="user.id"
            :to="{ name: 'messages', query: { user: user.id } }"
            class="online-user"
          >
            <span class="user-dot" :class="{ 'in-chat': user.inChat }"></span>
            <span class="online-name">{{ user.display_name || user.username }}</span>
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
const defaults = { starred: true, projects: true, chat: true }
const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || 'null') || defaults
const open = ref({ ...defaults, ...saved })

function toggle(section) {
  open.value[section] = !open.value[section]
  localStorage.setItem(STORAGE_KEY, JSON.stringify(open.value))
}

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

// All users sorted: in-chat users first (marked), then the rest
const sortedUsers = computed(() => {
  const chatSet = new Set(sidebarStore.chatUsers.map(u => u.id))
  const inChat = sidebarStore.allUsers
    .filter(u => chatSet.has(u.id) && u.id !== auth.user?.id)
    .map(u => ({ ...u, inChat: true }))
  const notInChat = sidebarStore.allUsers
    .filter(u => !chatSet.has(u.id) && u.id !== auth.user?.id)
    .map(u => ({ ...u, inChat: false }))
  return [...inChat, ...notInChat]
})

let pollInterval = null

onMounted(() => {
  sidebarStore.fetchStarred()
  sidebarStore.fetchAllProjects()
  sidebarStore.fetchAllUsers()
  sidebarStore.fetchChatUsers()
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

.user-list { display: flex; flex-direction: column; }

.online-user {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 16px;
  font-size: 13px;
  text-decoration: none;
  color: var(--color-text);
  cursor: pointer;
}
.online-user:hover { background: var(--color-bg); }

.user-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-border);
  flex-shrink: 0;
}
.user-dot.in-chat {
  background: var(--color-success);
}

.online-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text);
}
</style>
