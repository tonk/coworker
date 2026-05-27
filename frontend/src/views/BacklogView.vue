<template>
  <div class="backlog-layout">
    <div class="backlog-toolbar">
      <div class="backlog-toolbar-left">
        <img v-if="projectAvatar(projectStore.currentProject)" :src="projectAvatar(projectStore.currentProject)" class="board-project-avatar" alt="" />
        <h1 class="board-project-name">{{ projectStore.currentProject?.name }}</h1>
      </div>
      <div class="backlog-toolbar-right">
        <RouterLink :to="`/projects/${slug}`" class="btn btn-ghost btn-sm">
          📌 Board
        </RouterLink>
        <RouterLink v-if="sprintStore.activeSprint" :to="`/projects/${slug}/sprint`" class="btn btn-ghost btn-sm">
          🏃 {{ $t('sprint.board') }}
        </RouterLink>
        <RouterLink :to="`/projects/${slug}/charts`" class="btn btn-ghost btn-sm">
          📊 {{ $t('sprint.charts') }}
        </RouterLink>
      </div>
    </div>

    <!-- Tab bar: Backlog | Velocity -->
    <div class="backlog-tabs" role="tablist" :aria-label="$t('sprint.backlog_title')">
      <button :class="['tab', { active: view === 'backlog' }]" @click="view = 'backlog'" role="tab" :aria-selected="view === 'backlog'" aria-controls="tab-panel-backlog" id="tab-btn-backlog">
        📋 {{ $t('sprint.backlog_title') }}
      </button>
      <button :class="['tab', { active: view === 'velocity' }]" @click="view = 'velocity'" role="tab" :aria-selected="view === 'velocity'" aria-controls="tab-panel-velocity" id="tab-btn-velocity">
        📈 {{ $t('sprint.velocity_title') }}
      </button>
    </div>

    <!-- Backlog view -->
    <div v-show="view === 'backlog'" class="backlog-body" role="tabpanel" id="tab-panel-backlog" aria-labelledby="tab-btn-backlog">
      <!-- Left: Backlog cards -->
      <div class="backlog-column">
        <div class="backlog-column-header">
          <h2>{{ $t('sprint.backlog_title') }}</h2>
          <span class="card-count">{{ sprintStore.backlog.length }} {{ $t('sprint.cards') }}</span>
        </div>
        <div v-if="!sprintStore.backlog.length" class="backlog-empty">
          {{ $t('sprint.no_backlog') }}
        </div>
        <div
          v-for="card in sprintStore.backlog"
          :key="card.id"
          class="backlog-card"
          @click="openCard(card)"
        >
          <div class="backlog-card-title">
            <span class="card-ref">{{ projectStore.currentProject?.key_prefix }}-{{ card.card_number }}</span>
            {{ card.title }}
          </div>
          <div class="backlog-card-meta">
            <span v-if="card.story_points != null" class="sp-badge">{{ card.story_points }} SP</span>
            <span v-if="card.priority && card.priority !== 'none'" class="priority-badge" :class="card.priority">{{ card.priority }}</span>
            <div class="backlog-card-actions">
              <select
                v-if="planningOrActiveSprints.length"
                class="sprint-select"
                @change="assignToSprint(card, $event)"
                :title="$t('sprint.add_to_sprint')"
              >
                <option value="">+ {{ $t('sprint.add_to_sprint') }}</option>
                <option v-for="s in planningOrActiveSprints" :key="s.id" :value="s.id">
                  {{ s.name }}
                </option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Sprints -->
      <div class="sprints-column">
        <div class="sprints-column-header">
          <h2>{{ $t('sprint.title') }}</h2>
          <button v-if="canManage" class="btn btn-primary btn-sm" @click="showCreateSprint = true">
            + {{ $t('sprint.new_sprint') }}
          </button>
        </div>

        <div v-if="!sprintStore.sprints.length" class="backlog-empty">
          {{ $t('sprint.no_sprints') }}
        </div>

        <div v-for="sprint in sprintStore.sprints" :key="sprint.id" class="sprint-block">
          <div class="sprint-block-header">
            <div class="sprint-block-title">
              <span class="sprint-status-badge" :class="sprint.status">{{ $t(`sprint.status_${sprint.status}`) }}</span>
              <strong>{{ sprint.name }}</strong>
              <span v-if="sprint.goal" class="sprint-goal-text">{{ sprint.goal }}</span>
            </div>
            <div class="sprint-block-meta">
              <span class="sprint-dates" v-if="sprint.start_date || sprint.end_date">
                {{ fmtDate(sprint.start_date) }} – {{ fmtDate(sprint.end_date) }}
              </span>
              <span class="sp-badge">{{ sprint.completed_points }}/{{ sprint.total_points }} SP</span>
              <div class="sprint-block-actions">
                <button v-if="canManage && sprint.status === 'planning'" class="btn btn-primary btn-sm" @click="startSprint(sprint)">
                  {{ $t('sprint.start') }}
                </button>
                <button v-if="canManage && sprint.status === 'active'" class="btn btn-warning btn-sm" @click="completeSprint(sprint)">
                  {{ $t('sprint.complete') }}
                </button>
                <button v-if="canManage" class="btn btn-ghost btn-sm" @click="editSprint(sprint)">
                  {{ $t('common.edit') }}
                </button>
                <button v-if="canManage && sprint.status !== 'active'" class="btn btn-danger btn-sm" @click="deleteSprint(sprint)">
                  {{ $t('common.delete') }}
                </button>
              </div>
            </div>
          </div>

          <!-- Cards in this sprint -->
          <div class="sprint-cards">
            <div v-if="!sprint.card_ids.length" class="sprint-empty">
              {{ $t('sprint.no_cards_yet') }}
            </div>
            <div
              v-for="cardId in sprint.card_ids"
              :key="cardId"
              class="sprint-card-row"
            >
              <template v-if="boardCardById(cardId)">
                <span class="card-ref">{{ projectStore.currentProject?.key_prefix }}-{{ boardCardById(cardId).card_number }}</span>
                <span class="sprint-card-title" @click="openCardById(cardId)">{{ boardCardById(cardId).title }}</span>
                <span v-if="boardCardById(cardId).story_points != null" class="sp-badge sp-sm">{{ boardCardById(cardId).story_points }}</span>
                <span class="closed-badge" v-if="boardCardById(cardId).closed">✓</span>
                <button v-if="canManage && sprint.status !== 'completed'" class="btn-icon remove-btn" @click="removeFromSprint(sprint, cardId)" :title="$t('sprint.remove_from_sprint')" :aria-label="$t('sprint.remove_from_sprint')">✕</button>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Velocity view -->
    <div v-show="view === 'velocity'" class="velocity-body" role="tabpanel" id="tab-panel-velocity" aria-labelledby="tab-btn-velocity">
      <h2>{{ $t('sprint.velocity_title') }}</h2>
      <div v-if="!completedSprints.length" class="backlog-empty">
        {{ $t('sprint.no_completed_sprints') }}
      </div>
      <div v-else class="velocity-chart-wrap">
        <svg class="velocity-svg" :viewBox="`0 0 ${svgWidth} ${svgHeight}`" :width="svgWidth" :height="svgHeight">
          <!-- Y axis labels -->
          <text v-for="tick in yTicks" :key="tick" :x="yAxisWidth - 6" :y="svgPadTop + chartHeight - (tick / maxPoints * chartHeight) + 4" class="axis-label" text-anchor="end">{{ tick }}</text>
          <!-- Grid lines -->
          <line v-for="tick in yTicks" :key="`g${tick}`"
            :x1="yAxisWidth" :y1="svgPadTop + chartHeight - (tick / maxPoints * chartHeight)"
            :x2="svgWidth - svgPadRight" :y2="svgPadTop + chartHeight - (tick / maxPoints * chartHeight)"
            class="grid-line" />
          <!-- Bars -->
          <g v-for="(sprint, i) in completedSprints" :key="sprint.id">
            <!-- Total points bar (background) -->
            <rect
              :x="yAxisWidth + i * barGroupW + barGap"
              :y="svgPadTop + chartHeight - (sprint.total_points / maxPoints * chartHeight)"
              :width="barW / 2 - 2"
              :height="sprint.total_points / maxPoints * chartHeight"
              class="bar-total"
            />
            <!-- Completed points bar (foreground) -->
            <rect
              :x="yAxisWidth + i * barGroupW + barGap + barW / 2"
              :y="svgPadTop + chartHeight - (sprint.completed_points / maxPoints * chartHeight)"
              :width="barW / 2 - 2"
              :height="sprint.completed_points / maxPoints * chartHeight"
              class="bar-completed"
            />
            <!-- Sprint name label -->
            <text
              :x="yAxisWidth + i * barGroupW + barGroupW / 2"
              :y="svgPadTop + chartHeight + 14"
              class="axis-label sprint-label"
              text-anchor="middle"
            >{{ sprint.name.length > 10 ? sprint.name.slice(0, 9) + '…' : sprint.name }}</text>
          </g>
          <!-- X axis line -->
          <line :x1="yAxisWidth" :y1="svgPadTop + chartHeight" :x2="svgWidth - svgPadRight" :y2="svgPadTop + chartHeight" class="axis-line" />
        </svg>
        <!-- Legend -->
        <div class="velocity-legend">
          <span class="legend-item"><span class="legend-dot total"></span> {{ $t('sprint.total_points') }}</span>
          <span class="legend-item"><span class="legend-dot completed"></span> {{ $t('sprint.completed_points') }}</span>
        </div>
        <!-- Table -->
        <table class="data-table velocity-table">
          <thead>
            <tr>
              <th>{{ $t('sprint.sprint_name') }}</th>
              <th>{{ $t('sprint.total_points') }}</th>
              <th>{{ $t('sprint.completed_points') }}</th>
              <th>{{ $t('sprint.card_count_label') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in completedSprints" :key="s.id">
              <td>{{ s.name }}</td>
              <td>{{ s.total_points }}</td>
              <td>{{ s.completed_points }}</td>
              <td>{{ s.card_count }}</td>
            </tr>
          </tbody>
        </table>
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

    <!-- Create/edit sprint modal -->
    <BaseModal v-if="showCreateSprint || editingSprintId" :title="editingSprintId ? $t('sprint.edit') : $t('sprint.new_sprint')" @close="closeSprintModal">
      <div class="form-group">
        <label class="form-label">{{ $t('sprint.sprint_name') }}</label>
        <input class="form-input" v-model="sprintForm.name" autofocus />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('sprint.sprint_goal') }}</label>
        <textarea class="form-input" v-model="sprintForm.goal" rows="2"></textarea>
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('sprint.start_date') }}</label>
        <div style="position:relative;display:inline-block;width:100%">
          <input type="text" class="form-input" v-model="displayStartDate" :placeholder="dateOnlyFormat()" @change="onTextStartDate" />
          <input type="date" style="position:absolute;opacity:0;right:0;top:0;width:28px;height:100%;cursor:pointer" @change="onPickerStartDate" />
        </div>
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('sprint.end_date') }}</label>
        <div style="position:relative;display:inline-block;width:100%">
          <input type="text" class="form-input" v-model="displayEndDate" :placeholder="dateOnlyFormat()" @change="onTextEndDate" />
          <input type="date" style="position:absolute;opacity:0;right:0;top:0;width:28px;height:100%;cursor:pointer" @change="onPickerEndDate" />
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="closeSprintModal">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="!sprintForm.name.trim()" @click="saveSprint">
          {{ editingSprintId ? $t('common.save') : $t('common.create') }}
        </button>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useSprintStore } from '@/stores/sprint'
