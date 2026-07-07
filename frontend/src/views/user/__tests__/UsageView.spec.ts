import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UsageView from '../UsageView.vue'

const {
  query,
  getStats,
  getDashboardModels,
  getDashboardSnapshotV2,
  listMyErrorRequests,
  list,
  getAvailable,
  showError,
  showWarning,
  showSuccess,
  showInfo,
} = vi.hoisted(() => ({
  query: vi.fn(),
  getStats: vi.fn(),
  getDashboardModels: vi.fn(),
  getDashboardSnapshotV2: vi.fn(),
  listMyErrorRequests: vi.fn(),
  list: vi.fn(),
  getAvailable: vi.fn(),
  showError: vi.fn(),
  showWarning: vi.fn(),
  showSuccess: vi.fn(),
  showInfo: vi.fn(),
}))

const messages: Record<string, string> = {
  'usage.costDetails': 'Cost Breakdown',
  'admin.usage.inputCost': 'Input Cost',
  'admin.usage.outputCost': 'Output Cost',
  'admin.usage.cacheCreationCost': 'Cache Creation Cost',
  'admin.usage.cacheReadCost': 'Cache Read Cost',
  'usage.inputTokenPrice': 'Input price',
  'usage.outputTokenPrice': 'Output price',
  'usage.perMillionTokens': '/ 1M tokens',
  'usage.serviceTier': 'Service tier',
  'usage.serviceTierPriority': 'Fast',
  'usage.serviceTierFlex': 'Flex',
  'usage.serviceTierStandard': 'Standard',
  'usage.rate': 'Rate',
  'usage.original': 'Original',
  'usage.billed': 'Billed',
  'usage.allApiKeys': 'All API Keys',
  'usage.tabs.usage': 'Usage',
  'usage.tabs.errors': 'Error Requests',
  'usage.errors.disabled': 'Error request records are not enabled',
  'usage.apiKeyFilter': 'API Key',
  'usage.model': 'Model',
  'usage.reasoningEffort': 'Reasoning Effort',
  'usage.endpoint': 'Endpoint',
  'usage.type': 'Type',
  'usage.tokens': 'Tokens',
  'usage.cacheHitRate': 'Cache Hit Rate',
  'usage.cost': 'Cost',
  'usage.firstToken': 'First Token',
  'usage.duration': 'Duration',
  'usage.time': 'Time',
  'usage.userAgent': 'User Agent',
  'usage.imageUnit': ' images',
  'usage.imageCount': 'Image count',
  'usage.imageBillingSize': 'Billing size',
  'usage.imageInputSize': 'Input size',
  'usage.imageOutputSize': 'Output size',
  'usage.imageSizeSource': 'Size source',
  'usage.imageSizeBreakdown': 'Size breakdown',
  'usage.imageSizeSourceOutput': 'Upstream output',
  'usage.imageSizeSourceInput': 'Request input',
  'usage.imageSizeSourceDefault': 'Default billing tier',
  'usage.imageSizeSourceLegacy': 'Legacy record',
  'usage.imageSizeSourceMissing': 'Not recorded',
  'usage.imageSizeNotRecorded': 'not recorded',
  'usage.imageSizeLegacyUnstandardized': 'legacy unstandardized',
  'usage.imageSizeUnknown': 'unknown',
  'usage.imageUnitPrice': 'Per-image price',
  'usage.imageTotalPrice': 'Image total price',
  'admin.usage.billingModeToken': 'Token',
  'admin.usage.billingModePerRequest': 'Per request',
  'admin.usage.billingModeImage': 'Image',
  'admin.usage.allGroups': 'All groups',
  'admin.usage.allModels': 'All models',
  'usage.ws': 'WS',
  'usage.stream': 'Stream',
  'usage.sync': 'Sync',
  'usage.exporting': 'Exporting',
  'usage.exportCsv': 'Export CSV',
  'usage.failedToLoad': 'Failed to load',
  'usage.noDataToExport': 'No data',
  'usage.preparingExport': 'Preparing export',
  'usage.exportSuccess': 'Export success',
  'usage.exportFailed': 'Export failed',
  'common.refresh': 'Refresh',
  'common.reset': 'Reset',
  'admin.dashboard.timeRange': 'Time range',
  'admin.users.columnSettings': 'Columns',
}

