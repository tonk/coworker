import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUIStore } from '@/stores/ui'

/** Maps vue-router route.name → help.pages.* key (camelCase). */
const ROUTE_HELP = {
  dashboard: 'dashboard',
  board: 'board',
  'project-settings': 'projectSettings',
  topics: 'topics',
  gantt: 'gantt',
  backlog: 'backlog',
  epics: 'epics',
  'sprint-board': 'sprintBoard',
  charts: 'charts',
  news: 'news',
  settings: 'settings',
  chats: 'chats',
  admin: 'admin',
  'time-tracking': 'timeTracking',
  customers: 'customers',
  'customer-detail': 'customerDetail',
  tickets: 'tickets',
  'ticket-detail': 'ticketDetail',
  inbox: 'inbox',
  'inbox-ticket-detail': 'ticketDetail',
}

/** Maps ui.helpContext → help.pages.* key for tabbed views and modals. */
const CONTEXT_HELP = {
  cardDetail: 'cardDetail',

  'projectSettings.general': 'projectSettingsGeneral',
  'projectSettings.members': 'projectSettingsMembers',
  'projectSettings.labels': 'projectSettingsLabels',
  'projectSettings.apikeys': 'projectSettingsApiKeys',
  'projectSettings.webhooks': 'projectSettingsWebhooks',
  'projectSettings.deletedcards': 'projectSettingsDeletedCards',

  'settings.profile': 'settingsProfile',
  'settings.workingHours': 'settingsWorkingHours',
  'settings.password': 'settingsPassword',
  'settings.mfa': 'settingsMfa',
  'settings.passkey': 'settingsPasskey',
  'settings.apiKeys': 'settingsApiKeys',

  'charts.velocity': 'chartsVelocity',
  'charts.throughput': 'chartsThroughput',
  'charts.burndown': 'chartsBurndown',
  'charts.burnup': 'chartsBurnup',
  'charts.cfd': 'chartsCfd',
  'charts.sprint-report': 'chartsSprintReport',
  'charts.epic-burndown': 'chartsEpicBurndown',
  'charts.release-burndown': 'chartsReleaseBurndown',
  'charts.cycle-time': 'chartsCycleTime',
  'charts.lead-time': 'chartsLeadTime',

  'admin.users': 'adminUsers',
  'admin.groups': 'adminGroups',
  'admin.customers': 'adminCustomers',
  'admin.projects': 'adminProjects',
  'admin.settings': 'adminSettings',
  'admin.news': 'adminNews',
  'admin.time-tracking': 'adminTimeTracking',
  'admin.sla': 'adminSla',
  'admin.macros': 'adminMacros',
  'admin.ticket-checklists': 'adminTicketChecklists',
  'admin.backup': 'adminBackup',
  'timeTracking.sheet': 'timeTrackingSheet',
  'timeTracking.report': 'timeTrackingReport',
  'timeTracking.board-report': 'timeTrackingBoardReport',
}

function messageList(tm, te, key) {
  if (!te(key)) return []
  const msgs = tm(key)
  return Array.isArray(msgs) ? msgs : []
}

/**
 * Resolves contextual in-app help for the current route (and optional tab context).
 * Content lives in i18n under help.pages.{key}.*
 */
export function usePageHelp() {
  const route = useRoute()
  const ui = useUIStore()
  const { t, te, tm } = useI18n()

  const helpKey = computed(() => {
    if (ui.helpContext) {
      const fromContext = CONTEXT_HELP[ui.helpContext]
      if (fromContext) return fromContext
    }
    return ROUTE_HELP[route.name] || 'default'
  })

  const pageBase = computed(() => `help.pages.${helpKey.value}`)

  const resolvedBase = computed(() => {
    const base = pageBase.value
    if (te(`${base}.title`)) return base
    if (helpKey.value !== 'default' && te('help.pages.default.title')) return 'help.pages.default'
    return base
  })

  const title = computed(() => t(`${resolvedBase.value}.title`))
  const intro = computed(() => (te(`${resolvedBase.value}.intro`) ? t(`${resolvedBase.value}.intro`) : ''))
  const tasks = computed(() => messageList(tm, te, `${resolvedBase.value}.tasks`))
  const shortcuts = computed(() => messageList(tm, te, `${resolvedBase.value}.shortcuts`))

  return { helpKey, title, intro, tasks, shortcuts }
}

/** Exported for unit tests. */
export { ROUTE_HELP, CONTEXT_HELP }
