import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { i18n, initLocale } from './i18n'
import { useSystemStore } from '@/stores/system'
import { useAuthStore } from '@/stores/auth'
import { setRuntimeServerUrl } from '@/api/serverConfig'
import './styles/main.css'
import '@fontsource/inter/400.css'
import '@fontsource/inter/500.css'
import '@fontsource/inter/600.css'
import '@fontsource/inter/700.css'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import '@fontsource/open-sans/400.css'
import '@fontsource/open-sans/500.css'
import '@fontsource/open-sans/600.css'
import '@fontsource/open-sans/700.css'
import '@fontsource/source-code-pro/400.css'
import '@fontsource/source-code-pro/500.css'
import '@fontsource/source-code-pro/600.css'

// Both WebView2 (Windows, https://tauri.localhost) and WebKitGTK 4.1 (Linux,
// tauri://localhost) treat the Tauri origin as a secure context and block
// http:// requests as mixed content.  Route all fetch calls through
// tauri-plugin-http so requests go via the native Rust HTTP client, which has
// no such restriction.  index.html installs the window.fetch proxy before the
// ES module bundle loads so Axios captures it at import time.
// DIAGNOSTIC: timing laps for the Windows startup-delay investigation.
// Mirrors into the native warmdesk-startup.log via the client_log command so
// JS and Rust timestamps land in one unified timeline (devtools console
// alone won't show what happened before the page finished loading).
// Remove once the investigation concludes.
const _t0 = performance.now()
function _lap(msg) {
  const line = `${msg} (+${(performance.now() - _t0).toFixed(0)}ms)`
  console.log(`[warmdesk-boot] ${line}`)
  if (window.__TAURI_INTERNALS__) {
    window.__TAURI_INTERNALS__.invoke('client_log', { msg: line }).catch(() => {})
  }
}

async function init() {
  _lap('init() started')
  if (window.__TAURI_INTERNALS__) {
    window.addEventListener('contextmenu', e => e.preventDefault())

    _lap('→ import tauri-plugin-http')
    const httpPlugin = await import('@tauri-apps/plugin-http')
    _lap('← import tauri-plugin-http')
    const tauriFetch =
      httpPlugin.fetch ||
      httpPlugin.default?.fetch ||
      httpPlugin.default
    if (typeof tauriFetch === 'function') {
      window.__tauriFetch = tauriFetch
    } else {
      console.error('[WarmDesk] tauri-plugin-http fetch not available', Object.keys(httpPlugin || {}))
    }
    try {
      _lap('→ invoke runtime_server_url')
      const runtimeServerUrl = await window.__TAURI_INTERNALS__.invoke('runtime_server_url')
      _lap('← invoke runtime_server_url')
      if (runtimeServerUrl) setRuntimeServerUrl(String(runtimeServerUrl))
    } catch { _lap('← invoke runtime_server_url (threw)') }
  }

  _lap('→ initLocale()')
  await initLocale()
  _lap('← initLocale()')

  const app = createApp(App)
  const pinia = createPinia()
  app.use(pinia)
  app.use(i18n)

  app.config.errorHandler = (err, _instance, info) => {
    console.error('[Vue error]', info, err)
  }

  // Restore the session before installing the router so the beforeEach guard
  // sees the correct isLoggedIn state on the very first navigation. Installing
  // the router earlier caused a race: the guard ran while isLoggedIn was still
  // false, allowed the /login route, and then initSession() completed and set
  // the user — leaving the login form rendered inside the app shell.
  if (!window.__TAURI_INTERNALS__) {
    _lap('→ initSession()')
    await useAuthStore().initSession().catch(() => {})
    _lap('← initSession()')
  }
  _lap('→ fetchAppMode()')
  await useSystemStore().fetchAppMode().catch(() => {})
  _lap('← fetchAppMode()')

  app.use(router)
  app.mount('#app')
  _lap('app.mount() done')
  useSystemStore().fetchSettings()
}

init()
