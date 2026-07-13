import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { i18n } from '@/i18n'
import ModelPlazaView from '../ModelPlazaView.vue'
import userChannelsAPI from '@/api/channels'
import { useAppStore } from '@/stores/app'
import type { UserModelCatalogItem } from '@/api/channels'

const routerMocks = vi.hoisted(() => ({
  route: { query: {} as Record<string, string> },
  replace: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('vue-router', () => ({
  useRoute: () => routerMocks.route,
  useRouter: () => ({ replace: routerMocks.replace }),
}))

vi.mock('@/stores/app', () => ({ useAppStore: vi.fn() }))
vi.mock('@/api/channels', () => ({
  default: { getAvailable: vi.fn(), getModelCatalog: vi.fn() },
}))
vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn().mockResolvedValue(true) }),
}))

function catalogItem(modelId: string, provider: string, group: string): UserModelCatalogItem {
  return {
    id: `${provider}:${modelId}`,
    display_name: modelId,
    model_id: modelId,
    provider,
    family: provider,
    status: 'available',
    status_reason: 'Ready',
    billing_multiplier: 1,
    billing_description: '1x',
    supports_streaming: true,
    supports_vision: false,
    supports_tools: true,
    supports_json: true,
    context_window: 128000,
    recommended_use: null,
    available_channel_count: 1,
    quick_start_url: `/quick-start?model=${modelId}`,
    updated_at: '2026-07-13T00:00:00Z',
    channels: [`${provider}-channel`],
    groups: [],
    pricing: null,
    offers: [{
      channel: `${provider}-channel`,
      platform: provider,
      groups: [{
        id: modelId.length,
        name: group,
        platform: provider,
        subscription_type: 'standard',
        rate_multiplier: 1,
        peak_rate_enabled: false,
        peak_start: '',
        peak_end: '',
        peak_rate_multiplier: 0,
        is_exclusive: false,
      }],
      pricing: {
        billing_mode: 'token', input_price: 0.000001, output_price: 0.000002,
        cache_write_price: null, cache_read_price: null, image_output_price: null,
        per_request_price: null, intervals: [],
      },
    }],
  }
}

const payload = [
  catalogItem('gpt-5', 'openai', 'Pro'),
  catalogItem('gemini-2.5', 'google', 'Basic'),
]

function mountView() {
  return mount(ModelPlazaView, {
    global: {
      stubs: {
        AppLayout: { template: '<main><slot /></main>' },
        Icon: { template: '<span />' },
        ModelPlazaCard: {
          props: ['item'],
          emits: ['details', 'copy'],
          template: '<button class="model-card" @click="$emit(\'details\', item)">{{ item.modelId }}</button>',
        },
        ModelPlazaDetailDialog: { props: ['show'], template: '<div v-if="show" data-testid="detail-dialog" />' },
      },
    },
  })
}

describe('ModelPlazaView', () => {
  beforeEach(() => {
    i18n.global.locale.value = 'en'
    routerMocks.route.query = {}
    routerMocks.replace.mockClear()
    vi.mocked(userChannelsAPI.getModelCatalog).mockReset()
    vi.mocked(useAppStore).mockReturnValue({
      cachedPublicSettings: { api_base_url: 'https://example.com' },
      fetchPublicSettings: vi.fn().mockResolvedValue({ available_channels_enabled: true }),
    } as unknown as ReturnType<typeof useAppStore>)
    vi.mocked(userChannelsAPI.getModelCatalog).mockResolvedValue(payload)
  })

  it('loads cards, filters locally, and keeps detail selection URL-compatible', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.findAll('.model-card')).toHaveLength(2)
    await wrapper.get('input[type="search"]').setValue('gemini')
    expect(wrapper.findAll('.model-card')).toHaveLength(1)
    expect(wrapper.text()).toContain('gemini-2.5')

    await wrapper.get('.model-card').trigger('click')
    expect(routerMocks.replace).toHaveBeenCalledWith({ query: { model: 'gemini-2.5' } })
    expect(wrapper.get('[data-testid="detail-dialog"]').exists()).toBe(true)
  })

  it('shows a retryable error without stale catalog cards', async () => {
    vi.mocked(userChannelsAPI.getModelCatalog).mockRejectedValueOnce(new Error('catalog offline'))
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('The model catalog could not be loaded.')
    expect(wrapper.text()).toContain('catalog offline')
    expect(wrapper.findAll('.model-card')).toHaveLength(0)
  })

  it('does not request catalog data when the compatibility feature switch is disabled', async () => {
    vi.mocked(useAppStore).mockReturnValue({
      cachedPublicSettings: { api_base_url: 'https://example.com' },
      fetchPublicSettings: vi.fn().mockResolvedValue({ available_channels_enabled: false }),
    } as unknown as ReturnType<typeof useAppStore>)
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('Model Plaza is not enabled')
    expect(userChannelsAPI.getModelCatalog).not.toHaveBeenCalled()
  })
})
