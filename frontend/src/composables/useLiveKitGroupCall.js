/**
 * LiveKit-based group video/audio for conversations with 3+ members.
 * Uses GET /conversations/:id/livekit-token from the backend.
 */

import { reactive, readonly } from 'vue'
import { Room, RoomEvent, Track } from 'livekit-client'
import { messagesApi } from '@/api/messages'
import { useCallSettings } from './useCallSettings'
import { setLiveKitCallActive } from './callsGate'
import { isWebRTCCallBusy } from './useWebRTCCall'

const _s = reactive({
  phase: 'idle', // idle | connecting | active | ended
  convId: null,
  title: '',
  hasVideo: false,
  isMuted: false,
  isCameraOff: false,
  errorMsg: '',
})

/** @type {Room | null} */
let _room = null

const _tiles = reactive([])
let _profilesByIdentity = {}

function _lkAudioOptions() {
  const { audioDeviceId } = useCallSettings()
  return audioDeviceId.value ? { deviceId: audioDeviceId.value } : {}
}

function _lkVideoOptions() {
  const { videoDeviceId } = useCallSettings()
  return videoDeviceId.value ? { deviceId: videoDeviceId.value } : {}
}

function _syncLocalFlags() {
  if (!_room) return
  const lp = _room.localParticipant
  const mic = lp.getTrackPublication(Track.Source.Microphone)
  const cam = lp.getTrackPublication(Track.Source.Camera)
  _s.isMuted = mic ? mic.isMuted : false
  _s.isCameraOff = cam ? cam.isMuted : true
}

function _syncTiles() {
  _tiles.splice(0, _tiles.length)
  if (!_room) return
  const lp = _room.localParticipant
  const localPub = lp.getTrackPublication(Track.Source.Camera)
  const localCam = localPub?.videoTrack ?? null
  const localProfile = _profilesByIdentity[String(lp.identity)] || {}
  _tiles.push({
    sid: 'local',
    identity: lp.identity,
    name: localProfile.name || '',
    avatar: localProfile.avatar || '',
    videoTrack: localCam,
    cameraOff: localPub ? localPub.isMuted : true,
    isLocal: true,
  })
  const remotes = [..._room.remoteParticipants.values()].sort((a, b) =>
    String(a.identity).localeCompare(String(b.identity))
  )
  for (const p of remotes) {
    const pub = p.getTrackPublication(Track.Source.Camera)
    const vt = pub?.videoTrack ?? null
    _tiles.push({
      sid: p.sid,
      identity: p.identity,
      name: _profilesByIdentity[String(p.identity)]?.name || p.name || '',
      avatar: _profilesByIdentity[String(p.identity)]?.avatar || '',
      videoTrack: vt,
      cameraOff: pub ? pub.isMuted : true,
      isLocal: false,
    })
  }
}

function _wireRoom(room) {
  const bump = () => {
    _syncTiles()
    _syncLocalFlags()
  }
  room
    .on(RoomEvent.ParticipantConnected, bump)
    .on(RoomEvent.ParticipantDisconnected, bump)
    .on(RoomEvent.TrackSubscribed, bump)
    .on(RoomEvent.TrackUnsubscribed, bump)
    .on(RoomEvent.LocalTrackPublished, bump)
    .on(RoomEvent.LocalTrackUnpublished, bump)
    .on(RoomEvent.TrackMuted, bump)
    .on(RoomEvent.TrackUnmuted, bump)
    .on(RoomEvent.Disconnected, () => {
      void leaveGroupCall()
    })
}

async function _disconnectRoom() {
  setLiveKitCallActive(false)
  const r = _room
  _room = null
  _tiles.splice(0, _tiles.length)
  if (!r) return
  try {
    r.removeAllListeners()
    await r.disconnect()
  } catch {}
}

function _clearToIdle() {
  _s.phase = 'idle'
  _s.convId = null
  _s.title = ''
  _s.hasVideo = false
  _s.errorMsg = ''
  _s.isMuted = false
  _s.isCameraOff = false
  _profilesByIdentity = {}
}

/**
 * @param {number} convId
 * @param {string} [title] header label
 * @param {{ identity: string|number, name?: string, avatar?: string }[]} [participantProfiles]
 */
async function joinGroupCall(convId, title = '', participantProfiles = []) {
  if (_s.phase === 'connecting' || _s.phase === 'active') return
  if (isWebRTCCallBusy()) return
  setLiveKitCallActive(true)
  _s.phase = 'connecting'
  _s.convId = convId
  _s.title = title || ''
  _s.errorMsg = ''
  _profilesByIdentity = {}
  for (const p of participantProfiles || []) {
    if (!p || p.identity === undefined || p.identity === null) continue
    _profilesByIdentity[String(p.identity)] = {
      name: p.name || '',
      avatar: p.avatar || '',
    }
  }

  let token
  let url
  try {
    const res = await messagesApi.getLiveKitToken(convId)
    token = res.data?.token
    url = res.data?.url
    if (!token || !url) {
      _s.errorMsg = 'livekit_unavailable'
      setLiveKitCallActive(false)
      _s.phase = 'idle'
      _s.convId = null
      _s.title = ''
      return
    }
  } catch (e) {
    _s.errorMsg =
      e?.response?.status === 503 ? 'livekit_unavailable' : 'livekit_connect_failed'
    setLiveKitCallActive(false)
    _s.phase = 'idle'
    _s.convId = null
    _s.title = ''
    return
  }

  const room = new Room({ adaptiveStream: true, dynacast: true })
  _room = room
  _wireRoom(room)

  try {
    await room.connect(url, token)
    try {
      await room.startAudio()
    } catch {}
    await room.localParticipant.setMicrophoneEnabled(true, _lkAudioOptions())
    let hasVideo = false
    try {
      await room.localParticipant.setCameraEnabled(true, _lkVideoOptions())
      hasVideo = true
    } catch {
      hasVideo = false
    }
    _s.hasVideo = hasVideo
    _s.phase = 'active'
    _syncTiles()
    _syncLocalFlags()
  } catch {
    await _disconnectRoom()
    _s.errorMsg = 'livekit_connect_failed'
    _clearToIdle()
  }
}

async function leaveGroupCall() {
  if (_s.phase !== 'connecting' && _s.phase !== 'active') return
  await _disconnectRoom()
  _s.phase = 'ended'
  setTimeout(() => {
    if (_s.phase === 'ended') _clearToIdle()
  }, 400)
}

async function toggleMute() {
  if (!_room || _s.phase !== 'active') return
  const next = !_s.isMuted
  await _room.localParticipant.setMicrophoneEnabled(!next, _lkAudioOptions())
  _syncLocalFlags()
}

async function toggleCamera() {
  if (!_room || _s.phase !== 'active') return
  const pub = _room.localParticipant.getTrackPublication(Track.Source.Camera)
  const wantOn = pub ? pub.isMuted : !_s.hasVideo
  try {
    await _room.localParticipant.setCameraEnabled(wantOn, _lkVideoOptions())
    _s.hasVideo = !!_room.localParticipant.getTrackPublication(Track.Source.Camera)?.videoTrack
  } catch {
    _s.hasVideo = false
  }
  _syncLocalFlags()
  _syncTiles()
}

export function useLiveKitGroupCall() {
  return {
    state: readonly(_s),
    tiles: readonly(_tiles),
    joinGroupCall,
    leaveGroupCall,
    toggleMute,
    toggleCamera,
  }
}
