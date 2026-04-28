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
              <div style="display:flex;gap:8px;align-items:center">
                <input class="form-input" v-model="form.avatar_url" :placeholder="$t('settings.avatar_url_placeholder')" style="flex:1" />
                <button type="button" class="btn btn-secondary btn-sm" @click="$refs.avatarFileInput.click()">{{ $t('settings.upload_avatar') }}</button>
                <button v-if="form.avatar_url" type="button" class="btn btn-danger btn-sm" @click="clearAvatar">{{ $t('common.clear') }}</button>
              </div>
              <input ref="avatarFileInput" type="file" accept="image/*" style="display:none" @change="onAvatarFileSelected" />
              <div class="avatar-preview" v-if="form.avatar_url" style="margin-top:8px">
                <img :src="resolveAssetUrl(form.avatar_url)" :alt="$t('settings.avatar_preview')" class="avatar-img" @error="avatarError = true" />
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
                <option value="da">Dansk</option>
                <option value="sv">Svenska</option>
                <option value="nb">Norsk</option>
                <option value="fi">Suomi</option>
                <option value="is">Íslenska</option>
                <option value="pt">Português</option>
                <option value="it">Italiano</option>
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
              <label class="form-label">{{ $t('settings.accent_color') }}</label>
              <div class="accent-swatches">
                <button
                  v-for="c in accentColors"
                  :key="c.value"
                  type="button"
                  class="accent-swatch"
                  :class="{ active: form.accent_color === c.value }"
                  :style="{ background: c.hex }"
                  :title="c.label"
                  @click="form.accent_color = c.value"
                ></button>
              </div>
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
                <option value="'FreeSans', sans-serif">FreeSans</option>
                <option value="'FreeSerif', serif">FreeSerif</option>
                <option value="'FreeMono', monospace">FreeMono (monospace)</option>
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
              <label class="form-label">Navigation</label>
              <label class="toggle-label">
                <input type="checkbox" v-model="form.show_breadcrumbs" />
                <span>Show breadcrumbs at the top</span>
              </label>
            </div>
            <div class="form-group">
              <label class="form-label">Email Notifications</label>
              <label class="toggle-label">
                <input type="checkbox" v-model="form.email_notifications" />
                <span>Receive email notifications for mentions, DMs, and card assignments</span>
              </label>
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('timeTracking.label') }}</label>
              <label class="toggle-label">
                <input type="checkbox" v-model="form.time_tracking_enabled" />
                <span>{{ $t('timeTracking.toggle_hint') }}</span>
              </label>
            </div>
            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="savingProfile">
                {{ savingProfile ? $t('common.loading') : $t('common.save') }}
              </button>
            </div>
          </form>
        </div>

        <div v-if="passwordExpired" class="auth-error" style="margin-bottom:16px;padding:12px 16px;border-radius:6px">
          {{ $t('auth.password_expired') }}
        </div>

        <div ref="pwCardRef" class="settings-card">
          <h2>{{ $t('auth.change_password') }}</h2>
          <form @submit.prevent="savePassword">
            <div class="form-group">
              <label class="form-label">{{ $t('auth.current_password') }}</label>
              <input class="form-input" type="password" v-model="pwForm.current_password" required />
            </div>
            <div class="form-group">
              <label class="form-label">{{ $t('auth.new_password') }}</label>
              <input class="form-input" type="password" v-model="pwForm.new_password" required :minlength="passwordPolicy.min_length" />
              <ul class="pw-requirements">
                <li>{{ $t('settings.req_min_length', { n: passwordPolicy.min_length }) }}</li>
                <li v-if="passwordPolicy.require_upper">{{ $t('settings.req_upper') }}</li>
                <li v-if="passwordPolicy.require_lower">{{ $t('settings.req_lower') }}</li>
                <li v-if="passwordPolicy.require_digit">{{ $t('settings.req_digit') }}</li>
                <li v-if="passwordPolicy.require_special">{{ $t('settings.req_special') }}</li>
              </ul>
            </div>
            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="savingPassword">
                {{ savingPassword ? $t('common.loading') : $t('auth.change_password') }}
              </button>
            </div>
          </form>
        </div>

        <!-- MFA card -->
        <div class="settings-card">
          <h2>{{ $t('mfa.title') }}</h2>
          <p class="form-hint" style="margin-bottom:16px">{{ $t('mfa.description') }}</p>

          <!-- MFA enabled state -->
          <div v-if="auth.user?.totp_enabled">
            <div class="mfa-status mfa-status-on">{{ $t('mfa.enabled') }}</div>
            <form @submit.prevent="disableMFA" style="margin-top:16px">
              <div class="form-group" style="max-width:320px">
                <label class="form-label">{{ $t('mfa.disable_instructions') }}</label>
                <input class="form-input" type="password" v-model="mfaDisablePassword" required :placeholder="$t('auth.password')" autocomplete="current-password" />
              </div>
              <div class="form-actions">
                <button type="submit" class="btn btn-danger" :disabled="mfaLoading || !mfaDisablePassword">
                  {{ mfaLoading ? $t('common.loading') : $t('mfa.confirm_disable') }}
                </button>
              </div>
            </form>
          </div>

          <!-- MFA disabled: setup flow -->
          <div v-else>
            <div class="mfa-status mfa-status-off">{{ $t('mfa.disabled') }}</div>

            <!-- Step 1: click to begin -->
            <div v-if="!mfaSetupData" style="margin-top:16px">
              <button class="btn btn-primary" @click="startMFASetup" :disabled="mfaLoading">
                {{ mfaLoading ? $t('common.loading') : $t('mfa.enable') }}
              </button>
            </div>

            <!-- Step 2: show QR + verify code -->
            <div v-else style="margin-top:16px">
              <p class="form-hint" style="margin-bottom:12px">{{ $t('mfa.setup_instructions') }}</p>
              <canvas ref="qrCanvas" class="mfa-qr"></canvas>
              <p class="form-hint" style="margin-top:12px">{{ $t('mfa.manual_entry') }}</p>
              <code class="mfa-secret">{{ mfaSetupData.secret }}</code>
              <form @submit.prevent="confirmMFAEnable" style="margin-top:20px">
                <div class="form-group" style="max-width:320px">
                  <label class="form-label">{{ $t('mfa.verify_code') }}</label>
                  <input
                    class="form-input mfa-code-input"
                    v-model="mfaEnableCode"
                    inputmode="numeric"
                    autocomplete="one-time-code"
                    maxlength="6"
                    required
                    placeholder="000000"
                  />
                </div>
                <div class="form-actions" style="gap:8px">
                  <button type="button" class="btn btn-secondary" @click="mfaSetupData = null; mfaEnableCode = ''">{{ $t('common.cancel') }}</button>
                  <button type="submit" class="btn btn-primary" :disabled="mfaLoading || mfaEnableCode.length !== 6">
                    {{ mfaLoading ? $t('common.loading') : $t('mfa.confirm_enable') }}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>

        <div class="settings-card">
          <h2>{{ $t('apikeys.personal_title') }}</h2>
          <p class="form-hint" style="margin-bottom:16px">{{ $t('apikeys.personal_description') }}</p>
          <div class="form-group" style="max-width:400px">
            <label class="form-label">{{ $t('apikeys.key_name') }}</label>
            <input class="form-input" v-model="newPersonalKeyName" :placeholder="$t('apikeys.key_name_placeholder')" />
          </div>
          <button class="btn btn-primary btn-sm" :disabled="!newPersonalKeyName.trim()" @click="generatePersonalKey">{{ $t('apikeys.generate') }}</button>

          <div v-if="generatedPersonalKey" class="personal-key-box">
            <p class="new-key-notice">{{ $t('apikeys.copy_notice') }}</p>
            <code class="new-key-value">{{ generatedPersonalKey }}</code>
            <button class="btn btn-secondary btn-sm" @click="copyPersonalKey">{{ $t('apikeys.copy') }}</button>
          </div>

          <table class="data-table" style="margin-top:24px">
            <thead>
              <tr>
                <th>{{ $t('apikeys.name') }}</th>
                <th>{{ $t('apikeys.prefix') }}</th>
                <th>{{ $t('apikeys.last_used') }}</th>
                <th>{{ $t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="personalKeys.length === 0">
                <td colspan="4" style="text-align:center;color:var(--color-text-muted)">{{ $t('apikeys.no_keys') }}</td>
              </tr>
              <tr v-for="key in personalKeys" :key="key.id">
                <td>{{ key.name }}</td>
                <td><code>{{ key.key_prefix }}…</code></td>
                <td>{{ key.last_used_at ? formatDateTime(key.last_used_at) : '—' }}</td>
                <td><button class="btn btn-danger btn-sm" @click="revokePersonalKey(key)">{{ $t('apikeys.revoke') }}</button></td>
              </tr>
            </tbody>
          </table>
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
import { ref, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { useTheme } from '@/composables/useTheme'
import { authApi } from '@/api/auth'
import { systemApi } from '@/api/system'
import { attachmentsApi } from '@/api/attachments'
import { resolveAssetUrl } from '@/api/serverConfig'
import { applyUserPreferences } from '@/composables/useUserPreferences'
import { useDateFormat } from '@/composables/useDateFormat'

const route = useRoute()
const auth = useAuthStore()
const ui = useUIStore()
const passwordExpired = ref(false)
const pwCardRef = ref(null)
const { setTheme, setAccentColor } = useTheme()

const accentColors = [
  { value: 'blue',   hex: '#6366f1' },
  { value: 'red',    hex: '#ef4444' },
  { value: 'green',  hex: '#22c55e' },
  { value: 'orange', hex: '#f97316' },
]
const { formatDateTime } = useDateFormat()
const { t: $t } = useI18n()

const form = ref({
  first_name: '',
  last_name: '',
  display_name: '',
  email: '',
  avatar_url: '',
  locale: 'en',
  theme: 'system',
  accent_color: 'blue',
  date_time_format: 'YYYY-MM-DD HH:mm',
  timezone: 'UTC',
  font: 'system',
  font_size: '14',
  sidebar_position: 'left',
  show_breadcrumbs: true,
  email_notifications: true,
  time_tracking_enabled: false
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
const passwordPolicy = ref({ min_length: 8, require_upper: false, require_lower: false, require_digit: false, require_special: false })
const savingProfile = ref(false)
const personalKeys = ref([])
const newPersonalKeyName = ref('')
const generatedPersonalKey = ref('')
const savingPassword = ref(false)
const avatarError = ref(false)

async function onAvatarFileSelected(e) {
  const file = e.target.files[0]
  if (!file) return
  e.target.value = ''
  try {
    const { data } = await attachmentsApi.uploadImage(file)
    form.value.avatar_url = data.url
    avatarError.value = false
  } catch {
    ui.error('Failed to upload image')
  }
}

// MFA
const mfaSetupData = ref(null)  // { secret, uri } during setup
const mfaEnableCode = ref('')
const mfaDisablePassword = ref('')
const mfaLoading = ref(false)
const qrCanvas = ref(null)

onMounted(async () => {
  loadPersonalKeys()
  try {
    const { data } = await systemApi.getSettings()
    if (data.password_policy) passwordPolicy.value = data.password_policy
  } catch {}
  if (route.query.expired === '1') {
    passwordExpired.value = true
    await nextTick()
    pwCardRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
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
      accent_color: u.accent_color || localStorage.getItem('accent_color') || 'blue',
      date_time_format: u.date_time_format || 'YYYY-MM-DD HH:mm',
      timezone: u.timezone || 'UTC',
      font: u.font || 'system',
      font_size: u.font_size || '14',
      sidebar_position: u.sidebar_position || 'left',
      show_breadcrumbs: u.show_breadcrumbs !== undefined ? u.show_breadcrumbs : true,
      email_notifications: u.email_notifications !== undefined ? u.email_notifications : true,
      time_tracking_enabled: !!u.time_tracking_enabled
    }
  }
})

async function clearAvatar() {
  form.value.avatar_url = ''
  await saveProfile()
}

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
      accent_color: form.value.accent_color,
      date_time_format: form.value.date_time_format,
      timezone: form.value.timezone,
      font: form.value.font,
      font_size: form.value.font_size,
      sidebar_position: form.value.sidebar_position,
      show_breadcrumbs: form.value.show_breadcrumbs,
      email_notifications: form.value.email_notifications,
      time_tracking_enabled: form.value.time_tracking_enabled
    })
    applyUserPreferences(auth.user)
    setTheme(form.value.theme)
    setAccentColor(form.value.accent_color)
    ui.success('Profile saved')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to save profile')
  } finally {
    savingProfile.value = false
  }
}

