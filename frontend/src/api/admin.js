import client from './client'

export const adminApi = {
  listUsers: () => client.get('/admin/users'),
  getUser: (id) => client.get(`/admin/users/${id}`),
  updateUser: (id, data) => client.put(`/admin/users/${id}`, data),
  deleteUser: (id) => client.delete(`/admin/users/${id}`),

  listProjects: (deleted = false) => client.get('/admin/projects', { params: deleted ? { deleted: 'true' } : {} }),
  restoreProject: (id) => client.post(`/admin/projects/${id}/restore`),
  createProject: (data) => client.post('/admin/projects', data),
  updateProject: (id, data) => client.put(`/admin/projects/${id}`, data),
  deleteProject: (id) => client.delete(`/admin/projects/${id}`),

  createUser: (data) => client.post('/admin/users', data),
  getUserProjects: (id) => client.get(`/admin/users/${id}/projects`),
  setUserProjects: (id, projectIds) => client.put(`/admin/users/${id}/projects`, { project_ids: projectIds }),
  getUserCustomers: (id) => client.get(`/admin/users/${id}/customers`),
  setUserCustomers: (id, customerIds, customerRoles) => client.put(`/admin/users/${id}/customers`, { customer_ids: customerIds, customer_roles: customerRoles || {} }),
  getSystemSettings: () => client.get('/admin/system'),
  updateSystemSettings: (data) => client.put('/admin/system', data),
  sendTestEmail: (to) => client.post('/admin/system/test-email', { to }),
  backupDatabase: () => client.post('/admin/system/backup'),
  listBackups: () => client.get('/admin/system/backups'),
  restoreBackup: (filename) => client.post('/admin/system/backups/restore', { filename }),
  downloadBackup: (filename) => client.get(`/admin/system/backups/${encodeURIComponent(filename)}`, { responseType: 'blob' }),
  deleteBackup: (filename) => client.delete(`/admin/system/backups/${encodeURIComponent(filename)}`),
  disableUserMFA: (id) => client.post(`/admin/users/${id}/mfa/disable`)
}
