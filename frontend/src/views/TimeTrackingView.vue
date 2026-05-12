<template>
  <div class="tt-view">
    <h1 class="sr-only">{{ $t('timeTracking.nav') }}</h1>

    <!-- ── Top bar ─────────────────────────────────────────────────────────── -->
    <div class="tt-bar">
      <div class="tt-employee">
        <span class="tt-emp-label">{{ $t('timeTracking.employee') }}</span>
        <select v-if="canViewOtherUsers" class="tt-user-select" v-model="selectedUserId" @change="onUserChange">
          <option :value="0">{{ $t('timeTracking.all_employees') }}</option>
          <option v-for="u in allUsers" :key="u.id" :value="u.id">{{ userName(u) }}</option>
        </select>
        <strong v-else class="tt-emp-name">{{ displayName }}</strong>
      </div>

      <div class="tt-week-nav">
        <button class="nav-btn" @click="shiftWeek(-1)" :title="$t('timeTracking.prev_week')">&#9664;</button>
        <span class="wk-label">{{ $t('timeTracking.week') }} &nbsp; {{ weekInfo.week }} &nbsp; {{ weekInfo.year }}</span>
        <button class="nav-btn" @click="shiftWeek(1)" :title="$t('timeTracking.next_week')">&#9654;</button>
      </div>

      <div class="tt-mode-tabs">
        <button class="tt-mode-btn" :class="{ active: mode === 'sheet' }" @click="mode = 'sheet'">
          {{ $t('timeTracking.tab_log') }}
        </button>
        <button class="tt-mode-btn" :class="{ active: mode === 'report' }" @click="mode = 'report'">
          {{ $t('timeTracking.tab_report') }}
        </button>
      </div>
    </div>

    <!-- ── Weekly timesheet ────────────────────────────────────────────────── -->
    <div v-show="mode === 'sheet'" class="tt-sheet-outer">
      <div v-if="loading" class="tt-loading">{{ $t('common.loading') }}</div>
      <div v-else class="tt-scroll">
        <table class="tt-table">
          <thead>
            <tr class="tt-head">
              <th class="c-nr"></th>
              <th class="c-info">
                <span>{{ $t('timeTracking.customer') }}</span>
                <span class="sub">{{ $t('timeTracking.project') }}</span>
              </th>
              <th class="c-desc">{{ $t('timeTracking.activity') }}</th>
              <th v-for="d in weekDays" :key="d.iso" class="c-day">
                <div class="dh-abbr">{{ d.abbr }}</div>
                <div class="dh-date">{{ d.mmdd }}</div>
              </th>
              <th class="c-total">{{ $t('timeTracking.total') }}</th>
              <th class="c-act"></th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="(row, idx) in allRows" :key="row.key"
                :class="['tt-row', idx % 2 === 1 ? 'alt' : '', deletingRow === row.key ? 'tt-row-deleting' : '']">
              <td class="c-nr">{{ idx + 1 }}</td>

              <!-- Edit mode -->
              <template v-if="editingRow === row.key">
                <td class="c-info tt-editing">
                  <select class="nr-sel" v-model="editForm.customer_id" @change="editForm.project_id = null">
                    <option :value="null">{{ $t('timeTracking.no_customer') }}</option>
                    <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
                  </select>
                  <select class="nr-sel" v-model="editForm.project_id">
                    <option :value="null">{{ $t('timeTracking.no_project') }}</option>
                    <option v-for="p in editRowProjects" :key="p.id" :value="p.id">{{ p.name }}</option>
                  </select>
                </td>
                <td class="c-desc tt-editing">
                  <input class="nr-desc" type="text"
                    v-model="editForm.description"
                    :placeholder="$t('timeTracking.description')"
                    @keydown.enter="confirmEditRow(row)"
                    @keydown.escape="cancelEditRow"
                  />
                </td>
              </template>

              <!-- Normal mode -->
              <template v-else>
                <td class="c-info">
                  <div class="rc-cust">{{ row.customer_name || '—' }}</div>
                  <div class="rc-proj">{{ row.project_name || '—' }}</div>
                </td>
                <td class="c-desc">{{ row.description || '—' }}</td>
              </template>

              <td v-for="d in weekDays" :key="d.iso" class="c-day">
                <input
                  type="number"
                  class="h-inp"
                  :class="{ 'h-inp-filled': !!cellVal(row, d.iso) }"
                  step="0.25"
                  min="0"
                  :placeholder="savingCell === row.key + d.iso ? '…' : ''"
                  :value="cellVal(row, d.iso)"
                  :disabled="viewingOther || savingCell === row.key + d.iso || editingRow === row.key"
                  @focus="$event.target.select()"
                  @blur="onCellBlur(row, d.iso, $event.target.value)"
                  @keydown.enter="$event.target.blur()"
                />
              </td>
              <td class="c-total c-rowtotal">{{ rowTotal(row) }}</td>

              <!-- Actions -->
              <td class="c-act">
                <template v-if="!viewingOther">
                  <template v-if="editingRow === row.key">
                    <button class="act-btn act-ok" @click="confirmEditRow(row)" :title="$t('common.save')">✓</button>
                    <button class="act-btn act-no" @click="cancelEditRow" :title="$t('common.cancel')">✕</button>
                  </template>
                  <template v-else-if="deletingRow === row.key">
                    <button class="act-btn act-ok" @click="confirmDeleteRow(row)" :title="$t('common.yes')">✓</button>
                    <button class="act-btn act-no" @click="cancelDeleteRow" :title="$t('common.no')">✕</button>
                  </template>
                  <template v-else>
                    <button class="act-btn act-edit" @click="startEditRow(row)" :title="$t('common.edit')">✎</button>
                    <button class="act-btn act-del" @click="startDeleteRow(row)" :title="$t('common.delete')">🗑</button>
                  </template>
                </template>
              </td>
            </tr>

            <!-- Inline new-row editor -->
            <tr v-if="addingRow" class="tt-row tt-newrow">
              <td class="c-nr">{{ allRows.length + 1 }}</td>
              <td class="c-info">
                <select class="nr-sel" v-model="newRow.customer_id" @change="newRow.project_id = null">
                  <option :value="null">{{ $t('timeTracking.no_customer') }}</option>
                  <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
                <select class="nr-sel" v-model="newRow.project_id">
                  <option :value="null">{{ $t('timeTracking.no_project') }}</option>
                  <option v-for="p in newRowProjects" :key="p.id" :value="p.id">{{ p.name }}</option>
                </select>
              </td>
              <td class="c-desc">
                <input ref="newDescRef" class="nr-desc" type="text"
                  v-model="newRow.description"
                  :placeholder="$t('timeTracking.description')"
                  @keydown.enter="confirmNewRow"
                  @keydown.escape="cancelNewRow"
                />
              </td>
              <td v-for="d in weekDays" :key="d.iso" class="c-day">
                <input class="h-inp" type="number" step="0.25" min="0" value="" disabled />
              </td>
              <td class="c-total"></td>
              <td class="c-act"></td>
            </tr>
          </tbody>

          <tfoot>
            <tr class="tt-foot">
              <td colspan="3" class="foot-lbl">{{ $t('timeTracking.total') }}</td>
              <td v-for="d in weekDays" :key="d.iso" class="c-day c-total c-dttotal">
                {{ dayTotal(d.iso) }}
              </td>
              <td class="c-total grand-total">{{ grandTotal }}</td>
              <td class="c-act"></td>
            </tr>
          </tfoot>
        </table>
      </div>

      <!-- Bottom bar: add row + exports -->
      <div class="tt-add-bar">
        <template v-if="!viewingOther && !addingRow">
          <button class="btn-add-row" @click="startAddRow">
            ＋ {{ $t('timeTracking.add_row') }}
          </button>
        </template>
        <template v-else-if="!viewingOther">
          <button class="btn btn-primary" @click="confirmNewRow">{{ $t('common.save') }}</button>
          <button class="btn btn-secondary" @click="cancelNewRow">{{ $t('common.cancel') }}</button>
        </template>
        <div class="tt-export-group" v-if="allRows.length > 0">
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
          <button class="btn btn-secondary" @click="exportSheetXLSX">{{ $t('timeTracking.export_xlsx') }}</button>
          <button class="btn btn-secondary" @click="exportSheetPDF">{{ $t('timeTracking.export_pdf') }}</button>
        </div>
      </div>
    </div>

    <!-- ── Report ──────────────────────────────────────────────────────────── -->
    <div v-if="mode === 'report'" class="tt-report-outer">
      <div class="report-filters">
        <select class="form-input fi-sm" v-model="rpt.period" @change="loadReport">
          <option value="week">{{ $t('report.week') }}</option>
          <option value="month">{{ $t('report.month') }}</option>
          <option value="year">{{ $t('report.year') }}</option>
        </select>
        <input type="number" class="form-input fi-sm fi-year"
          v-model.number="rpt.year" min="2000" max="2100" @change="loadReport" />
        <select v-if="rpt.period === 'month'" class="form-input fi-sm"
          v-model.number="rpt.month" @change="loadReport">
          <option v-for="(m, i) in months" :key="i" :value="i + 1">{{ m }}</option>
        </select>
        <select v-if="rpt.period === 'week'" class="form-input fi-sm"
          v-model.number="rpt.week" @change="loadReport">
          <option v-for="w in 53" :key="w" :value="w">
            {{ $t('timeTracking.week') }} {{ w }}
          </option>
        </select>
        <button class="btn btn-secondary" @click="loadReport">{{ $t('timeTracking.refresh') }}</button>
        <div class="tt-export-group" v-if="report && report.total_minutes > 0">
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
          <button class="btn btn-secondary" @click="exportReportXLSX">{{ $t('timeTracking.export_xlsx') }}</button>
          <button class="btn btn-secondary" @click="exportReportPDF">{{ $t('timeTracking.export_pdf') }}</button>
        </div>
      </div>

      <div v-if="loadingReport" class="tt-loading">{{ $t('common.loading') }}</div>
      <template v-else-if="report">
        <!-- Report header: logo + company name -->
        <div class="rpt-header">
          <div class="rpt-header-left">
            <img v-if="report.company_logo" :src="resolveAssetUrl(report.company_logo)" alt="" class="rpt-logo" @error="report.company_logo = ''" />
          </div>
          <div class="rpt-header-center">
            <div v-if="report.company_name" class="rpt-company-name">{{ report.company_name }}</div>
            <div class="rpt-title-label">{{ $t('timeTracking.title') }}</div>
            <div class="rpt-period-label">{{ report.period_label }}</div>
          </div>
          <div class="rpt-header-right"></div>
        </div>
        <div v-if="report.groups.length === 0" class="rpt-empty">{{ $t('timeTracking.no_entries') }}</div>
        <div v-for="grp in report.groups" :key="grp.label" class="rpt-group">
          <div class="rpt-group-hd">
            <span>{{ grp.label }}</span>
            <span class="rpt-grp-total">{{ fmtDecimal(grp.total_minutes) }}</span>
          </div>
          <table v-if="grp.entries.length" class="rpt-table">
            <colgroup>
              <col class="rpt-col-date" />
              <col class="rpt-col-customer" />
              <col class="rpt-col-project" />
              <col class="rpt-col-activity" />
              <col class="rpt-col-time" />
            </colgroup>
            <thead>
              <tr>
                <th>{{ $t('timeTracking.date') }}</th>
                <th>{{ $t('timeTracking.customer') }}</th>
                <th>{{ $t('timeTracking.project') }}</th>
                <th>{{ $t('timeTracking.activity') }}</th>
                <th class="rpt-th-time">{{ $t('timeTracking.time') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="e in grp.entries" :key="e.id">
                <td>{{ e.date.slice(0,10) }}</td>
                <td>{{ e.customer?.name || '—' }}</td>
                <td>{{ e.project?.name || '—' }}</td>
                <td>{{ e.description || '—' }}</td>
                <td class="rpt-th-time">{{ fmtDecimal(e.minutes) }}</td>
              </tr>
            </tbody>
          </table>
          <div v-else class="rpt-grp-empty">{{ $t('timeTracking.no_entries_group') }}</div>
        </div>
        <div class="rpt-grand-total">
          <span>{{ $t('timeTracking.total') }}</span>
          <span>{{ fmtDecimal(report.total_minutes) }}</span>
        </div>
      </template>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { timeEntriesApi } from '@/api/timeEntries'
import { customersApi } from '@/api/customers'
import { projectsApi } from '@/api/projects'
import client from '@/api/client'
import { resolveAssetUrl } from '@/api/serverConfig'

const { t } = useI18n()
const auth = useAuthStore()
const ui = useUIStore()

// ── Shared lookup data ────────────────────────────────────────────────────
const customers = ref([])
const projects  = ref([])

// ── User switching (admin / time_tracking_viewer only) ────────────────────
const canViewOtherUsers = computed(() => auth.isAdmin || !!auth.user?.time_tracking_viewer)
const allUsers = ref([])
const selectedUserId = ref(auth.user?.id ?? null)
const viewingOther = computed(() => !!selectedUserId.value && selectedUserId.value !== auth.user?.id)

function userName(u) {
  return [u.first_name, u.last_name].filter(Boolean).join(' ') || u.display_name || u.username
}

const displayName = computed(() => {
  if (canViewOtherUsers.value) {
    if (selectedUserId.value === 0) return t('timeTracking.all_employees')
    const u = allUsers.value.find(u => u.id === selectedUserId.value) || auth.user
    return u ? userName(u) : ''
  }
  return auth.user ? userName(auth.user) : ''
})

function onUserChange() {
  localRows.value = []
  editingRow.value = null
  deletingRow.value = null
  addingRow.value = false
  loadWeek()
  loadReport()
}

// ── PDF export options ────────────────────────────────────────────────────
const pdfFont = ref(localStorage.getItem('timeTracking.pdfFont') || 'inter')
watch(pdfFont, v => localStorage.setItem('timeTracking.pdfFont', v))
const pdfLang = ref(localStorage.getItem('timeTracking.pdfLang') || 'auto')
watch(pdfLang, v => localStorage.setItem('timeTracking.pdfLang', v))

// ── Mode ──────────────────────────────────────────────────────────────────
const mode = ref('sheet')

// ── Week navigation ───────────────────────────────────────────────────────
const anchor = ref(new Date())

const weekStart = computed(() => {
  const d = new Date(anchor.value)
  d.setHours(0, 0, 0, 0)
  const day = d.getDay()
  d.setDate(d.getDate() + (day === 0 ? -6 : 1 - day))
  return d
})

const weekInfo = computed(() => {
  const s = weekStart.value
  const d = new Date(Date.UTC(s.getFullYear(), s.getMonth(), s.getDate()))
  d.setUTCDate(d.getUTCDate() + 4 - (d.getUTCDay() || 7))
  const y0 = new Date(Date.UTC(d.getUTCFullYear(), 0, 1))
  return { year: d.getUTCFullYear(), week: Math.ceil(((d - y0) / 86400000 + 1) / 7) }
})

const weekDays = computed(() => {
  const abbr = new Intl.DateTimeFormat(undefined, { weekday: 'short' })
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(weekStart.value)
    d.setDate(d.getDate() + i)
    const iso = d.toISOString().slice(0, 10)
    return { iso, mmdd: iso.slice(5), abbr: abbr.format(d) }
  })
})

