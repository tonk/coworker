/**
 * Applies user preferences (font, font-size, sidebar position, theme) to the document.
 * Falls back to global system defaults when the user has not set a preference.
 * Call applyUserPreferences(user) whenever the user object changes.
 */
import { useSystemStore } from '@/stores/system'
import { useTheme } from '@/composables/useTheme'

export function applyUserPreferences(user) {
  if (!user) return

  const root = document.documentElement
  const systemStore = useSystemStore()
  const sysDefaults = systemStore.defaults

  // Font family — user value takes priority, fall back to system default
  const font = user.font || sysDefaults.font || 'system'
  if (font === 'system') {
    root.style.removeProperty('--user-font')
  } else {
    root.style.setProperty('--user-font', font)
  }

  // Font size — user value takes priority, fall back to system default
  const size = parseInt(user.font_size || sysDefaults.font_size) || 14
  root.style.setProperty('--user-font-size', `${size}px`)

  // Sidebar position — stored in localStorage so App.vue can react
  const pos = user.sidebar_position === 'right' ? 'right' : 'left'
  localStorage.setItem('sidebar_position', pos)
  root.setAttribute('data-sidebar', pos)

  // Accent colour
  const { setAccentColor } = useTheme()
  setAccentColor(user.accent_color || 'blue')
}
