<template>
  <div class="reactions-wrap" v-if="reactions && reactions.length || showPicker">
    <div class="reactions">
      <button
        v-for="r in reactions"
        :key="r.emoji"
        :class="['reaction-pill', { 'reacted': hasReacted(r) }]"
        @click="$emit('toggle', r.emoji)"
      >
        {{ r.emoji }} <span class="reaction-count">{{ r.count }}</span>
        <span class="reaction-tooltip">{{ reactionTooltip(r) }}</span>
      </button>
    </div>
    <div class="add-reaction-wrap">
      <button class="add-reaction-btn" @click.stop="showPicker = !showPicker" title="Add reaction">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M8 13s1.5 2 4 2 4-2 4-2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>
      </button>
      <EmojiPicker v-if="showPicker" @pick="onPick" @close="showPicker = false" />
    </div>
  </div>
  <div class="add-reaction-wrap" v-else>
    <button class="add-reaction-btn" @click.stop="showPicker = !showPicker" title="Add reaction">
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M8 13s1.5 2 4 2 4-2 4-2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>
    </button>
    <EmojiPicker v-if="showPicker" @pick="onPick" @close="showPicker = false" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import EmojiPicker from './EmojiPicker.vue'

const props = defineProps({
  reactions: { type: Array, default: () => [] },
  users:     { type: Array, default: () => [] },
})
const emit = defineEmits(['toggle'])

const auth = useAuthStore()
const showPicker = ref(false)

function hasReacted(r) {
  return r.user_ids?.includes(auth.user?.id)
}

function reactionTooltip(r) {
  if (!r.user_ids?.length) return r.emoji
  const names = r.user_ids.map(uid => {
    if (uid === auth.user?.id) return 'You'
    const u = props.users.find(u => u.id === uid)
    return u ? (u.display_name || u.username) : `User ${uid}`
  })
  return names.join(', ')
}

function onPick(emoji) {
  showPicker.value = false
  emit('toggle', emoji)
}
</script>

<style scoped>
.reactions-wrap {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 4px;
}
.reactions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.reaction-pill {
  position: relative;
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 2px 7px;
  border-radius: 12px;
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  font-size: 13px;
  cursor: pointer;
  line-height: 1.4;
  transition: border-color .15s, background .15s;
}
.reaction-pill:hover { border-color: var(--color-primary); }
.reaction-pill.reacted {
  border-color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 12%, transparent);
}
.reaction-count { font-size: 11px; color: var(--color-text-muted); }

.reaction-tooltip {
  display: none;
  position: absolute;
  bottom: calc(100% + 6px);
  left: 50%;
  transform: translateX(-50%);
  background: var(--color-tooltip-bg, #1e293b);
  color: var(--color-tooltip-text, #f1f5f9);
  font-size: 11px;
  white-space: nowrap;
  padding: 4px 8px;
  border-radius: 5px;
  pointer-events: none;
  z-index: 100;
}
.reaction-tooltip::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 5px solid transparent;
  border-top-color: var(--color-tooltip-bg, #1e293b);
}
.reaction-pill:hover .reaction-tooltip { display: block; }

.add-reaction-wrap { position: relative; }
.add-reaction-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 11px;
  border: 1px dashed var(--color-border);
  background: none;
  cursor: pointer;
  color: var(--color-text-muted);
  opacity: 0;
  transition: opacity .15s;
}
.reactions-wrap:hover .add-reaction-btn,
.add-reaction-wrap:hover .add-reaction-btn,
.add-reaction-btn:focus { opacity: 1; }
.add-reaction-btn:hover { border-color: var(--color-primary); color: var(--color-primary); }
</style>
