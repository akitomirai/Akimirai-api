<template>
  <Teleport to="body">
    <Transition name="usage-diagnostics-fade">
      <div v-if="show" class="fixed inset-0 z-[100]">
        <button
          type="button"
          class="absolute inset-0 bg-black/35 backdrop-blur-[1px]"
          :aria-label="t('common.close')"
          @click="close"
        ></button>

        <aside
          class="absolute inset-y-0 right-0 flex w-full max-w-2xl flex-col border-l border-gray-200 bg-white shadow-2xl dark:border-dark-700 dark:bg-dark-900"
          role="dialog"
          aria-modal="true"
          :aria-label="t('admin.usage.diagnostics.title')"
        >
          <header class="flex min-h-16 items-center justify-between gap-4 border-b border-gray-200 px-5 py-3 dark:border-dark-700">
            <div class="min-w-0">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('admin.usage.diagnostics.title') }}
              </h2>
              <p class="truncate font-mono text-xs text-gray-500 dark:text-gray-400">
                {{ diagnostics?.request_id || `#${usageId ?? '-'}` }}
              </p>
            </div>
            <button
              type="button"
              class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary-500 dark:hover:bg-dark-800 dark:hover:text-white"
              :title="t('common.close')"
              @click="close"
            >
              <Icon name="x" size="sm" />
            </button>
          </header>

          <div v-if="loading" class="flex flex-1 items-center justify-center">
            <div class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
              <span class="h-4 w-4 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></span>
              {{ t('common.loading') }}
            </div>
          </div>

          <div v-else-if="loadError" class="flex flex-1 flex-col items-center justify-center gap-3 px-6 text-center">
            <Icon name="exclamationCircle" size="lg" class="text-rose-500" />
            <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('admin.usage.diagnostics.loadFailed') }}</p>
            <button type="button" class="btn btn-secondary" @click="loadDiagnostics">
              <Icon name="refresh" size="sm" />
              {{ t('admin.usage.diagnostics.retry') }}
            </button>
          </div>

          <div v-else-if="diagnostics" class="flex-1 overflow-y-auto">
            <section class="grid grid-cols-2 gap-x-5 gap-y-4 border-b border-gray-200 px-5 py-5 sm:grid-cols-4 dark:border-dark-700">
              <div>
                <div class="diagnostic-label">{{ t('admin.usage.diagnostics.requestStartedAt') }}</div>
                <div class="diagnostic-value">{{ formatDate(diagnostics.request_started_at) }}</div>
              </div>
              <div>
                <div class="diagnostic-label">{{ t('admin.usage.diagnostics.recordedAt') }}</div>
                <div class="diagnostic-value">{{ formatDate(diagnostics.created_at) }}</div>
              </div>
              <div>
                <div class="diagnostic-label">{{ t('admin.usage.diagnostics.totalDuration') }}</div>
                <div class="diagnostic-value">{{ formatDuration(diagnostics.request_total_ms) }}</div>
              </div>
              <div>
                <div class="diagnostic-label">{{ t('admin.usage.diagnostics.upstreamStatus') }}</div>
                <div class="diagnostic-value">{{ diagnostics.final_upstream_status ?? '-' }}</div>
              </div>
            </section>

            <section v-if="hasPromptCacheDiagnostics" data-testid="prompt-cache-diagnostics" class="border-b border-gray-200 px-5 py-5 dark:border-dark-700">
              <h3 class="diagnostic-heading mb-4">{{ t('admin.usage.diagnostics.promptCache') }}</h3>
              <dl class="grid grid-cols-[max-content_minmax(0,1fr)] gap-x-4 gap-y-2 text-sm">
                <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.diagnostics.promptCacheKeySource') }}</dt>
                <dd class="text-right font-medium text-gray-900 dark:text-white">{{ promptCacheSourceLabel }}</dd>
                <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.diagnostics.promptCacheKeyHash') }}</dt>
                <dd class="truncate text-right font-mono text-xs text-gray-700 dark:text-gray-300" :title="diagnostics.prompt_cache_key_hash || ''">
                  {{ shortHash(diagnostics.prompt_cache_key_hash) }}
                </dd>
                <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.diagnostics.promptCachePrefixHash') }}</dt>
                <dd class="truncate text-right font-mono text-xs text-gray-700 dark:text-gray-300" :title="diagnostics.prompt_cache_prefix_hash || ''">
                  {{ shortHash(diagnostics.prompt_cache_prefix_hash) }}
                </dd>
                <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.diagnostics.promptCacheToolsHash') }}</dt>
                <dd class="truncate text-right font-mono text-xs text-gray-700 dark:text-gray-300" :title="diagnostics.prompt_cache_tools_hash || ''">
                  {{ shortHash(diagnostics.prompt_cache_tools_hash) }}
                </dd>
                <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.diagnostics.promptCacheSystemHash') }}</dt>
                <dd class="truncate text-right font-mono text-xs text-gray-700 dark:text-gray-300" :title="diagnostics.prompt_cache_system_hash || ''">
                  {{ shortHash(diagnostics.prompt_cache_system_hash) }}
                </dd>
              </dl>
            </section>

            <section class="border-b border-gray-200 px-5 py-5 dark:border-dark-700">
              <div class="mb-4 flex items-center justify-between gap-3">
                <h3 class="diagnostic-heading">{{ t('admin.usage.diagnostics.route') }}</h3>
                <span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="routeBadgeClass">
                  {{ routeLabel }}
                </span>
              </div>
              <dl class="grid grid-cols-[max-content_minmax(0,1fr)] gap-x-4 gap-y-2 text-sm">
                <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.diagnostics.account') }}</dt>
                <dd class="truncate text-right font-medium text-gray-900 dark:text-white">
                  {{ diagnostics.account?.name || '-' }}<span v-if="diagnostics.account_id"> #{{ diagnostics.account_id }}</span>
                </dd>
                <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.diagnostics.proxy') }}</dt>
                <dd class="truncate text-right font-medium text-gray-900 dark:text-white">
                  {{ proxyLabel }}
                </dd>
                <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.diagnostics.fingerprint') }}</dt>
                <dd class="truncate text-right font-mono text-xs text-gray-700 dark:text-gray-300" :title="diagnostics.route_fingerprint || ''">
                  {{ shortFingerprint(diagnostics.route_fingerprint) }}
                </dd>
                <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.diagnostics.retrySwitch') }}</dt>
                <dd class="text-right font-medium text-gray-900 dark:text-white">
                  {{ diagnostics.retry_count ?? 0 }} / {{ diagnostics.account_switch_count ?? 0 }}
                </dd>
              </dl>
            </section>

            <section class="border-b border-gray-200 px-5 py-5 dark:border-dark-700">
              <h3 class="diagnostic-heading mb-4">{{ t('admin.usage.diagnostics.timeline') }}</h3>
              <p class="mb-3 text-[11px] leading-4 text-gray-400 dark:text-gray-500">
                {{ t('admin.usage.diagnostics.timelineDetail') }}
              </p>
              <div class="relative space-y-0 pl-5">
                <div class="absolute bottom-2 left-[5px] top-2 w-px bg-gray-200 dark:bg-dark-700"></div>
                <div
                  v-for="step in timingSteps"
                  :key="step.key"
                  :data-testid="`timing-step-${step.key}`"
                  class="relative flex min-h-10 items-start justify-between gap-4 py-2.5"
                >
                  <span class="absolute -left-5 top-3.5 h-2.5 w-2.5 rounded-full border-2 border-white bg-primary-500 ring-1 ring-primary-200 dark:border-dark-900 dark:ring-primary-900"></span>
                  <span class="min-w-0">
                    <span class="block text-sm text-gray-700 dark:text-gray-200">{{ step.label }}</span>
                    <span class="mt-0.5 block text-[11px] leading-4 text-gray-400 dark:text-gray-500">{{ step.detail }}</span>
                  </span>
                  <span class="whitespace-nowrap font-mono text-sm font-medium text-gray-900 dark:text-white">
                    {{ formatDuration(step.value) }}
                  </span>
                </div>
              </div>
              <div v-if="diagnostics.request_body_bytes != null" class="mt-3 text-xs text-gray-500 dark:text-gray-400">
                {{ t('admin.usage.diagnostics.bodyBytes') }}: {{ formatBytes(diagnostics.request_body_bytes) }}
              </div>
            </section>

            <section v-if="attempts.length" class="border-b border-gray-200 px-5 py-5 dark:border-dark-700">
              <h3 class="diagnostic-heading mb-4">{{ t('admin.usage.diagnostics.attempts') }}</h3>
              <div class="space-y-3">
                <article v-for="attempt in attempts" :key="attempt.sequence" class="rounded-md border border-gray-200 p-3 dark:border-dark-700">
                  <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                      <div class="text-sm font-medium text-gray-900 dark:text-white">
                        #{{ attempt.sequence }} {{ attempt.account_name || t('admin.usage.diagnostics.accountFallback', { id: attempt.account_id }) }}
                      </div>
                      <div class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">
                        {{ attemptRouteLabel(attempt) }}
                      </div>
                    </div>
                    <span class="rounded px-2 py-0.5 text-xs font-medium" :class="outcomeClass(attempt.outcome)">
                      {{ attempt.outcome || '-' }}
                    </span>
                  </div>
                  <div class="mt-3 grid grid-cols-2 gap-2 text-xs sm:grid-cols-4">
                    <div><span class="text-gray-400">{{ t('admin.usage.diagnostics.startedOffset') }}</span><div class="font-mono text-gray-800 dark:text-gray-200">{{ formatDuration(attempt.started_ms) }}</div></div>
                    <div><span class="text-gray-400">{{ t('admin.usage.diagnostics.writtenOffset') }}</span><div class="font-mono text-gray-800 dark:text-gray-200">{{ formatDuration(attempt.request_written_ms) }}</div></div>
                    <div><span class="text-gray-400">{{ t('admin.usage.diagnostics.firstByteOffset') }}</span><div class="font-mono text-gray-800 dark:text-gray-200">{{ formatDuration(attempt.first_byte_ms) }}</div></div>
                    <div><span class="text-gray-400">{{ t('admin.usage.diagnostics.status') }}</span><div class="font-mono text-gray-800 dark:text-gray-200">{{ attempt.upstream_status ?? '-' }}</div></div>
                  </div>
                  <p v-if="attempt.reason" class="mt-3 break-words text-xs text-rose-600 dark:text-rose-400">{{ attempt.reason }}</p>
                </article>
              </div>
            </section>
          </div>

          <footer v-if="diagnostics" class="flex items-center justify-between gap-3 border-t border-gray-200 px-5 py-3 dark:border-dark-700">
            <span class="truncate font-mono text-xs text-gray-500 dark:text-gray-400">#{{ diagnostics.id }}</span>
            <button type="button" class="btn btn-secondary" :disabled="!diagnostics.request_id" @click="openErrors">
              <Icon name="externalLink" size="sm" />
              {{ t('admin.usage.diagnostics.viewErrors') }}
            </button>
          </footer>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminUsageAPI } from '@/api/admin/usage'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import type { AdminUsageAttemptEvent, AdminUsageDiagnostics } from '@/types'

