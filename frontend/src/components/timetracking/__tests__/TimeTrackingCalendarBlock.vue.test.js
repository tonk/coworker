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
})
