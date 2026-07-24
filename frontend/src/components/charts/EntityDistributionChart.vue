<template>
  <div class="card p-4">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ title }}</h3>
      <div
        v-if="showMetricToggle"
        class="inline-flex rounded-lg border border-gray-200 bg-gray-50 p-0.5 dark:border-gray-700 dark:bg-dark-800"
      >
        <button
          type="button"
          class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
          :class="metric === 'tokens'
            ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
            : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
          @click="emit('update:metric', 'tokens')"
        >
          {{ t('admin.dashboard.metricTokens') }}
        </button>
        <button
          type="button"
          class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
          :class="metric === 'actual_cost'
            ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
            : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
          @click="emit('update:metric', 'actual_cost')"
        >
          {{ t('admin.dashboard.metricActualCost') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="displayItems.length > 0 && chartData" class="flex flex-col gap-5 sm:flex-row sm:items-start">
      <div
        v-if="visualization === 'horizontal-bar' && barChartData"
        ref="barScrollRef"
        data-testid="horizontal-bar-scroll"
        class="entity-distribution-bar-scroll max-h-52 w-full shrink-0 overflow-x-hidden overflow-y-auto focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500/40 sm:w-72"
        :aria-label="title"
        role="region"
        tabindex="0"
        @scroll="syncVerticalScroll($event, tableScrollRef)"
      >
        <div aria-hidden="true" class="h-6" data-testid="horizontal-bar-header-spacer" />
        <div
          data-testid="horizontal-bar-canvas"
          class="relative w-full"
          :style="{ height: `${barChartHeight}px` }"
        >
          <Bar
            :data="barChartData"
            :options="barChartOptions"
            :plugins="[barValueLabelsPlugin]"
            role="img"
            :aria-label="`${title}: ${entityLabel}`"
          />
        </div>
      </div>
      <div v-else class="mx-auto h-44 w-44 shrink-0 sm:mx-0 sm:h-48 sm:w-48">
        <Doughnut :data="chartData" :options="doughnutOptions" />
      </div>
      <div
        ref="tableScrollRef"
        data-testid="distribution-table-scroll"
        class="max-h-52 min-w-0 flex-1 overflow-auto"
        @scroll="syncVerticalScroll($event, barScrollRef)"
      >
        <table class="w-full min-w-[480px] text-xs">
          <thead>
            <tr class="h-6 text-gray-500 dark:text-gray-400">
              <th class="p-0 text-left">{{ entityLabel }}</th>
              <th class="p-0 text-right">{{ t('admin.dashboard.requests') }}</th>
              <th class="p-0 text-right">{{ t('admin.dashboard.tokens') }}</th>
              <th class="p-0 text-right">{{ t('admin.dashboard.actual') }}</th>
              <th v-if="showAccountCost" class="p-0 text-right">{{ t('admin.dashboard.accountCost') }}</th>
              <th v-if="showStandardCost" class="p-0 text-right">{{ t('admin.dashboard.standard') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in displayItems"
              :key="item.id"
              class="h-[30px] border-t border-gray-100 dark:border-gray-700"
            >
              <td class="max-w-[180px] truncate p-0 font-medium text-gray-900 dark:text-white" :title="item.label">
                {{ item.label }}
              </td>
              <td class="p-0 text-right text-gray-600 dark:text-gray-400">{{ formatNumber(item.requests) }}</td>
              <td class="p-0 text-right text-gray-600 dark:text-gray-400">{{ formatTokenCount(item.total_tokens) }}</td>
              <td class="p-0 text-right text-green-600 dark:text-green-400">${{ formatCost(item.actual_cost) }}</td>
              <td v-if="showAccountCost" class="p-0 text-right text-orange-500 dark:text-orange-400">
                ${{ formatCost(item.account_cost) }}
              </td>
              <td v-if="showStandardCost" class="p-0 text-right text-gray-400 dark:text-gray-500">
                ${{ formatCost(item.cost) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div v-else class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ArcElement,
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  Tooltip,
} from 'chart.js'
import type { ChartOptions, Plugin, TooltipItem } from 'chart.js'
import { Bar, Doughnut } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { formatTokenCount } from '@/utils/format'

ChartJS.register(ArcElement, BarElement, CategoryScale, LinearScale, Tooltip, Legend)

interface EntityDistributionItem {
  id: string | number
  label: string
  requests: number
  total_tokens: number
  cost: number
  actual_cost: number
  account_cost?: number
}

type DistributionMetric = 'tokens' | 'actual_cost'
type DistributionVisualization = 'doughnut' | 'horizontal-bar'

const props = withDefaults(defineProps<{
  title: string
  entityLabel: string
  items: EntityDistributionItem[]
  loading?: boolean
  metric?: DistributionMetric
  showMetricToggle?: boolean
  showStandardCost?: boolean
  showAccountCost?: boolean
  visualization?: DistributionVisualization
}>(), {
  loading: false,
  metric: 'tokens',
  showMetricToggle: true,
  showStandardCost: true,
  showAccountCost: false,
  visualization: 'doughnut',
})

const emit = defineEmits<{
  'update:metric': [value: DistributionMetric]
}>()

const { t } = useI18n()
const chartColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316', '#6366f1', '#84cc16']
const BAR_ROW_HEIGHT = 30
const barScrollRef = ref<HTMLElement | null>(null)
const tableScrollRef = ref<HTMLElement | null>(null)
const toFiniteNumber = (value: unknown): number => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

const displayItems = computed(() => {
  const metricKey = props.metric === 'actual_cost' ? 'actual_cost' : 'total_tokens'
  return [...props.items].sort((a, b) => toFiniteNumber(b[metricKey]) - toFiniteNumber(a[metricKey]))
})

const chartData = computed(() => {
  if (displayItems.value.length === 0) return null
  return {
    labels: displayItems.value.map((item) => item.label),
    datasets: [{
      data: displayItems.value.map((item) => toFiniteNumber(props.metric === 'actual_cost' ? item.actual_cost : item.total_tokens)),
      backgroundColor: displayItems.value.map((_, index) => chartColors[index % chartColors.length]),
      borderWidth: 0,
    }],
  }
})

const barChartData = computed(() => {
  if (!chartData.value) return null
  return {
    ...chartData.value,
    datasets: chartData.value.datasets.map((dataset) => ({
      ...dataset,
      borderRadius: 4,
      borderSkipped: false as const,
      barThickness: 14,
      maxBarThickness: 14,
    })),
  }
})

const barChartHeight = computed(() => displayItems.value.length * BAR_ROW_HEIGHT)

const formatNumber = (value: number): string => toFiniteNumber(value).toLocaleString()
const formatCost = (value: number | null | undefined): string => {
  const amount = toFiniteNumber(value)
  if (amount >= 1000) return `${(amount / 1000).toFixed(2)}K`
  if (amount >= 1) return amount.toFixed(2)
  if (amount >= 0.01) return amount.toFixed(3)
  return amount.toFixed(4)
}

const formatMetricValue = (value: unknown): string => {
  const amount = toFiniteNumber(value)
  return props.metric === 'actual_cost' ? `$${formatCost(amount)}` : formatTokenCount(amount)
}

const formatTooltipLabel = (label: string, value: unknown, values: readonly unknown[]): string => {
  const amount = toFiniteNumber(value)
  const total = values.reduce<number>((sum, item) => sum + toFiniteNumber(item), 0)
  const percentage = total > 0 ? ((amount / total) * 100).toFixed(1) : '0.0'
  return `${label}: ${formatMetricValue(amount)} (${percentage}%)`
}

const isDarkMode = (): boolean => {
  return typeof document !== 'undefined' && document.documentElement.classList.contains('dark')
}

const getChartTextColor = (): string => isDarkMode() ? '#9ca3af' : '#6b7280'
const getChartGridColor = (): string => isDarkMode()
  ? 'rgba(148, 163, 184, 0.16)'
  : 'rgba(148, 163, 184, 0.22)'

const syncVerticalScroll = (event: Event, target: HTMLElement | null): void => {
  if (!target) return
  const source = event.currentTarget as HTMLElement
  if (target.scrollTop !== source.scrollTop) target.scrollTop = source.scrollTop
}

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: TooltipItem<'doughnut'>) => {
          return formatTooltipLabel(context.label || '', context.raw, context.dataset.data)
        },
      },
    },
  },
}))

