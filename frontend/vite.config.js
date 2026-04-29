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
    rolldownOptions: {
      output: {
        manualChunks(id) {
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
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true
      }
    }
  }
})