function shiftWeek(delta) {
  const d = new Date(anchor.value)
  d.setDate(d.getDate() + delta * 7)
  anchor.value = d
  loadWeek()
}

// ── Entry state ───────────────────────────────────────────────────────────
const rawEntries = ref([])  // TimeEntry[] from API for this week
const localRows  = ref([])  // rows added locally but with no entries yet
const loading    = ref(false)
const savingCell = ref('')  // "rowKey+dateISO" while a save is in flight

function rowKey(customerId, projectId, description) {
  return `${customerId ?? ''}|${projectId ?? ''}|${description ?? ''}`
}

// Rows derived from existing entries this week
const entryRows = computed(() => {
  const seen = new Map()
  for (const e of rawEntries.value) {
    const k = rowKey(e.customer_id, e.project_id, e.description)
    if (!seen.has(k)) {
      seen.set(k, {
        key:           k,
        customer_id:   e.customer_id,
        customer_name: e.customer?.name || '',
        project_id:    e.project_id,
        project_name:  e.project?.name || '',
        description:   e.description || '',
      })
    }
  }
  return [...seen.values()]
})

// All rows = entry-derived rows + locally-added rows not yet in entries
const allRows = computed(() => {
  const keys = new Set(entryRows.value.map(r => r.key))
  const extras = localRows.value.filter(r => !keys.has(r.key))
  return [...entryRows.value, ...extras]
})

