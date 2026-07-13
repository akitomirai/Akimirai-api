import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import UserSpendingRankingChart from '../UserSpendingRankingChart.vue'

const messages: Record<string, string> = {
  'admin.dashboard.spendingRankingTitle': 'User Spending Ranking',
  'admin.dashboard.spendingRankingUser': 'User',
  'admin.dashboard.spendingRankingRequests': 'Requests',
  'admin.dashboard.spendingRankingTokens': 'Tokens',
  'admin.dashboard.spendingRankingSpend': 'Spend',
  'admin.dashboard.spendingRankingOther': 'Others',
  'admin.dashboard.noDataAvailable': 'No data available',
  'admin.dashboard.failedToLoad': 'Failed to load',
  'admin.redeem.userPrefix': 'User #{id}'
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => messages[key] ?? key })
  }
})

vi.mock('vue-chartjs', () => ({
  Doughnut: {
    props: ['data'],
    template: '<div class="chart-data">{{ JSON.stringify(data) }}</div>'
  }
}))

describe('UserSpendingRankingChart', () => {
  const items = [
    { user_id: 1, email: 'alpha@example.com', actual_cost: 12, requests: 10, tokens: 1000 },
    { user_id: 2, email: 'beta@example.com', actual_cost: 8, requests: 6, tokens: 600 }
  ]

  it('renders ranked users and a separately colored Others slice', () => {
    const wrapper = mount(UserSpendingRankingChart, {
      props: {
        items,
        totalActualCost: 30,
        totalRequests: 20,
        totalTokens: 2000
      },
      global: { stubs: { LoadingSpinner: true } }
    })

    const chartData = JSON.parse(wrapper.get('.chart-data').text())
    expect(chartData.labels).toEqual(['#1 alpha@example.com', '#2 beta@example.com', 'Others'])
    expect(chartData.datasets[0].data).toEqual([12, 8, 10])
    expect(chartData.datasets[0].backgroundColor[0]).toBe('#3b82f6')
    expect(chartData.datasets[0].backgroundColor[2]).toBe('#94a3b8')

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(3)
    expect(rows[2].text()).toContain('Others')
    expect(rows[2].text()).toContain('4')
    expect(rows[2].text()).toContain('400')
    expect(rows[2].text()).toContain('$10.00')
  })

  it('emits selection only for concrete users', async () => {
    const wrapper = mount(UserSpendingRankingChart, {
      props: {
        items,
        totalActualCost: 30,
        totalRequests: 20,
        totalTokens: 2000
      },
      global: { stubs: { LoadingSpinner: true } }
    })

    const rows = wrapper.findAll('tbody tr')
    await rows[0].trigger('click')
    await rows[2].trigger('click')

    expect(wrapper.emitted('select-user')).toEqual([[items[0]]])
  })

  it('keeps loading, error, and empty states distinct', async () => {
    const wrapper = mount(UserSpendingRankingChart, {
      props: { items: [], loading: true },
      global: { stubs: { LoadingSpinner: { template: '<div data-test="spinner" />' } } }
    })

    expect(wrapper.find('[data-test="spinner"]').exists()).toBe(true)

    await wrapper.setProps({ loading: false, error: true })
    expect(wrapper.text()).toContain('Failed to load')

    await wrapper.setProps({ error: false })
    expect(wrapper.text()).toContain('No data available')
  })
})
