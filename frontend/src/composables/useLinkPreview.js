import { reactive } from 'vue'
import client from '@/api/client'

// Module-level cache so previews survive component re-renders
const cache = reactive({})

export function useLinkPreview() {
  async function fetchPreview(url) {
    if (!url || url in cache) return
    cache[url] = null  // sentinel: fetch in progress
    try {
      const { data } = await client.get('/link-preview', { params: { url } })
      // Only cache if there's something useful to show
      cache[url] = (data.title || data.description || data.image) ? data : false
    } catch {
      cache[url] = false
    }
  }

  return { cache, fetchPreview }
}
