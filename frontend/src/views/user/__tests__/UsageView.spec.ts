import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UsageView from '../UsageView.vue'

const {
  listRequestLogs,
  getStats,
  getDashboardModels,
  getDashboardSnapshotV2,
  listKeys,
  getAvailableGroups,
  publicSettings,
  showError,
  showWarning,
  showSuccess,
  showInfo,
} = vi.hoisted(() => ({
  listRequestLogs: vi.fn(),
  getStats: vi.fn(),
  getDashboardModels: vi.fn(),
  getDashboardSnapshotV2: vi.fn(),
  listKeys: vi.fn(),
  getAvailableGroups: vi.fn(),
  publicSettings: { allow_user_view_error_requests: true },
  showError: vi.fn(),
  showWarning: vi.fn(),
  showSuccess: vi.fn(),
  showInfo: vi.fn(),
}))

const messages: Record<string, string> = {
  'usage.logs.typeFilter': 'Type',
  'usage.logs.kinds.all': 'All types',
  'usage.logs.kinds.consumption': 'Consumption',
  'usage.logs.kinds.error': 'Error',
  'usage.apiKeyFilter': 'API Key',
  'usage.apiKeyDistribution': 'API Key Distribution',
  'usage.apiKeyFallback': 'API key',
  'usage.allApiKeys': 'All API Keys',
  'usage.model': 'Model',
  'usage.exporting': 'Exporting',
  'usage.exportCsv': 'Export CSV',
  'usage.noDataToExport': 'No data',
  'usage.preparingExport': 'Preparing export',
  'usage.exportSuccess': 'Export success',
  'usage.exportFailed': 'Export failed',
  'admin.usage.group': 'Group',
  'admin.usage.allGroups': 'All groups',
  'admin.usage.allModels': 'All models',
  'admin.dashboard.timeRange': 'Time range',
  'common.refresh': 'Refresh',
  'common.reset': 'Reset',
}

vi.mock('@/api', () => ({
  usageAPI: {
    listRequestLogs,
    getStats,
    getDashboardModels,
    getDashboardSnapshotV2,
  },
  keysAPI: { list: listKeys },
  userGroupsAPI: { getAvailable: getAvailableGroups },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: publicSettings,
    showError,
    showWarning,
    showSuccess,
    showInfo,
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => messages[key] ?? key }),
  }
})

const requestLog = {
  id: 1,
  kind: 'consumption' as const,
  created_at: '2026-07-13T02:13:32Z',
  request_id: 'req-user-1',
  api_key_id: 1,
  api_key_name: 'demo-key',
  api_key_deleted: false,
  group_id: 1,
  group_name: 'Pro pool',
  rate_multiplier: 1,
  model: 'gpt-5.6-sol',
  reasoning_effort: 'max',
  first_token_ms: 3000,
  duration_ms: 4000,
  total_tokens: 17_248,
  actual_cost: 0.011508,
  status_code: null,
  error_code: null,
  error_message: null,
}

const AppLayoutStub = { template: '<div><slot /></div>' }
const UsageStatsCardsStub = { template: '<div data-test="usage-stats" />' }
const DateRangePickerStub = {
  name: 'DateRangePicker',
  props: ['presetValues'],
  template: '<div data-test="date-range" />',
}
const SelectStub = {
  props: ['options'],
  template: '<div data-test="select"><span v-for="option in options" :key="String(option.value)">{{ option.label }}</span></div>',
}
const UserRequestLogTableStub = {
  name: 'UserRequestLogTable',
  props: ['rows', 'loading'],
  emits: ['sort'],
  template: '<div data-test="request-log-table">{{ rows.map((row) => row.request_id).join(",") }}</div>',
}
const EntityDistributionChartStub = {
  props: ['title', 'entityLabel', 'items'],
  template: '<div data-test="api-key-distribution-chart">{{ title }} {{ items.map((item) => item.label).join(",") }}</div>',
}

const mountUsageView = () => mount(UsageView, {
  global: {
    stubs: {
      AppLayout: AppLayoutStub,
      UsageStatsCards: UsageStatsCardsStub,
      DateRangePicker: DateRangePickerStub,
      Select: SelectStub,
      Pagination: true,
      UserRequestLogTable: UserRequestLogTableStub,
      EntityDistributionChart: EntityDistributionChartStub,
    },
  },
})

