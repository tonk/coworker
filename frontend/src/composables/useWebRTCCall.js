/**
 * useWebRTCCall — singleton WebRTC audio/video call composable.
 *
 * All components share one call instance via module-level state so that
 * IncomingCallOverlay and ActiveCallBar can both react to the same call.
 *
 * 1:1 only. Group calls require an SFU (e.g. LiveKit).
 */

import { reactive, readonly } from 'vue'
import { useCallSettings } from './useCallSettings'
import { isLiveKitCallActive } from './callsGate'
import { useUIStore } from '@/stores/ui'
import { i18n } from '@/i18n'
import { webrtcApi } from '@/api/webrtc'

// ── Module-level singleton ──────────────────────────────────────────────────

const _s = reactive({
  phase:       'idle',   // idle | calling | ringing | active | ended
  remoteUserId: null,
  remoteName:  '',
  remoteAvatar: '',
  convId:      null,
  isMuted:     false,
  isCameraOff: false,
  hasVideo:    false,
  isScreenSharing: false,
  // Remote party's reported state — set only from call.mute/call.screen_share
  // signals, since track.enabled=false on the sender is invisible to the
  // receiver otherwise (WebRTC has no built-in "peer is muted" signal).
  remoteMuted: false,
  remoteScreenSharing: false,
  errorMsg:    '',
  // True while acceptCall() is negotiating (getUserMedia can take several
  // seconds with no other visual feedback). Lets the UI disable the accept
  // button so a second click can't start a concurrent, colliding negotiation.
  accepting:   false,
})

let _pc                = null
let _localStream       = null
/** Display capture stream while screen-sharing (video sender replaced); cleaned up on end. */
let _screenStream      = null
let _remoteStream      = null
let _pendingCandidates = []
let _pendingOffer      = null   // SDP stored while ringing
let _sendFn            = null   // set by App.vue
let _onRemoteStream    = null   // set by ActiveCallBar
let _onLocalStream     = null   // set by ActiveCallBar (self-preview)
let _ringTimeout       = null   // auto-cancel unanswered outgoing calls

const RING_TIMEOUT_MS = 45_000

// STUN-only fallback, used if the backend's /ice-servers call fails or the
// server has no TURN server configured.
const DEFAULT_ICE_SERVERS = [
  { urls: 'stun:stun.l.google.com:19302' },
  { urls: 'stun:stun1.l.google.com:19302' },
]

let _iceServersPromise = null

// Fetched once per session (TURN credentials don't change at runtime) and
// reused for every call. Falls back to STUN-only on any failure so a
// misconfigured/unreachable endpoint never blocks calling — it just loses
// TURN relay.
function _loadIceServers() {
  if (!_iceServersPromise) {
    _iceServersPromise = webrtcApi.getIceServers()
      .then((res) => res.data?.ice_servers || DEFAULT_ICE_SERVERS)
      .catch(() => DEFAULT_ICE_SERVERS)
  }
  return _iceServersPromise
}

// ── Helpers ─────────────────────────────────────────────────────────────────

function _send(msgObj) {
  if (typeof _sendFn === 'function') _sendFn(msgObj)
}

function _reset() {
  _s.phase        = 'idle'
  _s.remoteUserId = null
  _s.remoteName   = ''
  _s.remoteAvatar = ''
  _s.convId       = null
  _s.isMuted      = false
  _s.isCameraOff  = false
  _s.hasVideo     = false
  _s.isScreenSharing = false
  _s.remoteMuted  = false
  _s.remoteScreenSharing = false
  _s.errorMsg     = ''
  _s.accepting    = false
  _pendingCandidates = []
  _pendingOffer   = null
  _remoteStream   = null
}

function _teardown() {
  if (_screenStream) {
    _screenStream.getTracks().forEach(t => t.stop())
    _screenStream = null
    _s.isScreenSharing = false
  }
  if (_localStream) {
    _localStream.getTracks().forEach(t => t.stop())
    _localStream = null
  }
  if (_pc) {
    _pc.onicecandidate = null
    _pc.ontrack = null
    _pc.onconnectionstatechange = null
    _pc.close()
    _pc = null
  }
}

// ── RTCPeerConnection setup ──────────────────────────────────────────────────

