<template>
  <BaseDialog
    :show="show"
    :title="t('keys.ccsImport.title')"
    width="narrow"
    @close="emit('close')"
  >
    <form id="cc-switch-import-form" class="space-y-4" @submit.prevent="submit">
      <fieldset>
        <legend class="input-label">{{ t('keys.ccsImport.app') }}</legend>
        <div class="grid grid-cols-3 gap-1 rounded-lg bg-gray-100 p-1 dark:bg-dark-700">
          <button
            v-for="option in appOptions"
            :key="option.value"
            type="button"
            :class="[
              'min-h-9 rounded-md px-2 text-sm font-medium transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500/40',
              form.app === option.value
                ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-600 dark:text-white'
                : 'text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
            ]"
            :aria-pressed="form.app === option.value"
            @click="form.app = option.value"
          >
            {{ option.label }}
          </button>
        </div>
      </fieldset>

      <div>
        <label for="cc-switch-name" class="input-label">{{ t('keys.ccsImport.name') }}</label>
        <input
          id="cc-switch-name"
          v-model="form.name"
          type="text"
          required
          class="input"
          :placeholder="t('keys.ccsImport.namePlaceholder')"
        />
      </div>

      <div>
        <label for="cc-switch-main-model" class="input-label">
          {{ t('keys.ccsImport.mainModel') }} <span class="text-red-500">*</span>
        </label>
        <input
          id="cc-switch-main-model"
          v-model="form.model"
          type="text"
          required
          class="input"
          :list="modelListId"
          :placeholder="t('keys.ccsImport.modelPlaceholder')"
          data-test="cc-switch-main-model"
        />
        <p v-if="catalogLoading" class="input-hint">{{ t('keys.modelRestriction.loading') }}</p>
        <p v-else-if="catalogError" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
          {{ t('keys.ccsImport.catalogFallback') }}
        </p>
      </div>

      <template v-if="form.app === 'claude'">
        <div v-for="field in claudeModelFields" :key="field.key">
          <label :for="`cc-switch-${field.key}`" class="input-label">{{ field.label }}</label>
          <input
            :id="`cc-switch-${field.key}`"
            v-model="form[field.key]"
            type="text"
            class="input"
            :list="modelListId"
            :placeholder="t('keys.ccsImport.modelPlaceholder')"
          />
        </div>
      </template>

      <datalist :id="modelListId">
        <option v-for="model in modelOptions" :key="model" :value="model" />
      </datalist>

      <p v-if="validationError" class="text-sm text-red-600 dark:text-red-400" role="alert">
        {{ validationError }}
      </p>
    </form>

    <template #footer>
      <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <label
          v-if="form.app === 'codex'"
          class="flex min-h-10 min-w-0 flex-1 cursor-pointer items-center gap-2 text-sm text-gray-700 dark:text-gray-200"
          :title="t('keys.ccsImport.remoteCompactionHint')"
        >
          <input
            v-model="form.remoteCompaction"
            type="checkbox"
            class="h-4 w-4 shrink-0 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-dark-500 dark:bg-dark-700"
            data-test="cc-switch-remote-compaction"
          />
          <span class="min-w-0">
            <span class="font-medium">{{ t('keys.ccsImport.remoteCompaction') }}</span>
            <span class="block text-xs text-gray-500 dark:text-gray-400">
              {{ t('keys.ccsImport.remoteCompactionHint') }}
            </span>
          </span>
        </label>
        <span v-else aria-hidden="true" />

        <div class="flex shrink-0 justify-end gap-2">
          <button type="button" class="btn btn-secondary whitespace-nowrap" @click="emit('close')">
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            form="cc-switch-import-form"
            class="btn btn-primary whitespace-nowrap"
          >
            {{ t('keys.ccsImport.open') }}
          </button>
        </div>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import {
  GROK_CC_SWITCH_MODEL,
  OPENAI_CC_SWITCH_CODEX_MODEL,
  type CcSwitchApp,
  type CcSwitchImportFormData
} from '@/utils/ccswitchImport'
import type { GroupPlatform } from '@/types'

