import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import CommunityEmojiPicker from '../CommunityEmojiPicker.vue'

const pickerMock = vi.hoisted(() => ({
  options: [] as Array<Record<string, unknown>>,
}))

vi.mock('emoji-picker-element', () => ({
  Picker: class {
    constructor(options: Record<string, unknown>) {
      pickerMock.options.push(options)
      const element = document.createElement('div')
      element.className = 'mock-emoji-picker'
      return element
    }
  },
}))

describe('CommunityEmojiPicker', () => {
  beforeEach(() => {
    pickerMock.options = []
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('opens a localized picker with local emoji data', async () => {
    const wrapper = mount(CommunityEmojiPicker, {
      attachTo: document.body,
      props: { locale: 'zh-CN', label: '选择表情' },
    })

    const trigger = wrapper.get('button')
    expect(trigger.attributes('aria-expanded')).toBe('false')
    await trigger.trigger('click')

    expect(trigger.attributes('aria-expanded')).toBe('true')
    expect(document.body.querySelector('[role="dialog"]')).not.toBeNull()
    expect(pickerMock.options).toHaveLength(1)
    expect(pickerMock.options[0]?.locale).toBe('zh')
    expect(String(pickerMock.options[0]?.dataSource)).toContain('data.json')
    expect((pickerMock.options[0]?.i18n as { favoritesLabel: string }).favoritesLabel).toBe('常用表情')

    wrapper.unmount()
  })

  it('emits the selected Unicode emoji and stays open for consecutive choices', async () => {
    const wrapper = mount(CommunityEmojiPicker, {
      attachTo: document.body,
      props: { locale: 'en', label: 'Choose emoji' },
    })

    await wrapper.get('button').trigger('click')
    const picker = document.body.querySelector('.mock-emoji-picker')
    picker?.dispatchEvent(new CustomEvent('emoji-click', {
      detail: { unicode: '👍' },
    }))

    expect(wrapper.emitted('select')).toEqual([['👍']])
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('true')

    wrapper.unmount()
  })

  it('closes on Escape and restores focus to the trigger', async () => {
    const wrapper = mount(CommunityEmojiPicker, {
      attachTo: document.body,
      props: { locale: 'en', label: 'Choose emoji' },
    })

    const trigger = wrapper.get('button')
    await trigger.trigger('click')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await wrapper.vm.$nextTick()

    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(trigger.element)

    wrapper.unmount()
  })

  it('closes when the user clicks outside the picker', async () => {
    const wrapper = mount(CommunityEmojiPicker, {
      attachTo: document.body,
      props: { locale: 'en', label: 'Choose emoji' },
    })

    const trigger = wrapper.get('button')
    await trigger.trigger('click')
    document.body.dispatchEvent(new MouseEvent('pointerdown', { bubbles: true }))
    await wrapper.vm.$nextTick()

    expect(trigger.attributes('aria-expanded')).toBe('false')

    wrapper.unmount()
  })

  it('fits above the composer trigger on a narrow mobile viewport', async () => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(320)
    vi.spyOn(window, 'innerHeight', 'get').mockReturnValue(568)
    const wrapper = mount(CommunityEmojiPicker, {
      attachTo: document.body,
      props: { locale: 'en', label: 'Choose emoji' },
    })
    const trigger = wrapper.get('button')
    vi.spyOn(trigger.element, 'getBoundingClientRect').mockReturnValue({
      x: 264,
      y: 520,
      top: 520,
      right: 304,
      bottom: 560,
      left: 264,
      width: 40,
      height: 40,
      toJSON: () => ({}),
    })

    await trigger.trigger('click')
    const popover = document.body.querySelector<HTMLElement>('.community-emoji-popover')

    expect(popover?.style.width).toBe('304px')
    expect(Number.parseInt(popover?.style.top || '', 10) + Number.parseInt(popover?.style.height || '', 10)).toBeLessThanOrEqual(510)

    wrapper.unmount()
    vi.restoreAllMocks()
  })
})
