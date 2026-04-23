const KEY = 'warmdesk_server_url'
const RUNTIME_KEY = 'warmdesk_runtime_server_url'

export function getServerUrl() {
  const runtimeOverride = typeof window !== 'undefined'
    ? window.__WARMDESK_RUNTIME_SERVER_URL__
    : ''
  const runtimeSession = typeof window !== 'undefined'
    ? sessionStorage.getItem(RUNTIME_KEY)
    : ''
  return runtimeOverride || runtimeSession || localStorage.getItem(KEY) || ''
}

export function setServerUrl(url) {
  // Normalize: strip trailing slash
  localStorage.setItem(KEY, url.replace(/\/+$/, ''))
}

export function clearServerUrl() {
  localStorage.removeItem(KEY)
}

export function setRuntimeServerUrl(url) {
  const normalized = (url || '').trim().replace(/\/+$/, '')
  if (!normalized) return
  if (typeof window !== 'undefined') {
    window.__WARMDESK_RUNTIME_SERVER_URL__ = normalized
    sessionStorage.setItem(RUNTIME_KEY, normalized)
  }
}

export function isServerConfigured() {
  return !!getServerUrl()
}

// Resolve a server-relative asset URL (e.g. /uploads/...) to an absolute URL
// when running in Tauri/desktop mode where relative paths don't resolve correctly.
export function resolveAssetUrl(url) {
  if (!url || url.startsWith('http') || url.startsWith('data:') || url.startsWith('blob:')) return url
  const server = getServerUrl()
  return server ? `${server}${url}` : url
}

// Convert an HTTP/HTTPS server URL to a WebSocket URL for the given path.
// e.g. https://warmdesk.example.com + /api/v1/ws/slug → wss://warmdesk.example.com/api/v1/ws/slug
export function getWsUrl(path) {
  const base = getServerUrl()
  if (!base) return null
  const wsBase = base.replace(/^http/, 'ws') // http→ws, https→wss
  return `${wsBase}${path}`
}
