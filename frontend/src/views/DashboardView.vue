<template>
  <main class="dashboard-main">
      <div class="dashboard-container">
        <div class="dashboard-header">
          <h1>{{ $t('project.projects') }}</h1>
          <div class="dashboard-header-controls">
            <select v-if="customerOptions.length > 1" class="form-input customer-filter" v-model="selectedCustomer">
              <option value="">{{ $t('customer.all_customers') }}</option>
              <option v-for="c in customerOptions" :key="c.value" :value="c.value">{{ c.label }}</option>
            </select>
            <button class="btn btn-primary" @click="showCreate = true">+ {{ $t('project.new_project') }}</button>
          </div>
        </div>

        <div v-if="projectStore.loading" class="loading-state">
          <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
        </div>

        <div v-else-if="projectStore.projects.length === 0" class="empty-state">
          <p>{{ $t('project.no_projects') }}</p>
        </div>

        <div v-else class="projects-grid" ref="gridEl">
          <div
            v-for="project in filteredProjects"
            :key="project.id"
            class="project-card"
            :data-id="project.id"
            @click="router.push(`/projects/${project.slug}`)"
          >
            <div v-if="isAdmin && !selectedCustomer" class="drag-handle" title="Drag to reorder" @click.stop>⠿</div>
            <div class="project-color-bar" :style="{ background: project.color || '#6366f1' }"></div>
            <div class="project-card-body">
              <h3 class="project-title-row">
                <img v-if="projectAvatar(project)" :src="projectAvatar(project)" class="project-avatar" alt="" />
                <span>{{ project.name }}</span>
              </h3>
              <p v-if="project.customer" class="project-customer">🏢 {{ project.customer.name }}</p>
              <p v-if="project.description" class="project-desc">{{ project.description }}</p>
              <p class="project-open-cards">{{ project.open_card_count }} {{ $t('board.open_cards') }}</p>
              <div class="project-actions">
                <button
                  class="btn btn-ghost btn-sm star-btn"
                  :class="{ starred: sidebarStore.isStarred(project.slug) }"
                  @click.stop="toggleStar(project)"
                  :title="sidebarStore.isStarred(project.slug) ? $t('project.unstar') : $t('project.star')"
                >★</button>
                <RouterLink :to="`/projects/${project.slug}/settings`" class="btn btn-ghost btn-sm" @click.stop>
                  ⚙
                </RouterLink>
              </div>
            </div>
          </div>
        </div>
      </div>
  </main>

  <BaseModal v-if="showCreate" :title="$t('project.new_project')" @close="showCreate = false; newProject.key_prefix = ''; prefixTouched = false">
      <form @submit.prevent="handleCreate">
        <div class="form-group">
          <label class="form-label">{{ $t('project.project_name') }}</label>
          <input class="form-input" v-model="newProject.name" required autofocus />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('project.key_prefix') }} *</label>
          <div style="display:flex;align-items:center;gap:8px">
            <input
              class="form-input"
              style="width:120px;text-transform:uppercase;font-family:monospace"
              :value="newProject.key_prefix"
              maxlength="10"
              @input="e => { const v = e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, ''); e.target.value = v; newProject.key_prefix = v; prefixTouched = true }"
            />
            <span style="font-size:13px;color:var(--color-text-muted)">{{ $t('project.key_prefix_hint') }} &nbsp;<code style="color:var(--color-primary)">{{ newProject.key_prefix || '???' }}-1</code></span>
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('project.description') }}</label>
          <textarea class="form-input" v-model="newProject.description" rows="3"></textarea>
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('project.color') }}</label>
          <input type="color" class="form-input" v-model="newProject.color" style="height:40px;padding:4px" />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('project.customer') }} *</label>
          <select class="form-input" v-model="newProject.customer_id" style="max-width:400px" required>
            <option :value="null" disabled>— {{ $t('project.customer') }} —</option>
            <option v-for="c in createCustomers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('sprint.board_type') }}</label>
          <select class="form-input" v-model="newProject.board_type" style="max-width:300px">
            <option value="kanban">{{ $t('sprint.board_type_kanban') }}</option>
            <option value="scrum">{{ $t('sprint.board_type_scrum') }}</option>
          </select>
          <p style="font-size:12px;color:var(--color-text-muted);margin-top:4px">{{ $t('sprint.board_type_hint') }}</p>
        </div>
      </form>
      <template #footer>
        <button class="btn btn-secondary" @click="showCreate = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="handleCreate" :disabled="creating || !newProject.key_prefix.trim() || !newProject.customer_id">{{ $t('project.create') }}</button>
      </template>
  </BaseModal>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import Sortable from 'sortablejs'
import BaseModal from '@/components/common/BaseModal.vue'
import { useProjectStore } from '@/stores/project'
import { useUIStore } from '@/stores/ui'
import { useSidebarStore } from '@/stores/sidebar'
import { useAuthStore } from '@/stores/auth'
import { resolveAssetUrl } from '@/api/serverConfig'
import { customersApi } from '@/api/customers'