describe('user UsageView unified logs', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    publicSettings.allow_user_view_error_requests = true
    listRequestLogs.mockResolvedValue({ items: [requestLog], total: 1, page: 1, page_size: 20 })
    getStats.mockResolvedValue({
      total_requests: 1,
      total_input_tokens: 10,
      total_output_tokens: 20,
      total_cache_tokens: 0,
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
    getDashboardSnapshotV2.mockResolvedValue({
      generated_at: '2026-07-13T00:00:00Z',
      start_date: '2026-07-13',
      end_date: '2026-07-13',
      granularity: 'hour',
      api_keys: [{ api_key_id: 1, api_key_name: 'demo-key', requests: 1, total_tokens: 30, cost: 0.1, actual_cost: 0.08 }],
    })
    listKeys.mockResolvedValue({ items: [{ id: 1, name: 'demo-key' }] })
    getAvailableGroups.mockResolvedValue([{ id: 1, name: 'Pro pool' }])
  })

  it('loads one server-paginated timeline with the requested filters', async () => {
    const wrapper = mountUsageView()
    await flushPromises()

    expect(listRequestLogs).toHaveBeenCalledWith(expect.objectContaining({
      kind: 'all',
      page: 1,
      period: '24h',
      sort_by: 'created_at',
      sort_order: 'desc',
    }), expect.objectContaining({ signal: expect.any(AbortSignal) }))
    expect(listRequestLogs.mock.calls[0][0]).not.toHaveProperty('start_date')
    expect(listRequestLogs.mock.calls[0][0]).not.toHaveProperty('end_date')
    expect(wrapper.get('[data-test="request-log-table"]').text()).toContain('req-user-1')
    expect(wrapper.get('[data-test="api-key-distribution-chart"]').text()).toContain('demo-key')
    expect(wrapper.get('[data-test="api-key-distribution-chart"]').element.parentElement?.className)
      .not.toContain('lg:grid-cols-2')
    expect(wrapper.findComponent({ name: 'DateRangePicker' }).props('presetValues')).toEqual([
      'yesterday', 'today', 'last24Hours', 'last48Hours', '7days', '14days', '30days',
    ])
    expect(wrapper.findAll('label.input-label').map((label) => label.text())).toEqual([
      'Type', 'API Key', 'Model', 'Group',
    ])
    expect(wrapper.text()).not.toContain('Error Requests')
    expect(getStats).toHaveBeenCalledWith(expect.objectContaining({ period: '24h' }))
    expect(getDashboardModels).toHaveBeenCalledWith(expect.objectContaining({ period: '24h' }))
    expect(getDashboardSnapshotV2).toHaveBeenCalledWith(expect.objectContaining({ period: '24h' }))
  })

  it('uses dashboard period and granularity semantics for presets and reset', async () => {
    const wrapper = mountUsageView()
    await flushPromises()
    vi.clearAllMocks()

    wrapper.findComponent({ name: 'DateRangePicker' }).vm.$emit('change', {
      startDate: '2026-07-11',
      endDate: '2026-07-13',
      preset: 'last48Hours',
    })
    await flushPromises()

    expect(listRequestLogs).toHaveBeenCalledWith(expect.objectContaining({ period: '48h' }), expect.anything())
    expect(listRequestLogs.mock.calls[0][0]).not.toHaveProperty('start_date')
    expect(listRequestLogs.mock.calls[0][0]).not.toHaveProperty('end_date')
    expect(getStats).toHaveBeenCalledWith(expect.objectContaining({ period: '48h' }))
    expect(getDashboardModels).toHaveBeenCalledWith(expect.objectContaining({ period: '48h' }))
    expect(getDashboardSnapshotV2).toHaveBeenCalledWith(expect.objectContaining({
      period: '48h',
      granularity: '2h',
    }))

    vi.clearAllMocks()
    wrapper.findComponent({ name: 'DateRangePicker' }).vm.$emit('change', {
      startDate: '2026-07-01',
      endDate: '2026-07-05',
      preset: null,
    })
    await flushPromises()
    expect(listRequestLogs).toHaveBeenCalledWith(expect.objectContaining({
      start_date: '2026-07-01',
      end_date: '2026-07-05',
    }), expect.anything())
    expect(listRequestLogs.mock.calls[0][0]).not.toHaveProperty('period')

    vi.clearAllMocks()
    ;(wrapper.vm as any).resetFilters()
    await flushPromises()
    expect(listRequestLogs).toHaveBeenCalledWith(expect.objectContaining({ period: '24h' }), expect.anything())
  })

  it('requests consumption only when user error visibility is disabled', async () => {
    publicSettings.allow_user_view_error_requests = false
    const wrapper = mountUsageView()
    await flushPromises()

    expect(listRequestLogs).toHaveBeenCalledWith(expect.objectContaining({ kind: 'consumption' }), expect.anything())
    expect(wrapper.text()).not.toContain('Error')
    expect(wrapper.text()).toContain('Consumption')
  })

  it('forwards table sorting to the unified endpoint', async () => {
    const wrapper = mountUsageView()
    await flushPromises()
    listRequestLogs.mockClear()

    wrapper.findComponent({ name: 'UserRequestLogTable' }).vm.$emit('sort', 'duration_ms', 'asc')
    await flushPromises()

    expect(listRequestLogs).toHaveBeenCalledWith(expect.objectContaining({
      sort_by: 'duration_ms',
      sort_order: 'asc',
    }), expect.anything())
  })

  it('exports the unified consumption and error fields', async () => {
    const wrapper = mountUsageView()
    await flushPromises()

    let csvContent = ''
    const OriginalBlob = globalThis.Blob
    vi.stubGlobal('Blob', vi.fn((parts: BlobPart[], options?: BlobPropertyBag) => {
      csvContent = parts.map((part) => String(part)).join('')
      return new OriginalBlob(parts, options)
    }))
    const originalCreateObjectURL = window.URL.createObjectURL
    const originalRevokeObjectURL = window.URL.revokeObjectURL
    window.URL.createObjectURL = vi.fn(() => 'blob:logs-export') as typeof window.URL.createObjectURL
    window.URL.revokeObjectURL = vi.fn(() => {}) as typeof window.URL.revokeObjectURL
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    await (wrapper.vm as any).exportToCSV()

    expect(csvContent).toContain('Time,Request ID,Type,API Key,Group,Rate Multiplier,Model,Reasoning Effort')
    expect(csvContent).toContain('req-user-1')
    expect(csvContent).not.toContain('Upstream Endpoint')
    expect(clickSpy).toHaveBeenCalled()
    expect(showSuccess).toHaveBeenCalled()

    window.URL.createObjectURL = originalCreateObjectURL
    window.URL.revokeObjectURL = originalRevokeObjectURL
    vi.unstubAllGlobals()
    clickSpy.mockRestore()
  })
})
