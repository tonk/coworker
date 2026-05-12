<template>
  <div class="toast-container" role="status" aria-live="polite" aria-atomic="false">
    <div
      v-for="toast in ui.toasts"
      :key="toast.id"
      :class="`toast toast-${toast.type}`"
    >
      <template v-if="toast.type === 'mention'">
        <div class="toast-mention-title">
          <span class="toast-mention-icon">@</span>
          <strong>{{ toast.senderName }}</strong> mentioned you
        </div>
        <div class="toast-mention-body">{{ toast.context }}</div>
        <div class="toast-mention-preview">{{ toast.body }}</div>
      </template>
      <template v-else>{{ toast.message }}</template>
    </div>
  </div>
</template>

<script setup>
import { useUIStore } from '@/stores/ui'
const ui = useUIStore()
</script>

<style scoped>
.toast-mention-title {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 600;
  margin-bottom: 2px;
}
.toast-mention-icon {
  background: var(--color-primary);
  color: white;
  border-radius: 50%;
  width: 18px;
  height: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}
.toast-mention-body {
  font-size: 11px;
  opacity: 0.7;
  margin-bottom: 4px;
}
.toast-mention-preview {
  font-size: 12px;
  opacity: 0.9;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
</style>
