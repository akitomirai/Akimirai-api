import { mount } from '@vue/test-utils'
import { nextTick, ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { UserDashboardStats } from '@/api/usage'

const currentLocale = ref('zh')

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      locale: currentLocale,
      t: (key: string) => key,
    }),
  }
})

import UserDashboardStats from '../UserDashboardStats.vue'

const stats: UserDashboardStats = {
  total_api_keys: 3,
  active_api_keys: 3,
  total_requests: 100,
  total_input_tokens: 250_000_000,
  total_output_tokens: 150_000_000,
  total_cache_creation_tokens: 0,
  total_cache_read_tokens: 0,
  total_tokens: 400_000_000,
  total_cost: 0,
  total_actual_cost: 0,
  today_requests: 10,
  today_input_tokens: 25_000_000,
  today_output_tokens: 15_000_000,
  today_cache_creation_tokens: 0,
  today_cache_read_tokens: 0,
  today_tokens: 40_000_000,
  today_cost: 0,
  today_actual_cost: 0,
  average_duration_ms: 120,
  rpm: 2,
  tpm: 400_000_000,
}

const mountStats = (isSimple = true) => mount(UserDashboardStats, {
  props: {
    stats,
    balance: 0,
    isSimple,
    platformQuotas: [],
  },
  global: {
    stubs: {
      Icon: { template: '<span />' },
    },
  },
})

describe('UserDashboardStats', () => {
  beforeEach(() => {
    currentLocale.value = 'zh'
  })

  it('uses one complete grid and emphasizes total tokens in simple mode', () => {
    const wrapper = mountStats()
    const grid = wrapper.get('[data-testid="dashboard-stats-grid"]')
    const cards = Array.from(grid.element.children)

    expect(wrapper.findAll('[data-testid="dashboard-stats-grid"]')).toHaveLength(1)
    expect(cards).toHaveLength(7)
    expect(cards[3]?.textContent).toContain('dashboard.todayTokens')
    expect(cards[4]?.textContent).toContain('dashboard.totalTokens')
    expect(wrapper.get('[data-testid="dashboard-total-tokens"]').classes()).toContain('col-span-2')
    expect(wrapper.get('[data-testid="dashboard-total-tokens"]').text()).toContain('4.00亿')
    expect(wrapper.text()).not.toContain('dashboard.input')
    expect(wrapper.text()).not.toContain('dashboard.output')
  })

  it('formats token totals according to the active locale', async () => {
    const wrapper = mountStats()
    expect(wrapper.get('[data-testid="dashboard-total-tokens"]').text()).toContain('4.00亿')
    expect(wrapper.get('[data-testid="dashboard-total-tokens"] [title="400,000,000"]').exists()).toBe(true)

    currentLocale.value = 'en'
    await nextTick()

    expect(wrapper.get('[data-testid="dashboard-total-tokens"]').text()).toContain('40000.00w')
    expect(wrapper.get('[data-testid="dashboard-total-tokens"] [title="400,000,000"]').exists()).toBe(true)
  })

  it('keeps equal-width token cards when billing stats are enabled', () => {
    const wrapper = mountStats(false)
    const grid = wrapper.get('[data-testid="dashboard-stats-grid"]')

    expect(grid.element.children).toHaveLength(8)
    expect(wrapper.get('[data-testid="dashboard-total-tokens"]').classes()).not.toContain('col-span-2')
    expect(wrapper.text()).toContain('dashboard.balance')
  })
})
