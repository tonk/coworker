<template>
  <a v-if="preview" :href="url" target="_blank" rel="noopener noreferrer" class="link-preview-card">
    <img v-if="previewImage && !imgBroken" :src="previewImage" class="preview-thumb" alt="" aria-hidden="true" @error="imgBroken = true" />
    <div class="preview-text">
      <div v-if="preview.title" class="preview-title">{{ preview.title }}</div>
      <div v-if="preview.description" class="preview-desc">{{ preview.description }}</div>
      <div class="preview-site">{{ preview.site_name || displayHost }}</div>
    </div>
  </a>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useLinkPreview } from '@/composables/useLinkPreview'
import { resolveAssetUrl } from '@/api/serverConfig'

const props = defineProps({ url: { type: String, required: true } })

const { cache, fetchPreview } = useLinkPreview()
const imgBroken = ref(false)

const preview = computed(() => cache[props.url] || null)
const previewImage = computed(() => resolveAssetUrl(preview.value?.image || ''))

const displayHost = computed(() => {
  try { return new URL(props.url).hostname } catch { return props.url }
})

onMounted(() => fetchPreview(props.url))
watch(() => props.url, url => { imgBroken.value = false; fetchPreview(url) })
</script>

<style scoped>
.link-preview-card {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  margin-top: 6px;
  padding: 8px 10px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-left: 3px solid var(--color-primary);
  border-radius: 6px;
  text-decoration: none;
  color: var(--color-text);
  max-width: 420px;
  overflow: hidden;
  transition: border-color .15s;
}
.link-preview-card:hover { border-left-color: var(--color-primary-hover, var(--color-primary)); }

.preview-thumb {
  width: 64px;
  height: 64px;
  object-fit: cover;
  border-radius: 4px;
  flex-shrink: 0;
}

.preview-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.preview-title {
  font-weight: 600;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.preview-desc {
  font-size: 12px;
  color: var(--color-text-muted);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.preview-site {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-top: 2px;
}
</style>
