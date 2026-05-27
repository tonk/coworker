<template>
  <div class="charts-layout">
    <div class="board-toolbar">
      <div class="board-toolbar-left">
        <img v-if="projectAvatar(projectStore.currentProject)" :src="projectAvatar(projectStore.currentProject)" class="board-project-avatar" alt="" />
        <h1 class="board-project-name">{{ projectStore.currentProject?.name }}</h1>
      </div>
      <div class="board-toolbar-right">
        <RouterLink :to="`/projects/${slug}`" class="btn btn-ghost btn-sm">📌 Board</RouterLink>
        <RouterLink v-if="!isKanban" :to="`/projects/${slug}/backlog`" class="btn btn-ghost btn-sm">📋 {{ $t('sprint.backlog') }}</RouterLink>
        <RouterLink v-if="activeSprint" :to="`/projects/${slug}/sprint`" class="btn btn-ghost btn-sm">🏃 {{ $t('sprint.board') }}</RouterLink>
      </div>
    </div>

    <div class="charts-body">
      <div class="charts-tabs">
        <button v-for="tab in tabs" :key="tab.key" :class="['tab', { active: activeTab === tab.key }]" @click="activeTab = tab.key">
          {{ tab.icon }} {{ $t(tab.label) }}
        </button>
      </div>

      <!-- ── Velocity ─────────────────────────────────────────────────── -->
      <div v-if="activeTab === 'velocity'" class="chart-panel">
        <h2>{{ $t('sprint.velocity_title') }}</h2>
        <div v-if="velocityLoading" class="chart-loading"><div class="spinner"></div></div>
        <div v-else-if="!velocitySprints.length" class="chart-empty">{{ $t('sprint.no_completed_sprints') }}</div>
        <div v-else class="chart-wrap"><canvas ref="velocityCanvas"></canvas></div>
      </div>

      <!-- ── Throughput ───────────────────────────────────────────────── -->
      <div v-if="activeTab === 'throughput'" class="chart-panel">
        <div class="chart-panel-header">
          <h2>{{ $t('sprint.throughput') }}</h2>
          <select v-if="isKanban" class="form-input sprint-select" v-model="throughputWeeks" @change="loadWeeklyThroughput">
            <option :value="4">4 weeks</option>
            <option :value="8">8 weeks</option>
            <option :value="12">12 weeks</option>
            <option :value="26">26 weeks</option>
          </select>
        </div>
        <template v-if="isKanban">
          <div v-if="weeklyThroughputLoading" class="chart-loading"><div class="spinner"></div></div>
          <div v-else-if="!weeklyThroughput.some(w => w.count > 0)" class="chart-empty">{{ $t('sprint.no_closed_cards') }}</div>
          <div v-else class="chart-wrap"><canvas ref="weeklyThroughputCanvas"></canvas></div>
        </template>
        <template v-else>
          <div v-if="velocityLoading" class="chart-loading"><div class="spinner"></div></div>
          <div v-else-if="!velocitySprints.length" class="chart-empty">{{ $t('sprint.no_completed_sprints') }}</div>
          <div v-else class="chart-wrap"><canvas ref="throughputCanvas"></canvas></div>
        </template>
      </div>

      <!-- ── Burndown ─────────────────────────────────────────────────── -->
      <div v-if="activeTab === 'burndown'" class="chart-panel">
        <div class="chart-panel-header">
          <h2>{{ $t('sprint.burndown') }}</h2>
          <select class="form-input sprint-select" v-model="selectedSprintId" @change="loadBurndown">
            <option :value="null" disabled>— {{ $t('sprint.select_sprint') }} —</option>
            <option v-for="s in sprintsWithDates" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </div>
        <div v-if="burndownLoading" class="chart-loading"><div class="spinner"></div></div>
        <div v-else-if="!selectedSprintId" class="chart-empty">{{ $t('sprint.select_sprint') }}</div>
        <div v-else-if="burndownError" class="chart-empty">{{ burndownError }}</div>
        <div v-else-if="!burndownData.length" class="chart-empty">{{ $t('sprint.no_chart_data') }}</div>
        <div v-else class="chart-wrap"><canvas ref="burndownCanvas"></canvas></div>
      </div>

      <!-- ── Burnup ───────────────────────────────────────────────────── -->
      <div v-if="activeTab === 'burnup'" class="chart-panel">
        <div class="chart-panel-header">
          <h2>{{ $t('sprint.burnup') }}</h2>
          <select class="form-input sprint-select" v-model="selectedBurnupSprintId" @change="loadBurnup">
            <option :value="null" disabled>— {{ $t('sprint.select_sprint') }} —</option>
            <option v-for="s in sprintsWithDates" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </div>
        <div v-if="burnupLoading" class="chart-loading"><div class="spinner"></div></div>
        <div v-else-if="!selectedBurnupSprintId" class="chart-empty">{{ $t('sprint.select_sprint') }}</div>
        <div v-else-if="burnupError" class="chart-empty">{{ burnupError }}</div>
        <div v-else-if="!burnupData.length" class="chart-empty">{{ $t('sprint.no_chart_data') }}</div>
        <div v-else class="chart-wrap"><canvas ref="burnupCanvas"></canvas></div>
      </div>

      <!-- ── CFD ──────────────────────────────────────────────────────── -->
      <div v-if="activeTab === 'cfd'" class="chart-panel">
        <div class="chart-panel-header">
          <h2>{{ $t('sprint.cfd') }}</h2>
          <select class="form-input sprint-select" v-model="cfdDays" @change="loadCFD">
            <option :value="30">30 days</option>
            <option :value="60">60 days</option>
            <option :value="90">90 days</option>
            <option :value="180">180 days</option>
          </select>
        </div>
        <div v-if="cfdLoading" class="chart-loading"><div class="spinner"></div></div>
        <div v-else-if="!cfdLabels.length" class="chart-empty">{{ $t('sprint.no_chart_data') }}</div>
        <div v-else class="chart-wrap"><canvas ref="cfdCanvas"></canvas></div>
      </div>

      <!-- ── Release Burndown ─────────────────────────────────────────── -->
      <div v-if="activeTab === 'release-burndown'" class="chart-panel">
        <div class="chart-panel-header">
          <h2>{{ $t('sprint.release_burndown') }}</h2>
          <select class="form-input sprint-select" v-model="selectedReleaseId" @change="loadReleaseBurndown">
            <option :value="null" disabled>— {{ $t('sprint.select_release') }} —</option>
            <option v-for="r in releases" :key="r.id" :value="r.id">{{ r.name }}</option>
          </select>
          <button class="btn btn-secondary btn-sm" @click="showReleaseManager = !showReleaseManager">
            ⚙ {{ $t('sprint.manage_releases') }}
          </button>
        </div>

        <!-- Release manager -->
        <div v-if="showReleaseManager" class="release-manager">
          <div class="release-manager-header">
            <strong>{{ $t('sprint.manage_releases') }}</strong>
            <button class="btn btn-primary btn-sm" @click="openCreateRelease">+ {{ $t('sprint.new_release') }}</button>
          </div>
          <div v-if="!releases.length" class="release-empty">{{ $t('sprint.no_releases') }}</div>
          <div v-for="r in releases" :key="r.id" class="release-row">
            <div class="release-row-header">
              <span class="release-row-name">{{ r.name }}</span>
              <span v-if="r.target_date" class="release-row-date">🎯 {{ formatDate(r.target_date) }}</span>
              <button class="btn btn-ghost btn-sm" @click="openEditRelease(r)">✎</button>
              <button class="btn btn-ghost btn-sm danger" @click="confirmDeleteRelease(r)">✕</button>
            </div>
            <div class="release-sprints-label">{{ $t('sprint.title') }}:</div>
            <div class="release-sprint-chips">
              <template v-for="s in sprintStore.sprints" :key="s.id">
                <button
                  :class="['sprint-chip', { assigned: releaseHasSprint(r, s.id) }]"
                  @click="toggleSprintInRelease(r, s)"
                >{{ s.name }}</button>
              </template>
            </div>
          </div>
        </div>

        <div v-if="releaseBurndownLoading" class="chart-loading"><div class="spinner"></div></div>
        <div v-else-if="!selectedReleaseId" class="chart-empty">{{ $t('sprint.select_release') }}</div>
        <div v-else-if="releaseBurndownError" class="chart-empty">{{ releaseBurndownError }}</div>
        <div v-else-if="!releaseBurndownData.length" class="chart-empty">{{ $t('sprint.no_chart_data') }}</div>
        <div v-else class="chart-wrap"><canvas ref="releaseBurndownCanvas"></canvas></div>
      </div>

      <!-- ── Cycle Time ────────────────────────────────────────────────── -->
      <div v-if="activeTab === 'cycle-time'" class="chart-panel">
        <h2>{{ $t('sprint.cycle_time') }}</h2>
        <div v-if="cycleTimeLoading" class="chart-loading"><div class="spinner"></div></div>
        <div v-else-if="!cycleTimeCards.length" class="chart-empty">{{ $t('sprint.no_closed_cards') }}</div>
        <div v-else class="chart-wrap"><canvas ref="cycleTimeCanvas"></canvas></div>
      </div>

      <!-- ── Lead Time ─────────────────────────────────────────────────── -->
      <div v-if="activeTab === 'lead-time'" class="chart-panel">
        <h2>{{ $t('sprint.lead_time') }}</h2>
        <div v-if="cycleTimeLoading" class="chart-loading"><div class="spinner"></div></div>
        <div v-else-if="!cycleTimeCards.length" class="chart-empty">{{ $t('sprint.no_closed_cards') }}</div>
        <div v-else class="chart-wrap"><canvas ref="leadTimeCanvas"></canvas></div>
      </div>
    </div>
  </div>

  <!-- Create / Edit release modal -->
  <BaseModal v-if="showReleaseModal" :title="editingRelease ? $t('sprint.edit_release') : $t('sprint.new_release')" @close="closeReleaseModal">
    <div class="form-group">
      <label class="form-label">{{ $t('sprint.release_name') }}</label>
      <input class="form-input" v-model="releaseForm.name" autofocus />
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('sprint.sprint_goal') }}</label>
      <textarea class="form-input" v-model="releaseForm.goal" rows="2"></textarea>
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('sprint.target_date') }}</label>
      <input type="date" class="form-input" v-model="releaseForm.target_date" style="max-width:200px" />
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="closeReleaseModal">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" @click="saveRelease" :disabled="!releaseForm.name.trim()">{{ $t('common.save') }}</button>
    </template>
  </BaseModal>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Chart, registerables } from 'chart.js'
