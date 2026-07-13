import { describe, expect, it } from 'vitest'
import {
  completeModelUsageBuckets,
  createModelUsageTickFormatter,
} from '../modelUsageTrendAxis'

describe('completeModelUsageBuckets', () => {
  it('fills two-hour buckets from the exact range start', () => {
    expect(completeModelUsageBuckets(
      ['2026-07-10 13:25', '2026-07-12 11:25'],
      '2h',
      '2026-07-10T13:25:00+08:00',
      '2026-07-12T13:25:00+08:00',
    )).toHaveLength(24)
    expect(completeModelUsageBuckets(
      [],
      '2h',
      '2026-07-10T13:25:00+08:00',
      '2026-07-10T19:25:00+08:00',
    )).toEqual([
      '2026-07-10 13:25',
      '2026-07-10 15:25',
      '2026-07-10 17:25',
    ])
  })

  it('fills inactive hours inside a short hourly range', () => {
    expect(completeModelUsageBuckets([
      '2026-07-12 10:00',
      '2026-07-12 13:00',
    ], 'hour')).toEqual([
      '2026-07-12 10:00',
      '2026-07-12 11:00',
      '2026-07-12 12:00',
      '2026-07-12 13:00',
    ])
  })

  it('leaves daily data and long hourly ranges unchanged', () => {
    expect(completeModelUsageBuckets(['2026-07-11', '2026-07-12'], 'day')).toEqual([
      '2026-07-11',
      '2026-07-12',
    ])
    expect(completeModelUsageBuckets([
      '2026-07-01 00:00',
      '2026-07-08 00:00',
    ], 'hour')).toEqual([
      '2026-07-01 00:00',
      '2026-07-08 00:00',
    ])
  })

  it('fills RFC3339 instants without collapsing the repeated DST hour', () => {
    expect(completeModelUsageBuckets([
      '2026-11-01T05:00:00Z',
      '2026-11-01T06:00:00Z',
    ], 'hour', '2026-11-01T04:00:00Z', '2026-11-01T08:00:00Z')).toEqual([
      '2026-11-01T04:00:00Z',
      '2026-11-01T05:00:00Z',
      '2026-11-01T06:00:00Z',
      '2026-11-01T07:00:00Z',
    ])
  })
})

describe('createModelUsageTickFormatter', () => {
  it('shows only time for hourly buckets within one day', () => {
    const format = createModelUsageTickFormatter([
      '2026-07-12 08:00',
      '2026-07-12 14:00',
    ], 'hour')

    expect(format(0)).toBe('08:00')
    expect(format(1)).toBe('14:00')
  })

  it('adds month and day when hourly buckets cross a date boundary', () => {
    const format = createModelUsageTickFormatter([
      '2026-07-11 23:00',
      '2026-07-12 00:00',
    ], 'hour')

    expect(format(0)).toBe('07-11 23:00')
    expect(format(1)).toBe('07-12 00:00')
  })

  it('keeps daily labels short unless the range crosses years', () => {
    const sameYear = createModelUsageTickFormatter(['2026-01-01', '2026-07-12'], 'day')
    const crossYear = createModelUsageTickFormatter(['2025-12-31', '2026-01-01'], 'day')

    expect(sameYear(1)).toBe('07-12')
    expect(crossYear(1)).toBe('2026-01-01')
  })

  it('falls back to the source label for unexpected bucket formats', () => {
    const format = createModelUsageTickFormatter(['unknown'], 'hour')

    expect(format(0)).toBe('unknown')
  })

  it('formats RFC3339 buckets in the requested non-server timezone', () => {
    const format = createModelUsageTickFormatter([
      '2026-07-13T00:00:00Z',
      '2026-07-13T01:00:00Z',
    ], 'hour', 'Asia/Shanghai')

    expect(format(0)).toBe('08:00')
    expect(format(1)).toBe('09:00')
  })

  it('adds UTC offsets when a DST fallback repeats the same wall-clock hour', () => {
    const format = createModelUsageTickFormatter([
      '2026-11-01T05:00:00Z',
      '2026-11-01T06:00:00Z',
    ], 'hour', 'America/New_York')

    expect(format(0)).toContain('01:00')
    expect(format(1)).toContain('01:00')
    expect(format(0)).not.toBe(format(1))
    expect(format(0)).toMatch(/GMT-4|EDT/)
    expect(format(1)).toMatch(/GMT-5|EST/)
  })
})
