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
 * Builds a Map<id, color> for a list of items (customers or projects): items with
 * an explicit `.color` keep it; the rest are assigned the first palette color not
 * already taken (by an explicit color or an earlier auto-assignment), processed
 * in id order so the result is stable across re-renders for the same item set.
 */
function assignColors(items) {
  const map = new Map()
  const used = new Set()

  const withColor = []
  const withoutColor = []
  for (const item of items || []) {
    if (item.color) withColor.push(item)
    else withoutColor.push(item)
  }

  for (const item of withColor) {
    map.set(item.id, item.color)
    used.add(item.color)
  }

  for (const item of [...withoutColor].sort((a, b) => a.id - b.id)) {
    const next = PALETTE.find((color) => !used.has(color))
    const color = next || hashColor(item.id)
    map.set(item.id, color)
    used.add(color)
  }

  return map
}

export function assignCustomerColors(customers) {
  return assignColors(customers)
}

export function assignProjectColors(projects) {
  return assignColors(projects)
}
