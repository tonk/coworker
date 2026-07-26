import client from './client'

export const searchApi = {
  search: (q) => client.get('/search', { params: { q } }),
  preview: (q, replace, types, limit) => client.post('/search/replace/preview', { q, replace, types, limit }),
  replace: (q, replace, items) => client.post('/search/replace/apply', { q, replace, items })
}
