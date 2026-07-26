<template>
  <BaseModal :title="$t('search.replace_title')" width="640px" @close="$emit('close')">
    <div class="sr-form">
      <div class="sr-field">
        <label for="sr-find">{{ $t('search.find') }}</label>
        <input id="sr-find" v-model="findText" type="text" class="form-input" autofocus />
      </div>
      <div class="sr-field">
        <label for="sr-replace">{{ $t('search.replace_with') }}</label>
        <input id="sr-replace" v-model="replaceText" type="text" class="form-input" />
      </div>
      <div class="sr-field sr-field-limit">
        <label for="sr-limit">{{ $t('search.results_per_type') }}</label>
        <input
          id="sr-limit"
          v-model.number="resultsLimit"
          type="number"
          class="form-input"
          min="1"
          :max="MAX_RESULTS_LIMIT"
          step="1"
        />
      </div>

      <fieldset class="sr-scope">
        <legend>{{ $t('search.scope') }}</legend>
        <label class="sr-scope-item">
          <input type="checkbox" v-model="scopeCard" />
          {{ $t('search.scope_cards') }}
        </label>
        <label class="sr-scope-item">
          <input type="checkbox" v-model="scopeComment" />
          {{ $t('board.comments') }}
        </label>
        <label class="sr-scope-item">
          <input type="checkbox" v-model="scopeDM" />
          {{ $t('search.scope_dms') }}
        </label>
        <label v-if="showTickets" class="sr-scope-item">
          <input type="checkbox" v-model="scopeTicket" />
          {{ $t('ticket.tickets') }}
        </label>
        <label v-if="showTimeEntries" class="sr-scope-item">
          <input type="checkbox" v-model="scopeTimeEntry" />
          {{ $t('timeTracking.nav') }}
        </label>
      </fieldset>

      <div class="sr-actions">
        <button class="btn btn-secondary btn-sm" :disabled="!canPreview || previewing" @click="runPreview">
          {{ $t('common.preview') }}
        </button>
        <span v-if="previewError" class="sr-error" role="alert">{{ previewError }}</span>
      </div>
    </div>

    <div v-if="results !== null" class="sr-results">
      <p class="sr-summary">{{ summaryText }}</p>
      <div v-if="!results.length" class="sr-empty">{{ $t('common.no_results') }}</div>

      <template v-for="(group, gName) in grouped" :key="gName">
        <div class="sr-group-label">
          {{ groupLabel(gName) }}
          <span v-if="truncatedTypes.has(gName)" class="sr-truncated-note">{{ $t('search.results_truncated', { limit: appliedLimit }) }}</span>
        </div>
        <label
          v-for="r in group"
          :key="rowKey(r)"
          class="sr-row"
          :class="{ 'sr-row-disabled': !r.editable }"
        >
          <input type="checkbox" v-model="checked[rowKey(r)]" :disabled="!r.editable" />
          <span class="sr-diff">
            <span class="sr-before">{{ r.before }}</span>
            <span class="sr-arrow" aria-hidden="true">&rarr;</span>
            <span class="sr-after">{{ r.after }}</span>
          </span>
          <span v-if="!r.editable" class="sr-reason">{{ r.reason }}</span>
        </label>
      </template>
    </div>

    <template #footer>
      <button class="btn btn-ghost btn-sm" @click="$emit('close')">{{ $t('common.cancel') }}</button>
      <button
        v-if="results !== null"
        class="btn btn-primary btn-sm"
        :disabled="selectedCount === 0 || applying"
        @click="runApply"
      >
        {{ $t('search.replace_selected', { count: selectedCount }) }}
      </button>
    </template>
  </BaseModal>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { searchApi } from '@/api/search'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import BaseModal from './BaseModal.vue'

const { t } = useI18n()
const auth = useAuthStore()
const ui = useUIStore()

defineEmits(['close'])

const MAX_RESULTS_LIMIT = 500

const findText = ref('')
const replaceText = ref('')
const resultsLimit = ref(20)

const scopeCard = ref(true)
const scopeComment = ref(true)
const scopeDM = ref(true)
const scopeTicket = ref(true)
const scopeTimeEntry = ref(true)

const showTickets = computed(() => auth.helpdeskEnabled || auth.user?.global_role === 'customer')
const showTimeEntries = computed(() => auth.timeTrackingEnabled || auth.isAdmin)

const canPreview = computed(() => findText.value.trim().length >= 3)

const previewing = ref(false)
const applying = ref(false)
const previewError = ref('')
const results = ref(null) // null = no preview run yet
const checked = reactive({})
const appliedLimit = ref(20)
const truncatedTypes = ref(new Set())

function rowKey(r) {
  return `${r.type}:${r.id}:${r.field}`
}

