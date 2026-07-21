<template>
  <div
    ref="blockEl"
    role="button"
    :tabindex="readOnly ? -1 : 0"
    class="cal-block"
    :class="{ 'cal-block-dragging': !!drag, 'cal-block-readonly': readOnly }"
    :style="blockStyle"
    :aria-label="accessibleLabel"
    @pointerdown="onMovePointerDown"
    @keydown.enter="onOpen"
    @keydown.space.prevent="onOpen"
    @click="onClick"
    @contextmenu.prevent="onContextMenu"
  >
    <div class="cal-block-body">
      <div class="cal-block-title">{{ customerName || $t('timeTracking.no_customer') }}</div>
      <div class="cal-block-sub">{{ projectName || $t('timeTracking.no_project') }}</div>
      <div v-if="!dense && entry.description" class="cal-block-activity">{{ entry.description }}</div>
    </div>
    <div
      v-if="!readOnly"
      class="cal-resize-handle cal-resize-top"
      aria-hidden="true"
      @pointerdown.stop="onResizePointerDown('top', $event)"
    />
    <div
      v-if="!readOnly"
      class="cal-resize-handle cal-resize-bottom"
      aria-hidden="true"
      @pointerdown.stop="onResizePointerDown('bottom', $event)"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDateFormat } from '@/composables/useDateFormat'
import { pxToWallClock, dayColumnIndexFromX, MIN_BLOCK_HEIGHT_PX, DEFAULT_SNAP_MINUTES } from '@/utils/calendarLayout'
import { parseWallClock, fmtWallClock, wallClockSpanMinutes } from '@/utils/shiftTimeEntries'
import { NO_CUSTOMER_COLOR } from '@/utils/calendarColors'

const props = defineProps({
  entry: { type: Object, required: true },
  top: { type: Number, required: true },
  height: { type: Number, required: true },
  pxPerHour: { type: Number, required: true },
  dayIndex: { type: Number, required: true },
  weekDays: { type: Array, required: true }, // [{ iso }, ...] in column order
  getColumnRects: { type: Function, required: true }, // () => DOMRect[] of day columns, measured live for cross-day drag
  customerName: { type: String, default: '' },
  projectName: { type: String, default: '' },
  color: { type: String, default: NO_CUSTOMER_COLOR }, // resolved by the parent (own customer color, or the assigned fallback)
  readOnly: { type: Boolean, default: false },
  dense: { type: Boolean, default: false },
})
const emit = defineEmits(['open', 'contextmenu', 'move', 'resize'])

const { t } = useI18n()
const { formatTime } = useDateFormat()

const blockEl = ref(null)
const drag = ref(null) // { kind: 'move' | 'resize', edge?, startX, startY, dx, dy, moved }
// A completed drag still fires a native 'click' on pointerup; suppress that one click.
let suppressNextClick = false

const CLICK_THRESHOLD_PX = 4

const blockStyle = computed(() => {
  const style = { top: `${props.top}px`, height: `${props.height}px`, background: props.color }
  if (drag.value) {
    if (drag.value.kind === 'move') {
      style.transform = `translate(${drag.value.dx}px, ${drag.value.dy}px)`
      style.zIndex = 50
    } else if (drag.value.edge === 'top') {
      const clampedDy = Math.min(drag.value.dy, props.height - MIN_BLOCK_HEIGHT_PX)
      style.top = `${props.top + clampedDy}px`
      style.height = `${props.height - clampedDy}px`
    } else {
      const clampedDy = Math.max(drag.value.dy, MIN_BLOCK_HEIGHT_PX - props.height)
      style.height = `${props.height + clampedDy}px`
    }
  }
  return style
})

const accessibleLabel = computed(() => {
  const dateOnly = props.entry.date ? props.entry.date.slice(0, 10) : ''
  const start = props.entry.start_time ? formatTime(`${dateOnly}T${props.entry.start_time}:00`) : ''
  const end = props.entry.end_time ? formatTime(`${dateOnly}T${props.entry.end_time}:00`) : ''
  const range = start && end ? `${start}–${end}` : ''
  const who = [props.customerName, props.projectName].filter(Boolean).join(' / ') || t('timeTracking.no_customer')
  let label = t('timeTracking.calendar_block_label', { range, who })
  if (props.entry.description) label += `: ${props.entry.description}`
  return label
})

function onOpen() {
  emit('open', props.entry)
}

function onContextMenu(e) {
  emit('contextmenu', { x: e.clientX, y: e.clientY, entry: props.entry })
}

function onClick() {
  if (suppressNextClick) { suppressNextClick = false; return }
  emit('open', props.entry)
}

function onMovePointerDown(e) {
  if (props.readOnly || e.button !== 0) return
  e.preventDefault()
  const el = blockEl.value
  el.setPointerCapture(e.pointerId)
  drag.value = { kind: 'move', pointerId: e.pointerId, startX: e.clientX, startY: e.clientY, dx: 0, dy: 0 }
  el.addEventListener('pointermove', onMovePointerMove)
  el.addEventListener('pointerup', onMovePointerUp, { once: true })
}

