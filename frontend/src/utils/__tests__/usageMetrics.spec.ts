import { describe, expect, it } from 'vitest'

import {
  cacheHitProgressWidth,
  cacheHitRatio,
  formatUsageDuration,
  hasKnownTokenMetrics,
  totalUsageTokens,
} from '@/utils/usageMetrics'

describe('usageMetrics', () => {
  it('totals all four token categories without treating missing values as unknown totals', () => {
    expect(totalUsageTokens({
      input_tokens: 100,
      output_tokens: 20,
      cache_creation_tokens: 30,
      cache_read_tokens: 850,
    })).toBe(1_000)
    expect(hasKnownTokenMetrics({ input_tokens: null, output_tokens: null })).toBe(false)
    expect(hasKnownTokenMetrics({ input_tokens: 0, output_tokens: 0 })).toBe(true)
  })

  it('reuses cache read over input, read, and creation as the hit ratio', () => {
    const metrics = {
      input_tokens: 100,
      cache_read_tokens: 850,
      cache_creation_tokens: 50,
    }
    expect(cacheHitRatio(metrics)).toBe(85)
    expect(cacheHitProgressWidth(metrics)).toBe('85%')
    expect(cacheHitRatio({ input_tokens: 0, cache_read_tokens: 0, cache_creation_tokens: 0 })).toBeNull()
  })

  it('formats latency across millisecond, second, minute, and hour boundaries', () => {
    expect(formatUsageDuration(null)).toBe('-')
    expect(formatUsageDuration(450)).toBe('450ms')
    expect(formatUsageDuration(2_370)).toBe('2.37s')
    expect(formatUsageDuration(60_000)).toBe('1m 0s')
    expect(formatUsageDuration(3_660_000)).toBe('1h 1m')
  })
})
