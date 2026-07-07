<template>
  <AppLayout>
    <div class="space-y-6">
      <UsageStatsCards :stats="usageStats" :show-account-cost="false" :strike-standard-cost="true" />

      <div class="card px-4 py-3">
        <div class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
          <div class="flex flex-1 flex-wrap items-end gap-4">
            <div class="w-full sm:w-[220px]">
              <label class="input-label">{{ t('usage.apiKeyFilter') }}</label>
              <Select v-model="filters.api_key_id" :options="apiKeyOptions" @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto">
              <label class="input-label">{{ t('admin.dashboard.timeRange') }}</label>
              <DateRangePicker
                v-model:start-date="startDate"
                v-model:end-date="endDate"
                @change="onDateRangeChange"
              />
            </div>
          </div>

          <div class="flex w-full flex-wrap items-center justify-end gap-2 xl:w-auto">
            <button
              type="button"
              @click="refreshData"
              :disabled="loading"
              class="btn btn-secondary h-9 rounded-lg px-3 py-0"
            >
              <Icon name="refresh" size="sm" />
              <span>{{ t('common.refresh') }}</span>
            </button>
            <button type="button" @click="resetFilters" class="btn btn-secondary h-9 rounded-lg px-3 py-0">
              <Icon name="x" size="sm" />
              <span>{{ t('common.reset') }}</span>
            </button>
            <div class="relative" ref="columnDropdownRef">
              <button
                type="button"
                @click="showColumnDropdown = !showColumnDropdown"
                class="btn btn-secondary h-9 rounded-lg px-3 py-0"
                :title="t('admin.users.columnSettings')"
              >
                <Icon name="grid" size="sm" />
                <span>{{ t('admin.users.columnSettings') }}</span>
              </button>
              <div
                v-if="showColumnDropdown"
                class="absolute right-0 top-full z-50 mt-1 max-h-80 w-48 overflow-y-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
              >
                <button
                  v-for="col in toggleableColumns"
                  :key="col.key"
                  type="button"
                  @click="toggleColumn(col.key)"
                  class="flex w-full items-center justify-between px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
                >
                  <span>{{ col.label }}</span>
                  <Icon v-if="isColumnVisible(col.key)" name="check" size="sm" class="text-primary-500" />
                </button>
              </div>
            </div>
            <button
              type="button"
              @click="exportToCSV"
              :disabled="exporting"
              class="btn btn-primary h-9 rounded-lg px-3 py-0"
            >
              <Icon name="download" size="sm" />
              <span>{{ exporting ? t('usage.exporting') : t('usage.exportCsv') }}</span>
            </button>
          </div>
        </div>
      </div>

      <div class="card overflow-hidden">
        <div class="flex border-b border-gray-200 px-4 dark:border-dark-700">
          <button
            type="button"
            class="usage-record-tab"
            :class="{ 'usage-record-tab-active': activeTab === 'usage' }"
            @click="activeTab = 'usage'"
          >
            {{ t('usage.tabs.usage') }}
          </button>
          <button
            type="button"
            class="usage-record-tab"
            :class="{ 'usage-record-tab-active': activeTab === 'errors' }"
            @click="switchToErrors"
          >
            {{ t('usage.tabs.errors') }}
          </button>
        </div>

        <template v-if="activeTab === 'usage'">
          <UsageTable
            :data="usageLogs"
            :loading="loading"
            :columns="visibleColumns"
            :server-side-sort="true"
            :show-account-billing="false"
            :show-upstream-endpoint="false"
            :dense="true"
            :framed="false"
            default-sort-key="created_at"
            default-sort-order="desc"
            @sort="handleSort"
            @ipGeoBatchFailed="handleIpGeoBatchFailed"
          />

          <div v-if="pagination.total > 0" class="border-t border-gray-100 px-4 py-3 dark:border-dark-700">
            <Pagination
              :page="pagination.page"
              :total="pagination.total"
              :page-size="pagination.page_size"
              @update:page="handlePageChange"
              @update:pageSize="handlePageSizeChange"
            />
          </div>
        </template>

        <UserErrorRequestsTable
          v-else-if="errorViewEnabled"
          :rows="errorRows"
          :total="errorTotal"
          :loading="errorLoading"
          :page="errorPage"
          :page-size="errorPageSize"
          :api-keys="apiKeys"
          @filter="onErrorFilter"
          @update:page="onErrorPage"
          @update:pageSize="onErrorPageSize"
        />

        <div v-else class="flex min-h-[260px] items-center justify-center px-6 py-12">
          <div class="text-center">
            <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 text-gray-400 dark:bg-dark-800 dark:text-dark-500">
              <Icon name="inbox" size="lg" />
            </div>
            <p class="mt-4 text-sm text-gray-500 dark:text-dark-400">{{ t('usage.errors.disabled') }}</p>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>

