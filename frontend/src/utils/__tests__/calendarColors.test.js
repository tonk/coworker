import { describe, it, expect } from 'vitest'
import { colorForCustomer, assignCustomerColors } from '../calendarColors'

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

describe('assignCustomerColors', () => {
  it("keeps a customer's own explicit color", () => {
    const map = assignCustomerColors([{ id: 1, color: '#123456' }])
    expect(map.get(1)).toBe('#123456')
  })

  it('assigns colorless customers a palette color not used by any explicit color', () => {
    const map = assignCustomerColors([
      { id: 1, color: '#2563eb' }, // takes the first palette color explicitly
      { id: 2, color: '' },        // no color set
    ])
    expect(map.get(2)).not.toBe('#2563eb')
  })

  it('never assigns the same auto-picked color to two different colorless customers (while the palette lasts)', () => {
    const customers = Array.from({ length: 5 }, (_, i) => ({ id: i + 1, color: '' }))
    const map = assignCustomerColors(customers)
    const colors = customers.map((c) => map.get(c.id))
    expect(new Set(colors).size).toBe(colors.length)
  })

  it('is stable across calls for the same input', () => {
    const customers = [{ id: 1, color: '' }, { id: 2, color: '' }, { id: 3, color: '#dc2626' }]
    const first = assignCustomerColors(customers)
    const second = assignCustomerColors(customers)
    for (const c of customers) expect(first.get(c.id)).toBe(second.get(c.id))
  })
})