function getEntry(row, dateISO) {
  return rawEntries.value.find(
    e => rowKey(e.customer_id, e.project_id, e.description) === row.key
      && e.date.slice(0, 10) === dateISO
  ) ?? null
}

function cellVal(row, dateISO) {
  const e = getEntry(row, dateISO)
  return e ? fmtDecimal(e.minutes) : ''
}

function rowTotal(row) {
  const m = weekDays.value.reduce((s, d) => s + (getEntry(row, d.iso)?.minutes || 0), 0)
  return m ? fmtDecimal(m) : '0.00'
}

function dayTotal(dateISO) {
  const m = allRows.value.reduce((s, row) => s + (getEntry(row, dateISO)?.minutes || 0), 0)
  return m ? fmtDecimal(m) : '0.00'
}

const grandTotal = computed(() => {
  const m = rawEntries.value.reduce((s, e) => s + e.minutes, 0)
  return fmtDecimal(m)
})

// ── Load week ─────────────────────────────────────────────────────────────
async function loadWeek() {
  loading.value = true
  try {
    const from = weekDays.value[0].iso
    const to   = weekDays.value[6].iso
    const params = { from, to }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const { data } = await timeEntriesApi.list(params)
    rawEntries.value = data
    // Drop any local rows that now have entries
    const keys = new Set(data.map(e => rowKey(e.customer_id, e.project_id, e.description)))
    localRows.value = localRows.value.filter(r => !keys.has(r.key))
  } catch {
    ui.error(t('timeTracking.load_error'))
  } finally {
    loading.value = false
  }
}

