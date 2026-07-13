<template>
  <AppLayout>
    <div class="space-y-5">
      <header class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ text('title') }}</h1>
        </div>
        <button class="btn btn-secondary self-start sm:self-auto" type="button" :disabled="loading" @click="loadCatalog">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          {{ text('refresh') }}
        </button>
      </header>

      <ModelPlazaFilters
        :model-value="filters"
        :options="filterOptions"
        @update:model-value="filters = $event"
        @reset="resetFilters"
      />

      <div class="flex items-center justify-between gap-3 text-xs text-gray-500 dark:text-dark-400">
        <span>{{ text('results', { count: filteredModels.length }) }}</span>
      </div>

      <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3" aria-live="polite">
        <div v-for="index in 6" :key="index" class="min-h-[18rem] animate-pulse rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800/60">
          <div class="h-10 w-2/3 rounded bg-gray-100 dark:bg-dark-700" />
          <div class="mt-6 h-4 w-full rounded bg-gray-100 dark:bg-dark-700" />
          <div class="mt-3 h-4 w-4/5 rounded bg-gray-100 dark:bg-dark-700" />
          <div class="mt-10 h-20 rounded bg-gray-100 dark:bg-dark-700" />
        </div>
        <span class="sr-only">{{ text('loading') }}</span>
      </div>

      <section v-else-if="featureDisabled" class="rounded-lg border border-amber-200 bg-amber-50 p-8 text-center dark:border-amber-800 dark:bg-amber-900/20">
        <Icon name="inbox" size="xl" class="mx-auto text-amber-500" />
        <p class="mt-3 text-sm font-medium text-amber-800 dark:text-amber-200">{{ text('disabled') }}</p>
      </section>

      <section v-else-if="loadError" class="rounded-lg border border-red-200 bg-red-50 p-8 text-center dark:border-red-800 dark:bg-red-900/20">
        <p class="text-sm font-medium text-red-700 dark:text-red-300">{{ text('loadFailed') }}</p>
        <p v-if="errorMessage" class="mt-1 text-xs text-red-600/80 dark:text-red-300/80">{{ errorMessage }}</p>
        <button class="btn btn-sm btn-secondary mt-4" type="button" @click="loadCatalog">{{ text('retry') }}</button>
      </section>

      <section v-else-if="filteredModels.length === 0" class="rounded-lg border border-gray-200 bg-white p-10 text-center dark:border-dark-700 dark:bg-dark-800/50">
        <Icon name="inbox" size="xl" class="mx-auto text-gray-400" />
        <p class="mt-3 text-sm font-medium text-gray-800 dark:text-dark-100">{{ text('empty') }}</p>
        <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ text('emptyHint') }}</p>
        <button class="btn btn-sm btn-secondary mt-4" type="button" @click="resetFilters">{{ text('reset') }}</button>
      </section>

      <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <ModelPlazaCard
          v-for="model in filteredModels"
          :key="`${model.provider}-${model.id}`"
          :item="model"
          @copy="copyModel"
          @details="openDetails"
        />
      </div>
    </div>

    <ModelPlazaDetailDialog
      :show="selectedModel !== null"
      :model="selectedModel"
      :base-url="baseUrl"
      @close="closeDetails"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import ModelPlazaFilters from '@/components/model-plaza/ModelPlazaFilters.vue'
import ModelPlazaCard from '@/components/model-plaza/ModelPlazaCard.vue'
import ModelPlazaDetailDialog from '@/components/model-plaza/ModelPlazaDetailDialog.vue'
import { useModelPlazaText } from '@/components/model-plaza/modelPlazaText'
import userChannelsAPI from '@/api/channels'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { extractApiErrorMessage } from '@/utils/apiError'
import { getBrowserOriginFallback, normalizeOpenAIBaseUrl } from '@/utils/quickStart'
import {
  filterModelCatalogItems,
  findCatalogModel,
  getModelCatalogFilterOptions,
  toModelCatalogItems,
  type ModelCatalogFilters,
  type ModelCatalogItem,
} from '@/utils/modelCatalog'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const { text } = useModelPlazaText()

const loading = ref(false)
const loadError = ref(false)
const featureDisabled = ref(false)
const errorMessage = ref('')
const models = ref<ModelCatalogItem[]>([])
const selectedModel = ref<ModelCatalogItem | null>(null)
const filters = ref<ModelCatalogFilters>(emptyFilters())
let requestController: AbortController | null = null

const baseUrl = computed(() => normalizeOpenAIBaseUrl(
  appStore.cachedPublicSettings?.api_base_url,
  getBrowserOriginFallback(),
))
const filterOptions = computed(() => getModelCatalogFilterOptions(models.value))
const filteredModels = computed(() => filterModelCatalogItems(models.value, filters.value))

watch([models, () => route.query.model], () => {
  const requested = route.query.model
  selectedModel.value = requested ? findCatalogModel(models.value, requested) : null
}, { immediate: true })

async function loadCatalog(): Promise<void> {
  requestController?.abort()
  const controller = new AbortController()
  requestController = controller
  loading.value = true
  loadError.value = false
  featureDisabled.value = false
  errorMessage.value = ''

  try {
    const settings = await appStore.fetchPublicSettings()
    if (settings?.available_channels_enabled === false) {
      featureDisabled.value = true
      models.value = []
      return
    }
    const payload = await userChannelsAPI.getModelCatalog({ signal: controller.signal })
    models.value = toModelCatalogItems(payload)
  } catch (error) {
    if (controller.signal.aborted) return
    loadError.value = true
    models.value = []
    errorMessage.value = extractApiErrorMessage(error, text('loadFailed'))
  } finally {
    if (requestController === controller) loading.value = false
  }
}

function resetFilters(): void {
  filters.value = emptyFilters()
}

function copyModel(modelId: string): void {
  void copyToClipboard(modelId, text('copied'))
}

function openDetails(model: ModelCatalogItem): void {
  selectedModel.value = model
  void router.replace({ query: { ...route.query, model: model.modelId } })
}

function closeDetails(): void {
  selectedModel.value = null
  const query = { ...route.query }
  delete query.model
  void router.replace({ query })
}

function emptyFilters(): ModelCatalogFilters {
  return { query: '', group: '', provider: '', billingMode: '', status: '', capability: '' }
}

onMounted(loadCatalog)
onBeforeUnmount(() => requestController?.abort())
</script>
