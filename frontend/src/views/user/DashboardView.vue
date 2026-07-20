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
          :platform-quotas="platformQuotas"
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
          <div class="min-w-0">
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
import { useAuthStore } from '@/stores'
import {
  usageAPI,
  type ModelUsageTrendPoint,
  type ModelUsageTrendGranularity,
  type UserDashboardStats as UserStatsType,
} from '@/api/usage'
import { getMyPlatformQuotas } from '@/api/user'
import type { PlatformQuotaItem } from '@/types'
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
const platformQuotas = ref<PlatformQuotaItem[] | null>(null)
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
    const period = getDashboardPresetPeriod(activePreset.value)
    const response = await usageAPI.getDashboardModelTrend({
      ...(period
        ? { period }
        : {
            start_time: toLocalDateTimeParam(startDate.value, startTime.value),
            end_time: toLocalDateTimeParam(endDate.value, endTime.value),
          }),
      granularity: granularity.value,
      model_source: 'requested',
      timezone: dashboardTimezone,
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

const handleDateRangeChange = (range: DateRangeChange) => {
  activePreset.value = range.preset
  const defaultGranularity = getDashboardPresetGranularity(range.preset)
  if (defaultGranularity) granularity.value = defaultGranularity
  void loadModelTrends()
}

const loadPlatformQuotas = async () => {
  try {
    const data = await getMyPlatformQuotas()
    platformQuotas.value = data.platform_quotas ?? []
  } catch (error) {
    console.warn('Failed to load platform quotas:', error instanceof Error ? error.message : error)
    platformQuotas.value = []
  }
}

const refreshAll = () => {
  loadStats()
  loadModelTrends()
  loadPlatformQuotas()
}

onMounted(refreshAll)
</script>
