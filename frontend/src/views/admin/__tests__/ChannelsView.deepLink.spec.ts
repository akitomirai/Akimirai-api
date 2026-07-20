import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import ChannelsView from '../ChannelsView.vue'
import type { Channel } from '@/api/admin/channels'

const {
  getById,
  listChannels,
  getAllGroups,
  getWebSearchEmulationConfig,
  showError,
} = vi.hoisted(() => ({
  getById: vi.fn(),
  listChannels: vi.fn(),
  getAllGroups: vi.fn(),
  getWebSearchEmulationConfig: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (_key: string, paramsOrFallback?: unknown, fallback?: string) => {
        if (fallback) return fallback
        return typeof paramsOrFallback === 'string' ? paramsOrFallback : _key
      },
    }),
  }
})

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: { edit: '42' } }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess: vi.fn() }),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    channels: {
      list: listChannels,
      getById,
      create: vi.fn(),
      update: vi.fn(),
      remove: vi.fn(),
      syncPricingModels: vi.fn(),
    },
    groups: { getAll: getAllGroups },
    settings: { getWebSearchEmulationConfig },
    accounts: { getById: vi.fn(), list: vi.fn() },
  },
}))

const channel: Channel = {
  id: 42,
  name: 'Deep linked channel',
  description: 'Opened from model pricing',
  status: 'active',
  billing_model_source: 'channel_mapped',
  restrict_models: true,
  group_ids: [],
  model_pricing: [],
  model_mapping: {},
  apply_pricing_to_account_stats: false,
  account_stats_pricing_rules: [],
  created_at: '2026-07-13T00:00:00Z',
  updated_at: '2026-07-13T00:00:00Z',
}

describe('ChannelsView edit deep link', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listChannels.mockResolvedValue({ items: [], total: 0 })
    getById.mockResolvedValue(channel)
    getAllGroups.mockResolvedValue([])
    getWebSearchEmulationConfig.mockResolvedValue({ enabled: false, providers: [] })
  })

  it('loads the requested channel and opens the existing edit dialog', async () => {
    const wrapper = shallowMount(ChannelsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          BaseDialog: {
            props: ['show', 'title'],
            template: '<div class="base-dialog" :data-show="String(show)" :data-title="title"><slot /></div>',
          },
          DataTable: true,
          Pagination: true,
          ConfirmDialog: true,
          EmptyState: true,
          Select: true,
          Icon: true,
          PlatformIcon: true,
          Toggle: true,
          PricingEntryCard: true,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    expect(getById).toHaveBeenCalledWith(42)
    const dialog = wrapper.find('.base-dialog')
    expect(dialog.attributes('data-show')).toBe('true')
    expect(dialog.attributes('data-title')).toContain('Edit Channel')
    expect(showError).not.toHaveBeenCalled()
  })
})
