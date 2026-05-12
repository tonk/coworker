<template>
  <Teleport to="body">
    <div
      class="call-settings-panel"
      ref="panelEl"
      :style="{ top: pos.top + 'px', right: pos.right + 'px' }"
    >
      <div class="csp-header">
        <span class="csp-title">{{ $t('call.settings') }}</span>
        <button class="csp-close" @click="$emit('close')">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      <div v-if="loading" class="csp-loading">…</div>
      <template v-else>

        <!-- ── Microphone ─────────────────────────────────────────── -->
        <div class="csp-section">
          <label class="csp-label">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/></svg>
            {{ $t('call.microphone') }}
          </label>
          <select class="csp-select" v-model="localAudioIn" @change="onAudioInChange">
            <option v-for="d in devices.audioInputs" :key="d.deviceId" :value="d.deviceId">
              {{ d.label || $t('call.microphone') }}
            </option>
          </select>
          <!-- Level meter -->
          <div class="csp-level-wrap" :title="$t('call.input_level')">
            <div class="csp-level-bar" :style="{ width: micLevel + '%' }"></div>
          </div>
        </div>

        <!-- ── Camera ────────────────────────────────────────────── -->
        <div class="csp-section">
          <label class="csp-label">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>
            {{ $t('call.camera') }}
          </label>
          <select class="csp-select" v-model="localVideoIn" @change="onVideoInChange">
            <option v-for="d in devices.videoInputs" :key="d.deviceId" :value="d.deviceId">
              {{ d.label || $t('call.camera') }}
            </option>
            <option v-if="!devices.videoInputs.length" value="">{{ $t('call.no_camera') }}</option>
          </select>
          <!-- Camera preview -->
          <div class="csp-preview-wrap">
            <video ref="previewVideoEl" autoplay playsinline muted class="csp-preview"></video>
            <div v-if="!cameraActive" class="csp-preview-off">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><line x1="1" y1="1" x2="23" y2="23"/><path d="M21 21H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h3m3-3h6l2 3h4a2 2 0 0 1 2 2v9.34"/></svg>
            </div>
          </div>
        </div>

        <!-- ── Speaker (only shown when setSinkId is supported and devices found) ── -->
        <div v-if="showSpeaker" class="csp-section">
          <label class="csp-label">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><path d="M19.07 4.93a10 10 0 0 1 0 14.14"/><path d="M15.54 8.46a5 5 0 0 1 0 7.07"/></svg>
            {{ $t('call.speaker') }}
          </label>
          <div class="csp-row">
            <select class="csp-select" v-model="localAudioOut" @change="onAudioOutChange">
              <option v-for="d in devices.audioOutputs" :key="d.deviceId" :value="d.deviceId">
                {{ d.label || $t('call.speaker') }}
              </option>
            </select>
            <button class="csp-test-btn" @click="testSpeaker" :title="$t('call.test_speaker')">
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            </button>
          </div>
          <audio ref="testAudioEl" style="display:none"></audio>
        </div>

      </template>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useCallSettings } from '@/composables/useCallSettings'

const props = defineProps({
  pos: { type: Object, default: () => ({ top: 0, right: 0 }) },
})
const emit = defineEmits(['close'])

const { t } = useI18n()
const { devices, audioDeviceId, videoDeviceId, audioOutputId, loadDevices, setAudioDevice, setVideoDevice, setAudioOutput } = useCallSettings()

const panelEl      = ref(null)
const previewVideoEl = ref(null)
const testAudioEl  = ref(null)
const loading      = ref(true)
const micLevel     = ref(0)
const cameraActive = ref(false)

const localAudioIn  = ref(audioDeviceId.value)
const localVideoIn  = ref(videoDeviceId.value)
const localAudioOut = ref(audioOutputId.value)

const showSpeaker = computed(() =>
  devices.audioOutputs.length > 0 &&
  typeof HTMLAudioElement !== 'undefined' &&
  typeof HTMLAudioElement.prototype.setSinkId === 'function'
)

// ── Device change handlers ────────────────────────────────────────────────
function onAudioInChange() {
  setAudioDevice(localAudioIn.value)
  restartMic()
}
function onVideoInChange() {
  setVideoDevice(localVideoIn.value)
  restartCamera()
}
function onAudioOutChange() {
  setAudioOutput(localAudioOut.value)
}

// ── Microphone level monitor ──────────────────────────────────────────────
let _micStream   = null
let _audioCtx    = null
let _analyser    = null
let _animFrame   = null

async function startMic() {
  stopMic()
  try {
    const constraints = localAudioIn.value
      ? { audio: { deviceId: { ideal: localAudioIn.value } }, video: false }
      : { audio: true, video: false }
    _micStream = await navigator.mediaDevices.getUserMedia(constraints)
    _audioCtx  = new (window.AudioContext || window.webkitAudioContext)()
    const src  = _audioCtx.createMediaStreamSource(_micStream)
    _analyser  = _audioCtx.createAnalyser()
    _analyser.fftSize = 256
    src.connect(_analyser)
    const buf = new Uint8Array(_analyser.frequencyBinCount)
    function tick() {
      _analyser.getByteFrequencyData(buf)
      const avg = buf.reduce((s, v) => s + v, 0) / buf.length
      micLevel.value = Math.min(100, avg * 2.5)
      _animFrame = requestAnimationFrame(tick)
    }
    tick()
  } catch {}
}

