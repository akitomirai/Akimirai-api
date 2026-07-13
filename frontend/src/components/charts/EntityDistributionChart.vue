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
    <div v-else-if="displayItems.length > 0 && chartData" class="flex flex-col gap-5 sm:flex-row sm:items-center">
      <div class="mx-auto h-44 w-44 shrink-0 sm:mx-0 sm:h-48 sm:w-48">
        <Doughnut :data="chartData" :options="doughnutOptions" />
      </div>
      <div class="max-h-52 min-w-0 flex-1 overflow-auto">
        <table class="w-full min-w-[480px] text-xs">
          <thead>
            <tr class="text-gray-500 dark:text-gray-400">
              <th class="pb-2 text-left">{{ entityLabel }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.requests') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.tokens') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.actual') }}</th>
              <th v-if="showAccountCost" class="pb-2 text-right">{{ t('admin.dashboard.accountCost') }}</th>
              <th v-if="showStandardCost" class="pb-2 text-right">{{ t('admin.dashboard.standard') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in displayItems"
              :key="item.id"
              class="border-t border-gray-100 dark:border-gray-700"
            >
              <td class="max-w-[180px] truncate py-1.5 font-medium text-gray-900 dark:text-white" :title="item.label">
                {{ item.label }}
              </td>
              <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">{{ formatNumber(item.requests) }}</td>
              <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">{{ formatTokenCount(item.total_tokens) }}</td>
              <td class="py-1.5 text-right text-green-600 dark:text-green-400">${{ formatCost(item.actual_cost) }}</td>
              <td v-if="showAccountCost" class="py-1.5 text-right text-orange-500 dark:text-orange-400">
                ${{ formatCost(item.account_cost) }}
              </td>
              <td v-if="showStandardCost" class="py-1.5 text-right text-gray-400 dark:text-gray-500">
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
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArcElement, Chart as ChartJS, Legend, Tooltip } from 'chart.js'
import type { TooltipItem } from 'chart.js'
import { Doughnut } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { formatTokenCount } from '@/utils/format'

ChartJS.register(ArcElement, Tooltip, Legend)

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

const props = withDefaults(defineProps<{
  title: string
  entityLabel: string
  items: EntityDistributionItem[]
  loading?: boolean
  metric?: DistributionMetric
  showMetricToggle?: boolean
  showStandardCost?: boolean
  showAccountCost?: boolean
}>(), {
  loading: false,
  metric: 'tokens',
  showMetricToggle: true,
  showStandardCost: true,
  showAccountCost: false,
})

const emit = defineEmits<{
  'update:metric': [value: DistributionMetric]
}>()

const { t } = useI18n()
const chartColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316', '#6366f1', '#84cc16']
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

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: TooltipItem<'doughnut'>) => {
          const value = toFiniteNumber(context.raw)
          const total = context.dataset.data.reduce<number>((sum, item) => sum + toFiniteNumber(item), 0)
          const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0'
          const formatted = props.metric === 'actual_cost' ? `$${formatCost(value)}` : formatTokenCount(value)
          return `${context.label || ''}: ${formatted} (${percentage}%)`
        },
      },
    },
  },
}))

const formatNumber = (value: number): string => toFiniteNumber(value).toLocaleString()
const formatCost = (value: number | null | undefined): string => {
  const amount = toFiniteNumber(value)
  if (amount >= 1000) return `${(amount / 1000).toFixed(2)}K`
  if (amount >= 1) return amount.toFixed(2)
  if (amount >= 0.01) return amount.toFixed(3)
  return amount.toFixed(4)
}
</script>
