<template>
  <main class="inbox-main">
    <div class="inbox-container">
      <header class="inbox-header">
        <h1>{{ $t('inbox.title') }}</h1>
        <div class="header-actions">
          <div class="view-toggle" role="tablist">
            <button role="tab" :aria-selected="viewMode === 'cards'" :class="['view-toggle-btn', { active: viewMode === 'cards' }]" @click="viewMode = 'cards'">☰ {{ $t('ticket.card_view') }}</button>
            <button role="tab" :aria-selected="viewMode === 'group'" :class="['view-toggle-btn', { active: viewMode === 'group' }]" @click="viewMode = 'group'">⊞ {{ $t('ticket.group_view') }}</button>
            <button role="tab" :aria-selected="viewMode === 'list'" :class="['view-toggle-btn', { active: viewMode === 'list' }]" @click="viewMode = 'list'">☷ {{ $t('ticket.list_view') }}</button>
          </div>
          <button class="btn btn-primary btn-sm" @click="showCreate = true">+ {{ $t('ticket.new_ticket') }}</button>
        </div>
      </header>

      <div v-if="loading" class="loading-state">
        <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
      </div>

      <template v-else>
        <div v-if="!tickets.length" class="empty-state">{{ $t('inbox.empty') }}</div>

        <!-- Card view -->
        <template v-else-if="viewMode === 'cards'">
          <div v-if="regularTickets.length" class="ticket-grid">
            <div
              v-for="t in regularTickets" :key="t.id"
              class="ticket-card" :class="'ticket-' + t.status"
              @click="openTicket(t)" role="button" tabindex="0"
              @keydown.enter="openTicket(t)" @keydown.space.prevent="openTicket(t)"
              :aria-label="t.title"
            >
              <div class="ticket-card-header">
                <span class="ticket-id">#{{ t.id }}</span>
                <span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span>
                <span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span>
                <span v-if="t.from_email" class="email-badge" :title="t.from_email" aria-label="From email">✉</span>
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
                <span v-if="t.assigned_to" class="ticket-assignee">{{ t.assigned_to.display_name || t.assigned_to.username }}</span>
                <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
              </div>
            </div>
          </div>

          <div v-if="pendingReminderTickets.length" class="section-divider">
            <span>{{ $t('ticket.pending_reminders') }}</span>
          </div>
          <div v-if="pendingReminderTickets.length" class="ticket-grid">
            <div
              v-for="t in pendingReminderTickets" :key="t.id"
              class="ticket-card" :class="'ticket-' + t.status"
              @click="openTicket(t)" role="button" tabindex="0"
              @keydown.enter="openTicket(t)" @keydown.space.prevent="openTicket(t)"
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
                <span v-if="t.reminder_at" class="reminder-badge">{{ $t('ticket.reminder') }}: {{ formatDate(t.reminder_at) }}</span>
                <span v-if="t.tags?.length" class="ticket-tags">
                  <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                </span>
                <span v-if="t.assigned_to" class="ticket-assignee">{{ t.assigned_to.display_name || t.assigned_to.username }}</span>
                <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
              </div>
            </div>
          </div>

          <div v-if="pendingCloseTickets.length" class="section-divider">
            <span>{{ $t('ticket.resolved_closed') }}</span>
          </div>
          <div v-if="pendingCloseTickets.length" class="ticket-grid">
            <div
              v-for="t in pendingCloseTickets" :key="t.id"
              class="ticket-card" :class="'ticket-' + t.status"
              @click="openTicket(t)" role="button" tabindex="0"
              @keydown.enter="openTicket(t)" @keydown.space.prevent="openTicket(t)"
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
                <span v-if="t.tags?.length" class="ticket-tags">
                  <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                </span>
                <span v-if="t.assigned_to" class="ticket-assignee">{{ t.assigned_to.display_name || t.assigned_to.username }}</span>
                <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
              </div>
            </div>
          </div>

          <div v-if="closedTickets.length" class="section-divider">
            <span>{{ $t('ticket.status_closed') }}</span>
          </div>
          <div v-if="closedTickets.length" class="ticket-grid">
            <div
              v-for="t in closedTickets" :key="t.id"
              class="ticket-card" :class="'ticket-' + t.status"
              @click="openTicket(t)" role="button" tabindex="0"
              @keydown.enter="openTicket(t)" @keydown.space.prevent="openTicket(t)"
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
                <span v-if="t.tags?.length" class="ticket-tags">
                  <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                </span>
                <span v-if="t.assigned_to" class="ticket-assignee">{{ t.assigned_to.display_name || t.assigned_to.username }}</span>
                <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
              </div>
            </div>
          </div>
        </template>

        <!-- Group view -->
        <template v-else-if="viewMode === 'group'">
          <div class="group-sub-toggle">
            <button :class="['group-sub-btn', { active: groupSubMode === 'cards' }]" @click="groupSubMode = 'cards'">☰ {{ $t('ticket.card_view') }}</button>
            <button :class="['group-sub-btn', { active: groupSubMode === 'list' }]" @click="groupSubMode = 'list'">☷ {{ $t('ticket.list_view') }}</button>
          </div>
          <div v-for="g in groupedTickets" :key="g.status" class="group-section">
            <div class="group-header">
              <h2 class="group-title">{{ g.label }}</h2>
              <span class="group-count">{{ g.tickets.length }}</span>
            </div>
            <div v-if="g.tickets.length && groupSubMode === 'cards'" class="ticket-grid">
              <div
                v-for="t in g.tickets" :key="t.id"
                class="ticket-card" :class="'ticket-' + t.status"
                @click="openTicket(t)" role="button" tabindex="0"
                @keydown.enter="openTicket(t)" @keydown.space.prevent="openTicket(t)"
                :aria-label="t.title"
              >
                <div class="ticket-card-header">
                  <span class="ticket-id">#{{ t.id }}</span>
                  <span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span>
                  <span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span>
                  <span v-if="t.from_email" class="email-badge" :title="t.from_email" aria-label="From email">✉</span>
                </div>
                <h3 class="ticket-card-title">{{ t.title }}</h3>
                <div class="ticket-card-meta">
                  <span v-if="t.status === 'pending' && t.reminder_at" class="reminder-badge">{{ $t('ticket.reminder') }}: {{ formatDate(t.reminder_at) }}</span>
                  <span v-if="t.sla_response_breached" class="sla-badge sla-breach" :title="slaTitle(t)">{{ $t('sla.breached') }}</span>
                  <span v-else-if="slaWarning(t)" class="sla-badge sla-warning" :title="slaTitle(t)">{{ $t('sla.warning') }}</span>
                  <span v-else-if="t.sla_policy_id" class="sla-badge sla-ok" :title="slaTitle(t)">{{ $t('sla.on_track') }}</span>
                  <span v-if="t.tags?.length" class="ticket-tags">
                    <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                    <span v-if="t.tags.length > 3" class="mini-tag more">+{{ t.tags.length - 3 }}</span>
                  </span>
                  <span v-if="t.assigned_to" class="ticket-assignee">{{ t.assigned_to.display_name || t.assigned_to.username }}</span>
                  <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
                </div>
              </div>
            </div>
            <table v-else-if="g.tickets.length && groupSubMode === 'list'" class="group-table">
              <thead>
                <tr>
                  <th class="th-sort" :class="groupSortClass('id')" @click="groupToggleSort('id')"># <span class="sort-arrow">{{ groupSortArrow('id') }}</span></th>
                  <th class="th-sort" :class="groupSortClass('title')" @click="groupToggleSort('title')">{{ $t('ticket.title') }} <span class="sort-arrow">{{ groupSortArrow('title') }}</span></th>
                  <th class="th-sort" :class="groupSortClass('priority')" @click="groupToggleSort('priority')">{{ $t('ticket.priority') }} <span class="sort-arrow">{{ groupSortArrow('priority') }}</span></th>
                  <th class="th-sort" :class="groupSortClass('type')" @click="groupToggleSort('type')">{{ $t('ticket.type') }} <span class="sort-arrow">{{ groupSortArrow('type') }}</span></th>
                  <th class="th-sort" :class="groupSortClass('assigned_to')" @click="groupToggleSort('assigned_to')">{{ $t('ticket.assigned_to') }} <span class="sort-arrow">{{ groupSortArrow('assigned_to') }}</span></th>
                  <th class="th-sort" :class="groupSortClass('created_at')" @click="groupToggleSort('created_at')">{{ $t('ticket.created_at') }} <span class="sort-arrow">{{ groupSortArrow('created_at') }}</span></th>
                  <th>{{ $t('ticket.tags') }}</th>
                  <th>SLA</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="t in g.tickets" :key="t.id" class="table-row" @click="openTicket(t)" tabindex="0" @keydown.enter="openTicket(t)">
                  <td class="td-id">#{{ t.id }}</td>
                  <td class="td-title">{{ t.title }}</td>
                  <td><span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span></td>
                  <td><span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span></td>
                  <td class="td-assignee">{{ t.assigned_to?.display_name || t.assigned_to?.username || '—' }}</td>
                  <td class="td-date">{{ formatDate(t.created_at) }}</td>
                  <td>
                    <span v-if="t.tags?.length" class="ticket-tags">
                      <span v-for="tag in t.tags.slice(0, 2)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                      <span v-if="t.tags.length > 2" class="mini-tag more">+{{ t.tags.length - 2 }}</span>
                    </span>
                  </td>
                  <td>
                    <span v-if="t.sla_response_breached" class="sla-badge sla-breach">{{ $t('sla.breached') }}</span>
                    <span v-else-if="slaWarning(t)" class="sla-badge sla-warning">{{ $t('sla.warning') }}</span>
                    <span v-else-if="t.sla_policy_id" class="sla-badge sla-ok">{{ $t('sla.on_track') }}</span>
                    <span v-else>—</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>

        <!-- List view -->
        <template v-else-if="viewMode === 'list'">
          <table class="ticket-table">
            <thead>
              <tr>
                <th class="th-sort" :class="sortClass('id')" @click="toggleSort('id')"># <span class="sort-arrow">{{ sortArrow('id') }}</span></th>
                <th class="th-sort th-title" :class="sortClass('title')" @click="toggleSort('title')">{{ $t('ticket.title') }} <span class="sort-arrow">{{ sortArrow('title') }}</span></th>
                <th class="th-sort" :class="sortClass('status')" @click="toggleSort('status')">{{ $t('ticket.status') }} <span class="sort-arrow">{{ sortArrow('status') }}</span></th>
                <th class="th-sort" :class="sortClass('priority')" @click="toggleSort('priority')">{{ $t('ticket.priority') }} <span class="sort-arrow">{{ sortArrow('priority') }}</span></th>
                <th class="th-sort" :class="sortClass('type')" @click="toggleSort('type')">{{ $t('ticket.type') }} <span class="sort-arrow">{{ sortArrow('type') }}</span></th>
                <th class="th-sort" :class="sortClass('assigned_to')" @click="toggleSort('assigned_to')">{{ $t('ticket.assigned_to') }} <span class="sort-arrow">{{ sortArrow('assigned_to') }}</span></th>
                <th class="th-sort" :class="sortClass('created_at')" @click="toggleSort('created_at')">{{ $t('ticket.created_at') }} <span class="sort-arrow">{{ sortArrow('created_at') }}</span></th>
                <th class="th-tags">{{ $t('ticket.tags') }}</th>
                <th>SLA</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="t in sortedTickets" :key="t.id" class="table-row" @click="openTicket(t)" tabindex="0" @keydown.enter="openTicket(t)" :aria-label="t.title">
                <td class="td-id">#{{ t.id }}</td>
                <td class="td-title">{{ t.title }}</td>
                <td><span class="ticket-status" :class="'status-' + t.status">{{ $t('ticket.status_' + t.status) }}</span></td>
                <td><span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span></td>
                <td><span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span></td>
                <td class="td-assignee">{{ t.assigned_to?.display_name || t.assigned_to?.username || '—' }}</td>
                <td class="td-date">{{ formatDate(t.created_at) }}</td>
                <td>
                  <span v-if="t.tags?.length" class="ticket-tags">
                    <span v-for="tag in t.tags.slice(0, 2)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                    <span v-if="t.tags.length > 2" class="mini-tag more">+{{ t.tags.length - 2 }}</span>
                  </span>
                </td>
                <td>
                  <span v-if="t.sla_response_breached" class="sla-badge sla-breach">{{ $t('sla.breached') }}</span>
                  <span v-else-if="slaWarning(t)" class="sla-badge sla-warning">{{ $t('sla.warning') }}</span>
                  <span v-else-if="t.sla_policy_id" class="sla-badge sla-ok">{{ $t('sla.on_track') }}</span>
                  <span v-else>—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </template>
      </template>
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
  </main>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
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

