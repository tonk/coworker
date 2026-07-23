import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import en from '@/i18n/en.json'
import TimeTrackingCalendarBlock from '../TimeTrackingCalendarBlock.vue'

const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })

const weekDays = [
  { iso: '2026-07-20' }, { iso: '2026-07-21' }, { iso: '2026-07-22' },
  { iso: '2026-07-23' }, { iso: '2026-07-24' }, { iso: '2026-07-25' }, { iso: '2026-07-26' },
]

const entry = {
  id: 1,
  date: '2026-07-21T00:00:00Z',
  start_time: '09:00',
  end_time: '10:30',
  description: 'Bugfixing',
}

function mountBlock(props = {}) {
  return mount(TimeTrackingCalendarBlock, {
    props: {
      entry,
      top: 100,
      height: 90,
      pxPerHour: 60,
      dayIndex: 1,
      weekDays,
      columnRects: [],
      customerName: 'Acme Corp',
      projectName: 'Website redesign',
      ...props,
    },
    global: { plugins: [i18n] },
  })
}

beforeEach(() => {
  setActivePinia(createPinia())
  // jsdom doesn't implement the Pointer Events capture API used by the drag/resize handlers.
  if (!Element.prototype.setPointerCapture) {
    Element.prototype.setPointerCapture = () => {}
    Element.prototype.releasePointerCapture = () => {}
  }
})

describe('TimeTrackingCalendarBlock', () => {
  it('renders its style from the top/height props', () => {
    const wrapper = mountBlock()
    const el = wrapper.get('.cal-block')
    expect(el.attributes('style')).toContain('top: 100px')
    expect(el.attributes('style')).toContain('height: 90px')
  })

  it('emits open on Enter', async () => {
    const wrapper = mountBlock()
    await wrapper.get('.cal-block').trigger('keydown.enter')
    expect(wrapper.emitted('open')).toBeTruthy()
    expect(wrapper.emitted('open')[0][0]).toEqual(entry)
  })

  it('hides resize handles from the tab order and from assistive tech', () => {
    const wrapper = mountBlock()
    const handles = wrapper.findAll('.cal-resize-handle')
    expect(handles.length).toBe(2)
    for (const handle of handles) {
      expect(handle.attributes('tabindex')).toBeUndefined()
      expect(handle.attributes('aria-hidden')).toBe('true')
    }
  })

  it('is keyboard-focusable and has a non-empty accessible label', () => {
    const wrapper = mountBlock()
    const el = wrapper.get('.cal-block')
    expect(el.attributes('tabindex')).toBe('0')
    expect(el.attributes('aria-label')).toBeTruthy()
    expect(el.attributes('aria-label')).toContain('Acme Corp')
    expect(el.attributes('aria-label')).toContain('Bugfixing')
  })

  it('is not tabbable when read-only', () => {
    const wrapper = mountBlock({ readOnly: true })
    expect(wrapper.get('.cal-block').attributes('tabindex')).toBe('-1')
    expect(wrapper.findAll('.cal-resize-handle').length).toBe(0)
  })

  it('hides the bottom handle on a split "start" segment (its bottom edge is midnight, not the real end)', () => {
    const wrapper = mountBlock({ segment: 'start' })
    expect(wrapper.find('.cal-resize-top').exists()).toBe(true)
    expect(wrapper.find('.cal-resize-bottom').exists()).toBe(false)
  })

  it('hides the top handle on a split "continuation" segment (its top edge is midnight, not the real start)', () => {
    const wrapper = mountBlock({ segment: 'continuation' })
    expect(wrapper.find('.cal-resize-top').exists()).toBe(false)
    expect(wrapper.find('.cal-resize-bottom').exists()).toBe(true)
  })

  function firePointer(el, type, opts) {
    el.dispatchEvent(new PointerEvent(type, { bubbles: true, cancelable: true, button: 0, ...opts }))
  }

  it('computes overnight-aware minutes when resizing the top handle of a split "start" segment', () => {
    // 19:00 -> 07:00 (next day), rendered as a 'start' segment from 19:00 to midnight.
    const overnightEntry = { id: 2, date: '2026-07-21T00:00:00Z', start_time: '19:00', end_time: '07:00', description: 'Standby' }
    const wrapper = mountBlock({ entry: overnightEntry, top: 19 * 60, height: 5 * 60, segment: 'start' })
    const blockEl = wrapper.get('.cal-block').element
    const topHandle = wrapper.get('.cal-resize-top').element

    // Drag the start edge up by one hour (60px at pxPerHour=60): 19:00 -> 18:00.
    firePointer(topHandle, 'pointerdown', { clientY: 19 * 60, pointerId: 1 })
    firePointer(blockEl, 'pointermove', { clientY: 18 * 60, pointerId: 1 })
    firePointer(blockEl, 'pointerup', { clientY: 18 * 60, pointerId: 1 })

    const payload = wrapper.emitted('resize')[0][0]
    expect(payload.newStartTime).toBe('18:00')
    expect(payload.newEndTime).toBe('07:00')
    // 18:00 -> 07:00 next day = 13 hours, NOT the ~15 minutes a same-day-only
    // calculation would (wrongly) produce.
    expect(payload.newMinutes).toBe(13 * 60)
  })

  it('computes overnight-aware minutes when resizing the bottom handle of a split "continuation" segment', () => {
    // Same entry, rendered as the 'continuation' segment from midnight to 07:00.
    const overnightEntry = { id: 2, date: '2026-07-21T00:00:00Z', start_time: '19:00', end_time: '07:00', description: 'Standby' }
    const wrapper = mountBlock({ entry: overnightEntry, top: 0, height: 7 * 60, segment: 'continuation' })
    const blockEl = wrapper.get('.cal-block').element
    const bottomHandle = wrapper.get('.cal-resize-bottom').element

    // Drag the end edge down by one hour: 07:00 -> 08:00.
    firePointer(bottomHandle, 'pointerdown', { clientY: 7 * 60, pointerId: 1 })
    firePointer(blockEl, 'pointermove', { clientY: 8 * 60, pointerId: 1 })
    firePointer(blockEl, 'pointerup', { clientY: 8 * 60, pointerId: 1 })

    const payload = wrapper.emitted('resize')[0][0]
    expect(payload.newStartTime).toBe('19:00')
    expect(payload.newEndTime).toBe('08:00')
    expect(payload.newMinutes).toBe(13 * 60)
  })
})
