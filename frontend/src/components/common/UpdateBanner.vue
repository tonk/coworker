<template>
  <div v-if="visible" class="update-banner">
    <span>
      {{ $t('update.available', { version: latestVersion }) }}
      <button class="update-link" @click="openLink(releaseUrl)">{{ $t('update.release_notes') }}</button>
    </span>
    <button class="update-dismiss" @click="dismiss" :aria-label="$t('common.close')">✕</button>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  latestVersion: { type: String, required: true },
  releaseUrl: { type: String, required: true }
})

async function openLink(url) {
  if (window.__TAURI_INTERNALS__) {
    await window.__TAURI_INTERNALS__.invoke('plugin:opener|open_url', { url, with: null })
  } else {
    window.open(url, '_blank', 'noopener,noreferrer')
  }
}

const DISMISS_KEY = 'update_dismissed'

const visible = ref(sessionStorage.getItem(DISMISS_KEY) !== props.latestVersion)

function dismiss() {
  sessionStorage.setItem(DISMISS_KEY, props.latestVersion)
  visible.value = false
}
</script>

<style scoped>
.update-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 16px;
  background: var(--color-primary, #3b82f6);
  color: #fff;
  font-size: 13px;
  line-height: 1.4;
  z-index: 200;
}
.update-link {
  background: none;
  border: none;
  padding: 0;
  color: #fff;
  font-weight: 600;
  margin-left: 6px;
  text-decoration: underline;
  cursor: pointer;
  font-size: inherit;
}
.update-link:hover { opacity: 0.85; }
.update-dismiss {
  background: none;
  border: none;
  color: #fff;
  cursor: pointer;
  font-size: 14px;
  padding: 2px 6px;
  opacity: 0.8;
  flex-shrink: 0;
}
.update-dismiss:hover { opacity: 1; }
</style>
