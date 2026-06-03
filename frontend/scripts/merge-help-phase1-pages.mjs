#!/usr/bin/env node
/** Merge help-content-phase1-pages.json into every help-content.{locale}.json pages object. */
import { readFileSync, writeFileSync, existsSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const additions = JSON.parse(readFileSync(join(__dirname, 'help-content-phase1-pages.json'), 'utf8'))
const locales = ['en', 'nl', 'de', 'fr', 'es', 'da', 'sv', 'nb', 'fi', 'is', 'pt', 'it']

for (const locale of locales) {
  const path = join(__dirname, `help-content.${locale}.json`)
  if (!existsSync(path)) {
    console.warn(`skip ${locale} — file missing`)
    continue
  }
  const data = JSON.parse(readFileSync(path, 'utf8'))
  data.field_button = data.field_button || (locale === 'en' ? 'More information' : data.button)
  data.pages = { ...data.pages, ...additions }
  writeFileSync(path, `${JSON.stringify(data, null, 2)}\n`)
  console.log(`merged phase1 pages → ${locale}`)
}
