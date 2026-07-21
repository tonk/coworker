import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import en from '@/i18n/en.json'
import TimeEntryModal from '../TimeEntryModal.vue'

const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })

const regularCustomer = { id: 1, name: 'Acme Corp' }
const ttCustomer = { id: 2, name: 'Personal' }
const regularProject = { id: 10, customer_id: 1, name: 'Website redesign' }
const ttProject = { id: 20, name: 'Internal' }

function mountModal(props = {}) {
  return mount(TimeEntryModal, {
    props: {
      entry: null,
      prefill: { date: '2026-07-21', start_time: '09:00', end_time: '10:00' },
      allCustomers: [regularCustomer, ttCustomer],
      allProjects: [regularProject, ttProject],
      ttCustomers: [ttCustomer],
      ttProjects: [ttProject],
      projects: [regularProject],
      ...props,
    },
    global: { plugins: [i18n], stubs: { teleport: true, Teleport: true } },
    attachTo: document.body,
  })
}

beforeEach(() => {
  setActivePinia(createPinia())
})

describe('TimeEntryModal', () => {
  it('shows only time-tracking-only projects when a time-tracking-only customer is selected', async () => {
    const wrapper = mountModal()
    const customerSelect = wrapper.get('#te-customer')
    await customerSelect.setValue(String(ttCustomer.id))
    const options = wrapper.get('#te-project').findAll('option')
    const labels = options.map((o) => o.text())
    expect(labels).toContain('Internal')
    expect(labels).not.toContain('Website redesign')
  })

  it("shows a regular customer's own projects plus time-tracking-only ones", async () => {
    const wrapper = mountModal()
    const customerSelect = wrapper.get('#te-customer')
    await customerSelect.setValue(String(regularCustomer.id))
    const labels = wrapper.get('#te-project').findAll('option').map((o) => o.text())
    expect(labels).toContain('Website redesign')
    expect(labels).toContain('Internal')
  })

  it('disables Save when end time is not after start time', async () => {
    const wrapper = mountModal({ prefill: { date: '2026-07-21', start_time: '10:00', end_time: '09:00' } })
    const saveBtn = wrapper.findAll('button').find((b) => b.text() === 'Save')
    expect(saveBtn.attributes('disabled')).toBeDefined()
  })

  it('emits a save payload shaped for the time-entries API', async () => {
    const wrapper = mountModal()
    const saveBtn = wrapper.findAll('button').find((b) => b.text() === 'Save')
    await saveBtn.trigger('click')
    const payload = wrapper.emitted('save')[0][0]
    expect(payload).toMatchObject({
      customer_id: null,
      project_id: null,
      date: '2026-07-21',
      start_time: '09:00',
      end_time: '10:00',
      minutes: 60,
    })
    expect(payload).toHaveProperty('description')
    expect(payload).toHaveProperty('distance')
  })

  it('emits delete with the entry when editing an existing entry', async () => {
    const existing = { id: 5, date: '2026-07-21', start_time: '09:00', end_time: '10:00', description: 'x' }
    const wrapper = mountModal({ entry: existing, prefill: null })
    const deleteBtn = wrapper.findAll('button').find((b) => b.text() === 'Delete')
    expect(deleteBtn).toBeTruthy()
    await deleteBtn.trigger('click')
    expect(wrapper.emitted('delete')[0][0]).toEqual(existing)
  })
})
