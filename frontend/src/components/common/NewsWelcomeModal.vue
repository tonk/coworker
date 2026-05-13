<template>
  <Teleport to="body">
    <div class="welcome-backdrop" @click.self="close">
      <div class="welcome-modal" role="dialog" aria-modal="true" :aria-labelledby="'welcome-title-' + current.id">
        <div class="welcome-header" :style="current.sidebar_color ? { borderTopColor: current.sidebar_color } : {}">
          <div class="welcome-logo-row">
            <img src="/logo.svg" alt="WarmDesk" class="welcome-logo" />
            <span class="welcome-tag">{{ $t('dashboard.news_title') }}</span>
            <span v-if="items.length > 1" class="welcome-counter">{{ idx + 1 }} / {{ items.length }}</span>
          </div>
          <button class="welcome-close" :aria-label="$t('common.close')" @click="close">✕</button>
        </div>

        <div class="welcome-body">
          <h2 class="welcome-title" :id="'welcome-title-' + current.id">{{ current.title }}</h2>
          <div class="welcome-content markdown-body" v-html="renderMarkdown(current.text)"></div>
        </div>

        <div class="welcome-footer">
          <button v-if="idx < items.length - 1" class="btn btn-primary" @click="next">
            {{ $t('news.welcome_next') }}
            <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-left:4px">
              <polyline points="6,3 11,8 6,13"/>
            </svg>
          </button>
          <button v-else class="btn btn-primary" @click="close">{{ $t('news.welcome_got_it') }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const props = defineProps({
  items: { type: Array, required: true },
})
const emit = defineEmits(['close'])

const idx = ref(0)
const current = computed(() => props.items[idx.value])

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}

function next() {
  if (idx.value < props.items.length - 1) idx.value++
}

function close() {
  emit('close')
}

function onKeyDown(e) {
  if (e.key === 'Escape') close()
  if (e.key === 'ArrowRight' || e.key === 'Enter') {
    if (idx.value < props.items.length - 1) next()
    else close()
  }
}

onMounted(() => document.addEventListener('keydown', onKeyDown))
onBeforeUnmount(() => document.removeEventListener('keydown', onKeyDown))
</script>

<style scoped>
.welcome-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 24px;
}

.welcome-modal {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-md);
  width: 100%;
  max-width: 560px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.welcome-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 22px 14px;
  border-bottom: 1px solid var(--color-border);
  border-top: 4px solid var(--color-primary);
  flex-shrink: 0;
}

.welcome-logo-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.welcome-logo {
  height: 20px;
  width: auto;
  opacity: 0.85;
}

.welcome-tag {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .07em;
  color: var(--color-primary);
}

.welcome-counter {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-left: 4px;
}

.welcome-close {
  background: transparent;
  border: none;
  font-size: 16px;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  line-height: 1;
}
.welcome-close:hover { background: var(--color-bg); color: var(--color-text); }

.welcome-body {
  padding: 22px;
  overflow-y: auto;
  flex: 1;
}

.welcome-title {
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 14px;
  line-height: 1.3;
  color: var(--color-text);
}

.welcome-content {
  font-size: 14px;
  color: var(--color-text-muted);
  line-height: 1.6;
}
.welcome-content :deep(p)  { margin: 0 0 8px; }
.welcome-content :deep(h1),.welcome-content :deep(h2),.welcome-content :deep(h3) { font-weight: 700; margin: 10px 0 6px; color: var(--color-text); }
.welcome-content :deep(h1) { font-size: 1.05em; }
.welcome-content :deep(h2) { font-size: 1em; }
.welcome-content :deep(ul),.welcome-content :deep(ol) { padding-left: 20px; margin: 0 0 8px; }
.welcome-content :deep(li) { margin: 3px 0; }
.welcome-content :deep(code) { background: var(--color-bg-alt); border: 1px solid var(--color-border); border-radius: 3px; padding: 1px 5px; font-size: .85em; font-family: var(--font-mono); }
.welcome-content :deep(pre) { background: var(--color-bg-alt); border: 1px solid var(--color-border); border-radius: var(--radius); padding: 10px 12px; overflow-x: auto; margin: 0 0 8px; font-size: .85em; }
.welcome-content :deep(pre code) { background: none; border: none; padding: 0; }
.welcome-content :deep(blockquote) { border-left: 3px solid var(--color-border); padding-left: 12px; color: var(--color-text-muted); margin: 0 0 8px; }
.welcome-content :deep(a) { color: var(--color-primary); text-decoration: underline; }
.welcome-content :deep(strong) { font-weight: 700; color: var(--color-text); }
.welcome-content :deep(> *:last-child) { margin-bottom: 0; }

.welcome-footer {
  padding: 14px 22px;
  border-top: 1px solid var(--color-border);
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
}
</style>
