<template>
  <header class="glass sticky top-0 z-30 border-b border-gray-200/50 dark:border-dark-700/50">
    <div class="flex h-16 items-center justify-between gap-2 px-2 sm:px-4 md:px-6">
      <!-- Left: Mobile Menu Toggle + Page Title -->
      <div class="flex min-w-0 flex-1 items-center gap-2 sm:gap-4">
        <button
          @click="toggleMobileSidebar"
          class="btn-ghost btn-icon lg:hidden"
          :aria-label="t('common.toggleMenu')"
        >
          <Icon name="menu" size="md" />
        </button>

        <div class="hidden min-w-0 items-baseline gap-3 lg:flex">
          <slot name="title">
            <h1 class="shrink-0 whitespace-nowrap text-lg font-semibold text-gray-950 dark:text-white">
              {{ pageTitle }}
            </h1>
          </slot>
          <button
            type="button"
            class="header-quote"
            :class="{ 'header-quote-loading': quoteLoading }"
            :title="quoteTitle"
            aria-label="Refresh quote"
            @click="loadRandomQuote"
          >
            <span class="header-quote-text">{{ quoteText }}</span>
          </button>
        </div>
      </div>

      <!-- Right: Announcements + Docs + Language + Subscriptions + Balance + User Dropdown -->
      <div class="flex min-w-0 shrink-0 items-center gap-1 sm:gap-3">
        <button
          v-if="user"
          data-testid="daily-check-in-desktop"
          type="button"
          class="hidden xl:inline-flex h-9 min-w-[7.25rem] items-center justify-center gap-1.5 rounded-lg px-2.5 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 disabled:cursor-default dark:focus-visible:ring-offset-dark-900"
          :class="checkInButtonClass"
          :disabled="checkInDisabled"
          :aria-busy="checkInBusy"
          :title="checkInTitle"
          aria-live="polite"
          @click="handleCheckInAction"
        >
          <Icon
            :name="checkInIcon"
            size="sm"
            :class="{ 'animate-spin': checkInBusy }"
          />
          <span class="whitespace-nowrap">{{ checkInLabel }}</span>
        </button>

        <button
          v-if="user && modelPlazaEnabled"
          type="button"
          class="flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-sm font-medium transition-colors"
          :class="isModelPlazaActive
            ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
            : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white'"
          :aria-pressed="isModelPlazaActive"
          :title="isModelPlazaActive ? t('common.close', 'Close') : t('nav.modelPlaza')"
          @click="handleModelPlazaClick"
        >
          <Icon :name="isModelPlazaActive ? 'x' : 'grid'" size="sm" />
          <span class="hidden sm:inline">{{ t('nav.modelPlaza') }}</span>
        </button>

        <button
          v-if="user"
          type="button"
          class="flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-sm font-medium transition-colors"
          :class="isGptImageActive
            ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
            : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white'"
          :aria-pressed="isGptImageActive"
          :title="isGptImageActive ? '关闭 GPTImage' : 'GPTImage'"
          @click="handleGptImageClick"
        >
          <Icon :name="isGptImageActive ? 'x' : 'sparkles'" size="sm" />
          <span class="hidden sm:inline">GPTImage</span>
        </button>
        <!-- Announcement Bell -->
        <AnnouncementBell v-if="user" />

        <!-- Docs Link -->
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white sm:flex"
        >
          <Icon name="book" size="sm" />
          <span class="hidden sm:inline">{{ t('nav.docs') }}</span>
        </a>

        <!-- Language Switcher -->
        <LocaleSwitcher />

        <!-- Subscription Progress (for users with active subscriptions) -->
        <SubscriptionProgressMini v-if="user" />

        <!-- Balance Display -->
        <div
          v-if="user"
