import client from './client'

export const timeEntriesApi = {
  list: (params) => client.get('/time-entries', { params }),
  create: (data) => client.post('/time-entries', data),
  update: (id, data) => client.put(`/time-entries/${id}`, data),
  remove: (id) => client.delete(`/time-entries/${id}`),
  report: (params) => client.get('/time-entries/report', { params }),
  reportPDF: (params) => client.get('/time-entries/report/pdf', { params, responseType: 'blob' }),
  reportXLSX: (params) => client.get('/time-entries/report/xlsx', { params, responseType: 'blob' }),
  sheetXLSX: (params) => client.get('/time-entries/sheet/xlsx', { params, responseType: 'blob' }),
}
