<template>
  <div class="lk-cell">
    <video
      v-if="track"
      ref="videoEl"
      class="lk-video"
      autoplay
      playsinline
      :class="{ mirror: isLocal && !isScreenShare, 'lk-video--screen': isScreenShare }"
    />
    <div v-else class="lk-placeholder">
      <img v-if="avatar" :src="avatar" class="lk-avatar" @error="onAvatarError" />
      <span v-else class="lk-initial">{{ initial }}</span>
    </div>
    <div v-if="track && cameraOff" class="lk-off-overlay">
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="1" y1="1" x2="23" y2="23"/>
        <path d="M21 21H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h3m3-3h6l2 3h4a2 2 0 0 1 2 2v9.34"/>
      </svg>
    </div>
    <div class="lk-name">{{ displayLabel }}</div>
  </div>
</template>

<script setup>
import { ref, watch, onUnmounted, computed } from 'vue'

const props = defineProps({
  /** @type {import('livekit-client').LocalVideoTrack | import('livekit-client').RemoteVideoTrack | null} */
  track: { type: Object, default: null },
  label: { type: String, default: '' },
  avatar: { type: String, default: '' },
  isLocal: { type: Boolean, default: false },
  /** Publication muted (camera off) but track object may still exist */
  cameraOff: { type: Boolean, default: false },
  /** Screen share — fit entire frame, no mirror */
  isScreenShare: { type: Boolean, default: false },
})

const videoEl = ref(null)
let _attached = null
const _avatarHidden = ref(false)

const displayLabel = computed(() => props.label || '')
const avatar = computed(() => (_avatarHidden.value ? '' : (props.avatar || '')))

const initial = computed(() => {
  const s = displayLabel.value.trim()
  if (s.length) return s.charAt(0).toUpperCase()
  return '?'
})

watch(
  () => [props.track, videoEl.value],
  () => {
    if (_attached && videoEl.value) {
      try {
        _attached.detach(videoEl.value)
      } catch {}
      _attached = null
    }
    const t = props.track
    const el = videoEl.value
    if (t && el) {
      try {
        t.attach(el)
        _attached = t
      } catch {}
    }
  },
  { flush: 'post' }
)

watch(
  () => props.avatar,
  () => {
    _avatarHidden.value = false
  }
)

onUnmounted(() => {
  if (_attached && videoEl.value) {
    try {
      _attached.detach(videoEl.value)
    } catch {}
  }
  _attached = null
})

function onAvatarError() {
  _avatarHidden.value = true
}
</script>

<style scoped>
.lk-cell {
  position: relative;
  aspect-ratio: 16 / 10;
  border-radius: 10px;
  overflow: hidden;
  background: #111;
  border: 1px solid rgba(255, 255, 255, 0.12);
}
.lk-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.lk-video.mirror {
  transform: scaleX(-1);
}
.lk-video--screen {
  object-fit: contain;
  background: #0a0a0a;
}
.lk-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(145deg, #1e293b, #0f172a);
}
.lk-initial {
  font-size: 28px;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.35);
}
.lk-avatar {
  width: 72px;
  height: 72px;
  border-radius: 999px;
  object-fit: cover;
  border: 2px solid rgba(255, 255, 255, 0.25);
}
.lk-off-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.55);
  color: rgba(255, 255, 255, 0.45);
}
.lk-name {
  position: absolute;
  left: 8px;
  bottom: 6px;
  right: 8px;
  font-size: 12px;
  font-weight: 600;
  color: #fff;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.85);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
