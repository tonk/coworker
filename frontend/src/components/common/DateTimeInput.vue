<template>
  <div>
    <input
      class="form-input"
      :class="{ 'input-error': error }"
      type="text"
      :value="displayValue"
      :placeholder="fmt"
      :aria-label="label || undefined"
      :aria-describedby="hint ? hintId : undefined"
      autocomplete="off"
      spellcheck="false"
      @change="onChange($event.target.value)"
    />
    <p v-if="hint && !error" :id="hintId" class="form-hint">{{ hint }}</p>
    <p v-if="error" class="field-error" role="alert">{{ error }}</p>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'

const props = defineProps({
  modelValue: { type: String, default: null },
  label:      { type: String, default: '' },
  hint:       { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue'])
const { t } = useI18n()

const auth   = useAuthStore()
const system = useSystemStore()

const fmt = computed(() =>
  auth.user?.date_time_format || system.defaults?.date_time_format || 'YYYY-MM-DD HH:mm'
)

const hintId = `dti-hint-${Math.random().toString(36).slice(2)}`
const error  = ref('')

// ── Formatting ────────────────────────────────────────────────────────────────

function pad(n) { return String(n).padStart(2, '0') }

function applyFmt(d, f) {
  const h   = d.getHours()
  const h12 = h % 12 || 12
  return f
    .replace('YYYY', d.getFullYear())
    .replace('MM',   pad(d.getMonth() + 1))
    .replace('DD',   pad(d.getDate()))
    .replace('HH',   pad(h))
    .replace('hh',   pad(h12))
    .replace('mm',   pad(d.getMinutes()))
    .replace('a',    h < 12 ? 'am' : 'pm')
}

const displayValue = computed(() => {
  if (!props.modelValue) return ''
  const d = new Date(props.modelValue)
  return isNaN(d) ? '' : applyFmt(d, fmt.value)
})

// ── Parsing ───────────────────────────────────────────────────────────────────

function buildRegex(f) {
  const caps = {
    YYYY: '(?<fYYYY>\\d{4})',
    MM:   '(?<fMM>\\d{1,2})',
    DD:   '(?<fDD>\\d{1,2})',
    HH:   '(?<fHH>\\d{1,2})',
    hh:   '(?<fhh>\\d{1,2})',
    mm:   '(?<fmm>\\d{1,2})',
    a:    '(?<fa>[aApP][mM])',
  }
  const tokenRe = /YYYY|MM|DD|HH|hh|mm|a/g
  let result = '', last = 0, m
  tokenRe.lastIndex = 0
  while ((m = tokenRe.exec(f)) !== null) {
    result += f.slice(last, m.index).replace(/[-/\\^$*+?.()|[\]{}]/g, '\\$&')
    result += caps[m[0]]
    last = m.index + m[0].length
  }
  result += f.slice(last).replace(/[-/\\^$*+?.()|[\]{}]/g, '\\$&')
  return new RegExp('^' + result + '$', 'i')
}

function parseValue(val) {
  const m = val.trim().match(buildRegex(fmt.value))
  if (!m) return null
  const g     = m.groups
  const year  = parseInt(g.fYYYY ?? new Date().getFullYear())
  const month = parseInt(g.fMM   ?? 1) - 1
  const day   = parseInt(g.fDD   ?? 1)
  let   hours = parseInt(g.fHH   != null ? g.fHH : (g.fhh ?? 0))
  const mins  = parseInt(g.fmm   ?? 0)
  if (g.fa) {
    const ap = g.fa.toLowerCase()
    if (ap === 'pm' && hours < 12) hours += 12
    else if (ap === 'am' && hours === 12) hours = 0
  }
  const d = new Date(year, month, day, hours, mins)
  return isNaN(d.getTime()) ? null : d
}

function onChange(val) {
  error.value = ''
  if (!val.trim()) {
    emit('update:modelValue', null)
    return
  }
  const d = parseValue(val)
  if (!d) {
    error.value = t('common.invalid_date', { fmt: fmt.value })
    return
  }
  emit('update:modelValue', d.toISOString())
}
</script>
