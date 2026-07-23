<template>
  <div
    ref="dayEl"
    class="cal-day"
    :class="{ 'cal-day-today': isToday }"
    :style="{ '--hour-height': pxPerHour + 'px' }"
    @pointerdown="onSlotPointerDown"
    @contextmenu.prevent="onSlotContextMenu"
  >
    <div v-if="dragSelect" class="cal-drag-selection" :style="selectionStyle" aria-hidden="true" />
    <TimeTrackingCalendarBlock
      v-for="item in entries"
      :key="`${item.entry.id}-${item.segment}`"
      :entry="item.entry"
      :top="item.top"
      :height="item.height"
      :segment="item.segment"
      :px-per-hour="pxPerHour"
      :day-index="dayIndex"
      :week-days="weekDays"
      :get-column-rects="getColumnRects"
      :customer-name="item.customerName"
      :project-name="item.projectName"
      :color="item.color"
      :read-only="readOnly"
      :dense="item.height < 40"
      @open="$emit('block-open', $event)"
      @contextmenu="$emit('block-contextmenu', $event)"
      @move="$emit('block-move', $event)"
      @resize="$emit('block-resize', $event)"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import TimeTrackingCalendarBlock from './TimeTrackingCalendarBlock.vue'
import { pxToWallClock, DEFAULT_SNAP_MINUTES } from '@/utils/calendarLayout'
import { parseWallClock, fmtWallClock } from '@/utils/shiftTimeEntries'

const CLICK_THRESHOLD_PX = 4

const props = defineProps({
  dateISO: { type: String, required: true },
  isToday: { type: Boolean, default: false },
  entries: { type: Array, default: () => [] }, // [{ entry, top, height, customerName, projectName }]
  pxPerHour: { type: Number, required: true },
  dayIndex: { type: Number, required: true },
  weekDays: { type: Array, required: true },
  getColumnRects: { type: Function, required: true },
  readOnly: { type: Boolean, default: false },
})
const emit = defineEmits(['slot-click', 'slot-contextmenu', 'block-contextmenu', 'block-open', 'block-move', 'block-resize'])

const dayEl = ref(null)
const dragSelect = ref(null) // { pointerId, startY, currentY } while click-dragging to select a range

const selectionStyle = computed(() => {
  if (!dragSelect.value) return {}
  const { startY, currentY } = dragSelect.value
  const rect = dayEl.value.getBoundingClientRect()
  const top = Math.min(startY, currentY) - rect.top
  const height = Math.abs(currentY - startY)
  return { top: `${top}px`, height: `${height}px` }
})

function slotTimeAt(e) {
  const rect = dayEl.value.getBoundingClientRect()
  return pxToWallClock(e.clientY - rect.top, props.pxPerHour, DEFAULT_SNAP_MINUTES)
}

function onSlotPointerDown(e) {
  if (props.readOnly || e.target !== dayEl.value || e.button !== 0) return
  e.preventDefault()
  dayEl.value.setPointerCapture(e.pointerId)
  dragSelect.value = { pointerId: e.pointerId, startY: e.clientY, currentY: e.clientY }
  dayEl.value.addEventListener('pointermove', onSlotPointerMove)
  dayEl.value.addEventListener('pointerup', onSlotPointerUp, { once: true })
}

function onSlotPointerMove(e) {
  if (!dragSelect.value) return
  dragSelect.value.currentY = e.clientY
}

function onSlotPointerUp(e) {
  dayEl.value.removeEventListener('pointermove', onSlotPointerMove)
  const d = dragSelect.value
  dragSelect.value = null
  if (!d) return
  dayEl.value.releasePointerCapture?.(e.pointerId)

  const rect = dayEl.value.getBoundingClientRect()
  const startPx = Math.min(d.startY, d.currentY) - rect.top
  const dragDistance = Math.abs(d.currentY - d.startY)
  const startTime = pxToWallClock(startPx, props.pxPerHour, DEFAULT_SNAP_MINUTES)

  // A plain click (no real drag) leaves endTime unset — the caller falls back to
  // a default duration, matching the previous click-to-create behaviour.
  let endTime = null
  if (dragDistance >= CLICK_THRESHOLD_PX) {
    const endPx = Math.max(d.startY, d.currentY) - rect.top
    const startM = parseWallClock(startTime)
    let endM = parseWallClock(pxToWallClock(endPx, props.pxPerHour, DEFAULT_SNAP_MINUTES))
    if (endM <= startM) endM = startM + DEFAULT_SNAP_MINUTES
    endTime = fmtWallClock(endM)
  }

  emit('slot-click', { date: props.dateISO, startTime, endTime })
}

function onSlotContextMenu(e) {
  if (props.readOnly || e.target !== dayEl.value) return
  emit('slot-contextmenu', { x: e.clientX, y: e.clientY, date: props.dateISO, startTime: slotTimeAt(e) })
}

defineExpose({ getRect: () => dayEl.value?.getBoundingClientRect() })
</script>

<style scoped>
.cal-day {
  position: relative;
  flex: 1;
  min-width: 0;
  border-left: 1px solid var(--color-border);
  background-image: repeating-linear-gradient(
    to bottom,
    var(--color-border) 0,
    var(--color-border) 1px,
    transparent 1px,
    transparent var(--hour-height)
  );
}

.cal-day-today { background-color: rgba(99, 102, 241, 0.05); }

.cal-drag-selection {
  position: absolute;
  left: 2px;
  right: 2px;
  background: rgba(99, 102, 241, 0.25);
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-sm);
  pointer-events: none;
}
</style>
