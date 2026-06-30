<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-sm font-medium text-primary-600 dark:text-primary-400">
              {{ t('quickStart.kicker') }}
            </p>
            <h1 class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ t('quickStart.title') }}
            </h1>
            <p class="mt-2 max-w-3xl text-sm text-gray-600 dark:text-dark-300">
              {{ t('quickStart.description') }}
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <router-link class="btn btn-secondary" to="/available-channels">
              <Icon name="server" size="sm" />
              {{ t('quickStart.viewModels') }}
            </router-link>
            <router-link class="btn btn-secondary" to="/keys">
              <Icon name="key" size="sm" />
              {{ t('quickStart.createKey') }}
            </router-link>
            <router-link class="btn btn-secondary" to="/usage">
              <Icon name="chart" size="sm" />
              {{ t('quickStart.viewUsage') }}
            </router-link>
          </div>
        </div>

        <div v-if="loading" class="mt-6 flex justify-center py-8">
          <LoadingSpinner />
        </div>

        <div v-else-if="loadError" class="mt-6 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300">
          {{ t('quickStart.loadFailed') }}
        </div>

        <div v-else class="mt-6">
          <div
            v-if="queryModelHint"
            class="mb-4 rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-200"
          >
            {{ queryModelHint }}
          </div>
          <QuickStartExamples
            :base-url="baseUrl"
            :model="selectedModel"
          />
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import QuickStartExamples from '@/components/user/quickstart/QuickStartExamples.vue'
import { useAppStore } from '@/stores'
import userChannelsAPI from '@/api/channels'
import { getBrowserOriginFallback, normalizeOpenAIBaseUrl } from '@/utils/quickStart'
import {
  selectQuickStartCatalogModel,
  toModelCatalogItems,
  type ModelCatalogItem,
} from '@/utils/modelCatalog'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()

const loading = ref(false)
const loadError = ref(false)
const modelCatalog = ref<ModelCatalogItem[]>([])

const baseUrl = computed(() =>
  normalizeOpenAIBaseUrl(appStore.cachedPublicSettings?.api_base_url, getBrowserOriginFallback())
)

const quickStartSelection = computed(() => selectQuickStartCatalogModel(modelCatalog.value, route.query.model))

const selectedModel = computed(() => quickStartSelection.value.selected?.id || '')

const queryModelHint = computed(() => {
  const selection = quickStartSelection.value
  if (!selection.requested || !selection.usedFallback) return ''
  if (selectedModel.value) {
    return t(
      'quickStart.modelFallbackHint',
      `Requested model ${selection.requested} is not currently available. Examples now use ${selectedModel.value}.`
    )
  }
  return t(
    'quickStart.modelUnavailableHint',
    `Requested model ${selection.requested} is not currently available. Choose a model from the model list.`
  )
})

async function loadData() {
  loading.value = true
  loadError.value = false
  try {
    await appStore.fetchPublicSettings()
    const catalog = await userChannelsAPI.getModelCatalog()
    modelCatalog.value = toModelCatalogItems(catalog)
  } catch (error) {
    console.error('[QuickStartView] failed to load quick start data:', error)
    loadError.value = true
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>
