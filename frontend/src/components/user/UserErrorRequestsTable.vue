<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div class="card flex min-h-0 flex-1 flex-col overflow-hidden">
      <IpGeoBatchToolbar :ips="rows.map((row) => row.client_ip)" @failed="emit('ipGeoBatchFailed')" />

      <DataTable
        :columns="columns"
        :data="rows"
        :loading="loading"
        clickable-rows
        server-side-sort
        default-sort-key="created_at"
        default-sort-order="desc"
        @sort="onSort"
        @row-click="(row) => openDetail(row.id)"
      >
        <template #cell-model="{ row }">
          <span v-if="row.model" class="text-sm font-medium text-gray-900 dark:text-white">{{ row.model }}</span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-key_name="{ row }">
          <div class="text-sm">
            <span class="text-gray-900 dark:text-white">{{ row.key_name || '-' }}</span>
            <span
              v-if="row.key_deleted"
              class="ml-1 inline-flex items-center rounded bg-rose-100 px-1 py-px text-[10px] font-medium leading-tight text-rose-600 ring-1 ring-inset ring-rose-200 dark:bg-rose-500/20 dark:text-rose-400 dark:ring-rose-500/30"
            >{{ t('usage.errors.keyDeleted') }}</span>
          </div>
        </template>

        <template #cell-endpoint="{ row }">
          <div class="max-w-[320px] text-xs">
            <div class="break-all text-gray-700 dark:text-gray-300">
              <span class="font-medium text-gray-500 dark:text-gray-400">{{ t('usage.inbound') }}:</span>
              <span class="ml-1">{{ row.inbound_endpoint?.trim() || '-' }}</span>
            </div>
          </div>
        </template>

        <template #cell-status="{ row }">
          <span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="statusClass(row.status_code)">
            {{ row.status_code || '-' }}
          </span>
        </template>

        <template #cell-category="{ row }">
          <span class="text-sm text-gray-900 dark:text-white">{{ t(`usage.errors.categories.${row.category}`) }}</span>
        </template>

        <template #cell-message="{ row }">
          <span
            v-if="row.message"
            class="block max-w-[280px] truncate text-sm text-gray-600 dark:text-gray-400"
            :title="row.message"
          >{{ row.message }}</span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-group="{ row }">
          <span
            v-if="row.group_name"
            class="inline-flex items-center rounded bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200"
          >{{ row.group_name }}</span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-type="{ row }">
          <span
            v-if="requestTypeBadge(row)"
            class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium"
            :class="requestTypeBadge(row)!.className"
          >{{ requestTypeBadge(row)!.label }}</span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-platform="{ row }">
          <span class="text-sm text-gray-900 dark:text-white">{{ row.platform || '-' }}</span>
        </template>

        <template #cell-client_ip="{ row }">
          <div @click.stop>
            <div v-if="row.client_ip">
              <span class="font-mono text-sm text-gray-600 dark:text-gray-400">{{ row.client_ip }}</span>
              <IpGeoCell :ip="row.client_ip" />
            </div>
            <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
          </div>
        </template>

        <template #cell-created_at="{ row }">
          <span class="text-sm text-gray-600 dark:text-gray-400">{{ formatDateTime(row.created_at) }}</span>
        </template>

        <template #cell-user_agent="{ row }">
          <span
            v-if="row.user_agent"
            class="block max-w-[320px] truncate text-sm text-gray-600 dark:text-gray-400"
            :title="row.user_agent"
          >{{ row.user_agent }}</span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #empty><EmptyState :message="t('usage.errors.empty')" /></template>
      </DataTable>
    </div>

    <Pagination
      v-if="total > 0"
      :page="page"
      :page-size="pageSize"
      :total="total"
      @update:page="emit('update:page', $event)"
      @update:page-size="emit('update:pageSize', $event)"
    />

    <UserErrorDetailModal v-model:show="showDetail" :error-id="selectedId" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Pagination from '@/components/common/Pagination.vue'
import UserErrorDetailModal from '@/components/user/UserErrorDetailModal.vue'
import IpGeoCell from '@/components/common/IpGeoCell.vue'
import IpGeoBatchToolbar from '@/components/common/IpGeoBatchToolbar.vue'
import { formatDateTime } from '@/utils/format'
import {
  mapErrorSortKey,
  numericRequestTypeKind,
  requestTypeBadgeClass,
  requestTypeLabelKey,
  statusCodeBadgeClass,
} from '@/utils/errorBadges'
import type { UserErrorRequest } from '@/types'
import type { Column } from '@/components/common/types'

const props = defineProps<{
  rows: UserErrorRequest[]
  total: number
  loading: boolean
  page: number
  pageSize: number
  visibleColumnKeys?: string[]
}>()

const emit = defineEmits<{
  (event: 'update:page', value: number): void
  (event: 'update:pageSize', value: number): void
  (event: 'ipGeoBatchFailed'): void
  (event: 'sort', sortBy: string, sortOrder: 'asc' | 'desc'): void
}>()

const { t } = useI18n()
const showDetail = ref(false)
const selectedId = ref<number | null>(null)

const allColumns = computed<Column[]>(() => [
  { key: 'key_name', label: t('usage.errors.keyName') },
  { key: 'model', label: t('usage.errors.model'), sortable: true },
  { key: 'endpoint', label: t('usage.errors.endpoint') },
  { key: 'client_ip', label: 'IP' },
  { key: 'group', label: t('admin.usage.group') },
  { key: 'type', label: t('usage.type') },
  { key: 'platform', label: t('usage.errors.platform') },
  { key: 'category', label: t('usage.errors.category') },
  { key: 'status', label: t('usage.errors.status'), sortable: true },
  { key: 'message', label: t('usage.errors.message') },
  { key: 'created_at', label: t('usage.errors.time'), sortable: true },
  { key: 'user_agent', label: t('usage.userAgent') },
])

const columns = computed(() => props.visibleColumnKeys
  ? allColumns.value.filter((column) => props.visibleColumnKeys!.includes(column.key))
  : allColumns.value)

const onSort = (key: string, order: 'asc' | 'desc') => emit('sort', mapErrorSortKey(key), order)

const requestTypeBadge = (row: UserErrorRequest): { label: string; className: string } | null => {
  const kind = numericRequestTypeKind(row.request_type, row.stream)
  if (!kind) return null
  return { label: t(requestTypeLabelKey(kind)), className: requestTypeBadgeClass(kind) }
}

const openDetail = (id: number) => {
  selectedId.value = id
  showDetail.value = true
}

const statusClass = statusCodeBadgeClass
</script>
