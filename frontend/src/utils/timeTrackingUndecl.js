/** Undeclarable minutes for a single time entry (capped at entry length). */
export function entryUndeclMins(entry) {
  const pu = entry?.project?.undeclarable_minutes || 0
  if (pu <= 0 || !entry?.minutes) return 0
  return Math.min(entry.minutes, pu)
}

/** Declarable minutes for a week row from its day entries. */
export function rowDeclarableMins(dayEntries) {
  const raw = dayEntries.reduce((s, e) => s + (e?.minutes || 0), 0)
  if (raw === 0) return 0
  const undecl = dayEntries.reduce((s, e) => s + entryUndeclMins(e), 0)
  return Math.max(0, raw - undecl)
}
