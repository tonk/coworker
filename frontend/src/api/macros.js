import client from './client'

export const macrosApi = {
  list: () => client.get('/macros'),
  adminList: () => client.get('/admin/macros'),
  create: (data) => client.post('/admin/macros', data),
  update: (id, data) => client.put(`/admin/macros/${id}`, data),
  delete: (id) => client.delete(`/admin/macros/${id}`),
  apply: (customerId, ticketId, macroId) =>
    client.post(`/customers/${customerId}/tickets/${ticketId}/macros/${macroId}`),
  applyInbox: (ticketId, macroId) =>
    client.post(`/tickets/inbox/${ticketId}/macros/${macroId}`),
}
