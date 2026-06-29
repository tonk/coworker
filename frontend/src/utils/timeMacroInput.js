/** Parse hours/minutes input according to the user's time notation preference. */
export function parseTimeNotationMinutes(val, notation = 'decimal') {
  if (!val && val !== 0) return 0
  const s = String(val).trim()
  if (!s) return 0
  if (notation === 'hhmm') {
    if (s.includes(':')) {
      const [h, m] = s.split(':')
      return (parseInt(h, 10) || 0) * 60 + (parseInt(m, 10) || 0)
    }
    return (parseInt(s, 10) || 0) * 60
  }
  return Math.round((parseFloat(s) || 0) * 60)
}

/** Macro templates store HH:MM or decimal hours — parse both regardless of user notation. */
export function parseMacroTimeInput(val, notation = 'decimal') {
  if (val === null || val === undefined) return 0
  const s = String(val).trim()
  if (!s) return 0
  if (s.includes(':')) {
    const [h, m] = s.split(':')
    return (parseInt(h, 10) || 0) * 60 + (parseInt(m, 10) || 0)
  }
  return parseTimeNotationMinutes(val, notation)
}
