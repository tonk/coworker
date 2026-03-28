<template>
  <BaseModal :title="$t('board.edit_card')" @close="$emit('close')" :resizable="true" style="--modal-width: 700px">
    <div class="card-detail">
      <div v-if="cardRef" class="card-ref-badge">{{ cardRef }}</div>
      <div class="form-group">
        <label class="form-label">{{ $t('board.card_title') }}</label>
        <input v-if="!locked" class="form-input" v-model="form.title" />
        <div v-else class="description-text">{{ form.title }}</div>
      </div>

      <div class="form-group">
        <label class="form-label">{{ $t('board.description') }}</label>
        <CardEditor v-if="!locked" v-model="form.description" />
        <div v-else class="description-text comment-text" v-html="renderMarkdown(form.description)"></div>
      </div>

      <div class="detail-row">
        <div class="form-group half">
          <label class="form-label">{{ $t('board.priority') }}</label>
          <select class="form-input" v-model="form.priority">
            <option v-for="p in priorities" :key="p" :value="p">{{ $t(`board.priorities.${p}`) }}</option>
          </select>
        </div>
        <div class="form-group half">
          <label class="form-label">{{ $t('board.due_date') }}</label>
          <input class="form-input" type="date" v-model="form.due_date" />
          <span v-if="form.due_date" class="form-hint">{{ formatDate(form.due_date) }}</span>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">{{ $t('board.assignee') }}</label>
        <select class="form-input" v-model="form.assignee_id">
          <option :value="null">—</option>
          <option v-for="m in members" :key="m.user.id" :value="m.user.id">{{ m.user.display_name || m.user.username }}</option>
        </select>
      </div>

      <div class="form-group">
        <label class="form-label">{{ $t('board.assignees') }}</label>
        <div class="labels-picker">
          <span
            v-for="m in members"
            :key="m.user.id"
            class="label-chip watcher-chip"
            :class="{ active: isAssigned(m.user.id) }"
            @click="toggleAssignee(m.user)"
          >{{ m.user.display_name || m.user.username }}</span>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">{{ $t('board.labels') }}</label>
        <div class="labels-picker">
          <span
            v-for="label in labels"
            :key="label.id"
            class="label-chip"
            :class="{ active: hasLabel(label.id) }"
            :style="{ borderColor: label.color, color: hasLabel(label.id) ? '#fff' : label.color, background: hasLabel(label.id) ? label.color : 'transparent' }"
            @click="toggleLabel(label)"
          >{{ label.name }}</span>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">{{ $t('board.tags') }}</label>
        <div class="tags-editor">
          <div class="tags-list" v-if="card.tags?.length">
            <span v-for="tag in card.tags" :key="tag.id" class="tag-chip">
              #{{ tag.name }}
              <button class="tag-remove" @click="removeTag(tag)" title="Remove tag">×</button>
            </span>
          </div>
          <div class="tag-input-row">
            <input
              class="form-input tag-input"
              v-model="newTagName"
              :placeholder="$t('board.add_tag_placeholder')"
              @keydown.enter.prevent="addTag"
              @keydown.comma.prevent="addTag"
            />
            <button class="btn btn-secondary btn-sm" @click="addTag" :disabled="!newTagName.trim()">
              {{ $t('common.add') }}
            </button>
          </div>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">{{ $t('board.watchers') }}</label>
        <div class="labels-picker">
          <span
            v-for="m in members"
            :key="m.user.id"
            class="label-chip watcher-chip"
            :class="{ active: isWatching(m.user.id) }"
            @click="toggleWatcher(m.user)"
          >{{ m.user.display_name || m.user.username }}</span>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">Attachments</label>
        <AttachmentList :attachments="attachments" :can-delete="true" @remove="deleteAttachment" />
        <div
          class="upload-drop-zone"
          :class="{ dragging: isDragging }"
          @dragover.prevent="isDragging = true"
          @dragleave="isDragging = false"
          @drop.prevent="onDrop"
          @click="fileInput.click()"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          <span>Click or drop files to attach</span>
        </div>
        <input ref="fileInput" type="file" multiple style="display:none" @change="onFileSelected" />
        <div v-if="uploading" class="upload-progress">Uploading…</div>
      </div>

      <!-- Checklist -->
      <div class="checklist-section">
        <div class="checklist-header">
          <h4>{{ $t('checklist.title') }}</h4>
          <span v-if="checklist.length" class="checklist-progress">
            {{ checklist.filter(i => i.is_completed).length }}/{{ checklist.length }}
          </span>
        </div>
        <div v-if="checklist.length" class="checklist-progress-bar">
          <div class="checklist-progress-fill" :style="{ width: checklistPct + '%' }"></div>
        </div>
        <div class="checklist-items">
          <div v-for="item in checklist" :key="item.id" class="checklist-item">
            <input
              type="checkbox"
              class="checklist-checkbox"
              :checked="item.is_completed"
              @change="toggleChecklistItem(item)"
            />
            <span v-if="editingItemId !== item.id" class="checklist-body" :class="{ completed: item.is_completed }">{{ item.body }}</span>
            <input v-else class="form-input checklist-edit-input" v-model="editItemBody" @blur="saveItemEdit(item)" @keydown.enter.prevent="saveItemEdit(item)" @keydown.esc="cancelItemEdit" />
            <button v-if="editingItemId !== item.id" class="btn-icon-xs" @click="startItemEdit(item)" title="Edit">✏</button>
            <button class="btn-icon-xs btn-danger" @click="removeChecklistItem(item)" title="Delete">×</button>
          </div>
        </div>
        <div class="checklist-add-row">
          <input
            class="form-input checklist-new-input"
            v-model="newChecklistItem"
            :placeholder="$t('checklist.add_item_placeholder')"
            @keydown.enter.prevent="addChecklistItem"
          />
          <button class="btn btn-secondary btn-sm" @click="addChecklistItem" :disabled="!newChecklistItem.trim()">
            {{ $t('checklist.add_item') }}
          </button>
        </div>
      </div>

      <div class="comments-section">
        <h4>{{ $t('board.comments') }}</h4>
        <div class="comment-list">
          <div v-for="comment in card.comments" :key="comment.id" class="comment" :class="{ 'comment-reply': comment.body.trimStart().startsWith('>') }">
            <div class="comment-avatar">
              <img v-if="avatarUrl(comment.user)" :src="avatarUrl(comment.user)" :alt="comment.user.display_name" class="comment-avatar-img" @error="e => e.target.style.display='none'" />
              <span v-else>{{ comment.user.display_name?.slice(0,2).toUpperCase() }}</span>
            </div>
            <div class="comment-body">
              <div class="comment-meta">
                <strong>{{ comment.user.display_name || comment.user.username }}</strong>
                <span class="comment-time">{{ formatDateTime(comment.created_at) }}</span>
                <span v-if="comment.is_edited" class="edited-badge">✎</span>
              </div>
              <div class="comment-text" v-html="renderMarkdown(comment.body)"></div>
              <button class="btn btn-ghost btn-sm reply-btn" @click="replyTo(comment)">{{ $t('board.reply') }}</button>
            </div>
          </div>
        </div>

        <div class="add-comment">
          <CardEditor v-model="newComment" :min-height="'80px'" :placeholder="$t('board.add_comment')" />
          <button class="btn btn-primary btn-sm" @click="submitComment" :disabled="!newComment.trim()">
            {{ $t('board.add_comment') }}
          </button>
        </div>
      </div>

      <div v-if="history.length" class="history-section">
        <h4>{{ $t('board.column_history') }}</h4>
        <div class="history-list">
          <div v-for="h in history" :key="h.id" class="history-entry">
            <span class="history-time">{{ formatDateTime(h.created_at) }}</span>
            <span class="history-who">{{ h.user.display_name || h.user.username }}</span>
            <span class="history-move">
              <span class="history-col">{{ h.from_column.name }}</span>
              →
              <span class="history-col">{{ h.to_column.name }}</span>
            </span>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <button class="btn btn-danger btn-sm" @click="confirmDelete">{{ $t('board.delete_card') }}</button>
      <button class="btn btn-secondary" @click="$emit('close')">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" @click="save" :disabled="saving">{{ saving ? $t('common.loading') : $t('common.save') }}</button>
    </template>
  </BaseModal>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import BaseModal from '@/components/common/BaseModal.vue'