const barChartOptions = computed<ChartOptions<'bar'>>(() => ({
  indexAxis: 'y',
  responsive: true,
  maintainAspectRatio: false,
  animation: {
    duration: 250,
  },
  layout: {
    padding: {
      right: 4,
    },
  },
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: TooltipItem<'bar'>) => {
          return formatTooltipLabel(context.label || '', context.raw, context.dataset.data)
        },
      },
    },
  },
  scales: {
    x: {
      beginAtZero: true,
      border: { display: false },
      grid: { color: () => getChartGridColor() },
      ticks: { display: false },
    },
    y: {
      border: { display: false },
      grid: { display: false },
      ticks: {
        display: false,
        autoSkip: false,
      },
    },
  },
}))

const barValueLabelsPlugin: Plugin<'bar'> = {
  id: 'entityDistributionValueLabels',
  afterDatasetsDraw(chart) {
    const dataset = chart.data.datasets[0]
    if (!dataset) return

    const meta = chart.getDatasetMeta(0)
    const { ctx, chartArea } = chart
    ctx.save()
    ctx.font = '500 10px ui-sans-serif, system-ui, sans-serif'
    ctx.textBaseline = 'middle'

    meta.data.forEach((element, index) => {
      const value = dataset.data[index]
      const label = formatMetricValue(value)
      const { x, y } = element.getProps(['x', 'y'], true) as { x: number; y: number }
      const labelWidth = ctx.measureText(label).width
      const renderInside = x + labelWidth + 6 > chartArea.right

      ctx.textAlign = renderInside ? 'right' : 'left'
      ctx.fillStyle = renderInside ? '#ffffff' : getChartTextColor()
      ctx.fillText(label, renderInside ? x - 4 : x + 4, y)
    })

    ctx.restore()
  },
}
</script>

<style scoped>
.entity-distribution-bar-scroll {
  -ms-overflow-style: none;
  scrollbar-width: none !important;
}

.entity-distribution-bar-scroll::-webkit-scrollbar {
  display: none;
  width: 0 !important;
  height: 0 !important;
}
</style>
