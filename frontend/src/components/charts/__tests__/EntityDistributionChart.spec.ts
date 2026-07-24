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
    props: ['data', 'options'],
    template: '<div class="chart-data doughnut-data">{{ JSON.stringify(data) }}</div>',
  },
  Bar: {
    name: 'BarChartStub',
    props: ['data', 'options'],
    template: '<div class="chart-data bar-data" :data-index-axis="options.indexAxis">{{ JSON.stringify(data) }}</div>',
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
    expect(wrapper.find('.doughnut-data').exists()).toBe(true)
    expect(wrapper.find('.bar-data').exists()).toBe(false)
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

  it('renders a vertically scrollable horizontal bar chart and updates it with the metric', async () => {
    const barItems = [
      ...items,
      { id: 3, label: 'key-c-with-a-very-long-label', requests: 1, total_tokens: 300, cost: 0.4, actual_cost: 0.3 },
      { id: 4, label: 'key-d', requests: 1, total_tokens: 200, cost: 0.3, actual_cost: 0.2 },
      { id: 5, label: 'key-e', requests: 1, total_tokens: 100, cost: 0.2, actual_cost: 0.1 },
      { id: 6, label: 'key-f', requests: 1, total_tokens: Number.NaN, cost: 0.1, actual_cost: 0.05 },
      { id: 7, label: 'key-g', requests: 1, total_tokens: Number.POSITIVE_INFINITY, cost: 0.1, actual_cost: 0.01 },
    ]
    const wrapper = mount(EntityDistributionChart, {
      props: {
        title: 'User Distribution',
        entityLabel: 'User',
        items: barItems,
        visualization: 'horizontal-bar',
      },
      global: { stubs: { LoadingSpinner: true } },
    })

    const chartData = JSON.parse(wrapper.get('.bar-data').text())
    expect(chartData.labels).toEqual([
      'key-a',
      'key-b',
      'key-c-with-a-very-long-label',
      'key-d',
      'key-e',
      'key-f',
      'key-g',
    ])
    expect(chartData.datasets[0].data).toEqual([1200, 600, 300, 200, 100, 0, 0])
    expect(wrapper.get('.bar-data').attributes('data-index-axis')).toBe('y')
    expect(wrapper.findComponent({ name: 'BarChartStub' }).props('options').scales.y.ticks.display).toBe(false)
    expect(wrapper.find('.doughnut-data').exists()).toBe(false)

    const scrollArea = wrapper.get('[data-testid="horizontal-bar-scroll"]')
    const tableScrollArea = wrapper.get('[data-testid="distribution-table-scroll"]')
    const chartCanvas = wrapper.get('[data-testid="horizontal-bar-canvas"]')
    expect(scrollArea.classes()).toContain('overflow-y-auto')
    expect(scrollArea.classes()).toContain('max-h-52')
    expect(scrollArea.classes()).toContain('sm:w-72')
    expect(scrollArea.attributes('role')).toBe('region')
    expect(scrollArea.attributes('tabindex')).toBe('0')
    expect(wrapper.get('[data-testid="horizontal-bar-header-spacer"]').classes()).toContain('h-6')
    expect(chartCanvas.attributes('style')).toContain('height: 210px')
    expect(wrapper.findAll('tbody tr').every((row) => row.classes().includes('h-[30px]'))).toBe(true)

    const barScrollElement = scrollArea.element as HTMLElement
    const tableScrollElement = tableScrollArea.element as HTMLElement
    barScrollElement.scrollTop = 60
    await scrollArea.trigger('scroll')
    expect(tableScrollElement.scrollTop).toBe(60)
    tableScrollElement.scrollTop = 30
    await tableScrollArea.trigger('scroll')
    expect(barScrollElement.scrollTop).toBe(30)

    await wrapper.setProps({ metric: 'actual_cost' })

    const actualCostData = JSON.parse(wrapper.get('.bar-data').text())
    expect(actualCostData.labels).toEqual([
      'key-b',
      'key-c-with-a-very-long-label',
      'key-d',
      'key-a',
      'key-e',
      'key-f',
      'key-g',
    ])
    expect(actualCostData.datasets[0].data).toEqual([0.9, 0.3, 0.2, 0.1, 0.1, 0.05, 0.01])
    expect(wrapper.findAll('tbody tr')[0].text()).toContain('key-b')
  })
})
