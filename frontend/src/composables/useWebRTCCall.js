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
  errorMsg:    '',
})

let _pc                = null
let _localStream       = null
let _remoteStream      = null
let _pendingCandidates = []
let _pendingOffer      = null   // SDP stored while ringing
let _sendFn            = null   // set by App.vue
let _onRemoteStream    = null   // set by ActiveCallBar
let _onLocalStream     = null   // set by ActiveCallBar (self-preview)

const ICE_SERVERS = [
  { urls: 'stun:stun.l.google.com:19302' },
  { urls: 'stun:stun1.l.google.com:19302' },
]

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
  _s.errorMsg     = ''
  _pendingCandidates = []
  _pendingOffer   = null
  _remoteStream   = null
}

function _teardown() {
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

function _setupPC() {
  _pc = new RTCPeerConnection({ iceServers: ICE_SERVERS })

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

async function startCall(userId, userName, avatar, convId) {
  if (_s.phase !== 'idle') return

  _s.remoteUserId = userId
  _s.remoteName   = userName || ''
  _s.remoteAvatar = avatar  || ''
  _s.convId       = convId
  _s.phase        = 'calling'
  _s.errorMsg     = ''

  let stream, hasVideo
  try {
    ;({ stream, hasVideo } = await _getMedia(true))
  } catch {
    _s.errorMsg = 'no_mic'
    _reset()
    return
  }

  _localStream  = stream
  _s.hasVideo   = hasVideo
  if (typeof _onLocalStream === 'function') _onLocalStream(_localStream)

  _setupPC()
  _localStream.getTracks().forEach(t => _pc.addTrack(t, _localStream))

  const offer = await _pc.createOffer()
  await _pc.setLocalDescription(offer)

  _send({
    type: 'call.offer',
    payload: {
      to_user_id:      userId,
      conversation_id: convId,
      sdp:             offer.sdp,
      has_video:       hasVideo,
    },
  })
}

function handleSignal(msg) {
  switch (msg.type) {
    case 'call.ring':        return _onRing(msg.payload)
    case 'call.answer':      return _onAnswer(msg.payload)
    case 'call.ice':         return _onICE(msg.payload)
    case 'call.hangup':      return _onHangup()
    case 'call.reject':      return _onReject()
    case 'call.unavailable': return _onUnavailable()
  }
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
  try {
    await _pc.setRemoteDescription(new RTCSessionDescription({ type: 'answer', sdp: payload.sdp }))
    await _applyPendingCandidates()
    _s.phase = 'active'
  } catch {}
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
function _onReject()      { _teardown(); _s.phase = 'ended'; setTimeout(_reset, 2000) }
function _onUnavailable() { _teardown(); _s.errorMsg = 'unavailable'; _s.phase = 'ended'; setTimeout(_reset, 3000) }

async function acceptCall() {
  if (_s.phase !== 'ringing' || !_pendingOffer) return

  let stream, hasVideo
  try {
    ;({ stream, hasVideo } = await _getMedia(_s.hasVideo))
  } catch {
    _s.errorMsg = 'no_mic'
    rejectCall()
    return
  }

  _localStream = stream
  _s.hasVideo  = hasVideo
  if (typeof _onLocalStream === 'function') _onLocalStream(_localStream)

  _setupPC()
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
}

function rejectCall() {
  if (_s.phase !== 'ringing') return
  _send({ type: 'call.reject', payload: { to_user_id: _s.remoteUserId, conversation_id: _s.convId } })
  _teardown()
  _reset()
}

function endCall(sendMsg = true) {
  if (_s.phase === 'idle') return
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
}

function toggleCamera() {
  if (!_localStream) return
  _localStream.getVideoTracks().forEach(t => { t.enabled = !t.enabled })
  _s.isCameraOff = !_s.isCameraOff
}

// ── Composable export ────────────────────────────────────────────────────────

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
  }
}
