<template>
  <div class="topics-layout">
    <div class="topics-toolbar">
      <img v-if="projectAvatar(projectStore.currentProject)" :src="projectAvatar(projectStore.currentProject)" class="project-avatar" alt="" />
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
          <h1>{{ $t('topics.title') }}</h1>
          <button class="btn btn-primary btn-sm" @click="showNew = true">
            + {{ $t('topics.new_topic') }}
          </button>
        </div>

        <button class="topic-item chat-nav-item" :class="{ active: chatOpen }" @click="openChat">
          <span aria-hidden="true">💬</span> {{ $t('topics.project_chat') }}
        </button>

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

      <!-- ── Topic detail / project chat ──────────────────── -->
      <main class="topics-main">
        <div v-if="chatOpen" class="chat-view">
          <h2 class="chat-view-title">{{ $t('topics.project_chat') }}</h2>

          <div class="chat-messages" ref="chatMessagesEl" role="log">
            <div v-if="chatStore.loading && !chatStore.messages.length" class="topics-loading">
              <div class="spinner"></div>
            </div>
            <div v-else-if="!chatStore.messages.length" class="topics-empty">
              {{ $t('topics.no_chat_messages') }}
            </div>
            <div v-for="msg in visibleChatMessages" :key="msg.id" class="reply-item">
              <div class="comment-avatar">
                <img v-if="avatarUrl(msg.user)" :src="avatarUrl(msg.user)" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else>{{ (msg.user?.display_name || msg.user?.username || '?').slice(0,2).toUpperCase() }}</span>
              </div>
              <div class="reply-content">
                <div class="reply-meta">
                  <strong>{{ msg.user?.display_name || msg.user?.username }}</strong>
                  <span class="topic-date">{{ formatDateTime(msg.created_at) }}</span>
                  <span v-if="msg.is_edited" class="edited-badge">({{ $t('topics.edited') }})</span>
                  <button v-if="canDeleteMessage(msg)" class="btn btn-ghost btn-xs btn-danger" @click="deleteChatMessage(msg)">{{ $t('topics.delete') }}</button>
                </div>
                <!-- nosemgrep: javascript.vue.security.audit.xss.templates.avoid-v-html.avoid-v-html -- renderMarkdown sanitizes with DOMPurify -->
                <div class="reply-body" v-html="renderMarkdown(msg.body)"></div>
                <AttachmentList v-if="msg.attachments?.length" :attachments="msg.attachments" />
              </div>
            </div>
          </div>

          <div class="add-reply">
            <AttachmentList v-if="pendingChatFiles.length" :attachments="pendingChatFiles" :can-delete="true" @remove="removePendingChatFile" />
            <div class="compose-outer">
              <InlineEmojiPicker
                v-if="chatEmojiOpen"
                :initial-search="chatEmojiQuery || ''"
                @pick="onChatEmojiPick"
                @escape="onChatEmojiEscape"
                @close="chatEmojiOpen = false"
              />
              <MentionDropdown
                v-if="chatMentionUsers.length"
                :users="chatMentionUsers"
                :active-index="chatMentionIndex"
                @pick="pickChatMention"
                @update:activeIndex="chatMentionIndex = $event"
              />
              <div class="topic-editor-wrap">
                <button class="emoji-trigger-btn" type="button" aria-label="Add emoji reaction" @click="chatEmojiOpen = !chatEmojiOpen">😊</button>
                <FileUploadButton @files-selected="onChatFilesSelected" />
                <textarea
                  ref="chatTextareaEl"
                  class="topic-textarea topic-textarea-sm"
                  v-model="newChatBody"
                  :placeholder="$t('topics.chat_placeholder')"
                  rows="3"
                  spellcheck="true"
                  :lang="auth.user?.locale || 'en'"
                  @input="onChatInput"
                  @keydown="onChatKeydownCompose"
                  @paste="onChatPaste"
                ></textarea>
              </div>
            </div>
            <button class="btn btn-primary btn-sm" @click="sendChatMessage" :disabled="!newChatBody.trim() && !pendingChatFiles.length">
              {{ $t('topics.send') }}
            </button>
          </div>
        </div>

        <div v-else-if="!activeTopic" class="topics-placeholder">
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
              <div class="compose-outer">
                <InlineEmojiPicker
                  v-if="editTopicEmojiOpen"
                  :initial-search="editTopicEmojiQuery || ''"
                  @pick="onEditTopicEmojiPick"
                  @escape="onEditTopicEmojiEscape"
                  @close="editTopicEmojiOpen = false"
                />
                <MentionDropdown
                  v-if="editTopicMentionUsers.length"
                  :users="editTopicMentionUsers"
                  :active-index="editTopicMentionIndex"
                  @pick="pickEditTopicMention"
                  @update:activeIndex="editTopicMentionIndex = $event"
                />
                <div class="topic-editor-wrap">
                  <button class="emoji-trigger-btn" type="button" aria-label="Add emoji reaction" @click="editTopicEmojiOpen = !editTopicEmojiOpen">😊</button>
                  <textarea
                    ref="editTopicTextareaEl"
                    class="topic-textarea"
                    v-model="editForm.body"
                    rows="6"
                    spellcheck="true"
                    :lang="auth.user?.locale || 'en'"
                    @input="onEditTopicInput"
                    @keydown="onEditTopicKeydownCompose"
                  ></textarea>
                </div>
              </div>
              <div class="edit-actions">
                <button class="btn btn-secondary btn-sm" @click="cancelEditTopic">{{ $t('common.cancel') }}</button>
                <button class="btn btn-primary btn-sm" @click="saveTopicEdit">{{ $t('common.save') }}</button>
              </div>
            </div>
            <!-- nosemgrep: javascript.vue.security.audit.xss.templates.avoid-v-html.avoid-v-html -- renderMarkdown sanitizes with DOMPurify -->
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
                    <div class="compose-outer">
                      <InlineEmojiPicker
                        v-if="editReplyEmojiOpen"
                        :initial-search="editReplyEmojiQuery || ''"
                        @pick="onEditReplyEmojiPick"
                        @escape="onEditReplyEmojiEscape"
                        @close="editReplyEmojiOpen = false"
                      />
                      <MentionDropdown
                        v-if="editReplyMentionUsers.length"
                        :users="editReplyMentionUsers"
                        :active-index="editReplyMentionIndex"
                        @pick="pickEditReplyMention"
                        @update:activeIndex="editReplyMentionIndex = $event"
                      />
                      <div class="topic-editor-wrap">
                        <button class="emoji-trigger-btn" type="button" aria-label="Add emoji reaction" @click="editReplyEmojiOpen = !editReplyEmojiOpen">😊</button>
                        <textarea
                          ref="editReplyTextareaEl"
                          class="topic-textarea topic-textarea-sm"
                          v-model="editReplyBody"
                          rows="4"
                          spellcheck="true"
                          :lang="auth.user?.locale || 'en'"
                          @input="onEditReplyInput"
                          @keydown="onEditReplyKeydownCompose"
                        ></textarea>
                      </div>
                    </div>
                    <div class="edit-actions">
                      <button class="btn btn-secondary btn-sm" @click="cancelEditReply">{{ $t('common.cancel') }}</button>
                      <button class="btn btn-primary btn-sm" @click="saveReplyEdit(reply)">{{ $t('common.save') }}</button>
                    </div>
                  </div>
                  <!-- nosemgrep: javascript.vue.security.audit.xss.templates.avoid-v-html.avoid-v-html -- renderMarkdown sanitizes with DOMPurify -->
                  <div v-else class="reply-body" v-html="renderMarkdown(reply.body)"></div>
                </div>
              </div>
            </div>

            <!-- Add reply -->
            <div class="add-reply">
              <div class="compose-outer">
                <InlineEmojiPicker
                  v-if="newReplyEmojiOpen"
                  :initial-search="newReplyEmojiQuery || ''"
                  @pick="onNewReplyEmojiPick"
                  @escape="onNewReplyEmojiEscape"
                  @close="newReplyEmojiOpen = false"
                />
                <MentionDropdown
                  v-if="newReplyMentionUsers.length"
                  :users="newReplyMentionUsers"
                  :active-index="newReplyMentionIndex"
                  @pick="pickNewReplyMention"
                  @update:activeIndex="newReplyMentionIndex = $event"
                />
                <div class="topic-editor-wrap">
                  <button class="emoji-trigger-btn" type="button" aria-label="Add emoji reaction" @click="newReplyEmojiOpen = !newReplyEmojiOpen">😊</button>
                  <textarea
                    ref="newReplyTextareaEl"
                    class="topic-textarea topic-textarea-sm"
                    v-model="newReplyBody"
                    :placeholder="$t('topics.add_reply')"
                    rows="4"
                    spellcheck="true"
                    :lang="auth.user?.locale || 'en'"
                    @input="onNewReplyInput"
                    @keydown="onNewReplyKeydownCompose"
                  ></textarea>
                </div>
              </div>
              <button class="btn btn-primary btn-sm" @click="postReply" :disabled="!newReplyBody.trim()">
                {{ $t('topics.post_reply') }}
              </button>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- New Topic modal -->
    <BaseModal v-if="showNew" :title="$t('topics.new_topic')" @close="showNew = false" :resizable="true">
      <div class="form-group">
        <label class="form-label" for="new-topic-title">{{ $t('topics.topic_title') }}</label>
        <input id="new-topic-title" class="form-input" v-model="newTopic.title" required autofocus />
      </div>
      <div class="form-group">
        <label class="form-label" for="new-topic-body">{{ $t('topics.topic_body') }}</label>
        <div class="compose-outer">
          <InlineEmojiPicker
            v-if="newTopicEmojiOpen"
            :initial-search="newTopicEmojiQuery || ''"
            @pick="onNewTopicEmojiPick"
            @escape="onNewTopicEmojiEscape"
            @close="newTopicEmojiOpen = false"
          />
          <MentionDropdown
            v-if="newTopicMentionUsers.length"
            :users="newTopicMentionUsers"
            :active-index="newTopicMentionIndex"
            @pick="pickNewTopicMention"
            @update:activeIndex="newTopicMentionIndex = $event"
          />
          <div class="topic-editor-wrap">
            <button class="emoji-trigger-btn" type="button" aria-label="Add emoji reaction" @click="newTopicEmojiOpen = !newTopicEmojiOpen">😊</button>
            <textarea
              id="new-topic-body"
              ref="newTopicTextareaEl"
              class="topic-textarea"
              v-model="newTopic.body"
              rows="6"
              spellcheck="true"
              :lang="auth.user?.locale || 'en'"
              @input="onNewTopicInput"
              @keydown="onNewTopicKeydownCompose"
            ></textarea>
          </div>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="showNew = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="createTopic" :disabled="!newTopic.title.trim()">{{ $t('topics.create') }}</button>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import BaseModal from '@/components/common/BaseModal.vue'