import { useProjectStore } from '@/stores/project'
import { useSprintStore } from '@/stores/sprint'
import { useUIStore } from '@/stores/ui'
import { projectsApi } from '@/api/projects'
import { resolveAssetUrl } from '@/api/serverConfig'
import BaseModal from '@/components/common/BaseModal.vue'
import { useDateFormat } from '@/composables/useDateFormat'

Chart.register(...registerables)

const { t } = useI18n()
const route = useRoute()
const projectStore = useProjectStore()
const sprintStore = useSprintStore()
const ui = useUIStore()
const { formatDate } = useDateFormat()

const slug = computed(() => route.params.slug)

const isKanban = computed(() => projectStore.currentProject?.board_type === 'kanban')

const tabs = computed(() => {
  if (isKanban.value) {
    return [
      { key: 'cfd',        icon: '🌊', label: 'sprint.cfd' },
      { key: 'cycle-time', icon: '⏱',  label: 'sprint.cycle_time' },
      { key: 'lead-time',  icon: '📐', label: 'sprint.lead_time' },
      { key: 'throughput', icon: '🔢', label: 'sprint.throughput' },
    ]
  }
  return [
    { key: 'velocity',         icon: '📊', label: 'sprint.velocity_title' },
    { key: 'throughput',       icon: '🔢', label: 'sprint.throughput' },
    { key: 'burndown',         icon: '📉', label: 'sprint.burndown' },
    { key: 'burnup',           icon: '📈', label: 'sprint.burnup' },
    { key: 'cfd',              icon: '🌊', label: 'sprint.cfd' },
    { key: 'release-burndown', icon: '🚀', label: 'sprint.release_burndown' },
    { key: 'cycle-time',       icon: '⏱',  label: 'sprint.cycle_time' },
    { key: 'lead-time',        icon: '📐', label: 'sprint.lead_time' },
  ]
})