class="group relative hidden items-center gap-2 rounded-xl bg-primary-50 px-3 py-1.5 dark:bg-primary-900/20 sm:flex"
          :title="`${t('common.balance')}: ${formatHeaderMoney(availableBalance)}`"
        >
          <span class="balance-chip-icon">
            <Icon name="dollar" size="sm" :stroke-width="2" />
          </span>
          <span class="text-sm font-semibold text-primary-700 dark:text-primary-300">
            {{ formatHeaderMoney(availableBalance) }}
          </span>
          <span
            v-if="frozenBalance > 0"
            class="rounded-full bg-amber-100 px-1.5 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-900/40 dark:text-amber-200"
          >
            {{ balanceFrozenLabel }}
          </span>
          <div
            class="pointer-events-none absolute right-0 top-full mt-2 hidden w-56 rounded-lg border border-gray-200 bg-white p-3 text-xs shadow-lg group-hover:block dark:border-dark-700 dark:bg-dark-800"
          >
            <div class="flex items-center justify-between">
              <span class="text-gray-500 dark:text-dark-400">{{ balanceAvailableText }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ formatHeaderMoney(availableBalance) }}</span>
            </div>
            <div class="mt-2 flex items-center justify-between">
              <span class="text-gray-500 dark:text-dark-400">{{ balanceFrozenText }}</span>
              <span class="font-medium text-amber-700 dark:text-amber-200">{{ formatHeaderMoney(frozenBalance) }}</span>
            </div>
            <div class="mt-2 border-t border-gray-100 pt-2 dark:border-dark-700">
              <div class="flex items-center justify-between">
                <span class="text-gray-500 dark:text-dark-400">{{ balanceTotalText }}</span>
                <span class="font-semibold text-gray-900 dark:text-white">{{ formatHeaderMoney(totalBalance) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- User Dropdown -->
        <div v-if="user" class="relative" ref="dropdownRef">
          <button
            @click="toggleDropdown"
            class="flex items-center gap-2 rounded-xl p-1.5 transition-colors hover:bg-gray-100 dark:hover:bg-dark-800"
            :aria-label="t('common.userMenu')"
          >
            <div class="flex h-8 w-8 items-center justify-center overflow-hidden rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 text-sm font-medium text-white shadow-sm">
              <img
                v-if="avatarUrl"
                :src="avatarUrl"
                :alt="displayName"
                class="h-full w-full object-cover"
              >
              <span v-else>{{ userInitials }}</span>
            </div>
            <div class="hidden text-left md:block">
              <div class="text-sm font-medium text-gray-900 dark:text-white">
                {{ displayName }}
              </div>
              <div class="text-xs capitalize text-gray-500 dark:text-dark-400">
                {{ user.role }}
              </div>
            </div>
            <Icon name="chevronDown" size="sm" class="hidden text-gray-400 md:block" />
          </button>

          <!-- Dropdown Menu -->
          <transition name="dropdown">
            <div v-if="dropdownOpen" class="dropdown right-0 mt-2 w-56">
              <!-- User Info -->
              <div class="border-b border-gray-100 px-4 py-3 dark:border-dark-700">
                <div class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ displayName }}
                </div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ user.email }}</div>
              </div>

              <!-- Balance (mobile only) -->
              <div class="border-b border-gray-100 px-4 py-2 dark:border-dark-700 sm:hidden">
                <div class="text-xs text-gray-500 dark:text-dark-400">
                  {{ t('common.balance') }}
                </div>
                <div class="text-sm font-semibold text-primary-600 dark:text-primary-400">
                  {{ formatHeaderMoney(availableBalance) }}
                </div>
                <div v-if="frozenBalance > 0" class="mt-1 text-xs text-amber-600 dark:text-amber-300">
                  {{ balanceFrozenText }} {{ formatHeaderMoney(frozenBalance) }}
                </div>
              </div>

              <div class="border-b border-gray-100 py-1 dark:border-dark-700 xl:hidden">
                <button
                  data-testid="daily-check-in-mobile"
                  type="button"
                  class="flex h-10 w-full items-center gap-2 px-4 text-left text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-emerald-500 disabled:cursor-default"
                  :class="checkInMobileButtonClass"
                  :disabled="checkInDisabled"
                  :aria-busy="checkInBusy"
                  :title="checkInTitle"
                  aria-live="polite"
                  @click.stop="handleCheckInAction"
                >
                  <Icon
                    :name="checkInIcon"
                    size="sm"
                    :class="{ 'animate-spin': checkInBusy }"
                  />
                  <span class="min-w-0 flex-1 truncate">{{ checkInLabel }}</span>
                </button>
              </div>

              <div class="py-1">
                <router-link to="/profile" @click="closeDropdown" class="dropdown-item">
                  <Icon name="user" size="sm" />
                  {{ t('nav.profile') }}
                </router-link>

                <router-link to="/keys" @click="closeDropdown" class="dropdown-item">
                  <Icon name="key" size="sm" />
                  {{ t('nav.apiKeys') }}
                </router-link>

                <a
                  v-if="authStore.isAdmin"
                  href="https://github.com/Wei-Shaw/sub2api"
                  target="_blank"
                  rel="noopener noreferrer"
                  @click="closeDropdown"
                  class="dropdown-item"
                >
                  <svg class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
                    <path
                      fill-rule="evenodd"
                      clip-rule="evenodd"
                      d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.17 6.839 9.49.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.604-3.369-1.34-3.369-1.34-.454-1.156-1.11-1.464-1.11-1.464-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.167 22 16.418 22 12c0-5.523-4.477-10-10-10z"
                    />
                  </svg>
                  {{ t('nav.github') }}
                </a>

              </div>

              <!-- Contact Support (only show if configured) -->
              <div
                v-if="contactInfo"
                class="border-t border-gray-100 px-4 py-2.5 dark:border-dark-700"
              >
                <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                  <svg
                    class="h-3.5 w-3.5 flex-shrink-0"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976 1.057-1.976 2.192v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155"
                    />
                  </svg>
                  <span>{{ t('common.contactSupport') }}:</span>
                  <span class="font-medium text-gray-700 dark:text-gray-300">{{
                    contactInfo
                  }}</span>
                </div>
              </div>

              <div v-if="showOnboardingButton" class="border-t border-gray-100 py-1 dark:border-dark-700">
                <button @click="handleReplayGuide" class="dropdown-item w-full">
                  <svg class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
                    <path
                      d="M12 2a10 10 0 100 20 10 10 0 000-20zm0 14a1 1 0 110 2 1 1 0 010-2zm1.07-7.75c0-.6-.49-1.25-1.32-1.25-.7 0-1.22.4-1.43 1.02a1 1 0 11-1.9-.62A3.41 3.41 0 0111.8 5c2.02 0 3.25 1.4 3.25 2.9 0 2-1.83 2.55-2.43 3.12-.43.4-.47.75-.47 1.23a1 1 0 01-2 0c0-1 .16-1.82 1.1-2.7.69-.64 1.82-1.05 1.82-2.06z"
                    />
                  </svg>
                  {{ $t('onboarding.restartTour') }}
                </button>
              </div>

              <div class="border-t border-gray-100 py-1 dark:border-dark-700">
                <button
                  @click="handleLogout"
                  class="dropdown-item w-full text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
                >
                  <svg
                    class="h-4 w-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75"
                    />
                  </svg>
                  {{ t('nav.logout') }}
                </button>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import { resolveUserDisplayName } from '@/utils/userDisplay'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import SubscriptionProgressMini from '@/components/common/SubscriptionProgressMini.vue'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { createFeatureReturnNavigation } from '@/utils/featureReturnNavigation'
