import type {
  UserAvailableChannel,
  UserAvailableGroup,
  UserSupportedModelPricing,
} from '@/api/channels'

export type ModelAvailabilityStatus =
  | 'available'
  | 'maintenance'
  | 'temporarily_unavailable'
  | 'unknown'

export interface ModelCatalogGroup {
  id: number
  name: string
  platform: string
  subscriptionType: string
  rateMultiplier: number
  effectiveRateMultiplier: number
  isExclusive: boolean
}

export interface ModelCatalogItem {
  id: string
  displayName: string
  platform: string
  status: ModelAvailabilityStatus
  channelNames: string[]
  groups: ModelCatalogGroup[]
  pricing: UserSupportedModelPricing | null
  supportsStreaming: 'unknown'
}

export interface ModelStatusSource {
  modelEnabled?: boolean | null
  channelEnabled?: boolean | null
  hasAvailableChannel?: boolean | null
  hasModelConfig?: boolean | null
  hasSufficientData?: boolean | null
}

export const MODEL_MULTIPLIER_EXPLANATION =
  '倍率表示该模型按平台基础计费单位的倍数消耗，最终扣费以用量账单为准。'

export function deriveModelAvailabilityStatus(source: ModelStatusSource): ModelAvailabilityStatus {
  if (source.hasSufficientData === false) return 'unknown'
  if (source.modelEnabled === false || source.channelEnabled === false) return 'maintenance'
  if (source.hasModelConfig === true && source.hasAvailableChannel === false) {
    return 'temporarily_unavailable'
  }
  if (source.hasAvailableChannel === true) return 'available'
  return 'unknown'
}

export function deriveModelCatalog(
  channels: UserAvailableChannel[],
  userGroupRates: Record<number, number> = {},
): ModelCatalogItem[] {
  const items = new Map<string, ModelCatalogItem>()

  for (const channel of channels) {
    const channelName = `${channel.name ?? ''}`.trim()
    for (const section of channel.platforms || []) {
      const platform = `${section.platform ?? ''}`.trim()
      if (!platform) continue

      for (const model of section.supported_models || []) {
        const modelId = `${model.name ?? ''}`.trim()
        if (!modelId) continue

        const key = `${platform.toLowerCase()}\u0000${modelId.toLowerCase()}`
        const existing = items.get(key)
        const groups = (section.groups || []).map((group) =>
          toCatalogGroup(group, userGroupRates),
        )

        if (existing) {
          if (channelName && !existing.channelNames.includes(channelName)) {
            existing.channelNames.push(channelName)
          }
          mergeGroups(existing.groups, groups)
          if (!existing.pricing && model.pricing) existing.pricing = model.pricing
          continue
        }

        items.set(key, {
          id: modelId,
          displayName: modelId,
          platform,
          status: deriveModelAvailabilityStatus({
            hasAvailableChannel: true,
            hasModelConfig: true,
            hasSufficientData: true,
          }),
          channelNames: channelName ? [channelName] : [],
          groups,
          pricing: model.pricing,
          supportsStreaming: 'unknown',
        })
      }
    }
  }

  return Array.from(items.values()).sort((a, b) => {
    const byPlatform = a.platform.localeCompare(b.platform)
    if (byPlatform !== 0) return byPlatform
    return a.id.localeCompare(b.id)
  })
}

export function findCatalogModel(
  models: ModelCatalogItem[],
  rawModel?: string | (string | null)[] | null,
): ModelCatalogItem | null {
  const value = Array.isArray(rawModel) ? rawModel[0] : rawModel
  const wanted = `${value ?? ''}`.trim().toLowerCase()
  if (!wanted) return null
  return models.find((item) => item.id.toLowerCase() === wanted) || null
}

export function pickRecommendedCatalogModel(models: ModelCatalogItem[]): ModelCatalogItem | null {
  return models.find((item) => item.status === 'available') || models[0] || null
}

export function selectQuickStartCatalogModel(
  models: ModelCatalogItem[],
  rawModel?: string | (string | null)[] | null,
): {
  selected: ModelCatalogItem | null
  requested: string
  usedFallback: boolean
} {
  const requested = `${Array.isArray(rawModel) ? rawModel[0] : rawModel ?? ''}`.trim()
  const match = findCatalogModel(models, requested)
  if (match && match.status === 'available') {
    return { selected: match, requested, usedFallback: false }
  }
  return {
    selected: pickRecommendedCatalogModel(models),
    requested,
    usedFallback: requested.length > 0,
  }
}

export function getMultiplierRange(item: ModelCatalogItem): {
  min: number | null
  max: number | null
} {
  const values = item.groups
    .map((group) => group.effectiveRateMultiplier)
    .filter((value) => Number.isFinite(value) && value >= 0)

  if (values.length === 0) return { min: null, max: null }
  return {
    min: Math.min(...values),
    max: Math.max(...values),
  }
}

export function formatMultiplierRange(item: ModelCatalogItem): string {
  const range = getMultiplierRange(item)
  if (range.min == null || range.max == null) return '-'
  if (range.min === range.max) return `${formatMultiplier(range.min)}x`
  return `${formatMultiplier(range.min)}x - ${formatMultiplier(range.max)}x`
}

export function isModelAvailabilityErrorCode(code?: string | null): boolean {
  const normalized = `${code ?? ''}`.trim().toUpperCase()
  return normalized === 'MODEL_DISABLED' || normalized === 'NO_AVAILABLE_CHANNEL'
}

function toCatalogGroup(
  group: UserAvailableGroup,
  userGroupRates: Record<number, number>,
): ModelCatalogGroup {
  const override = userGroupRates[group.id]
  const effectiveRateMultiplier =
    typeof override === 'number' && Number.isFinite(override)
      ? override
      : group.rate_multiplier

  return {
    id: group.id,
    name: group.name,
    platform: group.platform,
    subscriptionType: group.subscription_type,
    rateMultiplier: group.rate_multiplier,
    effectiveRateMultiplier,
    isExclusive: group.is_exclusive,
  }
}

function mergeGroups(target: ModelCatalogGroup[], incoming: ModelCatalogGroup[]) {
  const seen = new Set(target.map((group) => group.id))
  for (const group of incoming) {
    if (seen.has(group.id)) continue
    target.push(group)
    seen.add(group.id)
  }
}

function formatMultiplier(value: number): string {
  return Number.isInteger(value) ? String(value) : value.toFixed(2).replace(/0+$/, '').replace(/\.$/, '')
}