import MentionDropdown from '@/components/common/MentionDropdown.vue'
import InlineEmojiPicker from '@/components/common/InlineEmojiPicker.vue'
import AttachmentList from '@/components/common/AttachmentList.vue'
import FileUploadButton from '@/components/common/FileUploadButton.vue'
import { useTopicsStore } from '@/stores/topics'
import { useProjectStore } from '@/stores/project'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { topicsApi } from '@/api/topics'
import { projectsApi } from '@/api/projects'
import { messagesApi } from '@/api/messages'
import { attachmentsApi } from '@/api/attachments'
import { useWebSocket } from '@/composables/useWebSocket'
import { useDateFormat } from '@/composables/useDateFormat'
import { avatarUrl } from '@/composables/useAvatar'
import { useCompose } from '@/composables/useCompose'
import { resolveAssetUrl } from '@/api/serverConfig'

const route = useRoute()
const slug = computed(() => route.params.slug)

const topicsStore = useTopicsStore()
const projectStore = useProjectStore()
const chatStore = useChatStore()
const auth = useAuthStore()
const ui = useUIStore()
const { formatDateTime } = useDateFormat()

const activeTopic = ref(null)
const chatOpen = ref(false)
const chatMessagesEl = ref(null)
const chatTextareaEl = ref(null)
const newChatBody = ref('')
const pendingChatFiles = ref([])
const visibleChatMessages = computed(() => chatStore.messages.filter(m => !m.is_deleted))

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}
const replies = ref([])
const projectMembers = ref([])
const newReplyBody = ref('')
const showNew = ref(false)
const newTopic = ref({ title: '', body: '' })
const editingTopic = ref(false)
const editForm = ref({ title: '', body: '' })
const editingReplyId = ref(null)
const editReplyBody = ref('')
const projectMentionUsers = computed(() =>
  (projectMembers.value || [])
    .map(m => m?.user || m)
    .filter(u => u && u.username)
)

