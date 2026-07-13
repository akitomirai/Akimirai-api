import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import type { DashboardStats } from '@/types'
import DashboardView from '../DashboardView.vue'

const { getSnapshotV2, getUserUsageTrend, getUserSpendingRanking, routerPush } = vi.hoisted(() => ({
  getSnapshotV2: vi.fn(),
  getUserUsageTrend: vi.fn(),
  getUserSpendingRanking: vi.fn(),
  routerPush: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    dashboard: {
      getSnapshotV2,
      getUserUsageTrend,
      getUserSpendingRanking
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn()
  })
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: routerPush
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const formatLocalDate = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const createDashboardStats = (): DashboardStats => ({
  total_users: 0,
  today_new_users: 0,
  active_users: 0,
  hourly_active_users: 0,
  stats_updated_at: '',
  stats_stale: false,
  total_api_keys: 0,
  active_api_keys: 0,
  total_accounts: 0,
  normal_accounts: 0,
  error_accounts: 0,
  ratelimit_accounts: 0,
  overload_accounts: 0,
  total_requests: 0,
  total_input_tokens: 0,
  total_output_tokens: 0,
  total_cache_creation_tokens: 0,
  total_cache_read_tokens: 0,
  total_tokens: 0,
  total_cost: 0,
  total_actual_cost: 0,
  today_requests: 0,
  today_input_tokens: 0,
  today_output_tokens: 0,
  today_cache_creation_tokens: 0,
  today_cache_read_tokens: 0,
  today_tokens: 0,
  today_cost: 0,
  today_actual_cost: 0,
  average_duration_ms: 0,
  uptime: 0,
  rpm: 0,
  tpm: 0
})

describe('admin DashboardView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())

    getSnapshotV2.mockReset()
    getUserUsageTrend.mockReset()
    getUserSpendingRanking.mockReset()
    routerPush.mockReset()

    getSnapshotV2.mockResolvedValue({
      stats: createDashboardStats(),
      trend: [],
      models: []
    })
    getUserUsageTrend.mockResolvedValue({
      trend: [],
      start_date: '',
      end_date: '',
      granularity: 'hour'
    })
    getUserSpendingRanking.mockResolvedValue({
      ranking: [],
      total_actual_cost: 0,
      total_requests: 0,
      total_tokens: 0,
      start_date: '',
      end_date: ''
    })
  })

  it('uses last 24 hours as default dashboard range', async () => {
    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: true,
          Select: true,
          UserSpendingRankingChart: true,
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          Line: true
        }
      }
    })

    await flushPromises()

    const now = new Date()
    const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000)

    expect(getSnapshotV2).toHaveBeenCalledTimes(1)
    expect(getSnapshotV2).toHaveBeenCalledWith(expect.objectContaining({
      start_date: formatLocalDate(yesterday),
      end_date: formatLocalDate(now),
      granularity: 'hour'
    }))
    expect(wrapper.text()).not.toContain('admin.dashboard.quickActions')
  })

  it('places spending ranking beside model distribution and token trend above recent usage', async () => {
    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: true,
          Select: true,
          UserSpendingRankingChart: {
            name: 'UserSpendingRankingChart',
            emits: ['select-user'],
            template: '<button @click="$emit(\'select-user\', { user_id: 42 })" />'
          },
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          Line: true
        }
      }
    })

    await flushPromises()

    const sectionOrder = wrapper
      .findAll('[data-testid^="dashboard-"]')
      .map((section) => section.attributes('data-testid'))

    expect(sectionOrder).toEqual([
      'dashboard-user-spending-ranking',
      'dashboard-model-distribution',
      'dashboard-token-usage-trend',
      'dashboard-recent-usage'
    ])

    const ranking = wrapper.get('[data-testid="dashboard-user-spending-ranking"]')
    const models = wrapper.get('[data-testid="dashboard-model-distribution"]')
    expect(ranking.element.parentElement).toBe(models.element.parentElement)
    expect(ranking.element.parentElement?.className).toContain('lg:grid-cols-2')

    await ranking.trigger('click')
    expect(routerPush).toHaveBeenCalledWith({
      path: '/admin/usage',
      query: expect.objectContaining({
        user_id: '42',
        period: '24h',
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
      })
    })
  })

  it('uses the shared preset granularity for the last 48 hours', async () => {
    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: {
            props: ['startDate', 'endDate', 'presetValues'],
            emits: ['update:startDate', 'update:endDate', 'change'],
            template: `
              <button
                data-testid="last-48-hours"
                :data-presets="presetValues.join(',')"
                @click="$emit('update:startDate', '2026-07-11'); $emit('update:endDate', '2026-07-13'); $emit('change', { startDate: '2026-07-11', endDate: '2026-07-13', preset: 'last48Hours' })"
              />
            `
          },
          Select: true,
          UserSpendingRankingChart: true,
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          Line: true
        }
      }
    })

    await flushPromises()
    expect(wrapper.get('[data-testid="last-48-hours"]').attributes('data-presets')).toBe(
      'yesterday,today,last24Hours,last48Hours,7days,14days,30days'
    )

    await wrapper.get('[data-testid="last-48-hours"]').trigger('click')
    await flushPromises()

    expect(getSnapshotV2).toHaveBeenLastCalledWith(
      expect.objectContaining({
        start_date: '2026-07-11',
        end_date: '2026-07-13',
        period: '48h',
        granularity: '2h'
      })
    )
    expect(getUserUsageTrend).toHaveBeenLastCalledWith(
      expect.objectContaining({ granularity: '2h' })
    )
  })
})