function onMovePointerMove(e) {
  if (!drag.value || drag.value.kind !== 'move') return
  drag.value.dx = e.clientX - drag.value.startX
  drag.value.dy = e.clientY - drag.value.startY
}

function onMovePointerUp(e) {
  const el = blockEl.value
  el.removeEventListener('pointermove', onMovePointerMove)
  const d = drag.value
  drag.value = null
  if (!d) return
  el.releasePointerCapture?.(e.pointerId)

  const dayDelta = dayColumnIndexFromX(e.clientX, props.getColumnRects()) - props.dayIndex
  // Below the drag threshold: treat as a plain click and let the browser's own
  // 'click' event (fired after this pointerup) open the entry via onClick.
  if (Math.abs(d.dy) < CLICK_THRESHOLD_PX && dayDelta === 0) return

  const startM = parseWallClock(props.entry.start_time)
  const endM = parseWallClock(props.entry.end_time)
  const duration = wallClockSpanMinutes(startM, endM) ?? 60
  const newStartTime = pxToWallClock(props.top + d.dy, props.pxPerHour, DEFAULT_SNAP_MINUTES)
  const newStartM = parseWallClock(newStartTime)
  const newEndM = Math.min(24 * 60 - 1, newStartM + duration)
  const newDayIndex = Math.max(0, Math.min(props.weekDays.length - 1, props.dayIndex + dayDelta))

  suppressNextClick = true
  emit('move', {
    entry: props.entry,
    newDate: props.weekDays[newDayIndex].iso,
    newStartTime,
    newEndTime: fmtWallClock(newEndM),
    newMinutes: newEndM - newStartM,
  })
}

function onResizePointerDown(edge, e) {
  if (props.readOnly || e.button !== 0) return
  e.preventDefault()
  const el = blockEl.value
  el.setPointerCapture(e.pointerId)
  drag.value = { kind: 'resize', edge, pointerId: e.pointerId, startX: e.clientX, startY: e.clientY, dx: 0, dy: 0 }
  el.addEventListener('pointermove', onResizePointerMove)
  el.addEventListener('pointerup', onResizePointerUp, { once: true })
}

function onResizePointerMove(e) {
  if (!drag.value || drag.value.kind !== 'resize') return
  drag.value.dy = e.clientY - drag.value.startY
}

function onResizePointerUp(e) {
  const el = blockEl.value
  el.removeEventListener('pointermove', onResizePointerMove)
  const d = drag.value
  drag.value = null
  if (!d) return
  el.releasePointerCapture?.(e.pointerId)
  suppressNextClick = true

  const startM = parseWallClock(props.entry.start_time)
  const endM = parseWallClock(props.entry.end_time)

  if (d.edge === 'top') {
    const clampedDy = Math.min(d.dy, props.height - MIN_BLOCK_HEIGHT_PX)
    const newStartTime = pxToWallClock(props.top + clampedDy, props.pxPerHour, DEFAULT_SNAP_MINUTES)
    const newStartM = Math.min(endM - DEFAULT_SNAP_MINUTES, parseWallClock(newStartTime))
    emit('resize', {
      entry: props.entry,
      newDate: props.entry.date,
      newStartTime: fmtWallClock(newStartM),
      newEndTime: props.entry.end_time,
      newMinutes: endM - newStartM,
    })
  } else {
    const clampedDy = Math.max(d.dy, MIN_BLOCK_HEIGHT_PX - props.height)
    const newEndTime = pxToWallClock(props.top + props.height + clampedDy, props.pxPerHour, DEFAULT_SNAP_MINUTES)
    const newEndM = Math.max(startM + DEFAULT_SNAP_MINUTES, parseWallClock(newEndTime))
    emit('resize', {
      entry: props.entry,
      newDate: props.entry.date,
      newStartTime: props.entry.start_time,
      newEndTime: fmtWallClock(newEndM),
      newMinutes: newEndM - startM,
    })
  }
}
</script>

<style scoped>
.cal-block {
  position: absolute;
  left: 2px;
  right: 2px;
  overflow: hidden;
  color: #fff;
  border-radius: var(--radius-sm);
  cursor: grab;
  font-size: 12px;
  line-height: 1.3;
  box-shadow: var(--shadow);
}

.cal-block:focus-visible {
  outline: 2px solid var(--color-text);
  outline-offset: 2px;
}

.cal-block-dragging { cursor: grabbing; opacity: 0.9; }
.cal-block-readonly { cursor: default; }

.cal-block-body { padding: 2px 6px; pointer-events: none; }
.cal-block-title { font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cal-block-sub, .cal-block-activity { opacity: 0.9; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.cal-resize-handle {
  position: absolute;
  left: 0;
  right: 0;
  height: 6px;
  cursor: ns-resize;
}

.cal-resize-top { top: 0; }
.cal-resize-bottom { bottom: 0; }
</style>
