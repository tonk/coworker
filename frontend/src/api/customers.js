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
}
