<template>
  <div
    v-if="ratio !== null"
    data-test="cache-hit-rate"
    class="w-[112px] min-w-[112px] max-w-[112px] text-left"
  >
    <div class="flex items-baseline justify-between gap-2 whitespace-nowrap">
      <span class="font-semibold text-sky-600 dark:text-sky-400">{{ ratio }}%</span>
      <span class="text-left text-xs text-gray-400 dark:text-gray-500">
        {{ formatTokenCount(metrics.cache_read_tokens ?? 0, { unitStyle }) }}
      </span>
    </div>
    <div class="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-sky-100 dark:bg-sky-950/50">
      <div
        class="h-full rounded-full bg-sky-500 dark:bg-sky-400"
        :style="{ width: cacheHitProgressWidth(metrics) }"
      ></div>
    </div>
  </div>
  <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatTokenCount } from '@/utils/format'
import {
  cacheHitProgressWidth,
  cacheHitRatio,
  type UsageTokenMetrics,
} from '@/utils/usageMetrics'

const props = withDefaults(defineProps<{
  metrics: UsageTokenMetrics
  unitStyle?: 'locale' | 'wan-latin'
}>(), {
  unitStyle: 'locale',
})

const ratio = computed(() => cacheHitRatio(props.metrics))
</script>
