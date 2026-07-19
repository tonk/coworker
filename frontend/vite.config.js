import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { execSync } from 'node:child_process'

const gitVersion = (() => {
  try {
    return execSync('git describe --tags --always', { encoding: 'utf8' }).trim().replace(/^v/, '')
  } catch {
    return 'dev'
  }
})()

export default defineConfig({
  plugins: [vue()],
  define: {
    __APP_VERSION__: JSON.stringify(gitVersion)
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      'frappe-gantt/dist/frappe-gantt.css': fileURLToPath(new URL('./node_modules/frappe-gantt/dist/frappe-gantt.css', import.meta.url))
    }
  },
  build: {
    chunkSizeWarningLimit: 700,
    rolldownOptions: {
      output: {
        manualChunks(id) {
          // Split large app-internal shared modules out of the entry chunk.
          if (id.includes('/src/stores/'))
            return 'chunk-stores'
          if (id.includes('/src/api/') || id.includes('/src/composables/useWebSocket'))
            return 'chunk-app-net'
          if (id.includes('/src/i18n/')) {
            const localeMatch = id.match(/\/src\/i18n\/([a-z]{2})\.json$/)
            if (localeMatch) return `chunk-i18n-${localeMatch[1]}`
            return 'chunk-app-i18n'
          }
          if (
            id.includes('/src/components/layout/') ||
            id.includes('/src/components/common/ToastContainer.vue') ||
            id.includes('/src/components/common/UpdateBanner.vue') ||
            id.includes('/src/composables/useProjectChatUnread') ||
            id.includes('/src/composables/useUserPreferences') ||
            id.includes('/src/composables/useUpdateCheck')
          )
            return 'chunk-app-shell'
          if (
            id.includes('/src/components/call/') ||
            id.includes('/src/composables/useWebRTCCall') ||
            id.includes('/src/composables/useLiveKitGroupCall') ||
            id.includes('/src/composables/useCallSettings')
          )
            return 'chunk-calls'

          if (!id.includes('node_modules')) return

          // Vue core + router + state
          if (/node_modules\/(vue|vue-router|pinia|@vue)\//.test(id))
            return 'chunk-vue'

          // i18n (vue-i18n + intl-messageformat + related)
          if (/node_modules\/(vue-i18n|@intlify|intl-messageformat)\//.test(id))
            return 'chunk-i18n'

          // Markdown pipeline (renderer + highlighter)
          if (/node_modules\/(marked|highlight\.js)\//.test(id))
            return 'chunk-markdown'

          // HTTP + sanitisation utilities
          if (/node_modules\/(axios|dompurify|follow-redirects)\//.test(id))
            return 'chunk-http'
        }
      }
    }
  },
  server: {
    port: 5173,
    watch: {
      // src-tauri/target churns constantly during `cargo build`/`tauri dev`.
      // On Windows, cargo holds an exclusive lock on .dll/.rlib files while
      // writing them, and Vite's fs watcher trying to watch a locked file
      // crashes the dev server outright (EBUSY). Never watch Rust build output.
      ignored: ['**/src-tauri/target/**']
    },
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