// ── Cell save ─────────────────────────────────────────────────────────────
async function onCellBlur(row, dateISO, rawVal) {
  const hours   = parseFloat(rawVal) || 0
  const minutes = Math.round(hours * 60)
  const existing = getEntry(row, dateISO)

  if (minutes === (existing?.minutes || 0)) return   // no change

  const ck = row.key + dateISO
  savingCell.value = ck

  try {
    if (minutes === 0 && existing) {
      await timeEntriesApi.remove(existing.id)
      rawEntries.value = rawEntries.value.filter(e => e.id !== existing.id)
    } else if (minutes > 0 && existing) {
      const { data } = await timeEntriesApi.update(existing.id, {
        customer_id:  row.customer_id  || null,
        project_id:   row.project_id   || null,
        date:         dateISO,
        minutes,
        description:  row.description,
      })
      const idx = rawEntries.value.findIndex(e => e.id === existing.id)
      rawEntries.value[idx] = data
    } else if (minutes > 0) {
      const { data } = await timeEntriesApi.create({
        customer_id:  row.customer_id  || null,
        project_id:   row.project_id   || null,
        date:         dateISO,
        minutes,
        description:  row.description,
      })
      rawEntries.value.push(data)
      // Row is now in entries — remove from localRows
      localRows.value = localRows.value.filter(r => r.key !== row.key)
    }
  } catch {
    ui.error(t('timeTracking.save_error'))
  } finally {
    if (savingCell.value === ck) savingCell.value = ''
  }
}

