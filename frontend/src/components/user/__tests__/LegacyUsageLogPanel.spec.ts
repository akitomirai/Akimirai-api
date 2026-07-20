import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import LegacyUsageLogPanel from '../LegacyUsageLogPanel.vue'

const { query, listMyErrorRequests, showError, showWarning, showSuccess } = vi.hoisted(() => ({
  query: vi.fn(),
  listMyErrorRequests: vi.fn(),
  showError: vi.fn(),
  showWarning: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api', () => ({
  usageAPI: { query, listMyErrorRequests },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showWarning, showSuccess }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'usage.tabs.usage': 'Usage Details',
    'usage.tabs.errors': 'Error Requests',
    'admin.users.columnSettings': 'Column Settings',
    'common.refresh': 'Refresh',
    'common.reset': 'Reset',
    'usage.exportCsv': 'Export CSV',
    'usage.apiKeyFilter': 'API Key',
    'usage.model': 'Model',
    'usage.reasoningEffort': 'Reasoning',
    'usage.type': 'Type',
    'admin.usage.group': 'Group',
    'admin.usage.billingMode': 'Billing Mode',
    'usage.tokens': 'Tokens',
    'usage.cacheHitRate': 'Cache Hit Rate',
    'usage.cost': 'Cost',
    'usage.latency': 'Latency',
    'usage.time': 'Time',
  }
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => messages[key] ?? key }),
  }
})

const UsageTableStub = {
  name: 'UsageTable',
  props: ['data', 'columns', 'tokenBreakdown'],
  template: `<div data-test="legacy-usage-table" :data-token-breakdown="String(tokenBreakdown)" :data-column-keys="columns.map((column) => column.key).join('|')">{{ columns.map((column) => column.label).join(',') }}</div>`,
}

const UserErrorRequestsTableStub = {
  name: 'UserErrorRequestsTable',
  props: ['rows'],
  template: '<div data-test="legacy-error-table">{{ rows.length }}</div>',
}

const SelectStub = {
  name: 'Select',
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  template: '<div />',
}

describe('LegacyUsageLogPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    query.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20 })
    listMyErrorRequests.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20 })
  })

  it('uses the legacy usage source and keeps breakdown, tabs, and column settings', async () => {
    const wrapper = mount(LegacyUsageLogPanel, {
      props: {
        startDate: '2026-07-12',
        endDate: '2026-07-13',
        startTime: '22:00',
        endTime: '22:00',
        period: '24h',
        timezone: 'Asia/Shanghai',
        apiKeys: [{ id: 1, name: 'Baka1' }] as any,
        groups: [{ id: 1, name: 'Pro pool' }] as any,
        models: ['gpt-5.6-sol'],
        errorViewEnabled: true,
      },
      global: {
        stubs: {
          UsageTable: UsageTableStub,
          UserErrorRequestsTable: UserErrorRequestsTableStub,
          Select: SelectStub,
          Pagination: true,
          Icon: true,
        },
      },
    })
    await flushPromises()

    expect(query).toHaveBeenCalledWith(expect.objectContaining({
      period: '24h',
      page: 1,
      sort_by: 'created_at',
      sort_order: 'desc',
    }))
    expect(wrapper.get('[data-test="legacy-usage-table"]').attributes('data-token-breakdown')).toBe('true')
    expect(wrapper.get('[data-test="legacy-usage-table"]').text()).toContain('Tokens')
    expect(wrapper.get('[data-test="legacy-usage-table"]').attributes('data-column-keys')).not.toMatch(/reasoning_effort|ip_address|stream/)
    expect(wrapper.get('[data-test="legacy-usage-table"]').attributes('data-column-keys')).toContain('billing_mode')
    expect(wrapper.text()).toContain('Column Settings')
    expect(wrapper.text()).toContain('Usage Details')
    expect(wrapper.text()).toContain('Error Requests')

    await wrapper.findAll('button').find((button) => button.text().includes('Column Settings'))!.trigger('click')
    expect(wrapper.text()).toContain('Reasoning')
    expect(wrapper.text()).toContain('IP')
    expect(wrapper.text()).toContain('Type')
    expect(wrapper.text()).toContain('Billing Mode')

    const errorTab = wrapper.findAll('button').find((button) => button.text().includes('Error Requests'))
    await errorTab!.trigger('click')
    await flushPromises()

    expect(listMyErrorRequests).toHaveBeenCalledWith(expect.objectContaining({
      period: '24h',
      page: 1,
      sort_by: 'created_at',
    }))
    expect(listMyErrorRequests.mock.calls[0][0]).not.toHaveProperty('start_time')
    expect(wrapper.find('[data-test="legacy-error-table"]').exists()).toBe(true)
  })
})
