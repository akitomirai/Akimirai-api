<template>
  <section
    data-testid="profile-privacy-filter-card"
    class="card overflow-hidden border border-gray-100 bg-white/90 dark:border-dark-700 dark:bg-dark-900/50"
  >
    <div class="flex items-start justify-between gap-4 border-b border-gray-100 px-6 py-5 dark:border-dark-700">
      <div class="min-w-0">
        <div class="flex items-center gap-3">
          <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-300">
            <Icon name="shield" size="md" />
          </span>
          <div>
            <h2 class="text-lg font-medium text-gray-900 dark:text-white">
              {{ t('profile.privacyFilter.title') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('profile.privacyFilter.description') }}
            </p>
          </div>
        </div>
        <p class="mt-3 text-sm text-gray-500 dark:text-gray-400">
          {{ t('profile.privacyFilter.noRetention') }}
        </p>
      </div>

      <Toggle
        :model-value="privacyConfig.enabled"
        :disabled="saving"
        data-testid="privacy-filter-enabled-toggle"
        :aria-label="t('profile.privacyFilter.enabled')"
        @update:model-value="handleEnabledChange"
      />
    </div>

    <div class="grid gap-3 px-6 py-5 sm:grid-cols-2 lg:grid-cols-3">
      <label
        v-for="option in privacyFilterOptions"
        :key="option.value"
        :data-testid="`privacy-filter-type-${option.value}`"
        :class="[
          'flex min-h-[44px] items-center justify-between gap-3 rounded-lg border px-4 py-3 text-left transition',
          isTypeSelected(option.value)
            ? 'border-primary-300 bg-primary-50 text-primary-800 dark:border-primary-700 dark:bg-primary-900/20 dark:text-primary-100'
            : 'border-gray-200 bg-gray-50/80 text-gray-700 hover:border-primary-200 hover:bg-white dark:border-dark-700 dark:bg-dark-800/60 dark:text-gray-300 dark:hover:border-primary-800 dark:hover:bg-dark-800',
          saving ? 'cursor-wait opacity-70' : 'cursor-pointer',
        ]"
      >
        <input
          type="checkbox"
          class="sr-only"
          :checked="isTypeSelected(option.value)"
          :disabled="saving"
          :data-testid="`privacy-filter-type-${option.value}-input`"
          :aria-label="option.label"
          @change="handleTypeToggle(option.value)"
        >
        <span class="flex min-w-0 items-center gap-2">
          <span class="truncate text-sm font-medium">{{ option.label }}</span>
          <Icon
            name="questionCircle"
            size="xs"
            class="shrink-0 text-gray-400"
            :title="option.hint"
          />
        </span>
        <span
          :class="[
            'flex h-5 w-5 shrink-0 items-center justify-center rounded-md border',
            isTypeSelected(option.value)
              ? 'border-primary-500 bg-primary-500 text-white'
              : 'border-gray-300 bg-white text-transparent dark:border-dark-600 dark:bg-dark-900',
          ]"
        >
          <Icon name="check" size="xs" :stroke-width="2" />
        </span>
      </label>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@/components/icons'
import Toggle from '@/components/common/Toggle.vue'
import { userAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  DEFAULT_PRIVACY_FILTER_TYPES,
  normalizePrivacyFilterConfig,
  type PrivacyFilterConfig,
  type PrivacyFilterType,
} from '@/utils/privacyFilter'
import type { User } from '@/types'

const props = defineProps<{
  config?: PrivacyFilterConfig | null
}>()

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const saving = ref(false)
const privacyConfig = ref(normalizePrivacyFilterConfig(props.config))

const privacyFilterOptions = computed(() =>
  DEFAULT_PRIVACY_FILTER_TYPES.map((value) => ({
    value,
    label: t(`profile.privacyFilter.types.${value}`),
    hint: t(`profile.privacyFilter.hints.${value}`),
  })),
)

watch(
  () => props.config,
  (next) => {
    privacyConfig.value = normalizePrivacyFilterConfig(next)
  },
  { deep: true },
)

function isTypeSelected(type: PrivacyFilterType): boolean {
  return privacyConfig.value.types.includes(type)
}

async function handleEnabledChange(enabled: boolean): Promise<void> {
  await saveConfig(
    {
      ...privacyConfig.value,
      enabled,
    },
    enabled
      ? t('profile.privacyFilter.enabledSuccess')
      : t('profile.privacyFilter.disabledSuccess'),
  )
}

async function handleTypeToggle(type: PrivacyFilterType): Promise<void> {
  const selected = new Set(privacyConfig.value.types)
  if (selected.has(type)) {
    selected.delete(type)
  } else {
    selected.add(type)
  }

  await saveConfig({
    ...privacyConfig.value,
    types: DEFAULT_PRIVACY_FILTER_TYPES.filter((item) => selected.has(item)),
  })
}

async function saveConfig(nextConfig: PrivacyFilterConfig, successMessage = t('common.saved')): Promise<void> {
  const previous = privacyConfig.value
  const normalized = normalizePrivacyFilterConfig(nextConfig)
  privacyConfig.value = normalized
  saving.value = true

  try {
    const updated = await userAPI.updateProfile({ privacy_filter_config: normalized })
    const merged = mergeUpdatedUserPrivacyConfig(updated, normalized)
    authStore.user = merged.user
    privacyConfig.value = merged.config
    appStore.showSuccess(successMessage)
  } catch (err: unknown) {
    privacyConfig.value = previous
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    saving.value = false
  }
}

function mergeUpdatedUserPrivacyConfig(
  updated: User,
  submittedConfig: PrivacyFilterConfig,
): { user: User; config: PrivacyFilterConfig } {
  const returnedConfig = hasReturnedPrivacyFilterConfig(updated)
    ? normalizePrivacyFilterConfig(updated.privacy_filter_config)
    : submittedConfig
  return {
    user: {
      ...updated,
      privacy_filter_config: returnedConfig,
    },
    config: returnedConfig,
  }
}

function hasReturnedPrivacyFilterConfig(
  updated: User,
): updated is User & { privacy_filter_config: PrivacyFilterConfig } {
  return (
    typeof updated.privacy_filter_config?.enabled === 'boolean' &&
    Array.isArray(updated.privacy_filter_config.types)
  )
}
</script>