const props = defineProps<{
  show: boolean
  usageId: number | null
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  openErrors: [requestId: string]
}>()

const { t } = useI18n()
const diagnostics = ref<AdminUsageDiagnostics | null>(null)
const loading = ref(false)
const loadError = ref(false)
let requestSequence = 0
let previousBodyOverflow = ''

const close = () => emit('update:show', false)

const loadDiagnostics = async () => {
  const id = props.usageId
  if (!props.show || id == null) return
  const sequence = ++requestSequence
  loading.value = true
  loadError.value = false
  try {
    const result = await adminUsageAPI.getDiagnostics(id)
    if (sequence !== requestSequence) return
    diagnostics.value = result
  } catch {
    if (sequence !== requestSequence) return
    diagnostics.value = null
    loadError.value = true
  } finally {
    if (sequence === requestSequence) loading.value = false
  }
}

const timingSteps = computed(() => {
  const row = diagnostics.value
  if (!row) return []

  return [
    {
      key: 'body',
      label: t('admin.usage.diagnostics.bodyRead'),
      detail: t('admin.usage.diagnostics.bodyReadDetail'),
      value: row.request_body_read_ms,
    },
    {
      key: 'routing',
      label: t('admin.usage.diagnostics.routing'),
      detail: t('admin.usage.diagnostics.routingDetail'),
      value: row.routing_latency_ms,
    },
    {
      key: 'written',
      label: t('admin.usage.diagnostics.requestWritten'),
      detail: t('admin.usage.diagnostics.requestWrittenDetail'),
      value: row.upstream_request_written_ms,
    },
    {
      key: 'first-byte',
      label: t('admin.usage.diagnostics.firstByte'),
      detail: t('admin.usage.diagnostics.firstByteDetail'),
      value: row.upstream_first_byte_ms,
    },
    {
      key: 'first-token',
      label: t('admin.usage.diagnostics.firstToken'),
      detail: t('admin.usage.diagnostics.firstTokenDetail'),
      value: row.request_first_token_ms,
    },
    {
      key: 'total',
      label: t('admin.usage.diagnostics.completed'),
      detail: t('admin.usage.diagnostics.completedDetail'),
      value: row.request_total_ms,
    },
  ]
})

