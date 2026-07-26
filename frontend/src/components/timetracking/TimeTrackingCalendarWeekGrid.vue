<template>
  <div class="cal-week">
    <div class="cal-week-header">
      <div class="cal-hour-gutter-spacer" aria-hidden="true" />
      <div v-for="d in weekDays" :key="d.iso" class="cal-day-header" :class="{ 'cal-day-header-today': d.isToday }">
        <span class="cal-day-abbr">{{ d.abbr }}</span>
        <span class="cal-day-date">{{ d.mmdd }}</span>
      </div>
    </div>

    <div v-if="hasUnscheduled" class="cal-unscheduled-row">
      <div class="cal-hour-gutter-spacer" aria-hidden="true">{{ $t('timeTracking.calendar_unscheduled') }}</div>
      <div v-for="d in weekDays" :key="d.iso + '-u'" class="cal-unscheduled-cell">
        <button
          v-for="e in unscheduledByDay[d.iso] || []"
          :key="e.entry.id"
          :id="`tt-entry-${e.entry.id}`"
          type="button"
          class="cal-unscheduled-chip"
          :style="{ borderLeftColor: e.color }"
          @click="$emit('block-open', e.entry)"
          @contextmenu.prevent="$emit('block-contextmenu', { x: $event.clientX, y: $event.clientY, entry: e.entry })"
        >{{ e.customerName || e.entry.description || $t('timeTracking.no_customer') }}</button>
      </div>
    </div>

    <div ref="scrollEl" class="cal-scroll">
      <div class="cal-grid-body" :style="{ height: pxPerHour * 24 + 'px' }">
        <div class="cal-hour-gutter">
          <div v-for="h in 24" :key="h" class="cal-hour-label" :style="{ top: (h - 1) * pxPerHour + 'px' }">
            {{ hourLabel(h - 1) }}
          </div>
        </div>
        <TimeTrackingCalendarDay
          v-for="(d, i) in weekDays"
          :key="d.iso"
          :ref="(el) => setColumnRef(i, el)"
          :date-i-s-o="d.iso"
          :is-today="d.isToday"
          :entries="scheduledByDay[d.iso] || []"
          :px-per-hour="pxPerHour"
          :day-index="i"
          :week-days="weekDays"
          :get-column-rects="getColumnRects"
          :read-only="readOnly"
          @slot-click="$emit('slot-click', $event)"
          @slot-contextmenu="$emit('slot-contextmenu', $event)"
          @block-contextmenu="$emit('block-contextmenu', $event)"
          @block-open="$emit('block-open', $event)"
          @block-move="$emit('block-move', $event)"
          @block-resize="$emit('block-resize', $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import TimeTrackingCalendarDay from './TimeTrackingCalendarDay.vue'
import { topOffsetPx, heightPx, PX_PER_HOUR, MIN_BLOCK_HEIGHT_PX } from '@/utils/calendarLayout'
import { parseWallClock, addDaysISO } from '@/utils/shiftTimeEntries'
import { useDateFormat } from '@/composables/useDateFormat'

const { formatTime } = useDateFormat()

const props = defineProps({
  weekDays: { type: Array, required: true }, // [{ iso, mmdd, abbr, isToday }]
  entries: { type: Array, default: () => [] }, // raw TimeEntry rows for the displayed week
  pxPerHour: { type: Number, default: PX_PER_HOUR },
  customerName: { type: Function, required: true }, // (entry) => string
  projectName: { type: Function, required: true },  // (entry) => string
  entryColor: { type: Function, required: true }, // (entry) => hex color string (customer or project, per user setting)
  readOnly: { type: Boolean, default: false },
})
defineEmits(['slot-click', 'slot-contextmenu', 'block-contextmenu', 'block-open', 'block-move', 'block-resize'])

const scrollEl = ref(null)
const dayRefs = ref([])

function setColumnRef(i, el) {
  dayRefs.value[i] = el
}

// Measured fresh on every call (not cached) — the calendar panel can be display:none
// at mount time (v-show, table view is the default), which would otherwise freeze
// every column at a zero-size rect and make day-to-day dragging register nonsense jumps.
function getColumnRects() {
  return dayRefs.value.map((day) => day?.getRect()).filter(Boolean)
}

function hourLabel(hour) {
  return formatTime(new Date(2000, 0, 1, hour, 0))
}

const entriesByDay = computed(() => {
  const map = {}
  for (const e of props.entries) {
    if (!e.date) continue
    const iso = e.date.slice(0, 10)
    ;(map[iso] ||= []).push(e)
  }
  return map
})

