<template>
  <main class="ticket-list-main">
    <div class="ticket-list-container">
      <DashboardNews />
      <header class="ticket-list-header">
        <RouterLink :to="`/customers/${customerId}`" class="back-link">{{ $t('common.go_back') }}</RouterLink>
        <h1>{{ customer?.name || $t('ticket.tickets') }}</h1>
        <button class="btn btn-primary btn-sm" @click="showCreate = true">+ {{ $t('ticket.new_ticket') }}</button>
      </header>

      <div v-if="loading" class="loading-state">
        <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
      </div>

      <template v-else>
        <div v-if="!tickets.length" class="empty-state">
          {{ $t('ticket.no_tickets') }}
        </div>

        <div v-else class="ticket-grid">
          <div
            v-for="t in tickets"
            :key="t.id"
            class="ticket-card"
            :class="'ticket-' + t.status"
            @click="openTicket(t)"
            role="button"
            tabindex="0"
            @keydown.enter="openTicket(t)"
            @keydown.space.prevent="openTicket(t)"
            :aria-label="t.title"
          >
            <div class="ticket-card-header">
              <span class="ticket-id">#{{ t.id }}</span>
              <span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span>
              <span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span>
            </div>
            <h3 class="ticket-card-title">{{ t.title }}</h3>
            <div class="ticket-card-meta">
              <span class="ticket-status" :class="'status-' + t.status">{{ $t('ticket.status_' + t.status) }}</span>
              <span v-if="t.sla_response_breached" class="sla-badge sla-breach" :title="slaTitle(t)">{{ $t('sla.breached') }}</span>
              <span v-else-if="slaWarning(t)" class="sla-badge sla-warning" :title="slaTitle(t)">{{ $t('sla.warning') }}</span>
              <span v-else-if="t.sla_policy_id" class="sla-badge sla-ok" :title="slaTitle(t)">{{ $t('sla.on_track') }}</span>
              <span v-if="t.tags?.length" class="ticket-tags">
                <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                <span v-if="t.tags.length > 3" class="mini-tag more">+{{ t.tags.length - 3 }}</span>
              </span>
              <span v-if="t.assigned_to" class="ticket-assignee">
                {{ t.assigned_to.display_name || t.assigned_to.username }}
              </span>
              <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
            </div>
          </div>
        </div>
      </template>
    </div>

    <BaseModal v-if="showCreate" :title="$t('ticket.new_ticket')" @close="showCreate = false">
      <form @submit.prevent="submitCreate">
        <div class="form-group">
          <label>{{ $t('ticket.title') }}</label>
          <input v-model="newTicket.title" class="form-input" required :placeholder="$t('ticket.title_placeholder')" />
        </div>
        <div class="form-group">
          <label>{{ $t('ticket.description') }}</label>
          <div class="md-editor">
            <div class="md-editor-tabs" role="tablist">
              <button role="tab" :aria-selected="descTab === 'edit'" aria-controls="ticket-create-panel-edit" id="ticket-create-tab-edit" :class="['md-tab', { active: descTab === 'edit' }]" type="button" @click="descTab = 'edit'">{{ $t('common.edit') }}</button>
              <button role="tab" :aria-selected="descTab === 'preview'" aria-controls="ticket-create-panel-preview" id="ticket-create-tab-preview" :class="['md-tab', { active: descTab === 'preview' }]" type="button" @click="descTab = 'preview'">{{ $t('common.preview') }}</button>
            </div>
            <textarea v-if="descTab === 'edit'" id="ticket-create-panel-edit" role="tabpanel" aria-labelledby="ticket-create-tab-edit" v-model="newTicket.description" class="form-input md-textarea" rows="4" @paste="onDescPaste"></textarea>
            <div v-else id="ticket-create-panel-preview" role="tabpanel" aria-labelledby="ticket-create-tab-preview" class="md-preview markdown-body" v-html="renderMarkdown(newTicket.description)"></div>
          </div>
          <div v-if="pendingFiles.length" class="pending-attachments">
            <span class="pending-label">{{ $t('ticket.pending_files') }}</span>
            <AttachmentList :attachments="pendingFiles" :can-delete="true" @remove="removePending" />
          </div>
          <FileUploadButton @files-selected="onFilesSelected" />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>{{ $t('ticket.type') }}</label>
            <select v-model="newTicket.type" class="form-input">
              <option value="incident">{{ $t('ticket.type_incident') }}</option>
              <option value="problem">{{ $t('ticket.type_problem') }}</option>
              <option value="service_request">{{ $t('ticket.type_service_request') }}</option>
              <option value="change_request">{{ $t('ticket.type_change_request') }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>{{ $t('ticket.priority') }}</label>
            <select v-model="newTicket.priority" class="form-input">
              <option value="low">{{ $t('ticket.priority_low') }}</option>
              <option value="medium" selected>{{ $t('ticket.priority_medium') }}</option>
              <option value="high">{{ $t('ticket.priority_high') }}</option>
              <option value="critical">{{ $t('ticket.priority_critical') }}</option>
            </select>
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>{{ $t('ticket.group') }}</label>
            <select v-model="newTicket.group_id" class="form-input">
              <option :value="null">—</option>
              <option v-for="g in customerGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>{{ $t('ticket.owner') }}</label>
            <select v-model="newTicket.owner_id" class="form-input">
              <option :value="null">—</option>
              <option v-for="u in customerUsers" :key="u.id" :value="u.id">{{ u.display_name || u.username }}</option>
            </select>
          </div>
        </div>
        <div class="form-group">
          <label>{{ $t('ticket.tags') }}</label>
          <div class="tags-editor">
            <div class="tags-list" v-if="newTicketTags.length">
              <span v-for="(tag, i) in newTicketTags" :key="i" class="tag-chip">
                #{{ tag }}
                <button class="tag-remove" @click="newTicketTags.splice(i, 1)" title="Remove tag" aria-label="Remove tag">×</button>
              </span>
            </div>
            <input class="form-input tag-input" v-model="newTagName" :placeholder="$t('ticket.add_tag_placeholder')" @keydown.enter.prevent="addNewTag" @keydown.comma.prevent="addNewTag" />
          </div>
        </div>
        <div class="modal-footer" slot="footer">
          <button type="button" class="btn btn-secondary" @click="showCreate = false">{{ $t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary">{{ $t('common.create') }}</button>
        </div>
      </form>
    </BaseModal>
  </main>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import client from '@/api/client'
import { customersApi } from '@/api/customers'
import { ticketsApi } from '@/api/tickets'
import { attachmentsApi } from '@/api/attachments'
import { useUIStore } from '@/stores/ui'
import BaseModal from '@/components/common/BaseModal.vue'
import DashboardNews from '@/components/common/DashboardNews.vue'
import AttachmentList from '@/components/common/AttachmentList.vue'
import FileUploadButton from '@/components/common/FileUploadButton.vue'
import { useDateFormat } from '@/composables/useDateFormat'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const { formatDate } = useDateFormat()

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}

const route = useRoute()
const router = useRouter()
const ui = useUIStore()

const customerId = computed(() => Number(route.params.id))
const customer = ref(null)
const tickets = ref([])
const loading = ref(true)
const showCreate = ref(false)
const descTab = ref('edit')
const newTicket = ref({ title: '', description: '', type: 'incident', priority: 'medium', group_id: null, owner_id: null })
const newTicketTags = ref([])
const newTagName = ref('')
const pendingFiles = ref([])
const customerGroups = ref([])
const customerUsers = ref([])

async function fetchData() {
  loading.value = true
  pendingFiles.value = []
  try {
    const { data } = await customersApi.get(customerId.value)
    customer.value = data.customer || data
  } catch {}
  try {
    const { data } = await ticketsApi.list(customerId.value)
    tickets.value = data || []
  } catch {}
  try {
    const { data } = await client.get(`/customers/${customerId.value}/groups`)
    customerGroups.value = data || []
  } catch {}
  try {
    const { data } = await client.get(`/customers/${customerId.value}/members`)
    customerUsers.value = data || []
  } catch {}
  loading.value = false
}

watch(() => route.params.id, () => {
  fetchData()
})

onMounted(fetchData)

function openTicket(t) {
  router.push(`/customers/${customerId.value}/tickets/${t.id}`)
}

function slaWarning(t) {
  if (!t.sla_response_deadline && !t.sla_resolution_deadline) return false
  if (t.status === 'resolved' || t.status === 'closed') return false
  const now = Date.now()
  if (t.sla_response_deadline && !t.first_response_at) {
    const deadline = new Date(t.sla_response_deadline).getTime()
    if (deadline - now < 3600000 && deadline - now > 0) return true
  }
  if (t.sla_resolution_deadline) {
    const deadline = new Date(t.sla_resolution_deadline).getTime()
    if (deadline - now < 3600000 && deadline - now > 0) return true
  }
  return false
}

function slaTitle(t) {
  const parts = []
  if (t.sla_policy?.name) parts.push(t.sla_policy.name)
  if (t.sla_response_deadline) {
    parts.push(`Response: ${formatDate(t.sla_response_deadline)}`)
  }
  if (t.sla_resolution_deadline) {
    parts.push(`Resolution: ${formatDate(t.sla_resolution_deadline)}`)
  }
  return parts.join(' | ')
}

async function submitCreate() {
  try {
    const res = await ticketsApi.create(customerId.value, newTicket.value)
    const ticket = res.data || res

    if (pendingFiles.value.length) {
      const filesToUpload = [...pendingFiles.value]
      pendingFiles.value = []
      filesToUpload.forEach(pf => { if (pf._previewUrl) URL.revokeObjectURL(pf._previewUrl) })
      for (const pf of filesToUpload) {
        const fd = new FormData()
        fd.append('file', pf._file)
        fd.append('owner_type', 'ticket')
        fd.append('owner_id', String(ticket.id))
        try {
          await attachmentsApi.upload(fd)
        } catch {}
      }
    }

    // Add tags after creation
    if (newTicketTags.value.length) {
      for (const tagName of newTicketTags.value) {
        try {
          await ticketsApi.addTag(customerId.value, ticket.id, tagName)
        } catch {}
      }
    }

    tickets.value.unshift(ticket)
    showCreate.value = false
    descTab.value = 'edit'
    newTicket.value = { title: '', description: '', type: 'incident', priority: 'medium', group_id: null, owner_id: null }
    newTicketTags.value = []
    newTagName.value = ''
    ui.success('Ticket created')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to create ticket')
  }
}

function onFilesSelected(files) {
  for (const f of files) {
    pendingFiles.value.push({
      id: Math.random(),
      filename: f.name,
      size_bytes: f.size,
      mime_type: f.type || 'application/octet-stream',
      _file: f,
      _previewUrl: f.type?.startsWith('image/') ? URL.createObjectURL(f) : null,
    })
  }
}

function removePending(a) {
  if (a._previewUrl) URL.revokeObjectURL(a._previewUrl)
  pendingFiles.value = pendingFiles.value.filter(p => p.id !== a.id)
}

function addNewTag() {
  const name = newTagName.value.trim().replace(/^#/, '')
  if (!name || newTicketTags.value.includes(name)) return
  newTicketTags.value.push(name)
  newTagName.value = ''
}

async function onDescPaste(e) {
  const items = Array.from(e.clipboardData?.items || [])
  const images = items.filter(it => it.kind === 'file' && it.type.startsWith('image/'))
  if (images.length) {
    e.preventDefault()
    const files = await Promise.all(images.map(it => {
      const file = it.getAsFile()
      if (file) return file
      return it.getType(it.type).then(blob => {
        const ext = it.type.split('/')[1]?.split('+')[0] || 'png'
        return new File([blob], `clipboard.${ext}`, { type: it.type })
      })
    }))
    const valid = files.filter(Boolean)
    if (valid.length) {
      onFilesSelected(valid)
      ui.success(valid.length > 1 ? `${valid.length} images pasted` : 'Image pasted')
    }
    return
  }
  if (window.__TAURI_INTERNALS__ && navigator.clipboard?.read) {
    try {
      const clipItems = await navigator.clipboard.read()
      const files = []
      for (const item of clipItems) {
        for (const type of item.types) {
          if (type.startsWith('image/')) {
            const blob = await item.getType(type)
            const ext = type.split('/')[1]?.split('+')[0] || 'png'
            files.push(new File([blob], `paste.${ext}`, { type }))
          }
        }
      }
      if (files.length) { e.preventDefault(); onFilesSelected(files); ui.success('Image pasted') }
    } catch {}
  }
}
</script>

<style scoped>
.ticket-list-main { padding: 24px; max-width: 1200px; margin: 0 auto; }
.ticket-list-header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.ticket-list-header h1 { flex: 1; margin: 0; font-size: 20px; }
.back-link { font-size: 13px; color: var(--color-primary); text-decoration: none; }
.back-link:hover { text-decoration: underline; }
.loading-state { display: flex; justify-content: center; padding: 48px; }
.empty-state { text-align: center; padding: 64px 24px; color: var(--color-text-muted); }
.ticket-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 12px; }
.ticket-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: box-shadow .15s, border-color .15s;
}
.ticket-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,.08); border-color: var(--color-primary); }
.ticket-card:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }
.ticket-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.ticket-id { font-size: 11px; font-weight: 700; color: var(--color-text-muted); }
.ticket-type { font-size: 10px; font-weight: 600; padding: 2px 6px; border-radius: 4px; }
.type-incident { background: #fecaca; color: #b91c1c; }
.type-problem { background: #fed7aa; color: #9a3412; }
.type-service_request { background: #d1fae5; color: #065f46; }
.type-change_request { background: #dbeafe; color: #1e40af; }
.mini-tag { font-size: 10px; background: var(--color-bg-alt); padding: 1px 5px; border-radius: 3px; color: var(--color-text-muted); }
.mini-tag.more { font-weight: 700; }
.ticket-tags { display: flex; gap: 3px; align-items: center; }
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; }
.tags-editor { display: flex; flex-direction: column; gap: 6px; }
.tags-list { display: flex; flex-wrap: wrap; gap: 4px; }
.tag-chip { display: inline-flex; align-items: center; gap: 3px; font-size: 12px; background: var(--color-bg-alt); padding: 2px 8px; border-radius: 4px; }
.tag-remove { background: none; border: none; cursor: pointer; font-size: 14px; line-height: 1; color: var(--color-text-muted); padding: 0; }
.tag-input { flex: 1; }
.ticket-priority { font-size: 10px; font-weight: 600; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; }
.pri-low { background: #e0f2fe; color: #0369a1; }
.pri-medium { background: #fef3c7; color: #92400e; }
.pri-high { background: #fde68a; color: #b45309; }
.pri-critical { background: #fecaca; color: #b91c1c; }
.ticket-card-title { font-size: 14px; font-weight: 600; margin: 0 0 8px; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.ticket-card-meta { display: flex; gap: 8px; align-items: center; font-size: 11px; color: var(--color-text-muted); flex-wrap: wrap; }
.ticket-status { padding: 2px 6px; border-radius: 4px; font-weight: 600; }
.status-open { background: #dbeafe; color: #1e40af; }
.status-in_progress { background: #fef3c7; color: #92400e; }
.status-resolved { background: #d1fae5; color: #065f46; }
.status-closed { background: #e5e7eb; color: #374151; }
.ticket-assignee { display: flex; align-items: center; gap: 4px; }
.sla-badge { font-size: 10px; font-weight: 700; padding: 1px 5px; border-radius: 3px; text-transform: uppercase; }
.sla-ok { background: #d1fae5; color: #065f46; }
.sla-warning { background: #fef3c7; color: #92400e; }
.sla-breach { background: #fecaca; color: #b91c1c; }

/* Markdown editor tabs (matches AdminView pattern) */
:deep(.md-editor) { border: 1px solid var(--color-border); border-radius: var(--radius); overflow: hidden; }
:deep(.md-editor-tabs) { display: flex; background: var(--color-bg-alt); border-bottom: 1px solid var(--color-border); }
:deep(.md-tab) { padding: 6px 16px; font-size: 12px; font-weight: 600; cursor: pointer; background: none; border: none; border-bottom: 2px solid transparent; color: var(--color-text-muted); transition: color .15s, border-color .15s; }
:deep(.md-tab:hover) { color: var(--color-text); }
:deep(.md-tab.active) { color: var(--color-primary); border-bottom-color: var(--color-primary); }
:deep(.md-textarea) { border: none !important; border-radius: 0 !important; resize: vertical; }
:deep(.md-preview) { padding: 10px 12px; min-height: 80px; }
.pending-attachments { background: var(--color-bg-alt); border-radius: 6px; padding: 8px; margin-top: 4px; }
.pending-label { font-size: 11px; font-weight: 600; color: var(--color-text-muted); display: block; margin-bottom: 4px; }
</style>
