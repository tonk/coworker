<template>
  <aside class="app-sidebar" :style="{ width: sidebarWidth + 'px' }">
    <div class="resize-handle" :class="sidebarPos === 'right' ? 'handle-left' : 'handle-right'" @mousedown="startResize"></div>

    <!-- Starred Projects -->
    <section
      class="sidebar-section"
      :style="sectionStyle('starred')"
      :class="{ 'section-drag-over': sectionDragOver === 'starred' }"
      data-section-key="starred"
    >
      <button class="section-header" @click="toggle('starred')" :aria-expanded="open.starred" aria-controls="section-body-starred">
        <span class="section-drag-handle" aria-hidden="true" @pointerdown.prevent.stop="onSectionHandleDown($event, 'starred')">⠿</span>
        <span class="section-title">{{ $t('sidebar.starred') }}</span>
        <span class="chevron" :class="{ open: open.starred }" aria-hidden="true">›</span>
      </button>
      <div v-show="open.starred" class="section-body" id="section-body-starred">
        <div v-if="!orderedStarredProjects.length" class="section-empty">
          {{ $t('sidebar.no_starred') }}
        </div>
        <nav class="sidebar-nav">
          <RouterLink
            v-for="project in orderedStarredProjects"
            :key="project.id"
            :to="`/projects/${project.slug}`"
            class="sidebar-link"
            :class="{ 'drag-over': dragOverId === project.id }"
            draggable="false"
            :data-item-id="project.id"
            data-item-type="project"
            :title="project.customer?.name"
          >
            <span class="drag-handle" @pointerdown.prevent.stop="onItemHandleDown($event, project, 'project')">⠿</span>
            <img v-if="projectAvatar(project) && !avatarErrors.has('p'+project.id)" :src="projectAvatar(project)" class="project-avatar" alt="" @error="avatarErrors.add('p'+project.id)" />
            <span v-else class="project-dot" :style="{ background: project.color || '#6366f1' }"></span>
            <span class="link-text">{{ project.name }}</span>
            <button class="fav-btn fav-btn-active" @click.prevent="sidebarStore.unstarProject(project.slug)" :aria-label="$t('project.unstar')"><span aria-hidden="true">★</span></button>
          </RouterLink>
        </nav>
      </div>
    </section>

    <!-- All Projects -->
    <section
      class="sidebar-section"
      :style="sectionStyle('projects')"
      :class="{ 'section-drag-over': sectionDragOver === 'projects' }"
      data-section-key="projects"
    >
      <button class="section-header" @click="toggle('projects')" :aria-expanded="open.projects" aria-controls="section-body-projects">
        <span class="section-drag-handle" aria-hidden="true" @pointerdown.prevent.stop="onSectionHandleDown($event, 'projects')">⠿</span>
        <span class="section-title">{{ $t('sidebar.all_projects') }}</span>
        <span class="chevron" :class="{ open: open.projects }" aria-hidden="true">›</span>
      </button>
      <div v-show="open.projects" class="section-body indented" id="section-body-projects">
        <div v-if="!sortedProjects.length" class="section-empty">
          {{ $t('sidebar.no_projects') }}
        </div>
        <nav class="sidebar-nav">
          <RouterLink
            v-for="project in sortedProjects"
            :key="project.id"
            :to="`/projects/${project.slug}`"
            class="sidebar-link"
            :title="project.customer?.name"
          >
            <img v-if="projectAvatar(project) && !avatarErrors.has('p'+project.id)" :src="projectAvatar(project)" class="project-avatar" alt="" @error="avatarErrors.add('p'+project.id)" />
            <span v-else class="project-dot" :style="{ background: project.color || '#6366f1' }"></span>
            <span class="link-text">{{ project.name }}</span>
            <span v-if="project.starred" class="star-mark">★</span>
          </RouterLink>
        </nav>
      </div>
    </section>

    <!-- Favorite Customers -->
    <section
      class="sidebar-section"
      :style="sectionStyle('customers')"
      :class="{ 'section-drag-over': sectionDragOver === 'customers' }"
      data-section-key="customers"
    >
      <button class="section-header" @click="toggle('customers')" :aria-expanded="open.customers" aria-controls="section-body-customers">
        <span class="section-drag-handle" aria-hidden="true" @pointerdown.prevent.stop="onSectionHandleDown($event, 'customers')">⠿</span>
        <span class="section-title">{{ $t('sidebar.customers') }}</span>
        <span class="chevron" :class="{ open: open.customers }" aria-hidden="true">›</span>
      </button>
      <div v-show="open.customers" class="section-body" id="section-body-customers">
        <div v-if="!orderedStarredCustomers.length" class="section-empty">
          {{ $t('sidebar.no_customers') }}
        </div>
        <nav class="sidebar-nav">
          <RouterLink
            v-for="c in orderedStarredCustomers"
            :key="c.id"
            :to="`/customers/${c.id}`"
            class="sidebar-link"
            :class="{ 'drag-over': dragOverId === c.id }"
            draggable="false"
            :data-item-id="c.id"
            data-item-type="customer"
          >
            <span class="drag-handle" @pointerdown.prevent.stop="onItemHandleDown($event, c, 'customer')">⠿</span>
            <img v-if="customerAvatar(c) && !avatarErrors.has('c'+c.id)" :src="customerAvatar(c)" class="customer-avatar" alt="" @error="avatarErrors.add('c'+c.id)" />
            <span v-else class="customer-avatar-fallback">{{ customerInitial(c) }}</span>
            <span class="link-text">{{ c.name }}</span>
            <button class="fav-btn fav-btn-active" @click.prevent="customersStore.toggleFavorite(c.id)" :aria-label="$t('customer.unstar')"><span aria-hidden="true">★</span></button>
          </RouterLink>
        </nav>
      </div>
    </section>

    <!-- All Customers -->
    <section
      class="sidebar-section"
      :style="sectionStyle('allCustomers')"
      :class="{ 'section-drag-over': sectionDragOver === 'allCustomers' }"
      data-section-key="allCustomers"
    >
      <button class="section-header" @click="toggle('allCustomers')" :aria-expanded="open.allCustomers" aria-controls="section-body-allcustomers">
        <span class="section-drag-handle" aria-hidden="true" @pointerdown.prevent.stop="onSectionHandleDown($event, 'allCustomers')">⠿</span>
        <span class="section-title">{{ $t('customer.all_customers') }}</span>
        <span class="chevron" :class="{ open: open.allCustomers }" aria-hidden="true">›</span>
      </button>
      <div v-show="open.allCustomers" class="section-body indented" id="section-body-allcustomers">
        <div v-if="!sortedCustomers.length" class="section-empty">
          {{ $t('sidebar.no_customers') }}
        </div>
        <nav class="sidebar-nav">
          <RouterLink
            v-for="c in sortedCustomers"
            :key="c.id"
            :to="`/customers/${c.id}`"
            class="sidebar-link"
          >
            <img v-if="customerAvatar(c) && !avatarErrors.has('c'+c.id)" :src="customerAvatar(c)" class="customer-avatar" alt="" @error="avatarErrors.add('c'+c.id)" />
            <span v-else class="customer-avatar-fallback">{{ customerInitial(c) }}</span>
            <span class="link-text">{{ c.name }}</span>
            <span v-if="c.starred" class="star-mark">★</span>
          </RouterLink>
        </nav>
      </div>
    </section>

    <!-- Favorite People -->
    <section
      class="sidebar-section"
      :style="sectionStyle('favorites')"
      :class="{ 'section-drag-over': sectionDragOver === 'favorites' }"
      data-section-key="favorites"
    >
      <button class="section-header" @click="toggle('favorites')" :aria-expanded="open.favorites" aria-controls="section-body-favorites">
        <span class="section-drag-handle" aria-hidden="true" @pointerdown.prevent.stop="onSectionHandleDown($event, 'favorites')">⠿</span>
        <span class="section-title">{{ $t('sidebar.favorites') }}</span>
        <span class="chevron" :class="{ open: open.favorites }" aria-hidden="true">›</span>
      </button>
      <div v-show="open.favorites" class="section-body" id="section-body-favorites">
        <div v-if="!favoritedUsers.length" class="section-empty">
          {{ $t('sidebar.no_favorites') }}
        </div>
        <div class="user-list">
          <RouterLink
            v-for="user in favoritedUsers"
            :key="user.id"
            :to="{ name: 'chats', query: { user: user.id } }"
            class="user-row"
            :title="user.username">
            <span class="presence-dot" :class="{ online: isOnline(user.id) }" aria-hidden="true"></span>
            <span class="sr-only">{{ isOnline(user.id) ? $t('sidebar.online') : $t('sidebar.offline') }}</span>
            <img v-if="userAvatar(user) && !avatarErrors.has('u'+user.id)" :src="userAvatar(user)" class="user-avatar" alt="" @error="avatarErrors.add('u'+user.id)" />
            <span v-else class="user-avatar-fallback">{{ userInitials(user) }}</span>
            <span class="user-row-name">{{ user.display_name || user.username }}</span>
            <button class="fav-btn fav-btn-active" @click.prevent="unfavorite(user)" :aria-label="$t('sidebar.unfavorite')"><span aria-hidden="true">★</span></button>
          </RouterLink>
        </div>
      </div>
    </section>

    <!-- All People -->
    <section
      class="sidebar-section"
      :style="sectionStyle('people')"
      :class="{ 'section-drag-over': sectionDragOver === 'people' }"
      data-section-key="people"
    >
      <button class="section-header" @click="toggle('people')" :aria-expanded="open.people" aria-controls="section-body-people">
        <span class="section-drag-handle" aria-hidden="true" @pointerdown.prevent.stop="onSectionHandleDown($event, 'people')">⠿</span>
        <span class="section-title">{{ $t('sidebar.users') }}</span>
        <span class="badge-count" v-if="onlineCount" aria-hidden="true">{{ onlineCount }}</span>
        <span class="chevron" :class="{ open: open.people }" aria-hidden="true">›</span>
      </button>
      <div v-show="open.people" class="section-body indented" id="section-body-people">
        <div class="user-list">
          <RouterLink
            v-for="user in sortedUsers"
            :key="user.id"
            :to="{ name: 'chats', query: { user: user.id } }"
            class="user-row"
            :title="user.username">
            <span class="presence-dot" :class="{ online: user.online }" aria-hidden="true"></span>
            <span class="sr-only">{{ user.online ? $t('sidebar.online') : $t('sidebar.offline') }}</span>
            <img v-if="userAvatar(user) && !avatarErrors.has('u'+user.id)" :src="userAvatar(user)" class="user-avatar" alt="" @error="avatarErrors.add('u'+user.id)" />
            <span v-else class="user-avatar-fallback">{{ userInitials(user) }}</span>
            <span class="user-row-name">{{ user.display_name || user.username }}</span>
            <button
              class="fav-btn"
              :class="{ 'fav-btn-active': sidebarStore.isFavorite(user.id) }"
              @click.prevent="toggleFavorite(user)"
              :aria-label="sidebarStore.isFavorite(user.id) ? $t('sidebar.unfavorite') : $t('sidebar.favorite')"
            ><span aria-hidden="true">{{ sidebarStore.isFavorite(user.id) ? '★' : '☆' }}</span></button>
          </RouterLink>
          <div v-if="!sortedUsers.length" class="section-empty">
            {{ $t('sidebar.no_users') }}
          </div>
        </div>
      </div>
    </section>

    <!-- Chats -->
    <section
      class="sidebar-section"
      :style="sectionStyle('chats')"
      :class="{ 'section-drag-over': sectionDragOver === 'chats' }"
      data-section-key="chats"
    >
      <button class="section-header" @click="toggle('chats')" :aria-expanded="open.chats" aria-controls="section-body-chats">
        <span class="section-drag-handle" aria-hidden="true" @pointerdown.prevent.stop="onSectionHandleDown($event, 'chats')">⠿</span>
        <span class="section-title">{{ $t('nav.messages') }}</span>
        <span v-if="notificationsStore.hasUnread" class="unread-dot" aria-hidden="true"></span>
        <span class="sr-only" v-if="notificationsStore.hasUnread">{{ $t('sidebar.unread_messages') }}</span>
        <span class="chevron" :class="{ open: open.chats }" aria-hidden="true">›</span>
      </button>
      <div v-show="open.chats" class="section-body indented" id="section-body-chats">
        <nav class="sidebar-nav">
          <RouterLink
            v-for="conv in recentConversations"
            :key="conv.id"
            :to="convLink(conv)"
            class="sidebar-link conv-link"
          >
            <span class="conv-indicator" :class="{ unread: notificationsStore.isConvUnread(conv) }" aria-hidden="true"></span>
            <span v-if="notificationsStore.isConvUnread(conv)" class="sr-only">{{ $t('sidebar.unread_messages') }}</span>
            <img v-if="conversationAvatar(conv) && !avatarErrors.has('cv'+conv.id)" :src="conversationAvatar(conv)" class="conv-avatar" alt="" @error="avatarErrors.add('cv'+conv.id)" />
            <span v-else class="conv-avatar-fallback">{{ conversationInitials(conv) }}</span>
            <span class="link-text">{{ convSidebarName(conv) }}</span>
          </RouterLink>
          <RouterLink v-if="!recentConversations.length" to="/chats" class="sidebar-link">
            <span class="link-text">{{ $t('dm.no_conversations') }}</span>
          </RouterLink>
        </nav>
        <RouterLink to="/chats" class="sidebar-link sidebar-link-all">{{ $t('sidebar.all_chats') }}</RouterLink>
      </div>
    </section>

  </aside>
