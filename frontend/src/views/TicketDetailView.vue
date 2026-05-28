<template>
  <main class="ticket-detail-main">
    <div v-if="loading" class="loading-state">
      <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <template v-else-if="ticket">
      <header class="detail-header">
        <RouterLink :to="isInbox ? '/tickets/inbox' : `/customers/${customerId}/tickets`" class="back-link">{{ $t('common.go_back') }}</RouterLink>
        <div class="detail-title-row">
          <div class="title-edit-wrap">
            <h1 v-if="!editingTitle" @dblclick="startTitleEdit" :title="$t('common.double_click_edit')">{{ ticket.title }}</h1>
            <input v-else ref="titleInput" v-model="editTitleVal" class="form-input title-input" @blur="saveTitle" @keydown.enter="saveTitle" @keydown.escape="cancelTitleEdit" />
          </div>
          <div class="detail-actions">
            <select class="form-input form-input-sm" :value="ticket.status" @change="updateStatus($event.target.value)">
              <option value="new">{{ $t('ticket.status_new') }}</option>
              <option value="open">{{ $t('ticket.status_open') }}</option>
              <option value="pending">{{ $t('ticket.status_pending') }}</option>
              <option value="pending_close">{{ $t('ticket.status_pending_close') }}</option>
              <option value="closed">{{ $t('ticket.status_closed') }}</option>
            </select>
            <select class="form-input form-input-sm" :value="ticket.priority" @change="updatePriority($event.target.value)">
              <option value="low">{{ $t('ticket.priority_low') }}</option>
              <option value="medium">{{ $t('ticket.priority_medium') }}</option>
              <option value="high">{{ $t('ticket.priority_high') }}</option>
              <option value="critical">{{ $t('ticket.priority_critical') }}</option>
            </select>
            <select v-if="macros.length" class="form-input form-input-sm" :value="''" :disabled="applyingMacro" @change="applyMacro($event.target.value); $event.target.value = ''">
              <option value="" disabled>{{ $t('macro.apply_macro') }}</option>
              <option v-for="m in macros" :key="m.id" :value="m.id">{{ m.name }}</option>
            </select>
            <button class="btn btn-danger btn-sm" @click="deleteTicket">{{ $t('common.delete') }}</button>
          </div>
        </div>
        <div class="detail-meta">
          <span class="ticket-type" :class="'type-' + ticket.type">{{ $t('ticket.type_' + ticket.type) }}</span>
          <span class="ticket-status" :class="'status-' + ticket.status">{{ $t('ticket.status_' + ticket.status) }}</span>
          <span v-if="ticket.sla_policy_id" :class="['sla-badge', ticket.sla_response_breached || ticket.sla_resolution_breached ? 'sla-breach' : slaWarning(ticket) ? 'sla-warning' : 'sla-ok']">
            {{ ticket.sla_policy?.name || $t('sla.sla') }}
            <template v-if="ticket.sla_response_breached">· {{ $t('sla.response_breached') }}</template>
            <template v-else-if="ticket.sla_resolution_breached">· {{ $t('sla.resolution_breached') }}</template>
          </span>
          <span class="ticket-id-copy" title="Click to copy ticket reference" @click="copyTicketRef">#{{ ticket.id }}</span>
          <span>{{ $t('ticket.created_by') }} {{ ticket.created_by?.display_name || ticket.created_by?.username }}</span>
          <span>{{ formatDateTime(ticket.created_at) }}</span>
          <span v-if="ticket.from_email" class="from-email-badge" :title="ticket.from_email">✉ {{ ticket.from_email }}</span>
          <span v-if="!isInbox" class="assignee-wrap">
            <label class="sr-only" for="assignee-select">{{ $t('ticket.assigned_to') }}</label>
            <select id="assignee-select" class="form-input form-input-sm" :value="ticket.assigned_to_id" @change="updateAssignedTo($event.target.value)">
              <option :value="null">—</option>
              <option v-for="u in customerUsers" :key="u.user_id" :value="u.user_id">{{ u.display_name || u.username }}</option>
            </select>
          </span>
          <span v-if="ticket.owner">{{ $t('ticket.owner') }} {{ ticket.owner.display_name || ticket.owner.username }}</span>
          <span v-if="ticket.group">{{ $t('ticket.group') }} {{ ticket.group.name }}</span>
        </div>
      </header>

      <div v-if="ticket.status === 'pending'" class="reminder-row">
        <span class="reminder-label">{{ $t('ticket.reminder_date') }}</span>
        <DatePicker :model-value="reminderDateValue" @update:model-value="updateReminderDate" />
      </div>
      <div v-if="ticket.status === 'pending_close'" class="reminder-row">
        <span class="reminder-label">{{ $t('ticket.close_date') }}</span>
        <DatePicker :model-value="closeDateValue" @update:model-value="updateCloseDate" />
      </div>

      <div v-if="!isInbox" class="detail-fields">
        <div class="detail-tags" v-if="ticket.tags?.length">
          <span v-for="tag in ticket.tags" :key="tag.id" class="tag-chip">
            #{{ tag.name }}
            <button class="tag-remove" @click="removeTag(tag)" title="Remove tag" aria-label="Remove tag">×</button>
          </span>
        </div>
        <div class="tag-input-row">
          <input class="form-input form-input-sm tag-input" v-model="newTagName" :placeholder="$t('ticket.add_tag_placeholder')" @keydown.enter.prevent="addTag" @keydown.comma.prevent="addTag" />
          <button class="btn btn-secondary btn-sm" @click="addTag" :disabled="!newTagName.trim()">{{ $t('common.add') }}</button>
        </div>
      </div>

      <div v-if="ticket.description" class="detail-description">
        <div class="detail-desc-header">
          <span class="detail-desc-label">{{ $t('ticket.original_email') }}</span>
          <span v-if="ticket.from_name" class="from-name">{{ ticket.from_name }}</span>
          <a v-if="ticket.from_email" :href="'mailto:' + ticket.from_email" class="from-email-link" :title="ticket.from_email">{{ ticket.from_email }}</a>
          <button v-if="ticket.email_message_id" class="btn btn-ghost btn-sm" @click="showOriginalEmail" style="margin-left:auto">{{ $t('ticket.show_original') }}</button>
        </div>
        <div class="email-body markdown-body selectable" v-html="renderMarkdown(ticket.description)"></div>
      </div>
      <AttachmentList v-if="ticket.attachments?.length" :attachments="ticket.attachments" :can-delete="false" />

      <div v-if="ticket.sla_policy_id && !isInbox" class="sla-card">
        <h4>{{ ticket.sla_policy?.name || $t('sla.sla') }}</h4>
        <div class="sla-card-body">
          <div class="sla-card-row" v-if="ticket.sla_response_deadline">
            <span class="sla-card-label">{{ $t('sla.response_by') }}</span>
            <span :class="['sla-card-value', ticket.sla_response_breached ? 'sla-card-breach' : '']">{{ formatDateTime(ticket.sla_response_deadline) }}</span>
            <span v-if="ticket.first_response_at" class="sla-card-check met"><span class="feat-check" style="color:var(--color-success)">✓</span> {{ $t('sla.first_response') }} {{ formatDateTime(ticket.first_response_at) }}</span>
            <span v-else-if="ticket.sla_response_breached" class="sla-card-status breach">{{ $t('sla.response_breached') }}</span>
            <span v-else-if="slaWarning(ticket)" class="sla-card-status warn">{{ $t('sla.warning') }}</span>
            <span v-else class="sla-card-status ok">{{ $t('sla.on_track') }}</span>
          </div>
          <div class="sla-card-row" v-if="ticket.sla_resolution_deadline">
            <span class="sla-card-label">{{ $t('sla.resolution_by') }}</span>
            <span :class="['sla-card-value', ticket.sla_resolution_breached ? 'sla-card-breach' : '']">{{ formatDateTime(ticket.sla_resolution_deadline) }}</span>
            <span v-if="ticket.status === 'pending_close'" class="sla-card-check met"><span class="feat-check" style="color:var(--color-success)">✓</span> {{ $t('ticket.status_' + ticket.status) }}</span>
            <span v-else-if="ticket.sla_resolution_breached" class="sla-card-status breach">{{ $t('sla.resolution_breached') }}</span>
            <span v-else-if="slaWarning(ticket)" class="sla-card-status warn">{{ $t('sla.warning') }}</span>
            <span v-else class="sla-card-status ok">{{ $t('sla.on_track') }}</span>
          </div>
        </div>
      </div>

      <div class="linked-tickets-section" v-if="linkedTickets !== null && !isInbox">
        <h4>{{ $t('ticket.linked_tickets') }} <span v-if="linkedTickets.length" class="count">({{ linkedTickets.length }})</span></h4>
        <div v-if="linkedTickets.length" class="linked-list">
          <div v-for="lt in linkedTickets" :key="lt.link_id" class="linked-row" @click="openTicket(lt)">
            <span class="linked-status" :class="'status-' + lt.status">{{ $t('ticket.status_' + lt.status) }}</span>
            <span class="linked-priority" :class="'pri-' + lt.priority">{{ lt.priority }}</span>
            <span class="linked-title">#{{ lt.id }} {{ lt.title }}</span>
            <button class="btn-icon-xs" @click.stop="removeLinkedTicket(lt)" :title="$t('ticket.remove_link')" aria-label="Remove link">✕</button>
          </div>
        </div>
        <div class="linked-add-row">
          <input class="form-input form-input-sm" v-model="newLinkId" :placeholder="$t('ticket.add_link_placeholder')" @keydown.enter.prevent="addLinkedTicket" />
          <button class="btn btn-secondary btn-sm" @click="addLinkedTicket" :disabled="!newLinkId.trim()">{{ $t('common.add') }}</button>
        </div>
      </div>

      <div v-if="linkedCards !== null && !isInbox" class="linked-tickets-section">
        <h4>{{ $t('card_ref.linked_cards') }} <span v-if="linkedCards.length" class="count">({{ linkedCards.length }})</span></h4>
        <div v-if="linkedCards.length" class="linked-list">
          <div v-for="lc in linkedCards" :key="lc.link_id" class="linked-row" @click="openCard(lc)">
            <span class="linked-ref">{{ lc.project_key }}-{{ lc.card_number }}</span>
            <span class="linked-title">{{ lc.title }}</span>
            <button class="btn-icon-xs" @click.stop="removeCardLink(lc)" :title="$t('ticket.remove_link')" aria-label="Remove link">✕</button>
          </div>
        </div>
        <div class="linked-add-row">
          <input class="form-input form-input-sm" v-model="newCardLink" :placeholder="$t('card_ref.add_link_placeholder')" @keydown.enter.prevent="addCardLink" />
          <button class="btn btn-secondary btn-sm" @click="addCardLink" :disabled="!newCardLink.trim()">{{ $t('common.add') }}</button>
        </div>
      </div>

      <div v-if="isInbox" class="move-section">
        <h4>{{ $t('inbox.customer') }}</h4>
        <div class="move-row">
          <label class="sr-only" for="assign-customer-select">{{ $t('inbox.assign_customer') }}</label>
          <select id="assign-customer-select" class="form-input form-input-sm" v-model="moveTargetCustomer">
            <option :value="null">—</option>
            <option v-for="c in allCustomers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
          <button class="btn btn-primary btn-sm" @click="assignToCustomer" :disabled="!moveTargetCustomer || moving">{{ $t('inbox.assign') }}</button>
        </div>
      </div>
      <div v-else class="move-section">
        <h4>{{ $t('ticket.move_to_customer') }}</h4>
        <div class="move-row">
          <label class="sr-only" for="move-customer-select">{{ $t('ticket.move_to_customer') }}</label>
          <select id="move-customer-select" class="form-input form-input-sm" v-model="moveTargetCustomer">
            <option :value="null">—</option>
            <option v-for="c in allCustomers" :key="c.id" :value="c.id" :disabled="c.id === customerId">{{ c.name }}</option>
          </select>
          <button class="btn btn-secondary btn-sm" @click="moveTicket" :disabled="!moveTargetCustomer || moveTargetCustomer === customerId || moving">{{ $t('ticket.move') }}</button>
        </div>
      </div>

      <section v-if="auth.timeTrackingEnabled && !isInbox" class="time-section">
        <h4>{{ $t('ticket.time_logged') }} <span v-if="ticketTimeEntries.length" class="count">({{ formatMinutes(totalMinutes) }})</span></h4>

        <div v-if="ticketTimeEntries.length" class="time-list">
          <div v-for="te in ticketTimeEntries" :key="te.id" class="time-row">
            <span class="time-date">{{ formatDate(te.date) }}</span>
            <span class="time-who">{{ te.user?.display_name || te.user?.username }}</span>
            <span class="time-dur">{{ formatMinutes(te.minutes) }}</span>
            <span v-if="te.project" class="time-proj">{{ te.project.name }}</span>
            <span class="time-desc">{{ te.description }}</span>
            <button v-if="te.user_id === auth.user?.id" class="btn-icon-xs" @click="deleteTimeEntry(te)" :title="$t('ticket.time_delete_entry')" :aria-label="$t('ticket.time_delete_entry')">✕</button>
          </div>
        </div>
        <p v-else class="time-empty">{{ $t('ticket.no_time_logged') }}</p>

        <form class="time-add-form" @submit.prevent="logTime">
          <label class="sr-only" for="te-date">{{ $t('ticket.time_date') }}</label>
          <DatePicker id="te-date" :model-value="teDate" @update:model-value="teDate = $event" />
          <label class="sr-only" for="te-hours">{{ $t('ticket.time_hours') }}</label>
          <input id="te-hours" class="form-input form-input-sm te-num" type="number" min="0" max="23" v-model.number="teHours" :placeholder="$t('ticket.time_hours')" />
          <label class="sr-only" for="te-mins">{{ $t('ticket.time_minutes') }}</label>
          <input id="te-mins" class="form-input form-input-sm te-num" type="number" min="0" max="59" step="5" v-model.number="teMins" :placeholder="$t('ticket.time_minutes')" />
          <label class="sr-only" for="te-project">{{ $t('board.select_project') }}</label>
          <select id="te-project" class="form-input form-input-sm te-project" v-model="teProjectId">
            <option :value="null">— {{ $t('board.select_project') }} —</option>
            <option v-for="p in customerProjects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <label class="sr-only" for="te-desc">{{ $t('ticket.time_note') }}</label>
          <input id="te-desc" class="form-input form-input-sm te-desc" type="text" v-model="teDesc" :placeholder="$t('ticket.time_note')" />
          <button type="submit" class="btn btn-secondary btn-sm" :disabled="teSubmitting || (teHours === 0 && teMins === 0)">{{ $t('ticket.log_time') }}</button>
        </form>
      </section>

      <section class="messages-section">
        <h2>{{ $t('ticket.messages') }} ({{ ticket.messages?.length || 0 }})</h2>

        <div v-if="!ticket.messages?.length" class="empty-state">
          {{ $t('ticket.no_messages') }}
        </div>

        <div v-else class="messages-list">
          <div v-for="msg in ticket.messages" :key="msg.id" class="message">
            <div class="message-header">
              <span class="msg-author-group">
                <strong>{{ msg.from_name || msg.user?.display_name || msg.user?.username }}</strong>
                <span v-if="msg.email_sent" class="email-sent-badge" :title="$t('ticket.email_sent_hint')">✉</span>
              </span>
              <span class="message-time">{{ formatDateTime(msg.created_at) }}</span>
            </div>
            <div class="message-body markdown-body selectable" v-html="renderMarkdown(msg.body)"></div>
            <AttachmentList v-if="msg.attachments?.length" :attachments="msg.attachments" :can-delete="false" />
          </div>
        </div>

        <form class="message-form" @submit.prevent="submitMessage">
          <div class="md-editor">
            <div class="md-editor-tabs" role="tablist">
              <button role="tab" :aria-selected="msgEditorTab === 'edit'" aria-controls="ticket-msg-panel-edit" id="ticket-msg-tab-edit" :class="['md-tab', { active: msgEditorTab === 'edit' }]" type="button" @click="msgEditorTab = 'edit'">{{ $t('common.edit') }}</button>
              <button role="tab" :aria-selected="msgEditorTab === 'preview'" aria-controls="ticket-msg-panel-preview" id="ticket-msg-tab-preview" :class="['md-tab', { active: msgEditorTab === 'preview' }]" type="button" @click="msgEditorTab = 'preview'">{{ $t('common.preview') }}</button>
            </div>
            <textarea v-if="msgEditorTab === 'edit'" id="ticket-msg-panel-edit" role="tabpanel" aria-labelledby="ticket-msg-tab-edit" v-model="newMessage" class="form-input md-textarea" rows="3" :placeholder="$t('ticket.message_placeholder')" required @paste="onMsgPaste"></textarea>
            <div v-else id="ticket-msg-panel-preview" role="tabpanel" aria-labelledby="ticket-msg-tab-preview" class="md-preview markdown-body" v-html="renderMarkdown(newMessage)"></div>
          </div>
          <div v-if="pendingFiles.length" class="pending-attachments">
            <span class="pending-label">{{ $t('ticket.pending_files') }}</span>
            <AttachmentList :attachments="pendingFiles" :can-delete="true" @remove="removePending" />
          </div>
          <div class="message-form-actions">
            <FileUploadButton @files-selected="onFilesSelected" />
            <button type="submit" class="btn btn-primary btn-sm" :disabled="(!newMessage.trim() && !pendingFiles.length) || sending">{{ $t('ticket.send') }}<span v-if="pendingFiles.length" class="pending-badge">· {{ pendingFiles.length }}</span></button>
          </div>
        </form>
      </section>

      <section v-if="!isInbox" class="activity-section">
        <button class="activity-toggle" @click="showActivity = !showActivity" :aria-expanded="showActivity">
          <span class="activity-toggle-icon" :class="{ rotated: showActivity }">▸</span>
          {{ $t('ticket.history') }} <span class="activity-count">({{ history.length }})</span>
        </button>
        <div v-if="showActivity">
          <div v-if="!history.length" class="activity-empty">{{ $t('common.no_results') }}</div>
          <div v-else class="activity-list">
            <div v-for="h in history" :key="h.id" class="activity-entry">
              <span class="activity-icon" :class="'activity-type-' + h.event_type" aria-hidden="true">{{ activityIcon(h.event_type) }}</span>
              <span class="activity-time">{{ formatDateTime(h.created_at) }}</span>
              <span class="activity-who">{{ h.user?.display_name || h.user?.username }}</span>
              <span class="activity-detail">{{ activityLabel(h) }}</span>
            </div>
          </div>
        </div>
      </section>
    </template>

    <div v-else class="empty-state">
      {{ $t('ticket.not_found') }}
    </div>

    <BaseModal v-if="showRawEmail" title="Original email" @close="showRawEmail = false" :resizable="true" style="--modal-width: 800px">
      <pre class="raw-email-pre">{{ rawEmailContent }}</pre>
    </BaseModal>
  </main>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { ticketsApi } from '@/api/tickets'
