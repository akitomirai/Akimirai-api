<template>
  <div class="space-y-4">
    <div class="card p-4">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div v-if="activeTab === 'errors'" class="flex flex-1 flex-wrap items-end gap-4">
          <div class="w-full sm:w-auto sm:min-w-[190px]">
            <label class="input-label">{{ t('usage.errors.keyName') }}</label>
            <Select v-model="errorFilters.api_key_id" :options="apiKeyOptions" @change="applyErrorFilters" />
          </div>
          <div class="w-full sm:w-auto sm:min-w-[190px]">
            <label class="input-label">{{ t('usage.errors.model') }}</label>
            <Select v-model="errorFilters.model" :options="modelOptions" searchable clearable @change="applyErrorFilters" />
          </div>
          <div class="w-full sm:w-auto sm:min-w-[180px]">
            <label class="input-label">{{ t('usage.errors.category') }}</label>
            <Select v-model="errorFilters.category" :options="errorCategoryOptions" @change="applyErrorFilters" />
          </div>
          <div class="w-full sm:w-auto sm:min-w-[160px]">
            <label class="input-label">{{ t('usage.errors.status') }}</label>
            <Select v-model="errorFilters.status_code" :options="errorStatusOptions" @change="applyErrorFilters" />
          </div>
        </div>

        <div v-else class="flex flex-1 flex-wrap items-end gap-4">
          <div class="w-full sm:w-auto sm:min-w-[190px]">
            <label class="input-label">{{ t('usage.apiKeyFilter') }}</label>
            <Select v-model="usageFilters.api_key_id" :options="apiKeyOptions" @change="applyUsageFilters" />
          </div>
          <div class="w-full sm:w-auto sm:min-w-[190px]">
            <label class="input-label">{{ t('usage.model') }}</label>
            <Select v-model="usageFilters.model" :options="modelOptions" searchable clearable @change="applyUsageFilters" />
          </div>
          <div class="w-full sm:w-auto sm:min-w-[180px]">
            <label class="input-label">{{ t('admin.usage.group') }}</label>
            <Select v-model="usageFilters.group_id" :options="groupOptions" searchable @change="applyUsageFilters" />
          </div>
        </div>

        <div class="flex w-full flex-wrap items-center justify-end gap-3 sm:w-auto">
          <button type="button" class="btn btn-secondary" :disabled="currentLoading" @click="refresh">
            <Icon name="refresh" size="sm" />
            {{ t('common.refresh') }}
          </button>
          <button type="button" class="btn btn-secondary" @click="resetFilters">
            <Icon name="x" size="sm" />
            {{ t('common.reset') }}
          </button>
          <div ref="columnDropdownRef" class="relative">
            <button
              type="button"
              class="btn btn-secondary"
              :title="t('admin.users.columnSettings')"
              @click="showColumnDropdown = !showColumnDropdown"
            >
              <Icon name="grid" size="sm" />
              <span>{{ t('admin.users.columnSettings') }}</span>
            </button>
            <div
              v-if="showColumnDropdown"
              class="absolute right-0 top-full z-50 mt-1 max-h-80 w-52 overflow-y-auto rounded-md border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
            >
              <button
                v-for="column in currentToggleableColumns"
                :key="column.key"
                type="button"
                class="flex w-full items-center justify-between px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
                @click="toggleCurrentColumn(column.key)"
              >
                <span>{{ column.label }}</span>
                <Icon v-if="isCurrentColumnVisible(column.key)" name="check" size="sm" class="text-primary-500" />
              </button>
            </div>
          </div>
          <button
            v-if="activeTab === 'usage'"
            type="button"
            class="btn btn-primary"
            :disabled="exporting"
            @click="exportToCSV"
          >
            <Icon name="download" size="sm" />
            {{ exporting ? t('usage.exporting') : t('usage.exportCsv') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="errorViewEnabled" class="flex gap-2 border-b border-gray-200 dark:border-dark-700">
      <button
        type="button"
        class="border-b-2 px-4 py-2 text-sm font-medium transition-colors"
        :class="activeTab === 'usage' ? 'border-primary-500 text-primary-600 dark:text-primary-400' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'"
        @click="activeTab = 'usage'"
      >
        {{ t('usage.tabs.usage') }}
      </button>
      <button
        type="button"
        class="border-b-2 px-4 py-2 text-sm font-medium transition-colors"
        :class="activeTab === 'errors' ? 'border-primary-500 text-primary-600 dark:text-primary-400' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'"
        @click="openErrors"
      >
        {{ t('usage.tabs.errors') }}
      </button>
    </div>

    <template v-if="activeTab === 'usage'">
      <UsageTable
        :data="usageRows"
        :columns="visibleUsageColumns"
        :loading="usageLoading"
        :server-side-sort="true"
        default-sort-key="created_at"
        default-sort-order="desc"
        :show-account-billing="false"
        :show-upstream-endpoint="false"
        :token-breakdown="true"
        @sort="onUsageSort"
        @ip-geo-batch-failed="showIpGeoFailure"
      />
      <Pagination
        v-if="usagePagination.total > 0"
        :page="usagePagination.page"
        :page-size="usagePagination.pageSize"
        :total="usagePagination.total"
        @update:page="onUsagePage"
        @update:page-size="onUsagePageSize"
      />
    </template>

    <UserErrorRequestsTable
      v-else
      :rows="errorRows"
      :total="errorPagination.total"
      :loading="errorLoading"
      :page="errorPagination.page"
      :page-size="errorPagination.pageSize"
      :visible-column-keys="visibleErrorColumnKeys"
      @sort="onErrorSort"
      @update:page="onErrorPage"
      @update:page-size="onErrorPageSize"
      @ip-geo-batch-failed="showIpGeoFailure"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { usageAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import UsageTable from '@/components/admin/usage/UsageTable.vue'
import UserErrorRequestsTable from '@/components/user/UserErrorRequestsTable.vue'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { formatReasoningEffort } from '@/utils/format'
import { toLocalDateTimeParam } from '@/utils/dashboardTimeRange'
import type { Column } from '@/components/common/types'
import type {
  ApiKey,
  Group,
  UsageLog,
  UserErrorRequest,
  UserUsagePeriod,
} from '@/types'

const props = defineProps<{
  startDate: string
  endDate: string
  startTime: string
  endTime: string
  period: UserUsagePeriod | null
  timezone: string
  apiKeys: ApiKey[]
  groups: Group[]
  models: string[]
  errorViewEnabled: boolean
  refreshKey?: number
}>()

const { t } = useI18n()
const appStore = useAppStore()
const activeTab = ref<'usage' | 'errors'>('usage')
const usageRows = ref<UsageLog[]>([])
const errorRows = ref<UserErrorRequest[]>([])
const usageLoading = ref(false)
const errorLoading = ref(false)
const exporting = ref(false)

const usageFilters = reactive<{ api_key_id: number | null; group_id: number | null; model: string | null }>({
  api_key_id: null,
  group_id: null,
  model: null,
})
const errorFilters = reactive<{ api_key_id: number | null; model: string | null; category: string; status_code: number | null }>({
  api_key_id: null,
  model: null,
  category: '',
  status_code: null,
})
const usagePagination = reactive({ page: 1, pageSize: getPersistedPageSize(), total: 0 })
const errorPagination = reactive({ page: 1, pageSize: getPersistedPageSize(), total: 0 })
const usageSort = reactive({ by: 'created_at', order: 'desc' as 'asc' | 'desc' })
const errorSort = reactive({ by: 'created_at', order: 'desc' as 'asc' | 'desc' })

const rangeParams = computed(() => props.period
  ? { period: props.period }
  : {
      start_time: toLocalDateTimeParam(props.startDate, props.startTime),
      end_time: toLocalDateTimeParam(props.endDate, props.endTime),
    })

const apiKeyOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('usage.allApiKeys') },
  ...props.apiKeys.map((key) => ({ value: key.id, label: key.name })),
])
const groupOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allGroups') },
  ...props.groups.map((group) => ({ value: group.id, label: group.name })),
])
const modelOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allModels') },
  ...props.models.map((model) => ({ value: model, label: model })),
])
const errorCategoryOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('usage.errors.allCategories') },
  ...['auth', 'rate_limit', 'quota', 'invalid_request', 'service_unavailable', 'upstream', 'internal', 'other', 'cyber']
    .map((category) => ({ value: category, label: t(`usage.errors.categories.${category}`) })),
])
const errorStatusOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('usage.errors.allStatuses') },
  ...[400, 401, 403, 404, 429, 500, 502, 503].map((status) => ({ value: status, label: String(status) })),
])

