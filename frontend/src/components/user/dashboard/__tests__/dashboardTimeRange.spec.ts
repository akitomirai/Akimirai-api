import { describe, expect, it } from 'vitest'
import {
  dashboardDatePresets,
  getDashboardPresetGranularity,
  getDashboardPresetPeriod,
} from '@/utils/dashboardTimeRange'

describe('dashboard time range contract', () => {
  it('keeps the requested preset order', () => {
    expect(dashboardDatePresets).toEqual([
      'yesterday',
      'today',
      'last24Hours',
      'last48Hours',
      '7days',
      '14days',
      '30days',
    ])
  })

  it.each([
    ['yesterday', 'yesterday', 'hour'],
    ['today', 'today', 'hour'],
    ['last24Hours', '24h', 'hour'],
    ['last48Hours', '48h', '2h'],
    ['7days', '7d', '4h'],
    ['14days', '14d', '8h'],
    ['30days', '30d', 'day'],
  ] as const)('maps %s to period %s and granularity %s', (preset, period, granularity) => {
    expect(getDashboardPresetPeriod(preset)).toBe(period)
    expect(getDashboardPresetGranularity(preset)).toBe(granularity)
  })

  it('does not impose rolling defaults on a custom range', () => {
    expect(getDashboardPresetPeriod(null)).toBeNull()
    expect(getDashboardPresetGranularity(null)).toBeNull()
  })
})
