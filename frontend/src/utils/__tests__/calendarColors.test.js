import { describe, it, expect } from 'vitest'
import { colorForCustomer } from '../calendarColors'

describe('colorForCustomer', () => {
  it('returns a stable color for the same customer id', () => {
    expect(colorForCustomer(7)).toBe(colorForCustomer(7))
  })

  it('returns different colors for different ids (within the palette size)', () => {
    expect(colorForCustomer(1)).not.toBe(colorForCustomer(2))
  })

  it('returns a fixed color for no customer', () => {
    expect(colorForCustomer(null)).toBe(colorForCustomer(undefined))
  })

  it('returns a valid hex color', () => {
    expect(colorForCustomer(42)).toMatch(/^#[0-9a-f]{6}$/i)
    expect(colorForCustomer(null)).toMatch(/^#[0-9a-f]{6}$/i)
  })
})