// ── Edit row ──────────────────────────────────────────────────────────────
const editingRow = ref(null)
const editForm   = ref({ customer_id: null, project_id: null, description: '' })
const deletingRow = ref(null)

const editRowProjects = computed(() => {
  if (!editForm.value.customer_id) return projects.value
  return projects.value.filter(p => p.customer_id === editForm.value.customer_id)
})

function startEditRow(row) {
  cancelDeleteRow()
  editForm.value = { customer_id: row.customer_id, project_id: row.project_id, description: row.description }
  editingRow.value = row.key
}

function cancelEditRow() {
  editingRow.value = null
}

async function confirmEditRow(row) {
  const r = editForm.value
  const cust = customers.value.find(c => c.id === r.customer_id)
  const proj = projects.value.find(p => p.id === r.project_id)
  const newKey = rowKey(r.customer_id, r.project_id, r.description)

  const toUpdate = rawEntries.value.filter(
    e => rowKey(e.customer_id, e.project_id, e.description) === row.key
  )
  try {
    for (const e of toUpdate) {
      const { data } = await timeEntriesApi.update(e.id, {
        customer_id: r.customer_id || null,
        project_id:  r.project_id  || null,
        date:        e.date.slice(0, 10),
        minutes:     e.minutes,
        description: r.description,
      })
      const idx = rawEntries.value.findIndex(x => x.id === e.id)
      rawEntries.value[idx] = data
    }
    const li = localRows.value.findIndex(x => x.key === row.key)
    if (li >= 0) {
      localRows.value[li] = {
        ...localRows.value[li],
        key:           newKey,
        customer_id:   r.customer_id,
        customer_name: cust?.name || '',
        project_id:    r.project_id,
        project_name:  proj?.name || '',
        description:   r.description,
      }
    }
  } catch {
    ui.error(t('timeTracking.save_error'))
  } finally {
    editingRow.value = null
  }
}

function startDeleteRow(row) {
  cancelEditRow()
  deletingRow.value = row.key
}

function cancelDeleteRow() {
  deletingRow.value = null
}

async function confirmDeleteRow(row) {
  const toDelete = rawEntries.value.filter(
    e => rowKey(e.customer_id, e.project_id, e.description) === row.key
  )
  try {
    for (const e of toDelete) {
      await timeEntriesApi.remove(e.id)
    }
    rawEntries.value = rawEntries.value.filter(
      e => rowKey(e.customer_id, e.project_id, e.description) !== row.key
    )
    localRows.value = localRows.value.filter(r => r.key !== row.key)
  } catch {
    ui.error(t('timeTracking.delete_error'))
  } finally {
    deletingRow.value = null
  }
}

// ── Add row ───────────────────────────────────────────────────────────────
const addingRow   = ref(false)
const newDescRef  = ref(null)
const newRow      = ref({ customer_id: null, project_id: null, description: '' })

const newRowProjects = computed(() => {
  if (!newRow.value.customer_id) return projects.value
  return projects.value.filter(p => p.customer_id === newRow.value.customer_id)
})

function startAddRow() {
  newRow.value = { customer_id: null, project_id: null, description: '' }
  addingRow.value = true
  nextTick(() => newDescRef.value?.focus())
}

function confirmNewRow() {
  const r = newRow.value
  const cust = customers.value.find(c => c.id === r.customer_id)
  const proj = projects.value.find(p => p.id === r.project_id)
  const k = rowKey(r.customer_id, r.project_id, r.description)
  if (!allRows.value.find(x => x.key === k)) {
    localRows.value.push({
      key:           k,
      customer_id:   r.customer_id,
      customer_name: cust?.name || '',
      project_id:    r.project_id,
      project_name:  proj?.name || '',
      description:   r.description,
    })
  }
  addingRow.value = false
}

function cancelNewRow() {
  addingRow.value = false
}

// ── Report ────────────────────────────────────────────────────────────────
const now  = new Date()
const rpt  = ref({ period: 'month', year: now.getFullYear(), month: now.getMonth() + 1, week: currentISOWeek() })
const report       = ref(null)
const loadingReport = ref(false)

