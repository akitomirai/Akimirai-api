<template>
  <section class="space-y-6">
    <div class="rounded-xl border border-gray-200 bg-white px-5 py-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('channelStatus.modelStats.title') }}</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ todayRangeLabel }}</p>
        </div>

        <div class="grid grid-cols-2 gap-x-6 gap-y-3 text-sm sm:grid-cols-4 lg:min-w-[520px]">
          <div class="lg:text-right">
            <p class="text-gray-500 dark:text-dark-400">{{ t('channelStatus.modelStats.models') }}</p>
            <p class="mt-1 font-semibold text-gray-900 dark:text-white">{{ rows.length }}</p>
          </div>
          <div class="lg:text-right">
            <p class="text-gray-500 dark:text-dark-400">{{ t('channelStatus.modelStats.successRate') }}</p>
            <p class="mt-1 font-semibold text-gray-900 dark:text-white">{{ formatPercent(overallAvailability) }}</p>
          </div>
          <div class="lg:text-right">
            <p class="text-gray-500 dark:text-dark-400">{{ t('channelStatus.modelStats.token24h') }}</p>
            <p class="mt-1 font-semibold text-gray-900 dark:text-white">{{ formatTokenCount(totalTokens) }}</p>
          </div>
          <div class="lg:text-right">
            <p class="text-gray-500 dark:text-dark-400">{{ t('channelStatus.modelStats.normalModels') }}</p>
            <p class="mt-1 font-semibold text-gray-900 dark:text-white">{{ operationalCount }} / {{ rows.length }}</p>
          </div>
        </div>
      </div>
    </div>

    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-dark-400">
        <span v-for="item in legendItems" :key="item.label" class="inline-flex items-center gap-1.5">
          <span class="h-2 w-2 rounded-full" :class="item.className"></span>
          {{ item.label }}
        </span>
      </div>

      <div class="flex w-full items-center gap-3 sm:w-auto">
        <div class="relative min-w-0 flex-1 sm:w-72 sm:flex-none">
        <input
          v-model="searchTerm"
          type="search"
          class="input h-9 w-full rounded-lg pr-9 text-sm"
          :placeholder="t('channelStatus.modelStats.searchPlaceholder')"
        />
        </div>
        <button
          type="button"
          class="btn btn-secondary h-9 px-4 text-sm"
          :disabled="loading"
          :title="t('common.refresh')"
          @click="emit('refresh')"
        >
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <div v-if="loading && rows.length === 0" class="grid gap-x-10 gap-y-9 lg:grid-cols-2">
      <div v-for="i in 6" :key="i" class="animate-pulse">
        <div class="mb-4 flex items-center justify-between">
          <div class="h-5 w-32 rounded bg-gray-200 dark:bg-dark-700"></div>
          <div class="h-5 w-12 rounded bg-gray-200 dark:bg-dark-700"></div>
        </div>
        <div class="h-[34px] rounded bg-gray-100 dark:bg-dark-900/50"></div>
        <div class="mt-4 h-4 w-4/5 rounded bg-gray-100 dark:bg-dark-900/50"></div>
      </div>
    </div>

    <EmptyState
      v-else-if="filteredRows.length === 0"
      :title="t('channelStatus.empty.title')"
      :description="t('channelStatus.empty.description')"
    />

    <div v-else data-test="model-status-grid" class="grid gap-x-10 gap-y-9 lg:grid-cols-2">
      <button
        v-for="row in filteredRows"
        :key="row.model"
        type="button"
        class="group min-w-0 rounded-lg px-0.5 py-1 text-left transition-colors hover:bg-white/60 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-400 dark:hover:bg-dark-800/45"
        @click="emit('modelClick', row.primaryMonitor)"
      >
        <div class="mb-5 flex items-center justify-between gap-3">
          <div class="flex min-w-0 items-center gap-2">
            <span class="h-2 w-2 shrink-0 rounded-full" :class="modelDotClass(row.status)"></span>
            <span class="truncate text-[15px] font-semibold text-gray-950 dark:text-white">{{ row.model }}</span>
            <span
              v-for="provider in row.providers"
              :key="provider"
              class="rounded-md px-1.5 py-0.5 text-[10px] font-semibold leading-none"
              :class="providerBadgeClass(provider)"
            >
              {{ providerLabel(provider) }}
            </span>
          </div>
          <span class="shrink-0 text-xs font-medium" :class="row.status === 'operational' ? 'text-emerald-600 dark:text-emerald-300' : 'text-amber-600 dark:text-amber-300'">
            {{ row.status === 'operational' ? t('channelStatus.modelStats.normal') : t('channelStatus.modelStats.needsAttention') }}
          </span>
        </div>

        <div class="flex h-[34px] items-end gap-[4px] overflow-hidden">
          <span
            v-for="(bar, idx) in row.bars"
            :key="`${row.model}-${idx}`"
            data-test="model-status-bar"
            class="min-w-[3px] max-w-[8px] flex-1 rounded-[2px] transition-opacity group-hover:opacity-90"
            :class="bar.colorClass"
            :style="{ height: `${bar.heightPct}%` }"
            :title="bar.title"
          ></span>
        </div>

        <div class="mt-1 flex justify-between text-[11px] text-gray-400 dark:text-dark-500">
          <span>{{ t('channelStatus.modelStats.past') }}</span>
          <span>{{ t('channelStatus.modelStats.now') }}</span>
        </div>

        <div class="mt-4 flex flex-wrap items-center gap-x-6 gap-y-2 text-xs text-slate-500 dark:text-dark-400">
          <span>{{ t('channelStatus.modelStats.successRate') }} <b>{{ formatPercent(row.availability) }}</b></span>
          <span>{{ t('channelStatus.modelStats.successFailed') }} <b>{{ formatNumber(row.successCount) }}/{{ formatNumber(row.failureCount) }}</b></span>
          <span>{{ t('channelStatus.modelStats.token24h') }} <b>{{ formatTokenCount(row.totalTokens) }}</b></span>
          <span>{{ t('channelStatus.modelStats.avgLatency') }} <b>{{ row.avgLatencyMs == null ? '-' : `${row.avgLatencyMs}ms` }}</b></span>
          <span>{{ t('channelStatus.modelStats.cacheRate') }} <b>{{ formatPercent(row.cacheRate) }}</b></span>
          <span>{{ t('channelStatus.modelStats.monitors') }} <b>{{ row.monitorCount }}</b></span>
        </div>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type {
  MonitorTimelinePoint,
  MonitorStatus,
  UserMonitorDetail,
  UserMonitorView,
} from '@/api/channelMonitor'
import type { ModelStat } from '@/types'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatTokenCount } from '@/utils/format'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