function _setupPC(iceServers) {
  _pc = new RTCPeerConnection({ iceServers })

  _pc.onicecandidate = (e) => {
    if (!e.candidate) return
    _send({
      type: 'call.ice',
      payload: {
        to_user_id:      _s.remoteUserId,
        conversation_id: _s.convId,
        candidate:       JSON.stringify(e.candidate),
      },
    })
  }

  _pc.ontrack = (e) => {
    _remoteStream = e.streams[0] || new MediaStream([e.track])
    if (typeof _onRemoteStream === 'function') _onRemoteStream(_remoteStream)
  }

  _pc.onconnectionstatechange = () => {
    if (!_pc) return
    const st = _pc.connectionState
    if (st === 'connected') _s.phase = 'active'
    else if (st === 'failed' || st === 'closed') endCall(false)
  }
}

async function _applyPendingCandidates() {
  while (_pendingCandidates.length > 0) {
    const raw = _pendingCandidates.shift()
    try {
      const c = typeof raw === 'string' ? JSON.parse(raw) : raw
      await _pc.addIceCandidate(new RTCIceCandidate(c))
    } catch {}
  }
}

// ── Media acquisition (with video→audio fallback) ───────────────────────────

async function _getMedia(wantVideo) {
  const { getMediaConstraints } = useCallSettings()
  if (wantVideo) {
    try {
      const stream = await navigator.mediaDevices.getUserMedia(getMediaConstraints(true))
      return { stream, hasVideo: true }
    } catch {}
    // Camera unavailable — fall back to audio only
    const stream = await navigator.mediaDevices.getUserMedia(getMediaConstraints(false))
    return { stream, hasVideo: false }
  }
  const stream = await navigator.mediaDevices.getUserMedia(getMediaConstraints(false))
  return { stream, hasVideo: false }
}

// ── Public API ───────────────────────────────────────────────────────────────

function setSendFn(fn) {
  _sendFn = fn
}

function setRemoteStreamCallback(fn) {
  _onRemoteStream = fn
  if (_remoteStream && typeof fn === 'function') fn(_remoteStream)
}

function setLocalStreamCallback(fn) {
  _onLocalStream = fn
  if (_localStream && typeof fn === 'function') fn(_localStream)
}

async function startCall(userId, userName, avatar, convId, wantVideo = true) {
  if (_s.phase !== 'idle' || isLiveKitCallActive()) return

  _s.remoteUserId = userId
  _s.remoteName   = userName || ''
  _s.remoteAvatar = avatar  || ''
  _s.convId       = convId
  _s.phase        = 'calling'
  _s.errorMsg     = ''

  // Auto-cancel if callee doesn't answer within the timeout
  _ringTimeout = setTimeout(() => {
    if (_s.phase === 'calling') endCall(true)
  }, RING_TIMEOUT_MS)

  let stream, hasVideo, iceServers
  try {
    ;[{ stream, hasVideo }, iceServers] = await Promise.all([_getMedia(wantVideo), _loadIceServers()])
  } catch {
    clearTimeout(_ringTimeout); _ringTimeout = null
    _s.errorMsg = 'no_mic'
    _reset()
    return
  }

  // User may have cancelled while getUserMedia was pending
  if (_s.phase !== 'calling') {
    stream.getTracks().forEach(t => t.stop())
    return
  }

  _localStream  = stream
  _s.hasVideo   = hasVideo

  try {
    if (typeof _onLocalStream === 'function') _onLocalStream(_localStream)

    _setupPC(iceServers)
    _localStream.getTracks().forEach(t => _pc.addTrack(t, _localStream))

    const offer = await _pc.createOffer()
    await _pc.setLocalDescription(offer)

    if (_s.phase !== 'calling') return

    _send({
      type: 'call.offer',
      payload: {
        to_user_id:      userId,
        conversation_id: convId,
        sdp:             offer.sdp,
        has_video:       hasVideo,
      },
    })
  } catch {
    _failCall()
  }
}

