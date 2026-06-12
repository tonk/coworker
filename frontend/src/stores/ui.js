import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUIStore = defineStore('ui', () => {
  const toasts = ref([])
  let nextId = 0

  function toast(message, type = 'info', duration = 3000) {
    const id = ++nextId
    toasts.value.push({ id, message, type })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, duration)
  }

  function success(message) { toast(message, 'success') }
  function error(message) { toast(message, 'error', 5000) }
  function info(message) { toast(message, 'info') }

  function mention(senderName, body, context) {
    const id = ++nextId
    toasts.value.push({ id, type: 'mention', senderName, body, context })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, 6000)
  }

  // Promise-based confirm dialog — options: { confirmLabel?, destructive? }
  const confirmState = ref(null) // { message, confirmLabel, destructive, resolve }

  function confirm(message, options = {}) {
    return new Promise(resolve => {
      confirmState.value = {
        message,
        confirmLabel: options.confirmLabel ?? null,
        destructive: options.destructive === true,
        resolve,
      }
    })
  }

  function _confirmResolve(result) {
    confirmState.value?.resolve(result)
    confirmState.value = null
  }

  /** Optional tab/section id for contextual help, e.g. "admin.settings". */
  const helpContext = ref(null)

  function setHelpContext(context) {
    helpContext.value = context || null
  }

  return {
    toasts, toast, success, error, info, mention,
    confirmState, confirm, _confirmResolve,
    helpContext, setHelpContext,
  }
})