const attempts = computed(() => diagnostics.value?.attempt_timeline ?? [])
const hasPromptCacheDiagnostics = computed(() => {
  const row = diagnostics.value
  return Boolean(row && (
    row.prompt_cache_key_source
    || row.prompt_cache_key_hash
    || row.prompt_cache_prefix_hash
    || row.prompt_cache_tools_hash
    || row.prompt_cache_system_hash
  ))
})
const promptCacheSourceLabel = computed(() => {
  const source = diagnostics.value?.prompt_cache_key_source || 'none'
  return t(`admin.usage.diagnostics.promptCacheSource.${source}`)
})
const routeLabel = computed(() => diagnostics.value?.route_kind === 'proxy'
  ? t('admin.usage.diagnostics.proxyRoute')
  : diagnostics.value?.route_kind === 'direct'
    ? t('admin.usage.diagnostics.directRoute')
    : t('admin.usage.diagnostics.unavailable'))
const routeBadgeClass = computed(() => diagnostics.value?.route_kind === 'proxy'
  ? 'bg-sky-100 text-sky-700 dark:bg-sky-500/20 dark:text-sky-300'
  : diagnostics.value?.route_kind === 'direct'
    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300'
    : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300')
const proxyLabel = computed(() => {
  const row = diagnostics.value
  if (!row || row.route_kind !== 'proxy') return routeLabel.value
  const name = row.proxy_name_snapshot || t('admin.usage.diagnostics.proxyFallback')
  const id = row.proxy_id_snapshot ? ` #${row.proxy_id_snapshot}` : ''
  const protocol = row.proxy_protocol_snapshot ? ` (${row.proxy_protocol_snapshot})` : ''
  return `${name}${id}${protocol}`
})

const formatDate = (value: string | null | undefined) => value ? formatDateTime(value) : t('admin.usage.diagnostics.unavailable')
const formatDuration = (value: number | null | undefined) => {
  if (value == null) return t('admin.usage.diagnostics.unavailable')
  if (value < 1000) return `${value}ms`
  if (value < 60_000) return `${(value / 1000).toFixed(2)}s`
  const seconds = Math.round(value / 1000)
  return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
}
const formatBytes = (value: number) => value < 1024
  ? `${value} B`
  : value < 1024 * 1024
    ? `${(value / 1024).toFixed(1)} KiB`
    : `${(value / (1024 * 1024)).toFixed(1)} MiB`
const shortFingerprint = (value: string | null | undefined) => value ? value.slice(0, 8) : t('admin.usage.diagnostics.unavailable')
const shortHash = (value: string | null | undefined) => value ? value.slice(0, 12) : t('admin.usage.diagnostics.unavailable')
const attemptRouteLabel = (attempt: AdminUsageAttemptEvent) => {
  if (attempt.route.kind !== 'proxy') return t('admin.usage.diagnostics.directRoute')
  const name = attempt.route.proxy_name || t('admin.usage.diagnostics.proxyFallback')
  return attempt.route.proxy_id ? `${name} #${attempt.route.proxy_id}` : name
}
const outcomeClass = (outcome: string) => outcome === 'success'
  ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300'
  : outcome === 'unknown'
    ? 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
    : 'bg-rose-100 text-rose-700 dark:bg-rose-500/20 dark:text-rose-300'

const openErrors = () => {
  if (!diagnostics.value?.request_id) return
  emit('openErrors', diagnostics.value.request_id)
}

const onKeydown = (event: KeyboardEvent) => {
  if (props.show && event.key === 'Escape') close()
}

watch(() => [props.show, props.usageId] as const, ([show]) => {
  if (show) {
    void loadDiagnostics()
    previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
  } else {
    requestSequence++
    document.body.style.overflow = previousBodyOverflow
  }
}, { immediate: true })

onMounted(() => document.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = previousBodyOverflow
})
</script>

<style scoped>
.diagnostic-label {
  font-size: 0.6875rem;
  color: rgb(107 114 128);
}

.diagnostic-value {
  margin-top: 0.25rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: rgb(17 24 39);
}

.diagnostic-heading {
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(17 24 39);
}

.dark .diagnostic-label {
  color: rgb(156 163 175);
}

.dark .diagnostic-value,
.dark .diagnostic-heading {
  color: rgb(243 244 246);
}

.usage-diagnostics-fade-enter-active,
.usage-diagnostics-fade-leave-active {
  transition: opacity 160ms ease;
}

.usage-diagnostics-fade-enter-from,
.usage-diagnostics-fade-leave-to {
  opacity: 0;
}

.usage-diagnostics-fade-enter-active aside,
.usage-diagnostics-fade-leave-active aside {
  transition: transform 180ms ease;
}

.usage-diagnostics-fade-enter-from aside,
.usage-diagnostics-fade-leave-to aside {
  transform: translateX(100%);
}
</style>
