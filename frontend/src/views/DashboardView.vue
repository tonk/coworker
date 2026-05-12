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

        <!-- ── Dashboard news widgets ─────────────────────────────────────── -->
        <template v-if="visibleNews.length">
          <div class="dashboard-widgets">
            <div
              v-for="item in visibleNews"
              :key="item.id"
              class="widget news-widget"
            >
              <button class="widget-dismiss" :aria-label="$t('common.close')" @click="dismissNewsItem(item.id)">×</button>
              <div class="widget-tag">{{ $t('dashboard.news_title') }}</div>
              <h2 class="widget-title">{{ item.title }}</h2>
              <p class="widget-date">{{ formatNewsDate(item.created_at) }}</p>
              <p class="widget-body">{{ item.text }}</p>
            </div>
          </div>
        </template>

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
                <span class="board-type-badge" :class="project.board_type || 'kanban'">{{ $t(`sprint.board_type_${project.board_type || 'kanban'}`) }}</span>
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

  <BaseModal v-if="showCreate" :title="$t('project.new_project')" @close="showCreate = false; newProject.key_prefix = ''; prefixTouched = false; customerError = false">
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
          <select class="form-input" :class="{ 'input-error': customerError }" v-model="newProject.customer_id" style="max-width:400px" @change="customerError = false">
            <option :value="null" disabled>— {{ $t('project.customer') }} —</option>
            <option v-for="c in createCustomers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
          <p v-if="customerError" class="field-error">{{ $t('project.customer_required') }}</p>
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
        <button class="btn btn-primary" @click="handleCreate" :disabled="creating || !newProject.key_prefix.trim()">{{ $t('project.create') }}</button>
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
import { newsApi } from '@/api/news'

const router = useRouter()
const projectStore = useProjectStore()
const ui = useUIStore()
const sidebarStore = useSidebarStore()
const auth = useAuthStore()
const showCreate = ref(false)

// ── News widgets ──────────────────────────────────────────────────────────────
const NEWS_DISMISSED_KEY = 'dashboard_news_dismissed_ids'
const allNews = ref([])

function getDismissedIds() {
  try { return new Set(JSON.parse(localStorage.getItem(NEWS_DISMISSED_KEY) || '[]')) } catch { return new Set() }
}
const dismissedIds = ref(getDismissedIds())

const visibleNews = computed(() =>
  allNews.value.filter(n => !dismissedIds.value.has(n.id))
)

function dismissNewsItem(id) {
  dismissedIds.value = new Set([...dismissedIds.value, id])
  try { localStorage.setItem(NEWS_DISMISSED_KEY, JSON.stringify([...dismissedIds.value])) } catch {}
}

function formatNewsDate(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' })
}
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
const customerError = ref(false)

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
  newsApi.listActive().then(r => { allNews.value = r.data || [] }).catch(() => {})
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
    customerError.value = true
    return
  }
  creating.value = true
  try {
    const project = await projectStore.createProject(newProject.value)
    showCreate.value = false
    newProject.value = { name: '', description: '', color: '#6366f1', key_prefix: '', board_type: 'kanban', customer_id: null }
    prefixTouched.value = false
    customerError.value = false
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
.board-type-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 9999px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  flex-shrink: 0;
  margin-left: auto;
}
.board-type-badge.scrum {
  background: color-mix(in srgb, var(--color-primary) 15%, transparent);
  color: var(--color-primary);
}
.board-type-badge.kanban {
  background: color-mix(in srgb, var(--color-success) 15%, transparent);
  color: var(--color-success);
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

/* ── Dashboard news widgets ──────────────────────────────────────────────── */
.dashboard-widgets {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  margin-bottom: 28px;
}

.widget {
  position: relative;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-left: 3px solid var(--color-primary);
  border-radius: var(--radius);
  padding: 18px 20px;
}

.widget-tag {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .07em;
  color: var(--color-primary);
  margin-bottom: 8px;
}

.widget-title {
  font-size: 15px;
  font-weight: 700;
  margin-bottom: 4px;
  line-height: 1.3;
  padding-right: 24px;
}
.widget-date {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-bottom: 10px;
}
.widget-body {
  font-size: 13px;
  color: var(--color-text-muted);
  line-height: 1.55;
  white-space: pre-wrap;
}

.widget-dismiss {
  position: absolute;
  top: 10px;
  right: 12px;
  background: transparent;
  border: none;
  color: var(--color-text-muted);
  font-size: 16px;
  line-height: 1;
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
}
.widget-dismiss:hover { color: var(--color-text); background: var(--color-bg); }

.input-error { border-color: var(--color-danger, #ef4444) !important; }
.field-error { margin-top: 4px; font-size: 12px; color: var(--color-danger, #ef4444); }
</style>
