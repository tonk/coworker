<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-logo">
        <img :src="systemStore.logoSrc" alt="WarmDesk" style="height:36px;width:auto" />
      </div>
      <h1 class="auth-title">{{ $t('auth.register_title') }}</h1>
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label class="form-label">{{ $t('auth.email') }}</label>
          <input class="form-input" type="email" v-model="form.email" required />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('auth.username') }}</label>
          <input class="form-input" v-model="form.username" required minlength="3" />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('auth.display_name') }}</label>
          <input class="form-input" v-model="form.display_name" />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('auth.password') }}</label>
          <input class="form-input" type="password" v-model="form.password" required minlength="8" />
        </div>
        <p v-if="error" class="auth-error" role="alert">{{ error }}</p>
        <p v-if="serverReachabilityError" class="auth-warning">{{ serverReachabilityError }}</p>
        <button type="submit" class="btn btn-primary" style="width:100%" :disabled="loading">
          <span v-if="loading" class="spinner" style="width:16px;height:16px;border-width:2px"></span>
          {{ $t('auth.register') }}
        </button>
      </form>
      <p class="auth-link">
        {{ $t('auth.have_account') }}
        <RouterLink to="/login">{{ $t('auth.login') }}</RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'
import { getServerUrl } from '@/api/serverConfig'

const auth = useAuthStore()
const systemStore = useSystemStore()
const router = useRouter()
const form = ref({ email: '', username: '', display_name: '', password: '' })
const error = ref('')
const serverReachabilityError = ref('')
const loading = ref(false)
const isTauri = !!window.__TAURI_INTERNALS__

async function handleSubmit() {
  error.value = ''
  serverReachabilityError.value = ''
  loading.value = true
  try {
    await auth.register(form.value)
    router.push('/')
  } catch (e) {
    const msg = e.response?.data?.error || 'Registration failed'
    const unreachable =
      !e.response &&
      /network|fetch|timeout|scope|not allowed|failed/i.test(msg)
    if (isTauri && unreachable) {
      serverReachabilityError.value = 'Could not reach server. Check the URL and network connection.'
    }
    error.value = isTauri ? `${msg} (server: ${getServerUrl() || '(empty)'})` : msg
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg);
  padding: 24px;
}
.auth-card {
  background: #fff;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 40px;
  width: 100%;
  max-width: 400px;
  box-shadow: var(--shadow-md);
}
.auth-logo { font-size: 24px; font-weight: 800; color: var(--color-primary); text-align: center; margin-bottom: 8px; }
.auth-title { font-size: 18px; font-weight: 600; text-align: center; margin-bottom: 28px; }
.auth-error { color: var(--color-danger); font-size: 13px; margin-bottom: 12px; }
.auth-warning { color: var(--color-danger); font-size: 12px; margin-bottom: 12px; }
.auth-link { text-align: center; margin-top: 20px; font-size: 13px; color: var(--color-text-muted); }
</style>
