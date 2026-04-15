import client from './client'

export const reportsApi = {
  getTimeReport: (params) => client.get('/reports/time', { params }),
  getTimeReportPDF: (params) => client.get('/reports/time/pdf', { params, responseType: 'blob' })
}
