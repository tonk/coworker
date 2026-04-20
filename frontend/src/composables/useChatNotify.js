import { ref } from 'vue'

const STORAGE_KEY = 'chat_notify'
const enabled = ref(localStorage.getItem(STORAGE_KEY) !== 'false')

function requestPermission() {
  if ('Notification' in window && Notification.permission === 'default') {
    Notification.requestPermission()
  }
}

// Ask on first load if notifications are already enabled
if (enabled.value) requestPermission()

function toggle() {
  enabled.value = !enabled.value
  localStorage.setItem(STORAGE_KEY, enabled.value ? 'true' : 'false')
  if (enabled.value) requestPermission()
}

function desktopNotify(title, body) {
  if (!enabled.value) return
  if (!('Notification' in window)) return
  if (Notification.permission !== 'granted') return
  if (document.hasFocus()) return
  new Notification(title, { body, icon: '/favicon.ico' })
}

export function useChatNotify() {
  return { notifyEnabled: enabled, toggleNotify: toggle, desktopNotify }
}
