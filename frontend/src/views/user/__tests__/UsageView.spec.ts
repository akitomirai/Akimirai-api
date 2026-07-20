import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UsageView from '../UsageView.vue'
import { resetTokenCountModeForTests } from '@/composables/useTokenCountMode'

const {
  getStats,
  getDashboardModels,
  listKeys,
  getAvailableGroups,
  publicSettings,
} = vi.hoisted(() => ({
  getStats: vi.fn(),
  getDashboardModels: vi.fn(),
  listKeys: vi.fn(),
  getAvailableGroups: vi.fn(),
  publicSettings: { allow_user_view_error_requests: true },
}))

const messages: Record<string, string> = {
  'admin.dashboard.timeRange': 'Time range',
  'usage.countMode.modern': 'w/unit',
  'usage.countMode.legacy': 'k/unit',
  'usage.countMode.label': 'Token count mode',
}

vi.mock('@/api', () => ({
  usageAPI: { getStats, getDashboardModels },
  keysAPI: { list: listKeys },
  userGroupsAPI: { getAvailable: getAvailableGroups },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ cachedPublicSettings: publicSettings }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => messages[key] ?? key }),
  }
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const UsageStatsCardsStub = {
  name: 'UsageStatsCards',
  props: ['showTokenBreakdown'],
  template: '<div data-test="usage-stats" :data-show-token-breakdown="String(showTokenBreakdown)" />',
}
const DateRangePickerStub = {
  name: 'DateRangePicker',
  props: { presetValues: Array, showTimeInputs: Boolean },
  emits: ['change'],
  template: '<div data-test="date-range" />',
}
const LegacyUsageLogPanelStub = {
  name: 'LegacyUsageLogPanel',
  props: ['period', 'errorViewEnabled', 'startTime', 'endTime'],
  template: '<div data-test="legacy-usage-panel" :data-period="period" :data-errors="String(errorViewEnabled)" />',
}

const mountUsageView = () => mount(UsageView, {
  global: {
    stubs: {
      AppLayout: AppLayoutStub,
      UsageStatsCards: UsageStatsCardsStub,
      DateRangePicker: DateRangePickerStub,
      LegacyUsageLogPanel: LegacyUsageLogPanelStub,
    },
  },
})

describe('user UsageView unified logs', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    resetTokenCountModeForTests()
    publicSettings.allow_user_view_error_requests = true
    getStats.mockResolvedValue({
      total_requests: 1,
      total_input_tokens: 10,
      total_output_tokens: 20,
      total_cache_tokens: 0,
      total_cache_read_tokens: 0,
      total_tokens: 30,
      total_cost: 0.1,
      total_actual_cost: 0.08,
      average_duration_ms: 12,
      endpoints: [],
      upstream_endpoints: [],
      endpoint_paths: [],
    })
    getDashboardModels.mockResolvedValue({
      models: [{ model: 'gpt-5.6-sol', requests: 1, total_tokens: 30 }],
      start_date: '2026-07-13',
      end_date: '2026-07-13',
    })
    listKeys.mockResolvedValue({ items: [{ id: 1, name: 'demo-key' }] })
    getAvailableGroups.mockResolvedValue([{ id: 1, name: 'Pro pool' }])
  })

  it('uses the same legacy detail table for both count modes', async () => {
    const wrapper = mountUsageView()
    await flushPromises()

    expect(wrapper.get('[data-test="legacy-usage-panel"]').attributes()).toMatchObject({
      'data-period': '24h',
      'data-errors': 'true',
    })
    expect(wrapper.get('[data-test="usage-stats"]').attributes('data-show-token-breakdown')).toBe('true')
    expect(wrapper.findComponent({ name: 'DateRangePicker' }).props('showTimeInputs')).toBe(true)

    const legacyButton = wrapper.findAll('button').find((button) => button.text() === 'k/unit')
    await legacyButton!.trigger('click')
    await flushPromises()

    expect(localStorage.getItem('token-count-mode')).toBe('legacy')
    expect(wrapper.get('[data-test="legacy-usage-panel"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="request-log-table"]').exists()).toBe(false)
  })

  it('does not call the removed unified request-log projection', async () => {
    const wrapper = mountUsageView()
    await flushPromises()

    expect(getStats).toHaveBeenCalledWith(expect.objectContaining({ period: '24h' }))
    expect(getDashboardModels).toHaveBeenCalledWith(expect.objectContaining({ period: '24h' }))
    expect(wrapper.find('[data-test="request-log-table"]').exists()).toBe(false)
  })

  it('uses period semantics for presets and minute precision for custom ranges', async () => {
    const wrapper = mountUsageView()
    await flushPromises()
    vi.clearAllMocks()

    wrapper.findComponent({ name: 'DateRangePicker' }).vm.$emit('change', {
      startDate: '2026-07-11',
      endDate: '2026-07-13',
      startTime: '22:00',
      endTime: '22:00',
      preset: 'last48Hours',
    })
    await flushPromises()

    expect(getStats).toHaveBeenCalledWith(expect.objectContaining({ period: '48h' }))
    expect(getDashboardModels).toHaveBeenCalledWith(expect.objectContaining({ period: '48h' }))

    vi.clearAllMocks()
    wrapper.findComponent({ name: 'DateRangePicker' }).vm.$emit('change', {
      startDate: '2026-07-01',
      endDate: '2026-07-05',
      startTime: '08:30',
      endTime: '19:45',
      preset: null,
    })
    await flushPromises()

    expect(getStats).toHaveBeenCalledWith(expect.objectContaining({
      start_time: '2026-07-01T08:30',
      end_time: '2026-07-05T19:45',
    }))
    expect(getStats.mock.calls[0][0]).not.toHaveProperty('period')
  })
})
