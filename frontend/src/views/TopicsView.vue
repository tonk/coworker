<template>
  <div class="topics-layout">
    <div class="topics-toolbar">
      <RouterLink :to="`/projects/${slug}`" class="breadcrumb-link">
        {{ projectStore.currentProject?.name || slug }}
      </RouterLink>
      <span class="breadcrumb-sep">/</span>
      <span class="breadcrumb-cur">{{ $t('topics.title') }}</span>
    </div>

    <div class="topics-body">
      <!-- ── Topics list ──────────────────────────────────── -->
      <aside class="topics-sidebar">
        <div class="topics-sidebar-header">
          <h2>{{ $t('topics.title') }}</h2>
          <button class="btn btn-primary btn-sm" @click="showNew = true">
            + {{ $t('topics.new_topic') }}
          </button>
        </div>

        <div v-if="topicsStore.loading" class="topics-loading">
          <div class="spinner"></div>
        </div>

        <div v-else-if="!topicsStore.topics.length" class="topics-empty">
          {{ $t('topics.no_topics') }}
        </div>

        <div v-else class="topic-list">
          <div
            v-for="topic in topicsStore.topics"
            :key="topic.id"
            class="topic-item"
            :class="{ active: activeTopic?.id === topic.id, pinned: topic.is_pinned }"
            @click="openTopic(topic)"
          >
            <div class="topic-item-header">
              <span v-if="topic.is_pinned" class="pin-icon" title="Pinned">📌</span>
              <span class="topic-item-title">{{ topic.title }}</span>
            </div>
            <div class="topic-item-meta">
              <span class="topic-author">{{ topic.user?.display_name || topic.user?.username }}</span>
              <span class="topic-reply-count">{{ topic.reply_count }} {{ $t('topics.replies') }}</span>
              <span class="topic-date">{{ formatDateTime(topic.created_at) }}</span>
            </div>
          </div>
        </div>
      </aside>

      <!-- ── Topic detail ─────────────────────────────────── -->
      <main class="topics-main">
        <div v-if="!activeTopic" class="topics-placeholder">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.3"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          <p>{{ $t('topics.no_topics') }}</p>
        </div>

        <div v-else class="topic-detail">
          <!-- Topic header -->
          <div class="topic-detail-header">
            <div class="topic-detail-title-row">
              <h2 class="topic-detail-title">
                <span v-if="activeTopic.is_pinned" class="pin-icon">📌</span>
                <span v-if="editingTopic">
                  <input class="form-input topic-title-input" v-model="editForm.title" />
                </span>
                <span v-else>{{ activeTopic.title }}</span>
              </h2>
              <div class="topic-actions">
                <button v-if="canEditTopic" class="btn btn-ghost btn-sm" @click="startEditTopic">{{ $t('topics.edit') }}</button>
                <button v-if="canEditTopic" class="btn btn-ghost btn-sm" @click="togglePin">
                  {{ activeTopic.is_pinned ? $t('topics.unpin') : $t('topics.pin') }}
                </button>
                <button v-if="canEditTopic" class="btn btn-ghost btn-sm btn-danger" @click="deleteTopic">{{ $t('topics.delete') }}</button>
              </div>
            </div>
            <div class="topic-detail-meta">
              <div class="comment-avatar">
                <img v-if="avatarUrl(activeTopic.user)" :src="avatarUrl(activeTopic.user)" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else>{{ (activeTopic.user?.display_name || activeTopic.user?.username || '?').slice(0,2).toUpperCase() }}</span>
              </div>
              <strong>{{ activeTopic.user?.display_name || activeTopic.user?.username }}</strong>
              <span class="meta-sep">·</span>
              <span class="topic-date">{{ formatDateTime(activeTopic.created_at) }}</span>
              <span v-if="activeTopic.is_edited" class="edited-badge">({{ $t('topics.edited') }})</span>
            </div>
          </div>

          <!-- Topic body -->
          <div class="topic-body-section">
            <div v-if="editingTopic">
              <CardEditor v-model="editForm.body" />
              <div class="edit-actions">
                <button class="btn btn-secondary btn-sm" @click="cancelEditTopic">{{ $t('common.cancel') }}</button>
                <button class="btn btn-primary btn-sm" @click="saveTopicEdit">{{ $t('common.save') }}</button>
              </div>
            </div>
            <div v-else class="topic-body-text" v-html="renderMarkdown(activeTopic.body)"></div>
          </div>

          <!-- Replies -->
          <div class="replies-section">
            <h4 class="replies-title">{{ replies.length }} {{ $t('topics.replies') }}</h4>

            <div class="reply-list">
              <div v-for="reply in replies" :key="reply.id" class="reply-item">
                <div class="comment-avatar">
                  <img v-if="avatarUrl(reply.user)" :src="avatarUrl(reply.user)" class="avatar-img" @error="e => e.target.style.display='none'" />
                  <span v-else>{{ (reply.user?.display_name || reply.user?.username || '?').slice(0,2).toUpperCase() }}</span>
                </div>
                <div class="reply-content">
                  <div class="reply-meta">
                    <strong>{{ reply.user?.display_name || reply.user?.username }}</strong>
                    <span class="topic-date">{{ formatDateTime(reply.created_at) }}</span>
                    <span v-if="reply.is_edited" class="edited-badge">({{ $t('topics.edited') }})</span>
                    <button v-if="canEditReply(reply)" class="btn btn-ghost btn-xs" @click="startEditReply(reply)">{{ $t('topics.edit') }}</button>
                    <button v-if="canEditReply(reply)" class="btn btn-ghost btn-xs btn-danger" @click="deleteReply(reply)">{{ $t('topics.delete') }}</button>
                  </div>
                  <div v-if="editingReplyId === reply.id">
                    <CardEditor v-model="editReplyBody" />
                    <div class="edit-actions">
                      <button class="btn btn-secondary btn-sm" @click="cancelEditReply">{{ $t('common.cancel') }}</button>
                      <button class="btn btn-primary btn-sm" @click="saveReplyEdit(reply)">{{ $t('common.save') }}</button>
                    </div>
                  </div>
                  <div v-else class="reply-body" v-html="renderMarkdown(reply.body)"></div>
                </div>
              </div>
            </div>

            <!-- Add reply -->
            <div class="add-reply">
              <CardEditor v-model="newReplyBody" :placeholder="$t('topics.add_reply')" :min-height="'80px'" />
              <button class="btn btn-primary btn-sm" @click="postReply" :disabled="!newReplyBody.trim()">
                {{ $t('topics.post_reply') }}
              </button>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- New Topic modal -->
    <BaseModal v-if="showNew" :title="$t('topics.new_topic')" @close="showNew = false">
      <div class="form-group">
        <label class="form-label">{{ $t('topics.topic_title') }}</label>
        <input class="form-input" v-model="newTopic.title" required autofocus />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('topics.topic_body') }}</label>
        <CardEditor v-model="newTopic.body" />
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="showNew = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="createTopic" :disabled="!newTopic.title.trim()">{{ $t('topics.create') }}</button>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import BaseModal from '@/components/common/BaseModal.vue'
import CardEditor from '@/components/board/CardEditor.vue'
import { useTopicsStore } from '@/stores/topics'
import { useProjectStore } from '@/stores/project'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { topicsApi } from '@/api/topics'
import { useWebSocket } from '@/composables/useWebSocket'
import { useDateFormat } from '@/composables/useDateFormat'
import { avatarUrl } from '@/composables/useAvatar'