</template>

<script setup>
import { ref, computed, reactive, onMounted, onUnmounted } from 'vue'

const avatarErrors = reactive(new Set())

const SIDEBAR_WIDTH_KEY = 'sidebar_width'
const MIN_WIDTH = 150
const MAX_WIDTH = 480
const DEFAULT_WIDTH = 220

const sidebarWidth = ref(
  Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, parseInt(localStorage.getItem(SIDEBAR_WIDTH_KEY) || DEFAULT_WIDTH)))
)

let resizing = false
let startX = 0
let startWidth = 0

function startResize(e) {
  resizing = true
  startX = e.clientX
  startWidth = sidebarWidth.value
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

function onResize(e) {
  if (!resizing) return
  const delta = e.clientX - startX
  const sign = sidebarPos.value === 'right' ? -1 : 1
  sidebarWidth.value = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, startWidth + sign * delta))
}

function stopResize() {
  if (!resizing) return
  resizing = false
  localStorage.setItem(SIDEBAR_WIDTH_KEY, sidebarWidth.value)
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}
import { RouterLink } from 'vue-router'
import { useSidebarStore } from '@/stores/sidebar'
import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'
import { useCustomersStore } from '@/stores/customers'
import { resolveAssetUrl } from '@/api/serverConfig'

const sidebarStore = useSidebarStore()
const auth = useAuthStore()
const notificationsStore = useNotificationsStore()
const customersStore = useCustomersStore()