function handleSignal(msg) {
  switch (msg.type) {
    case 'call.ring':        return _onRing(msg.payload)
    case 'call.answer':      return _onAnswer(msg.payload)
    case 'call.ice':         return _onICE(msg.payload)
    case 'call.hangup':      return _onHangup()
    case 'call.reject':      return _onReject()
    case 'call.unavailable': return _onUnavailable()
    case 'call.failed':      return _onCallFailed()
    case 'call.mute':        return _onMuteState(msg.payload)
    case 'call.screen_share': return _onScreenShareState(msg.payload)
  }
}

function _onMuteState(payload) {
  _s.remoteMuted = !!payload.muted
}

function _onScreenShareState(payload) {
  _s.remoteScreenSharing = !!payload.sharing
}

function _onRing(payload) {
  if (_s.phase !== 'idle') {
    _send({ type: 'call.reject', payload: { to_user_id: payload.from_user_id, conversation_id: payload.conversation_id } })
    return
  }
  _s.remoteUserId = payload.from_user_id
  _s.remoteName   = payload.from_name   || ''
  _s.remoteAvatar = payload.from_avatar || ''
  _s.convId       = payload.conversation_id
  _s.hasVideo     = payload.has_video   || false
  _pendingOffer   = payload.sdp
  _s.phase        = 'ringing'
}

async function _onAnswer(payload) {
  if (_s.phase !== 'calling' || !_pc) return
  clearTimeout(_ringTimeout); _ringTimeout = null
  try {
    await _pc.setRemoteDescription(new RTCSessionDescription({ type: 'answer', sdp: payload.sdp }))
    await _applyPendingCandidates()
    _s.phase = 'active'
  } catch {
    _failCall()
  }
}

// Negotiation failed after signaling was already underway (as opposed to
// hangup/reject, which are deliberate user actions) — notify the other side,
// tear down locally, and surface an error instead of leaving the UI hung.
function _failCall() {
  clearTimeout(_ringTimeout); _ringTimeout = null
  _send({ type: 'call.failed', payload: { to_user_id: _s.remoteUserId, conversation_id: _s.convId } })
  _teardown()
  _s.errorMsg = 'failed'
  _s.phase = 'ended'
  setTimeout(_reset, 3000)
}

function _onCallFailed() {
  clearTimeout(_ringTimeout); _ringTimeout = null
  _teardown()
  _s.errorMsg = 'failed'
  _s.phase = 'ended'
  setTimeout(_reset, 3000)
}

async function _onICE(payload) {
  if (!_pc || !_pc.remoteDescription) {
    _pendingCandidates.push(payload.candidate)
    return
  }
  try {
    const c = typeof payload.candidate === 'string' ? JSON.parse(payload.candidate) : payload.candidate
    await _pc.addIceCandidate(new RTCIceCandidate(c))
  } catch {}
}

function _onHangup()      { endCall(false) }
function _onReject()      { clearTimeout(_ringTimeout); _ringTimeout = null; _teardown(); _s.errorMsg = 'rejected'; _s.phase = 'ended'; setTimeout(_reset, 3000) }
function _onUnavailable() { clearTimeout(_ringTimeout); _ringTimeout = null; _teardown(); _s.errorMsg = 'unavailable'; _s.phase = 'ended'; setTimeout(_reset, 3000) }

async function acceptCall() {
  // _s.accepting guards against re-entrancy: getUserMedia can take several
  // seconds (permission prompt, camera init) with no other visual feedback,
  // and phase stays 'ringing' throughout — so an impatient second click on
  // Accept would otherwise start a second, concurrent negotiation that races
  // the first one for the shared _pc/_pendingOffer module state.
  if (_s.phase !== 'ringing' || !_pendingOffer || _s.accepting) return
  _s.accepting = true

  try {
    let stream, hasVideo, iceServers
    try {
      ;[{ stream, hasVideo }, iceServers] = await Promise.all([_getMedia(_s.hasVideo), _loadIceServers()])
    } catch {
      _s.errorMsg = 'no_mic'
      rejectCall()
      return
    }

    // Caller may have hung up while getUserMedia was pending
    if (_s.phase !== 'ringing') {
      stream.getTracks().forEach(t => t.stop())
      return
    }

    _localStream = stream
    _s.hasVideo  = hasVideo

    try {
      if (typeof _onLocalStream === 'function') _onLocalStream(_localStream)

      _setupPC(iceServers)
      _localStream.getTracks().forEach(t => _pc.addTrack(t, _localStream))

      await _pc.setRemoteDescription(new RTCSessionDescription({ type: 'offer', sdp: _pendingOffer }))
      _pendingOffer = null
      await _applyPendingCandidates()

      const answer = await _pc.createAnswer()
      await _pc.setLocalDescription(answer)

      _send({
        type: 'call.answer',
        payload: { to_user_id: _s.remoteUserId, conversation_id: _s.convId, sdp: answer.sdp },
      })

      _s.phase = 'active'
    } catch {
      _failCall()
    }
  } finally {
    _s.accepting = false
  }
}