function selectedTypes() {
  const types = []
  if (scopeCard.value) types.push('card')
  if (scopeComment.value) types.push('card_comment')
  if (scopeDM.value) types.push('dm_message')
  if (showTickets.value && scopeTicket.value) types.push('ticket')
  if (showTimeEntries.value && scopeTimeEntry.value) types.push('time_entry')
  return types
}

async function runPreview() {
  previewError.value = ''
  const types = selectedTypes()
  if (!types.length) {
    previewError.value = t('search.select_scope')
    return
  }
  previewing.value = true
  try {
    const limit = Math.min(Math.max(Math.trunc(resultsLimit.value) || 20, 1), MAX_RESULTS_LIMIT)
    resultsLimit.value = limit
    const { data } = await searchApi.preview(findText.value.trim(), replaceText.value, types, limit)
    results.value = data.results || []
    appliedLimit.value = data.limit || limit
    truncatedTypes.value = new Set(
      Object.entries(data.counts || {})
        .filter(([, count]) => count >= appliedLimit.value)
        .map(([type]) => type)
    )
    for (const key of Object.keys(checked)) delete checked[key]
    for (const r of results.value) {
      checked[rowKey(r)] = r.editable
    }
  } catch (e) {
    previewError.value = e?.response?.data?.error || t('common.error')
    results.value = null
  } finally {
    previewing.value = false
  }
}

const grouped = computed(() => {
  const g = {}
  for (const r of results.value || []) {
    if (!g[r.type]) g[r.type] = []
    g[r.type].push(r)
  }
  return g
})

const selectedCount = computed(() =>
  (results.value || []).filter((r) => r.editable && checked[rowKey(r)]).length
)

const summaryText = computed(() => {
  const total = (results.value || []).length
  const notEditable = (results.value || []).filter((r) => !r.editable).length
  return t('search.match_summary', { total, notEditable })
})

function groupLabel(type) {
  const map = {
    card: t('search.scope_cards'),
    card_comment: t('board.comments'),
    dm_message: t('search.scope_dms'),
    ticket: t('ticket.tickets'),
    time_entry: t('timeTracking.nav')
  }
  return map[type] || type
}

async function runApply() {
  const items = (results.value || [])
    .filter((r) => r.editable && checked[rowKey(r)])
    .map((r) => ({ type: r.type, id: r.id, field: r.field }))
  if (!items.length) return

  applying.value = true
  try {
    const { data } = await searchApi.replace(findText.value.trim(), replaceText.value, items)
    const updatedKeys = new Set((data.updated || []).map((u) => `${u.type}:${u.id}`))
    results.value = (results.value || []).filter((r) => !updatedKeys.has(`${r.type}:${r.id}`))
    const skippedCount = (data.skipped || []).length
    if (skippedCount > 0) {
      ui.error(t('search.replace_partial', { updated: updatedKeys.size, skipped: skippedCount }))
    } else {
      ui.success(t('search.replace_success', { count: updatedKeys.size }))
    }
  } catch (e) {
    ui.error(e?.response?.data?.error || t('common.error'))
  } finally {
    applying.value = false
  }
}
</script>

<style scoped>
.sr-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.sr-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.sr-field label {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}
.sr-field-limit {
  max-width: 160px;
}
.sr-scope {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 8px 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.sr-scope legend {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  padding: 0 4px;
}
.sr-scope-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--color-text);
}
.sr-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
.sr-error {
  color: var(--color-danger, #d33);
  font-size: 12px;
}
.sr-results {
  margin-top: 16px;
  border-top: 1px solid var(--color-border);
  padding-top: 12px;
}
.sr-summary {
  font-size: 12px;
  color: var(--color-text-muted);
  margin: 0 0 8px;
}
.sr-empty {
  padding: 16px;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 13px;
}
.sr-group-label {
  padding: 6px 0 2px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-muted);
}
.sr-truncated-note {
  text-transform: none;
  font-weight: 400;
  letter-spacing: normal;
  margin-left: 6px;
  color: var(--color-warning, #b58900);
}
.sr-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 6px 4px;
  border-radius: 6px;
  cursor: pointer;
}
.sr-row:hover {
  background: var(--color-bg);
}
.sr-row-disabled {
  cursor: not-allowed;
  opacity: 0.55;
}
.sr-diff {
  flex: 1;
  font-size: 12px;
  color: var(--color-text);
  word-break: break-word;
}
.sr-before {
  text-decoration: line-through;
  color: var(--color-text-muted);
}
.sr-arrow {
  margin: 0 6px;
}
.sr-after {
  font-weight: 600;
}
.sr-reason {
  font-size: 11px;
  color: var(--color-text-muted);
  white-space: nowrap;
}
</style>