import { useDailyCheckIn } from '@/composables/useDailyCheckIn'

interface HeaderQuote {
  text: string
  author?: string
  source: string
}

const fallbackQuotes: HeaderQuote[] = [
  { text: '今天也要把小问题打包成小蛋糕。', author: 'MiraiAPI', source: 'Local' },
  { text: '请求先排队，快乐不排队。', author: 'MiraiAPI', source: 'Local' },
  { text: '缓存命中时，连风都轻了一点。', author: 'MiraiAPI', source: 'Local' },
  { text: '把 bug 放进日志里，像收服一只小怪。', author: 'MiraiAPI', source: 'Local' },
  { text: '今日份魔法值已续杯，继续开工。', author: 'MiraiAPI', source: 'Local' },
  { text: '别慌，进度条也在努力奔跑。', author: 'MiraiAPI', source: 'Local' },
  { text: '服务器在值班，猫耳耳机先戴好。', author: 'MiraiAPI', source: 'Local' },
  { text: '愿你的请求一路绿灯，延迟乖乖坐下。', author: 'MiraiAPI', source: 'Local' },
  { text: '今天的 API 也在认真发光。', author: 'MiraiAPI', source: 'Local' },
  { text: '小小参数，大大冒险。', author: 'MiraiAPI', source: 'Local' },
  { text: '先喝口水，再和报错讲道理。', author: 'MiraiAPI', source: 'Local' },
  { text: '世界很大，先把这一页跑通。', author: 'MiraiAPI', source: 'Local' }
]

