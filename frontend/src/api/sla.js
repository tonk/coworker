import client from './client'

export const slaApi = {
  list: () => client.get('/admin/sla-policies'),
  create: (data) => client.post('/admin/sla-policies', data),
  update: (id, data) => client.put(`/admin/sla-policies/${id}`, data),
  delete: (id) => client.delete(`/admin/sla-policies/${id}`),
}
