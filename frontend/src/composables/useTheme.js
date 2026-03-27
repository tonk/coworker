import { ref, watchEffect } from 'vue'

const theme = ref(localStorage.getItem('theme') || 'light')

function applyTheme(value) {
  const root = document.documentElement
  if (value === 'dark') {
    root.setAttribute('data-theme', 'dark')
  } else if (value === 'light') {
    root.removeAttribute('data-theme')
  } else {
    // system
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    if (prefersDark) {
      root.setAttribute('data-theme', 'dark')
    } else {
      root.removeAttribute('data-theme')
    }
  }
}

// Listen for system theme changes when theme is set to 'system'
const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
mediaQuery.addEventListener('change', () => {
  if (theme.value === 'system') applyTheme('system')
})

// Apply theme on initial load
applyTheme(theme.value)

export function useTheme() {
  function setTheme(value) {
    theme.value = value
    localStorage.setItem('theme', value)
    applyTheme(value)
  }

  watchEffect(() => applyTheme(theme.value))

  return { theme, setTheme }
}
