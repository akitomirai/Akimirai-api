<template>
  <div
    class="auth-shell relative flex min-h-screen items-center justify-center overflow-hidden bg-slate-50 px-4 py-8 text-slate-950 dark:bg-dark-950 dark:text-white sm:px-6 lg:px-8"
  >
    <div class="pointer-events-none absolute inset-0 auth-grid"></div>
    <div class="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-slate-300 to-transparent dark:via-white/15"></div>

    <div class="relative z-10 grid w-full max-w-6xl items-center gap-7 lg:grid-cols-[minmax(0,1.08fr)_minmax(360px,430px)]">
      <section class="hidden min-h-[560px] flex-col justify-between lg:flex">
        <div>
          <div v-if="settingsLoaded" class="mb-16 flex items-center gap-3">
            <div
              class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/10"
            >
              <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
            </div>
            <div>
              <p class="text-sm font-semibold text-slate-950 dark:text-white">{{ siteName }}</p>
              <p class="text-xs text-slate-500 dark:text-dark-300">{{ siteSubtitle }}</p>
            </div>
          </div>

          <p class="mb-5 text-xs font-semibold uppercase text-primary-700 dark:text-primary-300">
            {{ t('auth.layoutHeroKicker') }}
          </p>
          <h1 class="max-w-3xl text-5xl font-black leading-[1.08] text-slate-950 dark:text-white xl:text-6xl">
            {{ t('auth.layoutHeroTitle') }}
          </h1>
          <p class="mt-6 max-w-xl text-base leading-8 text-slate-600 dark:text-dark-200">
            {{ t('auth.layoutHeroDescription') }}
          </p>
        </div>

        <div class="grid max-w-3xl grid-cols-3 overflow-hidden rounded-lg border border-slate-200/80 bg-white/70 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5">
          <div
            v-for="metric in heroMetrics"
            :key="metric.label"
            class="border-r border-slate-200/80 px-5 py-4 last:border-r-0 dark:border-white/10"
          >
            <p class="text-2xl font-black tabular-nums text-slate-950 dark:text-white">{{ metric.value }}</p>
            <p class="mt-1 text-xs font-medium text-slate-500 dark:text-dark-300">{{ metric.label }}</p>
          </div>
        </div>
      </section>

      <section class="mx-auto w-full max-w-md lg:mx-0">
        <div class="mb-6 text-center lg:hidden">
          <template v-if="settingsLoaded">
            <div
            class="mb-3 inline-flex h-14 w-14 items-center justify-center overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/10"
            >
              <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
            </div>
            <h1 class="text-3xl font-black text-slate-950 dark:text-white">{{ siteName }}</h1>
            <p class="mt-2 text-sm text-slate-500 dark:text-dark-300">{{ siteSubtitle }}</p>
          </template>
        </div>

        <div class="auth-card rounded-lg border border-slate-200/90 bg-white/85 p-6 shadow-[0_24px_80px_rgba(15,23,42,0.10)] backdrop-blur-xl dark:border-white/10 dark:bg-dark-900/80 sm:p-8">
          <slot />
        </div>

        <div class="mt-6 text-center text-sm">
          <slot name="footer" />
        </div>

        <div class="mt-8 text-center text-xs text-slate-400 dark:text-dark-500">
          &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const { t } = useI18n()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

const heroMetrics = computed(() => [
  { value: t('auth.layoutHeroMetricOneValue'), label: t('auth.layoutHeroMetricOneLabel') },
  { value: t('auth.layoutHeroMetricTwoValue'), label: t('auth.layoutHeroMetricTwoLabel') },
  { value: t('auth.layoutHeroMetricThreeValue'), label: t('auth.layoutHeroMetricThreeLabel') }
])

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.auth-shell {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.98)),
    #f8fafc;
}

:global(.dark) .auth-shell {
  background:
    linear-gradient(180deg, rgba(2, 6, 23, 0.98), rgba(15, 23, 42, 0.98)),
    #020617;
}

.auth-grid {
  background-image:
    linear-gradient(rgba(15, 23, 42, 0.055) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.055) 1px, transparent 1px),
    linear-gradient(rgba(15, 23, 42, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.035) 1px, transparent 1px);
  background-size:
    96px 96px,
    96px 96px,
    24px 24px,
    24px 24px;
  mask-image: linear-gradient(to bottom, black 0%, black 72%, transparent 100%);
  -webkit-mask-image: linear-gradient(to bottom, black 0%, black 72%, transparent 100%);
}

:global(.dark) .auth-grid {
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.08) 1px, transparent 1px),
    linear-gradient(rgba(255, 255, 255, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.04) 1px, transparent 1px);
}

.auth-card :deep(.input) {
  border-radius: 0.5rem;
  background-color: rgba(248, 250, 252, 0.78);
}

:global(.dark) .auth-card :deep(.input) {
  background-color: rgba(15, 23, 42, 0.72);
}

.auth-card :deep(.btn) {
  border-radius: 0.5rem;
}
</style>
