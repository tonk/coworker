import client from './client'

export const adminApi = {
  listUsers: () => client.get('/admin/users'),
  getUser: (id) => client.get(`/admin/users/${id}`),
  updateUser: (id, data) => client.put(`/admin/users/${id}`, data),
  deleteUser: (id) => client.delete(`/admin/users/${id}`),

  listProjects: () => client.get('/admin/projects'),
  createProject: (data) => client.post('/admin/projects', data),
  updateProject: (id, data) => client.put(`/admin/projects/${id}`, data),
  deleteProject: (id) => client.delete(`/admin/projects/${id}`),

  createUser: (data) => client.post('/admin/users', data),
  getUserProjects: (id) => client.get(`/admin/users/${id}/projects`),
  setUserProjects: (id, projectIds) => client.put(`/admin/users/${id}/projects`, { project_ids: projectIds }),
  getSystemSettings: () => client.get('/admin/system'),
  updateSystemSettings: (data) => client.put('/admin/system', data)
}
