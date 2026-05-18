import client, { fetchBinary } from './client'

export const reportsApi = {
  getTimeReport: (params) => client.get('/reports/time', { params }),
  getTimeReportPDF: (params) => fetchBinary('/reports/time/pdf', params),
  getTimeReportXLSX: (params) => fetchBinary('/reports/time/xlsx', params),
}
