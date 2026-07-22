<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="card p-4 sm:p-6">
          <div class="flex flex-wrap items-end justify-between gap-4">
            <div class="flex flex-1 flex-wrap items-end gap-4">
              <div>
                <label class="input-label">{{ t('admin.dailyCheckIns.history') }}</label>
                <div
                  class="inline-flex h-10 items-center rounded-md border border-gray-200 bg-gray-50 p-1 dark:border-dark-600 dark:bg-dark-800"
                  role="group"
                  :aria-label="t('admin.dailyCheckIns.history')"
                >
                  <button
                    type="button"
                    data-test="date-mode-current"
                    class="h-8 rounded px-3 text-sm font-medium transition-colors"
                    :class="dateMode === 'current' ? activeModeClass : inactiveModeClass"
                    :aria-pressed="dateMode === 'current'"
                    @click="setDateMode('current')"
                  >
                    {{ t('admin.dailyCheckIns.currentDay') }}
                  </button>
                  <button
                    type="button"
                    data-test="date-mode-history"
                    class="h-8 rounded px-3 text-sm font-medium transition-colors"
                    :class="dateMode === 'history' ? activeModeClass : inactiveModeClass"
                    :aria-pressed="dateMode === 'history'"
                    @click="setDateMode('history')"
                  >
                    {{ t('admin.dailyCheckIns.allHistory') }}
                  </button>
                </div>
              </div>

              <div v-if="dateMode === 'history'" class="w-full sm:w-auto sm:min-w-[180px]">
                <label class="input-label" for="daily-check-in-service-date">
                  {{ t('admin.dailyCheckIns.serviceDate') }}
                </label>
                <input
                  id="daily-check-in-service-date"
                  v-model="serviceDate"
                  data-test="service-date"
                  type="date"
                  class="input"
                  @change="applyServiceDate"
                />
              </div>

              <div class="w-full sm:w-auto sm:min-w-[260px]">
                <label class="input-label" for="daily-check-in-query">
                  {{ t('admin.dailyCheckIns.keyword') }}
                </label>
                <div class="relative">
                  <Icon
                    name="search"
                    size="md"
                    class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
                  />
                  <input
                    id="daily-check-in-query"
                    v-model.trim="query"
                    type="search"
                    class="input pl-10"
                    :placeholder="t('admin.dailyCheckIns.keywordPlaceholder')"
                    @keyup.enter="search"
                  />
                </div>
              </div>
            </div>

            <div class="flex w-full items-center justify-end gap-3 sm:w-auto">
              <button data-test="search" type="button" class="btn btn-primary" :disabled="loading" @click="search">
                <Icon name="search" size="sm" class="mr-1.5" />
                {{ t('common.search') }}
              </button>
              <button type="button" class="btn btn-secondary" :disabled="loading" @click="resetFilters">
                <Icon name="refresh" size="sm" class="mr-1.5" />
                {{ t('common.reset') }}
              </button>
            </div>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="records" :loading="loading" row-key="id">
          <template #cell-user="{ row }">
            <div class="min-w-0 max-w-[260px] whitespace-normal">
              <div class="break-all font-medium text-gray-900 dark:text-white">{{ row.email }}</div>
              <div class="mt-0.5 break-all text-xs text-gray-500 dark:text-gray-400">
                {{ row.username || '-' }} · #{{ row.user_id }}
              </div>
            </div>
          </template>

          <template #cell-service_date="{ value }">
            <span class="whitespace-nowrap font-mono text-gray-700 dark:text-gray-300">{{ value }}</span>
          </template>

          <template #cell-reward_amount="{ value }">
            <span class="inline-flex items-center gap-1.5 rounded-full bg-green-100 px-2.5 py-1 text-sm font-semibold text-green-700 dark:bg-green-900/30 dark:text-green-300">
              <Icon name="gift" size="sm" />
              +{{ formatAmount(value) }}
            </span>
          </template>

          <template #cell-balance="{ row }">
            <div class="inline-flex items-center gap-2 whitespace-nowrap font-mono text-sm">
              <span class="text-gray-500 dark:text-gray-400">{{ formatAmount(row.balance_before) }}</span>
              <Icon name="arrowRight" size="sm" class="text-gray-400" />
              <span class="font-semibold text-gray-900 dark:text-white">{{ formatAmount(row.balance_after) }}</span>
            </div>
          </template>

          <template #cell-checked_in_at="{ value }">
            <span class="whitespace-nowrap text-gray-600 dark:text-gray-300">{{ formatDateTime(value) }}</span>
          </template>

          <template #empty>
            <div class="flex flex-col items-center py-8">
              <Icon name="calendar" size="xl" class="mb-4 h-12 w-12 text-gray-300 dark:text-dark-600" />
              <p class="text-sm font-medium text-gray-500 dark:text-gray-400">
                {{ t('admin.dailyCheckIns.empty') }}
              </p>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="total > 0"
          :total="total"
          :page="page"
          :page-size="pageSize"
          @update:page="onPageChange"
          @update:pageSize="onPageSizeChange"
        />
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { adminAPI, type DailyCheckInAdminQuery, type DailyCheckInAdminRecord } from '@/api/admin'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import { useAppStore } from '@/stores'

type DateMode = 'current' | 'history'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const records = ref<DailyCheckInAdminRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const query = ref('')
const serviceDate = ref('')
const dateMode = ref<DateMode>('current')

const activeModeClass = 'bg-white text-primary-700 shadow-sm dark:bg-dark-700 dark:text-primary-300'
const inactiveModeClass = 'text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200'

const columns = computed<Column[]>(() => [
  { key: 'user', label: t('admin.dailyCheckIns.user'), class: 'min-w-[16rem] whitespace-normal' },
  { key: 'service_date', label: t('admin.dailyCheckIns.serviceDate') },
  { key: 'reward_amount', label: t('admin.dailyCheckIns.reward') },
  { key: 'balance', label: t('admin.dailyCheckIns.balanceChange') },
  { key: 'checked_in_at', label: t('admin.dailyCheckIns.checkedInAt') }
])

function buildQuery(): DailyCheckInAdminQuery {
  const params: DailyCheckInAdminQuery = {
    page: page.value,
    page_size: pageSize.value
  }
  const keyword = query.value.trim()
  if (keyword) params.q = keyword
  if (dateMode.value === 'history') {
    if (serviceDate.value) params.service_date = serviceDate.value
    else params.all = true
  }
  return params
}

async function fetchRecords() {
  loading.value = true
  try {
    const result = await adminAPI.dailyCheckIns.list(buildQuery())
    records.value = result.items
    total.value = result.total
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.dailyCheckIns.loadFailed'))
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  fetchRecords()
}

function setDateMode(mode: DateMode) {
  if (dateMode.value === mode) return
  dateMode.value = mode
  if (mode === 'current') serviceDate.value = ''
  search()
}

function applyServiceDate() {
  search()
}

function resetFilters() {
  query.value = ''
  serviceDate.value = ''
  dateMode.value = 'current'
  search()
}

function onPageChange(nextPage: number) {
  page.value = nextPage
  fetchRecords()
}

function onPageSizeChange(nextPageSize: number) {
  pageSize.value = nextPageSize
  page.value = 1
  fetchRecords()
}

function formatAmount(value: number): string {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return '-'
  return numeric.toFixed(2).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1')
}

function formatDateTime(value: string): string {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

onMounted(fetchRecords)
</script>
