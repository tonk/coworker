import { describe, it, expect } from 'vitest'
import { filterProjectsForCustomer } from '../projectFilter'

const regularProjectA = { id: 1, customer_id: 10, name: 'Regular A' }
const regularProjectB = { id: 2, customer_id: 20, name: 'Regular B' }
const ttProjectA = { id: 3, name: 'TT Project A' }
const ttProjectB = { id: 4, name: 'TT Project B' }

const ctx = {
  allProjects: [regularProjectA, regularProjectB, ttProjectA, ttProjectB],
  ttCustomers: [{ id: 100, name: 'TT Customer' }],
  ttProjects: [ttProjectA, ttProjectB],
  projects: [regularProjectA, regularProjectB],
}

describe('filterProjectsForCustomer', () => {
  it('returns every project when no customer is selected', () => {
    expect(filterProjectsForCustomer(null, ctx)).toEqual(ctx.allProjects)
    expect(filterProjectsForCustomer(undefined, ctx)).toEqual(ctx.allProjects)
  })

  it('returns only time-tracking-only projects for a time-tracking-only customer', () => {
    expect(filterProjectsForCustomer(100, ctx)).toEqual(ctx.ttProjects)
  })

  it('returns a regular customer\'s own projects plus the time-tracking-only ones', () => {
    expect(filterProjectsForCustomer(10, ctx)).toEqual([regularProjectA, ttProjectA, ttProjectB])
  })

  it('returns only time-tracking-only projects for a regular customer with no projects of its own', () => {
    expect(filterProjectsForCustomer(30, ctx)).toEqual([ttProjectA, ttProjectB])
  })
})