import { useProjectStore } from '@/stores/project'
import { useBoardStore } from '@/stores/board'
import { useUIStore } from '@/stores/ui'
import { useAuthStore } from '@/stores/auth'
import { useWebSocket } from '@/composables/useWebSocket'
import { useDateFormat } from '@/composables/useDateFormat'
import { projectsApi } from '@/api/projects'
import { resolveAssetUrl } from '@/api/serverConfig'
import CardDetail from '@/components/board/CardDetail.vue'
import BaseModal from '@/components/common/BaseModal.vue'

const route = useRoute()
const { t } = useI18n()
const slug = computed(() => route.params.slug)

const sprintStore = useSprintStore()
const projectStore = useProjectStore()
const boardStore = useBoardStore()
const ui = useUIStore()
const auth = useAuthStore()

const { formatDate, dateOnlyFormat } = useDateFormat()

const view = ref('backlog')
const selectedCard = ref(null)
const projectMembers = ref([])
const showCreateSprint = ref(false)
const editingSprintId = ref(null)
const sprintForm = ref({ name: '', goal: '', startDate: '', endDate: '' })
const displayStartDate = ref('')
const displayEndDate = ref('')

const { connect, disconnect } = useWebSocket(slug)

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}

const completedSprints = computed(() => sprintStore.completedSprints)
const planningOrActiveSprints = computed(() => sprintStore.sprints.filter(s => s.status !== 'completed'))

