<template>
  <div v-if="hasKnownTokenMetrics(metrics)" class="space-y-1 text-xs tabular-nums">
    <div class="flex items-center gap-3 whitespace-nowrap">
      <span class="inline-flex items-center gap-1 text-emerald-600 dark:text-emerald-400">
        <Icon name="arrowDown" size="xs" />
        <span>{{ formatTokenCount(metrics.input_tokens ?? 0, { unitStyle }) }}</span>
      </span>
      <span class="inline-flex items-center gap-1 text-violet-600 dark:text-violet-400">
        <Icon name="arrowUp" size="xs" />
        <span>{{ formatTokenCount(metrics.output_tokens ?? 0, { unitStyle }) }}</span>
      </span>
    </div>
    <div class="inline-flex items-center gap-1 whitespace-nowrap text-sky-600 dark:text-sky-400">
      <Icon name="database" size="xs" />
      <span>{{ formatTokenCount(metrics.cache_read_tokens ?? 0, { unitStyle }) }}</span>
    </div>
  </div>
  <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import { formatTokenCount } from '@/utils/format'
import { hasKnownTokenMetrics, type UsageTokenMetrics } from '@/utils/usageMetrics'

withDefaults(defineProps<{
  metrics: UsageTokenMetrics
  unitStyle?: 'locale' | 'wan-latin'
}>(), {
  unitStyle: 'locale',
})
</script>
