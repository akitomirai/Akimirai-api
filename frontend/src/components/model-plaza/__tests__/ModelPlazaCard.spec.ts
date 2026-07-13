import { beforeEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ModelPlazaCard from '../ModelPlazaCard.vue'
import { i18n } from '@/i18n'
import type { ModelCatalogItem } from '@/utils/modelCatalog'

const item: ModelCatalogItem = {
  id: 'gpt-5',
  displayName: 'GPT 5',
  modelId: 'gpt-5',
  provider: 'openai',
  platform: 'openai',
  family: 'gpt',
  status: 'available',
  statusReason: 'Available from two sources',
  billingMultiplier: 1,
  billingDescription: '1x',
  availableChannelCount: 2,
  quickStartUrl: '/quick-start?model=gpt-5',
  updatedAt: '2026-07-13T00:00:00Z',
  channelNames: ['channel-a', 'channel-b'],
  groups: [],
  pricing: null,
  offers: [
    {
      channel: 'channel-a',
      platform: 'openai',
      groups: [{
        id: 1,
        name: 'Pro',
        platform: 'openai',
        subscriptionType: 'standard',
        rateMultiplier: 1,
        effectiveRateMultiplier: 1,
        peakRateEnabled: false,
        peakStart: '',
        peakEnd: '',
        peakRateMultiplier: 0,
        isExclusive: false,
      }],
      pricing: {
        billing_mode: 'token', input_price: 0.000005, output_price: 0.00003,
        cache_write_price: null, cache_read_price: 0.0000005, image_output_price: null,
        per_request_price: null, intervals: [],
      },
    },
    {
      channel: 'channel-b',
      platform: 'openai',
      groups: [{
        id: 2,
        name: 'Basic',
        platform: 'openai',
        subscriptionType: 'standard',
        rateMultiplier: 0.8,
        effectiveRateMultiplier: 0.8,
        peakRateEnabled: false,
        peakStart: '',
        peakEnd: '',
        peakRateMultiplier: 0,
        isExclusive: false,
      }],
      pricing: {
        billing_mode: 'token', input_price: 0.000004, output_price: 0.000025,
        cache_write_price: null, cache_read_price: 0.0000004, image_output_price: null,
        per_request_price: null, intervals: [],
      },
    },
  ],
  supportsStreaming: true,
  supportsVision: false,
  supportsTools: true,
  supportsJson: true,
  contextWindow: 128000,
  recommendedUse: null,
}

describe('ModelPlazaCard', () => {
  beforeEach(() => {
    i18n.global.locale.value = 'en'
  })

  it('shows offer-aware pricing and emits copy/detail actions', async () => {
    const wrapper = mount(ModelPlazaCard, {
      props: { item },
      global: {
        stubs: {
          Icon: { template: '<span />' },
          ModelIcon: { template: '<span />' },
        },
      },
    })

    expect(wrapper.text()).toContain('2 channel-specific prices')
    expect(wrapper.text()).toContain('Pro')
    expect(wrapper.text()).toContain('Basic')

    await wrapper.get('button[title="Copy model ID"]').trigger('click')
    await wrapper.findAll('button').at(-1)!.trigger('click')
    expect(wrapper.emitted('copy')).toEqual([['gpt-5']])
    expect(wrapper.emitted('details')?.[0]?.[0]).toMatchObject({ modelId: 'gpt-5' })
  })
})
