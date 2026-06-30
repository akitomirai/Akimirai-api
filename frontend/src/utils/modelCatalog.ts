import type {
  UserAvailableGroup,
  UserModelCatalogItem,
  UserSupportedModelPricing,
} from '@/api/channels'

export type ModelAvailabilityStatus = 'available' | 'maintenance' | 'unavailable' | 'unknown'

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
  modelId: string
  provider: string
  platform: string
  family: string | null
  status: ModelAvailabilityStatus
  statusReason: string
  billingMultiplier: number | null
  billingDescription: string
  availableChannelCount: number
  quickStartUrl: string
  updatedAt: string | null
  channelNames: string[]
  groups: ModelCatalogGroup[]
  pricing: UserSupportedModelPricing | null
  supportsStreaming: boolean | null
  supportsVision: boolean | null
  supportsTools: boolean | null
  supportsJson: boolean | null
  contextWindow: number | null
  recommendedUse: string | null
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
    return 'unavailable'
  }
  if (source.hasAvailableChannel === true) return 'available'
  return 'unknown'
}

export function toModelCatalogItems(items: UserModelCatalogItem[] = []): ModelCatalogItem[] {
  return items
    .map((item) => ({
      id: cleanString(item.model_id || item.id),
      displayName: cleanString(item.display_name || item.model_id || item.id),
      modelId: cleanString(item.model_id),
      provider: cleanString(item.provider),
      platform: cleanString(item.provider),
      family: cleanNullableString(item.family),
      status: normalizeStatus(item.status),
      statusReason: cleanString(item.status_reason),
      billingMultiplier: finiteNumberOrNull(item.billing_multiplier),
      billingDescription: cleanString(item.billing_description),
      availableChannelCount: Math.max(0, Number(item.available_channel_count) || 0),
      quickStartUrl: cleanString(item.quick_start_url),
      updatedAt: cleanNullableString(item.updated_at),
      channelNames: uniqueStrings(item.channels),
      groups: (item.groups || []).map(toCatalogGroup),
      pricing: item.pricing || null,
      supportsStreaming: nullableBoolean(item.supports_streaming),
      supportsVision: nullableBoolean(item.supports_vision),
      supportsTools: nullableBoolean(item.supports_tools),
      supportsJson: nullableBoolean(item.supports_json),
      contextWindow: positiveIntOrNull(item.context_window),
      recommendedUse: cleanNullableString(item.recommended_use),
    }))
    .filter((item) => item.id && item.modelId && item.provider)
    .sort((a, b) => {
      const byProvider = a.provider.localeCompare(b.provider)
      if (byProvider !== 0) return byProvider
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
  return models.find((item) => item.id.toLowerCase() === wanted || item.modelId.toLowerCase() === wanted) || null
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

  if (values.length === 0) {
    return item.billingMultiplier == null
      ? { min: null, max: null }
      : { min: item.billingMultiplier, max: item.billingMultiplier }
  }
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

function toCatalogGroup(group: UserAvailableGroup): ModelCatalogGroup {
  return {
    id: group.id,
    name: cleanString(group.name),
    platform: cleanString(group.platform),
    subscriptionType: cleanString(group.subscription_type),
    rateMultiplier: finiteNumberOrFallback(group.rate_multiplier, 0),
    effectiveRateMultiplier: finiteNumberOrFallback(group.rate_multiplier, 0),
    isExclusive: group.is_exclusive === true,
  }
}

function normalizeStatus(status: string): ModelAvailabilityStatus {
  return status === 'available' || status === 'maintenance' || status === 'unavailable' || status === 'unknown'
    ? status
    : 'unknown'
}

function cleanString(value: unknown): string {
  return `${value ?? ''}`.trim()
}

function cleanNullableString(value: unknown): string | null {
  const text = cleanString(value)
  return text || null
}

function uniqueStrings(values: unknown): string[] {
  if (!Array.isArray(values)) return []
  return Array.from(new Set(values.map(cleanString).filter(Boolean)))
}

function finiteNumberOrNull(value: unknown): number | null {
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}

function finiteNumberOrFallback(value: unknown, fallback: number): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback
}

function positiveIntOrNull(value: unknown): number | null {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? Math.trunc(value) : null
}

function nullableBoolean(value: unknown): boolean | null {
  return typeof value === 'boolean' ? value : null
}

function formatMultiplier(value: number): string {
  return Number.isInteger(value) ? String(value) : value.toFixed(2).replace(/0+$/, '').replace(/\.$/, '')
}
