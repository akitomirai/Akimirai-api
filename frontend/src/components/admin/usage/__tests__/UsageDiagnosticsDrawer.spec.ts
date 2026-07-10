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
      request_body_read_ms: 12,
      upstream_request_written_ms: 150,
      upstream_first_byte_ms: 700,
      request_first_token_ms: 900,
      route_kind: 'proxy',
      proxy_id_snapshot: 8,
      proxy_name_snapshot: 'jp-egress',
      proxy_protocol_snapshot: 'socks5',
      route_fingerprint: '1234567890abcdef',
      retry_count: 1,
      account_switch_count: 0,
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

    const errorsButton = wrapper.findAll('button').find((button) => button.text().includes('viewErrors'))
    expect(errorsButton).toBeDefined()
    await errorsButton!.trigger('click')
    expect(wrapper.emitted('openErrors')).toEqual([['req-diagnostics']])
  })
})
