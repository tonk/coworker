import client from './client'

export const ticketChecklistsApi = {
  listTemplates: () => client.get('/ticket-checklist-templates'),
  adminList: () => client.get('/admin/ticket-checklist-templates'),
  create: (data) => client.post('/admin/ticket-checklist-templates', data),
  update: (id, data) => client.put(`/admin/ticket-checklist-templates/${id}`, data),
  delete: (id) => client.delete(`/admin/ticket-checklist-templates/${id}`),
  apply: (customerId, ticketId, templateId) =>
    client.post(`/customers/${customerId}/tickets/${ticketId}/checklist/templates/${templateId}`),
  applyInbox: (ticketId, templateId) =>
    client.post(`/tickets/inbox/${ticketId}/checklist/templates/${templateId}`),
  updateItem: (customerId, ticketId, itemId, data) =>
    client.put(`/customers/${customerId}/tickets/${ticketId}/checklist/${itemId}`, data),
  updateItemInbox: (ticketId, itemId, data) =>
    client.put(`/tickets/inbox/${ticketId}/checklist/${itemId}`, data),
  deleteItem: (customerId, ticketId, itemId) =>
    client.delete(`/customers/${customerId}/tickets/${ticketId}/checklist/${itemId}`),
  deleteItemInbox: (ticketId, itemId) =>
    client.delete(`/tickets/inbox/${ticketId}/checklist/${itemId}`),
  reorder: (customerId, ticketId, items) =>
    client.patch(`/customers/${customerId}/tickets/${ticketId}/checklist/reorder`, items),
  reorderInbox: (ticketId, items) =>
    client.patch(`/tickets/inbox/${ticketId}/checklist/reorder`, items),
}
