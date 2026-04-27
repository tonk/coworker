<template>
  <div class="auth-page">

    <!-- ── Branded layout: branding card left, form card right ── -->
    <div v-if="branding.enabled" class="auth-split">

      <!-- Left: company branding panel -->
      <div class="auth-brand-panel">
        <div class="auth-brand-body">
          <img v-if="branding.logo" :src="branding.logo" class="auth-brand-logo" alt="" />
          <span v-if="!branding.logo && branding.name" class="auth-brand-initials">
            {{ branding.name.charAt(0) }}
          </span>
        </div>
        <div v-if="branding.name" class="auth-brand-name">{{ branding.name }}</div>
      </div>

      <!-- Right: login form -->
      <div class="auth-card">
        <div class="auth-logo">
          <img src="/logo.svg" alt="WarmDesk" style="height:32px;width:auto" />
        </div>
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
              <input class="form-input" type="password" autocomplete="current-password"
                spellcheck="false" autocorrect="off" autocapitalize="off"
                v-model="form.password" required />
            </div>
            <p v-if="error" class="auth-error">{{ error }}</p>
            <button type="submit" class="btn btn-primary" style="width:100%" :disabled="loading">
              <span v-if="loading" class="spinner" style="width:16px;height:16px;border-width:2px"></span>
              {{ $t('auth.login') }}
            </button>
          </form>
          <p class="auth-link"><RouterLink to="/forgot-password">{{ $t('auth.forgot_password') }}</RouterLink></p>
          <p v-if="registrationEnabled" class="auth-link">
            {{ $t('auth.no_account') }} <RouterLink to="/register">{{ $t('auth.register') }}</RouterLink>
          </p>
          <div v-if="isTauri" class="auth-server">
            <span class="auth-server-url">{{ currentServer }}</span>
            <RouterLink to="/connect" class="auth-server-change">Change</RouterLink>
          </div>
          <div v-if="isTauri && serverReachabilityError" class="auth-server-warning">{{ serverReachabilityError }}</div>
        </template>
        <template v-else>
          <h1 class="auth-title">{{ $t('mfa.mfa_required_title') }}</h1>
          <p class="auth-mfa-hint">{{ $t('mfa.mfa_required_instructions') }}</p>
          <form @submit.prevent="handleMFASubmit">
            <div class="form-group">
              <label class="form-label">{{ $t('mfa.code_placeholder') }}</label>
              <input class="form-input mfa-code-input" v-model="mfaCode"
                inputmode="numeric" autocomplete="one-time-code" maxlength="6"
                required autofocus placeholder="000000" />
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
    </div>

    <!-- ── Plain centered layout (no branding) ── -->
    <div v-else class="auth-card">
      <div class="auth-logo">
        <img src="/logo.svg" alt="WarmDesk" style="height:36px;width:auto" />
      </div>
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
            <input class="form-input" type="password" autocomplete="current-password"
              spellcheck="false" autocorrect="off" autocapitalize="off"
              v-model="form.password" required />
          </div>
          <p v-if="error" class="auth-error">{{ error }}</p>
          <button type="submit" class="btn btn-primary" style="width:100%" :disabled="loading">
            <span v-if="loading" class="spinner" style="width:16px;height:16px;border-width:2px"></span>
            {{ $t('auth.login') }}
          </button>
        </form>
        <p class="auth-link"><RouterLink to="/forgot-password">{{ $t('auth.forgot_password') }}</RouterLink></p>
        <p v-if="registrationEnabled" class="auth-link">
          {{ $t('auth.no_account') }} <RouterLink to="/register">{{ $t('auth.register') }}</RouterLink>
        </p>
        <div v-if="isTauri" class="auth-server">
          <span class="auth-server-url">{{ currentServer }}</span>
          <RouterLink to="/connect" class="auth-server-change">Change</RouterLink>
        </div>
        <div v-if="isTauri && serverReachabilityError" class="auth-server-warning">{{ serverReachabilityError }}</div>
      </template>
      <template v-else>
        <h1 class="auth-title">{{ $t('mfa.mfa_required_title') }}</h1>
        <p class="auth-mfa-hint">{{ $t('mfa.mfa_required_instructions') }}</p>
        <form @submit.prevent="handleMFASubmit">
          <div class="form-group">
            <label class="form-label">{{ $t('mfa.code_placeholder') }}</label>
            <input class="form-input mfa-code-input" v-model="mfaCode"
              inputmode="numeric" autocomplete="one-time-code" maxlength="6"
              required autofocus placeholder="000000" />
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

    <button class="login-theme-toggle" @click="toggleLoginTheme"
      :title="loginTheme === 'system' ? 'Following system theme — click for light' : loginTheme === 'light' ? 'Light mode — click for dark' : 'Dark mode — click for system'">
      <!-- monitor: system/auto theme -->
      <svg v-if="loginTheme === 'system'" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="2" y="3" width="20" height="14" rx="2"/>
        <polyline points="8 21 12 17 16 21"/>
      </svg>
      <!-- sun: shown in dark mode to switch to light -->
      <svg v-else-if="loginTheme === 'dark'" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="5"/>
        <line x1="12" y1="1" x2="12" y2="3"/>
        <line x1="12" y1="21" x2="12" y2="23"/>
        <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
        <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
        <line x1="1" y1="12" x2="3" y2="12"/>
        <line x1="21" y1="12" x2="23" y2="12"/>
        <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
        <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
      </svg>
      <!-- moon: shown in light mode to switch to dark -->
      <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
      </svg>
    </button>

    <div class="auth-version">WarmDesk v{{ appVersion }}</div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
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
const branding = ref({ enabled: false, name: '', logo: '' })
const isTauri = !!window.__TAURI_INTERNALS__
const currentServer = getServerUrl()
const serverReachabilityError = ref('')

