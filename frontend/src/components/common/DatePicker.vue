<template>
  <div class="dp-wrap" ref="wrapEl">
    <span
      class="dp-trigger"
      :class="{ 'dp-trigger-empty': !modelValue }"
      @click="toggle"
      @keydown.enter="toggle"
      @keydown.space.prevent="toggle"
      role="button"
      tabindex="0"
      :aria-label="modelValue ? displayed : placeholder"
    >{{ modelValue ? displayed : placeholder }}</span>

    <Teleport to="body">
      <div v-if="open" class="dp-scrim" @click="close" />
      <div
        v-if="open"
        class="dp-popup"
        :style="popupStyle"
        @keydown.escape="close"
        role="dialog"
        aria-modal="true"
        aria-label="Date picker"
      >
        <div class="dp-header">
          <button type="button" class="dp-nav-btn" @click="prevMonth" aria-label="Previous month">‹</button>
          <span class="dp-month-year">{{ monthYear }}</span>
          <button type="button" class="dp-nav-btn" @click="nextMonth" aria-label="Next month">›</button>
        </div>
        <div class="dp-weekdays" aria-hidden="true">
          <span v-for="wd in weekdayLabels" :key="wd">{{ wd }}</span>
        </div>
        <div class="dp-days" role="grid">
          <button
            v-for="(cell, i) in calDays"
            :key="i"
            type="button"
            class="dp-day"
            :class="{
              'dp-day-today': cell.isToday,
              'dp-day-selected': cell.isSelected,
              'dp-day-empty': !cell.date,
            }"
            :tabindex="cell.date ? 0 : -1"
            :aria-pressed="cell.isSelected || undefined"
            :aria-current="cell.isToday ? 'date' : undefined"
            :aria-label="cell.date || undefined"
            @click="cell.date && pick(cell.date)"
          >{{ cell.label }}</button>
        </div>
        <div v-if="modelValue" class="dp-footer">
          <button type="button" class="dp-clear-btn" @click="clear">{{ $t('common.clear') }}</button>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'
import { applyFormat, dateOnlyFmt } from '@/composables/useDateFormat'

const props = defineProps({
  modelValue: { type: String, default: null }, // YYYY-MM-DD or null
})
const emit = defineEmits(['update:modelValue'])

const auth = useAuthStore()
const system = useSystemStore()

const open = ref(false)
const wrapEl = ref(null)
const popupStyle = ref({})

const dateFmt = computed(() => {
  const full = auth.user?.date_time_format || system.defaults?.date_time_format || 'YYYY-MM-DD HH:mm'
  return dateOnlyFmt(full)
})

const placeholder = computed(() => dateFmt.value)
const displayed = computed(() => props.modelValue ? applyFormat(props.modelValue, dateFmt.value) : '')

const weekStart = computed(() => auth.user?.week_start || 'monday')
const weekdayLabels = computed(() => {
  const mo = ['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su']
  return weekStart.value === 'sunday' ? ['Su', ...mo.slice(0, 6)] : mo
})

const viewYear = ref(new Date().getFullYear())
const viewMonth = ref(new Date().getMonth() + 1)

function toggle() {
  if (open.value) { close(); return }
  if (props.modelValue) {
    const [y, m] = props.modelValue.split('-').map(Number)
    viewYear.value = y
    viewMonth.value = m
  } else {
    const now = new Date()
    viewYear.value = now.getFullYear()
    viewMonth.value = now.getMonth() + 1
  }
  open.value = true
  nextTick(positionPopup)
}

function close() { open.value = false }

function positionPopup() {
  if (!wrapEl.value) return
  const rect = wrapEl.value.getBoundingClientRect()
  const PW = 264
  const PH = 310
  let top = rect.bottom + 6
  let left = rect.left
  if (top + PH > window.innerHeight) top = rect.top - PH - 6
  if (left + PW > window.innerWidth) left = Math.max(8, window.innerWidth - PW - 8)
  popupStyle.value = { top: `${top}px`, left: `${left}px` }
}