const editTopicTextareaEl = ref(null)
const editReplyTextareaEl = ref(null)
const newReplyTextareaEl = ref(null)
const newTopicTextareaEl = ref(null)

const {
  mentionUsers: editTopicMentionUsers,
  mentionIndex: editTopicMentionIndex,
  onTextareaInput: onEditTopicInput,
  onTextareaKeydown: onEditTopicKeydownCompose,
  pickMention: pickEditTopicMention,
  emojiQuery: editTopicEmojiQuery,
  pickEmoji: pickEditTopicEmoji,
} = useCompose({
  textareaEl: editTopicTextareaEl,
  getValue: () => editForm.value.body,
  setValue: (v) => { editForm.value.body = v },
  users: projectMentionUsers,
})
const editTopicEmojiOpen = ref(false)
watch(editTopicEmojiQuery, (q) => { editTopicEmojiOpen.value = q !== null })

const {
  mentionUsers: editReplyMentionUsers,
  mentionIndex: editReplyMentionIndex,
  onTextareaInput: onEditReplyInput,
  onTextareaKeydown: onEditReplyKeydownCompose,
  pickMention: pickEditReplyMention,
  emojiQuery: editReplyEmojiQuery,
  pickEmoji: pickEditReplyEmoji,
} = useCompose({
  textareaEl: editReplyTextareaEl,
  getValue: () => editReplyBody.value,
  setValue: (v) => { editReplyBody.value = v },
  users: projectMentionUsers,
})
const editReplyEmojiOpen = ref(false)
watch(editReplyEmojiQuery, (q) => { editReplyEmojiOpen.value = q !== null })

