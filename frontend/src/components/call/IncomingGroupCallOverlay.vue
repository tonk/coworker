<template>
  <Transition name="call-overlay">
    <div v-if="invite.pending" class="incoming-call-overlay">
      <div class="call-avatar">
        <img
          v-if="invite.fromAvatar"
          :src="invite.fromAvatar"
          class="avatar-img"
          @error="e => e.target.style.display = 'none'"
        />
        <span v-else class="avatar-initials">{{ initials }}</span>
      </div>
      <div class="call-info">
        <div class="call-label">{{ $t('call.group_invite') }}</div>
        <div class="call-name">{{ invite.fromName || $t('common.unknown') }}</div>
      </div>
      <div class="call-actions">
        <button class="call-btn accept-btn" :title="$t('call.join_call')" @click="acceptGroupInvite">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="23 7 16 12 23 17 23 7"/>
            <rect x="1" y="5" width="15" height="14" rx="2" ry="2"/>
          </svg>
        </button>
        <button class="call-btn decline-btn" :title="$t('call.decline')" @click="declineGroupInvite">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M10.68 13.31a16 16 0 0 0 3.41 2.6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7 2 2 0 0 1 1.72 2v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.42 19.42 0 0 1-3.33-2.67m-2.67-3.34a19.79 19.79 0 0 1-3.07-8.63A2 2 0 0 1 3.6 1.27h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L7.91 9.91M1 1l22 22"/>
          </svg>
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { computed } from 'vue'
import { useLiveKitGroupCall } from '@/composables/useLiveKitGroupCall'

const { invite, acceptGroupInvite, declineGroupInvite } = useLiveKitGroupCall()

const initials = computed(() => {
  const name = invite.fromName || '?'
  return name.slice(0, 2).toUpperCase()
})
</script>

<style scoped>
.incoming-call-overlay {
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: 501;
  display: flex;
  align-items: center;
  gap: 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 16px;
  padding: 14px 18px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.22);
  min-width: 280px;
}

.call-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}
.avatar-img { width: 100%; height: 100%; object-fit: cover; }
.avatar-initials { color: #fff; font-size: 16px; font-weight: 700; }

.call-info { flex: 1; min-width: 0; }
.call-label {
  font-size: 11px;
  color: var(--color-text-muted);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 2px;
}
.call-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.call-actions { display: flex; gap: 10px; flex-shrink: 0; }
.call-btn {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  transition: opacity 0.15s, transform 0.1s;
  flex-shrink: 0;
}
.call-btn:hover { opacity: 0.88; transform: scale(1.06); }
.call-btn:active { transform: scale(0.96); }
.accept-btn { background: #3b82f6; }
.decline-btn { background: #ef4444; }

.call-overlay-enter-from,
.call-overlay-leave-to { opacity: 0; transform: translateY(12px) scale(0.95); }
.call-overlay-enter-active,
.call-overlay-leave-active { transition: opacity 0.2s, transform 0.2s; }
</style>