const loadUsage = async () => {
  usageLoading.value = true
  try {
    const response = await usageAPI.query({
      page: usagePagination.page,
      page_size: usagePagination.pageSize,
      ...rangeParams.value,
      timezone: props.timezone,
      api_key_id: usageFilters.api_key_id ?? undefined,
      group_id: usageFilters.group_id ?? undefined,
      model: usageFilters.model?.trim() || undefined,
      sort_by: usageSort.by,
      sort_order: usageSort.order,
    })
    usageRows.value = response.items
    usagePagination.total = response.total
  } catch (error) {
    console.error('[LegacyUsageLogPanel] failed to load usage:', error)
    appStore.showError(t('usage.failedToLoad'))
  } finally {
    usageLoading.value = false
  }
}

const loadErrors = async () => {
  if (!props.errorViewEnabled) return
  errorLoading.value = true
  try {
    const response = await usageAPI.listMyErrorRequests({
      page: errorPagination.page,
      page_size: errorPagination.pageSize,
      ...(props.period
        ? { period: props.period }
        : {
            start_time: toLocalDateTimeParam(props.startDate, props.startTime),
            end_time: toLocalDateTimeParam(props.endDate, props.endTime),
          }),
      timezone: props.timezone,
      api_key_id: errorFilters.api_key_id ?? undefined,
      model: errorFilters.model?.trim() || undefined,
      category: errorFilters.category || undefined,
      status_code: errorFilters.status_code ?? undefined,
      sort_by: errorSort.by,
      sort_order: errorSort.order,
    })
    errorRows.value = response.items
    errorPagination.total = response.total
  } catch (error) {
    console.error('[LegacyUsageLogPanel] failed to load errors:', error)
    appStore.showError(t('usage.errors.failedToLoad'))
  } finally {
    errorLoading.value = false
  }
}

