<template>
  <div class="grid gap-4 xl:grid-cols-3">
    <section class="card p-4">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.commercial.balanceTitle') }}</p>
          <p v-if="hasBalance" class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">
            ${{ formatMoney(balanceValue) }}
          </p>
          <p v-else class="mt-2 text-sm text-gray-500 dark:text-dark-400">
            {{ t('dashboard.commercial.noBalance') }}
          </p>
        </div>
        <div class="rounded-lg bg-emerald-100 p-2 dark:bg-emerald-900/30">
          <Icon name="dollar" class="text-emerald-600 dark:text-emerald-400" />
        </div>
      </div>
      <div class="mt-4 flex flex-wrap gap-2">
        <router-link v-if="paymentEnabled" class="btn btn-sm btn-primary" to="/purchase">
          <Icon name="creditCard" size="sm" />
          {{ t('dashboard.commercial.recharge') }}
        </router-link>
        <router-link class="btn btn-sm btn-secondary" to="/redeem">
          <Icon name="gift" size="sm" />
          {{ t('dashboard.commercial.redeem') }}
        </router-link>
      </div>
    </section>

    <section class="card p-4">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.commercial.todayUsageTitle') }}</p>
          <template v-if="stats">
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              ${{ formatMoney(stats.today_actual_cost || 0) }}
            </p>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
              {{ formatNumber(stats.today_tokens || 0) }} tokens · {{ formatNumber(stats.today_requests || 0) }} {{ t('dashboard.commercial.requests') }}
            </p>
          </template>
          <p v-else class="mt-2 text-sm text-gray-500 dark:text-dark-400">
            {{ t('dashboard.commercial.noTodayStats') }}
          </p>
        </div>
        <div class="rounded-lg bg-blue-100 p-2 dark:bg-blue-900/30">
          <Icon name="chart" class="text-blue-600 dark:text-blue-400" />
        </div>
      </div>
      <router-link class="btn btn-sm btn-secondary mt-4" to="/usage">
        <Icon name="externalLink" size="sm" />
        {{ t('dashboard.commercial.viewUsage') }}
      </router-link>
    </section>

    <section class="card p-4">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.commercial.keyStatusTitle') }}</p>
          <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
            {{ apiKeyCount }}
          </p>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ activeApiKeyCount }} {{ t('common.active') }}
          </p>
        </div>
        <div class="rounded-lg bg-indigo-100 p-2 dark:bg-indigo-900/30">
          <Icon name="key" class="text-indigo-600 dark:text-indigo-400" />
        </div>
      </div>
      <div class="mt-3 text-xs text-gray-500 dark:text-dark-400">
        <span v-if="apiKeysLoading">{{ t('common.loading') }}</span>
        <span v-else-if="apiKeysError">{{ t('dashboard.commercial.keyLoadFailed') }}</span>
        <span v-else-if="lastUsedAt">{{ t('dashboard.commercial.lastUsedAt', { time: formatDateTime(lastUsedAt) }) }}</span>
        <span v-else>{{ t('dashboard.commercial.noKeyUsage') }}</span>
      </div>
      <div class="mt-4 flex flex-wrap gap-2">
        <router-link class="btn btn-sm btn-primary" to="/keys">
          <Icon name="plus" size="sm" />
          {{ t('dashboard.commercial.createKey') }}
        </router-link>
        <router-link class="btn btn-sm btn-secondary" to="/keys">
          <Icon name="key" size="sm" />
          {{ t('dashboard.commercial.manageKeys') }}
        </router-link>
      </div>
    </section>

    <section class="card p-4 xl:col-span-2">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.commercial.quickStartTitle') }}</p>
          <p class="mt-2 text-sm text-gray-600 dark:text-dark-300">
            {{ t('dashboard.commercial.quickStartDescription') }}
          </p>
        </div>
        <button class="btn btn-sm btn-secondary" type="button" @click="copyBaseUrl">
          <Icon name="copy" size="sm" />
          {{ t('dashboard.commercial.copyBaseUrl') }}
        </button>
      </div>
      <div class="mt-4 grid gap-3 md:grid-cols-2">
        <div class="rounded-lg border border-gray-200 p-3 dark:border-dark-700">
          <p class="text-xs font-medium uppercase text-gray-500 dark:text-dark-400">Base URL</p>
          <code class="mt-2 block truncate rounded bg-gray-50 px-2 py-1.5 font-mono text-xs text-gray-900 dark:bg-dark-800 dark:text-dark-100">
            {{ baseUrl }}
          </code>
        </div>
        <div class="rounded-lg border border-gray-200 p-3 dark:border-dark-700">
          <p class="text-xs font-medium uppercase text-gray-500 dark:text-dark-400">Model</p>
          <code class="mt-2 block truncate rounded bg-gray-50 px-2 py-1.5 font-mono text-xs text-gray-900 dark:bg-dark-800 dark:text-dark-100">
            {{ recommendedModel || '<MODEL_NAME>' }}
          </code>
        </div>
      </div>
      <div class="mt-4 flex flex-wrap gap-2">
        <router-link class="btn btn-sm btn-primary" to="/quick-start">
          <Icon name="book" size="sm" />
          {{ t('dashboard.commercial.viewQuickStart') }}
        </router-link>
        <router-link class="btn btn-sm btn-secondary" to="/keys">
          <Icon name="key" size="sm" />
          {{ t('dashboard.commercial.createKey') }}
        </router-link>
        <router-link class="btn btn-sm btn-secondary" to="/available-channels">
          <Icon name="server" size="sm" />
          {{ t('dashboard.commercial.viewModels') }}
        </router-link>
        <router-link class="btn btn-sm btn-secondary" to="/usage">
          <Icon name="chart" size="sm" />
          {{ t('dashboard.commercial.viewUsage') }}
        </router-link>
      </div>
    </section>

    <section class="card p-4">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.commercial.paymentTitle') }}</p>
          <p class="mt-2 text-sm text-gray-600 dark:text-dark-300">
            {{ paymentEnabled ? t('dashboard.commercial.paymentEnabled') : t('dashboard.commercial.paymentDisabled') }}
          </p>
        </div>
        <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
          <Icon name="creditCard" class="text-amber-600 dark:text-amber-400" />
        </div>
      </div>
      <div class="mt-4 flex flex-wrap gap-2">
        <router-link v-if="paymentEnabled" class="btn btn-sm btn-primary" to="/purchase">
          {{ t('dashboard.commercial.recharge') }}
        </router-link>
        <router-link v-if="paymentEnabled" class="btn btn-sm btn-secondary" to="/orders">
          {{ t('dashboard.commercial.orders') }}
        </router-link>
        <router-link class="btn btn-sm btn-secondary" to="/subscriptions">
          {{ t('dashboard.commercial.subscriptions') }}
        </router-link>
      </div>
    </section>

    <section class="card p-4 xl:col-span-3">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.commercial.recentErrorsTitle') }}</p>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ t('dashboard.commercial.recentErrorsDescription') }}
          </p>
        </div>
        <router-link class="btn btn-sm btn-secondary" to="/usage?tab=errors">
          <Icon name="externalLink" size="sm" />
          {{ t('dashboard.commercial.viewErrors') }}
        </router-link>
      </div>

      <div v-if="errorsLoading" class="py-6 text-center text-sm text-gray-500 dark:text-dark-400">
        {{ t('common.loading') }}
      </div>
      <div v-else-if="!errorViewEnabled" class="py-6 text-center text-sm text-gray-500 dark:text-dark-400">
        {{ t('dashboard.commercial.errorViewDisabled') }}
      </div>
      <div v-else-if="errorsError" class="py-6 text-center text-sm text-red-600 dark:text-red-300">
        {{ t('dashboard.commercial.errorLoadFailed') }}
      </div>
      <div v-else-if="recentErrors.length === 0" class="py-6 text-center text-sm text-gray-500 dark:text-dark-400">
        {{ t('dashboard.commercial.noRecentErrors') }}
      </div>
      <div v-else class="mt-4 space-y-3">
        <article
          v-for="error in recentErrors"
          :key="error.id"
          class="rounded-lg border border-gray-200 p-3 dark:border-dark-700"
        >
          <div class="flex flex-wrap items-center gap-2">
            <span class="rounded bg-red-50 px-2 py-1 font-mono text-xs font-medium text-red-700 dark:bg-red-900/20 dark:text-red-300">
              {{ error.error_code || '-' }}
            </span>
            <span class="rounded px-2 py-1 text-xs font-medium" :class="error.retryable ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-300' : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'">
              {{ error.retryable ? t('usage.errors.retryable') : t('usage.errors.notRetryable') }}
            </span>
            <span class="rounded px-2 py-1 text-xs font-medium" :class="error.charged ? 'bg-yellow-50 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-300' : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'">
              {{ error.charged ? t('usage.errors.charged') : t('usage.errors.notCharged') }}
            </span>
          </div>
          <p class="mt-2 text-sm text-gray-900 dark:text-white">
            {{ error.explanation || '-' }}
          </p>
          <p v-if="error.suggestion" class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ error.suggestion }}
          </p>
          <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-dark-400">
            <span>{{ t('dashboard.commercial.requestId') }}: <span class="font-mono">{{ error.request_id || '-' }}</span></span>
            <span>{{ formatDateTime(error.created_at) }}</span>
            <router-link class="text-primary-600 hover:text-primary-700 dark:text-primary-400" to="/usage?tab=errors">
              {{ t('dashboard.commercial.details') }}
            </router-link>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { formatDateTime } from '@/utils/format'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'
