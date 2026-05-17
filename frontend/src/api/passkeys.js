import client from './client'

export const passkeysApi = {
  registerBegin: () => client.get('/auth/passkey/register/begin'),
  registerFinish: (data) => client.post('/auth/passkey/register/finish', data),
  loginBegin: () => client.post('/auth/passkey/login/begin'),
  loginFinish: (data) => client.post('/auth/passkey/login/finish', data),
  list: () => client.get('/auth/passkeys'),
  delete: (id) => client.delete(`/auth/passkeys/${id}`),
}

// Convert ArrayBuffer / TypedArray to base64url string.
export function bufferToBase64url(buffer) {
  const bytes = new Uint8Array(buffer)
  let str = ''
  for (const b of bytes) str += String.fromCharCode(b)
  return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
}

// Convert base64url string to ArrayBuffer.
export function base64urlToBuffer(b64url) {
  const b64 = b64url.replace(/-/g, '+').replace(/_/g, '/')
  const bin = atob(b64)
  return Uint8Array.from(bin, (c) => c.charCodeAt(0)).buffer
}

// Decode the raw PublicKeyCredentialCreationOptions coming from the server
// (challenges and IDs arrive as base64url strings and must be converted to ArrayBuffers).
export function decodeCreationOptions(opts) {
  const pk = { ...opts }
  pk.challenge = base64urlToBuffer(pk.challenge)
  pk.user = { ...pk.user, id: base64urlToBuffer(pk.user.id) }
  if (pk.excludeCredentials) {
    pk.excludeCredentials = pk.excludeCredentials.map((c) => ({
      ...c,
      id: base64urlToBuffer(c.id),
    }))
  }
  return pk
}

// Decode the raw PublicKeyCredentialRequestOptions from the server.
export function decodeRequestOptions(opts) {
  const pk = { ...opts }
  pk.challenge = base64urlToBuffer(pk.challenge)
  if (pk.allowCredentials) {
    pk.allowCredentials = pk.allowCredentials.map((c) => ({
      ...c,
      id: base64urlToBuffer(c.id),
    }))
  }
  return pk
}

// Serialize a PublicKeyCredential (registration) for sending to the server.
export function serializeRegistrationCredential(cred) {
  return {
    id: cred.id,
    rawId: bufferToBase64url(cred.rawId),
    type: cred.type,
    response: {
      clientDataJSON: bufferToBase64url(cred.response.clientDataJSON),
      attestationObject: bufferToBase64url(cred.response.attestationObject),
      transports: cred.response.getTransports?.() ?? [],
    },
  }
}

// Serialize a PublicKeyCredential (authentication) for sending to the server.
export function serializeAuthenticationCredential(cred) {
  return {
    id: cred.id,
    rawId: bufferToBase64url(cred.rawId),
    type: cred.type,
    response: {
      clientDataJSON: bufferToBase64url(cred.response.clientDataJSON),
      authenticatorData: bufferToBase64url(cred.response.authenticatorData),
      signature: bufferToBase64url(cred.response.signature),
      userHandle: cred.response.userHandle ? bufferToBase64url(cred.response.userHandle) : undefined,
    },
  }
}