const activeTab = ref('velocity')

// ── Shared data ──────────────────────────────────────────────────────────────
const velocityLoading = ref(false)
const velocitySprints = ref([])

const cycleTimeLoading = ref(false)
const cycleTimeCards = ref([])

// ── Per-chart state ──────────────────────────────────────────────────────────
const selectedSprintId = ref(null)
const selectedBurnupSprintId = ref(null)
const burndownLoading = ref(false)
const burndownData = ref([])
const burndownError = ref('')
const burnupLoading = ref(false)
const burnupData = ref([])
const burnupError = ref('')

const cfdDays = ref(90)
const cfdLoading = ref(false)
const cfdLabels = ref([])
const cfdSeries = ref([])

const releases = ref([])
const selectedReleaseId = ref(null)
const releaseBurndownLoading = ref(false)
const releaseBurndownData = ref([])
const releaseBurndownError = ref('')
const showReleaseManager = ref(false)
const showReleaseModal = ref(false)
const editingRelease = ref(null)
const releaseForm = ref({ name: '', goal: '', target_date: '' })

// ── Weekly throughput (kanban) ────────────────────────────────────────────────
const weeklyThroughput = ref([])
const weeklyThroughputLoading = ref(false)
const throughputWeeks = ref(12)

// ── Canvas refs ──────────────────────────────────────────────────────────────
const velocityCanvas = ref(null)
const throughputCanvas = ref(null)
const weeklyThroughputCanvas = ref(null)
const burndownCanvas = ref(null)
const burnupCanvas = ref(null)
const cfdCanvas = ref(null)
const releaseBurndownCanvas = ref(null)
const cycleTimeCanvas = ref(null)
const leadTimeCanvas = ref(null)

