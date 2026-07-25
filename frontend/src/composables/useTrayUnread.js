import { watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'
import { useTicketsStore } from '@/stores/tickets'
import { useSystemStore } from '@/stores/system'
import { useProjectChatUnread } from '@/composables/useProjectChatUnread'

const isTauri = !!window.__TAURI_INTERNALS__

export function useTrayUnread() {
  const auth = useAuthStore()
  const notificationsStore = useNotificationsStore()
  const ticketsStore = useTicketsStore()
  const systemStore = useSystemStore()
  const { projectChatUnread } = useProjectChatUnread()

  watch(
    [
      () => notificationsStore.hasUnread,
      projectChatUnread,
      () => ticketsStore.inboxUnread,
      () => auth.user?.tray_icon_enabled,
      () => auth.user?.close_to_tray_enabled,
      () => systemStore.isTimetrackingMode,
      () => systemStore.isTestInstance,
    ],
    async () => {
      if (!isTauri || !auth.isLoggedIn) return
      const enabled = auth.user?.tray_icon_enabled !== false
      const closeToTray = auth.user?.close_to_tray_enabled !== false
      const convUnread = notificationsStore.hasUnread ? 1 : 0
      const total = convUnread + projectChatUnread.value + ticketsStore.inboxUnread
      try {
        const { invoke } = await import('@tauri-apps/api/core')
        await invoke('set_tray_unread', {
          count: total,
          enabled,
          isTimetracking: systemStore.isTimetrackingMode,
          isTest: systemStore.isTestInstance,
          closeToTray,
        })
      } catch {
        // Tray command may fail if tray was not created (non-Tauri)
      }
    },
    { immediate: true, deep: true },
  )
}
