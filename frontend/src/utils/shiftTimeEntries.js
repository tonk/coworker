const MINUTES_PER_DAY = 24 * 60

export function parseWallClock(s) {
  if (!s) return -1
  const trimmed = s.trim()
  const m = trimmed.match(/^(\d{1,2}):(\d{2})$/)
  if (!m) return -1
  const h = parseInt(m[1], 10)
  const min = parseInt(m[2], 10)
  if (h > 23 || min > 59) return -1
  return h * 60 + min
}

export function fmtWallClock(mins) {
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`
}

/** Minutes between two wall-clock times on the same day (supports overnight when end <= start). */
export function wallClockSpanMinutes(startM, endM) {
  if (startM < 0 || endM < 0 || startM === endM) return null
  if (endM > startM) return endM - startM
  return (MINUTES_PER_DAY - startM) + endM
}

export function addDaysISO(iso, days) {
  const d = new Date(`${iso}T12:00:00`)
  d.setDate(d.getDate() + days)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

/** Split a multi-day shift into per-calendar-day time entries. */
export function splitShiftIntoDayEntries(startDateISO, startTime, endDateISO, endTime) {
  const startM = parseWallClock(startTime)
  const endM = parseWallClock(endTime)
  if (startM < 0 || endM < 0 || !startDateISO || !endDateISO) return []
  if (endDateISO < startDateISO) return []

  const segments = []
  let curISO = startDateISO

  while (curISO <= endDateISO) {
    const isFirst = curISO === startDateISO
    const isLast = curISO === endDateISO

    if (isFirst && isLast) {
      const minutes = wallClockSpanMinutes(startM, endM)
      if (minutes == null || minutes <= 0) break
      segments.push({
        date: curISO,
        start_time: fmtWallClock(startM),
        end_time: fmtWallClock(endM),
        minutes,
      })
    } else if (isFirst) {
      const minutes = MINUTES_PER_DAY - startM
      if (minutes > 0) {
        segments.push({
          date: curISO,
          start_time: fmtWallClock(startM),
          end_time: '23:59',
          minutes,
        })
      }
    } else if (isLast) {
      if (endM > 0) {
        segments.push({
          date: curISO,
          start_time: '00:00',
          end_time: fmtWallClock(endM),
          minutes: endM,
        })
      }
    } else {
      segments.push({
        date: curISO,
        start_time: '00:00',
        end_time: '23:59',
        minutes: MINUTES_PER_DAY,
      })
    }

    if (curISO === endDateISO) break
    curISO = addDaysISO(curISO, 1)
  }

  return segments
}

/** Friday 19:00 → Monday 07:00 relative to a week whose days are ISO date strings (index 0 = week start). */
export function weekendStandbyDefaults(weekDayISOs) {
  let fridayISO = null
  for (const iso of weekDayISOs) {
    if (new Date(`${iso}T12:00:00`).getDay() === 5) {
      fridayISO = iso
      break
    }
  }
  if (!fridayISO) {
    fridayISO = weekDayISOs[Math.min(4, weekDayISOs.length - 1)]
  }
  return {
    start_date: fridayISO,
    start_time: '19:00',
    end_date: addDaysISO(fridayISO, 3),
    end_time: '07:00',
  }
}
