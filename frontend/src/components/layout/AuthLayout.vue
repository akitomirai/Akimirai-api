<template>
  <div
    class="auth-shell relative flex min-h-screen items-center justify-center overflow-hidden bg-slate-50 px-4 py-8 text-slate-950 dark:bg-dark-950 dark:text-white sm:px-6 lg:px-8"
  >
    <div class="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-slate-300 to-transparent dark:via-white/15"></div>

    <div class="relative z-10 grid w-full max-w-[1680px] items-center gap-10 lg:min-h-[720px] lg:grid-cols-[minmax(0,1fr)_minmax(560px,620px)] xl:gap-12">
      <section class="hidden min-h-[680px] lg:block" aria-hidden="true"></section>

      <section class="auth-panel-stack mx-auto flex w-full max-w-[620px] flex-col lg:mx-0">
        <div v-if="settingsLoaded" class="auth-reveal auth-reveal-brand mb-5 hidden items-center gap-3 lg:flex">
          <div
            class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/10"
          >
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <div>
            <p class="text-sm font-semibold text-slate-950 dark:text-white">{{ authBrandName }}</p>
            <p class="text-xs text-slate-500 dark:text-dark-300">{{ siteSubtitle }}</p>
          </div>
        </div>

        <div class="auth-reveal auth-reveal-copy auth-hero-copy hidden lg:block">
          <p class="mb-3 text-xs font-semibold uppercase text-primary-700 dark:text-primary-300">
            {{ t('auth.layoutHeroKicker') }}
          </p>
          <h1 class="auth-hero-title max-w-none whitespace-nowrap text-[2rem] font-black leading-[1.08] text-slate-950 dark:text-white xl:text-[2.15rem] 2xl:text-[2.25rem]">
            {{ t('auth.layoutHeroTitle') }}
          </h1>
          <p class="mt-2 text-xl font-bold text-slate-900 dark:text-white/90 xl:text-2xl">
            {{ t('auth.layoutHeroSubtitle') }}
          </p>
          <p class="mt-2 max-w-xl text-sm leading-7 text-slate-600 dark:text-dark-200 xl:text-base">
            {{ t('auth.layoutHeroDescription') }}
          </p>
        </div>

        <div class="auth-reveal auth-reveal-mobile mb-6 text-center lg:hidden">
          <template v-if="settingsLoaded">
            <div
              class="mb-3 inline-flex h-14 w-14 items-center justify-center overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/10"
            >
              <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
            </div>
            <h1 class="text-3xl font-black text-slate-950 dark:text-white">{{ authBrandName }}</h1>
            <p class="mt-2 text-sm text-slate-500 dark:text-dark-300">{{ siteSubtitle }}</p>
          </template>
        </div>

        <div class="auth-reveal auth-reveal-card auth-card mt-8 w-full self-start rounded-lg border border-slate-200/90 bg-white/85 p-6 shadow-[0_24px_80px_rgba(15,23,42,0.10)] backdrop-blur-xl dark:border-white/10 dark:bg-dark-900/80 sm:p-8 lg:max-w-[430px]">
          <slot />
        </div>

        <div class="auth-reveal auth-reveal-footer mt-6 w-full self-start text-center text-sm lg:max-w-[430px]">
          <slot name="footer" />
        </div>

        <div class="auth-reveal auth-reveal-metrics auth-metrics mt-6 grid w-full self-start grid-cols-3 overflow-hidden rounded-lg border border-slate-200/80 bg-white/70 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5 lg:max-w-[620px]">
          <div
            v-for="metric in heroMetrics"
            :key="metric.label"
            class="border-r border-slate-200/80 px-5 py-4 last:border-r-0 dark:border-white/10"
          >
            <p class="auth-metric-value font-black tabular-nums text-slate-950 dark:text-white">{{ metric.value }}</p>
            <p class="mt-1 text-xs font-medium text-slate-500 dark:text-dark-300">{{ metric.label }}</p>
          </div>
        </div>

        <div class="auth-reveal auth-reveal-page-footer mt-6 w-full self-start text-center text-xs text-slate-400 dark:text-dark-500 lg:max-w-[430px]">
          {{ t('auth.layoutFooter') }}
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

