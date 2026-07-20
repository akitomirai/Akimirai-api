<template>
  <div ref="rootRef" class="space-y-2">
    <button
      ref="triggerRef"
      type="button"
      class="flex min-h-11 w-full items-center gap-2 rounded-xl border border-gray-200 bg-white px-3 py-2 text-left text-sm transition-colors focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/30 disabled:cursor-not-allowed disabled:bg-gray-100 disabled:opacity-70 dark:border-dark-600 dark:bg-dark-800 dark:disabled:bg-dark-900"
      :disabled="disabled || loading || Boolean(error)"
      :aria-expanded="open"
      aria-haspopup="listbox"
      data-test="model-multi-select-trigger"
      @click="open = !open"
    >
      <div v-if="modelValue.length" class="flex min-w-0 flex-1 flex-wrap gap-1.5">
        <span
          v-for="model in modelValue"
          :key="model"
          class="inline-flex max-w-full items-center gap-1 rounded-md bg-gray-100 px-2 py-1 text-xs text-gray-700 dark:bg-dark-700 dark:text-gray-200"
        >
          <span class="truncate">{{ model }}</span>
          <button
            type="button"
            class="shrink-0 rounded p-0.5 text-gray-400 hover:bg-gray-200 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-white"
            :aria-label="t('keys.modelRestriction.remove', { model })"
            @click.stop="remove(model)"
          >
            <Icon name="x" size="xs" />
          </button>
        </span>
      </div>
      <span v-else class="min-w-0 flex-1 truncate text-gray-400 dark:text-dark-400">
        {{ placeholder }}
      </span>
      <Icon
        name="chevronDown"
        size="md"
        :class="['shrink-0 text-gray-400 transition-transform', open && 'rotate-180']"
      />
    </button>

    <p v-if="loading" class="text-xs text-gray-500 dark:text-gray-400">
      {{ t('keys.modelRestriction.loading') }}
    </p>
    <p v-else-if="error" class="text-xs text-red-600 dark:text-red-400">
      {{ error }}
    </p>

    <Teleport to="body">
      <div
        v-if="open && !loading && !error"
        ref="dropdownRef"
        class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-lg shadow-black/10 dark:border-dark-600 dark:bg-dark-800 dark:shadow-black/30"
        :style="dropdownStyle"
        data-test="model-multi-select-dropdown"
      >
      <div class="flex items-center gap-2 border-b border-gray-100 px-3 py-2 dark:border-dark-700">
        <Icon name="search" size="sm" class="shrink-0 text-gray-400" />
        <input
          ref="searchRef"
          v-model="query"
          type="search"
          class="min-w-0 flex-1 bg-transparent text-sm text-gray-900 outline-none placeholder:text-gray-400 dark:text-white"
          :placeholder="t('keys.modelRestriction.search')"
          data-test="model-multi-select-search"
        />
      </div>
      <div class="max-h-56 overflow-y-auto py-1" role="listbox" aria-multiselectable="true">
        <button
          v-for="model in filteredOptions"
          :key="model"
          type="button"
          role="option"
          :aria-selected="modelValue.includes(model)"
          class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-gray-700 transition-colors hover:bg-gray-50 dark:text-gray-200 dark:hover:bg-dark-700"
          @click="toggle(model)"
        >
          <span
            :class="[
              'flex h-4 w-4 shrink-0 items-center justify-center rounded border',
              modelValue.includes(model)
                ? 'border-primary-500 bg-primary-500 text-white'
                : 'border-gray-300 dark:border-dark-500'
            ]"
          >
            <Icon v-if="modelValue.includes(model)" name="check" size="xs" />
          </span>
          <span class="min-w-0 flex-1 truncate">{{ model }}</span>
        </button>
        <div v-if="filteredOptions.length === 0" class="px-4 py-6 text-center text-sm text-gray-500 dark:text-gray-400">
          {{ options.length ? t('keys.modelRestriction.noMatch') : t('keys.modelRestriction.empty') }}
        </div>
      </div>
      <div v-if="modelValue.length" class="border-t border-gray-100 px-3 py-2 text-right dark:border-dark-700">
        <button type="button" class="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400" @click="clear">
          {{ t('keys.modelRestriction.clear') }}
        </button>
      </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(defineProps<{
  modelValue: string[]
  options: string[]
  disabled?: boolean
  loading?: boolean
  error?: string
  placeholder?: string
}>(), {
  disabled: false,
  loading: false,
  error: '',
  placeholder: ''
})

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const { t } = useI18n()
const rootRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const searchRef = ref<HTMLInputElement | null>(null)
const open = ref(false)
const query = ref('')
const triggerRect = ref<DOMRect | null>(null)
const dropdownPosition = ref<'bottom' | 'top'>('bottom')

const dropdownStyle = computed(() => {
  const rect = triggerRect.value
  if (!rect) return { display: 'none' }

  const viewportPadding = 8
  const width = Math.min(rect.width, window.innerWidth - viewportPadding * 2)
  const left = Math.min(
    Math.max(rect.left, viewportPadding),
    window.innerWidth - width - viewportPadding
  )
  const style: Record<string, string> = {
    position: 'fixed',
    left: `${left}px`,
    width: `${width}px`,
    zIndex: '100000020'
  }

  if (dropdownPosition.value === 'top') {
    style.bottom = `${window.innerHeight - rect.top + 4}px`
  } else {
    style.top = `${rect.bottom + 4}px`
  }
  return style
})

const filteredOptions = computed(() => {
  const normalized = query.value.trim().toLowerCase()
  if (!normalized) return props.options
  return props.options.filter((model) => model.toLowerCase().includes(normalized))
})

watch(open, async (isOpen) => {
  if (!isOpen) {
    query.value = ''
    window.removeEventListener('scroll', updateDropdownPosition, { capture: true })
    window.removeEventListener('resize', updateDropdownPosition)
    return
  }
  await nextTick()
  updateDropdownPosition()
  searchRef.value?.focus()
  window.addEventListener('scroll', updateDropdownPosition, { capture: true, passive: true })
  window.addEventListener('resize', updateDropdownPosition)
})

watch(() => [props.disabled, props.loading, props.error] as const, ([disabled, loading, error]) => {
  if (disabled || loading || error) open.value = false
})

const toggle = (model: string) => {
  emit(
    'update:modelValue',
    props.modelValue.includes(model)
      ? props.modelValue.filter((item) => item !== model)
      : [...props.modelValue, model]
  )
}

const remove = (model: string) => {
  emit('update:modelValue', props.modelValue.filter((item) => item !== model))
}

const clear = () => emit('update:modelValue', [])

const updateDropdownPosition = () => {
  const trigger = triggerRef.value
  if (!trigger) return

  const rect = trigger.getBoundingClientRect()
  triggerRect.value = rect
  const dropdownHeight = dropdownRef.value?.offsetHeight || 280
  const spaceBelow = window.innerHeight - rect.bottom
  const spaceAbove = rect.top
  dropdownPosition.value = spaceBelow < dropdownHeight + 8 && spaceAbove > spaceBelow
    ? 'top'
    : 'bottom'
}

const onDocumentClick = (event: MouseEvent) => {
  const target = event.target as Node
  if (!rootRef.value?.contains(target) && !dropdownRef.value?.contains(target)) open.value = false
}

onMounted(() => document.addEventListener('click', onDocumentClick))
onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick)
  window.removeEventListener('scroll', updateDropdownPosition, { capture: true })
  window.removeEventListener('resize', updateDropdownPosition)
})
</script>
