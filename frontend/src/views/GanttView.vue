<template>
  <div class="gantt-layout">
    <div class="gantt-toolbar">
      <div class="gantt-toolbar-left">
        <img v-if="projectAvatar(projectStore.currentProject)" :src="projectAvatar(projectStore.currentProject)" class="project-avatar" alt="" />
        <RouterLink :to="`/projects/${slug}`" class="breadcrumb-link">
          {{ projectStore.currentProject?.name }}
        </RouterLink>
        <span class="breadcrumb-sep">›</span>
        <span class="breadcrumb-cur">{{ $t('gantt.title') }}</span>
      </div>
      <div class="gantt-toolbar-right">
        <div class="view-mode-btns">
          <button
            v-for="mode in viewModes"
            :key="mode.value"
            class="btn btn-sm"
            :class="viewMode === mode.value ? 'btn-primary' : 'btn-ghost'"
            @click="setViewMode(mode.value)"
          >{{ $t(mode.label) }}</button>
        </div>
      </div>
    </div>

    <div class="gantt-body">
      <div v-if="loading" class="gantt-empty">
        <div class="spinner"></div>
      </div>
      <div v-else-if="!tasks.length" class="gantt-empty">
        {{ $t('gantt.no_dates') }}
      </div>
      <div v-show="!loading && tasks.length" ref="ganttEl" class="gantt-wrapper"></div>
    </div>

    <CardDetail
      v-if="selectedCard"
      :card="selectedCard"
      :project-slug="slug"
      :members="members"
      :labels="labels"
      @close="selectedCard = null; reload()"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Gantt from 'frappe-gantt'
import 'frappe-gantt/dist/frappe-gantt.css'
import { useBoardStore } from '@/stores/board'
import { useProjectStore } from '@/stores/project'
import { projectsApi } from '@/api/projects'
import { resolveAssetUrl } from '@/api/serverConfig'
import CardDetail from '@/components/board/CardDetail.vue'

const route = useRoute()
const { t } = useI18n()
const slug = computed(() => route.params.slug)

const boardStore = useBoardStore()
const projectStore = useProjectStore()

const loading = ref(true)
const ganttEl = ref(null)
const selectedCard = ref(null)
let ganttInstance = null

const viewModes = [
  { value: 'Day',   label: 'gantt.view_day' },
  { value: 'Week',  label: 'gantt.view_week' },
  { value: 'Month', label: 'gantt.view_month' },
]
const viewMode = ref('Week')

// Gather members and labels from project store for CardDetail
const members = computed(() => projectStore.currentProject?.members?.map(m => m.user) || [])
const labels  = computed(() => projectStore.currentProject?.labels || [])

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}

// Build frappe-gantt task list from board cards
const tasks = computed(() => {
  const result = []
  for (const col of boardStore.columns) {
    for (const card of col.cards) {
      const start = card.start_date
        ? card.start_date.slice(0, 10)
        : card.due_date
          ? card.due_date.slice(0, 10)
          : null
      const end = card.due_date
        ? card.due_date.slice(0, 10)
        : card.start_date
          ? card.start_date.slice(0, 10)
          : null

      if (!start && !end) continue

      // Ensure end >= start (frappe-gantt requires end > start for rendering)
      let s = start, e = end
      if (s && e && s > e) e = s
      if (!s) s = e
      if (!e) e = s

      // If same day, extend end by one day so the bar is visible
      if (s === e) {
        const d = new Date(e)
        d.setDate(d.getDate() + 1)
        e = d.toISOString().slice(0, 10)
      }

      const progress = card.closed ? 100
        : card.sub_card_count > 0
          ? Math.round(card.sub_cards_done / card.sub_card_count * 100)
          : 0

      result.push({
        id:       String(card.id),
        name:     `${projectStore.currentProject?.key_prefix || ''}-${card.card_number} ${card.title}`,
        start:    s,
        end:      e,
        progress,
        _card:    card,
        custom_class: card.closed ? 'bar-closed' : `bar-priority-${card.priority}`,
      })
    }
  }
  return result
})

