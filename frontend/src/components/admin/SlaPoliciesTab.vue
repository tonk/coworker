<template>
  <div>
    <div class="tab-toolbar">
      <button class="btn btn-primary btn-sm" @click="startCreate">+ {{ $t('sla.new_policy') }}</button>
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
        <tr v-if="editing && editing.id === 0">
          <td colspan="6" style="padding:8px">
            <div class="sla-form-row">
              <label class="sr-only" for="sla-name">{{ $t('sla.name') }}</label>
              <input id="sla-name" class="form-input" v-model="form.name" :placeholder="$t('sla.name')" style="flex:2" />
              <label class="sr-only" for="sla-resp">{{ $t('sla.response_time') }}</label>
              <input id="sla-resp" class="form-input" type="number" min="0" v-model.number="form.response_time_minutes" :placeholder="'0'" style="width:80px" title="minutes" />
              <label class="sr-only" for="sla-resol">{{ $t('sla.resolution_time') }}</label>
              <input id="sla-resol" class="form-input" type="number" min="0" v-model.number="form.resolution_time_minutes" :placeholder="'0'" style="width:80px" title="minutes" />
              <label class="sr-only" for="sla-filter">{{ $t('sla.priority_filter') }}</label>
              <input id="sla-filter" class="form-input" v-model="form.priority_filter" :placeholder="$t('sla.priority_filter_placeholder')" style="flex:1" />
              <label class="toggle-label" style="white-space:nowrap;font-size:12px">
                <input type="checkbox" v-model="form.is_active" />
                {{ $t('sla.active') }}
              </label>
              <button class="btn btn-primary btn-sm" @click="confirmSave">{{ $t('common.save') }}</button>
              <button class="btn btn-secondary btn-sm" @click="cancelEdit">{{ $t('common.cancel') }}</button>
            </div>
          </td>
        </tr>
        <tr v-for="p in policies" :key="p.id">
          <template v-if="editing && editing.id === p.id">
            <td colspan="6" style="padding:8px">
              <div class="sla-form-row">
                <input class="form-input" v-model="form.name" style="flex:2" />
                <input class="form-input" type="number" min="0" v-model.number="form.response_time_minutes" style="width:80px" />
                <input class="form-input" type="number" min="0" v-model.number="form.resolution_time_minutes" style="width:80px" />
                <input class="form-input" v-model="form.priority_filter" style="flex:1" />
                <label class="toggle-label" style="white-space:nowrap;font-size:12px">
                  <input type="checkbox" v-model="form.is_active" />
                  {{ $t('sla.active') }}
                </label>
                <button class="btn btn-primary btn-sm" @click="confirmSave">{{ $t('common.save') }}</button>
                <button class="btn btn-secondary btn-sm" @click="editing = null">{{ $t('common.cancel') }}</button>
              </div>
            </td>
          </template>
          <template v-else>
            <td><strong>{{ p.name }}</strong></td>
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
          </template>
        </tr>
        <tr v-if="!policies.length && !(editing && editing.id === 0)">
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
const form = ref({ name: '', response_time_minutes: 0, resolution_time_minutes: 0, priority_filter: '', is_active: true })

onMounted(load)

async function load() {
  loading.value = true
  try {
    const { data } = await slaApi.list()
    policies.value = data || []
  } catch { policies.value = [] }
  loading.value = false
}

function startCreate() {
  form.value = { name: '', response_time_minutes: 0, resolution_time_minutes: 0, priority_filter: '', is_active: true }
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
  if (!await ui.confirm('Delete this SLA policy?')) return
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
.sla-form-row { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
.loading-state { display: flex; justify-content: center; padding: 48px; }
</style>