const sidebarPos = computed(() => auth.user?.sidebar_position || localStorage.getItem('sidebar_position') || 'left')

// Collapse state — persisted in localStorage
const STORAGE_KEY = 'sidebar_open'
const defaults = { starred: true, projects: true, customers: false, allCustomers: false, favorites: true, chats: true, people: true }
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

// ── Section drag-to-reorder (pointer events — works on WebKitGTK/Linux) ───────
const SECTION_ORDER_KEY = 'sidebar_section_order'
const DEFAULT_SECTION_ORDER = ['starred', 'projects', 'customers', 'allCustomers', 'favorites', 'people', 'chats']

const sectionOrder = ref(
  JSON.parse(localStorage.getItem(SECTION_ORDER_KEY) || 'null') || [...DEFAULT_SECTION_ORDER]
)
const sectionDragOver = ref(null)
let _sectionDragKey = null

function sectionStyle(key) {
  return { order: sectionOrder.value.indexOf(key) }
}

function onSectionHandleDown(e, key) {
  _sectionDragKey = key
  sectionDragOver.value = null
  document.addEventListener('pointermove', onSectionPointerMove, { passive: true })
  document.addEventListener('pointerup', onSectionPointerUp)
}

function onSectionPointerMove(e) {
  const el = document.elementFromPoint(e.clientX, e.clientY)
  const sec = el?.closest('[data-section-key]')
  sectionDragOver.value = sec?.dataset.sectionKey || null
}

