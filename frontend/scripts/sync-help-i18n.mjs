#!/usr/bin/env node
/**
 * Merges help-content.{locale}.json into frontend/src/i18n/{locale}.json.
 * Missing locale files fall back to help-content.en.json.
 *
 * Usage: node scripts/sync-help-i18n.mjs
 * Add or edit: scripts/help-content.{locale}.json
 */
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const i18nDir = join(__dirname, '../src/i18n')
const locales = ['en', 'nl', 'de', 'fr', 'es', 'da', 'sv', 'nb', 'fi', 'is', 'pt', 'it']

const fallback = JSON.parse(readFileSync(join(__dirname, 'help-content.en.json'), 'utf8'))

function loadHelp(locale) {
  const path = join(__dirname, `help-content.${locale}.json`)
  if (existsSync(path)) {
    return JSON.parse(readFileSync(path, 'utf8'))
  }
  if (locale !== 'en') {
    console.warn(`help-content.${locale}.json not found — using English fallback`)
  }
  return fallback
}

for (const locale of locales) {
  const path = join(i18nDir, `${locale}.json`)
  const data = JSON.parse(readFileSync(path, 'utf8'))
  data.help = loadHelp(locale)
  writeFileSync(path, `${JSON.stringify(data, null, 2)}\n`)
  console.log(`updated ${locale}.json`)
}
