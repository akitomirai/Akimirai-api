<template>
  <section class="card p-4">
    <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
      <div>
        <p class="text-sm font-medium text-gray-500 dark:text-dark-400">
          {{ t('availableChannels.modelCatalog.kicker', 'Model Catalog') }}
        </p>
        <h2 class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('availableChannels.modelCatalog.title', 'Available Models') }}
        </h2>
        <p class="mt-1 max-w-3xl text-xs text-gray-500 dark:text-dark-400">
          {{ MODEL_MULTIPLIER_EXPLANATION }}
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <router-link class="btn btn-sm btn-secondary" to="/quick-start">
          <Icon name="book" size="sm" />
          {{ t('availableChannels.modelCatalog.quickStart', 'Quick Start') }}
        </router-link>
        <button class="btn btn-sm btn-secondary" type="button" @click="$emit('retry')">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          {{ t('common.refresh', 'Refresh') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="mt-4 py-8 text-center text-sm text-gray-500 dark:text-dark-400">
      <Icon name="refresh" size="lg" class="inline-block animate-spin text-gray-400" />
      <p class="mt-2">{{ t('common.loading', 'Loading...') }}</p>
    </div>

    <div
      v-else-if="error"
      class="mt-4 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300"
    >
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <span>{{ t('availableChannels.modelCatalog.loadFailed', 'Failed to load available models.') }}</span>
        <button class="btn btn-sm btn-secondary" type="button" @click="$emit('retry')">
          {{ t('common.refresh', 'Refresh') }}
        </button>
      </div>
    </div>

    <div v-else-if="items.length === 0" class="mt-4 py-8 text-center text-sm text-gray-500 dark:text-dark-400">
      <Icon name="inbox" size="xl" class="mx-auto mb-3 h-12 w-12 text-gray-400" />
      <p>{{ t('availableChannels.modelCatalog.empty', 'No available models yet.') }}</p>
      <p class="mt-1 text-xs">
        {{ t('availableChannels.modelCatalog.emptyHint', 'No model data will be fabricated here.') }}
      </p>
    </div>

    <div v-else class="mt-4 overflow-x-auto">
      <table class="w-full min-w-[760px] table-fixed border-collapse text-sm">
        <thead>
          <tr class="border-b border-gray-100 text-xs font-medium uppercase text-gray-500 dark:border-dark-700 dark:text-gray-400">
            <th class="w-[230px] px-3 py-2 text-left">{{ t('availableChannels.modelCatalog.columns.model', 'Model ID') }}</th>
            <th class="w-[130px] px-3 py-2 text-left">{{ t('availableChannels.modelCatalog.columns.status', 'Status') }}</th>
            <th class="w-[120px] px-3 py-2 text-left">{{ t('availableChannels.modelCatalog.columns.platform', 'Platform') }}</th>
            <th class="w-[140px] px-3 py-2 text-left">{{ t('availableChannels.modelCatalog.columns.multiplier', 'Multiplier') }}</th>
            <th class="px-3 py-2 text-left">{{ t('availableChannels.modelCatalog.columns.channels', 'Visible source') }}</th>
            <th class="w-[210px] px-3 py-2 text-right">{{ t('common.actions', 'Actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in items"
            :key="`${item.provider}-${item.id}`"
            class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
          >
            <td class="px-3 py-3 align-top">
              <code class="block truncate rounded bg-gray-50 px-2 py-1 font-mono text-xs text-gray-900 dark:bg-dark-800 dark:text-dark-100">
                {{ item.id }}
              </code>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                {{ capabilitySummary(item) }}
              </p>
            </td>
            <td class="px-3 py-3 align-top">
              <span class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium" :class="statusClass(item.status)">
                {{ statusLabel(item.status) }}
              </span>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                {{ item.statusReason || '-' }}
              </p>
            </td>
            <td class="px-3 py-3 align-top">
              <span class="inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase" :class="platformBadgeClass(item.provider)">
                <PlatformIcon :platform="item.provider as GroupPlatform" size="xs" />
                {{ item.provider }}
              </span>
            </td>
            <td class="px-3 py-3 align-top">
              <span class="text-sm font-medium text-gray-900 dark:text-white">{{ formatMultiplierRange(item) }}</span>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                {{ pricingLabel(item) }}
              </p>
            </td>
            <td class="px-3 py-3 align-top">
              <div class="space-y-1">
                <p class="truncate text-sm text-gray-700 dark:text-dark-200">
                  {{ item.channelNames.join(', ') || '-' }}
                </p>
                <p class="truncate text-xs text-gray-500 dark:text-dark-400">
                  {{ sourceSummary(item) }}
                </p>
              </div>
            </td>
            <td class="px-3 py-3 align-top">
              <div class="flex justify-end gap-2">
                <button class="btn btn-sm btn-secondary" type="button" @click="copyModel(item.id)">
                  <Icon name="copy" size="sm" />
                  {{ copiedModel === item.id ? t('common.copied', 'Copied') : t('common.copy', 'Copy') }}
                </button>
                <router-link class="btn btn-sm btn-primary" :to="quickStartLink(item)">
                  <Icon name="externalLink" size="sm" />
                  {{ t('availableChannels.modelCatalog.useModel', 'Use') }}
                </router-link>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { platformBadgeClass } from '@/utils/platformColors'
import {
  formatMultiplierRange,
  MODEL_MULTIPLIER_EXPLANATION,
  type ModelAvailabilityStatus,
  type ModelCatalogItem,
} from '@/utils/modelCatalog'
import type { GroupPlatform } from '@/types'

defineProps<{
  items: ModelCatalogItem[]
  loading: boolean
  error: boolean
}>()

defineEmits<{
  retry: []
}>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const copiedModel = ref('')

function statusLabel(status: ModelAvailabilityStatus): string {
  switch (status) {
    case 'available':
      return t('availableChannels.modelCatalog.status.available', 'Available')
    case 'maintenance':
      return t('availableChannels.modelCatalog.status.maintenance', 'Maintenance')
    case 'unavailable':
      return t('availableChannels.modelCatalog.status.unavailable', 'Unavailable')
    default:
      return t('availableChannels.modelCatalog.status.unknown', 'Unknown')
  }
}

function statusClass(status: ModelAvailabilityStatus): string {
  switch (status) {
    case 'available':
      return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300'
    case 'maintenance':
      return 'bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-300'
    case 'unavailable':
      return 'bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-300'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'
  }
}

function pricingLabel(item: ModelCatalogItem): string {
  if (item.billingDescription) {
    return item.billingDescription
  }
  if (!item.pricing) {
    return t('availableChannels.modelCatalog.noPricing', 'No model pricing details')
  }
  return t('availableChannels.modelCatalog.pricingAvailable', 'Pricing details available')
}

function capabilitySummary(item: ModelCatalogItem): string {
  const parts = [
    item.family,
    item.contextWindow ? `${item.contextWindow} context` : null,
    capabilityLabel('stream', item.supportsStreaming),
    capabilityLabel('vision', item.supportsVision),
    capabilityLabel('tools', item.supportsTools),
    capabilityLabel('json', item.supportsJson),
  ].filter(Boolean)
  return parts.join(' / ') || t('availableChannels.modelCatalog.capabilityUnknown', 'Capabilities not declared')
}

function capabilityLabel(name: string, supported: boolean | null): string | null {
  if (supported == null) return null
  return supported ? name : `${name}: no`
}

function sourceSummary(item: ModelCatalogItem): string {
  const groupNames = item.groups.map((group) => group.name).join(', ')
  const count = `${item.availableChannelCount} available channel(s)`
  return groupNames ? `${groupNames} / ${count}` : count
}

function quickStartLink(item: ModelCatalogItem) {
  if (item.quickStartUrl) return item.quickStartUrl
  return { path: '/quick-start', query: { model: item.id } }
}

async function copyModel(modelId: string) {
  const ok = await copyToClipboard(modelId)
  if (!ok) return
  copiedModel.value = modelId
  window.setTimeout(() => {
    if (copiedModel.value === modelId) copiedModel.value = ''
  }, 1600)
}
</script>
