<template>
  <AppLayout>
    <div class="space-y-6">
      <UsageStatsCards
        :stats="usageStats"
        :show-account-cost="false"
        :strike-standard-cost="true"
      />

      <div class="card p-4">
        <div class="flex flex-wrap items-center gap-4">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.dashboard.timeRange') }}:
            </span>
            <DateRangePicker
              v-model:start-date="startDate"
              v-model:end-date="endDate"
              :preset-values="dashboardDatePresets"
              @change="onDateRangeChange"
            />
          </div>
        </div>
      </div>

      <EntityDistributionChart
        v-model:metric="apiKeyDistributionMetric"
        :title="t('usage.apiKeyDistribution')"
        :entity-label="t('usage.apiKeyFilter')"
        :items="apiKeyDistributionItems"
        :loading="chartsLoading"
      />

      <div class="card p-4">
        <div class="flex flex-wrap items-end justify-between gap-4">
          <div class="flex flex-1 flex-wrap items-end gap-4">
            <div class="w-full sm:w-auto sm:min-w-[170px]">
              <label class="input-label">{{ t('usage.logs.typeFilter') }}</label>
              <Select v-model="logKind" :options="kindOptions" @change="applyLogFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[210px]">
              <label class="input-label">{{ t('usage.apiKeyFilter') }}</label>
              <Select v-model="filters.api_key_id" :options="apiKeyOptions" @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[210px]">
              <label class="input-label">{{ t('usage.model') }}</label>
              <Select
                v-model="filters.model"
                :options="modelOptions"
                searchable
                clearable
                @change="applyFilters"
              />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[190px]">
              <label class="input-label">{{ t('admin.usage.group') }}</label>
              <Select
                v-model="filters.group_id"
                :options="groupOptions"
                searchable
                @change="applyFilters"
              />
            </div>
          </div>

          <div class="flex w-full flex-wrap items-center justify-end gap-3 sm:w-auto">
            <button type="button" class="btn btn-secondary" :disabled="loading" @click="refreshData">
              {{ t('common.refresh') }}
            </button>
            <button type="button" class="btn btn-secondary" @click="resetFilters">
              {{ t('common.reset') }}
            </button>
            <button type="button" class="btn btn-primary" :disabled="exporting" @click="exportToCSV">
              {{ exporting ? t('usage.exporting') : t('usage.exportCsv') }}
            </button>
          </div>
        </div>
      </div>

      <UserRequestLogTable
        :rows="requestLogs"
        :loading="loading"
        @sort="handleSort"
      />

      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { keysAPI, usageAPI, userGroupsAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import AppLayout from '@/components/layout/AppLayout.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import UsageStatsCards from '@/components/admin/usage/UsageStatsCards.vue'
import EntityDistributionChart from '@/components/charts/EntityDistributionChart.vue'
import UserRequestLogTable from '@/components/user/UserRequestLogTable.vue'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import {
  dashboardDatePresets,
  getDashboardPresetGranularity,
  getDashboardPresetPeriod,
} from '@/utils/dashboardTimeRange'
import type { ModelUsageTrendGranularity, UserUsagePeriod } from '@/api/usage'
import type {
  ApiKey,
  ApiKeyStat,
  Group,
  ModelStat,
  UsageQueryParams,
  UsageStatsResponse,
  UserRequestLog,
  UserRequestLogQueryParams,
} from '@/types'

type DistributionMetric = 'tokens' | 'actual_cost'
type LogKindFilter = NonNullable<UserRequestLogQueryParams['kind']>

const { t } = useI18n()
const appStore = useAppStore()

const formatLocalDate = (date: Date): string =>
  `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`

const getLast24HoursRangeDates = () => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return { start: formatLocalDate(start), end: formatLocalDate(end) }
}

const getGranularityForRange = (start: string, end: string): ModelUsageTrendGranularity => {
  const startTime = new Date(`${start}T00:00:00`).getTime()
  const endTime = new Date(`${end}T00:00:00`).getTime()
  return Math.ceil((endTime - startTime) / 86_400_000) <= 1 ? 'hour' : 'day'
}

const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)
const activePeriod = ref<UserUsagePeriod | null>('24h')
const granularity = ref<ModelUsageTrendGranularity>(getGranularityForRange(defaultRange.start, defaultRange.end))
const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone

const usageStats = ref<UsageStatsResponse | null>(null)
const requestLogs = ref<UserRequestLog[]>([])
const apiKeyStats = ref<ApiKeyStat[]>([])
const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const modelOptionValues = ref<string[]>([])

