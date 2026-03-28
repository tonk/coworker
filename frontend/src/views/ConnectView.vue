<template>
  <div class="connect-page">
    <div class="connect-card">
      <img src="/logo.svg" alt="Coworker" class="connect-logo" />
      <h1 class="connect-title">Connect to Coworker</h1>
      <p class="connect-hint">Enter the URL of your Coworker server.</p>

      <form @submit.prevent="connect" class="connect-form">
        <div class="form-group">
          <label class="form-label">Server URL</label>
          <input
            ref="inputRef"
            v-model="serverUrl"
            class="form-input"
            type="url"
            placeholder="https://coworker.example.com"
            autocomplete="url"
            required
            :disabled="loading"
          />
        </div>

        <div v-if="error" class="connect-error">{{ error }}</div>

        <button class="btn btn-primary connect-btn" type="submit" :disabled="loading || !serverUrl.trim()">
          <span v-if="loading">Connecting…</span>
          <span v-else>Connect</span>
        </button>
      </form>

      <div v-if="currentServer" class="connect-current">
        Currently connected to <code>{{ currentServer }}</code>
        <button class="btn btn-ghost btn-sm" @click="continueWithCurrent">Continue</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { setServerUrl, getServerUrl } from '@/api/serverConfig'

const router = useRouter()
const serverUrl = ref('')
const loading = ref(false)
const error = ref('')
const currentServer = ref('')
const inputRef = ref(null)

onMounted(() => {
  currentServer.value = getServerUrl()
  inputRef.value?.focus()
})

async function connect() {
  error.value = ''
  loading.value = true

  const url = serverUrl.value.trim().replace(/\/+$/, '')

  try {
    // Probe the server — system/settings is public and lightweight.
    // We use globalThis.fetch directly (patched to the Tauri HTTP plugin in
    // Tauri mode) so the request bypasses CORS restrictions.
    const res = await globalThis.fetch(`${url}/api/v1/system/settings`, {
      signal: AbortSignal.timeout(8000),
    })
    if (res.ok || res.status === 401 || res.status === 403) {
      // Got a real response from a Coworker server
      setServerUrl(url)
      router.push('/login')
    } else {
      error.value = `Unexpected response from server (HTTP ${res.status}). Check the URL.`
    }
  } catch (e) {
    if (e.name === 'TimeoutError' || e.name === 'AbortError') {
      error.value = 'Connection timed out. Check the URL and try again.'
    } else {
      error.value = 'Could not reach the server. Check the URL and make sure it is reachable.'
    }
  } finally {
    loading.value = false
  }
}

function continueWithCurrent() {
  router.push('/login')
}
</script>

<style scoped>
.connect-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg);
  padding: 24px;
}

.connect-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 40px 36px;
  width: 100%;
  max-width: 420px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  box-shadow: 0 4px 24px rgba(0,0,0,.08);
}

.connect-logo {
  width: 56px;
  height: 56px;
  margin-bottom: 8px;
}

.connect-title {
  font-size: 22px;
  font-weight: 700;
  margin: 0;
}

.connect-hint {
  font-size: 14px;
  color: var(--color-text-muted);
  margin: 0 0 12px;
  text-align: center;
}

.connect-form {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.connect-btn {
  width: 100%;
  padding: 10px;
  font-size: 15px;
}

.connect-error {
  font-size: 13px;
  color: var(--color-danger);
  background: #fff5f5;
  border: 1px solid #fecaca;
  border-radius: var(--radius-sm);
  padding: 10px 12px;
}

.connect-current {
  margin-top: 16px;
  font-size: 13px;
  color: var(--color-text-muted);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  text-align: center;
}

.connect-current code {
  font-size: 12px;
  word-break: break-all;
}
</style>