let charts = {}

const activeSprint = computed(() => sprintStore.activeSprint)
const sprintsWithDates = computed(() => sprintStore.sprints.filter(s => s.start_date && s.end_date))

function projectAvatar(project) { return resolveAssetUrl(project?.avatar || '') }
function cssVar(name) { return getComputedStyle(document.documentElement).getPropertyValue(name).trim() }

// ── Velocity + Throughput ────────────────────────────────────────────────────

async function loadVelocity() {
  velocityLoading.value = true
  try {
    const r = await projectsApi.getVelocityChart(slug.value)
    velocitySprints.value = r.data.sprints || []
  } catch { velocitySprints.value = [] }
  finally { velocityLoading.value = false }
}

function renderVelocity() {
  if (!velocityCanvas.value || !velocitySprints.value.length) return
  charts.velocity?.destroy()
  const primary = cssVar('--color-primary') || '#6366f1'
  const success = cssVar('--color-success') || '#22c55e'
  charts.velocity = new Chart(velocityCanvas.value, {
    type: 'bar',
    data: {
      labels: velocitySprints.value.map(s => s.name),
      datasets: [
        { label: t('sprint.total_points'), data: velocitySprints.value.map(s => s.total_points), backgroundColor: primary + '55', borderColor: primary, borderWidth: 1 },
        { label: t('sprint.completed_points'), data: velocitySprints.value.map(s => s.completed_points), backgroundColor: success + '88', borderColor: success, borderWidth: 1 },
      ],
    },
    options: { responsive: true, plugins: { legend: { position: 'top' } }, scales: { y: { beginAtZero: true, ticks: { stepSize: 1 } } } },
  })
}

function renderThroughput() {
  if (!throughputCanvas.value || !velocitySprints.value.length) return
  charts.throughput?.destroy()
  const warning = cssVar('--color-warning') || '#f59e0b'
  const success = cssVar('--color-success') || '#22c55e'
  charts.throughput = new Chart(throughputCanvas.value, {
    type: 'bar',
    data: {
      labels: velocitySprints.value.map(s => s.name),
      datasets: [
        { label: t('sprint.total_cards'), data: velocitySprints.value.map(s => s.total_cards), backgroundColor: warning + '55', borderColor: warning, borderWidth: 1 },
        { label: t('sprint.completed_cards'), data: velocitySprints.value.map(s => s.completed_cards), backgroundColor: success + '88', borderColor: success, borderWidth: 1 },
      ],
    },
    options: { responsive: true, plugins: { legend: { position: 'top' } }, scales: { y: { beginAtZero: true, ticks: { stepSize: 1 } } } },
  })
}

// ── Weekly Throughput (kanban) ────────────────────────────────────────────────

async function loadWeeklyThroughput() {
  weeklyThroughputLoading.value = true
  try {
    const r = await projectsApi.getThroughputChart(slug.value, throughputWeeks.value)
    weeklyThroughput.value = r.data.weeks || []
  } catch { weeklyThroughput.value = [] }
  finally { weeklyThroughputLoading.value = false }
  await nextTick(); renderWeeklyThroughput()
}

