<template>
  <div>
    <div class="tab-toolbar">
      <button class="btn btn-primary btn-sm" @click="startCreate">+ {{ $t('macro.new_macro') }}</button>
    </div>

    <!-- Create / Edit form -->
    <div v-if="editing" class="macro-form-card">
      <h3 class="macro-form-title">{{ editing.id ? $t('macro.edit_macro') : $t('macro.new_macro') }}</h3>

      <div class="macro-form-row">
        <div class="macro-form-field" style="flex:2">
          <label class="macro-label" for="macro-name">{{ $t('macro.name') }}</label>
          <input id="macro-name" class="form-input" v-model="form.name" :placeholder="$t('macro.name')" />
        </div>
        <div class="macro-form-field" style="flex:1;min-width:60px">
          <label class="macro-label" for="macro-order">{{ $t('macro.sort_order') }}</label>
          <input id="macro-order" class="form-input" type="number" min="0" v-model.number="form.sort_order" />
        </div>
        <div class="macro-form-field" style="align-self:flex-end">
          <label class="toggle-label">
            <input type="checkbox" v-model="form.is_active" />
            {{ $t('macro.active') }}
          </label>
        </div>
      </div>

      <div class="macro-form-field">
        <label class="macro-label" for="macro-desc">{{ $t('macro.description') }}</label>
        <input id="macro-desc" class="form-input" v-model="form.description" :placeholder="$t('macro.description')" />
      </div>

      <div class="macro-actions-section">
        <div class="macro-actions-header">
          <span class="macro-label">{{ $t('macro.actions') }}</span>
          <button class="btn btn-secondary btn-sm" @click="addAction">+ {{ $t('macro.add_action') }}</button>
        </div>

        <div v-if="!form.actions.length" class="macro-no-actions">{{ $t('macro.no_actions') }}</div>

        <div v-for="(action, idx) in form.actions" :key="idx" class="macro-action-row">
          <label class="sr-only" :for="'act-type-' + idx">{{ $t('macro.action_type') }}</label>
          <select :id="'act-type-' + idx" class="form-input form-input-sm" v-model="action.type" @change="onActionTypeChange(action)">
            <option value="set_status">{{ $t('macro.action_set_status') }}</option>
            <option value="set_priority">{{ $t('macro.action_set_priority') }}</option>
            <option value="set_type">{{ $t('macro.action_set_type') }}</option>
            <option value="add_tag">{{ $t('macro.action_add_tag') }}</option>
            <option value="add_message">{{ $t('macro.action_add_message') }}</option>
          </select>

          <!-- Status select -->
          <template v-if="action.type === 'set_status'">
            <label class="sr-only" :for="'act-val-' + idx">{{ $t('macro.action_value') }}</label>
            <select :id="'act-val-' + idx" class="form-input form-input-sm" v-model="action.value">
              <option value="new">{{ $t('ticket.status_new') }}</option>
              <option value="open">{{ $t('ticket.status_open') }}</option>
              <option value="pending">{{ $t('ticket.status_pending') }}</option>
              <option value="pending_close">{{ $t('ticket.status_pending_close') }}</option>
              <option value="closed">{{ $t('ticket.status_closed') }}</option>
            </select>
          </template>

          <!-- Priority select -->
          <template v-else-if="action.type === 'set_priority'">
            <label class="sr-only" :for="'act-val-' + idx">{{ $t('macro.action_value') }}</label>
            <select :id="'act-val-' + idx" class="form-input form-input-sm" v-model="action.value">
              <option value="low">{{ $t('ticket.priority_low') }}</option>
              <option value="medium">{{ $t('ticket.priority_medium') }}</option>
              <option value="high">{{ $t('ticket.priority_high') }}</option>
              <option value="critical">{{ $t('ticket.priority_critical') }}</option>
            </select>
          </template>

          <!-- Type select -->
          <template v-else-if="action.type === 'set_type'">
            <label class="sr-only" :for="'act-val-' + idx">{{ $t('macro.action_value') }}</label>
            <select :id="'act-val-' + idx" class="form-input form-input-sm" v-model="action.value">
              <option value="incident">{{ $t('ticket.type_incident') }}</option>
              <option value="problem">{{ $t('ticket.type_problem') }}</option>
              <option value="service_request">{{ $t('ticket.type_service_request') }}</option>
              <option value="change_request">{{ $t('ticket.type_change_request') }}</option>
            </select>
          </template>

          <!-- Tag / message text input -->
          <template v-else-if="action.type === 'add_tag'">
            <label class="sr-only" :for="'act-val-' + idx">{{ $t('macro.action_value') }}</label>
            <input :id="'act-val-' + idx" class="form-input form-input-sm" v-model="action.value" :placeholder="$t('ticket.add_tag_placeholder')" style="flex:1" />
          </template>

          <!-- Message textarea -->
          <template v-else-if="action.type === 'add_message'">
            <div class="macro-msg-wrap">
              <label class="sr-only" :for="'act-val-' + idx">{{ $t('macro.action_value') }}</label>
              <textarea :id="'act-val-' + idx" :ref="el => { if (el) msgRefs[idx] = el }" class="form-input form-input-sm macro-msg-area" v-model="action.value" :placeholder="$t('ticket.message_placeholder')" rows="2"></textarea>
              <div class="macro-ph-row">
                <span class="macro-ph-label">{{ $t('macro.placeholders') }}:</span>
                <button v-for="ph in PLACEHOLDERS" :key="ph" type="button" class="macro-ph-chip" @click="insertPlaceholder(action, ph, idx)">{{ ph }}</button>
              </div>
            </div>
          </template>

          <button class="btn-icon-xs macro-remove-btn" @click="removeAction(idx)" :aria-label="$t('common.delete')">✕</button>
        </div>
      </div>

      <div class="macro-form-footer">
        <button class="btn btn-primary btn-sm" @click="confirmSave" :disabled="saving">{{ $t('common.save') }}</button>
        <button class="btn btn-secondary btn-sm" @click="cancelEdit">{{ $t('common.cancel') }}</button>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <table v-else class="data-table">
      <thead>
        <tr>
          <th>{{ $t('macro.name') }}</th>
          <th>{{ $t('macro.description') }}</th>
          <th>{{ $t('macro.actions') }}</th>
          <th>{{ $t('macro.active') }}</th>
          <th>{{ $t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="m in macros" :key="m.id">
          <td><strong>{{ m.name }}</strong></td>
          <td class="text-muted">{{ m.description }}</td>
          <td>{{ m.actions?.length || 0 }}</td>
          <td>
            <span :class="['badge', m.is_active ? 'badge-active' : 'badge-inactive']">
              {{ m.is_active ? $t('sla.yes') : $t('sla.no') }}
            </span>
          </td>
          <td>
            <div class="actions-cell">
              <button class="btn btn-ghost btn-sm" @click="startEdit(m)">{{ $t('common.edit') }}</button>
              <button class="btn btn-ghost btn-sm btn-danger" @click="deleteMacro(m)">{{ $t('common.delete') }}</button>
            </div>
          </td>
        </tr>
        <tr v-if="!macros.length && !editing">
          <td colspan="5" style="text-align:center;color:var(--color-text-muted)">{{ $t('macro.no_macros') }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { macrosApi } from '@/api/macros'
import { useUIStore } from '@/stores/ui'

const ui = useUIStore()
const loading = ref(true)
const saving = ref(false)
const macros = ref([])
const editing = ref(null)
const form = ref(emptyForm())
const msgRefs = ref({})

const PLACEHOLDERS = ['{email}', '{fname}', '{name}', '{subject}', '{ticket_id}', '{agent}', '{agent_fname}']

function insertPlaceholder(action, ph, idx) {
  const el = msgRefs.value[idx]
  if (el) {
    const start = el.selectionStart
    const end = el.selectionEnd
    const val = action.value || ''
    action.value = val.slice(0, start) + ph + val.slice(end)
    nextTick(() => {
      el.selectionStart = el.selectionEnd = start + ph.length
      el.focus()
    })
  } else {
    action.value = (action.value || '') + ph
  }
}

onMounted(load)

function emptyForm() {
  return { name: '', description: '', is_active: true, sort_order: 0, actions: [] }
}

async function load() {
  loading.value = true
  try {
    const { data } = await macrosApi.adminList()
    macros.value = data || []
  } catch { macros.value = [] }
  loading.value = false
}

function startCreate() {
  form.value = emptyForm()
  editing.value = { id: 0 }
}

function startEdit(m) {
  form.value = {
    name: m.name,
    description: m.description,
    is_active: m.is_active,
    sort_order: m.sort_order,
    actions: (m.actions || []).map(a => ({ ...a })),
  }
  editing.value = m
}

function cancelEdit() {
  editing.value = null
}

function addAction() {
  form.value.actions.push({ type: 'set_status', value: 'open' })
}

function removeAction(idx) {
  form.value.actions.splice(idx, 1)
}

function onActionTypeChange(action) {
  const defaults = {
    set_status: 'open',
    set_priority: 'medium',
    set_type: 'incident',
    add_tag: '',
    add_message: '',
  }
  action.value = defaults[action.type] ?? ''
}

async function confirmSave() {
  if (!form.value.name.trim()) {
    ui.error('Name is required')
    return
  }
  saving.value = true
  try {
    if (editing.value.id === 0) {
      await macrosApi.create(form.value)
      ui.success('Macro created')
    } else {
      await macrosApi.update(editing.value.id, form.value)
      ui.success('Macro updated')
    }
    editing.value = null
    await load()
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to save macro')
  } finally {
    saving.value = false
  }
}

async function deleteMacro(m) {
  if (!await ui.confirm(`Delete macro "${m.name}"?`)) return
  try {
    await macrosApi.delete(m.id)
    ui.success('Macro deleted')
    await load()
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to delete macro')
  }
}
</script>

<style scoped>
.macro-form-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}
.macro-form-title { font-size: 15px; font-weight: 600; margin: 0 0 16px; }
.macro-form-row { display: flex; gap: 12px; align-items: flex-start; flex-wrap: wrap; margin-bottom: 12px; }
.macro-form-field { display: flex; flex-direction: column; gap: 4px; flex: 1; min-width: 120px; }
.macro-label { font-size: 12px; font-weight: 600; color: var(--color-text-muted); }
.macro-actions-section { margin-top: 16px; }
.macro-actions-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.macro-no-actions { font-size: 13px; color: var(--color-text-muted); padding: 8px 0; }
.macro-action-row { display: flex; gap: 8px; align-items: flex-start; margin-bottom: 8px; }
.macro-action-row > select.form-input-sm:first-of-type { width: 160px; flex: 0 0 160px; }
.macro-action-row > select.form-input-sm:not(:first-of-type) { flex: 1; min-width: 0; }
.macro-action-row > input.form-input-sm { flex: 1; min-width: 0; }
.macro-msg-wrap { display: flex; flex-direction: column; gap: 4px; flex: 1; min-width: 0; }
.macro-msg-area { resize: vertical; min-height: 60px; width: 100%; box-sizing: border-box; }
.macro-ph-row { display: flex; flex-wrap: wrap; gap: 4px; align-items: center; }
.macro-ph-label { font-size: 11px; color: var(--color-text-muted); flex-shrink: 0; }
.macro-ph-chip { font-size: 11px; font-family: monospace; padding: 1px 6px; border: 1px solid var(--color-border); border-radius: 4px; background: var(--color-surface); color: var(--color-primary); cursor: pointer; line-height: 1.6; }
.macro-ph-chip:hover { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.macro-remove-btn { align-self: flex-start; flex-shrink: 0; margin-top: 4px; }
.macro-form-footer { display: flex; gap: 8px; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--color-border); }
.text-muted { color: var(--color-text-muted); font-size: 13px; }
.loading-state { display: flex; justify-content: center; padding: 48px; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); border: 0; }

/* Column widths for the macros table */
:deep(.data-table) th:nth-child(1),
:deep(.data-table) td:nth-child(1) { width: 20%; min-width: 160px; }
:deep(.data-table) th:nth-child(2),
:deep(.data-table) td:nth-child(2) { width: auto; }
:deep(.data-table) th:nth-child(3),
:deep(.data-table) td:nth-child(3) { width: 90px; min-width: 90px; text-align: center; }
:deep(.data-table) th:nth-child(4),
:deep(.data-table) td:nth-child(4) { width: 90px; min-width: 90px; text-align: center; }
:deep(.data-table) th:nth-child(5),
:deep(.data-table) td:nth-child(5) { width: 120px; min-width: 120px; }
</style>
