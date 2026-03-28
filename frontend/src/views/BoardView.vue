<template>
  <div class="board-layout">
    <div class="board-toolbar">
      <div class="board-toolbar-left">
        <RouterLink :to="`/projects/${slug}/settings`" class="btn btn-ghost btn-sm">
          ⚙ {{ $t('project.settings') }}<span v-if="sidebarStore.isStarred(slug)" class="settings-star">★</span>
        </RouterLink>
        <button class="btn btn-ghost btn-sm star-btn" :class="{ starred: sidebarStore.isStarred(slug) }" @click="toggleStar" :title="sidebarStore.isStarred(slug) ? $t('board.unstar') : $t('board.star')">
          {{ sidebarStore.isStarred(slug) ? '★' : '☆' }}
        </button>
      </div>
      <div class="board-toolbar-right">
        <RouterLink :to="`/projects/${slug}/topics`" class="btn btn-ghost btn-sm">
          💬 {{ $t('topics.title') }}
        </RouterLink>
        <button class="btn btn-ghost btn-sm" @click="chatOpen = !chatOpen">
          💬 {{ $t('chat.title') }}
        </button>
        <button class="btn btn-secondary btn-sm" @click="showAddColumn = true">
          + {{ $t('board.add_column') }}
        </button>
      </div>
    </div>

    <div class="board-body" :style="{ marginRight: chatOpen ? '340px' : '0' }">
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
            @add-card="openAddCard"
            @open-card="openCardDetail"
            @card-moved="onCardMoved"
            @rename-column="onRenameColumn"
          />
        </div>
      </div>
    </div>

    <ChatPanel
      :open="chatOpen"
      :project-slug="slug"
      :ws-send="wsSend"
      @close="chatOpen = false"
    />

    <!-- Add card modal -->
    <BaseModal v-if="showAddCard" :title="$t('board.add_card')" @close="showAddCard = false">
      <form @submit.prevent="submitAddCard">
        <div class="form-group">
          <label class="form-label">{{ $t('board.card_title') }}</label>
          <input class="form-input" v-model="newCard.title" required autofocus />
        </div>
      </form>
      <template #footer>
        <button class="btn btn-secondary" @click="showAddCard = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="submitAddCard">{{ $t('common.create') }}</button>
      </template>
    </BaseModal>

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
      </form>
      <template #footer>
        <button class="btn btn-secondary" @click="showAddColumn = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="submitAddColumn">{{ $t('common.create') }}</button>
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import Sortable from 'sortablejs'
import BoardColumn from '@/components/board/BoardColumn.vue'
import CardDetail from '@/components/board/CardDetail.vue'
import ChatPanel from '@/components/chat/ChatPanel.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import { useBoardStore } from '@/stores/board'
import { useProjectStore } from '@/stores/project'
import { useUIStore } from '@/stores/ui'
import { useSidebarStore } from '@/stores/sidebar'
import { useWebSocket } from '@/composables/useWebSocket'
import { projectsApi } from '@/api/projects'

const route = useRoute()
const slug = computed(() => route.params.slug)

const boardStore = useBoardStore()
const projectStore = useProjectStore()
const ui = useUIStore()
const sidebarStore = useSidebarStore()

const chatOpen = ref(false)
const showAddCard = ref(false)
const showAddColumn = ref(false)
const selectedCard = ref(null)
const addCardColumnId = ref(null)
const newCard = ref({ title: '' })
const newColumn = ref({ name: '', color: '#94a3b8' })
const columnsEl = ref(null)
let columnSortable = null

const projectMembers = ref([])

const { connected, presenceUsers, connect, disconnect, send: wsSend } = useWebSocket(slug.value)

onMounted(async () => {
  await Promise.all([
    boardStore.loadBoard(slug.value),
    projectStore.fetchProject(slug.value),
    sidebarStore.fetchStarred()
  ])
  loadMembers()
  connect()
  initColumnSortable()
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

async function toggleStar() {
  if (!slug.value) return
  if (sidebarStore.isStarred(slug.value)) {
    await sidebarStore.unstarProject(slug.value)
  } else {
    await sidebarStore.starProject(slug.value)
  }
}

function openAddCard(columnId) {
  addCardColumnId.value = columnId
  newCard.value = { title: '' }
  showAddCard.value = true
}

async function submitAddCard() {
  if (!newCard.value.title) return
  try {
    await boardStore.createCard(addCardColumnId.value, newCard.value)
    showAddCard.value = false
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to create card')
  }
}

async function submitAddColumn() {
  if (!newColumn.value.name) return
  try {
    await boardStore.createColumn(newColumn.value.name, { color: newColumn.value.color })
    showAddColumn.value = false
    newColumn.value = { name: '', color: '#94a3b8' }
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to create column')
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
.board-layout { display: flex; flex-direction: column; height: 100%; overflow: hidden; }

.board-toolbar {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  padding: 8px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.board-toolbar-left, .board-toolbar-right { display: flex; gap: 8px; align-items: center; }
.star-btn { font-size: 18px; line-height: 1; color: var(--color-text-muted); }
.star-btn.starred { color: #f59e0b; }
.settings-star { color: #f59e0b; margin-left: 4px; font-size: 13px; }

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
</style>
