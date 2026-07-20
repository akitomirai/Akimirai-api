import { describe, expect, it } from 'vitest'
import {
  deriveModelAvailabilityStatus,
  filterModelCatalogItems,
  findCatalogModel,
  formatMultiplierRange,
  getModelCatalogFilterOptions,
  getModelPricingSummary,
  isModelAvailabilityErrorCode,
  selectQuickStartCatalogModel,
  toModelCatalogItems,
} from '../modelCatalog'
import type { UserModelCatalogItem } from '@/api/channels'

const catalogPayload: UserModelCatalogItem[] = [
  {
    id: 'openai:gpt-test-model',
    display_name: 'GPT test model',
    model_id: 'gpt-test-model',
    provider: 'openai',
    family: 'GPT',
    status: 'available',
    status_reason: '当前有可用渠道',
    billing_multiplier: 1.5,
    billing_description: '1.5x',
    supports_streaming: true,
    supports_vision: false,
    supports_tools: true,
    supports_json: true,
    context_window: 128000,
    recommended_use: null,
    available_channel_count: 2,
    quick_start_url: '/quick-start?model=gpt-test-model',
    updated_at: '2026-06-30T01:00:00Z',
    channels: ['safe-visible-channel'],
    groups: [
      {
        id: 10,
        name: 'Pro',
        platform: 'openai',
        subscription_type: 'standard',
        rate_multiplier: 1.5,
        peak_rate_enabled: false,
        peak_start: '',
        peak_end: '',
        peak_rate_multiplier: 0,
        is_exclusive: false,
      },
    ],
    pricing: {
      billing_mode: 'token',
      input_price: 0.000001,
      output_price: 0.000002,
      cache_write_price: null,
      cache_read_price: null,
      image_output_price: null,
      per_request_price: null,
      intervals: [],
    },
    offers: [
      {
        channel: 'channel-a',
        platform: 'openai',
        groups: [
          {
            id: 10,
            name: 'Pro',
            platform: 'openai',
            subscription_type: 'standard',
            rate_multiplier: 1.5,
            peak_rate_enabled: true,
            peak_start: '09:00',
            peak_end: '12:00',
            peak_rate_multiplier: 2,
            is_exclusive: false,
          },
        ],
        pricing: {
          billing_mode: 'token',
          input_price: 0.000001,
          output_price: 0.000002,
          cache_write_price: null,
          cache_read_price: null,
          image_output_price: null,
          per_request_price: null,
          intervals: [],
        },
      },
    ],
  },
]

describe('modelCatalog', () => {
  it('maps backend catalog DTOs without deriving fake availability', () => {
    const catalog = toModelCatalogItems(catalogPayload)

    expect(catalog).toHaveLength(1)
    expect(catalog[0]).toMatchObject({
      id: 'gpt-test-model',
      modelId: 'gpt-test-model',
      displayName: 'GPT test model',
      status: 'available',
      provider: 'openai',
      platform: 'openai',
      family: 'GPT',
      billingMultiplier: 1.5,
      availableChannelCount: 2,
      supportsStreaming: true,
      supportsVision: false,
      supportsTools: true,
      supportsJson: true,
      contextWindow: 128000,
      channelNames: ['channel-a'],
    })
    expect(formatMultiplierRange(catalog[0])).toBe('1.5x')
    expect(catalog[0].offers[0].groups[0]).toMatchObject({
      peakRateEnabled: true,
      peakStart: '09:00',
      peakEnd: '12:00',
      peakRateMultiplier: 2,
    })
  })

  it('keeps channel pricing associated with each offer and filters against offer data', () => {
    const firstOffer = catalogPayload[0].offers![0]
    const payload: UserModelCatalogItem[] = [{
      ...catalogPayload[0],
      offers: [
        firstOffer,
        {
          ...firstOffer,
          channel: 'channel-b',
          groups: [{
            ...firstOffer.groups[0],
            id: 20,
            name: 'Basic',
          }],
          pricing: {
            ...firstOffer.pricing!,
            input_price: 0.000003,
            output_price: 0.000006,
          },
        },
      ],
    }]

    const catalog = toModelCatalogItems(payload)
    expect(catalog[0].offers.map((offer) => ({
      channel: offer.channel,
      group: offer.groups[0].name,
      input: offer.pricing?.input_price,
    }))).toEqual([
      { channel: 'channel-a', group: 'Pro', input: 0.000001 },
      { channel: 'channel-b', group: 'Basic', input: 0.000003 },
    ])
    expect(getModelPricingSummary(catalog[0])).toMatchObject({ kind: 'multiple', pricedOfferCount: 2 })
    expect(filterModelCatalogItems(catalog, { group: 'Basic' })).toHaveLength(1)
    expect(filterModelCatalogItems(catalog, { group: 'Missing' })).toHaveLength(0)
    expect(getModelCatalogFilterOptions(catalog)).toMatchObject({
      groups: ['Basic', 'Pro'],
      providers: ['openai'],
      billingModes: ['token'],
    })
  })

  it('keeps status derivation conservative without faking availability', () => {
    expect(deriveModelAvailabilityStatus({
      modelEnabled: true,
      hasAvailableChannel: true,
      hasModelConfig: true,
      hasSufficientData: true,
    })).toBe('available')

    expect(deriveModelAvailabilityStatus({
      modelEnabled: false,
      hasAvailableChannel: true,
      hasModelConfig: true,
      hasSufficientData: true,
    })).toBe('maintenance')

    expect(deriveModelAvailabilityStatus({
      modelEnabled: true,
      hasAvailableChannel: false,
      hasModelConfig: true,
      hasSufficientData: true,
    })).toBe('unavailable')

    expect(deriveModelAvailabilityStatus({
      hasSufficientData: false,
    })).toBe('unknown')
  })

  it('selects the requested Quick Start model only when it is available', () => {
    const catalog = toModelCatalogItems(catalogPayload)

    expect(findCatalogModel(catalog, 'GPT-TEST-MODEL')?.id).toBe('gpt-test-model')
    expect(selectQuickStartCatalogModel(catalog, 'gpt-test-model')).toMatchObject({
      selected: catalog[0],
      requested: 'gpt-test-model',
      usedFallback: false,
    })
    expect(selectQuickStartCatalogModel(catalog, 'missing-model')).toMatchObject({
      selected: catalog[0],
      requested: 'missing-model',
      usedFallback: true,
    })
  })

  it('does not copy sensitive extra payload fields into catalog items', () => {
    const unsafePayload = [
      {
        ...catalogPayload[0],
        upstream_token: 'token-should-not-render',
        cookie: 'cookie-should-not-render',
        private_key: 'private-key-should-not-render',
        prompt: 'prompt-should-not-render',
        api_key: 'sk-test-placeholder',
      },
    ] as unknown as UserModelCatalogItem[]

    const serialized = JSON.stringify(toModelCatalogItems(unsafePayload))

    expect(serialized).not.toContain('token-should-not-render')
    expect(serialized).not.toContain('cookie-should-not-render')
    expect(serialized).not.toContain('private-key-should-not-render')
    expect(serialized).not.toContain('prompt-should-not-render')
    expect(serialized).not.toContain('sk-test-placeholder')
  })

  it('recognizes model availability error codes for dashboard guidance', () => {
    expect(isModelAvailabilityErrorCode('MODEL_DISABLED')).toBe(true)
    expect(isModelAvailabilityErrorCode('NO_AVAILABLE_CHANNEL')).toBe(true)
    expect(isModelAvailabilityErrorCode('UPSTREAM_5XX')).toBe(false)
  })
})
