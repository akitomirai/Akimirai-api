<template>
  <BaseDialog
    :show="show"
    :title="model?.displayName || text('details')"
    width="full"
    :close-on-click-outside="true"
    @close="emit('close')"
  >
    <div v-if="model" class="space-y-5" data-testid="model-plaza-detail">
      <div class="flex flex-col gap-3 border-b border-gray-100 pb-4 dark:border-dark-700 sm:flex-row sm:items-start sm:justify-between">
        <div class="flex min-w-0 items-start gap-3">
          <div class="flex h-11 w-11 flex-none items-center justify-center rounded-lg border border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-900/70">
            <ModelIcon :model="model.modelId" size="26px" />
          </div>
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <code class="break-all font-mono text-sm font-semibold text-gray-900 dark:text-white">{{ model.modelId }}</code>
              <span class="rounded-full px-2 py-0.5 text-[11px] font-medium" :class="statusClass(model.status)">{{ text(model.status) }}</span>
            </div>
            <p class="mt-1 text-xs uppercase text-gray-500 dark:text-dark-400">{{ model.provider }}</p>
          </div>
        </div>
        <button class="btn btn-sm btn-secondary flex-none" type="button" @click="copy(model.modelId)">
          <Icon name="copy" size="sm" />
          {{ text('copyModel') }}
        </button>
      </div>

      <div class="grid grid-cols-3 rounded-lg bg-gray-100 p-1 dark:bg-dark-900/70" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab"
          class="min-h-9 rounded-md px-3 py-2 text-xs font-medium transition sm:text-sm"
          :class="activeTab === tab ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-500 hover:text-gray-800 dark:text-dark-400 dark:hover:text-dark-200'"
          type="button"
          role="tab"
          :aria-selected="activeTab === tab"
          @click="activeTab = tab"
        >
          {{ text(tab) }}
        </button>
      </div>

      <section v-if="activeTab === 'overview'" class="space-y-5" data-testid="overview-tab">
        <dl class="grid gap-px overflow-hidden rounded-lg border border-gray-200 bg-gray-200 text-sm dark:border-dark-700 dark:bg-dark-700 sm:grid-cols-2 lg:grid-cols-4">
          <div class="bg-white p-4 dark:bg-dark-800">
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ text('provider') }}</dt>
            <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ model.provider }}</dd>
          </div>
          <div class="bg-white p-4 dark:bg-dark-800">
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ text('contextWindow') }}</dt>
            <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ model.contextWindow ? formatTokenCount(model.contextWindow) : text('notReported') }}</dd>
          </div>
          <div class="bg-white p-4 dark:bg-dark-800">
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ text('visibleChannels') }}</dt>
            <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ model.availableChannelCount }}</dd>
          </div>
          <div class="bg-white p-4 dark:bg-dark-800">
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ text('lastUpdated') }}</dt>
            <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ model.updatedAt ? formatDateTime(model.updatedAt) : text('notReported') }}</dd>
          </div>
        </dl>

        <div>
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ text('capabilities') }}</h3>
          <div class="mt-2 flex flex-wrap gap-2">
            <span v-for="capability in capabilities" :key="capability" class="rounded-md border border-gray-200 px-2.5 py-1 text-xs text-gray-700 dark:border-dark-600 dark:text-dark-200">
              {{ text(capability) }}
            </span>
            <span v-if="capabilities.length === 0" class="text-sm text-gray-500 dark:text-dark-400">{{ text('notReported') }}</span>
          </div>
        </div>

        <div>
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ text('sourceOffers') }}</h3>
          <div class="mt-3 space-y-3">
            <section
              v-for="(offer, index) in offers"
              :key="`${offer.channel}-${offer.platform}-${index}`"
              class="rounded-lg border border-gray-200 p-4 dark:border-dark-700"
              data-testid="catalog-offer"
            >
              <div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ offer.channel || text('legacyOffer') }}</p>
                  <p class="mt-0.5 text-xs uppercase text-gray-500 dark:text-dark-400">{{ offer.platform || model.provider }}</p>
                </div>
                <span v-if="offer.pricing" class="rounded-md bg-gray-100 px-2 py-1 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-300">
                  {{ billingModeLabel(offer.pricing.billing_mode) }}
                </span>
              </div>

              <div class="mt-4 grid gap-4 lg:grid-cols-[minmax(0,1.2fr)_minmax(0,1fr)]">
                <div>
                  <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ text('groups') }}</p>
                  <div class="mt-2 flex flex-wrap gap-2">
                    <span v-for="group in offer.groups" :key="group.id" class="rounded-md border border-gray-200 px-2 py-1 text-xs text-gray-700 dark:border-dark-600 dark:text-dark-200">
                      {{ group.name }} · {{ text('normalRate') }} {{ formatMultiplier(group.effectiveRateMultiplier) }}x
                      <template v-if="group.peakRateEnabled">
                        · {{ text('peakRate', { start: group.peakStart || '-', end: group.peakEnd || '-' }) }} {{ formatMultiplier(group.peakRateMultiplier) }}x
                      </template>
                    </span>
                    <span v-if="offer.groups.length === 0" class="text-sm text-gray-500 dark:text-dark-400">{{ text('notReported') }}</span>
                  </div>
                </div>

                <div>
                  <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ text('pricing') }}</p>
                  <dl v-if="offer.pricing" class="mt-2 space-y-2 text-xs">
                    <div v-for="row in priceRows(offer.pricing)" :key="row.label" class="flex items-center justify-between gap-4">
                      <dt class="text-gray-500 dark:text-dark-400">{{ row.label }}</dt>
                      <dd class="font-mono font-medium text-gray-900 dark:text-white">{{ row.value }}</dd>
                    </div>
                  </dl>
                  <p v-else class="mt-2 text-sm text-gray-500 dark:text-dark-400">{{ text('noPricing') }}</p>
                </div>
              </div>

              <div v-if="offer.pricing?.intervals.length" class="mt-4 overflow-x-auto border-t border-gray-100 pt-4 dark:border-dark-700">
                <p class="mb-2 text-xs font-medium text-gray-500 dark:text-dark-400">{{ text('intervals') }}</p>
                <table class="w-full min-w-[34rem] text-left text-xs">
                  <thead class="text-gray-500 dark:text-dark-400">
                    <tr>
                      <th class="pb-2 font-medium">{{ text('tier') }} / {{ text('range') }}</th>
                      <th class="pb-2 font-medium">{{ text('input') }}</th>
                      <th class="pb-2 font-medium">{{ text('output') }}</th>
                      <th class="pb-2 font-medium">{{ text('perRequest') }}</th>
                    </tr>
                  </thead>
                  <tbody class="text-gray-800 dark:text-dark-100">
                    <tr v-for="(interval, intervalIndex) in offer.pricing.intervals" :key="intervalIndex" class="border-t border-gray-100 dark:border-dark-700">
                      <td class="py-2 pr-3">{{ interval.tier_label || intervalRange(interval.min_tokens, interval.max_tokens) }}</td>
                      <td class="py-2 pr-3 font-mono">{{ formatScaled(interval.input_price, 1_000_000) }}</td>
                      <td class="py-2 pr-3 font-mono">{{ formatScaled(interval.output_price, 1_000_000) }}</td>
                      <td class="py-2 font-mono">{{ formatScaled(interval.per_request_price, 1) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>
          </div>
        </div>
      </section>

      <section v-else-if="activeTab === 'performance'" class="space-y-4" data-testid="performance-tab">
        <dl class="grid gap-px overflow-hidden rounded-lg border border-gray-200 bg-gray-200 text-sm dark:border-dark-700 dark:bg-dark-700 sm:grid-cols-3">
          <div class="bg-white p-4 dark:bg-dark-800">
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ text('currentStatus') }}</dt>
            <dd class="mt-1 font-semibold text-gray-900 dark:text-white">{{ text(model.status) }}</dd>
          </div>
          <div class="bg-white p-4 dark:bg-dark-800">
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ text('visibleChannels') }}</dt>
            <dd class="mt-1 font-semibold text-gray-900 dark:text-white">{{ model.availableChannelCount }}</dd>
          </div>
          <div class="bg-white p-4 dark:bg-dark-800">
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ text('lastUpdated') }}</dt>
            <dd class="mt-1 font-semibold text-gray-900 dark:text-white">{{ model.updatedAt ? formatDateTime(model.updatedAt) : text('notReported') }}</dd>
          </div>
        </dl>
        <div class="rounded-lg border border-gray-200 bg-gray-50 p-5 dark:border-dark-700 dark:bg-dark-900/40">
          <p class="text-sm font-medium text-gray-800 dark:text-dark-100">{{ text('performanceUnavailable') }}</p>
          <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ text('performanceHint') }}</p>
        </div>
      </section>

      <section v-else class="space-y-4" data-testid="api-tab">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ text('apiExample') }}</h3>
          <div class="flex flex-wrap rounded-lg bg-gray-100 p-1 dark:bg-dark-900/70">
            <button
              v-for="example in examples"
              :key="example.id"
              class="rounded-md px-3 py-1.5 text-xs font-medium transition"
              :class="activeExample === example.id ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-500 dark:text-dark-400'"
              type="button"
              @click="activeExample = example.id"
            >
              {{ example.label }}
            </button>
          </div>
        </div>
        <div class="relative overflow-hidden rounded-lg border border-gray-200 bg-gray-950 dark:border-dark-700">
          <button class="absolute right-3 top-3 rounded-md border border-white/15 bg-white/10 p-2 text-gray-200 hover:bg-white/20" type="button" :title="text('copyCode')" @click="copy(activeExampleContent)">
            <Icon name="copy" size="sm" />
          </button>
          <pre class="max-h-[28rem] overflow-auto p-5 pr-14 text-xs leading-6 text-gray-100"><code>{{ activeExampleContent }}</code></pre>
        </div>
      </section>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { buildQuickStartExamples } from '@/utils/quickStart'
