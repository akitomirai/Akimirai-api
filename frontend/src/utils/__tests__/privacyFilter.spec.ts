import { describe, expect, it } from 'vitest'
import {
  DEFAULT_PRIVACY_FILTER_TYPES,
  normalizePrivacyFilterConfig,
} from '@/utils/privacyFilter'

describe('normalizePrivacyFilterConfig', () => {
  it('uses default types when types are missing', () => {
    expect(normalizePrivacyFilterConfig({ enabled: true })).toEqual({
      enabled: true,
      types: DEFAULT_PRIVACY_FILTER_TYPES,
    })
  })

  it('filters invalid values, removes duplicates, and keeps canonical order', () => {
    expect(normalizePrivacyFilterConfig({
      enabled: true,
      types: ['token', 'email', 'invalid' as never, 'token', 'api_key'],
    })).toEqual({
      enabled: true,
      types: ['email', 'api_key', 'token'],
    })
  })
})