const viewMode = ref(localStorage.getItem('inbox_view_mode') || 'cards')
watch(viewMode, val => localStorage.setItem('inbox_view_mode', val))

const groupSubMode = ref(localStorage.getItem('inbox_group_sub_mode') || 'cards')
watch(groupSubMode, val => localStorage.setItem('inbox_group_sub_mode', val))

const sortField = ref('created_at')
const sortDir = ref(-1)
const groupSortField = ref('created_at')
const groupSortDir = ref(-1)

const priorityRank = { low: 1, medium: 2, high: 3, critical: 4 }

const regularTickets = computed(() =>
  tickets.value.filter(t => t.status !== 'pending_close' && t.status !== 'closed' && !(t.status === 'pending' && t.reminder_at))
)
const pendingReminderTickets = computed(() =>
  tickets.value.filter(t => t.status === 'pending' && t.reminder_at)
    .sort((a, b) => new Date(a.reminder_at) - new Date(b.reminder_at))
)
const pendingCloseTickets = computed(() =>
  tickets.value.filter(t => t.status === 'pending_close')
)
const closedTickets = computed(() =>
  tickets.value.filter(t => t.status === 'closed')
)

const statusGroups = [
  { status: 'new', label: 'New' },
  { status: 'open', label: 'Open' },
  { status: 'pending', label: 'Pending reminder' },
  { status: 'pending_close', label: 'Pending close' },
  { status: 'closed', label: 'Closed' },
]