</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { keysAPI, usageAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import UsageStatsCards from '@/components/admin/usage/UsageStatsCards.vue'
import UsageTable from '@/components/admin/usage/UsageTable.vue'
import Icon from '@/components/icons/Icon.vue'
import UserErrorRequestsTable from '@/components/user/UserErrorRequestsTable.vue'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { formatReasoningEffort } from '@/utils/format'
import { BILLING_MODE_IMAGE, getBillingModeLabel } from '@/utils/billingMode'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import type {
  ApiKey,
  UsageLog,
  UsageQueryParams,
  UsageStatsResponse,
  UserErrorRequest,
} from '@/types'
import type { Column } from '@/components/common/types'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()

const usageStats = ref<UsageStatsResponse | null>(null)
const usageLogs = ref<UsageLog[]>([])

const loading = ref(false)
const exporting = ref(false)
const errorRows = ref<UserErrorRequest[]>([])
const errorLoading = ref(false)
const errorPage = ref(1)
const errorPageSize = ref(20)
const errorTotal = ref(0)
const errorFilter = ref<{ model: string; category: string; api_key_id: number | null }>({
  model: '',
  category: '',
  api_key_id: null,
})

let abortController: AbortController | null = null
let statsReqSeq = 0

const formatLocalDate = (date: Date): string =>
  `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`

const getLast24HoursRangeDates = () => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return { start: formatLocalDate(start), end: formatLocalDate(end) }
}

const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)

const activeTab = ref<'usage' | 'errors'>('usage')
const errorViewEnabled = computed(() => appStore.cachedPublicSettings?.allow_user_view_error_requests ?? false)

const filters = ref<UsageQueryParams>({
  start_date: startDate.value,
  end_date: endDate.value,
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
})
const sortState = reactive({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc',
})

const apiKeys = ref<ApiKey[]>([])

const apiKeyOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('usage.allApiKeys') },
  ...apiKeys.value.map((key) => ({ value: key.id, label: key.name })),
])

const normalizedFilters = computed<UsageQueryParams>(() => ({
  ...filters.value,
  api_key_id: filters.value.api_key_id ?? undefined,
  start_date: startDate.value,
  end_date: endDate.value,
}))

const buildUsageListParams = (page: number, pageSize: number): UsageQueryParams => ({
  page,
  page_size: pageSize,
  ...normalizedFilters.value,
  sort_by: sortState.sort_by,
  sort_order: sortState.sort_order,
})

const loadLogs = async () => {
  abortController?.abort()
  const controller = new AbortController()
  abortController = controller
  loading.value = true
  try {
    const res = await usageAPI.query(buildUsageListParams(pagination.page, pagination.page_size), {
      signal: controller.signal,
    })
    if (!controller.signal.aborted) {
      usageLogs.value = res.items
      pagination.total = res.total
    }
  } catch (error: any) {
    if (error?.name !== 'AbortError' && error?.code !== 'ERR_CANCELED') {
      appStore.showError(t('usage.failedToLoad'))
    }
  } finally {
    if (abortController === controller) loading.value = false
  }
}

const loadStats = async () => {
  const seq = ++statsReqSeq
  try {
    const stats = await usageAPI.getStats(normalizedFilters.value)
    if (seq !== statsReqSeq) return
    usageStats.value = stats
  } catch (error) {
    if (seq !== statsReqSeq) return
    console.error('Failed to load usage stats:', error)
    usageStats.value = null
  }
}

const applyFilters = () => {
  pagination.page = 1
  void loadLogs()
  void loadStats()
  resetErrorRows()
}

const refreshData = () => {
  void loadLogs()
  void loadStats()
  if (activeTab.value === 'errors' && errorViewEnabled.value) void loadErrors()
}

const resetFilters = () => {
  const range = getLast24HoursRangeDates()
  startDate.value = range.start
  endDate.value = range.end
  filters.value = {
    start_date: range.start,
    end_date: range.end,
  }
  applyFilters()
}

const onDateRangeChange = (range: { startDate: string; endDate: string; preset: string | null }) => {
  startDate.value = range.startDate
  endDate.value = range.endDate
  filters.value.start_date = range.startDate
  filters.value.end_date = range.endDate
  applyFilters()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  void loadLogs()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadLogs()
}

const handleSort = (key: string, order: 'asc' | 'desc') => {
  sortState.sort_by = key
  sortState.sort_order = order
  pagination.page = 1
  void loadLogs()
}

const handleIpGeoBatchFailed = () => {
  appStore.showError(t('usage.ipGeo.batchFailed'))
}

