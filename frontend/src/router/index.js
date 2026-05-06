import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { isServerConfigured } from '@/api/serverConfig'
import { lazyWithReload } from '@/router/lazyWithReload'

const routes = [
  { path: '/connect', name: 'connect', component: lazyWithReload(() => import('@/views/ConnectView.vue')), meta: { public: true, serverIndependent: true } },
  { path: '/login', name: 'login', component: lazyWithReload(() => import('@/views/LoginView.vue')), meta: { public: true } },
  { path: '/register', name: 'register', component: lazyWithReload(() => import('@/views/RegisterView.vue')), meta: { public: true } },
  { path: '/forgot-password', name: 'forgot-password', component: lazyWithReload(() => import('@/views/ForgotPasswordView.vue')), meta: { public: true } },
  { path: '/reset-password', name: 'reset-password', component: lazyWithReload(() => import('@/views/ResetPasswordView.vue')), meta: { public: true } },
  { path: '/', name: 'dashboard', component: lazyWithReload(() => import('@/views/DashboardView.vue')) },
  { path: '/projects/:slug', name: 'board', component: lazyWithReload(() => import('@/views/BoardView.vue')) },
  { path: '/projects/:slug/settings', name: 'project-settings', component: lazyWithReload(() => import('@/views/ProjectSettingsView.vue')) },
  { path: '/projects/:slug/topics', name: 'topics', component: lazyWithReload(() => import('@/views/TopicsView.vue')) },
  { path: '/projects/:slug/gantt', name: 'gantt', component: lazyWithReload(() => import('@/views/GanttView.vue')) },
  { path: '/projects/:slug/backlog', name: 'backlog', component: lazyWithReload(() => import('@/views/BacklogView.vue')) },
  { path: '/projects/:slug/sprint', name: 'sprint-board', component: lazyWithReload(() => import('@/views/SprintBoardView.vue')) },
  { path: '/projects/:slug/charts', name: 'charts', component: lazyWithReload(() => import('@/views/ChartsView.vue')) },
  { path: '/settings', name: 'settings', component: lazyWithReload(() => import('@/views/SettingsView.vue')) },
  { path: '/chats', name: 'chats', component: lazyWithReload(() => import('@/views/DirectMessagesView.vue')) },
  { path: '/messages', redirect: '/chats' },
  { path: '/admin', name: 'admin', component: lazyWithReload(() => import('@/views/AdminView.vue')), meta: { adminOnly: true } },
  { path: '/reports', name: 'reports', component: lazyWithReload(() => import('@/views/ReportView.vue')), meta: { reportsOnly: true } },
  { path: '/time-tracking', name: 'time-tracking', component: lazyWithReload(() => import('@/views/TimeTrackingView.vue')), meta: { timeTrackingOnly: true } },
  { path: '/customers', name: 'customers', component: lazyWithReload(() => import('@/views/CustomersView.vue')) },
  { path: '/customers/:id', name: 'customer-detail', component: lazyWithReload(() => import('@/views/CustomerDetailView.vue')) },
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
  if (to.meta.reportsOnly && !auth.canViewReports) return '/'
  if (to.meta.timeTrackingOnly && !auth.timeTrackingEnabled) return '/'
  if (to.meta.public && auth.isLoggedIn && (to.name === 'login' || to.name === 'register')) return '/'
  return true
})

export default router
