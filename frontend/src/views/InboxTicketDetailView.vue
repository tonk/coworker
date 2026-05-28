<template>
  <div class="ticket-detail-wrap" v-if="ticket">
    <div class="ticket-detail-header">
      <RouterLink to="/tickets/inbox" class="back-link">← {{ $t('inbox.title') }}</RouterLink>
      <span class="ticket-detail-id">#{{ ticket.id }}</span>
    </div>

    <!-- Title -->
    <div class="td-title-row">
      <input
        v-if="editingTitle"
        ref="titleInputRef"
        class="td-title-input"
        v-model="titleDraft"
        @blur="saveTitle"
        @keydown.enter.prevent="saveTitle"
        @keydown.escape="cancelTitle"
        :aria-label="$t('ticket.title')"
      />
      <h2 v-else class="td-title" @click="startEditTitle" :title="$t('common.click_to_edit')">{{ ticket.title }}</h2>
    </div>

    <div class="td-body">
      <!-- Left: metadata -->
      <div class="td-meta">
        <div class="td-meta-row">
          <span class="td-label">{{ $t('ticket.status') }}</span>
          <select class="td-select" :value="ticket.status" @change="updateField('status', $event.target.value)" :aria-label="$t('ticket.status')">
            <option v-for="s in statuses" :key="s" :value="s">{{ s }}</option>
          </select>
        </div>
        <div class="td-meta-row">
          <span class="td-label">{{ $t('ticket.priority') }}</span>
          <select class="td-select" :value="ticket.priority" @change="updateField('priority', $event.target.value)" :aria-label="$t('ticket.priority')">
            <option v-for="p in priorities" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div class="td-meta-row">
          <span class="td-label">{{ $t('ticket.type') }}</span>
          <select class="td-select" :value="ticket.type" @change="updateField('type', $event.target.value)" :aria-label="$t('ticket.type')">
            <option v-for="tp in types" :key="tp" :value="tp">{{ tp }}</option>
          </select>
        </div>
        <div class="td-meta-row td-meta-assign-customer">
          <span class="td-label">{{ $t('inbox.customer') }}</span>
          <div class="td-assign-customer">
            <select class="td-select" v-model="selectedCustomer" :aria-label="$t('inbox.assign_customer')">
              <option :value="null">{{ $t('inbox.no_customer') }}</option>
              <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
            <button
              v-if="selectedCustomer"
              class="td-btn td-btn-primary"
              @click="assignCustomer"
              :disabled="assigning"
            >{{ assigning ? $t('common.saving') : $t('inbox.assign') }}</button>
          </div>
        </div>
        <div class="td-meta-row">
          <span class="td-label">{{ $t('ticket.created_by') }}</span>
          <span class="td-value">{{ ticket.created_by?.display_name || ticket.created_by?.username }}</span>
        </div>
        <div class="td-meta-row">
          <span class="td-label">{{ $t('ticket.created_at') }}</span>
          <span class="td-value">{{ fmtDate(ticket.created_at) }}</span>
        </div>
        <div v-if="ticket.tags?.length || isAdmin" class="td-meta-row td-tags-row">
          <span class="td-label">{{ $t('ticket.tags') }}</span>
          <div class="td-tags">
            <span v-for="tag in ticket.tags" :key="tag.id" class="td-tag">{{ tag.name }}</span>
          </div>
        </div>
      </div>

      <!-- Right: description + messages -->
      <div class="td-content">
        <div class="td-section">
          <h3 class="td-section-title">{{ $t('ticket.description') }}</h3>
          <textarea
            class="td-desc"
            v-model="descDraft"
            :placeholder="$t('ticket.description_placeholder')"
            @blur="saveDescription"
            :aria-label="$t('ticket.description')"
          />
        </div>

        <div class="td-section">
          <h3 class="td-section-title">{{ $t('ticket.messages') }}</h3>
          <div class="td-messages">
            <div v-for="msg in ticket.messages" :key="msg.id" class="td-message">
              <div class="td-msg-meta">
                <strong>{{ msg.user?.display_name || msg.user?.username }}</strong>
                <span class="td-msg-date">{{ fmtDate(msg.created_at) }}</span>
              </div>
              <div class="td-msg-body">{{ msg.body }}</div>
            </div>
            <div v-if="!ticket.messages?.length" class="td-no-messages">{{ $t('ticket.no_messages') }}</div>
          </div>
          <div class="td-reply">
            <textarea
              class="td-reply-input"
              v-model="replyBody"
              :placeholder="$t('ticket.reply_placeholder')"
              rows="3"
              :aria-label="$t('ticket.reply')"
            />
            <button class="td-btn td-btn-primary" @click="sendMessage" :disabled="!replyBody.trim() || sendingMsg">
              {{ sendingMsg ? $t('common.saving') : $t('ticket.send_reply') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>

  <div v-else-if="loading" class="ticket-detail-loading">{{ $t('common.loading') }}</div>
  <div v-else class="ticket-detail-notfound">{{ $t('ticket.not_found') }}</div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ticketsApi } from '@/api/tickets'
import { customersApi } from '@/api/customers'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { useI18n } from 'vue-i18n'

const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()
const ui     = useUIStore()
const { t }  = useI18n()

const ticketId = computed(() => Number(route.params.ticketId))
const ticket   = ref(null)
const loading  = ref(false)

const editingTitle = ref(false)
const titleDraft   = ref('')
const titleInputRef = ref(null)
const descDraft    = ref('')

const statuses   = ['new', 'open', 'pending', 'pending_close', 'closed']
const priorities = ['low', 'medium', 'high', 'critical']
const types      = ['incident', 'problem', 'service_request', 'change_request']

const customers       = ref([])
const selectedCustomer = ref(null)
const assigning       = ref(false)
const replyBody       = ref('')
const sendingMsg      = ref(false)

const isAdmin = computed(() => auth.globalRole === 'admin')

function fmtDate(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString(undefined, { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true
  try {
    const { data } = await ticketsApi.inboxGet(ticketId.value)
    ticket.value = data
    titleDraft.value = data.title
    descDraft.value  = data.description || ''
  } catch {
    ticket.value = null
  } finally {
    loading.value = false
  }
}

async function loadCustomers() {
  try {
    const { data } = await customersApi.list()
    customers.value = Array.isArray(data) ? data : []
  } catch {
    customers.value = []
  }
}

function startEditTitle() {
  editingTitle.value = true
  nextTick(() => titleInputRef.value?.focus())
}

async function saveTitle() {
  editingTitle.value = false
  if (!titleDraft.value.trim() || titleDraft.value === ticket.value.title) return
  const { data } = await ticketsApi.inboxUpdate(ticketId.value, { title: titleDraft.value.trim() })
  ticket.value = data
}

function cancelTitle() {
  editingTitle.value = false
  titleDraft.value = ticket.value.title
}

async function saveDescription() {
  if (descDraft.value === (ticket.value.description || '')) return
  const { data } = await ticketsApi.inboxUpdate(ticketId.value, { description: descDraft.value })
  ticket.value = data
}

async function updateField(field, value) {
  const { data } = await ticketsApi.inboxUpdate(ticketId.value, { [field]: value })
  ticket.value = data
}

async function assignCustomer() {
  if (!selectedCustomer.value) return
  assigning.value = true
  try {
    await ticketsApi.inboxUpdate(ticketId.value, { customer_id: selectedCustomer.value })
    ui.success(t('inbox.assigned_success'))
    router.push(`/customers/${selectedCustomer.value}/tickets`)
  } catch {
    ui.error(t('common.error'))
  } finally {
    assigning.value = false
  }
}

async function sendMessage() {
  if (!replyBody.value.trim()) return
  sendingMsg.value = true
  try {
    await ticketsApi.inboxMessage(ticketId.value, replyBody.value.trim())
    replyBody.value = ''
    await load()
  } catch {
    ui.error(t('common.error'))
  } finally {
    sendingMsg.value = false
  }
}

onMounted(() => {
  load()
  loadCustomers()
})
</script>

<style scoped>
.ticket-detail-wrap { padding: 24px; max-width: 1100px; margin: 0 auto; }
.ticket-detail-header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.back-link { color: var(--color-text-muted); text-decoration: none; font-size: 0.9rem; }
.back-link:hover { color: var(--color-primary); }
.ticket-detail-id { color: var(--color-text-muted); font-size: 0.9rem; }

.td-title-row { margin-bottom: 20px; }
.td-title { font-size: 1.4rem; font-weight: 600; margin: 0; cursor: pointer; }
.td-title:hover { color: var(--color-primary); }
.td-title-input { font-size: 1.4rem; font-weight: 600; width: 100%; border: 1px solid var(--color-primary); border-radius: 6px; padding: 4px 8px; background: var(--color-surface); color: var(--color-text); outline: none; }

.td-body { display: grid; grid-template-columns: 260px 1fr; gap: 24px; }

.td-meta { display: flex; flex-direction: column; gap: 12px; }
.td-meta-row { display: flex; flex-direction: column; gap: 4px; }
.td-label { font-size: 0.75rem; font-weight: 600; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: .04em; }
.td-value { font-size: 0.9rem; }
.td-select { font-size: 0.85rem; padding: 5px 8px; border: 1px solid var(--color-border); border-radius: 6px; background: var(--color-surface); color: var(--color-text); cursor: pointer; }
.td-assign-customer { display: flex; gap: 8px; align-items: center; }
.td-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.td-tag { font-size: 0.75rem; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 4px; padding: 2px 7px; }

.td-content { display: flex; flex-direction: column; gap: 24px; }
.td-section-title { font-size: 0.9rem; font-weight: 600; margin: 0 0 10px; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: .04em; }
.td-desc { width: 100%; min-height: 100px; padding: 10px; border: 1px solid var(--color-border); border-radius: 6px; resize: vertical; font-size: 0.9rem; background: var(--color-surface); color: var(--color-text); font-family: inherit; }

.td-messages { display: flex; flex-direction: column; gap: 12px; margin-bottom: 16px; }
.td-message { background: var(--color-bg); border-radius: 6px; padding: 10px 14px; }
.td-msg-meta { display: flex; align-items: center; gap: 10px; margin-bottom: 4px; font-size: 0.82rem; }
.td-msg-date { color: var(--color-text-muted); }
.td-msg-body { font-size: 0.9rem; white-space: pre-wrap; }
.td-no-messages { color: var(--color-text-muted); font-size: 0.9rem; }
.td-reply { display: flex; flex-direction: column; gap: 8px; }
.td-reply-input { padding: 8px 10px; border: 1px solid var(--color-border); border-radius: 6px; resize: vertical; font-size: 0.9rem; background: var(--color-surface); color: var(--color-text); font-family: inherit; }

.td-btn { padding: 7px 16px; border-radius: 6px; border: none; cursor: pointer; font-size: 0.85rem; font-weight: 500; }
.td-btn-primary { background: var(--color-primary); color: #fff; }
.td-btn-primary:disabled { opacity: .5; cursor: not-allowed; }

.ticket-detail-loading, .ticket-detail-notfound { padding: 40px; text-align: center; color: var(--color-text-muted); }

@media (max-width: 700px) {
  .td-body { grid-template-columns: 1fr; }
}
</style>
