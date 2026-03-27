import client from './client'

export const projectsApi = {
  list: () => client.get('/projects'),
  create: (data) => client.post('/projects', data),
  get: (slug) => client.get(`/projects/${slug}`),
  update: (slug, data) => client.put(`/projects/${slug}`, data),
  delete: (slug) => client.delete(`/projects/${slug}`),

  // Members
  listMembers: (slug) => client.get(`/projects/${slug}/members`),
  inviteMember: (slug, data) => client.post(`/projects/${slug}/members`, data),
  updateMemberRole: (slug, userId, role) => client.put(`/projects/${slug}/members/${userId}/role`, { role }),
  removeMember: (slug, userId) => client.delete(`/projects/${slug}/members/${userId}`),

  // Labels
  listLabels: (slug) => client.get(`/projects/${slug}/labels`),
  createLabel: (slug, data) => client.post(`/projects/${slug}/labels`, data),
  updateLabel: (slug, labelId, data) => client.put(`/projects/${slug}/labels/${labelId}`, data),
  deleteLabel: (slug, labelId) => client.delete(`/projects/${slug}/labels/${labelId}`),

  // Columns
  listColumns: (slug) => client.get(`/projects/${slug}/columns`),
  createColumn: (slug, data) => client.post(`/projects/${slug}/columns`, data),
  updateColumn: (slug, columnId, data) => client.put(`/projects/${slug}/columns/${columnId}`, data),
  deleteColumn: (slug, columnId) => client.delete(`/projects/${slug}/columns/${columnId}`),
  reorderColumns: (slug, items) => client.patch(`/projects/${slug}/columns/reorder`, items),

  // Cards
  listCards: (slug, columnId) => client.get(`/projects/${slug}/columns/${columnId}/cards`),
  createCard: (slug, columnId, data) => client.post(`/projects/${slug}/columns/${columnId}/cards`, data),
  reorderCards: (slug, columnId, items) => client.patch(`/projects/${slug}/columns/${columnId}/cards/reorder`, items),
  getCard: (slug, cardId) => client.get(`/projects/${slug}/cards/${cardId}`),
  getCardHistory: (slug, cardId) => client.get(`/projects/${slug}/cards/${cardId}/history`),
  updateCard: (slug, cardId, data) => client.put(`/projects/${slug}/cards/${cardId}`, data),
  deleteCard: (slug, cardId) => client.delete(`/projects/${slug}/cards/${cardId}`),
  moveCard: (slug, cardId, data) => client.patch(`/projects/${slug}/cards/${cardId}/move`, data),
  assignLabel: (slug, cardId, labelId) => client.post(`/projects/${slug}/cards/${cardId}/labels/${labelId}`),
  removeLabel: (slug, cardId, labelId) => client.delete(`/projects/${slug}/cards/${cardId}/labels/${labelId}`),
  addWatcher: (slug, cardId, userId) => client.post(`/projects/${slug}/cards/${cardId}/watchers/${userId}`),
  removeWatcher: (slug, cardId, userId) => client.delete(`/projects/${slug}/cards/${cardId}/watchers/${userId}`),
  updateAssignee: (slug, cardId, userId) => client.put(`/projects/${slug}/cards/${cardId}/assignee`, { user_id: userId }),

  // Comments
  listComments: (slug, cardId) => client.get(`/projects/${slug}/cards/${cardId}/comments`),
  createComment: (slug, cardId, body) => client.post(`/projects/${slug}/cards/${cardId}/comments`, { body }),
  updateComment: (slug, cardId, commentId, body) => client.put(`/projects/${slug}/cards/${cardId}/comments/${commentId}`, { body }),
  deleteComment: (slug, cardId, commentId) => client.delete(`/projects/${slug}/cards/${cardId}/comments/${commentId}`),

  // Chat
  listMessages: (slug, params) => client.get(`/projects/${slug}/chat/messages`, { params }),
  deleteMessage: (slug, msgId) => client.delete(`/projects/${slug}/chat/messages/${msgId}`),

  // Starred
  starProject: (slug) => client.post(`/projects/${slug}/star`),
  unstarProject: (slug) => client.delete(`/projects/${slug}/star`),
  listStarred: () => client.get('/starred-projects')
}
