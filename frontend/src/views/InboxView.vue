<template>
  <div class="inbox-view">
    <div class="inbox-header">
      <h1 class="inbox-title">{{ $t('inbox.title') }}</h1>
      <span class="inbox-count" v-if="tickets.length">{{ tickets.length }}</span>
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
          <span :class="['ticket-priority', `priority-${t.priority}`]">{{ t.priority }}</span>
          <span :class="['ticket-status', `status-${t.status}`]">{{ t.status }}</span>
        </div>
        <div class="ticket-card-title">{{ t.title }}</div>
        <div class="ticket-card-meta">
          <span v-if="t.assigned_to" class="ticket-assignee">{{ t.assigned_to.display_name || t.assigned_to.username }}</span>
          <span v-else class="ticket-unassigned">{{ $t('ticket.unassigned') }}</span>
          <span class="ticket-date">{{ fmtDate(t.created_at) }}</span>
        </div>
        <div v-if="t.tags?.length" class="ticket-tags">
          <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="ticket-tag">{{ tag.name }}</span>
          <span v-if="t.tags.length > 3" class="ticket-tag ticket-tag-more">+{{ t.tags.length - 3 }}</span>
        </div>
      </RouterLink>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ticketsApi } from '@/api/tickets'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const tickets = ref([])
const loading = ref(false)

function fmtDate(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString(undefined, { day: 'numeric', month: 'short', year: 'numeric' })
}

onMounted(async () => {
  loading.value = true
  try {
    const { data } = await ticketsApi.inboxList()
    tickets.value = data
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
.ticket-id { font-size: 0.78rem; color: var(--color-text-muted); font-weight: 500; }
.ticket-card-title { font-weight: 500; margin-bottom: 8px; }
.ticket-card-meta { display: flex; align-items: center; gap: 12px; font-size: 0.8rem; color: var(--color-text-muted); margin-bottom: 6px; }
.ticket-card-meta .ticket-date { margin-left: auto; }
.ticket-tags { display: flex; gap: 4px; flex-wrap: wrap; }
.ticket-tag { font-size: 0.72rem; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 4px; padding: 1px 6px; color: var(--color-text-muted); }
.ticket-tag-more { color: var(--color-text-muted); }
.ticket-priority, .ticket-status { font-size: 0.72rem; border-radius: 4px; padding: 2px 6px; font-weight: 500; text-transform: capitalize; }
.priority-low    { background: #d4edda; color: #155724; }
.priority-medium { background: #fff3cd; color: #856404; }
.priority-high   { background: #ffe0b2; color: #e65100; }
.priority-critical { background: #f8d7da; color: #721c24; }
.status-new      { background: var(--color-bg); border: 1px solid var(--color-border); color: var(--color-text-muted); }
.status-open     { background: #cce5ff; color: #004085; }
.status-pending  { background: #fff3cd; color: #856404; }
.status-closed   { background: var(--color-bg); color: var(--color-text-muted); opacity: .7; }
[data-theme="dark"] .priority-low     { background: #1a3a24; color: #6fcf97; }
[data-theme="dark"] .priority-medium  { background: #3a3010; color: #f2c94c; }
[data-theme="dark"] .priority-high    { background: #3a2010; color: #f2994a; }
[data-theme="dark"] .priority-critical{ background: #3a1015; color: #eb5757; }
[data-theme="dark"] .status-open      { background: #103060; color: #56b4f7; }
[data-theme="dark"] .status-pending   { background: #3a3010; color: #f2c94c; }
</style>
