#!/usr/bin/env node
/** Merge phase1 bundle JSON into help-content.{de,fr,fi,is}.json */
import { readFileSync, writeFileSync, existsSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const locales = ['de', 'fr', 'fi', 'is']

for (const locale of locales) {
  const bundlePath = join(__dirname, 'phase1-locales', `${locale}.json`)
  const targetPath = join(__dirname, `help-content.${locale}.json`)
  if (!existsSync(bundlePath)) {
    console.error(`missing ${bundlePath}`)
    process.exit(1)
  }
  const bundle = JSON.parse(readFileSync(bundlePath, 'utf8'))
  const data = JSON.parse(readFileSync(targetPath, 'utf8'))
  data.pages = { ...data.pages, ...bundle.pages }
  data.field_button = bundle.field_button
  data.fields = bundle.fields
  writeFileSync(targetPath, `${JSON.stringify(data, null, 2)}\n`)
  console.log(`merged phase1 → help-content.${locale}.json (${Object.keys(bundle.pages).length} pages)`)
}
