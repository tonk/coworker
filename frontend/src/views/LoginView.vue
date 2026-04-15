<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-logo">
        <img src="/logo.svg" alt="WarmDesk" style="height:36px;width:auto" />
      </div>

      <!-- Step 1: credentials -->
      <template v-if="!mfaStep">
        <h1 class="auth-title">{{ $t('auth.login_title') }}</h1>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label class="form-label">{{ $t('auth.email') }} / {{ $t('auth.username') }}</label>
            <input class="form-input" v-model="form.login" required autofocus
              spellcheck="false" autocorrect="off" autocapitalize="off" />
          </div>
          <div class="form-group">
            <label class="form-label">{{ $t('auth.password') }}</label>
            <input
              class="form-input"
              type="password"
              autocomplete="current-password"
              spellcheck="false" autocorrect="off" autocapitalize="off"
              v-model="form.password"
              required
            />
          </div>
          <p v-if="error" class="auth-error">{{ error }}</p>
          <button type="submit" class="btn btn-primary" style="width:100%" :disabled="loading">
            <span v-if="loading" class="spinner" style="width:16px;height:16px;border-width:2px"></span>
            {{ $t('auth.login') }}
          </button>
        </form>
        <p class="auth-link">
          <RouterLink to="/forgot-password">{{ $t('auth.forgot_password') }}</RouterLink>
        </p>
        <p v-if="registrationEnabled" class="auth-link">
          {{ $t('auth.no_account') }}
          <RouterLink to="/register">{{ $t('auth.register') }}</RouterLink>
        </p>
        <div v-if="isTauri" class="auth-server">
          <span class="auth-server-url">{{ currentServer }}</span>
          <RouterLink to="/connect" class="auth-server-change">Change</RouterLink>
        </div>
      </template>

      <!-- Step 2: TOTP code -->
      <template v-else>
        <h1 class="auth-title">{{ $t('mfa.mfa_required_title') }}</h1>
        <p class="auth-mfa-hint">{{ $t('mfa.mfa_required_instructions') }}</p>
        <form @submit.prevent="handleMFASubmit">
          <div class="form-group">
            <label class="form-label">{{ $t('mfa.code_placeholder') }}</label>
            <input
              class="form-input mfa-code-input"
              v-model="mfaCode"
              inputmode="numeric"
              autocomplete="one-time-code"
              maxlength="6"
              required
              autofocus
              placeholder="000000"
            />
          </div>
          <p v-if="error" class="auth-error">{{ error }}</p>
          <button type="submit" class="btn btn-primary" style="width:100%" :disabled="loading">
            <span v-if="loading" class="spinner" style="width:16px;height:16px;border-width:2px"></span>
            {{ $t('auth.login') }}
          </button>
          <button type="button" class="btn btn-secondary" style="width:100%;margin-top:8px" @click="mfaStep = false; error = ''">
            {{ $t('common.cancel') }}
          </button>
        </form>
      </template>
    </div>
    <div class="auth-version">WarmDesk v{{ appVersion }}</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

const appVersion = __APP_VERSION__
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { systemApi } from '@/api/system'
import { getServerUrl } from '@/api/serverConfig'

const { t: $t } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const form = ref({ login: '', password: '' })
const mfaStep = ref(false)
const mfaCode = ref('')
const error = ref('')
const loading = ref(false)
const registrationEnabled = ref(true)
const isTauri = !!window.__TAURI_INTERNALS__
const currentServer = getServerUrl()

onMounted(async () => {
  try {
    const { data } = await systemApi.getSettings()
    registrationEnabled.value = data.registration_enabled
  } catch {}
})

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    const result = await auth.login(form.value.login, form.value.password)
    if (result.mfa_required) {
      mfaStep.value = true
      mfaCode.value = ''
      return
    }
    router.push('/')
  } catch (e) {
    const data = e.response?.data
    const serverMsg = data?.error
      ?? (typeof data === 'string' ? (() => { try { return JSON.parse(data).error } catch { return data } })() : null)
    error.value = serverMsg || e.message || 'Login failed'
  } finally {
    loading.value = false
  }
}

async function handleMFASubmit() {
  error.value = ''
  loading.value = true
  try {
    await auth.verifyMFA(mfaCode.value)
    router.push('/')
  } catch (e) {
    const serverError = e.response?.data?.error
    if (serverError === 'mfa_session_expired') {
      // MFA token expired or invalid — go back to step 1
      mfaStep.value = false
      mfaCode.value = ''
      error.value = $t('mfa.session_expired')
    } else {
      error.value = serverError === 'invalid_code' ? $t('mfa.invalid_code') : (e.message || $t('mfa.invalid_code'))
    }
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
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 40px;
  width: 100%;
  max-width: 400px;
  box-shadow: var(--shadow-md);
}

.auth-logo {
  font-size: 24px;
  font-weight: 800;
  color: var(--color-primary);
  text-align: center;
  margin-bottom: 8px;
}

.auth-title {
  font-size: 18px;
  font-weight: 600;
  text-align: center;
  margin-bottom: 28px;
  color: var(--color-text);
}

.auth-error { color: var(--color-danger); font-size: 13px; margin-bottom: 12px; }
.auth-link { text-align: center; margin-top: 20px; font-size: 13px; color: var(--color-text-muted); }
.auth-mfa-hint { font-size: 13px; color: var(--color-text-muted); margin-bottom: 20px; text-align: center; }
.mfa-code-input { font-size: 24px; letter-spacing: 8px; text-align: center; font-family: monospace; }

.auth-server {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 16px;
  font-size: 12px;
  color: var(--color-text-muted);
}

.auth-server-url {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 260px;
}

.auth-server-change {
  flex-shrink: 0;
  color: var(--color-primary);
  text-decoration: none;
}

.auth-server-change:hover {
  text-decoration: underline;
}

.auth-version {
  position: fixed;
  bottom: 16px;
  left: 0;
  right: 0;
  text-align: center;
  font-size: 12px;
  color: var(--color-text-muted);
  opacity: 0.6;
}

</style>
