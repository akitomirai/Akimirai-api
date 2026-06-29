import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ModelCatalogPanel from '../ModelCatalogPanel.vue'
import { MODEL_MULTIPLIER_EXPLANATION, type ModelCatalogItem } from '@/utils/modelCatalog'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (_key: string, fallback?: string) => fallback || _key
  })
}))

const copyToClipboard = vi.fn().mockResolvedValue(true)

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard })
}))

const item: ModelCatalogItem = {
  id: 'gpt-test-model',
  displayName: 'gpt-test-model',
  platform: 'openai',
  status: 'available',
  channelNames: ['safe-channel'],
  groups: [
    {
      id: 1,
      name: 'Pro',
      platform: 'openai',
      subscriptionType: 'standard',
      rateMultiplier: 1.2,
      effectiveRateMultiplier: 1.2,
      isExclusive: false,
    },
  ],
  pricing: null,
  supportsStreaming: 'unknown',
}

const mountComponent = (props = {}) => mount(ModelCatalogPanel, {
  props: {
    items: [item],
    loading: false,
    error: false,
    ...props,
  },
  global: {
    stubs: {
      RouterLink: {
        props: ['to'],
        template: '<a :data-to="JSON.stringify(to)"><slot /></a>',
      },
      Icon: { template: '<span />' },
      PlatformIcon: { template: '<span />' },
      Teleport: true,
    },
  },
})

describe('ModelCatalogPanel', () => {
  it('renders model id, availability, multiplier explanation, and Quick Start entry', () => {
    const wrapper = mountComponent()
    const text = wrapper.text()

    expect(text).toContain('Available Models')
    expect(text).toContain('gpt-test-model')
    expect(text).toContain('Available')
    expect(text).toContain('1.2x')
    expect(text).toContain(MODEL_MULTIPLIER_EXPLANATION)
    expect(wrapper.html()).toContain('/quick-start')
  })

  it('copies the model id without exposing sensitive fields', async () => {
    const wrapper = mountComponent()

    await wrapper.findAll('button.btn-secondary')[1].trigger('click')

    expect(copyToClipboard).toHaveBeenCalledWith('gpt-test-model')
    expect(wrapper.text()).not.toContain('sk-test-placeholder')
    expect(wrapper.text()).not.toContain('private_key')
    expect(wrapper.text()).not.toContain('cookie')
  })

  it('renders loading, error, and empty states without fake models', () => {
    expect(mountComponent({ loading: true }).text()).toContain('Loading')
    expect(mountComponent({ error: true, items: [] }).text()).toContain('Failed to load available models')
    expect(mountComponent({ items: [] }).text()).toContain('No available models yet')
  })
})
