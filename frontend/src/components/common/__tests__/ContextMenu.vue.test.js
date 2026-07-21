import { describe, it, expect, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import ContextMenu from '../ContextMenu.vue'

const items = [
  { key: 'edit', label: 'Edit' },
  { key: 'delete', label: 'Delete', danger: true },
]

let wrapper

afterEach(() => {
  wrapper?.unmount()
  wrapper = undefined
})

async function mountMenu(props = {}) {
  wrapper = mount(ContextMenu, { props: { x: 10, y: 10, items, ...props }, attachTo: document.body })
  await wrapper.vm.$nextTick() // let the async onMounted (measure + attach doc listener) settle
  return wrapper
}

describe('ContextMenu', () => {
  it('renders a menu with menuitems for each item', async () => {
    await mountMenu()
    const menu = document.body.querySelector('[role="menu"]')
    expect(menu).toBeTruthy()
    const menuitems = document.body.querySelectorAll('[role="menuitem"]')
    expect(menuitems.length).toBe(2)
    expect(menuitems[0].textContent).toContain('Edit')
    expect(menuitems[1].textContent).toContain('Delete')
  })

  it('emits select with the item key when clicked', async () => {
    await mountMenu()
    const menuitems = document.body.querySelectorAll('[role="menuitem"]')
    await menuitems[1].dispatchEvent(new MouseEvent('click', { bubbles: true }))
    expect(wrapper.emitted('select')).toEqual([['delete']])
  })

  it('emits close on Escape', async () => {
    await mountMenu()
    const menu = document.body.querySelector('[role="menu"]')
    menu.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('emits close on an outside click', async () => {
    await mountMenu()
    document.body.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('moves the active item with ArrowDown/ArrowUp', async () => {
    await mountMenu()
    const menu = document.body.querySelector('[role="menu"]')
    menu.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }))
    await wrapper.vm.$nextTick()
    const menuitems = document.body.querySelectorAll('[role="menuitem"]')
    expect(menuitems[1].getAttribute('tabindex')).toBe('0')
    expect(menuitems[0].getAttribute('tabindex')).toBe('-1')
  })
})