import CardEditor from './CardEditor.vue'
import AttachmentList from '@/components/common/AttachmentList.vue'
import { useBoardStore } from '@/stores/board'
import { useProjectStore } from '@/stores/project'
import { projectsApi } from '@/api/projects'
import { attachmentsApi } from '@/api/attachments'
import { useUIStore } from '@/stores/ui'
import { useDateFormat } from '@/composables/useDateFormat'
import { avatarUrl } from '@/composables/useAvatar'

const props = defineProps({
  card: { type: Object, required: true },
  labels: { type: Array, default: () => [] },
  members: { type: Array, default: () => [] },
  projectSlug: { type: String, required: true }
})
const emit = defineEmits(['close', 'deleted'])

const boardStore = useBoardStore()
const projectStore = useProjectStore()
const ui = useUIStore()

const cardRef = computed(() => {
  const prefix = projectStore.currentProject?.key_prefix
  return prefix && props.card.card_number ? `${prefix}-${props.card.card_number}` : null
})
const { formatDateTime, formatDate } = useDateFormat()
const locked = ref(!!props.card.description)
const newComment = ref('')
const history = ref([])
const saving = ref(false)
const newTagName = ref('')
const attachments = ref([...(props.card.attachments || [])])
const uploading = ref(false)
const isDragging = ref(false)
const fileInput = ref(null)
const checklist = ref([])
const newChecklistItem = ref('')
const editingItemId = ref(null)
const editItemBody = ref('')
const assignees = ref([...(props.card.assignees || [])])

