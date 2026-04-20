import client from './client'
import { getServerUrl } from './serverConfig'

export const attachmentsApi = {
  upload: (formData) => client.post('/attachments', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  delete: (id) => client.delete(`/attachments/${id}`),
  // Returns an absolute URL when a server is configured (Tauri/desktop mode)
  // so that <img src="..."> and <a href="..."> resolve correctly outside the browser.
  // Appends ?token= so <img> tags (which can't send Authorization headers) still authenticate.
  url: (id) => {
    const server = getServerUrl()
    const base = server ? `${server}/api/v1/attachments/${id}` : `/api/v1/attachments/${id}`
    const token = sessionStorage.getItem('access_token') || ''
    return token ? `${base}?token=${encodeURIComponent(token)}` : base
  },
  // Upload an image file (avatar, logo, etc.) and return the server path (/uploads/...).
  uploadImage: (file) => {
    const fd = new FormData()
    fd.append('file', file)
    return client.post('/upload/image', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
  }
}
