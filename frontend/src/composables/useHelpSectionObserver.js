import { onMounted, onBeforeUnmount } from 'vue'
import { useUIStore } from '@/stores/ui'

/**
 * Updates ui.helpContext based on which marked section is most visible while scrolling.
 * Mark sections with data-help-context="settings.profile" (or any CONTEXT_HELP key prefix).
 */
export function useHelpSectionObserver(rootRef) {
  const ui = useUIStore()
  let observer = null

  onMounted(() => {
    const root = rootRef.value
    if (!root) return

    const sections = [...root.querySelectorAll('[data-help-context]')]
    if (!sections.length) return

    observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter(e => e.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio)
        const top = visible[0]?.target?.dataset?.helpContext
        if (top) ui.setHelpContext(top)
      },
      { root: null, rootMargin: '-35% 0px -45% 0px', threshold: [0, 0.15, 0.35, 0.6] },
    )

    for (const el of sections) observer.observe(el)

    const initial = sections[0]?.dataset?.helpContext
    if (initial) ui.setHelpContext(initial)
  })

  onBeforeUnmount(() => {
    observer?.disconnect()
    observer = null
    ui.setHelpContext(null)
  })
}