const currentLoading = computed(() => activeTab.value === 'usage' ? usageLoading.value : errorLoading.value)
const refresh = () => activeTab.value === 'usage' ? void loadUsage() : void loadErrors()
const applyUsageFilters = () => { usagePagination.page = 1; void loadUsage() }
const applyErrorFilters = () => { errorPagination.page = 1; void loadErrors() }
const resetFilters = () => {
  if (activeTab.value === 'usage') {
    Object.assign(usageFilters, { api_key_id: null, group_id: null, model: null })
    applyUsageFilters()
  } else {
    Object.assign(errorFilters, { api_key_id: null, model: null, category: '', status_code: null })
    applyErrorFilters()
  }
}
const openErrors = () => {
  activeTab.value = 'errors'
  if (errorRows.value.length === 0) void loadErrors()
}
const onUsageSort = (key: string, order: 'asc' | 'desc') => {
  usageSort.by = key
  usageSort.order = order
  usagePagination.page = 1
  void loadUsage()
}
const onUsagePage = (page: number) => { usagePagination.page = page; void loadUsage() }
const onUsagePageSize = (pageSize: number) => { usagePagination.pageSize = pageSize; usagePagination.page = 1; void loadUsage() }
const onErrorSort = (key: string, order: 'asc' | 'desc') => { errorSort.by = key; errorSort.order = order; errorPagination.page = 1; void loadErrors() }
const onErrorPage = (page: number) => { errorPagination.page = page; void loadErrors() }
const onErrorPageSize = (pageSize: number) => { errorPagination.pageSize = pageSize; errorPagination.page = 1; void loadErrors() }
const showIpGeoFailure = () => appStore.showError(t('usage.ipGeo.batchFailed'))

