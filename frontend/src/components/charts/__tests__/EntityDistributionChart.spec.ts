import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import EntityDistributionChart from '../EntityDistributionChart.vue'

const messages: Record<string, string> = {
  'admin.dashboard.requests': 'Requests',
  'admin.dashboard.tokens': 'Tokens',
  'admin.dashboard.actual': 'Actual',
  'admin.dashboard.accountCost': 'Account Cost',
  'admin.dashboard.standard': 'Standard',
  'admin.dashboard.metricTokens': 'By Tokens',
  'admin.dashboard.metricActualCost': 'By Actual Cost',
  'admin.dashboard.noDataAvailable': 'No data available',
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => messages[key] ?? key }),
  }
})

vi.mock('vue-chartjs', () => ({
  Doughnut: {
    props: ['data'],
    template: '<div class="chart-data">{{ JSON.stringify(data) }}</div>',
  },
}))

const items = [
  { id: 1, label: 'key-a', requests: 3, total_tokens: 1200, cost: 1.8, actual_cost: 0.1 },
  { id: 2, label: 'key-b', requests: 2, total_tokens: 600, cost: 0.7, actual_cost: 0.9 },
]

describe('EntityDistributionChart', () => {
  it('sorts and charts by tokens by default', () => {
    const wrapper = mount(EntityDistributionChart, {
      props: { title: 'API Key Distribution', entityLabel: 'API Key', items },
      global: { stubs: { LoadingSpinner: true } },
    })

    const chartData = JSON.parse(wrapper.get('.chart-data').text())
    expect(chartData.labels).toEqual(['key-a', 'key-b'])
    expect(chartData.datasets[0].data).toEqual([1200, 600])
    expect(wrapper.findAll('tbody tr')[0].text()).toContain('key-a')
    expect(wrapper.findAll('tbody tr')[0].text()).toContain('1,200')
    expect(wrapper.text()).not.toContain('Account Cost')
  })

  it('sorts by actual cost and can show account cost', () => {
    const wrapper = mount(EntityDistributionChart, {
      props: {
        title: 'User Distribution',
        entityLabel: 'User',
        items: items.map((item) => ({ ...item, account_cost: item.actual_cost / 2 })),
        metric: 'actual_cost',
        showAccountCost: true,
      },
      global: { stubs: { LoadingSpinner: true } },
    })

    const chartData = JSON.parse(wrapper.get('.chart-data').text())
    expect(chartData.labels).toEqual(['key-b', 'key-a'])
    expect(chartData.datasets[0].data).toEqual([0.9, 0.1])
    expect(wrapper.text()).toContain('Account Cost')
  })
})
