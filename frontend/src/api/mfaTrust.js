const MFA_TRUST_KEY = 'mfa_trust_token'

export function getMfaTrustToken() {
  return localStorage.getItem(MFA_TRUST_KEY) || ''
}

export function setMfaTrustToken(token) {
  if (token) localStorage.setItem(MFA_TRUST_KEY, token)
  else localStorage.removeItem(MFA_TRUST_KEY)
}

export function clearMfaTrustToken() {
  localStorage.removeItem(MFA_TRUST_KEY)
}
