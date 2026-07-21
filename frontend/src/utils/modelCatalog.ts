import type {
  UserAvailableGroup,
  UserModelCatalogItem,
  UserModelCatalogOffer,
  UserPricingInterval,
  UserSupportedModelPricing,
} from '@/api/channels'
import type { BillingMode } from '@/constants/channel'

export type ModelAvailabilityStatus = 'available' | 'maintenance' | 'unavailable' | 'unknown'

export interface ModelCatalogGroup {
  id: number
  name: string
  platform: string
  subscriptionType: string
  rateMultiplier: number
  effectiveRateMultiplier: number
  peakRateEnabled: boolean
  peakStart: string
  peakEnd: string
  peakRateMultiplier: number
  isExclusive: boolean
}

export interface ModelCatalogOffer {
  channel: string
  platform: string
  groups: ModelCatalogGroup[]
  pricing: UserSupportedModelPricing | null
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
  offers: ModelCatalogOffer[]
  supportsStreaming: boolean | null
  supportsVision: boolean | null
  supportsTools: boolean | null
  supportsJson: boolean | null
  contextWindow: number | null
  recommendedUse: string | null
}

export type ModelCatalogCapability = 'streaming' | 'vision' | 'tools' | 'json'

export interface ModelCatalogFilters {
  query?: string
  group?: string
  provider?: string
  billingMode?: BillingMode | ''
  status?: ModelAvailabilityStatus | ''
  capability?: ModelCatalogCapability | ''
}

export interface ModelCatalogFilterOptions {
  groups: string[]
  providers: string[]
  billingModes: BillingMode[]
  statuses: ModelAvailabilityStatus[]
  capabilities: ModelCatalogCapability[]
}

export interface ModelCatalogPricingSummary {
  kind: 'none' | 'single' | 'multiple'
  pricedOfferCount: number
  billingModes: BillingMode[]
  pricing: UserSupportedModelPricing | null
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
      channelNames: normalizedChannelNames(item),
      groups: (item.groups || []).map(toCatalogGroup),
      pricing: normalizePricing(item.pricing),
      offers: normalizeOffers(item.offers),
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

export function getModelCatalogGroups(item: ModelCatalogItem): ModelCatalogGroup[] {
  if (item.offers.length === 0) return item.groups

  const groups = new Map<number, ModelCatalogGroup>()
  for (const offer of item.offers) {
    for (const group of offer.groups) {
      if (!groups.has(group.id)) groups.set(group.id, group)
    }
  }
  return Array.from(groups.values()).sort((a, b) => a.name.localeCompare(b.name))
}

export function getModelCatalogFilterOptions(items: ModelCatalogItem[]): ModelCatalogFilterOptions {
  const groups = new Set<string>()
  const providers = new Set<string>()
  const billingModes = new Set<BillingMode>()
  const statuses = new Set<ModelAvailabilityStatus>()
  const capabilities = new Set<ModelCatalogCapability>()

  for (const item of items) {
    providers.add(item.provider)
    statuses.add(item.status)
    for (const group of getModelCatalogGroups(item)) groups.add(group.name)
    for (const pricing of getItemPricing(item)) billingModes.add(pricing.billing_mode)
    if (item.supportsStreaming === true) capabilities.add('streaming')
    if (item.supportsVision === true) capabilities.add('vision')
    if (item.supportsTools === true) capabilities.add('tools')
    if (item.supportsJson === true) capabilities.add('json')
  }

  return {
    groups: Array.from(groups).sort((a, b) => a.localeCompare(b)),
    providers: Array.from(providers).sort((a, b) => a.localeCompare(b)),
    billingModes: Array.from(billingModes).sort(),
    statuses: Array.from(statuses).sort(),
    capabilities: Array.from(capabilities).sort(),
  }
}

export function filterModelCatalogItems(
  items: ModelCatalogItem[],
  filters: ModelCatalogFilters,
): ModelCatalogItem[] {
  const query = cleanString(filters.query).toLowerCase()
  const group = cleanString(filters.group).toLowerCase()
  const provider = cleanString(filters.provider).toLowerCase()

  return items.filter((item) => {
    const itemGroups = getModelCatalogGroups(item)
    const searchable = [
      item.id,
      item.displayName,
      item.modelId,
      item.provider,
      item.family || '',
      ...item.channelNames,
      ...itemGroups.map((candidate) => candidate.name),
    ].join(' ').toLowerCase()

    if (query && !searchable.includes(query)) return false
    if (provider && item.provider.toLowerCase() !== provider) return false
    if (group && !itemGroups.some((candidate) => candidate.name.toLowerCase() === group)) return false
    if (filters.status && item.status !== filters.status) return false
    if (filters.billingMode && !getItemPricing(item).some((pricing) => pricing.billing_mode === filters.billingMode)) {
      return false
    }
    if (filters.capability && !supportsCapability(item, filters.capability)) return false
    return true
  })
}

export function getModelPricingSummary(item: ModelCatalogItem): ModelCatalogPricingSummary {
  const offersWithPricing = item.offers.filter((offer) => offer.pricing !== null)
  const pricing = offersWithPricing.length > 0
    ? offersWithPricing.map((offer) => offer.pricing as UserSupportedModelPricing)
    : item.pricing ? [item.pricing] : []
  const signatures = new Set(pricing.map(pricingSignature))

  if (pricing.length === 0) {
    return { kind: 'none', pricedOfferCount: 0, billingModes: [], pricing: null }
  }

  return {
    kind: signatures.size > 1 ? 'multiple' : 'single',
    pricedOfferCount: offersWithPricing.length || 1,
    billingModes: Array.from(new Set(pricing.map((entry) => entry.billing_mode))).sort(),
    pricing: signatures.size === 1 ? pricing[0] : null,
  }
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
    peakRateEnabled: group.peak_rate_enabled === true,
    peakStart: cleanString(group.peak_start),
    peakEnd: cleanString(group.peak_end),
    peakRateMultiplier: finiteNumberOrFallback(group.peak_rate_multiplier, 0),
    isExclusive: group.is_exclusive === true,
  }
}