const checklistPct = computed(() => {
  if (!checklist.value.length) return 0
  return Math.round(checklist.value.filter(i => i.is_completed).length / checklist.value.length * 100)
})

function isAssigned(userId) {
  return assignees.value.some(a => a.id === userId)
}

async function toggleAssignee(user) {
  try {
    if (isAssigned(user.id)) {
      await projectsApi.removeAssignee(props.projectSlug, props.card.id, user.id)
      assignees.value = assignees.value.filter(a => a.id !== user.id)
    } else {
      const { data } = await projectsApi.addAssignee(props.projectSlug, props.card.id, user.id)
      assignees.value = data
    }
    boardStore.updateCard({ ...props.card, assignees: [...assignees.value] })
  } catch {
    ui.error('Failed to update assignees')
  }
}

async function addChecklistItem() {
  const body = newChecklistItem.value.trim()
  if (!body) return
  try {
    const { data } = await projectsApi.createChecklistItem(props.projectSlug, props.card.id, body)
    checklist.value = [...checklist.value, data]
    newChecklistItem.value = ''
  } catch {
    ui.error('Failed to add checklist item')
  }
}

async function toggleChecklistItem(item) {
  try {
    const { data } = await projectsApi.updateChecklistItem(props.projectSlug, props.card.id, item.id, { is_completed: !item.is_completed })
    const idx = checklist.value.findIndex(i => i.id === item.id)
    if (idx !== -1) checklist.value[idx] = data
  } catch {
    ui.error('Failed to update item')
  }
}

function startItemEdit(item) {
  editingItemId.value = item.id
  editItemBody.value = item.body
}

function cancelItemEdit() {
  editingItemId.value = null
  editItemBody.value = ''
}

async function saveItemEdit(item) {
  if (!editItemBody.value.trim()) { cancelItemEdit(); return }
  try {
    const { data } = await projectsApi.updateChecklistItem(props.projectSlug, props.card.id, item.id, { body: editItemBody.value })
    const idx = checklist.value.findIndex(i => i.id === item.id)
    if (idx !== -1) checklist.value[idx] = data
    cancelItemEdit()
  } catch {
    ui.error('Failed to update item')
  }
}

async function removeChecklistItem(item) {
  try {
    await projectsApi.deleteChecklistItem(props.projectSlug, props.card.id, item.id)
    checklist.value = checklist.value.filter(i => i.id !== item.id)
  } catch {
    ui.error('Failed to delete item')
  }
}