function sortTickets(arr, field, dir) {
  return [...arr].sort((a, b) => {
    let va, vb
    if (field === 'priority') { va = priorityRank[a.priority] || 0; vb = priorityRank[b.priority] || 0 }
    else if (field === 'assigned_to') { va = (a.assigned_to?.display_name || a.assigned_to?.username || '').toLowerCase(); vb = (b.assigned_to?.display_name || b.assigned_to?.username || '').toLowerCase() }
    else if (field === 'id') { va = a.id; vb = b.id }
    else if (field === 'title') { va = a.title.toLowerCase(); vb = b.title.toLowerCase() }
    else { va = a[field]; vb = b[field] }
    if (va < vb) return -1 * dir
    if (va > vb) return 1 * dir
    return 0
  })
}

const groupedTickets = computed(() =>
  statusGroups
    .map(g => ({ ...g, tickets: sortTickets(tickets.value.filter(t => t.status === g.status), groupSortField.value, groupSortDir.value) }))
    .filter(g => g.tickets.length > 0)
)

const sortedTickets = computed(() => sortTickets(tickets.value, sortField.value, sortDir.value))

function toggleSort(field) {
  if (sortField.value === field) sortDir.value *= -1
  else { sortField.value = field; sortDir.value = -1 }
}
function sortClass(field) { return sortField.value !== field ? '' : sortDir.value === -1 ? 'sort-desc' : 'sort-asc' }
function sortArrow(field) { return sortField.value !== field ? '▽' : sortDir.value === -1 ? '▽' : '△' }

