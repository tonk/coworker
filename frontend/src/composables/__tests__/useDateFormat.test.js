import { describe, it, expect } from 'vitest'
import { pad, applyFormat, dateOnlyFmt } from '../useDateFormat'

describe('pad', () => {
  it('pads single-digit numbers', () => {
    expect(pad(0)).toBe('00')
    expect(pad(5)).toBe('05')
    expect(pad(9)).toBe('09')
  })

  it('does not pad two-digit numbers', () => {
    expect(pad(10)).toBe('10')
    expect(pad(59)).toBe('59')
  })
})

describe('applyFormat', () => {
  it('formats Date objects with full format', () => {
    const d = new Date(2025, 0, 15, 14, 30, 0)
    expect(applyFormat(d, 'YYYY-MM-DD HH:mm')).toBe('2025-01-15 14:30')
  })

  it('formats ISO date-only string as local date', () => {
    expect(applyFormat('2025-01-15', 'YYYY-MM-DD')).toBe('2025-01-15')
  })

  it('formats ISO date-time string', () => {
    const result = applyFormat('2025-01-15T14:30:00Z', 'YYYY-MM-DD HH:mm')
    expect(result).toMatch(/2025-01-1[5-6] 1[4-6]:30/)
  })

  it('handles 12-hour format with am/pm', () => {
    const morning = new Date(2025, 0, 15, 9, 15)
    const evening = new Date(2025, 0, 15, 21, 15)
    expect(applyFormat(morning, 'hh:mm a')).toBe('09:15 am')
    expect(applyFormat(evening, 'hh:mm a')).toBe('09:15 pm')
  })

  it('returns the input string for invalid dates', () => {
    expect(applyFormat('not-a-date', 'YYYY-MM-DD')).toBe('not-a-date')
  })

  it('handles midnight as 12 am', () => {
    const d = new Date(2025, 0, 15, 0, 0)
    expect(applyFormat(d, 'hh:mm a')).toBe('12:00 am')
  })

  it('handles noon as 12 pm', () => {
    const d = new Date(2025, 0, 15, 12, 0)
    expect(applyFormat(d, 'hh:mm a')).toBe('12:00 pm')
  })
})

describe('dateOnlyFmt', () => {
  it('strips time portion from format', () => {
    expect(dateOnlyFmt('YYYY-MM-DD HH:mm')).toBe('YYYY-MM-DD')
    expect(dateOnlyFmt('MM/DD/YYYY HH:mm a')).toBe('MM/DD/YYYY')
  })

  it('returns the same string when no time portion', () => {
    expect(dateOnlyFmt('YYYY-MM-DD')).toBe('YYYY-MM-DD')
  })
})
