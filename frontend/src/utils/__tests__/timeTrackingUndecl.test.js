import { describe, it, expect } from 'vitest'
import { entryUndeclMins, rowDeclarableMins } from '../timeTrackingUndecl.js'

describe('timeTrackingUndecl', () => {
  it('caps undeclarable at entry minutes', () => {
    expect(entryUndeclMins({ minutes: 30, project: { undeclarable_minutes: 45 } })).toBe(30)
    expect(entryUndeclMins({ minutes: 120, project: { undeclarable_minutes: 45 } })).toBe(45)
    expect(entryUndeclMins({ minutes: 60, project: { undeclarable_minutes: 0 } })).toBe(0)
  })

  it('sums declarable minutes for a row', () => {
    const days = [
      { minutes: 480, project: { undeclarable_minutes: 60 } },
      { minutes: 120, project: { undeclarable_minutes: 45 } },
      null,
    ]
    expect(rowDeclarableMins(days)).toBe(420 + 75)
  })
})
