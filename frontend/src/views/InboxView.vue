<template>
  <div class="inbox-view">
    <div class="inbox-header">
      <h1 class="inbox-title">{{ $t('inbox.title') }}</h1>
      <span class="inbox-count" v-if="tickets.length">{{ tickets.length }}</span>
      <button class="btn btn-primary btn-sm" style="margin-left:auto" @click="showCreate = true">+ {{ $t('ticket.new_ticket') }}</button>
    </div>

    <div v-if="loading" class="inbox-loading">{{ $t('common.loading') }}</div>

    <div v-else-if="!tickets.length" class="inbox-empty">
      <p>{{ $t('inbox.empty') }}</p>
    </div>

    <div v-else class="ticket-list">
      <RouterLink
        v-for="t in tickets"
        :key="t.id"
        :to="`/tickets/inbox/${t.id}`"
        class="ticket-card"
      >
        <div class="ticket-card-top">
          <span class="ticket-id">#{{ t.id }}</span>
          <span :class="['ticket-type-badge', `type-${t.type}`]">{{ $t('ticket.type_' + t.type) }}</span>
          <span :class="['ticket-priority', `priority-${t.priority}`]">{{ t.priority }}</span>
          <span :class="['ticket-status', `status-${t.status}`]">{{ t.status }}</span>
          <span v-if="t.from_email" class="ticket-email-badge" :title="t.from_email" aria-label="From email">✉</span>
        </div>
        <div class="ticket-card-title">{{ t.title }}</div>
        <div class="ticket-card-meta">
          <span v-if="t.assigned_to" class="ticket-assignee">{{ t.assigned_to.display_name || t.assigned_to.username }}</span>
          <span v-else class="ticket-unassigned">{{ $t('ticket.unassigned') }}</span>
          <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
        </div>
        <div v-if="t.tags?.length" class="ticket-tags">
          <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="ticket-tag">{{ tag.name }}</span>
          <span v-if="t.tags.length > 3" class="ticket-tag ticket-tag-more">+{{ t.tags.length - 3 }}</span>
        </div>
      </RouterLink>
    </div>

    <BaseModal v-if="showCreate" :title="$t('ticket.new_ticket')" @close="cancelCreate">
      <form @submit.prevent="submitCreate">
        <div class="form-group">
          <label for="inbox-title">{{ $t('ticket.title') }}</label>
          <input id="inbox-title" v-model="newTicket.title" class="form-input" required :placeholder="$t('ticket.title_placeholder')" />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label for="inbox-type">{{ $t('ticket.type') }}</label>
            <select id="inbox-type" v-model="newTicket.type" class="form-input">
              <option value="incident">{{ $t('ticket.type_incident') }}</option>
              <option value="problem">{{ $t('ticket.type_problem') }}</option>
              <option value="service_request">{{ $t('ticket.type_service_request') }}</option>
              <option value="change_request">{{ $t('ticket.type_change_request') }}</option>
            </select>
          </div>
          <div class="form-group">
            <label for="inbox-priority">{{ $t('ticket.priority') }}</label>
            <select id="inbox-priority" v-model="newTicket.priority" class="form-input">
              <option value="low">{{ $t('ticket.priority_low') }}</option>
              <option value="medium">{{ $t('ticket.priority_medium') }}</option>
              <option value="high">{{ $t('ticket.priority_high') }}</option>
              <option value="critical">{{ $t('ticket.priority_critical') }}</option>
            </select>
          </div>
        </div>
        <div class="form-group">
          <label for="inbox-description">{{ $t('ticket.description') }}</label>
          <textarea id="inbox-description" v-model="newTicket.description" class="form-input" rows="4" :placeholder="$t('ticket.description_placeholder')"></textarea>
        </div>
      </form>
      <template #footer>
        <button class="btn btn-secondary" @click="cancelCreate">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="submitCreate" :disabled="!newTicket.title.trim() || creating">{{ $t('common.create') }}</button>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ticketsApi } from '@/api/tickets'
import { useTicketsStore } from '@/stores/tickets'
import { useUIStore } from '@/stores/ui'
import { useDateFormat } from '@/composables/useDateFormat'
import { useI18n } from 'vue-i18n'
import BaseModal from '@/components/common/BaseModal.vue'

const { t } = useI18n()
const router = useRouter()
const ticketsStore = useTicketsStore()
const ui = useUIStore()
const { formatDate } = useDateFormat()
const tickets = ref([])
const loading = ref(false)
const showCreate = ref(false)
const creating = ref(false)

const defaultTicket = () => ({ title: '', description: '', type: 'service_request', priority: 'medium' })
const newTicket = ref(defaultTicket())

