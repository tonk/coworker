import { describe, it, expect } from 'vitest'
import { parseTimeNotationMinutes, parseMacroTimeInput } from '../timeMacroInput'

describe('parseTimeNotationMinutes', () => {
  it('parses decimal hours', () => {
    expect(parseTimeNotationMinutes('1.5', 'decimal')).toBe(90)
    expect(parseTimeNotationMinutes('2', 'decimal')).toBe(120)
    expect(parseTimeNotationMinutes('', 'decimal')).toBe(0)
  })

  it('parses hhmm notation', () => {
    expect(parseTimeNotationMinutes('1:30', 'hhmm')).toBe(90)
    expect(parseTimeNotationMinutes('8', 'hhmm')).toBe(480)
    expect(parseTimeNotationMinutes('0:45', 'hhmm')).toBe(45)
  })
})

describe('parseMacroTimeInput', () => {
  it('always parses HH:MM regardless of notation', () => {
    expect(parseMacroTimeInput('0:45', 'decimal')).toBe(45)
    expect(parseMacroTimeInput('2:00', 'decimal')).toBe(120)
    expect(parseMacroTimeInput('1:30', 'hhmm')).toBe(90)
  })

  it('falls back to notation for decimal values', () => {
    expect(parseMacroTimeInput('1.5', 'decimal')).toBe(90)
    expect(parseMacroTimeInput('2', 'decimal')).toBe(120)
  })

  it('returns 0 for empty input', () => {
    expect(parseMacroTimeInput('', 'decimal')).toBe(0)
    expect(parseMacroTimeInput(null, 'decimal')).toBe(0)
  })
})
