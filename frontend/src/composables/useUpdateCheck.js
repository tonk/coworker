import { ref } from 'vue'

const GITHUB_REPO = 'tonk/warmdesk'
const CACHE_KEY = 'update_check'
const CACHE_TTL = 60 * 60 * 1000 // 1 hour

function parse(v) {
  return v.replace(/^v/, '').split('.').map(Number)
}

export function isNewer(current, latest) {
  const [ma, mi, pa] = parse(current)
  const [mb, mi2, pb] = parse(latest)
  if (mb !== ma) return mb > ma
  if (mi2 !== mi) return mi2 > mi
  return pb > pa
}

let _installMethod = null

export async function detectInstallMethod() {
  if (_installMethod) return _installMethod

  if (window.__TAURI_INTERNALS__) {
    try {
      const { invoke } = await import('@tauri-apps/api/core')
      _installMethod = await invoke('installation_method')
      return _installMethod
    } catch {}
  }

  const plat = (navigator.platform || '').toLowerCase()
  if (plat.includes('win')) _installMethod = 'windows'
  else if (plat.includes('mac')) _installMethod = 'dmg'
  else if (plat.includes('linux')) _installMethod = 'portable'
  else _installMethod = 'unknown'
  return _installMethod
}

export function pickAsset(assets, tag, method) {
  if (!window.__TAURI_INTERNALS__ || !assets || !method) return null

  const candidates = []

  if (method === 'dmg') {
    candidates.push(`WarmDesk-${tag}-universal.dmg`)
  } else if (method === 'windows') {
    candidates.push(`WarmDesk-${tag}-x64-portable.zip`)
    candidates.push(`WarmDesk-${tag}-x64-setup.exe`)
  } else if (method === 'appimage') {
    candidates.push(`WarmDesk-${tag}-x86_64.AppImage`)
  } else if (method === 'deb') {
    candidates.push(`WarmDesk-${tag}-amd64.deb`)
  } else if (method === 'rpm') {
    candidates.push(`WarmDesk-${tag}-x86_64.rpm`)
  } else if (method === 'portable') {
    candidates.push(`warmdesk-${tag}-linux-amd64.tar.gz`)
  }

  for (const name of candidates) {
    const asset = assets.find(a => a.name === name)
    if (asset) return asset.browser_download_url
  }
  return null
}

export function useUpdateCheck() {
  const updateAvailable = ref(false)
  const latestVersion = ref(null)
  const releaseUrl = ref(null)
  const downloadUrl = ref(null)

  async function check(currentVersion) {
    const method = await detectInstallMethod()

    const cached = sessionStorage.getItem(CACHE_KEY)
    if (cached) {
      const { tag, url, assets, expires } = JSON.parse(cached)
      if (Date.now() < expires) {
        if (isNewer(currentVersion, tag)) {
          latestVersion.value = tag.replace(/^v/, '')
          releaseUrl.value = url
          downloadUrl.value = pickAsset(assets, tag, method)
          updateAvailable.value = true
        }
        return
      }
    }

    try {
      const res = await fetch(
        `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`,
        { headers: { Accept: 'application/vnd.github+json' } }
      )
      if (!res.ok) return
      const data = await res.json()
      const tag = data.tag_name
      const url = data.html_url
      const assets = data.assets || []
      sessionStorage.setItem(CACHE_KEY, JSON.stringify({ tag, url, assets, expires: Date.now() + CACHE_TTL }))
      if (tag && isNewer(currentVersion, tag)) {
        latestVersion.value = tag.replace(/^v/, '')
        releaseUrl.value = url
        downloadUrl.value = pickAsset(assets, tag, method)
        updateAvailable.value = true
      }
    } catch {
      // Silently ignore network errors / rate limits
    }
  }

  return { updateAvailable, latestVersion, releaseUrl, downloadUrl, check }
}
