import client, { fetchBinary } from './client'

export const timeEntriesApi = {
  list: (params) => client.get('/time-entries', { params }),
  create: (data) => client.post('/time-entries', data),
  update: (id, data) => client.put(`/time-entries/${id}`, data),
  remove: (id) => client.delete(`/time-entries/${id}`),
  report: (params) => client.get('/time-entries/report', { params }),
  reportPDF: (params) => fetchBinary('/time-entries/report/pdf', params),
  reportXLSX: (params) => fetchBinary('/time-entries/report/xlsx', params),
  sheetXLSX: (params) => fetchBinary('/time-entries/sheet/xlsx', params),
}
