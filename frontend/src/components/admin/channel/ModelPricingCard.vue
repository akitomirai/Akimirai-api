<template>
  <article
    class="flex min-h-[236px] flex-col rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition-colors hover:border-gray-300 dark:border-dark-700 dark:bg-dark-800/70 dark:hover:border-dark-600"
    data-test="model-pricing-card"
  >
    <div class="flex min-w-0 items-start justify-between gap-3">
      <div class="flex min-w-0 items-center gap-3">
        <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-gray-50 dark:bg-dark-700">
          <ModelIcon :model="card.model" size="24px" />
        </div>
        <div class="min-w-0">
          <h2 class="truncate font-mono text-sm font-semibold text-gray-900 dark:text-white" :title="card.model">
            {{ card.model }}
          </h2>
          <div class="mt-1 flex flex-wrap items-center gap-1.5">
            <span
              class="inline-flex items-center rounded-md border px-2 py-0.5 text-[11px] font-medium"
              :class="platformBadgeClass(card.platform)"
            >
              {{ platformLabel(card.platform) }}
            </span>
            <span
              v-if="card.isPattern"
              class="inline-flex items-center rounded-md bg-amber-50 px-2 py-0.5 text-[11px] font-medium text-amber-700 dark:bg-amber-900/20 dark:text-amber-300"
            >
              {{ t('admin.modelPricing.pattern', 'Pattern') }}
            </span>
          </div>
        </div>
      </div>

      <span
        class="inline-flex flex-shrink-0 items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
        :class="card.activeSourceCount > 0
          ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300'
          : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'"
      >
        <span class="h-1.5 w-1.5 rounded-full" :class="card.activeSourceCount > 0 ? 'bg-emerald-500' : 'bg-gray-400'"></span>
        {{ card.activeSourceCount > 0 ? t('common.active', 'Active') : t('common.disabled', 'Disabled') }}
      </span>
    </div>

    <div class="mt-4 flex flex-wrap gap-1.5">
      <span
        v-for="mode in card.billingModes"
        :key="mode"
        class="inline-flex rounded-md bg-gray-100 px-2 py-1 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-300"
      >
        {{ billingModeLabel(mode) }}
      </span>
    </div>

    <div class="mt-4 min-h-[48px] rounded-lg bg-gray-50 px-3 py-2.5 dark:bg-dark-700/60">
      <p class="text-xs font-medium text-gray-500 dark:text-dark-400">
        {{ t('admin.modelPricing.priceSummary', 'Price summary') }}
      </p>
      <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">
        {{ priceSummary }}
      </p>
    </div>

    <div class="mt-auto flex items-end justify-between gap-3 pt-4">
      <div class="min-w-0 text-xs text-gray-500 dark:text-dark-400">
        <p>
          {{ t('admin.modelPricing.sourceCount', { count: card.sources.length }, `${card.sources.length} source rule(s)`) }}
        </p>
        <p v-if="card.disabledSourceCount > 0" class="mt-1 truncate">
          {{ t('admin.modelPricing.disabledSourceCount', { count: card.disabledSourceCount }, `${card.disabledSourceCount} disabled`) }}
        </p>
      </div>
      <button type="button" class="btn btn-sm btn-secondary flex-shrink-0" @click="$emit('details', card)">
        <Icon name="eye" size="sm" />
        {{ t('admin.modelPricing.details', 'Details') }}
      </button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { BillingMode } from '@/constants/channel'
import type { AdminModelPricingCard } from '@/utils/adminModelPricing'
import { formatScaled } from '@/utils/pricing'
import { platformBadgeClass, platformLabel } from '@/utils/platformColors'
import Icon from '@/components/icons/Icon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'

const props = defineProps<{
  card: AdminModelPricingCard
}>()

defineEmits<{
  details: [card: AdminModelPricingCard]
}>()

const { t } = useI18n()

function billingModeLabel(mode: BillingMode): string {
  switch (mode) {
    case 'token': return t('admin.channels.billingMode.token', 'By token')
    case 'per_request': return t('admin.channels.billingMode.perRequest', 'By request')
    case 'image': return t('admin.channels.billingMode.image', 'By image')
  }
}

function formatRange(values: Array<number | null>, scale: number): string | null {
  const numbers = [...new Set(values.filter((value): value is number => value != null))].sort((a, b) => a - b)
  if (numbers.length === 0) return null
  if (numbers.length === 1) return formatScaled(numbers[0], scale)
  return `${formatScaled(numbers[0], scale)}-${formatScaled(numbers[numbers.length - 1], scale)}`
}

const priceSummary = computed(() => {
  if (props.card.billingModes.length !== 1) {
    return t(
      'admin.modelPricing.mixedPricingSummary',
      { count: props.card.sources.length },
      `${props.card.sources.length} rules / mixed billing`,
    )
  }

  const mode = props.card.billingModes[0]
  if (mode === 'token') {
    const input = formatRange(props.card.sources.map((source) => source.pricing.input_price), 1_000_000)
    const output = formatRange(props.card.sources.map((source) => source.pricing.output_price), 1_000_000)
    const parts = [
      input ? `${t('admin.channels.form.inputPrice', 'Input')} ${input}` : null,
      output ? `${t('admin.channels.form.outputPrice', 'Output')} ${output}` : null,
    ].filter(Boolean)
    return parts.length > 0 ? `${parts.join(' / ')} / 1M` : t('admin.modelPricing.tieredOrUnset', 'Tiered or unset')
  }

  const requestPrice = formatRange(props.card.sources.map((source) => (
    mode === 'image' ? source.pricing.image_output_price : source.pricing.per_request_price
  )), 1)
  if (requestPrice) {
    return `${requestPrice} / ${mode === 'image' ? t('admin.modelPricing.imageUnit', 'image') : t('admin.modelPricing.requestUnit', 'request')}`
  }
  return t('admin.modelPricing.tieredOrUnset', 'Tiered or unset')
})
</script>
