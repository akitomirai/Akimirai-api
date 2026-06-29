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

        <UserDashboardCommercialPath
          :stats="stats"
          :balance="balance"
          :api-keys="apiKeys"
          :api-keys-loading="loadingApiKeys"
          :api-keys-error="apiKeysError"
          :recent-errors="recentErrors"
          :errors-loading="loadingErrors"
          :errors-error="errorsError"
          :error-view-enabled="errorViewEnabled"
          :base-url="baseUrl"
          :recommended-model="recommendedModel"
          :payment-enabled="paymentEnabled"
        />

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

        <UserDashboardCharts
          v-if="stats"
          v-model:startDate="startDate"
          v-model:endDate="endDate"
          v-model:granularity="granularity"
          :loading="loadingCharts"
          :trend="trendData"
          :models="modelStats"
          @dateRangeChange="loadCharts"
          @granularityChange="loadCharts"
          @refresh="refreshAll"
        />

        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2">
            <UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" />
          </div>
          <div class="lg:col-span-1">
            <UserDashboardQuickActions />
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
import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'
import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import UserDashboardCommercialPath from '@/components/user/dashboard/UserDashboardCommercialPath.vue'
import { useAppStore, useAuthStore } from '@/stores'
import { keysAPI } from '@/api/keys'
import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import { getMyPlatformQuotas } from '@/api/user'
import { getBrowserOriginFallback, normalizeOpenAIBaseUrl } from '@/utils/quickStart'
import type { ApiKey, ModelStat, PlatformQuotaItem, TrendDataPoint, UsageLog, UserErrorRequest } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const stats = ref<UserStatsType | null>(null)
const loading = ref(false)
const loadingUsage = ref(false)
const loadingCharts = ref(false)
const loadingApiKeys = ref(false)
const loadingErrors = ref(false)
const dashboardError = ref(false)
const apiKeysError = ref(false)
const errorsError = ref(false)

const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const recentUsage = ref<UsageLog[]>([])
const apiKeys = ref<ApiKey[]>([])
const recentErrors = ref<UserErrorRequest[]>([])
const platformQuotas = ref<PlatformQuotaItem[] | null>(null)

const formatLocalDate = (d: Date) => d.toISOString().split('T')[0]
const startDate = ref(formatLocalDate(new Date(Date.now() - 6 * 86400000)))
const endDate = ref(formatLocalDate(new Date()))
const granularity = ref<'day' | 'hour'>('day')

const balance = computed(() => {
  const value = user.value?.balance
  return typeof value === 'number' && Number.isFinite(value) ? value : null
})

const paymentEnabled = computed(() => appStore.cachedPublicSettings?.payment_enabled ?? false)
const errorViewEnabled = computed(() => appStore.cachedPublicSettings?.allow_user_view_error_requests ?? false)
const baseUrl = computed(() =>
  normalizeOpenAIBaseUrl(appStore.cachedPublicSettings?.api_base_url, getBrowserOriginFallback())
)
const recommendedModel = computed(() => modelStats.value[0]?.model || '')

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

const loadCharts = async () => {
  loadingCharts.value = true
  try {
    const [trend, models] = await Promise.all([
      usageAPI.getDashboardTrend({
        start_date: startDate.value,
        end_date: endDate.value,
        granularity: granularity.value
      }),
      usageAPI.getDashboardModels({
        start_date: startDate.value,
        end_date: endDate.value
      })
    ])
    trendData.value = trend.trend || []
    modelStats.value = models.models || []
  } catch (error) {
    console.error('Failed to load charts:', error instanceof Error ? error.message : error)
    trendData.value = []
    modelStats.value = []
  } finally {
    loadingCharts.value = false
  }
}

const loadRecent = async () => {
  loadingUsage.value = true
  try {
    const res = await usageAPI.getByDateRange(startDate.value, endDate.value)
    recentUsage.value = res.items.slice(0, 5)
  } catch (error) {
    console.error('Failed to load recent usage:', error instanceof Error ? error.message : error)
    recentUsage.value = []
  } finally {
    loadingUsage.value = false
  }
}

const loadApiKeys = async () => {
  loadingApiKeys.value = true
  apiKeysError.value = false
  try {
    const res = await keysAPI.list(1, 100)
    apiKeys.value = res.items || []
  } catch (error) {
    console.error('Failed to load API key summary:', error instanceof Error ? error.message : error)
    apiKeysError.value = true
    apiKeys.value = []
  } finally {
    loadingApiKeys.value = false
  }
}

const loadRecentErrors = async () => {
  if (!errorViewEnabled.value) {
    recentErrors.value = []
    return
  }

  loadingErrors.value = true
  errorsError.value = false
  try {
    const resp = await usageAPI.listMyErrorRequests({
      page: 1,
      page_size: 5,
    })
    recentErrors.value = resp.items || []
  } catch (error) {
    console.error('Failed to load recent user errors:', error instanceof Error ? error.message : error)
    errorsError.value = true
    recentErrors.value = []
  } finally {
    loadingErrors.value = false
  }
}

const loadPublicSettingsAndErrors = async () => {
  try {
    await appStore.fetchPublicSettings()
  } catch (error) {
    console.error('Failed to load public settings:', error instanceof Error ? error.message : error)
  }
  await loadRecentErrors()
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
  loadCharts()
  loadRecent()
  loadApiKeys()
  loadPlatformQuotas()
  loadPublicSettingsAndErrors()
}

onMounted(refreshAll)
</script>