function renderWeeklyThroughput() {
  if (!weeklyThroughputCanvas.value || !weeklyThroughput.value.length) return
  charts.weeklyThroughput?.destroy()
  const primary = cssVar('--color-primary') || '#6366f1'
  charts.weeklyThroughput = new Chart(weeklyThroughputCanvas.value, {
    type: 'bar',
    data: {
      labels: weeklyThroughput.value.map(w => w.week_start),
      datasets: [{
        label: t('sprint.completed_cards'),
        data: weeklyThroughput.value.map(w => w.count),
        backgroundColor: primary + '88', borderColor: primary, borderWidth: 1,
      }],
    },
    options: {
      responsive: true,
      plugins: { legend: { display: false } },
      scales: { y: { beginAtZero: true, ticks: { stepSize: 1 }, title: { display: true, text: t('sprint.completed_cards') } } },
    },
  })
}

// ── Burndown ─────────────────────────────────────────────────────────────────

async function loadBurndown() {
  if (!selectedSprintId.value) return
  burndownLoading.value = true; burndownError.value = ''; burndownData.value = []
  try {
    const r = await projectsApi.getBurndownChart(slug.value, selectedSprintId.value)
    burndownData.value = r.data.data || []
  } catch (e) { burndownError.value = e.response?.data?.error || t('sprint.no_chart_data') }
  finally { burndownLoading.value = false }
  await nextTick(); renderBurndown()
}

function renderBurndown() {
  if (!burndownCanvas.value || !burndownData.value.length) return
  charts.burndown?.destroy()
  const danger = cssVar('--color-danger') || '#ef4444'
  const muted  = cssVar('--color-text-muted') || '#9ca3af'
  charts.burndown = new Chart(burndownCanvas.value, {
    type: 'line',
    data: {
      labels: burndownData.value.map(d => d.date),
      datasets: [
        { label: t('sprint.remaining_points'), data: burndownData.value.map(d => d.remaining), borderColor: danger, backgroundColor: danger + '22', fill: true, tension: 0.1, pointRadius: 3 },
        { label: t('sprint.ideal_line'), data: burndownData.value.map(d => d.ideal), borderColor: muted, borderDash: [6, 4], fill: false, tension: 0, pointRadius: 0 },
      ],
    },
    options: { responsive: true, plugins: { legend: { position: 'top' } }, scales: { y: { beginAtZero: true } } },
  })
}

// ── Burnup ────────────────────────────────────────────────────────────────────

async function loadBurnup() {
  if (!selectedBurnupSprintId.value) return
  burnupLoading.value = true; burnupError.value = ''; burnupData.value = []
  try {
    const r = await projectsApi.getBurnupChart(slug.value, selectedBurnupSprintId.value)
    burnupData.value = r.data.data || []
  } catch (e) { burnupError.value = e.response?.data?.error || t('sprint.no_chart_data') }
  finally { burnupLoading.value = false }
  await nextTick(); renderBurnup()
}

function renderBurnup() {
  if (!burnupCanvas.value || !burnupData.value.length) return
  charts.burnup?.destroy()
  const success = cssVar('--color-success') || '#22c55e'
  const primary = cssVar('--color-primary') || '#6366f1'
  charts.burnup = new Chart(burnupCanvas.value, {
    type: 'line',
    data: {
      labels: burnupData.value.map(d => d.date),
      datasets: [
        { label: t('sprint.completed_points'), data: burnupData.value.map(d => d.completed), borderColor: success, backgroundColor: success + '33', fill: true, tension: 0.1, pointRadius: 3 },
        { label: t('sprint.scope_line'), data: burnupData.value.map(d => d.total), borderColor: primary, borderDash: [6, 4], fill: false, tension: 0, pointRadius: 0 },
      ],
    },
    options: { responsive: true, plugins: { legend: { position: 'top' } }, scales: { y: { beginAtZero: true } } },
  })
}

// ── CFD ───────────────────────────────────────────────────────────────────────

async function loadCFD() {
  cfdLoading.value = true; cfdLabels.value = []; cfdSeries.value = []
  try {
    const r = await projectsApi.getCFDChart(slug.value, cfdDays.value)
    cfdLabels.value = r.data.labels || []
    cfdSeries.value = r.data.series || []
  } catch { cfdLabels.value = [] }
  finally { cfdLoading.value = false }
  await nextTick(); renderCFD()
}