function pickFallbackQuote(): HeaderQuote {
  return fallbackQuotes[Math.floor(Math.random() * fallbackQuotes.length)]
}

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()
const onboardingStore = useOnboardingStore()
const {
  phase: checkInPhase,
  status: checkInStatus,
  handleAction: handleCheckInAction,
} = useDailyCheckIn({ authStore, appStore })

const user = computed(() => authStore.user)
const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const contactInfo = computed(() => appStore.contactInfo)
const docUrl = computed(() => sanitizeUrl(appStore.docUrl))
const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')
const randomQuote = ref<HeaderQuote>(pickFallbackQuote())
const quoteText = ref(randomQuote.value.text)
const quoteLoading = ref(false)
let quoteLoadSeq = 0
let quoteRefreshTimer: number | undefined
const availableBalance = computed(() => Number(user.value?.balance || 0))
const frozenBalance = computed(() => Number(user.value?.frozen_balance || 0))
const totalBalance = computed(() => availableBalance.value + frozenBalance.value)
const balanceAvailableText = computed(() => t('common.availableBalance') === 'common.availableBalance' ? '可用余额' : t('common.availableBalance'))
const balanceFrozenText = computed(() => t('common.frozenBalance') === 'common.frozenBalance' ? '冻结金额' : t('common.frozenBalance'))
const balanceTotalText = computed(() => t('common.totalBalance') === 'common.totalBalance' ? '总余额' : t('common.totalBalance'))
const balanceFrozenLabel = computed(() => `${balanceFrozenText.value} ${formatHeaderMoney(frozenBalance.value)}`)
const checkInBusy = computed(() => checkInPhase.value === 'loading' || checkInPhase.value === 'claiming')
const checkInDisabled = computed(() => checkInBusy.value || checkInPhase.value === 'claimed')
const checkInIcon = computed(() => {
  switch (checkInPhase.value) {
    case 'loading':
    case 'claiming':
      return 'refresh'
    case 'claimed':
      return 'checkCircle'
    case 'error':
      return 'exclamationCircle'
    case 'available':
      return 'gift'
  }
  return 'gift'
})
const checkInLabel = computed(() => {
  switch (checkInPhase.value) {
    case 'loading':
      return t('checkIn.loading')
    case 'claiming':
      return t('checkIn.claiming')
    case 'claimed':
      return t('checkIn.claimed', { amount: formatCheckInReward(checkInStatus.value?.reward_amount ?? 0) })
    case 'error':
      return t('checkIn.error')
    case 'available':
      return t('checkIn.available')
  }
  return t('checkIn.available')
})
const checkInTitle = computed(() => {
  switch (checkInPhase.value) {
    case 'claimed':
      return t('checkIn.claimedTitle', { amount: formatCheckInReward(checkInStatus.value?.reward_amount ?? 0) })
    case 'error':
      return t('checkIn.errorTitle')
    case 'available':
      return t('checkIn.availableTitle')
    case 'loading':
      return t('checkIn.loading')
    case 'claiming':
      return t('checkIn.claiming')
  }
  return t('checkIn.availableTitle')
})
const checkInButtonClass = computed(() => {
  switch (checkInPhase.value) {
    case 'available':
      return 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100 dark:bg-emerald-900/25 dark:text-emerald-300 dark:hover:bg-emerald-900/40'
    case 'claimed':
      return 'bg-gray-100 text-gray-600 dark:bg-dark-800 dark:text-dark-300'
    case 'error':
      return 'bg-red-50 text-red-700 hover:bg-red-100 dark:bg-red-900/25 dark:text-red-300 dark:hover:bg-red-900/40'
    case 'loading':
    case 'claiming':
      return 'bg-gray-50 text-gray-500 dark:bg-dark-800 dark:text-dark-400'
  }
  return 'bg-gray-50 text-gray-500 dark:bg-dark-800 dark:text-dark-400'
})
const checkInMobileButtonClass = computed(() => {
  switch (checkInPhase.value) {
    case 'available':
      return 'text-emerald-700 hover:bg-emerald-50 dark:text-emerald-300 dark:hover:bg-emerald-900/20'
    case 'claimed':
      return 'text-gray-600 dark:text-dark-300'
    case 'error':
      return 'text-red-700 hover:bg-red-50 dark:text-red-300 dark:hover:bg-red-900/20'
    case 'loading':
    case 'claiming':
      return 'text-gray-500 dark:text-dark-400'
  }
  return 'text-gray-500 dark:text-dark-400'
})