const canManage = computed(() => {
  if (auth.user?.global_role === 'admin') return true
  const me = projectMembers.value.find(m => m.user_id === auth.user?.id)
  return me ? ['owner', 'admin', 'member'].includes(me.role) : false
})

// Build a card lookup map from all board columns
const cardMap = computed(() => {
  const m = new Map()
  for (const col of boardStore.columns) {
    for (const card of col.cards) m.set(card.id, card)
  }
  return m
})

function boardCardById(id) {
  return cardMap.value.get(id) || null
}

onMounted(async () => {
  await Promise.all([
    sprintStore.loadSprints(slug.value),
    sprintStore.loadBacklog(slug.value),
    boardStore.loadBoard(slug.value),
    projectStore.fetchProject(slug.value),
  ])
  const { data } = await projectsApi.listMembers(slug.value)
  projectMembers.value = data
  connect()
})

onUnmounted(() => {
  disconnect()
  sprintStore.reset()
  boardStore.reset()
})

function fmtDate(d) {
  if (!d) return '?'
  return formatDate(d)
}

function onPickerStartDate(e) {
  sprintForm.value.startDate = e.target.value
  displayStartDate.value = e.target.value ? formatDate(e.target.value) : ''
}

function onPickerEndDate(e) {
  sprintForm.value.endDate = e.target.value
  displayEndDate.value = e.target.value ? formatDate(e.target.value) : ''
}

