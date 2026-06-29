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
          <QuickStartExamples
            :base-url="baseUrl"
            :model="recommendedModel"
          />
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import QuickStartExamples from '@/components/user/quickstart/QuickStartExamples.vue'
import { useAppStore } from '@/stores'
import { usageAPI } from '@/api/usage'
import { getBrowserOriginFallback, normalizeOpenAIBaseUrl } from '@/utils/quickStart'
import type { ModelStat } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const loadError = ref(false)
const modelStats = ref<ModelStat[]>([])

const baseUrl = computed(() =>
  normalizeOpenAIBaseUrl(appStore.cachedPublicSettings?.api_base_url, getBrowserOriginFallback())
)

const recommendedModel = computed(() => modelStats.value[0]?.model || '')

async function loadData() {
  loading.value = true
  loadError.value = false
  try {
    await appStore.fetchPublicSettings()
    const models = await usageAPI.getDashboardModels()
    modelStats.value = models.models || []
  } catch (error) {
    console.error('[QuickStartView] failed to load quick start data:', error)
    loadError.value = true
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>
