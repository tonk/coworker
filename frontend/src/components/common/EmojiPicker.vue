<template>
  <div class="emoji-picker" ref="el">
    <button v-for="e in emojis" :key="e" class="emoji-btn" @click="$emit('pick', e)">{{ e }}</button>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'

const emit = defineEmits(['pick', 'close'])

const emojis = ['👍', '❤️', '😂', '😮', '😢', '👎']
const el = ref(null)

function onClickOutside(e) {
  if (el.value && !el.value.contains(e.target)) {
    Promise.resolve().then(() => emit('close'))
  }
}

onMounted(() => document.addEventListener('mousedown', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('mousedown', onClickOutside))
</script>

<style scoped>
.emoji-picker {
  display: flex;
  gap: 2px;
  padding: 4px 6px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 20px;
  box-shadow: var(--shadow-md, 0 4px 12px rgba(0,0,0,.1));
  position: absolute;
  z-index: 300;
  bottom: 100%;
  left: 0;
  margin-bottom: 4px;
  white-space: nowrap;
}
.emoji-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 18px;
  padding: 2px 3px;
  border-radius: 6px;
  line-height: 1;
  transition: transform .1s;
}
.emoji-btn:hover { transform: scale(1.3); background: var(--color-bg); }
</style>