const months = computed(() => {
  const fmt = new Intl.DateTimeFormat(undefined, { month: 'long' })
  return Array.from({ length: 12 }, (_, i) => fmt.format(new Date(2000, i, 1)))
})

function currentISOWeek() {
  const d = new Date(Date.UTC(now.getFullYear(), now.getMonth(), now.getDate()))
  d.setUTCDate(d.getUTCDate() + 4 - (d.getUTCDay() || 7))
  const y0 = new Date(Date.UTC(d.getUTCFullYear(), 0, 1))
  return Math.ceil(((d - y0) / 86400000 + 1) / 7)
}

async function loadReport() {
  loadingReport.value = true
  try {
    const params = {
      period: rpt.value.period,
      year:   rpt.value.year,
      month:  rpt.value.month,
      week:   rpt.value.week,
    }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const { data } = await timeEntriesApi.report(params)
    report.value = data
  } catch {
    ui.error(t('timeTracking.load_error'))
  } finally {
    loadingReport.value = false
  }
}

// ── Formatting ────────────────────────────────────────────────────────────
function fmtDecimal(minutes) {
  if (!minutes) return '0.00'
  return (minutes / 60).toFixed(2)
}

// ── Exports ───────────────────────────────────────────────────────────────

// Weekly timesheet → XLSX: rows are customer/project combos, columns are days.
async function exportSheetXLSX() {
  try {
    const params = { year: weekInfo.value.year, week: weekInfo.value.week }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const { data } = await timeEntriesApi.sheetXLSX(params)
    triggerDownload(data, `time-tracking-week${weekInfo.value.week}-${weekInfo.value.year}.xlsx`, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')
  } catch {
    ui.error(t('timeTracking.export_error'))
  }
}

// Weekly timesheet → PDF: delegates to backend report with period=week.
async function exportSheetPDF() {
  try {
    const params = { period: 'week', year: weekInfo.value.year, week: weekInfo.value.week, font: pdfFont.value, lang: pdfLang.value }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const { data } = await timeEntriesApi.reportPDF(params)
    triggerDownload(data, `time-tracking-week${weekInfo.value.week}-${weekInfo.value.year}.pdf`, 'application/pdf')
  } catch {
    ui.error(t('timeTracking.export_error'))
  }
}

// Report tab → XLSX: date-list grouped by period.
async function exportReportXLSX() {
  if (!report.value) return
  try {
    const params = {
      period: rpt.value.period,
      year:   rpt.value.year,
      month:  rpt.value.month,
      week:   rpt.value.week,
    }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const { data } = await timeEntriesApi.reportXLSX(params)
    const slug = report.value.period_label.replace(/\s+/g, '-').toLowerCase()
    triggerDownload(data, `time-tracking-${slug}.xlsx`, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')
  } catch {
    ui.error(t('timeTracking.export_error'))
  }
}

// Report tab → PDF: delegates to backend.
async function exportReportPDF() {
  if (!report.value) return
  try {
    const params = {
      period: rpt.value.period,
      year:   rpt.value.year,
      month:  rpt.value.month,
      week:   rpt.value.week,
      font:   pdfFont.value,
      lang:   pdfLang.value,
    }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const { data } = await timeEntriesApi.reportPDF(params)
    const slug = report.value.period_label.replace(/\s+/g, '-').toLowerCase()
    triggerDownload(data, `time-tracking-${slug}.pdf`, 'application/pdf')
  } catch {
    ui.error(t('timeTracking.export_error'))
  }
}

function triggerDownload(blob, filename, type) {
  const url = URL.createObjectURL(new Blob([blob], { type }))
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

// ── Init ──────────────────────────────────────────────────────────────────
onMounted(async () => {
  const fetches = [
    customersApi.list().catch(() => ({ data: [] })),
    projectsApi.list().catch(() => ({ data: [] })),
  ]
  if (canViewOtherUsers.value) {
    fetches.push(client.get('/users').catch(() => ({ data: [] })))
  }
  const results = await Promise.all(fetches)
  customers.value = results[0].data
  projects.value  = results[1].data.filter(p => !p.archived)
  if (canViewOtherUsers.value) {
    allUsers.value = results[2].data
    selectedUserId.value = auth.user?.id ?? null
  }
  await loadWeek()
  await loadReport()
})
</script>

<style scoped>
/* ── Shell ── */
.tt-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--color-bg);
  font-size: 13px;
}

/* ── Top bar ── */
.tt-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 16px;
  height: 48px;
  background: var(--color-primary);
  color: #fff;
  flex-shrink: 0;
  flex-wrap: wrap;
}

.tt-emp-label { font-size: 11px; opacity: .7; margin-right: 6px; }
.tt-emp-name  { font-weight: 600; }
.tt-user-select {
  background: rgba(255,255,255,.15);
  border: 1px solid rgba(255,255,255,.3);
  border-radius: 4px;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  padding: 3px 8px;
  cursor: pointer;
}
.tt-user-select option { background: var(--color-surface); color: var(--color-text); }

