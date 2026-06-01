import client from './client'

export const epicsApi = {
  list:     (slug)           => client.get(`/projects/${slug}/epics`),
  create:   (slug, data)     => client.post(`/projects/${slug}/epics`, data),
  update:   (slug, id, data) => client.put(`/projects/${slug}/epics/${id}`, data),
  delete:   (slug, id)       => client.delete(`/projects/${slug}/epics/${id}`),
  reorder:  (slug, items)    => client.patch(`/projects/${slug}/epics/reorder`, items),
  listCards:(slug, id)       => client.get(`/projects/${slug}/epics/${id}/cards`),
}
