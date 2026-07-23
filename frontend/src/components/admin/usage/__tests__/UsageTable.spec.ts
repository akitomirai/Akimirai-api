const ipGeoMocks = vi.hoisted(() => ({
  getEntry: vi.fn(() => ({ status: 'idle' as const })),
  fetchOne: vi.fn(),
  fetchBatch: vi.fn(),
}))

vi.mock('@/utils/ipGeoLookup', () => ipGeoMocks)

import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import UsageTable from '../UsageTable.vue'

const messages: Record<string, string> = {
  'admin.usage.userDeletedBadge': 'Deleted',
  'usage.costDetails': 'Cost Breakdown',
  'admin.usage.inputCost': 'Input Cost',
  'admin.usage.outputCost': 'Output Cost',
  'admin.usage.cacheCreationCost': 'Cache Creation Cost',
  'admin.usage.cacheReadCost': 'Cache Read Cost',
  'admin.usage.inputTokens': 'Input Tokens',
  'admin.usage.outputTokens': 'Output Tokens',
  'admin.usage.cacheCreationTokens': 'Cache Creation Tokens',
  'admin.usage.cacheReadTokens': 'Cache Read Tokens',
  'usage.inputTokenPrice': 'Input price',
  'usage.outputTokenPrice': 'Output price',
  'usage.cacheReadTokenPrice': 'Cache read price',
  'usage.tokenDetails': 'Token Details',
  'usage.tokenUnitPrice': 'Unit price',
  'usage.tokenCost': 'Cost',
  'usage.totalTokens': 'Total Tokens',
  'usage.cacheHitRate': 'Cache hit rate',
  'usage.latencyFirstToken': 'First Token',
  'usage.latencyDuration': 'Total',
  'usage.perMillionTokens': '/ 1M tokens',
  'usage.serviceTier': 'Service tier',
  'usage.serviceTierPriority': 'Fast',
  'usage.serviceTierFlex': 'Flex',
  'usage.serviceTierStandard': 'Standard',
  'usage.rate': 'Rate',
  'usage.accountMultiplier': 'Account rate',
  'usage.original': 'Original',
  'usage.userBilled': 'User billed',
  'usage.accountBilled': 'Account billed',
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
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

const DataTableStub = {
  props: ['data'],
  template: `
    <div>
      <div v-for="row in data" :key="row.request_id">
        <slot name="cell-model" :row="row" :value="row.model" />
        <slot name="cell-group" :row="row" />
        <slot name="cell-account" :row="row" />
        <slot name="cell-billing_mode" :row="row" />
        <slot name="cell-tokens" :row="row" />
        <slot name="cell-cache_hit_rate" :row="row" />
        <slot name="cell-cost" :row="row" />
        <slot name="cell-latency" :row="row" />
        <slot name="cell-diagnostics" :row="row" />
      </div>
    </div>
  `,
}

const baseImageRow = {
  request_id: 'req-admin-image',
  model: 'gpt-image-2',
  actual_cost: 0.4,
  total_cost: 0.4,
  account_rate_multiplier: 1,
  rate_multiplier: 1,
  service_tier: null,
  input_cost: 0,
  output_cost: 0,
  cache_creation_cost: 0,
  cache_read_cost: 0,
  input_tokens: 0,
  output_tokens: 0,
  cache_creation_tokens: 0,
  cache_read_tokens: 0,
  cache_creation_5m_tokens: 0,
  cache_creation_1h_tokens: 0,
  cache_ttl_overridden: false,
  billing_mode: 'image',
  image_count: 2,
  image_size: '2K',
  image_input_size: null,
  image_output_size: null,
  image_size_source: null,
  image_size_breakdown: null,
}

describe('admin UsageTable tooltip', () => {
  beforeEach(() => {
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
      x: 0,
      y: 0,
      top: 20,
      left: 20,
      right: 120,
      bottom: 40,
      width: 100,
      height: 20,
      toJSON: () => ({}),
    } as DOMRect)
  })

  it('marks only usage rows that actually applied long-context billing', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            ...baseImageRow,
            request_id: 'req-long-context-enabled',
            long_context_billing_applied: true,
          },
          {
            ...baseImageRow,
            request_id: 'req-long-context-disabled',
            long_context_billing_applied: false,
          },
        ],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.findAll('[data-testid="long-context-billing-marker"]')).toHaveLength(1)
    expect(wrapper.get('[data-testid="long-context-billing-marker"]').text()).toBe('x2')
  })

  it('shows service tier and billing breakdown in cost tooltip', async () => {
    const row = {
      request_id: 'req-admin-1',
      actual_cost: 0.092883,
      total_cost: 0.092883,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      service_tier: 'priority',
      input_cost: 0.020285,
      output_cost: 0.00303,
      cache_creation_cost: 0,
      cache_read_cost: 0.069568,
      input_tokens: 4057,
      output_tokens: 101,
      cache_read_tokens: 100_000,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await wrapper.get('[data-testid="cost-detail-trigger"]').trigger('click')
    await nextTick()
    expect(wrapper.text()).not.toContain('Unit price')

    await wrapper.get('[data-testid="cost-detail-trigger"]').trigger('mouseenter')
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Service tier')
    expect(text).toContain('Fast')
    expect(text).toContain('Rate')
    expect(text).toContain('1.00x')
    expect(text).toContain('Account rate')
    expect(text).toContain('User billed')
    expect(text).toContain('Account billed')
    expect(text).toContain('$0.092883')
    expect(text).toContain('$5.0000 / 1M tokens')
    expect(text).toContain('$30.0000 / 1M tokens')
    expect(text).toContain('$0.6957 / 1M tokens')
    expect(text).toContain('$0.069568')
  })

  it('shows requested and upstream models separately for admin rows', () => {
    const row = {
      request_id: 'req-admin-model-1',
      model: 'claude-sonnet-4',
      upstream_model: 'claude-sonnet-4-20250514',
      actual_cost: 0,
      total_cost: 0,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('claude-sonnet-4')
    expect(text).toContain('claude-sonnet-4-20250514')
  })

  it('shows route and retry indicators and emits diagnostics click', async () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          id: 77,
          request_id: 'req-route',
          model: 'gpt-5.6-sol',
          account: { id: 3, name: 'codex-jp' },
          route_kind: 'proxy',
          proxy_name_snapshot: 'jp-egress',
          retry_count: 2,
          account_switch_count: 1,
          actual_cost: 0,
          total_cost: 0,
          input_cost: 0,
          output_cost: 0,
          rate_multiplier: 1,
          input_tokens: 0,
          output_tokens: 0,
        }],
        loading: false,
        columns: [],
      },
      global: { stubs: { DataTable: DataTableStub, EmptyState: true, Teleport: true } },
    })

    expect(wrapper.text()).toContain('jp-egress')
    expect(wrapper.text()).toContain('R2')
    expect(wrapper.text()).toContain('S1')
    await wrapper.get('[data-testid="diagnostics-trigger"]').trigger('click')
    expect(wrapper.emitted('diagnosticsClick')).toEqual([[77]])
  })

  it('appends lowercase reasoning effort badge to the requested model', () => {
    const row = {
      request_id: 'req-admin-model-reasoning',
      model: 'gpt-5.5',
      reasoning_effort: 'x-high',
      actual_cost: 0,
      total_cost: 0,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
      cache_creation_tokens: 0,
      cache_read_tokens: 0,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.text()).toContain('gpt-5.5')
    expect(wrapper.get('.reasoning-effort-badge').text()).toBe('xhigh')
    expect(wrapper.text()).not.toContain('XHigh')
  })

  it('keeps cache hit token text in a compact fixed-width lane', () => {
    const row = {
      request_id: 'req-admin-cache-hit',
      model: 'gpt-5.5',
      actual_cost: 0,
      total_cost: 0,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 631,
      output_tokens: 28,
      cache_creation_tokens: 0,
      cache_read_tokens: 172000,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const cacheCell = wrapper.get('[data-test="cache-hit-rate"]')
    expect(cacheCell.text()).toContain('99.6%')
    expect(cacheCell.text()).toContain('17.20w')
    expect(cacheCell.attributes('class')).toContain('w-[112px]')
    expect(cacheCell.attributes('class')).toContain('text-left')
    expect(wrapper.get('[data-testid="usage-total-token-count"]').text()).toBe('17.27w')
    expect(wrapper.text()).not.toContain('631')
    expect(wrapper.text()).not.toContain('28')
  })

  it('shows the token detail trigger beside both token layouts', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          request_id: 'req-token-trigger',
          model: 'gpt-5.5',
          actual_cost: 0,
          total_cost: 0,
          input_tokens: 12,
          output_tokens: 3,
          cache_read_tokens: 5,
        }],
        loading: false,
        columns: [],
        tokenBreakdown: true,
      },
      global: {
        stubs: { DataTable: DataTableStub, EmptyState: true, Teleport: true },
      },
    })

    expect(wrapper.get('[data-testid="token-detail-trigger"]').exists()).toBe(true)
  })

  it('opens the cost detail popover when hovering the cost icon', async () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          request_id: 'req-token-breakdown',
          model: 'gpt-5.5',
          input_tokens: 1000,
          output_tokens: 200,
          cache_read_tokens: 3000,
          input_cost: 0.005,
          output_cost: 0.006,
          cache_read_cost: 0.0015,
          total_cost: 0.0125,
          actual_cost: 0.01,
          rate_multiplier: 0.8,
        }],
        loading: false,
        columns: [],
      },
      global: {
        stubs: { DataTable: DataTableStub, EmptyState: true, Teleport: true },
      },
    })

    await wrapper.get('[data-testid="cost-detail-trigger"]').trigger('mouseenter')
    await nextTick()

    expect(wrapper.text()).toContain('Unit price')
    expect(wrapper.text()).toContain('Cost')
    expect(wrapper.text()).toContain('Total Tokens')
    expect(wrapper.get('[data-testid="cost-input-token-row"]').text()).toContain('1,000')
    expect(wrapper.get('[data-testid="cost-output-token-row"]').text()).toContain('200')
    expect(wrapper.get('[data-testid="cost-cache-token-row"]').text()).toContain('3,000')
    expect(wrapper.get('[data-testid="cost-input-token-icon"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="cost-output-token-icon"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="cost-cache-token-icon"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('0.80x')
    expect(wrapper.text()).toContain('$0.010000')

    await wrapper.get('[data-testid="cost-detail-trigger"]').trigger('mouseleave')
    await nextTick()
    expect(wrapper.text()).not.toContain('Unit price')
  })

  it('opens detailed token quantities when hovering the token icon', async () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          request_id: 'req-token-counts',
          model: 'gpt-5.5',
          input_tokens: 1000,
          output_tokens: 200,
          cache_creation_tokens: 400,
          cache_read_tokens: 3000,
          input_cost: 0.005,
          output_cost: 0.006,
          cache_read_cost: 0.0015,
          total_cost: 0.0125,
          actual_cost: 0.01,
          rate_multiplier: 0.8,
        }],
        loading: false,
        columns: [],
      },
      global: {
        stubs: { DataTable: DataTableStub, EmptyState: true, Teleport: true },
      },
    })

    await wrapper.get('[data-testid="token-detail-trigger"]').trigger('click')
    await nextTick()
    expect(wrapper.text()).not.toContain('Token Details')

    await wrapper.get('[data-testid="token-detail-trigger"]').trigger('mouseenter')
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Token Details')
    expect(text).toContain('Input Tokens1,000')
    expect(text).toContain('Output Tokens200')
    expect(text).toContain('Cache Creation Tokens400')
    expect(text).toContain('Cache Read Tokens3,000')
    expect(text).toContain('Total Tokens4,600')
    expect(text).not.toContain('Unit price')
    expect(text).not.toContain('User billed')

    await wrapper.get('[data-testid="token-detail-trigger"]').trigger('mouseleave')
    await nextTick()
    expect(wrapper.text()).not.toContain('Token Details')
  })

  it('joins the group name and multiplier in the group badge', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          request_id: 'req-group-label',
          model: 'gpt-5.5',
          group: { name: 'Pro pool', rate_multiplier: 0.07 },
          actual_cost: 0,
          total_cost: 0,
          input_tokens: 0,
          output_tokens: 0,
          cache_read_tokens: 0,
        }],
        loading: false,
        columns: [],
      },
      global: {
        stubs: { DataTable: DataTableStub, EmptyState: true, Teleport: true },
      },
    })

    expect(wrapper.text()).toContain('Pro pool · 0.07x')
  })

  it.each([
    {
      name: 'defaulted row',
      row: {
        ...baseImageRow,
        request_id: 'req-admin-default-image',
        image_size: '2K',
        image_input_size: 'auto',
        image_output_size: null,
        image_size_source: 'default',
      },
      expected: ['2K', 'Default billing tier', 'auto', 'unknown'],
    },
    {
      name: 'output-sourced row',
      row: {
        ...baseImageRow,
        request_id: 'req-admin-output-image',
        image_size: '4K',
        image_input_size: '1024x1024',
        image_output_size: '3840x2160',
        image_size_source: 'output',
        image_size_breakdown: { '4K': 1 },
      },
      expected: ['4K', 'Upstream output', '1024x1024', '3840x2160', '4K x 1'],
    },
    {
      name: 'input-sourced row',
      row: {
        ...baseImageRow,
        request_id: 'req-admin-input-image',
        image_size: '1K',
        image_input_size: '1024x1024',
        image_output_size: null,
        image_size_source: 'input',
      },
      expected: ['1K', 'Request input', '1024x1024', 'unknown'],
    },
    {
      name: 'legacy unstandardized row',
      row: {
        ...baseImageRow,
        request_id: 'req-admin-legacy-unstandardized-image',
        image_size: '512x512',
        image_input_size: null,
        image_output_size: null,
        image_size_source: null,
      },
      expected: ['legacy unstandardized: 512x512', 'Legacy record', 'unknown'],
    },
  ])('shows image usage metadata for $name', async ({ row, expected }) => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await wrapper.get('[data-testid="cost-detail-trigger"]').trigger('mouseenter')
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Image count')
    expect(text).toContain('Billing size')
    expect(text).toContain('Size source')
    expect(text).toContain('Input size')
    expect(text).toContain('Output size')
    expect(text).toContain('Per-image price')
    expect(text).toContain('Image total price')
    for (const value of expected) {
      expect(text).toContain(value)
    }
  })

  it('displays historical image rows with missing billing_mode as image usage without a 2K fallback', async () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            ...baseImageRow,
            request_id: 'req-admin-legacy-missing-image',
            billing_mode: null,
            image_size: null,
            image_input_size: null,
            image_output_size: null,
            image_size_source: null,
            image_size_breakdown: null,
          },
        ],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await wrapper.get('[data-testid="cost-detail-trigger"]').trigger('mouseenter')
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Image')
    expect(text).toContain('Image count')
    expect(text).toContain('Per-image price')
    expect(text).toContain('not recorded')
    expect(text).not.toContain('(2K)')
  })
})

