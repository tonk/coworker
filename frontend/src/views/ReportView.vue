<template>
  <div class="report-page">
    <!-- Filter panel (hidden when printing) -->
    <div class="report-filters no-print">
      <div class="filters-inner">
        <h1 class="filters-title">{{ $t('report.title') }}</h1>
        <div class="filters-row">
          <div class="filter-group">
            <label class="filter-label">{{ $t('report.period') }}</label>
            <select class="form-input" v-model="filters.period">
              <option value="all">{{ $t('report.period_all') }}</option>
              <option value="year">{{ $t('report.period_year') }}</option>
              <option value="month">{{ $t('report.period_month') }}</option>
              <option value="week">{{ $t('report.period_week') }}</option>
            </select>
          </div>

          <div class="filter-group" v-if="filters.period === 'year' || filters.period === 'month' || filters.period === 'week'">
            <label class="filter-label">{{ $t('report.year') }}</label>
            <select class="form-input" v-model.number="filters.year">
              <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}</option>
            </select>
          </div>

          <div class="filter-group" v-if="filters.period === 'month'">
            <label class="filter-label">{{ $t('report.month') }}</label>
            <select class="form-input" v-model.number="filters.month">
              <option v-for="(name, idx) in monthNames" :key="idx" :value="idx + 1">{{ name }}</option>
            </select>
          </div>

          <div class="filter-group" v-if="filters.period === 'week'">
            <label class="filter-label">{{ $t('report.week') }}</label>
            <select class="form-input" v-model.number="filters.week">
              <option v-for="w in 53" :key="w" :value="w">{{ $t('report.week') }} {{ w }}</option>
            </select>
          </div>

          <div class="filter-group">
            <label class="filter-label">{{ $t('report.project_filter') }}</label>
            <select class="form-input" v-model="filters.project">
              <option value="all">{{ $t('report.all_projects') }}</option>
              <option v-for="p in projects" :key="p.id" :value="p.slug">{{ p.name }}</option>
            </select>
          </div>

          <!-- Assignee multi-select -->
          <div class="filter-group assignee-filter-group" ref="assigneeDropdownRef">
            <label class="filter-label">{{ $t('report.assignee_filter') }}</label>
            <div class="assignee-select" @click="showAssigneeDropdown = !showAssigneeDropdown">
              <span class="assignee-select-label">
                <template v-if="filters.assignees.length === 0">{{ $t('report.all_assignees') }}</template>
                <template v-else>{{ selectedAssigneeNames }}</template>
              </span>
              <svg class="select-chevron" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
            </div>
            <div v-if="showAssigneeDropdown" class="assignee-dropdown">
              <label class="assignee-option">
                <input type="checkbox" :checked="filters.assignees.length === 0" @change="filters.assignees = []" />
                {{ $t('report.all_assignees') }}
              </label>
              <label v-for="u in allUsers" :key="u.id" class="assignee-option">
                <input type="checkbox" :value="u.id" v-model="filters.assignees" />
                {{ u.display_name || u.username }}
              </label>
            </div>
          </div>

          <div class="filter-group filter-actions">
            <button class="btn btn-primary" @click="loadReport" :disabled="loading">
              {{ loading ? $t('report.loading') : $t('report.run') }}
            </button>
          </div>
        </div>

        <div class="export-row" v-if="report">
          <div class="pdf-font-group">
            <label class="filter-label">{{ $t('report.pdf_font') }}</label>
            <select class="form-input" v-model="pdfFont">
              <option value="inter">Inter</option>
              <option value="roboto">Roboto</option>
              <option value="opensans">Open Sans</option>
              <option value="sourcecode">Source Code Pro</option>
              <option value="freesans">{{ $t('report.pdf_font_freesans') }}</option>
              <option value="freeserif">{{ $t('report.pdf_font_freeserif') }}</option>
              <option value="freemono">{{ $t('report.pdf_font_freemono') }}</option>
            </select>
          </div>
          <div class="pdf-font-group">
            <label class="filter-label">{{ $t('report.pdf_lang') }}</label>
            <select class="form-input" v-model="pdfLang">
              <option value="auto">{{ $t('report.pdf_lang_auto') }}</option>
              <option value="en">English</option>
              <option value="nl">Nederlands</option>
              <option value="de">Deutsch</option>
              <option value="fr">Français</option>
              <option value="es">Español</option>
              <option value="da">Dansk</option>
              <option value="sv">Svenska</option>
              <option value="nb">Norsk</option>
              <option value="fi">Suomi</option>
              <option value="is">Íslenska</option>
              <option value="pt">Português</option>
              <option value="it">Italiano</option>
            </select>
          </div>
          <button class="btn btn-secondary" @click="exportPDF">{{ $t('report.export_pdf') }}</button>
          <button class="btn btn-secondary" @click="exportXLSX">{{ $t('report.export_xlsx') }}</button>
        </div>
      </div>
    </div>

    <!-- Report content (visible on screen and when printing) -->
    <!-- Per-page header: hidden on screen, repeats on every printed page via position:fixed -->
    <div class="print-page-header" v-if="report">
      <img src="/logo.svg" alt="WarmDesk" class="print-logo" />
      <span class="print-app-name">WarmDesk</span>
    </div>

    <div class="report-content" v-if="report">
      <!-- Report Header -->
      <div class="report-header">
        <div class="report-header-left">
          <img v-if="report.company_logo" :src="resolveAssetUrl(report.company_logo)" alt="Logo" class="report-logo" @error="report.company_logo = ''" />
        </div>
        <div class="report-header-center">
          <div v-if="report.company_name" class="report-company-name">{{ report.company_name }}</div>
          <div class="report-title">{{ $t('report.title') }}</div>
          <div class="report-period-label">{{ report.period_label }}</div>
        </div>
        <div class="report-header-right">
          <div class="report-meta">{{ $t('report.generated_at') }}: {{ formatUTCTimestamp(report.generated_at) }}</div>
        </div>
      </div>

      <div v-if="report.projects.length === 0" class="report-empty">
        {{ $t('report.no_data') }}
      </div>

      <!-- Project sections -->
      <div v-for="proj in report.projects" :key="proj.project_id" class="report-project">
        <div class="project-header">
          <span class="project-name">{{ proj.project_name }}</span>
          <span class="project-total">{{ formatMinutes(proj.total_minutes) }}</span>
        </div>
        <table class="report-table">
          <thead>
            <tr>
              <th class="col-ref">{{ $t('report.col_ref') }}</th>
              <th class="col-title">{{ $t('report.col_title') }}</th>
              <th class="col-assignees">{{ $t('report.col_assignees') }}</th>
              <th class="col-updated">{{ $t('report.col_updated') }}</th>
              <th class="col-time">{{ $t('report.col_time') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="card in proj.cards" :key="card.card_id" :class="{ 'row-closed': card.closed }">
              <td class="col-ref">
                <span class="card-ref-badge" v-if="card.card_ref">{{ card.card_ref }}</span>
              </td>
              <td class="col-title">
                <span :class="{ 'title-closed': card.closed }">{{ card.title }}</span>
                <span v-if="card.closed" class="closed-badge">{{ $t('board.closed') }}</span>
              </td>
              <td class="col-assignees">{{ card.assignees.join(', ') || '—' }}</td>
              <td class="col-updated">{{ card.updated_at }}</td>
              <td class="col-time time-value">{{ formatMinutes(card.time_spent_minutes) }}</td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="subtotal-row">
              <td colspan="4" class="subtotal-label">{{ $t('report.subtotal') }}</td>
              <td class="col-time time-value">{{ formatMinutes(proj.total_minutes) }}</td>
            </tr>
          </tfoot>
        </table>
      </div>

      <!-- Grand Total -->
      <div class="grand-total-row" v-if="report.projects.length > 0">
        <span class="grand-total-label">{{ $t('report.grand_total') }}</span>
        <span class="grand-total-value">{{ formatMinutes(report.total_minutes) }}</span>
      </div>
    </div>

    <!-- Empty state before first load -->
    <div class="report-placeholder no-print" v-if="!report && !loading">
      <div class="placeholder-icon">📊</div>
      <p>{{ $t('report.run') }} to generate the report.</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { projectsApi } from '@/api/projects'
import { reportsApi } from '@/api/reports'
import { messagesApi } from '@/api/messages'
import { useDateFormat } from '@/composables/useDateFormat'
import { resolveAssetUrl } from '@/api/serverConfig'
import { useUIStore } from '@/stores/ui'

const { t, locale } = useI18n()
const { formatDateTime } = useDateFormat()
const ui = useUIStore()

const loading = ref(false)
const report = ref(null)
const pdfFont = ref(localStorage.getItem('report.pdfFont') || 'inter')
watch(pdfFont, v => localStorage.setItem('report.pdfFont', v))
const pdfLang = ref(localStorage.getItem('report.pdfLang') || 'auto')
watch(pdfLang, v => localStorage.setItem('report.pdfLang', v))
const projects = ref([])
const allUsers = ref([])
const showAssigneeDropdown = ref(false)
const assigneeDropdownRef = ref(null)

const now = new Date()
const filters = ref({
  period: 'month',
  year: now.getFullYear(),
  month: now.getMonth() + 1,
  week: getISOWeek(now),
  project: 'all',
  assignees: []
})

function getISOWeek(date) {
  const tmp = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()))
  tmp.setUTCDate(tmp.getUTCDate() + 4 - (tmp.getUTCDay() || 7))
  const yearStart = new Date(Date.UTC(tmp.getUTCFullYear(), 0, 1))
  return Math.ceil((((tmp - yearStart) / 86400000) + 1) / 7)
}

