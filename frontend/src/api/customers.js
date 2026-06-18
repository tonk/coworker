import client from './client'

export const customersApi = {
  list:           ()                         => client.get('/customers'),
  get:            (id)                       => client.get(`/customers/${id}`),
  create:         (data)                     => client.post('/customers', data),
  update:         (id, data)                 => client.put(`/customers/${id}`, data),
  delete:         (id)                       => client.delete(`/customers/${id}`),
  addFavorite:    (id)                       => client.post(`/customers/${id}/favorite`),
  removeFavorite: (id)                       => client.delete(`/customers/${id}/favorite`),
  listContracts:  (cid)                      => client.get(`/customers/${cid}/contracts`),
  createContract: (cid, data)               => client.post(`/customers/${cid}/contracts`, data),
  updateContract: (cid, rid, data)          => client.put(`/customers/${cid}/contracts/${rid}`, data),
  deleteContract: (cid, rid)                => client.delete(`/customers/${cid}/contracts/${rid}`),
  listMembers:    (cid)                      => client.get(`/customers/${cid}/members`),
  setMembers:     (cid, members)             => client.put(`/customers/${cid}/members`, { members }),
  listAllContractRates: ()                   => client.get('/customers/rates'),

  // Invoices
  listInvoices:   (cid)            => client.get(`/customers/${cid}/invoices`),
  getInvoice:     (cid, iid)       => client.get(`/customers/${cid}/invoices/${iid}`),
  createInvoice:  (cid, data)      => client.post(`/customers/${cid}/invoices`, data),
  updateInvoice:  (cid, iid, data) => client.put(`/customers/${cid}/invoices/${iid}`, data),
  deleteInvoice:  (cid, iid)       => client.delete(`/customers/${cid}/invoices/${iid}`),

  // Time-tracking-only customers (no CRM entry)
  listTimeTracking:   ()           => client.get('/time-tracking-customers'),
  createTimeTracking: (data)       => client.post('/time-tracking-customers', data),
  updateTimeTracking: (id, data)   => client.put(`/time-tracking-customers/${id}`, data),
  deleteTimeTracking: (id)         => client.delete(`/time-tracking-customers/${id}`),
}