function onSectionPointerUp() {
  const target = sectionDragOver.value
  sectionDragOver.value = null
  document.removeEventListener('pointermove', onSectionPointerMove)
  document.removeEventListener('pointerup', onSectionPointerUp)
  if (!_sectionDragKey || !target || _sectionDragKey === target) { _sectionDragKey = null; return }
  const order = [...sectionOrder.value]
  const fi = order.indexOf(_sectionDragKey)
  const ti = order.indexOf(target)
  if (fi !== -1 && ti !== -1) { order.splice(fi, 1); order.splice(ti, 0, _sectionDragKey) }
  sectionOrder.value = order
  localStorage.setItem(SECTION_ORDER_KEY, JSON.stringify(order))
  _sectionDragKey = null
}

// ── Item drag-to-reorder for starred sections (pointer events) ────────────────
const STARRED_PROJECTS_ORDER_KEY  = 'sidebar_starred_projects_order'
const STARRED_CUSTOMERS_ORDER_KEY = 'sidebar_starred_customers_order'

const starredProjectOrder  = ref(JSON.parse(localStorage.getItem(STARRED_PROJECTS_ORDER_KEY)  || 'null') || [])
const starredCustomerOrder = ref(JSON.parse(localStorage.getItem(STARRED_CUSTOMERS_ORDER_KEY) || 'null') || [])
const dragOverId = ref(null)
let _dragItem = null
let _dragType = null