import { macrosApi } from '@/api/macros'
import client from '@/api/client'
import { attachmentsApi } from '@/api/attachments'
import { customersApi } from '@/api/customers'
import { timeEntriesApi } from '@/api/timeEntries'
import { useUIStore } from '@/stores/ui'
import { useAuthStore } from '@/stores/auth'
import { useTicketsStore } from '@/stores/tickets'
import { useDateFormat } from '@/composables/useDateFormat'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import AttachmentList from '@/components/common/AttachmentList.vue'
import FileUploadButton from '@/components/common/FileUploadButton.vue'
import DatePicker from '@/components/common/DatePicker.vue'
import BaseModal from '@/components/common/BaseModal.vue'

const { formatDateTime, formatDate } = useDateFormat()
const auth = useAuthStore()
const ticketsStore = useTicketsStore()

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}

async function showOriginalEmail() {
  if (rawEmailContent.value) { showRawEmail.value = true; return }
  rawEmailLoading.value = true
  try {
    const res = await client.get(`/customers/${customerId.value}/tickets/${ticketId.value}/raw-email`, { responseType: 'text' })
    rawEmailContent.value = res.data
    showRawEmail.value = true
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to load original email')
  } finally {
    rawEmailLoading.value = false
  }
}

const descEditorTab = ref('preview')
const msgEditorTab = ref('edit')

