import type {
  Channel,
  ChannelModelPricing,
} from '@/api/admin/channels'
import type { BillingMode, ChannelStatus } from '@/constants/channel'

export interface AdminModelPricingSource {
  key: string
  channelId: number
  channelName: string
  channelDescription: string
  channelStatus: ChannelStatus
  billingModelSource: Channel['billing_model_source']
  groupIds: number[]
  restrictModels: boolean
  ruleIndex: number
  sharedModels: string[]
  pricing: ChannelModelPricing
}

export interface AdminModelPricingCard {
  key: string
  model: string
  platform: string
  isPattern: boolean
  activeSourceCount: number
  disabledSourceCount: number
  billingModes: BillingMode[]
  sources: AdminModelPricingSource[]
}

export interface AdminModelPricingFilters {
  search?: string
  provider?: string
  billingMode?: BillingMode | ''
  status?: ChannelStatus | ''
}

export interface AdminChannelPage {
  items: Channel[]
  total: number
}

export type AdminChannelPageFetcher = (
  page: number,
  pageSize: number,
  filters?: {
    status?: string
    search?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: { signal?: AbortSignal },
) => Promise<AdminChannelPage>

const DEFAULT_PAGE_SIZE = 200
const MAX_PAGES = 500

function clonePricingRule(rule: ChannelModelPricing): ChannelModelPricing {
  return {
    ...rule,
    models: [...(rule.models || [])],
    intervals: (rule.intervals || []).map((interval) => ({ ...interval })),
  }
}

export function isModelPattern(model: string): boolean {
  return /[*?]/.test(model)
}

/**
 * Loads the complete admin channel collection without relying on a single
 * oversized page. IDs are de-duplicated so page shifts cannot duplicate a
 * source in the model projection.
 */
export async function fetchAllAdminChannels(
  fetchPage: AdminChannelPageFetcher,
  options: { signal?: AbortSignal; pageSize?: number } = {},
): Promise<Channel[]> {
  const pageSize = options.pageSize ?? DEFAULT_PAGE_SIZE
  const channelsById = new Map<number, Channel>()

  for (let page = 1; page <= MAX_PAGES; page += 1) {
    const response = await fetchPage(
      page,
      pageSize,
      { sort_by: 'id', sort_order: 'asc' },
      { signal: options.signal },
    )
    const items = response.items || []

    for (const channel of items) {
      channelsById.set(channel.id, channel)
    }

    if (items.length === 0 || channelsById.size >= response.total || items.length < pageSize) {
      return [...channelsById.values()]
    }
  }

  throw new Error('Channel pagination exceeded the safety limit')
}

/**
 * Projects channel-owned pricing rules into model-first cards. Each card keeps
 * the complete source rule so shared model lists and wildcard patterns remain
 * visible instead of being flattened into an independently editable record.
 */
export function projectAdminModelPricing(channels: Channel[]): AdminModelPricingCard[] {
  const cards = new Map<string, AdminModelPricingCard>()

  for (const channel of channels) {
    for (const [ruleIndex, rawRule] of (channel.model_pricing || []).entries()) {
      const sharedModels = (rawRule.models || []).map((model) => model.trim()).filter(Boolean)
      if (sharedModels.length === 0) continue

      for (const model of sharedModels) {
        const platform = rawRule.platform || 'unknown'
        const key = `${platform.toLowerCase()}::${model.toLowerCase()}`
        let card = cards.get(key)
        if (!card) {
          card = {
            key,
            model,
            platform,
            isPattern: isModelPattern(model),
            activeSourceCount: 0,
            disabledSourceCount: 0,
            billingModes: [],
            sources: [],
          }
          cards.set(key, card)
        }

        const pricing = clonePricingRule(rawRule)
        card.sources.push({
          key: `${channel.id}:${rawRule.id ?? ruleIndex}:${model.toLowerCase()}`,
          channelId: channel.id,
          channelName: channel.name,
          channelDescription: channel.description || '',
          channelStatus: channel.status,
          billingModelSource: channel.billing_model_source,
          groupIds: [...(channel.group_ids || [])],
          restrictModels: channel.restrict_models,
          ruleIndex,
          sharedModels: [...sharedModels],
          pricing,
        })

        if (channel.status === 'active') card.activeSourceCount += 1
        else card.disabledSourceCount += 1
        if (!card.billingModes.includes(rawRule.billing_mode)) {
          card.billingModes.push(rawRule.billing_mode)
        }
      }
    }
  }

  return [...cards.values()].sort((a, b) => {
    const platformOrder = a.platform.localeCompare(b.platform)
    return platformOrder || a.model.localeCompare(b.model)
  })
}

export function filterAdminModelPricing(
  cards: AdminModelPricingCard[],
  filters: AdminModelPricingFilters,
): AdminModelPricingCard[] {
  const query = filters.search?.trim().toLowerCase() || ''

  return cards.filter((card) => {
    if (filters.provider && card.platform !== filters.provider) return false
    if (filters.billingMode && !card.billingModes.includes(filters.billingMode)) return false
    if (filters.status === 'active' && card.activeSourceCount === 0) return false
    if (filters.status === 'disabled' && card.disabledSourceCount === 0) return false
    if (!query) return true

    return [
      card.model,
      card.platform,
      ...card.sources.flatMap((source) => [
        source.channelName,
        source.channelDescription,
        ...source.sharedModels,
      ]),
    ].some((value) => value.toLowerCase().includes(query))
  })
}
