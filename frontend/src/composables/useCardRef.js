import { marked } from 'marked'
import { markedHighlight } from 'marked-highlight'
import hljs from 'highlight.js/lib/core'
import DOMPurify from 'dompurify'
import { useRouter } from 'vue-router'
import client from '@/api/client'
import { resolveAssetUrl } from '@/api/serverConfig'

// Register the languages most likely to appear in chat code blocks
import bash from 'highlight.js/lib/languages/bash'
import css from 'highlight.js/lib/languages/css'
import go from 'highlight.js/lib/languages/go'
import ini from 'highlight.js/lib/languages/ini'
import javascript from 'highlight.js/lib/languages/javascript'
import json from 'highlight.js/lib/languages/json'
import python from 'highlight.js/lib/languages/python'
import rust from 'highlight.js/lib/languages/rust'
import sql from 'highlight.js/lib/languages/sql'
import typescript from 'highlight.js/lib/languages/typescript'
import xml from 'highlight.js/lib/languages/xml'
import yaml from 'highlight.js/lib/languages/yaml'

hljs.registerLanguage('bash', bash)
hljs.registerLanguage('sh', bash)
hljs.registerLanguage('shell', bash)
hljs.registerLanguage('css', css)
hljs.registerLanguage('go', go)
hljs.registerLanguage('ini', ini)
hljs.registerLanguage('toml', ini)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('js', javascript)
hljs.registerLanguage('json', json)
hljs.registerLanguage('python', python)
hljs.registerLanguage('py', python)
hljs.registerLanguage('rust', rust)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('ts', typescript)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('yml', yaml)

marked.use(markedHighlight({
  langPrefix: 'hljs language-',
  highlight(code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      return hljs.highlight(code, { language: lang }).value
    }
    return hljs.highlightAuto(code).value
  },
}))

const CARD_REF_RE = /#([A-Z][A-Z0-9]*-\d+)\b/g
const URL_RE = /https?:\/\/[^\s<>"')\]]+/g

export function renderMarkdown(text) {
  const withRefs = (text || '').replace(CARD_REF_RE, (_, ref) =>
    `<span class="card-ref-link" data-card-ref="${ref}">#${ref}</span>`
  )
  const safe = DOMPurify.sanitize(marked.parse(withRefs), { ADD_ATTR: ['data-card-ref'] })
  const doc = new DOMParser().parseFromString(safe, 'text/html')
  doc.querySelectorAll('img[src]').forEach((img) => {
    const src = img.getAttribute('src') || ''
    img.setAttribute('src', resolveAssetUrl(src))
  })
  return doc.body.innerHTML
}

export function firstUrl(text) {
  const m = (text || '').match(URL_RE)
  return m?.[0] ?? null
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