// 只在标准模式的管理员下显示新手引导按钮

const showOnboardingButton = computed(() => {
  return !authStore.isSimpleMode && user.value?.role === 'admin'
})

function isGptImagePath(path: string): boolean {
  return path === '/images' || path.startsWith('/images/') || path.startsWith('/image-management')
}

function isModelPlazaPath(path: string): boolean {
  return path === '/model-plaza' || path.startsWith('/model-plaza/')
}

const fallbackFeaturePath = () => authStore.isAdmin ? '/admin/dashboard' : '/dashboard'
const gptImageNavigation = createFeatureReturnNavigation({
  storageKey: 'sub2api:gpt-image-return-path',
  isFeaturePath: isGptImagePath,
  entryPath: '/images',
  fallbackPath: fallbackFeaturePath,
})
const modelPlazaNavigation = createFeatureReturnNavigation({
  storageKey: 'sub2api:model-plaza-return-path',
  isFeaturePath: isModelPlazaPath,
  entryPath: '/model-plaza',
  fallbackPath: fallbackFeaturePath,
})
const isGptImageActive = computed(() => gptImageNavigation.isActive(route.path))
const isModelPlazaActive = computed(() => modelPlazaNavigation.isActive(route.path))
const modelPlazaEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.modelPlaza))

const quoteTitle = computed(() => {
  const author = randomQuote.value.author ? ` - ${randomQuote.value.author}` : ''
  return `${randomQuote.value.text}${author} (${randomQuote.value.source})`
})

function setRandomQuote(quote: HeaderQuote) {
  randomQuote.value = quote
  quoteText.value = quote.text
}

const userInitials = computed(() => {
  if (!user.value) return ''
  return displayName.value.substring(0, 2).toUpperCase()
})

const displayName = computed(() => {
  if (!user.value) return ''
  return resolveUserDisplayName(user.value)
})

const pageTitle = computed(() => {
  // For custom pages, use the menu item's label instead of generic "自定义页面"
  if (route.name === 'CustomPage') {
    const id = route.params.id as string
    const publicItems = appStore.cachedPublicSettings?.custom_menu_items ?? []
    const menuItem = publicItems.find((item) => item.id === id)
      ?? (authStore.isAdmin ? adminSettingsStore.customMenuItems.find((item) => item.id === id) : undefined)
    if (menuItem?.label) return menuItem.label
  }
  const titleKey = route.meta.titleKey as string
  if (titleKey) {
    return t(titleKey)
  }
  return (route.meta.title as string) || ''
})

