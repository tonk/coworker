<template>
  <div class="search-container" ref="containerEl">
    <div class="search-input-wrap">
      <svg class="search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
      <input
        class="search-input"
        v-model="query"
        placeholder="Search..."
        @focus="open = true"
        @keydown.escape="close"
        @keydown.down.prevent="moveDown"
        @keydown.up.prevent="moveUp"
        @keydown.enter.prevent="selectCurrent"
        ref="inputEl"
      />
      <span v-if="loading" class="search-spinner"></span>
    </div>

    <div v-if="open && (results.length || query.length >= 2)" class="search-results">
      <div v-if="!results.length && query.length >= 2 && !loading" class="search-empty">No results</div>

      <template v-for="(group, gName) in grouped" :key="gName">
        <div class="result-group-label">{{ groupLabel(gName) }}</div>
        <div
          v-for="(r, idx) in group"
          :key="r.item.id + gName"
          :class="['result-item', { active: activeIdx === globalIdx(gName, idx) }]"
          @mousedown.prevent="navigate(r)"
          @mouseover="activeIdx = globalIdx(gName, idx)"
        >
          <div class="result-icon">
            <svg v-if="r.type === 'card'" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="3" y1="9" x2="21" y2="9"/></svg>
            <svg v-else width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          </div>
          <div class="result-text">
            <div class="result-title">{{ resultTitle(r) }}</div>
            <div class="result-sub">{{ resultSub(r) }}</div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { searchApi } from '@/api/search'

const router = useRouter()
const query = ref('')
const results = ref([])
const open = ref(false)
const loading = ref(false)
const activeIdx = ref(0)
const containerEl = ref(null)
const inputEl = ref(null)

let debounceTimer = null

watch(query, (val) => {
  clearTimeout(debounceTimer)
  results.value = []
  if (val.length < 2) return
  loading.value = true
  debounceTimer = setTimeout(async () => {
    try {
      const { data } = await searchApi.search(val)
      results.value = data || []
    } catch {}
    loading.value = false
  }, 300)
})

const grouped = computed(() => {
  const g = {}
  for (const r of results.value) {
    if (!g[r.type]) g[r.type] = []
    g[r.type].push(r)
  }
  return g
})

function groupLabel(type) {
  return { card: 'Cards', chat_message: 'Chat', dm_message: 'Messages', card_comment: 'Comments' }[type] || type
}

function resultTitle(r) {
  const item = r.item
  if (r.type === 'card') return item.title || 'Card'
  if (r.type === 'chat_message') return item.body?.slice(0, 60) || ''
  if (r.type === 'dm_message') return item.body?.slice(0, 60) || ''
  if (r.type === 'card_comment') return item.body?.slice(0, 60) || ''
  return ''
}

function resultSub(r) {
  const item = r.item
  if (r.type === 'chat_message') return 'by ' + (item.user?.display_name || item.user?.username || '')
  if (r.type === 'dm_message') return 'from ' + (item.sender?.display_name || item.sender?.username || '')
  if (r.type === 'card_comment') return 'comment by ' + (item.user?.display_name || item.user?.username || '')
  return ''
}

// Flat index for keyboard navigation
const flatResults = computed(() => {
  const flat = []
  for (const [gName, group] of Object.entries(grouped.value)) {
    for (let i = 0; i < group.length; i++) {
      flat.push({ gName, idx: i, result: group[i] })
    }
  }
  return flat
})

function globalIdx(gName, idx) {
  let n = 0
  for (const [g, group] of Object.entries(grouped.value)) {
    if (g === gName) return n + idx
    n += group.length
  }
  return -1
}

function moveDown() { activeIdx.value = Math.min(activeIdx.value + 1, flatResults.value.length - 1) }
function moveUp() { activeIdx.value = Math.max(activeIdx.value - 1, 0) }

function selectCurrent() {
  const entry = flatResults.value[activeIdx.value]
  if (entry) navigate(entry.result)
}

function navigate(r) {
  close()
  const item = r.item
  if (r.type === 'card') {
    router.push(`/projects/${item.project_id || ''}`)
  } else if (r.type === 'chat_message') {
    router.push(`/projects/${item.project_id || ''}`)
  } else if (r.type === 'dm_message') {
    router.push(`/messages`)
  } else if (r.type === 'card_comment') {
    router.push(`/projects/${item.project_id || ''}`)
  }
}

function close() {
  open.value = false
  query.value = ''
  results.value = []
}

function onClickOutside(e) {
  if (containerEl.value && !containerEl.value.contains(e.target)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('mousedown', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('mousedown', onClickOutside))
</script>

<style scoped>
.search-container {
  position: relative;
  width: 260px;
}
.search-input-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 5px 10px;
  transition: border-color .15s;
}
.search-input-wrap:focus-within { border-color: var(--color-primary); }
.search-icon { color: var(--color-text-muted); flex-shrink: 0; }
.search-input {
  flex: 1;
  border: none;
  background: transparent;
  outline: none;
  font-size: 13px;
  color: var(--color-text);
}
.search-input::placeholder { color: var(--color-text-muted); }
.search-spinner {
  width: 12px;
  height: 12px;
  border: 2px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin .6s linear infinite;
  flex-shrink: 0;
}
@keyframes spin { to { transform: rotate(360deg); } }

.search-results {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(0,0,0,.12);
  z-index: 500;
  max-height: 360px;
  overflow-y: auto;
  padding: 4px 0;
}
.search-empty {
  padding: 16px;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 13px;
}
.result-group-label {
  padding: 6px 12px 2px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .06em;
  color: var(--color-text-muted);
}
.result-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  cursor: pointer;
  transition: background .1s;
}
.result-item.active,
.result-item:hover { background: var(--color-bg); }
.result-icon { color: var(--color-text-muted); flex-shrink: 0; }
.result-text { overflow: hidden; }
.result-title {
  font-size: 13px;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.result-sub {
  font-size: 11px;
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
