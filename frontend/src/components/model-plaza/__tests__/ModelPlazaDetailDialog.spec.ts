import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import ModelPlazaDetailDialog from '../ModelPlazaDetailDialog.vue'
import { i18n } from '@/i18n'
import type { ModelCatalogItem } from '@/utils/modelCatalog'

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn().mockResolvedValue(true) }),
}))

const model: ModelCatalogItem = {
  id: 'gpt-5', displayName: 'GPT 5', modelId: 'gpt-5', provider: 'openai', platform: 'openai',
  family: 'gpt', status: 'available', statusReason: 'Ready', billingMultiplier: 1,
  billingDescription: '1x', availableChannelCount: 2, quickStartUrl: '',
  updatedAt: '2026-07-13T00:00:00Z', channelNames: ['channel-a', 'channel-b'], groups: [],
  pricing: null, supportsStreaming: true, supportsVision: false, supportsTools: true,
  supportsJson: true, contextWindow: 128000, recommendedUse: null,
  offers: [
    {
      channel: 'channel-a', platform: 'openai',
      groups: [{ id: 1, name: 'Pro', platform: 'openai', subscriptionType: 'standard', rateMultiplier: 1, effectiveRateMultiplier: 1, peakRateEnabled: true, peakStart: '09:00', peakEnd: '12:00', peakRateMultiplier: 2, isExclusive: false }],
      pricing: { billing_mode: 'token', input_price: 0.000005, output_price: 0.00003, cache_write_price: null, cache_read_price: 0.0000005, image_input_price: 0.000007, image_output_price: 0.000009, per_request_price: null, intervals: [] },
    },
    {
      channel: 'channel-b', platform: 'openai',
      groups: [{ id: 2, name: 'Basic', platform: 'openai', subscriptionType: 'standard', rateMultiplier: 0.8, effectiveRateMultiplier: 0.8, peakRateEnabled: false, peakStart: '', peakEnd: '', peakRateMultiplier: 0, isExclusive: false }],
      pricing: { billing_mode: 'per_request', input_price: null, output_price: null, cache_write_price: null, cache_read_price: null, image_input_price: null, image_output_price: null, per_request_price: 0.02, intervals: [] },
    },
  ],
}

describe('ModelPlazaDetailDialog', () => {
  beforeEach(() => { i18n.global.locale.value = 'en' })

  it('renders separate offers and only verified performance data', async () => {
    const wrapper = mount(ModelPlazaDetailDialog, {
      props: { show: true, model, baseUrl: 'https://example.com/v1' },
      global: {
        plugins: [createPinia()],
        stubs: {
          BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' },
          ModelIcon: { template: '<span />' },
          Icon: { template: '<span />' },
        },
      },
    })

    expect(wrapper.findAll('[data-testid="catalog-offer"]')).toHaveLength(2)
    expect(wrapper.text()).toContain('channel-a')
    expect(wrapper.text()).toContain('channel-b')
    expect(wrapper.text()).toContain('Peak 09:00-12:00')
    expect(wrapper.text()).toContain('Image input')
    expect(wrapper.text()).toContain('Image output')

    await wrapper.findAll('[role="tab"]').find((node) => node.text() === 'Performance')!.trigger('click')
    const performance = wrapper.get('[data-testid="performance-tab"]').text()
    expect(performance).toContain('No model-safe latency measurement')
    expect(performance).not.toMatch(/TPS|rate limit|parameter/i)

    await wrapper.findAll('[role="tab"]').find((node) => node.text() === 'API')!.trigger('click')
    expect(wrapper.get('[data-testid="api-tab"]').text()).toContain('https://example.com/v1/chat/completions')
    expect(wrapper.get('[data-testid="api-tab"]').text()).toContain('gpt-5')
  })
})
