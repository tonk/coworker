import client from './client'

export const topicsApi = {
  list: (slug) => client.get(`/projects/${slug}/topics`),
  create: (slug, data) => client.post(`/projects/${slug}/topics`, data),
  get: (slug, topicId) => client.get(`/projects/${slug}/topics/${topicId}`),
  update: (slug, topicId, data) => client.put(`/projects/${slug}/topics/${topicId}`, data),
  delete: (slug, topicId) => client.delete(`/projects/${slug}/topics/${topicId}`),

  createReply: (slug, topicId, body) => client.post(`/projects/${slug}/topics/${topicId}/replies`, { body }),
  updateReply: (slug, topicId, replyId, body) => client.put(`/projects/${slug}/topics/${topicId}/replies/${replyId}`, { body }),
  deleteReply: (slug, topicId, replyId) => client.delete(`/projects/${slug}/topics/${topicId}/replies/${replyId}`),
}