function cancelCreate() {
  showCreate.value = false
  newTicket.value = defaultTicket()
}

async function submitCreate() {
  if (!newTicket.value.title.trim()) return
  creating.value = true
  try {
    const { data } = await ticketsApi.inboxCreate(newTicket.value)
    tickets.value.unshift(data)
    ticketsStore.inboxCount++
    if (data.status === 'new') ticketsStore.inboxUnread++
    cancelCreate()
    router.push(`/tickets/inbox/${data.id}`)
  } catch (e) {
    ui.error(e.response?.data?.error || t('common.error'))
  } finally {
    creating.value = false
  }
}

onMounted(async () => {
  loading.value = true
  try {
    const { data } = await ticketsApi.inboxList()
    tickets.value = data
    const active = (data || []).filter(t => t.status !== 'closed')
    ticketsStore.inboxCount = active.length
    ticketsStore.inboxUnread = active.filter(t => t.status === 'new').length
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.inbox-view { padding: 24px; max-width: 900px; margin: 0 auto; }
.inbox-header { display: flex; align-items: center; gap: 10px; margin-bottom: 24px; }
.inbox-title { font-size: 1.4rem; font-weight: 600; margin: 0; }
.inbox-count { background: var(--color-primary); color: #fff; border-radius: 12px; font-size: 0.78rem; font-weight: 600; padding: 2px 8px; }
.inbox-loading, .inbox-empty { color: var(--color-text-muted); padding: 40px 0; text-align: center; }
.ticket-list { display: flex; flex-direction: column; gap: 8px; }
.ticket-card { display: block; background: var(--color-surface); border: 1px solid var(--color-border); border-radius: 8px; padding: 14px 16px; text-decoration: none; color: var(--color-text); transition: border-color .15s; }
.ticket-card:hover { border-color: var(--color-primary); }
.ticket-card-top { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.ticket-email-badge { font-size: 0.75rem; color: var(--color-text-muted); margin-left: auto; }
.ticket-id { font-size: 0.78rem; color: var(--color-text-muted); font-weight: 500; }
.ticket-card-title { font-weight: 500; margin-bottom: 8px; }
.ticket-card-meta { display: flex; align-items: center; gap: 12px; font-size: 0.8rem; color: var(--color-text-muted); margin-bottom: 6px; }
.ticket-card-meta .ticket-date { margin-left: auto; }
.ticket-tags { display: flex; gap: 4px; flex-wrap: wrap; }
.ticket-tag { font-size: 0.72rem; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 4px; padding: 1px 6px; color: var(--color-text-muted); }
.ticket-tag-more { color: var(--color-text-muted); }
.ticket-priority, .ticket-status, .ticket-type-badge { font-size: 0.72rem; border-radius: 4px; padding: 2px 6px; font-weight: 500; text-transform: capitalize; }
.priority-low    { background: #d4edda; color: #155724; }
.priority-medium { background: #fff3cd; color: #856404; }
.priority-high   { background: #ffe0b2; color: #e65100; }
.priority-critical { background: #f8d7da; color: #721c24; }
.status-new      { background: var(--color-bg); border: 1px solid var(--color-border); color: var(--color-text-muted); }
.status-open     { background: #cce5ff; color: #004085; }
.status-pending  { background: #fff3cd; color: #856404; }
.status-closed   { background: var(--color-bg); color: var(--color-text-muted); opacity: .7; }
.type-incident       { background: #fecaca; color: #b91c1c; }
.type-problem        { background: #fed7aa; color: #9a3412; }
.type-service_request { background: #d1fae5; color: #065f46; }
.type-change_request { background: #dbeafe; color: #1e40af; }
[data-theme="dark"] .priority-low     { background: #1a3a24; color: #6fcf97; }
[data-theme="dark"] .priority-medium  { background: #3a3010; color: #f2c94c; }
[data-theme="dark"] .priority-high    { background: #3a2010; color: #f2994a; }
[data-theme="dark"] .priority-critical{ background: #3a1015; color: #eb5757; }
[data-theme="dark"] .status-open      { background: #103060; color: #56b4f7; }
[data-theme="dark"] .status-pending   { background: #3a3010; color: #f2c94c; }
[data-theme="dark"] .type-incident    { background: #3a1015; color: #ef4444; }
[data-theme="dark"] .type-problem     { background: #3a2010; color: #f97316; }
[data-theme="dark"] .type-service_request { background: #0a2a1a; color: #34d399; }
[data-theme="dark"] .type-change_request  { background: #0a1a3a; color: #60a5fa; }
.form-row { display: flex; gap: 16px; }
.form-row .form-group { flex: 1; }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 13px; font-weight: 600; margin-bottom: 4px; }
</style>
