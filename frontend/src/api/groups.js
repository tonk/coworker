import client from './client'

export const groupsApi = {
  // Admin: group CRUD
  list: () => client.get('/admin/groups'),
  get: (id) => client.get(`/admin/groups/${id}`),
  create: (data) => client.post('/admin/groups', data),
  update: (id, data) => client.patch(`/admin/groups/${id}`, data),
  delete: (id) => client.delete(`/admin/groups/${id}`),

  // Admin: group membership
  addMember: (id, userId) => client.post(`/admin/groups/${id}/members`, { user_id: userId }),
  removeMember: (id, userId) => client.delete(`/admin/groups/${id}/members/${userId}`),

  // Admin: group access on projects / customers
  setProjectAccess: (id, projectId, role) => client.put(`/admin/groups/${id}/projects/${projectId}`, { role }),
  removeProjectAccess: (id, projectId) => client.delete(`/admin/groups/${id}/projects/${projectId}`),
  setCustomerAccess: (id, customerId, role) => client.put(`/admin/groups/${id}/customers/${customerId}`, { role }),
  removeCustomerAccess: (id, customerId) => client.delete(`/admin/groups/${id}/customers/${customerId}`),

  // Project-scoped (project owners)
  listProjectGroups: (slug) => client.get(`/projects/${slug}/groups`),
  projectAddMember: (slug, groupId, userId) => client.post(`/projects/${slug}/groups/${groupId}/members`, { user_id: userId }),
  projectRemoveMember: (slug, groupId, userId) => client.delete(`/projects/${slug}/groups/${groupId}/members/${userId}`),

  // Customer-scoped (customer owners)
  listCustomerGroups: (customerId) => client.get(`/customers/${customerId}/groups`),
  customerAddMember: (customerId, groupId, userId) => client.post(`/customers/${customerId}/groups/${groupId}/members`, { user_id: userId }),
  customerRemoveMember: (customerId, groupId, userId) => client.delete(`/customers/${customerId}/groups/${groupId}/members/${userId}`),
}