const usageColumns = computed<Column[]>(() => [
  { key: 'api_key', label: t('usage.apiKeyFilter') },
  { key: 'model', label: t('usage.model'), sortable: true },
  { key: 'reasoning_effort', label: t('usage.reasoningEffort') },
  { key: 'endpoint', label: t('usage.endpoint') },
  { key: 'ip_address', label: 'IP' },
  { key: 'group', label: t('admin.usage.group') },
  { key: 'stream', label: t('usage.type') },
  { key: 'billing_mode', label: t('admin.usage.billingMode') },
  { key: 'tokens', label: t('usage.tokens') },
  { key: 'cache_hit_rate', label: t('usage.cacheHitRate') },
  { key: 'cost', label: t('usage.cost') },
  { key: 'latency', label: t('usage.latency') },
  { key: 'created_at', label: t('usage.time'), sortable: true },
  { key: 'user_agent', label: t('usage.userAgent') },
])
const errorColumns = computed<Column[]>(() => [
  { key: 'key_name', label: t('usage.errors.keyName') },
  { key: 'model', label: t('usage.errors.model') },
  { key: 'endpoint', label: t('usage.errors.endpoint') },
  { key: 'client_ip', label: 'IP' },
  { key: 'group', label: t('admin.usage.group') },
  { key: 'type', label: t('usage.type') },
  { key: 'platform', label: t('usage.errors.platform') },
  { key: 'category', label: t('usage.errors.category') },
  { key: 'status', label: t('usage.errors.status') },
  { key: 'message', label: t('usage.errors.message') },
  { key: 'created_at', label: t('usage.errors.time') },
  { key: 'user_agent', label: t('usage.userAgent') },
])

const hiddenUsageColumns = reactive(new Set<string>())
const hiddenErrorColumns = reactive(new Set<string>())
const alwaysVisibleUsage = new Set(['created_at'])
const alwaysVisibleError = new Set(['status', 'created_at'])
const usageStorageKey = 'user-usage-hidden-columns'
const errorStorageKey = 'user-usage-error-hidden-columns'
const usageDefaultsVersionKey = 'user-usage-columns-defaults-v2'
const defaultHiddenUsageColumns = ['user_agent', 'reasoning_effort', 'ip_address', 'stream']
const visibleUsageColumns = computed(() => usageColumns.value.filter((column) => alwaysVisibleUsage.has(column.key) || !hiddenUsageColumns.has(column.key)))
const visibleErrorColumnKeys = computed(() => errorColumns.value.filter((column) => alwaysVisibleError.has(column.key) || !hiddenErrorColumns.has(column.key)).map((column) => column.key))
const currentToggleableColumns = computed(() => (activeTab.value === 'usage' ? usageColumns.value : errorColumns.value).filter((column) => !(activeTab.value === 'usage' ? alwaysVisibleUsage : alwaysVisibleError).has(column.key)))
const isCurrentColumnVisible = (key: string) => !(activeTab.value === 'usage' ? hiddenUsageColumns : hiddenErrorColumns).has(key)
const toggleCurrentColumn = (key: string) => {
  const target = activeTab.value === 'usage' ? hiddenUsageColumns : hiddenErrorColumns
  const storageKey = activeTab.value === 'usage' ? usageStorageKey : errorStorageKey
  if (target.has(key)) target.delete(key)
  else target.add(key)
  localStorage.setItem(storageKey, JSON.stringify([...target]))
}
const loadColumns = () => {
  const restore = (key: string, target: Set<string>, defaults: string[]) => {
    try {
      const saved = localStorage.getItem(key)
      const values = saved ? JSON.parse(saved) as string[] : defaults
      values.forEach((value) => target.add(value))
    } catch {
      defaults.forEach((value) => target.add(value))
    }
  }
  restore(usageStorageKey, hiddenUsageColumns, defaultHiddenUsageColumns)
  restore(errorStorageKey, hiddenErrorColumns, ['user_agent'])

  if (localStorage.getItem(usageDefaultsVersionKey) !== '1') {
    defaultHiddenUsageColumns.forEach((value) => hiddenUsageColumns.add(value))
    localStorage.setItem(usageStorageKey, JSON.stringify([...hiddenUsageColumns]))
    localStorage.setItem(usageDefaultsVersionKey, '1')
  }
}

