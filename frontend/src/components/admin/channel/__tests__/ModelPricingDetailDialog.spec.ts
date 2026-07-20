import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ModelPricingDetailDialog from '../ModelPricingDetailDialog.vue'
import type { AdminModelPricingCard } from '@/utils/adminModelPricing'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (_key: string, fallback?: string) => fallback || _key,
  }),
}))

const model: AdminModelPricingCard = {
  key: 'openai::gpt-*',
  model: 'gpt-*',
  platform: 'openai',
  isPattern: true,
  activeSourceCount: 1,
  disabledSourceCount: 0,
  billingModes: ['token'],
  sources: [{
    key: '42:5:gpt-*',
    channelId: 42,
    channelName: 'OpenAI source',
    channelDescription: 'Primary route',
    channelStatus: 'active',
    billingModelSource: 'channel_mapped',
    groupIds: [7, 8],
    restrictModels: true,
    ruleIndex: 0,
    sharedModels: ['gpt-*', 'o3-*'],
    pricing: {
      id: 5,
      platform: 'openai',
      models: ['gpt-*', 'o3-*'],
      billing_mode: 'token',
      input_price: 0.000005,
      output_price: 0.00003,
      cache_write_price: 0.00000625,
      cache_read_price: 0.0000005,
      image_output_price: null,
      per_request_price: null,
      intervals: [{
        id: 9,
        min_tokens: 0,
        max_tokens: 272000,
        tier_label: 'standard',
        input_price: 0.000005,
        output_price: 0.00003,
        cache_write_price: null,
        cache_read_price: 0.0000005,
        per_request_price: null,
        sort_order: 0,
      }],
    },
  }],
}

describe('ModelPricingDetailDialog', () => {
  it('shows every source rule relationship and emits the selected channel edit', async () => {
    const wrapper = mount(ModelPricingDetailDialog, {
      props: {
        show: true,
        model,
        groups: {
          7: { name: 'Pro pool', platform: 'openai' },
          8: { name: 'Claude pool', platform: 'anthropic' },
        },
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title'],
            emits: ['close'],
            template: '<div v-if="show"><h1>{{ title }}</h1><slot /><slot name="footer" /></div>',
          },
          Icon: { template: '<span />' },
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('OpenAI source')
    expect(text).toContain('gpt-*')
    expect(text).toContain('o3-*')
    expect(text).toContain('Pro pool')
    expect(text).not.toContain('Claude pool')
    expect(text).toContain('$5 / 1M')
    expect(text).toContain('$30 / 1M')
    expect(text).toContain('0-272000')

    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('edit-source')).toEqual([[42]])
  })

  it('renders image output pricing as a per-image amount', () => {
    const imageModel: AdminModelPricingCard = {
      ...model,
      billingModes: ['image'],
      sources: [{
        ...model.sources[0],
        pricing: {
          ...model.sources[0].pricing,
          billing_mode: 'image',
          input_price: null,
          output_price: null,
          image_output_price: 0.1,
          intervals: [],
        },
      }],
    }
    const wrapper = mount(ModelPricingDetailDialog, {
      props: { show: true, model: imageModel, groups: {} },
      global: {
        stubs: {
          BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' },
          Icon: { template: '<span />' },
        },
      },
    })

    expect(wrapper.text()).toContain('$0.1')
    expect(wrapper.text()).not.toContain('$100000 / 1M')
  })
})