const router = useRouter()
const projectStore = useProjectStore()
const ui = useUIStore()
const sidebarStore = useSidebarStore()
const auth = useAuthStore()
const showCreate = ref(false)
const gridEl = ref(null)
let sortable = null

const isAdmin = computed(() => auth.user?.global_role === 'admin')
const selectedCustomer = ref('')

const customerOptions = computed(() => {
  const seen = new Map()
  for (const p of projectStore.projects) {
    if (p.customer && !seen.has(p.customer.id)) {
      seen.set(p.customer.id, p.customer.name)
    }
  }
  return Array.from(seen.entries())
    .sort((a, b) => a[1].localeCompare(b[1]))
    .map(([id, name]) => ({ value: String(id), label: name }))
})

const filteredProjects = computed(() => {
  if (!selectedCustomer.value) return projectStore.projects
  return projectStore.projects.filter(p => String(p.customer?.id) === selectedCustomer.value)
})

const creating = ref(false)
const newProject = ref({ name: '', description: '', color: '#6366f1', key_prefix: '', board_type: 'kanban', customer_id: null })
const prefixTouched = ref(false)
const createCustomers = ref([])

function autoPrefix(name) {
  const words = name.toUpperCase().split(/[^A-Z0-9]+/).filter(Boolean)
  let r = ''
  for (const w of words) { if (r.length >= 3) break; r += w[0] }
  if (r.length < 3 && words.length > 0) {
    for (let i = 1; i < words[0].length && r.length < 3; i++) r += words[0][i]
  }
  while (r.length < 3) r += 'X'
  return r.slice(0, 3)
}

watch(() => newProject.value.name, (name) => {
  if (!prefixTouched.value) newProject.value.key_prefix = autoPrefix(name)
})

onMounted(async () => {
  await projectStore.fetchProjects()
  sidebarStore.fetchStarred()
  customersApi.list().then(r => { createCustomers.value = r.data || [] }).catch(() => {})
  if (isAdmin.value && gridEl.value) {
    sortable = Sortable.create(gridEl.value, {
      handle: '.drag-handle',
      animation: 150,
      onEnd({ oldIndex, newIndex }) {
        if (oldIndex === newIndex) return
        const reordered = [...projectStore.projects]
        const [moved] = reordered.splice(oldIndex, 1)
        reordered.splice(newIndex, 0, moved)
        projectStore.reorderProjects(reordered)
      }
    })
  }
})

onUnmounted(() => sortable?.destroy())

async function toggleStar(project) {
  if (sidebarStore.isStarred(project.slug)) {
    await sidebarStore.unstarProject(project.slug)
  } else {
    await sidebarStore.starProject(project.slug)
  }
}

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}

async function handleCreate() {
  if (!newProject.value.name) return
  if (!newProject.value.customer_id) {
    ui.error('A customer is required')
    return
  }
  creating.value = true
  try {
    const project = await projectStore.createProject(newProject.value)
    showCreate.value = false
    newProject.value = { name: '', description: '', color: '#6366f1', key_prefix: '', board_type: 'kanban', customer_id: null }
    prefixTouched.value = false
    router.push(`/projects/${project.slug}`)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to create project')
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.dashboard-main { flex: 1; padding: 32px 24px; }
.dashboard-container { max-width: 1200px; margin: 0 auto; }
.dashboard-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px; }
.dashboard-header h1 { font-size: 22px; font-weight: 700; }
.dashboard-header-controls { display: flex; align-items: center; gap: 10px; }
.customer-filter { width: auto; min-width: 160px; padding: 6px 10px; font-size: 13px; }

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.star-btn { color: var(--color-text-muted); }
.star-btn.starred { color: #f59e0b; }

.project-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  cursor: pointer;
  overflow: hidden;
  position: relative;
  transition: box-shadow .15s, transform .1s;
}
.project-card:hover { box-shadow: var(--shadow-md); transform: translateY(-1px); }

.project-color-bar { height: 4px; }

.project-card-body { padding: 16px; }
.project-card-body h3 { font-size: 15px; font-weight: 600; margin-bottom: 6px; }
.project-title-row { display: flex; align-items: center; gap: 8px; }
.project-avatar {
  width: 20px;
  height: 20px;
  border-radius: 5px;
  object-fit: cover;
  border: 1px solid var(--color-border);
  flex-shrink: 0;
}
.project-customer { font-size: 11px; color: var(--color-text-muted); margin-bottom: 4px; }
.project-desc { font-size: 13px; color: var(--color-text-muted); margin-bottom: 6px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.project-open-cards { font-size: 12px; color: var(--color-text-muted); margin-bottom: 12px; }

.project-actions { display: flex; justify-content: flex-end; }

.drag-handle {
  position: absolute;
  top: 6px;
  left: 6px;
  font-size: 14px;
  color: var(--color-text-muted);
  cursor: grab;
  opacity: 0;
  transition: opacity .15s;
  line-height: 1;
}
.project-card:hover .drag-handle { opacity: 0.6; }
.drag-handle:active { cursor: grabbing; }

.loading-state { display: flex; justify-content: center; padding: 60px; }
.empty-state { text-align: center; padding: 60px; color: var(--color-text-muted); }
</style>
