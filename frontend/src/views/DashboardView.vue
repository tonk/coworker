<template>
  <main class="dashboard-main">
      <div class="dashboard-container">
        <div class="dashboard-header">
          <h1>{{ $t('project.projects') }}</h1>
          <button class="btn btn-primary" @click="showCreate = true">+ {{ $t('project.new_project') }}</button>
        </div>

        <div v-if="projectStore.loading" class="loading-state">
          <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
        </div>

        <div v-else-if="projectStore.projects.length === 0" class="empty-state">
          <p>{{ $t('project.no_projects') }}</p>
        </div>

        <div v-else class="projects-grid" ref="gridEl">
          <div
            v-for="project in projectStore.projects"
            :key="project.id"
            class="project-card"
            :data-id="project.id"
            @click="router.push(`/projects/${project.slug}`)"
          >
            <div v-if="isAdmin" class="drag-handle" title="Drag to reorder" @click.stop>⠿</div>
            <div class="project-color-bar" :style="{ background: project.color || '#6366f1' }"></div>
            <div class="project-card-body">
              <h3>{{ project.name }}</h3>
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

  <BaseModal v-if="showCreate" :title="$t('project.new_project')" @close="showCreate = false">
      <form @submit.prevent="handleCreate">
        <div class="form-group">
          <label class="form-label">{{ $t('project.project_name') }}</label>
          <input class="form-input" v-model="newProject.name" required autofocus />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('project.description') }}</label>
          <textarea class="form-input" v-model="newProject.description" rows="3"></textarea>
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('project.color') }}</label>
          <input type="color" class="form-input" v-model="newProject.color" style="height:40px;padding:4px" />
        </div>
      </form>
      <template #footer>
        <button class="btn btn-secondary" @click="showCreate = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="handleCreate" :disabled="creating">{{ $t('project.create') }}</button>
      </template>
  </BaseModal>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import Sortable from 'sortablejs'
import BaseModal from '@/components/common/BaseModal.vue'
import { useProjectStore } from '@/stores/project'
import { useUIStore } from '@/stores/ui'
import { useSidebarStore } from '@/stores/sidebar'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const projectStore = useProjectStore()
const ui = useUIStore()
const sidebarStore = useSidebarStore()
const auth = useAuthStore()
const showCreate = ref(false)
const gridEl = ref(null)
let sortable = null

const isAdmin = computed(() => auth.user?.global_role === 'admin')
const creating = ref(false)
const newProject = ref({ name: '', description: '', color: '#6366f1' })

onMounted(async () => {
  await projectStore.fetchProjects()
  sidebarStore.fetchStarred()
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

async function handleCreate() {
  if (!newProject.value.name) return
  creating.value = true
  try {
    const project = await projectStore.createProject(newProject.value)
    showCreate.value = false
    newProject.value = { name: '', description: '', color: '#6366f1' }
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
