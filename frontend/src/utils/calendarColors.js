// Per-customer palette for calendar blocks. A customer's own `color` (set on the
// Customer/time-tracking-customer record) always wins; customers left without one
// get assigned the first palette color not already used by another customer, so
// two customers never end up visually identical unless the palette itself runs out.
// All colors are Tailwind ~600-level shades chosen for solid contrast with white text.
const PALETTE = [
  '#2563eb', // blue
  '#059669', // emerald
  '#dc2626', // red
  '#7c3aed', // violet
  '#d97706', // amber
  '#0d9488', // teal
  '#db2777', // pink
  '#0891b2', // cyan
  '#ea580c', // orange
  '#a21caf', // fuchsia
]

export const NO_CUSTOMER_COLOR = '#64748b' // slate

/** Deterministic hash-based fallback, used only once the palette is exhausted. */
function hashColor(customerId) {
  const n = Number(customerId)
  if (!Number.isFinite(n)) return NO_CUSTOMER_COLOR
  return PALETTE[Math.abs(n) % PALETTE.length]
}

export function colorForCustomer(customerId) {
  if (customerId == null) return NO_CUSTOMER_COLOR
  return hashColor(customerId)
}

/**
 * Builds a Map<customerId, color> for a list of customers: customers with an
 * explicit `.color` keep it; the rest are assigned the first palette color not
 * already taken (by an explicit color or an earlier auto-assignment), processed
 * in id order so the result is stable across re-renders for the same customer set.
 */
export function assignCustomerColors(customers) {
  const map = new Map()
  const used = new Set()

  const withColor = []
  const withoutColor = []
  for (const c of customers || []) {
    if (c.color) withColor.push(c)
    else withoutColor.push(c)
  }

  for (const c of withColor) {
    map.set(c.id, c.color)
    used.add(c.color)
  }

  for (const c of [...withoutColor].sort((a, b) => a.id - b.id)) {
    const next = PALETTE.find((color) => !used.has(color))
    const color = next || hashColor(c.id)
    map.set(c.id, color)
    used.add(color)
  }

  return map
}