const CFD_PALETTE = ['#6366f1','#22c55e','#f59e0b','#ef4444','#06b6d4','#8b5cf6','#ec4899','#14b8a6','#f97316','#84cc16']

function renderCFD() {
  if (!cfdCanvas.value || !cfdLabels.value.length) return
  charts.cfd?.destroy()
  const datasets = cfdSeries.value.map((col, i) => {
    const color = col.color || CFD_PALETTE[i % CFD_PALETTE.length]
    return { label: col.name, data: col.data, backgroundColor: color + 'cc', borderColor: color, borderWidth: 1, fill: true, tension: 0 }
  })
  charts.cfd = new Chart(cfdCanvas.value, {
    type: 'line',
    data: { labels: cfdLabels.value, datasets },
    options: {
      responsive: true,
      plugins: { legend: { position: 'top' } },
      scales: { y: { stacked: true, beginAtZero: true } },
      elements: { point: { radius: 0 } },
    },
  })
}

// ── Release Burndown + Release Management ─────────────────────────────────────

async function loadReleases() {
  try {
    const r = await projectsApi.listReleases(slug.value)
    releases.value = r.data || []
  } catch { releases.value = [] }
}

async function loadReleaseBurndown() {
  if (!selectedReleaseId.value) return
  releaseBurndownLoading.value = true; releaseBurndownError.value = ''; releaseBurndownData.value = []
  try {
    const r = await projectsApi.getReleaseBurndownChart(slug.value, selectedReleaseId.value)
    releaseBurndownData.value = r.data.data || []
  } catch (e) { releaseBurndownError.value = e.response?.data?.error || t('sprint.no_chart_data') }
  finally { releaseBurndownLoading.value = false }
  await nextTick(); renderReleaseBurndown()
}

function renderReleaseBurndown() {
  if (!releaseBurndownCanvas.value || !releaseBurndownData.value.length) return
  charts.releaseBurndown?.destroy()
  const danger = cssVar('--color-danger') || '#ef4444'
  const muted  = cssVar('--color-text-muted') || '#9ca3af'
  charts.releaseBurndown = new Chart(releaseBurndownCanvas.value, {
    type: 'line',
    data: {
      labels: releaseBurndownData.value.map(d => d.date),
      datasets: [
        { label: t('sprint.remaining_points'), data: releaseBurndownData.value.map(d => d.remaining), borderColor: danger, backgroundColor: danger + '22', fill: true, tension: 0.1, pointRadius: 3 },
        { label: t('sprint.ideal_line'), data: releaseBurndownData.value.map(d => d.ideal), borderColor: muted, borderDash: [6, 4], fill: false, tension: 0, pointRadius: 0 },
      ],
    },
    options: { responsive: true, plugins: { legend: { position: 'top' } }, scales: { y: { beginAtZero: true } } },
  })
}

function releaseHasSprint(release, sprintId) {
  return release.sprints?.some(s => s.id === sprintId)
}

async function toggleSprintInRelease(release, sprint) {
  try {
    if (releaseHasSprint(release, sprint.id)) {
      await projectsApi.removeSprintFromRelease(slug.value, release.id, sprint.id)
      release.sprints = release.sprints.filter(s => s.id !== sprint.id)
    } else {
      await projectsApi.addSprintToRelease(slug.value, release.id, sprint.id)
      release.sprints = [...(release.sprints || []), sprint]
    }
    if (selectedReleaseId.value === release.id) {
      releaseBurndownData.value = []; await loadReleaseBurndown()
    }
  } catch (e) { ui.error(e.response?.data?.error || 'Failed') }
}

function openCreateRelease() {
  editingRelease.value = null
  releaseForm.value = { name: '', goal: '', target_date: '' }
  showReleaseModal.value = true
}

function openEditRelease(release) {
  editingRelease.value = release
  releaseForm.value = {
    name: release.name,
    goal: release.goal || '',
    target_date: release.target_date ? release.target_date.slice(0, 10) : '',
  }
  showReleaseModal.value = true
}

function closeReleaseModal() { showReleaseModal.value = false; editingRelease.value = null }