const route = useRoute()
const slug = computed(() => route.params.slug)

const topicsStore = useTopicsStore()
const projectStore = useProjectStore()
const auth = useAuthStore()
const ui = useUIStore()
const { formatDateTime } = useDateFormat()

const activeTopic = ref(null)
const replies = ref([])
const newReplyBody = ref('')
const showNew = ref(false)
const newTopic = ref({ title: '', body: '' })
const editingTopic = ref(false)
const editForm = ref({ title: '', body: '' })
const editingReplyId = ref(null)
const editReplyBody = ref('')

const { connect, disconnect } = useWebSocket(slug.value)

onMounted(async () => {
  await projectStore.loadProject(slug.value)
  await topicsStore.loadTopics(slug.value)
  connect()
})

onUnmounted(() => {
  disconnect()
  topicsStore.reset()
})

// Re-fetch active topic detail when WS updates it
watch(() => topicsStore.topics, (topics) => {
  if (activeTopic.value) {
    const updated = topics.find(t => t.id === activeTopic.value.id)
    if (updated) activeTopic.value = { ...activeTopic.value, ...updated }
  }
}, { deep: true })

const canEditTopic = computed(() => {
  if (!activeTopic.value) return false
  return activeTopic.value.user_id === auth.user?.id || auth.isAdmin
})

