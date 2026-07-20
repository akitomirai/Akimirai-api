<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
            {{ t('admin.modelPricing.title', 'Model Pricing') }}
          </h1>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <button type="button" class="btn btn-secondary" @click="goToChannelConfiguration">
            <Icon name="cog" size="sm" />
            {{ t('admin.modelPricing.channelConfiguration', 'Channel configuration') }}
          </button>
          <button
            type="button"
            class="btn btn-secondary btn-icon"
            :disabled="loading"
            :title="t('common.refresh', 'Refresh')"
            @click="loadModelPricing"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <section class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800/70">
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-[minmax(260px,1fr)_180px_180px_180px_auto]">
          <div class="relative sm:col-span-2 xl:col-span-1">
            <Icon
              name="search"
              size="md"
              class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-dark-400"
            />
            <input
              v-model="filters.search"
              type="search"
              class="input pl-10"
              :placeholder="t('admin.modelPricing.searchPlaceholder', 'Search models or source channels...')"
            />
          </div>
          <Select v-model="filters.provider" :options="providerOptions" />
          <Select v-model="filters.billingMode" :options="billingModeOptions" />
          <Select v-model="filters.status" :options="statusOptions" />
          <button
            type="button"
            class="btn btn-secondary xl:px-3"
            :disabled="!hasActiveFilters"
            @click="resetFilters"
          >
            <Icon name="x" size="sm" />
            {{ t('common.reset', 'Reset') }}
          </button>
        </div>
        <div class="mt-3 text-xs text-gray-500 dark:text-dark-400">
          <span>{{ t('admin.modelPricing.resultCount', { count: filteredCards.length }, `${filteredCards.length} model(s)`) }}</span>
        </div>
      </section>

      <div v-if="loading" class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        <div
          v-for="index in 6"
          :key="index"
          class="h-[236px] animate-pulse rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800/70"
        >
          <div class="h-10 w-2/3 rounded bg-gray-100 dark:bg-dark-700"></div>
          <div class="mt-5 h-7 w-1/2 rounded bg-gray-100 dark:bg-dark-700"></div>
          <div class="mt-4 h-14 rounded bg-gray-100 dark:bg-dark-700"></div>
        </div>
      </div>

      <div
        v-else-if="loadError"
        class="rounded-lg border border-red-200 bg-red-50 px-5 py-8 text-center dark:border-red-800 dark:bg-red-900/20"
      >
        <Icon name="exclamationCircle" size="xl" class="mx-auto text-red-500" />
        <p class="mt-3 text-sm font-medium text-red-700 dark:text-red-300">
          {{ t('admin.modelPricing.loadError', 'Failed to load model pricing.') }}
        </p>
        <button type="button" class="btn btn-secondary mt-4" @click="loadModelPricing">
          <Icon name="refresh" size="sm" />
          {{ t('common.refresh', 'Refresh') }}
        </button>
      </div>

      <div
        v-else-if="filteredCards.length === 0"
        class="rounded-lg border border-gray-200 bg-white px-5 py-12 text-center dark:border-dark-700 dark:bg-dark-800/70"
      >
        <Icon name="inbox" size="xl" class="mx-auto text-gray-400" />
        <h2 class="mt-3 text-sm font-semibold text-gray-900 dark:text-white">
          {{ cards.length === 0
            ? t('admin.modelPricing.empty', 'No model pricing configured')
            : t('admin.modelPricing.noMatches', 'No models match the current filters') }}
        </h2>
        <p class="mx-auto mt-1 max-w-lg text-xs text-gray-500 dark:text-dark-400">
          {{ cards.length === 0
            ? t('admin.modelPricing.emptyHint', 'Add pricing rules from channel configuration.')
            : t('admin.modelPricing.noMatchesHint', 'Adjust or reset the filters to see more models.') }}
        </p>
        <button
          v-if="cards.length === 0"
          type="button"
          class="btn btn-primary mt-5"
          @click="goToChannelConfiguration"
        >
          <Icon name="cog" size="sm" />
          {{ t('admin.modelPricing.channelConfiguration', 'Channel configuration') }}
        </button>
        <button v-else type="button" class="btn btn-secondary mt-5" @click="resetFilters">
          {{ t('common.reset', 'Reset') }}
        </button>
      </div>

      <div v-else class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        <ModelPricingCard
          v-for="card in filteredCards"
          :key="card.key"
          :card="card"
          @details="selectedModel = $event"
        />
      </div>
    </div>

    <ModelPricingDetailDialog
      :show="selectedModel !== null"
      :model="selectedModel"
      :groups="groupsById"
      @close="selectedModel = null"
      @configure="goToChannelConfiguration"
      @edit-source="editSource"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { adminAPI } from '@/api/admin'
