import client from './client'
import { getServerUrl } from './serverConfig'

export const attachmentsApi = {
  upload: (formData) => client.post('/attachments', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  delete: (id) => client.delete(`/attachments/${id}`),
  // Returns an absolute URL when a server is configured (Tauri/desktop mode)
  // so that <img src="..."> and <a href="..."> resolve correctly outside the browser.
  // In Tauri mode appends ?token= because <img> tags can't set headers and the
  // native HTTP client has no browser cookie jar. In browser mode the httpOnly
  // access_token cookie is sent automatically with every request.
  url: (id) => {
    const server = getServerUrl()
    const base = server ? `${server}/api/v1/attachments/${id}` : `/api/v1/attachments/${id}`
    if (window.__TAURI_INTERNALS__) {
      const token = sessionStorage.getItem('access_token') || ''
      return token ? `${base}?token=${encodeURIComponent(token)}` : base
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