async function addTag() {
  const name = newTagName.value.trim().replace(/^#/, '')
  if (!name) return
  try {
    const { data } = await projectsApi.addCardTag(props.projectSlug, props.card.id, name)
    if (!props.card.tags) props.card.tags = []
    if (!props.card.tags.some(t => t.id === data.id)) {
      props.card.tags = [...props.card.tags, data]
    }
    boardStore.updateCard({ ...props.card })
    newTagName.value = ''
  } catch (e) {
    ui.error('Failed to add tag')
  }
}

async function removeTag(tag) {
  try {
    await projectsApi.removeCardTag(props.projectSlug, props.card.id, tag.id)
    props.card.tags = (props.card.tags || []).filter(t => t.id !== tag.id)
    boardStore.updateCard({ ...props.card })
  } catch (e) {
    ui.error('Failed to remove tag')
  }
}

async function uploadFiles(files) {
  if (!files.length) return
  uploading.value = true
  for (const file of files) {
    try {
      const fd = new FormData()
      fd.append('file', file)
      fd.append('owner_type', 'card')
      fd.append('owner_id', String(props.card.id))
      const { data } = await attachmentsApi.upload(fd)
      attachments.value = [...attachments.value, data]
    } catch (e) {
      ui.error(`Failed to upload ${file.name}`)
    }
  }
  uploading.value = false
}

function onFileSelected(e) {
  uploadFiles([...e.target.files])
  e.target.value = ''
}

function onDrop(e) {
  isDragging.value = false
  uploadFiles([...e.dataTransfer.files])
}

async function deleteAttachment(a) {
  try {
    await attachmentsApi.delete(a.id)
    attachments.value = attachments.value.filter(x => x.id !== a.id)
  } catch (e) {
    ui.error('Failed to delete attachment')
  }
}

onMounted(async () => {
  try {
    const [histRes, checkRes] = await Promise.all([
      projectsApi.getCardHistory(props.projectSlug, props.card.id),
      projectsApi.listChecklist(props.projectSlug, props.card.id)
    ])
    history.value = histRes.data
    checklist.value = checkRes.data || []
  } catch {}
})

const priorities = ['none', 'low', 'medium', 'high', 'critical']

const todayISO = new Date().toISOString().slice(0, 10)

const form = ref({
  title: props.card.title,
  description: props.card.description || '',
  priority: props.card.priority || 'none',
  due_date: props.card.due_date ? props.card.due_date.slice(0, 10) : todayISO,
  assignee_id: props.card.assignee_id || null
})

function hasLabel(labelId) {
  return props.card.labels?.some(l => l.id === labelId)
}

function isWatching(userId) {
  return props.card.watchers?.some(w => w.id === userId)
}

async function toggleWatcher(user) {
  try {
    if (isWatching(user.id)) {
      await projectsApi.removeWatcher(props.projectSlug, props.card.id, user.id)
      props.card.watchers = props.card.watchers.filter(w => w.id !== user.id)
    } else {
      await projectsApi.addWatcher(props.projectSlug, props.card.id, user.id)
      props.card.watchers = [...(props.card.watchers || []), user]
    }
  } catch (e) {
    ui.error('Failed to update watchers')
  }
}

async function toggleLabel(label) {
  try {
    if (hasLabel(label.id)) {
      await projectsApi.removeLabel(props.projectSlug, props.card.id, label.id)
      props.card.labels = props.card.labels.filter(l => l.id !== label.id)
    } else {
      await projectsApi.assignLabel(props.projectSlug, props.card.id, label.id)
      props.card.labels = [...(props.card.labels || []), label]
    }
    boardStore.updateCard({ ...props.card })
  } catch (e) {
    ui.error('Failed to update label')
  }
}

async function save() {
  saving.value = true
  try {
    const payload = {
      title: form.value.title,
      description: form.value.description,
      priority: form.value.priority,
      due_date: form.value.due_date || null,
      assignee_id: form.value.assignee_id
    }
    await boardStore.updateCardData(props.card.id, payload)
    locked.value = true
    if (newComment.value.trim()) await submitComment()
    ui.success('Saved')
    emit('close')
  } catch (e) {
    ui.error('Failed to save')
  } finally {
    saving.value = false
  }
}

async function submitComment() {
  if (!newComment.value.trim()) return
  try {
    const { data } = await projectsApi.createComment(props.projectSlug, props.card.id, newComment.value)
    props.card.comments = [...(props.card.comments || []), data]
    newComment.value = ''
  } catch (e) {
    ui.error('Failed to post comment')
  }
}

function replyTo(comment) {
  const author = comment.user.display_name || comment.user.username
  const quoted = comment.body.split('\n').map(l => `> ${l}`).join('\n')
  newComment.value = `> **${author}**\n${quoted}\n\n`
}

async function confirmDelete() {
  if (!confirm('Delete this card?')) return
  try {
    await boardStore.deleteCard(props.card.id, props.card.column_id)
    emit('deleted')
    emit('close')
  } catch (e) {
    ui.error('Failed to delete card')
  }
}

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}
</script>

<style scoped>
.card-ref-badge {
  display: inline-block;
  font-size: 11px;
  font-weight: 700;
  color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-primary) 25%, transparent);
  border-radius: 4px;
  padding: 2px 7px;
  margin-bottom: 12px;
  letter-spacing: 0.04em;
}
.card-detail { padding-bottom: 8px; }

.form-hint { font-size: 11px; color: var(--color-text-muted); margin-top: 4px; display: block; }

.detail-row { display: flex; gap: 16px; }
.half { flex: 1; }

.tags-editor { display: flex; flex-direction: column; gap: 8px; }
.tags-list { display: flex; flex-wrap: wrap; gap: 6px; }
.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
  background: transparent;
}
.tag-remove {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  color: var(--color-text-muted);
  padding: 0 1px;
}
.tag-remove:hover { color: var(--color-danger); }
.tag-input-row { display: flex; gap: 8px; align-items: center; }
.tag-input { flex: 1; }

