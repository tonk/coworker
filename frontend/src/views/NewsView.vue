<template>
  <main class="news-main">
    <div class="news-container">
      <div class="news-header">
        <h1>{{ $t('news.page_title') }}</h1>
        <div class="news-controls">
          <!-- Active / All toggle -->
          <div class="segment-control" role="group" :aria-label="$t('news.filter_label')">
            <button
              :class="['segment-btn', { active: !showAll }]"
              @click="showAll = false"
            >{{ $t('news.filter_active') }}</button>
            <button
              :class="['segment-btn', { active: showAll }]"
              @click="showAll = true"
            >{{ $t('news.filter_all') }}</button>
          </div>

          <!-- Sort -->
          <div class="sort-group">
            <select class="form-input sort-select" v-model="sortField" :aria-label="$t('news.sort_label')">
              <option value="created_at">{{ $t('news.sort_created') }}</option>
              <option value="start_date">{{ $t('news.sort_start') }}</option>
              <option value="end_date">{{ $t('news.sort_end') }}</option>
              <option value="title">{{ $t('news.sort_title') }}</option>
            </select>
            <button
              class="btn btn-ghost btn-sm sort-dir-btn"
              :title="sortDir === 'desc' ? $t('news.sort_desc') : $t('news.sort_asc')"
              @click="sortDir = sortDir === 'desc' ? 'asc' : 'desc'"
            >
              {{ sortDir === 'asc' ? '△' : '▽' }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="loading" class="loading-state">
        <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
      </div>

      <div v-else-if="!sortedItems.length" class="empty-state">
        {{ $t('news.empty') }}
      </div>

      <div v-else class="news-grid">
        <div
          v-for="item in sortedItems"
          :key="item.id"
          class="news-tile"
          :class="{ 'news-tile-inactive': !item.active || isExpired(item) || isFuture(item) }"
          :style="item.sidebar_color ? { borderLeftColor: item.sidebar_color } : {}"
        >
          <div class="news-tile-header">
            <img src="/logo.svg" alt="WarmDesk" class="news-tile-logo" />
            <span class="news-tile-tag">{{ $t('dashboard.news_title') }}</span>
            <div class="news-tile-badges">
              <span v-if="!item.active" class="news-badge badge-inactive">{{ $t('news.badge_inactive') }}</span>
              <span v-else-if="isExpired(item)" class="news-badge badge-expired">{{ $t('news.badge_expired') }}</span>
              <span v-else-if="isFuture(item)" class="news-badge badge-future">{{ $t('news.badge_future') }}</span>
              <span v-if="isDismissed(item.id)" class="news-badge badge-read">{{ $t('news.badge_read') }}</span>
            </div>
          </div>
          <h2 class="news-tile-title">{{ item.title }}</h2>
          <div class="news-tile-meta">
            <span class="news-tile-date">{{ formatDate(item.created_at) }}</span>
            <span v-if="item.start_date" class="news-tile-date">{{ $t('admin.news_start_date') }}: {{ formatDate(item.start_date) }}</span>
            <span v-if="item.end_date" class="news-tile-date">{{ $t('admin.news_end_date') }}: {{ formatDate(item.end_date) }}</span>
          </div>
          <div class="news-tile-body markdown-body" v-html="renderMarkdown(item.text)"></div>
          <div class="news-tile-footer">
            <button
              class="btn btn-ghost btn-sm read-toggle-btn"
              :class="{ 'read-toggle-btn--read': isDismissed(item.id) }"
              @click="isDismissed(item.id) ? undismiss(item.id) : dismiss(item.id)"
            >
              <svg v-if="isDismissed(item.id)" viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                <path d="M1 8s2.5-5 7-5 7 5 7 5-2.5 5-7 5-7-5-7-5z"/>
                <circle cx="8" cy="8" r="2"/>
              </svg>
              <svg v-else viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                <path d="M13.4 13.4L2.6 2.6M6.3 6.4A2 2 0 0 0 9.6 9.7M4.3 4.4C2.6 5.5 1.3 7.1 1 8c.8 2.3 3.7 5 7 5 1.3 0 2.5-.4 3.5-1M8 3c.4 0 .8 0 1.2.1C11.7 3.7 14 5.8 15 8c-.4 1-.9 2-1.7 2.8"/>
              </svg>
              {{ isDismissed(item.id) ? $t('news.mark_unread') : $t('news.mark_read') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { newsApi } from '@/api/news'
import { useDateFormat } from '@/composables/useDateFormat'

const { t } = useI18n()
const { formatDate } = useDateFormat()

const NEWS_DISMISSED_KEY = 'dashboard_news_dismissed_ids'

const loading = ref(true)
const allItems = ref([])
const showAll = ref(false)
const sortField = ref('created_at')
const sortDir = ref('desc')

function getDismissedIds() {
  try { return new Set(JSON.parse(localStorage.getItem(NEWS_DISMISSED_KEY) || '[]')) } catch { return new Set() }
}
const dismissedIds = ref(getDismissedIds())

function isDismissed(id) { return dismissedIds.value.has(id) }

function dismiss(id) {
  dismissedIds.value = new Set([...dismissedIds.value, id])
  try { localStorage.setItem(NEWS_DISMISSED_KEY, JSON.stringify([...dismissedIds.value])) } catch {}
}

function undismiss(id) {
  dismissedIds.value = new Set([...dismissedIds.value].filter(x => x !== id))
  try { localStorage.setItem(NEWS_DISMISSED_KEY, JSON.stringify([...dismissedIds.value])) } catch {}
}

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}

function isExpired(item) {
  return item.end_date && new Date(item.end_date) < new Date()
}

function isFuture(item) {
  return item.start_date && new Date(item.start_date) > new Date()
}

const filteredItems = computed(() => {
  if (showAll.value) return allItems.value
  const now = new Date()
  return allItems.value.filter(n =>
    n.active &&
    (!n.start_date || new Date(n.start_date) <= now) &&
    (!n.end_date || new Date(n.end_date) >= now)
  )
})

const sortedItems = computed(() => {
  const items = [...filteredItems.value]
  const dir = sortDir.value === 'asc' ? 1 : -1
  items.sort((a, b) => {
    let va = a[sortField.value]
    let vb = b[sortField.value]
    if (sortField.value === 'title') {
      return dir * (va || '').localeCompare(vb || '')
    }
    if (!va && !vb) return 0
    if (!va) return dir
    if (!vb) return -dir
    return dir * (new Date(va) - new Date(vb))
  })
  return items
})

onMounted(async () => {
  try {
    const r = await newsApi.listActive({ all: true })
    allItems.value = r.data || []
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.news-main {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}
.news-container {
  max-width: 900px;
  margin: 0 auto;
}
.news-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 24px;
}
.news-header h1 {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
}
.news-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

/* Segment control */
.segment-control {
  display: flex;
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
}
.segment-btn {
  padding: 5px 14px;
  font-size: 13px;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  transition: background .15s, color .15s;
}
.segment-btn.active {
  background: var(--color-primary);
  color: #fff;
}
.segment-btn:not(.active):hover {
  background: var(--color-surface);
  color: var(--color-text);
}

/* Sort controls */
.sort-group {
  display: flex;
  align-items: center;
  gap: 4px;
}
.sort-select {
  padding: 4px 8px;
  font-size: 13px;
  height: auto;
}
.sort-dir-btn {
  padding: 4px 6px;
  display: flex;
  align-items: center;
}

/* Grid */
.news-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Tile */
.news-tile {
  position: relative;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-left: 4px solid var(--color-primary);
  border-radius: var(--radius);
  padding: 18px 22px;
}
.news-tile-inactive {
  opacity: 0.65;
}
.news-tile-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}
.news-tile-logo {
  height: 18px;
  width: auto;
  display: block;
  opacity: 0.85;
}
.news-tile-tag {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .07em;
  color: var(--color-primary);
}
.news-tile-badges {
  display: flex;
  gap: 6px;
  margin-left: auto;
}
.news-badge {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 7px;
  border-radius: 99px;
  text-transform: uppercase;
  letter-spacing: .04em;
}
.badge-inactive { background: var(--color-bg-alt); color: var(--color-text-muted); border: 1px solid var(--color-border); }
.badge-expired  { background: #fef3c7; color: #92400e; border: 1px solid #fde68a; }
.badge-future   { background: #ede9fe; color: #5b21b6; border: 1px solid #ddd6fe; }
.badge-read     { background: var(--color-bg-alt); color: var(--color-text-muted); border: 1px solid var(--color-border); }

.news-tile-title {
  font-size: 16px;
  font-weight: 700;
  margin: 0 0 6px;
  line-height: 1.3;
}
.news-tile-meta {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}
.news-tile-date {
  font-size: 11px;
  color: var(--color-text-muted);
}
.news-tile-body {
  font-size: 13px;
  color: var(--color-text-muted);
  line-height: 1.6;
}
.news-tile-body :deep(p)   { margin: 0 0 6px; }
.news-tile-body :deep(h1),.news-tile-body :deep(h2),.news-tile-body :deep(h3) { font-weight: 700; margin: 8px 0 4px; color: var(--color-text); }
.news-tile-body :deep(h1) { font-size: 1.05em; }
.news-tile-body :deep(h2) { font-size: 1em; }
.news-tile-body :deep(ul),.news-tile-body :deep(ol) { padding-left: 18px; margin: 0 0 6px; }
.news-tile-body :deep(li) { margin: 2px 0; }
.news-tile-body :deep(code) { background: var(--color-bg-alt); border: 1px solid var(--color-border); border-radius: 3px; padding: 1px 4px; font-size: .85em; font-family: var(--font-mono); }
.news-tile-body :deep(pre) { background: var(--color-bg-alt); border: 1px solid var(--color-border); border-radius: var(--radius); padding: 8px 10px; overflow-x: auto; margin: 0 0 6px; font-size: .85em; }

.news-tile-footer {
  margin-top: 12px;
  min-height: 24px;
}

.loading-state { display: flex; justify-content: center; padding: 60px; }
.empty-state { text-align: center; padding: 60px; color: var(--color-text-muted); }

.read-toggle-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: var(--color-text-muted);
  padding: 3px 9px;
}
.read-toggle-btn:hover { color: var(--color-text); }
.read-toggle-btn--read { color: var(--color-primary); }
.read-toggle-btn--read:hover { color: var(--color-primary); opacity: 0.75; }
</style>