const MONTHS = ['January','February','March','April','May','June','July','August','September','October','November','December']
const monthYear = computed(() => `${MONTHS[viewMonth.value - 1]} ${viewYear.value}`)

function prevMonth() {
  if (viewMonth.value === 1) { viewMonth.value = 12; viewYear.value-- }
  else viewMonth.value--
}
function nextMonth() {
  if (viewMonth.value === 12) { viewMonth.value = 1; viewYear.value++ }
  else viewMonth.value++
}

const calDays = computed(() => {
  const y = viewYear.value
  const m = viewMonth.value
  const daysInMonth = new Date(y, m, 0).getDate()
  const now = new Date()
  const todayStr = `${now.getFullYear()}-${String(now.getMonth()+1).padStart(2,'0')}-${String(now.getDate()).padStart(2,'0')}`
  let firstDow = new Date(y, m - 1, 1).getDay() // 0=Sun
  const offset = weekStart.value === 'sunday' ? firstDow : (firstDow + 6) % 7
  const cells = []
  for (let i = 0; i < offset; i++) cells.push({ label: '', date: null, isToday: false, isSelected: false })
  for (let d = 1; d <= daysInMonth; d++) {
    const dateStr = `${y}-${String(m).padStart(2,'0')}-${String(d).padStart(2,'0')}`
    cells.push({ label: String(d), date: dateStr, isToday: dateStr === todayStr, isSelected: dateStr === props.modelValue })
  }
  while (cells.length % 7 !== 0) cells.push({ label: '', date: null, isToday: false, isSelected: false })
  return cells
})

function pick(dateStr) {
  emit('update:modelValue', dateStr)
  close()
}

function clear() {
  emit('update:modelValue', null)
  close()
}
</script>

<style scoped>
.dp-wrap { position: relative; display: inline-flex; align-items: center; }

.dp-trigger {
  font-size: 13px;
  cursor: pointer;
  user-select: none;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  color: var(--color-text);
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  transition: background 0.1s;
  white-space: nowrap;
}
.dp-trigger:hover { background: var(--color-bg); }
.dp-trigger:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }
.dp-trigger-empty { color: var(--color-text-muted); font-style: italic; }

.dp-scrim { position: fixed; inset: 0; z-index: 1500; }

.dp-popup {
  position: fixed;
  z-index: 1501;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-md);
  padding: 12px;
  width: 264px;
}

.dp-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.dp-nav-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-size: 18px;
  line-height: 1;
}
.dp-nav-btn:hover { background: var(--color-bg); }

.dp-month-year { font-size: 14px; font-weight: 600; color: var(--color-text); }

.dp-weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  margin-bottom: 4px;
}
.dp-weekdays span {
  text-align: center;
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  padding: 3px 0;
}

.dp-days { display: grid; grid-template-columns: repeat(7, 1fr); gap: 1px; }

.dp-day {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  border: none;
  background: none;
  border-radius: 50%;
  cursor: pointer;
  color: var(--color-text);
  transition: background 0.1s;
  padding: 0;
}
.dp-day:hover:not(.dp-day-empty) { background: var(--color-bg); }
.dp-day-empty { opacity: 0; pointer-events: none; cursor: default; }
.dp-day-today { font-weight: 700; color: var(--color-primary); }
.dp-day-selected { background: var(--color-primary) !important; color: #fff; font-weight: 600; }
.dp-day-selected.dp-day-today { color: #fff; }

.dp-footer { margin-top: 10px; display: flex; justify-content: flex-start; border-top: 1px solid var(--color-border); padding-top: 8px; }
.dp-clear-btn {
  background: none;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: 12px;
  cursor: pointer;
  padding: 3px 10px;
  color: var(--color-text-muted);
}
.dp-clear-btn:hover { background: var(--color-bg); color: var(--color-text); }
</style>
