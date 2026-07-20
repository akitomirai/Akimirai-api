import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  refreshUser: vi.fn(),
  getDashboardStats: vi.fn(),
  getDashboardModelTrend: vi.fn(),
  getDashboardTrend: vi.fn(),
  getDashboardModels: vi.fn(),
  getByDateRange: vi.fn(),
  getMyPlatformQuotas: vi.fn(),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    user: { balance: 25 },
    isSimpleMode: false,
    refreshUser: mocks.refreshUser,
  }),
}))

vi.mock('@/api/usage', () => ({
  usageAPI: {
    getDashboardStats: mocks.getDashboardStats,
    getDashboardModelTrend: mocks.getDashboardModelTrend,
    getDashboardTrend: mocks.getDashboardTrend,
    getDashboardModels: mocks.getDashboardModels,
    getByDateRange: mocks.getByDateRange,
  },
}))

vi.mock('@/api/user', () => ({
  getMyPlatformQuotas: mocks.getMyPlatformQuotas,
}))

import DashboardView from '../DashboardView.vue'

const trend = [
  {
    date: '2026-07-12',
    model: 'gpt-5.6-sol',
    is_other: false,
    requests: 4,
    total_tokens: 120,
    cost: 0.4,
    actual_cost: 0.3,
  },
]

const mountDashboard = () => mount(DashboardView, {
  global: {
    stubs: {
      AppLayout: { template: '<main><slot /></main>' },
      LoadingSpinner: { template: '<div data-testid="loading" />' },
      UserDashboardStats: { template: '<section data-testid="dashboard-stats" />' },
      UserDashboardFilters: {
        props: ['startDate', 'endDate', 'startTime', 'endTime', 'granularity', 'loading'],
        emits: [
          'update:startDate',
          'update:endDate',
          'update:startTime',
          'update:endTime',
          'update:granularity',
          'dateRangeChange',
          'granularityChange',
          'refresh',
        ],
        template: `
          <div data-testid="dashboard-filters">
            <button
              data-testid="set-hour"
              @click="$emit('update:granularity', 'hour'); $emit('granularityChange')"
            />
            <button
              data-testid="set-range"
              @click="$emit('update:startDate', '2026-07-01'); $emit('update:endDate', '2026-07-02'); $emit('update:startTime', '08:30'); $emit('update:endTime', '19:45'); $emit('dateRangeChange', { startDate: '2026-07-01', endDate: '2026-07-02', startTime: '08:30', endTime: '19:45', preset: null })"
            />
            <button
              data-testid="set-48-hours"
              @click="$emit('update:startDate', '2026-07-10'); $emit('update:endDate', '2026-07-12'); $emit('update:startTime', '22:00'); $emit('update:endTime', '22:00'); $emit('dateRangeChange', { startDate: '2026-07-10', endDate: '2026-07-12', startTime: '22:00', endTime: '22:00', preset: 'last48Hours' })"
            />
          </div>
        `,
      },
      ModelUsageTrendChart: {
        props: ['trendData', 'metric', 'loading', 'granularity', 'rangeStart', 'rangeEnd'],
        template: '<section :data-testid="`model-usage-trend-${metric}`" :data-granularity="granularity" :data-range-start="rangeStart" :data-range-end="rangeEnd">{{ trendData.length }}</section>',
      },
      UserDashboardAnnouncements: { template: '<aside data-testid="announcements" />' },
    },
  },
})

describe('user DashboardView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.refreshUser.mockResolvedValue(undefined)
    mocks.getDashboardStats.mockResolvedValue({ total_requests: 4 })
    mocks.getDashboardModelTrend.mockResolvedValue({
      trend,
      models: ['gpt-5.6-sol'],
      start_date: '2026-07-11',
      end_date: '2026-07-12',
      start_time: '2026-07-11T17:00:00+08:00',
      end_time: '2026-07-12T17:00:00+08:00',
      granularity: 'hour',
    })
    mocks.getMyPlatformQuotas.mockResolvedValue({ platform_quotas: [] })
  })

  it('renders both model trend charts with announcements and retires old data calls', async () => {
    const wrapper = mountDashboard()
    await flushPromises()

    expect(wrapper.find('[data-testid="model-usage-trend-actual_cost"]').text()).toBe('1')
    expect(wrapper.find('[data-testid="model-usage-trend-total_tokens"]').text()).toBe('1')
    expect(wrapper.find('[data-testid="model-usage-trend-total_tokens"]').attributes('data-granularity')).toBe('hour')
    expect(wrapper.find('[data-testid="model-usage-trend-total_tokens"]').attributes('data-range-start')).toBe('2026-07-11T17:00:00+08:00')
    expect(wrapper.find('[data-testid="model-usage-trend-total_tokens"]').attributes('data-range-end')).toBe('2026-07-12T17:00:00+08:00')
    expect(wrapper.find('[data-testid="announcements"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="dashboard-content-grid"]').classes()).toContain('lg:grid-cols-[minmax(0,5fr)_minmax(14rem,1fr)]')
    expect(mocks.getDashboardModelTrend).toHaveBeenCalledWith(expect.objectContaining({
      granularity: 'hour',
      period: '24h',
      model_source: 'requested',
    }))
    expect(mocks.getDashboardTrend).not.toHaveBeenCalled()
    expect(mocks.getDashboardModels).not.toHaveBeenCalled()
    expect(mocks.getByDateRange).not.toHaveBeenCalled()
  })

  it('reloads the model trend when granularity or date range changes', async () => {
    const wrapper = mountDashboard()
    await flushPromises()

    await wrapper.get('[data-testid="set-hour"]').trigger('click')
    await flushPromises()
    expect(mocks.getDashboardModelTrend).toHaveBeenLastCalledWith(expect.objectContaining({
      granularity: 'hour',
    }))

    await wrapper.get('[data-testid="set-range"]').trigger('click')
    await flushPromises()
    expect(mocks.getDashboardModelTrend).toHaveBeenLastCalledWith(expect.objectContaining({
      start_time: '2026-07-01T08:30',
      end_time: '2026-07-02T19:45',
    }))
    expect(mocks.getDashboardModelTrend).toHaveBeenLastCalledWith(expect.not.objectContaining({
      period: expect.anything(),
    }))
  })

  it('maps the rolling 48-hour preset to two-hour buckets', async () => {
    const wrapper = mountDashboard()
    await flushPromises()

    await wrapper.get('[data-testid="set-48-hours"]').trigger('click')
    await flushPromises()

    expect(mocks.getDashboardModelTrend).toHaveBeenLastCalledWith(expect.objectContaining({
      period: '48h',
      granularity: '2h',
    }))
  })
})