function parseSprintDate(str) {
  if (!str) return ''
  const fmt = dateOnlyFormat()
  const yPos = fmt.indexOf('YYYY')
  const mPos = fmt.indexOf('MM')
  const dPos = fmt.indexOf('DD')
  if (yPos < 0 || mPos < 0 || dPos < 0) return ''
  const y = parseInt(str.slice(yPos, yPos + 4))
  const m = parseInt(str.slice(mPos, mPos + 2))
  const d = parseInt(str.slice(dPos, dPos + 2))
  if (!y || m < 1 || m > 12 || d < 1 || d > 31) return ''
  return `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
}

function onTextStartDate() {
  const iso = parseSprintDate(displayStartDate.value)
  if (iso) {
    sprintForm.value.startDate = iso
    displayStartDate.value = formatDate(iso)
  } else if (!displayStartDate.value) {
    sprintForm.value.startDate = ''
  } else {
    displayStartDate.value = sprintForm.value.startDate ? formatDate(sprintForm.value.startDate) : ''
  }
}

function onTextEndDate() {
  const iso = parseSprintDate(displayEndDate.value)
  if (iso) {
    sprintForm.value.endDate = iso
    displayEndDate.value = formatDate(iso)
  } else if (!displayEndDate.value) {
    sprintForm.value.endDate = ''
  } else {
    displayEndDate.value = sprintForm.value.endDate ? formatDate(sprintForm.value.endDate) : ''
  }
}

async function openCard(card) {
  try {
    const { data } = await projectsApi.getCard(slug.value, card.id)
    selectedCard.value = data
  } catch {
    selectedCard.value = card
  }
}

async function openCardById(cardId) {
  const card = boardCardById(cardId)
  if (card) await openCard(card)
}

async function assignToSprint(card, event) {
  const sprintId = Number(event.target.value)
  event.target.value = ''
  if (!sprintId) return
  try {
    await sprintStore.addCardToSprint(sprintId, card.id)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed')
  }
}

async function removeFromSprint(sprint, cardId) {
  try {
    await sprintStore.removeCardFromSprint(sprint.id, cardId)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed')
  }
}

function editSprint(sprint) {
  editingSprintId.value = sprint.id
  const startIso = sprint.start_date ? sprint.start_date.slice(0, 10) : ''
  const endIso = sprint.end_date ? sprint.end_date.slice(0, 10) : ''
  sprintForm.value = { name: sprint.name, goal: sprint.goal || '', startDate: startIso, endDate: endIso }
  displayStartDate.value = startIso ? formatDate(startIso) : ''
  displayEndDate.value = endIso ? formatDate(endIso) : ''
}

function closeSprintModal() {
  showCreateSprint.value = false
  editingSprintId.value = null
  sprintForm.value = { name: '', goal: '', startDate: '', endDate: '' }
  displayStartDate.value = ''
  displayEndDate.value = ''
}

async function saveSprint() {
  if (!sprintForm.value.name.trim()) return
  const payload = {
    name: sprintForm.value.name,
    goal: sprintForm.value.goal,
    start_date: sprintForm.value.startDate || null,
    end_date: sprintForm.value.endDate || null,
  }
  try {
    if (editingSprintId.value) {
      await sprintStore.updateSprint(editingSprintId.value, payload)
    } else {
      await sprintStore.createSprint(payload)
    }
    closeSprintModal()
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed')
  }
}

async function startSprint(sprint) {
  if (!await ui.confirm(t('sprint.start_sprint_confirm'))) return
  try {
    await sprintStore.startSprint(sprint.id)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed')
  }
}

async function completeSprint(sprint) {
  if (!await ui.confirm(t('sprint.complete_sprint_confirm'))) return
  try {
    await sprintStore.completeSprint(sprint.id)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed')
  }
}

async function deleteSprint(sprint) {
  if (!await ui.confirm(t('sprint.delete_sprint_confirm'))) return
  try {
    await sprintStore.deleteSprint(sprint.id)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed')
  }
}

// Velocity chart dimensions
const svgPadTop = 20
const svgPadRight = 20
const yAxisWidth = 40
const chartHeight = 200
const barGap = 8
const barGroupW = computed(() => completedSprints.value.length > 0 ? Math.max(60, Math.min(120, 600 / completedSprints.value.length)) : 80)
const barW = computed(() => barGroupW.value - barGap * 2)
const svgWidth = computed(() => yAxisWidth + completedSprints.value.length * barGroupW.value + svgPadRight + barGap)
const svgHeight = svgPadTop + chartHeight + 30

const maxPoints = computed(() => {
  const max = Math.max(...completedSprints.value.map(s => s.total_points), 1)
  return Math.ceil(max / 5) * 5
})

const yTicks = computed(() => {
  const step = Math.max(1, Math.ceil(maxPoints.value / 4))
  const ticks = []
  for (let v = step; v <= maxPoints.value; v += step) ticks.push(v)
  return ticks
})
</script>

<style scoped>
.backlog-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.backlog-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
}

.backlog-toolbar-left, .backlog-toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.board-project-avatar {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  object-fit: cover;
  border: 1px solid var(--color-border);
}

.board-project-name {
  font-weight: 600;
  font-size: 15px;
}

.backlog-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
  padding: 0 16px;
}

.backlog-tabs .tab {
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  padding: 10px 16px;
  font-size: 13px;
  cursor: pointer;
  color: var(--color-text-muted);
  transition: color .15s, border-color .15s;
}
.backlog-tabs .tab.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.backlog-body {
  display: flex;
  flex: 1;
  overflow: hidden;
  gap: 0;
}

.backlog-column {
  flex: 1;
  min-width: 0;
  border-right: 1px solid var(--color-border);
  overflow-y: auto;
  padding: 16px;
}

.sprints-column {
  flex: 1.4;
  min-width: 0;
  overflow-y: auto;
  padding: 16px;
}

.backlog-column-header, .sprints-column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.backlog-column-header h2, .sprints-column-header h2 {
  font-size: 15px;
  font-weight: 600;
  margin: 0;
}

.card-count {
  background: var(--color-primary);
  color: #fff;
  border-radius: 9999px;
  padding: 1px 7px;
  font-size: 11px;
  font-weight: 600;
}

.backlog-empty {
  padding: 24px;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 13px;
}

.backlog-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 10px 12px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: box-shadow .1s;
}
.backlog-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,.1); }

.backlog-card-title {
  font-size: 13px;
  margin-bottom: 6px;
  line-height: 1.4;
}

.card-ref {
  font-size: 11px;
  color: var(--color-text-muted);
  font-family: monospace;
  margin-right: 6px;
}

.backlog-card-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.sp-badge {
  font-size: 11px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 9999px;
  padding: 1px 7px;
  font-weight: 600;
}
.sp-badge.sp-sm { font-size: 10px; padding: 1px 5px; }

.priority-badge {
  font-size: 11px;
  border-radius: 4px;
  padding: 1px 6px;
  font-weight: 600;
  text-transform: capitalize;
}
.priority-badge.high, .priority-badge.critical { background: #fee2e2; color: #b91c1c; }
.priority-badge.medium { background: #fef3c7; color: #92400e; }
.priority-badge.low { background: #dcfce7; color: #166534; }

.backlog-card-actions { margin-left: auto; }

.sprint-select {
  font-size: 12px;
  padding: 2px 6px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface);
  color: var(--color-text);
  cursor: pointer;
}

.sprint-block {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  margin-bottom: 16px;
  overflow: hidden;
}

.sprint-block-header {
  padding: 12px 14px;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-bg);
}

.sprint-block-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.sprint-goal-text {
  font-size: 12px;
  color: var(--color-text-muted);
  font-style: italic;
}

.sprint-block-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.sprint-block-actions {
  display: flex;
  gap: 6px;
  margin-left: auto;
}

.sprint-status-badge {
  font-size: 11px;
  border-radius: 9999px;
  padding: 2px 8px;
  font-weight: 600;
  text-transform: capitalize;
}
.sprint-status-badge.planning { background: #e0f2fe; color: #0369a1; }
.sprint-status-badge.active   { background: #dcfce7; color: #166534; }
.sprint-status-badge.completed { background: #f3f4f6; color: #6b7280; }

.sprint-dates { font-size: 12px; color: var(--color-text-muted); }

.sprint-cards {
  padding: 8px 12px;
}

.sprint-empty {
  padding: 8px;
  font-size: 12px;
  color: var(--color-text-muted);
  text-align: center;
}

.sprint-card-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 4px;
  border-radius: 4px;
  font-size: 13px;
}
.sprint-card-row:hover { background: var(--color-bg); }

.sprint-card-title {
  flex: 1;
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sprint-card-title:hover { color: var(--color-primary); }

.closed-badge {
  font-size: 12px;
  color: var(--color-success, #22c55e);
  font-weight: 700;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 2px 4px;
  font-size: 12px;
  border-radius: 3px;
  opacity: 0;
  transition: opacity .1s, color .1s;
}
.sprint-card-row:hover .btn-icon { opacity: 1; }
.btn-icon:hover { color: var(--color-danger, #ef4444); }

.remove-btn { flex-shrink: 0; }

/* Velocity */
.velocity-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.velocity-body h2 {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 20px;
}

.velocity-chart-wrap {
  max-width: 700px;
}

.velocity-svg {
  display: block;
  overflow: visible;
}

.axis-label {
  font-size: 11px;
  fill: var(--color-text-muted, #9ca3af);
}

.sprint-label {
  font-size: 10px;
}

.grid-line {
  stroke: var(--color-border, #e5e7eb);
  stroke-width: 1;
}

.axis-line {
  stroke: var(--color-text-muted, #9ca3af);
  stroke-width: 1.5;
}

.bar-total { fill: var(--color-border, #d1d5db); }
.bar-completed { fill: var(--color-primary, #6366f1); }

.velocity-legend {
  display: flex;
  gap: 20px;
  margin: 12px 0 20px;
  font-size: 12px;
}

.legend-item { display: flex; align-items: center; gap: 6px; }

.legend-dot {
  width: 12px;
  height: 12px;
  border-radius: 2px;
}
.legend-dot.total { background: var(--color-border, #d1d5db); }
.legend-dot.completed { background: var(--color-primary, #6366f1); }

.velocity-table { max-width: 500px; }


</style>
