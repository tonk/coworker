import { parseWallClock, fmtWallClock, wallClockSpanMinutes } from './shiftTimeEntries'

export const PX_PER_HOUR = 60
export const MIN_BLOCK_HEIGHT_PX = 16
export const DEFAULT_SNAP_MINUTES = 15

/** Vertical pixel offset of a wall-clock time within a 24h day column. */
export function topOffsetPx(startTime, pxPerHour = PX_PER_HOUR) {
  const startM = parseWallClock(startTime)
  if (startM < 0) return 0
  return (startM / 60) * pxPerHour
}

/** Pixel height of a same-day span, floored so short entries stay clickable. */
export function heightPx(startTime, endTime, pxPerHour = PX_PER_HOUR) {
  const startM = parseWallClock(startTime)
  const endM = parseWallClock(endTime)
  const span = wallClockSpanMinutes(startM, endM)
  if (span == null || endM <= startM) return MIN_BLOCK_HEIGHT_PX
  return Math.max(MIN_BLOCK_HEIGHT_PX, (span / 60) * pxPerHour)
}

/** Inverse of topOffsetPx: pixel offset within a day column -> snapped "HH:MM". */
export function pxToWallClock(px, pxPerHour = PX_PER_HOUR, snapMinutes = DEFAULT_SNAP_MINUTES) {
  const rawMinutes = (px / pxPerHour) * 60
  const snapped = Math.round(rawMinutes / snapMinutes) * snapMinutes
  const clamped = Math.max(0, Math.min(24 * 60 - snapMinutes, snapped))
  return fmtWallClock(clamped)
}

/** Which day column (by index into columnRects) a clientX falls under, clamped to the array bounds. */
export function dayColumnIndexFromX(x, columnRects) {
  if (!columnRects || !columnRects.length) return 0
  for (let i = 0; i < columnRects.length; i++) {
    const rect = columnRects[i]
    if (x >= rect.left && x < rect.right) return i
  }
  if (x < columnRects[0].left) return 0
  return columnRects.length - 1
}
