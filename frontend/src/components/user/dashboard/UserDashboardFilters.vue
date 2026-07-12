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
          @update:start-date="emit('update:startDate', $event)"
          @update:end-date="emit('update:endDate', $event)"
          @change="emit('dateRangeChange')"
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

type Granularity = 'day' | 'hour'

defineProps<{
  startDate: string
  endDate: string
  granularity: Granularity
  loading: boolean
}>()

const emit = defineEmits<{
  (event: 'update:startDate', value: string): void
  (event: 'update:endDate', value: string): void
  (event: 'update:granularity', value: Granularity): void
  (event: 'dateRangeChange'): void
  (event: 'granularityChange'): void
  (event: 'refresh'): void
}>()

const { t } = useI18n()

const granularityOptions = computed(() => [
  { value: 'day', label: t('dashboard.day') },
  { value: 'hour', label: t('dashboard.hour') },
])

const updateGranularity = (value: string | number | boolean | null) => {
  if (value !== 'day' && value !== 'hour') return
  emit('update:granularity', value)
  emit('granularityChange')
}
</script>
