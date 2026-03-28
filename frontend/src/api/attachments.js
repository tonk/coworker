import client from './client'
import { getServerUrl } from './serverConfig'

export const attachmentsApi = {
  upload: (formData) => client.post('/attachments', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  delete: (id) => client.delete(`/attachments/${id}`),
  // Returns an absolute URL when a server is configured (Tauri/desktop mode)
  // so that <img src="..."> and <a href="..."> resolve correctly outside the browser.
  url: (id) => {
    const server = getServerUrl()
    return server ? `${server}/api/v1/attachments/${id}` : `/api/v1/attachments/${id}`
  }
}