const scheduledByDay = computed(() => {
  const result = {}
  for (const d of props.weekDays) result[d.iso] = []

  for (const d of props.weekDays) {
    for (const e of (entriesByDay.value[d.iso] || [])) {
      if (!e.start_time || !e.end_time) continue
      const startM = parseWallClock(e.start_time)
      const endM = parseWallClock(e.end_time)
      if (startM < 0 || endM < 0) continue

      const base = {
        entry: e,
        customerName: props.customerName(e),
        projectName: props.projectName(e),
        color: props.entryColor(e),
      }

      if (endM === startM) {
        // Equal start/end (e.g. 00:00–00:00) is the whole-day convention — a single
        // full-height block, not a zero-length or overnight-wrapping one.
        result[d.iso].push({ ...base, top: 0, height: 24 * props.pxPerHour, segment: 'full' })
        continue
      }

      if (endM > startM) {
        result[d.iso].push({
          ...base,
          top: topOffsetPx(e.start_time, props.pxPerHour),
          height: heightPx(e.start_time, e.end_time, props.pxPerHour),
          segment: 'full',
        })
        continue
      }

      // endM < startM: a genuine overnight entry (e.g. 19:00 → 07:00 next day).
      // Render as two segments: start_time→midnight on this day, and
      // midnight→end_time on the next day (only if that day is in view).
      result[d.iso].push({
        ...base,
        top: topOffsetPx(e.start_time, props.pxPerHour),
        height: Math.max(MIN_BLOCK_HEIGHT_PX, ((24 * 60 - startM) / 60) * props.pxPerHour),
        segment: 'start',
      })

      const nextIso = addDaysISO(d.iso, 1)
      if (result[nextIso]) {
        result[nextIso].push({
          ...base,
          top: 0,
          height: heightPx('00:00', e.end_time, props.pxPerHour),
          segment: 'continuation',
        })
      }
    }
  }
  return result
})

const unscheduledByDay = computed(() => {
  const result = {}
  for (const d of props.weekDays) {
    const list = (entriesByDay.value[d.iso] || [])
      .filter((e) => !e.start_time || !e.end_time)
      .map((e) => ({ entry: e, customerName: props.customerName(e), projectName: props.projectName(e), color: props.entryColor(e) }))
    if (list.length) result[d.iso] = list
  }
  return result
})

const hasUnscheduled = computed(() => Object.keys(unscheduledByDay.value).length > 0)

onMounted(async () => {
  await nextTick()
  if (scrollEl.value) scrollEl.value.scrollTop = 7 * props.pxPerHour
})
</script>

<style scoped>
.cal-week { display: flex; flex-direction: column; height: 100%; }

.cal-week-header, .cal-unscheduled-row {
  display: flex;
  border-bottom: 1px solid var(--color-border);
}

.cal-hour-gutter-spacer {
  width: 56px;
  flex-shrink: 0;
  font-size: 11px;
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  padding: 4px;
}

.cal-day-header {
  flex: 1;
  min-width: 0;
  text-align: center;
  padding: 6px 4px;
  font-size: 12px;
  border-left: 1px solid var(--color-border);
}

.cal-day-header-today { color: var(--color-primary); font-weight: 600; }
.cal-day-abbr { display: block; text-transform: uppercase; }
.cal-day-date { display: block; color: var(--color-text-muted); }

.cal-unscheduled-cell {
  flex: 1;
  min-width: 0;
  border-left: 1px solid var(--color-border);
  padding: 2px 4px;
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
}

.cal-unscheduled-chip {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border);
  border-left: 3px solid var(--color-border);
  background: var(--color-bg);
  color: var(--color-text);
  cursor: pointer;
}
.cal-block-highlight {
  outline: 3px solid var(--color-primary);
  outline-offset: 2px;
  transition: outline-color .2s ease-in;
}
@media (prefers-reduced-motion: reduce) {
  .cal-block-highlight { transition: none; }
}

.cal-scroll {
  flex: 1;
  overflow-y: auto;
  position: relative;
}

.cal-grid-body { display: flex; position: relative; }

.cal-hour-gutter {
  position: relative;
  width: 56px;
  flex-shrink: 0;
}

.cal-hour-label {
  position: absolute;
  right: 6px;
  transform: translateY(-50%);
  font-size: 11px;
  color: var(--color-text-muted);
  white-space: nowrap;
}
</style>
