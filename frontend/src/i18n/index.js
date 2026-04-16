import { createI18n } from 'vue-i18n'
import en from './en.json'
import nl from './nl.json'
import de from './de.json'
import fr from './fr.json'
import es from './es.json'
import da from './da.json'
import sv from './sv.json'
import nb from './nb.json'
import fi from './fi.json'
import is from './is.json'
import pt from './pt.json'
import it from './it.json'

export const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('locale') || 'en',
  fallbackLocale: 'en',
  messages: { en, nl, de, fr, es, da, sv, nb, fi, is, pt, it }
})

export function setLocale(locale) {
  i18n.global.locale.value = locale
  localStorage.setItem('locale', locale)
  document.documentElement.lang = locale
}