async function saveRelease() {
  const data = {
    name: releaseForm.value.name,
    goal: releaseForm.value.goal,
    target_date: releaseForm.value.target_date || null,
  }
  try {
    if (editingRelease.value) {
      await projectsApi.updateRelease(slug.value, editingRelease.value.id, data)
    } else {
      await projectsApi.createRelease(slug.value, data)
    }
    await loadReleases()
    closeReleaseModal()
  } catch (e) { ui.error(e.response?.data?.error || 'Failed') }
}

async function confirmDeleteRelease(release) {
  if (!await ui.confirm(`${t('sprint.delete_release')} "${release.name}"?`)) return
  try {
    await projectsApi.deleteRelease(slug.value, release.id)
    releases.value = releases.value.filter(r => r.id !== release.id)
    if (selectedReleaseId.value === release.id) {
      selectedReleaseId.value = null; releaseBurndownData.value = []
    }
  } catch (e) { ui.error(e.response?.data?.error || 'Failed') }
}

// ── Cycle Time + Lead Time ────────────────────────────────────────────────────

async function loadCycleTime() {
  cycleTimeLoading.value = true
  try {
    const r = await projectsApi.getCycleTimeChart(slug.value)
    cycleTimeCards.value = r.data.cards || []
  } catch { cycleTimeCards.value = [] }
  finally { cycleTimeLoading.value = false }
}

function renderCycleTime() {
  if (!cycleTimeCanvas.value || !cycleTimeCards.value.length) return
  charts.cycleTime?.destroy()
  const primary = cssVar('--color-primary') || '#6366f1'
  const data = cycleTimeCards.value.map(c => ({ x: new Date(c.closed_at).getTime(), y: c.days_open, label: c.card_ref }))
  charts.cycleTime = new Chart(cycleTimeCanvas.value, {
    type: 'scatter',
    data: { datasets: [{ label: t('sprint.days_open'), data, backgroundColor: primary + '99', borderColor: primary, pointRadius: 5, pointHoverRadius: 7 }] },
    options: {
      responsive: true,
      plugins: {
        legend: { display: false },
        tooltip: { callbacks: { label: ctx => `${ctx.raw.label}: ${ctx.raw.y} days` } },
      },
      scales: {
        x: { type: 'linear', ticks: { callback: v => formatDate(new Date(v)) } },
        y: { beginAtZero: true, title: { display: true, text: t('sprint.days_open') } },
      },
    },
  })
}

function renderLeadTime() {
  if (!leadTimeCanvas.value || !cycleTimeCards.value.length) return
  charts.leadTime?.destroy()
  const buckets = [
    { label: '< 1d',   min: 0,   max: 1 },
    { label: '1-2d',   min: 1,   max: 2 },
    { label: '2-5d',   min: 2,   max: 5 },
    { label: '5-10d',  min: 5,   max: 10 },
    { label: '10-20d', min: 10,  max: 20 },
    { label: '20-30d', min: 20,  max: 30 },
    { label: '> 30d',  min: 30,  max: Infinity },
  ]
  const counts = buckets.map(b => cycleTimeCards.value.filter(c => c.days_open >= b.min && c.days_open < b.max).length)
  const primary = cssVar('--color-primary') || '#6366f1'
  charts.leadTime = new Chart(leadTimeCanvas.value, {
    type: 'bar',
    data: {
      labels: buckets.map(b => b.label),
      datasets: [{ label: t('sprint.count'), data: counts, backgroundColor: primary + '88', borderColor: primary, borderWidth: 1 }],
    },
    options: {
      responsive: true,
      plugins: { legend: { display: false } },
      scales: { y: { beginAtZero: true, ticks: { stepSize: 1 }, title: { display: true, text: t('sprint.count') } } },
    },
  })
}

// ── Tab switching ─────────────────────────────────────────────────────────────

