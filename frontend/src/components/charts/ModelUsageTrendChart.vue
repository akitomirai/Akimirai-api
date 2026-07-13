<template>
  <section
    class="card relative overflow-hidden p-4"
    :data-testid="`model-usage-trend-${metric}`"
  >
    <div
      v-if="loading"
      class="absolute inset-0 z-10 flex items-center justify-center bg-white/60 backdrop-blur-sm dark:bg-dark-800/60"
    >
      <LoadingSpinner size="md" />
    </div>

    <div class="mb-3 flex min-h-9 flex-col items-start justify-between gap-2 sm:flex-row sm:items-center">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
        {{ title }}
      </h3>

      <div class="flex w-full flex-wrap items-center gap-2 sm:w-auto sm:justify-end">
        <div
          class="inline-flex shrink-0 rounded-md border border-gray-200 bg-gray-50 p-0.5 dark:border-dark-600 dark:bg-dark-700"
          role="group"
          :aria-label="t('dashboard.metricType')"
        >
          <button
            v-for="option in metricOptions"
            :key="option.value"
            type="button"
            class="inline-flex h-7 items-center rounded px-2 text-xs font-medium transition-colors"
            :class="selectedMetric === option.value
              ? 'bg-white text-primary-600 shadow-sm dark:bg-dark-600 dark:text-primary-400'
              : 'text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-100'"
            :aria-pressed="selectedMetric === option.value"
            :title="option.label"
            :data-testid="`model-trend-metric-${option.value}`"
            @click="selectedMetric = option.value"
          >
            {{ option.label }}
          </button>
        </div>

        <div
          class="inline-flex shrink-0 rounded-md border border-gray-200 bg-gray-50 p-0.5 dark:border-dark-600 dark:bg-dark-700"
          role="group"
          :aria-label="t('dashboard.chartType')"
        >
          <button
            type="button"
            class="inline-flex h-7 items-center gap-1 rounded px-2 text-xs font-medium transition-colors"
            :class="chartType === 'bar'
              ? 'bg-white text-primary-600 shadow-sm dark:bg-dark-600 dark:text-primary-400'
              : 'text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-100'"
            :aria-pressed="chartType === 'bar'"
            :title="t('dashboard.barChart')"
            data-testid="model-trend-chart-bar"
            @click="chartType = 'bar'"
          >
            <Icon name="chartBar" size="xs" />
            <span>{{ t('dashboard.barChart') }}</span>
          </button>
          <button
            type="button"
            class="inline-flex h-7 items-center gap-1 rounded px-2 text-xs font-medium transition-colors"
            :class="chartType === 'area'
              ? 'bg-white text-primary-600 shadow-sm dark:bg-dark-600 dark:text-primary-400'
              : 'text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-100'"
            :aria-pressed="chartType === 'area'"
            :title="t('dashboard.areaChart')"
            data-testid="model-trend-chart-area"
            @click="chartType = 'area'"
          >
            <Icon name="trendingUp" size="xs" />
            <span>{{ t('dashboard.areaChart') }}</span>
          </button>
        </div>
      </div>
    </div>

    <div v-if="chartData" class="h-72 sm:h-80">
      <Bar
        v-if="chartType === 'bar'"
        :data="chartData"
        :options="chartOptions"
      />
      <Line
        v-else
        :data="chartData"
        :options="chartOptions"
      />
    </div>
    <div
      v-else
      class="flex h-64 items-center justify-center text-sm text-gray-500 dark:text-gray-400 sm:h-72"
    >
      {{ t('dashboard.noDataAvailable') }}
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Bar, Line } from 'vue-chartjs'
import {
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
} from 'chart.js'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import type { ModelUsageTrendGranularity, ModelUsageTrendPoint } from '@/api/usage'
import { formatCostFixed, formatTokenCount } from '@/utils/format'
import {
  completeModelUsageBuckets,
  createModelUsageTickFormatter,
} from './modelUsageTrendAxis'

ChartJS.register(
  BarElement,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler,
)

type Metric = 'total_tokens' | 'cost' | 'actual_cost'
type ChartType = 'bar' | 'area'

const props = withDefaults(defineProps<{
  trendData: ModelUsageTrendPoint[]
  metric: Metric
  granularity: ModelUsageTrendGranularity
  rangeStart?: string | null
  rangeEnd?: string | null
  timezone?: string
  loading?: boolean
}>(), {
  loading: false,
})

const { t } = useI18n()
const chartType = ref<ChartType>('area')
const selectedMetric = ref<Metric>(props.metric)

