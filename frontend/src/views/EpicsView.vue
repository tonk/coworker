<template>
  <div class="epics-layout">
    <div class="epics-toolbar">
      <div class="epics-toolbar-left">
        <img v-if="projectAvatar(projectStore.currentProject)" :src="projectAvatar(projectStore.currentProject)" class="board-project-avatar" alt="" />
        <h1 class="board-project-name">{{ projectStore.currentProject?.name }}</h1>
      </div>
      <div class="epics-toolbar-right">
        <RouterLink :to="`/projects/${slug}`" class="btn btn-ghost btn-sm">📌 {{ $t('board.board') }}</RouterLink>
        <RouterLink :to="`/projects/${slug}/backlog`" class="btn btn-ghost btn-sm">📋 {{ $t('sprint.backlog_title') }}</RouterLink>
        <button v-if="canManage" class="btn btn-primary btn-sm" @click="startCreate">+ {{ $t('epic.new_epic') }}</button>
      </div>
    </div>

    <!-- Create / edit form -->
    <div v-if="editing" class="epic-form-card">
      <h3 class="epic-form-title">{{ editing.id ? $t('epic.edit_epic') : $t('epic.new_epic') }}</h3>
      <div class="epic-form-row">
        <div class="epic-color-wrap">
          <label class="form-label sr-only" for="epic-color">{{ $t('epic.color') }}</label>
          <input id="epic-color" type="color" v-model="form.color" class="epic-color-input" :aria-label="$t('epic.color')" />
        </div>
        <div style="flex:1">
          <label class="form-label sr-only" for="epic-name">{{ $t('epic.name') }}</label>
          <input id="epic-name" ref="nameInputEl" class="form-input" v-model="form.name" :placeholder="$t('epic.name')" />
        </div>
        <select class="form-input" style="width:130px;flex-shrink:0" v-model="form.status" :aria-label="$t('epic.status')">
          <option value="open">{{ $t('epic.status_open') }}</option>
          <option value="done">{{ $t('epic.status_done') }}</option>
        </select>
      </div>
      <div class="form-group">
        <label class="form-label" for="epic-desc">{{ $t('epic.description') }}</label>
        <textarea id="epic-desc" class="form-input" v-model="form.description" rows="2" :placeholder="$t('epic.description')"></textarea>
      </div>
      <div class="epic-form-footer">
        <button class="btn btn-primary btn-sm" @click="confirmSave" :disabled="saving || !form.name.trim()">{{ $t('common.save') }}</button>
        <button class="btn btn-secondary btn-sm" @click="cancelEdit">{{ $t('common.cancel') }}</button>
      </div>
    </div>

    <div v-if="epicsStore.loading" class="loading-state">
      <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <div v-else-if="!epicsStore.epics.length && !editing" class="epics-empty">
      {{ $t('epic.no_epics') }}
    </div>

    <div v-else ref="epicsListEl" class="epics-list">
      <div v-for="epic in epicsStore.epics" :key="epic.id" class="epic-row">
        <span class="epic-drag-handle" aria-hidden="true" :title="$t('ticketChecklist.drag_reorder')">⠿</span>
        <span class="epic-color-dot" :style="{ background: epic.color }"></span>
        <div class="epic-row-main">
          <button type="button" class="epic-name-btn" @click="startEdit(epic)">{{ epic.name }}</button>
          <span v-if="epic.description" class="epic-desc-preview">{{ epic.description }}</span>
        </div>
        <span :class="['epic-status-badge', epic.status]">{{ $t(`epic.status_${epic.status}`) }}</span>
        <div class="epic-progress" :title="`${epic.done_count} / ${epic.card_count} ${$t('epic.cards_done')}`">
          <div class="epic-progress-bar">
            <div class="epic-progress-fill" :style="{ width: epic.card_count ? (epic.done_count / epic.card_count * 100) + '%' : '0%', background: epic.color }"></div>
          </div>
          <span class="epic-progress-label">{{ epic.done_count }}/{{ epic.card_count }}</span>
        </div>
        <div class="epic-row-actions">
          <button class="btn btn-ghost btn-sm" @click="startEdit(epic)">{{ $t('common.edit') }}</button>
          <button class="btn btn-ghost btn-sm btn-danger" @click="confirmDelete(epic)">{{ $t('common.delete') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Sortable from 'sortablejs'
import { useEpicsStore } from '@/stores/epics'
import { useProjectStore } from '@/stores/project'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { resolveAssetUrl } from '@/api/serverConfig'
import { projectsApi } from '@/api/projects'

const route = useRoute()
const { t } = useI18n()
const slug = computed(() => route.params.slug)
const epicsStore = useEpicsStore()
const projectStore = useProjectStore()
const auth = useAuthStore()
const ui = useUIStore()

const editing = ref(null)
const saving = ref(false)
const form = ref(emptyForm())
const nameInputEl = ref(null)
const epicsListEl = ref(null)
let sortableInstance = null
const projectMembers = ref([])

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}

const canManage = computed(() => {
  if (auth.user?.global_role === 'admin') return true
  const me = projectMembers.value.find(m => m.user_id === auth.user?.id)
  return me ? ['owner', 'admin', 'member'].includes(me.role) : false
})

function emptyForm() {
  return { name: '', description: '', color: '#6366f1', status: 'open' }
}

function startCreate() {
  form.value = emptyForm()
  editing.value = { id: 0 }
  nextTick(() => nameInputEl.value?.focus())
}

function startEdit(epic) {
  form.value = { name: epic.name, description: epic.description || '', color: epic.color, status: epic.status }
  editing.value = epic
  nextTick(() => nameInputEl.value?.focus())
}

function cancelEdit() {
  editing.value = null
}

async function confirmSave() {
  if (!form.value.name.trim()) return
  saving.value = true
  try {
    if (editing.value.id) {
      await epicsStore.updateEpic(editing.value.id, form.value)
    } else {
      await epicsStore.createEpic({ ...form.value, position: (epicsStore.epics.length + 1) * 1000 })
    }
    editing.value = null
  } catch (e) {
    ui.error(e.response?.data?.error || t('epic.save_failed'))
  } finally {
    saving.value = false
  }
}

async function confirmDelete(epic) {
  if (!await ui.confirm(`${t('epic.delete_confirm')} "${epic.name}"?`)) return
  try {
    await epicsStore.deleteEpic(epic.id)
    ui.success(t('epic.deleted'))
  } catch (e) {
    ui.error(e.response?.data?.error || t('common.error'))
  }
}

function initSortable() {
  if (!epicsListEl.value || sortableInstance) return
  sortableInstance = new Sortable(epicsListEl.value, {
    animation: 150,
    handle: '.epic-drag-handle',
    onEnd(evt) {
      const moved = epicsStore.epics.splice(evt.oldIndex, 1)[0]
      epicsStore.epics.splice(evt.newIndex, 0, moved)
      const items = epicsStore.epics.map((e, i) => ({ id: e.id, position: (i + 1) * 1000 }))
      epicsStore.epics.forEach((e, i) => { e.position = (i + 1) * 1000 })
      epicsStore.reorderEpics(items).catch(() => {})
    },
  })
}

watch(() => epicsStore.epics.length, async (len) => {
  if (sortableInstance) { sortableInstance.destroy(); sortableInstance = null }
  if (len) { await nextTick(); initSortable() }
})

onMounted(async () => {
  await Promise.all([
    epicsStore.loadEpics(slug.value),
    projectStore.fetchProject(slug.value),
  ])
  const { data } = await projectsApi.listMembers(slug.value)
  projectMembers.value = data
  if (epicsStore.epics.length) {
    await nextTick()
    initSortable()
  }
})

onUnmounted(() => {
  if (sortableInstance) { sortableInstance.destroy(); sortableInstance = null }
  epicsStore.reset()
})
</script>

<style scoped>
.epics-layout { display: flex; flex-direction: column; height: 100%; overflow: hidden; }

.epics-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
}
.epics-toolbar-left, .epics-toolbar-right { display: flex; align-items: center; gap: 8px; }
.board-project-avatar { width: 22px; height: 22px; border-radius: 6px; object-fit: cover; border: 1px solid var(--color-border); }
.board-project-name { font-weight: 600; font-size: 15px; }

