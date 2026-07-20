<template>
  <article class="flex min-h-[18rem] flex-col rounded-lg border border-gray-200 bg-white p-5 shadow-sm transition hover:border-gray-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800/60 dark:hover:border-dark-600">
    <div class="flex items-start gap-3">
      <div class="flex h-10 w-10 flex-none items-center justify-center rounded-lg border border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-900/70">
        <ModelIcon :model="item.modelId" size="24px" />
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-start justify-between gap-2">
          <div class="min-w-0">
            <h2 class="truncate text-sm font-semibold text-gray-900 dark:text-white" :title="item.displayName">
              {{ item.displayName }}
            </h2>
            <p class="mt-0.5 truncate font-mono text-xs text-gray-500 dark:text-dark-400" :title="item.modelId">
              {{ item.modelId }}
            </p>
          </div>
          <span class="inline-flex flex-none items-center gap-1.5 rounded-full px-2 py-1 text-[11px] font-medium" :class="statusClass">
            <span class="h-1.5 w-1.5 rounded-full bg-current" />
            {{ text(item.status) }}
          </span>
        </div>
        <p class="mt-2 text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ item.provider }}</p>
      </div>
    </div>

    <p v-if="item.recommendedUse" class="mt-4 line-clamp-2 text-sm text-gray-600 dark:text-dark-300">
      {{ item.recommendedUse }}
    </p>

    <div class="mt-4 flex min-h-7 flex-wrap gap-1.5">
      <span v-for="capability in capabilityLabels" :key="capability" class="rounded-md bg-gray-100 px-2 py-1 text-[11px] font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-300">
        {{ capability }}
      </span>
    </div>

    <div class="mt-4 border-t border-gray-100 pt-4 dark:border-dark-700">
      <p class="text-xs text-gray-500 dark:text-dark-400">{{ text('pricing') }}</p>
      <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ pricePrimary }}</p>
      <p v-if="priceSecondary" class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">{{ priceSecondary }}</p>
    </div>

    <div class="mt-4 flex flex-wrap gap-1.5">
      <span v-for="group in visibleGroups" :key="group.id" class="rounded-md border border-gray-200 px-2 py-1 text-[11px] text-gray-600 dark:border-dark-600 dark:text-dark-300">
        {{ group.name }}<span v-if="group.isExclusive"> · {{ text('exclusive') }}</span>
      </span>
      <span v-if="hiddenGroupCount > 0" class="rounded-md bg-gray-100 px-2 py-1 text-[11px] text-gray-500 dark:bg-dark-700 dark:text-dark-400">
        +{{ hiddenGroupCount }}
      </span>
    </div>

    <div class="mt-auto flex items-center justify-between gap-2 pt-5">
      <p class="text-xs text-gray-500 dark:text-dark-400">
        {{ item.availableChannelCount }} {{ text('channels').toLowerCase() }}
      </p>
      <div class="flex items-center gap-2">
        <button class="btn btn-icon btn-secondary p-2" type="button" :title="text('copyModel')" @click="emit('copy', item.modelId)">
          <Icon name="copy" size="sm" />
        </button>
        <button class="btn btn-sm btn-secondary" type="button" @click="emit('details', item)">
          <Icon name="eye" size="sm" />
          {{ text('details') }}
        </button>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import { formatScaled } from '@/utils/pricing'
import {
  getModelCatalogGroups,
  getModelPricingSummary,
  type ModelCatalogItem,
  type ModelCatalogCapability,
} from '@/utils/modelCatalog'
import { useModelPlazaText } from './modelPlazaText'

const props = defineProps<{ item: ModelCatalogItem }>()
const emit = defineEmits<{
  (event: 'copy', modelId: string): void
  (event: 'details', item: ModelCatalogItem): void
}>()
const { text } = useModelPlazaText()

const groups = computed(() => getModelCatalogGroups(props.item))
const visibleGroups = computed(() => groups.value.slice(0, 3))
const hiddenGroupCount = computed(() => Math.max(0, groups.value.length - visibleGroups.value.length))

const statusClass = computed(() => ({
  available: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/25 dark:text-emerald-300',
  maintenance: 'bg-amber-50 text-amber-700 dark:bg-amber-900/25 dark:text-amber-300',
  unavailable: 'bg-red-50 text-red-700 dark:bg-red-900/25 dark:text-red-300',
  unknown: 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300',
})[props.item.status])

const capabilityLabels = computed(() => {
  const values: ModelCatalogCapability[] = []
  if (props.item.supportsStreaming === true) values.push('streaming')
  if (props.item.supportsVision === true) values.push('vision')
  if (props.item.supportsTools === true) values.push('tools')
  if (props.item.supportsJson === true) values.push('json')
  return values.map((value) => text(value))
})

const pricingSummary = computed(() => getModelPricingSummary(props.item))
const pricePrimary = computed(() => {
  const summary = pricingSummary.value
  if (summary.kind === 'none') return text('noPricing')
  if (summary.kind === 'multiple') return text('priceVariants', { count: summary.pricedOfferCount })
  const pricing = summary.pricing
  if (!pricing) return text('noPricing')
  if (pricing.billing_mode === 'token') {
    return `${formatScaled(pricing.input_price, 1_000_000)} / ${formatScaled(pricing.output_price, 1_000_000)}`
  }
  const value = pricing.billing_mode === 'image' ? pricing.image_output_price : pricing.per_request_price
  return formatScaled(value, 1)
})

const priceSecondary = computed(() => {
  const summary = pricingSummary.value
  if (summary.kind === 'multiple') return summary.billingModes.map(billingModeLabel).join(' · ')
  const mode = summary.pricing?.billing_mode
  if (!mode) return ''
  return `${billingModeLabel(mode)} ${mode === 'token' ? text('perMillion') : text('each')}`
})

function billingModeLabel(mode: string): string {
  if (mode === 'token') return text('tokenBilling')
  if (mode === 'per_request') return text('requestBilling')
  return text('imageBilling')
}
</script>
