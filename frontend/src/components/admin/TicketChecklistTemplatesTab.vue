<template>
  <div>
    <div class="tab-toolbar">
      <button class="btn btn-primary btn-sm" @click="startCreate">+ {{ $t('ticketChecklist.new_template') }}</button>
    </div>

    <div v-if="editing" class="tcl-form-card">
      <h3 class="tcl-form-title">{{ editing.id ? $t('ticketChecklist.edit_template') : $t('ticketChecklist.new_template') }}</h3>

      <div class="tcl-form-row">
        <div class="tcl-form-field" style="flex:2">
          <label class="tcl-label" for="tcl-name">{{ $t('ticketChecklist.name') }}</label>
          <input id="tcl-name" class="form-input" v-model="form.name" :placeholder="$t('ticketChecklist.name')" />
        </div>
        <div class="tcl-form-field" style="flex:1;min-width:60px">
          <label class="tcl-label" for="tcl-order">{{ $t('ticketChecklist.sort_order') }}</label>
          <input id="tcl-order" class="form-input" type="number" min="0" v-model.number="form.sort_order" />
        </div>
        <div class="tcl-form-field" style="align-self:flex-end">
          <label class="toggle-label">
            <input type="checkbox" v-model="form.is_active" />
            {{ $t('ticketChecklist.active') }}
          </label>
        </div>
      </div>

      <div class="tcl-form-field">
        <label class="tcl-label" for="tcl-desc">{{ $t('ticketChecklist.description') }}</label>
        <input id="tcl-desc" class="form-input" v-model="form.description" :placeholder="$t('ticketChecklist.description')" />
      </div>

      <div class="tcl-items-section">
        <div class="tcl-items-header">
          <span class="tcl-label">{{ $t('ticketChecklist.items') }}</span>
          <button class="btn btn-secondary btn-sm" @click="addItem">+ {{ $t('ticketChecklist.add_item') }}</button>
        </div>
        <div v-if="!form.items.length" class="tcl-no-items">{{ $t('ticketChecklist.no_items') }}</div>
        <div v-for="(item, idx) in form.items" :key="idx" class="tcl-item-row">
          <label class="sr-only" :for="'tcl-item-' + idx">{{ $t('ticketChecklist.items') }}</label>
          <input :id="'tcl-item-' + idx" class="form-input form-input-sm" v-model="form.items[idx]" :placeholder="$t('ticketChecklist.item_placeholder')" />
          <button class="btn-icon-xs tcl-remove-btn" @click="removeItem(idx)" :aria-label="$t('common.delete')">✕</button>
        </div>
      </div>

      <div class="tcl-form-footer">
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
          <th>{{ $t('ticketChecklist.name') }}</th>
          <th>{{ $t('ticketChecklist.description') }}</th>
          <th>{{ $t('ticketChecklist.items') }}</th>
          <th>{{ $t('ticketChecklist.active') }}</th>
          <th>{{ $t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="t in templates" :key="t.id">
          <td>
            <button type="button" class="name-link" @click="startEdit(t)">{{ t.name }}</button>
          </td>
          <td class="text-muted">{{ t.description }}</td>
          <td>{{ t.items?.length || 0 }}</td>
          <td>
            <span :class="['badge', t.is_active ? 'badge-active' : 'badge-inactive']">
              {{ t.is_active ? $t('sla.yes') : $t('sla.no') }}
            </span>
          </td>
          <td>
            <div class="actions-cell">
              <button class="btn btn-ghost btn-sm" @click="startEdit(t)">{{ $t('common.edit') }}</button>
              <button class="btn btn-ghost btn-sm btn-danger" @click="deleteTemplate(t)">{{ $t('common.delete') }}</button>
            </div>
          </td>
        </tr>
        <tr v-if="!templates.length && !editing">
          <td colspan="5" style="text-align:center;color:var(--color-text-muted)">{{ $t('ticketChecklist.no_templates') }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ticketChecklistsApi } from '@/api/ticketChecklists'
import { useUIStore } from '@/stores/ui'

const { t } = useI18n()
const ui = useUIStore()
const loading = ref(true)
const saving = ref(false)
const templates = ref([])
const editing = ref(null)
const form = ref(emptyForm())
const formSnapshot = ref('')

onMounted(load)

function emptyForm() {
  return { name: '', description: '', is_active: true, sort_order: 0, items: [] }
}

function snapshotForm(f) {
  return JSON.stringify({
    name: f.name,
    description: f.description,
    is_active: f.is_active,
    sort_order: f.sort_order,
    items: [...(f.items || [])],
  })
}

function formIsDirty() {
  return snapshotForm(form.value) !== formSnapshot.value
}

function onEditEscape(e) {
  if (e.key !== 'Escape' || !editing.value || formIsDirty()) return
  e.preventDefault()
  cancelEdit()
}

watch(editing, (val) => {
  if (val) window.addEventListener('keydown', onEditEscape)
  else window.removeEventListener('keydown', onEditEscape)
})

onUnmounted(() => window.removeEventListener('keydown', onEditEscape))

async function load() {
  loading.value = true
  try {
    const { data } = await ticketChecklistsApi.adminList()
    templates.value = data || []
  } catch {
    templates.value = []
  }
  loading.value = false
}

function startCreate() {
  form.value = emptyForm()
  editing.value = { id: 0 }
  formSnapshot.value = snapshotForm(form.value)
}

function startEdit(tmpl) {
  form.value = {
    name: tmpl.name,
    description: tmpl.description,
    is_active: tmpl.is_active,
    sort_order: tmpl.sort_order,
    items: (tmpl.items || []).map(i => i),
  }
  editing.value = tmpl
  formSnapshot.value = snapshotForm(form.value)
}

function cancelEdit() {
  editing.value = null
}

function addItem() {
  form.value.items.push('')
}

function removeItem(idx) {
  form.value.items.splice(idx, 1)
}

async function confirmSave() {
  if (!form.value.name.trim()) {
    ui.error(t('ticketChecklist.name') + ' required')
    return
  }
  const payload = {
    name: form.value.name.trim(),
    description: form.value.description.trim(),
    is_active: form.value.is_active,
    sort_order: form.value.sort_order,
    items: form.value.items.map(i => i.trim()).filter(Boolean),
  }
  saving.value = true
  try {
    if (editing.value.id) {
      await ticketChecklistsApi.update(editing.value.id, payload)
    } else {
      await ticketChecklistsApi.create(payload)
    }
    editing.value = null
    await load()
    ui.success(t('ticketChecklist.saved'))
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to save')
  }
  saving.value = false
}

async function deleteTemplate(tmpl) {
  if (!confirm(t('common.confirm_delete'))) return
  try {
    await ticketChecklistsApi.delete(tmpl.id)
    await load()
    ui.success(t('ticketChecklist.deleted'))
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to delete')
  }
}
</script>

<style scoped>
.tcl-form-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}
.tcl-form-title { font-size: 15px; font-weight: 600; margin: 0 0 16px; }
.tcl-form-row { display: flex; gap: 12px; align-items: flex-start; flex-wrap: wrap; margin-bottom: 12px; }
.tcl-form-field { display: flex; flex-direction: column; gap: 4px; flex: 1; min-width: 120px; }
.tcl-label { font-size: 12px; font-weight: 600; color: var(--color-text-muted); }
.tcl-items-section { margin-top: 16px; }
.tcl-items-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.tcl-no-items { font-size: 13px; color: var(--color-text-muted); padding: 8px 0; }
.tcl-item-row { display: flex; gap: 8px; align-items: center; margin-bottom: 8px; }
.tcl-item-row .form-input { flex: 1; min-width: 0; }
.tcl-remove-btn { flex-shrink: 0; }
.tcl-form-footer { display: flex; gap: 8px; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--color-border); }
.text-muted { color: var(--color-text-muted); font-size: 13px; }
.loading-state { display: flex; justify-content: center; padding: 48px; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); border: 0; }

/* Column widths — match MacrosTab */
:deep(.data-table) th:nth-child(1),
:deep(.data-table) td:nth-child(1) { width: 28%; min-width: 200px; }
:deep(.data-table) th:nth-child(2),
:deep(.data-table) td:nth-child(2) { width: 40%; }
:deep(.data-table) th:nth-child(3),
:deep(.data-table) td:nth-child(3) { width: 90px; min-width: 90px; text-align: center; }
:deep(.data-table) th:nth-child(4),
:deep(.data-table) td:nth-child(4) { width: 90px; min-width: 90px; text-align: center; }
:deep(.data-table) th:nth-child(5),
:deep(.data-table) td:nth-child(5) { width: 120px; min-width: 120px; }
.name-link {
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  font-weight: 600;
  color: var(--color-primary);
  cursor: pointer;
  text-align: left;
}
.name-link:hover { text-decoration: underline; }
.name-link:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; border-radius: 2px; }
</style>
