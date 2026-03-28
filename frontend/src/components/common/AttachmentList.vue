<template>
  <div v-if="attachments && attachments.length" class="attachment-list">
    <div v-for="a in attachments" :key="a.id" class="attachment-item">
      <template v-if="a.mime_type?.startsWith('image/')">
        <a :href="attachmentsApi.url(a.id)" target="_blank" class="attachment-image-wrap">
          <img :src="attachmentsApi.url(a.id)" :alt="a.filename" class="attachment-thumb" @error="e => e.target.style.display='none'" />
        </a>
      </template>
      <template v-else>
        <a :href="attachmentsApi.url(a.id)" target="_blank" class="attachment-file">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
          <span class="attachment-name">{{ a.filename }}</span>
          <span class="attachment-size">{{ formatSize(a.size_bytes) }}</span>
        </a>
      </template>
      <button v-if="canDelete" class="attachment-remove" @click="$emit('remove', a)" title="Remove">×</button>
    </div>
  </div>
</template>

<script setup>
import { attachmentsApi } from '@/api/attachments'

defineProps({
  attachments: { type: Array, default: () => [] },
  canDelete: { type: Boolean, default: false }
})
defineEmits(['remove'])

function formatSize(bytes) {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>

<style scoped>
.attachment-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 6px;
}
.attachment-item {
  position: relative;
  display: flex;
  align-items: center;
}
.attachment-image-wrap {
  display: block;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--color-border);
}
.attachment-thumb {
  display: block;
  max-width: 180px;
  max-height: 160px;
  object-fit: cover;
}
.attachment-file {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 8px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 12px;
  color: var(--color-text);
  text-decoration: none;
  background: var(--color-bg);
  max-width: 200px;
}
.attachment-file:hover { border-color: var(--color-primary); }
.attachment-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 120px;
}
.attachment-size { color: var(--color-text-muted); flex-shrink: 0; }
.attachment-remove {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--color-danger, #ef4444);
  color: #fff;
  border: none;
  font-size: 11px;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}
</style>