const showColumnDropdown = ref(false)
const columnDropdownRef = ref<HTMLElement | null>(null)
const closeColumnDropdown = (event: MouseEvent) => {
  if (columnDropdownRef.value && !columnDropdownRef.value.contains(event.target as Node)) {
    showColumnDropdown.value = false
  }
}

const escapeCsv = (value: unknown): string => {
  if (value == null) return ''
  const text = String(value)
  const escaped = text.replace(/"/g, '""')
  if (/^[=+\-@\t\r]/.test(text)) return `"'${escaped}"`
  return /[,"\n\r]/.test(text) ? `"${escaped}"` : text
}
const exportToCSV = async () => {
  if (usagePagination.total === 0) {
    appStore.showWarning(t('usage.noDataToExport'))
    return
  }
  exporting.value = true
  try {
    const rows: UsageLog[] = []
    const pageSize = 100
    for (let page = 1; page <= Math.ceil(usagePagination.total / pageSize); page += 1) {
      const response = await usageAPI.query({
        page,
        page_size: pageSize,
        ...rangeParams.value,
        timezone: props.timezone,
        api_key_id: usageFilters.api_key_id ?? undefined,
        group_id: usageFilters.group_id ?? undefined,
        model: usageFilters.model?.trim() || undefined,
        sort_by: usageSort.by,
        sort_order: usageSort.order,
      })
      rows.push(...response.items)
    }
    const headers = ['Time', 'API Key', 'Model', 'Reasoning Effort', 'Input Tokens', 'Output Tokens', 'Cache Read Tokens', 'Cache Creation Tokens', 'Actual Cost', 'First Token (ms)', 'Duration (ms)']
    const body = rows.map((row) => [row.created_at, row.api_key?.name, row.model, formatReasoningEffort(row.reasoning_effort), row.input_tokens, row.output_tokens, row.cache_read_tokens, row.cache_creation_tokens, row.actual_cost, row.first_token_ms, row.duration_ms].map(escapeCsv).join(','))
    const blob = new Blob([`\uFEFF${[headers.join(','), ...body].join('\n')}`], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `usage_${props.startDate}_to_${props.endDate}.csv`
    link.click()
    URL.revokeObjectURL(url)
    appStore.showSuccess(t('usage.exportSuccess'))
  } catch (error) {
    console.error('[LegacyUsageLogPanel] CSV export failed:', error)
    appStore.showError(t('usage.exportFailed'))
  } finally {
    exporting.value = false
  }
}

watch(() => [props.startDate, props.endDate, props.startTime, props.endTime, props.period, props.timezone], () => {
  usagePagination.page = 1
  errorPagination.page = 1
  refresh()
})
watch(() => props.refreshKey, () => refresh())
watch(() => props.errorViewEnabled, (enabled) => {
  if (!enabled && activeTab.value === 'errors') activeTab.value = 'usage'
})

onMounted(() => {
  loadColumns()
  document.addEventListener('click', closeColumnDropdown)
  void loadUsage()
})
onUnmounted(() => document.removeEventListener('click', closeColumnDropdown))
</script>