function groupToggleSort(field) {
  if (groupSortField.value === field) groupSortDir.value *= -1
  else { groupSortField.value = field; groupSortDir.value = -1 }
}
function groupSortClass(field) { return groupSortField.value !== field ? '' : groupSortDir.value === -1 ? 'sort-desc' : 'sort-asc' }
function groupSortArrow(field) { return groupSortField.value !== field ? '▽' : groupSortDir.value === -1 ? '▽' : '△' }

function slaWarning(t) {
  if (!t.sla_response_deadline && !t.sla_resolution_deadline) return false
  if (t.status === 'pending_close') return false
  const now = Date.now()
  if (t.sla_response_deadline && !t.first_response_at) {
    const dl = new Date(t.sla_response_deadline).getTime()
    if (dl - now < 3600000 && dl - now > 0) return true
  }
  if (t.sla_resolution_deadline) {
    const dl = new Date(t.sla_resolution_deadline).getTime()
    if (dl - now < 3600000 && dl - now > 0) return true
  }
  return false
}
function slaTitle(t) {
  const parts = []
  if (t.sla_policy?.name) parts.push(t.sla_policy.name)
  if (t.sla_response_deadline) parts.push(`Response: ${formatDate(t.sla_response_deadline)}`)
  if (t.sla_resolution_deadline) parts.push(`Resolution: ${formatDate(t.sla_resolution_deadline)}`)
  return parts.join(' | ')
}

