import type { ModelUsageTrendGranularity, UserUsagePeriod } from '@/api/usage'

export type DashboardDatePreset =
  | 'yesterday'
  | 'today'
  | 'last24Hours'
  | 'last48Hours'
  | '7days'
  | '14days'
  | '30days'

export interface DateRangeChange {
  startDate: string
  endDate: string
  startTime: string
  endTime: string
  preset: string | null
}

export const toLocalDateTimeParam = (date: string, time: string): string =>
  `${date}T${time}`

export const dashboardDatePresets: DashboardDatePreset[] = [
  'yesterday',
  'today',
  'last24Hours',
  'last48Hours',
  '7days',
  '14days',
  '30days'
]

const presetPeriods: Record<DashboardDatePreset, UserUsagePeriod> = {
  yesterday: 'yesterday',
  today: 'today',
  last24Hours: '24h',
  last48Hours: '48h',
  '7days': '7d',
  '14days': '14d',
  '30days': '30d'
}

const dashboardPeriods = Array.from(new Set(Object.values(presetPeriods)))

const presetGranularities: Record<DashboardDatePreset, ModelUsageTrendGranularity> = {
  yesterday: 'hour',
  today: 'hour',
  last24Hours: 'hour',
  last48Hours: '2h',
  '7days': '4h',
  '14days': '8h',
  '30days': 'day'
}

export const dashboardGranularities: ModelUsageTrendGranularity[] = [
  'hour',
  '2h',
  '4h',
  '8h',
  'day'
]

export const isDashboardDatePreset = (value: string | null): value is DashboardDatePreset =>
  value !== null && dashboardDatePresets.includes(value as DashboardDatePreset)

export const getDashboardPresetPeriod = (preset: string | null): UserUsagePeriod | null =>
  isDashboardDatePreset(preset) ? presetPeriods[preset] : null

export const isDashboardUsagePeriod = (value: unknown): value is UserUsagePeriod =>
  typeof value === 'string' && dashboardPeriods.includes(value as UserUsagePeriod)

export const getDashboardPresetGranularity = (
  preset: string | null
): ModelUsageTrendGranularity | null =>
  isDashboardDatePreset(preset) ? presetGranularities[preset] : null

export const isDashboardGranularity = (value: unknown): value is ModelUsageTrendGranularity =>
  typeof value === 'string' && dashboardGranularities.includes(value as ModelUsageTrendGranularity)
