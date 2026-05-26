<template>
  <div v-if="visibleNews.length" class="dashboard-widgets">
    <div
      v-for="item in visibleNews"
      :key="item.id"
      class="widget news-widget"
      :style="item.sidebar_color ? { borderLeftColor: item.sidebar_color } : {}"
    >
      <button class="widget-dismiss" :aria-label="$t('common.close')" @click="dismiss(item.id)">×</button>
      <div class="widget-header">
        <img src="/logo.svg" alt="WarmDesk" class="widget-logo" />
        <span class="widget-tag">{{ $t('dashboard.news_title') }}</span>
      </div>
      <h2 class="widget-title">{{ item.title }}</h2>
      <p class="widget-date">{{ formatDate(item.created_at) }}</p>
      <div class="widget-body markdown-body" v-html="renderMarkdown(item.text)"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { newsApi } from '@/api/news'
import { useDateFormat } from '@/composables/useDateFormat'

const { formatDate } = useDateFormat()

const STORAGE_KEY = 'dashboard_news_dismissed_ids'

function getDismissedIds() {
  try { return new Set(JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]')) } catch { return new Set() }
}

const allNews = ref([])
const dismissedIds = ref(getDismissedIds())

const visibleNews = computed(() => allNews.value.filter(n => !dismissedIds.value.has(n.id)))

function dismiss(id) {
  dismissedIds.value = new Set([...dismissedIds.value, id])
  try { localStorage.setItem(STORAGE_KEY, JSON.stringify([...dismissedIds.value])) } catch {}
}

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}

onMounted(() => {
  newsApi.listActive().then(r => {
    const fetched = r.data || []
    allNews.value = fetched
    // prune stale dismissed IDs
    const activeIds = new Set(fetched.map(n => n.id))
    const pruned = new Set([...dismissedIds.value].filter(id => activeIds.has(id)))
    if (pruned.size !== dismissedIds.value.size) {
      dismissedIds.value = pruned
      try { localStorage.setItem(STORAGE_KEY, JSON.stringify([...pruned])) } catch {}
    }
  }).catch(() => {})
})
</script>

<style scoped>
.dashboard-widgets {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  margin-bottom: 28px;
}

.widget {
  position: relative;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-left: 3px solid var(--color-primary);
  border-radius: var(--radius);
  padding: 18px 20px;
}

.widget-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.widget-logo {
  height: 18px;
  width: auto;
  display: block;
  opacity: 0.85;
}

.widget-tag {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .07em;
  color: var(--color-primary);
}

.widget-title {
  font-size: 15px;
  font-weight: 700;
  margin-bottom: 4px;
  line-height: 1.3;
  padding-right: 24px;
}

.widget-date {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-bottom: 10px;
}

.widget-body {
  font-size: 13px;
  color: var(--color-text-muted);
  line-height: 1.55;
}
.widget-body :deep(p)   { margin: 0 0 6px; }
.widget-body :deep(h1),.widget-body :deep(h2),.widget-body :deep(h3) { font-weight: 700; margin: 8px 0 4px; color: var(--color-text); }
.widget-body :deep(h1) { font-size: 1.05em; }
.widget-body :deep(h2) { font-size: 1em; }
.widget-body :deep(ul),.widget-body :deep(ol) { padding-left: 18px; margin: 0 0 6px; }
.widget-body :deep(li) { margin: 2px 0; }
.widget-body :deep(code) { background: var(--color-bg-alt); border: 1px solid var(--color-border); border-radius: 3px; padding: 1px 4px; font-size: .85em; font-family: var(--font-mono); }
.widget-body :deep(pre) { background: var(--color-bg-alt); border: 1px solid var(--color-border); border-radius: var(--radius); padding: 8px 10px; overflow-x: auto; margin: 0 0 6px; font-size: .85em; }
.widget-body :deep(pre code) { background: none; border: none; padding: 0; }
.widget-body :deep(blockquote) { border-left: 3px solid var(--color-border); padding-left: 10px; color: var(--color-text-muted); margin: 0 0 6px; }
.widget-body :deep(a) { color: var(--color-primary); text-decoration: underline; }
.widget-body :deep(strong) { font-weight: 700; color: var(--color-text); }
.widget-body :deep(hr) { border: none; border-top: 1px solid var(--color-border); margin: 8px 0; }
.widget-body :deep(> *:last-child) { margin-bottom: 0; }

.widget-dismiss {
  position: absolute;
  top: 10px;
  right: 12px;
  background: transparent;
  border: none;
  color: var(--color-text-muted);
  font-size: 16px;
  line-height: 1;
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
}
.widget-dismiss:hover { color: var(--color-text); background: var(--color-bg); }
</style>
