import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ModelPricingView from '../ModelPricingView.vue'
import type { Channel } from '@/api/admin/channels'

const { routerPush, showError, listChannels, getAllGroups } = vi.hoisted(() => ({
  routerPush: vi.fn(),
  showError: vi.fn(),
  listChannels: vi.fn(),
  getAllGroups: vi.fn(),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (_key: string, paramsOrFallback?: unknown, fallback?: string) => {
        if (fallback) return fallback
        if (typeof paramsOrFallback === 'string') return paramsOrFallback
        if (paramsOrFallback && typeof paramsOrFallback === 'object' && 'count' in paramsOrFallback) {
          return `${(paramsOrFallback as { count: number }).count} model(s)`
        }
        return _key
      },
    }),
  }
})

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: routerPush }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError }),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    channels: { list: listChannels },
    groups: { getAllIncludingInactive: getAllGroups },
  },
}))

const sourceChannel: Channel = {
  id: 42,
  name: 'OpenAI Primary',
  description: 'Primary pricing source',
  status: 'active',
  billing_model_source: 'channel_mapped',
  restrict_models: true,
  group_ids: [7],
  model_pricing: [{
    id: 5,
    platform: 'openai',
    models: ['gpt-5'],
    billing_mode: 'token',
    input_price: 0.000005,
    output_price: 0.00003,
    cache_write_price: null,
    cache_read_price: 0.0000005,
    image_output_price: null,
    per_request_price: null,
    intervals: [],
  }],
  model_mapping: {},
  apply_pricing_to_account_stats: false,
  account_stats_pricing_rules: [],
  created_at: '2026-07-13T00:00:00Z',
  updated_at: '2026-07-13T00:00:00Z',
}

function mountView() {
  return mount(ModelPricingView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: { template: '<span />' },
        Select: {
          props: ['modelValue', 'options'],
          emits: ['update:modelValue'],
          template: '<div class="select-stub" />',
        },
        ModelPricingCard: {
          props: ['card'],
          emits: ['details'],
          template: '<button class="model-card" @click="$emit(\'details\', card)">{{ card.model }}</button>',
        },
        ModelPricingDetailDialog: {
          props: ['show', 'model', 'groups'],
          emits: ['close', 'configure', 'edit-source'],
          template: '<div v-if="show" class="detail-stub"><button class="edit-source" @click="$emit(\'edit-source\', model.sources[0].channelId)">Edit source</button></div>',
        },
      },
    },
  })
}

describe('ModelPricingView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listChannels.mockResolvedValue({ items: [sourceChannel], total: 1 })
    getAllGroups.mockResolvedValue([{ id: 7, name: 'Pro pool' }])
  })

  it('loads the complete projection and filters it by search text', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(listChannels).toHaveBeenCalledWith(
      1,
      200,
      { sort_by: 'id', sort_order: 'asc' },
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    )
    expect(wrapper.find('.model-card').text()).toBe('gpt-5')

    await wrapper.find('input[type="search"]').setValue('missing')
    expect(wrapper.find('.model-card').exists()).toBe(false)

    await wrapper.find('input[type="search"]').setValue('OpenAI Primary')
    expect(wrapper.find('.model-card').text()).toBe('gpt-5')
  })

  it('routes source edits and configuration to the channel owner', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('.model-card').trigger('click')
    await wrapper.find('.edit-source').trigger('click')

    expect(routerPush).toHaveBeenCalledWith({
      path: '/admin/channels/config',
      query: { edit: '42' },
    })

    const configButton = wrapper.findAll('button').find((button) => button.text().includes('Channel configuration'))
    expect(configButton).toBeDefined()
    await configButton!.trigger('click')
    expect(routerPush).toHaveBeenCalledWith('/admin/channels/config')
  })
})
