<template>
  <div class="space-y-6">
    <ModelUsageTrendChart
      :trend-data="trendData"
      metric="actual_cost"
      :loading="loading"
      :granularity="granularity"
      :range-start="rangeStart"
      :range-end="rangeEnd"
      :timezone="timezone"
    />
    <ModelUsageTrendChart
      :trend-data="trendData"
      metric="total_tokens"
      :loading="loading"
      :granularity="granularity"
      :range-start="rangeStart"
      :range-end="rangeEnd"
      :timezone="timezone"
    />
  </div>
</template>

<script setup lang="ts">
import ModelUsageTrendChart from '@/components/charts/ModelUsageTrendChart.vue'
import type { ModelUsageTrendPoint } from '@/api/usage'
import type { ModelUsageTrendGranularity } from '@/api/usage'

defineProps<{
  trendData: ModelUsageTrendPoint[]
  loading: boolean
  granularity: ModelUsageTrendGranularity
  rangeStart?: string | null
  rangeEnd?: string | null
  timezone?: string
}>()
</script>
