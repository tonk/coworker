/** Avoid circular imports between useWebRTCCall and useLiveKitGroupCall. */

let _liveKitActive = false

export function setLiveKitCallActive(v) {
  _liveKitActive = !!v
}

export function isLiveKitCallActive() {
  return _liveKitActive
}
