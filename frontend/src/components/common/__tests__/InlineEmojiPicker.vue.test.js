import { describe, it, expect, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import InlineEmojiPicker from '../InlineEmojiPicker.vue'

let wrapper

afterEach(() => {
  wrapper?.unmount()
  wrapper = undefined
})

async function search(query) {
  wrapper = mount(InlineEmojiPicker)
  await wrapper.find('.emoji-search').setValue(query)
  return wrapper.findAll('.emoji-btn').map(b => b.text())
}

describe('InlineEmojiPicker ASCII emoticon search', () => {
  it('finds the smiley for ":-)"', async () => {
    const results = await search(':-)')
    expect(results).toContain('😊')
    expect(wrapper.find('.emoji-empty').exists()).toBe(false)
  })

  it('finds the smiley for the shorter ":)"', async () => {
    const results = await search(':)')
    expect(results).toContain('😊')
  })

  it('finds the heart for "<3"', async () => {
    const results = await search('<3')
    expect(results).toContain('❤️')
  })

  it('still shows "No results" for genuine gibberish', async () => {
    await search('zzzznotanemoji')
    expect(wrapper.find('.emoji-empty').exists()).toBe(true)
  })

  it('still finds emoji by keyword name (existing behavior)', async () => {
    const results = await search('smile')
    expect(results.length).toBeGreaterThan(0)
  })
})

describe('InlineEmojiPicker focus stealing', () => {
  it('does not steal focus into its own search box when opened via inline ":" typing (initialSearch pre-filled)', async () => {
    const hostInput = document.createElement('textarea')
    document.body.appendChild(hostInput)
    hostInput.focus()

    wrapper = mount(InlineEmojiPicker, { props: { initialSearch: 's' }, attachTo: document.body })
    await wrapper.vm.$nextTick()
    await new Promise(r => setTimeout(r, 0)) // let the onMounted nextTick settle

    expect(document.activeElement).toBe(hostInput)
    hostInput.remove()
  })

  it('does focus its own search box when opened with nothing pre-typed (toolbar button case)', async () => {
    wrapper = mount(InlineEmojiPicker, { props: { initialSearch: '' }, attachTo: document.body })
    await wrapper.vm.$nextTick()
    await new Promise(r => setTimeout(r, 0))

    expect(document.activeElement).toBe(wrapper.find('.emoji-search').element)
  })

  it('keeps its search box (and results) in sync as initialSearch changes after mount', async () => {
    wrapper = mount(InlineEmojiPicker, { props: { initialSearch: 's' } })
    await wrapper.setProps({ initialSearch: 'sm' })
    await wrapper.vm.$nextTick()

    expect(wrapper.find('.emoji-search').element.value).toBe('sm')
  })
})
