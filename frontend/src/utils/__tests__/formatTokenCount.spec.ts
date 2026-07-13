import { describe, expect, it } from 'vitest'

import { formatTokenCount } from '@/utils/format'

describe('formatTokenCount', () => {
  it('uses Chinese 万 and 亿 units for compact token counts', () => {
    expect(formatTokenCount(9_999, { locale: 'zh-CN' })).toBe('9,999')
    expect(formatTokenCount(17_196, { locale: 'zh-CN' })).toBe('1.72万')
    expect(formatTokenCount(137_440_000, { locale: 'zh-CN' })).toBe('1.37亿')
    expect(formatTokenCount(1_500_000_000, { locale: 'zh-CN' })).toBe('15.00亿')
  })

  it('keeps exact token counts fully expanded and measured in individual tokens', () => {
    expect(formatTokenCount(137_440_000, { locale: 'zh-CN', display: 'exact' })).toBe(
      '137,440,000个'
    )
  })

  it('uses locale compact notation outside Chinese locales', () => {
    expect(formatTokenCount(1_500_000, { locale: 'en-US' })).toBe('1.50M')
    expect(formatTokenCount(1_500_000, { locale: 'en-US', display: 'exact' })).toBe('1,500,000')
  })

  it('normalizes missing and non-finite values to zero', () => {
    expect(formatTokenCount(null, { locale: 'zh-CN' })).toBe('0')
    expect(formatTokenCount(Number.NaN, { locale: 'zh-CN' })).toBe('0')
  })
})
