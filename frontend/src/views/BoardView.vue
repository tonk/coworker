<template>
  <div class="board-layout">
    <div class="board-toolbar">
      <div class="board-toolbar-left">
        <img v-if="projectAvatar(projectStore.currentProject)" :src="projectAvatar(projectStore.currentProject)" class="board-project-avatar" alt="" />
        <h1 class="board-project-name">{{ projectStore.currentProject?.name }}</h1>
        <button class="btn btn-ghost btn-sm star-btn" :class="{ starred: sidebarStore.isStarred(slug) }" @click="toggleStar" :title="sidebarStore.isStarred(slug) ? $t('board.unstar') : $t('board.star')">
          {{ sidebarStore.isStarred(slug) ? '★' : '☆' }}
        </button>
        <RouterLink v-if="canManageColumns" :to="`/projects/${slug}/settings`" class="btn btn-ghost btn-sm settings-link" :title="$t('project.settings')">
          ⚙
        </RouterLink>
      </div>
      <div class="board-toolbar-right">
        <template v-if="projectStore.currentProject?.board_type === 'scrum'">
          <RouterLink :to="`/projects/${slug}/backlog`" class="btn btn-ghost btn-sm">
            📋 {{ $t('sprint.backlog') }}
          </RouterLink>
          <RouterLink :to="`/projects/${slug}/sprint`" class="btn btn-ghost btn-sm">
            🏃 {{ $t('sprint.board') }}
          </RouterLink>
        </template>
        <RouterLink :to="`/projects/${slug}/charts`" class="btn btn-ghost btn-sm">
          📊 {{ $t('sprint.charts') }}
        </RouterLink>
        <RouterLink :to="`/projects/${slug}/gantt`" class="btn btn-ghost btn-sm">
          📅 {{ $t('gantt.title') }}
        </RouterLink>
        <RouterLink :to="`/projects/${slug}/topics`" class="btn btn-ghost btn-sm">
          💬 {{ $t('topics.title') }}
        </RouterLink>
        <button v-if="canManageColumns" class="btn btn-secondary btn-sm" @click="showAddColumn = true">
          + {{ $t('board.add_column') }}
        </button>
      </div>
    </div>

    <div class="board-body">
      <div class="board-columns-wrap">
        <div v-if="boardStore.loading" class="board-loading">
          <div class="spinner" style="width:40px;height:40px;border-width:3px"></div>
        </div>

        <div v-else class="board-columns" ref="columnsEl">
          <BoardColumn
            v-for="col in boardStore.columns"
            :key="col.id"
            :column="col"
            :data-column-id="col.id"
            :can-manage-columns="canManageColumns"
            @add-card="openAddCard"
            @open-card="openCardDetail"
            @card-moved="onCardMoved"
            @rename-column="onRenameColumn"
            @delete-column="onDeleteColumn"
            @edit-column="openEditColumn"
          />
        </div>
      </div>
    </div>

    <!-- Add column modal -->
    <BaseModal v-if="showAddColumn" :title="$t('board.add_column')" @close="showAddColumn = false">
      <form @submit.prevent="submitAddColumn">
        <div class="form-group">
          <label class="form-label">{{ $t('board.column_name') }}</label>
          <input class="form-input" v-model="newColumn.name" required autofocus />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('project.color') }}</label>
          <input type="color" class="form-input" v-model="newColumn.color" style="height:40px;padding:4px" />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('board.wip_limit') }}</label>
          <input class="form-input" type="text" inputmode="numeric" v-model="newColumn.wip" autocomplete="off" />
          <p class="form-hint">{{ $t('board.wip_limit_hint') }}</p>
        </div>
      </form>
      <template #footer>
        <button class="btn btn-secondary" @click="showAddColumn = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="submitAddColumn">{{ $t('common.create') }}</button>
      </template>
    </BaseModal>

    <!-- Edit column modal -->
    <BaseModal v-if="showEditColumn" :title="$t('board.edit_column')" @close="showEditColumn = false">
      <form @submit.prevent="submitEditColumn">
        <div class="form-group">
          <label class="form-label">{{ $t('board.column_name') }}</label>
          <input class="form-input" v-model="editColumn.name" required autofocus />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('project.color') }}</label>
          <input type="color" class="form-input" v-model="editColumn.color" style="height:40px;padding:4px" />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('board.wip_limit') }}</label>
          <input class="form-input" type="text" inputmode="numeric" v-model="editColumn.wip" autocomplete="off" />
          <p class="form-hint">{{ $t('board.wip_limit_hint') }}</p>
        </div>
      </form>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="showEditColumn = false">{{ $t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" @click="submitEditColumn">{{ $t('common.save') }}</button>
      </template>
    </BaseModal>

    <!-- Card detail -->
    <CardDetail
      v-if="selectedCard"
      :card="selectedCard"
      :labels="projectStore.currentProject?.labels || []"
      :members="projectMembers"
      :project-slug="slug"
      @close="selectedCard = null"
      @deleted="selectedCard = null"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Sortable from 'sortablejs'
import BoardColumn from '@/components/board/BoardColumn.vue'
import CardDetail from '@/components/board/CardDetail.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import { useBoardStore } from '@/stores/board'
import { useProjectStore } from '@/stores/project'
import { useUIStore } from '@/stores/ui'
import { useSidebarStore } from '@/stores/sidebar'
import { useAuthStore } from '@/stores/auth'
import { useWebSocket } from '@/composables/useWebSocket'
import { projectsApi } from '@/api/projects'
import { resolveAssetUrl } from '@/api/serverConfig'

const route = useRoute()
const { t } = useI18n()
const slug = computed(() => route.params.slug)

const boardStore = useBoardStore()
const projectStore = useProjectStore()
const ui = useUIStore()
const sidebarStore = useSidebarStore()
const auth = useAuthStore()

const showAddColumn = ref(false)
const showEditColumn = ref(false)
const editingColumnId = ref(null)
const selectedCard = ref(null)
const newColumn = ref({ name: '', color: '#94a3b8', wip: '' })
const editColumn = ref({ name: '', color: '#94a3b8', wip: '' })
const columnsEl = ref(null)
let columnSortable = null

const projectMembers = ref([])

const { connected, presenceUsers, connect, disconnect, send: wsSend } = useWebSocket(slug)

// Track whether we've had the first successful connection for the current board.
// On subsequent connections (reconnects), silently refresh to catch missed events.
let wsConnectedOnce = false

watch(connected, (isConnected) => {
  if (isConnected) {
    if (wsConnectedOnce) boardStore.silentRefresh(slug.value)
    wsConnectedOnce = true
  }
})

async function loadBoard(projectSlug) {
  wsConnectedOnce = false
  selectedCard.value = null
  await Promise.all([
    boardStore.loadBoard(projectSlug),
    projectStore.fetchProject(projectSlug),
    sidebarStore.fetchStarred()
  ])
  loadMembers()
  connect()
  await nextTick()
  initColumnSortable()
  if (route.query.card) openCardById(Number(route.query.card))
}

async function openCardById(cardId) {
  try {
    const { data } = await projectsApi.getCard(slug.value, cardId)
    selectedCard.value = data
  } catch {}
}

onMounted(() => loadBoard(slug.value))

watch(slug, async (newSlug) => {
  disconnect()
  boardStore.reset()
  columnSortable?.destroy()
  columnSortable = null
  await loadBoard(newSlug)
})

onUnmounted(() => {
  disconnect()
  boardStore.reset()
  columnSortable?.destroy()
})

async function loadMembers() {
  try {
    const { data } = await projectsApi.listMembers(slug.value)
    projectMembers.value = data
  } catch {}
}

// Global admins and project admins/owners can manage columns
const ADMIN_RANKS = { admin: 3, owner: 4 }
const canManageColumns = computed(() => {
  if (auth.user?.global_role === 'admin') return true
  const me = projectMembers.value.find(m => m.user_id === auth.user?.id)
  return me ? (ADMIN_RANKS[me.role] ?? 0) >= 3 : false
})

watch(canManageColumns, (val) => {
  columnSortable?.option('disabled', !val)
})

async function toggleStar() {
  if (!slug.value) return
  if (sidebarStore.isStarred(slug.value)) {
    await sidebarStore.unstarProject(slug.value)
  } else {
    await sidebarStore.starProject(slug.value)
  }
}

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}

function openAddCard(columnId) {
  selectedCard.value = { id: null, column_id: columnId, title: '' }
}

async function submitAddColumn() {
  if (!newColumn.value.name.trim()) return
  try {
    const extra = { color: newColumn.value.color }
    const raw = String(newColumn.value.wip ?? '').trim()
    if (raw !== '') {
      const n = parseInt(raw, 10)
      if (!Number.isFinite(n) || n < 1) {
        ui.error(t('board.wip_limit_invalid'))
        return
      }
      extra.wip_limit = n
    }
    await boardStore.createColumn(newColumn.value.name.trim(), extra)
    showAddColumn.value = false
    newColumn.value = { name: '', color: '#94a3b8', wip: '' }
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to create column')
  }
}

function openEditColumn(column) {
  editingColumnId.value = column.id
  editColumn.value = {
    name: column.name,
    color: column.color || '#94a3b8',
    wip: column.wip_limit != null && column.wip_limit !== undefined ? String(column.wip_limit) : ''
  }
  showEditColumn.value = true
}

async function submitEditColumn() {
  const name = editColumn.value.name.trim()
  if (!name || editingColumnId.value == null) return
  const payload = { name, color: editColumn.value.color || '#94a3b8' }
  const raw = String(editColumn.value.wip ?? '').trim()
  if (raw === '') {
    payload.wip_limit_clear = true
  } else {
    const n = parseInt(raw, 10)
    if (!Number.isFinite(n) || n < 1) {
      ui.error(t('board.wip_limit_invalid'))
      return
    }
    payload.wip_limit = n
  }
  try {
    const { data } = await projectsApi.updateColumn(slug.value, editingColumnId.value, payload)
    boardStore.updateColumn(data)
    showEditColumn.value = false
    editingColumnId.value = null
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to update column')
  }
}

async function openCardDetail(card) {
  // Fetch full card with comments
  try {
    const { data } = await projectsApi.getCard(slug.value, card.id)
    selectedCard.value = data
  } catch {
    selectedCard.value = card
  }
}

function initColumnSortable() {
  if (!columnsEl.value) return
  columnSortable = Sortable.create(columnsEl.value, {
    animation: 150,
    handle: '.column-header',
    ghostClass: 'column-ghost',
    dragClass: 'column-drag',
    disabled: !canManageColumns.value,
    onEnd(evt) {
      if (evt.oldIndex === evt.newIndex) return
      onColumnReordered(evt.oldIndex, evt.newIndex)
    }
  })
}

async function onColumnReordered(oldIndex, newIndex) {
  const cols = [...boardStore.columns]
  const [moved] = cols.splice(oldIndex, 1)
  cols.splice(newIndex, 0, moved)

  // Assign new positions
  const updates = cols.map((col, i) => ({ id: col.id, position: (i + 1) * 1000 }))
  boardStore.columns = cols

  try {
    await projectsApi.reorderColumns(slug.value, updates)
  } catch {
    ui.error('Failed to reorder columns')
    await boardStore.loadBoard(slug.value)
  }
}

async function onRenameColumn({ columnId, name }) {
  try {
    await projectsApi.updateColumn(slug.value, columnId, { name })
    const col = boardStore.columns.find(c => c.id === columnId)
    if (col) col.name = name
  } catch (e) {
    ui.error('Failed to rename column')
  }
}

async function onDeleteColumn(columnId) {
  if (!confirm(t('board.delete_column_confirm'))) return
  try {
    await projectsApi.deleteColumn(slug.value, columnId)
    boardStore.columns = boardStore.columns.filter(c => c.id !== columnId)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to delete column')
  }
}

async function onCardMoved({ cardId, fromColumnId, toColumnId, newIndex }) {
  const toCol = boardStore.columns.find(c => c.id === toColumnId)
  if (!toCol) return

  const cards = toCol.cards.filter(c => c.id !== cardId)
  let position
  if (newIndex === 0) {
    position = (cards[0]?.position || 1000) / 2
  } else if (newIndex >= cards.length) {
    position = (cards[cards.length - 1]?.position || 0) + 1000
  } else {
    position = ((cards[newIndex - 1]?.position || 0) + (cards[newIndex]?.position || cards[newIndex - 1]?.position + 2000)) / 2
  }

  try {
    await projectsApi.moveCard(slug.value, cardId, { column_id: toColumnId, position })
  } catch (e) {
    ui.error('Failed to move card')
    await boardStore.loadBoard(slug.value) // revert
  }
}
</script>

<style scoped>
.board-layout { display: flex; flex-direction: column; flex: 1; min-height: 0; overflow: hidden; }

.board-toolbar {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  padding: 8px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.board-toolbar-left, .board-toolbar-right { display: flex; gap: 8px; align-items: center; }
.board-project-avatar {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  object-fit: cover;
  border: 1px solid var(--color-border);
}
.board-project-name { font-size: 15px; font-weight: 600; color: var(--color-text); padding: 0 4px; }
.star-btn { font-size: 18px; line-height: 1; color: var(--color-text-muted); }
.star-btn.starred { color: #f59e0b; }
.settings-link { font-size: 15px; color: var(--color-text-muted); }

.board-body {
  flex: 1;
  overflow: hidden;
  transition: margin-right .25s;
}

.board-columns-wrap {
  height: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 20px;
}

.board-columns {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  height: 100%;
}

.board-loading { display: flex; justify-content: center; align-items: center; height: 200px; }

.breadcrumb-sep { color: var(--color-text-muted); margin: 0 4px; }
.breadcrumb-link { font-size: 14px; color: var(--color-text-muted); }
.breadcrumb-current { font-size: 14px; font-weight: 600; }

:global(.column-ghost) { opacity: 0.4; background: var(--color-primary) !important; }
:global(.column-drag) { transform: rotate(1deg); box-shadow: var(--shadow-md); }

.form-hint { font-size: 12px; color: var(--color-text-muted); margin-top: 4px; margin-bottom: 0; }
</style>
