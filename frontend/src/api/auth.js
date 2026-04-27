import client from './client'

export const authApi = {
  register: (data) => client.post('/auth/register', data),
  login: (data) => client.post('/auth/login', data),
  logout: () => client.post('/auth/logout'),
  refresh: (token) => client.post('/auth/refresh', token ? { refresh_token: token } : {}),
  me: () => client.get('/auth/me'),
  updateMe: (data) => client.put('/auth/me', data),
  changePassword: (data) => client.put('/auth/me/password', data),
  listApiKeys: () => client.get('/auth/api-keys'),
  createApiKey: (name) => client.post('/auth/api-keys', { name }),
  deleteApiKey: (id) => client.delete(`/auth/api-keys/${id}`),
  verifyMFA: (mfaToken, code) => client.post('/auth/mfa/verify', { mfa_token: mfaToken, code }),
  setupMFA: () => client.get('/auth/mfa/setup'),
  enableMFA: (code) => client.post('/auth/mfa/enable', { code }),
  disableMFA: (password) => client.post('/auth/mfa/disable', { password }),
  forgotPassword: (email) => client.post('/auth/forgot-password', { email }),
  resetPassword: (token, password) => client.post('/auth/reset-password', { token, password }),
  wsTicket: () => client.post('/auth/ws-ticket'),
  mediaTicket: () => client.post('/auth/media-ticket'),
}
