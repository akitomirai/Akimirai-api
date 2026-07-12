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

    <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">
      {{ title }}
    </h3>

    <div v-if="chartData" class="h-64 sm:h-72">
      <Line :data="chartData" :options="chartOptions" />
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
import { Line } from 'vue-chartjs'
import {
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
import type { ModelUsageTrendPoint } from '@/api/usage'
import { formatCostFixed, formatNumberLocaleString } from '@/utils/format'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

type Metric = 'actual_cost' | 'requests'

const props = withDefaults(defineProps<{
  trendData: ModelUsageTrendPoint[]
  metric: Metric
  loading?: boolean
}>(), {
  loading: false,
})

const { t } = useI18n()

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
  : t('dashboard.callTrend'))

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
    const value = props.metric === 'actual_cost' ? point.actual_cost : point.requests
    group.values.set(point.date, (group.values.get(point.date) ?? 0) + value)
    groups.set(key, group)
  }

  const orderedGroups = Array.from(groups.values())
    .sort((a, b) => Number(a.isOther) - Number(b.isOther))

  return {
    dates: Array.from(dates).sort(),
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
        backgroundColor: `${color}20`,
        borderWidth: 2,
        pointRadius: 0,
        pointHoverRadius: 4,
        fill: true,
        tension: 0.3,
      }
    }),
  }
})

const formatValue = (value: number): string => props.metric === 'actual_cost'
  ? `$${formatCostFixed(value, 4)}`
  : formatNumberLocaleString(value)

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const,
  },
  plugins: {
    legend: {
      position: 'top' as const,
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
        label: (context: any) => `${context.dataset.label}: ${formatValue(Number(context.raw ?? 0))}`,
      },
    },
  },
  scales: {
    x: {
      grid: { color: chartColors.value.grid },
      ticks: {
        color: chartColors.value.text,
        font: { size: 10 },
        maxRotation: 0,
        autoSkip: true,
      },
    },
    y: {
      beginAtZero: true,
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