function onItemHandleDown(e, item, type) {
  _dragItem = item
  _dragType = type
  dragOverId.value = null
  document.addEventListener('pointermove', onItemPointerMove, { passive: true })
  document.addEventListener('pointerup', onItemPointerUp)
}

function onItemPointerMove(e) {
  const el = document.elementFromPoint(e.clientX, e.clientY)
  const item = el?.closest(`[data-item-id][data-item-type="${_dragType}"]`)
  dragOverId.value = item ? Number(item.dataset.itemId) : null
}

function onItemPointerUp() {
  const targetId = dragOverId.value
  dragOverId.value = null
  document.removeEventListener('pointermove', onItemPointerMove)
  document.removeEventListener('pointerup', onItemPointerUp)
  if (!_dragItem || targetId === null || targetId === undefined || _dragItem.id === targetId) {
    _dragItem = null; _dragType = null; return
  }
  if (_dragType === 'project') {
    const ids = orderedStarredProjects.value.map(p => p.id)
    const fi = ids.indexOf(_dragItem.id), ti = ids.indexOf(targetId)
    if (fi !== -1 && ti !== -1) { ids.splice(fi, 1); ids.splice(ti, 0, _dragItem.id) }
    starredProjectOrder.value = ids
    localStorage.setItem(STARRED_PROJECTS_ORDER_KEY, JSON.stringify(ids))
  } else if (_dragType === 'customer') {
    const ids = orderedStarredCustomers.value.map(c => c.id)
    const fi = ids.indexOf(_dragItem.id), ti = ids.indexOf(targetId)
    if (fi !== -1 && ti !== -1) { ids.splice(fi, 1); ids.splice(ti, 0, _dragItem.id) }
    starredCustomerOrder.value = ids
    localStorage.setItem(STARRED_CUSTOMERS_ORDER_KEY, JSON.stringify(ids))
  }
  _dragItem = null; _dragType = null
}

