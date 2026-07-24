<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading && !stats" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else>
        <div
          v-if="dashboardError"
          class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300"
        >
          {{ t('dashboard.commercial.loadFailed') }}
        </div>

        <UserDashboardStats
          v-if="stats"
          :stats="stats"
          :balance="balance ?? 0"
          :is-simple="authStore.isSimpleMode"
        />
        <div v-else class="card p-6 text-center text-sm text-gray-500 dark:text-dark-400">
          {{ t('dashboard.noDataAvailable') }}
        </div>

        <UserDashboardFilters
          v-model:startDate="startDate"
          v-model:endDate="endDate"
          v-model:startTime="startTime"
          v-model:endTime="endTime"
          v-model:granularity="granularity"
          :loading="loadingTrends"
          @dateRangeChange="handleDateRangeChange"
          @granularityChange="loadModelTrends"
          @refresh="refreshAll"
        />

        <div
          data-testid="dashboard-content-grid"
          class="grid grid-cols-1 gap-6 lg:grid-cols-[minmax(0,5fr)_minmax(14rem,1fr)]"
        >
          <div class="min-w-0 space-y-6">
            <EntityDistributionChart
              data-testid="api-key-distribution"
              v-model:metric="apiKeyDistributionMetric"
              :title="t('usage.apiKeyDistribution')"
              :entity-label="t('usage.apiKeyFilter')"
              :items="apiKeyDistributionItems"
              :loading="loadingApiKeyStats"
              :show-account-cost="false"
              :show-standard-cost="true"
            />
            <UserDashboardModelTrends
              :trend-data="modelTrendData"
              :loading="loadingTrends"
              :granularity="granularity"
              :range-start="trendRangeStart"
              :range-end="trendRangeEnd"
              :timezone="dashboardTimezone"
            />
          </div>
          <div class="min-w-0">
            <UserDashboardAnnouncements />
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'
import UserDashboardFilters from '@/components/user/dashboard/UserDashboardFilters.vue'
import UserDashboardModelTrends from '@/components/user/dashboard/UserDashboardModelTrends.vue'
import UserDashboardAnnouncements from '@/components/user/dashboard/UserDashboardAnnouncements.vue'
import EntityDistributionChart from '@/components/charts/EntityDistributionChart.vue'
import { useAuthStore } from '@/stores'
import {
  usageAPI,
  type UsageDashboardSnapshotV2Response,
  type ModelUsageTrendPoint,
  type ModelUsageTrendGranularity,
  type UserDashboardStats as UserStatsType,
} from '@/api/usage'
import type { ApiKeyStat } from '@/types'
import {
  type DateRangeChange,
  getDashboardPresetGranularity,
  getDashboardPresetPeriod,
  toLocalDateTimeParam,
} from '@/utils/dashboardTimeRange'
import { formatDateLocalInput, formatTimeLocalInput } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const stats = ref<UserStatsType | null>(null)
const loading = ref(false)
const loadingTrends = ref(false)
const dashboardError = ref(false)

const modelTrendData = ref<ModelUsageTrendPoint[]>([])
const trendRangeStart = ref<string | null>(null)
const trendRangeEnd = ref<string | null>(null)
const apiKeyStats = ref<ApiKeyStat[]>([])
const loadingApiKeyStats = ref(false)
const apiKeyDistributionMetric = ref<'tokens' | 'actual_cost'>('actual_cost')
const dashboardTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone

const startDate = ref(formatDateLocalInput(new Date(Date.now() - 86400000)))
const endDate = ref(formatDateLocalInput(new Date()))
const startTime = ref(formatTimeLocalInput(new Date(Date.now() - 86400000)))
const endTime = ref(formatTimeLocalInput(new Date()))
const activePreset = ref<string | null>('last24Hours')
const granularity = ref<ModelUsageTrendGranularity>('hour')

const balance = computed(() => {
  const value = user.value?.balance
  return typeof value === 'number' && Number.isFinite(value) ? value : null
})

const getDashboardRangeParams = () => {
  const period = getDashboardPresetPeriod(activePreset.value)
  return period
    ? { period, timezone: dashboardTimezone }
    : {
        start_time: toLocalDateTimeParam(startDate.value, startTime.value),
        end_time: toLocalDateTimeParam(endDate.value, endTime.value),
        timezone: dashboardTimezone,
      }
}

const apiKeyDistributionItems = computed(() => apiKeyStats.value.map((item) => ({
  id: item.api_key_id,
  label: item.api_key_name?.trim() || t('usage.apiKeyFallback', { id: item.api_key_id }),
  requests: item.requests,
  total_tokens: item.total_tokens,
  cost: item.cost,
  actual_cost: item.actual_cost,
})))

const loadStats = async () => {
  loading.value = true
  dashboardError.value = false
  try {
    await authStore.refreshUser()
    stats.value = await usageAPI.getDashboardStats()
  } catch (error) {
    console.error('Failed to load dashboard stats:', error instanceof Error ? error.message : error)
    dashboardError.value = true
    stats.value = null
  } finally {
    loading.value = false
  }
}

const loadModelTrends = async () => {
  loadingTrends.value = true
  try {
    const response = await usageAPI.getDashboardModelTrend({
      ...getDashboardRangeParams(),
      granularity: granularity.value,
      model_source: 'requested',
    })
    modelTrendData.value = response.trend || []
    trendRangeStart.value = response.start_time
    trendRangeEnd.value = response.end_time
  } catch (error) {
    console.error('Failed to load model trends:', error instanceof Error ? error.message : error)
    modelTrendData.value = []
    trendRangeStart.value = null
    trendRangeEnd.value = null
  } finally {
    loadingTrends.value = false
  }
}

const loadApiKeyDistribution = async () => {
  loadingApiKeyStats.value = true
  try {
    const response: UsageDashboardSnapshotV2Response = await usageAPI.getDashboardSnapshotV2({
      ...getDashboardRangeParams(),
      include_trend: false,
      include_model_stats: false,
      include_group_stats: false,
      include_api_key_stats: true,
    })
    apiKeyStats.value = response.api_keys ?? []
  } catch (error) {
    console.error('Failed to load API key distribution:', error instanceof Error ? error.message : error)
    apiKeyStats.value = []
  } finally {
    loadingApiKeyStats.value = false
  }
}

const handleDateRangeChange = (range: DateRangeChange) => {
  activePreset.value = range.preset
  const defaultGranularity = getDashboardPresetGranularity(range.preset)
  if (defaultGranularity) granularity.value = defaultGranularity
  void loadModelTrends()
  void loadApiKeyDistribution()
}

const refreshAll = () => {
  loadStats()
  loadModelTrends()
  loadApiKeyDistribution()
}

onMounted(refreshAll)
</script>
