<template>
  <div v-if="visible" class="update-banner">
    <span>
      {{ $t('update.available', { version: latestVersion }) }}
      <a :href="releaseUrl" target="_blank" rel="noopener" class="update-link">{{ $t('update.release_notes') }}</a>
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
  color: #fff;
  font-weight: 600;
  margin-left: 6px;
  text-decoration: underline;
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