const {
  mentionUsers: newReplyMentionUsers,
  mentionIndex: newReplyMentionIndex,
  onTextareaInput: onNewReplyInput,
  onTextareaKeydown: onNewReplyKeydownCompose,
  pickMention: pickNewReplyMention,
  emojiQuery: newReplyEmojiQuery,
  pickEmoji: pickNewReplyEmoji,
} = useCompose({
  textareaEl: newReplyTextareaEl,
  getValue: () => newReplyBody.value,
  setValue: (v) => { newReplyBody.value = v },
  users: projectMentionUsers,
})
const newReplyEmojiOpen = ref(false)
watch(newReplyEmojiQuery, (q) => { newReplyEmojiOpen.value = q !== null })

const {
  mentionUsers: newTopicMentionUsers,
  mentionIndex: newTopicMentionIndex,
  onTextareaInput: onNewTopicInput,
  onTextareaKeydown: onNewTopicKeydownCompose,
  pickMention: pickNewTopicMention,
  emojiQuery: newTopicEmojiQuery,
  pickEmoji: pickNewTopicEmoji,
} = useCompose({
  textareaEl: newTopicTextareaEl,
  getValue: () => newTopic.value.body,
  setValue: (v) => { newTopic.value.body = v },
  users: projectMentionUsers,
})
const newTopicEmojiOpen = ref(false)
watch(newTopicEmojiQuery, (q) => { newTopicEmojiOpen.value = q !== null })

const {
  mentionUsers: chatMentionUsers,
  mentionIndex: chatMentionIndex,
  onTextareaInput: onChatInput,
  onTextareaKeydown: onChatKeydownCompose,
  pickMention: pickChatMention,
  emojiQuery: chatEmojiQuery,
  pickEmoji: pickChatEmoji,
} = useCompose({
  textareaEl: chatTextareaEl,
  getValue: () => newChatBody.value,
  setValue: (v) => { newChatBody.value = v },
  users: projectMentionUsers,
})
const chatEmojiOpen = ref(false)
watch(chatEmojiQuery, (q) => { chatEmojiOpen.value = q !== null })

