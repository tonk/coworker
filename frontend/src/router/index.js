import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { isServerConfigured } from '@/api/serverConfig'

const routes = [
  { path: '/connect', name: 'connect', component: () => import('@/views/ConnectView.vue'), meta: { public: true, serverIndependent: true } },
  { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
  { path: '/register', name: 'register', component: () => import('@/views/RegisterView.vue'), meta: { public: true } },
  { path: '/', name: 'dashboard', component: () => import('@/views/DashboardView.vue') },
  { path: '/projects/:slug', name: 'board', component: () => import('@/views/BoardView.vue') },
  { path: '/projects/:slug/settings', name: 'project-settings', component: () => import('@/views/ProjectSettingsView.vue') },
  { path: '/projects/:slug/topics', name: 'topics', component: () => import('@/views/TopicsView.vue') },
  { path: '/settings', name: 'settings', component: () => import('@/views/SettingsView.vue') },
  { path: '/messages', name: 'messages', component: () => import('@/views/DirectMessagesView.vue') },
  { path: '/admin', name: 'admin', component: () => import('@/views/AdminView.vue'), meta: { adminOnly: true } },
  { path: '/:pathMatch(.*)*', redirect: '/' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  // In Tauri (desktop) mode a server URL must be configured first.
  // In a regular browser the app is served from the server, so no config needed.
  if (window.__TAURI_INTERNALS__ && !to.meta.serverIndependent && !isServerConfigured()) {
    return '/connect'
  }

  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLoggedIn) return '/login'
  if (to.meta.adminOnly && !auth.isAdmin) return '/'
  if (to.meta.public && auth.isLoggedIn && (to.name === 'login' || to.name === 'register')) return '/'
  return true
})

export default router