.tt-week-nav {
  display: flex;
  align-items: center;
  gap: 10px;
}
.wk-label { font-weight: 600; font-size: 14px; letter-spacing: .5px; }
.nav-btn {
  background: rgba(255,255,255,.2);
  border: none;
  color: #fff;
  width: 26px;
  height: 26px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}
.nav-btn:hover { background: rgba(255,255,255,.35); }

.tt-mode-tabs { display: flex; gap: 2px; }
.tt-mode-btn {
  padding: 5px 14px;
  border: 1px solid rgba(255,255,255,.4);
  border-radius: 4px;
  background: transparent;
  color: rgba(255,255,255,.8);
  cursor: pointer;
  font-size: 12px;
}
.tt-mode-btn.active {
  background: #fff;
  color: var(--color-primary);
  font-weight: 600;
  border-color: #fff;
}

/* ── Sheet wrapper ── */
.tt-sheet-outer {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}
.tt-scroll {
  flex: 1;
  overflow: auto;
}
.tt-loading {
  padding: 40px;
  text-align: center;
  color: var(--color-text-muted);
}

/* ── Table ── */
.tt-table {
  border-collapse: collapse;
  min-width: 100%;
  table-layout: fixed;
}

/* Column widths */
.c-nr   { width: 40px; }
.c-info { width: 210px; }
.c-desc { width: 180px; }
.c-day  { width: 82px; }
.c-total { width: 70px; }
.c-act  { width: 54px; }

/* Sticky left columns */
.c-nr, .c-info, .c-desc {
  position: sticky;
  left: 0;
  z-index: 2;
  background: inherit;
}
.c-info { left: 40px; }
.c-desc { left: 250px; }

/* Head */
.tt-head th {
  background: var(--color-surface);
  border-bottom: 2px solid var(--color-border);
  border-right: 1px solid var(--color-border);
  padding: 6px 8px;
  text-align: center;
  font-weight: 600;
  color: var(--color-text);
  white-space: nowrap;
}
.tt-head .c-nr   { text-align: center; }
.tt-head .c-info { text-align: left; display: table-cell; }
.tt-head .c-info .sub { display: block; font-weight: 400; font-size: 11px; color: var(--color-text-muted); }
.tt-head .c-desc { text-align: left; }

.dh-abbr { font-weight: 600; }
.dh-date { font-size: 11px; color: var(--color-text-muted); }

/* Rows */
.tt-row td {
  border-bottom: 1px solid var(--color-border);
  border-right: 1px solid var(--color-border);
  padding: 3px 6px;
  vertical-align: middle;
  background: var(--color-surface);
}
.tt-row.alt td { background: var(--color-bg); }
.c-nr { text-align: center; color: var(--color-text-muted); font-size: 12px; }

.rc-cust { font-weight: 600; font-size: 12px; line-height: 1.3; color: var(--color-text); }
.rc-proj { font-size: 11px; color: var(--color-text-muted); line-height: 1.3; }
.c-desc  { font-size: 12px; color: var(--color-text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* Hour input cells */
.c-day { padding: 2px; }
.h-inp {
  width: 100%;
  border: 1px solid var(--color-border);
  border-radius: 3px;
  padding: 4px 6px;
  text-align: right;
  font-size: 13px;
  background: var(--color-surface);
  color: var(--color-text);
  outline: none;
  box-sizing: border-box;
  -moz-appearance: textfield;
}
.h-inp::-webkit-outer-spin-button,
.h-inp::-webkit-inner-spin-button { -webkit-appearance: none; }
.h-inp:focus { border-color: var(--color-primary); background: var(--color-surface); box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-primary) 20%, transparent); }
.h-inp.h-inp-filled { background: var(--color-surface); font-weight: 600; }
.h-inp:disabled { background: var(--color-bg); cursor: not-allowed; }

/* Totals */
.c-total { text-align: right; font-weight: 600; font-size: 13px; padding: 4px 10px; }
.c-rowtotal { color: var(--color-text); }

/* Footer */
.tt-foot td {
  background: var(--color-surface);
  border-top: 2px solid var(--color-border);
  border-right: 1px solid var(--color-border);
  padding: 6px 8px;
  font-weight: 700;
}
.foot-lbl { text-align: right; }
.c-dttotal { text-align: right; font-size: 13px; }
.grand-total { color: var(--color-primary); font-size: 14px; }