vi.mock('@/api', () => ({
  usageAPI: {
    query,
    getStats,
    getDashboardModels,
    getDashboardSnapshotV2,
    listMyErrorRequests,
  },
  keysAPI: {
    list,
  },
  userGroupsAPI: {
    getAvailable,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: { allow_user_view_error_requests: false },
    showError,
    showWarning,
    showSuccess,
    showInfo,
  }),
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
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

const AppLayoutStub = { template: '<div><slot /></div>' }
const UsageStatsCardsStub = { template: '<div data-test="usage-stats" />' }
const UsageTableStub = {
  props: ['columns', 'data', 'dense', 'framed'],
  template: `
    <div data-test="usage-table" :data-dense="String(dense)" :data-framed="String(framed)">
      <span v-for="column in columns" :key="column.key">{{ column.label }}</span>
      <span v-for="row in data" :key="row.request_id">{{ row.request_id }}</span>
    </div>
  `,
}
const UserErrorRequestsTableStub = {
  template: '<div data-test="error-requests-table" />',
}

const usageLog = {
  id: 1,
  user_id: 1,
  api_key_id: 1,
  account_id: null,
  request_id: 'req-user-1',
  model: 'gpt-5.5',
  service_tier: 'priority',
  reasoning_effort: null,
  inbound_endpoint: '/v1/responses',
  upstream_endpoint: '/v1/responses',
  group_id: null,
  subscription_id: null,
  input_tokens: 631,
  output_tokens: 28,
  cache_creation_tokens: 0,
  cache_read_tokens: 172000,
  cache_creation_5m_tokens: 0,
  cache_creation_1h_tokens: 0,
  input_cost: 0.001,
  output_cost: 0.002,
  cache_creation_cost: 0,
  cache_read_cost: 0.001,
  total_cost: 0.016202,
  actual_cost: 0.016202,
  rate_multiplier: 1,
  billing_type: 0,
  request_type: 'stream',
  stream: true,
  openai_ws_mode: false,
  duration_ms: 5420,
  first_token_ms: 4740,
  image_count: 0,
  image_size: null,
  image_input_size: null,
  image_output_size: null,
  image_size_source: null,
  image_size_breakdown: null,
  image_output_tokens: 0,
  image_output_cost: 0,
  user_agent: 'test-agent',
  ip_address: '203.0.113.10',
  cache_ttl_overridden: false,
  billing_mode: 'token',
  created_at: '2026-07-07T11:45:33Z',
  api_key: { id: 1, name: 'demo-key' },
}

function mountUsageView() {
  return mount(UsageView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        Pagination: true,
        Select: true,
        DateRangePicker: true,
        Icon: true,
        UsageStatsCards: UsageStatsCardsStub,
        UsageTable: UsageTableStub,
        UserErrorRequestsTable: UserErrorRequestsTableStub,
      },
    },
  })
}

