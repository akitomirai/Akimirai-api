<template>
  <div class="space-y-6">
    <ModelUsageTrendChart
      :trend-data="trendData"
      metric="actual_cost"
      :loading="loading"
    />
    <ModelUsageTrendChart
      :trend-data="trendData"
      metric="requests"
      :loading="loading"
    />
  </div>
</template>

<script setup lang="ts">
import ModelUsageTrendChart from '@/components/charts/ModelUsageTrendChart.vue'
import type { ModelUsageTrendPoint } from '@/api/usage'

defineProps<{
  trendData: ModelUsageTrendPoint[]
  loading: boolean
}>()
</script>
