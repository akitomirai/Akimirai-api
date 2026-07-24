import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UsageView from '../UsageView.vue'

const { list, getStats, getSnapshotV2, getById, getModelStats, getUserBreakdown, listErrorLogs, routeQuery } = vi.hoisted(() => {
  vi.stubGlobal('localStorage', {
    getItem: vi.fn(() => null),
    setItem: vi.fn(),
    removeItem: vi.fn(),
  })

  return {
    list: vi.fn(),
    getStats: vi.fn(),
    getSnapshotV2: vi.fn(),
    getById: vi.fn(),
    getModelStats: vi.fn(),
    getUserBreakdown: vi.fn(),
    listErrorLogs: vi.fn(),
    routeQuery: {} as Record<string, string>,
  }
})

const messages: Record<string, string> = {
  'admin.dashboard.timeRange': 'Time Range',
  'admin.dashboard.day': 'Day',
  'admin.dashboard.hour': 'Hour',
  'admin.usage.failedToLoadUser': 'Failed to load user',
  'admin.users.columnSettings': 'Column settings',
  'admin.usage.diagnostics.timings': 'Stage Timings',
  'admin.usage.diagnostics.requestStartedAt': 'Request started',
}

vi.mock('@/api/admin', () => ({
  adminAPI: {
    usage: {
      list,
      getStats,
    },
    dashboard: {
      getSnapshotV2,
      getModelStats,
      getUserBreakdown,
    },
    users: {
      getById,
    },
  },
}))

vi.mock('@/api/admin/usage', () => ({
  adminUsageAPI: {
    list: vi.fn(),
  },
}))

vi.mock('@/api/admin/ops', () => ({
  listErrorLogs,
}))

vi.mock('@/components/admin/usage/UsageDiagnosticsDrawer.vue', () => ({
  default: { props: ['show', 'usageId'], emits: ['update:show', 'openErrors'], template: '<div />' },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showWarning: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn(),
  }),
}))