function openTicket(t) {
  router.push(`/tickets/inbox/${t.id}`)
}

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

async function loadInbox() {
  loading.value = true
  try {
    const { data } = await ticketsApi.inboxList()
    tickets.value = data || []
    const active = (data || []).filter(t => t.status !== 'closed')
    ticketsStore.inboxCount = active.length
    ticketsStore.inboxUnread = active.filter(t => t.status === 'new').length
  } finally {
    loading.value = false
  }
}

onMounted(loadInbox)
watch(() => ticketsStore.inboxRefreshKey, loadInbox)
</script>

<style scoped>
.inbox-main { padding: 24px; margin: 0 auto; }
.inbox-main:has(.ticket-table) { max-width: 100%; padding: 24px 32px; }
.inbox-main:not(:has(.ticket-table)) { max-width: 1200px; }
.inbox-container { }
.inbox-header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.inbox-header h1 { flex: 1; margin: 0; font-size: 22px; }
.header-actions { display: flex; align-items: center; gap: 8px; }
.view-toggle { display: flex; border: 1px solid var(--color-border); border-radius: 6px; overflow: hidden; }
.view-toggle-btn { background: none; border: none; padding: 5px 12px; font-size: 12px; cursor: pointer; color: var(--color-text-muted); transition: background .15s, color .15s; }
.view-toggle-btn:not(:last-child) { border-right: 1px solid var(--color-border); }
.view-toggle-btn:hover { background: var(--color-bg-alt); }
.view-toggle-btn.active { background: var(--color-primary); color: #fff; }
.loading-state { display: flex; justify-content: center; padding: 48px; }
.empty-state { text-align: center; padding: 64px 24px; color: var(--color-text-muted); }
.ticket-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 12px; }
.ticket-card { background: var(--color-surface); border: 1px solid var(--color-border); border-radius: 8px; padding: 16px; cursor: pointer; transition: box-shadow .15s, border-color .15s; }
.ticket-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,.08); border-color: var(--color-primary); }
.ticket-card:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }
.ticket-card-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.ticket-id { font-size: 11px; font-weight: 700; color: var(--color-text-muted); }
.email-badge { font-size: 0.75rem; color: var(--color-text-muted); margin-left: auto; }
.ticket-type { font-size: 10px; font-weight: 600; padding: 2px 6px; border-radius: 4px; }
.type-incident { background: #fecaca; color: #b91c1c; }
.type-problem { background: #fed7aa; color: #9a3412; }
.type-service_request { background: #d1fae5; color: #065f46; }
.type-change_request { background: #dbeafe; color: #1e40af; }
.ticket-priority { font-size: 10px; font-weight: 600; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; }
.pri-low { background: #e0f2fe; color: #0369a1; }
.pri-medium { background: #fef3c7; color: #92400e; }
.pri-high { background: #fde68a; color: #b45309; }
.pri-critical { background: #fecaca; color: #b91c1c; }
.ticket-card-title { font-size: 14px; font-weight: 600; margin: 0 0 8px; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.ticket-card-meta { display: flex; gap: 8px; align-items: center; font-size: 11px; color: var(--color-text-muted); flex-wrap: wrap; }
.ticket-status { padding: 2px 6px; border-radius: 4px; font-weight: 600; }
.status-new { background: #dbeafe; color: #1e40af; }
.status-open { background: #fef3c7; color: #92400e; }
.status-pending { background: #f0e6ff; color: #6b21a8; }
.status-pending_close { background: #d1fae5; color: #065f46; }
.status-closed { background: #e5e7eb; color: #374151; }
.ticket-assignee { display: flex; align-items: center; gap: 4px; }
.ticket-tags { display: flex; gap: 3px; align-items: center; }
.mini-tag { font-size: 10px; background: var(--color-bg-alt); padding: 1px 5px; border-radius: 3px; color: var(--color-text-muted); }
.mini-tag.more { font-weight: 700; }
.sla-badge { font-size: 10px; font-weight: 700; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; }
.sla-ok { background: #d1fae5; color: #065f46; }
.sla-warning { background: #fef3c7; color: #92400e; }
.sla-breach { background: #fecaca; color: #b91c1c; }
.reminder-badge { font-size: 10px; font-weight: 700; padding: 1px 5px; border-radius: 3px; background: #fef3c7; color: #92400e; }
.section-divider { display: flex; align-items: center; gap: 12px; margin: 24px 0 16px; color: var(--color-text-muted); font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; }
.section-divider::before, .section-divider::after { content: ''; flex: 1; height: 1px; background: var(--color-border); }
/* Group view */
.group-section { margin-bottom: 32px; }
.group-header { display: flex; align-items: center; gap: 10px; margin-bottom: 16px; }
.group-title { font-size: 16px; font-weight: 700; margin: 0; }
.group-count { font-size: 11px; font-weight: 600; color: #fff; background: var(--color-primary); padding: 1px 7px; border-radius: 9999px; line-height: 20px; }
.group-section:not(:last-child)::after { content: ''; display: block; height: 1px; background: var(--color-border); margin-top: 32px; }
.group-sub-toggle { display: flex; gap: 4px; margin-bottom: 20px; }
.group-sub-btn { font-size: 12px; font-weight: 600; padding: 4px 12px; border: 1px solid var(--color-border); border-radius: 6px; background: var(--color-bg); color: var(--color-text-muted); cursor: pointer; transition: all .15s; }
.group-sub-btn.active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.group-sub-btn:hover:not(.active) { border-color: var(--color-primary); color: var(--color-primary); }
.group-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.group-table th { text-align: left; padding: 6px 8px; border-bottom: 2px solid var(--color-border); font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; color: var(--color-text-muted); white-space: nowrap; }
.group-table td { padding: 6px 8px; border-bottom: 1px solid var(--color-border); vertical-align: middle; }
.group-table .table-row { cursor: pointer; }
.group-table .table-row:hover { background: var(--color-bg-alt); }
/* List view */
.ticket-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.ticket-table th { text-align: left; padding: 8px 10px; border-bottom: 2px solid var(--color-border); font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; color: var(--color-text-muted); white-space: nowrap; user-select: none; }
.ticket-table td { padding: 8px 10px; border-bottom: 1px solid var(--color-border); vertical-align: middle; }
.th-sort { cursor: pointer; }
.th-sort:hover { color: var(--color-text); }
.sort-arrow { font-size: 10px; margin-left: 2px; }
.th-sort.sort-desc .sort-arrow, .th-sort.sort-asc .sort-arrow { color: var(--color-primary); }
.th-title { min-width: 200px; }
.td-title { font-weight: 600; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.td-id { font-family: monospace; font-size: 12px; color: var(--color-text-muted); font-weight: 700; }
.td-assignee { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 120px; color: var(--color-text-muted); }
.td-date { white-space: nowrap; color: var(--color-text-muted); font-size: 12px; }
.table-row { cursor: pointer; transition: background .1s; }
.table-row:hover { background: var(--color-bg-alt); }
.table-row:focus-visible { outline: 2px solid var(--color-primary); outline-offset: -2px; }
/* Dark mode */
[data-theme="dark"] .type-incident { background: #3a1015; color: #ef4444; }
[data-theme="dark"] .type-problem { background: #3a2010; color: #f97316; }
[data-theme="dark"] .type-service_request { background: #0a2a1a; color: #34d399; }
[data-theme="dark"] .type-change_request { background: #0a1a3a; color: #60a5fa; }
[data-theme="dark"] .pri-low { background: #1a3a24; color: #6fcf97; }
[data-theme="dark"] .pri-medium { background: #3a3010; color: #f2c94c; }
[data-theme="dark"] .pri-high { background: #3a2010; color: #f2994a; }
[data-theme="dark"] .pri-critical { background: #3a1015; color: #eb5757; }
[data-theme="dark"] .status-new { background: #0a1a3a; color: #60a5fa; }
[data-theme="dark"] .status-open { background: #3a3010; color: #f2c94c; }
[data-theme="dark"] .status-pending { background: #2a1040; color: #c084fc; }
[data-theme="dark"] .status-pending_close { background: #0a2a1a; color: #34d399; }
[data-theme="dark"] .status-closed { background: #1f2937; color: #9ca3af; }
/* Create modal */
.form-row { display: flex; gap: 16px; }
.form-row .form-group { flex: 1; }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 13px; font-weight: 600; margin-bottom: 4px; }
</style>
