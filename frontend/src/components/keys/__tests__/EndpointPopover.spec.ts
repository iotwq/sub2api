import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const copyToClipboard = vi.fn().mockResolvedValue(true)

const messages: Record<string, string> = {
  'keys.endpoints.title': '直连地址',
  'keys.endpoints.optimized': '国内优化地址',
  'keys.endpoints.default': '默认',
  'keys.endpoints.copied': '已复制',
  'keys.endpoints.copiedHint': '已复制到剪贴板',
  'keys.endpoints.clickToCopy': '点击可复制此端点',
  'keys.endpoints.speedTest': '测速',
}

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => messages[key] ?? key,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

import EndpointPopover from '../EndpointPopover.vue'

describe('EndpointPopover', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('将说明提示渲染到 URL 上方而不是旧的 title 图标上', () => {
    const wrapper = mount(EndpointPopover, {
      props: {
        apiBaseUrl: 'https://default.example.com/v1',
        customEndpoints: [
          {
            name: '备用线路',
            endpoint: 'https://backup.example.com/v1',
            description: '自定义说明',
          },
        ],
      },
    })

    expect(wrapper.text()).toContain('自定义说明')
    expect(wrapper.text()).toContain('点击可复制此端点')
    expect(wrapper.find('[role="button"]').attributes('title')).toBeUndefined()
    expect(wrapper.find('[title="自定义说明"]').exists()).toBe(false)
  })

  it('点击 URL 后会复制并切换为已复制提示', async () => {
    const wrapper = mount(EndpointPopover, {
      props: {
        apiBaseUrl: 'https://default.example.com/v1',
        customEndpoints: [],
      },
    })

    await wrapper.find('[role="button"]').trigger('click')
    await flushPromises()

    expect(copyToClipboard).toHaveBeenCalledWith('https://default.example.com/v1', '已复制')
    expect(wrapper.text()).toContain('已复制到剪贴板')
    expect(wrapper.find('button[aria-label="已复制到剪贴板"]').exists()).toBe(true)
  })

  it('分别展示直连和国内优化地址，且仅直连地址标记为默认', () => {
    const wrapper = mount(EndpointPopover, {
      props: {
        apiBaseUrl: 'https://direct.example.com',
        optimizedApiBaseUrl: 'https://optimized.example.com',
        customEndpoints: [],
      },
    })

    expect(wrapper.text()).toContain('直连地址')
    expect(wrapper.text()).toContain('https://direct.example.com')
    expect(wrapper.text()).toContain('国内优化地址')
    expect(wrapper.text()).toContain('https://optimized.example.com')
    expect(wrapper.findAll('span').filter((node) => node.text() === '默认')).toHaveLength(1)

    const optimizedSpeedTest = wrapper.findAll('a').find(
      (node) => node.attributes('href') ===
        `https://www.tcptest.cn/http/${encodeURIComponent('https://optimized.example.com')}`,
    )
    expect(optimizedSpeedTest).toBeDefined()
  })
})