vi.mock('@/utils/format', () => ({
  formatReasoningEffort: (value: string | null | undefined) => value ?? '-',
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

vi.mock('vue-router', () => ({
  useRoute: () => ({
    query: routeQuery
  })
}))

beforeEach(() => {
  Object.keys(routeQuery).forEach((key) => delete routeQuery[key])
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const UsageFiltersStub = { template: '<div><slot name="before-refresh" /></div>' }
const UsageTableStub = {
  props: {
    tokenBreakdown: Boolean,
    columns: Array,
  },
  emits: ['userClick'],
  template: '<div data-test="usage-table" :data-token-breakdown="String(tokenBreakdown)" :data-columns="columns.map((column) => column.key).join(\'|\')"><button class="user-click" @click="$emit(\'userClick\', 2)">user</button></div>',
}
const UserTokenRankingStub = {
  emits: ['select-user'],
  template: '<div data-test="ranking"><button class="pick-user" @click="$emit(\'select-user\', 5, \'rank@test.com\')">pick</button></div>',
}
const RequestDiagnosticsTableStub = {
  props: ['columns'],
  template: '<div data-test="request-logs">{{ columns.map((column) => column.key).join("|") }}</div>',
}
const ModelDistributionChartStub = {
  props: ['metric'],
  emits: ['update:metric'],
  template: `
    <div data-test="model-chart">
      <span class="metric">{{ metric }}</span>
      <button class="switch-metric" @click="$emit('update:metric', 'actual_cost')">switch</button>
    </div>
  `,
}
const EntityDistributionChartStub = {
  props: ['metric', 'items', 'visualization'],
  emits: ['update:metric'],
  template: `
    <div data-test="user-chart">
      <span class="metric">{{ metric }}</span>
      <span class="items">{{ items.length }}</span>
      <span class="visualization">{{ visualization }}</span>
      <button class="switch-metric" @click="$emit('update:metric', 'actual_cost')">switch</button>
    </div>
  `,
}

describe('admin UsageView distribution metric toggles', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    list.mockReset()
    getStats.mockReset()
    getSnapshotV2.mockReset()
    getById.mockReset()
    getModelStats.mockReset()
    getUserBreakdown.mockReset()

    list.mockResolvedValue({
      items: [],
      total: 0,
      pages: 0,
    })
    getStats.mockResolvedValue({
      total_requests: 0,
      total_input_tokens: 0,
      total_output_tokens: 0,
      total_cache_tokens: 0,
      total_tokens: 0,
      total_cost: 0,
      total_actual_cost: 0,
      average_duration_ms: 0,
    })
    getSnapshotV2.mockResolvedValue({
      trend: [],
      models: [],
      groups: [],
    })
    getModelStats.mockResolvedValue({ models: [] })
    getUserBreakdown.mockResolvedValue({ users: [] })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('keeps previous model stats visible during refresh until new data arrives', async () => {
    // 首次加载返回 A
    getModelStats.mockResolvedValueOnce({ models: [{ model: 'A', total_tokens: 10 }] })

    const wrapper = mount(UsageView, {
      global: { stubs: {
        AppLayout: AppLayoutStub, UsageStatsCards: true, UsageFilters: UsageFiltersStub,
        UsageTable: true, UsageExportProgress: true, UsageCleanupDialog: true,
        UserBalanceHistoryModal: true, AuditLogModal: true, Pagination: true, Select: true,
        DateRangePicker: true, Icon: true,
        EntityDistributionChart: EntityDistributionChartStub,
        ModelDistributionChart: ModelDistributionChartStub, UserTokenRanking: true,
      } },
    })
    vi.advanceTimersByTime(120)
    await flushPromises()
    expect((wrapper.vm as any).requestedModelStats).toEqual([{ model: 'A', total_tokens: 10 }])

    // 刷新:让第二次 getModelStats 处于 pending,断言旧数据 A 仍在(不被清空成 [])
    let resolveSecond: (v: any) => void = () => {}
    getModelStats.mockReturnValueOnce(new Promise((res) => { resolveSecond = res }))
    ;(wrapper.vm as any).refreshData()
    await flushPromises()
    expect((wrapper.vm as any).requestedModelStats).toEqual([{ model: 'A', total_tokens: 10 }])

    // 新数据到达后替换为 B
    resolveSecond({ models: [{ model: 'B', total_tokens: 20 }] })
    await flushPromises()
    expect((wrapper.vm as any).requestedModelStats).toEqual([{ model: 'B', total_tokens: 20 }])
  })

  it('renders only user and model distributions without requesting retired snapshot charts', async () => {
    getUserBreakdown.mockResolvedValueOnce({
      users: [{
        user_id: 7,
        email: 'user@test.com',
        requests: 3,
        input_tokens: 10,
        output_tokens: 20,
        cache_tokens: 0,
        total_tokens: 30,
        cost: 0.3,
        actual_cost: 0.2,
        account_cost: 0.1,
      }],
    })
    const wrapper = mount(UsageView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          UsageStatsCards: true,
          UsageFilters: UsageFiltersStub,
          UsageTable: true,
          UsageExportProgress: true,
          UsageCleanupDialog: true,
          UserBalanceHistoryModal: true,
          Pagination: true,
          Select: true,
          DateRangePicker: true,
          Icon: true,
          EntityDistributionChart: EntityDistributionChartStub,
          ModelDistributionChart: ModelDistributionChartStub,
          UserTokenRanking: true,
        },
      },
    })

    vi.advanceTimersByTime(120)
    await flushPromises()

    expect(getSnapshotV2).not.toHaveBeenCalled()
    expect(getUserBreakdown).toHaveBeenCalledWith(expect.objectContaining({
      period: '24h',
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      limit: 200,
      sort_by: 'total_tokens',
    }))

    const modelChart = wrapper.find('[data-test="model-chart"]')
    const userChart = wrapper.find('[data-test="user-chart"]')

    expect(modelChart.find('.metric').text()).toBe('tokens')
    expect(userChart.find('.metric').text()).toBe('tokens')
    expect(userChart.find('.items').text()).toBe('1')
    expect(userChart.find('.visualization').text()).toBe('horizontal-bar')

    await modelChart.find('.switch-metric').trigger('click')
    await flushPromises()

    expect(modelChart.find('.metric').text()).toBe('actual_cost')
    expect(userChart.find('.metric').text()).toBe('tokens')
    expect(getUserBreakdown).toHaveBeenCalledTimes(1)

    await userChart.find('.switch-metric').trigger('click')
    await flushPromises()

    expect(modelChart.find('.metric').text()).toBe('actual_cost')
    expect(userChart.find('.metric').text()).toBe('actual_cost')
    expect(getUserBreakdown).toHaveBeenCalledTimes(1)
  })

  it('preserves a dashboard rolling period when opened from ranking', async () => {
    Object.assign(routeQuery, {
      user_id: '7',
      period: '48h',
      timezone: 'America/New_York',
    })

    mount(UsageView, {
      global: { stubs: {
        AppLayout: AppLayoutStub, UsageStatsCards: true, UsageFilters: UsageFiltersStub,
        UsageTable: true, UsageExportProgress: true, UsageCleanupDialog: true,
        UserBalanceHistoryModal: true, Pagination: true, Select: true,
        DateRangePicker: true, Icon: true, EntityDistributionChart: true,
        ModelDistributionChart: ModelDistributionChartStub, UserTokenRanking: true,
      } },
    })
    vi.advanceTimersByTime(120)
    await flushPromises()

    expect(list).toHaveBeenCalledWith(expect.objectContaining({
      user_id: 7,
      period: '48h',
      timezone: 'America/New_York',
    }), expect.anything())
    expect(list.mock.calls[0][0]).not.toHaveProperty('start_date')
    expect(getStats).toHaveBeenCalledWith(expect.objectContaining({
      period: '48h',
      timezone: 'America/New_York',
    }))
  })
})

describe('admin UsageView handleUserClick', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.mocked(localStorage.getItem).mockReset().mockReturnValue(null)
    vi.mocked(localStorage.setItem).mockClear()
    list.mockReset()
    getStats.mockReset()
    getSnapshotV2.mockReset()
    getById.mockReset()
    getUserBreakdown.mockReset()

    list.mockResolvedValue({ items: [], total: 0, pages: 0 })
    getStats.mockResolvedValue({
      total_requests: 0, total_input_tokens: 0, total_output_tokens: 0,
      total_cache_tokens: 0, total_tokens: 0, total_cost: 0, total_actual_cost: 0, average_duration_ms: 0,
    })
    getSnapshotV2.mockResolvedValue({ trend: [], models: [], groups: [] })
    getUserBreakdown.mockResolvedValue({ users: [] })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  const mountUsageTableView = () => mount(UsageView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        UsageStatsCards: true,
        UsageFilters: UsageFiltersStub,
        UsageTable: UsageTableStub,
        UsageExportProgress: true,
        UsageCleanupDialog: true,
        UserBalanceHistoryModal: true,
        AuditLogModal: true,
        Pagination: true,
        Select: true,
        DateRangePicker: true,
        Icon: true,
        TokenUsageTrend: true,
        ModelDistributionChart: true,
        GroupDistributionChart: true,
        EndpointDistributionChart: true,
        UserTokenRanking: true,
      },
    },
  })

  it('opens user via include_deleted when clicking a usage row user', async () => {
    getById.mockResolvedValue({ id: 2, email: 'd@test.com', deleted_at: '2026-05-28T00:00:00Z' })

    const wrapper = mountUsageTableView()

    vi.advanceTimersByTime(120)
    await flushPromises()

    const usageTable = wrapper.get('[data-test="usage-table"]')
    expect(usageTable.attributes('data-token-breakdown')).toBe('true')
    expect(usageTable.attributes('data-columns')).toContain('tokens|cache_hit_rate|cost')

    await wrapper.find('[data-test="usage-table"] .user-click').trigger('click')
    await flushPromises()

    expect(getById).toHaveBeenCalledWith(2, true)
  })

  it('keeps the cache-hit column visible with existing hidden-column preferences', async () => {
    vi.mocked(localStorage.getItem).mockImplementation((key) =>
      key === 'usage-hidden-columns' ? JSON.stringify(['api_key']) : null
    )

    const wrapper = mountUsageTableView()
    vi.advanceTimersByTime(120)
    await flushPromises()

    const columns = wrapper.get('[data-test="usage-table"]').attributes('data-columns')
    expect(columns).not.toContain('api_key')
    expect(columns).toContain('tokens|cache_hit_rate|cost')
  })
})

