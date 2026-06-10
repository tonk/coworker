import { describe, it, expect } from 'vitest'
import {
  parseSlotHHMM,
  slotDayTypeMatches,
  slotCoverageOnWeekday,
  slotPreviewReady,
  buildSlotPreviewDays,
  formatSlotPreviewTime,
  formatDayType,
} from '../contractSlotPreview'

describe('parseSlotHHMM', () => {
  it('parses valid HH:MM values', () => {
    expect(parseSlotHHMM('00:00')).toBe(0)
    expect(parseSlotHHMM('08:30')).toBe(510)
    expect(parseSlotHHMM('23:59')).toBe(1439)
  })

  it('returns -1 for invalid values', () => {
    expect(parseSlotHHMM('')).toBe(-1)
    expect(parseSlotHHMM('8:00')).toBe(-1)
    expect(parseSlotHHMM('0800')).toBe(-1)
    expect(parseSlotHHMM('24:00')).toBe(-1)
    expect(parseSlotHHMM('23:60')).toBe(-1)
    expect(parseSlotHHMM(null)).toBe(-1)
  })
})

describe('slotDayTypeMatches', () => {
  it('matches all', () => {
    expect(slotDayTypeMatches('all', 0)).toBe(true)
    expect(slotDayTypeMatches('all', 6)).toBe(true)
  })

  it('matches weekdays (Mon=0 to Fri=4)', () => {
    expect(slotDayTypeMatches('weekdays', 0)).toBe(true)
    expect(slotDayTypeMatches('weekdays', 4)).toBe(true)
    expect(slotDayTypeMatches('weekdays', 5)).toBe(false)
    expect(slotDayTypeMatches('weekdays', 6)).toBe(false)
  })

  it('matches weekends (Sat=5, Sun=6)', () => {
    expect(slotDayTypeMatches('weekends', 5)).toBe(true)
    expect(slotDayTypeMatches('weekends', 6)).toBe(true)
    expect(slotDayTypeMatches('weekends', 0)).toBe(false)
  })

  it('matches specific day by key', () => {
    expect(slotDayTypeMatches('monday', 0)).toBe(true)
    expect(slotDayTypeMatches('monday', 1)).toBe(false)
    expect(slotDayTypeMatches('sunday', 6)).toBe(true)
  })

  it('handles empty string as all', () => {
    expect(slotDayTypeMatches('', 3)).toBe(true)
  })

  it('matches comma-separated list', () => {
    const monThuList = 'monday,tuesday,wednesday,thursday'
    expect(slotDayTypeMatches(monThuList, 0)).toBe(true)   // Monday
    expect(slotDayTypeMatches(monThuList, 3)).toBe(true)   // Thursday
    expect(slotDayTypeMatches(monThuList, 4)).toBe(false)  // Friday
    expect(slotDayTypeMatches(monThuList, 5)).toBe(false)  // Saturday
  })
})

describe('formatDayType', () => {
  const t = key => key.split('.').pop()  // stub: returns last key segment

  it('returns slot_days_all for all/empty', () => {
    expect(formatDayType('all', t)).toBe('slot_days_all')
    expect(formatDayType('', t)).toBe('slot_days_all')
  })

  it('returns preset labels for weekdays/weekends', () => {
    expect(formatDayType('weekdays', t)).toBe('slot_days_weekdays')
    expect(formatDayType('weekends', t)).toBe('slot_days_weekends')
  })

  it('returns single day label for single-day value', () => {
    expect(formatDayType('friday', t)).toBe('slot_days_friday')
  })

  it('returns comma-joined abbrevs for multi-day list', () => {
    const result = formatDayType('monday,tuesday,wednesday,thursday', t)
    expect(result).toBe('slot_preview_dow_mon, slot_preview_dow_tue, slot_preview_dow_wed, slot_preview_dow_thu')
  })
})

describe('slotCoverageOnWeekday', () => {
  it('returns intervals for same-day slot', () => {
    const slot = { start_time: '09:00', end_time: '17:00', day_type: 'all', end_day_offset: 0 }
    const intervals = slotCoverageOnWeekday(slot, 0)
    expect(intervals).toEqual([[540, 1020]])
  })

  it('handles overnight slot spanning multiple days', () => {
    const slot = { start_time: '22:00', end_time: '06:00', day_type: 'all', end_day_offset: 1 }
    const sun = slotCoverageOnWeekday(slot, 6)
    expect(sun).toEqual([[0, 360], [1320, 1440]])
    const tue = slotCoverageOnWeekday(slot, 1)
    expect(tue).toEqual([[0, 360], [1320, 1440]])
  })

  it('returns empty when day type does not match', () => {
    const slot = { start_time: '09:00', end_time: '17:00', day_type: 'weekdays', end_day_offset: 0 }
    expect(slotCoverageOnWeekday(slot, 5)).toEqual([])
  })

  it('returns empty for invalid times', () => {
    const slot = { start_time: 'xx:yy', end_time: '17:00', day_type: 'all', end_day_offset: 0 }
    expect(slotCoverageOnWeekday(slot, 0)).toEqual([])
  })

  it('merges overlapping intervals into full-day coverage', () => {
    const slot = { start_time: '22:00', end_time: '06:00', day_type: 'all', end_day_offset: 2 }
    const sun = slotCoverageOnWeekday(slot, 6)
    expect(sun).toEqual([[0, 1440]])
  })
})

describe('slotPreviewReady', () => {
  it('returns true for valid slot', () => {
    expect(slotPreviewReady({ start_time: '09:00', end_time: '17:00' })).toBe(true)
  })

  it('returns true for equal times (24-hour cycle)', () => {
    expect(slotPreviewReady({ start_time: '09:00', end_time: '09:00' })).toBe(true)
  })

  it('returns false for invalid times', () => {
    expect(slotPreviewReady({ start_time: '', end_time: '17:00' })).toBe(false)
    expect(slotPreviewReady({ start_time: 'xx:yy', end_time: '17:00' })).toBe(false)
  })
})

describe('buildSlotPreviewDays', () => {
  const slot = { start_time: '09:00', end_time: '17:00', day_type: 'weekdays', end_day_offset: 0 }
  const labels = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

  it('builds 7 rows', () => {
    const rows = buildSlotPreviewDays(slot, labels)
    expect(rows).toHaveLength(7)
  })

  it('marks weekdays as active', () => {
    const rows = buildSlotPreviewDays(slot, labels)
    expect(rows[0].active).toBe(true)
    expect(rows[4].active).toBe(true)
    expect(rows[5].active).toBe(false)
    expect(rows[6].active).toBe(false)
  })

  it('computes segment positions as percentages', () => {
    const rows = buildSlotPreviewDays(slot, labels)
    const seg = rows[0].segments[0]
    expect(seg.left).toBe((540 / 1440) * 100)
    expect(seg.width).toBe(((1020 - 540) / 1440) * 100)
  })
})

describe('formatSlotPreviewTime', () => {
  it('formats minutes to HH:MM', () => {
    expect(formatSlotPreviewTime(0)).toBe('00:00')
    expect(formatSlotPreviewTime(510)).toBe('08:30')
    expect(formatSlotPreviewTime(1440)).toBe('00:00')
    expect(formatSlotPreviewTime(1500)).toBe('01:00')
  })
})
