import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import PricingEntryCard from '../PricingEntryCard.vue'
import ModelTagInput from '../ModelTagInput.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

describe('PricingEntryCard video pricing templates', () => {
  it('creates fixed 768P and 2K per-second tiers when MiniMax-H3 is added', async () => {
    const wrapper = mount(PricingEntryCard, {
      props: {
        platform: 'openai',
        entry: {
          models: [],
          billing_mode: 'token',
          input_price: null,
          output_price: null,
          cache_write_price: null,
          cache_read_price: null,
          image_input_price: null,
          image_output_price: null,
          per_request_price: null,
          intervals: [],
        },
      },
      global: {
        stubs: {
          Icon: true,
          Select: true,
          IntervalRow: true,
        },
      },
    })

    wrapper.getComponent(ModelTagInput).vm.$emit('update:models', ['MiniMax-H3'])
    await wrapper.vm.$nextTick()

    const updates = wrapper.emitted('update')
    expect(updates).toHaveLength(1)
    expect(updates?.[0]?.[0]).toMatchObject({
      models: ['MiniMax-H3'],
      billing_mode: 'video',
      per_request_price: null,
      intervals: [
        { tier_label: '768P', per_request_price: null },
        { tier_label: '2K', per_request_price: null },
      ],
    })
  })

  it('creates 480p, 720p, and 1080p per-second tiers for firefly-video-v2', async () => {
    const wrapper = mount(PricingEntryCard, {
      props: {
        platform: 'openai',
        entry: {
          models: [],
          billing_mode: 'token',
          input_price: null,
          output_price: null,
          cache_write_price: null,
          cache_read_price: null,
          image_input_price: null,
          image_output_price: null,
          per_request_price: null,
          intervals: [],
        },
      },
      global: {
        stubs: {
          Icon: true,
          Select: true,
          IntervalRow: true,
        },
      },
    })

    wrapper.getComponent(ModelTagInput).vm.$emit('update:models', ['firefly-video-v2'])
    await wrapper.vm.$nextTick()

    const updates = wrapper.emitted('update')
    expect(updates).toHaveLength(1)
    expect(updates?.[0]?.[0]).toMatchObject({
      models: ['firefly-video-v2'],
      billing_mode: 'video',
      per_request_price: null,
      intervals: [
        { tier_label: '480p', per_request_price: null },
        { tier_label: '720p', per_request_price: null },
        { tier_label: '1080p', per_request_price: null },
      ],
    })
  })

  it('creates only 480p and 720p per-second tiers for firefly-video-v2-fast', async () => {
    const wrapper = mount(PricingEntryCard, {
      props: {
        platform: 'openai',
        entry: {
          models: [],
          billing_mode: 'token',
          input_price: null,
          output_price: null,
          cache_write_price: null,
          cache_read_price: null,
          image_input_price: null,
          image_output_price: null,
          per_request_price: null,
          intervals: [],
        },
      },
      global: {
        stubs: {
          Icon: true,
          Select: true,
          IntervalRow: true,
        },
      },
    })

    wrapper.getComponent(ModelTagInput).vm.$emit('update:models', ['firefly-video-v2-fast'])
    await wrapper.vm.$nextTick()

    const updates = wrapper.emitted('update')
    expect(updates).toHaveLength(1)
    expect(updates?.[0]?.[0]).toMatchObject({
      models: ['firefly-video-v2-fast'],
      billing_mode: 'video',
      per_request_price: null,
      intervals: [
        { tier_label: '480p', per_request_price: null },
        { tier_label: '720p', per_request_price: null },
      ],
    })
  })

  it('migrates an existing unified SD2.0 price when a resolution price is entered', async () => {
    const wrapper = mount(PricingEntryCard, {
      props: {
        platform: 'openai',
        entry: {
          models: ['firefly-video-v2'],
          billing_mode: 'video',
          input_price: null,
          output_price: null,
          cache_write_price: null,
          cache_read_price: null,
          image_input_price: null,
          image_output_price: null,
          per_request_price: '0.17',
          intervals: [],
        },
      },
      global: {
        stubs: {
          Icon: true,
          Select: true,
          IntervalRow: true,
        },
      },
    })

    await wrapper.get('[data-testid="sd20-video-price-480p"]').setValue('0.18')

    expect(wrapper.emitted('update')?.at(-1)?.[0]).toMatchObject({
      models: ['firefly-video-v2'],
      billing_mode: 'video',
      per_request_price: null,
      intervals: [{ tier_label: '480p', per_request_price: '0.18' }],
    })
  })
})
