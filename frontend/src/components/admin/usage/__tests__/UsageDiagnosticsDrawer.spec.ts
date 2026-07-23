import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UsageDiagnosticsDrawer from '../UsageDiagnosticsDrawer.vue'

const getDiagnostics = vi.fn()

vi.mock('@/api/admin/usage', () => ({
  adminUsageAPI: {
    getDiagnostics: (...args: unknown[]) => getDiagnostics(...args),
  },
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params?.id ? `${key}:${params.id}` : key,
    }),
  }
})

describe('UsageDiagnosticsDrawer', () => {
  afterEach(() => {
    getDiagnostics.mockReset()
    document.body.style.overflow = ''
  })

  it('loads request diagnostics and opens correlated errors', async () => {
    getDiagnostics.mockResolvedValue({
      id: 42,
      request_id: 'req-diagnostics',
      model: 'gpt-5.6-sol',
      created_at: '2026-07-11T01:02:03Z',
      request_started_at: '2026-07-11T01:02:01Z',
      request_total_ms: 2000,
      auth_latency_ms: 25,
      request_body_read_ms: 12,
      routing_latency_ms: 35,
      upstream_connection_reused: false,
      upstream_connection_ready_ms: 100,
      upstream_dns_lookup_ms: 10,
      upstream_tcp_connect_ms: 20,
      upstream_tls_handshake_ms: 30,
      upstream_request_headers_written_ms: 120,
      upstream_request_written_ms: 150,
      upstream_latency_ms: 500,
      upstream_first_byte_ms: 700,
      upstream_response_headers_received_ms: 710,
      upstream_response_body_first_byte_ms: 730,
      upstream_first_event_ms: 760,
      request_first_output_character_ms: 800,
      request_first_token_ms: 900,
      request_body_bytes: 2048,
      route_kind: 'proxy',
      proxy_id_snapshot: 8,
      proxy_name_snapshot: 'jp-egress',
      proxy_protocol_snapshot: 'socks5',
      route_fingerprint: '1234567890abcdef',
      retry_count: 1,
      account_switch_count: 0,
      prompt_cache_key_hash: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
      prompt_cache_key_source: 'client_header',
      prompt_cache_prefix_hash: 'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb',
      prompt_cache_tools_hash: 'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc',
      prompt_cache_system_hash: 'dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd',
      account_id: 3,
      attempt_timeline: [],
    })

    const wrapper = mount(UsageDiagnosticsDrawer, {
      props: { show: true, usageId: 42 },
      global: { stubs: { Teleport: true, Icon: true } },
    })
    await flushPromises()

    expect(getDiagnostics).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('req-diagnostics')
    expect(wrapper.text()).toContain('jp-egress #8 (socks5)')
    expect(wrapper.text()).toContain('12345678')
    expect(wrapper.get('[data-testid="prompt-cache-diagnostics"]').text()).toContain('admin.usage.diagnostics.promptCacheSource.client_header')
    expect(wrapper.get('[data-testid="prompt-cache-diagnostics"]').text()).toContain('aaaaaaaaaaaa')
    expect(wrapper.get('[data-testid="prompt-cache-diagnostics"]').text()).toContain('bbbbbbbbbbbb')
    for (const checkpoint of [
      'bodyRead',
      'routing',
      'requestWritten',
      'firstByte',
      'firstToken',
      'completed',
    ]) {
      expect(wrapper.text()).toContain(`admin.usage.diagnostics.${checkpoint}`)
    }
    for (const detail of [
      'bodyReadDetail',
      'routingDetail',
      'requestWrittenDetail',
      'firstByteDetail',
      'firstTokenDetail',
      'completedDetail',
    ]) {
      expect(wrapper.text()).toContain(`admin.usage.diagnostics.${detail}`)
    }
    expect(wrapper.findAll('[data-testid^="timing-step-"]')).toHaveLength(6)
    expect(wrapper.get('[data-testid="timing-step-body"]').text()).toContain('12ms')
    expect(wrapper.get('[data-testid="timing-step-routing"]').text()).toContain('35ms')
    expect(wrapper.get('[data-testid="timing-step-written"]').text()).toContain('150ms')
    expect(wrapper.get('[data-testid="timing-step-first-byte"]').text()).toContain('700ms')
    expect(wrapper.get('[data-testid="timing-step-first-token"]').text()).toContain('900ms')
    expect(wrapper.get('[data-testid="timing-step-total"]').text()).toContain('2.00s')
    expect(wrapper.text()).toContain('2.00s')
    expect(wrapper.text()).toContain('2.0 KiB')
    expect(wrapper.text()).not.toContain('admin.usage.diagnostics.dnsResolution')
    expect(wrapper.text()).not.toContain('admin.usage.diagnostics.requestHeadersSent')

    const errorsButton = wrapper.findAll('button').find((button) => button.text().includes('viewErrors'))
    expect(errorsButton).toBeDefined()
    await errorsButton!.trigger('click')
    expect(wrapper.emitted('openErrors')).toEqual([['req-diagnostics']])
  })

  it('does not relabel legacy upstream-relative timings as request-scoped diagnostics', async () => {
    getDiagnostics.mockResolvedValue({
      id: 43,
      request_id: 'req-legacy-timings',
      model: 'gpt-5.6-sol',
      created_at: '2026-07-11T01:02:03Z',
      first_token_ms: 999,
      duration_ms: 2222,
      route_kind: 'direct',
      retry_count: 0,
      account_switch_count: 0,
      attempt_timeline: [],
    })

    const wrapper = mount(UsageDiagnosticsDrawer, {
      props: { show: true, usageId: 43 },
      global: { stubs: { Teleport: true, Icon: true } },
    })
    await flushPromises()

    expect(wrapper.get('[data-testid="timing-step-first-token"]').text()).toContain('admin.usage.diagnostics.unavailable')
    expect(wrapper.get('[data-testid="timing-step-total"]').text()).toContain('admin.usage.diagnostics.unavailable')
    expect(wrapper.text()).not.toContain('999ms')
    expect(wrapper.text()).not.toContain('2.22s')
  })
})
