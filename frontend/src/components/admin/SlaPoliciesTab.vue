<template>
  <div>
    <div class="tab-toolbar">
      <button class="btn btn-primary btn-sm" @click="startCreate">+ {{ $t('sla.new_policy') }}</button>
    </div>

    <!-- Create / Edit form card -->
    <div v-if="editing" class="sla-form-card">
      <h3 class="sla-form-title">{{ editing.id ? $t('sla.edit_policy') : $t('sla.new_policy') }}</h3>

      <div class="sla-form-row">
        <div class="sla-form-field" style="flex:2">
          <label class="sla-label" for="sla-name">{{ $t('sla.name') }}</label>
          <input id="sla-name" class="form-input" v-model="form.name" :placeholder="$t('sla.name')" />
        </div>
        <div class="sla-form-field" style="align-self:flex-end">
          <label class="toggle-label">
            <input type="checkbox" v-model="form.is_active" />
            {{ $t('sla.active') }}
          </label>
        </div>
      </div>

      <div class="sla-form-row">
        <div class="sla-form-field">
          <label class="sla-label" for="sla-resp">{{ $t('sla.response_time') }}</label>
          <div class="sla-input-unit">
            <input id="sla-resp" class="form-input" type="number" min="0" v-model.number="form.response_time_minutes" placeholder="0" />
            <span class="sla-unit">min</span>
          </div>
        </div>
        <div class="sla-form-field">
          <label class="sla-label" for="sla-resol">{{ $t('sla.resolution_time') }}</label>
          <div class="sla-input-unit">
            <input id="sla-resol" class="form-input" type="number" min="0" v-model.number="form.resolution_time_minutes" placeholder="0" />
            <span class="sla-unit">min</span>
          </div>
        </div>
        <div class="sla-form-field" style="flex:1;min-width:160px">
          <label class="sla-label" for="sla-filter">{{ $t('sla.priority_filter') }}</label>
          <input id="sla-filter" class="form-input" v-model="form.priority_filter" :placeholder="$t('sla.priority_filter_placeholder')" />
        </div>
      </div>

      <div class="sla-form-footer">
        <button class="btn btn-primary btn-sm" @click="confirmSave">{{ $t('common.save') }}</button>
        <button class="btn btn-secondary btn-sm" @click="cancelEdit">{{ $t('common.cancel') }}</button>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <table v-else class="data-table">
      <thead>
        <tr>
          <th>{{ $t('sla.name') }}</th>
          <th>{{ $t('sla.response_time') }}</th>
          <th>{{ $t('sla.resolution_time') }}</th>
          <th>{{ $t('sla.priority_filter') }}</th>
          <th>{{ $t('sla.active') }}</th>
          <th>{{ $t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in policies" :key="p.id">
          <td><button type="button" class="name-link" @click="startEdit(p)">{{ p.name }}</button></td>
          <td>{{ fmtMinutes(p.response_time_minutes) }}</td>
          <td>{{ fmtMinutes(p.resolution_time_minutes) }}</td>
          <td><code>{{ p.priority_filter || $t('sla.all') }}</code></td>
          <td>
            <span :class="['badge', p.is_active ? 'badge-active' : 'badge-inactive']">
              {{ p.is_active ? $t('sla.yes') : $t('sla.no') }}
            </span>
          </td>
          <td>
            <div class="actions-cell">
              <button class="btn btn-ghost btn-sm" @click="startEdit(p)">{{ $t('common.edit') }}</button>
              <button class="btn btn-ghost btn-sm btn-danger" @click="deletePolicy(p)">{{ $t('common.delete') }}</button>
            </div>
          </td>
        </tr>
        <tr v-if="!policies.length && !editing">
          <td colspan="6" style="text-align:center;color:var(--color-text-muted)">{{ $t('sla.no_policies') }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { slaApi } from '@/api/sla'
import { useUIStore } from '@/stores/ui'

const ui = useUIStore()
const loading = ref(true)
const policies = ref([])
const editing = ref(null)
const form = ref(emptyForm())

onMounted(load)

function emptyForm() {
  return { name: '', response_time_minutes: 0, resolution_time_minutes: 0, priority_filter: '', is_active: true }
}

async function load() {
  loading.value = true
  try {
    const { data } = await slaApi.list()
    policies.value = data || []
  } catch { policies.value = [] }
  loading.value = false
}

function startCreate() {
  form.value = emptyForm()
  editing.value = { id: 0 }
}

function startEdit(p) {
  form.value = { ...p }
  editing.value = p
}

function cancelEdit() {
  editing.value = null
}

async function confirmSave() {
  if (!form.value.name.trim()) {
    ui.error('Name is required')
    return
  }
  try {
    if (editing.value.id === 0) {
      await slaApi.create(form.value)
      ui.success('SLA policy created')
    } else {
      await slaApi.update(editing.value.id, form.value)
      ui.success('SLA policy updated')
    }
    editing.value = null
    await load()
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to save SLA policy')
  }
}

async function deletePolicy(p) {
  if (!await ui.confirm('Delete this SLA policy?', { destructive: true })) return
  try {
    await slaApi.delete(p.id)
    ui.success('SLA policy deleted')
    await load()
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to delete SLA policy')
  }
}

function fmtMinutes(m) {
  if (!m || m <= 0) return '—'
  if (m < 60) return `${m}m`
  const h = Math.floor(m / 60)
  const min = m % 60
  return min > 0 ? `${h}h ${min}m` : `${h}h`
}
</script>

<style scoped>
.sla-form-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}
.sla-form-title { font-size: 15px; font-weight: 600; margin: 0 0 16px; }
.sla-form-row { display: flex; gap: 12px; align-items: flex-start; flex-wrap: wrap; margin-bottom: 12px; }
.sla-form-field { display: flex; flex-direction: column; gap: 4px; min-width: 120px; }
.sla-label { font-size: 12px; font-weight: 600; color: var(--color-text-muted); }
.sla-input-unit { display: flex; align-items: center; gap: 6px; }
.sla-input-unit .form-input { width: 110px; }
.sla-unit { font-size: 12px; color: var(--color-text-muted); white-space: nowrap; }
.sla-form-footer { display: flex; gap: 8px; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--color-border); }
.loading-state { display: flex; justify-content: center; padding: 48px; }

.tab-toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.data-table { width: 100%; border-collapse: collapse; background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius); overflow: hidden; }
.data-table th, .data-table td { padding: 12px 16px; text-align: left; border-bottom: 1px solid var(--color-border); font-size: 13px; vertical-align: middle; }
.data-table th { font-weight: 600; color: var(--color-text-muted); font-size: 12px; background: var(--color-bg); }
.data-table small { color: var(--color-text-muted); font-size: 11px; }
.actions-cell { display: flex; gap: 6px; flex-wrap: wrap; }
.badge-active { background: #dcfce7; color: #166534; }
.badge-inactive { background: #fee2e2; color: #991b1b; }
[data-theme="dark"] .badge-active { background: #14532d; color: #86efac; }
[data-theme="dark"] .badge-inactive { background: #450a0a; color: #fca5a5; }

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