const { connect, disconnect } = useWebSocket(slug.value)

onMounted(async () => {
  await Promise.all([
    projectStore.fetchProject(slug.value),
    topicsStore.loadTopics(slug.value),
    loadProjectMembers(),
  ])
  connect()
})

onUnmounted(() => {
  disconnect()
  topicsStore.reset()
  chatStore.reset()
})

// Re-fetch active topic detail when WS updates it
watch(() => topicsStore.topics, (topics) => {
  if (activeTopic.value) {
    const updated = topics.find(t => t.id === activeTopic.value.id)
    if (updated) activeTopic.value = { ...activeTopic.value, ...updated }
  }
}, { deep: true })

// Scroll to the newest chat message as more arrive live via WebSocket
watch(() => chatStore.messages.length, () => {
  if (chatOpen.value) nextTick(() => scrollChatToBottom())
})

async function loadProjectMembers() {
  try {
    const { data } = await projectsApi.listMembers(slug.value)
    const members = data || []
    if (members.length) {
      projectMembers.value = members
      return
    }
    // Fallback for datasets where project membership rows are empty/missing.
    const usersResp = await messagesApi.listUsers()
    projectMembers.value = usersResp.data || []
  } catch {
    try {
      const usersResp = await messagesApi.listUsers()
      projectMembers.value = usersResp.data || []
    } catch {
      projectMembers.value = []
    }
  }
}

const canEditTopic = computed(() => {
  if (!activeTopic.value) return false
  return activeTopic.value.user_id === auth.user?.id || auth.isAdmin
})

function canEditReply(reply) {
  return reply.user_id === auth.user?.id || auth.isAdmin
}

