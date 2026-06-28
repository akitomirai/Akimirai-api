<template>
  <div class="space-y-2">
    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
      {{ t('payment.promoCode') }}
    </label>
    <div class="flex flex-col gap-2 sm:flex-row">
      <div class="relative flex-1">
        <Icon name="badge" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <input
          :value="code"
          type="text"
          class="input w-full pl-9 uppercase"
          :placeholder="t('payment.promoCodePlaceholder')"
          :disabled="disabled || checking"
          @input="handleInput"
          @keydown.enter.prevent="emit('apply')"
        />
      </div>
      <button
        v-if="preview"
        type="button"
        class="btn btn-secondary shrink-0"
        :disabled="disabled || checking"
        @click="emit('clear')"
      >
        {{ t('payment.clearPromoCode') }}
      </button>
      <button
        v-else
        type="button"
        class="btn btn-primary shrink-0"
        :disabled="disabled || checking || code.trim() === ''"
        @click="emit('apply')"
      >
        <span v-if="checking" class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
        {{ checking ? t('common.processing') : t('payment.applyPromoCode') }}
      </button>
    </div>
    <p v-if="preview" class="text-xs font-medium text-emerald-600 dark:text-emerald-300">
      {{ t('payment.promoCodeApplied', { code: preview.code, percent: formatPercent(preview.discount_percent) }) }}
    </p>
    <p v-else-if="error" class="text-xs text-red-600 dark:text-red-400">
      {{ error }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { PaymentPromoCodeQuote } from '@/types/payment'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  code: string
  preview: PaymentPromoCodeQuote | null
  checking: boolean
  error: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:code': [value: string]
  apply: []
  clear: []
}>()

const { t } = useI18n()

function handleInput(event: Event): void {
  emit('update:code', (event.target as HTMLInputElement).value.toUpperCase())
}

function formatPercent(value: number): string {
  return Number(value || 0).toFixed(2).replace(/\.?0+$/, '')
}
</script>