import { formatDateTime, formatTokenCount } from '@/utils/format'
import { formatScaled } from '@/utils/pricing'
import type { UserSupportedModelPricing } from '@/api/channels'
import type { ModelAvailabilityStatus, ModelCatalogCapability, ModelCatalogItem, ModelCatalogOffer } from '@/utils/modelCatalog'
import { useModelPlazaText } from './modelPlazaText'

type DetailTab = 'overview' | 'performance' | 'api'
type ExampleTab = 'curl' | 'openaiSdk' | 'codex'

const props = defineProps<{
  show: boolean
  model: ModelCatalogItem | null
  baseUrl: string
}>()
const emit = defineEmits<{ (event: 'close'): void }>()

const { text } = useModelPlazaText()
const { copyToClipboard } = useClipboard()
const activeTab = ref<DetailTab>('overview')
const activeExample = ref<ExampleTab>('curl')
const tabs: DetailTab[] = ['overview', 'performance', 'api']

const capabilities = computed<ModelCatalogCapability[]>(() => {
  if (!props.model) return []
  const values: ModelCatalogCapability[] = []
  if (props.model.supportsStreaming === true) values.push('streaming')
  if (props.model.supportsVision === true) values.push('vision')
  if (props.model.supportsTools === true) values.push('tools')
  if (props.model.supportsJson === true) values.push('json')
  return values
})