function rejectCall() {
  if (_s.phase !== 'ringing') return
  _send({ type: 'call.reject', payload: { to_user_id: _s.remoteUserId, conversation_id: _s.convId } })
  _teardown()
  _reset()
}

function endCall(sendMsg = true) {
  if (_s.phase === 'idle') return
  clearTimeout(_ringTimeout); _ringTimeout = null
  if (sendMsg && _s.remoteUserId && _s.convId) {
    _send({ type: 'call.hangup', payload: { to_user_id: _s.remoteUserId, conversation_id: _s.convId } })
  }
  _teardown()
  _s.phase = 'ended'
  setTimeout(_reset, 1500)
}

function toggleMute() {
  if (!_localStream) return
  _localStream.getAudioTracks().forEach(t => { t.enabled = !t.enabled })
  _s.isMuted = !_s.isMuted
  _send({ type: 'call.mute', payload: { to_user_id: _s.remoteUserId, conversation_id: _s.convId, muted: _s.isMuted } })
}

function toggleCamera() {
  if (!_localStream) return
  _localStream.getVideoTracks().forEach(t => { t.enabled = !t.enabled })
  _s.isCameraOff = !_s.isCameraOff
}

/**
 * Share screen by replacing the outbound camera video track (same transceiver — no renegotiation).
 * Only for established video calls. PiP keeps showing the local camera.
 */
async function toggleScreenShare() {
  if (_s.phase !== 'active' || !_pc || !_s.hasVideo) return
  const sender = _pc.getSenders().find(s => s.track?.kind === 'video')
  if (!sender) return

  if (_s.isScreenSharing) {
    if (_screenStream) {
      _screenStream.getTracks().forEach(t => t.stop())
      _screenStream = null
    }
    const cam = _localStream?.getVideoTracks?.()[0] ?? null
    try {
      await sender.replaceTrack(cam)
    } catch {}
    _s.isScreenSharing = false
    _send({ type: 'call.screen_share', payload: { to_user_id: _s.remoteUserId, conversation_id: _s.convId, sharing: false } })
    return
  }

  let stream
  try {
    stream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: false })
  } catch (err) {
    if (err?.name !== 'AbortError' && err?.name !== 'NotAllowedError') {
      useUIStore().error(i18n.global.t('call.screen_share_failed'))
    }
    return
  }
  const v = stream.getVideoTracks()[0]
  if (!v) {
    stream.getTracks().forEach(t => t.stop())
    return
  }
  _screenStream = stream
  v.addEventListener('ended', () => {
    if (_s.isScreenSharing) void toggleScreenShare()
  }, { once: true })
  try {
    await sender.replaceTrack(v)
    _s.isScreenSharing = true
    _send({ type: 'call.screen_share', payload: { to_user_id: _s.remoteUserId, conversation_id: _s.convId, sharing: true } })
  } catch {
    stream.getTracks().forEach(t => t.stop())
    _screenStream = null
    useUIStore().error(i18n.global.t('call.screen_share_failed'))
  }
}

// ── Composable export ────────────────────────────────────────────────────────

/** True while a 1:1 call is in progress (outgoing, ringing, or connected). */
export function isWebRTCCallBusy() {
  return _s.phase === 'calling' || _s.phase === 'ringing' || _s.phase === 'active'
}

export function useWebRTCCall() {
  return {
    state: readonly(_s),
    setSendFn,
    setRemoteStreamCallback,
    setLocalStreamCallback,
    startCall,
    handleSignal,
    acceptCall,
    rejectCall,
    endCall,
    toggleMute,
    toggleCamera,
    toggleScreenShare,
  }
}