const loading = ref(false)
const chartsLoading = ref(false)
const exporting = ref(false)
const apiKeyDistributionMetric = ref<DistributionMetric>('tokens')
const logKind = ref<LogKindFilter>('all')
const errorViewEnabled = computed(() => appStore.cachedPublicSettings?.allow_user_view_error_requests ?? false)

const filters = ref<{
  api_key_id: number | null
  group_id: number | null
  model: string | null
}>({
  api_key_id: null,
  group_id: null,
  model: null,
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
})

const sortState = reactive<{
  sort_by: 'created_at' | 'duration_ms'
  sort_order: 'asc' | 'desc'
}>({
  sort_by: 'created_at',
  sort_order: 'desc',
})

const kindOptions = computed<SelectOption[]>(() => {
  const options: SelectOption[] = [
    { value: 'all', label: t('usage.logs.kinds.all') },
    { value: 'consumption', label: t('usage.logs.kinds.consumption') },
  ]
  if (errorViewEnabled.value) {
    options.push({ value: 'error', label: t('usage.logs.kinds.error') })
  }
  return options
})

const apiKeyOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('usage.allApiKeys') },
  ...apiKeys.value.map((key) => ({ value: key.id, label: key.name })),
])

const groupOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allGroups') },
  ...groups.value.map((group) => ({ value: group.id, label: group.name })),
])

const modelOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allModels') },
  ...modelOptionValues.value.map((model) => ({ value: model, label: model })),
])

const apiKeyDistributionItems = computed(() => apiKeyStats.value.map((stat) => ({
  id: stat.api_key_id,
  label: stat.api_key_name || t('usage.apiKeyFallback', { id: stat.api_key_id }),
  requests: stat.requests,
  total_tokens: stat.total_tokens,
  cost: stat.cost,
  actual_cost: stat.actual_cost,
})))

const activeRangeParams = computed<Pick<UsageQueryParams, 'start_date' | 'end_date' | 'period'>>(() => activePeriod.value
  ? { period: activePeriod.value }
  : { start_date: startDate.value, end_date: endDate.value })

const normalizedUsageFilters = computed<UsageQueryParams>(() => ({
  ...activeRangeParams.value,
  timezone,
  api_key_id: filters.value.api_key_id ?? undefined,
  group_id: filters.value.group_id ?? undefined,
  model: filters.value.model?.trim() || undefined,
}))

const buildRequestLogParams = (page: number, pageSize: number): UserRequestLogQueryParams => ({
  page,
  page_size: pageSize,
  kind: errorViewEnabled.value ? logKind.value : 'consumption',
  ...activeRangeParams.value,
  timezone,
  api_key_id: filters.value.api_key_id ?? undefined,
  group_id: filters.value.group_id ?? undefined,
  model: filters.value.model?.trim() || undefined,
  sort_by: sortState.sort_by,
  sort_order: sortState.sort_order,
})

let logAbortController: AbortController | null = null
let dashboardRequestSequence = 0

const loadRequestLogs = async () => {
  logAbortController?.abort()
  const controller = new AbortController()
  logAbortController = controller
  loading.value = true
  try {
    const response = await usageAPI.listRequestLogs(
      buildRequestLogParams(pagination.page, pagination.page_size),
      { signal: controller.signal }
    )
    if (controller.signal.aborted) return
    requestLogs.value = response.items
    pagination.total = response.total
  } catch (error: any) {
    if (error?.name !== 'AbortError' && error?.code !== 'ERR_CANCELED') {
      appStore.showError(t('usage.logs.failedToLoad'))
    }
  } finally {
    if (logAbortController === controller) loading.value = false
  }
}

const refreshModelOptions = (models: ModelStat[]) => {
  const values = new Set(modelOptionValues.value)
  models.forEach((item) => {
    if (item.model) values.add(item.model)
  })
  if (filters.value.model) values.add(filters.value.model)
  modelOptionValues.value = Array.from(values).sort()
}

