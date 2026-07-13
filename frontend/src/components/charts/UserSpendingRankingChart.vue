<template>
  <div class="card p-4" data-testid="user-spending-ranking-chart">
    <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">
      {{ t('admin.dashboard.spendingRankingTitle') }}
    </h3>

    <div v-if="loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div
      v-else-if="error"
      class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('admin.dashboard.failedToLoad') }}
    </div>
    <div
      v-else-if="displayItems.length > 0 && chartData"
      class="flex flex-col items-center gap-6 sm:flex-row"
    >
      <div class="h-48 w-48 shrink-0">
        <Doughnut :data="chartData" :options="doughnutOptions" />
      </div>
      <div class="max-h-48 w-full min-w-0 flex-1 overflow-auto">
        <table class="w-full min-w-[28rem] text-xs">
          <thead>
            <tr class="text-gray-500 dark:text-gray-400">
              <th class="pb-2 text-left">{{ t('admin.dashboard.spendingRankingUser') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.spendingRankingRequests') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.spendingRankingTokens') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.spendingRankingSpend') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(item, index) in displayItems"
              :key="item.isOther ? 'others' : `${item.user_id}-${index}`"
              class="border-t border-gray-100 transition-colors dark:border-gray-700"
              :class="
                item.isOther
                  ? 'bg-gray-50/70 dark:bg-dark-700/20'
                  : 'cursor-pointer hover:bg-gray-50 dark:hover:bg-dark-700/40'
              "
              @click="item.isOther ? undefined : emit('select-user', item)"
            >
              <td class="py-1.5">
                <div class="flex min-w-0 items-center gap-2">
                  <span class="shrink-0 text-[11px] font-semibold text-gray-500 dark:text-gray-400">
                    {{ item.isOther ? '-' : `#${index + 1}` }}
                  </span>
                  <span
                    class="block max-w-[140px] truncate font-medium text-gray-900 dark:text-white"
                    :title="getRowLabel(item)"
                  >
                    {{ getRowLabel(item) }}
                  </span>
                </div>
              </td>
              <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                {{ formatNumber(item.requests) }}
              </td>
              <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                {{ formatTokenCount(item.tokens) }}
              </td>
              <td class="py-1.5 text-right text-green-600 dark:text-green-400">
                ${{ formatCost(item.actual_cost) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
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
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Doughnut } from 'vue-chartjs'

import type { UserSpendingRankingItem } from '@/types'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { formatTokenCount } from '@/utils/format'

ChartJS.register(ArcElement, Tooltip, Legend)

type RankingDisplayItem = UserSpendingRankingItem & { isOther?: boolean }

const props = withDefaults(
  defineProps<{
    items: UserSpendingRankingItem[]
    totalActualCost?: number
    totalRequests?: number
    totalTokens?: number
    loading?: boolean
    error?: boolean
  }>(),
  {
    totalActualCost: 0,
    totalRequests: 0,
    totalTokens: 0,
    loading: false,
    error: false
  }
)

const emit = defineEmits<{
  'select-user': [item: UserSpendingRankingItem]
}>()

const { t } = useI18n()

const chartColors = [
  '#3b82f6',
  '#10b981',
  '#f59e0b',
  '#ef4444',
  '#8b5cf6',
  '#ec4899',
  '#14b8a6',
  '#f97316',
  '#6366f1',
  '#84cc16',
  '#06b6d4',
  '#a855f7'
]

const toFiniteNumber = (value: unknown): number => {
  const numericValue = Number(value)
  return Number.isFinite(numericValue) ? numericValue : 0
}

const otherItem = computed<RankingDisplayItem | null>(() => {
  if (!props.items.length) return null

  const rankedActualCost = props.items.reduce((sum, item) => sum + toFiniteNumber(item.actual_cost), 0)
  const rankedRequests = props.items.reduce((sum, item) => sum + toFiniteNumber(item.requests), 0)
  const rankedTokens = props.items.reduce((sum, item) => sum + toFiniteNumber(item.tokens), 0)

  const actualCost = Math.max(props.totalActualCost - rankedActualCost, 0)
  const requests = Math.max(props.totalRequests - rankedRequests, 0)
  const tokens = Math.max(props.totalTokens - rankedTokens, 0)

  if (actualCost <= 0.000001 && requests <= 0 && tokens <= 0) return null

  return {
    user_id: 0,
    email: '',
    actual_cost: actualCost,
    requests,
    tokens,
    isOther: true
  }
})

const displayItems = computed<RankingDisplayItem[]>(() => {
  return otherItem.value ? [...props.items, otherItem.value] : [...props.items]
})

const getUserLabel = (item: UserSpendingRankingItem): string => {
  return item.email || t('admin.redeem.userPrefix', { id: item.user_id })
}

const getRowLabel = (item: RankingDisplayItem): string => {
  return item.isOther ? t('admin.dashboard.spendingRankingOther') : getUserLabel(item)
}

const chartData = computed(() => {
  if (!props.items.length) return null

  const labels = props.items.map((item, index) => `#${index + 1} ${getUserLabel(item)}`)
  const data = props.items.map((item) => toFiniteNumber(item.actual_cost))
  const backgroundColor = chartColors.slice(0, props.items.length)

  if (otherItem.value) {
    labels.push(t('admin.dashboard.spendingRankingOther'))
    data.push(otherItem.value.actual_cost)
    backgroundColor.push('#94a3b8')
  }

  return {
    labels,
    datasets: [{ data, backgroundColor, borderWidth: 0 }]
  }
})

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const value = context.raw as number
          const total = context.dataset.data.reduce((sum: number, item: number) => sum + item, 0)
          const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0'
          return `${context.label}: $${formatCost(value)} (${percentage}%)`
        }
      }
    }
  }
}))

const formatNumber = (value: number): string => toFiniteNumber(value).toLocaleString()

const formatCost = (value: number | null | undefined): string => {
  const amount = typeof value === 'number' && Number.isFinite(value) ? value : 0
  if (amount >= 1000) return `${(amount / 1000).toFixed(2)}K`
  if (amount >= 1) return amount.toFixed(2)
  if (amount >= 0.01) return amount.toFixed(3)
  return amount.toFixed(4)
}
</script>
