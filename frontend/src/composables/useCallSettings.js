/**
 * useCallSettings — singleton for preferred audio/video devices.
 * Persists selections to localStorage. Used by useWebRTCCall and CallSettingsDropdown.
 */

import { ref, reactive } from 'vue'

const LS = {
  audioIn:  'call.audioDeviceId',
  videoIn:  'call.videoDeviceId',
  audioOut: 'call.audioOutputId',
}

const devices = reactive({ audioInputs: [], videoInputs: [], audioOutputs: [] })
const audioDeviceId = ref(localStorage.getItem(LS.audioIn)  || '')
const videoDeviceId = ref(localStorage.getItem(LS.videoIn)  || '')
const audioOutputId = ref(localStorage.getItem(LS.audioOut) || '')

let _loaded = false

async function loadDevices(force = false) {
  if (_loaded && !force) return
  // Request permission briefly to get labelled device names
  try {
    const s = await navigator.mediaDevices.getUserMedia({ audio: true, video: true })
    s.getTracks().forEach(t => t.stop())
  } catch {
    try {
      const s = await navigator.mediaDevices.getUserMedia({ audio: true })
      s.getTracks().forEach(t => t.stop())
    } catch {}
  }
  try {
    const list = await navigator.mediaDevices.enumerateDevices()
    devices.audioInputs  = list.filter(d => d.kind === 'audioinput')
    devices.videoInputs  = list.filter(d => d.kind === 'videoinput')
    devices.audioOutputs = list.filter(d => d.kind === 'audiooutput')
    _loaded = true
  } catch {}
}

function getMediaConstraints(wantVideo) {
  return {
    audio: audioDeviceId.value ? { deviceId: { ideal: audioDeviceId.value } } : true,
    video: wantVideo
      ? (videoDeviceId.value ? { deviceId: { ideal: videoDeviceId.value } } : true)
      : false,
  }
}

function setAudioDevice(id) {
  audioDeviceId.value = id
  localStorage.setItem(LS.audioIn, id)
}
function setVideoDevice(id) {
  videoDeviceId.value = id
  localStorage.setItem(LS.videoIn, id)
}
function setAudioOutput(id) {
  audioOutputId.value = id
  localStorage.setItem(LS.audioOut, id)
}

export function useCallSettings() {
  return { devices, audioDeviceId, videoDeviceId, audioOutputId, loadDevices, getMediaConstraints, setAudioDevice, setVideoDevice, setAudioOutput }
}