const loadDashboardData = async () => {
  const sequence = ++dashboardRequestSequence
  chartsLoading.value = true
  try {
    const [stats, models, snapshot] = await Promise.all([
      usageAPI.getStats(normalizedUsageFilters.value),
      usageAPI.getDashboardModels({ ...normalizedUsageFilters.value, model_source: 'requested' }),
      usageAPI.getDashboardSnapshotV2({
        ...normalizedUsageFilters.value,
        granularity: granularity.value,
        include_trend: false,
        include_model_stats: false,
        include_group_stats: false,
        include_api_key_stats: true,
      }),
    ])
    if (sequence !== dashboardRequestSequence) return
    usageStats.value = stats
    refreshModelOptions(models.models || [])
    apiKeyStats.value = snapshot.api_keys || []
  } catch (error) {
    if (sequence !== dashboardRequestSequence) return
    console.error('[UsageView] failed to load usage summaries:', error)
    apiKeyStats.value = []
  } finally {
    if (sequence === dashboardRequestSequence) chartsLoading.value = false
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
  void loadRequestLogs()
  void loadDashboardData()
}

const applyLogFilters = () => {
  pagination.page = 1
  void loadRequestLogs()
}

const applyFilters = () => {
  pagination.page = 1
  refreshData()
}

const resetFilters = () => {
  const range = getLast24HoursRangeDates()
  startDate.value = range.start
  endDate.value = range.end
  activePeriod.value = '24h'
  granularity.value = getGranularityForRange(range.start, range.end)
  filters.value = { api_key_id: null, group_id: null, model: null }
  logKind.value = 'all'
  sortState.sort_by = 'created_at'
  sortState.sort_order = 'desc'
  applyFilters()
}

const onDateRangeChange = (range: { startDate: string; endDate: string; preset: string | null }) => {
  startDate.value = range.startDate
  endDate.value = range.endDate
  activePeriod.value = getDashboardPresetPeriod(range.preset)
  granularity.value = getDashboardPresetGranularity(range.preset)
    ?? getGranularityForRange(range.startDate, range.endDate)
  applyFilters()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  void loadRequestLogs()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadRequestLogs()
}

const handleSort = (key: 'created_at' | 'duration_ms', order: 'asc' | 'desc') => {
  sortState.sort_by = key
  sortState.sort_order = order
  pagination.page = 1
  void loadRequestLogs()
}

const escapeCSVValue = (value: unknown): string => {
  if (value == null) return ''
  const stringValue = String(value)
  const escaped = stringValue.replace(/"/g, '""')
  if (/^[=+\-@\t\r]/.test(stringValue)) return `"'${escaped}"`
  if (/[,"\n\r]/.test(stringValue)) return `"${escaped}"`
  return stringValue
}

const exportToCSV = async () => {
  if (pagination.total === 0) {
    appStore.showWarning(t('usage.noDataToExport'))
    return
  }

  exporting.value = true
  appStore.showInfo(t('usage.preparingExport'))
  try {
    const logs: UserRequestLog[] = []
    const pageSize = 100
    const pages = Math.ceil(pagination.total / pageSize)
    for (let page = 1; page <= pages; page += 1) {
      const response = await usageAPI.listRequestLogs(buildRequestLogParams(page, pageSize))
      logs.push(...response.items)
    }

    const headers = [
      'Time', 'Request ID', 'Type', 'API Key', 'Group', 'Rate Multiplier', 'Model', 'Reasoning Effort',
      'First Token (ms)', 'Duration (ms)', 'Total Tokens', 'Actual Cost', 'Status Code',
      'Error Code', 'Details',
    ]
    const rows = logs.map((log) => [
      log.created_at,
      log.request_id,
      log.kind,
      log.api_key_name,
      log.group_name,
      log.rate_multiplier,
      log.model,
      log.reasoning_effort,
      log.first_token_ms,
      log.duration_ms,
      log.total_tokens,
      log.actual_cost == null ? '' : log.actual_cost.toFixed(8),
      log.status_code,
      log.error_code,
      log.error_message,
    ].map(escapeCSVValue))
    const csv = [headers.map(escapeCSVValue).join(','), ...rows.map((row) => row.join(','))].join('\n')
    const blob = new Blob([`\uFEFF${csv}`], { type: 'text/csv;charset=utf-8;' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `logs_${startDate.value}_to_${endDate.value}.csv`
    link.click()
    window.URL.revokeObjectURL(url)
    appStore.showSuccess(t('usage.exportSuccess'))
  } catch (error) {
    console.error('[UsageView] CSV export failed:', error)
    appStore.showError(t('usage.exportFailed'))
  } finally {
    exporting.value = false
  }
}

watch(errorViewEnabled, (enabled, wasEnabled) => {
  if (!enabled && wasEnabled) {
    if (logKind.value === 'error') logKind.value = 'consumption'
    applyLogFilters()
    return
  }
  if (enabled && !wasEnabled && logKind.value === 'all') {
    applyLogFilters()
  }
})

onMounted(() => {
  void loadFilterOptions()
  refreshData()
})

onUnmounted(() => {
  logAbortController?.abort()
  dashboardRequestSequence += 1
})
</script>
