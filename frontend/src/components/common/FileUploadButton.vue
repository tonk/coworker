<template>
  <label class="file-upload-btn" :title="'Attach file'">
    <input type="file" multiple ref="inputEl" @change="onFiles" style="display:none" />
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48"/>
    </svg>
  </label>
</template>

<script setup>
const emit = defineEmits(['files-selected'])

function onFiles(e) {
  const files = Array.from(e.target.files)
  if (files.length) emit('files-selected', files)
  e.target.value = ''
}
</script>

<style scoped>
.file-upload-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  cursor: pointer;
  color: var(--color-text-muted);
  flex-shrink: 0;
  transition: color .15s, background .15s;
}
.file-upload-btn:hover {
  color: var(--color-text);
  background: var(--color-bg);
}
</style>