const getRequestTypeExportText = (log: UsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'cyber') return 'Cyber'
  if (requestType === 'ws_v2') return 'WS'
  if (requestType === 'stream') return 'Stream'
  if (requestType === 'sync') return 'Non-stream'
  return 'Unknown'
}

const getDisplayBillingMode = (
  row: Pick<UsageLog, 'billing_mode' | 'image_count'> | null | undefined
): string | null | undefined => {
  if ((row?.image_count ?? 0) > 0) return BILLING_MODE_IMAGE
  return row?.billing_mode
}

const escapeCSVValue = (value: unknown): string => {
  if (value == null) return ''
  const str = String(value)
  const escaped = str.replace(/"/g, '""')
  if (/^[=+\-@\t\r]/.test(str)) return `"\'${escaped}"`
  if (/[,"\n\r]/.test(str)) return `"${escaped}"`
  return str
}

const exportToCSV = async () => {
  if (pagination.total === 0) {
    appStore.showWarning(t('usage.noDataToExport'))
    return
  }
  exporting.value = true
  appStore.showInfo(t('usage.preparingExport'))
  try {
    const allLogs: UsageLog[] = []
    const pageSize = 100
    const totalPages = Math.ceil(pagination.total / pageSize)
    for (let page = 1; page <= totalPages; page++) {
      const response = await usageAPI.query(buildUsageListParams(page, pageSize))
      allLogs.push(...response.items)
    }
    if (allLogs.length === 0) {
      appStore.showWarning(t('usage.noDataToExport'))
      return
    }
    const headers = [
      'Time',
      'API Key Name',
      'Model',
      'Reasoning Effort',
      'Inbound Endpoint',
      'IP Address',
      'Type',
      'Billing Mode',
      'Input Tokens',
      'Output Tokens',
      'Cache Read Tokens',
      'Cache Creation Tokens',
      'Rate Multiplier',
      'Billed Cost',
      'Original Cost',
      'First Token (ms)',
      'Duration (ms)',
    ]
    const rows = allLogs.map((log) => [
      log.created_at,
      log.api_key?.name || '',
      log.model,
      formatReasoningEffort(log.reasoning_effort),
      log.inbound_endpoint || '',
      log.ip_address || '',
      getRequestTypeExportText(log),
      getBillingModeLabel(getDisplayBillingMode(log), t),
      log.input_tokens,
      log.output_tokens,
      log.cache_read_tokens,
      log.cache_creation_tokens,
      log.rate_multiplier,
      log.actual_cost.toFixed(8),
      log.total_cost.toFixed(8),
      log.first_token_ms ?? '',
      log.duration_ms ?? '',
    ].map(escapeCSVValue))
    const csvContent = [
      headers.map(escapeCSVValue).join(','),
      ...rows.map((row) => row.join(',')),
    ].join('\n')
    const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `usage_${startDate.value}_to_${endDate.value}.csv`
    link.click()
    window.URL.revokeObjectURL(url)
    appStore.showSuccess(t('usage.exportSuccess'))
  } catch (error) {
    console.error('CSV Export failed:', error)
    appStore.showError(t('usage.exportFailed'))
  } finally {
    exporting.value = false
  }
}

const ALWAYS_VISIBLE = ['created_at']
const DEFAULT_HIDDEN_COLUMNS = ['reasoning_effort', 'ip_address', 'group', 'user_agent']
const HIDDEN_COLUMNS_KEY = 'user-usage-hidden-columns-v2'

const allColumns = computed<Column[]>(() => [
  { key: 'api_key', label: t('usage.apiKeyFilter'), sortable: false },
  { key: 'model', label: t('usage.model'), sortable: true },
  { key: 'reasoning_effort', label: t('usage.reasoningEffort'), sortable: false },
  { key: 'endpoint', label: t('usage.endpoint'), sortable: false },
  { key: 'ip_address', label: 'IP', sortable: false },
  { key: 'group', label: t('admin.usage.group'), sortable: false },
  { key: 'stream', label: t('usage.type'), sortable: false },
  { key: 'billing_mode', label: t('admin.usage.billingMode'), sortable: false },
  { key: 'tokens', label: t('usage.tokens'), sortable: false },
  { key: 'cache_hit_rate', label: t('usage.cacheHitRate'), sortable: false, class: 'min-w-[132px]' },
  { key: 'cost', label: t('usage.cost'), sortable: false },
  { key: 'first_token', label: t('usage.firstToken'), sortable: false },
  { key: 'duration', label: t('usage.duration'), sortable: false },
  { key: 'created_at', label: t('usage.time'), sortable: true },
  { key: 'user_agent', label: t('usage.userAgent'), sortable: false },
])

const hiddenColumns = reactive<Set<string>>(new Set())
const toggleableColumns = computed(() => allColumns.value.filter((col) => !ALWAYS_VISIBLE.includes(col.key)))
const visibleColumns = computed(() =>
  allColumns.value.filter((col) => ALWAYS_VISIBLE.includes(col.key) || !hiddenColumns.has(col.key))
)
const isColumnVisible = (key: string) => !hiddenColumns.has(key)
const toggleColumn = (key: string) => {
  if (hiddenColumns.has(key)) hiddenColumns.delete(key)
  else hiddenColumns.add(key)
  localStorage.setItem(HIDDEN_COLUMNS_KEY, JSON.stringify([...hiddenColumns]))
}
const loadSavedColumns = () => {
  try {
    const saved = localStorage.getItem(HIDDEN_COLUMNS_KEY)
    const values = saved ? JSON.parse(saved) as string[] : DEFAULT_HIDDEN_COLUMNS
    values.forEach((key) => hiddenColumns.add(key))
  } catch {
    DEFAULT_HIDDEN_COLUMNS.forEach((key) => hiddenColumns.add(key))
  }
}

const showColumnDropdown = ref(false)
const columnDropdownRef = ref<HTMLElement | null>(null)
const handleColumnClickOutside = (event: MouseEvent) => {
  if (columnDropdownRef.value && !columnDropdownRef.value.contains(event.target as HTMLElement)) {
    showColumnDropdown.value = false
  }
}

const loadFilterOptions = async () => {
  try {
    const keys = await keysAPI.list(1, 100)
    apiKeys.value = keys.items
  } catch (error) {
    console.error('Failed to load usage filter options:', error)
  }
}

const resetErrorRows = () => {
  errorPage.value = 1
  if (activeTab.value === 'errors' && errorViewEnabled.value) {
    void loadErrors()
  } else {
    errorRows.value = []
    errorTotal.value = 0
  }
}

const loadErrors = async () => {
  if (!errorViewEnabled.value) {
    errorRows.value = []
    errorTotal.value = 0
    return
  }

  errorLoading.value = true
  try {
    const resp = await usageAPI.listMyErrorRequests({
      page: errorPage.value,
      page_size: errorPageSize.value,
      start_date: startDate.value,
      end_date: endDate.value,
      model: errorFilter.value.model || undefined,
      category: errorFilter.value.category || undefined,
      api_key_id: errorFilter.value.api_key_id ?? undefined,
    })
    errorRows.value = resp.items
    errorTotal.value = resp.total
  } catch (error) {
    console.error('[UsageView] loadErrors failed:', error)
    appStore.showError(t('usage.errors.failedToLoad'))
  } finally {
    errorLoading.value = false
  }
}

const onErrorFilter = (filter: { model: string; category: string; api_key_id: number | null }) => {
  errorFilter.value = filter
  errorPage.value = 1
  void loadErrors()
}

const onErrorPage = (page: number) => {
  errorPage.value = page
  void loadErrors()
}

const onErrorPageSize = (pageSize: number) => {
  errorPageSize.value = pageSize
  errorPage.value = 1
  void loadErrors()
}

const switchToErrors = () => {
  activeTab.value = 'errors'
  if (errorViewEnabled.value && errorRows.value.length === 0) void loadErrors()
}

watch(
  () => [route.query.tab, errorViewEnabled.value] as const,
  ([tab, enabled]) => {
    if (tab === 'errors' && activeTab.value !== 'errors') {
      switchToErrors()
      return
    }
    if (enabled && activeTab.value === 'errors' && errorRows.value.length === 0 && !errorLoading.value) {
      void loadErrors()
    }
  },
  { immediate: true }
)

onMounted(() => {
  loadSavedColumns()
  document.addEventListener('click', handleColumnClickOutside)
  void loadFilterOptions()
  refreshData()
})

onUnmounted(() => {
  abortController?.abort()
  document.removeEventListener('click', handleColumnClickOutside)
})
</script>

<style scoped>
.usage-record-tab {
  position: relative;
  margin-bottom: -1px;
  padding: 0.875rem 1rem;
  border-bottom: 2px solid transparent;
  font-size: 0.875rem;
  font-weight: 500;
  color: rgb(75 85 99);
  transition:
    color 0.16s ease,
    border-color 0.16s ease,
    background-color 0.16s ease;
}

.usage-record-tab:hover {
  color: rgb(17 24 39);
  background-color: rgb(249 250 251);
}

.usage-record-tab-active {
  color: rgb(17 24 39);
  border-bottom-color: rgb(20 184 166);
  background-color: white;
}

.dark .usage-record-tab {
  color: rgb(156 163 175);
}

.dark .usage-record-tab:hover {
  color: rgb(243 244 246);
  background-color: rgb(31 41 55 / 0.6);
}

.dark .usage-record-tab-active {
  color: white;
  border-bottom-color: rgb(45 212 191);
  background-color: rgb(17 24 39);
}
</style>
