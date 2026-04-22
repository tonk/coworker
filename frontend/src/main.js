import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { i18n } from './i18n'
import { useSystemStore } from '@/stores/system'
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
async function init() {
  if (window.__TAURI_INTERNALS__) {
    const { fetch: tauriFetch } = await import('@tauri-apps/plugin-http')
    window.__tauriFetch = tauriFetch
  }

  const app = createApp(App)
  const pinia = createPinia()
  app.use(pinia)
  app.use(router)
  app.use(i18n)

  app.config.errorHandler = (err, _instance, info) => {
    console.error('[Vue error]', info, err)
  }

  app.mount('#app')
  useSystemStore().fetchSettings()
}

init()
