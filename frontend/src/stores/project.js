import { defineStore } from 'pinia'
import { ref } from 'vue'
import { projectsApi } from '@/api/projects'

export const useProjectStore = defineStore('project', () => {
  const projects = ref([])
  const currentProject = ref(null)
  const loading = ref(false)

  async function fetchProjects() {
    loading.value = true
    try {
      const { data } = await projectsApi.list()
      projects.value = data
    } finally {
      loading.value = false
    }
  }

  async function fetchProject(slug) {
    loading.value = true
    try {
      const { data } = await projectsApi.get(slug)
      currentProject.value = data
      return data
    } finally {
      loading.value = false
    }
  }

  async function createProject(payload) {
    const { data } = await projectsApi.create(payload)
    projects.value.push(data)
    return data
  }

  async function updateProject(slug, payload) {
    const { data } = await projectsApi.update(slug, payload)
    const idx = projects.value.findIndex(p => p.slug === slug)
    if (idx !== -1) projects.value[idx] = data
    if (currentProject.value?.slug === slug) currentProject.value = data
    return data
  }

  async function deleteProject(slug) {
    await projectsApi.delete(slug)
    projects.value = projects.value.filter(p => p.slug !== slug)
    if (currentProject.value?.slug === slug) currentProject.value = null
  }

  async function reorderProjects(ordered) {
    // ordered: array of project objects in the desired new order
    const items = ordered.map((p, i) => ({ id: p.id, position: i }))
    projects.value = ordered
    await projectsApi.reorder(items)
  }

  return { projects, currentProject, loading, fetchProjects, fetchProject, createProject, updateProject, deleteProject, reorderProjects }
})
