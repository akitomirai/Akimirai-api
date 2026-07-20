import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import UsageStatsCards from '../UsageStatsCards.vue'
import { resetTokenCountModeForTests, setTokenCountMode } from '@/composables/useTokenCountMode'

const messages: Record<string, string> = {
  'usage.totalRequests': 'Total Requests',
  'usage.inSelectedRange': 'in selected range',
  'usage.totalTokens': 'Total Tokens',
  'usage.in': 'In',
  'usage.out': 'Out',
  'usage.cacheTotal': 'Cache',
  'usage.cacheBreakdown': 'Cache Token Breakdown',
  'usage.cacheCreationTokensLabel': 'Cache Creation',
  'usage.cacheReadTokensLabel': 'Cache Read',
  'usage.cacheRead': 'Read',
  'usage.totalCost': 'Total Cost',
  'usage.accountCost': 'Cost',
  'usage.standardCost': 'Standard',
  'usage.avgDuration': 'Avg Duration',
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

const stats = {
  total_requests: 1,
  total_input_tokens: 100,
  total_output_tokens: 50,
  total_cache_tokens: 34,
  total_cache_creation_tokens: 12,
  total_cache_read_tokens: 22,
  total_tokens: 137_440_000,
  total_cost: 0.001,
  total_actual_cost: 0.001,
  total_account_cost: 0.001,
  average_duration_ms: 250,
}

describe('UsageStatsCards', () => {
  beforeEach(() => {
    localStorage.clear()
    resetTokenCountModeForTests()
  })

  it('shows only the compact total token count', () => {
    const wrapper = mount(UsageStatsCards, {
      props: {
        stats,
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('Total Tokens')
    expect(text).toContain('13744.00w')
    expect(text).not.toContain('In:')
    expect(text).not.toContain('Out:')
    expect(text).not.toContain('Cache Token Breakdown')
  })

  it('shows legacy input, output, and cache-read counters with k/m/b units', () => {
    setTokenCountMode('legacy')
    const wrapper = mount(UsageStatsCards, {
      props: {
        stats: {
          ...stats,
          total_input_tokens: 1_250,
          total_output_tokens: 25_000,
          total_cache_read_tokens: 2_000_000,
        },
        showTokenBreakdown: true,
      },
      global: { stubs: { Icon: true } },
    })

    expect(wrapper.text()).toContain('In: 1.25k')
    expect(wrapper.text()).toContain('Out: 25.00k')
    expect(wrapper.text()).toContain('Read: 2.00m')
  })
})
