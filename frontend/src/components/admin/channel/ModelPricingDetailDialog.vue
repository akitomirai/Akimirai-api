<template>
  <BaseDialog
    :show="show"
    :title="model?.model || t('admin.modelPricing.detailsTitle', 'Model pricing details')"
    width="extra-wide"
    @close="$emit('close')"
  >
    <div v-if="model" class="max-h-[72vh] space-y-5 overflow-y-auto pr-1">
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <div class="rounded-lg bg-gray-50 px-3 py-2.5 dark:bg-dark-700/60">
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.modelPricing.provider', 'Provider') }}</p>
          <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ platformLabel(model.platform) }}</p>
        </div>
        <div class="rounded-lg bg-gray-50 px-3 py-2.5 dark:bg-dark-700/60">
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.modelPricing.sourceRules', 'Source rules') }}</p>
          <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ model.sources.length }}</p>
        </div>
        <div class="rounded-lg bg-gray-50 px-3 py-2.5 dark:bg-dark-700/60">
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('common.active', 'Active') }}</p>
          <p class="mt-1 text-sm font-semibold text-emerald-600 dark:text-emerald-400">{{ model.activeSourceCount }}</p>
        </div>
        <div class="rounded-lg bg-gray-50 px-3 py-2.5 dark:bg-dark-700/60">
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('common.disabled', 'Disabled') }}</p>
          <p class="mt-1 text-sm font-semibold text-gray-700 dark:text-dark-200">{{ model.disabledSourceCount }}</p>
        </div>
      </div>

      <div class="space-y-3">
        <section
          v-for="source in model.sources"
          :key="source.key"
          class="rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800"
          data-test="model-pricing-source"
        >
          <div class="flex flex-col gap-3 border-b border-gray-100 px-4 py-3 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ source.channelName }}</h3>
                <span
                  class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium"
                  :class="source.channelStatus === 'active'
                    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300'
                    : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'"
                >
                  {{ source.channelStatus === 'active' ? t('common.active', 'Active') : t('common.disabled', 'Disabled') }}
                </span>
              </div>
              <p v-if="source.channelDescription" class="mt-1 truncate text-xs text-gray-500 dark:text-dark-400">
                {{ source.channelDescription }}
              </p>
            </div>
            <button type="button" class="btn btn-sm btn-secondary flex-shrink-0" @click="$emit('edit-source', source.channelId)">
              <Icon name="edit" size="sm" />
              {{ t('admin.modelPricing.editSource', 'Edit source') }}
            </button>
          </div>

          <div class="space-y-4 px-4 py-4">
            <dl class="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2 lg:grid-cols-4">
              <div>
                <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.modelPricing.platform', 'Platform') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ platformLabel(source.pricing.platform) }}</dd>
              </div>
              <div>
                <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.modelPricing.billingMode', 'Billing mode') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ billingModeLabel(source.pricing.billing_mode) }}</dd>
              </div>
              <div>
                <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.modelPricing.modelSource', 'Pricing lookup') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ billingSourceLabel(source.billingModelSource) }}</dd>
              </div>
              <div>
                <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.modelPricing.restrictModels', 'Restricted models') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                  {{ source.restrictModels ? t('common.yes', 'Yes') : t('common.no', 'No') }}
                </dd>
              </div>
            </dl>

            <div>
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.modelPricing.sharedModels', 'Models sharing this rule') }}</p>
              <div class="mt-2 flex flex-wrap gap-1.5">
                <code
                  v-for="sharedModel in source.sharedModels"
                  :key="sharedModel"
                  class="rounded-md bg-gray-100 px-2 py-1 text-xs text-gray-700 dark:bg-dark-700 dark:text-dark-200"
                >
                  {{ sharedModel }}
                </code>
              </div>
            </div>

            <div>
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.modelPricing.groups', 'Associated groups') }}</p>
              <div class="mt-2 flex flex-wrap gap-1.5">
                <span
                  v-for="groupId in associatedGroupIds(source)"
                  :key="groupId"
                  class="rounded-md bg-primary-50 px-2 py-1 text-xs font-medium text-primary-700 dark:bg-primary-900/20 dark:text-primary-300"
                >
                  {{ groups[groupId]?.name || `#${groupId}` }}
                </span>
                <span v-if="associatedGroupIds(source).length === 0" class="text-xs text-gray-400">-</span>
              </div>
            </div>

            <div class="overflow-x-auto rounded-lg border border-gray-100 dark:border-dark-700">
              <table class="w-full min-w-[680px] text-sm">
                <thead class="bg-gray-50 text-xs text-gray-500 dark:bg-dark-700/60 dark:text-dark-400">
                  <tr>
                    <th class="px-3 py-2 text-left">{{ t('admin.channels.form.inputPrice', 'Input') }}</th>
                    <th class="px-3 py-2 text-left">{{ t('admin.channels.form.outputPrice', 'Output') }}</th>
                    <th class="px-3 py-2 text-left">{{ t('admin.channels.form.cacheWritePrice', 'Cache write') }}</th>
                    <th class="px-3 py-2 text-left">{{ t('admin.channels.form.cacheReadPrice', 'Cache read') }}</th>
                    <th class="px-3 py-2 text-left">{{ t('admin.channels.form.imageOutputPrice', 'Image output') }}</th>
                    <th class="px-3 py-2 text-left">{{ t('admin.channels.form.perRequestPrice', 'Per request') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr class="text-gray-800 dark:text-dark-100">
                    <td class="px-3 py-2 tabular-nums">{{ tokenPrice(source.pricing.input_price) }}</td>
                    <td class="px-3 py-2 tabular-nums">{{ tokenPrice(source.pricing.output_price) }}</td>
                    <td class="px-3 py-2 tabular-nums">{{ tokenPrice(source.pricing.cache_write_price) }}</td>
                    <td class="px-3 py-2 tabular-nums">{{ tokenPrice(source.pricing.cache_read_price) }}</td>
                    <td class="px-3 py-2 tabular-nums">{{ requestPrice(source.pricing.image_output_price) }}</td>
                    <td class="px-3 py-2 tabular-nums">{{ requestPrice(source.pricing.per_request_price) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div v-if="source.pricing.intervals.length > 0">
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.modelPricing.intervals', 'Pricing intervals') }}</p>
              <div class="mt-2 overflow-x-auto rounded-lg border border-gray-100 dark:border-dark-700">
                <table class="w-full min-w-[760px] text-sm">
                  <thead class="bg-gray-50 text-xs text-gray-500 dark:bg-dark-700/60 dark:text-dark-400">
                    <tr>
                      <th class="px-3 py-2 text-left">{{ t('admin.modelPricing.interval', 'Interval') }}</th>
                      <th class="px-3 py-2 text-left">{{ t('admin.modelPricing.tier', 'Tier') }}</th>
                      <th class="px-3 py-2 text-left">{{ t('admin.channels.form.inputPrice', 'Input') }}</th>
                      <th class="px-3 py-2 text-left">{{ t('admin.channels.form.outputPrice', 'Output') }}</th>
                      <th class="px-3 py-2 text-left">{{ t('admin.channels.form.cacheWritePrice', 'Cache write') }}</th>
                      <th class="px-3 py-2 text-left">{{ t('admin.channels.form.cacheReadPrice', 'Cache read') }}</th>
                      <th class="px-3 py-2 text-left">{{ t('admin.channels.form.perRequestPrice', 'Per request') }}</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                    <tr v-for="interval in source.pricing.intervals" :key="`${interval.id ?? interval.sort_order}:${interval.tier_label}`">
                      <td class="px-3 py-2 tabular-nums text-gray-700 dark:text-dark-200">{{ intervalRange(interval.min_tokens, interval.max_tokens) }}</td>
                      <td class="px-3 py-2 text-gray-700 dark:text-dark-200">{{ interval.tier_label || '-' }}</td>
                      <td class="px-3 py-2 tabular-nums">{{ tokenPrice(interval.input_price) }}</td>
                      <td class="px-3 py-2 tabular-nums">{{ tokenPrice(interval.output_price) }}</td>
                      <td class="px-3 py-2 tabular-nums">{{ tokenPrice(interval.cache_write_price) }}</td>
                      <td class="px-3 py-2 tabular-nums">{{ tokenPrice(interval.cache_read_price) }}</td>
                      <td class="px-3 py-2 tabular-nums">{{ requestPrice(interval.per_request_price) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>

    <template #footer>
      <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <button type="button" class="btn btn-secondary" @click="$emit('close')">
          {{ t('common.close', 'Close') }}
        </button>
        <button type="button" class="btn btn-primary" @click="$emit('configure')">
          <Icon name="cog" size="sm" />
          {{ t('admin.modelPricing.channelConfiguration', 'Channel configuration') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { BillingMode } from '@/constants/channel'
import type { AdminModelPricingCard, AdminModelPricingSource } from '@/utils/adminModelPricing'
import { formatScaled } from '@/utils/pricing'
import { platformLabel } from '@/utils/platformColors'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  show: boolean
  model: AdminModelPricingCard | null
  groups: Record<number, { name: string; platform: string }>
}>()

defineEmits<{
  close: []
  configure: []
  'edit-source': [channelId: number]
}>()

const { t } = useI18n()

function billingModeLabel(mode: BillingMode): string {
  switch (mode) {
    case 'token': return t('admin.channels.billingMode.token', 'By token')
    case 'per_request': return t('admin.channels.billingMode.perRequest', 'By request')
    case 'image': return t('admin.channels.billingMode.image', 'By image')
  }
}

function billingSourceLabel(source: string): string {
  return t(`admin.availableChannels.billingSource.${source}`, source)
}

function associatedGroupIds(source: AdminModelPricingSource): number[] {
  return source.groupIds.filter((groupId) => {
    const group = props.groups[groupId]
    return !group || group.platform === source.pricing.platform
  })
}

function tokenPrice(value: number | null): string {
  const formatted = formatScaled(value, 1_000_000)
  return formatted === '-' ? formatted : `${formatted} / 1M`
}

function requestPrice(value: number | null): string {
  return formatScaled(value, 1)
}

function intervalRange(min: number, max: number | null): string {
  return max == null ? `>${min}` : `${min}-${max}`
}
</script>
