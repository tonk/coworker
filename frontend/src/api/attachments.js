import client from './client'
import { getServerUrl } from './serverConfig'

// ── Media ticket cache (Tauri only) ──────────────────────────────────────────
// Short-lived tickets keep the long-lived JWT out of <img src="..."> URLs.
// Tickets are fetched once per session and refreshed every 4 minutes (5-min TTL).
let _mediaTicket = null
let _mediaTicketExpiry = 0
let _mediaTicketTimer = null

export async function refreshMediaTicket() {
  if (!window.__TAURI_INTERNALS__) return
  try {
    const { data } = await client.post('/auth/media-ticket')
    _mediaTicket = data.ticket
    // Treat the ticket as expired 30 s before its actual TTL to avoid races.
    _mediaTicketExpiry = Date.now() + (5 * 60 - 30) * 1000
  } catch {
    // Silently ignore — url() falls back to no-ticket and the image will 401
  }
}

export function startMediaTicketRefresh() {
  if (!window.__TAURI_INTERNALS__) return
  stopMediaTicketRefresh()
  _mediaTicketTimer = setInterval(refreshMediaTicket, 4 * 60 * 1000)
}

export function stopMediaTicketRefresh() {
  if (_mediaTicketTimer) {
    clearInterval(_mediaTicketTimer)
    _mediaTicketTimer = null
  }
  _mediaTicket = null
  _mediaTicketExpiry = 0
}

export const attachmentsApi = {
  upload: (formData) => client.post('/attachments', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  delete: (id) => client.delete(`/attachments/${id}`),
  // Returns an absolute URL when a server is configured (Tauri/desktop mode)
  // so that <img src="..."> and <a href="..."> resolve correctly outside the browser.
  // In Tauri mode appends ?ticket= (short-lived media ticket) instead of the full JWT.
  // In browser mode the httpOnly access_token cookie is sent automatically.
  url: (id) => {
    const server = getServerUrl()
    const base = server ? `${server}/api/v1/attachments/${id}` : `/api/v1/attachments/${id}`
    if (window.__TAURI_INTERNALS__) {
      if (_mediaTicket && Date.now() < _mediaTicketExpiry) {
        return `${base}?ticket=${encodeURIComponent(_mediaTicket)}`
      }
      return base
    }
    return base
  },
  // Upload an image file (avatar, logo, etc.) and return the server path (/uploads/...).
  uploadImage: (file) => {
    const fd = new FormData()
    fd.append('file', file)
    return client.post('/upload/image', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
  }
}
