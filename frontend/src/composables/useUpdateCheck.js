import { ref } from 'vue'

const GITHUB_REPO = 'tonk/warmdesk'
const CACHE_KEY = 'update_check'
const CACHE_TTL = 60 * 60 * 1000 // 1 hour

function parse(v) {
  return v.replace(/^v/, '').split('.').map(Number)
}

function isNewer(current, latest) {
  const [ma, mi, pa] = parse(current)
  const [mb, mi2, pb] = parse(latest)
  if (mb !== ma) return mb > ma
  if (mi2 !== mi) return mi2 > mi
  return pb > pa
}

function detectPlatform() {
  const plat = (navigator.platform || '').toLowerCase()
  if (plat.includes('win')) return 'windows'
  if (plat.includes('mac')) return 'macos'
  if (plat.includes('linux')) {
    if (plat.includes('aarch64') || plat.includes('arm64')) return 'linux-arm64'
    return 'linux'
  }
  return null
}

function pickAsset(assets, tag) {
  if (!window.__TAURI_INTERNALS__ || !assets) return null
  const platform = detectPlatform()
  if (!platform) return null

  const candidates = []

  if (platform === 'macos') {
    candidates.push(`WarmDesk-${tag}-universal.dmg`)
  } else if (platform === 'windows') {
    candidates.push(`WarmDesk-${tag}-x64-portable.zip`)
    candidates.push(`WarmDesk-${tag}-x64-setup.exe`)
  } else if (platform === 'linux') {
    candidates.push(`WarmDesk-${tag}-x86_64.AppImage`)
    candidates.push(`WarmDesk-${tag}-amd64.deb`)
    candidates.push(`WarmDesk-${tag}-x86_64.rpm`)
    candidates.push(`warmdesk-${tag}-linux-amd64.tar.gz`)
  } else if (platform === 'linux-arm64') {
    candidates.push(`warmdesk-${tag}-linux-arm64.tar.gz`)
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
    const cached = sessionStorage.getItem(CACHE_KEY)
    if (cached) {
      const { tag, url, assets, expires } = JSON.parse(cached)
      if (Date.now() < expires) {
        if (isNewer(currentVersion, tag)) {
          latestVersion.value = tag.replace(/^v/, '')
          releaseUrl.value = url
          downloadUrl.value = pickAsset(assets, tag)
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
        downloadUrl.value = pickAsset(assets, tag)
        updateAvailable.value = true
      }
    } catch {
      // Silently ignore network errors / rate limits
    }
  }

  return { updateAvailable, latestVersion, releaseUrl, downloadUrl, check }
}
