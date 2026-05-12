<template>
  <div class="board-layout">
    <div class="board-toolbar">
      <div class="board-toolbar-left">
        <img v-if="projectAvatar(projectStore.currentProject)" :src="projectAvatar(projectStore.currentProject)" class="board-project-avatar" alt="" />
        <h1 class="board-project-name">{{ projectStore.currentProject?.name }}</h1>
        <template v-if="sprintStore.activeSprint">
          <span class="sprint-badge active">{{ $t('sprint.status_active') }}</span>
          <span class="sprint-name-label">{{ sprintStore.activeSprint.name }}</span>
          <span v-if="sprintStore.activeSprint.goal" class="sprint-goal-chip">{{ sprintStore.activeSprint.goal }}</span>
        </template>
      </div>
      <div class="board-toolbar-right">
        <RouterLink :to="`/projects/${slug}/backlog`" class="btn btn-ghost btn-sm">
          📋 {{ $t('sprint.backlog') }}
        </RouterLink>
        <RouterLink :to="`/projects/${slug}`" class="btn btn-ghost btn-sm">
          📌 Board
        </RouterLink>
        <RouterLink :to="`/projects/${slug}/charts`" class="btn btn-ghost btn-sm">
          📊 {{ $t('sprint.charts') }}
        </RouterLink>
        <template v-if="sprintStore.activeSprint && canManage">
          <span class="sp-summary">
            {{ sprintStore.activeSprint.completed_points }}/{{ sprintStore.activeSprint.total_points }} SP
          </span>
          <button class="btn btn-warning btn-sm" @click="completeSprint">
            {{ $t('sprint.complete') }}
          </button>
        </template>
      </div>
    </div>

    <div class="board-body">
      <div v-if="!sprintStore.activeSprint && !loading" class="no-sprint-msg">
        <p>{{ $t('sprint.no_active_sprint') }}</p>
        <RouterLink :to="`/projects/${slug}/backlog`" class="btn btn-primary">
          {{ $t('sprint.go_to_backlog') }}
        </RouterLink>
      </div>

      <div v-else-if="loading" class="board-loading">
        <div class="spinner" style="width:40px;height:40px;border-width:3px"></div>
      </div>

      <div v-else class="board-columns-wrap">
        <div class="board-columns" ref="columnsEl">
          <BoardColumn
            v-for="col in filteredColumns"
            :key="col.id"
            :column="col"
            :data-column-id="col.id"
            :can-manage-columns="false"
            @open-card="openCardDetail"
            @card-moved="onCardMoved"
          />
        </div>
      </div>
    </div>

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
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useBoardStore } from '@/stores/board'
import { useSprintStore } from '@/stores/sprint'
import { useProjectStore } from '@/stores/project'
import { useUIStore } from '@/stores/ui'
import { useAuthStore } from '@/stores/auth'
import { useWebSocket } from '@/composables/useWebSocket'
import { projectsApi } from '@/api/projects'
import { resolveAssetUrl } from '@/api/serverConfig'
import BoardColumn from '@/components/board/BoardColumn.vue'
import CardDetail from '@/components/board/CardDetail.vue'

const route = useRoute()
const { t } = useI18n()
const slug = computed(() => route.params.slug)

const boardStore = useBoardStore()
const sprintStore = useSprintStore()
const projectStore = useProjectStore()
const ui = useUIStore()
const auth = useAuthStore()

const selectedCard = ref(null)
const projectMembers = ref([])
const loading = ref(true)
const columnsEl = ref(null)

const { connect, disconnect } = useWebSocket(slug)

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}

const canManage = computed(() => {
  if (auth.user?.global_role === 'admin') return true
  const me = projectMembers.value.find(m => m.user_id === auth.user?.id)
  return me ? ['owner', 'admin', 'member'].includes(me.role) : false
})

// Filter each column's cards to only those in the active sprint
const filteredColumns = computed(() => {
  const sprint = sprintStore.activeSprint
  if (!sprint) return []
  const ids = new Set(sprint.card_ids)
  return boardStore.columns.map(col => ({
    ...col,
    cards: col.cards.filter(c => ids.has(c.id))
  }))
})

async function load() {
  loading.value = true
  try {
    await Promise.all([
      boardStore.loadBoard(slug.value),
      sprintStore.loadSprints(slug.value),
      projectStore.fetchProject(slug.value),
    ])
    const { data } = await projectsApi.listMembers(slug.value)
    projectMembers.value = data
    connect()
  } finally {
    loading.value = false
  }
}

onMounted(() => load())

watch(slug, async (newSlug) => {
  disconnect()
  boardStore.reset()
  sprintStore.reset()
  await load()
})

onUnmounted(() => {
  disconnect()
  boardStore.reset()
  sprintStore.reset()
})

async function openCardDetail(card) {
  try {
    const { data } = await projectsApi.getCard(slug.value, card.id)
    selectedCard.value = data
  } catch {
    selectedCard.value = card
  }
}

function onCardMoved() {
  // card moves are handled by boardStore via WebSocket; nothing extra needed
}

async function completeSprint() {
  if (!confirm(t('sprint.complete_sprint_confirm'))) return
  try {
    await sprintStore.completeSprint(sprintStore.activeSprint.id)
    ui.success(t('sprint.complete_success'))
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed')
  }
}
</script>

<style scoped>
/* Reuses board-layout, board-toolbar, board-body, board-columns from global styles */

.board-project-avatar {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  object-fit: cover;
  border: 1px solid var(--color-border);
}

.sprint-badge {
  font-size: 11px;
  border-radius: 9999px;
  padding: 2px 8px;
  font-weight: 600;
}
.sprint-badge.active { background: #dcfce7; color: #166534; }

.sprint-name-label {
  font-weight: 600;
  font-size: 14px;
}

.sprint-goal-chip {
  font-size: 12px;
  color: var(--color-text-muted);
  font-style: italic;
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sp-summary {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-muted);
}

.no-sprint-msg {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--color-text-muted);
}

.no-sprint-msg p {
  font-size: 15px;
  margin: 0;
}

.board-loading {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-warning {
  background: var(--color-warning, #f59e0b);
  color: #fff;
  border: none;
  border-radius: var(--border-radius, 6px);
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}
.btn-warning:hover { opacity: .9; }
</style>