function canEditReply(reply) {
  return reply.user_id === auth.user?.id || auth.isAdmin
}

async function openTopic(topic) {
  activeTopic.value = topic
  replies.value = []
  newReplyBody.value = ''
  editingTopic.value = false
  editingReplyId.value = null
  try {
    const { data } = await topicsApi.get(slug.value, topic.id)
    activeTopic.value = data.topic
    replies.value = data.replies || []
  } catch {
    ui.error('Failed to load topic')
  }
}

async function createTopic() {
  try {
    await topicsApi.create(slug.value, { title: newTopic.value.title, body: newTopic.value.body })
    showNew.value = false
    newTopic.value = { title: '', body: '' }
    await topicsStore.loadTopics(slug.value)
  } catch {
    ui.error('Failed to create topic')
  }
}

function startEditTopic() {
  editForm.value = { title: activeTopic.value.title, body: activeTopic.value.body }
  editingTopic.value = true
}

function cancelEditTopic() {
  editingTopic.value = false
}

async function saveTopicEdit() {
  try {
    const { data } = await topicsApi.update(slug.value, activeTopic.value.id, editForm.value)
    activeTopic.value = data
    editingTopic.value = false
    await topicsStore.loadTopics(slug.value)
  } catch {
    ui.error('Failed to update topic')
  }
}

async function togglePin() {
  try {
    const { data } = await topicsApi.update(slug.value, activeTopic.value.id, { is_pinned: !activeTopic.value.is_pinned })
    activeTopic.value = data
    await topicsStore.loadTopics(slug.value)
  } catch {
    ui.error('Failed to update topic')
  }
}

async function deleteTopic() {
  if (!confirm('Delete this topic and all replies?')) return
  try {
    await topicsApi.delete(slug.value, activeTopic.value.id)
    activeTopic.value = null
    replies.value = []
    await topicsStore.loadTopics(slug.value)
  } catch {
    ui.error('Failed to delete topic')
  }
}

async function postReply() {
  if (!newReplyBody.value.trim()) return
  try {
    const { data } = await topicsApi.createReply(slug.value, activeTopic.value.id, newReplyBody.value)
    replies.value = [...replies.value, data]
    newReplyBody.value = ''
    // Update reply count on the topic in the list
    await topicsStore.loadTopics(slug.value)
  } catch {
    ui.error('Failed to post reply')
  }
}

function startEditReply(reply) {
  editingReplyId.value = reply.id
  editReplyBody.value = reply.body
}

function cancelEditReply() {
  editingReplyId.value = null
  editReplyBody.value = ''
}

async function saveReplyEdit(reply) {
  try {
    const { data } = await topicsApi.updateReply(slug.value, activeTopic.value.id, reply.id, editReplyBody.value)
    const idx = replies.value.findIndex(r => r.id === reply.id)
    if (idx !== -1) replies.value[idx] = data
    editingReplyId.value = null
  } catch {
    ui.error('Failed to update reply')
  }
}

async function deleteReply(reply) {
  if (!confirm('Delete this reply?')) return
  try {
    await topicsApi.deleteReply(slug.value, activeTopic.value.id, reply.id)
    replies.value = replies.value.filter(r => r.id !== reply.id)
    await topicsStore.loadTopics(slug.value)
  } catch {
    ui.error('Failed to delete reply')
  }
}

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}
</script>

<style scoped>
.topics-layout { display: flex; flex-direction: column; flex: 1; min-height: 0; overflow: hidden; }
.topics-toolbar { display: flex; align-items: center; padding: 8px 16px; border-bottom: 1px solid var(--color-border); flex-shrink: 0; }
.topics-body { display: flex; flex: 1; overflow: hidden; }

