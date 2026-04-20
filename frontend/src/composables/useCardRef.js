import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useRouter } from 'vue-router'
import client from '@/api/client'

// Matches card references like #PRJ-42 or #API2-7
const CARD_REF_RE = /#([A-Z][A-Z0-9]*-\d+)\b/g

export function renderMarkdown(text) {
  const withRefs = (text || '').replace(CARD_REF_RE, (_, ref) =>
    `<span class="card-ref-link" data-card-ref="${ref}">#${ref}</span>`
  )
  return DOMPurify.sanitize(marked.parse(withRefs), { ADD_ATTR: ['data-card-ref'] })
}

export function useCardRef() {
  const router = useRouter()

  async function handleCardRefClick(event) {
    const el = event.target.closest('[data-card-ref]')
    if (!el) return
    event.preventDefault()
    const ref = el.dataset.cardRef
    const newTab = event.ctrlKey || event.metaKey || event.button === 1
    try {
      const { data } = await client.get(`/cards/resolve/${ref}`)
      const location = { name: 'board', params: { slug: data.project_slug }, query: { card: String(data.id) } }
      if (newTab) {
        window.open(router.resolve(location).href, '_blank')
      } else {
        router.push(location)
      }
    } catch {}
  }

  return { handleCardRefClick }
}