const yearOptions = computed(() => {
  const y = now.getFullYear()
  return Array.from({ length: 5 }, (_, i) => y - i)
})

const selectedAssigneeNames = computed(() => {
  if (!filters.value.assignees.length) return t('report.all_assignees')
  return filters.value.assignees
    .map(id => {
      const u = allUsers.value.find(u => u.id === id)
      return u ? (u.display_name || u.username) : id
    })
    .join(', ')
})

function onClickOutsideAssignee(e) {
  if (assigneeDropdownRef.value && !assigneeDropdownRef.value.contains(e.target)) {
    showAssigneeDropdown.value = false
  }
}

const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December']

// Backend returns generated_at as a UTC string "YYYY-MM-DD HH:mm" without timezone marker.
// Append 'Z' so the JS Date constructor treats it as UTC, then format via the user's setting.
function formatUTCTimestamp(utcStr) {
  if (!utcStr) return utcStr
  return formatDateTime(utcStr.replace(' ', 'T') + ':00Z')
}

function formatMinutes(minutes) {
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return `${h}:${String(m).padStart(2, '0')}`
}

async function loadReport() {
  loading.value = true
  showAssigneeDropdown.value = false
  try {
    const params = { period: filters.value.period }
    if (filters.value.period !== 'all') params.year = filters.value.year
    if (filters.value.period === 'month') params.month = filters.value.month
    if (filters.value.period === 'week') params.week = filters.value.week
    if (filters.value.project !== 'all') params.project = filters.value.project
    if (filters.value.assignees.length) params.assignees = filters.value.assignees.join(',')
    const { data } = await reportsApi.getTimeReport(params)
    report.value = data
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function exportPDF() {
  if (!report.value) return
  try {
    const params = { period: filters.value.period }
    if (filters.value.period !== 'all') params.year = filters.value.year
    if (filters.value.period === 'month') params.month = filters.value.month
    if (filters.value.period === 'week') params.week = filters.value.week
    if (filters.value.project !== 'all') params.project = filters.value.project
    if (filters.value.assignees.length) params.assignees = filters.value.assignees.join(',')
    params.font = pdfFont.value
    params.lang = pdfLang.value === 'auto' ? locale.value : pdfLang.value
    const { data } = await reportsApi.getTimeReportPDF(params)
    await triggerDownload(data, 'time-report.pdf', 'application/pdf')
  } catch (e) {
    console.error('[export] PDF failed:', e)
    ui.error(String(e?.message || e))
  }
}

async function exportXLSX() {
  if (!report.value) return
  try {
    const params = { period: filters.value.period }
    if (filters.value.period !== 'all') params.year = filters.value.year
    if (filters.value.period === 'month') params.month = filters.value.month
    if (filters.value.period === 'week') params.week = filters.value.week
    if (filters.value.project !== 'all') params.project = filters.value.project
    if (filters.value.assignees.length) params.assignees = filters.value.assignees.join(',')
    const { data } = await reportsApi.getTimeReportXLSX(params)
    const filename = `time-report-${report.value.period_label.replace(/\s+/g, '-').toLowerCase()}.xlsx`
    await triggerDownload(data, filename, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')
  } catch (e) {
    console.error('[export] report XLSX failed:', e)
    ui.error(String(e?.message || e))
  }
}

async function triggerDownload(data, filename, type) {
  if (window.__TAURI_INTERNALS__) {
    const { save } = await import('@tauri-apps/plugin-dialog')
    const { writeFile } = await import('@tauri-apps/plugin-fs')
    const ext = filename.split('.').pop()
    const path = await save({
      defaultPath: filename,
      filters: [{ name: ext.toUpperCase(), extensions: [ext] }]
    })
    if (!path) return
    const bytes = data instanceof ArrayBuffer ? new Uint8Array(data) : new Uint8Array(await new Blob([data]).arrayBuffer())
    await writeFile(path, bytes)
    return
  }
  const url = URL.createObjectURL(new Blob([data], { type }))
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(async () => {
  try {
    const [projRes, userRes] = await Promise.all([
      projectsApi.list(),
      messagesApi.listUsers()
    ])
    projects.value = (projRes.data || []).filter(p => !p.is_archived)
    allUsers.value = userRes.data || []
  } catch {}
  document.addEventListener('click', onClickOutsideAssignee)
})

onUnmounted(() => {
  document.removeEventListener('click', onClickOutsideAssignee)
})
</script>

<style scoped>
.report-page {
  min-height: 100vh;
  background: var(--color-bg);
  font-family: inherit;
}

/* Filters */
.report-filters {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  padding: 20px 0;
  position: relative;
  z-index: 10;
}
.filters-inner {
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 24px;
}
.filters-title {
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 16px;
  color: var(--color-text);
}
.filters-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-end;
}
.filter-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.filter-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.filter-actions { justify-content: flex-end; padding-top: 2px; }

.assignee-filter-group { position: relative; min-width: 160px; }
.assignee-select {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  cursor: pointer;
  font-size: 14px;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
}
.assignee-select:hover { border-color: var(--color-primary); }
.assignee-select-label { overflow: hidden; text-overflow: ellipsis; flex: 1; }
.select-chevron { flex-shrink: 0; color: var(--color-text-muted); }
.assignee-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 100%;
  max-height: 220px;
  overflow-y: auto;
  background: var(--color-surface-raised);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,.12);
  z-index: 100;
  padding: 4px 0;
}
.assignee-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 14px;
  font-size: 13px;
  cursor: pointer;
  user-select: none;
}
.assignee-option:hover { background: var(--color-surface-hover); }
.assignee-option input[type="checkbox"] { accent-color: var(--color-primary); cursor: pointer; }
.export-row {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}
.pdf-font-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

