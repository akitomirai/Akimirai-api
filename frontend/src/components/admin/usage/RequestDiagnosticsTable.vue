<template>
  <DataTable
    :data="data"
    :columns="columns"
    :loading="loading"
    density="compact"
    row-key="id"
  >
    <template #cell-user="{ row }">
      <button
        v-if="row.user"
        type="button"
        class="text-left text-sm font-medium text-primary-600 hover:underline dark:text-primary-400"
        @click="emit('userClick', row.user.id, row.user.email)"
      >
        {{ row.user.email }}
        <span class="ml-1 text-xs font-normal text-gray-400">#{{ row.user.id }}</span>
      </button>
      <span v-else class="text-gray-400">-</span>
    </template>

    <template #cell-api_key="{ row }">
      <div v-if="row.api_key" class="min-w-[110px]">
        <div class="text-sm font-medium text-gray-900 dark:text-white">{{ row.api_key.name || `#${row.api_key.id}` }}</div>
        <div class="font-mono text-[11px] text-gray-400">#{{ row.api_key.id }}</div>
      </div>
      <span v-else class="text-gray-400">-</span>
    </template>

    <template #cell-created_at="{ row }">
      <div class="min-w-[138px] text-sm text-gray-700 dark:text-gray-300">
        {{ row.request_started_at ? formatDateTime(row.request_started_at) : t('admin.usage.diagnostics.unavailable') }}
      </div>
    </template>

    <template #cell-features="{ row }">
      <div class="min-w-[260px] space-y-1 text-xs">
        <div class="flex flex-wrap items-center gap-1.5">
          <span class="font-medium text-gray-900 dark:text-white">{{ row.model || '-' }}</span>
          <span class="rounded bg-blue-50 px-1.5 py-0.5 text-blue-700 dark:bg-blue-500/10 dark:text-blue-300">
            {{ requestTypeLabel(row) }}
          </span>
          <span v-if="row.request_body_bytes != null" class="text-gray-500 dark:text-gray-400">
            {{ formatBytes(row.request_body_bytes) }}
          </span>
        </div>
        <div class="truncate text-gray-500 dark:text-gray-400" :title="row.inbound_endpoint || ''">
          {{ row.inbound_endpoint || '-' }}
        </div>
        <div v-if="row.request_id" class="truncate font-mono text-[11px] text-gray-400" :title="row.request_id">
          {{ row.request_id }}
        </div>
      </div>
    </template>

    <template #cell-route="{ row }">
      <div class="min-w-[150px] space-y-1 text-xs">
        <div class="flex flex-wrap items-center gap-1.5">
          <span class="rounded px-1.5 py-0.5 font-medium" :class="routeBadgeClass(row.route_kind)">
            {{ routeLabel(row) }}
          </span>
          <span v-if="row.account" class="text-gray-600 dark:text-gray-300">{{ row.account.name }}</span>
        </div>
        <div v-if="row.route_fingerprint" class="font-mono text-[11px] text-gray-400" :title="row.route_fingerprint">
          {{ row.route_fingerprint.slice(0, 12) }}
        </div>
      </div>
    </template>

    <template #cell-timings="{ row }">
      <div class="grid min-w-[170px] grid-cols-[max-content_max-content] gap-x-2 gap-y-0.5 text-xs">
        <span class="text-gray-400">{{ t('admin.usage.diagnostics.firstByte') }}</span>
        <span class="font-medium tabular-nums text-gray-700 dark:text-gray-200">{{ formatDuration(row.upstream_first_byte_ms) }}</span>
        <span class="text-gray-400">{{ t('admin.usage.diagnostics.firstToken') }}</span>
        <span class="font-medium tabular-nums text-amber-600 dark:text-amber-300">{{ formatDuration(row.request_first_token_ms) }}</span>
        <span class="text-gray-400">{{ t('admin.usage.diagnostics.completed') }}</span>
        <span class="font-medium tabular-nums text-gray-700 dark:text-gray-200">{{ formatDuration(row.request_total_ms) }}</span>
      </div>
    </template>

    <template #cell-retries="{ row }">
      <div class="min-w-[92px] text-sm tabular-nums text-gray-700 dark:text-gray-300">
        {{ row.retry_count ?? 0 }} / {{ row.account_switch_count ?? 0 }}
      </div>
    </template>

    <template #cell-status="{ row }">
      <span class="inline-flex rounded px-2 py-0.5 text-xs font-medium" :class="statusBadgeClass(row.final_upstream_status)">
        {{ row.final_upstream_status ?? '-' }}
      </span>
    </template>

  </DataTable>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import DataTable from '@/components/common/DataTable.vue'
import { formatDateTime } from '@/utils/format'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import type { Column } from '@/components/common/types'
import type { AdminUsageLog, UsageRouteKind } from '@/types'

const props = withDefaults(defineProps<{
  data: AdminUsageLog[]
  loading?: boolean
  columns?: Column[]
}>(), {
  loading: false,
  columns: undefined,
})

const emit = defineEmits<{
  userClick: [userID: number, email?: string]
}>()

const { t } = useI18n()

const defaultColumns = computed<Column[]>(() => [
  { key: 'user', label: t('admin.usage.user') },
  { key: 'api_key', label: t('usage.apiKeyFilter') },
  { key: 'created_at', label: t('admin.usage.diagnostics.requestStartedAt') },
  { key: 'features', label: t('admin.usage.diagnostics.requestFeatures') },
  { key: 'route', label: t('admin.usage.diagnostics.route') },
  { key: 'timings', label: t('admin.usage.diagnostics.timings') },
  { key: 'retries', label: t('admin.usage.diagnostics.retrySwitch') },
  { key: 'status', label: t('admin.usage.diagnostics.upstreamStatus') },
])

const columns = computed<Column[]>(() => props.columns ?? defaultColumns.value)

const requestTypeLabel = (row: AdminUsageLog): string => {
  const requestType = resolveUsageRequestType(row)
  if (requestType === 'cyber') return t('usage.cyber')
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

const formatBytes = (value: number): string => {
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  return `${(value / (1024 * 1024)).toFixed(1)} MB`
}

const formatDuration = (value: number | null | undefined): string => {
  if (value == null) return t('admin.usage.diagnostics.unavailable')
  if (value < 1000) return `${value}ms`
  if (value < 60_000) return `${(value / 1000).toFixed(2)}s`
  const seconds = Math.round(value / 1000)
  return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
}

const routeLabel = (row: AdminUsageLog): string => {
  if (row.route_kind === 'proxy') {
    const name = row.proxy_name_snapshot || t('admin.usage.diagnostics.proxyFallback')
    return row.proxy_id_snapshot ? `${name} #${row.proxy_id_snapshot}` : name
  }
  if (row.route_kind === 'direct') return t('admin.usage.diagnostics.directRoute')
  return t('admin.usage.diagnostics.unavailable')
}

const routeBadgeClass = (kind: UsageRouteKind | null | undefined): string => {
  if (kind === 'proxy') return 'bg-sky-100 text-sky-700 dark:bg-sky-500/20 dark:text-sky-300'
  if (kind === 'direct') return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
}

const statusBadgeClass = (status: number | null | undefined): string => {
  if (status == null) return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
  if (status >= 200 && status < 300) return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300'
  if (status >= 400) return 'bg-rose-100 text-rose-700 dark:bg-rose-500/20 dark:text-rose-300'
  return 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300'
}
</script>
