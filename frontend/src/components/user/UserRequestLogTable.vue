<template>
  <div class="flex min-h-[360px] flex-col overflow-hidden rounded-md border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
    <DataTable
      :columns="columns"
      :data="rows"
      :loading="loading"
      :row-key="rowKey"
      density="compact"
      :estimate-row-height="72"
      :server-side-sort="true"
      :clickable-rows="true"
      default-sort-key="created_at"
      default-sort-order="desc"
      @sort="handleSort"
      @row-click="openDetail"
    >
      <template #cell-created_at="{ row }">
        <div class="min-w-[152px] space-y-1">
          <div class="tabular-nums text-gray-900 dark:text-white">{{ formatDateTime(row.created_at) }}</div>
          <span :class="kindBadgeClass(row.kind)" class="inline-flex rounded px-1.5 py-0.5 text-[11px] font-medium">
            {{ kindLabel(row.kind) }}
          </span>
        </div>
      </template>

      <template #cell-key="{ row }">
        <div class="min-w-[170px] space-y-1">
          <div class="flex items-center gap-1.5">
            <Icon name="key" size="xs" class="shrink-0 text-gray-400" />
            <span class="max-w-[210px] truncate font-medium text-gray-900 dark:text-white" :title="row.api_key_name || '-'">
              {{ row.api_key_name || '-' }}
            </span>
            <span
              v-if="row.api_key_deleted"
              class="rounded bg-gray-100 px-1 py-0.5 text-[10px] text-gray-500 dark:bg-dark-700 dark:text-dark-400"
            >
              {{ t('usage.errors.keyDeleted') }}
            </span>
          </div>
          <div class="flex flex-wrap items-center gap-1.5 text-xs text-gray-500 dark:text-dark-400">
            <span>{{ row.group_name || t('usage.logs.noGroup') }}</span>
            <span aria-hidden="true">·</span>
            <span class="tabular-nums">{{ multiplierLabel(row.rate_multiplier) }}</span>
          </div>
        </div>
      </template>

      <template #cell-model="{ row }">
        <div class="min-w-[150px] space-y-1">
          <div class="max-w-[240px] break-all font-medium text-gray-900 dark:text-white">
            {{ row.model || '-' }}
          </div>
          <span
            v-if="reasoningLabel(row.reasoning_effort)"
            class="inline-flex rounded bg-orange-50 px-1.5 py-0.5 text-[11px] font-medium text-orange-700 dark:bg-orange-500/10 dark:text-orange-300"
          >
            {{ reasoningLabel(row.reasoning_effort) }}
          </span>
        </div>
      </template>

      <template #cell-duration_ms="{ row }">
        <div class="grid min-w-[120px] grid-cols-[max-content_max-content] gap-x-2 gap-y-0.5 text-xs">
          <span class="text-gray-400 dark:text-dark-500">{{ t('usage.latencyFirstToken') }}</span>
          <span class="tabular-nums text-gray-700 dark:text-dark-200">{{ formatDuration(row.first_token_ms) }}</span>
          <span class="text-gray-400 dark:text-dark-500">{{ t('usage.latencyDuration') }}</span>
          <span class="font-medium tabular-nums text-gray-900 dark:text-white">{{ formatDuration(row.duration_ms) }}</span>
        </div>
      </template>

      <template #cell-details="{ row }">
        <button
          type="button"
          data-test="request-log-details"
          :data-kind="row.kind"
          class="group flex min-w-[170px] items-center justify-between gap-3 text-left focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500"
          :title="t('usage.logs.openDetail')"
          @click.stop="openDetail(row)"
        >
          <span v-if="row.kind === 'error'" class="min-w-0 space-y-0.5">
            <span class="flex items-center gap-2">
              <span class="badge" :class="statusBadgeClass(row.status_code)">{{ row.status_code || '-' }}</span>
              <span v-if="row.error_code" class="truncate font-mono text-xs text-red-600 dark:text-red-300">
                {{ row.error_code }}
              </span>
            </span>
            <span class="block max-w-[280px] truncate text-xs text-gray-500 dark:text-dark-400">
              {{ row.error_message || t('usage.logs.errorFallback') }}
            </span>
          </span>
          <span v-else class="space-y-0.5">
            <span class="block font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">
              {{ formatCurrency(row.actual_cost) }}
            </span>
            <span class="block text-xs tabular-nums text-gray-500 dark:text-dark-400">
              {{ formatTokenCount(row.total_tokens) }} {{ t('usage.tokens') }}
            </span>
          </span>
          <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-gray-400 transition-colors group-hover:bg-primary-50 group-hover:text-primary-600 dark:group-hover:bg-primary-500/10 dark:group-hover:text-primary-300">
            <Icon name="eye" size="sm" />
          </span>
        </button>
      </template>

      <template #empty>
        <div class="flex flex-col items-center py-4">
          <Icon name="inbox" size="xl" class="mb-3 text-gray-300 dark:text-dark-600" />
          <p class="font-medium text-gray-700 dark:text-dark-200">{{ t('usage.logs.empty') }}</p>
        </div>
      </template>
    </DataTable>

    <UserErrorDetailModal
      v-model:show="showErrorDetail"
      :error-id="selectedErrorId"
    />
    <UserUsageDetailModal
      v-model:show="showUsageDetail"
      :log="selectedUsageLog"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DataTable from '@/components/common/DataTable.vue'