describe('admin UsageTable latency semantics', () => {
  it('uses and labels the first-token metric when first-byte data is also present', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          request_id: 'req-first-byte',
          model: 'gpt-5.6-sol',
          upstream_first_byte_ms: 1200,
          request_first_token_ms: 2300,
          request_total_ms: 4500,
        }],
        columns: [{ key: 'latency', label: 'Latency' }],
      },
      global: { stubs: { DataTable: DataTableStub, Teleport: true } },
    })

    expect(wrapper.text()).toContain('First Token')
    expect(wrapper.text()).toContain('2.30s')
    expect(wrapper.text()).not.toContain('1.20s')
  })

  it('uses the legacy first-token field when request diagnostics are unavailable', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          request_id: 'req-first-token-fallback',
          model: 'gpt-5.6-sol',
          first_token_ms: 2300,
          duration_ms: 4500,
        }],
        columns: [{ key: 'latency', label: 'Latency' }],
      },
      global: { stubs: { DataTable: DataTableStub, Teleport: true } },
    })

    expect(wrapper.text()).toContain('First Token')
    expect(wrapper.text()).toContain('2.30s')
  })
})

describe('admin UsageTable IP geolocation batch toolbar', () => {
  const DataTableStubWithIp = {
    props: ['data'],
    template: `
      <div>
        <div v-for="row in data" :key="row.request_id">
          <slot name="cell-ip_address" :row="row" />
        </div>
      </div>
    `,
  }

  beforeEach(() => {
    ipGeoMocks.getEntry.mockReset()
    ipGeoMocks.fetchOne.mockReset()
    ipGeoMocks.fetchBatch.mockReset()
    ipGeoMocks.getEntry.mockReturnValue({ status: 'idle' })
  })

  it('does not render the batch toolbar when the ip_address column is not visible', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{ request_id: 'r1', ip_address: '8.8.8.8' }],
        loading: false,
        columns: [],
      },
      global: { stubs: { DataTable: DataTableStubWithIp, EmptyState: true, Teleport: true } },
    })
    expect(wrapper.text()).not.toContain('usage.ipGeo.batchFetch')
  })

  it('renders the batch toolbar with a pending count when the ip_address column is visible', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          { request_id: 'r1', ip_address: '8.8.8.8' },
          { request_id: 'r2', ip_address: '8.8.8.8' },
          { request_id: 'r3', ip_address: '1.1.1.1' },
        ],
        loading: false,
        columns: [{ key: 'ip_address', label: 'IP' }],
      },
      global: { stubs: { DataTable: DataTableStubWithIp, EmptyState: true, Teleport: true } },
    })
    expect(wrapper.text()).toContain('usage.ipGeo.pending')
    const button = wrapper.find('button')
    expect(button.exists()).toBe(true)
    expect((button.element as HTMLButtonElement).disabled).toBe(false)
  })

  it('fetches deduplicated IPs from the current page when the batch button is clicked', async () => {
    ipGeoMocks.fetchBatch.mockResolvedValue(true)
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          { request_id: 'r1', ip_address: '8.8.8.8' },
          { request_id: 'r2', ip_address: '8.8.8.8' },
          { request_id: 'r3', ip_address: '1.1.1.1' },
        ],
        loading: false,
        columns: [{ key: 'ip_address', label: 'IP' }],
      },
      global: { stubs: { DataTable: DataTableStubWithIp, EmptyState: true, Teleport: true } },
    })
    await wrapper.find('button').trigger('click')
    expect(ipGeoMocks.fetchBatch).toHaveBeenCalledWith(['8.8.8.8', '1.1.1.1'])
    expect(wrapper.emitted('ipGeoBatchFailed')).toBeUndefined()
  })

  it('emits ipGeoBatchFailed when the batch request reports a network-level failure', async () => {
    ipGeoMocks.fetchBatch.mockResolvedValue(false)
    const wrapper = mount(UsageTable, {
      props: {
        data: [{ request_id: 'r1', ip_address: '8.8.8.8' }],
        loading: false,
        columns: [{ key: 'ip_address', label: 'IP' }],
      },
      global: { stubs: { DataTable: DataTableStubWithIp, EmptyState: true, Teleport: true } },
    })
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('ipGeoBatchFailed')).toHaveLength(1)
  })

  it('renders IpGeoCell content for ip_address cells', () => {
    ipGeoMocks.getEntry.mockReturnValue({ status: 'success', label: 'CN · Guangdong · Shenzhen', detail: {} })
    const wrapper = mount(UsageTable, {
      props: {
        data: [{ request_id: 'r1', ip_address: '121.35.47.43' }],
        loading: false,
        columns: [{ key: 'ip_address', label: 'IP' }],
      },
      global: { stubs: { DataTable: DataTableStubWithIp, EmptyState: true, Teleport: true } },
    })
    expect(wrapper.text()).toContain('121.35.47.43')
    expect(wrapper.text()).toContain('CN · Guangdong · Shenzhen')
  })
})