async function openTopic(topic) {
  chatOpen.value = false
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

async function openChat() {
  activeTopic.value = null
  chatOpen.value = true
  if (!chatStore.messages.length) {
    await chatStore.loadMessages(slug.value)
  }
  await nextTick()
  scrollChatToBottom()
}

function scrollChatToBottom() {
  if (chatMessagesEl.value) chatMessagesEl.value.scrollTop = chatMessagesEl.value.scrollHeight
}

function canDeleteMessage(msg) {
  return msg.user_id === auth.user?.id || auth.isAdmin
}

async function deleteChatMessage(msg) {
  if (!await ui.confirm('Delete this message?', { destructive: true })) return
  try {
    await projectsApi.deleteMessage(slug.value, msg.id)
    chatStore.removeMessage({ id: msg.id })
  } catch {
    ui.error('Failed to delete message')
  }
}

function onChatFilesSelected(files) {
  for (const f of files) {
    pendingChatFiles.value.push({
      id: Math.random(),
      filename: f.name,
      size_bytes: f.size,
      mime_type: f.type || 'application/octet-stream',
      _file: f,
      _previewUrl: f.type?.startsWith('image/') ? URL.createObjectURL(f) : null,
    })
  }
}

function onChatPaste(e) {
  const items = Array.from(e.clipboardData?.items || [])
  const images = items.filter(it => it.kind === 'file' && it.type.startsWith('image/'))
  if (images.length) {
    e.preventDefault()
    onChatFilesSelected(images.map(it => it.getAsFile()).filter(Boolean))
  }
}

function removePendingChatFile(a) {
  if (a._previewUrl) URL.revokeObjectURL(a._previewUrl)
  pendingChatFiles.value = pendingChatFiles.value.filter(p => p.id !== a.id)
}

async function sendChatMessage() {
  if (!newChatBody.value.trim() && !pendingChatFiles.value.length) return
  try {
    const { data } = await projectsApi.sendMessage(slug.value, { body: newChatBody.value })
    data.attachments = []

    if (pendingChatFiles.value.length) {
      const filesToUpload = [...pendingChatFiles.value]
      pendingChatFiles.value = []
      filesToUpload.forEach(pf => { if (pf._previewUrl) URL.revokeObjectURL(pf._previewUrl) })
      for (const pf of filesToUpload) {
        const fd = new FormData()
        fd.append('file', pf._file)
        fd.append('owner_type', 'chat_message')
        fd.append('owner_id', String(data.id))
        try {
          const { data: att } = await attachmentsApi.upload(fd)
          data.attachments.push(att)
        } catch {
          ui.error(`Failed to upload ${pf.filename}`)
        }
      }
    }

    chatStore.addMessage(data)
    newChatBody.value = ''
    await nextTick()
    scrollChatToBottom()
  } catch {
    ui.error('Failed to send message')
  }
}

function onChatEmojiPick(emoji) {
  pickChatEmoji(emoji)
  chatEmojiOpen.value = false
}
function onChatEmojiEscape() {
  chatEmojiOpen.value = false
  nextTick(() => chatTextareaEl.value?.focus())
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
  if (!await ui.confirm('Delete this topic and all replies?', { destructive: true })) return
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
  if (!await ui.confirm('Delete this reply?', { destructive: true })) return
  try {
    await topicsApi.deleteReply(slug.value, activeTopic.value.id, reply.id)
    replies.value = replies.value.filter(r => r.id !== reply.id)
    await topicsStore.loadTopics(slug.value)
  } catch {
    ui.error('Failed to delete reply')
  }
}

function onEditTopicEmojiPick(emoji) {
  pickEditTopicEmoji(emoji)
  editTopicEmojiOpen.value = false
}
function onEditReplyEmojiPick(emoji) {
  pickEditReplyEmoji(emoji)
  editReplyEmojiOpen.value = false
}
function onNewReplyEmojiPick(emoji) {
  pickNewReplyEmoji(emoji)
  newReplyEmojiOpen.value = false
}
function onNewTopicEmojiPick(emoji) {
  pickNewTopicEmoji(emoji)
  newTopicEmojiOpen.value = false
}

function onEditTopicEmojiEscape() {
  editTopicEmojiOpen.value = false
  nextTick(() => editTopicTextareaEl.value?.focus())
}
function onEditReplyEmojiEscape() {
  editReplyEmojiOpen.value = false
  nextTick(() => editReplyTextareaEl.value?.focus())
}
function onNewReplyEmojiEscape() {
  newReplyEmojiOpen.value = false
  nextTick(() => newReplyTextareaEl.value?.focus())
}
function onNewTopicEmojiEscape() {
  newTopicEmojiOpen.value = false
  nextTick(() => newTopicTextareaEl.value?.focus())
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
.topics-sidebar-header h1 { margin: 0; font-size: 16px; }

.chat-nav-item {
  width: 100%;
  text-align: left;
  padding: 12px 16px;
  border: none;
  border-bottom: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}
.chat-nav-item:hover { background: var(--color-bg); }
.chat-nav-item.active { background: color-mix(in srgb, var(--color-primary) 8%, transparent); }

.chat-view { display: flex; flex-direction: column; height: 100%; max-width: 760px; }
.chat-view-title { margin: 0 0 16px; font-size: 20px; font-weight: 700; }
.chat-messages { flex: 1; overflow-y: auto; display: flex; flex-direction: column; gap: 16px; padding-right: 4px; margin-bottom: 16px; }

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

.compose-outer { position: relative; }
.topic-editor-wrap {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 8px;
}
.topic-editor-wrap:focus-within { border-color: var(--color-primary); }
.topic-textarea {
  width: 100%;
  border: none;
  outline: none;
  resize: vertical;
  min-height: 150px;
  background: transparent;
  color: var(--color-text);
  font-family: inherit;
  font-size: 14px;
  line-height: 1.5;
}
.topic-textarea-sm { min-height: 80px; }
.emoji-trigger-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  padding: 2px 3px;
  border-radius: 5px;
  line-height: 1;
  flex-shrink: 0;
  opacity: .55;
  transition: opacity .1s;
}
.emoji-trigger-btn:hover { opacity: 1; }
.emoji-trigger-btn:focus-visible { opacity: 1; }

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
.project-avatar {
  width: 20px;
  height: 20px;
  border-radius: 5px;
  object-fit: cover;
  border: 1px solid var(--color-border);
}
.breadcrumb-sep { color: var(--color-text-muted); margin: 0 6px; font-size: 14px; }
.breadcrumb-cur { font-size: 14px; font-weight: 500; }

.btn-xs { padding: 1px 6px; font-size: 11px; }
.btn-danger { color: var(--color-danger) !important; }
</style>