const route = useRoute()
const router = useRouter()
const ui = useUIStore()

const isInbox = computed(() => route.name === 'inbox-ticket-detail')
const customerId = computed(() => isInbox.value ? null : Number(route.params.id))
const ticketId = computed(() => Number(route.params.ticketId))
const ticket = ref(null)
const showRawEmail = ref(false)
const rawEmailContent = ref('')
const rawEmailLoading = ref(false)
const history = ref([])
const showActivity = ref(false)
const loading = ref(true)
const newMessage = ref('')
const pendingFiles = ref([])
const sending = ref(false)
const newTagName = ref('')
const linkedTickets = ref(null)
const newLinkId = ref('')
const linkedCards = ref(null)
const newCardLink = ref('')
const customerUsers = ref([])
const editingTitle = ref(false)
const editTitleVal = ref('')
const titleInput = ref(null)
const allCustomers = ref([])
const moveTargetCustomer = ref(null)
const moving = ref(false)
const macros = ref([])
const applyingMacro = ref(false)

function apiUpdate(data) {
  return isInbox.value
    ? ticketsApi.inboxUpdate(ticketId.value, data)
    : ticketsApi.update(customerId.value, ticketId.value, data)
}

function startTitleEdit() {
  editTitleVal.value = ticket.value?.title || ''
  editingTitle.value = true
  nextTick(() => titleInput.value?.focus())
}

