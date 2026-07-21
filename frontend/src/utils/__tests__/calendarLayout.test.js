import { describe, it, expect } from 'vitest'
import {
  topOffsetPx,
  heightPx,
  pxToWallClock,
  dayColumnIndexFromX,
  PX_PER_HOUR,
  MIN_BLOCK_HEIGHT_PX,
} from '../calendarLayout'

describe('topOffsetPx', () => {
  it('converts a wall-clock time to a pixel offset', () => {
    expect(topOffsetPx('00:00')).toBe(0)
    expect(topOffsetPx('09:30', 60)).toBe(9.5 * 60)
    expect(topOffsetPx('12:00', 30)).toBe(12 * 30)
  })

  it('returns 0 for an invalid time', () => {
    expect(topOffsetPx('')).toBe(0)
    expect(topOffsetPx('bad')).toBe(0)
  })
})

describe('heightPx', () => {
  it('computes proportional height for a same-day span', () => {
    expect(heightPx('09:00', '10:00', 60)).toBe(60)
    expect(heightPx('09:00', '09:30', 60)).toBe(30)
  })

  it('floors short entries at MIN_BLOCK_HEIGHT_PX', () => {
    expect(heightPx('09:00', '09:05', 60)).toBe(MIN_BLOCK_HEIGHT_PX)
  })

  it('floors zero/negative spans (e.g. end <= start) at MIN_BLOCK_HEIGHT_PX', () => {
    expect(heightPx('09:00', '09:00', 60)).toBe(MIN_BLOCK_HEIGHT_PX)
    expect(heightPx('10:00', '09:00', 60)).toBe(MIN_BLOCK_HEIGHT_PX)
  })
})

describe('pxToWallClock', () => {
  it('round-trips a time already on the snap grid', () => {
    expect(pxToWallClock(topOffsetPx('09:30', PX_PER_HOUR), PX_PER_HOUR, 15)).toBe('09:30')
  })

  it('snaps to the nearest grid line', () => {
    expect(pxToWallClock(topOffsetPx('09:37', PX_PER_HOUR), PX_PER_HOUR, 15)).toBe('09:30')
    expect(pxToWallClock(topOffsetPx('09:38', PX_PER_HOUR), PX_PER_HOUR, 15)).toBe('09:45')
  })

  it('clamps to the start and end of the day', () => {
    expect(pxToWallClock(-100, PX_PER_HOUR, 15)).toBe('00:00')
    expect(pxToWallClock(100000, PX_PER_HOUR, 15)).toBe('23:45')
  })
})

describe('dayColumnIndexFromX', () => {
  const rects = [
    { left: 0, right: 100 },
    { left: 100, right: 200 },
    { left: 200, right: 300 },
  ]

  it('finds the column containing x', () => {
    expect(dayColumnIndexFromX(50, rects)).toBe(0)
    expect(dayColumnIndexFromX(150, rects)).toBe(1)
    expect(dayColumnIndexFromX(250, rects)).toBe(2)
  })

  it('clamps out-of-range x to the nearest edge column', () => {
    expect(dayColumnIndexFromX(-50, rects)).toBe(0)
    expect(dayColumnIndexFromX(500, rects)).toBe(2)
  })

  it('returns 0 when there are no columns', () => {
    expect(dayColumnIndexFromX(50, [])).toBe(0)
  })
})
