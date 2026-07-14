import { afterEach, describe, expect, it } from 'vitest'

import { formatTokenCount } from '@/utils/format'
import { resetTokenCountModeForTests, setTokenCountMode } from '@/composables/useTokenCountMode'

describe('formatTokenCount', () => {
  afterEach(() => {
    localStorage.clear()
    resetTokenCountModeForTests()
  })

  it('uses w below 1e8 and promotes Chinese modern counts to 亿', () => {
    expect(formatTokenCount(9_999, { locale: 'zh-CN' })).toBe('9,999')
    expect(formatTokenCount(17_196, { locale: 'zh-CN' })).toBe('1.72w')
    expect(formatTokenCount(99_999_999, { locale: 'zh-CN' })).toBe('10000.00w')
    expect(formatTokenCount(137_440_000, { locale: 'zh-CN' })).toBe('1.37亿')
    expect(formatTokenCount(1_500_000_000, { locale: 'zh-CN' })).toBe('15.00亿')
  })

  it('keeps exact token counts fully expanded and measured in individual tokens', () => {
    expect(formatTokenCount(137_440_000, { locale: 'zh-CN', display: 'exact' })).toBe(
      '137,440,000'
    )
  })

  it('uses the same modern unit outside Chinese locales', () => {
    expect(formatTokenCount(1_500_000, { locale: 'en-US' })).toBe('150.00w')
    expect(formatTokenCount(1_500_000, { locale: 'en-US', display: 'exact' })).toBe('1,500,000')
  })

  it('normalizes missing and non-finite values to zero', () => {
    expect(formatTokenCount(null, { locale: 'zh-CN' })).toBe('0')
    expect(formatTokenCount(Number.NaN, { locale: 'zh-CN' })).toBe('0')
  })

  it('uses lower-case k, m, and b units in legacy mode', () => {
    setTokenCountMode('legacy')

    expect(formatTokenCount(1_475, { locale: 'zh-CN' })).toBe('1.48k')
    expect(formatTokenCount(87_280_000, { locale: 'zh-CN' })).toBe('87.28m')
    expect(formatTokenCount(1_500_000_000, { locale: 'zh-CN' })).toBe('1.50b')
    expect(formatTokenCount(1_475, { locale: 'zh-CN', display: 'exact' })).toBe('1,475')
  })

  it('supports the modern request-row lower-case w unit explicitly', () => {
    expect(formatTokenCount(9_999, { locale: 'zh-CN', unitStyle: 'wan-latin' })).toBe('9,999')
    expect(formatTokenCount(17_196, { locale: 'zh-CN', unitStyle: 'wan-latin' })).toBe('1.72w')
    expect(formatTokenCount(137_440_000, { locale: 'zh-CN', unitStyle: 'wan-latin' })).toBe('13744.00w')
  })
})
