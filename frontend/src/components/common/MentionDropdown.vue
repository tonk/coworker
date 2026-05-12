<template>
  <div class="mention-dropdown" ref="el">
    <div
      v-for="(user, i) in users"
      :key="user.id"
      class="mention-item"
      :class="{ active: i === activeIndex }"
      @mousedown.prevent="$emit('pick', user)"
      @mousemove="$emit('update:activeIndex', i)"
    >
      <div class="mention-avatar" :style="avatarBg(user)">
        <img v-if="avatarSrc(user)" :src="avatarSrc(user)" class="avatar-img" alt="" aria-hidden="true" @error="e => e.target.style.display='none'" />
        <span v-else class="avatar-initials">{{ initials(user) }}</span>
      </div>
      <span class="mention-name">{{ user.display_name || user.username }}</span>
      <span class="mention-handle">@{{ user.username }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { avatarUrl } from '@/composables/useAvatar'

defineProps({
  users: { type: Array, default: () => [] },
  activeIndex: { type: Number, default: 0 }
})
defineEmits(['pick', 'update:activeIndex'])

const COLORS = ['#6366f1','#8b5cf6','#ec4899','#f59e0b','#10b981','#3b82f6','#ef4444','#14b8a6']

function avatarBg(user) {
  const idx = (user?.username?.charCodeAt(0) || 0) % COLORS.length
  return { background: COLORS[idx] }
}

function avatarSrc(user) {
  return avatarUrl(user)
}

function initials(user) {
  const name = user?.display_name || user?.username || '?'
  return name.slice(0, 2).toUpperCase()
}
</script>

<style scoped>
.mention-dropdown {
  position: absolute;
  bottom: calc(100% + 4px);
  left: 0;
  min-width: 220px;
  max-width: 320px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: 0 6px 20px rgba(0,0,0,.12);
  z-index: 400;
  overflow: hidden;
}

.mention-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  cursor: pointer;
  transition: background .1s;
}
.mention-item:hover,
.mention-item.active {
  background: var(--color-bg);
}

.mention-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}
.avatar-img { width: 100%; height: 100%; object-fit: cover; border-radius: 50%; }
.avatar-initials { font-size: 9px; font-weight: 700; color: #fff; }

.mention-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mention-handle {
  font-size: 11px;
  color: var(--color-text-muted);
  flex-shrink: 0;
}
</style>
