import client from './client'

export const searchApi = {
  search: (q) => client.get('/search', { params: { q } })
}
