<template>
  <Transition name="call-ui">
    <div v-if="state.phase === 'calling' || state.phase === 'active'">

      <!-- ── VIDEO OVERLAY (active + video) ──────────────────────────────── -->
      <div v-if="state.phase === 'active' && state.hasVideo" class="video-overlay">
        <!-- Remote video fills the background -->
        <video ref="remoteVideo" autoplay playsinline class="remote-video"></video>

        <!-- Local self-preview (PiP, bottom-right) -->
        <div class="local-pip">
          <video ref="localVideo" autoplay playsinline muted class="local-video"></video>
          <div v-if="state.isCameraOff" class="pip-off">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M21 21H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h3m3-3h6l2 3h4a2 2 0 0 1 2 2v9.34"/>
            </svg>
          </div>
        </div>

        <!-- Remote name (top-left) -->
        <div class="remote-name">{{ state.remoteName }}</div>
        <div class="call-duration">{{ formattedDuration }}</div>

        <!-- Controls bar -->
        <div class="video-controls">
          <button :class="['vc-btn', { active: state.isMuted }]" :title="state.isMuted ? $t('call.unmute') : $t('call.mute')" @click="toggleMute">
            <svg v-if="!state.isMuted" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
              <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
              <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
            </svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
              <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"/>
              <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
            </svg>
          </button>

          <button :class="['vc-btn', { active: state.isCameraOff }]" :title="state.isCameraOff ? $t('call.camera_on') : $t('call.camera_off')" @click="toggleCamera">
            <svg v-if="!state.isCameraOff" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="23 7 16 12 23 17 23 7"/>
              <rect x="1" y="5" width="15" height="14" rx="2" ry="2"/>
            </svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M21 21H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h3m3-3h6l2 3h4a2 2 0 0 1 2 2v9.34"/>
            </svg>
          </button>

          <button class="vc-btn end-btn" :title="$t('call.hangup')" @click="endCall()">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M10.68 13.31a16 16 0 0 0 3.41 2.6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7 2 2 0 0 1 1.72 2v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.42 19.42 0 0 1-3.33-2.67m-2.67-3.34a19.79 19.79 0 0 1-3.07-8.63A2 2 0 0 1 3.6 1.27h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L7.91 9.91M1 1l22 22"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- ── AUDIO BAR (calling phase or active without video) ─────────────── -->
      <div v-else class="active-call-bar">
        <!-- Hidden audio element for remote stream in audio-only mode -->
        <audio ref="remoteAudio" autoplay playsinline></audio>

        <div class="call-bar-info">
          <div class="call-bar-name">{{ state.remoteName || '…' }}</div>
          <div class="call-bar-status">
            <span v-if="state.phase === 'calling'" class="status-calling">{{ $t('call.calling') }}</span>
            <span v-else class="status-duration">{{ formattedDuration }}</span>
          </div>
        </div>

        <div class="call-bar-actions">
          <button :class="['call-bar-btn', { active: state.isMuted }]" :title="state.isMuted ? $t('call.unmute') : $t('call.mute')" @click="toggleMute">
            <svg v-if="!state.isMuted" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
              <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
              <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
            </svg>
            <svg v-else width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
              <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"/>
              <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
            </svg>
          </button>

          <button class="call-bar-btn end-btn" :title="$t('call.hangup')" @click="endCall()">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M10.68 13.31a16 16 0 0 0 3.41 2.6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7 2 2 0 0 1 1.72 2v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.42 19.42 0 0 1-3.33-2.67m-2.67-3.34a19.79 19.79 0 0 1-3.07-8.63A2 2 0 0 1 3.6 1.27h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L7.91 9.91M1 1l22 22"/>
            </svg>
          </button>
        </div>
      </div>

    </div>
  </Transition>
</template>