function stopMic() {
  cancelAnimationFrame(_animFrame)
  _animFrame = null
  micLevel.value = 0
  if (_micStream) { _micStream.getTracks().forEach(t => t.stop()); _micStream = null }
  if (_audioCtx)  { try { _audioCtx.close() } catch {}; _audioCtx = null }
  _analyser = null
}

function restartMic() { startMic() }

// ── Camera preview ────────────────────────────────────────────────────────
let _camStream = null

async function startCamera() {
  stopCamera()
  if (!devices.videoInputs.length) return
  try {
    const constraints = localVideoIn.value
      ? { video: { deviceId: { ideal: localVideoIn.value } }, audio: false }
      : { video: true, audio: false }
    _camStream = await navigator.mediaDevices.getUserMedia(constraints)
    cameraActive.value = true
    await nextTick()
    if (previewVideoEl.value) previewVideoEl.value.srcObject = _camStream
  } catch {
    cameraActive.value = false
  }
}

function stopCamera() {
  if (_camStream) { _camStream.getTracks().forEach(t => t.stop()); _camStream = null }
  cameraActive.value = false
  if (previewVideoEl.value) previewVideoEl.value.srcObject = null
}

function restartCamera() { startCamera() }

// ── Speaker test ──────────────────────────────────────────────────────────
function testSpeaker() {
  try {
    const ctx = new (window.AudioContext || window.webkitAudioContext)()
    if (localAudioOut.value && typeof ctx.setSinkId === 'function') {
      ctx.setSinkId(localAudioOut.value).catch(() => {})
    }
    const osc  = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.frequency.value = 440
    osc.type = 'sine'
    gain.gain.setValueAtTime(0.3, ctx.currentTime)
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.7)
    osc.connect(gain)
    gain.connect(ctx.destination)
    osc.start()
    osc.stop(ctx.currentTime + 0.7)
    setTimeout(() => ctx.close(), 1000)
  } catch {}
}

// ── Click-outside / Escape to close ──────────────────────────────────────
function onPointerDown(e) {
  if (panelEl.value && !panelEl.value.contains(e.target)) emit('close')
}

function onKeyDown(e) {
  if (e.key === 'Escape') emit('close')
}

// ── Lifecycle ─────────────────────────────────────────────────────────────
onMounted(async () => {
  document.addEventListener('pointerdown', onPointerDown)
  document.addEventListener('keydown', onKeyDown)
  await loadDevices(true)
  localAudioIn.value  = audioDeviceId.value || devices.audioInputs[0]?.deviceId  || ''
  localVideoIn.value  = videoDeviceId.value || devices.videoInputs[0]?.deviceId  || ''
  localAudioOut.value = audioOutputId.value || devices.audioOutputs[0]?.deviceId || ''
  loading.value = false
  await startMic()
  await startCamera()
})

onUnmounted(() => {
  document.removeEventListener('pointerdown', onPointerDown)
  document.removeEventListener('keydown', onKeyDown)
  stopMic()
  stopCamera()
})
</script>

<style scoped>
.call-settings-panel {
  position: fixed;
  z-index: 600;
  width: 280px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.2);
  padding: 0;
  overflow: hidden;
}

.csp-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px 8px;
  border-bottom: 1px solid var(--color-border);
}
.csp-title {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-text);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.csp-close {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 2px;
  border-radius: 4px;
  display: flex;
  align-items: center;
}
.csp-close:hover { background: var(--color-bg); color: var(--color-text); }

.csp-loading {
  padding: 20px;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 13px;
}

.csp-section {
  padding: 10px 14px;
  border-bottom: 1px solid var(--color-border);
}
.csp-section:last-child { border-bottom: none; }

.csp-label {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 6px;
}

.csp-select {
  width: 100%;
  font-size: 13px;
  padding: 5px 8px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-bg);
  color: var(--color-text);
  cursor: pointer;
  margin-bottom: 6px;
}
.csp-select:last-child { margin-bottom: 0; }

/* Level meter */
.csp-level-wrap {
  height: 4px;
  background: var(--color-border);
  border-radius: 2px;
  overflow: hidden;
}
.csp-level-bar {
  height: 100%;
  background: #22c55e;
  border-radius: 2px;
  transition: width 0.05s linear;
  min-width: 0;
}

/* Camera preview */
.csp-preview-wrap {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  background: #111;
  border-radius: 6px;
  overflow: hidden;
  margin-top: 2px;
}
.csp-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scaleX(-1); /* mirror self-view */
}
.csp-preview-off {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255,255,255,0.3);
}

/* Speaker row */
.csp-row {
  display: flex;
  gap: 6px;
  align-items: center;
}
.csp-row .csp-select { flex: 1; margin-bottom: 0; }
.csp-test-btn {
  flex-shrink: 0;
  width: 30px;
  height: 30px;
  border-radius: 6px;
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  color: var(--color-text);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
.csp-test-btn:hover { background: var(--color-primary); border-color: var(--color-primary); color: #fff; }
</style>