/* Report content */
.report-content {
  max-width: 1100px;
  margin: 0 auto;
  padding: 32px 24px;
}

.report-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 3px solid var(--color-primary);
}
.report-header-left {
  flex: 0 0 auto;
  min-width: 80px;
}
.report-logo {
  max-height: 64px;
  max-width: 180px;
  object-fit: contain;
}
.report-header-center {
  flex: 1;
  text-align: center;
}
.report-company-name {
  font-size: 22px;
  font-weight: 800;
  color: var(--color-text);
  letter-spacing: -0.02em;
}
.report-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-top: 4px;
}
.report-period-label {
  font-size: 20px;
  font-weight: 700;
  color: var(--color-primary);
  margin-top: 6px;
}
.report-header-right {
  flex: 0 0 auto;
  text-align: right;
}
.report-meta {
  font-size: 12px;
  color: var(--color-text-muted);
  margin-top: 4px;
}

/* Project sections */
.report-project {
  margin-bottom: 36px;
}
.project-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--color-primary);
  color: #fff;
  padding: 10px 16px;
  border-radius: var(--radius) var(--radius) 0 0;
}
.project-name {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.01em;
}
.project-total {
  font-size: 14px;
  font-weight: 700;
  background: rgba(255,255,255,0.2);
  padding: 2px 10px;
  border-radius: 9999px;
}

