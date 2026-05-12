import client from './client'

export const newsApi = {
  listActive: () => client.get('/news'),

  adminList: () => client.get('/admin/news'),
  adminCreate: (data) => client.post('/admin/news', data),
  adminUpdate: (id, data) => client.put(`/admin/news/${id}`, data),
  adminDelete: (id) => client.delete(`/admin/news/${id}`),
}