<script setup>
import { ref, watch, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { useWebRTCCall } from '@/composables/useWebRTCCall'
import { useCallSettings } from '@/composables/useCallSettings'

const { state, toggleMute, toggleCamera, endCall, setRemoteStreamCallback, setLocalStreamCallback } = useWebRTCCall()
const { audioOutputId } = useCallSettings()

const remoteVideo = ref(null)
const localVideo  = ref(null)
const remoteAudio = ref(null)

// Store streams locally so we can re-apply when DOM elements appear
const _remoteStream = ref(null)
const _localStream  = ref(null)

function _applySinkId(el) {
  if (!el || !audioOutputId.value) return
  if (typeof el.setSinkId === 'function') {
    el.setSinkId(audioOutputId.value).catch(() => {})
  }
}

function _applyStreams() {
  nextTick(() => {
    if (_remoteStream.value) {
      if (remoteVideo.value) { remoteVideo.value.srcObject = _remoteStream.value; _applySinkId(remoteVideo.value) }
      if (remoteAudio.value) { remoteAudio.value.srcObject = _remoteStream.value; _applySinkId(remoteAudio.value) }
    }
    if (_localStream.value && localVideo.value) {
      localVideo.value.srcObject = _localStream.value
    }
  })
}

// Re-apply sink when output device changes mid-call
watch(audioOutputId, () => {
  _applySinkId(remoteVideo.value)
  _applySinkId(remoteAudio.value)
})

onMounted(() => {
  setRemoteStreamCallback((stream) => {
    _remoteStream.value = stream
    _applyStreams()
  })
  setLocalStreamCallback((stream) => {
    _localStream.value = stream
    _applyStreams()
  })
})

onUnmounted(() => {
  setRemoteStreamCallback(null)
  setLocalStreamCallback(null)
})

// Re-apply when phase changes (video overlay appears after 'active' transition)
watch(() => state.phase, _applyStreams)
watch(() => state.hasVideo, _applyStreams)

// ── Duration timer ──────────────────────────────────────────────────────────
const durationSeconds = ref(0)
let _durationTimer = null

watch(() => state.phase, (phase) => {
  if (phase === 'active') {
    durationSeconds.value = 0
    _durationTimer = setInterval(() => { durationSeconds.value++ }, 1000)
  } else {
    clearInterval(_durationTimer)
    _durationTimer = null
    durationSeconds.value = 0
    if (phase === 'ended' && state.errorMsg) {
      const key = state.errorMsg === 'unavailable'
        ? t('call.unavailable', { name: state.remoteName || '…' })
        : state.errorMsg === 'no_mic'
          ? t('call.no_mic')
          : null
      if (key) ui.error(key)
    }
  }
})

onUnmounted(() => clearInterval(_durationTimer))

const formattedDuration = computed(() => {
  const s = durationSeconds.value
  const m = Math.floor(s / 60)
  return `${String(m).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`
})
</script>

<style scoped>
/* ── VIDEO OVERLAY ─────────────────────────────────────────────────────────── */
.video-overlay {
  position: fixed;
  inset: 0;
  z-index: 500;
  background: #000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.remote-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.local-pip {
  position: absolute;
  bottom: 80px;
  right: 16px;
  width: 160px;
  height: 90px;
  border-radius: 10px;
  overflow: hidden;
  background: #1a1a1a;
  border: 2px solid rgba(255,255,255,0.15);
  box-shadow: 0 4px 16px rgba(0,0,0,0.5);
}
.local-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scaleX(-1); /* mirror self-view */
}
.pip-off {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1a1a1a;
  color: rgba(255,255,255,0.4);
}

.remote-name {
  position: absolute;
  top: 20px;
  left: 24px;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  text-shadow: 0 1px 6px rgba(0,0,0,0.7);
}
.call-duration {
  position: absolute;
  top: 44px;
  left: 24px;
  color: rgba(255,255,255,0.7);
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  text-shadow: 0 1px 4px rgba(0,0,0,0.6);
}

.video-controls {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 20px;
  background: linear-gradient(transparent, rgba(0,0,0,0.65));
}

.vc-btn {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: none;
  background: rgba(255,255,255,0.18);
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, transform 0.1s;
  backdrop-filter: blur(4px);
}
.vc-btn:hover { background: rgba(255,255,255,0.28); transform: scale(1.06); }
.vc-btn.active { background: rgba(255,255,255,0.85); color: #111; }
.vc-btn.end-btn { background: #ef4444; }
.vc-btn.end-btn:hover { background: #dc2626; transform: scale(1.06); }

/* ── AUDIO BAR ─────────────────────────────────────────────────────────────── */
.active-call-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 490;
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
  padding: 10px 24px;
  box-shadow: 0 -4px 16px rgba(0,0,0,0.12);
  height: 56px;
}
.call-bar-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
.call-bar-name { font-size: 14px; font-weight: 600; color: var(--color-text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.call-bar-status { font-size: 12px; }
.status-calling { color: var(--color-text-muted); animation: pulse 1.4s ease-in-out infinite; }
.status-duration { color: #22c55e; font-variant-numeric: tabular-nums; }

.call-bar-actions { display: flex; align-items: center; gap: 10px; }
.call-bar-btn {
  width: 36px; height: 36px;
  border-radius: 50%;
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  color: var(--color-text);
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.call-bar-btn:hover { background: var(--color-surface-raised, var(--color-bg)); }
.call-bar-btn.active { background: var(--color-primary); border-color: var(--color-primary); color: #fff; }
.call-bar-btn.end-btn { background: #ef4444; border-color: #ef4444; color: #fff; }
.call-bar-btn.end-btn:hover { opacity: 0.88; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}

/* ── Transitions ────────────────────────────────────────────────────────────── */
.call-ui-enter-from,
.call-ui-leave-to { opacity: 0; transform: translateY(100%); }
.call-ui-enter-active,
.call-ui-leave-active { transition: opacity 0.2s, transform 0.2s; }
</style>