async function saveTitle() {
  const val = editTitleVal.value.trim()
  if (!val || val === ticket.value?.title) {
    editingTitle.value = false
    return
  }
  try {
    const { data } = await apiUpdate({ title: val })
    ticket.value = data
    ui.success('Title updated')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to update title')
  }
  editingTitle.value = false
}

function cancelTitleEdit() {
  editingTitle.value = false
}

function titleWithDatePrefix(title, isoDate) {
  const stripped = title.replace(/^\[\d{4}-\d{2}-\d{2}\]\s*/, '')
  if (!isoDate) return stripped
  return `[${isoDate}] ${stripped}`
}

const reminderDateValue = computed(() => {
  if (!ticket.value?.reminder_at) return null
  const d = new Date(ticket.value.reminder_at)
  return d.toISOString().slice(0, 10)
})

async function updateReminderDate(val) {
  try {
    const d = val ? new Date(val + 'T12:00:00Z').toISOString() : null
    const newTitle = titleWithDatePrefix(ticket.value.title, val || null)
    const payload = { reminder_at: d }
    if (newTitle !== ticket.value.title) payload.title = newTitle
    const { data } = await apiUpdate(payload)
    ticket.value = data
    ui.success(val ? 'Reminder date updated' : 'Reminder cleared')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to update reminder')
  }
}

const closeDateValue = computed(() => {
  if (!ticket.value?.close_at) return null
  const d = new Date(ticket.value.close_at)
  return d.toISOString().slice(0, 10)
})

async function updateCloseDate(val) {
  try {
    const d = val ? new Date(val + 'T12:00:00Z').toISOString() : null
    const newTitle = titleWithDatePrefix(ticket.value.title, val || null)
    const payload = { close_at: d }
    if (newTitle !== ticket.value.title) payload.title = newTitle
    const { data } = await apiUpdate(payload)
    ticket.value = data
    ui.success(val ? 'Close date updated' : 'Close date cleared')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to update close date')
  }
}