// Login-page theme — separate from the logged-in user preference
// Three states: 'system' (follow OS) → 'light' → 'dark' → 'system'
const savedLoginTheme = localStorage.getItem('login_theme')
const loginTheme = ref(savedLoginTheme || 'system')
const userThemeBeforeLogin = document.documentElement.getAttribute('data-theme')

function osTheme() {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function applyLoginTheme(t) {
  document.documentElement.setAttribute('data-theme', t === 'system' ? osTheme() : t)
}

function toggleLoginTheme() {
  const next = { system: 'light', light: 'dark', dark: 'system' }
  loginTheme.value = next[loginTheme.value]
  localStorage.setItem('login_theme', loginTheme.value)
  applyLoginTheme(loginTheme.value)
}

applyLoginTheme(loginTheme.value)

onUnmounted(() => {
  // Restore whatever theme the app normally uses for the logged-in user
  const userTheme = localStorage.getItem('theme') || userThemeBeforeLogin || 'light'
  document.documentElement.setAttribute('data-theme', userTheme)
})

onMounted(async () => {
  if (isTauri) {
    if (currentServer) {
      try {
        const res = await window.__tauriFetch(`${currentServer}/api/v1/version`, {
          method: 'GET',
          headers: { Accept: 'application/json' },
        })
        if (!res.ok) {
          serverReachabilityError.value = `Could not reach server (${res.status}). Check the URL and try again.`
        }
      } catch (e) {
        serverReachabilityError.value = 'Could not reach server. Check the URL and network connection.'
      }
    }
  }
  try {
    const { data } = await systemApi.getSettings()
    registrationEnabled.value = data.registration_enabled
    branding.value = {
      enabled: !!data.login_branding_enabled && !!(data.company_name || data.company_logo),
      name: data.company_name || '',
      logo: data.company_logo || '',
    }
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
    const msg = serverMsg || e.message || 'Login failed'
    const unreachable =
      !e.response &&
      /network|fetch|timeout|scope|not allowed|failed/i.test(msg)
    if (isTauri && unreachable) {
      serverReachabilityError.value = 'Could not reach server. Check the URL and network connection.'
    }
    error.value = isTauri
      ? `${msg} (server: ${getServerUrl() || '(empty)'})`
      : msg
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
/* ── Page shell ─────────────────────────────────────── */
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg);
  padding: 24px;
}

/* ── Two-column split (branded mode) ────────────────── */
.auth-split {
  display: flex;
  align-items: stretch;
  gap: 24px;
  width: 100%;
  max-width: 800px;
}

/* ── Branding panel (left column) ───────────────────── */
.auth-brand-panel {
  flex: 0 0 340px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-md);
  padding: 36px 32px 28px;
  overflow: hidden;
  position: relative;
}

/* Subtle radial glow behind the logo */
.auth-brand-panel::before {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(
    ellipse 70% 55% at 50% 45%,
    color-mix(in srgb, var(--color-primary) 12%, transparent),
    transparent 70%
  );
  pointer-events: none;
}

.auth-brand-body {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.auth-brand-logo {
  max-width: 180px;
  max-height: 120px;
  width: auto;
  height: auto;
  object-fit: contain;
  filter: drop-shadow(0 4px 24px color-mix(in srgb, var(--color-primary) 30%, transparent));
}

.auth-brand-initials {
  font-size: 80px;
  font-weight: 800;
  color: var(--color-primary);
  opacity: 0.25;
  line-height: 1;
  user-select: none;
}

.auth-brand-name {
  position: relative;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--color-text-muted);
  margin-top: 24px;
  text-align: center;
}

/* ── Login form card ────────────────────────────────── */
.auth-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 40px;
  width: 100%;
  max-width: 400px;
  box-shadow: var(--shadow-md);
}

/* Inside branded split the card fills remaining space */
.auth-split .auth-card {
  flex: 1;
  max-width: none;
}

/* ── Form chrome ────────────────────────────────────── */
.auth-logo {
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
.auth-server-url { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 260px; }
.auth-server-change { flex-shrink: 0; color: var(--color-primary); text-decoration: none; }
.auth-server-change:hover { text-decoration: underline; }
.auth-server-warning { margin-top: 8px; font-size: 11px; color: var(--color-danger); text-align: center; word-break: break-all; }

/* ── Responsive: stack vertically on narrow screens ── */
@media (max-width: 640px) {
  .auth-split { flex-direction: column; max-width: 400px; }
  .auth-brand-panel { flex: none; min-height: 200px; }
  .auth-split .auth-card { max-width: none; }
}

/* ── Light/dark toggle ──────────────────────────────── */
.login-theme-toggle {
  position: fixed;
  top: 16px;
  right: 16px;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-muted);
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
  z-index: 100;
  box-shadow: var(--shadow-md);
}
.login-theme-toggle:hover {
  background: var(--color-bg-hover);
  color: var(--color-text);
}

/* ── Version watermark ──────────────────────────────── */
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
