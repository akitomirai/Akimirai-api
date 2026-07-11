import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import RequestDiagnosticsTable from '../RequestDiagnosticsTable.vue'
import type { AdminUsageLog } from '@/types'

const messages: Record<string, string> = {
  'admin.usage.user': 'User',
  'usage.apiKeyFilter': 'API Key',
  'admin.usage.diagnostics.requestStartedAt': 'Request started',
  'admin.usage.diagnostics.requestFeatures': 'Request Features',
  'admin.usage.diagnostics.route': 'Route',
  'admin.usage.diagnostics.timings': 'Stage Timings',
  'admin.usage.diagnostics.retrySwitch': 'Retries / switches',
  'admin.usage.diagnostics.upstreamStatus': 'Upstream status',
  'admin.usage.diagnostics.title': 'Request Diagnostics',
  'admin.usage.diagnostics.firstByte': 'First byte',
  'admin.usage.diagnostics.firstToken': 'First token',
  'admin.usage.diagnostics.completed': 'Completed',
  'admin.usage.diagnostics.proxyFallback': 'Proxy',
  'admin.usage.diagnostics.directRoute': 'Direct',
  'admin.usage.diagnostics.unavailable': 'Unavailable',
  'admin.usage.diagnostics.open': 'Open request diagnostics',
  'usage.stream': 'Stream',
  'usage.ws': 'WS',
  'usage.sync': 'Sync',
  'usage.cyber': 'Cyber',
  'usage.unknown': 'Unknown',
}

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => messages[key] ?? key }),
}))

vi.mock('@/utils/format', () => ({
  formatDateTime: (value: string) => value,
}))

const DataTableStub = {
  props: ['data', 'columns', 'density'],
  template: `
    <div>
      <div data-test="headers">{{ columns.map((column) => column.label).join('|') }}</div>
      <div data-test="density">{{ density }}</div>
      <div v-for="row in data" :key="row.id" data-test="row">
        <slot name="cell-user" :row="row" />
        <slot name="cell-api_key" :row="row" />
        <slot name="cell-created_at" :row="row" />
        <slot name="cell-features" :row="row" />
        <slot name="cell-route" :row="row" />
        <slot name="cell-timings" :row="row" />
        <slot name="cell-retries" :row="row" />
        <slot name="cell-status" :row="row" />
      </div>
    </div>
  `,
}

describe('RequestDiagnosticsTable', () => {
  it('shows request-oriented fields inline in a compact table', () => {
    const row = {
      id: 77,
      user: { id: 12, email: 'user@example.com' },
      api_key: { id: 33, name: 'AI02' },
      created_at: '2026-07-11T03:12:00Z',
      request_started_at: '2026-07-11T03:11:59Z',
      request_id: 'req-diagnostics',
      model: 'gpt-5.6-sol',
      inbound_endpoint: '/v1/responses',
      stream: true,
      request_type: 'stream',
      request_body_bytes: 2048,
      route_kind: 'proxy',
      proxy_id_snapshot: 8,
      proxy_name_snapshot: 'JP-01',
      route_fingerprint: 'abcdef1234567890',
      account: { id: 14, name: 'Account A' },
      upstream_first_byte_ms: 1200,
      request_first_token_ms: 2300,
      request_total_ms: 4500,
      retry_count: 2,
      account_switch_count: 1,
      final_upstream_status: 200,
    } as AdminUsageLog

    const wrapper = mount(RequestDiagnosticsTable, {
      props: { data: [row] },
      global: { stubs: { DataTable: DataTableStub, Icon: true } },
    })

    expect(wrapper.get('[data-test="headers"]').text()).toBe(
      'User|API Key|Request started|Request Features|Route|Stage Timings|Retries / switches|Upstream status',
    )
    expect(wrapper.get('[data-test="density"]').text()).toBe('compact')
    expect(wrapper.text()).toContain('user@example.com')
    expect(wrapper.text()).toContain('AI02')
    expect(wrapper.text()).toContain('gpt-5.6-sol')
    expect(wrapper.text()).toContain('/v1/responses')
    expect(wrapper.text()).toContain('2.0 KB')
    expect(wrapper.text()).toContain('JP-01 #8')
    expect(wrapper.text()).toContain('1.20s')
    expect(wrapper.text()).toContain('2 / 1')
    expect(wrapper.text()).toContain('200')

    expect(wrapper.find('button[title="Open request diagnostics"]').exists()).toBe(false)
  })

  it('renders only the columns selected by the parent view', () => {
    const wrapper = mount(RequestDiagnosticsTable, {
      props: {
        data: [],
        columns: [
          { key: 'user', label: 'User' },
          { key: 'timings', label: 'Stage Timings' },
        ],
      },
      global: { stubs: { DataTable: DataTableStub } },
    })

    expect(wrapper.get('[data-test="headers"]').text()).toBe('User|Stage Timings')
  })

  it('does not fabricate request offsets for historical rows', () => {
    const row = {
      id: 78,
      created_at: '2026-07-11T03:12:00Z',
      first_token_ms: 999,
      duration_ms: 2222,
      request_started_at: null,
      upstream_first_byte_ms: null,
      request_first_token_ms: null,
      request_total_ms: null,
    } as AdminUsageLog

    const wrapper = mount(RequestDiagnosticsTable, {
      props: { data: [row] },
      global: { stubs: { DataTable: DataTableStub } },
    })

    expect(wrapper.text()).not.toContain('2026-07-11T03:12:00Z')
    expect(wrapper.text()).not.toContain('999ms')
    expect(wrapper.text()).not.toContain('2.22s')
    expect(wrapper.text().match(/Unavailable/g)).toHaveLength(5)
  })
})
