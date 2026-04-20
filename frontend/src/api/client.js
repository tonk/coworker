import axios from 'axios'
import { getServerUrl } from './serverConfig'

// Base URL is resolved at request time so it picks up runtime server config.
// Falls back to relative '/api/v1' for the normal browser / Vite-proxy workflow.
function apiBase() {
  const server = getServerUrl()
  return server ? `${server}/api/v1` : '/api/v1'
}

// Use the fetch adapter so requests go through window.fetch.
// index.html installs a proxy for window.fetch before the bundle loads;
// on Windows Tauri that proxy is pointed at tauri-plugin-http by main.js,
// routing every request through the native Rust HTTP client.
const client = axios.create({
  headers: {
    'Content-Type': 'application/json',
    // Some WAFs/proxies block non-browser user-agents on POST endpoints.
    // When running inside Tauri the request goes through tauri-plugin-http
    // (reqwest) which sets its own UA; override it here so the server sees
    // a normal browser string instead.
    'User-Agent': navigator.userAgent,
    // Let the backend distinguish desktop app requests from browser requests.
    // Only set inside the Tauri runtime; absent = web browser.
    ...(window.__TAURI_INTERNALS__ && {
      'X-WarmDesk-Client': `tauri/${__APP_VERSION__}/${navigator.platform || 'unknown'}`,
    }),
  },
  adapter: 'fetch',
})

let isRefreshing = false
let refreshQueue = []

function processQueue(error, token = null) {
  refreshQueue.forEach(({ resolve, reject }) => {
    if (error) reject(error)
    else resolve(token)
  })
  refreshQueue = []
}

client.interceptors.request.use(config => {
  config.baseURL = apiBase()
  const token = sessionStorage.getItem('access_token')
    || (client.defaults.headers.common.Authorization || '').replace('Bearer ', '')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

client.interceptors.response.use(
  response => response,
  async error => {
    const original = error.config
    if (error.response?.status === 401 && !original._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          refreshQueue.push({ resolve, reject })
        }).then(token => {
          original.headers.Authorization = `Bearer ${token}`
          return client(original)
        })
      }

      original._retry = true
      isRefreshing = true

      const refreshToken = sessionStorage.getItem('refresh_token')
      if (!refreshToken) {
        isRefreshing = false
        sessionStorage.removeItem('access_token')
        sessionStorage.removeItem('refresh_token')
        window.location.href = '/login'
        return Promise.reject(error)
      }

      try {
        const { data } = await axios.post(`${apiBase()}/auth/refresh`, { refresh_token: refreshToken })
        sessionStorage.setItem('access_token', data.access_token)
        sessionStorage.setItem('refresh_token', data.refresh_token)
        client.defaults.headers.common.Authorization = `Bearer ${data.access_token}`
        processQueue(null, data.access_token)
        original.headers.Authorization = `Bearer ${data.access_token}`
        return client(original)
      } catch (err) {
        processQueue(err, null)
        sessionStorage.removeItem('access_token')
        sessionStorage.removeItem('refresh_token')
        window.location.href = '/login'
        return Promise.reject(err)
      } finally {
        isRefreshing = false
      }
    }
    return Promise.reject(error)
  }
)

export default client
