import client from './client'

export const messagesApi = {
  // Legacy 1-on-1 direct messages (kept for backward compatibility)
  listConversations: () => client.get('/direct-messages/conversations'),
  listUsers: () => client.get('/users'),
  listMessages: (userId) => client.get(`/direct-messages/${userId}`),
  sendMessage: (userId, data) => client.post(`/direct-messages/${userId}`, data),
  deleteMessage: (userId, msgId) => client.delete(`/direct-messages/${userId}/${msgId}`),

  // Conversations API (1-on-1 and group)
  getConversations: () => client.get('/conversations'),
  createConversation: (data) => client.post('/conversations', data),
  getMessages: (convId) => client.get(`/conversations/${convId}/messages`),
  sendConvMessage: (convId, data) => client.post(`/conversations/${convId}/messages`, data),
  deleteConvMessage: (convId, msgId) => client.delete(`/conversations/${convId}/messages/${msgId}`),
  addMember: (convId, data) => client.post(`/conversations/${convId}/members`, data),
  editConvMessage: (convId, msgId, body) => client.patch(`/conversations/${convId}/messages/${msgId}`, { body }),
  toggleConvReaction: (convId, msgId, emoji) => client.post(`/conversations/${convId}/messages/${msgId}/reactions`, { emoji }),
}
