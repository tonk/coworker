import { ref } from 'vue'

const STORAGE_KEY = 'chat_notify'
const enabled = ref(localStorage.getItem(STORAGE_KEY) !== 'false')
const isTauri = !!window.__TAURI_INTERNALS__

// Resolved to sendNotification once Tauri permission is granted
let tauriSend = null

async function requestPermission() {
  if (isTauri) {
    try {
      const { isPermissionGranted, requestPermission: reqPerm, sendNotification } =
        await import('@tauri-apps/plugin-notification')
      let granted = await isPermissionGranted()
      if (!granted) {
        const result = await reqPerm()
        granted = result === 'granted'
      }
      if (granted) tauriSend = sendNotification
    } catch {}
  } else if ('Notification' in window && Notification.permission === 'default') {
    Notification.requestPermission()
  }
}

if (enabled.value) requestPermission()

function toggle() {
  enabled.value = !enabled.value
  localStorage.setItem(STORAGE_KEY, enabled.value ? 'true' : 'false')
  if (enabled.value) requestPermission()
}

function desktopNotify(title, body) {
  if (!enabled.value) return
  if (document.hasFocus()) return
  if (isTauri) {
    tauriSend?.({ title, body })
    return
  }
  if (!('Notification' in window) || Notification.permission !== 'granted') return
  new Notification(title, { body, icon: '/favicon.ico' })
}

export function useChatNotify() {
  return { notifyEnabled: enabled, toggleNotify: toggle, desktopNotify }
}