type MonitorWindow = '7d' | '15d' | '30d'

interface TimelineBar {
  colorClass: string
  heightPct: number
  title: string
}

interface ModelStatusRow {
  model: string
  providers: string[]
  primaryMonitor: UserMonitorView
  status: MonitorStatus
  availability: number | null
  bars: TimelineBar[]
  successCount: number
  failureCount: number
  totalTokens: number
  cacheRate: number | null
  avgLatencyMs: number | null
  monitorCount: number
}

const props = defineProps<{
  items: UserMonitorView[]
  window: MonitorWindow
  loading: boolean
  detailCache: Record<number, UserMonitorDetail>
  usageModels: ModelStat[]
}>()

const emit = defineEmits<{
  (e: 'modelClick', item: UserMonitorView): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()
const { statusLabel, providerLabel, providerBadgeClass, formatPercent } = useChannelMonitorFormat()

const searchTerm = ref('')

const STATUS_RANK: Record<string, number> = {
  operational: 0,
  degraded: 1,
  failed: 2,
  error: 2,
}

const STATUS_COLOR: Record<string, string> = {
  operational: 'bg-emerald-500',
  degraded: 'bg-amber-500',
  failed: 'bg-red-500',
  error: 'bg-red-500',
  empty: 'bg-slate-100 dark:bg-dark-700',
}

const STATUS_HEIGHT: Record<string, number> = {
  operational: 88,
  degraded: 62,
  failed: 38,
  error: 38,
  empty: 88,
}

const legendItems = computed(() => [
  { label: t('channelStatus.modelStats.good'), className: STATUS_COLOR.operational },
  { label: t('channelStatus.modelStats.fair'), className: STATUS_COLOR.degraded },
  { label: t('channelStatus.modelStats.severe'), className: STATUS_COLOR.failed },
  { label: t('channelStatus.modelStats.noData'), className: STATUS_COLOR.empty },
])

const usageByModel = computed(() => {
  const map = new Map<string, ModelStat>()
  for (const item of props.usageModels) {
    map.set(item.model, item)
  }
  return map
})

const todayRangeLabel = computed(() => {
  const start = new Date()
  start.setHours(0, 0, 0, 0)
  const end = new Date(start)
  end.setDate(end.getDate() + 1)
  return `${formatMonthDay(start)} 00:00 - ${formatMonthDay(end)} 00:00`
})

const rows = computed<ModelStatusRow[]>(() => {
  const groups = new Map<string, UserMonitorView[]>()
  for (const item of props.items) {
    const model = item.primary_model?.trim() || item.name
    if (!groups.has(model)) groups.set(model, [])
    groups.get(model)!.push(item)
  }

  return [...groups.entries()]
    .map(([model, monitors]) => buildRow(model, monitors))
    .sort((a, b) => {
      const rankDiff = (STATUS_RANK[b.status] ?? 0) - (STATUS_RANK[a.status] ?? 0)
      if (rankDiff !== 0) return rankDiff
      return b.totalTokens - a.totalTokens
    })
})

const filteredRows = computed(() => {
  const q = searchTerm.value.trim().toLowerCase()
  if (!q) return rows.value
  return rows.value.filter(row => row.model.toLowerCase().includes(q))
})

const operationalCount = computed(() => rows.value.filter(row => row.status === 'operational').length)
const totalTokens = computed(() => rows.value.reduce((sum, row) => sum + row.totalTokens, 0))
const overallAvailability = computed(() => average(rows.value.map(row => row.availability)))

function buildRow(model: string, monitors: UserMonitorView[]): ModelStatusRow {
  const status = summarizeStatus(monitors.map(item => item.primary_status))
  const timeline = aggregateTimeline(monitors)
  const usage = usageByModel.value.get(model)
  const successCount = timeline.filter(point => point.status === 'operational').length
  const failureCount = timeline.filter(point => point.status !== 'operational').length
  const latencyAvg = average([
    ...timeline.map(point => point.latency_ms),
    ...monitors.map(item => item.primary_latency_ms),
  ])
  const cacheTotal = (usage?.input_tokens ?? 0) + (usage?.cache_read_tokens ?? 0) + (usage?.cache_creation_tokens ?? 0)

  return {
    model,
    providers: [...new Set(monitors.map(item => item.provider).filter(Boolean))],
    primaryMonitor: monitors[0],
    status,
    availability: average(monitors.map(resolveAvailability)),
    bars: timelineToBars(timeline),
    successCount,
    failureCount,
    totalTokens: usage?.total_tokens ?? 0,
    cacheRate: cacheTotal > 0 ? ((usage?.cache_read_tokens ?? 0) / cacheTotal) * 100 : null,
    avgLatencyMs: latencyAvg == null ? null : Math.round(latencyAvg),
    monitorCount: monitors.length,
  }
}

function resolveAvailability(item: UserMonitorView): number | null {
  if (props.window === '7d') return item.availability_7d ?? null
  const detail = props.detailCache[item.id]
  if (!detail) return null
  const primary = detail.models.find(m => m.model === item.primary_model)
  if (!primary) return null
  return props.window === '15d' ? primary.availability_15d ?? null : primary.availability_30d ?? null
}

function aggregateTimeline(monitors: UserMonitorView[], length = 60): MonitorTimelinePoint[] {
  const result: MonitorTimelinePoint[] = []

  for (let idx = 0; idx < length; idx += 1) {
    const points = monitors
      .map(item => item.timeline?.[idx])
      .filter((point): point is MonitorTimelinePoint => Boolean(point))

    if (points.length === 0) continue

    result.push({
      status: summarizeStatus(points.map(point => point.status)),
      latency_ms: average(points.map(point => point.latency_ms)),
      ping_latency_ms: average(points.map(point => point.ping_latency_ms)),
      checked_at: points[0]?.checked_at ?? '',
    })
  }

  return result
}

function timelineToBars(points: MonitorTimelinePoint[], length = 60): TimelineBar[] {
  const real = points.slice(0, length).reverse()
  const padCount = Math.max(0, length - real.length)
  const bars: TimelineBar[] = Array.from({ length: padCount }, () => ({
    colorClass: STATUS_COLOR.empty,
    heightPct: STATUS_HEIGHT.empty,
    title: t('channelStatus.modelStats.noData'),
  }))

  for (const point of real) {
    const status = point.status || 'empty'
    bars.push({
      colorClass: STATUS_COLOR[status] ?? STATUS_COLOR.empty,
      heightPct: STATUS_HEIGHT[status] ?? STATUS_HEIGHT.empty,
      title: `${statusLabel(point.status)}${point.latency_ms == null ? '' : ` - ${Math.round(point.latency_ms)}ms`}`,
    })
  }

  return bars
}

function summarizeStatus(statuses: MonitorStatus[]): MonitorStatus {
  if (statuses.some(status => status === 'failed' || status === 'error')) return 'failed'
  if (statuses.some(status => status === 'degraded')) return 'degraded'
  return 'operational'
}

function average(values: Array<number | null | undefined>): number | null {
  const valid = values.filter((value): value is number => typeof value === 'number' && Number.isFinite(value))
  if (valid.length === 0) return null
  return valid.reduce((sum, value) => sum + value, 0) / valid.length
}

function formatNumber(value: number): string {
  return value.toLocaleString()
}

function formatMonthDay(date: Date): string {
  return `${String(date.getMonth() + 1).padStart(2, '0')}/${String(date.getDate()).padStart(2, '0')}`
}

function modelDotClass(status: MonitorStatus): string {
  return STATUS_COLOR[status] ?? STATUS_COLOR.empty
}
</script>
