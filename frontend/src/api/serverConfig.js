const KEY = 'coworker_server_url'

export function getServerUrl() {
  return localStorage.getItem(KEY) || ''
}

export function setServerUrl(url) {
  // Normalize: strip trailing slash
  localStorage.setItem(KEY, url.replace(/\/+$/, ''))
}

export function clearServerUrl() {
  localStorage.removeItem(KEY)
}

export function isServerConfigured() {
  return !!getServerUrl()
}

// Convert an HTTP/HTTPS server URL to a WebSocket URL for the given path.
// e.g. https://coworker.example.com + /api/v1/ws/slug → wss://coworker.example.com/api/v1/ws/slug
export function getWsUrl(path) {
  const base = getServerUrl()
  if (!base) return null
  const wsBase = base.replace(/^http/, 'ws') // http→ws, https→wss
  return `${wsBase}${path}`
}
