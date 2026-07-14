<template>
  <div class="flex items-stretch gap-2">
    <span
      class="w-1 shrink-0 rounded-full"
      :class="barClass"
      aria-hidden="true"
    ></span>
    <div class="grid grid-cols-[max-content_max-content] items-baseline gap-x-2 gap-y-0.5 text-xs">
      <span class="text-gray-400 dark:text-gray-500">{{ t('usage.latencyFirstToken') }}</span>
      <span
        v-if="firstTokenMs != null"
        class="font-medium tabular-nums"
        :class="LATENCY_TEXT_CLASSES[firstTokenSeverity(firstTokenMs)]"
      >
        {{ formatUsageDuration(firstTokenMs) }}
      </span>
      <span v-else class="text-gray-400 dark:text-gray-500">-</span>
      <span class="text-gray-400 dark:text-gray-500">{{ t('usage.latencyDuration') }}</span>
      <span
        class="font-medium tabular-nums"
        :class="LATENCY_TEXT_CLASSES[durationSeverity(durationMs ?? 0)]"
      >
        {{ formatUsageDuration(durationMs) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  LATENCY_BAR_CLASSES,
  LATENCY_BAR_FROM_CLASSES,
  LATENCY_BAR_TO_CLASSES,
  LATENCY_TEXT_CLASSES,
  durationSeverity,
  firstTokenSeverity,
} from '@/utils/latencyHealth'
import { formatUsageDuration } from '@/utils/usageMetrics'

const props = defineProps<{
  firstTokenMs?: number | null
  durationMs?: number | null
}>()

const { t } = useI18n()

const barClass = computed(() => {
  if (props.firstTokenMs == null) {
    return LATENCY_BAR_CLASSES[durationSeverity(props.durationMs ?? 0)]
  }

  return [
    'bg-gradient-to-b from-40% to-60%',
    LATENCY_BAR_FROM_CLASSES[firstTokenSeverity(props.firstTokenMs)],
    LATENCY_BAR_TO_CLASSES[durationSeverity(props.durationMs ?? 0)],
  ]
})
</script>