function initGantt() {
  if (!ganttEl.value || !tasks.value.length) return
  ganttEl.value.innerHTML = ''
  const fmt = d => `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
  try {
    ganttInstance = new Gantt(ganttEl.value, tasks.value, {
      view_mode: viewMode.value,
      infinite_padding: false,
      scroll_to: 'today',
      on_click: (task) => {
        selectedCard.value = task._card
      },
      on_date_change: async (task, start, end) => {
        const card = task._card
        try {
          await projectsApi.updateCard(slug.value, card.id, {
            start_date: fmt(start),
            due_date:   fmt(end),
          })
          card.start_date = fmt(start)
          card.due_date   = fmt(end)
        } catch {}
      },
    })
  } catch (e) {
    console.error('Gantt init failed:', e)
  }
}

function setViewMode(mode) {
  viewMode.value = mode
  if (ganttInstance) ganttInstance.change_view_mode(mode)
}

async function reload() {
  loading.value = true
  try {
    await boardStore.loadBoard(slug.value)
    await projectStore.fetchProject(slug.value)
  } catch (e) {
    console.error('[Gantt] reload failed:', e)
  } finally {
    loading.value = false
  }
  initGantt()
}

onMounted(reload)
watch(slug, reload)
watch(tasks, initGantt)
</script>

<style scoped>
.gantt-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.gantt-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 20px;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
}

.gantt-toolbar-left { display: flex; align-items: center; gap: 8px; }
.project-avatar {
  width: 20px;
  height: 20px;
  border-radius: 5px;
  object-fit: cover;
  border: 1px solid var(--color-border);
}
.gantt-toolbar-right { display: flex; align-items: center; gap: 8px; }

.breadcrumb-link {
  color: var(--color-text-muted);
  text-decoration: none;
  font-size: 14px;
}
.breadcrumb-link:hover { color: var(--color-primary); }
.breadcrumb-sep { color: var(--color-text-muted); font-size: 14px; }
.breadcrumb-cur { font-size: 14px; font-weight: 600; }

.view-mode-btns { display: flex; gap: 4px; }

.gantt-body {
  flex: 1;
  overflow: auto;
  padding: 20px;
  background: var(--color-bg);
}

.gantt-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--color-text-muted);
  font-size: 14px;
}

.gantt-wrapper {
  min-width: 100%;
}

/* ── Frappe-gantt v1.2.2 theme overrides via CSS custom properties ── */
/* Setting variables on the container avoids specificity conflicts with
   frappe's own stylesheet which uses the same --g-* variables. */
:deep(.gantt-container) {
  font-family: inherit;
  /* colours */
  --g-header-background: var(--color-surface);
  --g-row-color:         var(--color-bg);
  --g-border-color:      var(--color-border);
  --g-row-border-color:  var(--color-border);
  --g-tick-color:        var(--color-border);
  --g-tick-color-thick:  var(--color-border);
  --g-text-muted:        var(--color-text-muted);
  --g-text-dark:         var(--color-text);
  /* bar defaults — overridden per-bar by priority classes below */
  --g-bar-color:         var(--color-primary);
  --g-bar-border:        var(--color-primary);
  --g-progress-color:    color-mix(in srgb, var(--color-primary) 60%, #000 40%);
  --g-text-light:        #fff;
}

/* Priority colours — set the variable on the group so the bar rect inherits it */
:deep(.bar-priority-critical) { --g-bar-color: #ef4444; --g-bar-border: #ef4444; }
:deep(.bar-priority-high)     { --g-bar-color: #f97316; --g-bar-border: #f97316; }
:deep(.bar-priority-medium)   { --g-bar-color: var(--color-primary); --g-bar-border: var(--color-primary); }
:deep(.bar-priority-low)      { --g-bar-color: #6b7280; --g-bar-border: #6b7280; }
:deep(.bar-priority-none)     { --g-bar-color: #9ca3af; --g-bar-border: #9ca3af; }
:deep(.bar-closed)            { --g-bar-color: #22c55e; --g-bar-border: #22c55e; opacity: .7; }

/* Text inside bars should always be legible */
:deep(.gantt .bar-label) { fill: var(--g-text-light) !important; font-family: inherit; font-size: 11px; }
</style>
