import client, { fetchBinary } from './client'

export const timeEntriesApi = {
  list: (params) => client.get('/time-entries', { params }),
  create: (data) => client.post('/time-entries', data),
  update: (id, data) => client.put(`/time-entries/${id}`, data),
  remove: (id) => client.delete(`/time-entries/${id}`),
  report: (params) => client.get('/time-entries/report', { params }),
  reportPDF: (params) => fetchBinary('/time-entries/report/pdf', params),
  reportChartPDF: (params) => fetchBinary('/time-entries/report/chart-pdf', params),
  reportXLSX: (params) => fetchBinary('/time-entries/report/xlsx', params),
  sheetXLSX: (params) => fetchBinary('/time-entries/sheet/xlsx', params),
  gridPDF: (params) => fetchBinary('/time-entries/grid/pdf', params),
  addHolidays: (data) => client.post('/time-entries/holidays', data),
  getRowOrder: (params) => client.get('/time-entries/row-order', { params }),
  setRowOrder: (keys, comments, params) => client.put('/time-entries/row-order', { keys, comments }, { params }),
  getMacroLibrary: () => client.get('/time-entries/macro-library'),
  setMacroLibrary: (library) => client.put('/time-entries/macro-library', { library }),
}
