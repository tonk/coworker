import { ref } from 'vue'

const GITHUB_REPO = 'tonk/coworker'
const CACHE_KEY = 'update_check'

// Compare two semver strings (strips leading 'v'). Returns true if b > a.
function isNewer(current, latest) {
  const parse = v => v.replace(/^v/, '').split('.').map(Number)
  const [ma, mi, pa] = parse(current)
  const [mb, mi2, pb] = parse(latest)
  if (mb !== ma) return mb > ma
  if (mi2 !== mi) return mi2 > mi
  return pb > pa
}

export function useUpdateCheck() {
  const updateAvailable = ref(false)
  const latestVersion = ref(null)
  const releaseUrl = ref(null)

  async function check(currentVersion) {
    // Use cached result within the same session
    const cached = sessionStorage.getItem(CACHE_KEY)
    if (cached) {
      const { tag, url } = JSON.parse(cached)
      if (isNewer(currentVersion, tag)) {
        latestVersion.value = tag.replace(/^v/, '')
        releaseUrl.value = url
        updateAvailable.value = true
      }
      return
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
      sessionStorage.setItem(CACHE_KEY, JSON.stringify({ tag, url }))
      if (tag && isNewer(currentVersion, tag)) {
        latestVersion.value = tag.replace(/^v/, '')
        releaseUrl.value = url
        updateAvailable.value = true
      }
    } catch {
      // Silently ignore network errors / rate limits
    }
  }

  return { updateAvailable, latestVersion, releaseUrl, check }
}