const authBrandName = computed(() => t('auth.layoutBrandName'))
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => t('auth.layoutHeroSubtitle'))
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const configuredAIPlatforms = computed(() => {
  const rawPlatforms = appStore.cachedPublicSettings?.configured_ai_platforms ?? []
  const seen = new Set<string>()
  return rawPlatforms
    .map((platform) => platform.trim())
    .filter((platform) => {
      if (!platform || seen.has(platform)) {
        return false
      }
      seen.add(platform)
      return true
    })
})

const platformMetricValue = computed(() => {
  const platforms = configuredAIPlatforms.value
  if (platforms.length === 0) {
    return t('auth.layoutHeroMetricOneValue')
  }
  const visiblePlatforms = platforms.slice(0, 3)
  const hiddenCount = platforms.length - visiblePlatforms.length
  return `${visiblePlatforms.join(' / ')}${hiddenCount > 0 ? ` +${hiddenCount}` : ''}`
})

const heroMetrics = computed(() => [
  { value: platformMetricValue.value, label: t('auth.layoutHeroMetricOneLabel') },
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

.auth-shell::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image:
    linear-gradient(90deg, rgba(248, 250, 252, 0.01) 0%, rgba(248, 250, 252, 0.04) 43%, rgba(248, 250, 252, 0.38) 65%, rgba(248, 250, 252, 0.66) 100%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.04), rgba(248, 250, 252, 0.3)),
    url('@/assets/auth-shrine-pixel.webp');
  background-position:
    center,
    center,
    44% 54%;
  background-repeat: no-repeat;
  background-size:
    100% 100%,
    100% 100%,
    cover;
}

:global(.dark .auth-shell) {
  background:
    linear-gradient(180deg, rgba(2, 6, 23, 0.98), rgba(15, 23, 42, 0.98)),
    #020617;
}

:global(.dark .auth-shell::before) {
  opacity: 0.35;
  filter: saturate(0.8) brightness(0.55);
}

.auth-hero-copy {
  max-width: 560px;
}

.auth-hero-title {
  text-shadow:
    0 2px 18px rgba(255, 255, 255, 0.94),
    0 0 2px rgba(255, 255, 255, 0.9);
}

.auth-metric-value {
  font-size: clamp(1rem, 1.3vw, 1.5rem);
  line-height: 1.1;
  white-space: nowrap;
}

:global(.dark .auth-hero-title) {
  text-shadow:
    0 2px 18px rgba(2, 6, 23, 0.65),
    0 0 2px rgba(2, 6, 23, 0.72);
}

.auth-reveal {
  animation: auth-reveal-rise 900ms cubic-bezier(0.22, 1, 0.36, 1) both;
  transform-origin: 50% 100%;
  will-change: transform, opacity;
}

.auth-reveal-brand,
.auth-reveal-mobile {
  animation-delay: 120ms;
}

.auth-reveal-copy {
  animation-delay: 420ms;
}

.auth-reveal-card {
  animation-delay: 820ms;
}

.auth-reveal-footer {
  animation-delay: 1140ms;
}

.auth-reveal-metrics {
  animation-delay: 1460ms;
}

.auth-reveal-page-footer {
  animation-delay: 1700ms;
}

@keyframes auth-reveal-rise {
  0% {
    opacity: 0;
    transform: translate3d(0, 72px, 0);
  }

  100% {
    opacity: 1;
    transform: translate3d(0, 0, 0);
  }
}

.auth-card :deep(.input) {
  border-radius: 0.5rem;
  background-color: rgba(248, 250, 252, 0.78);
}

:global(.dark .auth-card .input) {
  background-color: rgba(15, 23, 42, 0.72);
}

.auth-card :deep(.btn) {
  border-radius: 0.5rem;
}

@media (min-width: 1280px) {
  .auth-shell::before {
    background-position:
      center,
      center,
      44% 54%;
  }
}

@media (max-width: 1023px) {
  .auth-shell::before {
    opacity: 0.32;
    background-position:
      center,
      center,
      center top;
    background-size:
      100% 100%,
      100% 100%,
      cover;
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-reveal {
    animation: none;
    will-change: auto;
  }
}
</style>