.labels-picker { display: flex; flex-wrap: wrap; gap: 6px; }
.label-chip {
  padding: 3px 10px;
  border-radius: 9999px;
  font-size: 12px;
  font-weight: 600;
  border: 2px solid;
  cursor: pointer;
  transition: all .15s;
}

.upload-drop-zone {
  margin-top: 8px;
  border: 2px dashed var(--color-border);
  border-radius: var(--radius);
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: border-color .15s, background .15s;
}
.upload-drop-zone:hover, .upload-drop-zone.dragging {
  border-color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 5%, transparent);
  color: var(--color-primary);
}
.upload-progress { font-size: 12px; color: var(--color-text-muted); margin-top: 6px; }

.checklist-section { margin-top: 24px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.checklist-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.checklist-header h4 { margin: 0; font-size: 14px; }
.checklist-progress { font-size: 12px; font-weight: 600; color: var(--color-text-muted); }
.checklist-progress-bar { height: 4px; background: var(--color-border); border-radius: 2px; margin-bottom: 12px; overflow: hidden; }
.checklist-progress-fill { height: 100%; background: var(--color-primary); border-radius: 2px; transition: width .3s; }

.checklist-items { display: flex; flex-direction: column; gap: 4px; margin-bottom: 12px; }
.checklist-item { display: flex; align-items: center; gap: 8px; padding: 4px 0; }
.checklist-checkbox { width: 15px; height: 15px; cursor: pointer; flex-shrink: 0; accent-color: var(--color-primary); }
.checklist-body { flex: 1; font-size: 13px; line-height: 1.4; }
.checklist-body.completed { text-decoration: line-through; color: var(--color-text-muted); }
.checklist-edit-input { flex: 1; padding: 2px 8px; font-size: 13px; }

.checklist-add-row { display: flex; gap: 8px; align-items: center; }
.checklist-new-input { flex: 1; }

.btn-icon-xs {
  background: none; border: none; cursor: pointer; color: var(--color-text-muted);
  padding: 2px 4px; font-size: 13px; line-height: 1; border-radius: 3px; flex-shrink: 0;
}
.btn-icon-xs:hover { background: var(--color-bg); color: var(--color-text); }
.btn-icon-xs.btn-danger:hover { color: var(--color-danger); }

.comments-section { margin-top: 24px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.comments-section h4 { margin-bottom: 16px; font-size: 14px; }

.comment-list { display: flex; flex-direction: column; gap: 14px; margin-bottom: 20px; }
.comment { display: flex; gap: 10px; }
.comment-reply { margin-left: 28px; padding-left: 12px; border-left: 3px solid var(--color-border); }

.comment-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
}
.comment-avatar-img { width: 100%; height: 100%; object-fit: cover; border-radius: 50%; }

.comment-body { flex: 1; }
.comment-meta { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; font-size: 12px; }
.comment-time { color: var(--color-text-muted); }
.edited-badge { color: var(--color-text-muted); font-style: italic; font-size: 11px; }

.comment-text { font-size: 13px; line-height: 1.5; }
.comment-text :deep(p) { margin-bottom: 6px; }
.comment-text :deep(code) { background: #f1f5f9; padding: 1px 4px; border-radius: 3px; font-size: 12px; }

.watcher-chip {
  border-color: var(--color-text-muted) !important;
  color: var(--color-text-muted) !important;
  background: transparent !important;
}
.watcher-chip.active {
  border-color: var(--color-primary) !important;
  color: #fff !important;
  background: var(--color-primary) !important;
}

.description-text {
  padding: 8px 10px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  min-height: 40px;
  font-size: 13px;
  line-height: 1.5;
}

.reply-btn { margin-top: 4px; font-size: 12px; color: var(--color-text-muted); padding: 2px 8px; }
.reply-btn:hover { color: var(--color-primary); }

.add-comment { display: flex; flex-direction: column; gap: 8px; }
.add-comment .btn { align-self: flex-end; }

.history-section { margin-top: 24px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.history-section h4 { margin-bottom: 12px; font-size: 14px; color: var(--color-text-muted); }
.history-list { display: flex; flex-direction: column; gap: 6px; }
.history-entry { display: flex; align-items: center; gap: 10px; font-size: 12px; }
.history-time { color: var(--color-text-muted); flex-shrink: 0; }
.history-who { font-weight: 600; flex-shrink: 0; }
.history-move { display: flex; align-items: center; gap: 6px; color: var(--color-text-muted); }
.history-col { background: var(--color-bg); border: 1px solid var(--color-border); border-radius: var(--radius-sm); padding: 1px 6px; color: var(--color-text); font-size: 11px; }
</style>
