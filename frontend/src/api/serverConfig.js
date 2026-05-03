const KEY = 'warmdesk_server_url'
const RUNTIME_KEY = 'warmdesk_runtime_server_url'
let _externalImageProxyEnabled = true

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

export function setExternalImageProxyEnabled(enabled) {
  _externalImageProxyEnabled = enabled !== false
}

// Resolve a server-relative asset URL (e.g. /uploads/...) to an absolute URL
// when running in Tauri/desktop mode where relative paths don't resolve correctly.
export function resolveAssetUrl(url) {
  const raw = typeof url === 'string' ? url.trim() : ''
  if (!raw) return url

  // Keep data/blob URLs untouched.
  if (raw.startsWith('data:') || raw.startsWith('blob:')) {
    return raw
  }

  // Absolute URL or protocol-relative URL:
  // if same-origin => use directly, else route via same-origin proxy for CSP.
  if (/^(https?:)?\/\//i.test(raw)) {
    const protocol = typeof window !== 'undefined' ? window.location.protocol : 'https:'
    const absolute = raw.startsWith('//')
      ? `${protocol}${raw}`
      : raw
    try {
      const urlObj = new URL(absolute)
      const sameAsPage = typeof window !== 'undefined' && urlObj.origin === window.location.origin
      const server = getServerUrl()
      let sameAsServer = false
      if (server) {
        try { sameAsServer = urlObj.origin === new URL(server).origin } catch {}
      }
      if (sameAsPage || sameAsServer) return absolute
      if (!_externalImageProxyEnabled) return absolute
      const proxyPath = `/api/v1/media/proxy?url=${encodeURIComponent(absolute)}`
      return server ? `${server}${proxyPath}` : proxyPath
    } catch {
      return raw
    }
  }

  const server = getServerUrl()
  if (!server) return raw

  try {
    // Handles both '/uploads/a.png' and 'uploads/a.png' safely.
    return new URL(raw, `${server}/`).toString()
  } catch {
    return raw
  }
}

// Convert an HTTP/HTTPS server URL to a WebSocket URL for the given path.
// e.g. https://warmdesk.example.com + /api/v1/ws/slug → wss://warmdesk.example.com/api/v1/ws/slug
export function getWsUrl(path) {
  const base = getServerUrl()
  if (!base) return null
  const wsBase = base.replace(/^http/, 'ws') // http→ws, https→wss
  return `${wsBase}${path}`
}