.epic-form-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 16px 20px;
  margin: 16px;
}
.epic-form-title { font-size: 15px; font-weight: 600; margin: 0 0 12px; }
.epic-form-row { display: flex; gap: 8px; align-items: center; margin-bottom: 10px; }
.epic-color-wrap { flex-shrink: 0; }
.epic-color-input { width: 36px; height: 34px; border: 1px solid var(--color-border); border-radius: 6px; padding: 2px; cursor: pointer; }
.epic-form-footer { display: flex; gap: 8px; margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--color-border); }

.loading-state { display: flex; justify-content: center; padding: 48px; }
.epics-empty { text-align: center; padding: 48px 24px; color: var(--color-text-muted); font-size: 14px; }

.epics-list { flex: 1; overflow-y: auto; padding: 16px; display: flex; flex-direction: column; gap: 6px; }

.epic-row {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 10px 12px;
}
.epic-row:hover { box-shadow: 0 1px 4px rgba(0,0,0,.08); }

.epic-drag-handle { color: var(--color-text-muted); cursor: grab; font-size: 14px; flex-shrink: 0; user-select: none; opacity: 0; transition: opacity .1s; }
.epic-row:hover .epic-drag-handle { opacity: 1; }
.epic-drag-handle:active { cursor: grabbing; }

.epic-color-dot { width: 12px; height: 12px; border-radius: 50%; flex-shrink: 0; }

.epic-row-main { flex: 1; min-width: 0; }
.epic-name-btn { background: none; border: none; padding: 0; font: inherit; font-weight: 600; color: var(--color-primary); cursor: pointer; text-align: left; }
.epic-name-btn:hover { text-decoration: underline; }
.epic-desc-preview { display: block; font-size: 12px; color: var(--color-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 400px; }

.epic-status-badge { font-size: 11px; border-radius: 9999px; padding: 2px 8px; font-weight: 600; flex-shrink: 0; }
.epic-status-badge.open { background: #dbeafe; color: #1e40af; }
.epic-status-badge.done { background: #dcfce7; color: #166534; }
[data-theme="dark"] .epic-status-badge.open { background: #1e3a5f; color: #93c5fd; }
[data-theme="dark"] .epic-status-badge.done { background: #14532d; color: #86efac; }

.epic-progress { display: flex; align-items: center; gap: 6px; flex-shrink: 0; }
.epic-progress-bar { width: 80px; height: 6px; background: var(--color-border); border-radius: 9999px; overflow: hidden; }
.epic-progress-fill { height: 100%; border-radius: 9999px; transition: width .2s; }
.epic-progress-label { font-size: 11px; color: var(--color-text-muted); min-width: 30px; }

.epic-row-actions { display: flex; gap: 4px; flex-shrink: 0; }

.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); border: 0; }
</style>
