import { describe, it, expect } from 'vitest'
import {
  parseWallClock,
  fmtWallClock,
  wallClockSpanMinutes,
  addDaysISO,
  splitShiftIntoDayEntries,
  weekendStandbyDefaults,
} from '../shiftTimeEntries'

describe('parseWallClock', () => {
  it('parses valid times', () => {
    expect(parseWallClock('00:00')).toBe(0)
    expect(parseWallClock('08:00')).toBe(480)
    expect(parseWallClock('23:59')).toBe(1439)
    expect(parseWallClock(' 08:00 ')).toBe(480)
  })

  it('returns -1 for invalid inputs', () => {
    expect(parseWallClock('')).toBe(-1)
    expect(parseWallClock(null)).toBe(-1)
    expect(parseWallClock('0800')).toBe(-1)
    expect(parseWallClock('24:00')).toBe(-1)
    expect(parseWallClock('23:60')).toBe(-1)
    expect(parseWallClock('abc')).toBe(-1)
  })
})

describe('fmtWallClock', () => {
  it('formats minutes to HH:MM', () => {
    expect(fmtWallClock(0)).toBe('00:00')
    expect(fmtWallClock(480)).toBe('08:00')
    expect(fmtWallClock(1439)).toBe('23:59')
  })
})

describe('wallClockSpanMinutes', () => {
  it('computes span within same day', () => {
    expect(wallClockSpanMinutes(480, 1080)).toBe(600)
  })

  it('handles overnight span', () => {
    expect(wallClockSpanMinutes(1140, 420)).toBe(720)
  })

  it('returns null for invalid inputs', () => {
    expect(wallClockSpanMinutes(-1, 100)).toBeNull()
    expect(wallClockSpanMinutes(100, -1)).toBeNull()
    expect(wallClockSpanMinutes(100, 100)).toBeNull()
  })
})

describe('addDaysISO', () => {
  it('adds days to an ISO date', () => {
    expect(addDaysISO('2025-01-01', 0)).toBe('2025-01-01')
    expect(addDaysISO('2025-01-01', 1)).toBe('2025-01-02')
    expect(addDaysISO('2025-12-31', 1)).toBe('2026-01-01')
    expect(addDaysISO('2025-01-01', -1)).toBe('2024-12-31')
  })
})

describe('splitShiftIntoDayEntries', () => {
  it('splits a single-day shift', () => {
    const result = splitShiftIntoDayEntries('2025-01-15', '08:00', '2025-01-15', '17:00')
    expect(result).toHaveLength(1)
    expect(result[0]).toEqual({
      date: '2025-01-15',
      start_time: '08:00',
      end_time: '17:00',
      minutes: 540,
    })
  })

  it('splits a multi-day shift', () => {
    const result = splitShiftIntoDayEntries('2025-01-15', '22:00', '2025-01-17', '06:00')
    expect(result).toHaveLength(3)
    expect(result[0]).toEqual({ date: '2025-01-15', start_time: '22:00', end_time: '23:59', minutes: 120 })
    expect(result[1]).toEqual({ date: '2025-01-16', start_time: '00:00', end_time: '23:59', minutes: 1440 })
    expect(result[2]).toEqual({ date: '2025-01-17', start_time: '00:00', end_time: '06:00', minutes: 360 })
  })

  it('returns empty for end before start', () => {
    expect(splitShiftIntoDayEntries('2025-01-17', '08:00', '2025-01-15', '17:00')).toEqual([])
  })

  it('returns empty for invalid times', () => {
    expect(splitShiftIntoDayEntries('2025-01-15', 'abc', '2025-01-15', '17:00')).toEqual([])
    expect(splitShiftIntoDayEntries('2025-01-15', '08:00', '2025-01-15', 'abc')).toEqual([])
    expect(splitShiftIntoDayEntries(null, '08:00', '2025-01-15', '17:00')).toEqual([])
  })
})

describe('weekendStandbyDefaults', () => {
  it('returns Friday 19:00 to Monday 07:00', () => {
    const week = ['2025-01-13', '2025-01-14', '2025-01-15', '2025-01-16', '2025-01-17', '2025-01-18', '2025-01-19']
    const result = weekendStandbyDefaults(week)
    expect(result.start_date).toBe('2025-01-17')
    expect(result.start_time).toBe('19:00')
    expect(result.end_date).toBe('2025-01-20')
    expect(result.end_time).toBe('07:00')
  })

  it('falls back to index 4 when no Friday found', () => {
    const week = ['2025-01-12', '2025-01-13', '2025-01-14', '2025-01-15', '2025-01-16']
    const result = weekendStandbyDefaults(week)
    expect(result.start_date).toBe('2025-01-16')
  })
})
