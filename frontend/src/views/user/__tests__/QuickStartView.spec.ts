import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import QuickStartView from '../QuickStartView.vue'
import userChannelsAPI from '@/api/channels'
import { useAppStore } from '@/stores'
import type { UserModelCatalogItem } from '@/api/channels'

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      locale: 'en',
      t: (_key: string, fallback?: string) => fallback || _key
    }
  }),
  useI18n: () => ({
    t: (_key: string, fallback?: string) => fallback || _key
  })
}))

const routeState = {
  query: {} as Record<string, string>
}

vi.mock('vue-router', () => ({
  useRoute: () => routeState
}))

vi.mock('@/stores', () => ({
  useAppStore: vi.fn()
}))

vi.mock('@/api/channels', () => ({
  default: {
    getAvailable: vi.fn(),
    getModelCatalog: vi.fn(),
  }
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

const catalog: UserModelCatalogItem[] = [
  {
    id: 'openai:gpt-available',
    display_name: 'gpt-available',
    model_id: 'gpt-available',
    provider: 'openai',
    family: 'GPT',
    status: 'available',
    status_reason: '当前有可用渠道',
    billing_multiplier: 1,
    billing_description: '1x',
    supports_streaming: true,
    supports_vision: false,
    supports_tools: true,
    supports_json: true,
    context_window: 128000,
    recommended_use: null,
    available_channel_count: 1,
    quick_start_url: '/quick-start?model=gpt-available',
    updated_at: null,
    channels: ['safe-channel'],
    groups: [
      {
        id: 1,
        name: 'Pro',
        platform: 'openai',
        subscription_type: 'standard',
        rate_multiplier: 1,
        is_exclusive: false,
      },
    ],
    pricing: null,
  },
]

const mountComponent = async (query: Record<string, string> = {}) => {
  routeState.query = query
  vi.mocked(useAppStore).mockReturnValue({
    cachedPublicSettings: { api_base_url: 'https://example.com', payment_enabled: true },
    fetchPublicSettings: vi.fn().mockResolvedValue(undefined),
  } as unknown as ReturnType<typeof useAppStore>)
  vi.mocked(userChannelsAPI.getModelCatalog).mockResolvedValue(catalog)

  const wrapper = mount(QuickStartView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        LoadingSpinner: { template: '<span>loading</span>' },
        RouterLink: {
          props: ['to'],
          template: '<a><slot /></a>'
        },
        Icon: { template: '<span />' },
      },
    },
  })
  await flushPromises()
  return wrapper
}

describe('QuickStartView', () => {
  it('reads a query model and uses it in examples with one /v1', async () => {
    const wrapper = await mountComponent({ model: 'gpt-available' })
    const code = wrapper.findAll('pre code').map((node) => node.text()).join('\n\n')

    expect(code).toContain('gpt-available')
    expect(code).toContain('https://example.com/v1/chat/completions')
    expect(code).not.toContain('/v1/v1')
    expect(code).toContain('<YOUR_API_KEY>')
  })

  it('falls back safely when query model is not available', async () => {
    const wrapper = await mountComponent({ model: 'missing-model' })
    const text = wrapper.text()
    const code = wrapper.findAll('pre code').map((node) => node.text()).join('\n\n')

    expect(text).toContain('Requested model missing-model is not currently available')
    expect(code).toContain('gpt-available')
    expect(code).not.toContain('missing-model')
  })
})