/* Add-row bar */
.tt-add-bar {
  border-top: 1px solid var(--color-border);
  padding: 8px 12px;
  display: flex;
  gap: 8px;
  align-items: center;
  background: var(--color-surface);
  flex-shrink: 0;
}
.btn-add-row {
  background: none;
  border: 1px dashed var(--color-border);
  border-radius: var(--radius);
  padding: 5px 14px;
  cursor: pointer;
  font-size: 13px;
  color: var(--color-primary);
}
.btn-add-row:hover { background: var(--color-bg); border-color: var(--color-primary); }
.tt-export-group { margin-left: auto; display: flex; gap: 6px; align-items: flex-end; }
.pdf-font-group { display: flex; flex-direction: column; gap: 4px; }
.filter-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

/* Action column */
.c-act {
  padding: 0 4px;
  text-align: center;
  white-space: nowrap;
}
.tt-head .c-act { background: var(--color-surface); border-bottom: 2px solid var(--color-border); border-right: 1px solid var(--color-border); }
.tt-row .c-act { background: inherit; }
.tt-foot .c-act { background: var(--color-surface); border-top: 2px solid var(--color-border); }

.act-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  border-radius: 3px;
  cursor: pointer;
  font-size: 12px;
  background: transparent;
  color: var(--color-text-muted);
  opacity: 0;
  transition: opacity .15s, background .1s;
}
.tt-row:hover .act-btn { opacity: 1; }
.act-btn:hover { background: var(--color-bg); color: var(--color-text); }

.act-edit:hover { color: var(--color-primary); }
.act-del:hover  { color: var(--color-danger); }
.act-ok  { opacity: 1 !important; color: var(--color-success); }
.act-ok:hover { background: color-mix(in srgb, var(--color-success) 12%, transparent); }
.act-no  { opacity: 1 !important; color: var(--color-danger); }
.act-no:hover { background: color-mix(in srgb, var(--color-danger) 12%, transparent); }

/* Deleting row highlight */
.tt-row-deleting td { background: color-mix(in srgb, var(--color-danger) 8%, var(--color-surface)) !important; }

/* Editing row highlight */
.tt-editing { background: color-mix(in srgb, var(--color-warning, #f59e0b) 8%, var(--color-surface)) !important; }

/* New row editor */
.tt-newrow td { background: color-mix(in srgb, var(--color-warning, #f59e0b) 8%, var(--color-surface)) !important; }
.nr-sel {
  width: 100%;
  font-size: 12px;
  padding: 3px 4px;
  border: 1px solid var(--color-border);
  border-radius: 3px;
  display: block;
  margin-bottom: 2px;
  background: var(--color-surface);
  color: var(--color-text);
}
.nr-desc {
  width: 100%;
  font-size: 12px;
  padding: 3px 6px;
  border: 1px solid var(--color-border);
  border-radius: 3px;
  box-sizing: border-box;
  background: var(--color-surface);
  color: var(--color-text);
}

/* ── Report ── */
.tt-report-outer { flex: 1; overflow: auto; padding: 20px 24px; }
.report-filters { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; margin-bottom: 20px; }
.fi-sm { min-width: 80px; width: auto; }
.fi-year { width: 80px; }

/* Report header: logo + company name + period */
.rpt-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 3px solid var(--color-primary);
}
.rpt-header-left { flex: 0 0 auto; min-width: 60px; }
.rpt-header-center { flex: 1; text-align: center; }
.rpt-header-right { flex: 0 0 auto; min-width: 60px; }
.rpt-logo { max-height: 52px; max-width: 160px; object-fit: contain; }
.rpt-company-name { font-size: 20px; font-weight: 800; color: var(--color-text); letter-spacing: -0.02em; }
.rpt-title-label { font-size: 12px; font-weight: 600; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: 0.08em; margin-top: 4px; }
.rpt-period-label { font-size: 18px; font-weight: 700; color: var(--color-primary); margin-top: 4px; }

.rpt-group {
  margin-bottom: 16px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
}
.rpt-group-hd {
  display: flex;
  justify-content: space-between;
  padding: 8px 14px;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  font-weight: 600;
  font-size: 13px;
}
.rpt-grp-total { color: var(--color-primary); }
.rpt-grp-empty { padding: 10px 14px; color: var(--color-text-muted); font-size: 12px; }

.rpt-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  table-layout: fixed;
}
.rpt-col-date     { width: 110px; }
.rpt-col-customer { width: 18%; }
.rpt-col-project  { width: 18%; }
.rpt-col-activity { /* fills remaining space */ }
.rpt-col-time     { width: 80px; }
.rpt-table th,
.rpt-table td {
  padding: 6px 14px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rpt-table th { color: var(--color-text-muted); font-weight: 500; background: var(--color-surface); }
.rpt-th-time { text-align: right; }

.rpt-grand-total {
  display: flex;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--color-primary);
  color: #fff;
  font-weight: 700;
  border-radius: var(--radius);
  margin-top: 8px;
  font-size: 14px;
}
.rpt-empty { color: var(--color-text-muted); padding: 20px 0; text-align: center; }
</style>