// A DataTable stub that also renders cell-user, so the deleted badge can be asserted.
const DataTableStubWithUser = {
  props: ['data'],
  template: `
    <div>
      <div v-for="row in data" :key="row.request_id">
        <slot name="cell-user" :row="row" />
        <slot name="cell-model" :row="row" :value="row.model" />
        <slot name="cell-billing_mode" :row="row" />
        <slot name="cell-tokens" :row="row" />
        <slot name="cell-cost" :row="row" />
      </div>
    </div>
  `,
}

describe('admin UsageTable deleted-user badge', () => {
  it('renders deleted badge for a soft-deleted user row', () => {
    const row = {
      request_id: 'req-deleted-user-1',
      model: 'claude-3',
      user_id: 2,
      user: { id: 2, email: 'd@test.com', deleted_at: '2026-05-28T00:00:00Z' },
      actual_cost: 0,
      total_cost: 0,
      input_cost: 0,
      output_cost: 0,
      rate_multiplier: 1,
      input_tokens: 1,
      output_tokens: 1,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'user', label: 'User' }],
      },
      global: {
        stubs: {
          DataTable: DataTableStubWithUser,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.text()).toContain('Deleted')
    expect(wrapper.text()).toContain('d@test.com')
  })

  it('does NOT render deleted badge for an active user row', () => {
    const row = {
      request_id: 'req-active-user-1',
      model: 'claude-3',
      user_id: 3,
      user: { id: 3, email: 'active@test.com', deleted_at: null },
      actual_cost: 0,
      total_cost: 0,
      input_cost: 0,
      output_cost: 0,
      rate_multiplier: 1,
      input_tokens: 1,
      output_tokens: 1,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'user', label: 'User' }],
      },
      global: {
        stubs: {
          DataTable: DataTableStubWithUser,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.text()).not.toContain('Deleted')
    expect(wrapper.text()).toContain('active@test.com')
  })
})
