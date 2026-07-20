import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ModelUsageTrendChart from '../ModelUsageTrendChart.vue'
import { i18n } from '@/i18n'

const messages: Record<string, string> = {
  'dashboard.spendingDistribution': 'Spending distribution',
  'dashboard.callTrend': 'Call trend',
  'dashboard.chartType': 'Chart type',
  'dashboard.metricType': 'Y-axis metric',
  'dashboard.tokens': 'Tokens',
  'dashboard.quotaConsumption': 'Quota spend',
  'dashboard.actualConsumption': 'Actual spend',
  'dashboard.barChart': 'Bar',
  'dashboard.areaChart': 'Area',
  'dashboard.other': 'Other',
  'dashboard.noDataAvailable': 'No data available',
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

vi.mock('vue-chartjs', () => ({
  Bar: {
    name: 'Bar',
    props: ['data', 'options'],
    template: `
      <div>
        <div class="chart-data">{{ JSON.stringify(data) }}</div>
        <div class="chart-options">{{ JSON.stringify(options) }}</div>
      </div>
    `,
  },
  Line: {
    name: 'Line',
    props: ['data', 'options'],
    template: `
      <div>
        <div class="chart-data">{{ JSON.stringify(data) }}</div>
        <div class="chart-options">{{ JSON.stringify(options) }}</div>
      </div>
    `,
  },
}))

afterEach(() => {
  document.documentElement.classList.remove('dark')
  i18n.global.locale.value = 'en'
  vi.unstubAllGlobals()
})

describe('ModelUsageTrendChart', () => {
  const trendData = [
    { date: '2026-07-06', model: 'gpt-5.6-sol', is_other: false, requests: 8, total_tokens: 80, cost: 1.2, actual_cost: 1.1 },
    { date: '2026-07-07', model: 'gpt-5.6-sol', is_other: false, requests: 4, total_tokens: 40, cost: 0.4, actual_cost: 0.3 },
    { date: '2026-07-07', model: 'gpt-5.4', is_other: false, requests: 2, total_tokens: 20, cost: 0.2, actual_cost: 0.15 },
    { date: '2026-07-06', model: '', is_other: true, requests: 1, total_tokens: 10, cost: 0.1, actual_cost: 0.05 },
  ]

  it('builds cost datasets in canonical model order and fills missing buckets', () => {
    const wrapper = mount(ModelUsageTrendChart, {
      props: { trendData, metric: 'actual_cost', granularity: 'day' },
      global: { stubs: { LoadingSpinner: true } },
    })

    const chartData = JSON.parse(wrapper.find('.chart-data').text())
    expect(chartData.labels).toEqual(['2026-07-06', '2026-07-07'])
    expect(chartData.datasets.map((dataset: any) => dataset.label)).toEqual([
      'gpt-5.6-sol',
      'gpt-5.4',
      'Other',
    ])
    expect(chartData.datasets[1].data).toEqual([0, 0.15])
    expect(wrapper.text()).toContain('Spending distribution')
  })

  it('switches the y-axis metric and promotes large Chinese token values to 亿', async () => {
    i18n.global.locale.value = 'zh'
    const wrapper = mount(ModelUsageTrendChart, {
      props: { trendData, metric: 'actual_cost', granularity: 'day' },
      global: { stubs: { LoadingSpinner: true } },
    })

    expect(wrapper.get('[data-testid="model-trend-metric-actual_cost"]').attributes('aria-pressed')).toBe('true')

    await wrapper.get('[data-testid="model-trend-metric-total_tokens"]').trigger('click')

    const tokenData = JSON.parse(wrapper.find('.chart-data').text())
    const tokenOptions = wrapper.findComponent({ name: 'Line' }).props('options')
    expect(tokenData.datasets[0].data).toEqual([80, 40])
    expect(tokenOptions.scales.y.ticks.callback(123456789)).toBe('1.23亿')
    expect(tokenOptions.scales.y.ticks.callback(54321)).toBe('5.43w')
    expect(tokenOptions.plugins.tooltip.callbacks.label({
      dataset: { label: 'gpt-5.6-sol' },
      raw: 123456789,
    })).toBe('gpt-5.6-sol: 1.23亿')

    await wrapper.get('[data-testid="model-trend-metric-cost"]').trigger('click')

    const quotaData = JSON.parse(wrapper.find('.chart-data').text())
    expect(quotaData.datasets[0].data).toEqual([1.2, 0.4])
    expect(wrapper.get('[data-testid="model-trend-metric-cost"]').attributes('aria-pressed')).toBe('true')
  })

  it('places the model legend below the plot and switches between area and stacked bar charts', async () => {
    const wrapper = mount(ModelUsageTrendChart, {
      props: { trendData, metric: 'total_tokens', granularity: 'day' },
      global: { stubs: { LoadingSpinner: true, Icon: true } },
    })

    expect(wrapper.findComponent({ name: 'Line' }).exists()).toBe(true)
    expect(wrapper.findComponent({ name: 'Bar' }).exists()).toBe(false)
    expect(wrapper.findComponent({ name: 'Line' }).props('options').plugins.legend.position).toBe('bottom')
    expect(wrapper.get('[data-testid="model-trend-chart-area"]').attributes('aria-pressed')).toBe('true')

    await wrapper.get('[data-testid="model-trend-chart-bar"]').trigger('click')

    expect(wrapper.findComponent({ name: 'Line' }).exists()).toBe(false)
    expect(wrapper.findComponent({ name: 'Bar' }).exists()).toBe(true)
    const barOptions = wrapper.findComponent({ name: 'Bar' }).props('options')
    expect(barOptions.scales.x.stacked).toBe(true)
    expect(barOptions.scales.y.stacked).toBe(true)
    expect(barOptions.plugins.legend.position).toBe('bottom')
    expect(wrapper.get('[data-testid="model-trend-chart-bar"]').attributes('aria-pressed')).toBe('true')

    await wrapper.get('[data-testid="model-trend-chart-area"]').trigger('click')
    expect(wrapper.findComponent({ name: 'Line' }).exists()).toBe(true)
  })

  it('fills short hourly gaps and formats ticks from the category value index', () => {
    const hourlyData = [
      { date: '2026-07-12 08:00', model: 'gpt-5.6-sol', is_other: false, requests: 8, total_tokens: 80, cost: 1.2, actual_cost: 1.1 },
      { date: '2026-07-12 14:00', model: 'gpt-5.6-sol', is_other: false, requests: 4, total_tokens: 40, cost: 0.4, actual_cost: 0.3 },
    ]
    const wrapper = mount(ModelUsageTrendChart, {
      props: { trendData: hourlyData, metric: 'total_tokens', granularity: 'hour' },
      global: { stubs: { LoadingSpinner: true } },
    })

    const options = wrapper.findComponent({ name: 'Line' }).props('options') as {
      scales: { x: { ticks: { callback: (value: number, index: number) => string } } }
    }
    const chartData = JSON.parse(wrapper.find('.chart-data').text())

    expect(chartData.labels).toEqual([
      '2026-07-12 08:00',
      '2026-07-12 09:00',
      '2026-07-12 10:00',
      '2026-07-12 11:00',
      '2026-07-12 12:00',
      '2026-07-12 13:00',
      '2026-07-12 14:00',
    ])
    expect(chartData.datasets[0].data).toEqual([80, 0, 0, 0, 0, 0, 40])
    expect(options.scales.x.ticks.callback(6, 0)).toBe('14:00')
  })

  it('keeps model labels and colors stable across metrics with Other last', () => {
    const rankedByCost = [
      { date: '2026-07-06', model: 'gpt-a', is_other: false, requests: 5, total_tokens: 50, cost: 50, actual_cost: 50 },
      { date: '2026-07-06', model: 'gpt-b', is_other: false, requests: 50, total_tokens: 500, cost: 10, actual_cost: 10 },
      { date: '2026-07-06', model: '', is_other: true, requests: 20, total_tokens: 200, cost: 20, actual_cost: 20 },
    ]

    const costWrapper = mount(ModelUsageTrendChart, {
      props: { trendData: rankedByCost, metric: 'actual_cost', granularity: 'day' },
      global: { stubs: { LoadingSpinner: true } },
    })
    const tokenWrapper = mount(ModelUsageTrendChart, {
      props: { trendData: rankedByCost, metric: 'total_tokens', granularity: 'day' },
      global: { stubs: { LoadingSpinner: true } },
    })

    const costDatasets = JSON.parse(costWrapper.find('.chart-data').text()).datasets
    const tokenDatasets = JSON.parse(tokenWrapper.find('.chart-data').text()).datasets
    const projectIdentity = (dataset: any) => ({
      label: dataset.label,
      borderColor: dataset.borderColor,
    })

    expect(costDatasets.map(projectIdentity)).toEqual([
      { label: 'gpt-a', borderColor: '#7c3aed' },
      { label: 'gpt-b', borderColor: '#2563eb' },
      { label: 'Other', borderColor: '#06b6d4' },
    ])
    expect(tokenDatasets.map(projectIdentity)).toEqual(costDatasets.map(projectIdentity))
  })

  it('updates chart colors when the root theme class changes and disconnects on unmount', async () => {
    document.documentElement.classList.remove('dark')

    let observerCallback: MutationCallback | undefined
    const observe = vi.fn()
    const disconnect = vi.fn()

    class MutationObserverStub {
      observe = observe
      disconnect = disconnect

      constructor(callback: MutationCallback) {
        observerCallback = callback
      }
    }

    vi.stubGlobal('MutationObserver', MutationObserverStub)

    const wrapper = mount(ModelUsageTrendChart, {
      props: { trendData, metric: 'total_tokens', granularity: 'day' },
      global: { stubs: { LoadingSpinner: true } },
    })

    const lightOptions = JSON.parse(wrapper.find('.chart-options').text())
    expect(lightOptions.plugins.legend.labels.color).toBe('#374151')
    expect(lightOptions.scales.x.grid.color).toBe('#e5e7eb')
    expect(observerCallback).toBeDefined()
    expect(observe).toHaveBeenCalledWith(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    })

    document.documentElement.classList.add('dark')
    if (!observerCallback) {
      throw new Error('Expected the component to create a theme observer')
    }
    observerCallback([], {} as MutationObserver)
    await wrapper.vm.$nextTick()

    const darkOptions = JSON.parse(wrapper.find('.chart-options').text())
    expect(darkOptions.plugins.legend.labels.color).toBe('#e5e7eb')
    expect(darkOptions.scales.x.grid.color).toBe('#374151')

    wrapper.unmount()
    expect(disconnect).toHaveBeenCalledOnce()
  })

  it('renders loading and empty states without a chart', () => {
    const loading = mount(ModelUsageTrendChart, {
      props: { trendData: [], metric: 'total_tokens', granularity: 'day', loading: true },
      global: { stubs: { LoadingSpinner: true } },
    })
    expect(loading.find('.chart-data').exists()).toBe(false)
    expect(loading.findComponent({ name: 'LoadingSpinner' }).exists()).toBe(true)

    const empty = mount(ModelUsageTrendChart, {
      props: { trendData: [], metric: 'total_tokens', granularity: 'day' },
      global: { stubs: { LoadingSpinner: true } },
    })
    expect(empty.text()).toContain('No data available')
  })
})