.report-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  border: 1px solid var(--color-border);
  border-top: none;
}
.report-table th {
  background: var(--color-surface);
  color: var(--color-text-muted);
  font-weight: 600;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 8px 12px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}
.report-table td {
  padding: 9px 12px;
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text);
  vertical-align: middle;
}
.report-table tbody tr:last-child td {
  border-bottom: none;
}
.report-table tbody tr:hover {
  background: var(--color-bg);
}

.col-ref { width: 80px; }
.col-time { width: 100px; text-align: right; }
.col-updated { width: 100px; }
.col-assignees { width: 160px; }
.time-value { font-weight: 700; color: var(--color-primary); }

.card-ref-badge {
  font-size: 11px;
  font-weight: 700;
  color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-primary) 25%, transparent);
  border-radius: 4px;
  padding: 1px 5px;
}

.title-closed { text-decoration: line-through; color: var(--color-text-muted); }
.closed-badge {
  display: inline-block;
  margin-left: 6px;
  font-size: 10px;
  font-weight: 700;
  color: #dc2626;
  background: color-mix(in srgb, #ef4444 12%, transparent);
  border: 1px solid color-mix(in srgb, #ef4444 30%, transparent);
  border-radius: 4px;
  padding: 1px 5px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  vertical-align: middle;
}

.subtotal-row td {
  background: color-mix(in srgb, var(--color-primary) 5%, var(--color-surface));
  font-weight: 700;
  border-top: 2px solid var(--color-border);
  border-bottom: none;
}
.subtotal-label { text-align: right; color: var(--color-text-muted); font-size: 12px; text-transform: uppercase; letter-spacing: 0.06em; }

.grand-total-row {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 24px;
  margin-top: 20px;
  padding: 14px 20px;
  background: var(--color-surface);
  border: 2px solid var(--color-primary);
  border-radius: var(--radius);
}
.grand-total-label {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}
.grand-total-value {
  font-size: 22px;
  font-weight: 800;
  color: var(--color-primary);
}

.report-empty {
  text-align: center;
  padding: 48px;
  color: var(--color-text-muted);
  font-size: 15px;
}

.report-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 24px;
  color: var(--color-text-muted);
}
.placeholder-icon { font-size: 48px; margin-bottom: 16px; }
.report-placeholder p { font-size: 15px; }

/* Per-page print header — hidden on screen */
.print-page-header { display: none; }

/* Print styles */
@page {
  margin: 14mm 1cm 12mm 1cm;
  size: auto;
  /* App name top-left on pages 2+, page number top-right on all pages */
  @top-left {
    content: "WarmDesk";
    font-size: 11pt;
    font-weight: 700;
    color: #6366f1;
    vertical-align: middle;
  }
  @top-center  { content: ""; }
  @top-right {
    content: counter(page) " / " counter(pages);
    font-size: 9pt;
    color: #64748b;
    vertical-align: middle;
  }
  @bottom-left   { content: ""; }
  @bottom-center { content: ""; }
  @bottom-right  { content: ""; }
}

/* Page 1: logo banner replaces the margin-box "WarmDesk" text */
@page :first {
  @top-left  { content: ""; }
}

@media print {
  /* Logo banner at top of page 1 — in normal flow, not fixed */
  .print-page-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding-bottom: 4mm;
    margin-bottom: 6mm;
    border-bottom: 2px solid #6366f1;
    -webkit-print-color-adjust: exact;
    print-color-adjust: exact;
  }
  .print-logo { height: 26px; width: auto; }
  .print-app-name { font-size: 13pt; font-weight: 700; color: #6366f1; letter-spacing: 0.03em; }

  /* Hide everything outside the report content */
  :global(.app-shell-header),
  :global(.app-sidebar),
  :global(.app-footer),
  .no-print { display: none !important; }

  /* Make the shell fill the page without sidebar layout */
  :global(.app-shell-body) {
    display: block !important;
    overflow: visible !important;
    height: auto !important;
    min-height: 0 !important;
  }
  :global(.app-shell-content) {
    overflow: visible !important;
    height: auto !important;
    min-height: 0 !important;
  }

  .report-page { background: #fff; }
  .report-content { max-width: 100%; padding: 1cm; margin: 0; }
  .report-header { border-bottom: 3px solid #6366f1; }
  .report-company-name { color: #1e293b; }
  .report-period-label { color: #6366f1; }
  .project-header { background: #6366f1; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  .report-table th { background: #f8fafc; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  .subtotal-row td { background: #f0f4ff; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  .report-project { break-inside: auto; }
  .project-header { break-after: avoid; }
  .grand-total-row { border: 2px solid #6366f1; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  .time-value { color: #6366f1; }
  .card-ref-badge { color: #6366f1; border-color: #6366f1; }
}
</style>