describe('admin UsageView errors tab filter forwarding', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    list.mockReset()
    getStats.mockReset()
    getSnapshotV2.mockReset()
    getModelStats.mockReset()
    getUserBreakdown.mockReset()
    listErrorLogs.mockReset()

    list.mockResolvedValue({ items: [], total: 0, pages: 0 })
    getStats.mockResolvedValue({
      total_requests: 0, total_input_tokens: 0, total_output_tokens: 0,
      total_cache_tokens: 0, total_tokens: 0, total_cost: 0, total_actual_cost: 0, average_duration_ms: 0,
    })
    getSnapshotV2.mockResolvedValue({ trend: [], models: [], groups: [] })
    getModelStats.mockResolvedValue({ models: [] })
    getUserBreakdown.mockResolvedValue({ users: [] })
    listErrorLogs.mockResolvedValue({ items: [], total: 0, pages: 0 })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('forwards model/account_id/group_id to listErrorLogs on the errors tab', async () => {
    const wrapper = mount(UsageView, {
      global: { stubs: {
        AppLayout: AppLayoutStub, UsageStatsCards: true, UsageFilters: UsageFiltersStub,
        UsageTable: true, UsageExportProgress: true, UsageCleanupDialog: true,
        UserBalanceHistoryModal: true, AuditLogModal: true, Pagination: true, Select: true,
        DateRangePicker: true, Icon: true, TokenUsageTrend: true,
        ModelDistributionChart: true, GroupDistributionChart: true, EndpointDistributionChart: true,
        UserTokenRanking: true, OpsErrorLogTable: true, OpsErrorDetailModal: true,
      } },
    })
    vi.advanceTimersByTime(120)
    await flushPromises()

    // 模拟用户在过滤器里选择了模型/账户/分组
    const vm = wrapper.vm as any
    vm.filters.model = 'gpt-5.3-codex'
    vm.filters.account_id = 7
    vm.filters.group_id = 3
    await flushPromises()

    // 切换到「错误请求」标签（第二个 tab 按钮）触发 loadAdminErrors
    const tabs = wrapper.findAll('[data-testid="usage-detail-tab"]')
    await tabs[1].trigger('click')
    await flushPromises()

    expect(listErrorLogs).toHaveBeenCalledWith(expect.objectContaining({
      view: 'all',
      model: 'gpt-5.3-codex',
      account_id: 7,
      group_id: 3,
    }))
  })
})