async function loadRandomQuote() {
  const currentSeq = ++quoteLoadSeq
  quoteLoading.value = true

  try {
    if (currentSeq === quoteLoadSeq) {
      setRandomQuote(pickFallbackQuote())
    }
  } finally {
    if (currentSeq === quoteLoadSeq) {
      quoteLoading.value = false
    }
  }
}

function toggleMobileSidebar() {
  appStore.toggleMobileSidebar()
}

function toggleDropdown() {
  dropdownOpen.value = !dropdownOpen.value
}

function closeDropdown() {
  dropdownOpen.value = false
}

async function handleGptImageClick() {
  const wasActive = isGptImageActive.value
  const target = gptImageNavigation.getToggleTarget(route.fullPath)
  await (wasActive ? router.replace(target) : router.push(target))
}

async function handleModelPlazaClick() {
  const wasActive = isModelPlazaActive.value
  const target = modelPlazaNavigation.getToggleTarget(route.fullPath)
  await (wasActive ? router.replace(target) : router.push(target))
}

async function handleLogout() {
  closeDropdown()
  try {
    await authStore.logout()
  } catch (error) {
    // Ignore logout errors - still redirect to login
    console.error('Logout error:', error)
  }
  await router.push('/login')
}

function handleReplayGuide() {
  closeDropdown()
  onboardingStore.replay()
}

function formatHeaderMoney(value: number) {
  if (!Number.isFinite(value)) return '$0.00'
  return `$${value.toFixed(2)}`
}

function formatCheckInReward(value: number) {
  return Number.isInteger(value) ? String(value) : value.toFixed(2)
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    closeDropdown()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  loadRandomQuote()
  quoteRefreshTimer = window.setInterval(loadRandomQuote, 30000)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  if (quoteRefreshTimer !== undefined) {
    window.clearInterval(quoteRefreshTimer)
  }
})
</script>

<style scoped>
.header-quote {
  display: none;
  min-width: 0;
  max-width: min(34rem, 42vw);
  align-items: center;
  border: 0;
  background: transparent;
  padding: 0;
  text-align: left;
  transition:
    color 0.18s ease,
    opacity 0.18s ease;
}

.header-quote::before {
  content: '';
  display: inline-block;
  height: 1rem;
  width: 1px;
  flex: 0 0 auto;
  margin-right: 0.75rem;
  background: rgb(226 232 240);
}

.header-quote:hover {
  opacity: 0.78;
}

.header-quote-text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.8125rem;
  font-weight: 400;
  color: rgb(100 116 139);
}

.header-quote-loading {
  opacity: 0.55;
}

.balance-chip {
  min-height: 2.25rem;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.5rem;
  background: rgb(255 255 255);
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.035);
  padding: 0 0.75rem 0 0.625rem;
}

.balance-chip-icon {
  display: inline-flex;
  height: 1.25rem;
  width: 1.25rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 0.3125rem;
  background: rgb(240 253 250);
  color: rgb(15 118 110);
}

.balance-chip-amount {
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1;
  color: rgb(15 118 110);
  font-variant-numeric: tabular-nums;
}

.dark .header-quote {
  background: transparent;
}

.dark .header-quote::before {
  background: rgb(71 85 105);
}

.dark .header-quote-text {
  color: rgb(203 213 225);
}

.dark .balance-chip {
  border-color: rgb(51 65 85);
  background: rgb(15 23 42);
  box-shadow: none;
}

.dark .balance-chip-icon {
  background: rgba(45, 212, 191, 0.13);
  color: rgb(45 212 191);
}

.dark .balance-chip-amount {
  color: rgb(94 234 212);
}

@media (min-width: 1280px) {
  .header-quote {
    display: flex;
  }
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
