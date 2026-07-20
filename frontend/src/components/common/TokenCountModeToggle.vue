<template>
  <div
    class="inline-flex rounded-lg bg-gray-100 p-1 dark:bg-dark-700"
    role="group"
    :aria-label="t('usage.countMode.label')"
    data-testid="token-count-mode-toggle"
  >
    <button
      v-for="option in options"
      :key="option.value"
      type="button"
      class="min-h-9 rounded-md px-3 py-1.5 text-sm font-medium transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-1"
      :class="mode === option.value
        ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-800 dark:text-white'
        : 'text-gray-500 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
      :aria-pressed="mode === option.value"
      @click="setMode(option.value)"
    >
      {{ option.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useTokenCountMode, type TokenCountMode } from '@/composables/useTokenCountMode'

const { t } = useI18n()
const { mode, setMode } = useTokenCountMode()

const options = computed<Array<{ value: TokenCountMode; label: string }>>(() => [
  { value: 'modern', label: t('usage.countMode.modern') },
  { value: 'legacy', label: t('usage.countMode.legacy') },
])
</script>
