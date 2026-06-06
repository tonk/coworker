import { describe, it, expect } from 'vitest'
import { createI18n } from 'vue-i18n'
import en from '@/i18n/en.json'
import { ROUTE_HELP, CONTEXT_HELP, messageList } from '@/composables/usePageHelp'

describe('usePageHelp maps', () => {
  it('covers primary routes', () => {
    expect(ROUTE_HELP.board).toBe('board')
    expect(ROUTE_HELP['time-tracking']).toBe('timeTracking')
    expect(ROUTE_HELP['ticket-detail']).toBe('ticketDetail')
    expect(ROUTE_HELP['inbox-ticket-detail']).toBe('ticketDetail')
  })

  it('maps admin tab contexts', () => {
    expect(CONTEXT_HELP['admin.settings']).toBe('adminSettings')
    expect(CONTEXT_HELP['admin.macros']).toBe('adminMacros')
  })

  it('maps time tracking tab contexts', () => {
    expect(CONTEXT_HELP['timeTracking.sheet']).toBe('timeTrackingSheet')
    expect(CONTEXT_HELP['timeTracking.board-report']).toBe('timeTrackingBoardReport')
  })

  it('maps phase 1 sub-contexts', () => {
    expect(CONTEXT_HELP.cardDetail).toBe('cardDetail')
    expect(CONTEXT_HELP['projectSettings.webhooks']).toBe('projectSettingsWebhooks')
    expect(CONTEXT_HELP['settings.mfa']).toBe('settingsMfa')
    expect(CONTEXT_HELP['charts.burndown']).toBe('chartsBurndown')
  })

  it('messageList resolves array-valued i18n keys', () => {
    const { tm, te } = createI18n({ legacy: false, locale: 'en', messages: { en } }).global
    const key = 'help.pages.timeTrackingSheet.tasks'
    expect(te(key)).toBe(false)
    const tasks = messageList(tm, key)
    expect(tasks.length).toBe(5)
    expect(tasks[4]).toContain('undeclarable')
  })
})
