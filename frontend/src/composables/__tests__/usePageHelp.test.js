import { describe, it, expect, beforeEach } from 'vitest'
import { ref } from 'vue'
import { ROUTE_HELP, CONTEXT_HELP } from '@/composables/usePageHelp'

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
})