describe('admin UsageView ranking tab', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.mocked(localStorage.getItem).mockReset().mockReturnValue(null)
    vi.mocked(localStorage.setItem).mockClear()
    list.mockReset()
    getStats.mockReset()
    getSnapshotV2.mockReset()
    getModelStats.mockReset()
    getUserBreakdown.mockReset()

    list.mockResolvedValue({ items: [], total: 0, pages: 0 })
    getStats.mockResolvedValue({
      total_requests: 0, total_input_tokens: 0, total_output_tokens: 0,
      total_cache_tokens: 0, total_tokens: 0, total_cost: 0, total_actual_cost: 0, average_duration_ms: 0,
    })
    getSnapshotV2.mockResolvedValue({ trend: [], models: [], groups: [] })
    getModelStats.mockResolvedValue({ models: [] })
    getUserBreakdown.mockResolvedValue({ users: [] })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('mounts ranking lazily and drill-down sets user filter then jumps back to usage tab', async () => {
    const wrapper = mount(UsageView, {
      global: { stubs: {
        AppLayout: AppLayoutStub, UsageStatsCards: true, UsageFilters: UsageFiltersStub,
        UsageTable: true, UsageExportProgress: true, UsageCleanupDialog: true,
        UserBalanceHistoryModal: true, Pagination: true, Select: true,
        DateRangePicker: true, Icon: true, TokenUsageTrend: true,
        ModelDistributionChart: true, GroupDistributionChart: true, EndpointDistributionChart: true,
        UserTokenRanking: UserTokenRankingStub, RequestDiagnosticsTable: RequestDiagnosticsTableStub,
        OpsErrorLogTable: true, OpsErrorDetailModal: true,
      } },
    })
    vi.advanceTimersByTime(120)
    await flushPromises()

    // 懒挂载:切到排行 tab 前不渲染
    expect(wrapper.find('[data-test="ranking"]').exists()).toBe(false)

    const tabs = wrapper.findAll('[data-testid="usage-detail-tab"]')
    expect(tabs).toHaveLength(4)
    await tabs[2].trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-test="ranking"]').exists()).toBe(true)

    // 下钻:设置 user_id、切回用量明细 tab 并按新筛选重新拉取列表
    list.mockClear()
    await wrapper.find('[data-test="ranking"] .pick-user').trigger('click')
    await flushPromises()

    expect((wrapper.vm as any).activeTab).toBe('usage')
    expect((wrapper.vm as any).filters.user_id).toBe(5)
    expect(list).toHaveBeenCalledWith(expect.objectContaining({ user_id: 5 }), expect.anything())
  })

  it('shows request logs inline and lets admins choose visible columns', async () => {
    const wrapper = mount(UsageView, {
      global: { stubs: {
        AppLayout: AppLayoutStub, UsageStatsCards: true, UsageFilters: UsageFiltersStub,
        UsageTable: true, UsageExportProgress: true, UsageCleanupDialog: true,
        UserBalanceHistoryModal: true, Pagination: true, Select: true,
        DateRangePicker: true, Icon: true, EntityDistributionChart: true,
        ModelDistributionChart: true, UserTokenRanking: UserTokenRankingStub,
        RequestDiagnosticsTable: RequestDiagnosticsTableStub,
        OpsErrorLogTable: true, OpsErrorDetailModal: true,
      } },
    })
    vi.advanceTimersByTime(120)
    await flushPromises()

    const tabs = wrapper.findAll('[data-testid="usage-detail-tab"]')
    await tabs[3].trigger('click')
    await flushPromises()

    expect((wrapper.vm as any).activeTab).toBe('diagnostics')
    expect(wrapper.find('[data-test="request-logs"]').isVisible()).toBe(true)
    expect(wrapper.get('[data-test="request-logs"]').text()).toContain('timings')

    const columnSettings = wrapper.get('button[title="Column settings"]')
    await columnSettings.trigger('click')
    const timingToggle = wrapper.findAll('button').find(button => button.text() === 'Stage Timings')
    expect(timingToggle).toBeDefined()
    await timingToggle!.trigger('click')

    expect(wrapper.get('[data-test="request-logs"]').text()).not.toContain('timings')
    expect(localStorage.setItem).toHaveBeenCalledWith(
      'usage-diagnostics-hidden-columns',
      JSON.stringify(['timings']),
    )
  })

  it('restores request-log column preferences while keeping required columns visible', async () => {
    vi.mocked(localStorage.getItem).mockImplementation((key) =>
      key === 'usage-diagnostics-hidden-columns' ? JSON.stringify(['timings', 'user']) : null
    )
    const wrapper = mount(UsageView, {
      global: { stubs: {
        AppLayout: AppLayoutStub, UsageStatsCards: true, UsageFilters: UsageFiltersStub,
        UsageTable: true, UsageExportProgress: true, UsageCleanupDialog: true,
        UserBalanceHistoryModal: true, Pagination: true, Select: true,
        DateRangePicker: true, Icon: true, EntityDistributionChart: true,
        ModelDistributionChart: true, UserTokenRanking: UserTokenRankingStub,
        RequestDiagnosticsTable: RequestDiagnosticsTableStub,
        OpsErrorLogTable: true, OpsErrorDetailModal: true,
      } },
    })
    vi.advanceTimersByTime(120)
    await flushPromises()

    await wrapper.findAll('[data-testid="usage-detail-tab"]')[3].trigger('click')
    await flushPromises()

    const columns = wrapper.get('[data-test="request-logs"]').text()
    expect(columns).toContain('user')
    expect(columns).toContain('created_at')
    expect(columns).not.toContain('timings')
  })
})
