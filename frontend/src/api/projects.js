import client from './client'

export const projectsApi = {
  list: () => client.get('/projects'),

  // Time-tracking-only projects (no board)
  listTimeTracking: () => client.get('/time-tracking-projects'),
  createTimeTracking: (data) => client.post('/time-tracking-projects', data),
  updateTimeTracking: (id, data) => client.put(`/time-tracking-projects/${id}`, data),
  deleteTimeTracking: (id) => client.delete(`/time-tracking-projects/${id}`),

  create: (data) => client.post('/projects', data),
  reorder: (items) => client.patch('/projects/reorder', items),
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
  copyCard: (slug, cardId) => client.post(`/projects/${slug}/cards/${cardId}/copy`),
  transferCard: (slug, cardId, data) => client.post(`/projects/${slug}/cards/${cardId}/transfer`, data),
  assignLabel: (slug, cardId, labelId) => client.post(`/projects/${slug}/cards/${cardId}/labels/${labelId}`),
  removeLabel: (slug, cardId, labelId) => client.delete(`/projects/${slug}/cards/${cardId}/labels/${labelId}`),
  addWatcher: (slug, cardId, userId) => client.post(`/projects/${slug}/cards/${cardId}/watchers/${userId}`),
  removeWatcher: (slug, cardId, userId) => client.delete(`/projects/${slug}/cards/${cardId}/watchers/${userId}`),
  addCardTag: (slug, cardId, name) => client.post(`/projects/${slug}/cards/${cardId}/tags`, { name }),
  removeCardTag: (slug, cardId, tagId) => client.delete(`/projects/${slug}/cards/${cardId}/tags/${tagId}`),
  updateAssignee: (slug, cardId, userId) => client.put(`/projects/${slug}/cards/${cardId}/assignee`, { user_id: userId }),

  // Checklist
  listChecklist: (slug, cardId) => client.get(`/projects/${slug}/cards/${cardId}/checklist`),
  createChecklistItem: (slug, cardId, body) => client.post(`/projects/${slug}/cards/${cardId}/checklist`, { body }),
  reorderChecklistItems: (slug, cardId, items) => client.patch(`/projects/${slug}/cards/${cardId}/checklist/reorder`, items),
  updateChecklistItem: (slug, cardId, itemId, data) => client.put(`/projects/${slug}/cards/${cardId}/checklist/${itemId}`, data),
  deleteChecklistItem: (slug, cardId, itemId) => client.delete(`/projects/${slug}/cards/${cardId}/checklist/${itemId}`),

  // Multiple assignees
  addAssignee: (slug, cardId, userId) => client.post(`/projects/${slug}/cards/${cardId}/assignees/${userId}`),
  removeAssignee: (slug, cardId, userId) => client.delete(`/projects/${slug}/cards/${cardId}/assignees/${userId}`),

  // Comments
  listComments: (slug, cardId) => client.get(`/projects/${slug}/cards/${cardId}/comments`),
  createComment: (slug, cardId, data) => client.post(`/projects/${slug}/cards/${cardId}/comments`, data),
  updateComment: (slug, cardId, commentId, data) => client.put(`/projects/${slug}/cards/${cardId}/comments/${commentId}`, data),
  deleteComment: (slug, cardId, commentId) => client.delete(`/projects/${slug}/cards/${cardId}/comments/${commentId}`),

  // Chat
  listMessages: (slug, params) => client.get(`/projects/${slug}/chat/messages`, { params }),
  deleteMessage: (slug, msgId) => client.delete(`/projects/${slug}/chat/messages/${msgId}`),

  // Card git links
  getCardLinks: (slug, cardId) => client.get(`/projects/${slug}/cards/${cardId}/links`),

  // Chat reactions
  toggleChatReaction: (slug, msgId, emoji) => client.post(`/projects/${slug}/chat/messages/${msgId}/reactions`, { emoji }),

  // Webhooks
  listWebhooks: (slug) => client.get(`/projects/${slug}/webhooks`),
  createWebhook: (slug, data) => client.post(`/projects/${slug}/webhooks`, data),
  deleteWebhook: (slug, id) => client.delete(`/projects/${slug}/webhooks/${id}`),
  regenerateWebhook: (slug, id) => client.post(`/projects/${slug}/webhooks/${id}/regenerate`),

  // Starred
  starProject: (slug) => client.post(`/projects/${slug}/star`),
  unstarProject: (slug) => client.delete(`/projects/${slug}/star`),
  listStarred: () => client.get('/starred-projects'),

  // Project-scoped API keys
  listApiKeys: (slug) => client.get(`/projects/${slug}/api-keys`),
  createApiKey: (slug, name) => client.post(`/projects/${slug}/api-keys`, { name }),
  deleteApiKey: (slug, id) => client.delete(`/projects/${slug}/api-keys/${id}`),

  // Sub-cards
  listSubCards:  (slug, cardId)       => client.get(`/projects/${slug}/cards/${cardId}/subcards`),
  createSubCard: (slug, cardId, data) => client.post(`/projects/${slug}/cards/${cardId}/subcards`, data),

  // Deleted cards (project settings)
  listDeletedCards:    (slug)       => client.get(`/projects/${slug}/cards/deleted`),
  permanentDeleteCard: (slug, id)   => client.delete(`/projects/${slug}/cards/${id}/permanent`),
  restoreCard:         (slug, id)   => client.post(`/projects/${slug}/cards/${id}/restore`),

  // Card cross-references
  listCardRefs:   (slug, cardId)        => client.get(`/projects/${slug}/cards/${cardId}/refs`),
  createCardRef:  (slug, cardId, data)  => client.post(`/projects/${slug}/cards/${cardId}/refs`, data),
  deleteCardRef:  (slug, cardId, refId) => client.delete(`/projects/${slug}/cards/${cardId}/refs/${refId}`),

  // Ticket links (card side)
  listCardTickets:(slug, cardId)        => client.get(`/projects/${slug}/cards/${cardId}/tickets`),

  // Sprints (Scrum)
  listSprints:          (slug)                    => client.get(`/projects/${slug}/sprints`),
  createSprint:         (slug, data)              => client.post(`/projects/${slug}/sprints`, data),
  updateSprint:         (slug, sprintId, data)    => client.put(`/projects/${slug}/sprints/${sprintId}`, data),
  deleteSprint:         (slug, sprintId)          => client.delete(`/projects/${slug}/sprints/${sprintId}`),
  startSprint:          (slug, sprintId)          => client.post(`/projects/${slug}/sprints/${sprintId}/start`),
  completeSprint:       (slug, sprintId)          => client.post(`/projects/${slug}/sprints/${sprintId}/complete`),
  addCardToSprint:      (slug, sprintId, cardId)  => client.post(`/projects/${slug}/sprints/${sprintId}/cards/${cardId}`),
  removeCardFromSprint: (slug, sprintId, cardId)  => client.delete(`/projects/${slug}/sprints/${sprintId}/cards/${cardId}`),
  reorderSprintCards:   (slug, sprintId, items)   => client.patch(`/projects/${slug}/sprints/${sprintId}/cards/reorder`, items),
  listBacklog:          (slug)                    => client.get(`/projects/${slug}/backlog`),
  reorderBacklog:       (slug, items)             => client.patch(`/projects/${slug}/backlog/reorder`, items),
  reorderSprints:       (slug, items)             => client.patch(`/projects/${slug}/sprints/reorder`, items),

  // Releases (Scrum)
  listReleases:            (slug)                      => client.get(`/projects/${slug}/releases`),
  createRelease:           (slug, data)                => client.post(`/projects/${slug}/releases`, data),
  updateRelease:           (slug, releaseId, data)     => client.put(`/projects/${slug}/releases/${releaseId}`, data),
  deleteRelease:           (slug, releaseId)           => client.delete(`/projects/${slug}/releases/${releaseId}`),
  addSprintToRelease:      (slug, releaseId, sprintId) => client.post(`/projects/${slug}/releases/${releaseId}/sprints/${sprintId}`),
  removeSprintFromRelease: (slug, releaseId, sprintId) => client.delete(`/projects/${slug}/releases/${releaseId}/sprints/${sprintId}`),

  // Charts (Scrum)
  getVelocityChart:        (slug)           => client.get(`/projects/${slug}/charts/velocity`),
  getBurndownChart:        (slug, sprintId) => client.get(`/projects/${slug}/charts/burndown/${sprintId}`),
  getBurnupChart:          (slug, sprintId) => client.get(`/projects/${slug}/charts/burnup/${sprintId}`),
  getCFDChart:             (slug, days)     => client.get(`/projects/${slug}/charts/cfd`, { params: { days } }),
  getCycleTimeChart:       (slug)           => client.get(`/projects/${slug}/charts/cycle-time`),
  getThroughputChart:      (slug, weeks)    => client.get(`/projects/${slug}/charts/throughput`, { params: { weeks } }),
  getReleaseBurndownChart: (slug, releaseId) => client.get(`/projects/${slug}/charts/release-burndown/${releaseId}`),
  getSprintReport:         (slug, sprintId) => client.get(`/projects/${slug}/charts/sprint-report/${sprintId}`),
  getEpicBurndown:         (slug, epicId)   => client.get(`/projects/${slug}/epics/${epicId}/burndown`),
}
