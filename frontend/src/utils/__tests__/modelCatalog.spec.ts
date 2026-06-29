import { describe, expect, it } from 'vitest'
import {
  deriveModelAvailabilityStatus,
  deriveModelCatalog,
  findCatalogModel,
  formatMultiplierRange,
  isModelAvailabilityErrorCode,
  selectQuickStartCatalogModel,
} from '../modelCatalog'
import type { UserAvailableChannel } from '@/api/channels'

const channels: UserAvailableChannel[] = [
  {
    name: 'safe-visible-channel',
    description: 'visible',
    platforms: [
      {
        platform: 'openai',
        groups: [
          {
            id: 10,
            name: 'Pro',
            platform: 'openai',
            subscription_type: 'standard',
            rate_multiplier: 1.2,
            is_exclusive: false,
          },
        ],
        supported_models: [
          {
            name: 'gpt-test-model',
            platform: 'openai',
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
    ],
  },
]

describe('modelCatalog', () => {
  it('derives available models and effective multipliers from user-visible channels', () => {
    const catalog = deriveModelCatalog(channels, { 10: 1.5 })

    expect(catalog).toHaveLength(1)
    expect(catalog[0].id).toBe('gpt-test-model')
    expect(catalog[0].status).toBe('available')
    expect(catalog[0].platform).toBe('openai')
    expect(catalog[0].channelNames).toEqual(['safe-visible-channel'])
    expect(formatMultiplierRange(catalog[0])).toBe('1.5x')
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
    })).toBe('temporarily_unavailable')

    expect(deriveModelAvailabilityStatus({
      hasSufficientData: false,
    })).toBe('unknown')
  })

  it('selects the requested Quick Start model only when it is available', () => {
    const catalog = deriveModelCatalog(channels)

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
    const unsafeChannels = [
      {
        ...channels[0],
        upstream_token: 'token-should-not-render',
        cookie: 'cookie-should-not-render',
        private_key: 'private-key-should-not-render',
        prompt: 'prompt-should-not-render',
        platforms: [
          {
            ...channels[0].platforms[0],
            supported_models: [
              {
                ...channels[0].platforms[0].supported_models[0],
                api_key: 'sk-test-placeholder',
              },
            ],
          },
        ],
      },
    ] as unknown as UserAvailableChannel[]

    const serialized = JSON.stringify(deriveModelCatalog(unsafeChannels))

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