function normalizeOffers(offers: UserModelCatalogOffer[] | undefined): ModelCatalogOffer[] {
  if (!Array.isArray(offers)) return []
  return offers
    .map((offer) => ({
      channel: cleanString(offer.channel),
      platform: cleanString(offer.platform),
      groups: Array.isArray(offer.groups) ? offer.groups.map(toCatalogGroup) : [],
      pricing: normalizePricing(offer.pricing),
    }))
    .filter((offer) => offer.channel || offer.platform || offer.groups.length > 0 || offer.pricing !== null)
}

function normalizedChannelNames(item: UserModelCatalogItem): string[] {
  const offerNames = Array.isArray(item.offers)
    ? item.offers.map((offer) => cleanString(offer.channel)).filter(Boolean)
    : []
  return uniqueStrings(offerNames.length > 0 ? offerNames : item.channels)
}

function normalizePricing(pricing: UserSupportedModelPricing | null | undefined): UserSupportedModelPricing | null {
  if (!pricing) return null
  return {
    billing_mode: pricing.billing_mode,
    input_price: finiteNumberOrNull(pricing.input_price),
    output_price: finiteNumberOrNull(pricing.output_price),
    cache_write_price: finiteNumberOrNull(pricing.cache_write_price),
    cache_read_price: finiteNumberOrNull(pricing.cache_read_price),
    image_input_price: finiteNumberOrNull(pricing.image_input_price),
    image_output_price: finiteNumberOrNull(pricing.image_output_price),
    per_request_price: finiteNumberOrNull(pricing.per_request_price),
    intervals: Array.isArray(pricing.intervals) ? pricing.intervals.map(normalizeInterval) : [],
  }
}

function normalizeInterval(interval: UserPricingInterval): UserPricingInterval {
  return {
    min_tokens: Math.max(0, Number(interval.min_tokens) || 0),
    max_tokens: interval.max_tokens == null ? null : Math.max(0, Number(interval.max_tokens) || 0),
    tier_label: cleanNullableString(interval.tier_label) || undefined,
    input_price: finiteNumberOrNull(interval.input_price),
    output_price: finiteNumberOrNull(interval.output_price),
    cache_write_price: finiteNumberOrNull(interval.cache_write_price),
    cache_read_price: finiteNumberOrNull(interval.cache_read_price),
    per_request_price: finiteNumberOrNull(interval.per_request_price),
  }
}

function getItemPricing(item: ModelCatalogItem): UserSupportedModelPricing[] {
  const offerPricing = item.offers
    .map((offer) => offer.pricing)
    .filter((pricing): pricing is UserSupportedModelPricing => pricing !== null)
  if (offerPricing.length > 0) return offerPricing
  return item.pricing ? [item.pricing] : []
}

function supportsCapability(item: ModelCatalogItem, capability: ModelCatalogCapability): boolean {
  switch (capability) {
    case 'streaming': return item.supportsStreaming === true
    case 'vision': return item.supportsVision === true
    case 'tools': return item.supportsTools === true
    case 'json': return item.supportsJson === true
  }
}

function pricingSignature(pricing: UserSupportedModelPricing): string {
  return JSON.stringify(pricing)
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
