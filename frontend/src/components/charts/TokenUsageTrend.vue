<template>
  <div class="card p-4">
    <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">
      {{ t('admin.dashboard.tokenUsageTrend') }}
    </h3>
    <div v-if="loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="trendData.length > 0 && chartData" class="h-48">
      <Line :data="chartData" :options="lineOptions" />
    </div>
    <div
      v-else
      class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { TrendDataPoint } from '@/types'
import type { ModelUsageTrendGranularity } from '@/api/usage'
import { formatTokenCount } from '@/utils/format'
import {
  completeModelUsageBuckets,
  createModelUsageTickFormatter
} from './modelUsageTrendAxis'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const { t } = useI18n()

const props = withDefaults(defineProps<{
  trendData: TrendDataPoint[]
  loading?: boolean
  granularity?: ModelUsageTrendGranularity
  rangeStart?: string | null
  rangeEnd?: string | null
  timezone?: string
}>(), {
  granularity: 'hour',
  rangeStart: null,
  rangeEnd: null
})

const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb',
  input: '#3b82f6',
  output: '#10b981',
  cacheCreation: '#f59e0b',
  cacheRead: '#06b6d4',
  cacheHitRate: '#8b5cf6'
}))

const normalizedTrendData = computed<TrendDataPoint[]>(() => {
  if (!props.trendData?.length) return []
  const sourceDates = props.trendData.map((point) => point.date)
  const completedDates = completeModelUsageBuckets(
    sourceDates,
    props.granularity,
    props.rangeStart,
    props.rangeEnd
  )
  const completedDateSet = new Set(completedDates)
  if (!sourceDates.every((date) => completedDateSet.has(date))) return props.trendData

  const byDate = new Map(props.trendData.map((point) => [point.date, point]))
  return completedDates.map((date) => byDate.get(date) ?? {
    date,
    requests: 0,
    input_tokens: 0,
    output_tokens: 0,
    cache_creation_tokens: 0,
    cache_read_tokens: 0,
    total_tokens: 0,
    cost: 0,
    actual_cost: 0
  })
})

const chartData = computed(() => {
  if (!normalizedTrendData.value.length) return null

  const points = normalizedTrendData.value

  return {
    labels: points.map((d) => d.date),
    datasets: [
      {
        label: 'Input',
        data: points.map((d) => d.input_tokens),
        borderColor: chartColors.value.input,
        backgroundColor: `${chartColors.value.input}20`,
        fill: true,
        tension: 0.3
      },
      {
        label: 'Output',
        data: points.map((d) => d.output_tokens),
        borderColor: chartColors.value.output,
        backgroundColor: `${chartColors.value.output}20`,
        fill: true,
        tension: 0.3
      },
      {
        label: 'Cache Creation',
        data: points.map((d) => d.cache_creation_tokens),
        borderColor: chartColors.value.cacheCreation,
        backgroundColor: `${chartColors.value.cacheCreation}20`,
        fill: true,
        tension: 0.3
      },
      {
        label: 'Cache Read',
        data: points.map((d) => d.cache_read_tokens),
        borderColor: chartColors.value.cacheRead,
        backgroundColor: `${chartColors.value.cacheRead}20`,
        fill: true,
        tension: 0.3
      },
      {
        label: 'Cache Hit Rate',
        data: points.map((d) => {
          const totalPromptTokens = d.input_tokens + d.cache_read_tokens + d.cache_creation_tokens
          return totalPromptTokens > 0 ? (d.cache_read_tokens / totalPromptTokens) * 100 : 0
        }),
        borderColor: chartColors.value.cacheHitRate,
        backgroundColor: `${chartColors.value.cacheHitRate}20`,
        borderDash: [5, 5],
        fill: false,
        tension: 0.3,
        yAxisID: 'yPercent'
      }
    ]
  }
})

const tickFormatter = computed(() => createModelUsageTickFormatter(
  normalizedTrendData.value.map((point) => point.date),
  props.granularity,
  props.timezone
))

const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 15,
        font: {
          size: 11
        }
      }
    },
    tooltip: {
      callbacks: {
        title: (tooltipItems: any[]) => tickFormatter.value(tooltipItems[0]?.dataIndex ?? -1),
        label: (context: any) => {
          if (context.dataset.yAxisID === 'yPercent') {
            return `${context.dataset.label}: ${context.raw.toFixed(1)}%`
          }
          return `${context.dataset.label}: ${formatTokenCount(context.raw)}`
        },
        footer: (tooltipItems: any) => {
          const dataIndex = tooltipItems[0]?.dataIndex
          if (dataIndex !== undefined && normalizedTrendData.value[dataIndex]) {
            const data = normalizedTrendData.value[dataIndex]
            return `Actual: $${formatCost(data.actual_cost)} | Standard: $${formatCost(data.cost)}`
          }
          return ''
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (_value: string | number, index: number) => tickFormatter.value(index)
      }
    },
    y: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (value: string | number) => formatTokenCount(Number(value))
      }
    },
    yPercent: {
      position: 'right' as const,
      min: 0,
      max: 100,
      grid: {
        drawOnChartArea: false
      },
      ticks: {
        color: chartColors.value.cacheHitRate,
        font: {
          size: 10
        },
        callback: (value: string | number) => `${value}%`
      }
    }
  }
}))

const formatCost = (value: number): string => {
  if (value >= 1000) {
    return (value / 1000).toFixed(2) + 'K'
  } else if (value >= 1) {
    return value.toFixed(2)
  } else if (value >= 0.01) {
    return value.toFixed(3)
  }
  return value.toFixed(4)
}
</script>
