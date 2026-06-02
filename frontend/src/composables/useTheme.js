import { ref, watchEffect } from 'vue'

const theme = ref(localStorage.getItem('theme') || 'light')
const accentColor = ref(localStorage.getItem('accent_color') || 'blue')

function applyTheme(value) {
  const root = document.documentElement
  if (value === 'dark' || value === 'black') {
    root.setAttribute('data-theme', value)
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

function darkenHex(hex, amount = 0.82) {
  const r = parseInt(hex.slice(1, 3), 16)
  const g = parseInt(hex.slice(3, 5), 16)
  const b = parseInt(hex.slice(5, 7), 16)
  return `rgb(${Math.round(r * amount)}, ${Math.round(g * amount)}, ${Math.round(b * amount)})`
}

function applyAccentColor(value) {
  const root = document.documentElement
  if (!value || value === 'blue') {
    root.removeAttribute('data-accent')
    root.style.removeProperty('--color-primary')
    root.style.removeProperty('--color-primary-hover')
  } else if (value.startsWith('#')) {
    root.removeAttribute('data-accent')
    root.style.setProperty('--color-primary', value)
    root.style.setProperty('--color-primary-hover', darkenHex(value))
  } else {
    root.setAttribute('data-accent', value)
    root.style.removeProperty('--color-primary')
    root.style.removeProperty('--color-primary-hover')
  }
}

// Listen for system theme changes when theme is set to 'system'
const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
mediaQuery.addEventListener('change', () => {
  if (theme.value === 'system') applyTheme('system')
})

// Apply on initial load
applyTheme(theme.value)
applyAccentColor(accentColor.value)

export function useTheme() {
  function setTheme(value) {
    theme.value = value
    localStorage.setItem('theme', value)
    applyTheme(value)
  }

  function setAccentColor(value) {
    accentColor.value = value || 'blue'
    localStorage.setItem('accent_color', accentColor.value)
    applyAccentColor(accentColor.value)
  }

  watchEffect(() => applyTheme(theme.value))

  return { theme, setTheme, accentColor, setAccentColor }
}
