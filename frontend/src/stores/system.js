import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { systemApi } from '@/api/system'
import { setExternalImageProxyEnabled } from '@/api/serverConfig'
import { setLocale } from '@/i18n'

export const useSystemStore = defineStore('system', () => {
  const registrationEnabled = ref(true)
  const scrumStorypointsEnabled = ref(false)
  const externalImageProxyEnabled = ref(true)
  const sessionTimeoutMinutes = ref(60)
  const appMode = ref('')
  const isTimetrackingMode = computed(() => appMode.value === 'timetracking')
  const logoSrc     = computed(() => isTimetrackingMode.value ? '/timetracking.svg'      : '/logo.svg')
  const logoFullSrc = computed(() => isTimetrackingMode.value ? '/timetracking-full.svg' : '/logo-full.svg')
  const defaults = ref({
    date_time_format: 'YYYY-MM-DD HH:mm',
    timezone: 'UTC',
    theme: 'light',
    font: 'system',
    font_size: '14',
    locale: 'en'
  })

  async function fetchAppMode() {
    try {
      const { data } = await systemApi.getVersion()
      appMode.value = data.app_mode || ''
    } catch {}
  }

  async function fetchSettings() {
    try {
      const { data } = await systemApi.getSettings()
      registrationEnabled.value = data.registration_enabled !== false
      scrumStorypointsEnabled.value = data.scrum_storypoints_enabled === true
      externalImageProxyEnabled.value = data.external_image_proxy_enabled !== false
      setExternalImageProxyEnabled(externalImageProxyEnabled.value)
      sessionTimeoutMinutes.value = data.session_timeout_minutes || 0
      if (data.default_date_time_format) defaults.value.date_time_format = data.default_date_time_format
      if (data.default_timezone)         defaults.value.timezone         = data.default_timezone
      if (data.default_theme)            defaults.value.theme            = data.default_theme
      if (data.default_font)             defaults.value.font             = data.default_font
      if (data.default_font_size)        defaults.value.font_size        = data.default_font_size
      if (data.default_locale) {
        defaults.value.locale = data.default_locale
        // Apply only when the user has no stored preference
        if (!localStorage.getItem('locale')) await setLocale(data.default_locale)
      }
    } catch {}
  }

  return { registrationEnabled, scrumStorypointsEnabled, externalImageProxyEnabled, sessionTimeoutMinutes, appMode, isTimetrackingMode, logoSrc, logoFullSrc, defaults, fetchAppMode, fetchSettings }
})
