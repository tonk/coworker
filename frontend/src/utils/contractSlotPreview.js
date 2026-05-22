const MINUTES_PER_DAY = 24 * 60

// ISO weekday index: 0 = Monday … 6 = Sunday
const ISO_WEEKDAYS = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday']

export function parseSlotHHMM(value) {
  if (!value || value.length !== 5 || value[2] !== ':') return -1
  const h = parseInt(value.slice(0, 2), 10)
  const m = parseInt(value.slice(3, 5), 10)
  if (Number.isNaN(h) || Number.isNaN(m) || h < 0 || h > 23 || m < 0 || m > 59) return -1
  return h * 60 + m
}

export function slotDayTypeMatches(dayType, isoWeekday) {
  const key = ISO_WEEKDAYS[isoWeekday]
  switch (dayType) {
    case 'all':
    case '':
      return true
    case 'weekdays':
      return isoWeekday <= 4
    case 'weekends':
      return isoWeekday >= 5
    default:
      return dayType === key
  }
}

function mergeMinuteIntervals(intervals) {
  if (!intervals.length) return []
  const sorted = [...intervals].sort((a, b) => a[0] - b[0])
  const merged = [[sorted[0][0], sorted[0][1]]]
  for (let i = 1; i < sorted.length; i++) {
    const [s, e] = sorted[i]
    const last = merged[merged.length - 1]
    if (s <= last[1]) {
      if (e > last[1]) last[1] = e
    } else {
      merged.push([s, e])
    }
  }
  return merged
}

function slotEndDayOffset(slot) {
  const start = parseSlotHHMM(slot.start_time)
  const end = parseSlotHHMM(slot.end_time)
  if (start < 0 || end < 0 || end > start) return 0
  return slot.end_day_offset > 0 ? slot.end_day_offset : 1
}

/** Minute intervals [start, end) covered by the slot on one ISO weekday (0=Mon). */
export function slotCoverageOnWeekday(slot, isoWeekday) {
  const startM = parseSlotHHMM(slot.start_time)
  const endM = parseSlotHHMM(slot.end_time)
  if (startM < 0 || endM < 0 || startM === endM) return []

  const intervals = []
  const endOffset = slotEndDayOffset(slot)

  if (slotDayTypeMatches(slot.day_type, isoWeekday)) {
    if (endM > startM) {
      intervals.push([startM, endM])
    } else {
      intervals.push([startM, MINUTES_PER_DAY])
    }
  }

  if (endM <= startM) {
    for (let d = 1; d <= endOffset; d++) {
      const anchor = (isoWeekday - d + 7) % 7
      if (!slotDayTypeMatches(slot.day_type, anchor)) continue
      if (d < endOffset) {
        intervals.push([0, MINUTES_PER_DAY])
      } else {
        intervals.push([0, endM])
      }
    }
  }

  return mergeMinuteIntervals(intervals)
}

export function slotPreviewReady(slot) {
  return parseSlotHHMM(slot.start_time) >= 0 && parseSlotHHMM(slot.end_time) >= 0
    && parseSlotHHMM(slot.start_time) !== parseSlotHHMM(slot.end_time)
}

/** Build 7-day preview rows for the contract form timeline. */
export function buildSlotPreviewDays(slot, dayLabels) {
  return ISO_WEEKDAYS.map((key, isoWeekday) => {
    const intervals = slotCoverageOnWeekday(slot, isoWeekday)
    return {
      key,
      label: dayLabels[isoWeekday] || key.slice(0, 2),
      segments: intervals.map(([start, end]) => ({
        left: (start / MINUTES_PER_DAY) * 100,
        width: ((end - start) / MINUTES_PER_DAY) * 100,
      })),
      active: intervals.length > 0,
    }
  })
}

export function formatSlotPreviewTime(mins) {
  const h = Math.floor(mins / 60) % 24
  const m = mins % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`
}