import type { BillingMode, ChannelStatus } from '@/constants/channel'
import type { SelectOption } from '@/components/common/Select.vue'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { platformLabel } from '@/utils/platformColors'
import {
  fetchAllAdminChannels,
  filterAdminModelPricing,
  projectAdminModelPricing,
  type AdminModelPricingCard,
} from '@/utils/adminModelPricing'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import ModelPricingCard from '@/components/admin/channel/ModelPricingCard.vue'
import ModelPricingDetailDialog from '@/components/admin/channel/ModelPricingDetailDialog.vue'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const cards = ref<AdminModelPricingCard[]>([])
const groupsById = ref<Record<number, { name: string; platform: string }>>({})
const selectedModel = ref<AdminModelPricingCard | null>(null)
const loading = ref(false)
const loadError = ref(false)
const filters = reactive<{
  search: string
  provider: string
  billingMode: BillingMode | ''
  status: ChannelStatus | ''
}>({
  search: '',
  provider: '',
  billingMode: '',
  status: '',
})

let abortController: AbortController | null = null

const providerOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.modelPricing.allProviders', 'All providers') },
  ...[...new Set(cards.value.map((card) => card.platform))]
    .sort((a, b) => a.localeCompare(b))
    .map((provider) => ({ value: provider, label: platformLabel(provider) })),
])

const billingModeOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.modelPricing.allBillingModes', 'All billing modes') },
  { value: 'token', label: t('admin.channels.billingMode.token', 'By token') },
  { value: 'per_request', label: t('admin.channels.billingMode.perRequest', 'By request') },
  { value: 'image', label: t('admin.channels.billingMode.image', 'By image') },
])

const statusOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.modelPricing.allStatuses', 'All statuses') },
  { value: 'active', label: t('common.active', 'Active source') },
  { value: 'disabled', label: t('common.disabled', 'Disabled source') },
])

const filteredCards = computed(() => filterAdminModelPricing(cards.value, filters))
const hasActiveFilters = computed(() => Boolean(
  filters.search.trim() || filters.provider || filters.billingMode || filters.status,
))

async function loadModelPricing() {
  abortController?.abort()
  const controller = new AbortController()
  abortController = controller
  loading.value = true
  loadError.value = false

  try {
    const [channels, groups] = await Promise.all([
      fetchAllAdminChannels(adminAPI.channels.list, { signal: controller.signal }),
      adminAPI.groups.getAllIncludingInactive().catch(() => []),
    ])
    if (controller.signal.aborted || abortController !== controller) return

    cards.value = projectAdminModelPricing(channels)
    groupsById.value = Object.fromEntries(groups.map((group) => [group.id, {
      name: group.name,
      platform: group.platform,
    }]))

    if (selectedModel.value) {
      selectedModel.value = cards.value.find((card) => card.key === selectedModel.value?.key) || null
    }
  } catch (error: unknown) {
    const cancellation = error as { name?: string; code?: string }
    if (cancellation.name === 'AbortError' || cancellation.code === 'ERR_CANCELED') return
    cards.value = []
    loadError.value = true
    appStore.showError(extractApiErrorMessage(error, t('admin.modelPricing.loadError', 'Failed to load model pricing.')))
  } finally {
    if (abortController === controller) {
      abortController = null
      loading.value = false
    }
  }
}

function resetFilters() {
  filters.search = ''
  filters.provider = ''
  filters.billingMode = ''
  filters.status = ''
}

function goToChannelConfiguration() {
  selectedModel.value = null
  router.push('/admin/channels/config')
}

function editSource(channelId: number) {
  selectedModel.value = null
  router.push({ path: '/admin/channels/config', query: { edit: String(channelId) } })
}

onMounted(loadModelPricing)
onUnmounted(() => abortController?.abort())
</script>