import type { ApiKey, UserErrorRequest } from '@/types'

const props = defineProps<{
  stats: UserStatsType | null
  balance: number | null
  apiKeys: ApiKey[]
  apiKeysLoading: boolean
  apiKeysError: boolean
  recentErrors: UserErrorRequest[]
  errorsLoading: boolean
  errorsError: boolean
  errorViewEnabled: boolean
  baseUrl: string
  recommendedModel: string
  paymentEnabled: boolean
}>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const hasBalance = computed(() => typeof props.balance === 'number' && Number.isFinite(props.balance))
const balanceValue = computed(() => props.balance ?? 0)
const apiKeyCount = computed(() => props.stats?.total_api_keys ?? props.apiKeys.length)
const activeApiKeyCount = computed(() =>
  props.stats?.active_api_keys ?? props.apiKeys.filter((key) => key.status === 'active').length
)

const lastUsedAt = computed(() => {
  return props.apiKeys
    .map((key) => key.last_used_at)
    .filter((value): value is string => !!value)
    .sort((a, b) => new Date(b).getTime() - new Date(a).getTime())[0] || ''
})

function formatMoney(value: number): string {
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 4 }).format(value)
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat('en-US').format(value)
}

function copyBaseUrl() {
  copyToClipboard(props.baseUrl)
}
</script>
