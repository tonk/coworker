<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-logo">
        <img src="/logo.svg" alt="WarmDesk" style="height:36px;width:auto" />
      </div>

      <!-- No token in URL -->
      <template v-if="!token">
        <h1 class="auth-title">{{ $t('auth.reset_password_title') }}</h1>
        <p class="auth-error-block">{{ $t('auth.reset_password_invalid') }}</p>
        <p class="auth-link">
          <RouterLink to="/forgot-password">{{ $t('auth.forgot_password_title') }}</RouterLink>
        </p>
      </template>

      <!-- Success -->
      <template v-else-if="done">
        <h1 class="auth-title">{{ $t('auth.reset_password_title') }}</h1>
        <p class="auth-hint">{{ $t('auth.reset_password_success') }}</p>
        <p class="auth-link">
          <RouterLink to="/login">{{ $t('auth.back_to_login') }}</RouterLink>
        </p>
      </template>

      <!-- Reset form -->
      <template v-else>
        <h1 class="auth-title">{{ $t('auth.reset_password_title') }}</h1>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label class="form-label">{{ $t('auth.reset_password_new') }}</label>
            <input class="form-input" type="password" v-model="password" :minlength="passwordPolicy.min_length" required autofocus />
            <ul class="pw-requirements">
              <li>{{ $t('settings.req_min_length', { n: passwordPolicy.min_length }) }}</li>
              <li v-if="passwordPolicy.require_upper">{{ $t('settings.req_upper') }}</li>
              <li v-if="passwordPolicy.require_lower">{{ $t('settings.req_lower') }}</li>
              <li v-if="passwordPolicy.require_digit">{{ $t('settings.req_digit') }}</li>
              <li v-if="passwordPolicy.require_special">{{ $t('settings.req_special') }}</li>
            </ul>
          </div>
          <div class="form-group">
            <label class="form-label">{{ $t('auth.reset_password_confirm') }}</label>
            <input class="form-input" type="password" v-model="confirm" required />
          </div>
          <p v-if="error" class="auth-error">{{ error }}</p>
          <button type="submit" class="btn btn-primary" style="width:100%" :disabled="loading">
            <span v-if="loading" class="spinner" style="width:16px;height:16px;border-width:2px"></span>
            {{ $t('auth.reset_password_submit') }}
          </button>
        </form>
        <p class="auth-link">
          <RouterLink to="/login">{{ $t('auth.back_to_login') }}</RouterLink>
        </p>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authApi } from '@/api/auth'
import { systemApi } from '@/api/system'

const { t } = useI18n()
const route = useRoute()

const token = ref('')
const password = ref('')
const confirm = ref('')
const error = ref('')
const loading = ref(false)
const done = ref(false)
const passwordPolicy = ref({ min_length: 8, require_upper: false, require_lower: false, require_digit: false, require_special: false })

onMounted(async () => {
  token.value = route.query.token || ''
  try {
    const { data } = await systemApi.getSettings()
    if (data.password_policy) passwordPolicy.value = data.password_policy
  } catch {}
})

async function handleSubmit() {
  error.value = ''
  if (password.value !== confirm.value) {
    error.value = t('auth.reset_password_mismatch')
    return
  }
  loading.value = true
  try {
    await authApi.resetPassword(token.value, password.value)
    done.value = true
  } catch (e) {
    error.value = e.response?.data?.error || t('auth.reset_password_invalid')
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
  margin-bottom: 12px;
  color: var(--color-text);
}

.auth-hint { font-size: 13px; color: var(--color-text-muted); text-align: center; margin-bottom: 24px; }
.auth-error { color: var(--color-danger); font-size: 13px; margin-bottom: 12px; }
.auth-error-block { font-size: 13px; color: var(--color-danger); text-align: center; margin-bottom: 20px; }
.auth-link { text-align: center; margin-top: 20px; font-size: 13px; color: var(--color-text-muted); }

.pw-requirements {
  margin: 6px 0 0 0;
  padding: 0 0 0 16px;
  list-style: disc;
  font-size: 12px;
  color: var(--color-text-muted);
  line-height: 1.8;
}
</style>
