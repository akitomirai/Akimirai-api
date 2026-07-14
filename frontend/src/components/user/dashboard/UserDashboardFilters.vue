<template>
  <div class="card p-4">
    <div class="flex flex-wrap items-center gap-4">
      <div class="flex min-w-0 flex-wrap items-center gap-2">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('dashboard.timeRange') }}:
        </span>
        <DateRangePicker
          :start-date="startDate"
          :end-date="endDate"
          :start-time="startTime"
          :end-time="endTime"
          show-time-inputs
          :preset-values="dashboardDatePresets"
          @update:start-date="emit('update:startDate', $event)"
          @update:end-date="emit('update:endDate', $event)"
          @update:start-time="emit('update:startTime', $event)"
          @update:end-time="emit('update:endTime', $event)"
          @change="emit('dateRangeChange', $event)"
        />
      </div>

      <button
        type="button"
        class="btn btn-secondary"
        :disabled="loading"
        @click="emit('refresh')"
      >
        {{ t('common.refresh') }}
      </button>

      <div class="flex items-center gap-2 sm:ml-auto">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('dashboard.granularity') }}:
        </span>
        <div class="w-28">
          <Select
            :model-value="granularity"
            :options="granularityOptions"
            @update:model-value="updateGranularity"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import type { ModelUsageTrendGranularity } from '@/api/usage'
import { dashboardDatePresets, type DateRangeChange } from '@/utils/dashboardTimeRange'

type Granularity = ModelUsageTrendGranularity

defineProps<{
  startDate: string
  endDate: string
  startTime: string
  endTime: string
  granularity: Granularity
  loading: boolean
}>()

const emit = defineEmits<{
  (event: 'update:startDate', value: string): void
  (event: 'update:endDate', value: string): void
  (event: 'update:startTime', value: string): void
  (event: 'update:endTime', value: string): void
  (event: 'update:granularity', value: Granularity): void
  (event: 'dateRangeChange', range: DateRangeChange): void
  (event: 'granularityChange'): void
  (event: 'refresh'): void
}>()

const { t } = useI18n()

const granularityOptions = computed(() => [
  { value: 'hour', label: t('dashboard.oneHour') },
  { value: '2h', label: t('dashboard.twoHours') },
  { value: '4h', label: t('dashboard.fourHours') },
  { value: '8h', label: t('dashboard.eightHours') },
  { value: 'day', label: t('dashboard.day') },
])

const updateGranularity = (value: string | number | boolean | null) => {
  if (value !== 'day' && value !== 'hour' && value !== '2h' && value !== '4h' && value !== '8h') return
  emit('update:granularity', value)
  emit('granularityChange')
}
</script>
