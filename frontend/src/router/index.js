import { createRouter, createWebHistory } from 'vue-router'
import { nextTick } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'
import { isServerConfigured } from '@/api/serverConfig'
import { lazyWithReload } from '@/router/lazyWithReload'

const routes = [
  { path: '/connect', name: 'connect', component: lazyWithReload(() => import('@/views/ConnectView.vue')), meta: { public: true, serverIndependent: true } },
  { path: '/login', name: 'login', component: lazyWithReload(() => import('@/views/LoginView.vue')), meta: { public: true } },
  { path: '/register', name: 'register', component: lazyWithReload(() => import('@/views/RegisterView.vue')), meta: { public: true } },
  { path: '/forgot-password', name: 'forgot-password', component: lazyWithReload(() => import('@/views/ForgotPasswordView.vue')), meta: { public: true } },
  { path: '/reset-password', name: 'reset-password', component: lazyWithReload(() => import('@/views/ResetPasswordView.vue')), meta: { public: true } },
  { path: '/', name: 'dashboard', component: lazyWithReload(() => import('@/views/DashboardView.vue')) },
  { path: '/projects/:slug', name: 'board', component: lazyWithReload(() => import('@/views/BoardView.vue')), meta: { boardOnly: true } },
  { path: '/projects/:slug/settings', name: 'project-settings', component: lazyWithReload(() => import('@/views/ProjectSettingsView.vue')), meta: { boardOnly: true } },
  { path: '/projects/:slug/topics', name: 'topics', component: lazyWithReload(() => import('@/views/TopicsView.vue')), meta: { boardOnly: true } },
  { path: '/projects/:slug/gantt', name: 'gantt', component: lazyWithReload(() => import('@/views/GanttView.vue')), meta: { boardOnly: true } },
  { path: '/projects/:slug/backlog', name: 'backlog', component: lazyWithReload(() => import('@/views/BacklogView.vue')), meta: { boardOnly: true } },
  { path: '/projects/:slug/epics', name: 'epics', component: lazyWithReload(() => import('@/views/EpicsView.vue')), meta: { boardOnly: true } },
  { path: '/projects/:slug/sprint', name: 'sprint-board', component: lazyWithReload(() => import('@/views/SprintBoardView.vue')), meta: { boardOnly: true } },
  { path: '/projects/:slug/charts', name: 'charts', component: lazyWithReload(() => import('@/views/ChartsView.vue')), meta: { boardOnly: true } },
  { path: '/news', name: 'news', component: lazyWithReload(() => import('@/views/NewsView.vue')) },
  { path: '/settings', name: 'settings', component: lazyWithReload(() => import('@/views/SettingsView.vue')) },
  { path: '/chats', name: 'chats', component: lazyWithReload(() => import('@/views/DirectMessagesView.vue')), meta: { chatOnly: true } },
  { path: '/messages', redirect: '/chats' },
  { path: '/admin', name: 'admin', component: lazyWithReload(() => import('@/views/AdminView.vue')), meta: { adminOnly: true } },
  { path: '/reports', redirect: '/time-tracking?tab=board-report' },
  { path: '/time-tracking', name: 'time-tracking', component: lazyWithReload(() => import('@/views/TimeTrackingView.vue')), meta: { timeTrackingOnly: true } },
  { path: '/invoices', name: 'invoices', component: lazyWithReload(() => import('@/views/InvoicesView.vue')), meta: { adminOnly: true } },
  { path: '/customers', name: 'customers', component: lazyWithReload(() => import('@/views/CustomersView.vue')) },
  { path: '/customers/:id', name: 'customer-detail', component: lazyWithReload(() => import('@/views/CustomerDetailView.vue')) },
  { path: '/customers/:id/tickets', name: 'tickets', component: lazyWithReload(() => import('@/views/TicketListView.vue')), meta: { helpdeskOnly: true } },
  { path: '/customers/:id/tickets/:ticketId', name: 'ticket-detail', component: lazyWithReload(() => import('@/views/TicketDetailView.vue')), meta: { helpdeskOnly: true } },
  { path: '/tickets/inbox', name: 'inbox', component: lazyWithReload(() => import('@/views/InboxView.vue')), meta: { helpdeskOnly: true } },
  { path: '/tickets/inbox/:ticketId', name: 'inbox-ticket-detail', component: lazyWithReload(() => import('@/views/TicketDetailView.vue')), meta: { helpdeskOnly: true } },
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
  const system = useSystemStore()
  if (!to.meta.public && !auth.isLoggedIn) return '/login'
  if (system.isTimetrackingMode && to.path === '/') return '/time-tracking'
  if (to.meta.adminOnly && !auth.isAdmin) return '/'
  if (to.meta.reportsOnly && !auth.canViewReports) return '/'
  if (to.meta.timeTrackingOnly && !auth.timeTrackingEnabled && !auth.canViewReports) return '/'
  if (to.meta.helpdeskOnly && !auth.helpdeskEnabled) return '/'
  if (to.meta.boardOnly && !auth.boardEnabled) return '/'
  if (to.meta.chatOnly && !auth.chatEnabled) return '/'
  if (to.meta.public && auth.isLoggedIn && (to.name === 'login' || to.name === 'register')) return '/'
  return true
})

// After each navigation, move focus to main content so screen readers
// announce the new page without re-reading the sidebar/header. Skipped when
// a dialog is already open (e.g. the news welcome modal on first load) so
// this doesn't steal focus back from the dialog's own focus-on-open logic.
router.afterEach(() => {
  nextTick(() => {
    if (document.querySelector('[role="dialog"][aria-modal="true"]')) return
    const main = document.getElementById('main-content')
    if (main) main.focus()
  })
})

export default router