const orderedStarredProjects = computed(() => {
  const list = sidebarStore.starredProjects
  if (!starredProjectOrder.value.length) return list
  const map = new Map(list.map(p => [p.id, p]))
  const ordered = starredProjectOrder.value.map(id => map.get(id)).filter(Boolean)
  const rest = list.filter(p => !starredProjectOrder.value.includes(p.id))
  return [...ordered, ...rest]
})

const orderedStarredCustomers = computed(() => {
  const list = customersStore.starredCustomers
  if (!starredCustomerOrder.value.length) return list
  const map = new Map(list.map(c => [c.id, c]))
  const ordered = starredCustomerOrder.value.map(id => map.get(id)).filter(Boolean)
  const rest = list.filter(c => !starredCustomerOrder.value.includes(c.id))
  return [...ordered, ...rest]
})

const sortedCustomers = computed(() => {
  const starred = customersStore.customers.filter(c => c.is_favorite).map(c => ({ ...c, starred: true }))
  const rest    = customersStore.customers.filter(c => !c.is_favorite).map(c => ({ ...c, starred: false }))
  return [...starred, ...rest]
})

// ── All projects sorted: starred first (marked), then the rest ────────────────
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

// Chats section — most recently active conversations, capped at 8
const recentConversations = computed(() =>
  [...notificationsStore.conversations]
    .sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at))
    .slice(0, 8)
)

function convSidebarName(conv) {
  if (conv.name) return conv.name
  if (conv.is_group) {
    return conv.members
      ?.filter(m => m.user_id !== auth.user?.id)
      .map(m => m.user?.display_name || m.user?.username)
      .join(', ') || 'Group'
  }
  const other = conv.members?.find(m => m.user_id !== auth.user?.id)
  return other?.user?.display_name || other?.user?.username || 'Chat'
}

function convLink(conv) {
  if (!conv.is_group) {
    const other = conv.members?.find(m => m.user_id !== auth.user?.id)
    if (other) return { name: 'chats', query: { user: other.user_id } }
  }
  return { name: 'chats', query: { conv: conv.id } }
}

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}

function customerAvatar(customer) {
  return resolveAssetUrl(customer?.logo_url || '')
}

function customerInitial(customer) {
  return (customer?.name || '?').slice(0, 1).toUpperCase()
}

function userAvatar(user) {
  return resolveAssetUrl(user?.avatar_url || user?.gravatar_url || '')
}

function userInitials(user) {
  const base = user?.display_name || user?.username || '?'
  return base.slice(0, 2).toUpperCase()
}

function conversationAvatar(conv) {
  if (!conv) return ''
  if (conv.is_group && conv.avatar) return resolveAssetUrl(conv.avatar)
  const other = conv.members?.find(m => m.user_id !== auth.user?.id)?.user
  return userAvatar(other)
}

function conversationInitials(conv) {
  if (!conv) return '?'
  if (conv.name) return conv.name.slice(0, 2).toUpperCase()
  if (conv.is_group) return 'GR'
  const other = conv.members?.find(m => m.user_id !== auth.user?.id)?.user
  return userInitials(other)
}

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

const ONLINE_POLL_INTERVAL_MS = 5_000
let pollInterval = null
let unreadInterval = null

onMounted(() => {
  sidebarStore.fetchStarred()
  sidebarStore.fetchAllProjects()
  sidebarStore.fetchAllUsers()
  sidebarStore.fetchChatUsers()
  sidebarStore.fetchFavoriteUsers()
  customersStore.fetchCustomers()
  notificationsStore.checkUnread()
  pollInterval = setInterval(() => {
    sidebarStore.fetchAllUsers()
    sidebarStore.fetchChatUsers()
  }, ONLINE_POLL_INTERVAL_MS)
  unreadInterval = setInterval(() => {
    notificationsStore.checkUnread()
  }, 5_000)
})

