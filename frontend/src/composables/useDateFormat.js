import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'

/**
 * Returns formatting helpers that respect the user's date_time_format setting.
 *
 * Supported format tokens:
 *   YYYY  – 4-digit year
 *   MM    – 2-digit month
 *   DD    – 2-digit day
 *   HH    – 24-hour hours (00-23)
 *   mm    – minutes
 *   hh    – 12-hour hours (01-12)
 *   a     – am/pm
 */

export function pad(n) {
  return String(n).padStart(2, '0')
}

export function applyFormat(date, fmt) {
  let d
  // ISO date-only strings (YYYY-MM-DD) must be parsed as local date to avoid UTC midnight shift
  if (typeof date === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(date)) {
    const [y, m, day] = date.split('-').map(Number)
    d = new Date(y, m - 1, day)
  } else {
    d = new Date(date)
  }
  if (isNaN(d)) return String(date)

  const YYYY = d.getFullYear()
  const MM   = pad(d.getMonth() + 1)
  const DD   = pad(d.getDate())
  const HH   = pad(d.getHours())
  const min  = pad(d.getMinutes())
  const h12  = d.getHours() % 12 || 12
  const hh   = pad(h12)
  const a    = d.getHours() < 12 ? 'am' : 'pm'

  return fmt
    .replace('YYYY', YYYY)
    .replace('MM',   MM)
    .replace('DD',   DD)
    .replace('HH',   HH)
    .replace('hh',   hh)
    .replace('mm',   min)
    .replace('a',    a)
}

/** Returns only the date portion of the user's format (strips time part). */
export function dateOnlyFmt(fmt) {
  // Remove the time portion: everything from a space followed by HH or hh onwards
  return fmt.replace(/\s+(HH:mm( a)?|hh:mm a)/, '').trim()
}

export function useDateFormat() {
  const auth = useAuthStore()
  const system = useSystemStore()

  function userFmt() {
    return auth.user?.date_time_format || system.defaults.date_time_format || 'YYYY-MM-DD HH:mm'
  }

  /** Format a full date+time value using the user's format. */
  function formatDateTime(date) {
    if (!date) return '—'
    return applyFormat(date, userFmt())
  }

  /** Format only the date portion (no time) using the user's format. */
  function formatDate(date) {
    if (!date) return '—'
    return applyFormat(date, dateOnlyFmt(userFmt()))
  }

  /** Format only the time portion (HH:mm or hh:mm a) regardless of format. */
  function formatTime(date) {
    if (!date) return '—'
    const d = new Date(date)
    if (isNaN(d)) return String(date)
    const fmt = userFmt()
    if (fmt.includes('hh') && fmt.includes('a')) {
      const h12 = d.getHours() % 12 || 12
      return `${pad(h12)}:${pad(d.getMinutes())} ${d.getHours() < 12 ? 'am' : 'pm'}`
    }
    return `${pad(d.getHours())}:${pad(d.getMinutes())}`
  }

  function dateOnlyFormat() {
    return dateOnlyFmt(userFmt())
  }

  return { formatDateTime, formatDate, formatTime, dateOnlyFormat }
}