watch(activeTab, async (tab) => {
  await nextTick()
  if (tab === 'velocity'         && velocitySprints.value.length)            renderVelocity()
  if (tab === 'throughput' && !isKanban.value && velocitySprints.value.length) renderThroughput()
  if (tab === 'throughput' && isKanban.value  && weeklyThroughput.value.length) renderWeeklyThroughput()
  if (tab === 'burndown'         && burndownData.value.length)               renderBurndown()
  if (tab === 'burnup'           && burnupData.value.length)                 renderBurnup()
  if (tab === 'cfd'              && cfdLabels.value.length)                  renderCFD()
  if (tab === 'release-burndown' && releaseBurndownData.value.length)        renderReleaseBurndown()
  if (tab === 'cycle-time'       && cycleTimeCards.value.length)             renderCycleTime()
  if (tab === 'lead-time'        && cycleTimeCards.value.length)             renderLeadTime()
  if (tab === 'cfd'              && !cfdLabels.value.length)                 loadCFD()
  if (tab === 'release-burndown' && !releases.value.length)                  loadReleases()
  if ((tab === 'cycle-time' || tab === 'lead-time') && !cycleTimeCards.value.length) loadCycleTime()
  if (tab === 'throughput' && isKanban.value && !weeklyThroughput.value.length) loadWeeklyThroughput()
})

watch(() => slug.value, () => {
  Object.values(charts).forEach(c => c?.destroy())
  charts = {}
  velocitySprints.value = []; burndownData.value = []; burnupData.value = []
  cfdLabels.value = []; cfdSeries.value = []; releases.value = []
  releaseBurndownData.value = []; cycleTimeCards.value = []; weeklyThroughput.value = []
  selectedSprintId.value = null; selectedBurnupSprintId.value = null
  selectedReleaseId.value = null
  init()
})

async function init() {
  await projectStore.fetchProject(slug.value)
  if (isKanban.value) {
    activeTab.value = 'cfd'
    await Promise.all([loadCFD(), loadCycleTime()])
    return
  }
  activeTab.value = 'velocity'
  await Promise.all([
    sprintStore.loadSprints(slug.value),
    loadVelocity(),
  ])
  await nextTick()
  renderVelocity()
}

onMounted(init)
onBeforeUnmount(() => Object.values(charts).forEach(c => c?.destroy()))
</script>

<style scoped>
.charts-layout { display: flex; flex-direction: column; height: 100%; }

.charts-body {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  max-width: 1000px;
  margin: 0 auto;
  width: 100%;
}

.charts-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--color-border);
}

.tab {
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  padding: 8px 14px;
  margin-bottom: -1px;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-muted);
  cursor: pointer;
  white-space: nowrap;
  transition: color .15s, border-color .15s;
}
.tab:hover { color: var(--color-text); }
.tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); }

.chart-panel h2 { font-size: 16px; font-weight: 600; margin-bottom: 16px; }

.chart-panel-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.chart-panel-header h2 { margin-bottom: 0; }

.sprint-select { width: auto; min-width: 180px; font-size: 13px; padding: 6px 10px; }

.chart-wrap {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 16px;
}

.chart-loading { display: flex; justify-content: center; padding: 48px; }
.chart-empty { padding: 48px; text-align: center; color: var(--color-text-muted); font-size: 14px; }

/* Release manager */
.release-manager {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 16px;
  margin-bottom: 16px;
}

.release-manager-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.release-empty { font-size: 13px; color: var(--color-text-muted); padding: 8px 0; }

.release-row {
  border-top: 1px solid var(--color-border);
  padding: 10px 0;
}

.release-row-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.release-row-name { font-weight: 600; font-size: 14px; flex: 1; }
.release-row-date { font-size: 12px; color: var(--color-text-muted); }

.btn.danger { color: var(--color-danger, #ef4444); }
.btn.danger:hover { background: color-mix(in srgb, var(--color-danger, #ef4444) 10%, transparent); }

.release-sprints-label { font-size: 12px; color: var(--color-text-muted); margin-bottom: 6px; }

.release-sprint-chips { display: flex; flex-wrap: wrap; gap: 6px; }

.sprint-chip {
  font-size: 12px;
  padding: 3px 10px;
  border-radius: 9999px;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: background .1s, color .1s, border-color .1s;
}
.sprint-chip:hover { border-color: var(--color-primary); color: var(--color-primary); }
.sprint-chip.assigned { background: var(--color-primary); border-color: var(--color-primary); color: #fff; }

/* Board toolbar */
.board-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 16px; height: 48px;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface); flex-shrink: 0; gap: 8px; flex-wrap: wrap;
}
.board-toolbar-left { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; }
.board-toolbar-right { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.board-project-name { font-size: 14px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.board-project-avatar { width: 20px; height: 20px; border-radius: 4px; object-fit: cover; flex-shrink: 0; }

.spinner {
  width: 32px; height: 32px;
  border: 3px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