onUnmounted(() => {
  clearInterval(pollInterval)
  clearInterval(unreadInterval)
  stopResize()
  // Clean up any in-progress pointer drags
  document.removeEventListener('pointermove', onSectionPointerMove)
  document.removeEventListener('pointerup', onSectionPointerUp)
  document.removeEventListener('pointermove', onItemPointerMove)
  document.removeEventListener('pointerup', onItemPointerUp)
})
</script>

<style scoped>
.app-sidebar {
  flex-shrink: 0;
  position: relative;
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  padding: 12px 0;
}

.resize-handle {
  position: absolute;
  top: 0;
  width: 6px;
  height: 100%;
  cursor: col-resize;
  z-index: 10;
}
.resize-handle.handle-right { right: -3px; }
.resize-handle.handle-left  { left: -3px; }
.resize-handle:hover,
.resize-handle:active {
  background: var(--color-primary);
  opacity: 0.4;
}

.sidebar-section {
  margin-bottom: 12px;
}

.sidebar-section.section-drag-over {
  background: color-mix(in srgb, var(--color-primary) 8%, transparent);
  border-radius: 4px;
  outline: 1px dashed color-mix(in srgb, var(--color-primary) 40%, transparent);
}

.section-header {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 6px 12px 6px 8px;
  background: transparent;
  border: none;
  cursor: pointer;
  gap: 6px;
  text-align: left;
}
.section-header:hover { background: var(--color-bg); }

.section-drag-handle {
  flex-shrink: 0;
  color: var(--color-text-muted);
  font-size: 12px;
  opacity: 0;
  cursor: grab;
  touch-action: none;
  transition: opacity .1s;
  padding: 0 2px;
}
.section-drag-handle:active { cursor: grabbing; }
.sidebar-section:hover .section-drag-handle { opacity: 1; }
.sidebar-section:focus-within .section-drag-handle { opacity: 1; }

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
.section-body.indented .sidebar-link { padding-left: 36px; }
.section-body.indented .user-row     { padding-left: 36px; }

.section-empty {
  padding: 4px 16px 4px 36px;
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
.project-avatar {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  object-fit: cover;
  border: 1px solid var(--color-border);
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

.customer-avatar {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  object-fit: cover;
  border: 1px solid var(--color-border);
  flex-shrink: 0;
}
.customer-avatar-fallback {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  background: var(--color-border);
  color: var(--color-text-muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  font-weight: 700;
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
.user-avatar {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--color-border);
  flex-shrink: 0;
}
.user-avatar-fallback {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--color-border);
  color: var(--color-text-muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  font-weight: 700;
  flex-shrink: 0;
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
.fav-btn:focus-visible { opacity: 1; }
.fav-btn.fav-btn-active {
  color: var(--color-warning);
  opacity: 1;
}
.fav-btn:hover { color: var(--color-warning); }

/* Chats section */
.conv-link { gap: 8px; }

.conv-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-border);
  flex-shrink: 0;
}
.conv-indicator.unread {
  background: var(--color-danger, #ef4444);
  animation: pulse 1.4s ease-in-out infinite;
}
.conv-avatar {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--color-border);
  flex-shrink: 0;
}
.conv-avatar-fallback {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--color-border);
  color: var(--color-text-muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 8px;
  font-weight: 700;
  flex-shrink: 0;
}

.sidebar-link-all {
  font-size: 11px;
  color: var(--color-text-muted);
  padding-top: 4px;
  padding-bottom: 4px;
  padding-left: 36px;
  border-top: 1px solid var(--color-border);
  margin-top: 2px;
}

/* Item drag-to-reorder */
.drag-handle {
  flex-shrink: 0;
  color: var(--color-text-muted);
  font-size: 12px;
  opacity: 0;
  cursor: grab;
  touch-action: none;
  transition: opacity .1s;
}
.drag-handle:active { cursor: grabbing; }
.sidebar-link:hover .drag-handle { opacity: 1; }
.sidebar-link:focus-within .drag-handle { opacity: 1; }
.sidebar-link.drag-over {
  background: color-mix(in srgb, var(--color-primary) 15%, transparent);
  border-radius: 4px;
}
</style>
