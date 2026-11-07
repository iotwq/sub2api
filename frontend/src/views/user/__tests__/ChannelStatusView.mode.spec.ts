import { describe, expect, it, vi, beforeEach } from 'vitest'
import { defineComponent, h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

const isV1 = vi.fn(() => false)
const loaded = vi.hoisted(() => ({ v1: 0, v2: 0 }))

vi.mock('@/utils/featureFlags', () => ({
  isChannelMonitorV1Mode: () => isV1(),
}))

vi.mock('../ChannelStatusV1View.vue', () => {
  loaded.v1++
  return { __esModule: true, default: defineComponent({ name: 'ChannelStatusV1View', setup: () => () => h('div', { 'data-testid': 'v1' }) }) }
})
vi.mock('../ChannelStatusV2View.vue', () => {
  loaded.v2++
  return { __esModule: true, default: defineComponent({ name: 'ChannelStatusV2View', setup: () => () => h('div', { 'data-testid': 'v2' }) }) }
})

import ChannelStatusView from '../ChannelStatusView.vue'

describe('ChannelStatusView mode switch', () => {
  beforeEach(() => {
    isV1.mockReset()
  })

  it('renders V2 when not in v1 mode', async () => {
    isV1.mockReturnValue(false)
    const inactiveLoads = loaded.v1
    const wrapper = mount(ChannelStatusView)
    await flushPromises()
    expect(wrapper.find('[data-testid="v2"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="v1"]').exists()).toBe(false)
    expect(loaded.v1).toBe(inactiveLoads)
    wrapper.unmount()
  })

  it('renders V1 when in v1 mode', async () => {
    isV1.mockReturnValue(true)
    const inactiveLoads = loaded.v2
    const wrapper = mount(ChannelStatusView)
    await flushPromises()
    expect(wrapper.find('[data-testid="v1"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="v2"]').exists()).toBe(false)
    expect(loaded.v2).toBe(inactiveLoads)
    wrapper.unmount()
  })
})
