import client, { fetchBinary } from './client'

export const adminApi = {
  listUsers: (deleted = false) => client.get('/admin/users', { params: deleted ? { deleted: 'true' } : {} }),
  restoreUser: (id) => client.post(`/admin/users/${id}/restore`),
  purgeUser: (id) => client.delete(`/admin/users/${id}/purge`),
  getUser: (id) => client.get(`/admin/users/${id}`),
  updateUser: (id, data) => client.put(`/admin/users/${id}`, data),
  deleteUser: (id) => client.delete(`/admin/users/${id}`),

  listProjects: (deleted = false, closed = undefined) => {
    const params = {}
    if (deleted) params.deleted = 'true'
    if (closed !== undefined) params.closed = closed
    return client.get('/admin/projects', { params })
  },
  restoreProject: (id) => client.post(`/admin/projects/${id}/restore`),
  purgeProject: (id) => client.delete(`/admin/projects/${id}/purge`),
  createProject: (data) => client.post('/admin/projects', data),
  updateProject: (id, data) => client.put(`/admin/projects/${id}`, data),
  deleteProject: (id) => client.delete(`/admin/projects/${id}`),

  createUser: (data) => client.post('/admin/users', data),
  getUserProjects: (id) => client.get(`/admin/users/${id}/projects`),
  setUserProjects: (id, projectIds) => client.put(`/admin/users/${id}/projects`, { project_ids: projectIds }),
  getUserCustomers: (id) => client.get(`/admin/users/${id}/customers`),
  setUserCustomers: (id, customerIds, customerRoles) => client.put(`/admin/users/${id}/customers`, { customer_ids: customerIds, customer_roles: customerRoles || {} }),
  getUserGroups: (id) => client.get(`/admin/users/${id}/groups`),
  setUserGroups: (id, groupIds) => client.put(`/admin/users/${id}/groups`, { group_ids: groupIds }),
  getSystemSettings: () => client.get('/admin/system'),
  updateSystemSettings: (data) => client.put('/admin/system', data),
  sendTestEmail: (to) => client.post('/admin/system/test-email', { to }),
  backupDatabase: () => client.post('/admin/system/backup'),
  uploadBackup: (file) => {
    const fd = new FormData()
    fd.append('file', file)
    return client.post('/admin/system/backups/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
  },
  listBackups: () => client.get('/admin/system/backups'),
  restoreBackup: (filename, mode = 'replace') => client.post('/admin/system/backups/restore', { filename, mode }),
  downloadBackup: (filename) => fetchBinary(`/admin/system/backups/${encodeURIComponent(filename)}`),
  deleteBackup: (filename) => client.delete(`/admin/system/backups/${encodeURIComponent(filename)}`),
  disableUserMFA: (id) => client.post(`/admin/users/${id}/mfa/disable`),
  getUserPasskeys: (id) => client.get(`/admin/users/${id}/passkeys`),
  revokeUserPasskeys: (id) => client.delete(`/admin/users/${id}/passkeys`),

  getUserLoginHistory: (id) => client.get(`/admin/users/${id}/login-history`),
  listUserApiKeys: (id) => client.get(`/admin/users/${id}/api-keys`),
  createUserApiKey: (id, name) => client.post(`/admin/users/${id}/api-keys`, { name }),
  deleteUserApiKey: (userId, keyId) => client.delete(`/admin/users/${userId}/api-keys/${keyId}`),
}
