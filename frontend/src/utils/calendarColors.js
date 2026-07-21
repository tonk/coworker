// Deterministic per-customer palette for calendar blocks — same customer always
// gets the same color across sessions/reloads without needing a stored preference.
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

const NO_CUSTOMER_COLOR = '#64748b' // slate

export function colorForCustomer(customerId) {
  if (customerId == null) return NO_CUSTOMER_COLOR
  const n = Number(customerId)
  if (!Number.isFinite(n)) return NO_CUSTOMER_COLOR
  return PALETTE[Math.abs(n) % PALETTE.length]
}