async function fetchTicket() {
  loading.value = true
  pendingFiles.value = []
  newMessage.value = ''
  linkedTickets.value = null
  linkedCards.value = null
  try {
    const { data } = isInbox.value
      ? await ticketsApi.inboxGet(ticketId.value)
      : await ticketsApi.get(customerId.value, ticketId.value)
    ticket.value = data
    if (ticket.value.attachments === null) ticket.value.attachments = []
    if (ticket.value.messages) {
      for (const m of ticket.value.messages) {
        if (m.attachments === null) m.attachments = []
      }
    }

    // Load customers for assign/move panel
    try {
      const { data: customers } = await customersApi.list()
      allCustomers.value = customers || []
    } catch {
      allCustomers.value = []
    }

    if (!isInbox.value) {
      // Fetch customer members for assignee dropdown
      try {
        const { data: members } = await client.get(`/customers/${customerId.value}/members`)
        customerUsers.value = members || []
      } catch {
        customerUsers.value = []
      }
      // Fetch customer projects for time tracking
      try {
        const { data: custDetail } = await customersApi.get(customerId.value)
        const contractProjects = (custDetail.contracts || []).flatMap(cg => cg.projects || [])
        const unassigned = custDetail.projects || []
        customerProjects.value = [...contractProjects, ...unassigned]
      } catch {
        customerProjects.value = []
      }
      // Fetch linked tickets
      try {
        const { data: links } = await ticketsApi.listLinks(customerId.value, ticketId.value)
        linkedTickets.value = links || []
      } catch {
        linkedTickets.value = []
      }
      // Fetch linked cards
      try {
        const { data: cards } = await ticketsApi.listCards(customerId.value, ticketId.value)
        linkedCards.value = cards || []
      } catch {
        linkedCards.value = []
      }
      // Fetch time entries
      await fetchTimeEntries()
      // Fetch history
      try {
        const { data: h } = await ticketsApi.getHistory(customerId.value, ticketId.value)
        history.value = h || []
      } catch {
        history.value = []
      }
    }
  } catch {}
  loading.value = false
}

async function loadMacros() {
  try {
    const { data } = await macrosApi.list()
    macros.value = data || []
  } catch {
    macros.value = []
  }
}

async function applyMacro(macroId) {
  if (!macroId) return
  applyingMacro.value = true
  try {
    const { data } = isInbox.value
      ? await macrosApi.applyInbox(ticketId.value, macroId)
      : await macrosApi.apply(customerId.value, ticketId.value, macroId)
    ticket.value = data.ticket
    if (data.macro_messages?.length) {
      newMessage.value = data.macro_messages.join('\n\n')
      msgEditorTab.value = 'edit'
      await nextTick()
      document.getElementById('ticket-msg-panel-edit')?.focus()
    }
    ui.success('Macro applied')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to apply macro')
  } finally {
    applyingMacro.value = false
  }
}

watch([() => route.params.ticketId, () => route.name], () => {
  fetchTicket()
})

onMounted(() => { fetchTicket(); loadMacros() })

// Refresh messages when IMAP delivers a reply to this ticket
async function refreshMessages() {
  try {
    const { data } = isInbox.value
      ? await ticketsApi.inboxGet(ticketId.value)
      : await ticketsApi.get(customerId.value, ticketId.value)
    if (data?.messages && ticket.value) {
      for (const m of data.messages) {
        if (m.attachments === null) m.attachments = []
      }
      ticket.value = { ...ticket.value, messages: data.messages }
    }
  } catch {}
}

watch(() => ticketsStore.ticketRefreshKey, () => {
  if (ticketsStore.refreshForTicketId === ticketId.value) {
    refreshMessages()
  }
})

async function updateStatus(status) {
  try {
    const payload = { status }
    if (status === 'pending' && ticket.value.reminder_at) {
      const iso = new Date(ticket.value.reminder_at).toISOString().slice(0, 10)
      const newTitle = titleWithDatePrefix(ticket.value.title, iso)
      if (newTitle !== ticket.value.title) payload.title = newTitle
    }
    if (status === 'pending_close' && ticket.value.close_at) {
      const iso = new Date(ticket.value.close_at).toISOString().slice(0, 10)
      const newTitle = titleWithDatePrefix(ticket.value.title, iso)
      if (newTitle !== ticket.value.title) payload.title = newTitle
    }
    const { data } = await apiUpdate(payload)
    ticket.value = data
    ui.success('Status updated')
    if (status === 'pending' || status === 'pending_close') {
      await nextTick()
      document.querySelector('.reminder-row')?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
    }
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to update status')
  }
}

async function updatePriority(priority) {
  try {
    const { data } = await apiUpdate({ priority })
    ticket.value = data
    ui.success('Priority updated')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to update priority')
  }
}

