<template>
  <Transition name="call-overlay">
    <div
      v-if="state.phase === 'ringing'"
      class="incoming-call-overlay"
      role="alertdialog"
      aria-modal="true"
      :aria-labelledby="labelId"
    >
      <div class="call-avatar">
        <img
          v-if="state.remoteAvatar"
          :src="state.remoteAvatar"
          class="avatar-img"
          aria-hidden="true"
          @error="e => e.target.style.display = 'none'"
        />
        <span v-else class="avatar-initials" aria-hidden="true">{{ initials }}</span>
      </div>
      <div class="call-info">
        <div class="call-label" aria-hidden="true">
          <svg v-if="state.hasVideo" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="margin-right:4px;vertical-align:middle" aria-hidden="true"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>
          {{ state.hasVideo ? $t('call.incoming_video') : $t('call.incoming') }}
        </div>
        <div :id="labelId" class="call-name">{{ state.remoteName || $t('common.unknown') }}</div>
      </div>
      <div class="call-actions">
        <button ref="acceptBtnEl" class="call-btn accept-btn" :aria-label="$t('call.accept')" @click="acceptCall">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07A19.5 19.5 0 0 1 4.69 12a19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 3.6 1.27h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L7.91 9.91a16 16 0 0 0 6.07 6.07l1.22-1.22a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
          </svg>
        </button>
        <button class="call-btn decline-btn" :aria-label="$t('call.decline')" @click="rejectCall">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M10.68 13.31a16 16 0 0 0 3.41 2.6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7 2 2 0 0 1 1.72 2v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.42 19.42 0 0 1-3.33-2.67m-2.67-3.34a19.79 19.79 0 0 1-3.07-8.63A2 2 0 0 1 3.6 1.27h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L7.91 9.91M1 1l22 22"/>
          </svg>
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { computed, ref, watch, nextTick, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useWebRTCCall } from '@/composables/useWebRTCCall'

const { t } = useI18n()
const { state, acceptCall, rejectCall } = useWebRTCCall()

const labelId = 'incoming-call-label-' + Math.random().toString(36).slice(2, 8)
const acceptBtnEl = ref(null)

const initials = computed(() => {
  const name = state.remoteName || '?'
  return name.slice(0, 2).toUpperCase()
})

watch(() => state.phase, (phase) => {
  if (phase === 'ringing') nextTick(() => acceptBtnEl.value?.focus())
})

// ── Ringtone (Web Audio API oscillator) ──────────────────────────────────────
let _audioCtx = null
let _gainNode = null
let _beepTimer = null
let _activeOsc = null

function _beep() {
  if (!_audioCtx) return
  if (_activeOsc) {
    try { _activeOsc.stop() } catch {}
    _activeOsc = null
  }
  const osc = _audioCtx.createOscillator()
  const gain = _gainNode
  osc.type = 'sine'
  osc.frequency.value = 520
  osc.connect(gain)
  osc.start()
  osc.stop(_audioCtx.currentTime + 0.18)
  _activeOsc = osc
}

function _startRingtone() {
  try {
    _audioCtx = new (window.AudioContext || window.webkitAudioContext)()
    _gainNode = _audioCtx.createGain()
    _gainNode.gain.value = 0.25
    _gainNode.connect(_audioCtx.destination)

    // Pattern: two short beeps then a pause, repeat
    let tick = 0
    function ring() {
      if (tick % 4 < 2) _beep()
      tick++
      _beepTimer = setTimeout(ring, 300)
    }
    ring()
  } catch {}
}

function _stopRingtone() {
  clearTimeout(_beepTimer)
  _beepTimer = null
  if (_activeOsc) {
    try { _activeOsc.stop() } catch {}
    _activeOsc = null
  }
  if (_audioCtx) {
    try { _audioCtx.close() } catch {}
    _audioCtx = null
  }
  _gainNode = null
}

watch(() => state.phase, (phase) => {
  if (phase === 'ringing') _startRingtone()
  else _stopRingtone()
})
onUnmounted(_stopRingtone)
</script>

<style scoped>
.incoming-call-overlay {
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: 500;
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

.call-info {
  flex: 1;
  min-width: 0;
}
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

.call-actions {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}

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

.accept-btn { background: #22c55e; }
.decline-btn { background: #ef4444; }

/* Transition */
.call-overlay-enter-from,
.call-overlay-leave-to {
  opacity: 0;
  transform: translateY(12px) scale(0.95);
}
.call-overlay-enter-active,
.call-overlay-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}
</style>
