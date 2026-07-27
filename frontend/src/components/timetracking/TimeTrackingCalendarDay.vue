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
import { pxToWallClock, DEFAULT_SNAP_MINUTES, GRID_TOP_INSET_PX } from '@/utils/calendarLayout'
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
  const top = Math.min(startY, currentY)
  const height = Math.abs(currentY - startY)
  return { top: `${top}px`, height: `${height}px` }
})

// Deliberately not dayEl.getBoundingClientRect() (nor the pointer event's own offsetY,
// which is computed relative to that same rect): inside the calendar's scrolled
// container, WebKitGTK (Tauri Linux desktop) has been observed returning a stale
// getBoundingClientRect().top for the scrolled-past day column that doesn't reflect
// how far the view has actually scrolled, throwing off any offset computed from it —
// independent of zoom level. The scroll container's OWN rect doesn't move as its
// content scrolls, so combining its rect with its (always-accurate) scrollTop sidesteps
// the bad geometry query entirely.
//
// Separately, WebKitGTK has also been observed reporting clientY/getBoundingClientRect
// in a visual pixel space that's a constant multiple of the CSS pixel space pxPerHour is
// defined in — tied to the display's OS-level scale factor, present even unscrolled and
// at any zoom level. visualScale() recovers that live ratio from two hour-gutter labels
// (positioned by a pure `top: N*pxPerHour` inline style, no geometry query involved in
// their layout) so the click math is correct regardless of what that ratio is on a given
// machine, rather than assuming it's always 1.
function visualScale(scrollEl) {
  const labels = scrollEl.querySelectorAll('.cal-hour-label')
  if (labels.length < 2) return 1
  const top0 = labels[0].getBoundingClientRect().top
  const top1 = labels[1].getBoundingClientRect().top
  const scale = (top1 - top0) / props.pxPerHour
  return scale > 0 ? scale : 1
}

function pxFromEvent(e) {
  const scrollEl = dayEl.value.closest('.cal-scroll')
  const scrollRect = scrollEl.getBoundingClientRect()
  // GRID_TOP_INSET_PX shifts the day column's own top down within .cal-grid-body
  // (see TimeTrackingCalendarWeekGrid.vue) so 00:00 clears the header border visually —
  // subtract it back out here so the click math still resolves to the right time.
  return scrollEl.scrollTop + (e.clientY - scrollRect.top) / visualScale(scrollEl) - GRID_TOP_INSET_PX
}

function slotTimeAt(e) {
  return pxToWallClock(pxFromEvent(e), props.pxPerHour, DEFAULT_SNAP_MINUTES)
}

function onSlotPointerDown(e) {
  if (props.readOnly || e.target !== dayEl.value || e.button !== 0) return
  e.preventDefault()
  dayEl.value.setPointerCapture(e.pointerId)
  const px = pxFromEvent(e)
  dragSelect.value = { pointerId: e.pointerId, startY: px, currentY: px }
  dayEl.value.addEventListener('pointermove', onSlotPointerMove)
  dayEl.value.addEventListener('pointerup', onSlotPointerUp, { once: true })
}

function onSlotPointerMove(e) {
  if (!dragSelect.value) return
  dragSelect.value.currentY = pxFromEvent(e)
}

function onSlotPointerUp(e) {
  dayEl.value.removeEventListener('pointermove', onSlotPointerMove)
  const d = dragSelect.value
  dragSelect.value = null
  if (!d) return
  dayEl.value.releasePointerCapture?.(e.pointerId)

  const startPx = Math.min(d.startY, d.currentY)
  const dragDistance = Math.abs(d.currentY - d.startY)
  const startTime = pxToWallClock(startPx, props.pxPerHour, DEFAULT_SNAP_MINUTES)

  // A plain click (no real drag) leaves endTime unset — the caller falls back to
  // a default duration, matching the previous click-to-create behaviour.
  let endTime = null
  if (dragDistance >= CLICK_THRESHOLD_PX) {
    const endPx = Math.max(d.startY, d.currentY)
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
