<template>
  <div class="space-y-4">
    <div class="grid gap-3 md:grid-cols-2">
      <div class="rounded-lg border border-gray-200 p-3 dark:border-dark-700">
        <p class="text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
          {{ t('quickStart.baseUrl') }}
        </p>
        <div class="mt-2 flex items-center gap-2">
          <code class="min-w-0 flex-1 truncate rounded bg-gray-50 px-2 py-1.5 font-mono text-xs text-gray-900 dark:bg-dark-800 dark:text-dark-100">
            {{ examples.baseUrl }}
          </code>
          <button class="btn btn-sm btn-secondary" type="button" @click="copy('baseUrl', examples.baseUrl)">
            <Icon name="copy" size="sm" />
            {{ copiedKey === 'baseUrl' ? t('common.copied') : t('common.copy') }}
          </button>
        </div>
      </div>

      <div class="rounded-lg border border-gray-200 p-3 dark:border-dark-700">
        <p class="text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
          {{ t('quickStart.model') }}
        </p>
        <div class="mt-2 flex items-center gap-2">
          <code class="min-w-0 flex-1 truncate rounded bg-gray-50 px-2 py-1.5 font-mono text-xs text-gray-900 dark:bg-dark-800 dark:text-dark-100">
            {{ examples.model }}
          </code>
          <router-link class="btn btn-sm btn-secondary" to="/available-channels">
            <Icon name="externalLink" size="sm" />
            {{ t('quickStart.viewModels') }}
          </router-link>
        </div>
      </div>
    </div>

    <div class="rounded-lg border border-gray-200 p-3 dark:border-dark-700">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('quickStart.apiKey') }}
          </h3>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ apiKey ? t('quickStart.oneTimeKeyHint') : t('quickStart.placeholderKeyHint') }}
          </p>
        </div>
        <router-link class="btn btn-sm btn-primary" to="/keys">
          <Icon name="key" size="sm" />
          {{ t('quickStart.createKey') }}
        </router-link>
      </div>
      <code class="mt-3 block truncate rounded bg-gray-50 px-3 py-2 font-mono text-xs text-gray-900 dark:bg-dark-800 dark:text-dark-100">
        {{ examples.apiKey }}
      </code>
    </div>

    <div class="space-y-3">
      <section
        v-for="block in codeBlocks"
        :key="block.id"
        class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700"
      >
        <div class="flex items-center justify-between gap-3 border-b border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex min-w-0 items-center gap-2">
            <Icon :name="block.icon" size="sm" class="flex-shrink-0 text-gray-500 dark:text-dark-300" />
            <span class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ block.title }}</span>
          </div>
          <button class="btn btn-sm btn-secondary" type="button" @click="copy(block.id, block.content)">
            <Icon name="copy" size="sm" />
            {{ copiedKey === block.id ? t('common.copied') : t('common.copy') }}
          </button>
        </div>
        <pre class="max-h-80 overflow-auto bg-gray-950 p-4 text-xs text-gray-100"><code>{{ block.content }}</code></pre>
      </section>
    </div>

    <div class="rounded-lg border border-gray-200 p-3 dark:border-dark-700">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('quickStart.commonErrors') }}
          </h3>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ t('quickStart.commonErrorsHint') }}
          </p>
        </div>
        <router-link class="btn btn-sm btn-secondary" to="/usage?tab=errors">
          <Icon name="externalLink" size="sm" />
          {{ t('quickStart.viewErrors') }}
        </router-link>
      </div>
      <div class="mt-3 grid gap-2 text-xs sm:grid-cols-2">
        <div v-for="item in commonErrors" :key="item.code" class="rounded bg-gray-50 px-3 py-2 dark:bg-dark-800">
          <span class="font-mono font-semibold text-gray-900 dark:text-white">{{ item.code }}</span>
          <span class="ml-2 text-gray-600 dark:text-dark-300">{{ item.text }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { buildQuickStartExamples } from '@/utils/quickStart'

const props = withDefaults(defineProps<{
  baseUrl?: string | null
  model?: string | null
  apiKey?: string | null
}>(), {
  baseUrl: '',
  model: '',
  apiKey: ''
})

const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const copiedKey = ref<string | null>(null)

const examples = computed(() => buildQuickStartExamples({
  baseUrl: props.baseUrl,
  model: props.model,
  apiKey: props.apiKey
}))

const codeBlocks = computed(() => [
  { id: 'curl', title: t('quickStart.curlExample'), icon: 'terminal' as const, content: examples.value.curl },
  { id: 'openaiSdk', title: t('quickStart.openaiSdkExample'), icon: 'document' as const, content: examples.value.openaiSdk },
  { id: 'codex', title: t('quickStart.codexExample'), icon: 'book' as const, content: examples.value.codex },
])

const commonErrors = computed(() => [
  { code: '403', text: t('quickStart.errors.403') },
  { code: '429', text: t('quickStart.errors.429') },
  { code: '502', text: t('quickStart.errors.502') },
  { code: '503', text: t('quickStart.errors.503') },
  { code: 'stream disconnected', text: t('quickStart.errors.streamDisconnected') },
])

async function copy(key: string, content: string) {
  const ok = await copyToClipboard(content)
  if (!ok) return
  copiedKey.value = key
  window.setTimeout(() => {
    if (copiedKey.value === key) copiedKey.value = null
  }, 1600)
}
</script>