import Icon from '@/components/icons/Icon.vue'
import UserErrorDetailModal from '@/components/user/UserErrorDetailModal.vue'
import UserUsageDetailModal from '@/components/user/UserUsageDetailModal.vue'
import { formatCurrency, formatDateTime, formatReasoningEffort, formatTokenCount } from '@/utils/format'
import { formatMultiplier } from '@/utils/formatters'
import type { Column } from '@/components/common/types'
import type { UserRequestLog, UserRequestLogKind } from '@/types'

defineProps<{
  rows: UserRequestLog[]
  loading: boolean
}>()

const emit = defineEmits<{
  (event: 'sort', key: 'created_at' | 'duration_ms', order: 'asc' | 'desc'): void
}>()

const { t } = useI18n()
const selectedErrorId = ref<number | null>(null)
const selectedUsageLog = ref<UserRequestLog | null>(null)
const showErrorDetail = ref(false)
const showUsageDetail = ref(false)

const columns = computed<Column[]>(() => [
  { key: 'created_at', label: t('usage.logs.columns.time'), sortable: true },
  { key: 'key', label: t('usage.logs.columns.key') },
  { key: 'model', label: t('usage.logs.columns.model') },
  { key: 'duration_ms', label: t('usage.logs.columns.latency'), sortable: true },
  { key: 'details', label: t('usage.logs.columns.details') },
])

const rowKey = (row: UserRequestLog): string => `${row.kind}-${row.id}-${row.request_id}`

const handleSort = (key: string, order: 'asc' | 'desc') => {
  if (key === 'created_at' || key === 'duration_ms') emit('sort', key, order)
}

const openDetail = (row: UserRequestLog) => {
  if (row.kind === 'error') {
    selectedErrorId.value = row.id
    showErrorDetail.value = true
    return
  }
  selectedUsageLog.value = row
  showUsageDetail.value = true
}

const kindLabel = (kind: UserRequestLogKind): string =>
  kind === 'error' ? t('usage.logs.kinds.error') : t('usage.logs.kinds.consumption')

const kindBadgeClass = (kind: UserRequestLogKind): string =>
  kind === 'error'
    ? 'bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-300'
    : 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'

const multiplierLabel = (value: number | null): string =>
  value == null ? '-' : `${formatMultiplier(value)}x`

const reasoningLabel = (value: string | null): string => {
  const formatted = formatReasoningEffort(value)
  return formatted === '-' ? '' : formatted
}

const formatDuration = (milliseconds: number | null): string => {
  if (milliseconds == null) return '-'
  if (milliseconds < 1000) return `${milliseconds}ms`
  return `${(milliseconds / 1000).toFixed(milliseconds >= 10_000 ? 1 : 2)}s`
}

const statusBadgeClass = (statusCode: number | null): string => {
  if (statusCode == null) return 'badge-gray'
  if (statusCode >= 500) return 'badge-danger'
  if (statusCode === 429) return 'badge-warning'
  return 'badge-gray'
}
</script>