const props = withDefaults(defineProps<{
  show: boolean
  platform?: GroupPlatform | null
  providerName: string
  modelOptions: string[]
  catalogLoading?: boolean
  catalogError?: string
}>(), {
  platform: null,
  catalogLoading: false,
  catalogError: ''
})

const emit = defineEmits<{
  close: []
  confirm: [value: CcSwitchImportFormData]
}>()

const { t } = useI18n()
const modelListId = `cc-switch-models-${Math.random().toString(36).slice(2, 9)}`
const validationError = ref('')
const form = reactive<Required<CcSwitchImportFormData>>({
  app: 'claude',
  name: '',
  model: '',
  remoteCompaction: false,
  haikuModel: '',
  sonnetModel: '',
  opusModel: ''
})

const appOptions = computed<Array<{ value: CcSwitchApp; label: string }>>(() => [
  { value: 'claude', label: 'Claude' },
  { value: 'codex', label: 'Codex' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'grokbuild', label: 'Grok Build' }
])

const claudeModelFields = computed<Array<{
  key: 'haikuModel' | 'sonnetModel' | 'opusModel'
  label: string
}>>(() => [
  { key: 'haikuModel', label: t('keys.ccsImport.haikuModel') },
  { key: 'sonnetModel', label: t('keys.ccsImport.sonnetModel') },
  { key: 'opusModel', label: t('keys.ccsImport.opusModel') }
])

const inferApp = (): CcSwitchApp => {
  if (props.platform === 'openai') return 'codex'
  if (props.platform === 'grok') return 'grokbuild'
  if (props.platform === 'gemini') return 'gemini'
  return 'claude'
}

const preferredModel = (app: CcSwitchApp): string => {
  if (app === 'codex') {
    return props.modelOptions.find((model) => model === OPENAI_CC_SWITCH_CODEX_MODEL)
      || props.modelOptions.find((model) => model.startsWith('gpt-'))
      || ''
  }
  if (app === 'grokbuild') {
    return props.modelOptions.find((model) => model === GROK_CC_SWITCH_MODEL)
      || props.modelOptions.find((model) => model.toLowerCase().startsWith('grok-'))
      || ''
  }
  const prefix = app === 'gemini' ? 'gemini' : 'claude'
  return props.modelOptions.find((model) => model.toLowerCase().startsWith(prefix)) || ''
}

const reset = () => {
  form.app = inferApp()
  form.name = props.providerName
  form.model = preferredModel(form.app)
  form.remoteCompaction = form.app === 'codex'
  form.haikuModel = ''
  form.sonnetModel = ''
  form.opusModel = ''
  validationError.value = ''
}

watch(() => props.show, (show) => {
  if (show) reset()
}, { immediate: true })

watch(() => form.app, (app, previousApp) => {
  const previousDefault = preferredModel(previousApp)
  if (!form.model || form.model === previousDefault) form.model = preferredModel(app)
  form.remoteCompaction = app === 'codex'
  validationError.value = ''
})

const submit = () => {
  const name = form.name.trim()
  const model = form.model.trim()
  if (!name) {
    validationError.value = t('keys.ccsImport.nameRequired')
    return
  }
  if (!model) {
    validationError.value = t('keys.ccsImport.modelRequired')
    return
  }

  emit('confirm', {
    app: form.app,
    name,
    model,
    remoteCompaction: form.app === 'codex' && form.remoteCompaction,
    haikuModel: form.app === 'claude' ? form.haikuModel.trim() || undefined : undefined,
    sonnetModel: form.app === 'claude' ? form.sonnetModel.trim() || undefined : undefined,
    opusModel: form.app === 'claude' ? form.opusModel.trim() || undefined : undefined
  })
}
</script>
