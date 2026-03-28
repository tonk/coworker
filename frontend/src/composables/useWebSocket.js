import { ref, onUnmounted } from 'vue'
import { useBoardStore } from '@/stores/board'
import { useChatStore } from '@/stores/chat'
import { useTopicsStore } from '@/stores/topics'
import { getWsUrl } from '@/api/serverConfig'

export function useWebSocket(projectSlug) {
  const ws = ref(null)
  const connected = ref(false)
  const presenceUsers = ref([])
  let reconnectTimer = null
  let reconnectDelay = 1000

  const boardStore = useBoardStore()
  const chatStore = useChatStore()
  const topicsStore = useTopicsStore()

  function connect() {
    const token = localStorage.getItem('access_token')
    if (!token) return

    // Use the configured server URL when available (Tauri/desktop mode),
    // otherwise fall back to the current page's origin (browser mode).
    const wsUrlFromConfig = getWsUrl(`/api/v1/ws/${projectSlug}?token=${token}`)
    const url = wsUrlFromConfig || (() => {
      const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
      return `${protocol}//${location.host}/api/v1/ws/${projectSlug}?token=${token}`
    })()

    ws.value = new WebSocket(url)

    ws.value.onopen = () => {
      connected.value = true
      reconnectDelay = 1000
    }

    ws.value.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        handleMessage(msg)
      } catch (e) {
        console.error('WS parse error', e)
      }
    }

    ws.value.onclose = () => {
      connected.value = false
      scheduleReconnect()
    }

    ws.value.onerror = () => {
      ws.value?.close()
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer) return
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, reconnectDelay)
    reconnectDelay = Math.min(reconnectDelay * 2, 30000)
  }

  function handleMessage(msg) {
    const { type, payload } = msg

    if (type.startsWith('board.')) {
      boardStore.handleWsEvent(type, payload)
    } else if (type.startsWith('chat.')) {
      chatStore.handleWsEvent(type, payload)
    } else if (type.startsWith('topic.')) {
      topicsStore.handleWsEvent(type, payload)
    } else if (type === 'presence.joined') {
      if (!presenceUsers.value.find(u => u.id === payload.id)) {
        presenceUsers.value.push(payload)
      }
    } else if (type === 'presence.left') {
      presenceUsers.value = presenceUsers.value.filter(u => u.id !== payload.user_id)
    } else if (type === 'presence.list') {
      presenceUsers.value = payload.users
    }
  }

  function send(type, payload, id = null) {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type, payload, id }))
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    ws.value?.close()
    ws.value = null
    connected.value = false
  }

  onUnmounted(disconnect)

  return { connected, presenceUsers, connect, disconnect, send }
}
