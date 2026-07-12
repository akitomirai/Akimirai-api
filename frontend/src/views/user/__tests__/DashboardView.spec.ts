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
        props: ['startDate', 'endDate', 'granularity', 'loading'],
        emits: [
          'update:startDate',
          'update:endDate',
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
              @click="$emit('update:startDate', '2026-07-01'); $emit('update:endDate', '2026-07-02'); $emit('dateRangeChange')"
            />
          </div>
        `,
      },
      ModelUsageTrendChart: {
        props: ['trendData', 'metric', 'loading'],
        template: '<section :data-testid="`model-usage-trend-${metric}`">{{ trendData.length }}</section>',
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
      start_date: '2026-07-06',
      end_date: '2026-07-12',
      granularity: 'day',
    })
    mocks.getMyPlatformQuotas.mockResolvedValue({ platform_quotas: [] })
  })

  it('renders both model trend charts with announcements and retires old data calls', async () => {
    const wrapper = mountDashboard()
    await flushPromises()

    expect(wrapper.find('[data-testid="model-usage-trend-actual_cost"]').text()).toBe('1')
    expect(wrapper.find('[data-testid="model-usage-trend-requests"]').text()).toBe('1')
    expect(wrapper.find('[data-testid="announcements"]').exists()).toBe(true)
    expect(mocks.getDashboardModelTrend).toHaveBeenCalledWith(expect.objectContaining({
      granularity: 'day',
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
      start_date: '2026-07-01',
      end_date: '2026-07-02',
    }))
  })
})