const offers = computed<ModelCatalogOffer[]>(() => {
  if (!props.model) return []
  if (props.model.offers.length > 0) return props.model.offers
  return [{
    channel: props.model.channelNames.join(', '),
    platform: props.model.provider,
    groups: props.model.groups,
    pricing: props.model.pricing,
  }]
})

const quickStartExamples = computed(() => buildQuickStartExamples({
  baseUrl: props.baseUrl,
  model: props.model?.modelId,
}))

const examples = computed<Array<{ id: ExampleTab; label: string; content: string }>>(() => [
  { id: 'curl', label: text('curl'), content: quickStartExamples.value.curl },
  { id: 'openaiSdk', label: text('openaiSdk'), content: quickStartExamples.value.openaiSdk },
  { id: 'codex', label: text('codex'), content: quickStartExamples.value.codex },
])

const activeExampleContent = computed(() => examples.value.find((entry) => entry.id === activeExample.value)?.content || '')

watch(() => [props.show, props.model?.id] as const, () => {
  activeTab.value = 'overview'
  activeExample.value = 'curl'
})

function copy(value: string): void {
  void copyToClipboard(value, text('copied'))
}

function statusClass(status: ModelAvailabilityStatus): string {
  return {
    available: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/25 dark:text-emerald-300',
    maintenance: 'bg-amber-50 text-amber-700 dark:bg-amber-900/25 dark:text-amber-300',
    unavailable: 'bg-red-50 text-red-700 dark:bg-red-900/25 dark:text-red-300',
    unknown: 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300',
  }[status]
}

function billingModeLabel(mode: string): string {
  if (mode === 'token') return text('tokenBilling')
  if (mode === 'per_request') return text('requestBilling')
  return text('imageBilling')
}

function formatMultiplier(value: number): string {
  return Number.isInteger(value) ? String(value) : value.toFixed(2).replace(/0+$/, '').replace(/\.$/, '')
}

function priceRows(pricing: UserSupportedModelPricing): Array<{ label: string; value: string }> {
  if (pricing.billing_mode === 'token') {
    return [
      { label: text('input'), value: `${formatScaled(pricing.input_price, 1_000_000)} ${text('perMillion')}` },
      { label: text('output'), value: `${formatScaled(pricing.output_price, 1_000_000)} ${text('perMillion')}` },
      { label: text('cacheWrite'), value: `${formatScaled(pricing.cache_write_price, 1_000_000)} ${text('perMillion')}` },
      { label: text('cacheRead'), value: `${formatScaled(pricing.cache_read_price, 1_000_000)} ${text('perMillion')}` },
      { label: text('imageInput'), value: `${formatScaled(pricing.image_input_price, 1_000_000)} ${text('perMillion')}` },
      { label: text('imageOutput'), value: `${formatScaled(pricing.image_output_price, 1_000_000)} ${text('perMillion')}` },
    ].filter((row) => !row.value.startsWith('- '))
  }
  const amount = pricing.billing_mode === 'image' ? pricing.image_output_price : pricing.per_request_price
  return [{ label: pricing.billing_mode === 'image' ? text('imageOutput') : text('perRequest'), value: `${formatScaled(amount, 1)} ${text('each')}` }]
}

function intervalRange(min: number, max: number | null): string {
  return `${formatTokenCount(min)} - ${max == null ? '∞' : formatTokenCount(max)}`
}
</script>