async function loadPersonalKeys() {
  const { data } = await authApi.listApiKeys()
  personalKeys.value = data
}

async function generatePersonalKey() {
  try {
    const { data } = await authApi.createApiKey(newPersonalKeyName.value.trim())
    generatedPersonalKey.value = data.key
    newPersonalKeyName.value = ''
    loadPersonalKeys()
  } catch (e) {
    ui.error('Failed to generate key')
  }
}

async function revokePersonalKey(key) {
  if (!confirm('Revoke this API key?')) return
  await authApi.deleteApiKey(key.id)
  loadPersonalKeys()
}

function copyPersonalKey() {
  navigator.clipboard.writeText(generatedPersonalKey.value)
  ui.success('Copied!')
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

async function startMFASetup() {
  mfaLoading.value = true
  try {
    const { data } = await authApi.setupMFA()
    mfaSetupData.value = data
    mfaEnableCode.value = ''
    // Render QR code on next tick when canvas is in DOM
    await nextTick()
    const QRCode = (await import('qrcode')).default
    QRCode.toCanvas(qrCanvas.value, data.uri, { width: 200, margin: 2 })
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to start MFA setup')
  } finally {
    mfaLoading.value = false
  }
}

async function confirmMFAEnable() {
  mfaLoading.value = true
  try {
    await authApi.enableMFA(mfaEnableCode.value)
    mfaSetupData.value = null
    mfaEnableCode.value = ''
    await auth.fetchMe()
    ui.success($t('mfa.setup_success'))
  } catch (e) {
    ui.error(e.response?.data?.error || $t('mfa.invalid_code'))
  } finally {
    mfaLoading.value = false
  }
}

async function disableMFA() {
  mfaLoading.value = true
  try {
    await authApi.disableMFA(mfaDisablePassword.value)
    mfaDisablePassword.value = ''
    await auth.fetchMe()
    ui.success($t('mfa.disable_success'))
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to disable MFA')
  } finally {
    mfaLoading.value = false
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

.accent-swatches { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 4px; }
.accent-swatch {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 3px solid transparent;
  cursor: pointer;
  transition: transform .1s, border-color .1s;
  outline: 2px solid transparent;
  outline-offset: 2px;
}
.accent-swatch:hover { transform: scale(1.15); }
.accent-swatch.active {
  outline-color: var(--color-text);
  transform: scale(1.15);
}

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

.personal-key-box {
  margin-top: 16px;
  padding: 12px 16px;
  background: var(--color-surface-raised);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.new-key-notice { font-size: 13px; color: var(--color-warning); margin: 0; }
.new-key-value { font-size: 13px; word-break: break-all; }

.mfa-status {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 600;
}
.mfa-status-on { background: color-mix(in srgb, var(--color-success) 15%, transparent); color: var(--color-success); }
.mfa-status-off { background: color-mix(in srgb, var(--color-text-muted) 15%, transparent); color: var(--color-text-muted); }
.mfa-qr { display: block; margin-top: 12px; border-radius: var(--radius); border: 1px solid var(--color-border); }
.mfa-secret { display: block; font-size: 13px; word-break: break-all; margin-top: 6px; letter-spacing: 2px; }
.mfa-code-input { font-size: 20px; letter-spacing: 6px; text-align: center; font-family: monospace; }

.pw-requirements {
  margin: 6px 0 0 0;
  padding: 0 0 0 16px;
  list-style: disc;
  font-size: 12px;
  color: var(--color-text-muted);
  line-height: 1.8;
}
</style>
