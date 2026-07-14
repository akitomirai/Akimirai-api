import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import ModelStatusPanel from '../ModelStatusPanel.vue'
import type { UserMonitorView } from '@/api/channelMonitor'

const messages: Record<string, string> = {
  'channelStatus.modelStats.title': 'Model Status',
  'channelStatus.modelStats.models': 'Models',
  'channelStatus.modelStats.successRate': 'Success Rate',
  'channelStatus.modelStats.token24h': '24h Token',
  'channelStatus.modelStats.normalModels': 'Normal Models',
  'channelStatus.modelStats.searchPlaceholder': 'Search model...',
  'channelStatus.modelStats.good': 'Good',
  'channelStatus.modelStats.fair': 'Fair',
  'channelStatus.modelStats.severe': 'Severe',
  'channelStatus.modelStats.noData': 'No data',
  'channelStatus.modelStats.normal': 'Normal',
  'channelStatus.modelStats.needsAttention': 'Needs attention',
  'channelStatus.modelStats.past': 'Past',
  'channelStatus.modelStats.now': 'Now',
  'channelStatus.modelStats.successFailed': 'Success/Failed',
  'channelStatus.modelStats.avgLatency': 'Avg Latency',
  'channelStatus.modelStats.cacheRate': 'Cache Rate',
  'channelStatus.modelStats.monitors': 'Monitors',
  'channelStatus.empty.title': 'No channels available',
  'channelStatus.empty.description': 'No monitored channels have been configured yet.',
  'monitorCommon.status.operational': 'Operational',
  'monitorCommon.status.failed': 'Failed',
  'monitorCommon.providers.openai': 'OpenAI',
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

const timeline = [
  { status: 'operational', latency_ms: 1200, ping_latency_ms: 20, checked_at: '2026-07-07T00:00:00Z' },
  { status: 'operational', latency_ms: 1000, ping_latency_ms: 18, checked_at: '2026-07-07T00:01:00Z' },
  { status: 'failed', latency_ms: 2000, ping_latency_ms: 30, checked_at: '2026-07-07T00:02:00Z' },
] as UserMonitorView['timeline']

function monitor(model: string, id: number): UserMonitorView {
  return {
    id,
    name: model,
    provider: 'openai',
    group_name: 'default',
    primary_model: model,
    primary_status: 'operational',
    primary_latency_ms: 900,
    primary_ping_latency_ms: 21,
    availability_7d: 96,
    extra_models: [],
    timeline,
  }
}

function mountPanel(propOverrides: Record<string, unknown> = {}) {
  return mount(ModelStatusPanel, {
    props: {
      items: [monitor('gpt-5.5', 1), monitor('gpt-5.4-mini', 2)],
      window: '7d',
      loading: false,
      detailCache: {},
      usageModels: [
        {
          model: 'gpt-5.5',
          requests: 10,
          input_tokens: 10,
          output_tokens: 5,
          cache_creation_tokens: 0,
          cache_read_tokens: 90,
          total_tokens: 100_000_000,
          cost: 1,
          actual_cost: 0.5,
        },
      ],
      ...propOverrides,
    },
    global: {
      stubs: {
        EmptyState: { template: '<div data-test="empty-state"><slot /></div>' },
      },
    },
  })
}

describe('ModelStatusPanel', () => {
  it('renders aggregated model status rows with usage stats', () => {
    const wrapper = mountPanel()

    const text = wrapper.text()
    expect(text).toContain('Model Status')
    expect(text).toContain('gpt-5.5')
    expect(text).toContain('gpt-5.4-mini')
    expect(text).toContain('10000.00w')
    expect(text).toContain('90.00%')
    expect(text).toContain('2/1')
    expect(wrapper.findAll('[data-test="model-status-bar"]')).toHaveLength(120)
  })

  it('filters rows by model name', async () => {
    const wrapper = mountPanel()

    await wrapper.get('input[type="search"]').setValue('mini')
    await nextTick()

    expect(wrapper.text()).not.toContain('gpt-5.5')
    expect(wrapper.text()).toContain('gpt-5.4-mini')
  })

  it('keeps the desktop two-column model flow when only one model is visible', () => {
    const wrapper = mountPanel({
      items: [monitor('gpt-5.5', 1)],
    })

    const grid = wrapper.get('[data-test="model-status-grid"]')
    expect(grid.classes()).toContain('lg:grid-cols-2')
    expect(grid.classes()).not.toContain('grid-cols-1')
  })
})
