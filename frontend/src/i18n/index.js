import { createI18n } from 'vue-i18n'
import en from './en.json'

export const SUPPORTED_LOCALES = ['en', 'nl', 'de', 'fr', 'es', 'da', 'sv', 'nb', 'fi', 'is', 'pt', 'it']

const localeLoaders = {
  en: () => Promise.resolve({ default: en }),
  nl: () => import('./nl.json'),
  de: () => import('./de.json'),
  fr: () => import('./fr.json'),
  es: () => import('./es.json'),
  da: () => import('./da.json'),
  sv: () => import('./sv.json'),
  nb: () => import('./nb.json'),
  fi: () => import('./fi.json'),
  is: () => import('./is.json'),
  pt: () => import('./pt.json'),
  it: () => import('./it.json'),
}

const loadedLocales = new Set(['en'])
const loadingLocales = new Map()

function normalizeLocale(locale) {
  return SUPPORTED_LOCALES.includes(locale) ? locale : 'en'
}

function getStoredLocale() {
  return normalizeLocale(localStorage.getItem('locale') || 'en')
}

export const i18n = createI18n({
  legacy: false,
  locale: getStoredLocale(),
  fallbackLocale: 'en',
  messages: { en },
})

export async function loadLocale(locale) {
  const resolved = normalizeLocale(locale)
  if (loadedLocales.has(resolved)) return resolved

  if (loadingLocales.has(resolved)) return loadingLocales.get(resolved)

  const promise = localeLoaders[resolved]().then((mod) => {
    i18n.global.setLocaleMessage(resolved, mod.default)
    loadedLocales.add(resolved)
    loadingLocales.delete(resolved)
    return resolved
  })
  loadingLocales.set(resolved, promise)
  return promise
}

export async function setLocale(locale) {
  const resolved = await loadLocale(locale)
  i18n.global.locale.value = resolved
  localStorage.setItem('locale', resolved)
  document.documentElement.lang = resolved
}

/** Load the stored locale (if not English) before first render. */
export async function initLocale() {
  const locale = getStoredLocale()
  document.documentElement.lang = locale
  if (locale !== 'en') {
    await loadLocale(locale)
    i18n.global.locale.value = locale
  }
}