.topics-sidebar {
  width: 320px;
  flex-shrink: 0;
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.topics-sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.topics-sidebar-header h2 { margin: 0; font-size: 16px; }

.topics-loading, .topics-empty {
  display: flex; align-items: center; justify-content: center;
  padding: 32px 16px;
  color: var(--color-text-muted);
  font-size: 14px;
}

.topic-list { flex: 1; overflow-y: auto; }

.topic-item {
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid var(--color-border);
  transition: background .1s;
}
.topic-item:hover { background: var(--color-bg); }
.topic-item.active { background: color-mix(in srgb, var(--color-primary) 8%, transparent); }
.topic-item.pinned { border-left: 3px solid var(--color-primary); }

.topic-item-header { display: flex; align-items: flex-start; gap: 6px; margin-bottom: 4px; }
.topic-item-title { font-size: 13px; font-weight: 500; line-height: 1.4; }
.pin-icon { font-size: 12px; flex-shrink: 0; margin-top: 1px; }

.topic-item-meta { display: flex; align-items: center; gap: 8px; font-size: 11px; color: var(--color-text-muted); flex-wrap: wrap; }
.topic-reply-count { font-weight: 600; }

.topics-main { flex: 1; overflow-y: auto; padding: 24px; }

.topics-placeholder {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 100%; gap: 12px; color: var(--color-text-muted);
}

.topic-detail { max-width: 760px; }

.topic-detail-header { margin-bottom: 20px; }

.topic-detail-title-row {
  display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 8px;
}
.topic-detail-title { font-size: 20px; font-weight: 700; margin: 0; flex: 1; }
.topic-title-input { font-size: 18px; font-weight: 600; width: 100%; }

.topic-actions { display: flex; gap: 6px; flex-shrink: 0; }

.topic-detail-meta {
  display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--color-text-muted);
}
.meta-sep { opacity: 0.5; }
.topic-date { font-size: 12px; }
.edited-badge { font-size: 11px; font-style: italic; }

.comment-avatar {
  width: 28px; height: 28px; border-radius: 50%;
  background: var(--color-primary); color: #fff;
  font-size: 10px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; overflow: hidden;
}
.avatar-img { width: 100%; height: 100%; object-fit: cover; border-radius: 50%; }

.topic-body-section { margin-bottom: 32px; }
.topic-body-text {
  font-size: 14px; line-height: 1.6;
  padding: 16px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}
.topic-body-text :deep(p) { margin-bottom: 8px; }
.topic-body-text :deep(code) { background: #f1f5f9; padding: 1px 4px; border-radius: 3px; font-size: 13px; }

.edit-actions { display: flex; gap: 8px; margin-top: 8px; justify-content: flex-end; }

.replies-section { border-top: 1px solid var(--color-border); padding-top: 24px; }
.replies-title { margin: 0 0 16px; font-size: 14px; color: var(--color-text-muted); }

.reply-list { display: flex; flex-direction: column; gap: 16px; margin-bottom: 24px; }
.reply-item { display: flex; gap: 10px; }
.reply-content { flex: 1; }
.reply-meta { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; font-size: 12px; flex-wrap: wrap; }
.reply-body { font-size: 13px; line-height: 1.5; }
.reply-body :deep(p) { margin-bottom: 6px; }
.reply-body :deep(code) { background: #f1f5f9; padding: 1px 4px; border-radius: 3px; font-size: 12px; }

.add-reply { display: flex; flex-direction: column; gap: 8px; }
.add-reply .btn { align-self: flex-end; }

.breadcrumb-link { color: var(--color-text-muted); text-decoration: none; font-size: 14px; }
.breadcrumb-link:hover { color: var(--color-text); }
.breadcrumb-sep { color: var(--color-text-muted); margin: 0 6px; font-size: 14px; }
.breadcrumb-cur { font-size: 14px; font-weight: 500; }

.btn-xs { padding: 1px 6px; font-size: 11px; }
.btn-danger { color: var(--color-danger) !important; }
</style>
