<template>
  <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800/50">
    <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-[minmax(18rem,1fr)_repeat(5,minmax(8rem,0.55fr))_auto]">
      <label class="relative block sm:col-span-2 xl:col-span-1">
        <span class="sr-only">{{ text('searchPlaceholder') }}</span>
        <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <input
          :value="modelValue.query"
          class="input h-10 pl-9"
          type="search"
          :placeholder="text('searchPlaceholder')"
          @input="update('query', ($event.target as HTMLInputElement).value)"
        />
      </label>

      <select :value="modelValue.group" class="input h-10 py-1.5" @change="update('group', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ text('allGroups') }}</option>
        <option v-for="group in options.groups" :key="group" :value="group">{{ group }}</option>
      </select>
      <select :value="modelValue.provider" class="input h-10 py-1.5" @change="update('provider', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ text('allProviders') }}</option>
        <option v-for="provider in options.providers" :key="provider" :value="provider">{{ provider }}</option>
      </select>
      <select :value="modelValue.billingMode" class="input h-10 py-1.5" @change="update('billingMode', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ text('allBillingModes') }}</option>
        <option v-for="mode in options.billingModes" :key="mode" :value="mode">{{ billingModeLabel(mode) }}</option>
      </select>
      <select :value="modelValue.status" class="input h-10 py-1.5" @change="update('status', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ text('allStatuses') }}</option>
        <option v-for="status in options.statuses" :key="status" :value="status">{{ text(status) }}</option>
      </select>
      <select :value="modelValue.capability" class="input h-10 py-1.5" @change="update('capability', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ text('allCapabilities') }}</option>
        <option v-for="capability in options.capabilities" :key="capability" :value="capability">{{ text(capability) }}</option>
      </select>
      <button class="btn btn-sm btn-secondary h-10 px-3" type="button" :disabled="!hasFilters" @click="emit('reset')">
        <Icon name="refresh" size="sm" />
        {{ text('reset') }}
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import type { ModelCatalogFilterOptions, ModelCatalogFilters } from '@/utils/modelCatalog'
import type { BillingMode } from '@/constants/channel'
import { useModelPlazaText } from './modelPlazaText'

const props = defineProps<{
  modelValue: ModelCatalogFilters
  options: ModelCatalogFilterOptions
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: ModelCatalogFilters): void
  (event: 'reset'): void
}>()

const { text } = useModelPlazaText()

const hasFilters = computed(() => Object.values(props.modelValue).some((value) => Boolean(value)))

function update(key: keyof ModelCatalogFilters, value: string): void {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}

function billingModeLabel(mode: BillingMode): string {
  if (mode === 'token') return text('tokenBilling')
  if (mode === 'per_request') return text('requestBilling')
  return text('imageBilling')
}
</script>
