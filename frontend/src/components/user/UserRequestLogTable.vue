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
        <div class="flex min-w-[220px] items-center gap-3">
          <div class="flex min-w-0 flex-1 items-center gap-1.5">
            <Icon name="key" size="xs" class="shrink-0 text-gray-400" />
            <span class="truncate font-medium text-gray-900 dark:text-white" :title="row.api_key_name || '-'">
              {{ row.api_key_name || '-' }}
            </span>
            <span
              v-if="row.api_key_deleted"
              class="shrink-0 rounded bg-gray-100 px-1 py-0.5 text-[10px] text-gray-500 dark:bg-dark-700 dark:text-dark-400"
            >
              {{ t('usage.errors.keyDeleted') }}
            </span>
          </div>
          <span class="max-w-[120px] shrink-0 truncate rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300">
            {{ row.group_name || t('usage.logs.noGroup') }}
          </span>
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

      <template #cell-tokens="{ row }">
        <UsageTokenBreakdownCell :metrics="row" unit-style="wan-latin" />
      </template>

      <template #cell-cache_hit_rate="{ row }">
        <UsageCacheHitCell :metrics="row" unit-style="wan-latin" />
      </template>

      <template #cell-cost="{ row }">
        <div v-if="row.actual_cost != null" class="min-w-[112px] space-y-0.5">
          <span class="block font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">
            ${{ row.actual_cost.toFixed(6) }}
          </span>
          <span v-if="row.total_cost != null" class="block text-xs tabular-nums text-gray-400 dark:text-dark-500">
            {{ t('usage.original') }} ${{ row.total_cost.toFixed(6) }}
          </span>
        </div>
        <div v-else class="min-w-[112px] space-y-0.5">
          <span class="text-sm text-gray-400 dark:text-dark-500">-</span>
          <span v-if="row.kind === 'error'" class="block text-xs text-red-600 dark:text-red-300">
            {{ row.status_code || row.error_code || t('usage.logs.kinds.error') }}
          </span>
        </div>
      </template>

      <template #cell-duration_ms="{ row }">
        <UsageLatencyCell :first-token-ms="row.first_token_ms" :duration-ms="row.duration_ms" />
      </template>

      <template #empty>
        <div class="flex flex-col items-center py-4">
          <Icon name="inbox" size="xl" class="mb-3 text-gray-300 dark:text-dark-600" />
          <p class="font-medium text-gray-700 dark:text-dark-200">{{ t('usage.logs.empty') }}</p>
        </div>
      </template>
    </DataTable>

    <UserErrorDetailModal v-model:show="showErrorDetail" :error-id="selectedErrorId" />
    <UserUsageDetailModal v-model:show="showUsageDetail" :log="selectedUsageLog" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DataTable from '@/components/common/DataTable.vue'
import Icon from '@/components/icons/Icon.vue'
import UserErrorDetailModal from '@/components/user/UserErrorDetailModal.vue'
import UserUsageDetailModal from '@/components/user/UserUsageDetailModal.vue'
import UsageCacheHitCell from '@/components/usage/UsageCacheHitCell.vue'
import UsageLatencyCell from '@/components/usage/UsageLatencyCell.vue'
import UsageTokenBreakdownCell from '@/components/usage/UsageTokenBreakdownCell.vue'
import { formatDateTime, formatReasoningEffort } from '@/utils/format'
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
  { key: 'tokens', label: t('usage.tokens') },
  { key: 'cache_hit_rate', label: t('usage.cacheHitRate') },
  { key: 'cost', label: t('usage.cost') },
  { key: 'duration_ms', label: t('usage.logs.columns.latency'), sortable: true },
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

const reasoningLabel = (value: string | null): string => {
  const formatted = formatReasoningEffort(value)
  return formatted === '-' ? '' : formatted
}
</script>
