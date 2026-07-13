import { describe, expect, it, vi } from 'vitest'
import type { Channel, ChannelModelPricing } from '@/api/admin/channels'
import {
  fetchAllAdminChannels,
  filterAdminModelPricing,
  projectAdminModelPricing,
} from '@/utils/adminModelPricing'

function pricing(overrides: Partial<ChannelModelPricing> = {}): ChannelModelPricing {
  return {
    id: 10,
    platform: 'openai',
    models: ['gpt-5', 'gpt-5-mini'],
    billing_mode: 'token',
    input_price: 0.000005,
    output_price: 0.00003,
    cache_write_price: 0.00000625,
    cache_read_price: 0.0000005,
    image_output_price: null,
    per_request_price: null,
    intervals: [],
    ...overrides,
  }
}

function channel(id: number, overrides: Partial<Channel> = {}): Channel {
  return {
    id,
    name: `channel-${id}`,
    description: '',
    status: 'active',
    billing_model_source: 'channel_mapped',
    restrict_models: true,
    group_ids: [id],
    model_pricing: [pricing()],
    model_mapping: {},
    apply_pricing_to_account_stats: false,
    account_stats_pricing_rules: [],
    created_at: '2026-07-13T00:00:00Z',
    updated_at: '2026-07-13T00:00:00Z',
    ...overrides,
  }
}

describe('admin model pricing projection', () => {
  it('keeps every source rule and shared-model relationship', () => {
    const cards = projectAdminModelPricing([
      channel(1),
      channel(2, {
        status: 'disabled',
        group_ids: [7, 9],
        model_pricing: [pricing({ id: 20, models: ['gpt-5'], input_price: 0.000004 })],
      }),
    ])

    const gpt5 = cards.find((card) => card.model === 'gpt-5')
    expect(gpt5).toMatchObject({
      platform: 'openai',
      activeSourceCount: 1,
      disabledSourceCount: 1,
      billingModes: ['token'],
    })
    expect(gpt5?.sources).toHaveLength(2)
    expect(gpt5?.sources[0]).toMatchObject({
      channelId: 1,
      groupIds: [1],
      sharedModels: ['gpt-5', 'gpt-5-mini'],
    })
    expect(gpt5?.sources[0].pricing.models).toEqual(['gpt-5', 'gpt-5-mini'])
    expect(gpt5?.sources[1]).toMatchObject({ channelId: 2, groupIds: [7, 9] })
  })

  it('preserves wildcard patterns as cards without expanding them', () => {
    const cards = projectAdminModelPricing([
      channel(1, { model_pricing: [pricing({ models: ['gpt-5-*'] })] }),
    ])

    expect(cards).toHaveLength(1)
    expect(cards[0]).toMatchObject({ model: 'gpt-5-*', isPattern: true })
    expect(cards[0].sources[0].sharedModels).toEqual(['gpt-5-*'])
  })

  it('filters against model, provider, billing mode, source status, and source name', () => {
    const cards = projectAdminModelPricing([
      channel(1),
      channel(2, {
        name: 'Image source',
        status: 'disabled',
        model_pricing: [pricing({
          platform: 'gemini',
          models: ['imagen-4'],
          billing_mode: 'image',
          input_price: null,
          output_price: null,
          per_request_price: 0.08,
        })],
      }),
    ])

    expect(filterAdminModelPricing(cards, { search: 'image source' }).map((card) => card.model)).toEqual(['imagen-4'])
    expect(filterAdminModelPricing(cards, { provider: 'gemini' }).map((card) => card.model)).toEqual(['imagen-4'])
    expect(filterAdminModelPricing(cards, { billingMode: 'image' }).map((card) => card.model)).toEqual(['imagen-4'])
    expect(filterAdminModelPricing(cards, { status: 'disabled' }).map((card) => card.model)).toEqual(['imagen-4'])
  })
})

describe('fetchAllAdminChannels', () => {
  it('walks every page and de-duplicates shifted rows', async () => {
    const fetchPage = vi.fn(async (page: number) => {
      const pages = [
        { items: [channel(1), channel(2)], total: 4 },
        { items: [channel(2), channel(3)], total: 4 },
        { items: [channel(4)], total: 4 },
      ]
      return pages[page - 1] ?? { items: [], total: 4 }
    })

    const result = await fetchAllAdminChannels(fetchPage, { pageSize: 2 })

    expect(result.map((item) => item.id)).toEqual([1, 2, 3, 4])
    expect(fetchPage).toHaveBeenCalledTimes(3)
    expect(fetchPage).toHaveBeenNthCalledWith(
      1,
      1,
      2,
      { sort_by: 'id', sort_order: 'asc' },
      { signal: undefined },
    )
  })
})
