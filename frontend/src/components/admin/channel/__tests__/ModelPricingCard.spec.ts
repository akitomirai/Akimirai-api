import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ModelPricingCard from '../ModelPricingCard.vue'
import type { AdminModelPricingCard } from '@/utils/adminModelPricing'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (_key: string, fallback?: string) => fallback || _key,
  }),
}))

describe('ModelPricingCard', () => {
  it('uses image output pricing for image-billed models', () => {
    const card: AdminModelPricingCard = {
      key: 'openai::gpt-image',
      model: 'gpt-image',
      platform: 'openai',
      isPattern: false,
      activeSourceCount: 1,
      disabledSourceCount: 0,
      billingModes: ['image'],
      sources: [{
        key: '1:0:gpt-image',
        channelId: 1,
        channelName: 'Images',
        channelDescription: '',
        channelStatus: 'active',
        billingModelSource: 'channel_mapped',
        groupIds: [],
        restrictModels: true,
        ruleIndex: 0,
        sharedModels: ['gpt-image'],
        pricing: {
          platform: 'openai',
          models: ['gpt-image'],
          billing_mode: 'image',
          input_price: null,
          output_price: null,
          cache_write_price: null,
          cache_read_price: null,
          image_output_price: 0.1,
          per_request_price: 9,
          intervals: [],
        },
      }],
    }

    const wrapper = mount(ModelPricingCard, {
      props: { card },
      global: {
        stubs: {
          Icon: { template: '<span />' },
          ModelIcon: { template: '<span />' },
        },
      },
    })

    expect(wrapper.text()).toContain('$0.1 / image')
    expect(wrapper.text()).not.toContain('$9 / image')
  })
})