async function updateAssignedTo(userId) {
  try {
    const val = userId ? Number(userId) : null
    const { data } = await ticketsApi.update(customerId.value, ticketId.value, { assigned_to_id: val })
    ticket.value = data
    ui.success('Assignee updated')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to update assignee')
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

async function onMsgPaste(e) {
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

async function submitMessage() {
  const body = newMessage.value.trim()
  if (!body && !pendingFiles.value.length) return
  sending.value = true
  try {
    const sendBody = body || '📎'
    let newMsg
    if (isInbox.value) {
      const { data } = await ticketsApi.inboxMessage(ticketId.value, sendBody)
      newMsg = { ...data, attachments: [] }
    } else {
      const { data } = await ticketsApi.addMessage(customerId.value, ticketId.value, sendBody)
      newMsg = { ...data, attachments: [] }
      if (pendingFiles.value.length) {
        const filesToUpload = [...pendingFiles.value]
        pendingFiles.value = []
        filesToUpload.forEach(pf => { if (pf._previewUrl) URL.revokeObjectURL(pf._previewUrl) })
        for (const pf of filesToUpload) {
          const fd = new FormData()
          fd.append('file', pf._file)
          fd.append('owner_type', 'ticket_message')
          fd.append('owner_id', String(newMsg.id))
          try {
            const { data: att } = await attachmentsApi.upload(fd)
            newMsg.attachments.push(att)
          } catch {}
        }
      }
    }

    if (!ticket.value.messages) ticket.value.messages = []
    ticket.value.messages.push(newMsg)
    newMessage.value = ''
    await nextTick()
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to send message')
  } finally {
    sending.value = false
  }
}

async function addTag() {
  const name = newTagName.value.trim().replace(/^#/, '').toLowerCase()
  if (!name) return
  try {
    const { data } = await ticketsApi.addTag(customerId.value, ticketId.value, name)
    if (!ticket.value.tags) ticket.value.tags = []
    if (!ticket.value.tags.some(t => t.id === data.id)) {
      ticket.value.tags = [...ticket.value.tags, data]
    }
    newTagName.value = ''
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to add tag')
  }
}

async function removeTag(tag) {
  try {
    await ticketsApi.removeTag(customerId.value, ticketId.value, tag.id)
    ticket.value.tags = (ticket.value.tags || []).filter(t => t.id !== tag.id)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to remove tag')
  }
}

async function addLinkedTicket() {
  const targetId = parseInt(newLinkId.value.trim(), 10)
  if (!targetId) return
  try {
    const { data } = await ticketsApi.addLink(customerId.value, ticketId.value, targetId)
    if (!linkedTickets.value) linkedTickets.value = []
    linkedTickets.value.push(data)
    newLinkId.value = ''
    ui.success('Ticket linked')
  } catch (e) {
    const msg = e.response?.data?.error || 'Failed to link ticket'
    ui.error(msg)
  }
}

async function removeLinkedTicket(lt) {
  try {
    await ticketsApi.removeLink(customerId.value, ticketId.value, lt.link_id)
    linkedTickets.value = (linkedTickets.value || []).filter(l => l.link_id !== lt.link_id)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to remove link')
  }
}

async function addCardLink() {
  const val = newCardLink.value.trim()
  if (!val) return
  const body = /^\d+$/.test(val) ? { card_id: parseInt(val, 10) } : { ref: val }
  try {
    await ticketsApi.addCardLink(customerId.value, ticketId.value, body)
    const { data: cards } = await ticketsApi.listCards(customerId.value, ticketId.value)
    linkedCards.value = cards || []
    newCardLink.value = ''
    ui.success('Card linked')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to link card')
  }
}

async function removeCardLink(lc) {
  try {
    await ticketsApi.removeCardLink(customerId.value, ticketId.value, lc.link_id)
    linkedCards.value = (linkedCards.value || []).filter(c => c.link_id !== lc.link_id)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to remove link')
  }
}

function openTicket(t) {
  const cid = t.customer_id || customerId.value
  router.push(`/customers/${cid}/tickets/${t.id}`)
}

function openCard(c) {
  router.push({ name: 'board', params: { slug: c.project_slug }, query: { card: c.card_id } })
}

function slaWarning(t) {
  if (!t.sla_response_deadline && !t.sla_resolution_deadline) return false
  if (t.status === 'pending_close') return false
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

function activityIcon(type) {
  const icons = {
    created: '✦',
    status_changed: '⟳',
    title_changed: '✎',
    priority_changed: '⚑',
    type_changed: '▤',
    assignee_changed: '◉',
    owner_changed: '◎',
    group_changed: '⊞',
    comment_added: '✉',
    ticket_linked: '🔗',
    ticket_unlinked: '⛓',
    card_linked: '📋',
    card_unlinked: '📄',
    tag_added: '+',
    tag_removed: '−',
    dates_changed: '📅',
    customer_moved: '→',
  }
  return icons[type] || '•'
}

function activityLabel(h) {
  const t = (key) => { try { return h.detail || key } catch { return h.detail || key } }
  switch (h.event_type) {
    case 'created': return 'created this ticket'
    case 'status_changed': return `changed status to ${h.detail}`
    case 'title_changed': return `changed title to "${h.detail}"`
    case 'priority_changed': return `changed priority to ${h.detail}`
    case 'type_changed': return `changed type to ${h.detail}`
    case 'assignee_changed': return `changed assignee to ${h.detail}`
    case 'owner_changed': return `changed owner to ${h.detail}`
    case 'group_changed': return `changed group to ${h.detail}`
    case 'comment_added': return 'added a comment'
    case 'ticket_linked': return `linked ticket: ${h.detail}`
    case 'ticket_unlinked': return 'removed a ticket link'
    case 'card_linked': return 'linked a card'
    case 'card_unlinked': return 'unlinked a card'
    case 'tag_added': return `added tag #${h.detail}`
    case 'tag_removed': return `removed tag #${h.detail}`
    case 'dates_changed': return 'updated reminder / close dates'
    case 'customer_moved': return `moved to customer: ${h.detail}`
    default: return h.detail || h.event_type
  }
}

// --- Time tracking ---
const ticketTimeEntries = ref([])
const customerProjects = ref([])
const teDate = ref(new Date().toISOString().slice(0, 10))
const teHours = ref(0)
const teMins = ref(0)
const teProjectId = ref(null)
const teDesc = ref('')
const teSubmitting = ref(false)

const totalMinutes = computed(() => ticketTimeEntries.value.reduce((s, e) => s + e.minutes, 0))

function formatMinutes(mins) {
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}

async function fetchTimeEntries() {
  if (!auth.timeTrackingEnabled) return
  try {
    const { data } = await timeEntriesApi.list({ ticket_id: ticketId.value })
    ticketTimeEntries.value = data || []
  } catch {
    ticketTimeEntries.value = []
  }
}

async function logTime() {
  const minutes = teHours.value * 60 + teMins.value
  if (minutes <= 0) return
  teSubmitting.value = true
  try {
    const { data } = await timeEntriesApi.create({
      ticket_id: ticketId.value,
      project_id: teProjectId.value || undefined,
      date: teDate.value,
      minutes,
      description: teDesc.value.trim(),
    })
    ticketTimeEntries.value.unshift(data)
    teHours.value = 0
    teMins.value = 0
    teProjectId.value = null
    teDesc.value = ''
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to log time')
  } finally {
    teSubmitting.value = false
  }
}

async function deleteTimeEntry(te) {
  if (!await ui.confirm('Delete this time entry?')) return
  try {
    await timeEntriesApi.remove(te.id)
    ticketTimeEntries.value = ticketTimeEntries.value.filter(e => e.id !== te.id)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to delete time entry')
  }
}

async function copyTicketRef() {
  const text = `Ticket#${ticket.value.id}`
  try {
    await navigator.clipboard.writeText(text)
    ui.success('Copied ' + text)
  } catch {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    ui.success('Copied ' + text)
  }
}

async function deleteTicket() {
  if (!await ui.confirm('Delete this ticket?')) return
  try {
    if (isInbox.value) {
      await ticketsApi.inboxDelete(ticketId.value)
      router.push('/tickets/inbox')
    } else {
      await ticketsApi.delete(customerId.value, ticketId.value)
      router.push(`/customers/${customerId.value}/tickets`)
    }
    ui.success('Ticket deleted')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to delete ticket')
  }
}

async function moveTicket() {
  if (!moveTargetCustomer.value || moveTargetCustomer.value === customerId.value) return
  moving.value = true
  try {
    await ticketsApi.move(customerId.value, ticketId.value, moveTargetCustomer.value)
    ui.success('Ticket moved')
    router.push(`/customers/${moveTargetCustomer.value}/tickets/${ticketId.value}`)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to move ticket')
  } finally {
    moving.value = false
  }
}

async function assignToCustomer() {
  if (!moveTargetCustomer.value) return
  moving.value = true
  try {
    await ticketsApi.inboxUpdate(ticketId.value, { customer_id: moveTargetCustomer.value })
    ui.success('Ticket assigned')
    router.push(`/customers/${moveTargetCustomer.value}/tickets/${ticketId.value}`)
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to assign ticket')
  } finally {
    moving.value = false
  }
}
</script>

<style scoped>
.ticket-detail-main { padding: 24px; max-width: 900px; margin: 0 auto; }
.loading-state { display: flex; justify-content: center; padding: 48px; }
.empty-state { text-align: center; padding: 64px 24px; color: var(--color-text-muted); }
.back-link { font-size: 13px; color: var(--color-primary); text-decoration: none; }
.back-link:hover { text-decoration: underline; }
.detail-header { margin-bottom: 24px; }
.detail-title-row { display: flex; align-items: center; gap: 12px; margin: 8px 0; }
.detail-title-row h1 { flex: 1; margin: 0; font-size: 22px; }
.detail-actions { display: flex; gap: 8px; }
.detail-meta { display: flex; gap: 12px; align-items: center; font-size: 12px; color: var(--color-text-muted); flex-wrap: wrap; }
.from-email-badge { font-size: 11px; color: var(--color-text-muted); font-style: italic; }
.ticket-status { padding: 2px 8px; border-radius: 4px; font-weight: 600; font-size: 12px; }
.status-new { background: #dbeafe; color: #1e40af; }
.status-open { background: #fef3c7; color: #92400e; }
.status-pending { background: #f0e6ff; color: #6b21a8; }
.status-pending_close { background: #d1fae5; color: #065f46; }
.status-closed { background: #e5e7eb; color: #374151; }
.detail-desc-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}
.detail-desc-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-text-muted);
}
.from-email-link {
  font-size: 12px;
  color: var(--color-primary);
  text-decoration: none;
}
.from-email-link:hover {
  text-decoration: underline;
}
.from-name {
  font-size: 13px;
  color: var(--color-text);
  font-weight: 500;
  margin-left: 8px;
}
.email-body {
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--color-text);
}
.raw-email-pre {
  font-size: 12px;
  line-height: 1.4;
  white-space: pre;
  word-break: normal;
  overflow-x: auto;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 12px;
  max-height: 70vh;
  margin: 0;
  font-family: var(--font-mono);
}
.detail-description {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}
.msg-author-group {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.email-sent-badge {
  font-size: 13px;
  color: var(--color-primary);
  cursor: default;
}
.messages-section h2 { font-size: 16px; margin: 0 0 16px; }
.messages-list { display: flex; flex-direction: column; gap: 12px; margin-bottom: 16px; }
.message {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 12px;
}
.message-header { display: flex; justify-content: space-between; font-size: 12px; margin-bottom: 6px; }
.message-time { color: var(--color-text-muted); }
.message-body { font-size: 14px; }
.message-form { display: flex; flex-direction: column; gap: 8px; }
.message-form-actions { display: flex; align-items: center; gap: 8px; justify-content: flex-end; }
.form-input-sm { height: 28px; font-size: 12px; padding: 4px 8px; }
.pending-badge { font-size: 11px; color: var(--color-primary); font-weight: 600; }
.pending-attachments { background: var(--color-bg-alt); border-radius: 6px; padding: 8px; margin-top: 4px; }
.pending-label { font-size: 11px; font-weight: 600; color: var(--color-text-muted); display: block; margin-bottom: 4px; }
.sla-badge { font-size: 10px; font-weight: 700; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; }
.sla-ok { background: #d1fae5; color: #065f46; }
.sla-warning { background: #fef3c7; color: #92400e; }
.sla-breach { background: #fecaca; color: #b91c1c; }
.sla-deadline { font-size: 12px; color: var(--color-text-muted); }
.sla-deadline-breach { color: #b91c1c; font-weight: 600; }
.sla-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 24px;
}
.sla-card h4 { margin: 0 0 10px; font-size: 14px; }
.sla-card-body { display: flex; flex-direction: column; gap: 6px; }
.sla-card-row { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.sla-card-label { color: var(--color-text-muted); min-width: 100px; }
.sla-card-value { font-weight: 600; }
.sla-card-breach { color: #b91c1c; }
.sla-card-status { font-size: 11px; font-weight: 700; padding: 1px 6px; border-radius: 3px; text-transform: uppercase; }
.sla-card-status.ok { background: #d1fae5; color: #065f46; }
.sla-card-status.warn { background: #fef3c7; color: #92400e; }
.sla-card-status.breach { background: #fecaca; color: #b91c1c; }
.sla-card-check { font-size: 11px; color: var(--color-text-muted); }
.ticket-id-copy { cursor: pointer; font-variant-numeric: tabular-nums; }
.ticket-id-copy:hover { color: var(--color-primary); }
.ticket-type { font-size: 11px; font-weight: 600; padding: 2px 8px; border-radius: 4px; }
.type-incident { background: #fecaca; color: #b91c1c; }
.type-problem { background: #fed7aa; color: #9a3412; }
.type-service_request { background: #d1fae5; color: #065f46; }
.type-change_request { background: #dbeafe; color: #1e40af; }
.detail-fields { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; margin-bottom: 16px; }
.detail-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.tag-chip { display: inline-flex; align-items: center; gap: 3px; font-size: 12px; background: var(--color-bg-alt); padding: 2px 8px; border-radius: 4px; }
.tag-remove { background: none; border: none; cursor: pointer; font-size: 14px; line-height: 1; color: var(--color-text-muted); padding: 0; }
.tag-input-row { display: flex; gap: 4px; align-items: center; }
.tag-input { flex: 1; min-width: 120px; }
.linked-tickets-section { margin-bottom: 24px; }
.linked-tickets-section h4 { font-size: 14px; margin: 0 0 8px; display: flex; align-items: center; gap: 6px; }
.linked-tickets-section .count { font-weight: 400; color: var(--color-text-muted); }
.linked-list { display: flex; flex-direction: column; gap: 4px; margin-bottom: 8px; }
.linked-row { display: flex; align-items: center; gap: 8px; padding: 6px 8px; border-radius: 6px; cursor: pointer; font-size: 13px; }
.linked-row:hover { background: var(--color-bg-alt); }
.linked-row .linked-title { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.linked-status { padding: 1px 5px; border-radius: 3px; font-size: 10px; font-weight: 600; }
.linked-ref { font-family: monospace; font-size: 11px; font-weight: 700; color: var(--color-primary); min-width: 70px; }
.linked-priority { font-size: 10px; font-weight: 600; color: var(--color-text-muted); }
.linked-add-row { display: flex; gap: 4px; }
.linked-add-row input { flex: 1; }
.btn-icon-xs { background: none; border: none; cursor: pointer; color: var(--color-text-muted); padding: 2px; font-size: 12px; line-height: 1; }
.btn-icon-xs:hover { color: var(--color-danger, #ef4444); }

/* Markdown editor tabs (matches AdminView pattern) */
:deep(.md-editor) { border: 1px solid var(--color-border); border-radius: var(--radius); overflow: hidden; }
:deep(.md-editor-tabs) { display: flex; background: var(--color-bg-alt); border-bottom: 1px solid var(--color-border); }
:deep(.md-tab) { padding: 6px 16px; font-size: 12px; font-weight: 600; cursor: pointer; background: none; border: none; border-bottom: 2px solid transparent; color: var(--color-text-muted); transition: color .15s, border-color .15s; }
:deep(.md-tab:hover) { color: var(--color-text); }
:deep(.md-tab.active) { color: var(--color-primary); border-bottom-color: var(--color-primary); }
:deep(.md-textarea) { border: none !important; border-radius: 0 !important; resize: vertical; min-height: 80px; }
:deep(.md-preview) { padding: 10px 12px; min-height: 80px; }

.assignee-wrap { display: inline-flex; align-items: center; }
.assignee-wrap select { font-size: 12px; padding: 1px 6px; height: 22px; max-width: 140px; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); border: 0; }
.title-edit-wrap h1 { cursor: pointer; }
.title-edit-wrap h1:hover { background: var(--color-bg-alt); border-radius: 4px; }
.title-input { font-size: 22px; font-weight: 700; width: 100%; }
.reminder-row { display: flex; align-items: center; gap: 8px; margin-bottom: 16px; padding: 8px 12px; background: #f0e6ff; border-radius: 6px; }
.reminder-label { font-size: 12px; font-weight: 600; color: #6b21a8; }
.reminder-row :deep(.dp-trigger) { color: #6b21a8; border-color: color-mix(in srgb, #6b21a8 30%, transparent); background: transparent; font-weight: 600; }
.reminder-row :deep(.dp-trigger:hover) { background: color-mix(in srgb, #6b21a8 10%, transparent); }
.reminder-row :deep(.dp-trigger-empty) { color: #6b21a8; opacity: 0.6; }

/* Time tracking section */
.time-section { margin-bottom: 24px; }
.time-section h4 { font-size: 14px; margin: 0 0 8px; display: flex; align-items: center; gap: 6px; }
.time-section .count { font-weight: 400; color: var(--color-text-muted); }
.time-list { display: flex; flex-direction: column; gap: 4px; margin-bottom: 8px; }
.time-row { display: flex; align-items: center; gap: 8px; padding: 6px 8px; border-radius: 6px; font-size: 13px; background: var(--color-bg); border: 1px solid var(--color-border); }
.time-date { color: var(--color-text-muted); min-width: 80px; font-size: 12px; }
.time-who { color: var(--color-text-muted); min-width: 80px; font-size: 12px; }
.time-dur { font-weight: 600; min-width: 50px; }
.time-desc { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.time-proj { font-size: 11px; color: var(--color-primary); background: var(--color-bg-alt); padding: 1px 6px; border-radius: 3px; white-space: nowrap; }
.time-empty { font-size: 13px; color: var(--color-text-muted); margin: 0 0 8px; }
.time-add-form { display: flex; gap: 6px; align-items: center; flex-wrap: wrap; }
.te-num { width: 60px; text-align: center; }
.te-project { min-width: 130px; max-width: 180px; }
.te-desc { flex: 1; min-width: 120px; }

.activity-section { margin-top: 24px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.activity-toggle { display: flex; align-items: center; gap: 6px; width: 100%; background: none; border: none; cursor: pointer; font-size: 14px; color: var(--color-text-muted); padding: 4px 0; margin-bottom: 12px; }
.activity-toggle:hover { color: var(--color-text); }
.activity-toggle-icon { transition: transform .15s; font-size: 10px; }
.activity-toggle-icon.rotated { transform: rotate(90deg); }
.activity-count { font-weight: 400; color: var(--color-text-muted); }
.activity-empty { font-size: 12px; color: var(--color-text-muted); }
.activity-list { display: flex; flex-direction: column; gap: 6px; }
.activity-entry { display: flex; align-items: baseline; gap: 8px; font-size: 12px; }
.activity-icon { font-size: 11px; flex-shrink: 0; width: 14px; text-align: center; }
.activity-time { color: var(--color-text-muted); flex-shrink: 0; }
.activity-who { font-weight: 600; flex-shrink: 0; }
.activity-detail { color: var(--color-text-muted); }
.activity-type-created { color: #22c55e; }
.activity-type-status_changed { color: var(--color-primary); }
.activity-type-priority_changed { color: #f59e0b; }
.activity-type-title_changed,
.activity-type-type_changed,
.activity-type-assignee_changed,
.activity-type-owner_changed,
.activity-type-group_changed,
.activity-type-comment_added,
.activity-type-ticket_linked,
.activity-type-ticket_unlinked,
.activity-type-card_linked,
.activity-type-card_unlinked,
.activity-type-tag_added,
.activity-type-tag_removed,
.activity-type-dates_changed,
.activity-type-customer_moved { color: var(--color-text-muted); }

.move-section { margin-bottom: 24px; }
.move-section h4 { font-size: 14px; margin: 0 0 8px; }
.move-row { display: flex; gap: 6px; align-items: center; }
.move-row select { flex: 1; max-width: 280px; }
</style>
