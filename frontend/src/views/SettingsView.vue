<template>
  <main class="settings-main">
      <div class="settings-container">
        <h1>{{ $t('settings.title') }}</h1>

        <div class="settings-card">
          <h2>{{ $t('settings.profile') }}</h2>
          <form @submit.prevent="saveProfile">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">{{ $t('settings.first_name') }}</label>
                <input class="form-input" v-model="form.first_name" :placeholder="$t('settings.first_name')" />
              </div>
              <div class="form-group">
                <label class="form-label">{{ $t('settings.last_name') }}</label>
                <input class="form-input" v-model="form.last_name" :placeholder="$t('settings.last_name')" />
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('settings.display_name') }}</label>
              <input class="form-input" v-model="form.display_name" :placeholder="$t('settings.display_name')" />
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('auth.email') }}</label>
              <input class="form-input" v-model="form.email" type="email" :placeholder="$t('auth.email')" />
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('settings.avatar_url') }}</label>
              <input class="form-input" v-model="form.avatar_url" :placeholder="$t('settings.avatar_url_placeholder')" />
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('settings.avatar_preview') }}</label>
              <div class="avatar-preview" v-if="form.avatar_url">
                <img :src="form.avatar_url" :alt="$t('settings.avatar_preview')" class="avatar-img" @error="avatarError = true" />
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('common.language') }}</label>
              <select class="form-input" v-model="form.locale">
                <option value="en">English</option>
                <option value="nl">Nederlands</option>
                <option value="de">Deutsch</option>
                <option value="fr">Français</option>
                <option value="es">Español</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('settings.theme') }}</label>
              <select class="form-input" v-model="form.theme">
                <option value="light">{{ $t('settings.theme_light') }}</option>
                <option value="dark">{{ $t('settings.theme_dark') }}</option>
                <option value="system">{{ $t('settings.theme_system') }}</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('settings.date_time_format') }}</label>
              <select class="form-input" v-model="form.date_time_format">
                <option value="YYYY-MM-DD HH:mm">YYYY-MM-DD HH:mm (ISO)</option>
                <option value="DD/MM/YYYY HH:mm">DD/MM/YYYY HH:mm</option>
                <option value="MM/DD/YYYY hh:mm a">MM/DD/YYYY hh:mm a</option>
                <option value="DD-MM-YYYY HH:mm">DD-MM-YYYY HH:mm</option>
                <option value="DD.MM.YYYY HH:mm">DD.MM.YYYY HH:mm</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('settings.timezone') }}</label>
              <select class="form-input" v-model="form.timezone">
                <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('settings.font') }}</label>
              <select class="form-input" v-model="form.font">
                <option value="system">{{ $t('settings.font_system') }}</option>
                <option value="Inter, sans-serif">Inter</option>
                <option value="'Roboto', sans-serif">Roboto</option>
                <option value="'Open Sans', sans-serif">Open Sans</option>
                <option value="'Source Code Pro', monospace">Source Code Pro (monospace)</option>
                <option value="Georgia, serif">Georgia (serif)</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('settings.font_size') }}</label>
              <select class="form-input" v-model="form.font_size">
                <option value="12">12px</option>
                <option value="13">13px</option>
                <option value="14">14px</option>
                <option value="15">15px</option>
                <option value="16">16px</option>
                <option value="18">18px</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('settings.sidebar_position') }}</label>
              <select class="form-input" v-model="form.sidebar_position">
                <option value="left">{{ $t('settings.sidebar_left') }}</option>
                <option value="right">{{ $t('settings.sidebar_right') }}</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">Email Notifications</label>
              <label class="toggle-label">
                <input type="checkbox" v-model="form.email_notifications" />
                <span>Receive email notifications for mentions, DMs, and card assignments</span>
              </label>
            </div>
            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="savingProfile">
                {{ savingProfile ? $t('common.loading') : $t('common.save') }}
              </button>
            </div>
          </form>
        </div>

        <div class="settings-card">
          <h2>{{ $t('auth.change_password') }}</h2>
          <form @submit.prevent="savePassword">
            <div class="form-group">
              <label class="form-label">{{ $t('auth.current_password') }}</label>
              <input class="form-input" type="password" v-model="pwForm.current_password" required />
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('auth.new_password') }}</label>
              <input class="form-input" type="password" v-model="pwForm.new_password" required minlength="8" />
            </div>
            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="savingPassword">
                {{ savingPassword ? $t('common.loading') : $t('auth.change_password') }}
              </button>
            </div>
          </form>
        </div>

        <div class="settings-card info-card">
          <div class="info-row">
            <span class="info-label">{{ $t('settings.last_login') }}</span>
            <span class="info-value">{{ auth.user?.last_login_at ? formatDateTime(auth.user.last_login_at) : '-' }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">{{ $t('settings.settings_updated_at') }}</span>
            <span class="info-value">{{ auth.user?.settings_updated_at ? formatDateTime(auth.user.settings_updated_at) : '-' }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">{{ $t('auth.username') }}</span>
            <span class="info-value">{{ auth.user?.username }}</span>
          </div>
        </div>
      </div>
  </main>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { useTheme } from '@/composables/useTheme'
import { authApi } from '@/api/auth'
import { applyUserPreferences } from '@/composables/useUserPreferences'
import { useDateFormat } from '@/composables/useDateFormat'

const auth = useAuthStore()
const ui = useUIStore()
const { setTheme } = useTheme()
const { formatDateTime } = useDateFormat()

const form = ref({
  first_name: '',
  last_name: '',
  display_name: '',
  email: '',
  avatar_url: '',
  locale: 'en',
  theme: 'system',
  date_time_format: 'YYYY-MM-DD HH:mm',
  timezone: 'UTC',
  font: 'system',
  font_size: '14',
  sidebar_position: 'left',
  email_notifications: true
})

const timezones = [
  'UTC',
  'Europe/Amsterdam', 'Europe/Berlin', 'Europe/Brussels', 'Europe/London',
  'Europe/Madrid', 'Europe/Paris', 'Europe/Rome', 'Europe/Stockholm',
  'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles',
  'America/Toronto', 'America/Vancouver', 'America/Sao_Paulo',
  'Asia/Dubai', 'Asia/Istanbul', 'Asia/Jerusalem', 'Asia/Kolkata',
  'Asia/Singapore', 'Asia/Shanghai', 'Asia/Tokyo', 'Asia/Seoul',
  'Australia/Sydney', 'Pacific/Auckland'
]

const pwForm = ref({ current_password: '', new_password: '' })
const savingProfile = ref(false)
const savingPassword = ref(false)
const avatarError = ref(false)

onMounted(() => {
  const u = auth.user
  if (u) {
    form.value = {
      first_name: u.first_name || '',
      last_name: u.last_name || '',
      display_name: u.display_name || '',
      email: u.email || '',
      avatar_url: u.avatar_url || '',
      locale: u.locale || 'en',
      theme: u.theme || localStorage.getItem('theme') || 'system',
      date_time_format: u.date_time_format || 'YYYY-MM-DD HH:mm',
      timezone: u.timezone || 'UTC',
      font: u.font || 'system',
      font_size: u.font_size || '14',
      sidebar_position: u.sidebar_position || 'left',
      email_notifications: u.email_notifications !== undefined ? u.email_notifications : true
    }
  }
})

async function saveProfile() {
  savingProfile.value = true
  try {
    await auth.updateProfile({
      first_name: form.value.first_name,
      last_name: form.value.last_name,
      display_name: form.value.display_name,
      email: form.value.email,
      avatar_url: form.value.avatar_url,
      locale: form.value.locale,
      theme: form.value.theme,
      date_time_format: form.value.date_time_format,
      timezone: form.value.timezone,
      font: form.value.font,
      font_size: form.value.font_size,
      sidebar_position: form.value.sidebar_position,
      email_notifications: form.value.email_notifications
    })
    applyUserPreferences(auth.user)
    setTheme(form.value.theme)
    ui.success('Profile saved')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to save profile')
  } finally {
    savingProfile.value = false
  }
}

async function savePassword() {
  savingPassword.value = true
  try {
    await authApi.changePassword(pwForm.value)
    pwForm.value = { current_password: '', new_password: '' }
    ui.success('Password changed')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to change password')
  } finally {
    savingPassword.value = false
  }
}
</script>

<style scoped>
.settings-main { flex: 1; padding: 32px 24px; }
.settings-container { max-width: 640px; margin: 0 auto; }
h1 { font-size: 22px; font-weight: 700; margin-bottom: 24px; }

.settings-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 24px;
  margin-bottom: 24px;
}
.settings-card h2 { font-size: 16px; font-weight: 600; margin-bottom: 20px; }

.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.form-hint { font-size: 12px; color: var(--color-text-muted); margin-top: 4px; display: block; }

.form-actions { display: flex; justify-content: flex-end; margin-top: 8px; }

.avatar-preview { margin-top: 8px; }
.avatar-img { width: 64px; height: 64px; border-radius: 50%; object-fit: cover; border: 2px solid var(--color-border); }

.info-card { display: flex; flex-direction: column; gap: 12px; }
.info-row { display: flex; justify-content: space-between; align-items: center; }
.info-label { font-size: 13px; color: var(--color-text-muted); font-weight: 500; }
.info-value { font-size: 13px; }

.toggle-label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  font-size: 13px;
  color: var(--color-text);
}
.toggle-label input[type=checkbox] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: var(--color-primary);
}
</style>