const metricOptions = computed<Array<{ value: Metric; label: string }>>(() => [
  { value: 'total_tokens', label: t('dashboard.tokens') },
  { value: 'cost', label: t('dashboard.quotaConsumption') },
  { value: 'actual_cost', label: t('dashboard.actualConsumption') },
])

const colors = [
  '#7c3aed',
  '#2563eb',
  '#06b6d4',
  '#f97316',
  '#10b981',
  '#ec4899',
  '#eab308',
  '#64748b',
  '#94a3b8',
]

const title = computed(() => props.metric === 'actual_cost'
  ? t('dashboard.spendingDistribution')
  : t('dashboard.tokenUsageTrend'))

const readDarkMode = (): boolean => typeof document !== 'undefined'
  && document.documentElement.classList.contains('dark')

const isDarkMode = ref(readDarkMode())
let themeObserver: MutationObserver | null = null

onMounted(() => {
  isDarkMode.value = readDarkMode()
  if (typeof document === 'undefined' || typeof MutationObserver === 'undefined') return

  themeObserver = new MutationObserver(() => {
    isDarkMode.value = readDarkMode()
  })
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class'],
  })
})

onBeforeUnmount(() => {
  themeObserver?.disconnect()
  themeObserver = null
})

const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb',
}))

const groupedData = computed(() => {
  if (!props.trendData?.length) return null

  const dates = new Set<string>()
  const groups = new Map<string, {
    label: string
    isOther: boolean
    values: Map<string, number>
  }>()

  for (const point of props.trendData) {
    dates.add(point.date)
    const key = point.is_other ? '__other__' : point.model
    const group = groups.get(key) ?? {
      label: point.is_other ? t('dashboard.other') : point.model,
      isOther: point.is_other,
      values: new Map<string, number>(),
    }
    const value = point[selectedMetric.value]
    group.values.set(point.date, (group.values.get(point.date) ?? 0) + value)
    groups.set(key, group)
  }

  const orderedGroups = Array.from(groups.values())
    .sort((a, b) => Number(a.isOther) - Number(b.isOther))

  return {
    dates: completeModelUsageBuckets(
      Array.from(dates),
      props.granularity,
      props.rangeStart,
      props.rangeEnd,
    ),
    groups: orderedGroups,
  }
})

const chartData = computed(() => {
  if (!groupedData.value) return null

  return {
    labels: groupedData.value.dates,
    datasets: groupedData.value.groups.map((group, index) => {
      const color = colors[index % colors.length]
      return {
        label: group.label,
        data: groupedData.value!.dates.map((date) => group.values.get(date) ?? 0),
        borderColor: color,
        backgroundColor: chartType.value === 'bar' ? `${color}cc` : `${color}20`,
        borderWidth: chartType.value === 'bar' ? 1 : 2,
        borderRadius: chartType.value === 'bar' ? 2 : 0,
        maxBarThickness: 36,
        pointRadius: chartType.value === 'area' ? 0 : undefined,
        pointHoverRadius: chartType.value === 'area' ? 4 : undefined,
        fill: chartType.value === 'area',
        tension: chartType.value === 'area' ? 0.3 : 0,
      }
    }),
  }
})

const formatXAxisTick = computed(() => createModelUsageTickFormatter(
  groupedData.value?.dates ?? [],
  props.granularity,
  props.timezone,
))

const formatValue = (value: number): string => selectedMetric.value === 'total_tokens'
  ? formatTokenCount(value)
  : `$${formatCostFixed(value, 4)}`

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const,
  },
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 14,
        font: { size: 10 },
      },
    },
    tooltip: {
      itemSort: (a: any, b: any) => Number(b.raw ?? 0) - Number(a.raw ?? 0),
      callbacks: {
        title: (tooltipItems: any[]) => formatXAxisTick.value(tooltipItems[0]?.dataIndex ?? -1),
        label: (context: any) => `${context.dataset.label}: ${formatValue(Number(context.raw ?? 0))}`,
      },
    },
  },
  scales: {
    x: {
      stacked: chartType.value === 'bar',
      grid: { color: chartColors.value.grid },
      ticks: {
        color: chartColors.value.text,
        font: { size: 10 },
        maxRotation: 0,
        autoSkip: true,
        callback: (value: string | number) => {
          const labelIndex = typeof value === 'number' ? value : Number(value)
          return formatXAxisTick.value(Number.isInteger(labelIndex) ? labelIndex : -1)
        },
      },
    },
    y: {
      beginAtZero: true,
      stacked: chartType.value === 'bar',
      grid: { color: chartColors.value.grid },
      ticks: {
        color: chartColors.value.text,
        font: { size: 10 },
        callback: (value: string | number) => formatValue(Number(value)),
      },
    },
  },
}))
</script>
