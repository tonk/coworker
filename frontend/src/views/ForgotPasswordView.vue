<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-logo">
        <img src="/logo.svg" alt="WarmDesk" style="height:36px;width:auto" />
      </div>

      <template v-if="!sent">
        <h1 class="auth-title">{{ $t('auth.forgot_password_title') }}</h1>
        <p class="auth-hint">{{ $t('auth.forgot_password_instructions') }}</p>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label class="form-label">{{ $t('auth.email') }}</label>
            <input class="form-input" type="email" v-model="email" required autofocus />
          </div>
          <p v-if="error" class="auth-error">{{ error }}</p>
          <button type="submit" class="btn btn-primary" style="width:100%" :disabled="loading">
            <span v-if="loading" class="spinner" style="width:16px;height:16px;border-width:2px"></span>
            {{ $t('auth.forgot_password_submit') }}
          </button>
        </form>
      </template>

      <template v-else>
        <h1 class="auth-title">{{ $t('auth.forgot_password_title') }}</h1>
        <p class="auth-hint">{{ $t('auth.forgot_password_sent') }}</p>
      </template>

      <p class="auth-link">
        <RouterLink to="/login">{{ $t('auth.back_to_login') }}</RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { authApi } from '@/api/auth'

const email = ref('')
const sent = ref(false)
const error = ref('')
const loading = ref(false)

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    await authApi.forgotPassword(email.value)
    sent.value = true
  } catch {
    // Always show the "sent" confirmation to avoid account enumeration.
    sent.value = true
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
.auth-link { text-align: center; margin-top: 20px; font-size: 13px; color: var(--color-text-muted); }
</style>
