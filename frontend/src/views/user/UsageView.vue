<template>
  <AppLayout>
    <div class="space-y-6">
      <UsageStatsCards
        :stats="usageStats"
        :show-account-cost="false"
        :strike-standard-cost="true"
        :show-token-breakdown="true"
      />

      <div class="card p-4">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.dashboard.timeRange') }}:
            </span>
            <DateRangePicker
              v-model:start-date="startDate"
              v-model:end-date="endDate"
              v-model:start-time="startTime"
              v-model:end-time="endTime"
              show-time-inputs
              :preset-values="dashboardDatePresets"
              @change="onDateRangeChange"
            />
          </div>
          <TokenCountModeToggle />
        </div>
      </div>

      <LegacyUsageLogPanel
        :start-date="startDate"
        :end-date="endDate"
        :start-time="startTime"
        :end-time="endTime"
        :period="activePeriod"
        :timezone="timezone"
        :api-keys="apiKeys"
        :groups="groups"
        :models="modelOptionValues"
        :error-view-enabled="errorViewEnabled"
        :refresh-key="legacyRefreshKey"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { keysAPI, usageAPI, userGroupsAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import AppLayout from '@/components/layout/AppLayout.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import TokenCountModeToggle from '@/components/common/TokenCountModeToggle.vue'
import UsageStatsCards from '@/components/admin/usage/UsageStatsCards.vue'
import LegacyUsageLogPanel from '@/components/user/LegacyUsageLogPanel.vue'
import {
  dashboardDatePresets,
  type DateRangeChange,
  getDashboardPresetPeriod,
  toLocalDateTimeParam,
} from '@/utils/dashboardTimeRange'
import { formatDateLocalInput, formatTimeLocalInput } from '@/utils/format'
import type { UserUsagePeriod } from '@/api/usage'
import type {
  ApiKey,
  Group,
  ModelStat,
  UsageQueryParams,
  UsageStatsResponse,
} from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const getLast24HoursRange = () => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return {
    startDate: formatDateLocalInput(start),
    endDate: formatDateLocalInput(end),
    startTime: formatTimeLocalInput(start),
    endTime: formatTimeLocalInput(end),
  }
}

const defaultRange = getLast24HoursRange()
const startDate = ref(defaultRange.startDate)
const endDate = ref(defaultRange.endDate)
const startTime = ref(defaultRange.startTime)
const endTime = ref(defaultRange.endTime)
const activePeriod = ref<UserUsagePeriod | null>('24h')
const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone

const usageStats = ref<UsageStatsResponse | null>(null)
const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const modelOptionValues = ref<string[]>([])

const legacyRefreshKey = ref(0)
const errorViewEnabled = computed(() => appStore.cachedPublicSettings?.allow_user_view_error_requests ?? false)

const activeRangeParams = computed<Pick<UsageQueryParams, 'start_time' | 'end_time' | 'period'>>(() => activePeriod.value
  ? { period: activePeriod.value }
  : {
      start_time: toLocalDateTimeParam(startDate.value, startTime.value),
      end_time: toLocalDateTimeParam(endDate.value, endTime.value),
    })

const normalizedUsageFilters = computed<UsageQueryParams>(() => ({
  ...activeRangeParams.value,
  timezone,
}))
let dashboardRequestSequence = 0

const refreshModelOptions = (models: ModelStat[]) => {
  const values = new Set(modelOptionValues.value)
  models.forEach((item) => {
    if (item.model) values.add(item.model)
  })
  modelOptionValues.value = Array.from(values).sort()
}

const loadDashboardData = async () => {
  const sequence = ++dashboardRequestSequence
  try {
    const [stats, models] = await Promise.all([
      usageAPI.getStats(normalizedUsageFilters.value),
      usageAPI.getDashboardModels({ ...normalizedUsageFilters.value, model_source: 'requested' }),
    ])
    if (sequence !== dashboardRequestSequence) return
    usageStats.value = stats
    refreshModelOptions(models.models || [])
  } catch (error) {
    if (sequence !== dashboardRequestSequence) return
    console.error('[UsageView] failed to load usage summaries:', error)
  }
}

const loadFilterOptions = async () => {
  try {
    const [keyResponse, availableGroups] = await Promise.all([
      keysAPI.list(1, 100),
      userGroupsAPI.getAvailable(),
    ])
    apiKeys.value = keyResponse.items
    groups.value = availableGroups
  } catch (error) {
    console.error('[UsageView] failed to load filter options:', error)
  }
}

const refreshData = () => {
  legacyRefreshKey.value += 1
  void loadDashboardData()
}

const onDateRangeChange = (range: DateRangeChange) => {
  startDate.value = range.startDate
  endDate.value = range.endDate
  startTime.value = range.startTime
  endTime.value = range.endTime
  activePeriod.value = getDashboardPresetPeriod(range.preset)
  void loadDashboardData()
}

onMounted(() => {
  void loadFilterOptions()
  refreshData()
})

onUnmounted(() => {
  dashboardRequestSequence += 1
})
</script>
