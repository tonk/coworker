/**
 * LiveKit-based group video/audio for conversations with 3+ members.
 * Uses GET /conversations/:id/livekit-token from the backend.
 *
 * livekit-client is loaded dynamically so it ships in a separate chunk (~lazy).
 */

import { reactive, readonly } from 'vue'
import { messagesApi } from '@/api/messages'
import { useCallSettings } from './useCallSettings'
import { setLiveKitCallActive } from './callsGate'
import { isWebRTCCallBusy } from './useWebRTCCall'
import { useUIStore } from '@/stores/ui'
import { i18n } from '@/i18n'

/** Loaded on first group join; keeps initial bundle small. */
let _lkMod = null
async function _ensureLk() {
  if (!_lkMod) _lkMod = await import('livekit-client')
  return _lkMod
}

const _s = reactive({
  phase: 'idle', // idle | connecting | active | ended
  convId: null,
  title: '',
  hasVideo: false,
  isMuted: false,
  isCameraOff: false,
  isScreenSharing: false,
  errorMsg: '',
})

const _invite = reactive({
  pending: false,
  convId: null,
  convName: '',
  fromName: '',
  fromAvatar: '',
})

let _sendFn = null
function setSendFn(fn) { _sendFn = fn }
function _send(msg) { if (_sendFn) _sendFn(msg) }

/** @type {import('livekit-client').Room | null} */
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
  if (!_room || !_lkMod) return
  const { Track } = _lkMod
  const lp = _room.localParticipant
  const mic = lp.getTrackPublication(Track.Source.Microphone)
  const cam = lp.getTrackPublication(Track.Source.Camera)
  const ss = lp.getTrackPublication(Track.Source.ScreenShare)
  _s.isMuted = mic ? mic.isMuted : false
  _s.isCameraOff = cam ? cam.isMuted : true
  _s.isScreenSharing = !!(ss?.track && !ss.isMuted)
}

function _pushParticipantTiles(participant, isLocal) {
  if (!_lkMod) return
  const { Track } = _lkMod
  const sidPrefix = isLocal ? 'local' : participant.sid
  const prof = _profilesByIdentity[String(participant.identity)] || {}
  const name = isLocal ? (prof.name || '') : (prof.name || participant.name || '')
  const avatar = prof.avatar || ''

  const camPub = participant.getTrackPublication(Track.Source.Camera)
  const camTrack = camPub?.videoTrack ?? null
  _tiles.push({
    sid: `${sidPrefix}-cam`,
    identity: participant.identity,
    name,
    avatar,
    videoTrack: camTrack,
    cameraOff: camPub ? camPub.isMuted : true,
    isLocal,
    isScreenShare: false,
  })

  const screenPub = participant.getTrackPublication(Track.Source.ScreenShare)
  const scTrack = screenPub?.videoTrack ?? null
  if (scTrack) {
    _tiles.push({
      sid: `${sidPrefix}-scr`,
      identity: participant.identity,
      name,
      avatar,
      videoTrack: scTrack,
      cameraOff: false,
      isLocal,
      isScreenShare: true,
    })
  }
}

function _syncTiles() {
  _tiles.splice(0, _tiles.length)
  if (!_room) return
  const lp = _room.localParticipant
  _pushParticipantTiles(lp, true)
  const remotes = [..._room.remoteParticipants.values()].sort((a, b) =>
    String(a.identity).localeCompare(String(b.identity))
  )
  for (const p of remotes) _pushParticipantTiles(p, false)
}

function _wireRoom(room, RoomEvent) {
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
    try {
      await r.localParticipant.setScreenShareEnabled(false)
    } catch {}
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
  _s.isScreenSharing = false
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

  let lk
  try {
    lk = await _ensureLk()
  } catch {
    setLiveKitCallActive(false)
    _s.phase = 'idle'
    _s.convId = null
    _s.title = ''
    _s.errorMsg = 'livekit_connect_failed'
    return
  }

  const room = new lk.Room({ adaptiveStream: true, dynacast: true })
  _room = room
  _wireRoom(room, lk.RoomEvent)

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
  if (!_room || _s.phase !== 'active' || !_lkMod) return
  const { Track } = _lkMod
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

async function toggleScreenShare() {
  if (!_room || _s.phase !== 'active') return
  const next = !_s.isScreenSharing
  try {
    await _room.localParticipant.setScreenShareEnabled(next)
  } catch {
    useUIStore().error(i18n.global.t('call.screen_share_failed'))
    return
  }
  _syncLocalFlags()
  _syncTiles()
}

/** Send a call.group_invite for the given user IDs via the personal WebSocket. */
function inviteUsers(toUserIds) {
  if (!_s.convId || !toUserIds?.length) return
  _send({
    type: 'call.group_invite',
    payload: { to_user_ids: toUserIds, conversation_id: _s.convId },
  })
}

/** Send a group invite without requiring an active LiveKit call (used for 1:1 → group upgrade). */
function sendGroupInvite(convId, toUserIds) {
  if (!convId || !toUserIds?.length) return
  _send({
    type: 'call.group_invite',
    payload: { to_user_ids: toUserIds, conversation_id: convId },
  })
}

/** Called by App.vue when a call.group_invite message arrives on the personal WS. */
function handleGroupInvite(payload) {
  if (_invite.pending) return // already have a pending invite
  _invite.pending = true
  _invite.convId = payload.conversation_id
  _invite.convName = payload.conv_name || ''
  _invite.fromName = payload.from_name || ''
  _invite.fromAvatar = payload.from_avatar || ''
}

async function acceptGroupInvite() {
  if (!_invite.pending) return
  const { convId, convName, fromName } = _invite
  _invite.pending = false
  _invite.convId = null
  _invite.convName = ''
  _invite.fromName = ''
  _invite.fromAvatar = ''
  await joinGroupCall(convId, convName || fromName, [])
}

function declineGroupInvite() {
  _invite.pending = false
  _invite.convId = null
  _invite.convName = ''
  _invite.fromName = ''
  _invite.fromAvatar = ''
}

export function useLiveKitGroupCall() {
  return {
    state: readonly(_s),
    tiles: readonly(_tiles),
    invite: readonly(_invite),
    setSendFn,
    joinGroupCall,
    leaveGroupCall,
    toggleMute,
    toggleCamera,
    toggleScreenShare,
    inviteUsers,
    sendGroupInvite,
    handleGroupInvite,
    acceptGroupInvite,
    declineGroupInvite,
  }
}
