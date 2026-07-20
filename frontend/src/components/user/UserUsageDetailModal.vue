<template>
  <BaseDialog
    :show="show"
    :title="t('usage.logs.detailTitle')"
    width="wide"
    @close="emit('update:show', false)"
  >
    <div v-if="loading" class="flex justify-center py-12">
      <LoadingSpinner size="lg" />
    </div>

    <div v-else-if="loadError" class="py-10 text-center text-sm text-red-500">
      {{ t('usage.logs.detailLoadFailed') }}
    </div>

    <div v-else-if="detail" class="space-y-5 text-sm">
      <section>
        <h4 class="mb-3 font-semibold text-gray-900 dark:text-white">
          {{ t('usage.logs.requestDetails') }}
        </h4>
        <dl class="grid grid-cols-1 gap-x-8 gap-y-3 sm:grid-cols-2 lg:grid-cols-3">
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.logs.requestId') }}</dt>
            <dd class="mt-1 break-all font-mono text-xs text-gray-900 dark:text-white">
              {{ log?.request_id || '-' }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.time') }}</dt>
            <dd class="mt-1 text-gray-900 dark:text-white">{{ formatDateTime(log?.created_at) }}</dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.apiKeyFilter') }}</dt>
            <dd class="mt-1 text-gray-900 dark:text-white">{{ log?.api_key_name || '-' }}</dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.usage.group') }}</dt>
            <dd class="mt-1 text-gray-900 dark:text-white">{{ log?.group_name || '-' }}</dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.model') }}</dt>
            <dd class="mt-1 text-gray-900 dark:text-white">{{ log?.model || '-' }}</dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.reasoningEffort') }}</dt>
            <dd class="mt-1 text-gray-900 dark:text-white">
              {{ formatReasoningEffort(log?.reasoning_effort) }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.latencyFirstToken') }}</dt>
            <dd class="mt-1 tabular-nums text-gray-900 dark:text-white">
              {{ formatDuration(log?.first_token_ms) }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.latencyDuration') }}</dt>
            <dd class="mt-1 tabular-nums text-gray-900 dark:text-white">
              {{ formatDuration(log?.duration_ms) }}
            </dd>
          </div>
        </dl>
      </section>

      <section>
        <h4 class="mb-3 font-semibold text-gray-900 dark:text-white">
          {{ t('usage.tokenDetails') }}
        </h4>
        <dl class="grid grid-cols-2 gap-3 rounded-md border border-gray-200 p-4 dark:border-dark-700 sm:grid-cols-5">
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.in') }}</dt>
            <dd class="mt-1 font-medium tabular-nums text-gray-900 dark:text-white">
              {{ formatTokenCount(detail.input_tokens, { display: 'exact' }) }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.out') }}</dt>
            <dd class="mt-1 font-medium tabular-nums text-gray-900 dark:text-white">
              {{ formatTokenCount(detail.output_tokens, { display: 'exact' }) }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.cacheRead') }}</dt>
            <dd class="mt-1 font-medium tabular-nums text-gray-900 dark:text-white">
              {{ formatTokenCount(detail.cache_read_tokens, { display: 'exact' }) }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.cacheWrite') }}</dt>
            <dd class="mt-1 font-medium tabular-nums text-gray-900 dark:text-white">
              {{ formatTokenCount(detail.cache_creation_tokens, { display: 'exact' }) }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.totalTokens') }}</dt>
            <dd class="mt-1 font-semibold tabular-nums text-gray-900 dark:text-white">
              {{ formatTokenCount(totalTokens, { display: 'exact' }) }}
            </dd>
          </div>
        </dl>
      </section>

      <section>
        <h4 class="mb-3 font-semibold text-gray-900 dark:text-white">
          {{ t('usage.logs.billingDetails') }}
        </h4>
        <dl class="grid grid-cols-1 gap-3 rounded-md border border-gray-200 p-4 dark:border-dark-700 sm:grid-cols-3">
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.rate') }}</dt>
            <dd class="mt-1 font-medium tabular-nums text-gray-900 dark:text-white">
              {{ formatMultiplier(detail.rate_multiplier) }}x
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.actualCost') }}</dt>
            <dd class="mt-1 font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">
              {{ formatCurrency(detail.actual_cost) }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.standardCost') }}</dt>
            <dd class="mt-1 font-medium tabular-nums text-gray-900 dark:text-white">
              {{ formatCurrency(detail.total_cost) }}
            </dd>
          </div>
        </dl>
      </section>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { usageAPI } from '@/api'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { formatCurrency, formatDateTime, formatReasoningEffort, formatTokenCount } from '@/utils/format'
import { formatMultiplier } from '@/utils/formatters'
import type { UsageLog, UserRequestLog } from '@/types'

const props = defineProps<{
  show: boolean
  log: UserRequestLog | null
}>()

const emit = defineEmits<{
  (event: 'update:show', value: boolean): void
}>()

const { t } = useI18n()
const detail = ref<UsageLog | null>(null)
const loading = ref(false)
const loadError = ref(false)
let requestSequence = 0

const totalTokens = computed(() => {
  if (!detail.value) return 0
  return detail.value.input_tokens
    + detail.value.output_tokens
    + detail.value.cache_read_tokens
    + detail.value.cache_creation_tokens
})

const formatDuration = (milliseconds: number | null | undefined): string => {
  if (milliseconds == null) return '-'
  if (milliseconds < 1000) return `${milliseconds}ms`
  return `${(milliseconds / 1000).toFixed(milliseconds >= 10_000 ? 1 : 2)}s`
}

watch(
  () => [props.show, props.log?.id] as const,
  ([show, usageId]) => {
    const sequence = ++requestSequence
    if (!show || usageId == null) {
      detail.value = null
      loadError.value = false
      loading.value = false
      return
    }

    loading.value = true
    loadError.value = false
    detail.value = null
    void usageAPI.getById(usageId)
      .then((response) => {
        if (sequence === requestSequence) detail.value = response
      })
      .catch(() => {
        if (sequence === requestSequence) loadError.value = true
      })
      .finally(() => {
        if (sequence === requestSequence) loading.value = false
      })
  }
)
</script>