describe('user UsageView', () => {
  beforeEach(() => {
    query.mockReset()
    getStats.mockReset()
    getDashboardModels.mockReset()
    getDashboardSnapshotV2.mockReset()
    listMyErrorRequests.mockReset()
    list.mockReset()
    getAvailable.mockReset()
    showError.mockReset()
    showWarning.mockReset()
    showSuccess.mockReset()
    showInfo.mockReset()
    localStorage.clear()

    query.mockResolvedValue({ items: [usageLog], total: 1, pages: 1 })
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
      models: [{ model: 'gpt-5.4', requests: 1, input_tokens: 10, output_tokens: 20, cache_creation_tokens: 0, cache_read_tokens: 0, total_tokens: 30, cost: 0.1, actual_cost: 0.08 }],
      start_date: '2026-03-08',
      end_date: '2026-03-08',
    })
    getDashboardSnapshotV2.mockResolvedValue({
      generated_at: '2026-03-08T00:00:00Z',
      start_date: '2026-03-08',
      end_date: '2026-03-08',
      granularity: 'hour',
      trend: [],
      groups: [],
    })
    listMyErrorRequests.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20, pages: 0 })
    list.mockResolvedValue({ items: [{ id: 1, name: 'demo-key' }] })
    getAvailable.mockResolvedValue([{ id: 1, name: 'default' }])
  })

  it('loads compact usage data without chart requests on first render', async () => {
    const wrapper = mountUsageView()
    await flushPromises()

    expect(query).toHaveBeenCalled()
    expect(getStats).toHaveBeenCalled()
    expect(list).toHaveBeenCalledWith(1, 100)
    expect(getAvailable).not.toHaveBeenCalled()
    expect(getDashboardModels).not.toHaveBeenCalled()
    expect(getDashboardSnapshotV2).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Usage')
    expect(wrapper.text()).toContain('Error Requests')

    const table = wrapper.get('[data-test="usage-table"]')
    expect(table.attributes('data-dense')).toBe('true')
    expect(table.attributes('data-framed')).toBe('false')
    expect(table.text()).toContain('Cache Hit Rate')
    expect(table.text()).not.toContain('IP')
    expect(table.text()).not.toContain('Reasoning Effort')
    expect(table.text()).not.toContain('All groups')
  })

  it('keeps the error requests tab visible when the feature is disabled', async () => {
    const wrapper = mountUsageView()
    await flushPromises()

    const errorTab = wrapper.findAll('button').find((button) => button.text().includes('Error Requests'))
    expect(errorTab).toBeTruthy()

    await errorTab!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Error request records are not enabled')
    expect(listMyErrorRequests).not.toHaveBeenCalled()
  })

  it('exports csv with current filters and without admin-only fields', async () => {
    const wrapper = mountUsageView()
    await flushPromises()

    let exportedBlob: Blob | null = null
    let csvContent = ''
    const OriginalBlob = globalThis.Blob
    vi.stubGlobal('Blob', vi.fn((parts: BlobPart[], options?: BlobPropertyBag) => {
      csvContent = parts.map((part) => String(part)).join('')
      return new OriginalBlob(parts, options)
    }))
    const originalCreateObjectURL = window.URL.createObjectURL
    const originalRevokeObjectURL = window.URL.revokeObjectURL
    window.URL.createObjectURL = vi.fn((blob: Blob | MediaSource) => {
      exportedBlob = blob as Blob
      return 'blob:usage-export'
    }) as typeof window.URL.createObjectURL
    window.URL.revokeObjectURL = vi.fn(() => {}) as typeof window.URL.revokeObjectURL
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    await (wrapper.vm as any).exportToCSV()

    expect(exportedBlob).not.toBeNull()
    expect(query).toHaveBeenCalledWith(expect.objectContaining({
      page_size: 100,
      sort_by: 'created_at',
      sort_order: 'desc',
    }))
    expect(clickSpy).toHaveBeenCalled()
    expect(showSuccess).toHaveBeenCalled()
    expect(csvContent.startsWith('\uFEFF')).toBe(true)
    expect(csvContent.slice(1)).toBe([
      'Time,API Key Name,Model,Reasoning Effort,Inbound Endpoint,IP Address,Type,Billing Mode,Input Tokens,Output Tokens,Cache Read Tokens,Cache Creation Tokens,Rate Multiplier,Billed Cost,Original Cost,First Token (ms),Duration (ms)',
      '2026-07-07T11:45:33Z,demo-key,gpt-5.5,"\'-",/v1/responses,203.0.113.10,Stream,Token,631,28,172000,0,1,0.01620200,0.01620200,4740,5420',
    ].join('\n'))
    expect(csvContent).toContain('IP Address')
    expect(csvContent).toContain('203.0.113.10')
    expect(csvContent).toContain('Billed Cost')
    expect(csvContent).toContain('Original Cost')
    expect(csvContent).not.toContain('Upstream Endpoint')
    expect(csvContent).not.toContain('account_cost')
    expect(csvContent).not.toContain('account_rate_multiplier')

    window.URL.createObjectURL = originalCreateObjectURL
    window.URL.revokeObjectURL = originalRevokeObjectURL
    vi.unstubAllGlobals()
    clickSpy.mockRestore()
  })

  it('exports historical image rows with image billing mode derived from image_count', async () => {
    query.mockResolvedValue({
      items: [
        {
          ...usageLog,
          request_id: 'req-user-export-legacy-image',
          actual_cost: 0.2,
          total_cost: 0.2,
          input_cost: 0,
          output_cost: 0,
          cache_creation_cost: 0,
          cache_read_cost: 0,
          input_tokens: 0,
          output_tokens: 0,
          cache_creation_tokens: 0,
          cache_read_tokens: 0,
          image_count: 1,
          model: 'gpt-image-2',
          billing_mode: null,
          ip_address: null,
        },
      ],
      total: 1,
      pages: 1,
    })

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
    window.URL.createObjectURL = vi.fn(() => 'blob:usage-export') as typeof window.URL.createObjectURL
    window.URL.revokeObjectURL = vi.fn(() => {}) as typeof window.URL.revokeObjectURL
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    await (wrapper.vm as any).exportToCSV()

    expect(csvContent).toContain('Billing Mode')
    expect(csvContent).toContain('Image')
    expect(csvContent).not.toContain(',Token,0,0,0,0,')

    window.URL.createObjectURL = originalCreateObjectURL
    window.URL.revokeObjectURL = originalRevokeObjectURL
    vi.unstubAllGlobals()
    clickSpy.mockRestore()
  })
})
